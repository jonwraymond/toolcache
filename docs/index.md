# toolcache

Caching primitives for tool execution and discovery.

## Overview

toolcache provides cache interfaces and common helpers for tool execution results
and registry lookups. It is a pure data/cache library: **no execution, no
transport, no network I/O**. Backends are plugged in by the caller.

## Design Goals

1. Deterministic cache keys and stable serialization
2. Pluggable backends (in-memory, Redis optional)
3. Clear TTL and invalidation semantics
4. Thread-safe access under concurrency
5. Minimal dependencies

## Position in the Stack

```
toolrun/toolindex --> toolcache --> backends
```

## Core Types

| Type | Purpose |
|------|---------|
| `Cache` | Core interface for get/set/delete |
| `Keyer` | Deterministic key generation for tool inputs |
| `Policy` | TTL, size limits, and invalidation rules |
| `CacheMiddleware` | Wrap execution with cache lookup/store |

## Detailed Usage

### Basic Setup

```go
import (
    "context"
    "time"
    
    "github.com/jonwraymond/toolcache"
)

// Create cache with default policy
cache := toolcache.NewMemoryCache(toolcache.DefaultPolicy())

// Create keyer for deterministic key generation
keyer := toolcache.NewDefaultKeyer()

// Create custom policy
policy := toolcache.Policy{
    DefaultTTL:  5 * time.Minute,
    MaxTTL:      1 * time.Hour,
    AllowUnsafe: false,
}

// Create middleware
mw := toolcache.NewCacheMiddleware(cache, keyer, policy, toolcache.DefaultSkipRule)
```

### Executing Tools with Caching

```go
// Define your tool executor
executor := func(ctx context.Context, toolID string, input any) ([]byte, error) {
    // Your tool execution logic
    return []byte("result"), nil
}

// Execute with caching
ctx := context.Background()
result, err := mw.Execute(
    ctx,
    "myns:read_file",
    map[string]any{"path": "/tmp/file.txt"},
    []string{"read"},
    executor,
)
```

### Cache Key Generation

The `DefaultKeyer` generates deterministic keys using canonical JSON serialization:

```go
keyer := toolcache.NewDefaultKeyer()

// Same inputs always produce same keys
key1, _ := keyer.Key("myns:read_file", map[string]any{"path": "/tmp/file", "mode": "r"})
key2, _ := keyer.Key("myns:read_file", map[string]any{"mode": "r", "path": "/tmp/file"})
// key1 == key2 (field order doesn't matter)

// Different inputs produce different keys
key3, _ := keyer.Key("myns:read_file", map[string]any{"path": "/tmp/other"})
// key3 != key1
```

### TTL Policy Management

```go
// Default policy: 5m default, 1h max, no unsafe
policy := toolcache.DefaultPolicy()

// Compute effective TTL
ttl1 := policy.EffectiveTTL(0)              // Returns DefaultTTL (5m)
ttl2 := policy.EffectiveTTL(3 * time.Minute) // Returns 3m (override)
ttl3 := policy.EffectiveTTL(2 * time.Hour)   // Returns MaxTTL (1h, clamped)

// Disable caching
noCache := toolcache.NoCachePolicy()
ttl4 := noCache.EffectiveTTL(0) // Returns 0 (no caching)
```

### Custom Skip Rules

Skip rules control which tools bypass caching:

```go
// Skip tools in specific namespaces
customSkip := func(toolID string, tags []string) bool {
    return strings.HasPrefix(toolID, "dangerous:")
}

mw := toolcache.NewCacheMiddleware(cache, keyer, policy, customSkip)

// Never skip (cache everything, even unsafe operations)
// WARNING: Only use with AllowUnsafe: true
neverSkip := func(toolID string, tags []string) bool {
    return false
}

policy := toolcache.Policy{
    DefaultTTL:  5 * time.Minute,
    AllowUnsafe: true, // Required for caching unsafe operations
}
mw := toolcache.NewCacheMiddleware(cache, keyer, policy, neverSkip)
```

### Direct Cache Operations

You can use the `Cache` interface directly for custom caching needs:

```go
cache := toolcache.NewMemoryCache(toolcache.DefaultPolicy())

// Set a value with TTL
ctx := context.Background()
err := cache.Set(ctx, "mykey", []byte("myvalue"), 5*time.Minute)

// Get a value
value, ok := cache.Get(ctx, "mykey")
if ok {
    fmt.Println(string(value))
}

// Delete a value
err = cache.Delete(ctx, "mykey")
```

### Thread-Safety

All components are safe for concurrent use:

```go
cache := toolcache.NewMemoryCache(toolcache.DefaultPolicy())
keyer := toolcache.NewDefaultKeyer()
mw := toolcache.NewCacheMiddleware(cache, keyer, policy, toolcache.DefaultSkipRule)

// Safe to use from multiple goroutines
for i := 0; i < 10; i++ {
    go func(id int) {
        result, _ := mw.Execute(
            context.Background(),
            "myns:tool",
            map[string]any{"id": id},
            []string{"read"},
            executor,
        )
        fmt.Println(string(result))
    }(i)
}
```

### Key Validation

Cache keys are validated to ensure backend compatibility:

```go
// Valid keys
err := toolcache.ValidateKey("toolcache:myns:tool:abc123") // nil

// Invalid keys
err = toolcache.ValidateKey("")                    // ErrInvalidKey
err = toolcache.ValidateKey("key\nwith\nnewlines") // ErrInvalidKey
err = toolcache.ValidateKey(strings.Repeat("x", 513)) // ErrKeyTooLong
```

## Advanced Patterns

### Custom Cache Backend

Implement the `Cache` interface for custom backends:

```go
type RedisCache struct {
    client *redis.Client
}

func (c *RedisCache) Get(ctx context.Context, key string) ([]byte, bool) {
    val, err := c.client.Get(ctx, key).Bytes()
    if err == redis.Nil {
        return nil, false
    }
    if err != nil {
        return nil, false
    }
    return val, true
}

func (c *RedisCache) Set(ctx context.Context, key string, value []byte, ttl time.Duration) error {
    return c.client.Set(ctx, key, value, ttl).Err()
}

func (c *RedisCache) Delete(ctx context.Context, key string) error {
    return c.client.Del(ctx, key).Err()
}
```

### Custom Keyer

Implement the `Keyer` interface for custom key generation:

```go
type PrefixKeyer struct {
    prefix string
    base   toolcache.Keyer
}

func (k *PrefixKeyer) Key(toolID string, input any) (string, error) {
    baseKey, err := k.base.Key(toolID, input)
    if err != nil {
        return "", err
    }
    return k.prefix + ":" + baseKey, nil
}
```

## Versioning

toolcache follows semantic versioning aligned with the stack. The source of
truth is `ai-tools-stack/go.mod`, and `VERSIONS.md` is synchronized across repos.

## Next Steps

- [Design Notes](design-notes.md)
- [User Journey](user-journey.md)
- [Plans](plans/README.md)
