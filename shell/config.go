package shell

type Config struct {
	BufferSize       int64
	WorkingDirectory string
	CreateCommand    string
	ReadCommand      string
	ReadFormat       string
	ReadLinePrefix   string
	UpdateCommand    string
	DeleteCommand    string
	ExistsCommand    string
}
