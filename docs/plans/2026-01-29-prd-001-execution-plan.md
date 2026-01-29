# Toolcache Implementation Plan — Professional TDD Execution

Status: Ready for Implementation
Date: 2026-01-29
PRD: docs/plans/2026-01-29-prd-001-toolcache-library.md

## Overview

Implement caching primitives, deterministic keying, and read-through middleware.

## TDD Methodology

Each task follows strict TDD:
1. Red — write failing test
2. Red verification — run test, confirm failure
3. Green — minimal implementation
4. Green verification — run test, confirm pass
5. Commit — one commit per task

---

## Task 0 — Module Scaffolding

Commit:
- chore(toolcache): scaffold module and docs

---

## Task 1 — Core Interfaces + Errors

Tests:
- TestCacheErrors_NilCache
- TestCacheInterface_BasicGetSet

Commit:
- feat(toolcache): add cache interface and errors

---

## Task 2 — Deterministic Keyer

Tests:
- TestKeyer_DeterministicForMaps
- TestKeyer_ArrayOrderPreserved
- TestKeyer_SameInputsSameKey

Commit:
- feat(toolcache): add deterministic keyer

---

## Task 3 — TTL Policy

Tests:
- TestPolicy_DefaultTTL
- TestPolicy_OverrideTTL
- TestPolicy_DisabledCaching

Commit:
- feat(toolcache): add ttl policy

---

## Task 4 — In-memory Backend

Tests:
- TestMemoryCache_GetSetDelete
- TestMemoryCache_Expiry
- TestMemoryCache_ConcurrentAccess

Commit:
- feat(toolcache): add in-memory backend

---

## Task 5 — Middleware

Tests:
- TestMiddleware_CacheHit
- TestMiddleware_CacheMiss
- TestMiddleware_SkipUnsafeTags

Commit:
- feat(toolcache): add read-through middleware

---

## Task 6 — Docs + Diagrams

Steps:
- Expand README examples
- Add Mermaid diagram to user-journey
- Add D2 component diagram in ai-tools-stack

Commit:
- docs(toolcache): finalize documentation

---

## Quality Gates

- go test -v -race ./...
- go test -cover ./...
- go vet ./...
- golangci-lint run (if configured)

---

## Stack Integration

1. Add ai-tools-stack component docs + D2 diagram
2. Add mkdocs multirepo import
3. After first release, update version matrix

---

## Commit Order

1. chore(toolcache): scaffold module and docs
2. feat(toolcache): add cache interface and errors
3. feat(toolcache): add deterministic keyer
4. feat(toolcache): add ttl policy
5. feat(toolcache): add in-memory backend
6. feat(toolcache): add read-through middleware
7. docs(toolcache): finalize documentation
8. docs(ai-tools-stack): add toolcache component docs
9. chore(ai-tools-stack): add toolcache to version matrix (after release)
