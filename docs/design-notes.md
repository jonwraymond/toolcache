# Design Notes: Caching Semantics

## Core Principles

- **Determinism**: cache keys must be stable across runs.
- **Explicit TTLs**: cache entries always have explicit lifetimes.
- **Isolation**: no shared mutable state across caches.
- **Opt-in caching**: callers decide what to cache.

## Cache Keying

- Keys are derived from tool ID + normalized input payload.
- Keyer must be deterministic and order-insensitive for maps.

## Invalidation

- Support TTL-based expiration.
- Optional explicit invalidation by tool ID or key prefix.

## Concurrency

- Cache interfaces must be safe under concurrent access.
- Middleware must be re-entrant.

## Integration Points

- `toolrun`: wrap `RunTool` and `RunChain` with cache checks.
- `toolindex`: cache namespace/tool listings.
- `metatools-mcp`: optional caching in provider layer.
