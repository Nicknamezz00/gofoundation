package logger

import (
	"context"
	"time"
)

// Level represents log level
type Level int8

const (
	DebugLevel Level = iota - 1
	InfoLevel
	WarnLevel
	ErrorLevel
	FatalLevel
)

// String returns string representation of level
func (l Level) String() string {
	switch l {
	case DebugLevel:
		return "debug"
	case InfoLevel:
		return "info"
	case WarnLevel:
		return "warn"
	case ErrorLevel:
		return "error"
	case FatalLevel:
		return "fatal"
	default:
		return "unknown"
	}
}

// Field represents a log field
type Field struct {
	Key   string
	Value interface{}
}

// Logger is the interface for structured logging
type Logger interface {
	Debug(msg string, fields ...Field)
	Info(msg string, fields ...Field)
	Warn(msg string, fields ...Field)
	Error(msg string, fields ...Field)
	Fatal(msg string, fields ...Field)
	With(fields ...Field) Logger
	WithContext(ctx context.Context) Logger
}

// Config holds logger configuration
type Config struct {
	Level          Level
	DevMode        bool
	FilePath       string
	MaxSize        int           // MB
	MaxAge         int           // days
	MaxBackups     int
	Compress       bool
	RotateOnTime   bool
	RotateInterval time.Duration
}

// DefaultConfig returns default logger configuration
func DefaultConfig() Config {
	return Config{
		Level:          InfoLevel,
		DevMode:        true,
		MaxSize:        100,
		MaxAge:         7,
		MaxBackups:     3,
		Compress:       true,
		RotateOnTime:   false,
		RotateInterval: 24 * time.Hour,
	}
}
