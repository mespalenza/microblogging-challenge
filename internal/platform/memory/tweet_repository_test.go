package memory

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/mespalenza/microblogging-challenge/internal/timeline"
	"github.com/mespalenza/microblogging-challenge/internal/tweet"
)

func TestTweetRepository_SaveAndFindByID(t *testing.T) {
	repository := NewTweetRepository()
	want := tweet.Tweet{ID: "tweet-1", UserID: "user-1", Content: "hello"}

	if err := repository.Save(context.Background(), want); err != nil {
		t.Fatalf("Save() error = %v", err)
	}

	got, err := repository.FindByID(context.Background(), want.ID)
	if err != nil {
		t.Fatalf("FindByID() error = %v", err)
	}
	if got != want {
		t.Errorf("FindByID() = %+v, want %+v", got, want)
	}
}

func TestTweetRepository_CanceledContext(t *testing.T) {
	repository := NewTweetRepository()
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	want := tweet.Tweet{ID: "tweet-1", UserID: "user-1"}

	if err := repository.Save(ctx, want); !errors.Is(err, context.Canceled) {
		t.Fatalf("Save() error=%v", err)
	}
	if _, err := repository.FindByID(ctx, want.ID); !errors.Is(err, context.Canceled) {
		t.Fatalf("FindByID() error=%v", err)
	}
	if got, err := repository.FindByAuthors(ctx, []string{want.UserID}, nil, 20); got != nil || !errors.Is(err, context.Canceled) {
		t.Fatalf("FindByAuthors() tweets=%v error=%v", got, err)
	}

	if _, err := repository.FindByID(context.Background(), want.ID); err == nil {
		t.Fatal("canceled save changed state")
	}
}

func TestTweetRepository_FindByAuthors(t *testing.T) {
	repository := NewTweetRepository()
	ctx := context.Background()
	createdAt := time.Date(2026, 8, 11, 12, 0, 0, 0, time.UTC)
	values := []tweet.Tweet{
		{ID: "b", UserID: "u1", CreatedAt: createdAt},
		{ID: "a", UserID: "u1", CreatedAt: createdAt},
		{ID: "c", UserID: "u2", CreatedAt: createdAt.Add(time.Minute)},
		{ID: "ignored", UserID: "u3", CreatedAt: createdAt.Add(time.Hour)},
	}
	for _, value := range values {
		if err := repository.Save(ctx, value); err != nil {
			t.Fatal(err)
		}
	}

	got, err := repository.FindByAuthors(ctx, []string{"u1", "u2"}, nil, 2)
	if err != nil || len(got) != 2 || got[0].ID != "c" || got[1].ID != "b" {
		t.Fatalf("FindByAuthors() = %+v, err=%v", got, err)
	}

	cursor := &timeline.CursorPosition{CreatedAt: createdAt, TweetID: "b"}
	got, err = repository.FindByAuthors(ctx, []string{"u1", "u2"}, cursor, 10)
	if err != nil || len(got) != 1 || got[0].ID != "a" {
		t.Fatalf("FindByAuthors(cursor) = %+v, err=%v", got, err)
	}
}

func TestTweetRepository_FindByIDReturnsErrorWhenMissing(t *testing.T) {
	repository := NewTweetRepository()

	if _, err := repository.FindByID(context.Background(), "missing"); err == nil {
		t.Fatal("FindByID() error = nil, want an error")
	}
}

func TestTweetRepository_ConcurrentAccess(t *testing.T) {
	repository := NewTweetRepository()
	const count = 50
	var waitGroup sync.WaitGroup

	for index := 0; index < count; index++ {
		index := index
		waitGroup.Add(1)
		go func() {
			defer waitGroup.Done()
			value := tweet.Tweet{ID: fmt.Sprintf("tweet-%d", index)}
			if err := repository.Save(context.Background(), value); err != nil {
				t.Errorf("Save() error = %v", err)
			}
			if _, err := repository.FindByID(context.Background(), value.ID); err != nil {
				t.Errorf("FindByID() error = %v", err)
			}
		}()
	}

	waitGroup.Wait()
}
