package timeline

import (
	"context"
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/mespalenza/microblogging-challenge/internal/rest/errorcatalog"
	"github.com/mespalenza/microblogging-challenge/internal/rest/helper"
	domaintimeline "github.com/mespalenza/microblogging-challenge/internal/timeline"
)

const defaultLimit = 20

type TimelineGetter interface {
	GetTimeline(ctx context.Context, input domaintimeline.TimelineInput) (domaintimeline.TimelinePage, error)
}

type Handler struct {
	service TimelineGetter
}

func NewHandler(service TimelineGetter) Handler {
	return Handler{
		service: service,
	}
}

func (h Handler) GetTimeline(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query()

	limit := defaultLimit

	if values, exists := query["limit"]; exists {
		if len(values) != 1 {
			helper.WriteError(w, http.StatusBadRequest, "invalid_pagination_parameters", "pagination parameters have an invalid format")
			return
		}

		parsedLimit, err := strconv.Atoi(values[0])
		if err != nil {
			helper.WriteError(w, http.StatusBadRequest, "invalid_pagination_parameters", "pagination parameters have an invalid format")
			return
		}

		limit = parsedLimit
	}

	var cursor *domaintimeline.CursorPosition

	if values, exists := query["cursor"]; exists {
		if len(values) != 1 {
			helper.WriteError(w, http.StatusBadRequest, "invalid_pagination_parameters", "pagination parameters have an invalid format")
			return
		}

		decodedCursor, err := DecodeCursor(values[0])
		if err != nil {
			helper.WriteError(w, http.StatusBadRequest, "invalid_cursor", "the provided cursor is invalid")
			return
		}

		cursor = decodedCursor
	}

	page, err := h.service.GetTimeline(r.Context(), domaintimeline.TimelineInput{
		UserID: r.PathValue("user_id"),
		Limit:  limit,
		Cursor: cursor,
	},
	)
	if err != nil {
		mappedErr := errorcatalog.From(err)
		helper.WriteError(w, mappedErr.Status, mappedErr.Code, mappedErr.Message)
		return
	}

	response, err := NewTimelineResponse(page)
	if err != nil {
		mappedErr := errorcatalog.From(err)
		helper.WriteError(w, mappedErr.Status, mappedErr.Code, mappedErr.Message)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(response)
}
