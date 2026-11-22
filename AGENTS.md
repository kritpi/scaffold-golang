# AGENTS.md

Go HTTP service built **hexagonally**: `handler → core/service → port ← repository`.
Each layer has one constructor that covers every resource: `repository.New`, `service.New`, `handler.New`. `internal/server/server.go` calls them in that order.
The `widget` resource is the reference. Copy its shape. Don't invent a new one.

## Workflow for any change

1. **Locate.** Place each piece of new code with the flowchart in [docs/convention/02-project-structure.md](docs/convention/02-project-structure.md#where-does-new-code-go).
2. **Follow the recipe** for your change in [docs/convention/10-recipes.md](docs/convention/10-recipes.md):
   A endpoint · B resource · C storage backend · D external API · E event · F config key · G job · H new service · I schema change.
   Do every step of the recipe in order.
3. **Test** each use case you touched, in its own `internal/core/service/<resource>_<verb>_test.go`. Cover the success path, every error kind it can return, and every dependency failure ([docs/convention/07-testing.md](docs/convention/07-testing.md)).
4. **Done** means `make check` is green (it includes `lint-layout` and `lint-config`), `make secrets` finds nothing, `make up` serves the change, **and** every item in [docs/convention/11-review-checklist.md](docs/convention/11-review-checklist.md) holds for your diff.

## Invariants

- `internal/core` imports the standard library and other `internal/core` packages only: `domain` imports nothing, `port` imports `domain`, `service/helper` imports `domain` and `port`. `make lint-layout` fails on any violation.
  When it fails, move the code to the right layer. Change `scripts/check_layout.py` only with a design reason.
- One file per use case (`core/service/<resource>_<verb>.go`), per endpoint (`handler/<resource>_<verb>.go`) and per test function.
  A function that two or more use cases call moves to its own file, `core/service/helper/<resource>_<name>.go`, and takes its dependencies as arguments.
- Every outbound call (database, cache, HTTP API, broker) lives in `internal/repository/<store>` behind a port. A best-effort port returns no error and logs its own failure.
- `port` splits by kind: `service.go` (driving), `database.go` (`Repo`), `adapter.go` (`RedisRepo`, `Adapter`, `Clock`). Use cases call `s.repo.<Verb><Res>`, `s.redisRepo.<Method>`, `s.adpt.<Method>` and `s.clock.Now()`, nothing else.
- No env read, endpoint, IP or secret literal in Go code outside `internal/config`. Secrets have empty defaults and arrive as env.
- Each rule has one **owner** ([docs/convention/01-principles.md](docs/convention/01-principles.md#p2-every-piece-of-state-and-every-rule-has-one-owner)). Call the owner. Don't restate the rule.
- Ports speak intent in domain types. Table names, keys, streams and URLs stay inside their repository package.
- Errors: the repository maps driver codes to a `domain.Err*` kind, every layer wraps with `%w`, and `httpx.WriteError` logs once.
- Config is read once in `config.Load()`. Each layer constructor receives its own section.
- The actor is the verified bearer token's `email` claim (`auth.ActorFrom(ctx)`). No request body carries it.
- Time, IDs and transactions come from `port.Clock`, `port.IDGenerator` and `port.Transactor`.
- The Postgres schema changes only through a new `db/postgres/migrations` pair. `db/postgres/ddl/schema.sql` is regenerated with `make ddl`.

## Reference (read when the change touches it)

- Layer responsibilities, file anatomy, code shapes: [docs/convention/03-layers.md](docs/convention/03-layers.md)
- Request and data flow traces, failure points: [docs/convention/04-request-flow.md](docs/convention/04-request-flow.md)
- DTO, domain and record mapping, error kinds, envelopes: [docs/convention/05-models-and-errors.md](docs/convention/05-models-and-errors.md)
- Config sections, keys and wiring order: [docs/convention/06-config-and-wiring.md](docs/convention/06-config-and-wiring.md)
- Postgres, Redis, the catalog API, migrations, docker-compose, plus Firestore, Pub/Sub, GCS and tracing recipes: [docs/convention/08-datasources.md](docs/convention/08-datasources.md)
- Jobs, consumers, the `go.work` monorepo, CI: [docs/convention/09-jobs-and-monorepo.md](docs/convention/09-jobs-and-monorepo.md)
- Why the conventions differ from the origin repo: [docs/convention/00-source-analysis.md](docs/convention/00-source-analysis.md)
