package option

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

func (r *Repository) CreateOption(ctx context.Context, params dbsqlc.CreateOptionParams) (dbsqlc.Option, error) {
	return r.queries.CreateOption(ctx, params)
}

func (r *Repository) GetOptionByID(ctx context.Context, id string) (dbsqlc.Option, error) {
	uid, err := utils.StrToPgUUID(id)
	if err != nil {
		return dbsqlc.Option{}, err
	}
	return r.queries.GetOptionByID(ctx, uid)
}

func (r *Repository) ListOptionsByRoom(ctx context.Context, roomID string) ([]dbsqlc.Option, error) {
	uid, err := utils.StrToPgUUID(roomID)
	if err != nil {
		return nil, err
	}
	return r.queries.ListOptionsByRoom(ctx, uid)
}

func (r *Repository) UpdateOption(ctx context.Context, params dbsqlc.UpdateOptionParams) (dbsqlc.Option, error) {
	return r.queries.UpdateOption(ctx, params)
}

func (r *Repository) DeleteOption(ctx context.Context, id string) error {
	uid, err := utils.StrToPgUUID(id)
	if err != nil {
		return err
	}
	return r.queries.DeleteOption(ctx, uid)
}

func (r *Repository) GetRoomByID(ctx context.Context, id string) (dbsqlc.Room, error) {
	uid, err := utils.StrToPgUUID(id)
	if err != nil {
		return dbsqlc.Room{}, err
	}
	return r.queries.GetRoomByID(ctx, uid)
}

func (r *Repository) GetMediaByID(ctx context.Context, id string) (dbsqlc.Medium, error) {
	uid, err := utils.StrToPgUUID(id)
	if err != nil {
		return dbsqlc.Medium{}, err
	}
	return r.queries.GetMediaByID(ctx, uid)
}
