package timeline

import (
	"context"

	"github.com/mespalenza/microblogging-challenge/internal/tweet"
)

type FollowReader interface {
	FindFollowedIDs(ctx context.Context, followerID string) ([]string, error)
}

type TweetReader interface {
	FindByAuthors(ctx context.Context, authorIDs []string, position *CursorPosition, limit int) ([]tweet.Tweet, error)
}
