package shell

import "github.com/hashicorp/go-hclog"

type Config struct {
	BufferSize          int64
	Interpreter         []string
	CommandPrefix       string
	CommandSeparator    string
	CreateCommand       string
	ReadCommand         string
	ReadDeleteOnFailure bool
	ReadFormat          string
	ReadLinePrefix      string
	UpdateCommand       string
	DeleteCommand       string
	ExistsCommand       string
	Logger              hclog.Logger
	CommandLogLevel     hclog.Level
	CommandLogWidth     int
}
