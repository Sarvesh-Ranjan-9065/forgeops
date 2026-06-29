// Package log configures structured logging for ForgeOps using log/slog.
package log

import (
	"log/slog"
	"os"
	"strings"
)

// New returns a slog.Logger that writes JSON to standard error at the given
// level. Recognized levels are "debug", "info", "warn", and "error"; any other
// value defaults to "info".
func New(level string) *slog.Logger {
	var lvl slog.Level
	switch strings.ToLower(level) {
	case "debug":
		lvl = slog.LevelDebug
	case "warn":
		lvl = slog.LevelWarn
	case "error":
		lvl = slog.LevelError
	default:
		lvl = slog.LevelInfo
	}
	handler := slog.NewJSONHandler(os.Stderr, &slog.HandlerOptions{Level: lvl})
	return slog.New(handler)
}
