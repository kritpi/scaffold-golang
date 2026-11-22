# 10 — Recipes: reproduce the pattern step by step

Each recipe is an ordered checklist. Do the steps in order, then run `make check`.
Open the referenced file and copy its shape.

Placeholders: `<Res>` is the resource in PascalCase (`Banner`), `<res>` is lower case (`banner`), and `<verb>` is `create`, `get`, `list`, `update`, `delete` or another verb.

---

## A. Add an endpoint to an existing resource

Example: `POST /v1/widgets/:widget-id/duplicate`.

1. **Domain**, only when a new rule exists: add the method or constructor to `internal/core/domain/<res>.go`.
2. **Port:** add the method to `<Res>Service` in `internal/core/port/service.go`. Use domain types only.
3. **Service:** create `internal/core/service/<res>_<verb>.go` with the use-case order:
   validate and check preconditions, allocate, build, persist, emit.
   Wrap errors with `%w`. Do not log an error that you return.
4. **Helpers:** when the new file calls a function from another use-case file, move that function to its own file, `service/helper/<res>_<name>.go`. Export it, and pass its dependencies as arguments.
5. **Service test:** create `<res>_<verb>_test.go` with one `Test<Res><Verb>`. Cover the success path, each error kind, each dependency failure and the ordering guarantees.
6. **DTO:** in `internal/handler/dto/<res>.go`, add `<Verb><Res>Request` with `uri`, `form` or `json` tags and `ToDomain()`. Add `<Verb><Res>Response` and its `FromDomain`.
7. **Handler:** create `internal/handler/<res>_<verb>.go` with one method, `func (h *Handler) <Verb><Res>(ctx, dto.XRequest) (dto.XResponse, error)`, and its swag annotations.
8. **Route:** add one line to `router.go`, such as `widgets.POST("/:widget-id/duplicate", httpx.Handle(http.StatusCreated, h.DuplicateWidget))`.
9. `make check`, then `make up` and call the endpoint.

## B. Add a new resource (aggregate)

1. `internal/core/domain/<res>.go`: the aggregate, `<Res>Input` with `Validate()`, `New<Res>`, `Revise`, and the derived methods.
2. `internal/core/port/service.go`: `<Res>Service`, plus a `<Res>` field in `Services`.
3. `internal/core/port/database.go`: a `<Res>Repository` with resource-named methods (`Create<Res>`, `Get<Res>`, `Get<Res>ForUpdate`, ...), embedded in `Repo`.
   In `adapter.go`, add `Next<Res>ID` to `IDGenerator`, and a `<Res>Cache` or a `Publish<Res>Event` to `RedisRepo` when the resource needs them. `port.Repositories` does not change.
