# PRD-001: toolcache Library Implementation

> **For agents:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Build a caching library for tool execution and discovery with
predictable TTL semantics, deterministic keying, and pluggable backends.

**Architecture:** A pure cache layer that wraps execution with read-through
behavior. Backends are pluggable; the default implementation is in-memory.

**Tech Stack:** Go 1.24+, standard library (optional Redis adapter in future).

**Priority:** P2 (Phase 3 in the plan-of-record)

---

## Context and Stack Alignment

toolcache supports:
- `toolrun` execution caching (idempotent reads)
- `toolindex` query caching (namespaces/tools)
- `metatools-mcp` provider-level caching (optional)

The library must remain **protocol-agnostic** and **transport-agnostic**.

---

## Requirements

### Functional

1. Cache interface with `Get/Set/Delete` using `context.Context`.
2. Deterministic key generation that is order-insensitive.
3. TTL policies with default TTL and per-call overrides.
4. In-memory backend with safe concurrent access.
5. Middleware for read-through caching with safe skip rules.

### Non-functional

- Thread-safe under high concurrency.
- No I/O or network access by default.
- No caching of unsafe tools unless explicitly allowed.
- Deterministic serialization for keys.

---

## Cache Keying Contract

- Key format: `toolcache:<toolID>:<hash>`
- `toolID` is `namespace:name`.
- `hash` is derived from canonical JSON of input payload:
  - Object keys sorted lexicographically.
  - Map order does not affect output.
  - Arrays preserve order.

---

## Directory Structure

```
toolcache/
├── cache.go
├── cache_test.go
├── keyer.go
├── keyer_test.go
├── policy.go
├── policy_test.go
├── memory.go
├── memory_test.go
├── middleware.go
├── middleware_test.go
├── doc.go
├── README.md
├── go.mod
└── go.sum
```

---

## API Model (Target)

```go
// Cache provides the core caching interface.
type Cache interface {
    Get(ctx context.Context, key string) (value []byte, ok bool)
    Set(ctx context.Context, key string, value []byte, ttl time.Duration) error
    Delete(ctx context.Context, key string) error
}

// Keyer generates deterministic cache keys.
type Keyer interface {
    Key(toolID string, input any) (string, error)
}

// Policy controls TTL and cache behavior.
type Policy struct {
    DefaultTTL time.Duration
    AllowUnsafe bool
}
```

---

## TDD Task Breakdown (Detailed)

### Task 1 — Core Interfaces + Errors

**Files:** `cache.go`, `cache_test.go`

**Tests:**
- `TestCacheErrors_NilCache`
- `TestCacheInterface_BasicGetSet`

**Commit:** `feat(toolcache): add cache interface and errors`

---

### Task 2 — Deterministic Keyer

**Files:** `keyer.go`, `keyer_test.go`

**Tests:**
- `TestKeyer_DeterministicForMaps`
- `TestKeyer_ArrayOrderPreserved`
- `TestKeyer_SameInputsSameKey`

**Acceptance:** Map key ordering does not change the cache key.

**Commit:** `feat(toolcache): add deterministic keyer`

---

### Task 3 — TTL Policy

**Files:** `policy.go`, `policy_test.go`

**Tests:**
- `TestPolicy_DefaultTTL`
- `TestPolicy_OverrideTTL`
- `TestPolicy_DisabledCaching`

**Commit:** `feat(toolcache): add ttl policy`

---

### Task 4 — In-memory Backend

**Files:** `memory.go`, `memory_test.go`

**Tests:**
- `TestMemoryCache_GetSetDelete`
- `TestMemoryCache_Expiry`
- `TestMemoryCache_ConcurrentAccess`

**Commit:** `feat(toolcache): add in-memory backend`

---

### Task 5 — Middleware

**Files:** `middleware.go`, `middleware_test.go`

**Tests:**
- `TestMiddleware_CacheHit`
- `TestMiddleware_CacheMiss`
- `TestMiddleware_SkipUnsafeTags`

**Acceptance:** Tools tagged `write` or `danger` are skipped by default.

**Commit:** `feat(toolcache): add read-through middleware`

---

### Task 6 — Docs + Examples

**Files:** `README.md`, `docs/index.md`, `docs/user-journey.md`

**Acceptance:** Quick start example and Mermaid flow diagram included. Add a D2
component diagram in ai-tools-stack.

**Commit:** `docs(toolcache): finalize documentation`

---

## PR Process

1. Create branch: `feat/toolcache-<task>`
2. Implement TDD task in isolation
3. Run: `go test -race ./...`
4. Commit with scoped message
5. Open PR against `main`
6. Merge after CI green

---

## Versioning and Propagation

- **Source of truth:** `ai-tools-stack/go.mod`
- **Matrix:** `ai-tools-stack/VERSIONS.md` (auto-synced)
- **Propagation:** `ai-tools-stack/scripts/update-version-matrix.sh --apply`
- Tags: `vX.Y.Z` and `toolcache-vX.Y.Z`

---

## Integration with metatools-mcp

- Cache tool provider responses at the provider layer.
- Cache read-only tool execution results (configurable).
- Invalidate cache on toolindex change events.

---

## Definition of Done

- All tasks complete with tests passing
- `go test -race ./...` succeeds
- Docs + diagrams updated in ai-tools-stack
- CI green
- Version matrix updated after first release
