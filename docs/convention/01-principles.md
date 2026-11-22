# 01 — Principles

Every rule in the other docs follows from these principles. If a change fights one, redesign the change.
Don't add an exception.

## P1. Dependencies point inward

```
cmd → server → { router → httpx → handler } ──────────┐
             → { repository → memory | postgres | redis | catalog } ─┼─→ core/port → core/domain
             → { platform } ──────────────────────────┘      ↑
                                                core/service ─┘
```

- `core/domain` imports only the standard library.
- `core/port` imports `domain`.
- `core/service` imports `port`, `domain` and `core/service/helper`. `helper` imports `port` and `domain`.
- Every package outside `core` may import `core`. `core` never imports a package outside it.

Review enforces the rule with the command in [11-review-checklist.md](11-review-checklist.md#boundaries). That command prints nothing when the rule holds.

## P2. Every piece of state and every rule has one owner

| Thing | Owner |
|---|---|
| Invariants of an aggregate (date order, audit fields) | `domain.NewX` and `X.Revise` |
| Derived values (status) | a domain method (`StatusAt`) |
| Rules that two use cases share | `service/helper/<resource>_<name>.go`, one function per file |
| Wire formats (date strings, JSON names) | `handler/dto` |
| Storage formats (tables, columns, keys, streams) | the `repository/<store>` package |
| The Postgres schema | `db/postgres/migrations`. `db/postgres/ddl/schema.sql` is derived from it. |
| External API formats (URLs, headers, status codes) | `repository/catalog` |
| Env var names and defaults | `internal/config` |
| The HTTP status for an error | `httpx.classify` |
| Which concrete type implements a port | `repository.New` |
| Which services and handlers exist | `service.New`, `port.Services` |

Before you write a rule in a second place, call its owner.

## P3. One way to do each task

- One binder, `httpx.Handle`, for every endpoint.
- One error envelope, `dto.ErrorResponse`, and one mapping, `httpx.WriteError`.
- One use-case order: validate and check preconditions, allocate, build, persist, emit.
- One source for time (`port.Clock`), IDs (`port.IDGenerator`) and transactions (`port.Transactor`).
- One constructor per layer, so the wiring has one shape for every resource.

## P4. Callers know the least

- Ports speak intent in domain types: `NextWidgetID(ctx)`, not `Incr(ctx, "scaffold:widget-id")`.
- A port has no parameter for a collection, a table, a key, a topic, a URL or an `interface{}` value.
- Handlers don't know which database exists. Services don't know that HTTP exists.
- Constructors receive their config section. Nothing reads globals or env vars after startup.
- A best-effort port returns no error, so its caller cannot fail on it by mistake.

## P5. Every outbound call is a repository

`internal/repository/<store>` holds every call that leaves the process.
The origin repo split owned state from other teams' systems. The scaffold keeps one driven layer, so one constructor wires it.

| Store | Ports it implements |
|---|---|
| `memory` | `Repo` and `RedisRepo`, for local runs and as test fakes |
| `postgres` | `Repo` (`WidgetRepository`, `Transactor`), used as `s.repo` |
| `redis` | `RedisRepo` (`IDGenerator`, `WidgetCache`, `EventPublisher`), used as `s.redisRepo` |
| `catalog` | `Adapter` (`CatalogAdapter`, an HTTP API that another team owns), used as `s.adpt` |

Each store translates its failures into domain error kinds.

## P6. Errors: wrap on the way up, decide once at the edge

- The lowest layer that understands a failure turns it into a kind: `domain.ErrNotFound`, `ErrInvalidInput`, `ErrConflict` or `ErrUnavailable`.
- Every layer above adds context with `fmt.Errorf("verb noun %s: %w", id, err)`.
- No function logs and returns the same error. `httpx.WriteError` logs once and picks the status. A job logs once in `main`.
- A best-effort port logs its own failure, because no caller receives it.

## P7. Tests run the service layer through ports

Time, IDs, storage and outbound calls all reach the service through ports.
The service tests use the memory stand-ins and hand-written fakes, and they run in milliseconds.
The real datasources are checked with `make up` and a smoke run, not with unit tests.

## P8. Boring, explicit, greppable

- No reflection-driven persistence, no DI container, no `init()`.
- File names say what they hold: `widget_create.go`, `widget_create_test.go`, `000001_create_widgets.up.sql`.
- `server.go` shows the layer graph. `repository.New` shows which datasource backs each port.
