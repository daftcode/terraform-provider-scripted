package scripted

import (
	"testing"

	"fmt"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
	"os"
	"strings"
)

func TestAccScriptedResourceCRD_Basic(t *testing.T) {
	const testConfig = `
	provider "scripted" {
		create_command = "echo -n \"hi\" > test_file"
		read_command = "awk '{print \"out=\" $0}' test_file"
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
					testAccCheckResource("scripted_resource.test", "out", "hi"),
				),
			},
		},
	})
}

func TestAccScriptedResourceCRD_Base64(t *testing.T) {
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
					testAccCheckResource("scripted_resource.test", "out", "hi"),
				),
			},
		},
	})
}

func TestAccScriptedResourceCRD_Prefixed(t *testing.T) {
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
					testAccCheckResource("scripted_resource.test", "out", "hi"),
				),
			},
		},
	})
}

func TestAccScriptedResourceCRD_WeirdOutput(t *testing.T) {
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
					testAccCheckResource("scripted_resource.test", "out", " can you = read this"),
				),
			},
		},
	})
}

func TestAccScriptedResourceCRD_Parameters(t *testing.T) {
	const testConfig = `
	provider "scripted" {
		create_command = "echo -n \"{{.new.output}}\" > {{.new.file}}"
  		read_command = "echo -n \"out=$(cat '{{.new.file}}')\""
		delete_command = "rm {{.old.file}}"
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
					testAccCheckResource("scripted_resource.test", "out", "param value"),
				),
			},
		},
	})
}

func TestAccScriptedResourceCRD_Update(t *testing.T) {
	const testConfig1 = `
	provider "scripted" {
		create_command = "echo -n \"{{.new.output}}\" > {{.new.file}}"
		read_command = "echo -n \"out=$(cat '{{.new.file}}')\""
		delete_command = "rm {{.old.file}}"
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
		create_command = "echo -n \"{{.new.output}}\" > {{.new.file}}"
		read_command = "echo -n \"out=$(cat '{{.new.file}}')\""
		delete_command = "rm {{.old.file}}"
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
					testAccCheckResource("scripted_resource.test", "out", "hi"),
				),
			},
			{
				Config: testConfig2,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckResource("scripted_resource.test", "out", "hi all"),
				),
			},
		},
	})
}

func testAccCheckResource(name string, outparam string, value string) resource.TestCheckFunc {
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
		create_command = "echo -n \"{{.new.output}}\" > {{.new.file}}"
		read_command = "echo -n \"out=$(cat '{{.new.file}}')\""
		update_command = "rm {{.old.file}}; echo -n \"{{.new.output}}\" > {{.new.file}}"
		delete_command = "rm {{.old.file}}"
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
		create_command = "echo -n \"{{.new.output}}\" > {{.new.file}}"
		read_command = "echo -n \"out=$(cat '{{.new.file}}')\""
		update_command = "rm {{.old.file}}; echo -n \"{{.new.output}}\" > {{.new.file}}"
		delete_command = "rm {{.old.file}}"
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
					testAccCheckResource("scripted_resource.test", "out", "hi"),
				),
			},
			{
				Config: testConfig2,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckResource("scripted_resource.test", "out", "hi all"),
				),
			},
		},
	})
}

func TestAccScriptedResourceCRUD_DefaultUpdate(t *testing.T) {
	const testConfig1 = `
	provider "scripted" {
		create_command = "echo -n \"{{.new.output}}\" > {{.new.file}}"
		read_command = "echo -n \"out=$(cat '{{.new.file}}')\""
		delete_command = "rm {{.old.file}}"
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
		create_command = "echo -n \"{{.new.output}}\" > {{.new.file}}"
		read_command = "echo -n \"out=$(cat '{{.new.file}}')\""
		delete_command = "rm {{.old.file}}"
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
					testAccCheckResource("scripted_resource.test", "out", "hi"),
				),
			},
			{
				Config: testConfig2,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckResource("scripted_resource.test", "out", "hi all"),
				),
			},
		},
	})
}
func TestAccScriptedResourceCRUDE_Exists(t *testing.T) {
	const testConfig1 = `
	provider "scripted" {
		create_command = "echo -n \"{{.new.output}}\" > {{.new.file}}"
		read_command = "echo -n \"out=$(cat '{{.new.file}}')\""
		exists_command = "[ -f '{{.new.file}}' ] && exit 0 || exit 1"
		update_command = "rm {{.old.file}}; echo -n \"{{.new.output}}\" > {{.new.file}}"
		delete_command = "rm {{.old.file}}"
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
		create_command = "echo -n \"{{.new.output}}\" > {{.new.file}}"
		read_command = "echo -n \"out=$(cat '{{.new.file}}')\""
		exists_command = "[ -f '{{.new.file}}' ] && exit 0 || exit 1"
		update_command = "rm {{.old.file}}; echo -n \"{{.new.output}}\" > {{.new.file}}"
		delete_command = "rm {{.old.file}}"
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
					testAccCheckResource("scripted_resource.test", "out", "hi"),
				),
			},
			{
				Config: testConfig2,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckResource("scripted_resource.test", "out", "hi all"),
				),
			},
		},
	})
}
