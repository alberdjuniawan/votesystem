package vote

import (
	"context"
	"errors"
	"time"

	dbsqlc "github.com/alberdjuniawan/votesystem/internal/db/sqlc"
	"github.com/alberdjuniawan/votesystem/internal/modules/leaderboard"
	"github.com/alberdjuniawan/votesystem/internal/modules/realtime"
	"github.com/alberdjuniawan/votesystem/internal/shared/logger"
	"github.com/alberdjuniawan/votesystem/internal/shared/utils"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Service struct {
	repo        *Repository
	db          *pgxpool.Pool
	leaderboard *leaderboard.Service
	hub         *realtime.Hub
}

func NewService(
	repo *Repository,
	db *pgxpool.Pool,
	leaderboard *leaderboard.Service,
	hub *realtime.Hub,
) *Service {
	return &Service{
		repo:        repo,
		db:          db,
		leaderboard: leaderboard,
		hub:         hub,
	}
}

func (s *Service) CastVote(ctx context.Context, roomID, userID string, req CastVoteRequest) (*VoteResponse, error) {
	room, err := s.repo.GetRoomByID(ctx, roomID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrRoomNotFound
		}
		logger.Error(ctx, "Unexpected DB error getting room", "error", err)
		return nil, ErrInternal
	}

	if room.Status != dbsqlc.RoomStatusActive {
		return nil, ErrVotingClosed
	}

	if room.EndsAt.Valid && time.Now().After(room.EndsAt.Time) {
		return nil, ErrVotingClosed
	}
	if room.StartsAt.Valid && time.Now().Before(room.StartsAt.Time) {
		return nil, ErrVotingClosed
	}

	option, err := s.repo.GetOptionByID(ctx, req.OptionID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrOptionNotFound
		}
		logger.Error(ctx, "Unexpected DB error getting option", "error", err)
		return nil, ErrInternal
	}

	if utils.PgUUIDToStr(option.RoomID) != roomID {
		return nil, ErrOptionNotInRoom
	}

	_, err = s.repo.GetVoteByRoomAndUser(ctx, roomID, userID)
	if err == nil {
		return nil, ErrAlreadyVoted
	}
	if !errors.Is(err, pgx.ErrNoRows) {
		logger.Error(ctx, "Unexpected DB error checking existing vote", "error", err)
		return nil, ErrInternal
	}

	vote, err := s.repo.CreateVote(ctx, roomID, userID, req.OptionID)
	if err != nil {
		if isUniqueViolation(err) {
			return nil, ErrAlreadyVoted
		}
		logger.Error(ctx, "Failed to insert vote", "room_id", roomID, "user_id", userID, "error", err)
		return nil, ErrInternal
	}

	if err := s.leaderboard.IncrementVote(ctx, roomID, req.OptionID); err != nil {
		logger.Error(ctx, "Failed to update leaderboard in Redis", "room_id", roomID, "option_id", req.OptionID, "error", err)
	}

	voteCount, _ := s.leaderboard.GetVoteCount(ctx, roomID, req.OptionID)
	totalVotes, _ := s.leaderboard.TotalVotes(ctx, roomID)

	if err := s.hub.Broadcast(roomID, realtime.BroadcastMessage{
		Type:   "vote_update",
		RoomID: roomID,
		Payload: realtime.VoteUpdatePayload{
			OptionID:  req.OptionID,
			VoteCount: voteCount,
			Total:     totalVotes,
		},
	}); err != nil {
		logger.Error(ctx, "Failed to broadcast vote update", "room_id", roomID, "error", err)
	}

	return toVoteResponse(vote), nil
}

func (s *Service) GetMyVote(ctx context.Context, roomID, userID string) (*VoteResponse, error) {
	vote, err := s.repo.GetVoteByRoomAndUser(ctx, roomID, userID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		logger.Error(ctx, "Unexpected DB error getting vote", "error", err)
		return nil, ErrInternal
	}
	return toVoteResponse(vote), nil
}

func toVoteResponse(v dbsqlc.Vote) *VoteResponse {
	return &VoteResponse{
		ID:        utils.PgUUIDToStr(v.ID),
		RoomID:    utils.PgUUIDToStr(v.RoomID),
		UserID:    utils.PgUUIDToStr(v.UserID),
		OptionID:  utils.PgUUIDToStr(v.OptionID),
		CreatedAt: v.CreatedAt.Time,
	}
}

func isUniqueViolation(err error) bool {
	return err != nil && len(err.Error()) > 0 &&
		contains(err.Error(), "23505")
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) &&
		(s == substr || len(s) > 0 && containsStr(s, substr))
}

func containsStr(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
