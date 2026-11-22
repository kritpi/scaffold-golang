# scaffold-go-next

**A boilerplate starter for Golang with Gin (Fiber-swappable) shipped with a complete written convention set, built for AI agents and humans.**

It's a hexagonal (ports and adapters) HTTP service on Postgres and Redis. A working `widget` CRUD resource serves as the reference that every new resource copies,
and the docs explain how each layer connects and how to reproduce the pattern step by step.
It's derived from a production Go/GCP service. [docs/convention/00-source-analysis.md](docs/convention/00-source-analysis.md) breaks that service down and records what the scaffold changed and why.

## Quick start

Start Postgres, Redis, the migrations and the API in Docker:

```bash
make up
```

Every `/v1` route needs a bearer token. `make token` signs one with the local key from `make dev-auth`:

```bash
TOKEN=$(make -s token EMAIL=ops@example.com)
```

```bash
curl -s -XPOST localhost:8080/scaffold-api/v1/widgets -H "Authorization: Bearer $TOKEN" -d '{"header_name_th":"hello","start_date":"01-01-2025","end_date":"31-12-2030","widgets":[{"type":"CATEGORY","is_active":true}]}'
```

```bash
curl -s localhost:8080/scaffold-api/v1/widgets/1 -H "Authorization: Bearer $TOKEN"
```

The token's `email` claim becomes `created_by` and `updated_by`. The request body carries no actor.

Run the checks:

```bash
make check
```

`make check` runs tidy, vet, `lint-layout` (the structure rules), `lint-config` (no config or secret in Go code), and the service-layer tests with the race detector. Each use case has one test file.
`make secrets` runs gitleaks over the tree. CI in `.github/workflows/check.yml` runs `make check` and gitleaks over the full history.
`make run` starts the same API without Docker. Empty `POSTGRES_DSN` and `REDIS_ADDR` select the in-memory stores.

| Datasource | Holds | Source |
|---|---|---|
| Postgres 17 | widgets and their components | [db/postgres/migrations](db/postgres/migrations), with the derived [db/postgres/ddl/schema.sql](db/postgres/ddl/schema.sql) |
| Redis 7.4 | the widget ID counter, the read cache and the event stream | [db/redis/keys.md](db/redis/keys.md) |

## Architecture in one picture

```mermaid
flowchart LR
  Client -->|HTTP| R[router] --> HX[httpx.Handle] --> H[handler, one file per endpoint]
  H -->|port.Services| S[core/service]
  S --> D[core/domain]
  S -->|port.Repositories| REPO[repository/*]
  REPO --> PG[(Postgres)]
  REPO --> RD[(Redis)]
  REPO --> EXT[catalog HTTP API]
  SRV[server.go] -. repository.New, service.New, handler.New .-> REPO & S & H
```

Dependencies point inward. `core` never imports `handler`, `repository`, `config`, `gin` or a driver.
The import command in the [review checklist](docs/convention/11-review-checklist.md#boundaries) prints any violation.

## Docs

The convention set lives in [docs/convention](docs/convention).

| # | Doc | Read it to… |
|---|---|---|
| 00 | [Source analysis](docs/convention/00-source-analysis.md) | understand the origin repo's structure, its flows, and which patterns the scaffold kept, changed or flagged |
| 01 | [Principles](docs/convention/01-principles.md) | learn the rules everything else follows |
| 02 | [Project structure](docs/convention/02-project-structure.md) | see every file, the naming, and where new code goes |
| 03 | [Layers](docs/convention/03-layers.md) | see what each layer does, what it imports and what it must not do, with code |
| 04 | [Request & data flow](docs/convention/04-request-flow.md) | trace endpoint → service → repository, including failures |
| 05 | [Models & errors](docs/convention/05-models-and-errors.md) | understand the DTO, domain and record split, the mapping and the error kinds |
| 06 | [Config & wiring](docs/convention/06-config-and-wiring.md) | follow env → config → one constructor per layer |
| 07 | [Testing](docs/convention/07-testing.md) | write a service test: one file per use case, the fixture, the fakes |
| 08 | [Datasources](docs/convention/08-datasources.md) | run the Postgres and Redis stack, change the schema, copy the Firestore, Pub/Sub, GCS and tracing recipes |
| 09 | [Jobs & monorepo](docs/convention/09-jobs-and-monorepo.md) | add Cloud Run jobs, consumers, a `go.work` layout and CI |
| 10 | [Recipes](docs/convention/10-recipes.md) | follow step-by-step checklists to reproduce any part |
| 11 | [Review checklist](docs/convention/11-review-checklist.md) | review a PR, human or agent |

Agents start at [AGENTS.md](AGENTS.md). `CLAUDE.md` imports it.

> The root `main.go` (hello world) predates the scaffold and can be deleted. The real entrypoint is `cmd/server`.
