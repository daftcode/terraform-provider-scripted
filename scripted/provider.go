package scripted

import (
	"fmt"
	"github.com/hashicorp/go-hclog"
	"github.com/hashicorp/terraform/helper/schema"
	"github.com/hashicorp/terraform/helper/validation"
	"github.com/hashicorp/terraform/terraform"
	"io/ioutil"
	"os"
	"os/exec"
	"runtime"
	"strconv"
	"strings"
	"syscall"
)

var nextProviderId = 1

// Store original os.Stderr and os.Stdout, because it gets overwritten by go-plugin/server:Serve()
var parentStderr *os.File
var Stderr = os.Stderr

//noinspection GoUnusedGlobalVariable
var Stdout = os.Stdout

var EnvPrefix = envDefault("TF_SCRIPTED_ENV_PREFIX", DefaultEnvPrefix)
var Debug = getEnvBool("DEBUG", false)
var debugLogging = false

// String representing empty value, can be set to anything
var EnvEmptyString = getEnvMust("EMPTY_STRING", DefaultEmptyString)

func ParentStderr() *os.File {
	if parentStderr == nil {
		parentStderr = os.NewFile(uintptr(syscall.Stderr+1), fmt.Sprintf("/proc/%d/fd/%d", os.Getppid(), syscall.Stderr))
	}
	return parentStderr
}

