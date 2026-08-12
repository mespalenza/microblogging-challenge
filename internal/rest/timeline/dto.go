package timeline

import (
	resttweet "github.com/mespalenza/microblogging-challenge/internal/rest/tweet"
	domaintimeline "github.com/mespalenza/microblogging-challenge/internal/timeline"
)

type TimelineResponse struct {
	Tweets     []resttweet.TweetResponse `json:"tweets"`
	NextCursor *string                   `json:"next_cursor"`
}

func NewTimelineResponse(page domaintimeline.TimelinePage) (TimelineResponse, error) {
	nextCursor, err := EncodeCursor(page.NextCursor)
	if err != nil {
		return TimelineResponse{}, err
	}

	tweets := make([]resttweet.TweetResponse, 0, len(page.Tweets))
	for _, value := range page.Tweets {
		tweets = append(tweets, resttweet.NewTweetResponse(value))
	}

	return TimelineResponse{
		Tweets:     tweets,
		NextCursor: nextCursor,
	}, nil
}
