package logger

import (
	"sync"
	"time"
)

type LogMessage struct {
	Level   LogLevel
	Message string
	Time    time.Time
}

type Logger struct {
	logChan chan LogMessage
	done    chan struct{}
	once    sync.Once
}

type LogLevel string

// Log levels
const (
	InfoLevel  LogLevel = "INFO"
	WarnLevel  LogLevel = "WARNING"
	ErrorLevel LogLevel = "ERROR"
	FatalLevel LogLevel = "FATAL"
)

// ANSI color codes for terminal output
const (
	ColorReset   = "\033[0m"
	ColorBlue    = "\033[34m"   // Info
	ColorYellow  = "\033[33m"   // Warning
	ColorRed     = "\033[31m"   // Error
	ColorBoldRed = "\033[1;31m" // Fatal
)

var Log *Logger

var LevelColors = map[LogLevel]string{
	InfoLevel:  ColorBlue,
	WarnLevel:  ColorYellow,
	ErrorLevel: ColorRed,
	FatalLevel: ColorBoldRed,
}
