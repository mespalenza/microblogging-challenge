package errorcatalog

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"testing"

	"github.com/mespalenza/microblogging-challenge/internal/tweet"
)

func TestFrom(t *testing.T) {
	tests := []struct {
		name string
		err  error
		want Error
	}{
		{name: "invalid user ID", err: tweet.ErrInvalidUserID, want: Error{http.StatusBadRequest, "invalid_user_id", "user_id must not be empty"}},
		{name: "wrapped invalid content", err: fmt.Errorf("validate: %w", tweet.ErrInvalidContent), want: Error{http.StatusBadRequest, "invalid_content", "content must not be empty"}},
		{name: "content too long", err: tweet.ErrContentTooLong, want: Error{http.StatusBadRequest, "content_too_long", "content must not exceed 280 characters"}},
		{name: "unexpected error", err: errors.New("database unavailable"), want: Error{http.StatusInternalServerError, "internal_error", "an unexpected error occurred"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := From(tt.err); got != tt.want {
				t.Errorf("From() = %+v, want %+v", got, tt.want)
			}
		})
	}
}

func TestFromDecode(t *testing.T) {
	tests := []struct {
		name string
		err  error
		want Error
	}{
		{name: "invalid type", err: &json.UnmarshalTypeError{Value: "number", Type: nil}, want: Error{http.StatusBadRequest, "invalid_field_type", "a field has an invalid type"}},
		{name: "unknown field", err: errors.New(`json: unknown field "extra"`), want: Error{http.StatusBadRequest, "unknown_field", "request contains an unknown field"}},
		{name: "invalid JSON", err: errors.New("unexpected EOF"), want: Error{http.StatusBadRequest, "invalid_json", "request body contains invalid JSON"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := FromDecode(tt.err); got != tt.want {
				t.Errorf("FromDecode() = %+v, want %+v", got, tt.want)
			}
		})
	}
}
