## Argument reference

| Argument | Type | Description | Default |
|:---      | ---  | ---         | ---     |
| `commands_create` | [TypeString](https://www.terraform.io/docs/extend/schemas/schema-types.html#typestring) | Create command | `` |
| `commands_delete` | [TypeString](https://www.terraform.io/docs/extend/schemas/schema-types.html#typestring) | Delete command | `` |
| `commands_delete_on_not_exists` | [TypeBool](https://www.terraform.io/docs/extend/schemas/schema-types.html#typebool) | Delete resource when exists fails | `true` |
| `commands_delete_on_read_failure` | [TypeBool](https://www.terraform.io/docs/extend/schemas/schema-types.html#typebool) | Delete resource when read fails | `true` |
| `commands_environment_include_parent` | [TypeBool](https://www.terraform.io/docs/extend/schemas/schema-types.html#typebool) | Include parent environment in the command? | `true` |
| `commands_environment_prefix_new` | [TypeString](https://www.terraform.io/docs/extend/schemas/schema-types.html#typestring) | New environment prefix (skip if empty) | `` |
| `commands_environment_prefix_old` | [TypeString](https://www.terraform.io/docs/extend/schemas/schema-types.html#typestring) | Old environment prefix (skip if empty) | `` |
| `commands_exists` | [TypeString](https://www.terraform.io/docs/extend/schemas/schema-types.html#typestring) | Exists command | `` |
| `commands_exists_expected_exit_code` | [TypeInt](https://www.terraform.io/docs/extend/schemas/schema-types.html#typeint) | Exists command return status | `0` |
| `commands_id` | [TypeString](https://www.terraform.io/docs/extend/schemas/schema-types.html#typestring) | Command building resource id | `` |
| `commands_interpreter` | [TypeList](https://www.terraform.io/docs/extend/schemas/schema-types.html#typelist) | Interpreter and it's arguments, can be template with `command` variable. | `` |
| `commands_prefix` | [TypeString](https://www.terraform.io/docs/extend/schemas/schema-types.html#typestring) | Command prefix shared between all commands | `` |
| `commands_read` | [TypeString](https://www.terraform.io/docs/extend/schemas/schema-types.html#typestring) | Read command | `` |
| `commands_read_format` | [TypeString](https://www.terraform.io/docs/extend/schemas/schema-types.html#typestring) | Templates output types: raw `/^(?<key>[^=]+)=(?<value>[^\n]*)$/` or base64 `/^(?<key>[^=]+)=(?<value_base64>[^\n]*)$/`.  | `raw` |
| `commands_read_line_prefix` | [TypeString](https://www.terraform.io/docs/extend/schemas/schema-types.html#typestring) | Ignore lines in read command without this prefix. | `` |
| `commands_separator` | [TypeString](https://www.terraform.io/docs/extend/schemas/schema-types.html#typestring) | Format for joining 2 commands together without isolating them.  | `%s\\n%s` |
| `commands_should_update` | [TypeString](https://www.terraform.io/docs/extend/schemas/schema-types.html#typestring) | Command indicating whether resource should be updated, non-zero exit code to force update | `` |
| `commands_state_format` | [TypeString](https://www.terraform.io/docs/extend/schemas/schema-types.html#typestring) | State format type.  | `commands_read_format` |
| `commands_update` | [TypeString](https://www.terraform.io/docs/extend/schemas/schema-types.html#typestring) | Update command. Runs destroy then create by default. | `` |
| `commands_working_directory` | [TypeString](https://www.terraform.io/docs/extend/schemas/schema-types.html#typestring) | Working directory to run commands in | `` |
| `dependencies` | [TypeMap](https://www.terraform.io/docs/extend/schemas/schema-types.html#typemap) | Dependencies purely for provider graph walking, otherwise ignored. | `` |
| `logging_buffer_size` | [TypeInt](https://www.terraform.io/docs/extend/schemas/schema-types.html#typeint) | stdout and stderr buffer sizes | `1048576` |
| `logging_jsonformat` | [TypeBool](https://www.terraform.io/docs/extend/schemas/schema-types.html#typebool) | should logs be json instead of plain text?  | `$TF_SCRIPTED_LOGGING_JSONFORMAT` != "" |
| `logging_log_level` | [TypeString](https://www.terraform.io/docs/extend/schemas/schema-types.html#typestring) | Logging level: TRACE, DEBUG, INFO, WARN, ERROR.  | `$TF_SCRIPTED_LOGGING_LOG_LEVEL` |
| `logging_log_path` | [TypeString](https://www.terraform.io/docs/extend/schemas/schema-types.html#typestring) | Extra logs output path.  | `$TF_SCRIPTED_LOGGING_LOG_PATH` |
| `logging_output_line_width` | [TypeInt](https://www.terraform.io/docs/extend/schemas/schema-types.html#typeint) | Width of command's line to use during formatting.  | `$TF_SCRIPTED_LOGGING_OUTPUT_LINE_WIDTH` |
| `logging_output_logging_log_level` | [TypeString](https://www.terraform.io/docs/extend/schemas/schema-types.html#typestring) | Command stdout/stderr log level: TRACE, DEBUG, INFO, WARN, ERROR.  | `$TF_SCRIPTED_LOGGING_OUTPUT_LOG_LEVEL` |
| `logging_provider_name` | [TypeString](https://www.terraform.io/docs/extend/schemas/schema-types.html#typestring) | Name to display in log entries for this provider | `` |
| `templates_left_delim` | [TypeString](https://www.terraform.io/docs/extend/schemas/schema-types.html#typestring) | Left delimiter for templates.  | $TF_SCRIPTED_TEMPLATES_LEFT_DELIM or `{{` |
| `templates_right_delim` | [TypeString](https://www.terraform.io/docs/extend/schemas/schema-types.html#typestring) | Right delimiter for templates.  | $TF_SCRIPTED_TEMPLATES_RIGHT_DELIM or `{{` |
