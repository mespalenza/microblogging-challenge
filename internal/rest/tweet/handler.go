package tweet

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/mespalenza/microblogging-challenge/internal/rest/errorcatalog"
	"github.com/mespalenza/microblogging-challenge/internal/rest/helper"
	"github.com/mespalenza/microblogging-challenge/internal/tweet"
)

type TweetCreator interface {
	CreateTweet(ctx context.Context, input tweet.CreateInput) (tweet.Tweet, error)
}
type Handler struct {
	service TweetCreator
}

func NewHandler(service TweetCreator) Handler {
	return Handler{
		service: service,
	}
}

func (h Handler) CreateTweet(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	var tweetReq TweetRequest

	if err := helper.DecodeJSON(r.Body, &tweetReq); err != nil {
		mappedErr := errorcatalog.FromDecode(err)
		helper.WriteError(w, mappedErr.Status, mappedErr.Code, mappedErr.Message)
		return
	}

	createTweet, err := h.service.CreateTweet(ctx, tweetReq.ToDomain())
	if err != nil {
		mappedErr := errorcatalog.From(err)
		helper.WriteError(w, mappedErr.Status, mappedErr.Code, mappedErr.Message)
		return
	}

	tweetResp := NewTweetResponse(createTweet)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(tweetResp)
	return
}
