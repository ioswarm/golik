package golik

import (
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

type Loggable interface{
	Debug(msg string, values ...interface{})
	Info(msg string, values ...interface{})
	Warn(msg string, values ...interface{})
	Error(msg string, values ...interface{})
	Panic(msg string, values ...interface{})
}

// TODO create configured logger

var logrusInit bool

func initLogrus() {
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
		logrus.SetFormatter(&logrus.JSONFormatter{})
	default:
		logrus.SetFormatter(&logrus.TextFormatter{ 
			FullTimestamp: true,
		})
	}
}

func newLogrusLogger(meta map[string]interface{}) *LogrusLogger {
	if !logrusInit {
		initLogrus()
		logrusInit = true
	}
	return &LogrusLogger{ logrus.WithFields(logrus.Fields(meta)) }
}

type LogrusLogger struct {
	logger *logrus.Entry
}

func (l *LogrusLogger) Debug(msg string, values ...interface{}) {
	if len(values) > 0 {
		l.logger.Debugf(msg, values...)
	} else {
		l.logger.Debug(msg)
	}
}

func (l *LogrusLogger) Info(msg string, values ...interface{}) {
	if len(values) > 0 {
		l.logger.Infof(msg, values...)
	} else {
		l.logger.Info(msg)
	}
}

func (l *LogrusLogger) Warn(msg string, values ...interface{}) {
	if len(values) > 0 {
		l.logger.Warnf(msg, values...)
	} else {
		l.logger.Warn(msg)
	}
}

func (l *LogrusLogger) Error(msg string, values ...interface{}) {
	if len(values) > 0 {
		l.logger.Errorf(msg, values...)
	} else {
		l.logger.Error(msg)
	}
}

func (l *LogrusLogger) Panic(msg string, values ...interface{}) {
	if len(values) > 0 {
		l.logger.Panicf(msg, values...)
	} else {
		l.logger.Panic(msg)
	}
}
