
# scripted_resource

# Arguments reference

| Argument | Type | Description | Default |
|:---      | ---  | ---         | ---     |
| `context` | [map](https://www.terraform.io/docs/extend/schemas/schema-types.html#typemap) | Template context for rendering commands | not set |
| `environment` | [map](https://www.terraform.io/docs/extend/schemas/schema-types.html#typemap) | Environment to run commands in | not set |
| `output` | [map](https://www.terraform.io/docs/extend/schemas/schema-types.html#typemap) | Output from the read command | not set |
| `revision` | [string](https://www.terraform.io/docs/extend/schemas/schema-types.html#typestring) | Resource's revision | not set |
| `state` | [map](https://www.terraform.io/docs/extend/schemas/schema-types.html#typemap) | Output from create/update commands. Set key: `echo '{{ .StatePrefix }}key=value'`. Delete key: `echo '{{ .StatePrefix }}key={{ .EmptyString }}'` | not set |
