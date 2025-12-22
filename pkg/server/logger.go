package server

import (
	"log/slog"

	"github.com/NVIDIA/cloud-native-stack/pkg/logging"
)

// Logger interface for structured logging
type Logger interface {
	Info(msg string, fields map[string]interface{})
	Error(msg string, err error, fields map[string]interface{})
	Debug(msg string, fields map[string]interface{})
}

// SlogAdapter adapts slog.Logger to the Logger interface
type SlogAdapter struct {
	logger *slog.Logger
}

// NewLogger creates a new logger instance using the structured logger from pkg/logging
func NewLogger(level slog.Level) Logger {
	return &SlogAdapter{
		logger: logging.New(level),
	}
}

// Info logs informational messages
func (l *SlogAdapter) Info(msg string, fields map[string]interface{}) {
	args := make([]any, 0, len(fields)*2)
	for k, v := range fields {
		args = append(args, k, v)
	}
	l.logger.Info(msg, args...)
}

// Error logs error messages
func (l *SlogAdapter) Error(msg string, err error, fields map[string]interface{}) {
	args := make([]any, 0, (len(fields)+1)*2)
	if err != nil {
		args = append(args, "error", err.Error())
	}
	for k, v := range fields {
		args = append(args, k, v)
	}
	l.logger.Error(msg, args...)
}

// Debug logs debug messages
func (l *SlogAdapter) Debug(msg string, fields map[string]interface{}) {
	args := make([]any, 0, len(fields)*2)
	for k, v := range fields {
		args = append(args, k, v)
	}
	l.logger.Debug(msg, args...)
}
