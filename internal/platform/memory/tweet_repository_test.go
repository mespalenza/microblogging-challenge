package memory

import (
	"context"
	"fmt"
	"sync"
	"testing"

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
