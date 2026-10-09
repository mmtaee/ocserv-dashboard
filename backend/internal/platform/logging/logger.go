package logger

import (
	"context"
	"fmt"
	"os"
	"runtime"
	"strings"
	"time"
)

func Init(ctx context.Context, bufferSize int) {
	l := &Logger{
		logChan: make(chan LogMessage, bufferSize),
		done:    make(chan struct{}),
		output:  os.Stdout,
	}

	Log = l
	go l.worker(ctx)
}

func GetLogger() *Logger {
	return Log
}

func (l *Logger) worker(ctx context.Context) {
	defer close(l.done)
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

// Wait blocks until the logging worker has flushed its buffered messages.
func Wait() {
	if Log != nil {
		<-Log.done
	}
}

func (l *Logger) print(msg LogMessage) {
	color := LevelColors[msg.Level]
	fields := ""
	if msg.ClientIP != "" {
		fields += fmt.Sprintf(" [client_ip=%s]", msg.ClientIP)
	}
	if msg.Status != 0 {
		fields += fmt.Sprintf(" [status=%d]", msg.Status)
	}
	if msg.Latency != 0 {
		fields += fmt.Sprintf(" [latency=%s]", msg.Latency.Round(time.Microsecond))
	}
	fmt.Fprintf(
		l.output,
		"%s[%s] [%s] [service=%s]%s %s%s\n",
		color,
		msg.Time.Format("2006-01-02 15:04:05"),
		msg.Level,
		msg.Service,
		fields,
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

func send(msg LogMessage) {
	if Log != nil && Log.logChan != nil {
		if msg.Time.IsZero() {
			msg.Time = time.Now()
		}
		if msg.Service == "" {
			msg.Service = callerService()
		}
		select {
		case Log.logChan <- msg:
		default:
			// drop if buffer full
		}
	}
}

func callerService() string {
	_, file, _, ok := runtime.Caller(4)
	if !ok {
		return "backend"
	}
	return serviceForFile(file)
}

func serviceForFile(file string) string {
	switch {
	case strings.Contains(file, "/admin_api/"):
		return "admin-api"
	case strings.Contains(file, "/customer_api/"):
		return "customer-api"
	case strings.Contains(file, "/telegram_bot/"):
		return "telegram-bot"
	case strings.Contains(file, "/worker/"):
		return "worker"
	}
	return "backend"
}

func Debug(format string, args ...interface{}) {
	send(LogMessage{Level: DebugLevel, Message: SafeSprintf(format, args...)})
}

func Info(format string, args ...interface{}) {
	send(LogMessage{Level: InfoLevel, Message: SafeSprintf(format, args...)})
}

func Warn(format string, args ...interface{}) {
	send(LogMessage{Level: WarnLevel, Message: SafeSprintf(format, args...)})
}

func Error(format string, args ...interface{}) {
	send(LogMessage{Level: ErrorLevel, Message: SafeSprintf(format, args...)})
}

func Fatal(format string, args ...interface{}) {
	send(LogMessage{Level: FatalLevel, Message: SafeSprintf(format, args...)})
}

// Request records request metadata without inspecting headers, query strings, or bodies.
func Request(service, clientIP, message string, status int, latency time.Duration) {
	level := InfoLevel
	if status >= 500 {
		level = ErrorLevel
	} else if status >= 400 {
		level = WarnLevel
	}
	send(LogMessage{Service: service, Level: level, Message: message, ClientIP: clientIP, Status: status, Latency: latency})
}
