package scripted

import (
	"github.com/hashicorp/go-hclog"
	"github.com/hashicorp/terraform/terraform"
)

type EnvironmentConfig struct {
	PrefixNew          string
	PrefixOld          string
	IncludeParent      bool
	InheritVariables   []string
	IncludeJsonContext bool
}

type CommandTemplates struct {
	Create                   string
	CustomizeDiffComputeKeys string
	Delete                   string
	Dependencies             string
	Exists                   string
	Id                       string
	Interpreter              []string
	ModifyPrefix             string
	Prefix                   string
	PrefixFromEnv            string
	Read                     string
	NeedsUpdate              string
	Update                   string
}

type OutputConfig struct {
	LogLevel  hclog.Level
	LineWidth int
	LogPids   bool
	LogIids   bool
}

type CommandsConfig struct {
	Environment                 *EnvironmentConfig
	Templates                   *CommandTemplates
	Output                      *OutputConfig
	DeleteOnNotExists           bool
	DeleteOnReadFailure         bool
	Separator                   string
	WorkingDirectory            string
	TriggerString               string
	InterpreterIsProvider       bool
	InterpreterProviderCommands []string
	DependenciesNotMetError     bool
}

type TemplatesConfig struct {
	LeftDelim  string
	RightDelim string
}

type ProviderConfig struct {
	Commands                   *CommandsConfig
	StateComputeKeys           []string
	OutputComputeKeys          []string
	logging                    *Logging
	Templates                  *TemplatesConfig
	RunningMessageInterval     float64
	EmptyString                string
	LoggingBufferSize          int64
	OutputUseDefaultLinePrefix bool
	OutputLinePrefix           string
	OutputFormat               string
	StateFormat                string
	StateLinePrefix            string
	LinePrefix                 string
	Version                    string
	InstanceState              *terraform.InstanceState
	EnvPrefix                  string
	OpenParentStderr           bool
}
