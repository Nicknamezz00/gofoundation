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

var bufPool = sync.Pool{
	New: func() interface{} {
		b := make([]byte, 0, 1024)
		return &b
	},
}

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

	bp := bufPool.Get().(*[]byte)
	buf := (*bp)[:0]

	// Build JSON directly in fixed order — no map, no intermediate structs
	buf = append(buf, `{"level":"`...)
	buf = append(buf, level.String()...)
	buf = append(buf, `","message":`...)
	buf = appendJSONString(buf, msg)
	buf = append(buf, `,"timestamp":"`...)
	buf = time.Now().AppendFormat(buf, time.RFC3339Nano)
	buf = append(buf, `","caller":"`...)
	buf = append(buf, l.getCaller()...)
	buf = append(buf, '"')

	if err != nil {
		buf = append(buf, `,"error":`...)
		buf = appendJSONString(buf, err.Error())
	}

	for _, f := range l.fields {
		buf = appendField(buf, f)
	}
	for _, f := range fields {
		buf = appendField(buf, f)
	}

	buf = append(buf, "}\n"...)

	if _, writeErr := l.writer.Write(buf); writeErr != nil {
		fmt.Fprintf(os.Stderr, "failed to write log entry: %v\n", writeErr)
	}

	*bp = buf
	bufPool.Put(bp)
}

func appendField(buf []byte, f Field) []byte {
	buf = append(buf, ',', '"')
	buf = append(buf, f.Key...)
	buf = append(buf, '"', ':')
	val, _ := json.Marshal(f.Value)
	buf = append(buf, val...)
	return buf
}

func appendJSONString(buf []byte, s string) []byte {
	buf = append(buf, '"')
	for i := 0; i < len(s); i++ {
		c := s[i]
		switch c {
		case '"':
			buf = append(buf, '\\', '"')
		case '\\':
			buf = append(buf, '\\', '\\')
		case '\n':
			buf = append(buf, '\\', 'n')
		case '\r':
			buf = append(buf, '\\', 'r')
		case '\t':
			buf = append(buf, '\\', 't')
		default:
			if c < 0x20 {
				buf = append(buf, '\\', 'u', '0', '0', "0123456789abcdef"[c>>4], "0123456789abcdef"[c&0xf])
			} else {
				buf = append(buf, c)
			}
		}
	}
	buf = append(buf, '"')
	return buf
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
