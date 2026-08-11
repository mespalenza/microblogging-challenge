package timeline

import (
	"context"

	"github.com/mespalenza/microblogging-challenge/internal/tweet"
)

type Service struct {
	tweetReader  TweetReader
	followReader FollowReader
}

func NewService(tweetReader TweetReader, followReader FollowReader) *Service {
	return &Service{
		tweetReader:  tweetReader,
		followReader: followReader,
	}
}

func (s *Service) GetTimeline(ctx context.Context, input TimelineInput) (TimelinePage, error) {
	if input.Limit < 1 || input.Limit > 100 {
		return TimelinePage{}, ErrLimitOutOfRange
	}

	followedIDs, err := s.followReader.FindFollowedIDs(ctx, input.UserID)
	if err != nil {
		return TimelinePage{}, err
	}

	if len(followedIDs) == 0 {
		return TimelinePage{Tweets: make([]tweet.Tweet, 0)}, nil
	}

	tweets, err := s.tweetReader.FindByAuthors(ctx, followedIDs, input.Cursor, input.Limit+1)
	if err != nil {
		return TimelinePage{}, err
	}

	if len(tweets) <= input.Limit {
		return TimelinePage{
			Tweets: tweets,
		}, nil
	}

	tweets = tweets[:input.Limit]
	lastTweet := tweets[len(tweets)-1]

	nextCursor := CursorPosition{
		CreatedAt: lastTweet.CreatedAt,
		TweetID:   lastTweet.ID,
	}

	return TimelinePage{
		Tweets:     tweets,
		NextCursor: &nextCursor,
	}, nil
}
