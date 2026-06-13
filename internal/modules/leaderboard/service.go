package leaderboard

import (
	"context"
	"fmt"
	"sort"

	"github.com/alberdjuniawan/votesystem/internal/shared/utils"
	"github.com/redis/go-redis/v9"
)

func roomKey(roomID string) string {
	return fmt.Sprintf("leaderboard:%s", roomID)
}

type Service struct {
	redis *redis.Client
	repo  *Repository
}

func NewService(redis *redis.Client, repo *Repository) *Service {
	return &Service{redis: redis, repo: repo}
}

type OptionScore struct {
	OptionID  string `json:"option_id"`
	VoteCount int64  `json:"vote_count"`
	OrderNum  int    `json:"-"`
}

type LeaderboardResponse struct {
	RoomID string        `json:"room_id"`
	Scores []OptionScore `json:"scores"`
	Total  int64         `json:"total"`
}

func (s *Service) IncrementVote(ctx context.Context, roomID, optionID string) error {
	return s.redis.ZIncrBy(ctx, roomKey(roomID), 1, optionID).Err()
}

func (s *Service) GetVoteCount(ctx context.Context, roomID, optionID string) (int64, error) {
	score, err := s.redis.ZScore(ctx, roomKey(roomID), optionID).Result()
	if err == redis.Nil {
		return 0, nil
	}
	if err != nil {
		return 0, err
	}
	return int64(score), nil
}

func (s *Service) TotalVotes(ctx context.Context, roomID string) (int64, error) {
	results, err := s.redis.ZRangeWithScores(ctx, roomKey(roomID), 0, -1).Result()
	if err != nil {
		return 0, err
	}

	var total int64
	for _, r := range results {
		total += int64(r.Score)
	}
	return total, nil
}

func (s *Service) SeedLeaderboard(ctx context.Context, roomID string, scores []OptionScore) error {
	if len(scores) == 0 {
		return nil
	}
	members := make([]redis.Z, len(scores))
	for i, sc := range scores {
		members[i] = redis.Z{
			Score:  float64(sc.VoteCount),
			Member: sc.OptionID,
		}
	}
	return s.redis.ZAdd(ctx, roomKey(roomID), members...).Err()
}

func (s *Service) DeleteLeaderboard(ctx context.Context, roomID string) error {
	return s.redis.Del(ctx, roomKey(roomID)).Err()
}

func (s *Service) GetLeaderboard(ctx context.Context, roomID string) (*LeaderboardResponse, error) {
	dbOptions, err := s.repo.ListOptionsByRoom(ctx, roomID)
	if err != nil {
		return nil, err
	}

	results, err := s.redis.ZRevRangeWithScores(ctx, roomKey(roomID), 0, -1).Result()
	if err != nil {
		return nil, err
	}

	scoreMap := make(map[string]int64, len(results))
	for _, r := range results {
		scoreMap[r.Member.(string)] = int64(r.Score)
	}

	merged := make([]OptionScore, len(dbOptions))
	for i, o := range dbOptions {
		optID := utils.PgUUIDToStr(o.ID)
		merged[i] = OptionScore{
			OptionID:  optID,
			VoteCount: scoreMap[optID],
			OrderNum:  int(o.OrderNum),
		}
	}

	sort.SliceStable(merged, func(i, j int) bool {
		if merged[i].VoteCount != merged[j].VoteCount {
			return merged[i].VoteCount > merged[j].VoteCount
		}
		return merged[i].OrderNum < merged[j].OrderNum
	})

	var total int64
	for _, sc := range merged {
		total += sc.VoteCount
	}

	return &LeaderboardResponse{
		RoomID: roomID,
		Scores: merged,
		Total:  total,
	}, nil
}
