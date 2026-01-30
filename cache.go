package toolcache

import (
	"context"
	"errors"
	"strings"
	"time"
)

var (
	ErrNilCache   = errors.New("toolcache: cache is nil")
	ErrInvalidKey = errors.New("toolcache: key is invalid")
	ErrKeyTooLong = errors.New("toolcache: key exceeds max length")
)

const MaxKeyLength = 512

// Cache stores tool outputs with TTL semantics.
//
// Contract:
// - Concurrency: implementations must be safe for concurrent use.
// - Context: methods must honor cancellation/deadlines and return ctx.Err() when canceled.
// - Errors: invalid keys should return ErrInvalidKey/ErrKeyTooLong via ValidateKey.
// - Ownership: returned bytes are caller-owned; implementations should copy on write.
type Cache interface {
	Get(ctx context.Context, key string) (value []byte, ok bool)
	Set(ctx context.Context, key string, value []byte, ttl time.Duration) error
	Delete(ctx context.Context, key string) error
}

func ValidateKey(key string) error {
	if len(key) == 0 || len(strings.TrimSpace(key)) == 0 {
		return ErrInvalidKey
	}
	if len(key) > MaxKeyLength {
		return ErrKeyTooLong
	}
	if strings.ContainsAny(key, "\n\r") {
		return ErrInvalidKey
	}
	return nil
}
