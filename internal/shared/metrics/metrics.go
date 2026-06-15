package metrics

import (
	"context"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
)

var (
	meter = otel.Meter("votesystem")

	ActiveConnections metric.Int64UpDownCounter
	WSConnections     metric.Int64UpDownCounter

	RoomsCreated metric.Int64Counter
	VotesCast    metric.Int64Counter

	DBPoolActive metric.Int64ObservableGauge
	DBPoolIdle   metric.Int64ObservableGauge
	DBPoolMax    metric.Int64ObservableGauge

	UptimeSeconds metric.Int64ObservableGauge
	startTime     = time.Now()
)

func Init(poolStat func() (active, idle, max int64)) error {
	var err error

	ActiveConnections, err = meter.Int64UpDownCounter("http.active_connections",
		metric.WithDescription("Current number of active HTTP requests"),
	)
	if err != nil {
		return err
	}

	WSConnections, err = meter.Int64UpDownCounter("ws.connections",
		metric.WithDescription("Current number of active WebSocket connections"),
	)
	if err != nil {
		return err
	}

	RoomsCreated, err = meter.Int64Counter("votesystem.rooms.created",
		metric.WithDescription("Total number of rooms created"),
	)
	if err != nil {
		return err
	}

	VotesCast, err = meter.Int64Counter("votesystem.votes.cast",
		metric.WithDescription("Total number of votes cast"),
	)
	if err != nil {
		return err
	}

	DBPoolActive, err = meter.Int64ObservableGauge("db.pool.active",
		metric.WithDescription("Current number of active (acquired) DB connections from the pool"),
		metric.WithInt64Callback(func(ctx context.Context, obs metric.Int64Observer) error {
			if poolStat != nil {
				a, _, _ := poolStat()
				obs.Observe(a)
			}
			return nil
		}),
	)
	if err != nil {
		return err
	}

	DBPoolIdle, err = meter.Int64ObservableGauge("db.pool.idle",
		metric.WithDescription("Current number of idle DB connections in the pool"),
		metric.WithInt64Callback(func(ctx context.Context, obs metric.Int64Observer) error {
			if poolStat != nil {
				_, i, _ := poolStat()
				obs.Observe(i)
			}
			return nil
		}),
	)
	if err != nil {
		return err
	}

	DBPoolMax, err = meter.Int64ObservableGauge("db.pool.max",
		metric.WithDescription("Maximum number of DB connections the pool allows"),
		metric.WithInt64Callback(func(ctx context.Context, obs metric.Int64Observer) error {
			if poolStat != nil {
				_, _, m := poolStat()
				obs.Observe(m)
			}
			return nil
		}),
	)
	if err != nil {
		return err
	}

	UptimeSeconds, err = meter.Int64ObservableGauge("app.uptime_seconds",
		metric.WithDescription("Time since application started in seconds"),
		metric.WithInt64Callback(func(ctx context.Context, obs metric.Int64Observer) error {
			obs.Observe(int64(time.Since(startTime).Seconds()))
			return nil
		}),
	)
	if err != nil {
		return err
	}

	return nil
}

func RecordRoomCreated(ctx context.Context, ownerID string) {
	RoomsCreated.Add(ctx, 1, metric.WithAttributes(
		attribute.String("owner_id", ownerID),
	))
}

func RecordVoteCast(ctx context.Context, roomID, optionID string) {
	VotesCast.Add(ctx, 1, metric.WithAttributes(
		attribute.String("room_id", roomID),
		attribute.String("option_id", optionID),
	))
}

func RecordWSConnect() {
	WSConnections.Add(context.Background(), 1)
}

func RecordWSDisconnect() {
	WSConnections.Add(context.Background(), -1)
}
