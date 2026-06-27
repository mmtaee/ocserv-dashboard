package routing

import (
	"fmt"
	"io"

	"github.com/mmtaee/ocserv-dashboard/core/pkg/logger"

	LabstackLog "github.com/labstack/gommon/log"
)

type WrapperLogger struct {
	Log *logger.Logger
}

func NewLoggerWrapper(l *logger.Logger) *WrapperLogger {
	return &WrapperLogger{Log: l}
}

func (l *WrapperLogger) Output() io.Writer {
	return io.Discard
}

func (l *WrapperLogger) SetOutput(w io.Writer) {}

func (l *WrapperLogger) Prefix() string { return "" }

func (l *WrapperLogger) SetPrefix(p string) {}

func (l *WrapperLogger) Level() LabstackLog.Lvl {
	return LabstackLog.INFO
}

func (l *WrapperLogger) SetLevel(v LabstackLog.Lvl) {}

func (l *WrapperLogger) SetHeader(h string) {}

func (l *WrapperLogger) send(level logger.LogLevel, format string, args ...interface{}) {
	if l.Log == nil {
		return
	}

	msg := logger.SafeSprintf(format, args...)

	switch level {
	case logger.InfoLevel:
		logger.Info("%s", msg)
	case logger.WarnLevel:
		logger.Warn("%s", msg)
	case logger.ErrorLevel:
		logger.Error("%s", msg)
	case logger.FatalLevel:
		logger.Error("%s", msg)
	}
}
func (l *WrapperLogger) Print(i ...interface{}) {
	l.send(logger.InfoLevel, "%v", fmt.Sprint(i...))
}

func (l *WrapperLogger) Printf(format string, args ...interface{}) {
	l.send(logger.InfoLevel, format, args...)
}

func (l *WrapperLogger) Printj(j LabstackLog.JSON) {
	l.send(logger.InfoLevel, "%v", j)
}

func (l *WrapperLogger) Debug(i ...interface{}) {
	l.send(logger.InfoLevel, "%v", fmt.Sprint(i...))
}

func (l *WrapperLogger) Debugf(format string, args ...interface{}) {
	l.send(logger.InfoLevel, format, args...)
}

func (l *WrapperLogger) Debugj(j LabstackLog.JSON) {
	l.send(logger.InfoLevel, "%v", j)
}

func (l *WrapperLogger) Info(i ...interface{}) {
	l.send(logger.InfoLevel, "%v", fmt.Sprint(i...))
}

func (l *WrapperLogger) Infof(format string, args ...interface{}) {
	l.send(logger.InfoLevel, format, args...)
}

func (l *WrapperLogger) Infoj(j LabstackLog.JSON) {
	l.send(logger.InfoLevel, "%v", j)
}

func (l *WrapperLogger) Warn(i ...interface{}) {
	l.send(logger.WarnLevel, "%v", fmt.Sprint(i...))
}

func (l *WrapperLogger) Warnf(format string, args ...interface{}) {
	l.send(logger.WarnLevel, format, args...)
}

func (l *WrapperLogger) Warnj(j LabstackLog.JSON) {
	l.send(logger.WarnLevel, "%v", j)
}

func (l *WrapperLogger) Error(i ...interface{}) {
	l.send(logger.ErrorLevel, "%v", fmt.Sprint(i...))
}

func (l *WrapperLogger) Errorf(format string, args ...interface{}) {
	l.send(logger.ErrorLevel, format, args...)
}

func (l *WrapperLogger) Errorj(j LabstackLog.JSON) {
	l.send(logger.ErrorLevel, "%v", j)
}

func (l *WrapperLogger) Fatal(i ...interface{}) {
	l.send(logger.FatalLevel, "%v", fmt.Sprint(i...))
}

func (l *WrapperLogger) Fatalf(format string, args ...interface{}) {
	l.send(logger.FatalLevel, format, args...)
}

func (l *WrapperLogger) Fatalj(j LabstackLog.JSON) {
	l.send(logger.FatalLevel, "%v", j)
}

func (l *WrapperLogger) Panic(i ...interface{}) {
	msg := logger.SafeSprintf("%v", fmt.Sprint(i...))
	l.send(logger.FatalLevel, "%s", msg)
	panic(msg)
}

func (l *WrapperLogger) Panicf(format string, args ...interface{}) {
	msg := logger.SafeSprintf(format, args...)
	l.send(logger.FatalLevel, "%s", msg)
	panic(msg)
}

func (l *WrapperLogger) Panicj(j LabstackLog.JSON) {
	msg := logger.SafeSprintf("%v", j)
	l.send(logger.FatalLevel, "%s", msg)
	panic(msg)
}
