
# provider scripted

# Arguments reference

| Argument | Type | Description | Default |
|:---      | ---  | ---         | ---     |
|  `commands_create` | [string](https://www.terraform.io/docs/extend/schemas/schema-types.html#typestring) | Create command.  | `update_command` |
| REMOVED `commands_customizediff_computekeys` | [string](https://www.terraform.io/docs/extend/schemas/schema-types.html#typestring) | Command printing keys to be forced to recompute. Lines must be prefixed with LinePrefix and keys separated by whitespace characters | not set |
|  `commands_delete` | [string](https://www.terraform.io/docs/extend/schemas/schema-types.html#typestring) | Delete command | not set |
|  `commands_delete_on_not_exists` | [bool](https://www.terraform.io/docs/extend/schemas/schema-types.html#typebool) | Delete resource when exists fails | `true` |
|  `commands_delete_on_read_failure` | [bool](https://www.terraform.io/docs/extend/schemas/schema-types.html#typebool) | Delete resource when read fails | `false` |
|  `commands_dependencies` | [string](https://www.terraform.io/docs/extend/schemas/schema-types.html#typestring) | Command determining whether dependencies are met, dependencies met triggered by `{{ .TriggerString }}` | not set |
|  `commands_dependencies_error` | [bool](https://www.terraform.io/docs/extend/schemas/schema-types.html#typebool) | Should commands fail on dependencies not met? | `false` |
|  `commands_environment_include_json_context` | [bool](https://www.terraform.io/docs/extend/schemas/schema-types.html#typebool) | Should whole TemplateContext be passed as JSON serialized TF_SCRIPTED_CONTEXT environment variable to commands? | `false` |
|  `commands_environment_include_parent` | [bool](https://www.terraform.io/docs/extend/schemas/schema-types.html#typebool) | Include whole parent environment in the command? | `false` |
|  `commands_environment_inherit_variables` | [list](https://www.terraform.io/docs/extend/schemas/schema-types.html#typelist) | List of environment variables to inherit from parent.  | `$TF_SCRIPTED_ENVIRONMENT_INHERIT_VARIABLES` (JSON array) |
|  `commands_environment_prefix_new` | [string](https://www.terraform.io/docs/extend/schemas/schema-types.html#typestring) | New environment prefix (skip if empty) | not set |
|  `commands_environment_prefix_old` | [string](https://www.terraform.io/docs/extend/schemas/schema-types.html#typestring) | Old environment prefix (skip if empty) | not set |
|  `commands_exists` | [string](https://www.terraform.io/docs/extend/schemas/schema-types.html#typestring) | Exists command, not-exists triggered by `{{ .TriggerString }}` | not set |
|  `commands_id` | [string](https://www.terraform.io/docs/extend/schemas/schema-types.html#typestring) | Command building resource id | not set |
|  `commands_interpreter` | [list](https://www.terraform.io/docs/extend/schemas/schema-types.html#typelist) | Interpreter and it's arguments, can be a template with `command` variable.  | `$TF_SCRIPTED_COMMANDS_INTERPRETER` (JSON array), `["cmd","/C","{{ .command }}"]` (windows) or `["bash","-Eeuo","pipefail","-c","{{ .command }}"]` |
|  `commands_interpreter_is_provider` | [bool](https://www.terraform.io/docs/extend/schemas/schema-types.html#typebool) | Should interpreter be considered provider implementation? Should execude commands based based on TF_SCRIPTED_CONTEXT envvar (context's .Command) and ignore command line arguments. | `false` |
|  `commands_interpreter_provider_commands` | [list](https://www.terraform.io/docs/extend/schemas/schema-types.html#typelist) | Commands supported by interpreter-provider.  | result of running interpreter with `commands` argument |
|  `commands_modify_prefix` | [string](https://www.terraform.io/docs/extend/schemas/schema-types.html#typestring) | Modification commands (create and update) prefix | not set |
|  `commands_needs_update` | [string](https://www.terraform.io/docs/extend/schemas/schema-types.html#typestring) | Command indicating whether resource should be updated, update triggered by `{{ .TriggerString }}` | not set |
|  `commands_prefix` | [string](https://www.terraform.io/docs/extend/schemas/schema-types.html#typestring) | Command prefix shared between all commands | not set |
|  `commands_prefix_fromenv` | [string](https://www.terraform.io/docs/extend/schemas/schema-types.html#typestring) | Command prefix shared between all commands (added before `commands_prefix`)  | `$TF_SCRIPTED_COMMANDS_PREFIX_FROMENV` or not set |
|  `commands_read` | [string](https://www.terraform.io/docs/extend/schemas/schema-types.html#typestring) | Read command | not set |
|  `commands_read_use_default_line_prefix` | [bool](https://www.terraform.io/docs/extend/schemas/schema-types.html#typebool) | Ignore lines in read command without default line prefix instead of read-specific  | `$TF_SCRIPTED_COMMANDS_READ_USE_DEFAULT_LINE_PREFIX` == `""` |
|  `commands_separator` | [string](https://www.terraform.io/docs/extend/schemas/schema-types.html#typestring) | Format for joining 2 commands together without isolating them.  | `$TF_SCRIPTED_COMMANDS_SEPARATOR` or `%s\n%s` |
|  `commands_update` | [string](https://www.terraform.io/docs/extend/schemas/schema-types.html#typestring) | Update command. Deletes then creates if not set. Can be used in place of `create_command`. | not set |
|  `commands_working_directory` | [string](https://www.terraform.io/docs/extend/schemas/schema-types.html#typestring) | Working directory to run commands in  | `$TF_SCRIPTED_COMMANDS_WORKING_DIRECTORY` or not set |
|  `dependencies` | [map](https://www.terraform.io/docs/extend/schemas/schema-types.html#typemap) | Dependencies purely for provider graph walking, otherwise ignored. | not set |
|  `line_prefix` | [string](https://www.terraform.io/docs/extend/schemas/schema-types.html#typestring) | General line prefix  | `$TF_SCRIPTED_LINE_PREFIX` or `QmGRizGk1fdPEBVVZSGkCRPJRgAe9p07B` |
|  `logging_buffer_size` | [int](https://www.terraform.io/docs/extend/schemas/schema-types.html#typeint) | output (on error) buffer sizes | `8192` |
|  `logging_iids` | [bool](https://www.terraform.io/docs/extend/schemas/schema-types.html#typebool) | Should output lines contain `piid` (provider instance id) and `riid` (resource instance id?  | `$TF_SCRIPTED_LOGGING_IIDS` == `""` |
|  `logging_jsonformat` | [bool](https://www.terraform.io/docs/extend/schemas/schema-types.html#typebool) | should logs be json instead of plain text?  | `$TF_SCRIPTED_LOGGING_JSONFORMAT` != `""` |
|  `logging_jsonlist` | [bool](https://www.terraform.io/docs/extend/schemas/schema-types.html#typebool) | should json log formatter output lists instead of direct values?  | `$TF_SCRIPTED_LOGGING_JSONLIST` == `""` |
|  `logging_jsonlistpromote` | [bool](https://www.terraform.io/docs/extend/schemas/schema-types.html#typebool) | should json log formatter promote single values to lists and append?  | `$TF_SCRIPTED_LOGGING_JSONLISTPROMOTE` != `""` |
|  `logging_log_level` | [string](https://www.terraform.io/docs/extend/schemas/schema-types.html#typestring) | Logging level: TRACE, DEBUG, INFO, WARN, ERROR.  | `$TF_SCRIPTED_LOG_LEVEL` or `INFO` |
|  `logging_log_path` | [string](https://www.terraform.io/docs/extend/schemas/schema-types.html#typestring) | Additional logs output path.  | `$TF_SCRIPTED_LOG_PATH` or not set |
|  `logging_output_line_width` | [int](https://www.terraform.io/docs/extend/schemas/schema-types.html#typeint) | Width of command's line to use during formatting.  | `$TF_SCRIPTED_LOGGING_OUTPUT_LINE_WIDTH` |
|  `logging_output_logging_log_level` | [string](https://www.terraform.io/docs/extend/schemas/schema-types.html#typestring) | Command stdout/stderr log level: TRACE, DEBUG, INFO, WARN, ERROR.  | `$TF_SCRIPTED_OUTPUT_LOG_LEVEL` or `INFO` |
|  `logging_output_parent_stderr` | [bool](https://www.terraform.io/docs/extend/schemas/schema-types.html#typebool) | should we log directly to parent's stderr instead of our own?  | `$TF_SCRIPTED_LOGGING_OUTPUT_PARENT_STDERR` == `""` |
|  `logging_pids` | [bool](https://www.terraform.io/docs/extend/schemas/schema-types.html#typebool) | Should output lines contain `ppid` and `pid`?  | `$TF_SCRIPTED_LOGGING_PIDS` == `""` |
|  `logging_provider_name` | [string](https://www.terraform.io/docs/extend/schemas/schema-types.html#typestring) | Name to display in log entries for this provider | not set |
|  `logging_running_messages_interval` | [float](https://www.terraform.io/docs/extend/schemas/schema-types.html#typefloat) | should resources report still being in a running state? Trigger reports every N seconds.  | `$TF_SCRIPTED_LOGGING_RUNNING_MESSAGES_INTERVAL` |
|  `open_parent_stderr` | [bool](https://www.terraform.io/docs/extend/schemas/schema-types.html#typebool) | should we open 3rd file descriptor as parent's Stderr?  | `$TF_SCRIPTED_OPEN_PARENT_STDERR` == `""` |
|  `output_compute_keys` | [list](https://www.terraform.io/docs/extend/schemas/schema-types.html#typelist) | List of `output` keys which are forced to be computed on change. | not set |
|  `output_format` | [string](https://www.terraform.io/docs/extend/schemas/schema-types.html#typestring) | Templates output types: raw `/^(?<key>[^=]+)=(?<value>[^\n]*)$/`, base64 `/^(?<key>[^=]+)=(?<value_base64>[^\n]*)$/` or one JSON object per line overriding previously existing keys.  | `$TF_SCRIPTED_OUTPUT_FORMAT` or `raw` |
|  `output_line_prefix` | [string](https://www.terraform.io/docs/extend/schemas/schema-types.html#typestring) | Ignore lines in read command without this prefix.  | `$TF_SCRIPTED_OUTPUT_LINE_PREFIX` or not set |
|  `state_compute_keys` | [list](https://www.terraform.io/docs/extend/schemas/schema-types.html#typelist) | List of `state` keys which are forced to be computed on change. | not set |
|  `state_format` | [string](https://www.terraform.io/docs/extend/schemas/schema-types.html#typestring) | Create/Update state output format, for more info see `output_format`.  | `$TF_SCRIPTED_STATE_FORMAT` or `output_format` |
|  `state_line_prefix` | [string](https://www.terraform.io/docs/extend/schemas/schema-types.html#typestring) | State line prefix  | `$TF_SCRIPTED_STATE_LINE_PREFIX` or `WViRV1TbGAGehAYFL8g3ZL8o1cg1bxaq` |
|  `templates_left_delim` | [string](https://www.terraform.io/docs/extend/schemas/schema-types.html#typestring) | Left delimiter for templates.  | `$TF_SCRIPTED_TEMPLATES_LEFT_DELIM` or `{{` |
|  `templates_right_delim` | [string](https://www.terraform.io/docs/extend/schemas/schema-types.html#typestring) | Right delimiter for templates.  | `$TF_SCRIPTED_TEMPLATES_RIGHT_DELIM` or `}}` |
|  `trigger_string` | [string](https://www.terraform.io/docs/extend/schemas/schema-types.html#typestring) | TriggerString for exists, dependencies_met and needs_update  | `$TF_SCRIPTED_TRIGGER_STRING` or `ndn4VFxYG489bUmV6xKjKFE0RYQIJdts` |
