package shell

import (
	"testing"

	"github.com/hashicorp/terraform/helper/resource"
)


func TestAccGenericShellProviderCRUDE_Exists(t *testing.T) {
	const testConfig1 = `
	provider "shell" {
		create_command = "echo -n \"{{.output}}\" > {{.file}}"
		read_command = "awk '{print \"out=\" $0}' {{.file}}"
		update_command = "rm {{.old.file}}; echo -n \"{{.new.output}}\" > {{.new.file}}"
		delete_command = "rm {{.file}}"
	}
	resource "shell_crude" "test" {
		data {
			output = "hi"
			file = "testfileU1"
		}
	}
`
	const testConfig2 = `
	provider "shell" {
		create_command = "echo -n \"{{.output}}\" > {{.file}}"
		read_command = "awk '{print \"out=\" $0}' {{.file}}"
		update_command = "rm {{.old.file}}; echo -n \"{{.new.output}}\" > {{.new.file}}"
		exists_command = "[ -f '{{.file}}'] && echo -n true || echo -n false"
		delete_command = "rm {{.file}}"
	}
	resource "shell_crude" "test" {
		data {
			output = "hi all"
			file = "testfileU2"
		}

		provisioner "local-exec" {
			command = "rm ${self.data.file}"
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
					testAccCheckResource("shell_crude.test", "out", "hi"),
				),
			},
			{
				Config: testConfig2,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckResource("shell_crude.test", "out", "hi all"),
				),
			},
		},
	})
}
