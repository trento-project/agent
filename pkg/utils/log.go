package utils

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"os"
)

func NewDefaultLogger(logLevel string) *slog.Logger {
	return slog.New(NewDefaultTextHandler(
		os.Stdout,
		parseLogLevel(logLevel),
	))
}

func NewStderrLogger(logLevel string) *slog.Logger {
	return slog.New(NewDefaultTextHandler(
		os.Stderr,
		parseLogLevel(logLevel),
	))
}

type DefaultTextHandler struct {
	w     io.Writer
	level slog.Level
}

func NewDefaultTextHandler(w io.Writer, level slog.Level) *DefaultTextHandler {
	return &DefaultTextHandler{w: w, level: level}
}

func (h *DefaultTextHandler) Enabled(_ context.Context, level slog.Level) bool {
	return level >= h.level
}

func (h *DefaultTextHandler) Handle(_ context.Context, r slog.Record) error {
	// Format time as YYYY-MM-DD hh:mm:ss
	timeStr := r.Time.Format("2006-01-02 15:04:05")

	// Map slog.Level to uppercase string (WARNING for WARN, etc.)
	levelStr := r.Level.String()

	// Start building the log line
	line := fmt.Sprintf("%s %s %s", timeStr, levelStr, r.Message)

	// Append all key-value attributes
	r.Attrs(func(attr slog.Attr) bool {
		line += fmt.Sprintf(" %s=%v", attr.Key, attr.Value.Any())
		return true
	})

	// Write the line
	_, err := fmt.Fprintln(h.w, line)
	return err
}

func (h *DefaultTextHandler) WithAttrs(_ []slog.Attr) slog.Handler {
	// do nothing, all attributes are processed in Handle
	return h
}

func (h *DefaultTextHandler) WithGroup(_ string) slog.Handler {
	// TODO: handle group
	return h
}

// converts a string representation of a log level to slog.Level.
// Supported values are "debug", "info", "warn", "error".
func parseLogLevel(logLevel string) slog.Level {
	switch logLevel {
	case "debug":
		return slog.LevelDebug
	case "info":
		return slog.LevelInfo
	case "warn":
		return slog.LevelWarn
	case "error":
		return slog.LevelError
	default:
		return slog.LevelInfo
	}
}
