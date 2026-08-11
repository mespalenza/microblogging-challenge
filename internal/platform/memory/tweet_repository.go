package memory

import (
	"context"
	"errors"
	"sort"
	"sync"

	"github.com/mespalenza/microblogging-challenge/internal/timeline"
	"github.com/mespalenza/microblogging-challenge/internal/tweet"
)

type TweetRepository struct {
	mu             sync.RWMutex
	tweets         map[string]tweet.Tweet
	tweetsByAuthor map[string][]tweet.Tweet
}

func NewTweetRepository() *TweetRepository {
	return &TweetRepository{
		tweets:         make(map[string]tweet.Tweet),
		tweetsByAuthor: make(map[string][]tweet.Tweet),
	}
}

func (r *TweetRepository) Save(ctx context.Context, value tweet.Tweet) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.tweets[value.ID] = value
	r.tweetsByAuthor[value.UserID] = append(r.tweetsByAuthor[value.UserID], value)
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

func (r *TweetRepository) FindByAuthors(ctx context.Context, authorIDs []string, position *timeline.CursorPosition, limit int) ([]tweet.Tweet, error) {
	r.mu.RLock()

	tweets := make([]tweet.Tweet, 0)

	for _, authorID := range authorIDs {
		for _, value := range r.tweetsByAuthor[authorID] {
			if position != nil {
				isAfterCursor :=
					value.CreatedAt.Before(position.CreatedAt) ||
						(value.CreatedAt.Equal(position.CreatedAt) &&
							value.ID < position.TweetID)

				if !isAfterCursor {
					continue
				}
			}

			tweets = append(tweets, value)
		}
	}

	r.mu.RUnlock()

	sort.Slice(tweets, func(i, j int) bool {
		if !tweets[i].CreatedAt.Equal(tweets[j].CreatedAt) {
			return tweets[i].CreatedAt.After(tweets[j].CreatedAt)
		}

		return tweets[i].ID > tweets[j].ID
	})

	if len(tweets) > limit {
		tweets = tweets[:limit]
	}

	return tweets, nil
}
