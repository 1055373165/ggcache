package logger

import (
	"log"
	"os"
	"path"
	"time"

	"github.com/sirupsen/logrus"
)

var LogrusObj *logrus.Logger

// Init initializes the global LogrusObj. It can be called multiple times safely.
// If level or output need to be overridden, call this before other packages use logger.
func Init(level logrus.Level, json bool) {
	if LogrusObj != nil {
		LogrusObj.SetLevel(level)
		if json {
			LogrusObj.SetFormatter(&logrus.JSONFormatter{
				TimestampFormat: time.RFC3339,
			})
		}
		return
	}

	logger := logrus.New()
	logger.Out = os.Stdout
	logger.SetLevel(level)
	if json {
		logger.SetFormatter(&logrus.JSONFormatter{
			TimestampFormat: time.RFC3339,
		})
	} else {
		logger.SetFormatter(&logrus.TextFormatter{
			TimestampFormat: "2006-01-02 15:04:05",
			FullTimestamp:   true,
		})
	}
	LogrusObj = logger
}

func init() {
	if LogrusObj != nil {
		outputFile, _ := setOutputFile()
		LogrusObj.Out = outputFile
		return
	}

	// default init -> debug level JSON for easier parsing in k8s / cloud native env
	Init(logrus.DebugLevel, true)
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
