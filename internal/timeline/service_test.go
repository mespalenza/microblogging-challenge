package timeline

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/mespalenza/microblogging-challenge/internal/tweet"
)

type tweetReaderStub struct {
	tweets  []tweet.Tweet
	err     error
	authors []string
	limit   int
	cursor  *CursorPosition
	calls   int
}

func (s *tweetReaderStub) FindByAuthors(_ context.Context, authors []string, cursor *CursorPosition, limit int) ([]tweet.Tweet, error) {
	s.calls++
	s.authors = authors
	s.cursor = cursor
	s.limit = limit
	return s.tweets, s.err
}

type followReaderStub struct {
	ids    []string
	err    error
	userID string
	calls  int
}

func (s *followReaderStub) FindFollowedIDs(_ context.Context, userID string) ([]string, error) {
	s.calls++
	s.userID = userID
	return s.ids, s.err
}

func TestServiceGetTimeline(t *testing.T) {
	now := time.Date(2026, 8, 11, 12, 0, 0, 0, time.UTC)
	tweets := []tweet.Tweet{{ID: "3", CreatedAt: now}, {ID: "2", CreatedAt: now.Add(-time.Minute)}, {ID: "1", CreatedAt: now.Add(-2 * time.Minute)}}

	t.Run("rejects limit", func(t *testing.T) {
		for _, limit := range []int{0, 101} {
			tr, fr := &tweetReaderStub{}, &followReaderStub{}
			_, err := NewService(tr, fr).GetTimeline(context.Background(), TimelineInput{Limit: limit})
			if !errors.Is(err, ErrLimitOutOfRange) || fr.calls != 0 {
				t.Fatalf("limit=%d err=%v calls=%d", limit, err, fr.calls)
			}
		}
	})

	t.Run("accepts exact limit boundaries", func(t *testing.T) {
		for _, limit := range []int{1, 100} {
			tr := &tweetReaderStub{}
			fr := &followReaderStub{ids: []string{"u2"}}
			_, err := NewService(tr, fr).GetTimeline(context.Background(), TimelineInput{UserID: "u1", Limit: limit})
			if err != nil || fr.calls != 1 || tr.calls != 1 || tr.limit != limit+1 {
				t.Fatalf("limit=%d err=%v followCalls=%d tweetCalls=%d readerLimit=%d", limit, err, fr.calls, tr.calls, tr.limit)
			}
		}
	})

	t.Run("follow lookup error", func(t *testing.T) {
		want := errors.New("follow error")
		fr := &followReaderStub{err: want}
		_, err := NewService(&tweetReaderStub{}, fr).GetTimeline(context.Background(), TimelineInput{Limit: 20})
		if !errors.Is(err, want) {
			t.Fatalf("err=%v", err)
		}
	})

	t.Run("empty follows", func(t *testing.T) {
		tr := &tweetReaderStub{}
		page, err := NewService(tr, &followReaderStub{}).GetTimeline(context.Background(), TimelineInput{Limit: 20})
		if err != nil || page.Tweets == nil || len(page.Tweets) != 0 || tr.calls != 0 {
			t.Fatalf("page=%+v err=%v calls=%d", page, err, tr.calls)
		}
	})

	t.Run("tweet lookup error", func(t *testing.T) {
		want := errors.New("tweet error")
		tr := &tweetReaderStub{err: want}
		_, err := NewService(tr, &followReaderStub{ids: []string{"u2"}}).GetTimeline(context.Background(), TimelineInput{Limit: 2})
		if !errors.Is(err, want) {
			t.Fatalf("err=%v", err)
		}
	})

	t.Run("last page", func(t *testing.T) {
		tr := &tweetReaderStub{tweets: tweets[:2]}
		cursor := &CursorPosition{TweetID: "cursor"}
		page, err := NewService(tr, &followReaderStub{ids: []string{"u2"}}).GetTimeline(context.Background(), TimelineInput{UserID: "u1", Limit: 2, Cursor: cursor})
		if err != nil || len(page.Tweets) != 2 || page.NextCursor != nil || tr.limit != 3 || tr.cursor != cursor {
			t.Fatalf("page=%+v err=%v reader=%+v", page, err, tr)
		}
	})

	t.Run("page with next cursor", func(t *testing.T) {
		tr := &tweetReaderStub{tweets: tweets}
		page, err := NewService(tr, &followReaderStub{ids: []string{"u2"}}).GetTimeline(context.Background(), TimelineInput{Limit: 2})
		if err != nil || len(page.Tweets) != 2 || page.NextCursor == nil || page.NextCursor.TweetID != "2" || !page.NextCursor.CreatedAt.Equal(tweets[1].CreatedAt) {
			t.Fatalf("page=%+v err=%v", page, err)
		}
	})
}
