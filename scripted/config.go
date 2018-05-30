package scripted

import "github.com/hashicorp/go-hclog"

type ProviderConfig struct {
	BufferSize               int64
	Interpreter              []string
	WorkingDirectory         string
	CommandPrefix            string
	CommandJoiner            string
	CreateCommand            string
	ReadCommand              string
	DeleteOnReadFailure      bool
	OutputFormat             string
	OutputLinePrefix         string
	StateLinePrefix          string
	UpdateCommand            string
	DeleteCommand            string
	ExistsCommand            string
	Logger                   hclog.Logger
	FileLogger               hclog.Logger
	CommandLogLevel          hclog.Level
	CommandLogWidth          int
	IncludeParentEnvironment bool
	ExistsExpectedStatus     int
	DeleteBeforeUpdate       bool
	CreateAfterUpdate        bool
	CreateBeforeUpdate       bool
	DeleteOnNotExists        bool
	TemplatesRightDelim      string
	TemplatesLeftDelim       string
	NewEnvironmentPrefix     string
	OldEnvironmentPrefix     string
	EmptyString              string
	IdCommand                string
	StateFormat              string
}
