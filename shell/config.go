package shell

import (
	"log"
	"github.com/hashicorp/logutils"
)

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
	Logger              *log.Logger
	LogOutput           bool
	LogLevelFilter      *logutils.LevelFilter
}
