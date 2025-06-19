package storage

import (
	"context"
	"errors"
	"time"

	"github.com/redis/go-redis/v9"
)

const (
	// DefaultTTL is the default time-to-live for URL mappings (3 hours)
	DefaultTTL = 3 * time.Hour
)

// Error types for storage operations
var (
	ErrNotFound  = errors.New("url mapping not found")
	ErrKeyExists = errors.New("key already exists")
)

// Store represents the storage interface for URL mappings
type Store interface {
	Set(ctx context.Context, key, url string) error
	Get(ctx context.Context, key string) (string, error)
	Delete(ctx context.Context, key string) error
}

// RedisStore implements the Store interface using Redis
type RedisStore struct {
	client *redis.Client
	ttl    time.Duration
}

// NewRedisStore creates a new RedisStore instance
func NewRedisStore(addr, password string, db int) *RedisStore {
	client := redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: password,
		DB:       db,
	})

	return &RedisStore{
		client: client,
		ttl:    DefaultTTL,
	}
}

// Set stores a URL mapping with the specified key
func (s *RedisStore) Set(ctx context.Context, key, url string) error {
	// Validate inputs
	if key == "" {
		return errors.New("key cannot be empty")
	}
	if url == "" {
		return errors.New("url cannot be empty")
	}

	// Try to set the key only if it doesn't exist
	success, err := s.client.SetNX(ctx, key, url, s.ttl).Result()
	if err != nil {
		return err
	}
	if !success {
		return ErrKeyExists
	}
	return nil
}

// Get retrieves a URL mapping by key
func (s *RedisStore) Get(ctx context.Context, key string) (string, error) {
	url, err := s.client.Get(ctx, key).Result()
	if err == redis.Nil {
		return "", ErrNotFound
	}
	if err != nil {
		return "", err
	}

	// Refresh TTL on access
	if err := s.client.Expire(ctx, key, s.ttl).Err(); err != nil {
		// Log warning but don't fail the get operation
		// TODO: Add proper logging
		_ = err
	}

	return url, nil
}

// Delete removes a URL mapping
func (s *RedisStore) Delete(ctx context.Context, key string) error {
	result, err := s.client.Del(ctx, key).Result()
	if err != nil {
		return err
	}
	if result == 0 {
		return ErrNotFound
	}
	return nil
}

// Close closes the Redis connection
func (s *RedisStore) Close() error {
	return s.client.Close()
}

// FlushDB clears all data in the Redis database (for testing only)
func (s *RedisStore) FlushDB(ctx context.Context) error {
	return s.client.FlushDB(ctx).Err()
}
