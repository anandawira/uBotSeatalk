package log

import (
	"context"

	"go.uber.org/zap"

	"github.com/anandawira/uBotSeatalk/internal/pkg/contextid"
)

var logger *zap.SugaredLogger

func init() {
	l, err := zap.NewProduction()
	if err != nil {
		panic(err)
	}
	logger = l.Sugar()
}

func InfoCtx(ctx context.Context, msg string, kv ...interface{}) {
	contextID := contextid.Value(ctx)
	if contextID != "" {
		kv = append(kv, "context_id", contextID)
	}

	logger.Infow(msg, kv...)
}

func Info(msg string, kv ...interface{}) {
	logger.Infow(msg, kv...)
}

func ErrorCtx(ctx context.Context, msg string, kv ...interface{}) {
	contextID := contextid.Value(ctx)
	if contextID != "" {
		kv = append(kv, "context_id", contextID)
	}

	logger.Errorw(msg, kv...)
}

func Error(msg string, kv ...interface{}) {
	logger.Errorw(msg, kv...)
}
