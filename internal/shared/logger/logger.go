package logger

import (
	"context"
	"io"
	"log/slog"
	"os"

	"go.opentelemetry.io/otel/trace"
)

type contextKey string

const ReqIDKey contextKey = "request_id"

var log *slog.Logger

func Init(env string) {
	opts := &slog.HandlerOptions{Level: slog.LevelDebug}

	if env == "production" {
		opts.Level = slog.LevelInfo
	}

	os.MkdirAll("logs", 0755)
	logFile, err := os.OpenFile("logs/app.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		logFile = nil
	}

	var writer io.Writer
	if logFile != nil {
		writer = io.MultiWriter(os.Stdout, logFile)
	} else {
		writer = os.Stdout
	}

	handler := slog.NewJSONHandler(writer, opts)

	log = slog.New(handler)
	slog.SetDefault(log)
}

func getTraceAttrs(ctx context.Context) []any {
	var attrs []any

	if id, ok := ctx.Value(ReqIDKey).(string); ok && id != "" {
		attrs = append(attrs, slog.String("req_id", id))
	} else {
		attrs = append(attrs, slog.String("req_id", "N/A"))
	}

	span := trace.SpanFromContext(ctx)
	if span.SpanContext().IsValid() {
		attrs = append(attrs,
			slog.String("trace_id", span.SpanContext().TraceID().String()),
			slog.String("span_id", span.SpanContext().SpanID().String()),
		)
	}

	return attrs
}

func Info(ctx context.Context, msg string, args ...any) {
	log.Info(msg, append(getTraceAttrs(ctx), args...)...)
}

func Error(ctx context.Context, msg string, args ...any) {
	log.Error(msg, append(getTraceAttrs(ctx), args...)...)
}

func Warn(ctx context.Context, msg string, args ...any) {
	log.Warn(msg, append(getTraceAttrs(ctx), args...)...)
}

func Debug(ctx context.Context, msg string, args ...any) {
	log.Debug(msg, append(getTraceAttrs(ctx), args...)...)
}
