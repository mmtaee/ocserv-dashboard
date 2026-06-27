package logger

import (
	"context"
	"fmt"
	"os"
	"sync/atomic"
	"time"
)

var inSend int32

func Init(ctx context.Context, bufferSize int) {
	l := &Logger{
		logChan: make(chan LogMessage, bufferSize),
	}

	Log = l
	go l.worker(ctx)
}

func GetLogger() *Logger {
	return Log
}

func (l *Logger) worker(ctx context.Context) {
	for {
		select {
		case msg := <-l.logChan:
			l.print(msg)
		case <-ctx.Done():
			for {
				select {
				case msg := <-l.logChan:
					l.print(msg)
				default:
					// channel empty, stop draining
					return
				}
			}
		}
	}
}

func (l *Logger) print(msg LogMessage) {
	color := LevelColors[msg.Level]

	fmt.Printf(
		"%s[%s] [%s] %s%s\n",
		color,
		msg.Time.Format("2006-01-02 15:04:05"),
		msg.Level,
		msg.Message,
		ColorReset,
	)
}

func SafeSprintf(format string, args ...interface{}) (result string) {
	defer func() {
		if r := recover(); r != nil {
			// fallback message if formatting fails
			result = fmt.Sprintf("[LOG FORMAT ERROR] format=%q args=%v", format, args)
		}
	}()
	result = fmt.Sprintf(format, args...)
	return
}

func send(level LogLevel, message string) {
	if !atomic.CompareAndSwapInt32(&inSend, 0, 1) {
		// already inside send, drop message to prevent recursion
		return
	}
	defer atomic.StoreInt32(&inSend, 0)

	if Log != nil && Log.logChan != nil {
		select {
		case Log.logChan <- LogMessage{
			Level:   level,
			Message: message,
			Time:    time.Now(),
		}:
		default:
			// drop if buffer full
		}
	}
}

func Info(format string, args ...interface{}) {
	send(InfoLevel, SafeSprintf(format, args...))
}

func Warn(format string, args ...interface{}) {
	send(WarnLevel, SafeSprintf(format, args...))
}

func Error(format string, args ...interface{}) {
	send(ErrorLevel, SafeSprintf(format, args...))
}

func Fatal(format string, args ...interface{}) {
	msg := SafeSprintf(format, args...)
	fmt.Fprintf(os.Stderr, "[FATAL] %s\n", msg) // direct write
	os.Exit(1)
}
