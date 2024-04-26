package logger

import (
	"log"
	"os"
	"path"
	"time"

	"github.com/sirupsen/logrus"
)

var LogrusObj *logrus.Logger

func init() {
	if LogrusObj != nil {
		outputFile, _ := setOutputFile()
		LogrusObj.Out = outputFile
		return
	}

	logger := logrus.New()
	// outputFile, _ := setOutputFile()
	// logger.Out = outputFile
	logger.Out = os.Stdout
	logger.SetLevel(logrus.DebugLevel)
	logger.SetFormatter(&logrus.TextFormatter{
		TimestampFormat: "2006-01-02 15:04:05",
	})

	LogrusObj = logger
}

func setOutputFile() (*os.File, error) {
	now := time.Now()
	logFilePath := ""

	if dir, err := os.Getwd(); err == nil {
		logFilePath = dir + "/logs/"
	}

	_, err := os.Stat(logFilePath)
	if os.IsNotExist(err) {
		if err := os.MkdirAll(logFilePath, 0777); err != nil {
			log.Println(err.Error())
			return nil, err
		}
	}

	logFileName := now.Format("2006-01-02") + ".log"
	fileName := path.Join(logFilePath, logFileName)
	if _, err := os.Stat(fileName); err != nil {
		if _, err := os.Create(fileName); err != nil {
			log.Println(err.Error())
			return nil, err
		}
	}

	output, err := os.OpenFile(fileName, os.O_APPEND|os.O_WRONLY, os.ModeAppend)
	if err != nil {
		log.Println(err.Error())
		return nil, err
	}
	return output, nil
}
