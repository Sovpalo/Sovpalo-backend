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
NGINX_DIR="${ROOT_DIR}/deploy/nginx"
TARGET="/etc/nginx/sites-available/sovpalo-api-https.conf"
BOOTSTRAP_TEMPLATE="${NGINX_DIR}/sovpalo-api-https-bootstrap.conf.template"
HTTPS_TEMPLATE="${NGINX_DIR}/sovpalo-api-https.conf.template"
CERT_FULLCHAIN="/etc/letsencrypt/live/${DOMAIN}/fullchain.pem"
CERT_PRIVKEY="/etc/letsencrypt/live/${DOMAIN}/privkey.pem"

if [[ ! -f "${BOOTSTRAP_TEMPLATE}" || ! -f "${HTTPS_TEMPLATE}" ]]; then
  echo "error: nginx HTTPS templates are missing in ${NGINX_DIR}." >&2
  echo "Update the repository on the server and run this script again." >&2
  exit 1
fi

render_template() {
  local template="$1"
  local output="$2"
  sed \
    -e "s/__DOMAIN__/${DOMAIN}/g" \
    -e "s/__API_PORT__/${API_PORT}/g" \
    "${template}" > "${output}"
}

install_http_bootstrap() {
  install -D -m 644 "${NGINX_DIR}/upgrade-map.conf" /etc/nginx/conf.d/sovpalo-upgrade-map.conf
  mkdir -p /var/www/certbot
  render_template "${BOOTSTRAP_TEMPLATE}" "${TARGET}"
  ln -sf "${TARGET}" /etc/nginx/sites-enabled/sovpalo-api-https.conf
  if [[ -f /etc/nginx/sites-enabled/sovpalo-api.conf ]]; then
    rm -f /etc/nginx/sites-enabled/sovpalo-api.conf
  fi
  if [[ -f /etc/nginx/sites-enabled/default ]]; then
    rm -f /etc/nginx/sites-enabled/default
  fi
  nginx -t
  systemctl reload nginx
}

issue_certificate() {
  local certbot_args=(certonly --webroot -w /var/www/certbot -d "${DOMAIN}" --agree-tos --non-interactive)
  if [[ -n "${EMAIL}" ]]; then
    certbot_args+=(--email "${EMAIL}")
  else
    certbot_args+=(--register-unsafely-without-email)
  fi
  certbot "${certbot_args[@]}"
}

install_https_proxy() {
  if [[ ! -f "${CERT_FULLCHAIN}" || ! -f "${CERT_PRIVKEY}" ]]; then
    echo "error: certificate files for ${DOMAIN} were not created." >&2
    echo "Check DNS for ${DOMAIN} and retry certbot after HTTP bootstrap is active." >&2
    exit 1
  fi

  render_template "${HTTPS_TEMPLATE}" "${TARGET}"
  nginx -t
  systemctl reload nginx
}

echo "Installing temporary HTTP config for ${DOMAIN} before certificate issuance."
install_http_bootstrap

echo "Requesting Let's Encrypt certificate for ${DOMAIN}."
issue_certificate

echo "Enabling HTTPS reverse proxy for ${DOMAIN}."
install_https_proxy

echo "HTTPS enabled for https://${DOMAIN}"
echo "Use https://${DOMAIN}/telegram/webapp in BotFather and TELEGRAM_WEBAPP_URL."
