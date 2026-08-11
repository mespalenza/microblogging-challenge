package timeline

import (
	"time"

	"github.com/mespalenza/microblogging-challenge/internal/tweet"
)

type TimelineInput struct {
	UserID string
	Limit  int
	Cursor *CursorPosition
}

type TimelinePage struct {
	Tweets     []tweet.Tweet
	NextCursor *CursorPosition
}

type CursorPosition struct {
	CreatedAt time.Time
	TweetID   string
}
