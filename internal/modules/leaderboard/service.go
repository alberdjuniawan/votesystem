package leaderboard

import (
	"context"
	"fmt"

	"github.com/redis/go-redis/v9"
)

func roomKey(roomID string) string {
	return fmt.Sprintf("leaderboard:%s", roomID)
}

type Service struct {
	redis *redis.Client
}

func NewService(redis *redis.Client) *Service {
	return &Service{redis: redis}
}

type OptionScore struct {
	OptionID  string `json:"option_id"`
	VoteCount int64  `json:"vote_count"`
}

func (s *Service) IncrementVote(ctx context.Context, roomID, optionID string) error {
	return s.redis.ZIncrBy(ctx, roomKey(roomID), 1, optionID).Err()
}

func (s *Service) GetLeaderboard(ctx context.Context, roomID string) ([]OptionScore, error) {
	results, err := s.redis.ZRevRangeWithScores(ctx, roomKey(roomID), 0, -1).Result()
	if err != nil {
		return nil, err
	}

	scores := make([]OptionScore, len(results))
	for i, r := range results {
		scores[i] = OptionScore{
			OptionID:  r.Member.(string),
			VoteCount: int64(r.Score),
		}
	}
	return scores, nil
}

func (s *Service) SeedLeaderboard(ctx context.Context, roomID string, scores []OptionScore) error {
	if len(scores) == 0 {
		return nil
	}

	members := make([]redis.Z, len(scores))
	for i, s := range scores {
		members[i] = redis.Z{
			Score:  float64(s.VoteCount),
			Member: s.OptionID,
		}
	}

	return s.redis.ZAdd(ctx, roomKey(roomID), members...).Err()
}

func (s *Service) DeleteLeaderboard(ctx context.Context, roomID string) error {
	return s.redis.Del(ctx, roomKey(roomID)).Err()
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