4. `internal/core/service/service.go`: a `<Res>Config` section in `Config`, a `<res>Service{cfg, deps}` struct, and one line in `New`. Its use cases call `s.repo`, `s.redisRepo`, `s.adpt` and `s.clock`.
5. One file per use case, `<res>_<verb>.go`, plus one file per shared function in `service/helper`.
6. One test file per use case, `<res>_<verb>_test.go`, plus `<res>_fixture_test.go` for its `validInput`. Add the fakes to `fixture_test.go`.
7. `internal/repository/memory/<res>.go`: record, mapping, and the new methods on `memory.Repo` (and on `memory.RedisRepo` for IDs, cache and events). `repository.New` does not change.
8. Migrations: `make migrate-new NAME=create_<res>s`, then write the `.up.sql` and the `.down.sql`. Follow [recipe I](#i-change-the-postgres-schema).
9. `internal/repository/postgres/<res>.go`: queries, row scan, and the methods on `postgres.Repo`. Add the Redis methods to `redis.Repo`. The compiler reports any method a store is missing.
10. `internal/handler/dto/<res>.go`, then one `internal/handler/<res>_<verb>.go` per endpoint.
11. `router.go`: one route group.
12. `make check`, then `make up`, then the smoke run in [recipe C](#c-add-a-real-storage-backend).

`server.go` does not change.

## C. Add a real storage backend

1. Create `internal/repository/<store>/`. Copy the anatomy of `repository/postgres` or `repository/redis` ([08-datasources.md](08-datasources.md)).
2. `<store>.go`: `Config`, `Open(ctx, cfg)` that pings, `mapErr` from driver codes to domain kinds, and the names the store owns.
3. `<res>.go`: the queries or record struct, and the methods of the port the store backs (`port.Repo` or `port.RedisRepo`) on one `Repo` type. `New(client)` returns it.
4. `tx.go`, when the store backs `port.Repo`: `Repo.WithinTx`, which puts the transaction in ctx, and `txFrom(ctx)`. `Get<Res>ForUpdate` fails without it.
5. `repository.go`: add the section to `Config`, and a branch in `New` that opens the client, appends its close, and replaces `r.Repo` or `r.RedisRepo`.
6. `config.go`: add the env keys. Add them to `configs/local.env`, `configs/*.json` and the `server` service in `docker-compose.yml`.
7. `docker-compose.yml`: add the service with a healthcheck. Add it to `server.depends_on` with `condition: service_healthy`.
8. Run `make check`.
9. Run `make up` and the smoke run with `TOKEN=$(make -s token)`: create, get twice, update, list, delete, then get to see 404. A request without the token must get 401. Stop the new service with `docker compose stop <store>` and check that the API answers 503, not 500.

## D. Add an external API

1. `internal/core/port/adapter.go`: an intent-named `<System>Adapter` that speaks domain types and documents the error kinds it returns. Embed it in `Adapter`. Use cases call it as `s.adpt.<Method>`.
2. `internal/repository/<system>/http.go`: `Config{BaseURL, APIKey, Timeout}`, `New(cfg)`, and methods that build requests with `http.NewRequestWithContext`. Drain bodies and map status codes.
3. Add a stand-in such as `AllowAll`. `repository.New` uses it when the base URL is empty.
4. Add a fake to `fixture_test.go`, and cover the outage in each use case that calls the port.
5. Add the config section, the env keys and the branch in `repository.New`.
6. `make check`.

## E. Add an event

1. `domain/event.go`: a new `EventType` constant and its payload fields.
2. Service: call `s.redisRepo.PublishWidgetEvent(ctx, ...)` after the write commits. For guaranteed delivery, use the outbox in [08-datasources.md](08-datasources.md#delivery-guarantee-outbox).
3. `repository/redis/publisher.go`: map the new fields to stream values. Update `db/redis/keys.md`.
4. Test: assert `f.events.got` in the use case's test file.

## F. Add a config key

1. `internal/config/config.go`: read the key in `Load()` with `e.str`, `e.int`, `e.bool` or `e.dur` and a default.
2. Add the field to the consumer's own `Config` (`service.WidgetConfig`, `postgres.Config`, `redis.Config`, and so on). `server.go` needs no change.
3. `configs/local.env`, every `configs/<env>.json`, and `docker-compose.yml`.
4. Secrets go **only** into Secret Manager in the deploy config, never as values in `configs/`.

## G. Add a job (Cloud Run job)

1. Single module: `cmd/<job>/main.go` plus `server.Run<Job>(ctx)`, built from the same `repository.New` and `service.New`. Monorepo: copy the skeleton into `job/<job>/` ([09](09-jobs-and-monorepo.md)).
2. The job's service method takes its inputs as arguments and is idempotent.
3. Put the parser in a pure function, tested through the use case.
4. Add the trigger: `gcp-workflows/<job>/workflow.yaml`, or Cloud Scheduler.
5. Add a Dockerfile and a CI matrix entry.

## H. Start a new service from this scaffold

1. Copy the repo and change the module path in `go.mod`. Update the imports with `grep -rl kritpi/scaffold-golang . | xargs sed -i '' 's#github.com/kritpi/scaffold-golang#<new/module>#g'`.
2. Rename `BasePath` in `router.go`, `SERVICE_NAME` in `configs/`, the `name:` and the credentials in `docker-compose.yml`, and the key prefix in `repository/redis/redis.go`.
3. Add your first resource with recipe B. Then delete the widget files, the widget migrations and the widget rows in `db/redis/keys.md`.
4. `make reset`, then `make check` and `make up`.

## I. Change the Postgres schema

1. `make migrate-new NAME=<verb>_<object>` creates the next pair, `NNNNNN_<verb>_<object>.up.sql` and `.down.sql`.
2. Write the change in `.up.sql` and its exact reverse in `.down.sql`. Never edit a migration that has been applied in a shared environment. Add a new migration instead.
3. Update the queries in `internal/repository/postgres/<res>.go`, and the domain when the change is visible there.
4. Run `make up`, which applies the new migration. Then run `make migrate-down` and `make migrate-up` to prove that the pair reverses.
5. Run `make ddl` and commit `db/postgres/ddl/schema.sql` with the migration. The diff of `schema.sql` is what reviewers read. The `schema` CI job fails when the two disagree.
