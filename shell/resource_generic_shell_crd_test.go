package shell

import (
	"fmt"
	"os"
	"strings"
	"testing"

	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
)

func TestAccGenericShellProviderCRD_Basic(t *testing.T) {
	const testConfig = `
	provider "shell" {
		create_command = "echo -n \"hi\" > test_file"
		read_command = "awk '{print \"out=\" $0}' test_file"
		delete_command = "rm test_file"
	}
	resource "shell_crd" "test" {
	}
`

	resource.Test(t, resource.TestCase{
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckGenericShellDestroy,
		Steps: []resource.TestStep{
			{
				Config: testConfig,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckResource("shell_crd.test", "out", "hi"),
				),
			},
		},
	})
}

func TestAccGenericShellProviderCRD_Base64(t *testing.T) {
	const testConfig = `
	provider "shell" {
		create_command = "echo -n \"hi\" > test_file"
  		read_command = "echo -n \"out=$(base64 'test_file')\""
		read_format = "base64"
		delete_command = "rm test_file"
	}
	resource "shell_crd" "test" {
	}
`

	resource.Test(t, resource.TestCase{
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckGenericShellDestroy,
		Steps: []resource.TestStep{
			{
				Config: testConfig,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckResource("shell_crd.test", "out", "hi"),
				),
			},
		},
	})
}

func TestAccGenericShellProviderCRD_Prefixed(t *testing.T) {
	const testConfig = `
	provider "shell" {
		create_command = "echo -n \"hi\" > test_file"
  		read_command = "echo -n \"PREFIX_out=$(cat 'test_file')\""
		read_line_prefix =  "PREFIX_"
		delete_command = "rm test_file"
	}
	resource "shell_crd" "test" {
	}
`

	resource.Test(t, resource.TestCase{
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckGenericShellDestroy,
		Steps: []resource.TestStep{
			{
				Config: testConfig,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckResource("shell_crd.test", "out", "hi"),
				),
			},
		},
	})
}

func TestAccGenericShellProviderCRD_WeirdOutput(t *testing.T) {
	const testConfig = `
	provider "shell" {
		create_command = "echo -n \" can you = read this\" > test_file3"
  		read_command = "echo -n \"out=$(cat 'test_file3')\""
		delete_command = "rm test_file3"
	}
	resource "shell_crd" "test" {
	}
`

	resource.Test(t, resource.TestCase{
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckGenericShellDestroy,
		Steps: []resource.TestStep{
			{
				Config: testConfig,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckResource("shell_crd.test", "out", " can you = read this"),
				),
			},
		},
	})
}

func TestAccGenericShellProviderCRD_Parameters(t *testing.T) {
	const testConfig = `
	provider "shell" {
		create_command = "echo -n \"{{.new.output}}\" > {{.new.file}}"
  		read_command = "echo -n \"out=$(cat '{{.new.file}}')\""
		delete_command = "rm {{.old.file}}"
	}
	resource "shell_crd" "test" {
		context {
			output = "param value"
			file = "file4"
		}
	}
`

	resource.Test(t, resource.TestCase{
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckGenericShellDestroy,
		Steps: []resource.TestStep{
			{
				Config: testConfig,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckResource("shell_crd.test", "out", "param value"),
				),
			},
		},
	})
}

func TestAccGenericShellProviderCRD_Update(t *testing.T) {
	const testConfig1 = `
	provider "shell" {
		create_command = "echo -n \"{{.new.output}}\" > {{.new.file}}"
		read_command = "echo -n \"out=$(cat '{{.new.file}}')\""
		delete_command = "rm {{.old.file}}"
	}
	resource "shell_crd" "test" {
		context {
			output = "hi"
			file = "testfileU1"
		}
	}
`
	const testConfig2 = `
	provider "shell" {
		create_command = "echo -n \"{{.new.output}}\" > {{.new.file}}"
		read_command = "echo -n \"out=$(cat '{{.new.file}}')\""
		delete_command = "rm {{.old.file}}"
	}
	resource "shell_crd" "test" {
		context {
			output = "hi all"
			file = "testfileU2"
		}
	}
`

	resource.Test(t, resource.TestCase{
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckGenericShellDestroy,
		Steps: []resource.TestStep{
			{
				Config: testConfig1,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckResource("shell_crd.test", "out", "hi"),
				),
			},
			{
				Config: testConfig2,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckResource("shell_crd.test", "out", "hi all"),
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

func testAccCheckGenericShellDestroy(s *terraform.State) error {
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "shell_crd" {
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
