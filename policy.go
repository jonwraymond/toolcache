package toolcache

import "time"

// Policy defines caching behavior including TTL defaults and limits.
type Policy struct {
	// DefaultTTL is the TTL used when no override is provided.
	// A value of 0 disables caching by default.
	DefaultTTL time.Duration

	// MaxTTL is the maximum allowed TTL. If an effective TTL exceeds this,
	// it is clamped to MaxTTL. A value of 0 means no maximum.
	MaxTTL time.Duration

	// AllowUnsafe permits caching of results from tools marked as unsafe.
	// Default is false.
	AllowUnsafe bool
}

// EffectiveTTL computes the TTL to use given an optional override.
//
// Resolution rules:
//  1. If override > 0, use override
//  2. If override <= 0, use DefaultTTL
//  3. If MaxTTL > 0 and effective TTL > MaxTTL, clamp to MaxTTL
//  4. If effective TTL <= 0, return 0 (no caching)
func (p Policy) EffectiveTTL(override time.Duration) time.Duration {
	ttl := p.DefaultTTL
	if override > 0 {
		ttl = override
	}

	// Clamp to MaxTTL if set
	if p.MaxTTL > 0 && ttl > p.MaxTTL {
		ttl = p.MaxTTL
	}

	// Negative TTL means no caching
	if ttl < 0 {
		return 0
	}

	return ttl
}

// ShouldCache reports whether caching is enabled by default.
// Returns true if DefaultTTL > 0.
func (p Policy) ShouldCache() bool {
	return p.DefaultTTL > 0
}

// DefaultPolicy returns a Policy with sensible defaults:
//   - DefaultTTL: 5 minutes
//   - MaxTTL: 1 hour
//   - AllowUnsafe: false
func DefaultPolicy() Policy {
	return Policy{
		DefaultTTL:  5 * time.Minute,
		MaxTTL:      1 * time.Hour,
		AllowUnsafe: false,
	}
}

// NoCachePolicy returns a Policy that disables caching entirely.
func NoCachePolicy() Policy {
	return Policy{
		DefaultTTL:  0,
		MaxTTL:      0,
		AllowUnsafe: false,
	}
}
