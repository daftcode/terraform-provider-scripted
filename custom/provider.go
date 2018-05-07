package custom

import (
	"fmt"
	"github.com/hashicorp/go-hclog"
	"github.com/hashicorp/terraform/helper/schema"
	"github.com/hashicorp/terraform/helper/validation"
	"github.com/hashicorp/terraform/terraform"
	"os"
	"runtime"
	"strings"
)

// Store original os.Stderr and os.Stdout, because it gets overwritten by go-plugin/server:Serve()
var Stderr = os.Stderr
var Stdout = os.Stdout
var ValidLevelsStrings = []string{"TRACE", "DEBUG", "INFO", "WARN", "ERROR"}

func Provider() terraform.ResourceProvider {
	return &schema.Provider{
		Schema: map[string]*schema.Schema{
			"buffer_size": {
				Type:        schema.TypeInt,
				Optional:    true,
				Default:     1 * 1024 * 1024,
				Description: "stdout and stderr buffer sizes",
			},
			"log_level": {
				Type:         schema.TypeString,
				Optional:     true,
				Default:      "WARN",
				ValidateFunc: validation.StringInSlice(ValidLevelsStrings, true),
				Description:  fmt.Sprintf("Logging level: %s", strings.Join(ValidLevelsStrings, ", ")),
			},
			"command_log_level": {
				Type:         schema.TypeString,
				Optional:     true,
				Default:      "INFO",
				ValidateFunc: validation.StringInSlice(ValidLevelsStrings, true),
				Description:  fmt.Sprintf("Command outputs log level: %s", strings.Join(ValidLevelsStrings, ", ")),
			},
			"command_log_width": {
				Type:        schema.TypeInt,
				Optional:    true,
				Default:     1,
				Description: "Width of command's line to use during formatting.",
			},
			"interpreter": {
				Type:        schema.TypeList,
				Optional:    true,
				Elem:        &schema.Schema{Type: schema.TypeString},
				Description: "Interpreter to use for the commands",
			},
			"working_directory": {
				Type:        schema.TypeString,
				Required:    true,
				DefaultFunc: schema.EnvDefaultFunc("PWD", nil),
				Description: "The working directory where to run.",
			},
			"include_parent_environment": {
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     true,
				Description: "Include parent environment in the command?",
			},
			"command_prefix": {
				Type:        schema.TypeString,
				Optional:    true,
				Default:     "",
				Description: "Command prefix shared between all commands",
			},
			"command_separator": {
				Type:        schema.TypeString,
				Optional:    true,
				Default:     "\n",
				Description: "Commands separator used in specified interpreter",
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
			"delete_on_read_failure": {
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     true,
				Description: "Delete resource when read fails",
			},
			"read_format": {
				Type:         schema.TypeString,
				Optional:     true,
				Default:      "raw",
				ValidateFunc: validation.StringInSlice([]string{"raw", "base64"}, false),
				Description:  "Read command output type: raw or base64",
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
				Default:     "({{.delete_command}})\n({{.create_command}})",
				Description: "Update command, default is: ({{.delete_command}})\\n({{.create_command}})",
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
			"custom_crd":   resourceCustomCRD(),
			"custom_crde":  resourceCustomCRDE(),
			"custom_crud":  resourceCustomCRUD(),
			"custom_crude": resourceCustomCRUDE(),
		},

		ConfigureFunc: providerConfigure,
	}
}

func providerConfigure(d *schema.ResourceData) (interface{}, error) {
	// For some reason setting this via DefaultFunc results in an error
	interpreterI := d.Get("interpreter").([]interface{})
	if len(interpreterI) == 0 {
		if runtime.GOOS == "windows" {
			interpreterI = []interface{}{"cmd", "/C"}
		}
		interpreterI = []interface{}{"/bin/sh", "-c"}
	}
	interpreter := make([]string, len(interpreterI))
	for i, vI := range interpreterI {
		interpreter[i] = vI.(string)
	}
	logger := hclog.New(&hclog.LoggerOptions{
		JSONFormat: true,
		Output:     Stderr,
		Level:      hclog.LevelFromString(d.Get("log_level").(string)),
	})

	config := Config{
		Logger:                   logger,
		CommandLogLevel:          hclog.LevelFromString(d.Get("command_log_level").(string)),
		CommandLogWidth:          d.Get("command_log_width").(int),
		CommandPrefix:            d.Get("command_prefix").(string),
		Interpreter:              interpreter,
		WorkingDirectory:         d.Get("working_directory").(string),
		IncludeParentEnvironment: d.Get("include_parent_environment").(bool),
		CommandSeparator:         d.Get("command_separator").(string),
		BufferSize:               int64(d.Get("buffer_size").(int)),
		CreateCommand:            d.Get("create_command").(string),
		ReadCommand:              d.Get("read_command").(string),
		DeleteOnReadFailure:      d.Get("delete_on_read_failure").(bool),
		ReadFormat:               d.Get("read_format").(string),
		ReadLinePrefix:           d.Get("read_line_prefix").(string),
		UpdateCommand:            d.Get("update_command").(string),
		DeleteCommand:            d.Get("delete_command").(string),
		ExistsCommand:            d.Get("exists_command").(string),
	}

	return &config, nil
}
