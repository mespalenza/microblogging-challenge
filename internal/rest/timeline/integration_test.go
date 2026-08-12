package timeline_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"reflect"
	"testing"
	"time"

	domainfollow "github.com/mespalenza/microblogging-challenge/internal/follow"
	"github.com/mespalenza/microblogging-challenge/internal/platform/memory"
	restfollow "github.com/mespalenza/microblogging-challenge/internal/rest/follow"
	resttimeline "github.com/mespalenza/microblogging-challenge/internal/rest/timeline"
	domaintimeline "github.com/mespalenza/microblogging-challenge/internal/timeline"
	"github.com/mespalenza/microblogging-challenge/internal/tweet"
)

func TestTimelineHTTPIntegration(t *testing.T) {
	ctx := context.Background()
	tweetRepository := memory.NewTweetRepository()
	followRepository := memory.NewFollowRepository()

	baseTime := time.Date(2026, 8, 11, 12, 0, 0, 0, time.UTC)
	values := []tweet.Tweet{
		{ID: "t5", UserID: "u2", Content: "newest", CreatedAt: baseTime.Add(3 * time.Minute)},
		{ID: "t4", UserID: "u3", Content: "tie higher ID", CreatedAt: baseTime.Add(2 * time.Minute)},
		{ID: "t3", UserID: "u2", Content: "tie lower ID", CreatedAt: baseTime.Add(2 * time.Minute)},
		{ID: "t2", UserID: "u3", Content: "older", CreatedAt: baseTime.Add(time.Minute)},
		{ID: "t1", UserID: "u2", Content: "oldest", CreatedAt: baseTime},
		{ID: "own", UserID: "u1", Content: "must be excluded", CreatedAt: baseTime.Add(time.Hour)},
		{ID: "unrelated", UserID: "u4", Content: "must be excluded", CreatedAt: baseTime.Add(2 * time.Hour)},
	}
	for _, value := range values {
		if err := tweetRepository.Save(ctx, value); err != nil {
			t.Fatalf("Save(%s): %v", value.ID, err)
		}
	}

	// The tweets deliberately exist before the follow requests.
	router := http.NewServeMux()
	restfollow.NewHandler(domainfollow.NewService(followRepository)).RegisterRoutes(router)
	resttimeline.NewHandler(domaintimeline.NewService(tweetRepository, followRepository)).RegisterRoutes(router)

	for _, followedID := range []string{"u2", "u3"} {
		request := httptest.NewRequest(http.MethodPut, "/users/u1/following/"+followedID, nil)
		response := httptest.NewRecorder()
		router.ServeHTTP(response, request)
		if response.Code != http.StatusCreated {
			t.Fatalf("follow %s status=%d body=%s", followedID, response.Code, response.Body.String())
		}
	}

	wantPages := [][]string{{"t5", "t4"}, {"t3", "t2"}, {"t1"}}
	cursor := ""
	seen := make([]string, 0, 5)
	for index, wantIDs := range wantPages {
		path := "/users/u1/timeline?limit=2"
		if cursor != "" {
			path += "&cursor=" + url.QueryEscape(cursor)
		}

		response := httptest.NewRecorder()
		router.ServeHTTP(response, httptest.NewRequest(http.MethodGet, path, nil))
		if response.Code != http.StatusOK {
			t.Fatalf("page %d status=%d body=%s", index+1, response.Code, response.Body.String())
		}

		var page resttimeline.TimelineResponse
		if err := json.NewDecoder(response.Body).Decode(&page); err != nil {
			t.Fatalf("page %d decode: %v", index+1, err)
		}
		gotIDs := make([]string, len(page.Tweets))
		for i := range page.Tweets {
			gotIDs[i] = page.Tweets[i].ID
		}
		if !reflect.DeepEqual(gotIDs, wantIDs) {
			t.Fatalf("page %d ids=%v want=%v", index+1, gotIDs, wantIDs)
		}
		seen = append(seen, gotIDs...)

		if index < len(wantPages)-1 {
			if page.NextCursor == nil {
				t.Fatalf("page %d next_cursor=nil", index+1)
			}
			cursor = *page.NextCursor
		} else if page.NextCursor != nil {
			t.Fatalf("last page next_cursor=%q", *page.NextCursor)
		}
	}

	if !reflect.DeepEqual(seen, []string{"t5", "t4", "t3", "t2", "t1"}) {
		t.Fatalf("paginated ids=%v", seen)
	}
}

func TestTimelineHTTPIntegrationEmptyPage(t *testing.T) {
	tweetRepository := memory.NewTweetRepository()
	followRepository := memory.NewFollowRepository()
	router := resttimeline.NewHandler(domaintimeline.NewService(tweetRepository, followRepository)).Routes()

	response := httptest.NewRecorder()
	router.ServeHTTP(response, httptest.NewRequest(http.MethodGet, "/users/u1/timeline", nil))
	if response.Code != http.StatusOK {
		t.Fatalf("status=%d body=%s", response.Code, response.Body.String())
	}

	var body struct {
		Tweets     []json.RawMessage `json:"tweets"`
		NextCursor *string           `json:"next_cursor"`
	}
	if err := json.NewDecoder(response.Body).Decode(&body); err != nil {
		t.Fatal(err)
	}
	if body.Tweets == nil || len(body.Tweets) != 0 || body.NextCursor != nil {
		t.Fatalf("tweets=%v next_cursor=%v body=%s", body.Tweets, body.NextCursor, response.Body.String())
	}
}
