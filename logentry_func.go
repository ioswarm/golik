package golik

import (
	"fmt"

	"github.com/sirupsen/logrus"
	"google.golang.org/protobuf/types/known/timestamppb"
)


func (l *LogEntry) WithMeta(meta map[string]string) *LogEntry {
	l.Meta = meta
	return l
}

func (l *LogEntry) logrusEntry() *logrus.Entry {
	lf := logrus.Fields{}
	for key := range l.Meta {
		lf[key] = l.Meta[key]
	}
	return logrus.WithFields(lf)
}


func NewLogDebug(msg string, values ...interface{}) *LogEntry {
	m := msg
	if len(values) > 0 {
		m = fmt.Sprintf(m, values...)
	}
	return &LogEntry{
		Message: m,
		LogLevel: LogLevel_DEBUG,
		Timestamp: timestamppb.Now(),
	}
}

func NewLogInfo(msg string, values ...interface{}) *LogEntry {
	m := msg
	if len(values) > 0 {
		m = fmt.Sprintf(m, values...)
	}
	return &LogEntry{
		Message: m,
		LogLevel: LogLevel_INFO,
		Timestamp: timestamppb.Now(),
	}
}

func NewLogWarn(msg string, values ...interface{}) *LogEntry {
	m := msg
	if len(values) > 0 {
		m = fmt.Sprintf(m, values...)
	}
	return &LogEntry{
		Message: m,
		LogLevel: LogLevel_WARN,
		Timestamp: timestamppb.Now(),
	}
}

func NewLogError(msg string, values ...interface{}) *LogEntry {
	m := msg
	if len(values) > 0 {
		m = fmt.Sprintf(m, values...)
	}
	return &LogEntry{
		Message: m,
		LogLevel: LogLevel_ERROR,
		Timestamp: timestamppb.Now(),
	}
}
