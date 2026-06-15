package middleware

import (
	"log/slog"
	"strconv"
	"sync"
	"time"

	"github.com/alberdjuniawan/votesystem/internal/shared/metrics"
	"github.com/gin-gonic/gin"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
)

var (
	metricsOnce         sync.Once
	httpRequestsTotal   metric.Int64Counter
	httpRequestDuration metric.Float64Histogram
)

func initHTTPMetrics() {
	meter := otel.Meter("votesystem/http")

	counter, err := meter.Int64Counter("http.requests",
		metric.WithDescription("Total number of HTTP requests"),
	)
	if err != nil {
		slog.Error("failed to create http.requests counter", "error", err)
		return
	}

	histogram, err := meter.Float64Histogram("http.request.duration.seconds",
		metric.WithDescription("Duration of HTTP requests in seconds"),
		metric.WithUnit("s"),
	)
	if err != nil {
		slog.Error("failed to create http.request.duration histogram", "error", err)
		return
	}

	httpRequestsTotal = counter
	httpRequestDuration = histogram
}

func Metrics() gin.HandlerFunc {
	metricsOnce.Do(initHTTPMetrics)
	return func(c *gin.Context) {
		metrics.ActiveConnections.Add(c.Request.Context(), 1)
		defer metrics.ActiveConnections.Add(c.Request.Context(), -1)

		start := time.Now()
		c.Next()

		status := c.Writer.Status()
		route := c.FullPath()
		if route == "" {
			route = "unknown"
		}

		attrs := []attribute.KeyValue{
			attribute.String("http.method", c.Request.Method),
			attribute.String("http.route", route),
			attribute.String("http.status_code", strconv.Itoa(status)),
		}

		if httpRequestsTotal != nil {
			httpRequestsTotal.Add(c.Request.Context(), 1, metric.WithAttributes(attrs...))
		}
		if httpRequestDuration != nil {
			httpRequestDuration.Record(c.Request.Context(), time.Since(start).Seconds(), metric.WithAttributes(attrs...))
		}
	}
}
