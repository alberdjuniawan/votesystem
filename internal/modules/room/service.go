package room

import (
	"context"
	"crypto/rand"
	"errors"
	"math/big"

	dbsqlc "github.com/alberdjuniawan/votesystem/internal/db/sqlc"
	"github.com/alberdjuniawan/votesystem/internal/shared/logger"
	"github.com/alberdjuniawan/votesystem/internal/shared/metrics"
	"github.com/alberdjuniawan/votesystem/internal/shared/utils"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	sdktrace "go.opentelemetry.io/otel/trace"
)

var validTransitions = map[dbsqlc.RoomStatus]dbsqlc.RoomStatus{
	dbsqlc.RoomStatusDraft:  dbsqlc.RoomStatusActive,
	dbsqlc.RoomStatusActive: dbsqlc.RoomStatusClosed,
}

type Service struct {
	repo    *Repository
	baseURL string
	tracer  sdktrace.Tracer
}

func NewService(repo *Repository, baseURL string) *Service {
	return &Service{
		repo:    repo,
		baseURL: baseURL,
		tracer: otel.Tracer("github.com/alberdjuniawan/votesystem/internal/modules/room",
			sdktrace.WithInstrumentationVersion("1.0.0"),
		),
	}
}

func (s *Service) CreateRoom(ctx context.Context, ownerID string, req CreateRoomRequest) (*RoomResponse, error) {
	ctx, span := s.tracer.Start(ctx, "room.CreateRoom",
		sdktrace.WithAttributes(attribute.String("owner_id", ownerID)),
	)
	defer span.End()

	ownerUID, err := utils.StrToPgUUID(ownerID)
	if err != nil {
		return nil, ErrInternal
	}

	shareCode, err := s.generateUniqueShareCode(ctx)
	if err != nil {
		logger.Error(ctx, "Failed to generate share code", "error", err)
		return nil, ErrShareCodeCollision
	}

	if req.Type == "single_choice" {
		req.MaxVotes = 1
	} else if req.Type == "multiple_choice" && req.MaxVotes < 2 {
		return nil, ErrInvalidMaxVotes
	}

	params := dbsqlc.CreateRoomParams{
		OwnerID:      ownerUID,
		Title:        req.Title,
		Description:  utils.StrPtrToPgText(&req.Description),
		Type:         dbsqlc.RoomType(req.Type),
		ShowRealtime: req.ShowRealtime,
		MaxVotes:     int32(req.MaxVotes),
		ShareCode:    shareCode,
	}

	if req.StartsAt != nil {
		params.StartsAt = pgtype.Timestamptz{Time: *req.StartsAt, Valid: true}
	}
	if req.EndsAt != nil {
		params.EndsAt = pgtype.Timestamptz{Time: *req.EndsAt, Valid: true}
	}

	room, err := s.repo.CreateRoom(ctx, params)
	if err != nil {
		logger.Error(ctx, "Failed to create room", "owner_id", ownerID, "error", err)
		return nil, ErrInternal
	}

	metrics.RecordRoomCreated(ctx, ownerID)

	return toRoomResponse(room, s.baseURL), nil
}

func (s *Service) GetRoomByID(ctx context.Context, id string) (*RoomResponse, error) {
	ctx, span := s.tracer.Start(ctx, "room.GetRoomByID",
		sdktrace.WithAttributes(attribute.String("room_id", id)),
	)
	defer span.End()

	room, err := s.repo.GetRoomByID(ctx, id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrRoomNotFound
		}
		logger.Error(ctx, "Unexpected DB error getting room", "room_id", id, "error", err)
		return nil, ErrInternal
	}
	return toRoomResponse(room, s.baseURL), nil
}

func (s *Service) GetRoomByShareCode(ctx context.Context, code string) (*RoomResponse, error) {
	ctx, span := s.tracer.Start(ctx, "room.GetRoomByShareCode",
		sdktrace.WithAttributes(attribute.String("share_code", code)),
	)
	defer span.End()

	room, err := s.repo.GetRoomByShareCode(ctx, code)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrRoomNotFound
		}
		logger.Error(ctx, "Unexpected DB error getting room by share code", "code", code, "error", err)
		return nil, ErrInternal
	}
	return toRoomResponse(room, s.baseURL), nil
}

