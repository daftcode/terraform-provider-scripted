package scripted

import (
	"os"
	"regexp"
	"testing"

	"fmt"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
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
		Providers: testAccProviders,

		Steps: []resource.TestStep{
			{
				Config: testConfig,
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckResourceOutput("scripted_resource.test", "out", "hi"),
				),
			},
		},
	})
}

func TestAccScriptedResource_TemplateFuncInclude(t *testing.T) {
	const testConfig = `
	provider "scripted" {
		commands_read = <<EOF
{{- define "test" }}hi{{ end -}}
echo out={{ include "test" . | quote }}
EOF
	}
	resource "scripted_resource" "test" {
	}
`

	resource.Test(t, resource.TestCase{
		Providers: testAccProviders,

		Steps: []resource.TestStep{
			{
				Config: testConfig,
				Check: resource.ComposeAggregateTestCheckFunc(
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
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testConfig,
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckResourceOutput("scripted_resource.test", "out", "hi"),
				),
			},
		},
	})
}

func TestAccScriptedResource_Terraformify(t *testing.T) {
	const testConfig = `
	provider "scripted" {
        output_format = "json" 
		commands_read = <<EOF
jq -c '{data: .}' <<< "$content"
EOF
		commands_needs_update = <<EOF
set -x
content="$(jq -c "." <<< "$content")"
export output={{ default "" .Output | toJson | squote }}
if jq -e '. == (env.output | fromjson)' <<< "$content" ; then 
  echo {{ .TriggerString | squote }}
fi
EOF
	}
	resource "scripted_resource" "test" {
        environment {
            content = <<EOF
%s
EOF
        }
	}
`
	json := `
{
  "id": 1,
  "true": true,
  "nil": null,
  "person": {
   "name": "John",
   "age": 30
  },
  "cars": [
    {"car1": "Ford"},
    {"car2": "BMW"},
    {"car3": "Fiat"}
  ]
}
`

	d, _ := fromJson(json)
	var checks []resource.TestCheckFunc
	data := terraformify(d)
	for key, value := range data {
		checks = append(
			checks,
			testAccCheckResourceOutput("scripted_resource.test", fmt.Sprintf("data.%s", key), terraformifyPrimitive(value)),
		)
	}
	resource.Test(t, resource.TestCase{
		Providers: testAccProviders,

		Steps: []resource.TestStep{
			{
				Config:             fmt.Sprintf(testConfig, json),
				PlanOnly:           true,
				ExpectNonEmptyPlan: true,
			},
			{
				Config: fmt.Sprintf(testConfig, json),
				Check: resource.ComposeAggregateTestCheckFunc(
					checks...,
				),
			},
			{
				Config:             fmt.Sprintf(testConfig, json),
				PlanOnly:           true,
				ExpectNonEmptyPlan: false,
			},
		},
	})
}

