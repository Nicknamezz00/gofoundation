package logger

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"sync"
	"time"

	jsoniter "github.com/json-iterator/go"
	"github.com/Nicknamezz00/gofoundation/trace"
	"gopkg.in/natefinch/lumberjack.v2"
)

var json = jsoniter.ConfigCompatibleWithStandardLibrary

type logger struct {
	config Config
	writer io.Writer
	fields []Field
	mu     *sync.Mutex
}

// New creates a new logger
func New(config Config) (Logger, error) {
	l := &logger{
		config: config,
		fields: make([]Field, 0),
		mu:     &sync.Mutex{},
	}

	// Setup writer
	if config.DevMode {
		l.writer = os.Stdout
	} else {
		if config.FilePath == "" {
			return nil, fmt.Errorf("file path required for non-dev mode")
		}

		// Create directory if not exists
		dir := filepath.Dir(config.FilePath)
		if err := os.MkdirAll(dir, 0755); err != nil {
			return nil, fmt.Errorf("failed to create log directory: %w", err)
		}

		l.writer = &lumberjack.Logger{
			Filename:   config.FilePath,
			MaxSize:    config.MaxSize,
			MaxAge:     config.MaxAge,
			MaxBackups: config.MaxBackups,
			Compress:   config.Compress,
		}
	}

	return l, nil
}

func (l *logger) Debug(msg string, fields ...Field) {
	if l.config.Level > DebugLevel {
		return
	}
	l.log(DebugLevel, msg, nil, fields...)
}

func (l *logger) Info(msg string, fields ...Field) {
	if l.config.Level > InfoLevel {
		return
	}
	l.log(InfoLevel, msg, nil, fields...)
}

func (l *logger) Warn(msg string, fields ...Field) {
	if l.config.Level > WarnLevel {
		return
	}
	l.log(WarnLevel, msg, nil, fields...)
}

func (l *logger) Error(msg string, fields ...Field) {
	if l.config.Level > ErrorLevel {
		return
	}
	l.log(ErrorLevel, msg, nil, fields...)
}

func (l *logger) Fatal(msg string, fields ...Field) {
	l.log(FatalLevel, msg, nil, fields...)
	// Note: Fatal logs at fatal level but does NOT call os.Exit()
	// This is a library - the application should decide when to exit
}

func (l *logger) With(fields ...Field) Logger {
	newFields := make([]Field, len(l.fields)+len(fields))
	copy(newFields, l.fields)
	copy(newFields[len(l.fields):], fields)

	return &logger{
		config: l.config,
		writer: l.writer,
		fields: newFields,
		mu:     l.mu, // Share the mutex pointer
	}
}

func (l *logger) WithContext(ctx context.Context) Logger {
	traceInfo := trace.GetTraceInfo(ctx)
	if traceInfo == nil {
		return l
	}

	return l.With(
		Field{Key: "trace_id", Value: traceInfo.TraceID},
		Field{Key: "span_id", Value: traceInfo.SpanID},
	)
}

func (l *logger) log(level Level, msg string, err error, fields ...Field) {
	l.mu.Lock()
	defer l.mu.Unlock()

	// Build ordered map for JSON output
	entry := make(map[string]interface{})

	// Ordered keys: level, trace_id, span_id, error, message, caller, timestamp, ...fields
	entry["level"] = level.String()
	entry["message"] = msg
	entry["timestamp"] = time.Now().Format(time.RFC3339Nano)

	// Add caller info
	entry["caller"] = l.getCaller()

	// Add error if present
	if err != nil {
		entry["error"] = err.Error()
	}

	// Add base fields (including trace_id and span_id as top-level)
	for _, f := range l.fields {
		entry[f.Key] = f.Value
	}

	// Add additional fields
	for _, f := range fields {
		entry[f.Key] = f.Value
	}

	// Marshal to JSON
	data, err := json.Marshal(entry)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to marshal log entry: %v\n", err)
		return
	}

	// Write to output
	if _, err := l.writer.Write(data); err != nil {
		fmt.Fprintf(os.Stderr, "failed to write log entry: %v\n", err)
		return
	}
	if _, err := l.writer.Write([]byte("\n")); err != nil {
		fmt.Fprintf(os.Stderr, "failed to write log newline: %v\n", err)
	}
}

func (l *logger) getCaller() string {
	// Skip 3 frames: getCaller, log, and the public method (Debug/Info/etc)
	_, file, line, ok := runtime.Caller(3)
	if !ok {
		return "unknown"
	}

	// Get relative path from repository root
	// Try to find go.mod to determine repo root
	dir := filepath.Dir(file)
	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			relPath, err := filepath.Rel(dir, file)
			if err == nil {
				return fmt.Sprintf("%s:%d", relPath, line)
			}
			break
		}

		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}
		dir = parent
	}

	// Fallback to just filename
	return fmt.Sprintf("%s:%d", filepath.Base(file), line)
}

// Helper functions for creating fields
func String(key, value string) Field {
	return Field{Key: key, Value: value}
}

func Int(key string, value int) Field {
	return Field{Key: key, Value: value}
}

func Int64(key string, value int64) Field {
	return Field{Key: key, Value: value}
}

func Float64(key string, value float64) Field {
	return Field{Key: key, Value: value}
}

func Bool(key string, value bool) Field {
	return Field{Key: key, Value: value}
}

func Any(key string, value interface{}) Field {
	return Field{Key: key, Value: value}
}

func Err(err error) Field {
	if err == nil {
		return Field{Key: "error", Value: nil}
	}
	return Field{Key: "error", Value: err.Error()}
}
