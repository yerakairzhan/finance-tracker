package cache

import (
	"context"
	"testing"
	"time"

	miniredis "github.com/alicebob/miniredis/v2"
	"github.com/redis/go-redis/v9"
)

func TestRefreshSessionCRUD(t *testing.T) {
	mr := miniredis.RunT(t)
	client := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	store := NewRefreshSessionStore(client)

	ctx := context.Background()
	if err := store.CreateRefreshSession(ctx, "h1", 42, time.Hour); err != nil {
		t.Fatal(err)
	}
	s, err := store.GetRefreshSession(ctx, "h1")
	if err != nil || s == nil || s.UserID != 42 {
		t.Fatalf("unexpected session: %#v err=%v", s, err)
	}
	if err := store.DeleteRefreshSession(ctx, "h1"); err != nil {
		t.Fatal(err)
	}
	s, err = store.GetRefreshSession(ctx, "h1")
	if err != nil || s != nil {
		t.Fatalf("expected deleted session, got %#v err=%v", s, err)
	}
}

func TestRefreshRotationReplayProtection(t *testing.T) {
	mr := miniredis.RunT(t)
	client := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	store := NewRefreshSessionStore(client)
	ctx := context.Background()

	if err := store.CreateRefreshSession(ctx, "old", 7, time.Hour); err != nil {
		t.Fatal(err)
	}
	if err := store.RotateRefreshSession(ctx, "old", "new", 7, time.Hour); err != nil {
		t.Fatal(err)
	}
	oldS, _ := store.GetRefreshSession(ctx, "old")
	newS, _ := store.GetRefreshSession(ctx, "new")
	if oldS != nil || newS == nil {
		t.Fatalf("rotation failed old=%#v new=%#v", oldS, newS)
	}

	// replay attempt (old token) must fail
	if err := store.RotateRefreshSession(ctx, "old", "new2", 7, time.Hour); err == nil {
		t.Fatal("expected replay rotation failure")
	}
}
