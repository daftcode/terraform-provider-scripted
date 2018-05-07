package custom

import (
	"testing"

	"github.com/hashicorp/terraform/helper/resource"
)

func TestAccCustomProviderCRUDE_Exists(t *testing.T) {
	const testConfig1 = `
	provider "custom" {
		create_command = "echo -n \"{{.new.output}}\" > {{.new.file}}"
		read_command = "echo -n \"out=$(cat '{{.new.file}}')\""
		exists_command = "[ -f '{{.new.file}}' ] && echo -n true || echo -n false"
		update_command = "rm {{.old.file}}; echo -n \"{{.new.output}}\" > {{.new.file}}"
		delete_command = "rm {{.old.file}}"
	}
	resource "custom_crude" "test" {
		context {
			output = "hi"
			file = "testfileU1"
		}
	}
`
	const testConfig2 = `
	provider "custom" {
		create_command = "echo -n \"{{.new.output}}\" > {{.new.file}}"
		read_command = "echo -n \"out=$(cat '{{.new.file}}')\""
		exists_command = "[ -f '{{.new.file}}' ] && echo -n true || echo -n false"
		update_command = "rm {{.old.file}}; echo -n \"{{.new.output}}\" > {{.new.file}}"
		delete_command = "rm {{.old.file}}"
	}
	resource "custom_crude" "test" {
		context {
			output = "hi all"
			file = "testfileU2"
		}
	}
`

	resource.Test(t, resource.TestCase{
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckCustomDestroy,
		Steps: []resource.TestStep{
			{
				Config: testConfig1,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckResource("custom_crude.test", "out", "hi"),
				),
			},
			{
				Config: testConfig2,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckResource("custom_crude.test", "out", "hi all"),
				),
			},
		},
	})
}
