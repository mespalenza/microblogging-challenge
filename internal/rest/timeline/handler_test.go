package timeline

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/mespalenza/microblogging-challenge/internal/rest/helper"
	domain "github.com/mespalenza/microblogging-challenge/internal/timeline"
	"github.com/mespalenza/microblogging-challenge/internal/tweet"
)

type timelineServiceStub struct {
	page  domain.TimelinePage
	err   error
	input domain.TimelineInput
	calls int
}

func (s *timelineServiceStub) GetTimeline(_ context.Context, input domain.TimelineInput) (domain.TimelinePage, error) {
	s.calls++
	s.input = input
	return s.page, s.err
}

func TestGetTimeline(t *testing.T) {
	createdAt := time.Date(2026, 8, 11, 15, 0, 0, 0, time.UTC)
	t.Run("default pagination", func(t *testing.T) {
		s := &timelineServiceStub{page: domain.TimelinePage{Tweets: []tweet.Tweet{{ID: "t1", UserID: "u2", Content: "hi", CreatedAt: createdAt}}}}
		w := serve(s, "/users/u1/timeline")
		if w.Code != http.StatusOK || s.input.UserID != "u1" || s.input.Limit != defaultLimit || s.input.Cursor != nil {
			t.Fatalf("status=%d input=%+v body=%s", w.Code, s.input, w.Body.String())
		}
		var got TimelineResponse
		if err := json.NewDecoder(w.Body).Decode(&got); err != nil || len(got.Tweets) != 1 || got.Tweets[0].ID != "t1" {
			t.Fatalf("response=%+v err=%v", got, err)
		}
	})
	t.Run("limit and cursor", func(t *testing.T) {
		encoded, _ := EncodeCursor(&domain.CursorPosition{CreatedAt: createdAt, TweetID: "t9"})
		s := &timelineServiceStub{}
		w := serve(s, "/users/u1/timeline?limit=7&cursor="+*encoded)
		if w.Code != 200 || s.input.Limit != 7 || s.input.Cursor == nil || s.input.Cursor.TweetID != "t9" {
			t.Fatalf("status=%d input=%+v", w.Code, s.input)
		}
	})
	for _, tc := range []struct {
		name, path, code string
		status           int
	}{
		{"duplicate limit", "/users/u1/timeline?limit=1&limit=2", "invalid_pagination_parameters", 400},
		{"invalid limit", "/users/u1/timeline?limit=no", "invalid_pagination_parameters", 400},
		{"duplicate cursor", "/users/u1/timeline?cursor=a&cursor=b", "invalid_pagination_parameters", 400},
		{"empty cursor", "/users/u1/timeline?cursor=", "invalid_cursor", 400},
		{"invalid cursor", "/users/u1/timeline?cursor=not-base64!", "invalid_cursor", 400},
	} {
		t.Run(tc.name, func(t *testing.T) {
			s := &timelineServiceStub{}
			w := serve(s, tc.path)
			assertError(t, w, tc.status, tc.code)
			if s.calls != 0 {
				t.Errorf("calls=%d", s.calls)
			}
		})
	}
	t.Run("service error", func(t *testing.T) {
		s := &timelineServiceStub{err: domain.ErrLimitOutOfRange}
		w := serve(s, "/users/u1/timeline")
		assertError(t, w, 422, "limit_out_of_range")
	})
}

func serve(s *timelineServiceStub, path string) *httptest.ResponseRecorder {
	h := NewHandler(s).Routes()
	w := httptest.NewRecorder()
	h.ServeHTTP(w, httptest.NewRequest(http.MethodGet, path, nil))
	return w
}
func assertError(t *testing.T, w *httptest.ResponseRecorder, status int, code string) {
	t.Helper()
	if w.Code != status {
		t.Fatalf("status=%d want=%d body=%s", w.Code, status, w.Body.String())
	}
	var got helper.ErrorResponse
	if err := json.NewDecoder(w.Body).Decode(&got); err != nil || got.Error.Code != code {
		t.Fatalf("error=%+v decode=%v", got, err)
	}
}

func TestCursorAndResponse(t *testing.T) {
	position := &domain.CursorPosition{CreatedAt: time.Date(2026, 8, 11, 12, 30, 0, 123, time.FixedZone("ART", -3*3600)), TweetID: "tweet-1"}
	encoded, err := EncodeCursor(position)
	if err != nil || encoded == nil {
		t.Fatalf("encoded=%v err=%v", encoded, err)
	}
	decoded, err := DecodeCursor(*encoded)
	if err != nil || decoded.TweetID != position.TweetID || !decoded.CreatedAt.Equal(position.CreatedAt) {
		t.Fatalf("decoded=%+v err=%v", decoded, err)
	}
	if got, err := EncodeCursor(nil); err != nil || got != nil {
		t.Errorf("nil encode=%v err=%v", got, err)
	}
	if got, err := DecodeCursor(""); !errors.Is(err, ErrInvalidCursor) || got != nil {
		t.Errorf("empty decode=%v err=%v", got, err)
	}
	invalid := []string{"%%%", "e25vdC1qc29u", "eyJjcmVhdGVkX2F0Ijoibm8iLCJ0d2VldF9pZCI6InQifQ", "eyJjcmVhdGVkX2F0IjoiMjAyNi0wOC0xMVQxMjowMDowMFoiLCJ0d2VldF9pZCI6IiJ9"}
	for _, value := range invalid {
		if _, err := DecodeCursor(value); !errors.Is(err, ErrInvalidCursor) {
			t.Errorf("value=%q err=%v", value, err)
		}
	}
	page := domain.TimelinePage{Tweets: []tweet.Tweet{{ID: "t1", CreatedAt: position.CreatedAt}}, NextCursor: position}
	resp, err := NewTimelineResponse(page)
	if err != nil || len(resp.Tweets) != 1 || resp.NextCursor == nil {
		t.Fatalf("response=%+v err=%v", resp, err)
	}
}

func TestTimelineRoutes(t *testing.T) {
	h := NewHandler(&timelineServiceStub{}).Routes()
	for _, tc := range []struct {
		method, path string
		status       int
	}{{http.MethodPost, "/users/u1/timeline", 405}, {http.MethodGet, "/unknown", 404}} {
		w := httptest.NewRecorder()
		h.ServeHTTP(w, httptest.NewRequest(tc.method, tc.path, nil))
		if w.Code != tc.status {
			t.Errorf("status=%d want=%d", w.Code, tc.status)
		}
	}
}
