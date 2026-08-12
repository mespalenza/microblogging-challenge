package memory

import (
	"context"
	"sync"

	"github.com/mespalenza/microblogging-challenge/internal/follow"
)

type FollowRepository struct {
	mu       sync.RWMutex
	relation map[string]map[string]struct{}
}

func NewFollowRepository() *FollowRepository {
	return &FollowRepository{
		relation: make(map[string]map[string]struct{}),
	}
}

func (r *FollowRepository) SaveFollow(ctx context.Context, value follow.Follow) (bool, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if err := ctx.Err(); err != nil {
		return false, err
	}

	followedUsers, exists := r.relation[value.FollowerID]
	if !exists {
		followedUsers = make(map[string]struct{})
		r.relation[value.FollowerID] = followedUsers
	}

	if _, exists := followedUsers[value.FollowedID]; exists {
		return false, nil
	}

	followedUsers[value.FollowedID] = struct{}{}
	return true, nil
}

func (r *FollowRepository) FindFollowedIDs(ctx context.Context, followerID string) ([]string, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	if followedUsers, exists := r.relation[followerID]; exists {
		followedIDs := make([]string, 0, len(followedUsers))
		for followedID := range followedUsers {
			if err := ctx.Err(); err != nil {
				return nil, err
			}
			followedIDs = append(followedIDs, followedID)
		}
		return followedIDs, nil
	}
	return []string{}, nil
}
