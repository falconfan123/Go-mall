#!/usr/bin/env bash
# server-status-dashboard — 两台线上服务器状态 + 本地观测栈健康（Ghostty 右半 pane 用）
# 数据源：本地 Prometheus :9090 / Loki :3100 / Jaeger :16686（纯 HTTP API，无额外 agent）
# 降级判据（spec R4）：查询失败或无数据 → 从未成功过显示 DOWN；曾成功过显示 STALE(since 最后成功时间)
set -uo pipefail

PROM="http://localhost:9090"
LOKI="http://localhost:3100"
JAEG="http://localhost:16686"
INTERVAL="${1:-5}"

exec python3 - "$PROM" "$LOKI" "$JAEG" "$INTERVAL" <<'PY'
import json, sys, time, urllib.parse, urllib.request
from datetime import datetime

PROM, LOKI, JAEG, INTERVAL = sys.argv[1], sys.argv[2], sys.argv[3], float(sys.argv[4])
last_ok = {}  # source key -> epoch of last successful fetch

def http_json(url, timeout=6):
    req = urllib.request.Request(url, headers={"User-Agent": "gomall-dashboard/1.0"})
    with urllib.request.urlopen(req, timeout=timeout) as r:
        return json.loads(r.read().decode())

def mark_ok(key):
    last_ok[key] = time.time()

def mark_bad(key):
    t = last_ok.get(key)
    if t is None:
        return "DOWN"
    return "STALE(since %s)" % datetime.fromtimestamp(t).strftime("%H:%M:%S")

def prom_query(expr):
    """Return list of (instance, float value). Raises on failure."""
    url = PROM + "/api/v1/query?" + urllib.parse.urlencode({"query": expr})
    d = http_json(url)
    if d.get("status") != "success":
        raise RuntimeError("prom query failed")
    out = []
    for r in d["data"]["result"]:
        try:
            out.append((r["metric"].get("instance", ""), float(r["value"][1])))
        except (KeyError, ValueError):
            pass
    return out

def prom_scalar(expr, default=None):
    try:
        res = prom_query(expr)
        return res[0][1] if res else default
    except Exception:
        return default

def machine_lines(inst):
    key = "machine:" + inst
    up = prom_scalar('up{env="production",instance="%s"}' % inst)
    if up is None or up != 1:
        return ["  [%s] %s" % (inst, mark_bad(key))]
    try:
        cpu = prom_scalar('100 * (1 - avg by (instance)(rate(node_cpu_seconds_total{mode="idle",instance="%s"}[2m])))' % inst)
        mem = prom_scalar('100 * (1 - node_memory_MemAvailable_bytes{instance="%s"} / node_memory_MemTotal_bytes{instance="%s"})' % (inst, inst))
        rx = prom_scalar('sum by (instance)(rate(node_network_receive_bytes_total{instance="%s",device=~"eth0|ens.*"}[2m]))' % inst) or 0
        tx = prom_scalar('sum by (instance)(rate(node_network_transmit_bytes_total{instance="%s",device=~"eth0|ens.*"}[2m]))' % inst) or 0
        disk = prom_scalar('(1 - node_filesystem_avail_bytes{instance="%s",mountpoint="/",fstype!~"tmpfs.*"} / node_filesystem_size_bytes{instance="%s",mountpoint="/",fstype!~"tmpfs.*"}) * 100' % (inst, inst))
        load = prom_scalar('node_load1{instance="%s"}' % inst)
        if None in (cpu, mem, disk, load):
            raise RuntimeError("missing metric")
    except Exception:
        return ["  [%s] %s" % (inst, mark_bad(key))]
    mark_ok(key)
    return [
        "  [%s] UP" % inst,
        "      cpu %5.1f%%   mem %5.1f%%   disk %5.1f%%   load %4.2f" % (cpu, mem, disk, load),
        "      net  down %9.1f B/s   up %9.1f B/s" % (rx, tx),
    ]

def stack_lines():
    lines = []
    try:
        d = http_json(PROM + "/api/v1/targets")
        tg = d["data"]["activeTargets"]
        up = sum(1 for t in tg if t.get("health") == "up")
        mark_ok("stack:targets")
        lines.append("  targets: %d/%d UP" % (up, len(tg)))
    except Exception:
        lines.append("  targets: %s" % mark_bad("stack:targets"))
    try:
        q = urllib.parse.urlencode({"query": 'sum(count_over_time({env="development"} | json | level="error" [5m]))'})
        d = http_json(LOKI + "/loki/api/v1/query?" + q)
        res = d.get("data", {}).get("result", [])
        err = int(float(res[0]["value"][1])) if res else 0
        mark_ok("stack:loki")
        lines.append("  loki: %d error(5m)" % err)
    except Exception:
        lines.append("  loki: %s" % mark_bad("stack:loki"))
    try:
        http_json(JAEG + "/api/services")
        mark_ok("stack:jaeger")
        lines.append("  jaeger: UP")
    except Exception:
        lines.append("  jaeger: %s" % mark_bad("stack:jaeger"))
    return lines

def render():
    lines = ["── 线上服务器状态 (%s, 刷新 %ss) ──" % (datetime.now().strftime("%Y-%m-%d %H:%M:%S"), int(INTERVAL))]
    for inst in ("gomall-1", "gomall-2"):
        lines.extend(machine_lines(inst))
    lines.append("── 本地观测栈 ──")
    lines.extend(stack_lines())
    return lines

print("\033[2J\033[H", end="", flush=True)
while True:
    frame = render()
    print("\033[H\033[2J", end="")
    print("\n".join(frame), flush=True)
    time.sleep(INTERVAL)
PY