func (s *Service) ListMyRooms(ctx context.Context, ownerID string) ([]RoomResponse, error) {
	ctx, span := s.tracer.Start(ctx, "room.ListMyRooms",
		sdktrace.WithAttributes(attribute.String("owner_id", ownerID)),
	)
	defer span.End()

	rooms, err := s.repo.ListRoomsByOwner(ctx, ownerID)
	if err != nil {
		logger.Error(ctx, "Failed to list rooms", "owner_id", ownerID, "error", err)
		return nil, ErrInternal
	}

	result := make([]RoomResponse, len(rooms))
	for i, r := range rooms {
		result[i] = *toRoomResponse(r, s.baseURL)
	}
	return result, nil
}

func (s *Service) UpdateRoomStatus(ctx context.Context, id, ownerID, newStatus string) (*RoomResponse, error) {
	ctx, span := s.tracer.Start(ctx, "room.UpdateRoomStatus",
		sdktrace.WithAttributes(
			attribute.String("room_id", id),
			attribute.String("new_status", newStatus),
		),
	)
	defer span.End()

	room, err := s.repo.GetRoomByID(ctx, id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrRoomNotFound
		}
		logger.Error(ctx, "Unexpected DB error getting room for status update", "error", err)
		return nil, ErrInternal
	}

	if utils.PgUUIDToStr(room.OwnerID) != ownerID {
		return nil, ErrNotOwner
	}

	expectedNext, ok := validTransitions[room.Status]
	if !ok || expectedNext != dbsqlc.RoomStatus(newStatus) {
		return nil, ErrInvalidStatus
	}

	updated, err := s.repo.UpdateRoomStatus(ctx, id, dbsqlc.RoomStatus(newStatus))
	if err != nil {
		logger.Error(ctx, "Failed to update room status", "room_id", id, "error", err)
		return nil, ErrInternal
	}

	return toRoomResponse(updated, s.baseURL), nil
}

func (s *Service) DeleteRoom(ctx context.Context, id, ownerID string) error {
	ctx, span := s.tracer.Start(ctx, "room.DeleteRoom",
		sdktrace.WithAttributes(
			attribute.String("room_id", id),
			attribute.String("owner_id", ownerID),
		),
	)
	defer span.End()

	room, err := s.repo.GetRoomByID(ctx, id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrRoomNotFound
		}
		logger.Error(ctx, "Unexpected DB error getting room for delete", "error", err)
		return ErrInternal
	}

	if utils.PgUUIDToStr(room.OwnerID) != ownerID {
		return ErrNotOwner
	}

	if err := s.repo.DeleteRoom(ctx, id, ownerID); err != nil {
		logger.Error(ctx, "Failed to delete room", "room_id", id, "error", err)
		return ErrInternal
	}

	return nil
}

func (s *Service) generateUniqueShareCode(ctx context.Context) (string, error) {
	for i := 0; i < 5; i++ {
		code, err := randomBase62(8)
		if err != nil {
			return "", err
		}

		_, err = s.repo.GetRoomByShareCode(ctx, code)
		if errors.Is(err, pgx.ErrNoRows) {
			return code, nil
		}
		if err != nil {
			return "", err
		}
	}
	return "", ErrShareCodeCollision
}

const charsetBase62 = "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789"

func randomBase62(n int) (string, error) {
	result := make([]byte, n)
	for i := range result {
		idx, err := rand.Int(rand.Reader, big.NewInt(int64(len(charsetBase62))))
		if err != nil {
			return "", err
		}
		result[i] = charsetBase62[idx.Int64()]
	}
	return string(result), nil
}

func toRoomResponse(r dbsqlc.Room, baseURL string) *RoomResponse {
	res := &RoomResponse{
		ID:           utils.PgUUIDToStr(r.ID),
		OwnerID:      utils.PgUUIDToStr(r.OwnerID),
		Title:        r.Title,
		Type:         string(r.Type),
		Status:       string(r.Status),
		ShowRealtime: r.ShowRealtime,
		MaxVotes:     int(r.MaxVotes),
		ShareCode:    r.ShareCode,
		ShareURL:     baseURL + "/vote/" + r.ShareCode,
		CreatedAt:    r.CreatedAt.Time,
	}

	if r.Description.Valid {
		res.Description = r.Description.String
	}
	if r.StartsAt.Valid {
		t := r.StartsAt.Time
		res.StartsAt = &t
	}
	if r.EndsAt.Valid {
		t := r.EndsAt.Time
		res.EndsAt = &t
	}

	return res
}
