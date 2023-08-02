#!/usr/bin/env bash
# Copyright 2023 The Kubernetes Authors.
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#     http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

DIR="$(dirname "${BASH_SOURCE[0]}")"

ROOT_DIR="$(realpath "${DIR}/../..")"

export KWOK_LOGS_DIR="${ROOT_DIR}/logs"

function save_logs() {
  local name="${1}"
  shift 1
  kwokctl --name="${name}" export logs "${KWOK_LOGS_DIR}" "$@"
}

function create_cluster() {
  local name="${1}"
  local release="${2}"
  shift 2

  if ! KWOK_KUBE_VERSION="${release}" kwokctl \
    create cluster \
    --name "${name}" \
    --timeout 30m \
    --wait 30m \
    --quiet-pull \
    --disable-qps-limits \
    "$@"; then
    echo "Error: Cluster ${name} creation failed"
    exit 1
  fi
}

function delete_cluster() {
  local name="${1}"
  save_logs "${name}"
  if ! kwokctl delete cluster --name "${name}"; then
    echo "Error: Cluster ${name} deletion failed"
    exit 1
  fi
}

function child_timeout() {
  local to="${1}"
  shift
  "${@}" &
  local wp=$!
  local start=0
  while kill -0 "${wp}" 2>/dev/null; do
    if [[ "${start}" -ge "${to}" ]]; then
      kill "${wp}"
      echo "Error: Timeout ${to}s" >&2
      return 1
    fi
    ((start++))
    sleep 1
  done
  echo "Took ${start}s" >&2
}

function retry() {
  local times="${1}"
  shift
  local start=0
  while true; do
    if "${@}"; then
      return 0
    fi
    if [[ "${start}" -ge "${times}" ]]; then
      echo "Error: Retry ${times} times" >&2
      return 1
    fi
    ((start++))
    sleep 1
  done
}

GOOS="$(go env GOOS)"
GOARCH="$(go env GOARCH)"

function clear_testdata() {
  local name="${1}"

  sed '/^ *$/d' |
    sed "s|${ROOT_DIR}|<ROOT_DIR>|g" |
    sed "s|${HOME}|~|g" |
    sed 's|/root/|~/|g' |
    sed "s|${GOARCH}|<ARCH>|g" |
    sed "s|${GOOS}|<OS>|g" |
    sed "s|${name}|<CLUSTER_NAME>|g" |
    sed 's|\.tar\.gz|.<TAR>|g' |
    sed 's|\.zip|.<TAR>|g' |
    sed 's| --env=ETCD_UNSUPPORTED_ARCH=<ARCH> | |g' |
    sed 's| ETCD_UNSUPPORTED_ARCH=<ARCH> | |g'
}

function create_user() {
  local runtime="${1}"
  local name="${2}"
  local component="${3}"
  local uid="${4}"
  local username="${5}"
  local gid="${6}"
  local groupname="${7}"
  local home="${8}"
  local shell="${9}"
  container=kwok-"${name}"-"${component}"
  "${runtime}" exec "${container}" addgroup --gid "${gid}" "${groupname}"
  "${runtime}" exec "${container}" adduser -u "${uid}" -G "${groupname}" -h "${home}" -s "${shell}" "${username}" -D
}
