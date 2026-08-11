package timeline

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"time"

	domaintimeline "github.com/mespalenza/microblogging-challenge/internal/timeline"
)

var ErrInvalidCursor = errors.New("invalid cursor")

type cursorPayload struct {
	CreatedAt string `json:"created_at"`
	TweetID   string `json:"tweet_id"`
}

func EncodeCursor(position *domaintimeline.CursorPosition) (*string, error) {
	if position == nil {
		return nil, nil
	}

	payload := cursorPayload{
		CreatedAt: position.CreatedAt.UTC().Format(time.RFC3339Nano),
		TweetID:   position.TweetID,
	}

	data, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}

	encoded := base64.RawURLEncoding.EncodeToString(data)

	return &encoded, nil
}

func DecodeCursor(value string) (*domaintimeline.CursorPosition, error) {
	if value == "" {
		return nil, ErrInvalidCursor
	}

	data, err := base64.RawURLEncoding.DecodeString(value)
	if err != nil {
		return nil, ErrInvalidCursor
	}

	var payload cursorPayload
	if err := json.Unmarshal(data, &payload); err != nil {
		return nil, ErrInvalidCursor
	}

	createdAt, err := time.Parse(time.RFC3339Nano, payload.CreatedAt)
	if err != nil || payload.TweetID == "" {
		return nil, ErrInvalidCursor
	}

	return &domaintimeline.CursorPosition{
		CreatedAt: createdAt,
		TweetID:   payload.TweetID,
	}, nil
}
