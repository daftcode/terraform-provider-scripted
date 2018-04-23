package shell

type Config struct {
	BufferSize       int64
	Interpreter      []string
	CommandPrefix    string
	CommandSeparator string
	WorkingDirectory string
	CreateCommand    string
	ReadCommand      string
	ReadFormat       string
	ReadLinePrefix   string
	UpdateCommand    string
	DeleteCommand    string
	ExistsCommand    string
}
