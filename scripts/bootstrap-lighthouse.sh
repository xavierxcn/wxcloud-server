#!/usr/bin/env bash
set -euo pipefail

DEPLOY_PATH="${DEPLOY_PATH:-/opt/wxcloud-server}"
DEPLOY_USER="${DEPLOY_USER:-ubuntu}"

sudo apt-get update
sudo apt-get install -y ca-certificates curl git rsync

if ! command -v docker >/dev/null 2>&1; then
  if apt-cache show docker-ce >/dev/null 2>&1; then
    sudo apt-get install -y docker-ce docker-ce-cli containerd.io docker-buildx-plugin docker-compose-plugin
  else
    sudo apt-get install -y docker.io docker-compose-plugin
  fi
fi

sudo systemctl enable --now docker
sudo usermod -aG docker "${DEPLOY_USER}"

sudo mkdir -p "${DEPLOY_PATH}"
sudo chown -R "${DEPLOY_USER}:${DEPLOY_USER}" "${DEPLOY_PATH}"

if command -v ufw >/dev/null 2>&1; then
  sudo ufw allow 80/tcp || true
  sudo ufw allow 443/tcp || true
fi

echo "Lighthouse bootstrap complete. Reconnect SSH if docker group permissions are not active yet."