func Provider() terraform.ResourceProvider {
	return &schema.Provider{
		Schema: map[string]*schema.Schema{
			CommandCreate: {
				Type:        schema.TypeString,
				Optional:    true,
				DefaultFunc: defaultEmptyString,
				Description: "Create command. Defaults to: `update_command`",
			},
			CommandCustomizeDiffComputeKeys: {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Command printing keys to be forced to recompute. Lines must be prefixed with LinePrefix and keys separated by whitespace characters",
				Removed:     fmt.Sprintf("%s was removed since it was redundant", CommandCustomizeDiffComputeKeys),
			},
			CommandDelete: {
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
			CommandDependencies: {
				Type:        schema.TypeString,
				Optional:    true,
				DefaultFunc: defaultEmptyString,
				Description: fmt.Sprintf("Command determining whether dependencies are met, dependencies met triggered by `%s`", TriggerStringTpl),
			},
			"commands_dependencies_error": {
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     false,
				Description: "Should commands fail on dependencies not met?",
			},
			"commands_environment_include_json_context": {
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     false,
				Description: fmt.Sprintf("Should whole TemplateContext be passed as JSON serialized %s environment variable to commands?", JsonContextEnvKey),
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
			CommandExists: {
				Type:        schema.TypeString,
				Optional:    true,
				DefaultFunc: defaultEmptyString,
				Description: fmt.Sprintf("Exists command, not-exists triggered by `%s`", TriggerStringTpl),
			},
			CommandId: {
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
					dWI, _ := toJson(DefaultWindowsInterpreter)
					dI, _ := toJson(DefaultInterpreter)
					return fmt.Sprintf(
						"Interpreter and it's arguments, can be a template with `command` variable. "+
							"Defaults to: `$TF_SCRIPTED_COMMANDS_INTERPRETER` (JSON array), `%s` (windows) or `%s`",
						dWI,
						dI,
					)
				}(),
			},
			"commands_interpreter_is_provider": {
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     false,
				Description: "Should interpreter be considered provider implementation? Should execude commands based based on TF_SCRIPTED_CONTEXT envvar (context's .Command) and ignore command line arguments.",
			},
			"commands_interpreter_provider_commands": {
				Type:        schema.TypeList,
				Optional:    true,
				Elem:        &schema.Schema{Type: schema.TypeString},
				Description: "Commands supported by interpreter-provider. Defaults to: result of running interpreter with `commands` argument",
			},
			"commands_modify_prefix": {
				Type:        schema.TypeString,
				Optional:    true,
				DefaultFunc: defaultEmptyString,
				Description: "Modification commands (create and update) prefix",
			},
			CommandNeedsUpdate: {
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
			CommandRead: {
				Type:        schema.TypeString,
				Optional:    true,
				DefaultFunc: defaultEmptyString,
				Description: "Read command",
			},
			"output_format": stringDefaultSchema(
				&schema.Schema{
					ValidateFunc: validation.StringInSlice([]string{"raw", "base64", "json"}, false),
				},
				"output_format",
				"Templates output types: "+
					"raw `/^(?<key>[^=]+)=(?<value>[^\\n]*)$/`, "+
					"base64 `/^(?<key>[^=]+)=(?<value_base64>[^\\n]*)$/` or "+
					"one JSON object per line overriding previously existing keys.",
				"raw",
			),
			"commands_read_use_default_line_prefix": boolDefaultSchema(
				nil,
				"commands_read_use_default_line_prefix",
				"Ignore lines in read command without default line prefix instead of read-specific",
				false,
			),
			"output_line_prefix": stringDefaultSchemaEmpty(
				nil,
				"output_line_prefix",
				"Ignore lines in read command without this prefix.",
			),
			"state_format": stringDefaultSchemaEmptyMsgVal(
				&schema.Schema{
					ValidateFunc: validation.StringInSlice([]string{"raw", "base64", "json", EnvEmptyString}, false),
				},
				"state_format",
				"Create/Update state output format, for more info see `output_format`.",
				"`output_format`",
			),
			CommandUpdate: {
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
			"state_compute_keys": {
				Type:        schema.TypeList,
				Optional:    true,
				Elem:        &schema.Schema{Type: schema.TypeString},
				Description: "List of `state` keys which are forced to be computed on change.",
			},
			"output_compute_keys": {
				Type:        schema.TypeList,
				Optional:    true,
				Elem:        &schema.Schema{Type: schema.TypeString},
				Description: "List of `output` keys which are forced to be computed on change.",
			},
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
			"logging_output_parent_stderr": boolDefaultSchema(
				nil,
				"logging_output_parent_stderr",
				"should we log directly to parent's stderr instead of our own?",
				false,
			),
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
			"trigger_string": stringDefaultSchema(
				nil,
				"trigger_string",
				"TriggerString for exists, dependencies_met and needs_update",
				DefaultTriggerString,
			),
			"line_prefix": stringDefaultSchema(
				nil,
				"line_prefix",
				"General line prefix",
				DefaultLinePrefix,
			),
			"state_line_prefix": stringDefaultSchema(
				nil,
				"state_line_prefix",
				"State line prefix",
				DefaultStatePrefix,
			),
			"open_parent_stderr": boolDefaultSchema(
				nil,
				"open_parent_stderr",
				"should we open 3rd file descriptor as parent's Stderr?",
				false,
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
	jsonFormat := d.Get("logging_jsonformat").(bool)
	logToParent := d.Get("logging_output_parent_stderr").(bool)
	logProviderName := d.Get("logging_provider_name").(string)
	logLevel := hclog.LevelFromString(d.Get("logging_log_level").(string))
	jsonList := d.Get("logging_jsonlist").(bool)
	jsonListPromote := d.Get("logging_jsonlistpromote").(bool)
	output := Stderr

	if logToParent {
		if err := d.Set("open_parent_stderr", true); err != nil {
			return nil, err
		}
		output = ParentStderr()
	}

	logger := hclog.New(&hclog.LoggerOptions{
		JSONFormat:      jsonFormat,
		JSONList:        jsonList,
		JSONListPromote: jsonListPromote,
		Output:          output,
		Level:           logLevel,
	})
	hcloggers = append(hcloggers, logger)

	logPath := d.Get("logging_log_path").(string)
	var fileLogger hclog.Logger
	if logPath != EnvEmptyString {
		logFile, err := os.OpenFile(logPath, os.O_WRONLY|os.O_APPEND|os.O_CREATE, 0644)
		if err != nil {
			return nil, err
		}
		fileLogger = hclog.New(&hclog.LoggerOptions{
			JSONFormat:      jsonFormat,
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
	if logProviderName != EnvEmptyString {
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
		defVal = DefaultWindowsInterpreter
	} else {
		defVal = DefaultInterpreter
		if debugInterpreter {
			interpreter = append(defVal, "-x")
		}
	}
	interpreter, _, err = getEnvList("COMMANDS_INTERPRETER", defVal)

	return interpreter, err
}

func providerConfigure(d *schema.ResourceData) (interface{}, error) {
	logging, err := providerConfigureLogging(d)
	if err != nil {
		return nil, err
	}
	debugLogging = envDefault("TF_SCRIPTED_DEBUG_LOGGING", "") != ""

	interpreter, err := interpreterOrDefault(castConfigListString(d.Get("commands_interpreter")))
	if err != nil {
		return nil, err
	}
	if err := d.Set("commands_interpreter", interpreter); err != nil {
		return nil, err
	}

	if len(interpreter) < 1 {
		return nil, fmt.Errorf(`invalid interpreter: %s`, interpreter)
	}

	if err != nil {
		return nil, err
	}

	// Set default state_format
	if d.Get("state_format").(string) == EnvEmptyString {
		if err := d.Set("state_format", d.Get("output_format").(string)); err != nil {
			return nil, err
		}
	}

	// Set read prefix
	if d.Get("commands_read_use_default_line_prefix").(bool) {
		if err := d.Set("output_line_prefix", d.Get("line_prefix").(string)); err != nil {
			return nil, err
		}
	}

	// Set default commands_environment_inherit_variables
	inherit := castConfigListString(d.Get("commands_environment_inherit_variables"))
	if len(inherit) == 0 {
		inherit, _, err = getEnvList("ENVIRONMENT_INHERIT_VARIABLES", []string{})
		if err != nil {
			return nil, err
		}
		if err := d.Set("commands_environment_inherit_variables", inherit); err != nil {
			return nil, err
		}
	}

	interpreterProviderCommands := castConfigListString(d.Get("commands_interpreter_provider_commands"))
	if d.Get("commands_interpreter_is_provider").(bool) {
		if len(interpreterProviderCommands) == 0 {
			name := interpreter[0]
			args := interpreter[1:]
			args = append(args, "commands")
			cmd := exec.Command(name, args...)
			stdout, err := cmd.StdoutPipe()
			if err != nil {
				return nil, err
			}
			if err = cmd.Start(); err != nil {
				return nil, err
			}
			stdoutBytes, _ := ioutil.ReadAll(stdout)

			if err := cmd.Wait(); err != nil {
				return nil, err
			}
			interpreterProviderCommands = strings.Fields(string(stdoutBytes[:]))

			for _, command := range interpreterProviderCommands {
				if _, ok := AllowedCommands[command]; !ok {
					var allowedKeys []string
					for key := range AllowedCommands {
						allowedKeys = append(allowedKeys, fmt.Sprintf("%v", key))
					}
					return nil, fmt.Errorf(
						"command %v is not allowed, only: %s",
						command,
						strings.Join(allowedKeys, ", "),
					)
				}
				if err := d.Set(command, command); err != nil {
					return nil, err
				}
			}
		}
		if err := d.Set("commands_prefix", EnvEmptyString); err != nil {
			return nil, err
		}
		if err := d.Set("commands_modify_prefix", EnvEmptyString); err != nil {
			return nil, err
		}
		if err := d.Set("commands_prefix_fromenv", EnvEmptyString); err != nil {
			return nil, err
		}
		if err := d.Set("commands_environment_include_parent", true); err != nil {
			return nil, err
		}
		if err := d.Set("commands_environment_include_json_context", true); err != nil {
			return nil, err
		}
		if err := d.Set("commands_interpreter_provider_commands", interpreterProviderCommands); err != nil {
			return nil, err
		}
	}

	outputLinePrefix := d.Get("output_line_prefix").(string)
	if !isSet(outputLinePrefix) {
		outputLinePrefix = ""
		if err := d.Set("output_line_prefix", outputLinePrefix); err != nil {
			return nil, err
		}
	}
	config := ProviderConfig{
		Commands: &CommandsConfig{
			Environment: &EnvironmentConfig{
				PrefixNew:          d.Get("commands_environment_prefix_new").(string),
				PrefixOld:          d.Get("commands_environment_prefix_old").(string),
				IncludeParent:      d.Get("commands_environment_include_parent").(bool),
				InheritVariables:   castConfigListString(d.Get("commands_environment_inherit_variables")),
				IncludeJsonContext: d.Get("commands_environment_include_json_context").(bool),
			},
			Templates: &CommandTemplates{
				Interpreter:   interpreter,
				Dependencies:  d.Get(CommandDependencies).(string),
				ModifyPrefix:  d.Get("commands_modify_prefix").(string),
				Prefix:        d.Get("commands_prefix").(string),
				PrefixFromEnv: d.Get("commands_prefix_fromenv").(string),
				Create:        d.Get(CommandCreate).(string),
				Delete:        d.Get(CommandDelete).(string),
				Exists:        d.Get(CommandExists).(string),
				Id:            d.Get(CommandId).(string),
				NeedsUpdate:   d.Get(CommandNeedsUpdate).(string),
				Read:          d.Get(CommandRead).(string),
				Update:        d.Get(CommandUpdate).(string),
			},
			Output: &OutputConfig{
				LogLevel:  hclog.LevelFromString(d.Get("logging_output_logging_log_level").(string)),
				LineWidth: d.Get("logging_output_line_width").(int),
				LogPids:   d.Get("logging_pids").(bool),
				LogIids:   d.Get("logging_iids").(bool),
			},
			InterpreterIsProvider:       d.Get("commands_interpreter_is_provider").(bool),
			InterpreterProviderCommands: interpreterProviderCommands,
			DependenciesNotMetError:     d.Get("commands_dependencies_error").(bool),
			DeleteOnNotExists:           d.Get("commands_delete_on_not_exists").(bool),
			DeleteOnReadFailure:         d.Get("commands_delete_on_read_failure").(bool),
			Separator:                   d.Get("commands_separator").(string),
			WorkingDirectory:            d.Get("commands_working_directory").(string),
			TriggerString:               d.Get("trigger_string").(string),
		},
		Templates: &TemplatesConfig{
			LeftDelim:  d.Get("templates_left_delim").(string),
			RightDelim: d.Get("templates_right_delim").(string),
		},
		logging: logging,

		OpenParentStderr:       d.Get("open_parent_stderr").(bool),
		LoggingBufferSize:      int64(d.Get("logging_buffer_size").(int)),
		StateComputeKeys:       castConfigListString(d.Get("state_compute_keys")),
		OutputComputeKeys:      castConfigListString(d.Get("output_compute_keys")),
		OutputFormat:           d.Get("output_format").(string),
		OutputLinePrefix:       outputLinePrefix,
		EmptyString:            EnvEmptyString,
		StateFormat:            d.Get("state_format").(string),
		LinePrefix:             d.Get("line_prefix").(string),
		StateLinePrefix:        d.Get("state_line_prefix").(string),
		RunningMessageInterval: d.Get("logging_running_messages_interval").(float64),
		Version:                Version,
		EnvPrefix:              EnvPrefix,
		InstanceState:          d.State(),
	}

	if config.OpenParentStderr {
		ParentStderr()
	}

	logging.Log(hclog.Info, `Provider "scripted" initialized`)
	return &config, nil
}
