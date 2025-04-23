package main

import (
	"fmt"
	"os"
	"time"
)

type DivaLog struct {
	logLevel  int
	logFile *os.File
}

func NewDivaLog(_logLevel int, logPath string) (dl DivaLog) {
	dl = DivaLog{}

	dl.logLevel = _logLevel

	if logPath != "" {
		file, err := os.Create(logPath)
		if err != nil {
			dl.Log("Error opening log file", 0)
			dl.Log("Error details: " + err.Error(), 1)
		}

		dl.logFile = file
	}

	return
}

func (dl DivaLog) Log(message any, level int) {
	if dl.logLevel >= level {
		strMessage := fmt.Sprint(message)

		fmt.Println("[MultiDiva] " + strMessage)

		if(dl.logFile != nil) {
			_, err := dl.logFile.WriteString("[" + time.Now().Format("2006-01-02 15:04:05") + "] " + strMessage)
			if err != nil {
				fmt.Println("[MultiDiva] Failed to write log to file. Error details: " + err.Error())
			}
		}
	}
}