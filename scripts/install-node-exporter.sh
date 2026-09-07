#!/usr/bin/env bash
# Install/upgrade prometheus node_exporter on a gomall server via SSH.
# Usage: ./scripts/install-node-exporter.sh <ssh-alias> [version]
#   ssh-alias : gomall-1 | gomall-2 (entry in ~/.ssh/config)
#   version   : node_exporter release, default 1.9.1 (fallback 1.8.2)
set -euo pipefail

TARGET="${1:?usage: $0 <ssh-alias> [version]}"
VERSION="${2:-1.9.1}"
TARBALL="node_exporter-${VERSION}.linux-amd64.tar.gz"
URL="https://github.com/prometheus/node_exporter/releases/download/v${VERSION}/${TARBALL}"

echo "==> [${TARGET}] preflight: port 9100 must be free"
UPGRADED_FROM="$(ssh "$TARGET" "systemctl is-active node_exporter 2>/dev/null || true")"
if [ "$UPGRADED_FROM" = "active" ]; then
  echo "    node_exporter already active -> upgrade in place (stop first)"
  ssh "$TARGET" "systemctl stop node_exporter"
elif ssh "$TARGET" "ss -lntp | grep -q ':9100 '"; then
  echo "ERROR: ${TARGET}:9100 already in use by a foreign process."
  echo "处置路径：按 design Risks 调整 node_exporter 监听端口，并同步隧道(plist)/prometheus 配置——需人工确认后按新端口修订三处配置。"
  exit 1
else
  echo "    port 9100 free"
fi

echo "==> [${TARGET}] fetch node_exporter ${VERSION} (try remote download, fallback local+scp)"
if ssh "$TARGET" "curl -fsSL -m 90 -o /tmp/${TARBALL} '${URL}'"; then
  echo "    downloaded on server"
else
  echo "    server-side download failed -> local download + scp"
  curl -fsSL -m 180 -o "/tmp/${TARBALL}" "${URL}"
  scp -q "/tmp/${TARBALL}" "${TARGET}:/tmp/${TARBALL}"
fi

echo "==> [${TARGET}] install to /opt/node_exporter"
ssh "$TARGET" "set -e
  mkdir -p /opt/node_exporter
  tar -xzf /tmp/${TARBALL} -C /opt/node_exporter --strip-components=1
  rm -f /tmp/${TARBALL}
  cat > /etc/systemd/system/node_exporter.service <<'UNIT'
[Unit]
Description=Prometheus Node Exporter
After=network-online.target
Wants=network-online.target

[Service]
ExecStart=/opt/node_exporter/node_exporter --web.listen-address=127.0.0.1:9100
Restart=on-failure
RestartSec=3
NoNewPrivileges=true
ProtectHome=read-only

[Install]
WantedBy=multi-user.target
UNIT
  systemctl daemon-reload
  systemctl enable --now node_exporter
"

echo "==> [${TARGET}] verify"
sleep 2
ssh "$TARGET" "systemctl is-active node_exporter && curl -s 127.0.0.1:9100/metrics | grep -m1 node_load1"
echo "==> [${TARGET}] node_exporter ${VERSION} installed OK"
