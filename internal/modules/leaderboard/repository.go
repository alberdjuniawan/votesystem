package leaderboard

import (
	"context"

	dbsqlc "github.com/alberdjuniawan/votesystem/internal/db/sqlc"
	"github.com/alberdjuniawan/votesystem/internal/shared/utils"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Repository struct {
	queries *dbsqlc.Queries
	db      *pgxpool.Pool
}

func NewRepository(queries *dbsqlc.Queries, db *pgxpool.Pool) *Repository {
	return &Repository{queries: queries, db: db}
}

func (r *Repository) ListOptionsByRoom(ctx context.Context, roomID string) ([]dbsqlc.Option, error) {
	uid, err := utils.StrToPgUUID(roomID)
	if err != nil {
		return nil, err
	}
	return r.queries.ListOptionsByRoom(ctx, uid)
}
