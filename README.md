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

## Versioning

toolcache follows semantic versioning aligned with the stack. The source of
truth is `ai-tools-stack/go.mod`, and `VERSIONS.md` is synchronized across repos.

## Next Steps

- See `docs/index.md` for usage and design notes.
- PRD and execution plan live in `docs/plans/`.
