package shell

import (
	"github.com/hashicorp/terraform/helper/schema"
	"github.com/hashicorp/terraform/terraform"
	"runtime"
)

func Provider() terraform.ResourceProvider {
	return &schema.Provider{
		Schema: map[string]*schema.Schema{
			"buffer_size": {
				Type:        schema.TypeInt,
				Optional:    true,
				Default:     1 * 1024 * 1024,
				Description: "stdout and stderr buffer sizes",
			},
			"command_prefix": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "Command prefix shared between all commands",
			},
			"interpreter": {
				Type:     schema.TypeList,
				Optional: true,
				Default: []string{
					"/bin/sh",
					"-c",
				},
				DefaultFunc: func() (interface{}, error) {
					if runtime.GOOS == "windows" {
						return []string{"cmd", "/C"}, nil
					}
					return []string{"/bin/sh", "-c"}, nil
				},
				Description: "Interpreter to use for the commands",
			},
			"command_separator": {
				Type:        schema.TypeString,
				Optional:    true,
				Default:     "\n",
				Description: "Commands separator used in specified interpreter",
			},
			"working_directory": {
				Type:        schema.TypeString,
				Required:    true,
				DefaultFunc: schema.EnvDefaultFunc("PWD", nil),
				Description: "The working directory where to run.",
			},
			"create_command": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "Create command",
			},
			"read_command": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "Read command",
			},
			"read_format": {
				Type:        schema.TypeString,
				Optional:    true,
				Default:     "raw",
				Description: "Read command output type: raw or base64",
			},
			"read_line_prefix": {
				Type:        schema.TypeString,
				Optional:    true,
				Default:     "",
				Description: "Ignore lines without this prefix",
			},
			"update_command": {
				Type:        schema.TypeString,
				Optional:    true,
				Default:     "{{.delete_command}}\n{{.create_command}}",
				Description: "Update command",
			},
			"exists_command": {
				Type:        schema.TypeString,
				Optional:    true,
				Default:     "",
				Description: "Exists command",
			},
			"delete_command": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "Delete command",
			},
		},

		ResourcesMap: map[string]*schema.Resource{
			"shell_crd":   resourceGenericShellCRD(),
			"shell_crud":  resourceGenericShellCRUD(),
			"shell_crude": resourceGenericShellCRUDE(),
		},

		ConfigureFunc: providerConfigure,
	}
}

func providerConfigure(d *schema.ResourceData) (interface{}, error) {
	config := Config{
		CommandPrefix:    d.Get("command_prefix").(string),
		Interpreter:      d.Get("interpreter").([]string),
		CommandSeparator: d.Get("command_separator").(string),
		WorkingDirectory: d.Get("working_directory").(string),
		BufferSize:       int64(d.Get("buffer_size").(int)),
		CreateCommand:    d.Get("create_command").(string),
		ReadCommand:      d.Get("read_command").(string),
		ReadFormat:       d.Get("read_format").(string),
		ReadLinePrefix:   d.Get("read_line_prefix").(string),
		UpdateCommand:    d.Get("update_command").(string),
		DeleteCommand:    d.Get("delete_command").(string),
		ExistsCommand:    d.Get("exists_command").(string),
	}

	return &config, nil
}
