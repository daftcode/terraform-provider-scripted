package scripted

import (
	"fmt"
	"github.com/hashicorp/go-hclog"
	"github.com/hashicorp/terraform/helper/schema"
	"github.com/hashicorp/terraform/helper/validation"
	"github.com/hashicorp/terraform/terraform"
	"os"
	"runtime"
	"strconv"
	"strings"
)

// Store original os.Stderr and os.Stdout, because it gets overwritten by go-plugin/server:Serve()
var Stderr = os.Stderr
var Stdout = os.Stdout
var ValidLevelsStrings = []string{"TRACE", "DEBUG", "INFO", "WARN", "ERROR"}

var EmptyString = `ZVaXr3jCd80vqJRhBP9t83LrpWIdNKWJ` // String representing empty value, can be set to anything (eg. generate random each time)

func defaultEmptyString() (interface{}, error) {
	return EmptyString, nil
}

func Provider() terraform.ResourceProvider {
	return &schema.Provider{
		Schema: map[string]*schema.Schema{
			"commands_create": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Create command",
			},
			"commands_delete": {
				Type:        schema.TypeString,
				Optional:    true,
				DefaultFunc: defaultEmptyString,
				Description: "Delete command",
			},
			"commands_delete_on_read_failure": {
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     false,
				Description: "Delete resource when read fails",
			},
			"commands_delete_on_not_exists": {
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     true,
				Description: "Delete resource when exists fails",
			},
			"commands_environment_include_parent": {
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     true,
				Description: "Include parent environment in the command?",
			},
			"commands_environment_prefix_old": {
				Type:        schema.TypeString,
				Optional:    true,
				DefaultFunc: defaultEmptyString,
				Description: "Old environment prefix (skip if empty)",
			},
			"commands_environment_prefix_new": {
				Type:        schema.TypeString,
				Optional:    true,
				DefaultFunc: defaultEmptyString,
				Description: "New environment prefix (skip if empty)",
			},
			"commands_exists": {
				Type:        schema.TypeString,
				Optional:    true,
				DefaultFunc: defaultEmptyString,
				Description: "Exists command",
			},
			"commands_exists_expected_exit_code": {
				Type:        schema.TypeInt,
				Optional:    true,
				Default:     0,
				Description: "Exists command return status",
			},
			"commands_id": {
				Type:        schema.TypeString,
				Optional:    true,
				DefaultFunc: defaultEmptyString,
				Description: "Command building resource id",
			},
			"commands_should_update": {
				Type:        schema.TypeString,
				Optional:    true,
				DefaultFunc: defaultEmptyString,
				Description: "Command indicating whether resource should be updated, non-zero exit code to force update",
			},
			"commands_interpreter": {
				Type:        schema.TypeList,
				Optional:    true,
				Elem:        &schema.Schema{Type: schema.TypeString},
				Description: "Interpreter and it's arguments, can be template with `command` variable.",
			},
			"commands_prefix": {
				Type:        schema.TypeString,
				Optional:    true,
				DefaultFunc: defaultEmptyString,
				Description: "Command prefix shared between all commands",
			},
			"commands_separator": {
				Type:        schema.TypeString,
				Optional:    true,
				Default:     "%s\n%s",
				Description: "Format for joining 2 commands together without isolating them. Defaults to: `%s\\n%s`",
			},
			"commands_read": {
				Type:        schema.TypeString,
				Optional:    true,
				DefaultFunc: defaultEmptyString,
				Description: "Read command",
			},
			"commands_read_format": {
				Type:         schema.TypeString,
				Optional:     true,
				Default:      "raw",
				ValidateFunc: validation.StringInSlice([]string{"raw", "base64"}, false),
				Description:  "Templates output types: raw `/^(?<key>[^=]+)=(?<value>[^\\n]*)$/` or base64 `/^(?<key>[^=]+)=(?<value_base64>[^\\n]*)$/`. Defaults to: `raw`",
			},
			"commands_read_line_prefix": {
				Type:        schema.TypeString,
				Optional:    true,
				DefaultFunc: defaultEmptyString,
				Description: "Ignore lines in read command without this prefix.",
			},
			"commands_state_format": {
				Type:         schema.TypeString,
				Optional:     true,
				DefaultFunc:  defaultEmptyString,
				ValidateFunc: validation.StringInSlice([]string{"raw", "base64", EmptyString}, false),
				Description:  "State format type. Defaults to: `commands_read_format`",
			},
			"commands_update": {
				Type:        schema.TypeString,
				Optional:    true,
				DefaultFunc: defaultEmptyString,
				Description: "Update command. Runs destroy then create by default.",
			},
			"commands_working_directory": {
				Type:        schema.TypeString,
				Optional:    true,
				DefaultFunc: schema.EnvDefaultFunc("PWD", nil),
				Description: "Working directory to run commands in",
			},
			"dependencies": {
				Type:        schema.TypeMap,
				Optional:    true,
				Description: "Dependencies purely for provider graph walking, otherwise ignored.",
			},
			"logging_buffer_size": {
				Type:        schema.TypeInt,
				Optional:    true,
				Default:     1 * 1024 * 1024,
				Description: "stdout and stderr buffer sizes",
			},
			"logging_jsonformat": {
				Type: schema.TypeBool,
				DefaultFunc: func() (interface{}, error) {
					return os.Getenv("TF_SCRIPTED_LOGGING_JSONFORMAT") != "", nil
				},
				Optional:    true,
				Description: "should logs be json instead of plain text? Defaults to: `$TF_SCRIPTED_LOGGING_JSONFORMAT` != \"\"",
			},
			"logging_log_level": {
				Type:         schema.TypeString,
				Optional:     true,
				DefaultFunc:  schema.EnvDefaultFunc("TF_SCRIPTED_LOGGING_LOG_LEVEL", "INFO"),
				ValidateFunc: validation.StringInSlice(ValidLevelsStrings, true),
				Description:  fmt.Sprintf("Logging level: %s. Defaults to: `$TF_SCRIPTED_LOGGING_LOG_LEVEL`", strings.Join(ValidLevelsStrings, ", ")),
			},
			"logging_log_path": {
				Type:        schema.TypeString,
				DefaultFunc: schema.EnvDefaultFunc("TF_SCRIPTED_LOGGING_LOG_PATH", EmptyString),
				Optional:    true,
				Description: "Extra logs output path. Defaults to: `$TF_SCRIPTED_LOGGING_LOG_PATH`",
			},
			"logging_output_logging_log_level": {
				Type:         schema.TypeString,
				Optional:     true,
				DefaultFunc:  schema.EnvDefaultFunc("TF_SCRIPTED_LOGGING_OUTPUT_LOG_LEVEL", "INFO"),
				ValidateFunc: validation.StringInSlice(ValidLevelsStrings, true),
				Description:  fmt.Sprintf("Command stdout/stderr log level: %s. Defaults to: `$TF_SCRIPTED_LOGGING_OUTPUT_LOG_LEVEL`", strings.Join(ValidLevelsStrings, ", ")),
			},
			"logging_output_line_width": {
				Type:        schema.TypeInt,
				Optional:    true,
				Description: "Width of command's line to use during formatting. Defaults to: `$TF_SCRIPTED_LOGGING_OUTPUT_LINE_WIDTH`",
				DefaultFunc: func() (interface{}, error) {
					env, _ := schema.EnvDefaultFunc("TF_SCRIPTED_LOGGING_OUTPUT_LINE_WIDTH", "1")()
					val, err := strconv.Atoi(env.(string))
					return val, err
				},
			},
			"logging_provider_name": {
				Type:        schema.TypeString,
				DefaultFunc: defaultEmptyString,
				Optional:    true,
				Description: "Name to display in log entries for this provider",
			},
			"templates_left_delim": {
				Type:        schema.TypeString,
				DefaultFunc: schema.EnvDefaultFunc("TF_SCRIPTED_TEMPLATES_LEFT_DELIM", "{{"),
				Optional:    true,
				Description: "Left delimiter for templates. Defaults to: `$TF_SCRIPTED_TEMPLATES_LEFT_DELIM` or `{{`",
			},
			"templates_right_delim": {
				Type:        schema.TypeString,
				DefaultFunc: schema.EnvDefaultFunc("TF_SCRIPTED_TEMPLATES_RIGHT_DELIM", "}}"),
				Optional:    true,
				Description: "Right delimiter for templates. Defaults to: `$TF_SCRIPTED_TEMPLATES_RIGHT_DELIM` or `{{`",
			},
		},

		ResourcesMap: map[string]*schema.Resource{
			"scripted_resource": getScriptedResource(),
		},

		DataSourcesMap: map[string]*schema.Resource{
			"scripted_data": getScriptedDataSource(),
		},

		ConfigureFunc: providerConfigure,
	}
}

