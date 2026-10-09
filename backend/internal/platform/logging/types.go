package logger

import (
	"io"
	"time"
)

type LogMessage struct {
	Service  string
	Level    LogLevel
	Message  string
	Time     time.Time
	ClientIP string
	Status   int
	Latency  time.Duration
}

// StreamEntry is a timestamped line received from a service log source.
type StreamEntry struct {
	Message   string
	Timestamp time.Time
}

type Logger struct {
	logChan chan LogMessage
	done    chan struct{}
	output  io.Writer
}

type LogLevel string

// Log levels
const (
	DebugLevel LogLevel = "DEBUG"
	InfoLevel  LogLevel = "INFO"
	WarnLevel  LogLevel = "WARNING"
	ErrorLevel LogLevel = "ERROR"
	FatalLevel LogLevel = "FATAL"
)

// ANSI color codes for terminal output
const (
	ColorReset   = "\033[0m"
	ColorCyan    = "\033[36m"   // Debug
	ColorGreen   = "\033[32m"   // Info
	ColorYellow  = "\033[33m"   // Warning
	ColorRed     = "\033[31m"   // Error
	ColorBoldRed = "\033[1;31m" // Fatal
)

var Log *Logger

var LevelColors = map[LogLevel]string{
	DebugLevel: ColorCyan,
	InfoLevel:  ColorGreen,
	WarnLevel:  ColorYellow,
	ErrorLevel: ColorRed,
	FatalLevel: ColorBoldRed,
}
