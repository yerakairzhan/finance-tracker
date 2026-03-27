package cache

import (
	"context"
	"time"

	"github.com/redis/go-redis/v9"
)

type TokenBlocklist struct {
	client *redis.Client
}

func NewRedisClient(addr, password string, db int) *redis.Client {
	return redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: password,
		DB:       db,
	})
}

func NewTokenBlocklist(client *redis.Client) *TokenBlocklist {
	if client == nil {
		return nil
	}
	return &TokenBlocklist{client: client}
}

func (b *TokenBlocklist) Revoke(ctx context.Context, tokenID string, ttl time.Duration) error {
	if b == nil || b.client == nil {
		return nil
	}
	return b.client.Set(ctx, buildRevokedKey(tokenID), "1", ttl).Err()
}

func (b *TokenBlocklist) IsRevoked(ctx context.Context, tokenID string) (bool, error) {
	if b == nil || b.client == nil {
		return false, nil
	}
	exists, err := b.client.Exists(ctx, buildRevokedKey(tokenID)).Result()
	if err != nil {
		return false, err
	}
	return exists > 0, nil
}

func buildRevokedKey(tokenID string) string {
	return "revoked_access_token:" + tokenID
}
