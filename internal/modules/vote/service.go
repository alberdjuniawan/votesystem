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
	"github.com/jackc/pgx/v5/pgconn"
)

type Service struct {
	repo        *Repository
	leaderboard *leaderboard.Service
	hub         *realtime.Hub
}

func NewService(
	repo *Repository,
	leaderboard *leaderboard.Service,
	hub *realtime.Hub,
) *Service {
	return &Service{
		repo:        repo,
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

	if room.Type == dbsqlc.RoomTypeSingleChoice {
		votes, err := s.repo.GetVotesByRoomAndUser(ctx, roomID, userID)
		if err != nil {
			logger.Error(ctx, "Unexpected DB error checking vote", "error", err)
			return nil, ErrInternal
		}
		if len(votes) > 0 {
			return nil, ErrAlreadyVoted
		}
	} else {
		_, err = s.repo.GetVoteByRoomUserOption(ctx, roomID, userID, req.OptionID)
		if err == nil {
			return nil, ErrAlreadyVotedOption
		}
		if !errors.Is(err, pgx.ErrNoRows) {
			logger.Error(ctx, "Unexpected DB error checking vote", "error", err)
			return nil, ErrInternal
		}

		voteCount, err := s.repo.GetVoteCountByRoomAndUser(ctx, roomID, userID)
		if err != nil {
			logger.Error(ctx, "Unexpected DB error counting votes", "error", err)
			return nil, ErrInternal
		}
		if voteCount >= int64(room.MaxVotes) {
			return nil, ErrMaxVotesReached
		}
	}

	vote, err := s.repo.CreateVote(ctx, roomID, userID, req.OptionID)
	if err != nil {
		if isUniqueViolation(err) {
			if room.Type == dbsqlc.RoomTypeSingleChoice {
				return nil, ErrAlreadyVoted
			}
			return nil, ErrAlreadyVotedOption
		}
		logger.Error(ctx, "Failed to insert vote", "error", err)
		return nil, ErrInternal
	}

	if err := s.leaderboard.IncrementVote(ctx, roomID, req.OptionID); err != nil {
		logger.Error(ctx, "Failed to update leaderboard", "error", err)
	}

	voteCount, _ := s.leaderboard.GetVoteCount(ctx, roomID, req.OptionID)
	totalVotes, _ := s.leaderboard.TotalVotes(ctx, roomID)

	s.hub.Broadcast(roomID, realtime.BroadcastMessage{
		Type:   "vote_update",
		RoomID: roomID,
		Payload: realtime.VoteUpdatePayload{
			OptionID:  req.OptionID,
			VoteCount: voteCount,
			Total:     totalVotes,
		},
	})

	return toVoteResponse(vote), nil
}

func (s *Service) GetMyVote(ctx context.Context, roomID, userID string) (*MyVoteResponse, error) {
	room, err := s.repo.GetRoomByID(ctx, roomID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrRoomNotFound
		}
		logger.Error(ctx, "Unexpected DB error getting room", "error", err)
		return nil, ErrInternal
	}

	votes, err := s.repo.GetVotesByRoomAndUser(ctx, roomID, userID)
	if err != nil {
		logger.Error(ctx, "Unexpected DB error getting votes", "error", err)
		return nil, ErrInternal
	}

	voteResponses := make([]VoteResponse, len(votes))
	for i, v := range votes {
		voteResponses[i] = *toVoteResponse(v)
	}

	return &MyVoteResponse{
		RoomID:   roomID,
		RoomType: string(room.Type),
		Votes:    voteResponses,
		HasVoted: len(votes) > 0,
	}, nil
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
	var pgErr *pgconn.PgError
	return errors.As(err, &pgErr) && pgErr.Code == "23505"
}
