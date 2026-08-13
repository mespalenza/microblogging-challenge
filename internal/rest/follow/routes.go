package follow

import "net/http"

func (h Handler) RegisterRoutes(router *http.ServeMux) {
	router.HandleFunc(
		"PUT /users/{follower_id}/following/{followed_id}", h.FollowUser)
}

func (h Handler) Routes() http.Handler {
	router := http.NewServeMux()
	h.RegisterRoutes(router)
	return router
}
