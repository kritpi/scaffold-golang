# 03 — Layers, file by file

Each section says what the layer does, what it may import, what it must not do, and how its files are laid out.
Code excerpts come from the reference implementation in this repo.

Every layer has one constructor that covers every resource:

| Layer | Constructor | Returns | Adding a resource adds |
|---|---|---|---|
| repository | `repository.New(ctx, cfg, log)` | `port.Repositories{Repo, RedisRepo, Adpt}`, `close` | methods on each store's `Repo`; no new field |
| service | `service.New(cfg, deps)` | `port.Services` | one field and one struct |
| handler | `handler.New(svcs, cfg)` | `*handler.Handler` | nothing; methods go in new files |
| router | `router.Setup(engine, h, verifier)` | — | one route group |

---

## `cmd/<binary>/main.go`, the process entry

**Does:** sets up signal handling, calls `server.Run(ctx)`, and turns an error into exit code 1.
**Must not:** read config, build clients, or hold logic.

```go
func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	if err := server.Run(ctx); err != nil {
		slog.Error("server stopped", "err", err)
		os.Exit(1)
	}
}
```

Use one directory per binary, such as `cmd/server` and `cmd/worker`.

---

## `internal/config`, environment to typed struct

**Does:** reads every env var once and returns `Config`. Each section has the type that its consumer owns: `service.Config` and `repository.Config`.
It collects every parse error, so one bad deploy reports all of them.
**Must not:** be imported by `core` or `repository`. Those packages define the section types, and `server` passes the sections down.

