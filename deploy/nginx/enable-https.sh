#!/usr/bin/env bash
set -Eeuo pipefail

DOMAIN="${DOMAIN:-}"
EMAIL="${EMAIL:-}"
API_PORT="${API_PORT:-8000}"

if [[ -z "${DOMAIN}" ]]; then
  echo "usage: DOMAIN=api.example.com EMAIL=you@example.com ./deploy/nginx/enable-https.sh" >&2
  exit 1
fi

if [[ "${EUID:-$(id -u)}" -ne 0 ]]; then
  echo "error: run as root on the server." >&2
  exit 1
fi

if ! command -v nginx >/dev/null 2>&1; then
  echo "error: nginx is not installed." >&2
  exit 1
fi

if ! command -v certbot >/dev/null 2>&1; then
  echo "error: certbot is not installed." >&2
  exit 1
fi

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"
TEMPLATE="${ROOT_DIR}/deploy/nginx/sovpalo-api-https.conf.template"
TARGET="/etc/nginx/sites-available/sovpalo-api-https.conf"

sed \
  -e "s/__DOMAIN__/${DOMAIN}/g" \
  -e "s/__API_PORT__/${API_PORT}/g" \
  "${TEMPLATE}" > "${TARGET}"

ln -sf "${TARGET}" /etc/nginx/sites-enabled/sovpalo-api-https.conf
if [[ -f /etc/nginx/sites-enabled/sovpalo-api.conf ]]; then
  rm -f /etc/nginx/sites-enabled/sovpalo-api.conf
fi

nginx -t
systemctl reload nginx

CERTBOT_ARGS=(certonly --webroot -w /var/www/certbot -d "${DOMAIN}" --agree-tos --non-interactive)
if [[ -n "${EMAIL}" ]]; then
  CERTBOT_ARGS+=(--email "${EMAIL}")
else
  CERTBOT_ARGS+=(--register-unsafely-without-email)
fi

mkdir -p /var/www/certbot
certbot "${CERTBOT_ARGS[@]}"
nginx -t
systemctl reload nginx

echo "HTTPS enabled for https://${DOMAIN}"
echo "Use https://${DOMAIN}/telegram/webapp in BotFather and TELEGRAM_WEBAPP_URL."
