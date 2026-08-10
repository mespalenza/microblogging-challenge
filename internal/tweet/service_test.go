package tweet

import (
	"context"
	"errors"
	"strings"
	"testing"
)

type repositoryStub struct {
	savedTweet Tweet
	saveCalls  int
	saveErr    error
}

func (r *repositoryStub) Save(_ context.Context, value Tweet) error {
	r.saveCalls++
	r.savedTweet = value
	return r.saveErr
}

func TestServiceTweet_CreateTweet(t *testing.T) {
	tests := []struct {
		name        string
		request     CreateInput
		wantErr     error
		wantUserID  string
		wantContent string
	}{
		{
			name:    "rejects an empty user ID",
			request: CreateInput{Content: "hello"},
			wantErr: ErrInvalidUserID,
		},
		{
			name:    "rejects a user ID containing only spaces",
			request: CreateInput{UserID: "   ", Content: "hello"},
			wantErr: ErrInvalidUserID,
		},
		{
			name:    "validates the user ID before the content",
			request: CreateInput{},
			wantErr: ErrInvalidUserID,
		},
		{
			name:    "rejects empty content",
			request: CreateInput{UserID: "user-123"},
			wantErr: ErrInvalidContent,
		},
		{
			name:    "rejects content containing only spaces",
			request: CreateInput{UserID: "user-123", Content: "   "},
			wantErr: ErrInvalidContent,
		},
		{
			name: "accepts valid values surrounded by spaces",
			request: CreateInput{
				UserID:  "  user-123  ",
				Content: "  hello  ",
			},
			wantUserID:  "user-123",
			wantContent: "hello",
		},
		{
			name: "rejects content containing 281 ASCII characters",
			request: CreateInput{
				UserID:  "user-123",
				Content: strings.Repeat("a", 281),
			},
			wantErr: ErrContentTooLong,
		},
		{
			name: "accepts content containing exactly 280 ASCII characters",
			request: CreateInput{
				UserID:  "user-123",
				Content: strings.Repeat("a", 280),
			},
		},
		{
			name: "rejects content containing 281 emojis",
			request: CreateInput{
				UserID:  "user-123",
				Content: strings.Repeat("😀", 281),
			},
			wantErr: ErrContentTooLong,
		},
		{
			name: "accepts content containing exactly 280 emojis",
			request: CreateInput{
				UserID:  "user-123",
				Content: strings.Repeat("😀", 280),
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repository := &repositoryStub{}
			service := NewService(repository)

			got, err := service.CreateTweet(context.Background(), tt.request)

			if !errors.Is(err, tt.wantErr) {
				t.Fatalf("CreateTweet() error = %v, want %v", err, tt.wantErr)
			}

			if tt.wantUserID != "" && got.UserID != tt.wantUserID {
				t.Errorf("UserID = %q, want %q", got.UserID, tt.wantUserID)
			}

			if tt.wantContent != "" && got.Content != tt.wantContent {
				t.Errorf("Content = %q, want %q", got.Content, tt.wantContent)
			}

			if tt.wantErr != nil {
				if repository.saveCalls != 0 {
					t.Errorf("repository Save calls = %d, want 0", repository.saveCalls)
				}
				return
			}

			if repository.saveCalls != 1 {
				t.Fatalf("repository Save calls = %d, want 1", repository.saveCalls)
			}

			if repository.savedTweet != got {
				t.Errorf("saved tweet = %+v, want %+v", repository.savedTweet, got)
			}
		})
	}
}

func TestServiceTweet_CreateTweetReturnsRepositoryError(t *testing.T) {
	wantErr := errors.New("save tweet")
	repository := &repositoryStub{saveErr: wantErr}
	service := NewService(repository)

	got, err := service.CreateTweet(context.Background(), CreateInput{
		UserID:  "user-123",
		Content: "hello",
	})

	if !errors.Is(err, wantErr) {
		t.Fatalf("CreateTweet() error = %v, want %v", err, wantErr)
	}
	if got != (Tweet{}) {
		t.Errorf("CreateTweet() tweet = %+v, want zero value", got)
	}
	if repository.saveCalls != 1 {
		t.Errorf("repository Save calls = %d, want 1", repository.saveCalls)
	}
}
