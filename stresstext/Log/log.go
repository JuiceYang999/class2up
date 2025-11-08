package Log

import (
	"StressTest-INC-Cloud/Config"
	"log"
	"os"
)

var Log *log.Logger = nil

func init() {
	logFile := Config.GetAppConfig().Log.File
	if logFile == "" {
		logFile = "log.txt"
	}
	logIO, err := os.OpenFile(logFile, os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		logIO = os.Stdout
	}
	Log = log.New(logIO, "", log.LstdFlags)
}

func LogInfo(s string)  {
	Log.Println(s)
}