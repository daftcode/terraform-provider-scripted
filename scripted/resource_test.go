package scripted

import (
	"regexp"
	"testing"

	"fmt"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
	"os"
	"strings"
)

func TestAccScriptedResource_BasicCRD(t *testing.T) {
	const testConfig = `
	provider "scripted" {
		create_command = "echo -n \"hi\" > test_file"
		read_command = "echo -n \"out=$(cat test_file)\""
		delete_command = "rm test_file"
	}
	resource "scripted_resource" "test" {
	}
`

	resource.Test(t, resource.TestCase{
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckScriptedDestroy,
		Steps: []resource.TestStep{
			{
				Config: testConfig,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckResourceOutput("scripted_resource.test", "out", "hi"),
				),
			},
		},
	})
}
func TestAccScriptedResource_IdCommand(t *testing.T) {
	const testConfig = `
	provider "scripted" {
		id_command = "echo -n 'test-id'"
	}
	resource "scripted_resource" "test" {
	}
`

	resource.Test(t, resource.TestCase{
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckScriptedDestroy,
		Steps: []resource.TestStep{
			{
				Config: testConfig,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckResourceId("scripted_resource.test", "test-id"),
				),
			},
		},
	})
}

func TestAccScriptedResource_Base64(t *testing.T) {
	const testConfig = `
	provider "scripted" {
		create_command = "echo -n \"hi\" > test_file"
  		read_command = "echo -n \"out=$(base64 'test_file')\""
		read_format = "base64"
		delete_command = "rm test_file"
	}
	resource "scripted_resource" "test" {
	}
`

	resource.Test(t, resource.TestCase{
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckScriptedDestroy,
		Steps: []resource.TestStep{
			{
				Config: testConfig,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckResourceOutput("scripted_resource.test", "out", "hi"),
				),
			},
		},
	})
}

func TestAccScriptedResource_Prefixed(t *testing.T) {
	const testConfig = `
	provider "scripted" {
		create_command = "echo -n \"hi\" > test_file"
  		read_command = "echo -n \"PREFIX_out=$(cat 'test_file')\""
		read_line_prefix =  "PREFIX_"
		delete_command = "rm test_file"
	}
	resource "scripted_resource" "test" {
	}
`

	resource.Test(t, resource.TestCase{
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckScriptedDestroy,
		Steps: []resource.TestStep{
			{
				Config: testConfig,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckResourceOutput("scripted_resource.test", "out", "hi"),
				),
			},
		},
	})
}

func TestAccScriptedResource_WeirdOutput(t *testing.T) {
	const testConfig = `
	provider "scripted" {
		create_command = "echo -n \" can you = read this\" > test_file3"
  		read_command = "echo -n \"out=$(cat 'test_file3')\""
		delete_command = "rm test_file3"
	}
	resource "scripted_resource" "test" {
	}
`

	resource.Test(t, resource.TestCase{
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckScriptedDestroy,
		Steps: []resource.TestStep{
			{
				Config: testConfig,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckResourceOutput("scripted_resource.test", "out", " can you = read this"),
				),
			},
		},
	})
}

func TestAccScriptedResource_Parameters(t *testing.T) {
	const testConfig = `
	provider "scripted" {
		create_command = "echo -n \"{{.New.output}}\" > {{.New.file}}"
  		read_command = "echo -n \"out=$(cat '{{.New.file}}')\""
		delete_command = "rm {{.Old.file}}"
	}
	resource "scripted_resource" "test" {
		context {
			output = "param value"
			file = "file4"
		}
	}
`

	resource.Test(t, resource.TestCase{
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckScriptedDestroy,
		Steps: []resource.TestStep{
			{
				Config: testConfig,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckResourceOutput("scripted_resource.test", "out", "param value"),
				),
			},
		},
	})
}

func TestAccScriptedResource_EnvironmentTemplate(t *testing.T) {
	const testConfig = `
	provider "scripted" {
  		read_command = "echo -n \"out=$test_var\""
	}
	resource "scripted_resource" "test" {
		context {
			output = "param value"
		}
		environment {
			test_var = "{{.Cur.output}}"
		}
	}
`

	resource.Test(t, resource.TestCase{
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckScriptedDestroy,
		Steps: []resource.TestStep{
			{
				Config: testConfig,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckResourceOutput("scripted_resource.test", "out", "param value"),
				),
			},
		},
	})
}

