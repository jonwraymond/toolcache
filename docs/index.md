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
| `Middleware` | Wrap execution with cache lookup/store |

## Quick Start

```go
cache := toolcache.NewMemoryCache()
keyer := toolcache.NewKeyer()

mw := toolcache.NewMiddleware(cache, keyer)
wrapped := mw.Wrap(toolrunExecutor)
_ = wrapped
```

## Versioning

toolcache follows semantic versioning aligned with the stack. The source of
truth is `ai-tools-stack/go.mod`, and `VERSIONS.md` is synchronized across repos.

## Next Steps

- [Design Notes](design-notes.md)
- [User Journey](user-journey.md)
- [Plans](plans/README.md)
