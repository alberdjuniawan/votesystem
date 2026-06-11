package option

import (
	"context"
	"encoding/json"
	"errors"

	dbsqlc "github.com/alberdjuniawan/votesystem/internal/db/sqlc"
	"github.com/alberdjuniawan/votesystem/internal/shared/logger"
	"github.com/alberdjuniawan/votesystem/internal/shared/utils"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
)

type Service struct {
	repo    *Repository
	baseURL string
}

func NewService(repo *Repository, baseURL string) *Service {
	return &Service{repo: repo, baseURL: baseURL}
}

func (s *Service) CreateOption(ctx context.Context, roomID, userID string, req CreateOptionRequest) (*OptionResponse, error) {
	if err := s.validateRoomOwnerDraft(ctx, roomID, userID); err != nil {
		return nil, err
	}

	roomUID, err := utils.StrToPgUUID(roomID)
	if err != nil {
		return nil, ErrInternal
	}

	params := dbsqlc.CreateOptionParams{
		RoomID:      roomUID,
		Label:       req.Label,
		Description: utils.StrPtrToPgText(strPtr(req.Description)),
		OrderNum:    int32(req.OrderNum),
	}

	if req.Metadata != nil {
		params.Metadata = req.Metadata
	}

	if req.MediaID != "" {
		mediaUID, err := utils.StrToPgUUID(req.MediaID)
		if err != nil {
			return nil, ErrInternal
		}
		params.MediaID = pgtype.UUID{Bytes: mediaUID.Bytes, Valid: true}
	}

	opt, err := s.repo.CreateOption(ctx, params)
	if err != nil {
		logger.Error(ctx, "Failed to create option", "room_id", roomID, "error", err)
		return nil, ErrInternal
	}

	return s.toOptionResponse(ctx, opt), nil
}

func (s *Service) ListOptions(ctx context.Context, roomID string) ([]OptionResponse, error) {
	_, err := s.repo.GetRoomByID(ctx, roomID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrRoomNotFound
		}
		logger.Error(ctx, "Unexpected DB error getting room", "error", err)
		return nil, ErrInternal
	}

	opts, err := s.repo.ListOptionsByRoom(ctx, roomID)
	if err != nil {
		logger.Error(ctx, "Failed to list options", "room_id", roomID, "error", err)
		return nil, ErrInternal
	}

	result := make([]OptionResponse, len(opts))
	for i, o := range opts {
		result[i] = *s.toOptionResponse(ctx, o)
	}
	return result, nil
}

func (s *Service) UpdateOption(ctx context.Context, optionID, userID string, req UpdateOptionRequest) (*OptionResponse, error) {
	opt, err := s.repo.GetOptionByID(ctx, optionID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrOptionNotFound
		}
		logger.Error(ctx, "Unexpected DB error getting option", "error", err)
		return nil, ErrInternal
	}

	roomID := utils.PgUUIDToStr(opt.RoomID)
	if err := s.validateRoomOwnerDraft(ctx, roomID, userID); err != nil {
		return nil, err
	}

	label := opt.Label
	if req.Label != nil {
		label = *req.Label
	}

	description := opt.Description
	if req.Description != nil {
		description = utils.StrPtrToPgText(req.Description)
	}

	metadata := opt.Metadata
	if req.Metadata != nil {
		metadata = req.Metadata
	}

	mediaID := opt.MediaID
	if req.MediaID != nil {
		if *req.MediaID == "" {
			mediaID = pgtype.UUID{Valid: false}
		} else {
			uid, err := utils.StrToPgUUID(*req.MediaID)
			if err != nil {
				return nil, ErrInternal
			}
			mediaID = uid
		}
	}

	orderNum := opt.OrderNum
	if req.OrderNum != nil {
		orderNum = int32(*req.OrderNum)
	}

	optUID, _ := utils.StrToPgUUID(optionID)
	updated, err := s.repo.UpdateOption(ctx, dbsqlc.UpdateOptionParams{
		ID:          optUID,
		Label:       label,
		Description: description,
		Metadata:    metadata,
		MediaID:     mediaID,
		OrderNum:    orderNum,
	})
	if err != nil {
		logger.Error(ctx, "Failed to update option", "option_id", optionID, "error", err)
		return nil, ErrInternal
	}

	return s.toOptionResponse(ctx, updated), nil
}

func (s *Service) DeleteOption(ctx context.Context, optionID, userID string) error {
	opt, err := s.repo.GetOptionByID(ctx, optionID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrOptionNotFound
		}
		logger.Error(ctx, "Unexpected DB error getting option", "error", err)
		return ErrInternal
	}

	roomID := utils.PgUUIDToStr(opt.RoomID)
	if err := s.validateRoomOwnerDraft(ctx, roomID, userID); err != nil {
		return err
	}

	if err := s.repo.DeleteOption(ctx, optionID); err != nil {
		logger.Error(ctx, "Failed to delete option", "option_id", optionID, "error", err)
		return ErrInternal
	}

	return nil
}

func (s *Service) validateRoomOwnerDraft(ctx context.Context, roomID, userID string) error {
	room, err := s.repo.GetRoomByID(ctx, roomID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrRoomNotFound
		}
		logger.Error(ctx, "Unexpected DB error validating room", "error", err)
		return ErrInternal
	}

	if utils.PgUUIDToStr(room.OwnerID) != userID {
		return ErrNotRoomOwner
	}

	if room.Status != dbsqlc.RoomStatusDraft {
		return ErrRoomNotDraft
	}

	return nil
}

func (s *Service) toOptionResponse(ctx context.Context, o dbsqlc.Option) *OptionResponse {
	res := &OptionResponse{
		ID:        utils.PgUUIDToStr(o.ID),
		RoomID:    utils.PgUUIDToStr(o.RoomID),
		Label:     o.Label,
		OrderNum:  int(o.OrderNum),
		CreatedAt: o.CreatedAt.Time,
	}

	if o.Description.Valid {
		res.Description = o.Description.String
	}

	if o.Metadata != nil {
		res.Metadata = json.RawMessage(o.Metadata)
	}

	if o.MediaID.Valid {
		mediaID := utils.PgUUIDToStr(o.MediaID)
		res.MediaID = mediaID

		media, err := s.repo.GetMediaByID(ctx, mediaID)
		if err == nil {
			res.MediaURL = s.baseURL + "/media/" + media.StoragePath
		}
	}

	return res
}

func strPtr(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}