func TestAccScriptedResource_EnvironmentTemplateRecover(t *testing.T) {
	const config = `
	provider "scripted" {
  		read_command = "echo -n \"out=$test_var\""
	}
	resource "scripted_resource" "test" {
		context {
			output = "param value"
		}
		environment {
			test_var = "{{.Cur.output}}"
		}
	}
`
	resource.Test(t, resource.TestCase{
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckScriptedDestroy,
		Steps: []resource.TestStep{
			{
				Config: config,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckResourceOutput("scripted_resource.test", "out", "param value"),
				),
			},
			{
				Config: `
	provider "scripted" {
  		read_command = "echo -n \"out=$test_var\""
	}
	resource "scripted_resource" "test" {
		context {
			output = "param value"
		}
		environment {
			test_var = "{{ ZXVCASEQWSA }}"
		}
	}
`,
				ExpectError: regexp.MustCompile(`.*`),
			},
			{
				Config: config,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckResourceOutput("scripted_resource.test", "out", "param value"),
				),
			},
		},
	})
}

func TestAccScriptedResource_MultilineEnvironment(t *testing.T) {
	const testConfig = `
	provider "scripted" {
        read_format = "base64"
  		read_command = <<EOF
echo -n "out=$(echo -n "$test_var" | base64)"
EOF
	}
	resource "scripted_resource" "test" {
		environment {
			test_var = "line1\nline2"
		}
	}
`

	resource.Test(t, resource.TestCase{
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckScriptedDestroy,
		Steps: []resource.TestStep{
			{
				Config: testConfig,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckResourceOutput("scripted_resource.test", "out", "line1\nline2"),
				),
			},
		},
	})
}

func TestAccScriptedResource_OldNewEnvironment(t *testing.T) {
	const testConfig = `
	provider "scripted" {
		include_parent_environment = false
  		read_command = "env"
	}
	resource "scripted_resource" "test" {
		environment {
			var = "config1"
		}
	}
`
	const testConfig2 = `
	provider "scripted" {
		include_parent_environment = false
		old_environment_prefix = "old_"
		new_environment_prefix = "new_"
  		read_command = "env"
	}
	resource "scripted_resource" "test" {
		environment {
			var = "config2"
		}
	}
`
	resource.Test(t, resource.TestCase{
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckScriptedDestroy,
		Steps: []resource.TestStep{
			{
				Config: testConfig,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckResourceOutput("scripted_resource.test", "var", "config1"),
				),
			},
			{
				Config: testConfig2,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckResourceOutput("scripted_resource.test", "old_var", "config1"),
					testAccCheckResourceOutput("scripted_resource.test", "new_var", "config2"),
					testAccCheckResourceOutput("scripted_resource.test", "var", "config2"),
				),
			},
		},
	})
}
func TestAccScriptedResource_State(t *testing.T) {
	const testConfig = `
	provider "scripted" {
  		create_command = <<EOF
echo -n "{{ .StatePrefix }}value={{ .Cur.value }}"
EOF
		update_command = ""
		read_command = <<EOF
echo "old={{ .State.Old.value }}"
echo "new={{ .State.New.value }}"
EOF
	}
	resource "scripted_resource" "test" {
		context {
			value = "test"
		}
	}
`
	const testConfig2 = `
	provider "scripted" {
		create_before_update = true
  		create_command = ""
		update_command = ""
		read_command = <<EOF
echo "old={{ .State.Old.value }}"
echo "new={{ .State.New.value }}"
EOF
	}
	resource "scripted_resource" "test" {
		context {
			value = "test2"
		}
	}
`
	const testConfig3 = `
	provider "scripted" {
		create_before_update = true
  		create_command = " "
		update_command = ""
		read_command = <<EOF
echo "old={{ .State.Old.value }}"
echo "new={{ .State.New.value }}"
EOF
	}
	resource "scripted_resource" "test" {
		context {
			value = "test3"
		}
	}
`
	resource.Test(t, resource.TestCase{
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckScriptedDestroy,
		Steps: []resource.TestStep{
			{
				Config: testConfig,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckResourceState("scripted_resource.test", "value", "test"),
					testAccCheckResourceOutput("scripted_resource.test", "old", "<no value>"),
					testAccCheckResourceOutput("scripted_resource.test", "new", "test"),
				),
			},
			{
				Config: testConfig,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckResourceState("scripted_resource.test", "value", "test"),
					testAccCheckResourceOutput("scripted_resource.test", "old", "test"),
					testAccCheckResourceOutput("scripted_resource.test", "new", "test"),
				),
			},
			{
				Config: testConfig2,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckResourceState("scripted_resource.test", "value", "test"),
					testAccCheckResourceOutput("scripted_resource.test", "old", "test"),
					testAccCheckResourceOutput("scripted_resource.test", "new", "test"),
				),
			},
			{
				Config: testConfig3,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckResourceStateMissing("scripted_resource.test", "value"),
					testAccCheckResourceOutput("scripted_resource.test", "old", "test"),
					testAccCheckResourceOutput("scripted_resource.test", "new", "<no value>"),
				),
			},
		},
	})
}

