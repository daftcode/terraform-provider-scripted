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
		commands_create = "echo -n \"hi\" > test_file"
		commands_read = "echo -n \"out=$(cat test_file)\""
		commands_delete = "rm test_file"
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

func TestAccScriptedResource_Prefix(t *testing.T) {
	const testConfig = `
	provider "scripted" {
		commands_prefix = "true"
		commands_create = "echo -n \"hi\" > test_file"
		commands_read = "echo -n \"out=$(cat test_file)\""
		commands_delete = "rm test_file"
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

func TestAccScriptedResource_TwoStageApply(t *testing.T) {
	const testConfig = `
provider "scripted" {
	commands_dependencies = <<EOF
{{- if .Cur.dependency_path -}}
[ -f {{ .Cur.dependency_path | quote }} ] && echo -n true || echo -n false
{{- else -}}
echo -n true
{{- end -}}
EOF
	commands_create = "echo -n {{ .Cur.content | quote }} > {{ .Cur.path }}"
	commands_read = "echo \"out=$(cat {{ .Cur.path | quote }})\""
	commands_delete = "rm {{ .Cur.path | quote }}"
}
resource "scripted_resource" "dependency" {
	context {
		path = "dependency"
		content = "dependency"
	}
}
resource "scripted_resource" "test" {
	context {
		path = "test_file"
		content = "hi"
		dependency_path = "dependency"
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
					testAccCheckResourceOutput("scripted_resource.dependency", "out", "dependency"),
				),
				ExpectNonEmptyPlan: true,
			},
			{
				Config: testConfig,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckResourceOutput("scripted_resource.test", "out", "hi"),
				),
				ExpectNonEmptyPlan: false,
			},
		},
	})
}

func TestAccScriptedResource_StagedDeployment(t *testing.T) {
	const testConfig = `
resource "scripted_resource" "stage1" {
	context {
		path = "test_file"
	}
}

provider "scripted" {
	commands_needs_delete = "[ -n '${var.stage1}' ] || echo -n true"
	commands_create = "echo -n '%s' > {{ .Cur.path }}"
	commands_read = "echo \"out=$(cat {{ .Cur.path | quote }})\""
	commands_delete = "rm {{ .Cur.path | quote }}"
}

variable "stage1" {
	default = "%s"
}
`

	resource.Test(t, resource.TestCase{
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckScriptedDestroy,
		Steps: []resource.TestStep{
			{
				PreConfig:          func() { Stdout.WriteString(">>>>>>>>>>>> step 0\n") },
				Config:             fmt.Sprintf(testConfig, "hi", ""),
				PlanOnly:           true,
				ExpectNonEmptyPlan: true,
			},
			{
				PreConfig: func() { Stdout.WriteString(">>>>>>>>>>>> step 1\n") },
				Config:    fmt.Sprintf(testConfig, "hi", "1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckResourceOutput("scripted_resource.stage1", "out", "hi"),
				),
				ExpectNonEmptyPlan: false,
			},
			{
				PreConfig:          func() { Stdout.WriteString(">>>>>>>>>>>> step 2\n") },
				Config:             fmt.Sprintf(testConfig, "hi", "1"),
				PlanOnly:           true,
				ExpectNonEmptyPlan: false,
			},
			{
				PreConfig:          func() { Stdout.WriteString(">>>>>>>>>>>> step 3\n") },
				Config:             fmt.Sprintf(testConfig, "hi2", "1"),
				PlanOnly:           true,
				ExpectNonEmptyPlan: false,
			},
			{
				PreConfig: func() { Stdout.WriteString(">>>>>>>>>>>> step 4\n") },
				Config:    fmt.Sprintf(testConfig, "hi2", ""),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckResourceMissing("scripted_resource.stage1"),
				),
				ExpectNonEmptyPlan: true,
			},
			{
				PreConfig:          func() { Stdout.WriteString(">>>>>>>>>>>> step 5\n") },
				Config:             fmt.Sprintf(testConfig, "hi", ""),
				PlanOnly:           true,
				ExpectNonEmptyPlan: true,
			},
			{
				PreConfig: func() { Stdout.WriteString(">>>>>>>>>>>> step 6\n") },
				Config:    fmt.Sprintf(testConfig, "hi", "5"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckResourceOutput("scripted_resource.stage1", "out", "hi"),
				),
				ExpectNonEmptyPlan: false,
			},
			{
				PreConfig:          func() { Stdout.WriteString(">>>>>>>>>>>> step 7\n") },
				Config:             fmt.Sprintf(testConfig, "hi", "6"),
				PlanOnly:           true,
				ExpectNonEmptyPlan: false,
			},
		},
	})
}

