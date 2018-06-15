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
var ValidLogLevelsStrings = []string{"TRACE", "DEBUG", "INFO", "WARN", "ERROR"}

var DefaultEnvPrefix = "TF_SCRIPTED_"
var EnvPrefix = envDefault("TF_SCRIPTED_ENV_PREFIX", DefaultEnvPrefix)
var debugLogging = false

// String representing empty value, can be set to anything
var EmptyString, _ = getEnv("EMPTY_STRING", `ZVaXr3jCd80vqJRhBP9t83LrpWIdNKWJ`)

func defaultEmptyString() (interface{}, error) {
	return EmptyString, nil
}

func defaultMsg(msg, defVal string) string {
	defVal = strings.Replace(defVal, "$", "$"+EnvPrefix, -1)
	return fmt.Sprintf("%s Defaults to: %s", msg, defVal)
}

func envKey(key string) (ret string) {
	if strings.HasPrefix(key, EnvPrefix) {
		ret = key
	} else {
		ret = EnvPrefix + key
	}
	if debugLogging {
		Stderr.WriteString(fmt.Sprintf(`envKey("%s") -> ("%s")`+"\n", key, ret))
	}
	return ret
}

func envDefaultFunc(key, defVal string) schema.SchemaDefaultFunc {
	return func() (interface{}, error) {
		ret, _ := getEnv(key, defVal)
		return ret, nil
	}
}

func getEnv(key, defValue string) (value string, ok bool) {
	return envDefaultOk(envKey(key), defValue)
}

func getEnvBoolOk(key string, defVal bool) (value, ok bool) {
	str, ok := getEnv(key, EmptyString)
	if str == EmptyString {
		return defVal, false
	}
	value, err := strconv.ParseBool(str)
	if err != nil {
		ok = false
	}
	return value, ok
}

func getEnvBool(key string, defVal bool) (value bool) {
	value, _ = getEnvBoolOk(key, defVal)
	return value
}

func getEnvBoolFalse(key string) bool {
	return getEnvBool(key, false)
}

func getEnvBoolTrue(key string) bool {
	return getEnvBool(key, true)
}

func envDefault(key, defValue string) string {
	ret, _ := envDefaultOk(key, defValue)
	return ret
}

func envDefaultOk(key, defValue string) (value string, ok bool) {
	value, ok = os.LookupEnv(key)
	if !ok {
		value = defValue
	}
	if debugLogging {
		Stderr.WriteString(fmt.Sprintf(`envDefaultOk("%s", "%s") -> ("%s", %v)`+"\n", key, defValue, value, ok))
	}
	return value, ok
}

func stringDefaultSchemaEmpty(schema *schema.Schema, key, description string) *schema.Schema {
	return stringDefaultSchemaMsgVal(schema, key, description, "not set")
}

func stringDefaultSchemaEmptyMsgVal(s *schema.Schema, key, description, msgVal string) *schema.Schema {
	return stringDefaultSchemaBase(s, key, description, EmptyString, msgVal)
}
func stringDefaultSchema(s *schema.Schema, key, description, defVal string) *schema.Schema {
	return stringDefaultSchemaBase(s, key, description, defVal, fmt.Sprintf("`%s`", defVal))
}
func stringDefaultSchemaBase(s *schema.Schema, key, description, defVal, msgVal string) *schema.Schema {
	if s == nil {
		s = &schema.Schema{}
	}
	key = strings.ToUpper(key)
	s.Type = schema.TypeString
	s.Optional = true
	s.DefaultFunc = envDefaultFunc(key, defVal)
	s.Description = defaultMsg(description, fmt.Sprintf("`$%s` or %s", key, msgVal))
	return s
}

func stringDefaultSchemaMsgVal(s *schema.Schema, key, description, msgVal string) *schema.Schema {
	return stringDefaultSchemaBase(s, key, description, EmptyString, msgVal)
}

func boolDefaultSchema(s *schema.Schema, key, description string, defVal bool) *schema.Schema {
	key = strings.ToUpper(key)
	prefix := "="
	if defVal {
		prefix = "!"
	}
	s = stringDefaultSchemaMsgVal(s, key, description, fmt.Sprintf("`$%s` %s= `\"\"`", key, prefix))
	s.DefaultFunc = func() (interface{}, error) {
		value, ok := getEnvBoolOk(key, defVal)
		if !ok {
			value = defVal
		}
		return value, nil
	}
	s.Type = schema.TypeBool
	return s
}

