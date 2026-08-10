package errorcatalog

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"

	"github.com/mespalenza/microblogging-challenge/internal/follow"
	"github.com/mespalenza/microblogging-challenge/internal/tweet"
)

type Error struct {
	Status  int
	Code    string
	Message string
}

// From errores del servicio y dominio
func From(err error) Error {
	switch {
	case errors.Is(err, tweet.ErrInvalidUserID):
		return Error{
			Status:  http.StatusBadRequest,
			Code:    "invalid_user_id",
			Message: "user_id must not be empty",
		}

	case errors.Is(err, tweet.ErrInvalidContent):
		return Error{
			Status:  http.StatusBadRequest,
			Code:    "invalid_content",
			Message: "content must not be empty",
		}

	case errors.Is(err, tweet.ErrContentTooLong):
		return Error{
			Status:  http.StatusBadRequest,
			Code:    "content_too_long",
			Message: "content must not exceed 280 characters",
		}

	case errors.Is(err, follow.ErrCannotFollowSelf):
		return Error{
			Status:  http.StatusUnprocessableEntity,
			Code:    "cannot_follow_self",
			Message: "a user cannot follow themselves",
		}

	default:
		return Error{
			Status:  http.StatusInternalServerError,
			Code:    "internal_error",
			Message: "an unexpected error occurred",
		}
	}
}

// FromDecode errores generados al leer el JSON
func FromDecode(err error) Error {
	var typeError *json.UnmarshalTypeError

	switch {
	case errors.As(err, &typeError):
		return Error{
			Status:  http.StatusBadRequest,
			Code:    "invalid_field_type",
			Message: "a field has an invalid type",
		}

	case strings.HasPrefix(err.Error(), "json: unknown field "):
		return Error{
			Status:  http.StatusBadRequest,
			Code:    "unknown_field",
			Message: "request contains an unknown field",
		}

	default:
		return Error{
			Status:  http.StatusBadRequest,
			Code:    "invalid_json",
			Message: "request body contains invalid JSON",
		}
	}
}
