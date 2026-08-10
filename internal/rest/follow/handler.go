package follow

import (
	"context"
	"net/http"

	"github.com/mespalenza/microblogging-challenge/internal/follow"
	"github.com/mespalenza/microblogging-challenge/internal/rest/errorcatalog"
	"github.com/mespalenza/microblogging-challenge/internal/rest/helper"
)

type FollowUser interface {
	FollowUser(ctx context.Context, value follow.Follow) (bool, error)
}
type Handler struct {
	service FollowUser
}

func NewHandler(service FollowUser) Handler {
	return Handler{
		service: service,
	}
}

func (h Handler) FollowUser(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	followerID := r.PathValue("follower_id")
	followedID := r.PathValue("followed_id")

	relation := follow.Follow{
		FollowerID: followerID,
		FollowedID: followedID,
	}

	created, err := h.service.FollowUser(ctx, relation)
	if err != nil {
		mappedErr := errorcatalog.From(err)
		helper.WriteError(w, mappedErr.Status, mappedErr.Code, mappedErr.Message)
		return
	}

	if !created {
		w.WriteHeader(http.StatusNoContent)
		return
	}

	w.WriteHeader(http.StatusCreated)
}
