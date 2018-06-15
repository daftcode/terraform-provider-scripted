package scripted

import "github.com/hashicorp/go-hclog"

type EnvironmentConfig struct {
	PrefixNew     string
	PrefixOld     string
	IncludeParent bool
}

type CommandTemplates struct {
	Create        string
	Delete        string
	Dependencies  string
	Exists        string
	Id            string
	Interpreter   []string
	ModifyPrefix  string
	Prefix        string
	PrefixFromEnv string
	Read          string
	NeedsDelete   string
	NeedsUpdate   string
	Update        string
}

type OutputConfig struct {
	LogLevel  hclog.Level
	LineWidth int
	LogPids   bool
	LogIids   bool
}

type CommandsConfig struct {
	Environment               *EnvironmentConfig
	Templates                 *CommandTemplates
	Output                    *OutputConfig
	CreateAfterUpdate         bool
	DeleteBeforeUpdate        bool
	DeleteOnNotExists         bool
	DeleteOnReadFailure       bool
	Separator                 string
	WorkingDirectory          string
	NeedsUpdateExpectedOutput string
	ExistsExpectedOutput      string
	DependenciesTriggerOutput string
	NeedsDeleteExpectedOutput string
}

type TemplatesConfig struct {
	LeftDelim  string
	RightDelim string
}

type ProviderConfig struct {
	Commands          *CommandsConfig
	Logging           *Logging
	Templates         *TemplatesConfig
	EmptyString       string
	LoggingBufferSize int64
	OutputLinePrefix  string
	OutputFormat      string
	StateFormat       string
	StateLinePrefix   string
}
