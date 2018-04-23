package shell

import (
	"testing"

	"github.com/hashicorp/terraform/helper/resource"
)


func TestAccGenericShellProviderCRUD_Update(t *testing.T) {
	const testConfig1 = `
	provider "shell" {
		create_command = "echo -n \"{{.output}}\" > {{.file}}"
		read_command = "awk '{print \"out=\" $0}' {{.file}}"
		update_command = "rm {{.old.file}}; echo -n \"{{.new.output}}\" > {{.new.file}}"
		delete_command = "rm {{.file}}"
	}
	resource "shell_crud" "test" {
		context {
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
		delete_command = "rm {{.file}}"
	}
	resource "shell_crud" "test" {
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
					testAccCheckResource("shell_crud.test", "out", "hi"),
				),
			},
			{
				Config: testConfig2,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckResource("shell_crud.test", "out", "hi all"),
				),
			},
		},
	})
}

func TestAccGenericShellProviderCRUD_DefaultUpdate(t *testing.T) {
	const testConfig1 = `
	provider "shell" {
		create_command = "echo -n \"{{.output}}\" > {{.file}}"
		read_command = "awk '{print \"out=\" $0}' {{.file}}"
		delete_command = "rm {{.file}}"
	}
	resource "shell_crud" "test" {
		context {
			output = "hi"
			file = "testfileU1"
		}
	}
`
	const testConfig2 = `
	provider "shell" {
		create_command = "echo -n \"{{.output}}\" > {{.file}}"
		read_command = "awk '{print \"out=\" $0}' {{.file}}"
		delete_command = "rm {{.file}}"
	}
	locals {
		test = "hi all"
	}
	resource "shell_crud" "test" {
		context {
			output = "${local.test}"
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
					testAccCheckResource("shell_crud.test", "out", "hi"),
				),
			},
			{
				Config: testConfig2,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckResource("shell_crud.test", "out", "hi all"),
				),
			},
		},
	})
}
