package main

import (
	"errors"
	"log"
	"net/http"
	"time"

	domainfollow "github.com/mespalenza/microblogging-challenge/internal/follow"
	"github.com/mespalenza/microblogging-challenge/internal/platform/memory"
	restfollow "github.com/mespalenza/microblogging-challenge/internal/rest/follow"
	resttimeline "github.com/mespalenza/microblogging-challenge/internal/rest/timeline"
	resttweet "github.com/mespalenza/microblogging-challenge/internal/rest/tweet"
	domaintimeline "github.com/mespalenza/microblogging-challenge/internal/timeline"
	domaintweet "github.com/mespalenza/microblogging-challenge/internal/tweet"
)

func main() {
	tweetRepository := memory.NewTweetRepository()
	tweetService := domaintweet.NewService(tweetRepository)
	tweetHandler := resttweet.NewHandler(tweetService)

	followRepository := memory.NewFollowRepository()
	followService := domainfollow.NewService(followRepository)
	followHandler := restfollow.NewHandler(followService)

	timelineService := domaintimeline.NewService(tweetRepository, followRepository)
	timelineHandler := resttimeline.NewHandler(timelineService)

	router := http.NewServeMux()
	tweetHandler.RegisterRoutes(router)
	followHandler.RegisterRoutes(router)
	timelineHandler.RegisterRoutes(router)

	server := &http.Server{
		Addr:              ":8080",
		Handler:           router,
		ReadHeaderTimeout: 5 * time.Second,
	}

	log.Printf("server running on http://localhost%s", server.Addr)

	if err := server.ListenAndServe(); err != nil &&
		!errors.Is(err, http.ErrServerClosed) {
		log.Fatalf("server failed: %v", err)
	}
}
