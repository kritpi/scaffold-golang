# 04 — Request and data flow

## The shape

```mermaid
flowchart LR
  Client -->|HTTP| Router
  subgraph Transport
    Router --> Handle[httpx.Handle]
    Handle --> Handler[handler.Handler]
    Handler --> DTO[dto ToDomain / FromDomain]
  end
  subgraph Core
    Svc[service] --> Dom[domain]
  end
  subgraph Driven[repository, wired by repository.New]
    PG[repository/postgres]
    RD[repository/redis]
    CAT[repository/catalog]
  end
  Handler -->|port.Services.Widget| Svc
  Svc -->|WidgetRepository, Transactor| PG --> DB[(Postgres)]
  Svc -->|IDGenerator, WidgetCache, EventPublisher| RD --> Cache[(Redis: counter, cache, stream)]
  Svc -->|CatalogAdapter| CAT --> Ext[External HTTP API]
```

Data changes shape twice on the way in and twice on the way out:

```
JSON ──dto.ToDomain──▶ domain.WidgetInput ──domain.NewWidget──▶ domain.Widget ──SQL args──▶ widgets + widget_components rows
JSON ◀─dto.FromDomain─ domain.WidgetView  ◀──service adds Status── domain.Widget ◀──scanWidget── rows, or toDomain from the Redis widgetRecord
```

---

## Trace: `POST /scaffold-api/v1/widgets`

```mermaid
sequenceDiagram
  autonumber
  participant C as Client
  participant HX as httpx.Handle
  participant H as Handler
  participant S as widgetService
  participant CA as CatalogAdapter
  participant ID as IDGenerator
  participant R as WidgetRepository
  participant E as EventPublisher
  C->>HX: POST JSON body
  HX->>HX: ShouldBindJSON into dto.CreateWidgetRequest
  HX->>H: CreateWidget(ctx, req)
  H->>H: req.ToDomain() parses DD-MM-YYYY into time.Time
  H->>S: Create(ctx, domain.WidgetInput)
  S->>S: helper.CheckWidgetInput: in.Validate()
  S->>CA: ProductListExists(ctx, refID)
  CA-->>S: true
  S->>ID: NextWidgetID(ctx)
  ID-->>S: "42"
  S->>S: helper.ResolveWidgetDataSrc + domain.NewWidget(id, in, clock.Now())
  S->>R: Create(ctx, widget)
  R->>R: joinOrBeginTx: INSERT widgets + widget_components
  R-->>S: nil
  S->>E: PublishWidgetEvent(created), XADD. The publisher logs a failure.
  S-->>H: "42"
  H-->>HX: dto.CreateWidgetResponse{WidgetID:"42"}
  HX-->>C: 201 {"widget_id":"42"}
```

| Step | File | What can fail | Becomes |
|---|---|---|---|
| 2 | `httpx/bind.go` | malformed JSON | `ErrInvalidInput` → 400 |
| 4 | `handler/dto/widget.go` | bad date format | `ErrInvalidInput` → 400 |
| 6 | `core/domain/widget.go` | missing field, reversed dates, unknown type | `ErrInvalidInput` → 400 |
| 7 | `repository/catalog/http.go` | timeout, 5xx | `ErrUnavailable` → 503 |
| 7 | service | ref ID does not exist | `ErrInvalidInput` → 400 |
| 9 | `repository/redis/counter.go` | Redis down | `ErrUnavailable` → 503 |
| 12 | `repository/postgres/widget.go` | duplicate ID (`23505`) | `ErrConflict` → 409 |
| 12 | `repository/postgres/postgres.go` | Postgres down or restarting | `ErrUnavailable` → 503 |

---

## Trace: `PUT /scaffold-api/v1/widgets/:widget-id` (read-modify-write)

```mermaid
sequenceDiagram
  participant S as widgetService
  participant T as Transactor
  participant R as WidgetRepository
  S->>S: Validate + catalog check (outside the tx)
  S->>T: WithinTx(ctx, fn)
  T->>R: GetWidgetForUpdate(txCtx, id)
  R-->>T: current or ErrNotFound
  T->>T: current.Revise(in, now) keeps ID and Created*
  T->>R: Update(txCtx, next)
  T-->>S: commit / rollback
  S->>S: EvictWidget(id), then publish updated event
```

Keep external calls **out** of the transaction. Inside it, `GetWidgetForUpdate` runs `SELECT ... FOR UPDATE`, so a concurrent update waits.
The cache entry is evicted after the commit. A read between the two sees the old entry, and no read can store it after the eviction.

---

## Trace: `GET /scaffold-api/v1/widgets/:widget-id` (cache-aside)

1. `GetCachedWidget(ctx, id)`. A hit skips Postgres.
2. On a miss, `repo.GetWidget(ctx, id)` reads Postgres, and `CacheWidget` stores it for `WIDGET_CACHE_TTL`.
3. The service derives `Status` from `Clock.Now()` on every read, so a cached entry never holds a stale status.
4. With Redis down, every read is a miss. The API still answers, and the server logs `widget cache get failed`.

---

## Trace: `GET /scaffold-api/v1/widgets?page_id=2&page_size=10`

1. `httpx` binds the query into `dto.ListWidgetsRequest` (`form:` tags).
2. `req.ToDomain()` → `domain.Page{Number: 2, Size: 10}`.
3. The service calls `page.Normalize()`, so defaults and the max size live in one place.
4. `repo.List(ctx, page)` → `([]domain.Widget, total, error)`. The repo does the ordering and offset/limit natively.
5. The service builds `WidgetView{Status: w.StatusAt(now)}` for each item.
6. `dto.ListWidgetsFromDomain` adds the running `no`, formats dates and computes `last_page`.

---

## Trace: a job or message consumer

A job is a second driving adapter. It calls the **same** `port.Services` a handler would:

```mermaid
flowchart LR
  GCS[(GCS upload)] --> PS[[Pub/Sub]] --> WF[Workflow / Eventarc] --> JOB[cmd/job main]
  JOB --> JS[service.Import...] --> R[repository] --> DB[(Firestore)]
```

See [09-jobs-and-monorepo.md](09-jobs-and-monorepo.md).

---

## Cross-cutting concerns: where each one lives

| Concern | Where | How |
|---|---|---|
| Request context and cancellation | `httpx.Handle` passes `c.Request.Context()` | every port method takes `ctx` first |
| Tracing | middleware in `server.go` (otelgin) + a span per repository method | [08-datasources.md](08-datasources.md#recipe-tracing) |
| Logging | `slog` default set in `server.go`; errors logged in `httpx.WriteError`, best-effort failures in their repository | context-aware `slog.*Context` |
| Auth | `httpx.Authenticate(auth.Verifier)` on the `/v1` group in `router.go` verifies the ES256 bearer token and puts its `email` claim in ctx | the handler copies `auth.ActorFrom(ctx)` into `WidgetInput.Actor`, so the service never reads auth from ctx and no client sets the actor |
| Validation | shape in `dto.ToDomain`, rules in `domain.Validate` | — |
| Time zone | `API_TIME_ZONE` → `handler.Config.WireZone` → `dto` parse and format | domain and storage use `time.Time` |
