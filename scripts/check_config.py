"""Config and secrets come from the environment, read only in internal/config. Run: make lint-config"""
import glob
import re
import sys

SECRET_NAME = re.compile(r"(?i)(dsn|passw(or)?d|secret|token|api_?key|private_?key|credential)")
ENV_READ = re.compile(r"\b(os\.(Getenv|LookupEnv|Environ)|syscall\.Getenv)\b")
STRING = re.compile(r'"((?:[^"\\\n]|\\.)*)"')
ENDPOINT = [
    ("URL with a scheme", re.compile(r"^[a-z][a-z0-9+.-]*://")),
    ("host:port", re.compile(r"^[\w.-]*[a-z][\w.-]*:\d{2,5}$", re.I)),
    ("IP address", re.compile(r"\b\d{1,3}(\.\d{1,3}){3}\b")),
]
ASSIGN = re.compile(r'(?i)\b(\w*(passw(or)?d|secret|token|api_?key)\w*)\s*(:=|=|:)\s*"([^"]+)"')
DEFAULT = re.compile(r'e\.\w+\("(\w+)",\s*"([^"]*)"')



def strip_comment(line):
    quote = None
    i = 0
    while i < len(line):
        c = line[i]
        if quote:
            if c == "\\" and quote != "`":
                i += 1
            elif c == quote:
                quote = None
        elif c in "\"'`":
            quote = c
        elif line.startswith("//", i):
            return line[:i]
        i += 1
    return line


def blank_raw_strings(line):
    return re.sub(r"`[^`]*`", "``", line)


problems = []
for path in sorted(glob.glob("cmd/**/*.go", recursive=True) + glob.glob("internal/**/*.go", recursive=True)):
    in_config = path.startswith("internal/config/")
    for no, line in enumerate(open(path), 1):
        code = blank_raw_strings(strip_comment(line))
        where = f"{path}:{no}"
        if ENV_READ.search(code) and not in_config:
            problems.append(f"{where}: reads the environment outside internal/config")
        for m in ASSIGN.finditer(code):
            problems.append(f"{where}: {m.group(1)} is assigned a literal secret")
        if in_config:
            for key, default in DEFAULT.findall(code):
                if SECRET_NAME.search(key) and default:
                    problems.append(f"{where}: secret {key} has a non-empty default")
            continue
        if path.endswith("_test.go"):
            continue
        for lit in STRING.findall(code):
            for what, rx in ENDPOINT:
                if rx.search(lit):
                    problems.append(f"{where}: {what} {lit!r} belongs in a config key")

for p in problems:
    print(p)
print(f"{len(problems)} problem(s)")
sys.exit(1 if problems else 0)
