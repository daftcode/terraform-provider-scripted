
# provider scripted

# Arguments reference

| Argument | Type | Description | Default |
|:---      | ---  | ---         | ---     |
| `commands_create` | [string](https://www.terraform.io/docs/extend/schemas/schema-types.html#typestring) | Create command, defaults to `update_command` | not set |
| `commands_delete` | [string](https://www.terraform.io/docs/extend/schemas/schema-types.html#typestring) | Delete command | not set |
| `commands_delete_on_not_exists` | [bool](https://www.terraform.io/docs/extend/schemas/schema-types.html#typebool) | Delete resource when exists fails | `true` |
| `commands_delete_on_read_failure` | [bool](https://www.terraform.io/docs/extend/schemas/schema-types.html#typebool) | Delete resource when read fails | `false` |
| `commands_dependencies` | [string](https://www.terraform.io/docs/extend/schemas/schema-types.html#typestring) | Command determining whether dependencies are met | not set |
| `commands_dependencies_trigger_output` | [string](https://www.terraform.io/docs/extend/schemas/schema-types.html#typestring) | Exact output expected from `commands_dependencies` to pass the check.  | `$TF_SCRIPTED_COMMANDS_EXISTS_TRIGGER_OUTPUT` or `true` |
| `commands_environment_include_parent` | [bool](https://www.terraform.io/docs/extend/schemas/schema-types.html#typebool) | Include parent environment in the command? | `true` |
| `commands_environment_prefix_new` | [string](https://www.terraform.io/docs/extend/schemas/schema-types.html#typestring) | New environment prefix (skip if empty) | not set |
| `commands_environment_prefix_old` | [string](https://www.terraform.io/docs/extend/schemas/schema-types.html#typestring) | Old environment prefix (skip if empty) | not set |
| `commands_exists` | [string](https://www.terraform.io/docs/extend/schemas/schema-types.html#typestring) | Exists command | not set |
| `commands_exists_trigger_output` | [string](https://www.terraform.io/docs/extend/schemas/schema-types.html#typestring) | Exact output expected from `commands_exists` to trigger not-exists behaviour.  | `$TF_SCRIPTED_COMMANDS_EXISTS_TRIGGER_OUTPUT` or `false` |
| `commands_id` | [string](https://www.terraform.io/docs/extend/schemas/schema-types.html#typestring) | Command building resource id | not set |
| `commands_interpreter` | [list](https://www.terraform.io/docs/extend/schemas/schema-types.html#typelist) | Interpreter and it's arguments, can be a template with `command` variable. | not set |
| `commands_modify_prefix` | [string](https://www.terraform.io/docs/extend/schemas/schema-types.html#typestring) | Modification (create and update) commands prefix | not set |
| `commands_needs_delete` | [string](https://www.terraform.io/docs/extend/schemas/schema-types.html#typestring) | Command indicating whether resource should be deleted, non-zero exit code to force update | not set |
| `commands_needs_delete_trigger_output` | [string](https://www.terraform.io/docs/extend/schemas/schema-types.html#typestring) | Exact output expected from `commands_needs_delete` to trigger delete.  | `$TF_SCRIPTED_COMMANDS_NEEDS_DELETE_TRIGGER_OUTPUT` or `true` |
| `commands_needs_update` | [string](https://www.terraform.io/docs/extend/schemas/schema-types.html#typestring) | Command indicating whether resource should be updated, non-zero exit code to force update | not set |
| `commands_needs_update_trigger_output` | [string](https://www.terraform.io/docs/extend/schemas/schema-types.html#typestring) | Exact output expected from `commands_needs_update` to trigger an update.  | `$TF_SCRIPTED_COMMANDS_NEEDS_UPDATE_TRIGGER_OUTPUT` or `true` |
| `commands_prefix` | [string](https://www.terraform.io/docs/extend/schemas/schema-types.html#typestring) | Command prefix shared between all commands | not set |
| `commands_prefix_fromenv` | [string](https://www.terraform.io/docs/extend/schemas/schema-types.html#typestring) | Command prefix shared between all commands (added before `commands_prefix`)  | `$TF_SCRIPTED_COMMANDS_PREFIX_FROMENV` or not set |
| `commands_read` | [string](https://www.terraform.io/docs/extend/schemas/schema-types.html#typestring) | Read command | not set |
| `commands_read_format` | [string](https://www.terraform.io/docs/extend/schemas/schema-types.html#typestring) | Templates output types: raw `/^(?<key>[^=]+)=(?<value>[^\n]*)$/` or base64 `/^(?<key>[^=]+)=(?<value_base64>[^\n]*)$/`.  | `$TF_SCRIPTED_COMMANDS_READ_FORMAT` or `raw` |
| `commands_read_line_prefix` | [string](https://www.terraform.io/docs/extend/schemas/schema-types.html#typestring) | Ignore lines in read command without this prefix.  | `$TF_SCRIPTED_COMMANDS_READ_LINE_PREFIX` or not set |
| `commands_separator` | [string](https://www.terraform.io/docs/extend/schemas/schema-types.html#typestring) | Format for joining 2 commands together without isolating them.  | `$TF_SCRIPTED_COMMANDS_SEPARATOR` or %s\n%s |
| `commands_state_format` | [string](https://www.terraform.io/docs/extend/schemas/schema-types.html#typestring) | Ignore lines in read command without this prefix.  | `$TF_SCRIPTED_COMMANDS_STATE_FORMAT` or `commands_read_format` |
| `commands_update` | [string](https://www.terraform.io/docs/extend/schemas/schema-types.html#typestring) | Update command. Runs destroy then create by default. | not set |
| `commands_working_directory` | [string](https://www.terraform.io/docs/extend/schemas/schema-types.html#typestring) | Working directory to run commands in  | `$TF_SCRIPTED_COMMANDS_WORKING_DIRECTORY` or not set |
| `dependencies` | [map](https://www.terraform.io/docs/extend/schemas/schema-types.html#typemap) | Dependencies purely for provider graph walking, otherwise ignored. | not set |
| `logging_buffer_size` | [int](https://www.terraform.io/docs/extend/schemas/schema-types.html#typeint) | stdout and stderr buffer sizes | `8192` |
| `logging_iids` | [bool](https://www.terraform.io/docs/extend/schemas/schema-types.html#typebool) | Should output lines contain `piid` (provider instance id) and `riid` (resource instance id?  | `$TF_SCRIPTED_LOGGING_IIDS` or `$TF_SCRIPTED_LOGGING_IIDS` == `""` |
| `logging_jsonformat` | [bool](https://www.terraform.io/docs/extend/schemas/schema-types.html#typebool) | should logs be json instead of plain text?  | `$TF_SCRIPTED_LOGGING_JSONFORMAT` or `$TF_SCRIPTED_LOGGING_JSONFORMAT` != `""` |
| `logging_log_level` | [string](https://www.terraform.io/docs/extend/schemas/schema-types.html#typestring) | Logging level: TRACE, DEBUG, INFO, WARN, ERROR.  | `$TF_SCRIPTED_LOGGING_LOG_LEVEL` or `INFO` |
| `logging_log_path` | [string](https://www.terraform.io/docs/extend/schemas/schema-types.html#typestring) | Extra logs output path.  | `$TF_SCRIPTED_LOGGING_LOG_PATH` |
| `logging_output_line_width` | [int](https://www.terraform.io/docs/extend/schemas/schema-types.html#typeint) | Width of command's line to use during formatting.  | `$TF_SCRIPTED_LOGGING_OUTPUT_LINE_WIDTH` |
| `logging_output_logging_log_level` | [string](https://www.terraform.io/docs/extend/schemas/schema-types.html#typestring) | Command stdout/stderr log level: TRACE, DEBUG, INFO, WARN, ERROR.  | `$LOGGING_OUTPUT_LOG_LEVEL` |
| `logging_pids` | [bool](https://www.terraform.io/docs/extend/schemas/schema-types.html#typebool) | Should output lines contain `ppid` and `pid`?  | `$TF_SCRIPTED_LOGGING_PIDS` or `$TF_SCRIPTED_LOGGING_PIDS` == `""` |
| `logging_provider_name` | [string](https://www.terraform.io/docs/extend/schemas/schema-types.html#typestring) | Name to display in log entries for this provider | not set |
| `templates_left_delim` | [string](https://www.terraform.io/docs/extend/schemas/schema-types.html#typestring) | Left delimiter for templates.  | `$TF_SCRIPTED_TEMPLATES_LEFT_DELIM` or `{{` |
| `templates_right_delim` | [string](https://www.terraform.io/docs/extend/schemas/schema-types.html#typestring) | Right delimiter for templates.  | `$TF_SCRIPTED_TEMPLATES_RIGHT_DELIM` or `}}` |
