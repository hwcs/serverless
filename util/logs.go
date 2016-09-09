package util

import (
	"log"
	"os"
)

type Loggers struct {
	LogInfo    *log.Logger
	LogWarning *log.Logger
	LogError   *log.Logger
}

var Logs Loggers

func InitLogger() {
	fileName := "serverless.log"
	logFile, err := os.Create(fileName)
	if err != nil {
		log.Fatalln("InitLogger, open log file error !")
		return
	}
	defer logFile.Close()

	Logs.LogInfo = log.New(logFile, "[INFO]", log.Llongfile)
	Logs.LogWarning = log.New(logFile, "[WARNING]", log.Llongfile)
	Logs.LogError = log.New(logFile, "[ERROR]", log.Llongfile)
}

