package memory

import (
	"context"
	"sync"

	"github.com/mespalenza/microblogging-challenge/internal/follow"
)

type FollowRepository struct {
	mu       sync.Mutex
	relation map[follow.Follow]struct{}
}

func NewFollowRepository() *FollowRepository {
	return &FollowRepository{
		relation: make(map[follow.Follow]struct{}),
	}
}

func (r *FollowRepository) SaveFollow(ctx context.Context, value follow.Follow) (bool, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, ok := r.relation[value]; ok {
		return false, nil
	}

	r.relation[value] = struct{}{}
	return true, nil
}