func TestAccScriptedResource_NeedsUpdate(t *testing.T) {
	const testConfig = `
	provider "scripted" {
		alias = "file"
		commands_needs_update = <<EOF
[ "$(cat '{{ .Cur.path }}')" == {{ .Cur.content | quote }} ] || echo -n true
EOF
		commands_create = "echo -n {{ .Cur.content | quote }} > '{{ .Cur.path }}'"
		commands_read = "echo -n \"out=$(cat '{{ .Cur.path }}')\""
		commands_delete = "rm '{{ .Cur.path }}'"
	}
	resource "scripted_resource" "test" {
		provider = "scripted.file"
		context {
			path = "test_file"
			content = "hi"
		}
	}
`
	const testConfigNotCurrent = `
	provider "scripted" {
		alias = "file"
		commands_needs_update = "echo -n true"
		commands_create = "echo -n {{ .Cur.content | quote }} > '{{ .Cur.path }}'"
		commands_read = "echo -n \"out=$(cat '{{ .Cur.path }}')\""
		commands_delete = "rm '{{ .Cur.path }}'"
	}
	resource "scripted_resource" "test" {
		provider = "scripted.file"
		context {
			path = "test_file"
			content = "hi"
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
					testAccCheckResourceOutput("scripted_resource.test", "out", "hi"),
				),
			},
			{
				Config:             testConfig,
				PlanOnly:           true,
				ExpectNonEmptyPlan: false,
			},
			{
				Config:             testConfigNotCurrent,
				PlanOnly:           true,
				ExpectNonEmptyPlan: true,
			},
			{
				Config:             testConfig,
				PlanOnly:           true,
				ExpectNonEmptyPlan: false,
			},
		},
	})
}

func TestAccScriptedResource_IdCommand(t *testing.T) {
	const testConfig = `
	provider "scripted" {
		commands_id = "echo -n 'test-id'"
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
		commands_create = "echo -n \"hi\" > test_file"
  		commands_read = "echo -n \"out=$(base64 'test_file')\""
		commands_read_format = "base64"
		commands_delete = "rm test_file"
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
		commands_create = "echo -n \"hi\" > test_file"
  		commands_read = "echo -n \"PREFIX_out=$(cat 'test_file')\""
		commands_read_line_prefix =  "PREFIX_"
		commands_delete = "rm test_file"
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
		commands_create = "echo -n \" can you = read this\" > test_file3"
  		commands_read = "echo -n \"out=$(cat 'test_file3')\""
		commands_delete = "rm test_file3"
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
		commands_create = "echo -n \"{{.New.output}}\" > {{.New.file}}"
  		commands_read = "echo -n \"out=$(cat '{{.New.file}}')\""
		commands_delete = "rm {{.Old.file}}"
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
  		commands_read = "echo -n \"out=$test_var\""
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
  		commands_read = "echo -n \"out=$test_var\""
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
  		commands_read = "echo -n \"out=$test_var\""
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
        commands_read_format = "base64"
  		commands_read = <<EOF
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
		commands_environment_include_parent = false
  		commands_read = "env"
	}
	resource "scripted_resource" "test" {
		environment {
			var = "config1"
		}
	}
`
	const testConfig2 = `
	provider "scripted" {
		commands_environment_include_parent = false
		commands_environment_prefix_old = "old_"
		commands_environment_prefix_new = "new_"
  		commands_read = "env"
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
  		commands_create = <<EOF
echo -n "{{ .StatePrefix }}value={{ .Cur.value }}"
EOF
		commands_update = ""
		commands_read = <<EOF
echo old={{ .State.Old.value | quote}}
echo new={{ .State.New.value | quote}}
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
  		commands_create = ""
		commands_update = ""
		commands_read = <<EOF
echo old={{ .State.Old.value | quote}}
echo new={{ .State.New.value | quote}}
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
  		commands_create = ""
		commands_update = <<EOF
echo -n "{{ .StatePrefix }}value={{ .EmptyString }}"
EOF
		commands_read = <<EOF
echo old={{ .State.Old.value | quote}}
echo new={{ .State.New.value | quote}}
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
					testAccCheckResourceOutput("scripted_resource.test", "old", ""),
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
					testAccCheckResourceOutput("scripted_resource.test", "new", ""),
				),
			},
		},
	})
}