func Provider() terraform.ResourceProvider {
	return &schema.Provider{
		Schema: map[string]*schema.Schema{
			"commands_create": {
				Type:        schema.TypeString,
				Optional:    true,
				DefaultFunc: defaultEmptyString,
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
			"commands_exists_trigger_output": stringDefaultSchema(
				nil,
				"commands_exists_trigger_output",
				"Exact output expected from `commands_exists` to trigger not-exists behaviour.",
				"false",
			),
			"commands_id": {
				Type:        schema.TypeString,
				Optional:    true,
				DefaultFunc: defaultEmptyString,
				Description: "Command building resource id",
			},
			"commands_interpreter": {
				Type:        schema.TypeList,
				Optional:    true,
				Elem:        &schema.Schema{Type: schema.TypeString},
				Description: "Interpreter and it's arguments, can be a template with `command` variable.",
			},
			"commands_modify_prefix": {
				Type:        schema.TypeString,
				Optional:    true,
				DefaultFunc: defaultEmptyString,
				Description: "Modification (create and update) commands prefix",
			},
			"commands_should_update": {
				Type:        schema.TypeString,
				Optional:    true,
				DefaultFunc: defaultEmptyString,
				Description: "Command indicating whether resource should be updated, non-zero exit code to force update",
			},
			"commands_should_update_trigger_output": stringDefaultSchema(
				nil,
				"commands_should_update_trigger_output",
				"Exact output expected from `commands_should_update` to trigger an update.",
				"true",
			),
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
			"commands_separator": stringDefaultSchemaBase(
				nil,
				"commands_separator",
				"Format for joining 2 commands together without isolating them.",
				"%s\n%s",
				"%s\\n%s",
			),
			"commands_read": {
				Type:        schema.TypeString,
				Optional:    true,
				DefaultFunc: defaultEmptyString,
				Description: "Read command",
			},
			"commands_read_format": stringDefaultSchema(
				&schema.Schema{
					ValidateFunc: validation.StringInSlice([]string{"raw", "base64"}, false),
				},
				"commands_read_format",
				"Templates output types: raw `/^(?<key>[^=]+)=(?<value>[^\\n]*)$/` or base64 `/^(?<key>[^=]+)=(?<value_base64>[^\\n]*)$/`.",
				"raw",
			),
			"commands_read_line_prefix": stringDefaultSchemaEmpty(
				nil,
				"commands_read_line_prefix",
				"Ignore lines in read command without this prefix.",
			),
			"commands_state_format": stringDefaultSchemaEmptyMsgVal(
				&schema.Schema{
					ValidateFunc: validation.StringInSlice([]string{"raw", "base64", EmptyString}, false),
				},
				"commands_state_format",
				"Ignore lines in read command without this prefix.",
				"`commands_read_format`",
			),
			"commands_update": {
				Type:        schema.TypeString,
				Optional:    true,
				DefaultFunc: defaultEmptyString,
				Description: "Update command. Runs destroy then create by default.",
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
				Description: "stdout and stderr buffer sizes",
			},
			"logging_jsonformat": boolDefaultSchema(
				nil,
				"logging_jsonformat",
				"should logs be json instead of plain text?",
				true,
			),
			"logging_log_level": stringDefaultSchema(
				&schema.Schema{
					ValidateFunc: validation.StringInSlice(ValidLogLevelsStrings, true),
				},
				"logging_log_level",
				fmt.Sprintf("Logging level: %s.", strings.Join(ValidLogLevelsStrings, ", ")),
				"INFO",
			),
			"logging_log_path": {
				Type:        schema.TypeString,
				DefaultFunc: envDefaultFunc("LOGGING_LOG_PATH", EmptyString),
				Optional:    true,
				Description: defaultMsg("Extra logs output path.", "`$LOGGING_LOG_PATH`"),
			},
			"logging_output_logging_log_level": {
				Type:         schema.TypeString,
				Optional:     true,
				DefaultFunc:  envDefaultFunc("LOGGING_OUTPUT_LOG_LEVEL", "INFO"),
				ValidateFunc: validation.StringInSlice(ValidLogLevelsStrings, true),
				Description:  fmt.Sprintf("Command stdout/stderr log level: %s. Defaults to: `$LOGGING_OUTPUT_LOG_LEVEL`", strings.Join(ValidLogLevelsStrings, ", ")),
			},
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
			"logging_output_pids": boolDefaultSchema(
				nil,
				"logging_output_pids",
				"Should output lines contain `ppid` and `pid`?",
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
	debugLogging = envDefault("TF_SCRIPTED_DEBUG_LOGGING", "") != ""
	debugInterpreter := getEnvBoolFalse("INTERPRETER_DEBUG")
	// For some reason setting this via DefaultFunc results in an error
	interpreterI := d.Get("commands_interpreter").([]interface{})
	if len(interpreterI) == 0 {
		if runtime.GOOS == "windows" {
			interpreterI = []interface{}{"cmd", "/C"}
		} else {
			interpreterI = []interface{}{"bash", "-Eeuo", "pipefail", "-c", "{{ .command }}"}
			if debugInterpreter {
				interpreterI = append(interpreterI, "-x")
			}
		}
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
				Interpreter:   interpreter,
				ModifyPrefix:  d.Get("commands_modify_prefix").(string),
				Prefix:        d.Get("commands_prefix").(string),
				PrefixFromEnv: d.Get("commands_prefix_fromenv").(string),
				Create:        d.Get("commands_create").(string),
				Delete:        d.Get("commands_delete").(string),
				Exists:        d.Get("commands_exists").(string),
				Id:            d.Get("commands_id").(string),
				ShouldUpdate:  d.Get("commands_should_update").(string),
				Read:          d.Get("commands_read").(string),
				Update:        update,
			},
			Output: &OutputConfig{
				LogLevel:  hclog.LevelFromString(d.Get("logging_output_logging_log_level").(string)),
				LineWidth: d.Get("logging_output_line_width").(int),
				LogPids:   d.Get("logging_output_pids").(bool),
			},
			CreateAfterUpdate:          cau,
			DeleteBeforeUpdate:         dbu,
			DeleteOnNotExists:          d.Get("commands_delete_on_not_exists").(bool),
			DeleteOnReadFailure:        d.Get("commands_delete_on_read_failure").(bool),
			ShouldUpdateExpectedOutput: d.Get("commands_should_update_trigger_output").(string),
			ExistsExpectedOutput:       d.Get("commands_exists_trigger_output").(string),
			Separator:                  d.Get("commands_separator").(string),
			WorkingDirectory:           d.Get("commands_working_directory").(string),
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
