# 11 — Review checklist

Use this list on every PR, whether a human or an agent wrote it. A "no" needs a written reason in the PR.

## Boundaries
- [ ] `internal/core` imports only the standard library and other `internal/core` packages (`domain`, `port`, `service/helper`); `make lint-layout` checks it. By hand, check with `go list -f '{{join .Imports "\n"}}' ./internal/core/... | grep -v -e '^[a-z]*$' -e '^[a-z]*/' -e 'internal/core/'`, which prints nothing when the rule holds.
- [ ] No `gin`, `net/http`, SDK or driver types appear in a port signature.
- [ ] No port method takes a collection name, a table name, a key, a URL, or an `interface{}` or `any` payload.
- [ ] Only `repository.New` names concrete repository types. `server.go` names no resource.
- [ ] Nothing reads env vars outside `internal/config`.
- [ ] Every `/v1` route sits behind `httpx.Authenticate`. A request without a valid token gets 401.
- [ ] Every outbound call lives in `internal/repository/<store>`, behind a port.

## Modularity
- [ ] Each layer is built by its one constructor: `repository.New`, `service.New`, `handler.New`.
- [ ] A new resource adds interfaces embedded in `port.Repo`, `port.RedisRepo` or `port.Adapter`, and a field to `port.Services`. It adds nothing to `server.go` or `port.Repositories`.
- [ ] Use cases reach datasources only through `s.repo`, `s.redisRepo`, `s.adpt` and `s.clock` (`make lint-layout`).
- [ ] `port` keeps `service.go`, `database.go`, `adapter.go` and `port.go` (`make lint-layout`).
- [ ] One use case per service file, one endpoint per handler file, and one test function per test file.
- [ ] A function that two or more use-case files call lives in `service/helper/<resource>_<name>.go`, one function per file, with its dependencies as arguments. A function that one file calls stays in that file.

## Ownership
- [ ] Each new rule lives in one place: a domain method or constructor, never a copy in a handler or a second service.
- [ ] Derived values are computed, not stored.
- [ ] A caller cannot set audit fields or IDs, because they are absent from the `Input` types. The actor comes from `auth.ActorFrom(ctx)`, never from the body.
- [ ] Storage tags exist only on record structs in repository packages.
- [ ] Each schema change is a numbered `.up.sql` and `.down.sql` pair, and `db/postgres/ddl/schema.sql` was regenerated with `make ddl` (the `schema` CI job runs `make ddl-check`).
- [ ] Each new Redis key is listed in `db/redis/keys.md`.

## Use cases
- [ ] Order: validate and check preconditions, allocate, build, persist, emit.
- [ ] A read-modify-write runs inside `s.repo.WithinTx`, reads with `Get<Res>ForUpdate`, and uses the inner ctx.
- [ ] No external call runs inside a transaction. Cache invalidation and events run after the commit.
- [ ] Each handler method calls exactly one service method.

## Errors
- [ ] The repository translates driver errors to domain kinds by code, never by string matching.
- [ ] Every wrap uses `%w` with a "verb noun id" prefix.
- [ ] No function both logs and returns the same error. A best-effort port returns no error.
- [ ] A new error kind is added to `httpx.classify`.
- [ ] A stopped datasource makes the API answer 503, not 500.

## Tests
- [ ] Tests exist only in `internal/core/service`.
- [ ] Each use case: the success path, each error kind, each dependency failure, and the ordering and side-effect assertions.
- [ ] Each new assertion fails when the code it covers is broken.
- [ ] No `time.Now()` and no network call in a test.
- [ ] `make check` is green.

## Ops
- [ ] New config keys are in `configs/local.env`, every `configs/<env>.json` and `docker-compose.yml`.
- [ ] No env read outside `internal/config`, no endpoint or IP literal in Go code, and no secret default or secret literal (`make lint-config`).
- [ ] Gitleaks finds no secret (`make secrets`, and the `secrets` job in `.github/workflows/check.yml`).
- [ ] `make up` from a clean `make reset` reaches a listening server.
- [ ] Repository methods start a span and pass the returned ctx down.
- [ ] Outbound HTTP has a timeout shorter than the inbound deadline.
- [ ] Jobs are idempotent and validate their inputs before any I/O.
