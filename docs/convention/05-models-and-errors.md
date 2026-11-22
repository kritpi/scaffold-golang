# 05 — Models, mapping and errors

## Three models, three owners

| Model | Package | Tags | Changes when |
|---|---|---|---|
| **DTO** | `internal/handler/dto` | `json`, `form`, `uri` | the public API contract changes |
| **Domain** | `internal/core/domain` | none | business meaning changes |
| **Record / row** | `internal/repository/<store>` (unexported) | `json` on the Redis record. Postgres names columns in its SQL. | storage layout changes |

Each model changes for its own reason. Don't reuse one model for another's job, even when the fields match today.
A shared struct couples the API, the business model and the database, so a column rename turns into an API break.

### Mapping functions: names and placement

| Direction | Function | Lives in |
|---|---|---|
| wire → domain | `(r CreateWidgetRequest) ToDomain() (domain.WidgetInput, error)` | `dto` |
| domain → wire | `WidgetDetailFromDomain(v domain.WidgetView) WidgetDetailResponse` | `dto` |
| domain → storage | `toRecord(w domain.Widget) widgetRecord`, or the SQL arguments | repository |
| storage → domain | `(r widgetRecord) toDomain() domain.Widget`, or `scanWidget(row)` | repository |

Mapping is explicit field-by-field code. Don't use reflection mappers, and don't round-trip domain types through `json.Marshal`. The Redis cache marshals its own `widgetRecord`.
The compiler then flags every place a new field must be handled.

### Shared contract module (optional)

In a monorepo where other services import your request and response types (the source repo's `shared/model/v1`),
put the **DTOs** in a versioned module, `shared/model/v1`. `handler/dto` then declares
`type CreateWidgetRequest v1.CreateWidgetRequest` and adds the mapping methods. Never share domain types or records.

---

## Domain types

```go
type WidgetInput struct { Actor, HeaderTH string; StartDate, EndDate time.Time; Components []Component; ... }
func (in WidgetInput) Validate() error

type Widget struct { ID string; ...; CreatedAt time.Time; CreatedBy string; UpdatedAt time.Time; UpdatedBy string }
func NewWidget(id string, in WidgetInput, now time.Time) (Widget, error)
func (w Widget) Revise(in WidgetInput, now time.Time) (Widget, error)
func (w Widget) StatusAt(now time.Time) Status

type WidgetView struct { Widget; Status Status } // what reads return
```

Rules:
- Input types omit fields the caller must not set: ID and audit fields.
- Constructors and mutators take `now time.Time`. Domain code never calls `time.Now()`.
- Derived values are methods, never stored fields. Storing them means they go stale.
- Enums are typed strings (`type Status string`) with constants. `Validate` rejects unknown values.

---

## Errors

### Kinds (in `domain/errors.go`)

| Kind | Meaning | HTTP | Code |
|---|---|---|---|
| `ErrInvalidInput` | caller sent something wrong | 400 | `INVALID_REQUEST` |
| `ErrNotFound` | addressed thing does not exist | 404 | `NOT_FOUND` |
| `ErrUnauthenticated` | no valid bearer token | 401 | `UNAUTHENTICATED`, with `WWW-Authenticate: Bearer` |
| `ErrConflict` | state conflict: duplicate, version mismatch | 409 | `CONFLICT` |
| `ErrUnavailable` | a dependency is down or timed out; retry may work | 503 | `UNAVAILABLE` |
| anything else | bug or unexpected failure | 500 | `INTERNAL` (message hidden) |

Add a kind only when callers need to react to it differently (e.g. `ErrForbidden` → 403).

### Who does what

```
driver error ──▶ repository: map to kind ──▶ service: wrap with context ──▶ httpx: classify + log once
  pgx.ErrNoRows       fmt.Errorf("%w: widget %s", domain.ErrNotFound, id)
                                         fmt.Errorf("get widget %s: %w", id, err)
                                                                    errors.Is(err, domain.ErrNotFound) → 404
```

```go
// repository: translate
if errors.Is(err, pgx.ErrNoRows) { // internal/repository/postgres/widget.go
	return domain.Widget{}, fmt.Errorf("%w: widget %s", domain.ErrNotFound, id)
}

// service: add context, keep the chain
if err != nil {
	return fmt.Errorf("update widget %s: %w", id, err)
}

// domain: create
return fmt.Errorf("%w: end_date is before start_date", domain.ErrInvalidInput)
```

**Never:**
- `strings.Contains(err.Error(), "NotFound")`. Use the driver's code (`status.Code`, `pgconn.PgError.Code`, `redis.Nil`).
- `log.Errorf(...)` followed by `return err` in the same function. That produces duplicate logs.
- Returning driver errors to clients. `WriteError` replaces every 5xx message with its status text, such as `service unavailable`.
- `%v` when wrapping. Use `%w`, or `errors.Is` stops working.

### Response envelopes

```json
// success with payload: the response DTO itself
{"widget_id": "42"}

// command success without payload
{"code": "0000", "message": "successfully updated"}

// any error
{"code": "NOT_FOUND", "message": "get widget 9: not found: widget 9"}
```
