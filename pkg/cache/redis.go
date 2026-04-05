package cache

import (
	"context"
	"encoding/json"
	"errors"
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

type RefreshSession struct {
	UserID      int64     `json:"user_id"`
	CreatedAt   time.Time `json:"created_at"`
	ExpiresAt   time.Time `json:"expires_at"`
	RotatedFrom string    `json:"rotated_from,omitempty"`
}

type RefreshSessionStore struct {
	client *redis.Client
}

func NewRefreshSessionStore(client *redis.Client) *RefreshSessionStore {
	if client == nil {
		return nil
	}
	return &RefreshSessionStore{client: client}
}

func (s *RefreshSessionStore) CreateRefreshSession(ctx context.Context, tokenHash string, userID int64, ttl time.Duration) error {
	if s == nil || s.client == nil {
		return nil
	}
	if tokenHash == "" || userID <= 0 || ttl <= 0 {
		return errors.New("invalid refresh session input")
	}
	now := time.Now().UTC()
	payload, err := json.Marshal(RefreshSession{
		UserID:    userID,
		CreatedAt: now,
		ExpiresAt: now.Add(ttl),
	})
	if err != nil {
		return err
	}
	return s.client.Set(ctx, buildRefreshKey(tokenHash), payload, ttl).Err()
}

func (s *RefreshSessionStore) GetRefreshSession(ctx context.Context, tokenHash string) (*RefreshSession, error) {
	if s == nil || s.client == nil {
		return nil, nil
	}
	raw, err := s.client.Get(ctx, buildRefreshKey(tokenHash)).Bytes()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return nil, nil
		}
		return nil, err
	}
	var session RefreshSession
	if err := json.Unmarshal(raw, &session); err != nil {
		return nil, err
	}
	return &session, nil
}

func (s *RefreshSessionStore) DeleteRefreshSession(ctx context.Context, tokenHash string) error {
	if s == nil || s.client == nil || tokenHash == "" {
		return nil
	}
	return s.client.Del(ctx, buildRefreshKey(tokenHash)).Err()
}

func (s *RefreshSessionStore) RotateRefreshSession(ctx context.Context, oldTokenHash, newTokenHash string, userID int64, ttl time.Duration) error {
	if s == nil || s.client == nil {
		return nil
	}
	if oldTokenHash == "" || newTokenHash == "" || oldTokenHash == newTokenHash || userID <= 0 || ttl <= 0 {
		return errors.New("invalid refresh rotation input")
	}
	now := time.Now().UTC()
	payload, err := json.Marshal(RefreshSession{
		UserID:      userID,
		CreatedAt:   now,
		ExpiresAt:   now.Add(ttl),
		RotatedFrom: oldTokenHash,
	})
	if err != nil {
		return err
	}

	script := redis.NewScript(`
		local oldKey = KEYS[1]
		local newKey = KEYS[2]
		local payload = ARGV[1]
		local ttl = tonumber(ARGV[2])
		if redis.call("EXISTS", oldKey) == 0 then
			return 0
		end
		redis.call("DEL", oldKey)
		redis.call("SET", newKey, payload, "EX", ttl)
		return 1
	`)
	res, err := script.Run(ctx, s.client, []string{
		buildRefreshKey(oldTokenHash),
		buildRefreshKey(newTokenHash),
	}, payload, int(ttl.Seconds())).Int()
	if err != nil {
		return err
	}
	if res != 1 {
		return errors.New("refresh session not found or already rotated")
	}
	return nil
}

func buildRefreshKey(tokenHash string) string {
	return "auth:refresh:" + tokenHash
}
