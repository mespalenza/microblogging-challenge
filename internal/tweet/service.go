package tweet

import (
	"context"
	"strings"
	"time"

	"github.com/google/uuid"
)

type Service struct {
	repository Repository
}

func NewService(repository Repository) *Service {
	return &Service{
		repository: repository,
	}
}

func (s *Service) CreateTweet(ctx context.Context, request CreateInput) (Tweet, error) {
	userID := strings.TrimSpace(request.UserID)
	if userID == "" {
		return Tweet{}, ErrInvalidUserID
	}

	content := strings.TrimSpace(request.Content)
	if content == "" {
		return Tweet{}, ErrInvalidContent
	}

	if len([]rune(content)) > 280 {
		return Tweet{}, ErrContentTooLong
	}

	tweet := Tweet{
		ID:        uuid.NewString(),
		UserID:    userID,
		Content:   content,
		CreatedAt: time.Now().UTC(),
	}

	if err := s.repository.Save(ctx, tweet); err != nil {
		return Tweet{}, err
	}

	return tweet, nil
}
