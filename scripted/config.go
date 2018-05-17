package scripted

import "github.com/hashicorp/go-hclog"

type ProviderConfig struct {
	BufferSize               int64
	Interpreter              []string
	WorkingDirectory         string
	CommandPrefix            string
	CommandIsolator          string
	CommandJoiner            string
	CreateCommand            string
	ReadCommand              string
	DeleteOnReadFailure      bool
	ReadFormat               string
	ReadLinePrefix           string
	UpdateCommand            string
	DeleteCommand            string
	ExistsCommand            string
	Logger                   hclog.Logger
	FileLogger               hclog.Logger
	CommandLogLevel          hclog.Level
	CommandLogWidth          int
	IncludeParentEnvironment bool
	ExistsExpectedStatus     int
	LogProviderName          string
	DeleteBeforeUpdate       bool
	CreateAfterUpdate        bool
	CreateBeforeUpdate       bool
	DeleteOnNotExists        bool
	TemplatesRightDelim      string
	TemplatesLeftDelim       string
}
