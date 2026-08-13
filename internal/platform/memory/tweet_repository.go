package memory

import (
	"container/heap"
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

type authorCursor struct {
	tweets []tweet.Tweet
	index  int
}

type tweetHeap []authorCursor

func (h tweetHeap) Len() int {
	return len(h)
}

func (h tweetHeap) Less(i, j int) bool {
	left := h[i].tweets[h[i].index]
	right := h[j].tweets[h[j].index]

	return comesBefore(left, right)
}

func (h tweetHeap) Swap(i, j int) {
	h[i], h[j] = h[j], h[i]
}

func (h *tweetHeap) Push(value any) {
	*h = append(*h, value.(authorCursor))
}

func (h *tweetHeap) Pop() any {
	old := *h
	lastIndex := len(old) - 1
	value := old[lastIndex]

	old[lastIndex] = authorCursor{}
	*h = old[:lastIndex]

	return value
}

func comesBefore(left, right tweet.Tweet) bool {
	if !left.CreatedAt.Equal(right.CreatedAt) {
		return left.CreatedAt.After(right.CreatedAt)
	}

	return left.ID > right.ID
}

func isAfterCursor(value tweet.Tweet, position *timeline.CursorPosition) bool {
	if position == nil {
		return true
	}

	return value.CreatedAt.Before(position.CreatedAt) ||
		(value.CreatedAt.Equal(position.CreatedAt) &&
			value.ID < position.TweetID)
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
	if err := ctx.Err(); err != nil {
		return err
	}

	r.tweets[value.ID] = value

	authorTweets := r.tweetsByAuthor[value.UserID]

	insertAt := sort.Search(len(authorTweets), func(index int) bool {
		return !comesBefore(authorTweets[index], value)
	},
	)

	authorTweets = append(authorTweets, tweet.Tweet{})
	copy(authorTweets[insertAt+1:], authorTweets[insertAt:])
	authorTweets[insertAt] = value

	r.tweetsByAuthor[value.UserID] = authorTweets

	return nil
}

func (r *TweetRepository) FindByID(ctx context.Context, id string) (tweet.Tweet, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	if err := ctx.Err(); err != nil {
		return tweet.Tweet{}, err
	}

	if tweet, ok := r.tweets[id]; ok {
		return tweet, nil
	}
	return tweet.Tweet{}, errors.New("tweet not found")
}

func (r *TweetRepository) FindByAuthors(ctx context.Context, authorIDs []string, position *timeline.CursorPosition, limit int) ([]tweet.Tweet, error) {
	if limit <= 0 {
		return []tweet.Tweet{}, nil
	}

	r.mu.RLock()
	defer r.mu.RUnlock()
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	candidates := make(tweetHeap, 0, len(authorIDs))

	for _, authorID := range authorIDs {
		if err := ctx.Err(); err != nil {
			return nil, err
		}

		authorTweets := r.tweetsByAuthor[authorID]

		startAt := sort.Search(
			len(authorTweets),
			func(index int) bool {
				return isAfterCursor(authorTweets[index], position)
			},
		)

		if startAt < len(authorTweets) {
			candidates = append(candidates, authorCursor{
				tweets: authorTweets,
				index:  startAt,
			})
		}
	}

	heap.Init(&candidates)

	result := make([]tweet.Tweet, 0, limit)

	for len(result) < limit && candidates.Len() > 0 {
		if err := ctx.Err(); err != nil {
			return nil, err
		}

		current := heap.Pop(&candidates).(authorCursor)

		result = append(result, current.tweets[current.index])

		current.index++
		if current.index < len(current.tweets) {
			heap.Push(&candidates, current)
		}
	}

	return result, nil
}
