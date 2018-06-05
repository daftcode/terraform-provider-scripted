
# scripted_resource

# Argument reference

| Argument | Type | Description | Default |
|:---      | ---  | ---         | ---     |
| `context` | [TypeMap](https://www.terraform.io/docs/extend/schemas/schema-types.html#typemap) | Template context for rendering commands | not set |
| `environment` | [TypeMap](https://www.terraform.io/docs/extend/schemas/schema-types.html#typemap) | Environment to run commands in | not set |
| `log_name` | [TypeString](https://www.terraform.io/docs/extend/schemas/schema-types.html#typestring) | Resource name to display in log messages | not set |
| `needs_update` | [TypeBool](https://www.terraform.io/docs/extend/schemas/schema-types.html#typebool) | Helper indicating whether resource should be updated, ignore this. | not set |
| `output` | [TypeMap](https://www.terraform.io/docs/extend/schemas/schema-types.html#typemap) | Output from the read command | not set |
| `state` | [TypeMap](https://www.terraform.io/docs/extend/schemas/schema-types.html#typemap) | Output from create/update commands. Set key: `echo '{{ .StatePrefix }}key=value'`. Delete key: `echo '{{ .StatePrefix }}key={{ .EmptyString }}'` | not set |
