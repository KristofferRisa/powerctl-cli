---
name: Task
about: A planned unit of work — carries the Goal, Plan and Tasks that a PR delivers
title: ""
labels: task
assignees: ""
---

# Goal

What outcome are we after, and why does it matter? One or two sentences.
Describe the end state, not the implementation.

# Context

Background a reviewer needs: related issues, prior discussion, constraints,
links to the Tibber API docs, whatever led here.

Relates to #

# Plan

The intended approach. Enough that someone can disagree with it *before* the
code exists.

- **Approach:** ...
- **Command surface (if user-facing):**

  ```bash
  powerctl <command> --flag value
  ```

- **Output shape (if user-facing):** what `--format json` returns, since that's
  the contract agents and scripts depend on.
- **Files likely touched:** `internal/api/queries.go`, `internal/commands/...`,
  `internal/output/...`

# Tasks

- [ ] ...
- [ ] ...
- [ ] Tests added or updated
- [ ] `README.md` / `ARCHITECTURE.md` updated if behaviour changed

# Acceptance criteria

How we know this is done and correct.

- [ ] ...
- [ ] `go test ./...` passes
- [ ] `gofmt -s -l .` is clean and `go vet ./...` is clean

# Out of scope

What this deliberately does *not* cover, so the PR stays reviewable.
