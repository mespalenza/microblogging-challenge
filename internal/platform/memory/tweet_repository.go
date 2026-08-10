package memory

import (
	"context"
	"errors"
	"sync"

	"github.com/mespalenza/microblogging-challenge/internal/tweet"
)

type TweetRepository struct {
	mu     sync.RWMutex
	tweets map[string]tweet.Tweet
}

func NewTweetRepository() *TweetRepository {
	return &TweetRepository{
		tweets: make(map[string]tweet.Tweet),
	}
}

func (r *TweetRepository) Save(ctx context.Context, value tweet.Tweet) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.tweets[value.ID] = value
	return nil
}

func (r *TweetRepository) FindByID(ctx context.Context, id string) (tweet.Tweet, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	if tweet, ok := r.tweets[id]; ok {
		return tweet, nil
	}
	return tweet.Tweet{}, errors.New("tweet not found")
}
