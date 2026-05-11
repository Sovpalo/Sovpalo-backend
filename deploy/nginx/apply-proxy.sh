#!/usr/bin/env bash
set -Eeuo pipefail

if ! command -v nginx >/dev/null 2>&1; then
  echo "nginx is not installed; skipping reverse proxy update."
  exit 0
fi

if [[ "${EUID:-$(id -u)}" -ne 0 ]]; then
  echo "error: deploy/nginx/apply-proxy.sh must run as root on the server." >&2
  exit 1
fi

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"
NGINX_DIR="${ROOT_DIR}/deploy/nginx"

install -D -m 644 "${NGINX_DIR}/upgrade-map.conf" /etc/nginx/conf.d/sovpalo-upgrade-map.conf
install -D -m 644 "${NGINX_DIR}/sovpalo-api.conf" /etc/nginx/sites-available/sovpalo-api.conf
ln -sf /etc/nginx/sites-available/sovpalo-api.conf /etc/nginx/sites-enabled/sovpalo-api.conf

if [[ -f /etc/nginx/sites-enabled/default ]]; then
  rm -f /etc/nginx/sites-enabled/default
fi

nginx -t
systemctl reload nginx

echo "nginx reverse proxy updated for WebSocket support."