func TestAccScriptedResource_JSON(t *testing.T) {
	const testConfig = `
	provider "scripted" {
  		read_command = <<EOF
echo -n 'out={{ toJson (fromJson .Cur.val) }}'
EOF
	}
	resource "scripted_resource" "test" {
		context {
			val = <<EOF
{"a":[1,2],"b":"pc","d":4}
EOF
		}
	}
`

	resource.Test(t, resource.TestCase{
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckScriptedDestroy,
		Steps: []resource.TestStep{
			{
				Config: testConfig,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckResourceOutput("scripted_resource.test", "out", `{"a":[1,2],"b":"pc","d":4}`),
				),
			},
		},
	})
}

func TestAccScriptedResource_JSON_Nested(t *testing.T) {
	const testConfig = `
	provider "scripted" {
  		read_command = <<EOF
{{ $val := fromJson .Cur.val }}
echo -n 'out={{ toJson $val.a }}'
EOF
	}
	resource "scripted_resource" "test" {
		context {
			val = <<EOF
{"a":[1,2],"b":"pc","d":4}
EOF
		}
	}
`

	resource.Test(t, resource.TestCase{
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckScriptedDestroy,
		Steps: []resource.TestStep{
			{
				Config: testConfig,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckResourceOutput("scripted_resource.test", "out", `[1,2]`),
				),
			},
		},
	})
}

func TestAccScriptedResource_YAML(t *testing.T) {
	const testConfig = `
	provider "scripted" {
		read_format = "base64"
  		read_command = <<EOF
echo -n 'out={{ b64enc (toYaml (fromYaml .Cur.val)) }}'
EOF
	}
	resource "scripted_resource" "test" {
		context {
			val = <<EOF
a:
- 1
- 2
b: pc
d: 4
EOF
		}
	}
`

	resource.Test(t, resource.TestCase{
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckScriptedDestroy,
		Steps: []resource.TestStep{
			{
				Config: testConfig,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckResourceOutput("scripted_resource.test", "out", `a:
- 1
- 2
b: pc
d: 4
`),
				),
			},
		},
	})
}

func TestAccScriptedResourceCRD_Update(t *testing.T) {
	const testConfig1 = `
	provider "scripted" {
		create_command = "echo -n \"{{.New.output}}\" > {{.New.file}}"
		read_command = "echo -n \"out=$(cat '{{.New.file}}')\""
		delete_command = "rm {{.Old.file}}"
	}
	resource "scripted_resource" "test" {
		context {
			output = "hi"
			file = "testfileU1"
		}
	}
`
	const testConfig2 = `
	provider "scripted" {
		create_command = "echo -n \"{{.New.output}}\" > {{.New.file}}"
		read_command = "echo -n \"out=$(cat '{{.New.file}}')\""
		delete_command = "rm {{.Old.file}}"
	}
	resource "scripted_resource" "test" {
		context {
			output = "hi all"
			file = "testfileU2"
		}
	}
`

	resource.Test(t, resource.TestCase{
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckScriptedDestroy,
		Steps: []resource.TestStep{
			{
				Config: testConfig1,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckResourceOutput("scripted_resource.test", "out", "hi"),
				),
			},
			{
				Config: testConfig2,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckResourceOutput("scripted_resource.test", "out", "hi all"),
				),
			},
		},
	})
}

func testAccCheckResourceId(name string, id string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("resource not found: %s, found: %s", name, s.RootModule().Resources)
		}
		if rs.Primary.ID != id {
			return fmt.Errorf("id is not right: `%s` != `%s`", rs.Primary.ID, id)
		}
		return nil
	}
}
func testAccCheckResourceState(name string, outparam string, value string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("resource not found: %s, found: %s", name, s.RootModule().Resources)
		}
		if rs.Primary.ID == "" {
			return fmt.Errorf("no Record ID is set")
		}

		if expected, got := value, rs.Primary.Attributes["state."+outparam]; got != expected {
			return fmt.Errorf("wrong value in state '%s=%s', expected '%s'", outparam, got, expected)
		}
		return nil
	}
}
func testAccCheckResourceStateMissing(name string, outparam string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("resource not found: %s, found: %s", name, s.RootModule().Resources)
		}
		if rs.Primary.ID == "" {
			return fmt.Errorf("no Record ID is set")
		}
		if _, ok := rs.Primary.Attributes["state."+outparam]; ok {
			return fmt.Errorf("state key %s is present", outparam)
		}
		return nil
	}
}

func testAccCheckResourceOutput(name string, outparam string, value string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("resource not found: %s, found: %s", name, s.RootModule().Resources)
		}
		if rs.Primary.ID == "" {
			return fmt.Errorf("no Record ID is set")
		}

		if expected, got := value, rs.Primary.Attributes["output."+outparam]; got != expected {
			return fmt.Errorf("wrong value in output '%s=%s', expected '%s'", outparam, got, expected)
		}
		return nil
	}
}

