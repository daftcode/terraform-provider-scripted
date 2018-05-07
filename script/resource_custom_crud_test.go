package script

import (
	"testing"

	"github.com/hashicorp/terraform/helper/resource"
)

func TestAccScriptProviderCRUD_Update(t *testing.T) {
	const testConfig1 = `
	provider "script" {
		create_command = "echo -n \"{{.new.output}}\" > {{.new.file}}"
		read_command = "echo -n \"out=$(cat '{{.new.file}}')\""
		update_command = "rm {{.old.file}}; echo -n \"{{.new.output}}\" > {{.new.file}}"
		delete_command = "rm {{.old.file}}"
	}
	resource "script_crud" "test" {
		context {
			output = "hi"
			file = "testfileU1"
		}
	}
`
	const testConfig2 = `
	provider "script" {
		create_command = "echo -n \"{{.new.output}}\" > {{.new.file}}"
		read_command = "echo -n \"out=$(cat '{{.new.file}}')\""
		update_command = "rm {{.old.file}}; echo -n \"{{.new.output}}\" > {{.new.file}}"
		delete_command = "rm {{.old.file}}"
	}
	resource "script_crud" "test" {
		context {
			output = "hi all"
			file = "testfileU2"
		}
	}
`

	resource.Test(t, resource.TestCase{
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckScriptDestroy,
		Steps: []resource.TestStep{
			{
				Config: testConfig1,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckResource("script_crud.test", "out", "hi"),
				),
			},
			{
				Config: testConfig2,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckResource("script_crud.test", "out", "hi all"),
				),
			},
		},
	})
}

func TestAccScriptProviderCRUD_DefaultUpdate(t *testing.T) {
	const testConfig1 = `
	provider "script" {
		create_command = "echo -n \"{{.new.output}}\" > {{.new.file}}"
		read_command = "echo -n \"out=$(cat '{{.new.file}}')\""
		delete_command = "rm {{.old.file}}"
		log_level = "INFO"
	}
	resource "script_crud" "test" {
		context {
			output = "hi"
			file = "testfileU1"
		}
	}
`
	const testConfig2 = `
	provider "script" {
		create_command = "echo -n \"{{.new.output}}\" > {{.new.file}}"
		read_command = "echo -n \"out=$(cat '{{.new.file}}')\""
		delete_command = "rm {{.old.file}}"
		log_level = "INFO"
	}
	locals {
		test = "hi all"
	}
	resource "script_crud" "test" {
		context {
			output = "${local.test}"
			file = "testfileU2"
		}
	}
`

	resource.Test(t, resource.TestCase{
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckScriptDestroy,
		Steps: []resource.TestStep{
			{
				Config: testConfig1,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckResource("script_crud.test", "out", "hi"),
				),
			},
			{
				Config: testConfig2,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckResource("script_crud.test", "out", "hi all"),
				),
			},
		},
	})
}
