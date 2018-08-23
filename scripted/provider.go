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

var nextProviderId = 1

// Store original os.Stderr and os.Stdout, because it gets overwritten by go-plugin/server:Serve()
var Stderr = os.Stderr
var Stdout = os.Stdout
var ValidLogLevelsStrings = []string{"TRACE", "DEBUG", "INFO", "WARN", "ERROR"}

var DefaultEnvPrefix = "TF_SCRIPTED_"
var EnvPrefix = envDefault("TF_SCRIPTED_ENV_PREFIX", DefaultEnvPrefix)
var Debug = getEnvBool("DEBUG", false)
var debugLogging = false

// String representing empty value, can be set to anything
var EmptyString = getEnvMust("EMPTY_STRING", `ZVaXr3jCd80vqJRhBP9t83LrpWIdNKWJ`)
var DefaultTriggerString = `ndn4VFxYG489bUmV6xKjKFE0RYQIJdts`
var DefaultStatePrefix = `WViRV1TbGAGehAYFL8g3ZL8o1cg1bxaq`
var StatePrefixEnvKey = envKey("STATE_PREFIX")
var TriggerStringEnvKey = envKey("TRIGGER_STRING")
var TriggerStringTpl = `{{ .TriggerString }}`

var defaultWindowsInterpreter = []string{"cmd", "/C", "{{ .command }}"}
var defaultInterpreter = []string{"bash", "-Eeuo", "pipefail", "-c", "{{ .command }}"}