func testAccCheckScriptedDestroy(s *terraform.State) error {
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "scripted_resource" {
			continue
		}

		splitted := strings.Split(rs.Primary.Attributes["create_command"], " ")
		file := splitted[len(splitted)-1]
		if _, err := os.Stat(file); err == nil {
			return fmt.Errorf("file '%s' exists after delete", file)
		}
	}
	return nil
}

func TestAccScriptedResourceCRUD_Update(t *testing.T) {
	const testConfig1 = `
	provider "scripted" {
		create_command = "echo -n \"{{.New.output}}\" > {{.New.file}}"
		read_command = "echo -n \"out=$(cat '{{.New.file}}')\""
		update_command = "rm {{.Old.file}}; echo -n \"{{.New.output}}\" > {{.New.file}}"
		delete_command = "rm {{.Old.file}}"
	}
	resource "scripted_resource" "test" {
		context {
			output = "hi"
			file = "testfileU1"
		}
	}
`
	const testConfig2 = `
	provider "scripted" {
		create_command = "echo -n \"{{.New.output}}\" > {{.New.file}}"
		read_command = "echo -n \"out=$(cat '{{.New.file}}')\""
		update_command = "rm {{.Old.file}}; echo -n \"{{.New.output}}\" > {{.New.file}}"
		delete_command = "rm {{.Old.file}}"
	}
	resource "scripted_resource" "test" {
		context {
			output = "hi all"
			file = "testfileU2"
		}
	}
`

	resource.Test(t, resource.TestCase{
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckScriptedDestroy,
		Steps: []resource.TestStep{
			{
				Config: testConfig1,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckResourceOutput("scripted_resource.test", "out", "hi"),
				),
			},
			{
				Config: testConfig2,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckResourceOutput("scripted_resource.test", "out", "hi all"),
				),
			},
		},
	})
}

func TestAccScriptedResourceCRUD_DefaultUpdate(t *testing.T) {
	const testConfig1 = `
	provider "scripted" {
		create_command = "echo -n \"{{.New.output}}\" > {{.New.file}}"
		read_command = "echo -n \"out=$(cat '{{.New.file}}')\""
		delete_command = "rm {{.Old.file}}"
		log_level = "INFO"
	}
	resource "scripted_resource" "test" {
		context {
			output = "hi"
			file = "testfileU1"
		}
	}
`
	const testConfig2 = `
	provider "scripted" {
		create_command = "echo -n \"{{.New.output}}\" > {{.New.file}}"
		read_command = "echo -n \"out=$(cat '{{.New.file}}')\""
		delete_command = "rm {{.Old.file}}"
		log_level = "INFO"
	}
	locals {
		test = "hi all"
	}
	resource "scripted_resource" "test" {
		context {
			output = "${local.test}"
			file = "testfileU2"
		}
	}
`

	resource.Test(t, resource.TestCase{
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckScriptedDestroy,
		Steps: []resource.TestStep{
			{
				Config: testConfig1,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckResourceOutput("scripted_resource.test", "out", "hi"),
				),
			},
			{
				Config: testConfig2,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckResourceOutput("scripted_resource.test", "out", "hi all"),
				),
			},
		},
	})
}
func TestAccScriptedResourceCRUDE_Exists(t *testing.T) {
	const testConfig1 = `
	provider "scripted" {
		create_command = "echo -n \"{{.New.output}}\" > {{.New.file}}"
		read_command = "echo -n \"out=$(cat '{{.New.file}}')\""
		exists_command = "[ -f '{{.New.file}}' ] && exit 0 || exit 1"
		update_command = "rm {{.Old.file}}; echo -n \"{{.New.output}}\" > {{.New.file}}"
		delete_command = "rm {{.Old.file}}"
	}
	resource "scripted_resource" "test" {
		context {
			output = "hi"
			file = "testfileU1"
		}
	}
`
	const testConfig2 = `
	provider "scripted" {
		create_command = "echo -n \"{{.New.output}}\" > {{.New.file}}"
		read_command = "echo -n \"out=$(cat '{{.New.file}}')\""
		exists_command = "[ -f '{{.New.file}}' ] && exit 0 || exit 1"
		update_command = "rm {{.Old.file}}; echo -n \"{{.New.output}}\" > {{.New.file}}"
		delete_command = "rm {{.Old.file}}"
	}
	resource "scripted_resource" "test" {
		context {
			output = "hi all"
			file = "testfileU2"
		}
	}
`

	resource.Test(t, resource.TestCase{
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckScriptedDestroy,
		Steps: []resource.TestStep{
			{
				Config: testConfig1,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckResourceOutput("scripted_resource.test", "out", "hi"),
				),
			},
			{
				Config: testConfig2,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckResourceOutput("scripted_resource.test", "out", "hi all"),
				),
			},
		},
	})
}
