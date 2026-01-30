# toolcache

Caching library for tool execution and discovery.

## Overview

toolcache provides a small, composable caching layer for tool inputs/outputs and
registry lookups. It is a pure data/cache library: **no execution, no transport,
no network I/O**. Optional backends may be wired by the caller.

## Design Goals

1. Deterministic cache keys for tool execution
2. Pluggable backends (memory, Redis optional)
3. Explicit TTL and invalidation semantics
4. Thread-safe access under high concurrency
5. Minimal dependencies

## Position in the Stack

```
toolrun/toolindex --> toolcache --> backends
```

## Installation

```bash
go get github.com/jonwraymond/toolcache
```

## Quick Start

```go
package main

import (
    "context"
    "fmt"
    "time"
    
    "github.com/jonwraymond/toolcache"
)

func main() {
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
        // Your tool execution logic here
        return []byte("result"), nil
    }
    
    // Execute with caching
    ctx := context.Background()
    result, err := mw.Execute(ctx, "ns:mytool", map[string]any{"key": "value"}, []string{"read"}, executor)
    if err != nil {
        panic(err)
    }
    
    fmt.Println(string(result))
}
```

## API Documentation

### Cache Interface

The `Cache` interface defines the core caching operations:

```go
type Cache interface {
    Get(ctx context.Context, key string) (value []byte, ok bool)
    Set(ctx context.Context, key string, value []byte, ttl time.Duration) error
    Delete(ctx context.Context, key string) error
}
```

**Implementations:**
- `MemoryCache`: Thread-safe in-memory cache with TTL support

**Thread-Safety:** All `Cache` implementations are safe for concurrent use.

### Keyer Interface

The `Keyer` interface generates deterministic cache keys from tool inputs:

```go
type Keyer interface {
    Key(toolID string, input any) (string, error)
}
```

**Implementations:**
- `DefaultKeyer`: Generates keys using canonical JSON serialization and SHA-256 hashing

**Key Format:** `toolcache:<toolID>:<hash>`

**Determinism:** Identical inputs always produce identical keys, regardless of field order in maps.

### Policy

The `Policy` struct controls caching behavior:

```go
type Policy struct {
    DefaultTTL  time.Duration  // TTL when no override provided (0 = no caching)
    MaxTTL      time.Duration  // Maximum allowed TTL (0 = no limit)
    AllowUnsafe bool           // Allow caching unsafe operations (default: false)
}
```

**Methods:**
- `EffectiveTTL(override time.Duration) time.Duration`: Computes the TTL to use
- `ShouldCache() bool`: Reports whether caching is enabled

**Helpers:**
- `DefaultPolicy()`: Returns sensible defaults (5m default, 1h max, no unsafe)
- `NoCachePolicy()`: Returns a policy that disables all caching

#### TTL Resolution Rules

1. If `override > 0`, use override
2. If `override <= 0`, use `DefaultTTL`
3. If `MaxTTL > 0` and effective TTL > `MaxTTL`, clamp to `MaxTTL`
4. If effective TTL <= 0, return 0 (no caching)

### CacheMiddleware

The `CacheMiddleware` wraps tool execution with cache lookup and storage:

```go
type CacheMiddleware struct {
    // private fields
}

func NewCacheMiddleware(cache Cache, keyer Keyer, policy Policy, skipRule SkipRule) *CacheMiddleware

func (m *CacheMiddleware) Execute(
    ctx context.Context,
    toolID string,
    input any,
    tags []string,
    executor ToolExecutor,
) ([]byte, error)
```

**Execution Flow:**
1. Check if tool should be skipped (via `SkipRule`)
2. Generate cache key from toolID and input
3. Check cache for existing result
4. If miss, execute tool and store result with TTL
5. Return result

### Skip Rules

Skip rules determine which tools should bypass caching:

```go
type SkipRule func(toolID string, tags []string) bool
```

**Default Skip Rule:** Skips tools with tags: `write`, `danger`, `unsafe`, `mutation`, `delete`

**Custom Skip Rules:**
```go
// Skip all tools in a specific namespace
customSkip := func(toolID string, tags []string) bool {
    return strings.HasPrefix(toolID, "dangerous:")
}

// Never skip (cache everything)
neverSkip := func(toolID string, tags []string) bool {
    return false
}
```

## Thread-Safety Guarantees

- **MemoryCache**: All operations are protected by `sync.RWMutex`
- **CacheMiddleware**: Stateless and safe for concurrent use
- **DefaultKeyer**: Stateless and safe for concurrent use
- **Policy**: Immutable after creation, safe for concurrent use

## Key Validation

Cache keys are validated to ensure compatibility with backend systems:

- **Max Length:** 512 characters
- **Forbidden Characters:** Newlines (`\n`, `\r`)
- **Empty Keys:** Not allowed

Use `ValidateKey(key string) error` to check key validity.

## Examples

See `example_test.go` for runnable examples that appear in godoc.

## Versioning

toolcache follows semantic versioning aligned with the stack. The source of
truth is `ai-tools-stack/go.mod`, and `VERSIONS.md` is synchronized across repos.

## Next Steps

- See `docs/index.md` for detailed usage and design notes.
- PRD and execution plan live in `docs/plans/`.
