package tweet

import (
	"time"

	"github.com/mespalenza/microblogging-challenge/internal/tweet"
)

type TweetRequest struct {
	UserID  string `json:"user_id"`
	Content string `json:"content"`
}

type TweetResponse struct {
	ID      string `json:"id"`
	UserID  string `json:"user_id"`
	Content string `json:"content"`
	Date    string `json:"date"`
}

func (t *TweetRequest) ToDomain() tweet.CreateInput {
	return tweet.CreateInput{
		UserID:  t.UserID,
		Content: t.Content,
	}
}

func NewTweetResponse(value tweet.Tweet) TweetResponse {
	return TweetResponse{
		ID:      value.ID,
		UserID:  value.UserID,
		Content: value.Content,
		Date:    value.CreatedAt.UTC().Format(time.RFC3339),
	}
}
