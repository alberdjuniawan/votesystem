package vote

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

func (r *Repository) GetRoomByID(ctx context.Context, id string) (dbsqlc.Room, error) {
	uid, err := utils.StrToPgUUID(id)
	if err != nil {
		return dbsqlc.Room{}, err
	}
	return r.queries.GetRoomByID(ctx, uid)
}

func (r *Repository) GetOptionByID(ctx context.Context, id string) (dbsqlc.Option, error) {
	uid, err := utils.StrToPgUUID(id)
	if err != nil {
		return dbsqlc.Option{}, err
	}
	return r.queries.GetOptionByID(ctx, uid)
}

func (r *Repository) CreateVote(ctx context.Context, roomID, userID, optionID string) (dbsqlc.Vote, error) {
	roomUID, err := utils.StrToPgUUID(roomID)
	if err != nil {
		return dbsqlc.Vote{}, err
	}
	userUID, err := utils.StrToPgUUID(userID)
	if err != nil {
		return dbsqlc.Vote{}, err
	}
	optUID, err := utils.StrToPgUUID(optionID)
	if err != nil {
		return dbsqlc.Vote{}, err
	}
	return r.queries.CreateVote(ctx, dbsqlc.CreateVoteParams{
		RoomID:   roomUID,
		UserID:   userUID,
		OptionID: optUID,
	})
}

func (r *Repository) GetVoteByRoomUserOption(ctx context.Context, roomID, userID, optionID string) (dbsqlc.Vote, error) {
	roomUID, err := utils.StrToPgUUID(roomID)
	if err != nil {
		return dbsqlc.Vote{}, err
	}
	userUID, err := utils.StrToPgUUID(userID)
	if err != nil {
		return dbsqlc.Vote{}, err
	}
	optUID, err := utils.StrToPgUUID(optionID)
	if err != nil {
		return dbsqlc.Vote{}, err
	}
	return r.queries.GetVoteByRoomUserOption(ctx, dbsqlc.GetVoteByRoomUserOptionParams{
		RoomID:   roomUID,
		UserID:   userUID,
		OptionID: optUID,
	})
}

func (r *Repository) GetVoteCountByRoomAndUser(ctx context.Context, roomID, userID string) (int64, error) {
	roomUID, err := utils.StrToPgUUID(roomID)
	if err != nil {
		return 0, err
	}
	userUID, err := utils.StrToPgUUID(userID)
	if err != nil {
		return 0, err
	}
	return r.queries.GetVoteCountByRoomAndUser(ctx, dbsqlc.GetVoteCountByRoomAndUserParams{
		RoomID: roomUID,
		UserID: userUID,
	})
}

func (r *Repository) GetVotesByRoomAndUser(ctx context.Context, roomID, userID string) ([]dbsqlc.Vote, error) {
	roomUID, err := utils.StrToPgUUID(roomID)
	if err != nil {
		return nil, err
	}
	userUID, err := utils.StrToPgUUID(userID)
	if err != nil {
		return nil, err
	}
	return r.queries.GetVotesByRoomAndUser(ctx, dbsqlc.GetVotesByRoomAndUserParams{
		RoomID: roomUID,
		UserID: userUID,
	})
}
