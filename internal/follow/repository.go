package follow

import "context"

type Repository interface {
	SaveFollow(ctx context.Context, value Follow) (bool, error)
}
