package main

import (
	"errors"
	"log"
	"net/http"
	"time"

	"github.com/mespalenza/microblogging-challenge/internal/platform/memory"
	resttweet "github.com/mespalenza/microblogging-challenge/internal/rest/tweet"
	"github.com/mespalenza/microblogging-challenge/internal/tweet"
)

func main() {

	tweetRepository := memory.NewTweetRepository()
	tweetService := tweet.NewService(tweetRepository)
	handler := resttweet.NewHandler(tweetService)

	server := &http.Server{
		Addr:              ":8080",
		Handler:           handler.Routes(),
		ReadHeaderTimeout: 5 * time.Second,
	}

	log.Printf("server running on http://localhost%s", server.Addr)

	if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		log.Fatalf("server failed: %v", err)
	}
}
