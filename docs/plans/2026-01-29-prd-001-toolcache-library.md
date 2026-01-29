# PRD-001: toolcache Library Implementation

> **For agents:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Build a caching library for tool execution and discovery with clear
TTL semantics, deterministic keying, and pluggable backends.

**Architecture:** A pure cache layer that wraps execution with read-through
behavior. Backends are pluggable; default implementation is in-memory.

**Tech Stack:** Go 1.24+, standard library, optional Redis client (future)

**Priority:** P2 (Phase 3 in the plan-of-record)

---

## Context and Stack Alignment

toolcache supports:
- `toolrun` execution caching (idempotent reads)
- `toolindex` query caching (namespaces/tools)
- `metatools-mcp` provider-level caching (optional)

The library must remain **protocol-agnostic** and **transport-agnostic**.

---

## Scope

### In scope
- Cache interface (Get/Set/Delete)
- Deterministic key generation
- TTL and eviction policies
- In-memory backend
- Execution middleware (read-through)
- Unit tests for all exported behavior

### Out of scope
- Distributed cache coordination
- Cache warming strategies
- Persistence to disk
- Write-through semantics for non-idempotent tools

---

## Design Principles

1. **Determinism**: keying must be stable and order-insensitive.
2. **Explicit TTLs**: entries always have a defined lifetime.
3. **Isolation**: caches do not share global state.
4. **Safety**: no caching for unsafe or non-idempotent tools by default.
5. **Minimal dependencies**: no external services required.

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

## API Shape (Conceptual)

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
```

---

## Tasks (TDD)

### Task 1 — Core Interfaces

- Implement `Cache` interface + error types
- Implement `Keyer` with deterministic JSON normalization
- Tests: ordering insensitivity, stable keys

### Task 2 — Policy + TTL

- Implement TTL policy struct
- Tests: default TTL, override behavior

### Task 3 — In-memory Backend

- Thread-safe map-based cache
- TTL eviction on access
- Tests: get/set/delete + expiry

### Task 4 — Middleware

- Read-through middleware wrapper
- Skip caching for unsafe tags (`danger`, `write`)
- Tests: cache hit/miss, skip rules

### Task 5 — Docs + Examples

- Update README and docs/index.md
- Add Mermaid flow diagram
- Add D2 component diagram in ai-tools-stack

---

## Versioning and Propagation

- **Source of truth**: `ai-tools-stack/go.mod`
- **Version matrix**: `ai-tools-stack/VERSIONS.md` (auto-synced)
- **Propagation**: `ai-tools-stack/scripts/update-version-matrix.sh --apply`
- Tags: `vX.Y.Z` and `toolcache-vX.Y.Z`

---

## Integration with metatools-mcp

- Cache tool provider responses at the provider layer.
- Cache read-only tool execution results (configurable).
- Ensure cache invalidation when toolindex changes.

---

## Definition of Done

- All TDD tasks complete with tests passing
- `go test -race ./...` succeeds
- Docs include quick start + diagrams
- CI green
- Version matrix updated after first release
