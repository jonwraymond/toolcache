# User Journey: Caching Tool Results

## Scenario

You want to cache results for read-only tools to reduce latency and cost. Cache
entries expire after 30 seconds, and you want deterministic keys.

## Step 1: Initialize cache

```go
cache := toolcache.NewMemoryCache(
    toolcache.WithDefaultTTL(30*time.Second),
)
keyer := toolcache.NewKeyer()
```

## Step 2: Wrap execution

```go
mw := toolcache.NewMiddleware(cache, keyer)
wrapped := mw.Wrap(toolrunExecutor)
_ = wrapped
```

## Step 3: Execute tool

- First call populates cache
- Subsequent calls within TTL return cached result

## Flow Diagram

```mermaid
flowchart LR
  A[Tool Call] --> B[Cache Lookup]
  B -->|Hit| C[Return Cached]
  B -->|Miss| D[Execute Tool]
  D --> E[Store Result]
  E --> F[Return Result]
```