func TestAccScriptedResource_JSON(t *testing.T) {
	const testConfig = `
	provider "scripted" {
  		commands_read = <<EOF
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
  		commands_read = <<EOF
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
		commands_read_format = "base64"
  		commands_read = <<EOF
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
		commands_create = "echo -n \"{{.New.output}}\" > {{.New.file}}"
		commands_read = "echo -n \"out=$(cat '{{.New.file}}')\""
		commands_delete = "rm {{.Old.file}}"
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
		commands_create = "echo -n \"{{.New.output}}\" > {{.New.file}}"
		commands_read = "echo -n \"out=$(cat '{{.New.file}}')\""
		commands_delete = "rm {{.Old.file}}"
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
			return fmt.Errorf("state key `%s` should not be present", outparam)
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
func testAccCheckResourceMissing(name string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		_, ok := s.RootModule().Resources[name]
		if ok {
			return fmt.Errorf("resource should not be found: %s", name)
		}
		return nil
	}
}

func testAccCheckResourceOutputMissing(name string, outparam string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("resource not found: %s, found: %s", name, s.RootModule().Resources)
		}
		if rs.Primary.ID == "" {
			return fmt.Errorf("no Record ID is set")
		}

		if _, ok := rs.Primary.Attributes["output."+outparam]; ok {
			return fmt.Errorf("output key `%s` should not be present\n%s", outparam, rs.Primary)
		}
		return nil
	}
}

func testAccCheckScriptedDestroy(s *terraform.State) error {
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "scripted_resource" {
			continue
		}

		splitted := strings.Split(rs.Primary.Attributes["commands_create"], " ")
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
		commands_create = "echo -n \"{{.New.output}}\" > {{.New.file}}"
		commands_read = "echo -n \"out=$(cat '{{.New.file}}')\""
		commands_update = "rm {{.Old.file}}; echo -n \"{{.New.output}}\" > {{.New.file}}"
		commands_delete = "rm {{.Old.file}}"
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
		commands_create = "echo -n \"{{.New.output}}\" > {{.New.file}}"
		commands_read = "echo -n \"out=$(cat '{{.New.file}}')\""
		commands_update = "rm {{.Old.file}}; echo -n \"{{.New.output}}\" > {{.New.file}}"
		commands_delete = "rm {{.Old.file}}"
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
		commands_create = "echo -n \"{{.New.output}}\" > {{.New.file}}"
		commands_read = "echo -n \"out=$(cat '{{.New.file}}')\""
		commands_delete = "rm {{.Old.file}}"
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
		commands_create = "echo -n \"{{.New.output}}\" > {{.New.file}}"
		commands_read = "echo -n \"out=$(cat '{{.New.file}}')\""
		commands_delete = "rm {{.Old.file}}"
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
		commands_create = "echo -n \"{{.New.output}}\" > {{.New.file}}"
		commands_read = "echo -n \"out=$(cat '{{.New.file}}')\""
		commands_exists = "[ -f '{{.New.file}}' ] || echo -n false"
		commands_update = "rm {{.Old.file}}; echo -n \"{{.New.output}}\" > {{.New.file}}"
		commands_delete = "rm {{.Old.file}}"
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
		commands_create = "echo -n \"{{.New.output}}\" > {{.New.file}}"
		commands_read = "echo -n \"out=$(cat '{{.New.file}}')\""
		commands_exists = "[ -f '{{.New.file}}' ] || echo -n false"
		commands_update = "rm {{.Old.file}}; echo -n \"{{.New.output}}\" > {{.New.file}}"
		commands_delete = "rm {{.Old.file}}"
	}
	resource "scripted_resource" "test" {
		context {
			output = "hi all"
			file = "testfileU2"
		}
	}
`
	const testConfig3 = `
	provider "scripted" {
		commands_create = "echo -n \"{{.New.output}}\" > {{.New.file}}"
		commands_read = "echo -n \"out=$(cat '{{.New.file}}')\""
		commands_exists = "echo -n false"
		commands_update = "rm {{.Old.file}}; echo -n \"{{.New.output}}\" > {{.New.file}}"
		commands_delete = "rm {{.Old.file}}"
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
			{
				Config:             testConfig2,
				PlanOnly:           true,
				ExpectNonEmptyPlan: false,
			},
			{
				Config:             testConfig3,
				PlanOnly:           true,
				ExpectNonEmptyPlan: true,
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
