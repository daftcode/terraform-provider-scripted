package shell

import (
	"fmt"
	"github.com/hashicorp/logutils"
	"github.com/hashicorp/terraform/helper/schema"
	"github.com/hashicorp/terraform/helper/validation"
	"github.com/hashicorp/terraform/terraform"
	"log"
	"os"
	"runtime"
	"strings"
)

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
			"log_output": {
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     false,
				Description: "Should we log command outputs?",
			},
			"interpreter": {
				Type:        schema.TypeList,
				Optional:    true,
				Elem:        &schema.Schema{Type: schema.TypeString},
				Description: "Interpreter to use for the commands",
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
			"read_delete_on_failure": {
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     false,
				Description: "Delete resource when read fails",
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
			"shell_crd":   resourceGenericShellCRD(),
			"shell_crud":  resourceGenericShellCRUD(),
			"shell_crude": resourceGenericShellCRUDE(),
		},

		ConfigureFunc: providerConfigure,
	}
}

func stringToLogLevel(level string) logutils.LogLevel {
	return logutils.LogLevel(level)
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

	//noinspection GoPreferNilSlice
	logLevels := []logutils.LogLevel{}
	for _, level := range ValidLevelsStrings {
		logLevels = append(logLevels, stringToLogLevel(level))
	}

	filter := &logutils.LevelFilter{
		Levels:   logLevels,
		MinLevel: stringToLogLevel(d.Get("log_level").(string)),
		Writer:   os.Stderr,
	}
	logger := log.New(filter, "terraform-provider-shell: ", log.LstdFlags)

	config := Config{
		Logger:              logger,
		LogLevelFilter:      filter,
		LogOutput:           d.Get("log_output").(bool),
		CommandPrefix:       d.Get("command_prefix").(string),
		Interpreter:         interpreter,
		CommandSeparator:    d.Get("command_separator").(string),
		BufferSize:          int64(d.Get("buffer_size").(int)),
		CreateCommand:       d.Get("create_command").(string),
		ReadCommand:         d.Get("read_command").(string),
		ReadDeleteOnFailure: d.Get("read_delete_on_failure").(bool),
		ReadFormat:          d.Get("read_format").(string),
		ReadLinePrefix:      d.Get("read_line_prefix").(string),
		UpdateCommand:       d.Get("update_command").(string),
		DeleteCommand:       d.Get("delete_command").(string),
		ExistsCommand:       d.Get("exists_command").(string),
	}

	return &config, nil
}
