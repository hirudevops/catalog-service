#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
ENV_FILE="${SCRIPT_DIR}/.dockerhub.env"

if [[ ! -f "${ENV_FILE}" ]]; then
  echo "Missing env file: ${ENV_FILE}" >&2
  exit 1
fi

# shellcheck disable=SC1090
source "${ENV_FILE}"

if [[ -z "${DOCKERHUB_USERNAME:-}" || -z "${DOCKERHUB_PASSWORD:-}" ]]; then
  echo "DOCKERHUB_USERNAME/DOCKERHUB_PASSWORD must be set in ${ENV_FILE}" >&2
  exit 1
fi

if [[ -n "${1:-}" ]]; then
  IMAGE_TAG="${1}"
else
  ts="$(date -u +%Y%m%d%H%M%S)"
  git_sha=""
  if command -v git >/dev/null 2>&1 && git rev-parse --is-inside-work-tree >/dev/null 2>&1; then
    git_sha="-g$(git rev-parse --short HEAD)"
  fi
  IMAGE_TAG="${DOCKERHUB_USERNAME}/catalog-service:${ts}${git_sha}"
fi

cd "${SCRIPT_DIR}"

echo "Logging in to Docker Hub as ${DOCKERHUB_USERNAME}..."
printf '%s' "${DOCKERHUB_PASSWORD}" | docker login -u "${DOCKERHUB_USERNAME}" --password-stdin

echo "Building image ${IMAGE_TAG}..."
docker build -t "${IMAGE_TAG}" .

echo "Pushing image ${IMAGE_TAG}..."
docker push "${IMAGE_TAG}"

echo "Done."
