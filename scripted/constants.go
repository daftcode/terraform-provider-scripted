package scripted

const DefaultEnvPrefix = "TF_SCRIPTED_"
const DefaultTriggerString = `ndn4VFxYG489bUmV6xKjKFE0RYQIJdts`
const DefaultStatePrefix = `WViRV1TbGAGehAYFL8g3ZL8o1cg1bxaq`
const DefaultLinePrefix = DefaultStatePrefix
const DefaultEmptyString = `ZVaXr3jCd80vqJRhBP9t83LrpWIdNKWJ`

var DefaultWindowsInterpreter = []string{"cmd", "/C", "{{ .command }}"}
var DefaultInterpreter = []string{"bash", "-Eeuo", "pipefail", "-c", "{{ .command }}"}

var ValidLogLevelsStrings = []string{"TRACE", "DEBUG", "INFO", "WARN", "ERROR"}

const TriggerStringTpl = `{{ .TriggerString }}`
