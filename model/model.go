package model

const (
	CODE_TYPE_INLINE  = "inline"
	CODE_TYPE_ZIPFILE = "zip"
)

var RUNTIME_LANGUAGE = map[string]string{
	"Python 2.7":  ".py",
	"Node.js 4.3": ".js",
	"Java":        ".java",
	"C":           ".c",
	"C++":         ".cpp",
}

type Code struct {
	S3Bucket        string
	S3Key           string
	S3ObjectVersion string
	File            string
	CodeType        string // inline:inline file, zip: zip file
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

