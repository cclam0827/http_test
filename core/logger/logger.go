// Package logger configures the process-wide log/slog logger.
package logger

import (
	"io"
	"log/slog"
	"os"
	"strings"
)

// NewHandler returns a slog.Handler that writes to w in the given format:
//   - "json":  structured JSON (production default)
//   - "text":  key=value text
//   - "color": colorized single-line text for local development
//
// Unknown formats fall back to JSON.
func NewHandler(w io.Writer, format string, level slog.Level) slog.Handler {
	switch format {
	case "text":
		return slog.NewTextHandler(w, &slog.HandlerOptions{Level: level})
	case "color":
		return newColorHandler(w, level)
	default:
		return slog.NewJSONHandler(w, &slog.HandlerOptions{Level: level})
	}
}

// ParseLevel maps a LOG_LEVEL value to a slog level. Anything but "debug" is info.
func ParseLevel(s string) slog.Level {
	if strings.EqualFold(s, "debug") {
		return slog.LevelDebug
	}
	return slog.LevelInfo
}

// Setup installs the default logger (which also receives output from the
// standard log package) and returns it. An empty format means json when env is
// "production" and text otherwise.
func Setup(format, level, env string) *slog.Logger {
	if format == "" {
		format = "text"
		if env == "production" {
			format = "json"
		}
	}
	l := slog.New(NewHandler(os.Stdout, format, ParseLevel(level)))
	slog.SetDefault(l)
	return l
}
