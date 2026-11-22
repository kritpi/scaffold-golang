# Redis keys

Redis has no DDL. This table is the whole keyspace. The names live as constants in `internal/repository/redis/redis.go`; change both together.

| Key | Type | Written by | Read by | Lifetime |
|---|---|---|---|---|
| `scaffold:widget-id` | string (integer) | `Counter.NextWidgetID` (`INCR`) | same | permanent; reset only with the Postgres volume |
| `scaffold:widget-cache:<id>` | string (JSON `widgetRecord`) | `CacheWidget` | `GetCachedWidget` | `WIDGET_CACHE_TTL` (must be positive); replaced by `evicted` for `EvictHold` (2 × `REQUEST_TIMEOUT`) after an update or delete commits |
| `scaffold:widget-events` | stream | `Publisher.PublishWidgetEvent` (`XADD`, `MAXLEN ~ 10000`) | consumers (`XREADGROUP`) | trimmed to about 10000 entries |

Stream entry fields: `type` (`widget.created`, `widget.updated`, `widget.deleted`), `widget_id`, `actor`, `at` (RFC 3339, UTC).
