package tweet

import "net/http"

func (h Handler) RegisterRoutes(router *http.ServeMux) {
	router.HandleFunc("POST /tweets", h.CreateTweet)
}

func (h Handler) Routes() http.Handler {
	router := http.NewServeMux()
	h.RegisterRoutes(router)
	return router
}
