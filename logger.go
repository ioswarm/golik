package golik

import (
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

type LogLevel uint8

const (
	DEBUG LogLevel = iota
	INFO
	WARN
	ERROR
	PANIC
)

type LogEntry struct {
	Level LogLevel
	Meta map[string]interface{}
	Message string
	Values []interface{}
}

type Loggable interface{
	Logger() *logrus.Entry // TODO customize Logger 
	Log(entry LogEntry)
	Debug(msg string, values ...interface{})  // TODO impl debug-internal?
	Info(msg string, values ...interface{})
	Warn(msg string, values ...interface{})
	Error(msg string, values ...interface{})
	Panic(msg string, values ...interface{})
}


func HandleLogEntry(il *logrus.Entry, e LogEntry) {
	switch e.Level {
	case DEBUG:
		if len(e.Values) > 0 {
			il.Debugf(e.Message, e.Values...)
		} else {
			il.Debug(e.Message)
		}
	case WARN:
		if len(e.Values) > 0 {
			il.Warnf(e.Message, e.Values...)
		} else {
			il.Warn(e.Message)
		}
	case ERROR:
		if len(e.Values) > 0 {
			il.Errorf(e.Message, e.Values...)
		} else {
			il.Error(e.Message)
		}
	case PANIC:
		if len(e.Values) > 0 {
			il.Panicf(e.Message, e.Values...)
		} else {
			il.Panic(e.Message)
		}
	default:
		if len(e.Values) > 0 {
			il.Infof(e.Message, e.Values...)
		} else {
			il.Info(e.Message)
		}
	}
}

var loggingInit bool

func initLogging() {
	if !loggingInit {
		initSettings()

		switch viper.GetString("golik.log.level") {
		case "DEBUG": 
			logrus.SetLevel(logrus.DebugLevel)
		case "WARN":
			logrus.SetLevel(logrus.WarnLevel)
		case "ERROR":
			logrus.SetLevel(logrus.ErrorLevel)
		case "FATAL":
			logrus.SetLevel(logrus.FatalLevel)
		case "PANIC":
			logrus.SetLevel(logrus.PanicLevel)
		default:
			logrus.SetLevel(logrus.InfoLevel)
		}
	
		switch viper.GetString("golik.log.formatter") {
		case "json":
			logrus.SetFormatter(&logrus.JSONFormatter{
				TimestampFormat: "2006-01-02T15:04:05.000Z07:00",
			})
		default:
			logrus.SetFormatter(&logrus.TextFormatter{ 
				TimestampFormat: "2006-01-02T15:04:05.000Z07:00",
				FullTimestamp: true,
			})
		}
		loggingInit = true
	}
}


func init() {
	
	// logging
	viper.SetDefault("golik.log.level", "INFO")
	viper.SetDefault("golik.log.formatter", "text")

}