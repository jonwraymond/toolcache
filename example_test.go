package toolcache_test

import (
	"context"
	"fmt"
	"time"

	"github.com/jonwraymond/toolcache"
)

// ExampleNewCacheMiddleware demonstrates creating and using cache middleware
// to wrap tool execution with caching.
func ExampleNewCacheMiddleware() {
	// Create cache components
	cache := toolcache.NewMemoryCache(toolcache.DefaultPolicy())
	keyer := toolcache.NewDefaultKeyer()
	policy := toolcache.Policy{
		DefaultTTL:  5 * time.Minute,
		MaxTTL:      1 * time.Hour,
		AllowUnsafe: false,
	}

	// Create middleware
	mw := toolcache.NewCacheMiddleware(cache, keyer, policy, toolcache.DefaultSkipRule)

	// Define an executor
	executor := func(ctx context.Context, toolID string, input any) ([]byte, error) {
		return []byte("result"), nil
	}

	// Execute with caching
	ctx := context.Background()
	result, _ := mw.Execute(ctx, "ns:mytool", map[string]any{"key": "value"}, []string{"read"}, executor)

	fmt.Println(string(result))
	// Output: result
}

// ExampleDefaultKeyer demonstrates deterministic key generation.
// Keys are identical for the same inputs regardless of field order.
func ExampleDefaultKeyer() {
	keyer := toolcache.NewDefaultKeyer()

	// Same inputs produce same keys
	key1, _ := keyer.Key("myns:read_file", map[string]any{"path": "/tmp/file"})
	key2, _ := keyer.Key("myns:read_file", map[string]any{"path": "/tmp/file"})

	fmt.Println(key1 == key2)
	// Output: true
}

// ExampleDefaultKeyer_fieldOrder demonstrates that field order doesn't affect
// key generation due to canonical JSON serialization.
func ExampleDefaultKeyer_fieldOrder() {
	keyer := toolcache.NewDefaultKeyer()

	// Different field order produces same key
	key1, _ := keyer.Key("myns:tool", map[string]any{"a": 1, "b": 2})
	key2, _ := keyer.Key("myns:tool", map[string]any{"b": 2, "a": 1})

	fmt.Println(key1 == key2)
	// Output: true
}

// ExamplePolicy_EffectiveTTL demonstrates TTL resolution rules.
func ExamplePolicy_EffectiveTTL() {
	policy := toolcache.Policy{
		DefaultTTL: 5 * time.Minute,
		MaxTTL:     10 * time.Minute,
	}

	// No override: use DefaultTTL
	fmt.Println(policy.EffectiveTTL(0))

	// Override within MaxTTL: use override
	fmt.Println(policy.EffectiveTTL(3 * time.Minute))

	// Override exceeds MaxTTL: clamp to MaxTTL
	fmt.Println(policy.EffectiveTTL(15 * time.Minute))

	// Output:
	// 5m0s
	// 3m0s
	// 10m0s
}

// ExamplePolicy_ShouldCache demonstrates checking if caching is enabled.
func ExamplePolicy_ShouldCache() {
	// Policy with caching enabled
	policy1 := toolcache.Policy{DefaultTTL: 5 * time.Minute}
	fmt.Println(policy1.ShouldCache())

	// Policy with caching disabled
	policy2 := toolcache.NoCachePolicy()
	fmt.Println(policy2.ShouldCache())

	// Output:
	// true
	// false
}

// ExampleDefaultPolicy demonstrates the default policy configuration.
func ExampleDefaultPolicy() {
	policy := toolcache.DefaultPolicy()

	fmt.Println(policy.DefaultTTL)
	fmt.Println(policy.MaxTTL)
	fmt.Println(policy.AllowUnsafe)

	// Output:
	// 5m0s
	// 1h0m0s
	// false
}

// ExampleNoCachePolicy demonstrates a policy that disables all caching.
func ExampleNoCachePolicy() {
	policy := toolcache.NoCachePolicy()

	fmt.Println(policy.DefaultTTL)
	fmt.Println(policy.ShouldCache())

	// Output:
	// 0s
	// false
}

// ExampleMemoryCache demonstrates direct cache operations.
func ExampleMemoryCache() {
	cache := toolcache.NewMemoryCache(toolcache.DefaultPolicy())
	ctx := context.Background()

	// Set a value
	_ = cache.Set(ctx, "mykey", []byte("myvalue"), 5*time.Minute)

	// Get the value
	value, ok := cache.Get(ctx, "mykey")
	if ok {
		fmt.Println(string(value))
	}

	// Delete the value
	_ = cache.Delete(ctx, "mykey")

	// Value is gone
	_, ok = cache.Get(ctx, "mykey")
	fmt.Println(ok)

	// Output:
	// myvalue
	// false
}