func providerConfigureLogging(d *schema.ResourceData) (*Logging, error) {
	var hcloggers []hclog.Logger
	logProviderName := d.Get("logging_provider_name").(string)
	logLevel := hclog.LevelFromString(d.Get("logging_log_level").(string))
	logger := hclog.New(&hclog.LoggerOptions{
		JSONFormat: os.Getenv("TF_ACC") == "", // For logging in tests
		Output:     Stderr,
		Level:      logLevel,
	})
	hcloggers = append(hcloggers, logger)

	logPath := d.Get("logging_log_path").(string)
	var fileLogger hclog.Logger
	if logPath != EmptyString {
		logFile, err := os.OpenFile(logPath, os.O_WRONLY|os.O_APPEND|os.O_CREATE, 0644)
		if err != nil {
			return nil, err
		}
		fileLogger = hclog.New(&hclog.LoggerOptions{
			JSONFormat: d.Get("logging_jsonformat").(bool),
			Output:     logFile,
			Level:      logLevel,
		})
		hcloggers = append(hcloggers, fileLogger)
	}

	logging := newLogging(hcloggers)
	if logProviderName != EmptyString {
		logging.Push("provider_name", logProviderName)
	}
	return logging, nil
}

func providerConfigure(d *schema.ResourceData) (interface{}, error) {
	// For some reason setting this via DefaultFunc results in an error
	interpreterI := d.Get("commands_interpreter").([]interface{})
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

	logging, err := providerConfigureLogging(d)

	if err != nil {
		return nil, err
	}

	dbu := false
	cau := false

	update := d.Get("commands_update").(string)
	if update == EmptyString {
		dbu = true
		cau = true
	}
	of := d.Get("commands_read_format").(string)
	sf := d.Get("commands_state_format").(string)
	if sf == EmptyString {
		sf = of
	}
	config := ProviderConfig{
		Commands: &CommandsConfig{
			Environment: &EnvironmentConfig{
				PrefixNew:     d.Get("commands_environment_prefix_new").(string),
				PrefixOld:     d.Get("commands_environment_prefix_old").(string),
				IncludeParent: d.Get("commands_environment_include_parent").(bool),
			},
			Templates: &CommandTemplates{
				Interpreter:  interpreter,
				Prefix:       d.Get("commands_prefix").(string),
				Create:       d.Get("commands_create").(string),
				Delete:       d.Get("commands_delete").(string),
				Exists:       d.Get("commands_exists").(string),
				Id:           d.Get("commands_id").(string),
				ShouldUpdate: d.Get("commands_should_update").(string),
				Read:         d.Get("commands_read").(string),
				Update:       update,
			},
			Output: &OutputConfig{
				LogLevel:  hclog.LevelFromString(d.Get("logging_output_logging_log_level").(string)),
				LineWidth: d.Get("logging_output_line_width").(int),
			},
			CreateAfterUpdate:    cau,
			DeleteBeforeUpdate:   dbu,
			DeleteOnNotExists:    d.Get("commands_delete_on_not_exists").(bool),
			DeleteOnReadFailure:  d.Get("commands_delete_on_read_failure").(bool),
			ExistsExpectedStatus: d.Get("commands_exists_expected_exit_code").(int),
			Separator:            d.Get("commands_separator").(string),
			WorkingDirectory:     d.Get("commands_working_directory").(string),
		},
		Templates: &TemplatesConfig{
			LeftDelim:  d.Get("templates_left_delim").(string),
			RightDelim: d.Get("templates_right_delim").(string),
		},
		EmptyString:       EmptyString,
		Logging:           logging,
		LoggingBufferSize: int64(d.Get("logging_buffer_size").(int)),
		OutputFormat:      of,
		OutputLinePrefix:  d.Get("commands_read_line_prefix").(string),
		StateFormat:       sf,
		StateLinePrefix:   RandomSafeString(32),
	}
	logging.Log(hclog.Info, `Provider "scripted" initialized`)
	return &config, nil
}
