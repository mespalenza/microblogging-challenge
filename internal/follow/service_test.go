package follow

import (
	"context"
	"errors"
	"testing"
)

type repositoryStub struct {
	created bool
	err     error
	value   Follow
	calls   int
}

func (r *repositoryStub) SaveFollow(_ context.Context, value Follow) (bool, error) {
	r.calls++
	r.value = value
	return r.created, r.err
}

func TestServiceFollowUser(t *testing.T) {
	relation := Follow{FollowerID: "user-1", FollowedID: "user-2"}

	t.Run("creates relation", func(t *testing.T) {
		repo := &repositoryStub{created: true}
		created, err := NewService(repo).FollowUser(context.Background(), relation)
		if err != nil || !created || repo.value != relation || repo.calls != 1 {
			t.Fatalf("created=%v err=%v value=%+v calls=%d", created, err, repo.value, repo.calls)
		}
	})

	t.Run("existing relation", func(t *testing.T) {
		repo := &repositoryStub{}
		created, err := NewService(repo).FollowUser(context.Background(), relation)
		if err != nil || created {
			t.Fatalf("created=%v err=%v", created, err)
		}
	})

	t.Run("cannot follow self", func(t *testing.T) {
		repo := &repositoryStub{}
		created, err := NewService(repo).FollowUser(context.Background(), Follow{FollowerID: "user-1", FollowedID: "user-1"})
		if !errors.Is(err, ErrCannotFollowSelf) || created || repo.calls != 0 {
			t.Fatalf("created=%v err=%v calls=%d", created, err, repo.calls)
		}
	})

	t.Run("repository error", func(t *testing.T) {
		wantErr := errors.New("storage unavailable")
		repo := &repositoryStub{err: wantErr}
		created, err := NewService(repo).FollowUser(context.Background(), relation)
		if !errors.Is(err, wantErr) || created {
			t.Fatalf("created=%v err=%v", created, err)
		}
	})
}
