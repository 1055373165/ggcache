package logger

import (
	"os"
	"time"

	"github.com/charmbracelet/log"
)

var Logger *log.Logger

func Init() {
	Logger = log.NewWithOptions(os.Stderr, log.Options{
		ReportCaller:    true,
		ReportTimestamp: true,
		TimeFormat:      time.Kitchen,
		Prefix:          "Baking 🍪 ",
	})

	level := os.Getenv("LogLevel")
	switch level {
	case "debug":
		Logger.SetLevel(log.DebugLevel)
	case "info":
		Logger.SetLevel(log.InfoLevel)
	case "warn":
		Logger.SetLevel(log.WarnLevel)
	case "error":
		Logger.SetLevel(log.ErrorLevel)
	default:
		Logger.SetLevel(log.DebugLevel)
	}
	Logger.Debugf("log level: %v", level)
}
