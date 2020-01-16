package common

import (
	"fmt"
	"strings"

	"github.com/rs/zerolog"
)

// LeveledLogger interface for level logger compatible with
// go-retryablehttp lib
type LeveledLogger interface {
	Error(string, ...interface{})
	Info(string, ...interface{})
	Debug(string, ...interface{})
	Warn(string, ...interface{})
}

// RetryableHTTPLogger wrapper around zero logger compatible with
// LeveledLogger interface implemented by go-retryablehttp lib
type RetryableHTTPLogger struct {
	logger zerolog.Logger
}

// NewRetryableHTTPLogger create a new RetryableHTTPLogger logger
func NewRetryableHTTPLogger(logger zerolog.Logger) LeveledLogger {
	return RetryableHTTPLogger{
		logger: logger,
	}
}

func (l RetryableHTTPLogger) join(sep string, values ...interface{}) string {
	strs := make([]string, len(values))
	for i, v := range values {
		strs[i] = fmt.Sprintf("%s", v)
	}
	return strings.Join(strs, sep)
}

// Error print error level message
func (l RetryableHTTPLogger) Error(msg string, args ...interface{}) {
	l.logger.Error().Msg(fmt.Sprintf("%s %s", msg, l.join(" ", args...)))
}

// Warn print warn level message
func (l RetryableHTTPLogger) Warn(msg string, args ...interface{}) {
	l.logger.Warn().Msg(fmt.Sprintf("%s %s", msg, l.join(" ", args...)))
}

// Debug print debug level message
func (l RetryableHTTPLogger) Debug(msg string, args ...interface{}) {
	l.logger.Debug().Msg(fmt.Sprintf("%s %s", msg, l.join(" ", args...)))
}

// Info print info level message
func (l RetryableHTTPLogger) Info(msg string, args ...interface{}) {
	l.logger.Info().Msg(fmt.Sprintf("%s %s", msg, l.join(" ", args...)))
}
