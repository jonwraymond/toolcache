# Migration Guide: toolcache to toolops/cache

This guide helps you migrate from `github.com/jonwraymond/toolcache` to `github.com/jonwraymond/toolops/cache`.

## Quick Migration

### 1. Update go.mod

Remove the old dependency and add the new one:

```bash
go get github.com/jonwraymond/toolops
go mod tidy
```

### 2. Update Imports

Replace all imports:

```go
// Before
import "github.com/jonwraymond/toolcache"

// After
import "github.com/jonwraymond/toolops/cache"
```

### 3. Update Package References

Update all type and function references:

| Before | After |
|--------|-------|
| `toolcache.Cache` | `cache.Cache` |
| `toolcache.Keyer` | `cache.Keyer` |
| `toolcache.Policy` | `cache.Policy` |
| `toolcache.CacheMiddleware` | `cache.CacheMiddleware` |
| `toolcache.MemoryCache` | `cache.MemoryCache` |
| `toolcache.DefaultKeyer` | `cache.DefaultKeyer` |
| `toolcache.SkipRule` | `cache.SkipRule` |
| `toolcache.ToolExecutor` | `cache.ToolExecutor` |

### 4. Update Function Calls

| Before | After |
|--------|-------|
| `toolcache.NewMemoryCache()` | `cache.NewMemoryCache()` |
| `toolcache.NewDefaultKeyer()` | `cache.NewDefaultKeyer()` |
| `toolcache.NewCacheMiddleware()` | `cache.NewCacheMiddleware()` |
| `toolcache.DefaultPolicy()` | `cache.DefaultPolicy()` |
| `toolcache.NoCachePolicy()` | `cache.NoCachePolicy()` |
| `toolcache.DefaultSkipRule` | `cache.DefaultSkipRule` |
| `toolcache.ValidateKey()` | `cache.ValidateKey()` |

## Example Migration

### Before

```go
package main

import (
    "context"
    "time"

    "github.com/jonwraymond/toolcache"
)

func main() {
    cache := toolcache.NewMemoryCache(toolcache.DefaultPolicy())
    keyer := toolcache.NewDefaultKeyer()
    policy := toolcache.Policy{
        DefaultTTL:  5 * time.Minute,
        MaxTTL:      1 * time.Hour,
        AllowUnsafe: false,
    }

    mw := toolcache.NewCacheMiddleware(cache, keyer, policy, toolcache.DefaultSkipRule)

    executor := func(ctx context.Context, toolID string, input any) ([]byte, error) {
        return []byte("result"), nil
    }

    ctx := context.Background()
    result, _ := mw.Execute(ctx, "ns:mytool", map[string]any{"key": "value"}, []string{"read"}, executor)
    _ = result
}
```

### After

```go
package main

import (
    "context"
    "time"

    "github.com/jonwraymond/toolops/cache"
)

func main() {
    c := cache.NewMemoryCache(cache.DefaultPolicy())
    keyer := cache.NewDefaultKeyer()
    policy := cache.Policy{
        DefaultTTL:  5 * time.Minute,
        MaxTTL:      1 * time.Hour,
        AllowUnsafe: false,
    }

    mw := cache.NewCacheMiddleware(c, keyer, policy, cache.DefaultSkipRule)

    executor := func(ctx context.Context, toolID string, input any) ([]byte, error) {
        return []byte("result"), nil
    }

    ctx := context.Background()
    result, _ := mw.Execute(ctx, "ns:mytool", map[string]any{"key": "value"}, []string{"read"}, executor)
    _ = result
}
```

## Automated Migration

You can use `sed` or `gofmt` for bulk updates:

```bash
# Replace imports
find . -name "*.go" -exec sed -i '' 's|"github.com/jonwraymond/toolcache"|"github.com/jonwraymond/toolops/cache"|g' {} +

# Replace package references
find . -name "*.go" -exec sed -i '' 's|toolcache\.|cache.|g' {} +
```

## API Compatibility

The `toolops/cache` package maintains full API compatibility with `toolcache`. All types, functions, and behaviors remain identical. The only change is the import path.

## Questions?

Open an issue in the [toolops repository](https://github.com/jonwraymond/toolops/issues) if you encounter any migration issues.
