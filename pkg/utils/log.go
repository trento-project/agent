package utils

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"os"
	"strings"
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
	w      io.Writer
	level  slog.Level
	attrs  []slog.Attr
	groups []string
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
		line += formatAttr(attr, h.groups)
		return true
	})

	// Append any default attributes
	for _, attr := range h.attrs {
		line += formatAttr(attr, []string{})
	}

	// Write the line
	_, err := fmt.Fprintln(h.w, line)
	return err
}

func (h *DefaultTextHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	// prepend to every attribute key the current group prefix
	withPrefix := make([]slog.Attr, len(attrs))
	for i, attr := range attrs {
		withPrefix[i] = slog.Attr{
			Key:   formatAttrKey(attr.Key, h.groups),
			Value: attr.Value,
		}
	}

	// Combine existing attributes with new ones
	newAttrs := make([]slog.Attr, len(h.attrs)+len(withPrefix))
	copy(newAttrs, h.attrs)
	copy(newAttrs[len(h.attrs):], withPrefix)

	return &DefaultTextHandler{
		w:      h.w,
		level:  h.level,
		attrs:  newAttrs,
		groups: h.groups,
	}
}

func (h *DefaultTextHandler) WithGroup(group string) slog.Handler {
	newGroups := make([]string, len(h.groups)+1)
	copy(newGroups, h.groups)
	newGroups[len(h.groups)] = group

	return &DefaultTextHandler{
		w:      h.w,
		level:  h.level,
		attrs:  h.attrs,
		groups: newGroups,
	}
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

func formatAttr(attr slog.Attr, groups []string) string {
	return fmt.Sprintf(
		" %s=%v",
		formatAttrKey(attr.Key, groups),
		attr.Value.Any(),
	)
}

func formatAttrKey(key string, groups []string) string {
	if len(groups) == 0 {
		return key
	}
	return strings.Join(groups, ".") + "." + key
}
