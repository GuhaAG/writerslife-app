# PostgreSQL Content Persistence Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Persist every user, fiction, chapter, and follow route in PostgreSQL while preserving API contracts and concurrency-safe counters.

**Architecture:** Extend `db.Store` with explicit pgx repositories. Mutations that allocate chapter numbers, publish chapters, or change follow state run in transactions so parent counters change exactly once. Handlers retain HTTP validation and authorization only.

**Tech Stack:** Go 1.21, Gin, pgx/v5, PostgreSQL, testify.

## Global Constraints

- Preserve existing route, JSON payload, and status-code contracts.
- Use the existing `APP_ENV=test` and `writerslife_test` integration safety guard.
- Use TDD and do not commit.

---

### Task 1: Repository correctness fixes

**Files:** Modify `api/books-api/db/users.go`, `api/books-api/db/fictions.go`, `api/books-api/api_test.go`.

- [ ] Add failing safe-gated HTTP tests for duplicate username/email, literal `%`/`_` search, fresh-pool login, and fiction updates preserving counters.
- [ ] Implement unique-violation classification, escaped `ILIKE ... ESCAPE '\\'`, and metadata-only fiction updates.
- [ ] Run `go test ./...`.

### Task 2: Chapter persistence

**Files:** Create `api/books-api/db/chapters.go`; modify `api/books-api/main.go`, `api/books-api/api_test.go`.

- [ ] Add failing safe-gated HTTP tests for draft visibility, sequential chapter numbers, and one-time publishing counter changes.
- [ ] Implement transactional creation and update operations that lock the fiction row, assign `MAX(chapter_number)+1`, and update chapter count only on draft-to-published transitions.
- [ ] Replace all chapter map accesses with store calls and run `go test ./...`.

### Task 3: Follow persistence and cleanup

**Files:** Create `api/books-api/db/follows.go`; modify `api/books-api/main.go`, `api/books-api/api_test.go`.

- [ ] Add failing safe-gated HTTP tests for idempotent follow/unfollow, follow status, and followed-fiction lists.
- [ ] Implement transactional `INSERT ... ON CONFLICT DO NOTHING` and conditional delete mutations with atomic follower count changes.
- [ ] Remove remaining legacy maps/seed helpers, run `gofmt`, `go test ./...`, and `go vet ./...`.
