package utils

import (
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"sync"
	"time"
)

// LogLevel represents the severity of a log message.
type LogLevel int

const (
	LevelDebug LogLevel = iota
	LevelInfo
	LevelWarn
	LevelError
)

func (l LogLevel) String() string {
	switch l {
	case LevelDebug:
		return "DEBUG"
	case LevelInfo:
		return "INFO"
	case LevelWarn:
		return "WARN"
	case LevelError:
		return "ERROR"
	default:
		return "UNKNOWN"
	}
}

// Logger is a simple structured logger that writes to file and optionally stderr.
type Logger struct {
	mu       sync.Mutex
	file     *os.File
	logger   *log.Logger
	level    LogLevel
	debug    bool
}

// NewLogger creates a new logger. If debug is true, also logs to stderr.
func NewLogger(logPath string, debug bool) (*Logger, error) {
	if err := os.MkdirAll(filepath.Dir(logPath), 0700); err != nil {
		// best effort
	}

	f, err := os.OpenFile(logPath, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0600)
	if err != nil {
		// Fall back to stderr only
		l := &Logger{
			logger: log.New(os.Stderr, "", 0),
			level:  LevelInfo,
			debug:  debug,
		}
		if debug {
			l.level = LevelDebug
		}
		return l, nil
	}

	var w io.Writer
	if debug {
		w = io.MultiWriter(f, os.Stderr)
	} else {
		w = f
	}

	l := &Logger{
		file:   f,
		logger: log.New(w, "", 0),
		level:  LevelInfo,
		debug:  debug,
	}
	if debug {
		l.level = LevelDebug
	}
	return l, nil
}

func (l *Logger) log(level LogLevel, format string, args ...interface{}) {
	if level < l.level {
		return
	}
	l.mu.Lock()
	defer l.mu.Unlock()

	timestamp := time.Now().UTC().Format(time.RFC3339Nano)
	msg := fmt.Sprintf(format, args...)
	l.logger.Printf("%s [%-5s] %s", timestamp, level.String(), msg)
}

// Debug logs a debug message.
func (l *Logger) Debug(format string, args ...interface{}) {
	l.log(LevelDebug, format, args...)
}

// Info logs an info message.
func (l *Logger) Info(format string, args ...interface{}) {
	l.log(LevelInfo, format, args...)
}

// Warn logs a warning message.
func (l *Logger) Warn(format string, args ...interface{}) {
	l.log(LevelWarn, format, args...)
}

// Error logs an error message.
func (l *Logger) Error(format string, args ...interface{}) {
	l.log(LevelError, format, args...)
}

// Close closes the log file.
func (l *Logger) Close() {
	l.mu.Lock()
	defer l.mu.Unlock()
	if l.file != nil {
		l.file.Close()
	}
}
