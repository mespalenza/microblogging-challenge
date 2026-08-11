package timeline

import "net/http"

func (h Handler) RegisterRoutes(router *http.ServeMux) {
	router.HandleFunc("GET /users/{user_id}/timeline", h.GetTimeline)
}

func (h Handler) Routes() http.Handler {
	router := http.NewServeMux()
	h.RegisterRoutes(router)
	return router
}