To add a key, follow [recipe F](10-recipes.md#f-add-a-config-key).

---

## `internal/server`, the composition root

**Does:** calls one constructor per layer, then starts the HTTP server with graceful shutdown.

```go
cfg, err := config.Load()
verifier, err := auth.New(cfg.Auth)                                 // 0. bearer-token verifier
repos, closeRepos, err := repository.New(ctx, cfg.Repository, log) // 1. driven layer
defer closeRepos()
svcs := service.New(cfg.Service, service.Deps{Repositories: repos, Clock: platform.SystemClock{}}) // 2. core
router.Setup(engine, handler.New(svcs, cfg.Handler), verifier)        // 3. driving layer
return serve(ctx, cfg.Server, engine, log)
```

**Rule:** `server.go` names no resource. Adding a resource or swapping a datasource does not change it.

---

## `internal/router`, the URL table

**Does:** maps a method and a path to a handler method through `httpx.Handle(status, fn)`. It also mounts health checks and middleware.
**Must not:** contain closures with logic, bind requests itself, or read config.

```go
widgets := v1.Group("/widgets")
{
	widgets.GET("", httpx.Handle(http.StatusOK, h.ListWidgets))
	widgets.POST("", httpx.Handle(http.StatusCreated, h.CreateWidget))
	widgets.GET("/:widget-id", httpx.Handle(http.StatusOK, h.GetWidget))
	widgets.PUT("/:widget-id", httpx.Handle(http.StatusOK, h.UpdateWidget))
	widgets.DELETE("/:widget-id", httpx.Handle(http.StatusOK, h.DeleteWidget))
}
```

---

## `internal/httpx`, the only code that touches `*gin.Context`

| File | Function | Job |
|---|---|---|
| `bind.go` | `Handle[Req,Resp](status, fn)` | Builds `Req` from the URI (`uri:` tags), the query (`form:` tags) and the JSON body (`json:` tags). Calls `fn(ctx, req)` and writes `Resp` with `status`. |
| `auth.go` | `Authenticate(v)` | Reads `Authorization: Bearer`, verifies it with `auth.Verifier` (ES256, issuer, audience, expiry), and puts the `email` claim in ctx. Any failure is `ErrUnauthenticated`, so 401. |
| `deadline.go` | `Deadline(d)` | Ends every request's ctx after `REQUEST_TIMEOUT`. The cache eviction hold is derived from it. |
| `errors.go` | `WriteError(c, err)` | Maps `errors.Is(err, domain.ErrX)` to a status and a code. Logs once: 5xx at error level, 4xx at info level. Replaces every 5xx message with the status text. |

To move to Fiber or plain `net/http`, rewrite these two files and `router.go`. Handlers, DTOs and every layer below stay the same.

---

## `internal/handler` and `handler/dto`, the transport adapter

`handler.go` holds the one `Handler` type and `New(svc port.Services)`.
Each endpoint has its own file, `<resource>_<verb>.go`, with one method named `<Verb><Resource>`.

```go
// handler/widget_create.go
func (h *Handler) CreateWidget(ctx context.Context, req dto.CreateWidgetRequest) (dto.CreateWidgetResponse, error) {
	in, err := req.ToDomain(h.loc) // 1. wire to domain: parse dates in API_TIME_ZONE, map enums
	if err != nil {
		return dto.CreateWidgetResponse{}, err
	}
	in.Actor = auth.ActorFrom(ctx) // the verified token's email, never the body
	id, err := h.svc.Widget.Create(ctx, in) // 2. call one service method
	if err != nil {
		return dto.CreateWidgetResponse{}, err
	}
	return dto.CreateWidgetResponse{WidgetID: id}, nil // 3. domain to wire
}
```

**DTO file rules:**
- Struct tags appear here only: `json`, `form`, `uri`, `example`.
- `ToDomain()` parses wire formats and returns `ErrInvalidInput` for malformed input. `XFromDomain()` formats them.
- Create and update share the embedded `WidgetBody`, so the two request shapes cannot drift apart.
- Fields that the service computes (`data_src`, `status`, the running `no`) appear in responses only. Requests use `ComponentInput`, which has no `data_src`.
- Dates are parsed and formatted in `handler.Config.WireZone`, from `API_TIME_ZONE`, never in `time.Local`.

**Must not:** import repositories, compute business values, or call more than one service method. When an endpoint needs two calls, write a new use case.

Swagger annotations (`// @Summary ...`) sit on the handler method. Generate them with `make swag`.

---

## `internal/core/domain`, the model

**Does:** holds the types, the constructors that enforce invariants, derived values, error kinds and events.
**Imports:** the standard library only.

```go
func NewWidget(id string, in WidgetInput, now time.Time) (Widget, error) // create
func (w Widget) Revise(in WidgetInput, now time.Time) (Widget, error)    // update; keeps ID and Created*
func (w Widget) StatusAt(now time.Time) Status                           // derived, never stored
```

| Type | Purpose |
|---|---|
| `Widget` | the aggregate as stored |
| `WidgetInput` | what a caller may set: no ID, no audit fields |
| `WidgetView` | the read model: `Widget` plus the derived `Status` |
| `Page`, `PageResult[T]` | paging, with defaults and limits in `Normalize()` |
| `WidgetEvent` | what the core emits |

The domain has no struct tags, no driver types such as `pgtype.*`, and no `time.Now()`.

---

## `internal/core/port`, the interfaces

| File | Contains | Implemented by |
|---|---|---|
| `port.go` | the package doc and `Repositories`, the bundle that `repository.New` returns | — |
| `service.go` | driving ports (`WidgetService`) and `Services` | `core/service` |
| `database.go` | `Repo`, the database: it embeds `Transactor` and one `<Res>Repository` per resource | `postgres.Repo`, `memory.Repo` |
| `adapter.go` | `RedisRepo` (`IDGenerator`, `WidgetCache`, `EventPublisher`), `Adapter` (`CatalogAdapter`), and `Clock` | `redis.Repo`, `memory.RedisRepo`, `catalog.Client`, `catalog.AllowAll`, `platform.SystemClock` |

Port design checklist:
- Parameters and results are domain types or primitives.
- The doc comment names the domain error kinds that each method returns.
- No parameter names a collection, a table, a key, a topic or a URL.
- The method name states intent: `ProductListExists`, not `GetProductList` followed by a nil check.
- A best-effort port returns no error. `WidgetCache` and `EventPublisher` log their own failures, so the service has nothing to drop.

---

## `internal/core/service`, the use cases

`service.go` declares the parts that every resource shares:
- `Config`, with one section per resource (`Widget WidgetConfig`). Business settings only. `WidgetConfig.DataSrc` has the type `helper.WidgetDataSrc`, which its helper owns. `config.Load` rejects a template whose `%s` count is wrong.
- `Deps`, which embeds `port.Repositories` (`Repo`, `RedisRepo`, `Adpt`) and adds `Clock`.
- `deps`, the unexported copy that every resource's struct embeds, so use cases call `s.repo`, `s.redisRepo`, `s.adpt` and `s.clock`.
- `New(cfg, deps) port.Services`, plus one unexported struct per resource (`widgetService{cfg, deps}`).

Each use case has its own file, `<resource>_<verb>.go`, in this order:

```go
func (s *widgetService) Create(ctx context.Context, in domain.WidgetInput) (string, error) {
	if err := helper.CheckWidgetInput(ctx, s.adpt, in); err != nil { // 1. validate + preconditions, no side effects
		return "", err
	}
	id, err := s.redisRepo.NextWidgetID(ctx) // 2. allocate
	if err != nil {
		return "", fmt.Errorf("allocate widget id: %w", err)
	}
	now := s.clock.Now()
	in.Components = helper.ResolveWidgetDataSrc(s.cfg.DataSrc, id, in.Components)
	w, err := domain.NewWidget(id, in, now) // 3. build through the domain
	if err != nil {
		return "", err
	}
	if err := s.repo.CreateWidget(ctx, w); err != nil { // 4. persist
		return "", fmt.Errorf("create widget %s: %w", id, err)
	}
	s.redisRepo.PublishWidgetEvent(context.WithoutCancel(ctx), domain.WidgetEvent{Type: domain.EventWidgetCreated, WidgetID: id, Actor: in.Actor, At: now}) // 5. emit
	return id, nil
}
```

`service/helper` holds the functions that two or more use cases call, one function per file: `widget_check_input.go` and `widget_resolve_data_src.go`.
A helper is a plain function. It receives the port or the config it uses as an argument, and it imports only `domain` and `port`.
A function that one use case calls stays in that use case's file.

A read-modify-write runs inside `s.repo.WithinTx(ctx, func(ctx context.Context) error { ... })` and uses the inner `ctx`.
Cache invalidation and events run after the transaction commits, with `context.WithoutCancel(ctx)`.

**Must not:** import `gin`, `net/http`, drivers, `config`, or a repository package. Must not log errors.

---

## `internal/repository`, every outbound call

`repository.go` is the driven layer's wiring point. `New` starts from the in-process stand-ins.
For each non-empty config section it opens the client and replaces the stand-ins that the client backs.

| Section | Empty | Set |
|---|---|---|
| `Postgres.DSN` | `memory.Repo` | `postgres.Repo` |
| `Redis.Addr` | `memory.RedisRepo` | `redis.Repo` |
| `Catalog.BaseURL` | `catalog.AllowAll` | `catalog.Client` |

**File anatomy per store:**

| File | Holds |
|---|---|
| `<store>.go` | `Config`, `Open(ctx, cfg)` that pings, `mapErr` (driver codes to domain kinds), and the names the store owns (tables, keys, streams) |
| `<resource>.go` | the queries, the `<resource>Record` storage shape or row scan, and the repository methods |
| `tx.go` | `Repo.WithinTx` and the private `txFrom(ctx)` helper |
| `<capability>.go` | the other ports that the same store implements: `counter.go`, `cache.go`, `publisher.go` |

A constructor takes the client plus the settings it needs and returns the concrete type.
`repository.New` assigns it to `r.Repo`, `r.RedisRepo` or `r.Adpt`, and the compiler checks every method.

An HTTP API that another team owns is a repository too, `repository/catalog`.
Its `Config` holds the base URL, the key and the timeout. It drains response bodies and maps status codes to domain kinds.
Retries and circuit breakers live there, never in the service.

Datasource details are in [08-datasources.md](08-datasources.md).

---

## `internal/auth`, caller identity

`auth.New(Config)` reads the ES256 public key from `AUTH_JWT_PUBLIC_KEY_FILE`. `Verifier.Actor(token)` accepts only ES256, with the configured issuer and audience and an expiry, and returns the `email` claim.
`WithActor` and `ActorFrom` carry that email in ctx from `httpx.Authenticate` to the handler. `cmd/devtoken` makes a local keypair (`make dev-auth`, into git-ignored `.auth/`) and signs local tokens (`make token`).

---

## `internal/platform`, runtime wrappers

`platform` holds types that implement ports by calling the runtime: `SystemClock`, or a UUID generator.
They do no I/O, so `server.go` passes them to `service.Deps` directly.
