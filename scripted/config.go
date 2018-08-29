package scripted

import "github.com/hashicorp/go-hclog"

type EnvironmentConfig struct {
	PrefixNew        string
	PrefixOld        string
	IncludeParent    bool
	InheritVariables []string
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
	Environment         *EnvironmentConfig
	Templates           *CommandTemplates
	Output              *OutputConfig
	CreateAfterUpdate   bool
	DeleteBeforeUpdate  bool
	DeleteOnNotExists   bool
	DeleteOnReadFailure bool
	Separator           string
	WorkingDirectory    string
	TriggerString       string
}

type TemplatesConfig struct {
	LeftDelim  string
	RightDelim string
}

type ProviderConfig struct {
	Commands                   *CommandsConfig
	ComputeStateKeys           []string
	ComputeOutputKeys          []string
	Logging                    *Logging
	Templates                  *TemplatesConfig
	RunningMessageInterval     float64
	EmptyString                string
	LoggingBufferSize          int64
	OutputUseDefaultLinePrefix bool
	outputLinePrefix           string
	OutputFormat               string
	StateFormat                string
	StateLinePrefix            string
	LinePrefix                 string
}

func (pc *ProviderConfig) OutputLinePrefix() string {
	if pc.OutputUseDefaultLinePrefix {
		return pc.LinePrefix
	}
	return pc.outputLinePrefix
}
