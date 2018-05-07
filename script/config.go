package script

import "github.com/hashicorp/go-hclog"

type Config struct {
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
	CommandLogLevel          hclog.Level
	CommandLogWidth          int
	IncludeParentEnvironment bool
}
