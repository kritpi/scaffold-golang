"""Structure rules from docs/convention/11-review-checklist.md. Run: make lint-layout"""
import glob
import os
import re
import subprocess
import sys

root = sys.argv[1]
os.chdir(root)
fails = []


def check(goal, ok, detail):
    print(f"[{'PASS' if ok else 'FAIL'}] {goal}: {detail}")
    if not ok:
        fails.append(goal)


def uncommented(path):
    return re.sub(r"//[^\n]*", "", open(path).read())


def top_funcs(text):
    return re.findall(r"^func (?:\([^)]*\) )?(\w+)", text, re.M)


tests = sorted(glob.glob("**/*_test.go", recursive=True))
outside = [t for t in tests if os.path.dirname(t) != "internal/core/service"]
check("1 tests live only in internal/core/service", not outside, outside or f"{len(tests)} test files")
counts = {t: len(re.findall(r"^func Test\w+\(t \*testing\.T\)", uncommented(t), re.M)) for t in tests}
bad = {t: n for t, n in counts.items() if n > 1 or (n == 0 and not t.endswith("fixture_test.go"))}
check("1 one Test function per test file (fixtures hold none)", not bad, bad or {os.path.basename(k): v for k, v in counts.items()})

helpers = sorted(glob.glob("internal/core/service/helper/*.go"))
multi = {h: top_funcs(uncommented(h)) for h in helpers if len(top_funcs(uncommented(h))) != 1}
check("2 one function per helper file", helpers and not multi, multi or [os.path.basename(h) for h in helpers])
usecases = [f for f in glob.glob("internal/core/service/*.go") if not f.endswith("_test.go") and not f.endswith("service.go")]
shared = {}
for f in usecases:
    for name in top_funcs(uncommented(f)):
        if name[0].islower():
            users = [g for g in usecases if g != f and re.search(rf"\b{name}\(", uncommented(g))]
            if users:
                shared[f"{os.path.basename(f)}:{name}"] = users
check("2 no unexported function in a use-case file is called from another use-case file", not shared, shared or "none shared")

deps = subprocess.run(["go", "list", "-f", "{{.ImportPath}} {{join .Imports \" \"}}", "./..."], capture_output=True, text=True, check=True).stdout
drivers = ("github.com/jackc/pgx", "github.com/redis/go-redis", "net/http")
leaks = []
for line in deps.splitlines():
    pkg, *imps = line.split()
    rel = pkg.removeprefix("github.com/kritpi/scaffold-golang/")
    allowed_http = rel.startswith(("internal/httpx", "internal/router", "internal/server", "internal/handler"))
    for imp in imps:
        if imp.startswith(drivers) and not rel.startswith("internal/repository") and not (imp == "net/http" and allowed_http):
            leaks.append(f"{rel} -> {imp}")
        if rel.startswith("internal/core") and "/internal/" in imp and "/internal/core/" not in imp:
            leaks.append(f"{rel} -> {imp}")
check("3 drivers and outbound clients are imported only by internal/repository; core imports only core", not leaks, leaks or "clean")
check("3 no publish helper remains in core/service", not re.search(r"func \(s \*\w+\) publish\(", "".join(uncommented(f) for f in usecases)), "publish lives in repository/*/publisher.go")

fields = sorted({(os.path.basename(f), m) for f in usecases for m in re.findall(r"\bs\.(\w+)\.", uncommented(f))})
stray = [x for x in fields if x[1] not in ("repo", "redisRepo", "adpt", "clock", "cfg")]
check("3 use cases call s.repo, s.redisRepo, s.adpt or s.clock", not stray, stray or sorted({m for _, m in fields}))
ports = {os.path.basename(f): re.findall(r"^type (\w+) ", uncommented(f), re.M) for f in glob.glob("internal/core/port/*.go")}
want = {"port.go": ["Repositories"], "service.go": ["Services"], "database.go": ["Repo"], "adapter.go": ["RedisRepo", "Adapter"]}
check("3 port splits into service.go, database.go, adapter.go", set(ports) == set(want) and all(set(v) <= set(ports[k]) for k, v in want.items()), ports)

hfiles = sorted(f for f in glob.glob("internal/handler/*.go") if not f.endswith("handler.go"))
per = {os.path.basename(f): re.findall(r"^func \(h \*Handler\) (\w+)", uncommented(f), re.M) for f in hfiles}
check("4 one endpoint per handler file", all(len(v) == 1 for v in per.values()), per)
routes = re.findall(r"httpx\.Handle\([^,]+, h\.(\w+)\)", uncommented("internal/router/router.go"))
check("4 every route maps to a handler file", sorted(routes) == sorted(sum(per.values(), [])), routes)

svc = uncommented("internal/core/service/service.go")
check("5 one shared Deps embeds port.Repositories, and every service embeds deps", re.search(r"type Deps struct \{\s*port\.Repositories", svc) is not None and re.search(r"type widgetService struct \{[^}]*\n\tdeps\n", svc) is not None, "service.Deps -> deps")
check("5 no per-resource deps or service-constructor file", not glob.glob("internal/core/service/*_service.go") and "WidgetDeps" not in svc, "service.go only")

srv = uncommented("internal/server/server.go")
calls = {c: len(re.findall(rf"\b{c}\(", srv)) for c in ("repository.New", "service.New", "handler.New", "router.Setup")}
check("6 server.go calls each layer constructor once", all(n == 1 for n in calls.values()), calls)
check("6 server.go names no resource", not re.search(r"(?i)widget", srv), "no 'widget' in server.go")

print("ALL PASS" if not fails else f"FAILED: {fails}")
sys.exit(1 if fails else 0)
