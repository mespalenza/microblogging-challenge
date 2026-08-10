package tweet

import (
	"net/http"
)

func (h Handler) Routes() http.Handler {
	router := http.NewServeMux()
	router.HandleFunc("POST /tweets", h.CreateTweet)
	return router
}
