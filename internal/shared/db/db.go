package db

import (
	"context"
	"fmt"

	"github.com/alberdjuniawan/votesystem/internal/config"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	sdktrace "go.opentelemetry.io/otel/trace"
)

type queryTracer struct {
	tracer sdktrace.Tracer
}

func (t *queryTracer) TraceQueryStart(ctx context.Context, conn *pgx.Conn, data pgx.TraceQueryStartData) context.Context {
	_, span := t.tracer.Start(ctx, "db.query",
		sdktrace.WithAttributes(
			attribute.String("db.system", "postgresql"),
			attribute.String("db.statement", truncateSQL(data.SQL)),
		),
	)
	return sdktrace.ContextWithSpan(ctx, span)
}

func (t *queryTracer) TraceQueryEnd(ctx context.Context, conn *pgx.Conn, data pgx.TraceQueryEndData) {
	span := sdktrace.SpanFromContext(ctx)
	if data.Err != nil {
		span.RecordError(data.Err)
		span.SetStatus(codes.Error, data.Err.Error())
	}
	span.End()
}

func (t *queryTracer) TraceBatchStart(ctx context.Context, conn *pgx.Conn, data pgx.TraceBatchStartData) context.Context {
	return ctx
}

func (t *queryTracer) TraceBatchEnd(ctx context.Context, conn *pgx.Conn, data pgx.TraceBatchEndData) {}

func (t *queryTracer) TraceCopyFromStart(ctx context.Context, conn *pgx.Conn, data pgx.TraceCopyFromStartData) context.Context {
	return ctx
}

func (t *queryTracer) TraceCopyFromEnd(ctx context.Context, conn *pgx.Conn, data pgx.TraceCopyFromEndData) {}

func (t *queryTracer) TracePrepareStart(ctx context.Context, conn *pgx.Conn, data pgx.TracePrepareStartData) context.Context {
	return ctx
}

func (t *queryTracer) TracePrepareEnd(ctx context.Context, conn *pgx.Conn, data pgx.TracePrepareEndData) {}

func (t *queryTracer) TraceConnectStart(ctx context.Context, data pgx.TraceConnectStartData) context.Context {
	return ctx
}

func (t *queryTracer) TraceConnectEnd(ctx context.Context, data pgx.TraceConnectEndData) {}

func truncateSQL(sql string) string {
	if len(sql) > 200 {
		return sql[:200] + "..."
	}
	return sql
}

func NewPool(ctx context.Context, cfg config.DatabaseConfig) (*pgxpool.Pool, error) {
	poolConfig, err := pgxpool.ParseConfig(cfg.DSN())
	if err != nil {
		return nil, fmt.Errorf("failed to parse db config: %w", err)
	}

	poolConfig.MaxConns = cfg.MaxConns
	poolConfig.MinConns = cfg.MinConns

	poolConfig.ConnConfig.Tracer = &queryTracer{
		tracer: otel.Tracer("github.com/alberdjuniawan/votesystem/internal/shared/db"),
	}

	pool, err := pgxpool.NewWithConfig(ctx, poolConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create db pool: %w", err)
	}

	if err := pool.Ping(ctx); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	return pool, nil
}


