package golik

import (
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

func newLogger() *Clove {
	return &Clove{
		Name: "logger",
		Sync: true,
		Behavior: func(l *LogEntry) {
			entry := l.logrusEntry()
			switch l.LogLevel {
			case LogLevel_DEBUG:
				entry.Debug(l.Message)
			case LogLevel_WARN:
				entry.Warn(l.Message)
			case LogLevel_ERROR:
				entry.Error(l.Message)
			case LogLevel_FATAL:
				entry.Fatal(l.Message)
			default:
				entry.Info(l.Message)
			}
		},
		PreStart: func() {
			initLogging()
		},
		PostStart: func(ctx CloveContext) error {
			if err := ctx.Subscribe(func(data interface{}) bool {
				switch data.(type) {
				case LogEntry, *LogEntry:
					return true
				}
				return false
			}); err != nil {
				return err
			}
			ctx.Info("Logging is up")
			return nil
		},
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
				FullTimestamp:   true,
			})
		}

		loggingInit = true
	}
}

func init() {
	viper.SetDefault("golik.log.level", "INFO")
	viper.SetDefault("golik.log.formatter", "text")
}
