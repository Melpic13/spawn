package observability

import (
	"fmt"

	"go.uber.org/zap"
)

// Logger wraps zap logger.
type Logger struct {
	*zap.Logger
}

// NewLogger returns structured logger.
func NewLogger(level string) (*Logger, error) {
	cfg := zap.NewProductionConfig()
	if err := cfg.Level.UnmarshalText([]byte(level)); err != nil {
		return nil, fmt.Errorf("new logger: %w", err)
	}
	l, err := cfg.Build()
	if err != nil {
		return nil, fmt.Errorf("new logger: %w", err)
	}
	return &Logger{Logger: l}, nil
}
