# 07 — Testing

## Scope: the service layer only

Unit tests live in `internal/core/service` and nowhere else.
A use case is where the rules, the order and the side effects meet. A test there exercises the domain, the ports and the mapping in one go.

| Layer | How it is checked |
|---|---|
| `core/domain`, `core/service/helper` | through the use cases that call them: `StatusAt` in `TestWidgetGet`, `Validate` in `TestWidgetCreate`, `Revise` in `TestWidgetUpdate`, `helper.ResolveWidgetDataSrc` in `TestWidgetCreate` |
| `core/service` | `widget_<verb>_test.go`, one per use case |
| `repository/*`, `handler`, `httpx`, `router`, `auth` | `make up`, then the smoke run in [recipe C](10-recipes.md#c-add-a-real-storage-backend), with a token from `make token` |
| the dependency rule | review, with [11-review-checklist.md](11-review-checklist.md) |

`make check` runs `tidy`, `vet`, `lint-layout`, `lint-config` and `test` with the race detector. Work is done only when it passes.

## One file per test function

| File | Holds |
|---|---|
| `fixture_test.go` | `fixture`, `newFixture`, and every fake. All test files share it. |
| `<resource>_fixture_test.go` | the input that every test of the resource starts from: `validInput()` |
| `<resource>_<verb>_test.go` | exactly one `Test<Resource><Verb>`, with one `t.Run` per behaviour |

The test file mirrors the use-case file: `widget_create.go` has `widget_create_test.go` with `TestWidgetCreate`.
A second test function for the same use case is a `t.Run` inside the first one, never a new top-level function.

## The fixture

`newFixture` wires every port with a memory repository or a hand-written fake.
`f.newWidgetService()` builds the service from the current `f.deps`, so a test replaces one port and then calls it:

```go
f := newFixture(t)
f.deps.Repo = failingRepo{Repo: f.repo, err: domain.ErrUnavailable}

_, err := f.newWidgetService().Get(ctx, "1")
assert.ErrorIs(t, err, domain.ErrUnavailable)
```

| Fake | Port | Knobs |
|---|---|---|
| `memory.Repo`, `memory.RedisRepo` | `Repo`, `RedisRepo` | the real stand-ins |
| `fakeClock` | `Clock` | `now` |
| `fakeCatalog` | `Adapter` | `known`, `err`, `calls` |
| `recordingRedis` | `RedisRepo`: embeds `memory.RedisRepo` and records `PublishWidgetEvent` | `got` |
| `failingIDs`, `failingRepo` | embed the real `RedisRepo` or `Repo` and fail `NextWidgetID` or every widget method | `err` |

Name each `t.Run` after a behaviour, such as `"invalid input is rejected before any side effect"`. Never use a name such as `"error_case_2"`.

## What each use case covers

1. The success path. Assert the persisted state **and** the emitted event.
2. Each domain error kind that the use case can return.
3. Each dependency failure. Assert the wrap prefix and `errors.Is`.
4. Ordering guarantees. A rejected request consumes no ID, calls no catalog and emits nothing.
5. Cache behaviour, observed from outside. A write behind the service stays hidden by the cache. An update or delete makes the next read current.

Assert outcomes, not call sequences. The tests then survive refactors that keep the behaviour.
`fakeCatalog.calls` is the one counter, because "the catalog is not called" is part of the contract.

## Prove that the tests bite

Before you trust a new test, break the code it covers and watch it fail:

```bash
cd internal/core/service
cp widget_update.go widget_update.go.bak
sed -i '' 's/s.redisRepo.EvictWidget(ctx, id)/_ = id/' widget_update.go
go test -run TestWidgetUpdate .   # expect FAIL
mv widget_update.go.bak widget_update.go
```

## Time

Never call `time.Now()` in tests. Set `f.clock.now`, advance it, and build dates with `time.Date(..., time.UTC)`.