func Provider() terraform.ResourceProvider {
	return &schema.Provider{
		Schema: map[string]*schema.Schema{
			"commands_create": {
				Type:        schema.TypeString,
				Optional:    true,
				DefaultFunc: defaultEmptyString,
				Description: "Create command. Defaults to: `update_command`",
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
			"commands_dependencies": {
				Type:        schema.TypeString,
				Optional:    true,
				DefaultFunc: defaultEmptyString,
				Description: fmt.Sprintf("Command determining whether dependencies are met, dependencies met triggered by `%s`", TriggerStringTpl),
			},
			"commands_environment_include_parent": {
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     false,
				Description: "Include whole parent environment in the command?",
			},
			"commands_environment_inherit_variables": {
				Type:        schema.TypeList,
				Optional:    true,
				Elem:        &schema.Schema{Type: schema.TypeString},
				Description: "List of environment variables to inherit from parent. Defaults to: `$TF_SCRIPTED_ENVIRONMENT_INHERIT_VARIABLES` (JSON array)",
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
				Description: fmt.Sprintf("Exists command, not-exists triggered by `%s`", TriggerStringTpl),
			},
			"commands_id": {
				Type:        schema.TypeString,
				Optional:    true,
				DefaultFunc: defaultEmptyString,
				Description: "Command building resource id",
			},
			"commands_interpreter": {
				Type:     schema.TypeList,
				Optional: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
				Description: func() string {
					dWI, _ := toJson(defaultWindowsInterpreter)
					dI, _ := toJson(defaultInterpreter)
					return fmt.Sprintf(
						"Interpreter and it's arguments, can be a template with `command` variable. "+
							"Defaults to: `$TF_SCRIPTED_COMMANDS_INTERPRETER` (JSON array), `%s` (windows) or `%s`",
						dWI,
						dI,
					)
				}(),
			},
			"commands_modify_prefix": {
				Type:        schema.TypeString,
				Optional:    true,
				DefaultFunc: defaultEmptyString,
				Description: "Modification commands (create and update) prefix",
			},
			"commands_needs_update": {
				Type:        schema.TypeString,
				Optional:    true,
				DefaultFunc: defaultEmptyString,
				Description: fmt.Sprintf("Command indicating whether resource should be updated, update triggered by `%s`", TriggerStringTpl),
			},
			"commands_prefix": {
				Type:        schema.TypeString,
				Optional:    true,
				DefaultFunc: defaultEmptyString,
				Description: "Command prefix shared between all commands",
			},
			"commands_prefix_fromenv": stringDefaultSchemaEmpty(
				nil,
				"commands_prefix_fromenv",
				"Command prefix shared between all commands (added before `commands_prefix`)",
			),
			"commands_separator": stringDefaultSchemaBaseOr(
				nil,
				"commands_separator",
				"Format for joining 2 commands together without isolating them.",
				"%s\n%s",
				"`%s\\n%s`",
			),
			"commands_read": {
				Type:        schema.TypeString,
				Optional:    true,
				DefaultFunc: defaultEmptyString,
				Description: "Read command",
			},
			"commands_read_format": stringDefaultSchema(
				&schema.Schema{
					ValidateFunc: validation.StringInSlice([]string{"raw", "base64", "json"}, false),
				},
				"commands_read_format",
				"Templates output types: "+
					"raw `/^(?<key>[^=]+)=(?<value>[^\\n]*)$/`, "+
					"base64 `/^(?<key>[^=]+)=(?<value_base64>[^\\n]*)$/` or "+
					"one JSON object per line overriding previous keys.",
				"raw",
			),
			"commands_read_line_prefix": stringDefaultSchemaEmpty(
				nil,
				"commands_read_line_prefix",
				"Ignore lines in read command without this prefix.",
			),
			"commands_state_format": stringDefaultSchemaEmptyMsgVal(
				&schema.Schema{
					ValidateFunc: validation.StringInSlice([]string{"raw", "base64", "json", EmptyString}, false),
				},
				"commands_state_format",
				"Create/Update state output format, for more info see `commands_read_format`.",
				"`commands_read_format`",
			),
			"commands_update": {
				Type:        schema.TypeString,
				Optional:    true,
				DefaultFunc: defaultEmptyString,
				Description: "Update command. Deletes then creates if not set. Can be used in place of `create_command`.",
			},
			"commands_working_directory": stringDefaultSchemaEmpty(
				nil,
				"commands_working_directory",
				"Working directory to run commands in",
			),
			"dependencies": {
				Type:        schema.TypeMap,
				Optional:    true,
				Description: "Dependencies purely for provider graph walking, otherwise ignored.",
			},
			"logging_buffer_size": {
				Type:        schema.TypeInt,
				Optional:    true,
				Default:     8 * 1024,
				Description: "output (on error) buffer sizes",
			},
			"logging_jsonformat": boolDefaultSchema(
				nil,
				"logging_jsonformat",
				"should logs be json instead of plain text?",
				true,
			),
			"logging_jsonlist": boolDefaultSchema(
				nil,
				"logging_jsonlist",
				"should json log formatter output lists instead of direct values?",
				false,
			),
			"logging_jsonlistpromote": boolDefaultSchema(
				nil,
				"logging_jsonlistpromote",
				"should json log formatter promote single values to lists and append?",
				true,
			),
			"logging_running_messages_interval": floatDefaultSchema(
				nil,
				"logging_running_messages_interval",
				"should resources report still being in a running state? Trigger reports every N seconds.",
				0,
			),
			"logging_log_level": stringDefaultSchema(
				&schema.Schema{
					ValidateFunc: validation.StringInSlice(ValidLogLevelsStrings, true),
				},
				"log_level",
				fmt.Sprintf("Logging level: %s.", strings.Join(ValidLogLevelsStrings, ", ")),
				"INFO",
			),
			"logging_log_path": stringDefaultSchemaEmpty(
				nil,
				"log_path",
				"Additional logs output path.",
			),
			"logging_output_logging_log_level": stringDefaultSchema(
				&schema.Schema{
					ValidateFunc: validation.StringInSlice(ValidLogLevelsStrings, true),
				},
				"output_log_level",
				fmt.Sprintf("Command stdout/stderr log level: %s.", strings.Join(ValidLogLevelsStrings, ", ")),
				"INFO",
			),
			"logging_output_line_width": {
				Type:        schema.TypeInt,
				Optional:    true,
				Description: defaultMsg("Width of command's line to use during formatting.", "`$LOGGING_OUTPUT_LINE_WIDTH`"),
				DefaultFunc: func() (interface{}, error) {
					env, _ := envDefaultFunc("LOGGING_OUTPUT_LINE_WIDTH", "1")()
					val, err := strconv.Atoi(env.(string))
					return val, err
				},
			},
			"logging_pids": boolDefaultSchema(
				nil,
				"logging_pids",
				"Should output lines contain `ppid` and `pid`?",
				false,
			),
			"logging_iids": boolDefaultSchema(
				nil,
				"logging_iids",
				"Should output lines contain `piid` (provider instance id) and `riid` (resource instance id?",
				false,
			),
			"logging_provider_name": {
				Type:        schema.TypeString,
				DefaultFunc: defaultEmptyString,
				Optional:    true,
				Description: "Name to display in log entries for this provider",
			},
			"templates_left_delim": stringDefaultSchema(
				nil,
				"templates_left_delim",
				"Left delimiter for templates.",
				"{{",
			),
			"templates_right_delim": stringDefaultSchema(
				nil,
				"templates_right_delim",
				"Right delimiter for templates.",
				"}}",
			),
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
	jsonList := d.Get("logging_jsonlist").(bool)
	jsonListPromote := d.Get("logging_jsonlistpromote").(bool)
	logger := hclog.New(&hclog.LoggerOptions{
		JSONFormat:      os.Getenv("TF_ACC") == "", // For logging in tests
		JSONList:        jsonList,
		JSONListPromote: jsonListPromote,
		Output:          Stderr,
		Level:           logLevel,
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
			JSONFormat:      d.Get("logging_jsonformat").(bool),
			JSONList:        jsonList,
			JSONListPromote: jsonListPromote,
			Output:          logFile,
			Level:           logLevel,
		})
		hcloggers = append(hcloggers, fileLogger)
	}

	logging := newLogging(hcloggers)
	logging.level = logLevel
	if d.Get("logging_iids").(bool) {
		logging.Push("piid", nextProviderId)
	}
	nextProviderId++
	if logProviderName != EmptyString {
		logging.Push("provider_name", logProviderName)
	}
	return logging, nil
}

func interpreterOrDefault(cur []string) ([]string, error) {
	var interpreter []string
	var err error
	if len(cur) > 0 {
		return cur, nil
	}
	var defVal []string
	debugInterpreter := getEnvBoolFalse("INTERPRETER_DEBUG")

	if runtime.GOOS == "windows" {
		defVal = defaultWindowsInterpreter
	} else {
		defVal = defaultInterpreter
		if debugInterpreter {
			interpreter = append(defVal, "-x")
		}
	}
	interpreter, _, err = getEnvList("COMMANDS_INTERPRETER", defVal)

	return interpreter, err
}

func inheritVariablesOrDefault(d *schema.ResourceData) []string {
	inherit := castConfigListString(d.Get("commands_environment_inherit_variables"))
	if len(inherit) < 1 {
		inherit, _, _ = getEnvList("ENVIRONMENT_INHERIT_VARIABLES", []string{})
	}
	return inherit
}

func providerConfigure(d *schema.ResourceData) (interface{}, error) {
	logging, err := providerConfigureLogging(d)
	debugLogging = envDefault("TF_SCRIPTED_DEBUG_LOGGING", "") != ""
	interpreter, err := interpreterOrDefault(castConfigListString(d.Get("commands_interpreter")))
	if err != nil {
		return nil, err
	}

	if len(interpreter) < 1 {
		return nil, fmt.Errorf(`invalid interpreter: %s`, interpreter)
	}

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
				PrefixNew:        d.Get("commands_environment_prefix_new").(string),
				PrefixOld:        d.Get("commands_environment_prefix_old").(string),
				IncludeParent:    d.Get("commands_environment_include_parent").(bool),
				InheritVariables: inheritVariablesOrDefault(d),
			},
			Templates: &CommandTemplates{
				Interpreter:   interpreter,
				Dependencies:  d.Get("commands_dependencies").(string),
				ModifyPrefix:  d.Get("commands_modify_prefix").(string),
				Prefix:        d.Get("commands_prefix").(string),
				PrefixFromEnv: d.Get("commands_prefix_fromenv").(string),
				Create:        d.Get("commands_create").(string),
				Delete:        d.Get("commands_delete").(string),
				Exists:        d.Get("commands_exists").(string),
				Id:            d.Get("commands_id").(string),
				NeedsUpdate:   d.Get("commands_needs_update").(string),
				Read:          d.Get("commands_read").(string),
				Update:        update,
			},
			Output: &OutputConfig{
				LogLevel:  hclog.LevelFromString(d.Get("logging_output_logging_log_level").(string)),
				LineWidth: d.Get("logging_output_line_width").(int),
				LogPids:   d.Get("logging_pids").(bool),
				LogIids:   d.Get("logging_iids").(bool),
			},
			CreateAfterUpdate:   cau,
			DeleteBeforeUpdate:  dbu,
			DeleteOnNotExists:   d.Get("commands_delete_on_not_exists").(bool),
			DeleteOnReadFailure: d.Get("commands_delete_on_read_failure").(bool),
			Separator:           d.Get("commands_separator").(string),
			WorkingDirectory:    d.Get("commands_working_directory").(string),
			TriggerString:       envDefault(TriggerStringEnvKey, DefaultTriggerString),
		},
		Templates: &TemplatesConfig{
			LeftDelim:  d.Get("templates_left_delim").(string),
			RightDelim: d.Get("templates_right_delim").(string),
		},
		EmptyString:            EmptyString,
		Logging:                logging,
		LoggingBufferSize:      int64(d.Get("logging_buffer_size").(int)),
		OutputFormat:           of,
		OutputLinePrefix:       d.Get("commands_read_line_prefix").(string),
		StateFormat:            sf,
		StateLinePrefix:        envDefault(StatePrefixEnvKey, DefaultStatePrefix),
		RunningMessageInterval: d.Get("logging_running_messages_interval").(float64),
	}
	logging.Log(hclog.Info, `Provider "scripted" initialized`)
	return &config, nil
}
