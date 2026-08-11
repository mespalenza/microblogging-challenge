package memory

import (
	"context"
	"errors"
	"testing"

	"github.com/mespalenza/microblogging-challenge/internal/follow"
)

func TestFollowRepository(t *testing.T) {
	r := NewFollowRepository()
	ctx := context.Background()
	ids, err := r.FindFollowedIDs(ctx, "u1")
	if err != nil || len(ids) != 0 {
		t.Fatalf("initial ids=%v err=%v", ids, err)
	}
	for i, relation := range []follow.Follow{
		{FollowerID: "u1", FollowedID: "u2"},
		{FollowerID: "u1", FollowedID: "u3"},
	} {
		created, err := r.SaveFollow(ctx, relation)
		if err != nil || !created {
			t.Fatalf("save %d created=%v err=%v", i, created, err)
		}
	}
	created, err := r.SaveFollow(ctx, follow.Follow{FollowerID: "u1", FollowedID: "u2"})
	if err != nil || created {
		t.Fatalf("duplicate created=%v err=%v", created, err)
	}
	ids, err = r.FindFollowedIDs(ctx, "u1")
	if err != nil || len(ids) != 2 {
		t.Fatalf("ids=%v err=%v", ids, err)
	}
}

func TestFollowRepository_CanceledContext(t *testing.T) {
	repository := NewFollowRepository()
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	created, err := repository.SaveFollow(ctx, follow.Follow{FollowerID: "u1", FollowedID: "u2"})
	if created || !errors.Is(err, context.Canceled) {
		t.Fatalf("SaveFollow() created=%v error=%v", created, err)
	}

	ids, err := repository.FindFollowedIDs(ctx, "u1")
	if ids != nil || !errors.Is(err, context.Canceled) {
		t.Fatalf("FindFollowedIDs() ids=%v error=%v", ids, err)
	}

	ids, err = repository.FindFollowedIDs(context.Background(), "u1")
	if err != nil || len(ids) != 0 {
		t.Fatalf("canceled save changed state: ids=%v error=%v", ids, err)
	}
}
