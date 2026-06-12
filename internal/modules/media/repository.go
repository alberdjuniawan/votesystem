package media

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

func (r *Repository) CreateMedia(ctx context.Context, params dbsqlc.CreateMediaParams) (dbsqlc.Medium, error) {
	return r.queries.CreateMedia(ctx, params)
}

func (r *Repository) GetMediaByID(ctx context.Context, id string) (dbsqlc.Medium, error) {
	uid, err := utils.StrToPgUUID(id)
	if err != nil {
		return dbsqlc.Medium{}, err
	}
	return r.queries.GetMediaByID(ctx, uid)
}

func (r *Repository) DeleteMedia(ctx context.Context, id string) error {
	uid, err := utils.StrToPgUUID(id)
	if err != nil {
		return err
	}
	return r.queries.DeleteMedia(ctx, uid)
}
