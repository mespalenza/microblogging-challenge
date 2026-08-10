package tweet

import "context"

type Repository interface {
	Save(context.Context, Tweet) error
}
