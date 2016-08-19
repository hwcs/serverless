package model

type Code struct {
	S3Bucket        string
	S3Key           string
	S3ObjectVersion string
	ZipFile         []byte
}

type Function struct {
	FuncCode     Code
	Description  string
	FunctionName string
	Handler      string
	MemorySize   int
	Publish      bool
	Runtime      string
	Timeout      int
}

type EventMapping struct {
	BatchSize        int
	Enabled          bool
	EventSourceArn   string
	FunctionName     string
	StartingPosition string
}

