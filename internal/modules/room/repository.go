package room

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

func (r *Repository) CreateRoom(ctx context.Context, params dbsqlc.CreateRoomParams) (dbsqlc.Room, error) {
	return r.queries.CreateRoom(ctx, params)
}

func (r *Repository) GetRoomByID(ctx context.Context, id string) (dbsqlc.Room, error) {
	uid, err := utils.StrToPgUUID(id)
	if err != nil {
		return dbsqlc.Room{}, err
	}
	return r.queries.GetRoomByID(ctx, uid)
}

func (r *Repository) GetRoomByShareCode(ctx context.Context, code string) (dbsqlc.Room, error) {
	return r.queries.GetRoomByShareCode(ctx, code)
}

func (r *Repository) ListRoomsByOwner(ctx context.Context, ownerID string) ([]dbsqlc.Room, error) {
	uid, err := utils.StrToPgUUID(ownerID)
	if err != nil {
		return nil, err
	}
	return r.queries.ListRoomsByOwner(ctx, uid)
}

func (r *Repository) UpdateRoomStatus(ctx context.Context, id string, status dbsqlc.RoomStatus) (dbsqlc.Room, error) {
	uid, err := utils.StrToPgUUID(id)
	if err != nil {
		return dbsqlc.Room{}, err
	}
	return r.queries.UpdateRoomStatus(ctx, dbsqlc.UpdateRoomStatusParams{
		ID:     uid,
		Status: status,
	})
}

func (r *Repository) DeleteRoom(ctx context.Context, id, ownerID string) error {
	roomUID, err := utils.StrToPgUUID(id)
	if err != nil {
		return err
	}
	ownerUID, err := utils.StrToPgUUID(ownerID)
	if err != nil {
		return err
	}
	return r.queries.DeleteRoom(ctx, dbsqlc.DeleteRoomParams{
		ID:      roomUID,
		OwnerID: ownerUID,
	})
}
