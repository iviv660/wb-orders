package logger

import (
	"context"

	"go.uber.org/zap"
)

type NoopLogger struct{}

func (n *NoopLogger) Info(ctx context.Context, msg string, fields ...zap.Field)  {}
func (n *NoopLogger) Error(ctx context.Context, msg string, fields ...zap.Field) {}
