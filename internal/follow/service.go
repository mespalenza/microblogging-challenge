package follow

import "context"

type Service struct {
	repository Repository
}

func NewService(repository Repository) *Service {
	return &Service{
		repository: repository,
	}
}

func (s *Service) FollowUser(ctx context.Context, value Follow) (bool, error) {
	if value.FollowerID == value.FollowedID {
		return false, ErrCannotFollowSelf
	}

	created, err := s.repository.SaveFollow(ctx, value)
	if err != nil {
		return false, err
	}

	return created, nil
}
