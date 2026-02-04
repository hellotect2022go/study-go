python - <<'PY'
import os, random, string, time
path = "fake.log"
target_size = 1024 * 1024 *1024  # 1MB
levels = ["INFO", "DEBUG", "WARN", "ERROR"]
messages = [
    "User login successful",
    "User logout",
    "Cache miss for key",
    "Cache hit for key",
    "Request processed",
    "DB connection opened",
    "DB connection closed",
    "File uploaded",
    "File download started",
    "Heartbeat",
    "Retrying request",
    "Timeout occurred",
    "Validation passed",
    "Validation failed",
    "Config reloaded",
]

def random_ip():
    return ".".join(str(random.randint(1, 254)) for _ in range(4))

def random_user():
    return "user" + str(random.randint(1000, 9999))

with open(path, "w", encoding="utf-8") as f:
    size = 0
    while size < target_size:
        ts = time.strftime("%Y-%m-%d %H:%M:%S")
        level = random.choice(levels)
        msg = random.choice(messages)
        req_id = "".join(random.choices(string.ascii_lowercase + string.digits, k=8))
        line = f"{ts} [{level}] {msg} user={random_user()} ip={random_ip()} req_id={req_id}\n"
        f.write(line)
        size += len(line.encode("utf-8"))

# Trim to exactly 1MB if over
with open(path, "rb+") as f:
    f.truncate(target_size)

print(path)
PY