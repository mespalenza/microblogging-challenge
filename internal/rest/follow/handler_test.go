package follow

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	domainfollow "github.com/mespalenza/microblogging-challenge/internal/follow"
)

type serviceStub struct {
	created bool
	err     error
	value   domainfollow.Follow
	calls   int
}

func (s *serviceStub) FollowUser(_ context.Context, value domainfollow.Follow) (bool, error) {
	s.calls++
	s.value = value
	return s.created, s.err
}

func TestFollowUser(t *testing.T) {
	tests := []struct {
		name    string
		created bool
		err     error
		status  int
	}{
		{"created", true, nil, http.StatusCreated}, {"existing", false, nil, http.StatusNoContent},
		{"self", false, domainfollow.ErrCannotFollowSelf, http.StatusUnprocessableEntity}, {"failure", false, errors.New("boom"), http.StatusInternalServerError},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &serviceStub{created: tt.created, err: tt.err}
			h := NewHandler(s).Routes()
			r := httptest.NewRequest(http.MethodPut, "/users/u1/following/u2", nil)
			w := httptest.NewRecorder()
			h.ServeHTTP(w, r)
			if w.Code != tt.status {
				t.Fatalf("status=%d want=%d body=%s", w.Code, tt.status, w.Body.String())
			}
			if s.value != (domainfollow.Follow{FollowerID: "u1", FollowedID: "u2"}) {
				t.Errorf("value=%+v", s.value)
			}
		})
	}
}

func TestRoutes(t *testing.T) {
	h := NewHandler(&serviceStub{}).Routes()
	for _, tc := range []struct {
		method, path string
		status       int
	}{{http.MethodGet, "/users/u1/following/u2", 405}, {http.MethodPut, "/unknown", 404}} {
		w := httptest.NewRecorder()
		h.ServeHTTP(w, httptest.NewRequest(tc.method, tc.path, nil))
		if w.Code != tc.status {
			t.Errorf("%s %s status=%d", tc.method, tc.path, w.Code)
		}
	}
}
