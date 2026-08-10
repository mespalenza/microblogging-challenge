package tweet

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/mespalenza/microblogging-challenge/internal/rest/helper"
	"github.com/mespalenza/microblogging-challenge/internal/tweet"
)

type tweetServiceStub struct {
	result    tweet.Tweet
	err       error
	input     tweet.CreateInput
	callCount int
}

func (s *tweetServiceStub) CreateTweet(_ context.Context, input tweet.CreateInput) (tweet.Tweet, error) {
	s.callCount++
	s.input = input
	return s.result, s.err
}

func TestHandler_CreateTweet(t *testing.T) {
	createdAt := time.Date(2026, time.August, 10, 18, 30, 0, 0, time.UTC)
	service := &tweetServiceStub{
		result: tweet.Tweet{
			ID:        "tweet-1",
			UserID:    "user-1",
			Content:   "hello",
			CreatedAt: createdAt,
		},
	}
	handler := NewHandler(service)
	request := httptest.NewRequest(http.MethodPost, "/tweets", strings.NewReader(`{"user_id":" user-1 ","content":" hello "}`))
	recorder := httptest.NewRecorder()

	handler.CreateTweet(recorder, request)

	if recorder.Code != http.StatusCreated {
		t.Fatalf("status = %d, want %d; body = %s", recorder.Code, http.StatusCreated, recorder.Body.String())
	}
	if got := recorder.Header().Get("Content-Type"); got != "application/json" {
		t.Errorf("Content-Type = %q, want application/json", got)
	}

	var got TweetResponse
	if err := json.NewDecoder(recorder.Body).Decode(&got); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	want := TweetResponse{
		ID:      "tweet-1",
		UserID:  "user-1",
		Content: "hello",
		Date:    "2026-08-10T18:30:00Z",
	}
	if got != want {
		t.Errorf("response = %+v, want %+v", got, want)
	}
	if service.callCount != 1 {
		t.Fatalf("CreateTweet calls = %d, want 1", service.callCount)
	}
	wantInput := tweet.CreateInput{UserID: " user-1 ", Content: " hello "}
	if service.input != wantInput {
		t.Errorf("CreateTweet input = %+v, want %+v", service.input, wantInput)
	}
}

func TestHandler_CreateTweetErrors(t *testing.T) {
	tests := []struct {
		name       string
		body       string
		serviceErr error
		wantStatus int
		wantCode   string
		wantCalls  int
	}{
		{name: "malformed JSON", body: `{"user_id":`, wantStatus: http.StatusBadRequest, wantCode: "invalid_json"},
		{name: "unknown field", body: `{"user_id":"user-1","content":"hello","extra":true}`, wantStatus: http.StatusBadRequest, wantCode: "unknown_field"},
		{name: "invalid field type", body: `{"user_id":42,"content":"hello"}`, wantStatus: http.StatusBadRequest, wantCode: "invalid_field_type"},
		{name: "second JSON value", body: `{"user_id":"user-1","content":"hello"} {}`, wantStatus: http.StatusBadRequest, wantCode: "invalid_json"},
		{name: "invalid user ID", body: `{"user_id":" ","content":"hello"}`, serviceErr: tweet.ErrInvalidUserID, wantStatus: http.StatusBadRequest, wantCode: "invalid_user_id", wantCalls: 1},
		{name: "invalid content", body: `{"user_id":"user-1","content":" "}`, serviceErr: tweet.ErrInvalidContent, wantStatus: http.StatusBadRequest, wantCode: "invalid_content", wantCalls: 1},
		{name: "content too long", body: `{"user_id":"user-1","content":"` + strings.Repeat("a", 281) + `"}`, serviceErr: tweet.ErrContentTooLong, wantStatus: http.StatusBadRequest, wantCode: "content_too_long", wantCalls: 1},
		{name: "service failure", body: `{"user_id":"user-1","content":"hello"}`, serviceErr: errors.New("database unavailable"), wantStatus: http.StatusInternalServerError, wantCode: "internal_error", wantCalls: 1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			service := &tweetServiceStub{err: tt.serviceErr}
			handler := NewHandler(service)
			request := httptest.NewRequest(http.MethodPost, "/tweets", strings.NewReader(tt.body))
			recorder := httptest.NewRecorder()

			handler.CreateTweet(recorder, request)

			if recorder.Code != tt.wantStatus {
				t.Fatalf("status = %d, want %d; body = %s", recorder.Code, tt.wantStatus, recorder.Body.String())
			}
			var got helper.ErrorResponse
			if err := json.NewDecoder(recorder.Body).Decode(&got); err != nil {
				t.Fatalf("decode error response: %v", err)
			}
			if got.Error.Code != tt.wantCode {
				t.Errorf("error code = %q, want %q", got.Error.Code, tt.wantCode)
			}
			if service.callCount != tt.wantCalls {
				t.Errorf("CreateTweet calls = %d, want %d", service.callCount, tt.wantCalls)
			}
		})
	}
}

func TestHandler_Routes(t *testing.T) {
	handler := NewHandler(&tweetServiceStub{}).Routes()

	tests := []struct {
		name       string
		method     string
		path       string
		wantStatus int
	}{
		{name: "registered route", method: http.MethodPost, path: "/tweets", wantStatus: http.StatusBadRequest},
		{name: "wrong method", method: http.MethodGet, path: "/tweets", wantStatus: http.StatusMethodNotAllowed},
		{name: "unknown route", method: http.MethodPost, path: "/unknown", wantStatus: http.StatusNotFound},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			request := httptest.NewRequest(tt.method, tt.path, nil)
			recorder := httptest.NewRecorder()
			handler.ServeHTTP(recorder, request)
			if recorder.Code != tt.wantStatus {
				t.Errorf("status = %d, want %d", recorder.Code, tt.wantStatus)
			}
		})
	}
}

func TestTweetMapping(t *testing.T) {
	request := TweetRequest{UserID: "user-1", Content: "hello"}
	if got := request.ToDomain(); got != (tweet.CreateInput{UserID: "user-1", Content: "hello"}) {
		t.Errorf("ToDomain() = %+v", got)
	}

	createdAt := time.Date(2026, time.August, 10, 15, 30, 0, 0, time.FixedZone("ART", -3*60*60))
	got := NewTweetResponse(tweet.Tweet{ID: "tweet-1", UserID: "user-1", Content: "hello", CreatedAt: createdAt})
	want := TweetResponse{ID: "tweet-1", UserID: "user-1", Content: "hello", Date: "2026-08-10T18:30:00Z"}
	if got != want {
		t.Errorf("NewTweetResponse() = %+v, want %+v", got, want)
	}
}
