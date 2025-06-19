package storage

import (
	"context"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupTestRedis(t *testing.T) *RedisStore {
	// For local testing, we'll use a Redis instance on default port
	// In CI, this would be replaced with a test container
	store := NewRedisStore("localhost:6379", "", 0)

	// Clear the test database
	err := store.client.FlushDB(context.Background()).Err()
	require.NoError(t, err)

	return store
}

func TestRedisStore_Set(t *testing.T) {
	store := setupTestRedis(t)
	defer store.Close()
	ctx := context.Background()

	// Test successful set
	err := store.Set(ctx, "test1", "http://example.com")
	assert.NoError(t, err)

	// Verify TTL was set
	ttl, err := store.client.TTL(ctx, "test1").Result()
	assert.NoError(t, err)
	assert.True(t, ttl > 0 && ttl <= DefaultTTL)

	// Test duplicate key
	err = store.Set(ctx, "test1", "http://another.com")
	assert.Equal(t, ErrKeyExists, err)

	// Test empty key
	err = store.Set(ctx, "", "http://example.com")
	assert.Error(t, err)

	// Test empty URL
	err = store.Set(ctx, "test2", "")
	assert.Error(t, err)
}

func TestRedisStore_Get(t *testing.T) {
	store := setupTestRedis(t)
	defer store.Close()
	ctx := context.Background()

	// Set up test data
	err := store.Set(ctx, "test1", "http://example.com")
	require.NoError(t, err)

	// Test successful get
	url, err := store.Get(ctx, "test1")
	assert.NoError(t, err)
	assert.Equal(t, "http://example.com", url)

	// Test non-existent key
	_, err = store.Get(ctx, "nonexistent")
	assert.Equal(t, ErrNotFound, err)

	// Test empty key
	_, err = store.Get(ctx, "")
	assert.Error(t, err)

	// Test TTL refresh on get
	time.Sleep(time.Second) // Wait a bit to see TTL change

	// Get the current TTL
	originalTTL, err := store.client.TTL(ctx, "test1").Result()
	require.NoError(t, err)

	// Get the URL again to refresh TTL
	_, err = store.Get(ctx, "test1")
	require.NoError(t, err)

	// Check that TTL was refreshed
	newTTL, err := store.client.TTL(ctx, "test1").Result()
	require.NoError(t, err)
	assert.True(t, newTTL > originalTTL, "Expected TTL to be refreshed")
}

func TestRedisStore_Delete(t *testing.T) {
	store := setupTestRedis(t)
	defer store.Close()
	ctx := context.Background()

	// Set up test data
	err := store.Set(ctx, "test1", "http://example.com")
	require.NoError(t, err)

	// Test successful delete
	err = store.Delete(ctx, "test1")
	assert.NoError(t, err)

	// Verify key was deleted
	exists, err := store.client.Exists(ctx, "test1").Result()
	assert.NoError(t, err)
	assert.Equal(t, int64(0), exists)

	// Test delete non-existent key
	err = store.Delete(ctx, "nonexistent")
	assert.Equal(t, ErrNotFound, err)

	// Test delete empty key
	err = store.Delete(ctx, "")
	assert.Error(t, err)
}

func TestRedisStore_ConnectionFailure(t *testing.T) {
	// Try to connect to a non-existent Redis server
	store := NewRedisStore("localhost:6380", "", 0)
	defer store.Close()

	ctx := context.Background()

	// Test operations with bad connection
	err := store.Set(ctx, "test", "http://example.com")
	assert.Error(t, err)

	_, err = store.Get(ctx, "test")
	assert.Error(t, err)

	err = store.Delete(ctx, "test")
	assert.Error(t, err)
}

func TestRedisStore_Concurrent(t *testing.T) {
	store := setupTestRedis(t)
	defer store.Close()
	ctx := context.Background()

	// Number of concurrent operations
	n := 100
	var wg sync.WaitGroup
	wg.Add(n)

	// Channel to collect errors
	errCh := make(chan error, n)

	// Run concurrent Set operations
	for i := 0; i < n; i++ {
		go func(i int) {
			defer wg.Done()
			key := fmt.Sprintf("concurrent-%d", i)
			url := fmt.Sprintf("http://example.com/%d", i)

			if err := store.Set(ctx, key, url); err != nil {
				errCh <- fmt.Errorf("failed to set key %s: %v", key, err)
				return
			}

			// Verify the value was stored correctly
			storedURL, err := store.Get(ctx, key)
			if err != nil {
				errCh <- fmt.Errorf("failed to get key %s: %v", key, err)
				return
			}
			if storedURL != url {
				errCh <- fmt.Errorf("key %s: expected URL %s, got %s", key, url, storedURL)
				return
			}
		}(i)
	}

	// Wait for all operations to complete
	wg.Wait()
	close(errCh)

	// Check for any errors
	for err := range errCh {
		t.Error(err)
	}
}

func TestRedisStore_TTLExpiration(t *testing.T) {
	store := setupTestRedis(t)
	defer store.Close()
	ctx := context.Background()

	// Set a key with a very short TTL
	store.ttl = 1 * time.Second
	err := store.Set(ctx, "expiring", "http://example.com")
	require.NoError(t, err)

	// Verify the key exists
	url, err := store.Get(ctx, "expiring")
	assert.NoError(t, err)
	assert.Equal(t, "http://example.com", url)

	// Wait for the key to expire
	time.Sleep(2 * time.Second)

	// Verify the key has expired
	_, err = store.Get(ctx, "expiring")
	assert.Equal(t, ErrNotFound, err)
}