// ExampleMemoryCache_expiration demonstrates TTL expiration behavior.
func ExampleMemoryCache_expiration() {
	cache := toolcache.NewMemoryCache(toolcache.DefaultPolicy())
	ctx := context.Background()

	// Set a value with very short TTL
	_ = cache.Set(ctx, "mykey", []byte("myvalue"), 1*time.Nanosecond)

	// Value expires immediately
	time.Sleep(2 * time.Millisecond)
	_, ok := cache.Get(ctx, "mykey")
	fmt.Println(ok)

	// Output:
	// false
}

// ExampleDefaultSkipRule demonstrates the default skip rule behavior.
func ExampleDefaultSkipRule() {
	// Safe operations are not skipped
	skip1 := toolcache.DefaultSkipRule("myns:read_file", []string{"read"})
	fmt.Println(skip1)

	// Unsafe operations are skipped
	skip2 := toolcache.DefaultSkipRule("myns:delete_file", []string{"write", "delete"})
	fmt.Println(skip2)

	// Output:
	// false
	// true
}

// ExampleCacheMiddleware_Execute demonstrates the full execution flow
// with cache hit and miss scenarios.
func ExampleCacheMiddleware_Execute() {
	cache := toolcache.NewMemoryCache(toolcache.DefaultPolicy())
	keyer := toolcache.NewDefaultKeyer()
	policy := toolcache.Policy{
		DefaultTTL:  5 * time.Minute,
		AllowUnsafe: false,
	}
	mw := toolcache.NewCacheMiddleware(cache, keyer, policy, toolcache.DefaultSkipRule)

	callCount := 0
	executor := func(ctx context.Context, toolID string, input any) ([]byte, error) {
		callCount++
		return []byte(fmt.Sprintf("result-%d", callCount)), nil
	}

	ctx := context.Background()
	input := map[string]any{"key": "value"}

	// First call: cache miss, executor called
	result1, _ := mw.Execute(ctx, "ns:mytool", input, []string{"read"}, executor)
	fmt.Println(string(result1))

	// Second call: cache hit, executor not called
	result2, _ := mw.Execute(ctx, "ns:mytool", input, []string{"read"}, executor)
	fmt.Println(string(result2))

	// Executor was only called once
	fmt.Println(callCount)

	// Output:
	// result-1
	// result-1
	// 1
}

// ExampleCacheMiddleware_Execute_skipUnsafe demonstrates that unsafe
// operations bypass the cache.
func ExampleCacheMiddleware_Execute_skipUnsafe() {
	cache := toolcache.NewMemoryCache(toolcache.DefaultPolicy())
	keyer := toolcache.NewDefaultKeyer()
	policy := toolcache.Policy{
		DefaultTTL:  5 * time.Minute,
		AllowUnsafe: false, // Unsafe operations are skipped
	}
	mw := toolcache.NewCacheMiddleware(cache, keyer, policy, toolcache.DefaultSkipRule)

	callCount := 0
	executor := func(ctx context.Context, toolID string, input any) ([]byte, error) {
		callCount++
		return []byte(fmt.Sprintf("result-%d", callCount)), nil
	}

	ctx := context.Background()
	input := map[string]any{"key": "value"}

	// First call with unsafe tag
	result1, _ := mw.Execute(ctx, "ns:delete", input, []string{"write"}, executor)
	fmt.Println(string(result1))

	// Second call with same input and unsafe tag
	result2, _ := mw.Execute(ctx, "ns:delete", input, []string{"write"}, executor)
	fmt.Println(string(result2))

	// Executor was called twice (no caching)
	fmt.Println(callCount)

	// Output:
	// result-1
	// result-2
	// 2
}

// ExampleValidateKey demonstrates key validation rules.
func ExampleValidateKey() {
	// Valid key
	err1 := toolcache.ValidateKey("toolcache:myns:tool:abc123")
	fmt.Println(err1 == nil)

	// Empty key
	err2 := toolcache.ValidateKey("")
	fmt.Println(err2)

	// Key with newline
	err3 := toolcache.ValidateKey("key\nwith\nnewline")
	fmt.Println(err3)

	// Output:
	// true
	// toolcache: key is invalid
	// toolcache: key is invalid
}
