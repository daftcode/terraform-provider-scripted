package shell

import (
	"github.com/hashicorp/terraform/helper/schema"
	"github.com/hashicorp/terraform/terraform"
)

func Provider() terraform.ResourceProvider {
	return &schema.Provider{
		Schema: map[string]*schema.Schema{
			"working_directory": &schema.Schema{
				Type:        schema.TypeString,
				Required:    true,
				DefaultFunc: schema.EnvDefaultFunc("PWD", nil),
				Description: "The working directory where to run.",
			},
			"create_command": &schema.Schema{
				Type:        schema.TypeString,
				Required:    true,
				Description: "Create command",
			},
			"read_command": &schema.Schema{
				Type:        schema.TypeString,
				Required:    true,
				Description: "Read command",
			},
			"delete_command": &schema.Schema{
				Type:        schema.TypeString,
				Required:    true,
				Description: "Delete command",
			},
		},

		ResourcesMap: map[string]*schema.Resource{
			"shell_resource": resourceGenericShell(),
		},

		ConfigureFunc: providerConfigure,
	}
}

func providerConfigure(d *schema.ResourceData) (interface{}, error) {
	config := Config{
		WorkingDirectory: d.Get("working_directory").(string),
		CreateCommand:    d.Get("create_command").(string),
		ReadCommand:      d.Get("read_command").(string),
		DeleteCommand:    d.Get("delete_command").(string),
	}

	return &config, nil
}
