package atomic

import (
	"context"
	"fmt"
	"time"

	"github.com/cenkalti/backoff/v4"
	"github.com/redis/go-redis/v9"
)

const (
	AtomicPrefixKey = "ATOMIC_LOCK"
)

type atomic struct {
	Redis     redis.UniversalClient
	AtomicTTL time.Duration
}

type Atomic interface {
	Lock(ctx context.Context, key string) error
	Release(ctx context.Context, key string) error
	LockWithRetry(ctx context.Context, key string, maxElapsedTime time.Duration) error
}

func NewAtomic(redis redis.UniversalClient, atomicTTL time.Duration) Atomic {
	return &atomic{
		Redis:     redis,
		AtomicTTL: atomicTTL,
	}
}

// Lock พยายาม acquire lock ครั้งเดียว
func (a *atomic) Lock(ctx context.Context, key string) error {
	aKey := fmt.Sprintf("%s:%s", AtomicPrefixKey, key)

	r, err := a.Redis.SetNX(ctx, aKey, "LOCK", a.AtomicTTL).Result()
	if err != nil {
		return fmt.Errorf("fail to lock resource: %w", err)
	}

	if !r {
		return fmt.Errorf("resource already lock")
	}

	return nil
}

// LockWithRetry พยายาม acquire lock ด้วย exponential backoff
func (a *atomic) LockWithRetry(ctx context.Context, key string, maxElapsedTime time.Duration) error {

	// สร้าง exponential backoff config
	expBackoff := backoff.NewExponentialBackOff()
	expBackoff.InitialInterval = 50 * time.Millisecond
	expBackoff.MaxInterval = 1 * time.Second
	expBackoff.MaxElapsedTime = maxElapsedTime
	expBackoff.Multiplier = 2
	expBackoff.RandomizationFactor = 0.5

	// Wrap with context
	b := backoff.WithContext(expBackoff, ctx)

	operation := func() error {

		err := a.Lock(ctx, key)
		if err != nil {
			return err
		}

		return nil
	}

	err := backoff.Retry(operation, b)
	if err != nil {
		return fmt.Errorf("failed to acquire lock after retries: %w", err)
	}

	return nil
}

// Release ปลดปล่อย lock
func (a *atomic) Release(ctx context.Context, key string) error {

	aKey := fmt.Sprintf("%s:%s", AtomicPrefixKey, key)
	return a.Redis.Del(ctx, aKey).Err()
}
