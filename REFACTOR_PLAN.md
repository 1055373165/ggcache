# ggcache Refactor Plan

This document outlines the **step-by-step plan** to migrate the current codebase to a Go project layout that follows modern best practices while keeping the application runnable at every stage.

---

## Phase 0 — Preparation (Day 0)
1. **Freeze feature work** – only bug-fixes during refactor.
2. Make sure CI is green and tests pass.
3. Create `refactor` branch.  All steps below will be PRs targeting this branch until complete.

## Phase 1 — Directory Skeleton (Day 0-1)
Goal: introduce the new folders without moving any code.

Tasks:
- [ ] Create directories:
  ```
  cmd/ggcache
  internal/app
  internal/observability
  internal/transport/{grpcserver,httpserver}
  internal/cache/{core,transport}
  internal/data/mysql
  internal/domain/student
  configs
  deploy
  scripts
  docs
  examples
  ```
- [ ] Add empty `README.md` in each new top-level folder to keep them in git.

## Phase 2 — Move Entry Point (Day 1)
Goal: isolate the executable.

Tasks:
- [ ] Move current `main.go` → `cmd/ggcache/main.go`.
- [ ] Replace direct init logic with thin `app.Run()` call located in `internal/app/app.go`.
- [ ] Update Go modules import paths & run `go vet ./...`.
- [ ] CI green.

## Phase 3 — Config Refactor (Day 1-2)
Goal: centralise configuration.

Tasks:
- [ ] Move `config/` → `configs/`.
- [ ] Introduce `internal/app/config.go` to load YAML + env vars + flags.
- [ ] Deprecate hard-coded defaults.

## Phase 4 — Observability Pack (Day 2)
Goal: unify log, metrics, pprof.

Tasks:
- [ ] Create `internal/observability/obs.go` that initialises logrus, prometheus, pprof.
- [ ] Call `observability.Init(cfg.Observability)` from `app.Run()`.

## Phase 5 — Cache Core Split (Day 2-3)
Goal: split algorithmic code from transport.

Tasks:
- [ ] Move pure algorithms (consistent hash, eviction) into `internal/cache/core`.
- [ ] Move gRPC / HTTP fetcher & picker to `internal/cache/transport`.
- [ ] Adapt imports.

## Phase 6 — Transport Servers (Day 3-4)
Goal: encapsulate server start/stop.

Tasks:
- [ ] Create `internal/transport/grpcserver/server.go` (wrap existing `grpc_picker.Server`).
- [ ] Create `internal/transport/httpserver/server.go`.
- [ ] `app.Run()` just calls `grpcserver.Start()` & `httpserver.Start()`.

## Phase 7 — Data & Domain Layers (Day 4-5)
Goal: decouple business domain from MySQL.

Tasks:
- [ ] Move DAO code into `internal/data/mysql/student_repo.go`.
- [ ] Introduce `internal/domain/student/model.go` + `service.go`.
- [ ] Adjust cache getter to use `domain` service instead of DAO directly.

## Phase 8 — Examples & Tests Cleanup (Day 5)
Tasks:
- [ ] Move ad-hoc clients under `examples/`.
- [ ] Keep unit tests co-located; move integration tests to `internal/test/e2e`.
- [ ] Introduce Testcontainers-based e2e runner.

## Phase 9 — Makefile & CI (Day 5-6)
Tasks:
- [ ] Add `Makefile` with targets `build | test | lint | docker`.
- [ ] Update GitHub Actions to use cache and new targets.

## Phase 10 — Documentation & Clean-Up (Day 6)
Tasks:
- [ ] Update `README.md`, add `docs/architecture.md`, `CONTRIBUTING.md`.
- [ ] Run `go mod tidy`, remove obsolete packages, ensure `go vet`, `staticcheck`, `golangci-lint` all pass.

## Milestones & Acceptance
1. **Each Phase** merged only when CI green.
2. After Phase 6 the service must run as before (backwards compatibility).
3. After Phase 10 tag `v2.0.0`.

## Future Evolution

- **Client SDK**: move reusable client code to `pkg/client`, add client-side consistent-hash balancer for direct peer dial.
- **Hot-node optimisation**: leverage runtime metrics to implement auto-scaling / rate-limiting for hotspots.

## Documentation To-Dos

- `docs/architecture.md`: consistent hashing algorithm, etcd discovery flow, data path diagrams.
- `docs/deploy/k8s.md`: Helm chart & example values.yaml.
- `CONTRIBUTING.md`: coding style, commit message and PR guide, branch strategy.

---

*Generated: 2025-06-30*
