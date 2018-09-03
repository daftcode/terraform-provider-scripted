package scripted

import "github.com/daftcode/terraform-provider-scripted/version"

const JsonContextEnvKey = "TF_SCRIPTED_CONTEXT"
const DefaultEnvPrefix = "TF_SCRIPTED_"
const DefaultTriggerString = `ndn4VFxYG489bUmV6xKjKFE0RYQIJdts`
const DefaultStatePrefix = `WViRV1TbGAGehAYFL8g3ZL8o1cg1bxaq`
const DefaultLinePrefix = `QmGRizGk1fdPEBVVZSGkCRPJRgAe9p07B`
const DefaultEmptyString = `ZVaXr3jCd80vqJRhBP9t83LrpWIdNKWJ`
const Version = version.Version

var DefaultWindowsInterpreter = []string{"cmd", "/C", "{{ .command }}"}
var DefaultInterpreter = []string{"bash", "-Eeuo", "pipefail", "-c", "{{ .command }}"}

var ValidLogLevelsStrings = []string{"TRACE", "DEBUG", "INFO", "WARN", "ERROR"}

const TriggerStringTpl = `{{ .TriggerString }}`

type TerraformOperation string

const (
	OperationCreate        TerraformOperation = "create"
	OperationRead          TerraformOperation = "read"
	OperationExists        TerraformOperation = "exists"
	OperationUpdate        TerraformOperation = "update"
	OperationDelete        TerraformOperation = "delete"
	OperationCustomizeDiff TerraformOperation = "customizediff"
)

const (
	CommandId                       string = "commands_id"
	CommandDependencies             string = "commands_dependencies"
	CommandNeedsUpdate              string = "commands_needs_update"
	CommandCustomizeDiffComputeKeys string = "commands_customizediff_computekeys"
	CommandCreate                   string = "commands_create"
	CommandDelete                   string = "commands_delete"
	CommandExists                   string = "commands_exists"
	CommandRead                     string = "commands_read"
	CommandUpdate                   string = "commands_update"
)

var AllowedCommands = map[string]bool{
	CommandId:                       true,
	CommandDependencies:             true,
	CommandNeedsUpdate:              true,
	CommandCustomizeDiffComputeKeys: true,
	CommandCreate:                   true,
	CommandDelete:                   true,
	CommandExists:                   true,
	CommandRead:                     true,
	CommandUpdate:                   true,
}
