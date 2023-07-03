#!/usr/bin/env bash
# Copyright 2022 The Kubernetes Authors.
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

DIR="$(realpath "${DIR}")"

source "${DIR}/suite.sh"

RELEASES=()

function usage() {
  echo "Usage: $0 <kube-version...>"
  echo "  <kube-version> is the version of kubernetes to test against."
}

function args() {
  if [[ $# -eq 0 ]]; then
    usage
    exit 1
  fi
  while [[ $# -gt 0 ]]; do
    RELEASES+=("${1}")
    shift
  done
}

function get_snapshot_info() {
  local name="${1}"
  kwokctl --name "${name}" kubectl get pod | awk '{print $1}'
  kwokctl --name "${name}" kubectl get node | awk '{print $1}'
}

function test_snapshot_etcd() {
  local name="${1}"
  local empty_info
  local full_info
  local restore_empty_info
  local restore_full_info
  local empty_path="./snapshot-empty-${name}"
  local full_path="./snapshot-full-${name}"

  empty_info="$(get_snapshot_info "${name}")"

  if [[ "${SKIP_DRY_RUN}" != "true" ]]; then
    got="$(kwokctl snapshot save --name "${name}" --path "${empty_path}" --format etcd --dry-run | clear_testdata "${name}")"
    want="$(<"${DIR}/testdata/${KWOK_RUNTIME}/snapshot_save_etcd.txt")"
    if [[ "${got}" != "${want}" ]]; then
      echo "------------------------------"
      diff -u <(echo "${want}") <(echo "${got}")
      echo "${got}" >"${DIR}/testdata/${KWOK_RUNTIME}/snapshot_save_etcd.txt"
      echo "Error: dry run snapshot save etcd failed"
      if [[ "${UPDATE_DRY_RUN_TESTDATE}" == "true" ]]; then
        echo "${got}" >"${DIR}/testdata/${KWOK_RUNTIME}/snapshot_save_etcd.txt"
      fi
      echo "------------------------------"
      echo "cat <<ALL >${DIR}/testdata/${KWOK_RUNTIME}/snapshot_save_etcd.txt"
      echo "${got}"
      echo "ALL"
      echo "------------------------------"
      return 1
    fi
  fi
  kwokctl snapshot save --name "${name}" --path "${empty_path}" --format etcd

  for ((i = 0; i < 120; i++)); do
    kubectl kustomize "${DIR}" | kwokctl --name "${name}" kubectl apply -f - && break
    sleep 1
  done

  for ((i = 0; i < 120; i++)); do
    full_info="$(get_snapshot_info "${name}")"
    if [[ "${full_info}" != "${empty_info}" && "${full_info}" =~ "default pod/" ]]; then
      break
    fi
    sleep 1
  done

  if [[ "${full_info}" == "${empty_info}" ]]; then
    echo "Error: Resource creation failed"
    return 1
  fi

  kwokctl snapshot save --name "${name}" --path "${full_path}" --format etcd

  sleep 1

  if [[ "${SKIP_DRY_RUN}" != "true" ]]; then
    got="$(kwokctl snapshot restore --name "${name}" --path "${empty_path}" --format etcd --dry-run | clear_testdata "${name}")"
    want="$(<"${DIR}/testdata/${KWOK_RUNTIME}/snapshot_restore_etcd.txt")"
    if [[ "${got}" != "${want}" ]]; then
      echo "------------------------------"
      diff -u <(echo "${want}") <(echo "${got}")
      echo "${got}" >"${DIR}/testdata/${KWOK_RUNTIME}/snapshot_restore_etcd.txt"
      echo "Error: dry run snapshot restore etcd failed"
      if [[ "${UPDATE_DRY_RUN_TESTDATE}" == "true" ]]; then
        echo "${got}" >"${DIR}/testdata/${KWOK_RUNTIME}/snapshot_restore_etcd.txt"
      fi
      echo "------------------------------"
      echo "cat <<ALL >${DIR}/testdata/${KWOK_RUNTIME}/snapshot_restore_etcd.txt"
      echo "${got}"
      echo "ALL"
      echo "------------------------------"
      return 1
    fi
  fi
  kwokctl snapshot restore --name "${name}" --path "${empty_path}" --format etcd
  for ((i = 0; i < 120; i++)); do
    restore_empty_info="$(get_snapshot_info "${name}")"
    if [[ "${empty_info}" == "${restore_empty_info}" ]]; then
      break
    fi
    sleep 1
  done

  if [[ "${empty_info}" != "${restore_empty_info}" ]]; then
    echo "Error: Empty snapshot restore failed"
    echo "Expected: ${empty_info}"
    echo "Actual: ${restore_empty_info}"
    return 1
  fi

  sleep 1

  kwokctl snapshot restore --name "${name}" --path "${full_path}" --format etcd
  for ((i = 0; i < 120; i++)); do
    restore_full_info=$(get_snapshot_info "${name}")
    if [[ "${full_info}" == "${restore_full_info}" ]]; then
      break
    fi
    sleep 1
  done

  if [[ "${full_info}" != "${restore_full_info}" ]]; then
    echo "Error: Full snapshot restore failed"
    echo "Expected: ${full_info}"
    echo "Actual: ${restore_full_info}"
    return 1
  fi

  rm -rf "${empty_path}" "${full_path}"
}

function test_snapshot_k8s() {
  local name="${1}"
  local full_info
  local restore_full_info
  local full_path="./snapshot-k8s-${name}"

  for ((i = 0; i < 120; i++)); do
    kubectl kustomize "${DIR}" | kwokctl --name "${name}" kubectl apply -f - && break
    sleep 1
  done

  for ((i = 0; i < 120; i++)); do
    full_info="$(get_snapshot_info "${name}")"
    if [[ "${full_info}" =~ "default pod/" ]]; then
      break
    fi
    sleep 1
  done

  kwokctl snapshot save --name "${name}" --path "${full_path}" --format k8s

  for ((i = 0; i < 120; i++)); do
    kubectl kustomize "${DIR}" | kwokctl --name "${name}" kubectl delete -f - && break
    sleep 1
  done

  for ((i = 0; i < 120; i++)); do
    restore_full_info="$(get_snapshot_info "${name}")"
    if [[ ! "${restore_full_info}" =~ "default pod/" ]]; then
      break
    fi
    sleep 1
  done

  kwokctl snapshot restore --name "${name}" --path "${full_path}" --format k8s

  for ((i = 0; i < 120; i++)); do
    restore_full_info="$(get_snapshot_info "${name}")"
    if [[ "${restore_full_info}" =~ "default pod/" ]]; then
      break
    fi
    sleep 1
  done

  if [[ "${full_info}" != "${restore_full_info}" ]]; then
    echo "Error: Full snapshot restore failed"
    echo "Expected: ${full_info}"
    echo "Actual: ${restore_full_info}"
    return 1
  fi

  rm -rf "${full_path}"
}

function main() {
  local failed=()
  for release in "${RELEASES[@]}"; do
    echo "------------------------------"
    echo "Testing snapshot on ${KWOK_RUNTIME} for ${release}"
    name="snapshot-cluster-${KWOK_RUNTIME}-${release//./-}"
    create_cluster "etcd-${name}" "${release}"
    test_snapshot_etcd "etcd-${name}" || failed+=("snapshot_etcd_${name}")
    delete_cluster "etcd-${name}"

    create_cluster "k8s-${name}" "${release}"
    test_snapshot_k8s "k8s-${name}" || failed+=("snapshot_k8s_${name}")
    delete_cluster "k8s-${name}"
  done

  if [[ "${#failed[@]}" -ne 0 ]]; then
    echo "------------------------------"
    echo "Error: Some tests failed"
    for test in "${failed[@]}"; do
      echo " - ${test}"
    done
    exit 1
  fi
}

args "$@"

main