func TestAccScriptedResource_NeedsUpdate(t *testing.T) {
	const testConfigNeverUpdate = `
	provider "scripted" {
		commands_needs_update = ""
	}
	resource "scripted_resource" "test" {}
`
	const testConfigAlwaysUpdate = `
	provider "scripted" {
		commands_needs_update = "echo {{ .TriggerString | squote }}"
	}
	resource "scripted_resource" "test" {}
`
	printStep, _ := stepPrinter()
	resource.Test(t, resource.TestCase{
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				PreConfig: printStep,
				Config:    testConfigNeverUpdate,
			},
			{
				PreConfig:          printStep,
				Config:             testConfigNeverUpdate,
				PlanOnly:           true,
				ExpectNonEmptyPlan: false,
			},
			{
				PreConfig:          printStep,
				Config:             testConfigAlwaysUpdate,
				PlanOnly:           true,
				ExpectNonEmptyPlan: true,
			},
			{
				PreConfig:          printStep,
				Config:             testConfigNeverUpdate,
				PlanOnly:           true,
				ExpectNonEmptyPlan: false,
			},
			{
				PreConfig:          printStep,
				Config:             testConfigAlwaysUpdate,
				PlanOnly:           true,
				ExpectNonEmptyPlan: true,
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
		Providers: testAccProviders,

		Steps: []resource.TestStep{
			{
				Config: testConfig,
				Check: resource.ComposeAggregateTestCheckFunc(
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
		output_format = "base64"
		commands_delete = "rm test_file"
	}
	resource "scripted_resource" "test" {
	}
`

	resource.Test(t, resource.TestCase{
		Providers: testAccProviders,

		Steps: []resource.TestStep{
			{
				Config: testConfig,
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckResourceOutput("scripted_resource.test", "out", "hi"),
				),
			},
		},
	})
}

func TestAccScriptedResource_JsonWithOverride(t *testing.T) {
	const testConfig = `
	provider "scripted" {
		commands_create = "echo -n \"hi\" > test_file"
  		commands_read = <<EOF
	echo -n "{\"out\": \"$(cat test_file)\"}"
	EOF
		output_format = "json"
		commands_delete = "rm test_file"
	}
	resource "scripted_resource" "test" {
	}
`
	const testConfig2 = `
	provider "scripted" {
		commands_create = "echo -n \"hi\" > test_file"
  		commands_read = <<EOF
	echo "{\"out\": \"$(cat test_file)\"}"
	echo '{"out": "hi2"}'
	EOF
		output_format = "json"
		commands_delete = "rm test_file"
	}
	resource "scripted_resource" "test" {
	}
`

	resource.Test(t, resource.TestCase{
		Providers: testAccProviders,

		Steps: []resource.TestStep{
			{
				Config: testConfig,
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckResourceOutput("scripted_resource.test", "out", "hi"),
				),
			},
			{
				Config: testConfig2,
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckResourceOutput("scripted_resource.test", "out", "hi2"),
				),
			},
			{
				Config: testConfig,
				Check: resource.ComposeAggregateTestCheckFunc(
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
		output_line_prefix =  "PREFIX_"
		commands_delete = "rm test_file"
	}
	resource "scripted_resource" "test" {
	}
`

	resource.Test(t, resource.TestCase{
		Providers: testAccProviders,

		Steps: []resource.TestStep{
			{
				Config: testConfig,
				Check: resource.ComposeAggregateTestCheckFunc(
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
		Providers: testAccProviders,

		Steps: []resource.TestStep{
			{
				Config: testConfig,
				Check: resource.ComposeAggregateTestCheckFunc(
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
		Providers: testAccProviders,

		Steps: []resource.TestStep{
			{
				Config: testConfig,
				Check: resource.ComposeAggregateTestCheckFunc(
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
		Providers: testAccProviders,

		Steps: []resource.TestStep{
			{
				Config: testConfig,
				Check: resource.ComposeAggregateTestCheckFunc(
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
	//noinspection SpellCheckingInspection
	resource.Test(t, resource.TestCase{
		Providers: testAccProviders,

		Steps: []resource.TestStep{
			{
				Config: config,
				Check: resource.ComposeAggregateTestCheckFunc(
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
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckResourceOutput("scripted_resource.test", "out", "param value"),
				),
			},
		},
	})
}

func TestAccScriptedResource_MultilineEnvironment(t *testing.T) {
	const testConfig = `
	provider "scripted" {
        output_format = "base64"
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
		Providers: testAccProviders,

		Steps: []resource.TestStep{
			{
				Config: testConfig,
				Check: resource.ComposeAggregateTestCheckFunc(
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
		Providers: testAccProviders,

		Steps: []resource.TestStep{
			{
				Config: testConfig,
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckResourceOutput("scripted_resource.test", "var", "config1"),
				),
			},
			{
				Config: testConfig2,
				Check: resource.ComposeAggregateTestCheckFunc(
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
		Providers: testAccProviders,

		Steps: []resource.TestStep{
			{
				Config: testConfig,
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckResourceState("scripted_resource.test", "value", "test"),
					testAccCheckResourceOutput("scripted_resource.test", "old", "<nil>"),
					testAccCheckResourceOutput("scripted_resource.test", "new", "test"),
				),
			},
			{
				Config: testConfig,
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckResourceState("scripted_resource.test", "value", "test"),
					testAccCheckResourceOutput("scripted_resource.test", "old", "test"),
					testAccCheckResourceOutput("scripted_resource.test", "new", "test"),
				),
			},
			{
				Config: testConfig2,
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckResourceState("scripted_resource.test", "value", "test"),
					testAccCheckResourceOutput("scripted_resource.test", "old", "test"),
					testAccCheckResourceOutput("scripted_resource.test", "new", "test"),
				),
			},
			{
				Config: testConfig3,
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckResourceStateMissing("scripted_resource.test", "value"),
					testAccCheckResourceOutput("scripted_resource.test", "old", "test"),
					testAccCheckResourceOutput("scripted_resource.test", "new", "<nil>"),
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
		Providers: testAccProviders,

		Steps: []resource.TestStep{
			{
				Config: testConfig,
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckResourceOutput("scripted_resource.test", "out", `{"a":[1,2],"b":"pc","d":4}`),
				),
			},
		},
	})
}

func stepPrinter() (func(), resource.TestCheckFunc) {
	step := -1
	out := os.Stderr
	if Debug {
		_, _ = out.WriteString(fmt.Sprintf(">>>>>>>>>>> stepPrinter initialized with %d\n", step))
	}
	return func() {
			step++
			if Debug {
				_, _ = out.WriteString(fmt.Sprintf(">>>>>>>>>>> Step %d\n", step))
			}
		}, func(*terraform.State) error {
			if Debug {
				_, _ = out.WriteString(fmt.Sprintf(">>>>>>>>>>> Step %d check\n", step))
			}
			return nil
		}
}

func TestAccScriptedResource_RollbackDependenciesMet(t *testing.T) {
	const testConfig = `
	provider "scripted" {
  		commands_read = <<EOF
echo out={{ .Cur.val | quote }}
EOF
		commands_update = <<EOF
cat <<EOL | sed 's/^/{{.StatePrefix}}/g'
out={{ .Cur.val }}
EOL
EOF
	}
	resource "scripted_resource" "test" {
		context {
			val = "%v"
		}
	}
`

	const testConfigDoNothing = `
	provider "scripted" {
  		commands_dependencies = "true"
	}
	resource "scripted_resource" "test" {
		context {
			val = "%v"
		}
	}
`
	const testConfigErrUpdate = `
	provider "scripted" {
  		commands_read = <<EOF
echo out={{ .Cur.val | quote }}
EOF
		commands_update = "false"
	}
	resource "scripted_resource" "test" {
		context {
			val = "%v"
		}
	}
`

	printStep, _ := stepPrinter()
	resource.Test(t, resource.TestCase{
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				PreConfig: printStep,
				Config:    fmt.Sprintf(testConfig, "1"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckResourceOutput("scripted_resource.test", "out", `1`),
					testAccCheckResourceState("scripted_resource.test", "out", `1`),
				),
			},
			{
				PreConfig:   printStep,
				Config:      fmt.Sprintf(testConfigErrUpdate, "2"),
				ExpectError: regexp.MustCompile(`.*`),
			},
			{
				PreConfig: printStep,
				Config:    fmt.Sprintf(testConfigDoNothing, "1"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckResourceOutput("scripted_resource.test", "out", `1`),
					testAccCheckResourceState("scripted_resource.test", "out", `1`),
				),
			},
			{
				PreConfig: printStep,
				Config:    fmt.Sprintf(testConfig, "2"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckResourceOutput("scripted_resource.test", "out", `2`),
					testAccCheckResourceState("scripted_resource.test", "out", `2`),
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
		Providers: testAccProviders,

		Steps: []resource.TestStep{
			{
				Config: testConfig,
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckResourceOutput("scripted_resource.test", "out", `[1,2]`),
				),
			},
		},
	})
}

func TestAccScriptedResource_YAML(t *testing.T) {
	const testConfig = `
	provider "scripted" {
		output_format = "base64"
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
		Providers: testAccProviders,

		Steps: []resource.TestStep{
			{
				Config: testConfig,
				Check: resource.ComposeAggregateTestCheckFunc(
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
		Providers: testAccProviders,

		Steps: []resource.TestStep{
			{
				Config: testConfig1,
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckResourceOutput("scripted_resource.test", "out", "hi"),
				),
			},
			{
				Config: testConfig2,
				Check: resource.ComposeAggregateTestCheckFunc(
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
		primary := rs.Primary
		if primary.ID == "" {
			return fmt.Errorf("no Record ID is set")
		}

		if got, ok := primary.Attributes["state."+outparam]; !ok {
			return fmt.Errorf("state key `%s` is missing\n%v", outparam, rs.Primary)
		} else if got != value {
			return fmt.Errorf("wrong value in state `%s`, got %#v instead of %#v\n%v", outparam, got, value, primary)
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
		primary := rs.Primary
		if primary.ID == "" {
			return fmt.Errorf("no Record ID is set")
		}
		if _, ok := primary.Attributes["state."+outparam]; ok {
			return fmt.Errorf("state key `%s` should not be present\n%v", outparam, primary)
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
		primary := rs.Primary
		if primary.ID == "" {
			return fmt.Errorf("no Record ID is set")
		}
		if got, ok := primary.Attributes["output."+outparam]; !ok {
			return fmt.Errorf("output key `%s` is missing\n%v", outparam, rs.Primary)
		} else if got != value {
			return fmt.Errorf("wrong value in output `%s`, got %#v instead of %#v\n%v", outparam, got, value, primary)
		}
		return nil
	}
}

//noinspection GoUnusedFunction
func testAccCheckResourceMissing(name string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		_, ok := s.RootModule().Resources[name]
		if ok {
			return fmt.Errorf("resource should not be found: %s", name)
		}
		return nil
	}
}

//noinspection GoUnusedFunction
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
		Providers: testAccProviders,

		Steps: []resource.TestStep{
			{
				Config: testConfig1,
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckResourceOutput("scripted_resource.test", "out", "hi"),
				),
			},
			{
				Config: testConfig2,
				Check: resource.ComposeAggregateTestCheckFunc(
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
		Providers: testAccProviders,

		Steps: []resource.TestStep{
			{
				Config: testConfig1,
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckResourceOutput("scripted_resource.test", "out", "hi"),
				),
			},
			{
				Config: testConfig2,
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckResourceOutput("scripted_resource.test", "out", "hi all"),
				),
			},
		},
	})
}

func TestAccScriptedResource_Exists(t *testing.T) {
	const testExists = `
	provider "scripted" {
        commands_delete_on_not_exists = false
		commands_exists = ""
	}
	resource "scripted_resource" "test" {}
`
	const testNotExists = `
	provider "scripted" {
        commands_delete_on_not_exists = false
		commands_exists = "echo {{ .TriggerString | squote }}"
	}
	resource "scripted_resource" "test" {}
`

	resource.Test(t, resource.TestCase{
		Providers: testAccProviders,

		Steps: []resource.TestStep{
			{
				Config: testExists,
			},
			{
				Config:             testExists,
				PlanOnly:           true,
				ExpectNonEmptyPlan: false,
			},
			{
				Config:             testNotExists,
				PlanOnly:           true,
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccScriptedResourceCRUDE_Exists(t *testing.T) {
	const testConfig1 = `
	provider "scripted" {
		commands_create = "echo -n \"{{.New.output}}\" > {{.New.file}}"
		commands_read = "echo -n \"out=$(cat '{{.New.file}}')\""
		commands_exists = "[ -f '{{.New.file}}' ] || echo {{ .TriggerString | squote }}"
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
		commands_exists = "[ -f '{{.New.file}}' ] || echo -n {{ .TriggerString | squote }}"
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
		commands_exists = "echo -n {{ .TriggerString | squote }}"
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
		Providers: testAccProviders,

		Steps: []resource.TestStep{
			{
				Config: testConfig1,
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckResourceOutput("scripted_resource.test", "out", "hi"),
				),
			},
			{
				Config: testConfig2,
				Check: resource.ComposeAggregateTestCheckFunc(
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
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckResourceOutput("scripted_resource.test", "out", "hi all"),
				),
			},
		},
	})
}
