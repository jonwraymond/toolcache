package toolcache

import (
	"context"
	"strings"
)

type SkipRule func(toolID string, tags []string) bool

type ToolExecutor func(ctx context.Context, toolID string, input any) ([]byte, error)

var DefaultUnsafeTags = []string{"write", "danger", "unsafe", "mutation", "delete"}

func DefaultSkipRule(_ string, tags []string) bool {
	for _, tag := range tags {
		tagLower := strings.ToLower(tag)
		for _, unsafeTag := range DefaultUnsafeTags {
			if tagLower == unsafeTag {
				return true
			}
		}
	}
	return false
}

type CacheMiddleware struct {
	cache    Cache
	keyer    Keyer
	policy   Policy
	skipRule SkipRule
}

func NewCacheMiddleware(cache Cache, keyer Keyer, policy Policy, skipRule SkipRule) *CacheMiddleware {
	return &CacheMiddleware{
		cache:    cache,
		keyer:    keyer,
		policy:   policy,
		skipRule: skipRule,
	}
}

func (m *CacheMiddleware) Execute(ctx context.Context, toolID string, input any, tags []string, executor ToolExecutor) ([]byte, error) {
	if m.shouldSkip(toolID, tags) {
		return executor(ctx, toolID, input)
	}

	key, err := m.keyer.Key(toolID, input)
	if err != nil {
		return executor(ctx, toolID, input)
	}

	if cached, ok := m.cache.Get(ctx, key); ok {
		return cached, nil
	}

	result, err := executor(ctx, toolID, input)
	if err != nil {
		return nil, err
	}

	ttl := m.policy.EffectiveTTL(0)
	if ttl > 0 {
		_ = m.cache.Set(ctx, key, result, ttl)
	}

	return result, nil
}

func (m *CacheMiddleware) shouldSkip(toolID string, tags []string) bool {
	if m.policy.AllowUnsafe {
		return false
	}

	if m.skipRule != nil {
		return m.skipRule(toolID, tags)
	}

	return DefaultSkipRule(toolID, tags)
}
