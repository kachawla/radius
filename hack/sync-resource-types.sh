#!/bin/bash

# ============================================================================
# Sync Resource Types from resource-types-contrib Repository
# ============================================================================
# This script syncs resource type YAML files from the resource-types-contrib
# repository based on the configuration fetched from .radius-sync-config.yaml
# in the resource-types-contrib repository

set -euo pipefail

cleanup() {
  if [[ -n "${TEMP_DIR:-}" && -d "${TEMP_DIR}" ]]; then
    rm -rf "${TEMP_DIR}"
  fi
}

trap cleanup EXIT

readonly SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
readonly REPO_ROOT="$(cd "${SCRIPT_DIR}/.." && pwd)"
readonly DEFAULT_SOURCE_REPO="radius-project/resource-types-contrib"
readonly DEFAULT_SOURCE_BRANCH="main"
readonly CONFIG_FILE_NAME=".radius-sync-config.yaml"

TEMP_DIR=""
CHANGES_DETECTED=false
CONFIG_FILE=""
SOURCE_REPO="${DEFAULT_SOURCE_REPO}"
SOURCE_BRANCH="${DEFAULT_SOURCE_BRANCH}"

usage() {
  echo "Usage: $(basename "$0") [OPTIONS]"
  echo "Sync resource types from resource-types-contrib repository"
  echo ""
  echo "Options:"
  echo "  --source-repo REPO    Source repository (default: radius-project/resource-types-contrib)"
  echo "  --source-branch BRANCH Source branch (default: main)"
  echo "  --dry-run             Show what would be synced without making changes"
  echo "  -h, --help            Show this help message"
  exit 0
}

log_info() {
  echo "[INFO] $*"
}

log_error() {
  echo "[ERROR] $*" >&2
}

check_dependencies() {
  local missing_deps=()
  
  for cmd in curl jq yq; do
    if ! command -v "$cmd" &> /dev/null; then
      missing_deps+=("$cmd")
    fi
  done
  
  if [[ ${#missing_deps[@]} -gt 0 ]]; then
    log_error "Missing required dependencies: ${missing_deps[*]}"
    log_error "Please install them and try again"
    exit 1
  fi
}

fetch_config() {
  log_info "Fetching configuration from ${SOURCE_REPO}@${SOURCE_BRANCH}..."
  
  local config_url="https://raw.githubusercontent.com/${SOURCE_REPO}/${SOURCE_BRANCH}/${CONFIG_FILE_NAME}"
  
  TEMP_DIR="$(mktemp -d)"
  CONFIG_FILE="${TEMP_DIR}/${CONFIG_FILE_NAME}"
  
  if ! curl -sf "${config_url}" -o "${CONFIG_FILE}"; then
    log_error "Failed to fetch config file from ${config_url}"
    log_error "Make sure ${CONFIG_FILE_NAME} exists in the ${SOURCE_REPO} repository"
    exit 1
  fi
  
  log_info "Configuration file downloaded successfully"
}

parse_config() {
  if [[ ! -f "${CONFIG_FILE}" ]]; then
    log_error "Config file not found: ${CONFIG_FILE}"
    exit 1
  fi
  
  # Get target directories - these should be defined in the config from upstream
  TARGET_DIR_COUNT=$(yq eval '.targetDirectories | length' "${CONFIG_FILE}")
  if [[ "${TARGET_DIR_COUNT}" == "0" || "${TARGET_DIR_COUNT}" == "null" ]]; then
    log_error "Invalid config file: no targetDirectories specified"
    exit 1
  fi
  
  log_info "Configuration loaded:"
  log_info "  Source: ${SOURCE_REPO}@${SOURCE_BRANCH}"
  log_info "  Target directories: ${TARGET_DIR_COUNT}"
}

fetch_resource_type() {
  local namespace="$1"
  local name="$2"
  local file="$3"
  
  local source_url="https://raw.githubusercontent.com/${SOURCE_REPO}/${SOURCE_BRANCH}/${namespace}/${name}/${file}"
  
  log_info "Fetching ${namespace}/${name}/${file}..."
  
  TEMP_DIR="$(mktemp -d)"
  local temp_file="${TEMP_DIR}/${file}"
  
  if ! curl -sf "${source_url}" -o "${temp_file}"; then
    log_error "Failed to fetch ${source_url}"
    return 1
  fi
  
  # Validate the YAML file
  if ! yq eval '.' "${temp_file}" > /dev/null 2>&1; then
    log_error "Invalid YAML in fetched file: ${source_url}"
    return 1
  fi
  
  # Get target directories count
  local target_dir_count
  target_dir_count=$(yq eval '.targetDirectories | length' "${CONFIG_FILE}")
  
  local changed=false
  
  # Copy to each target directory
  for ((j=0; j<target_dir_count; j++)); do
    local target_dir
    target_dir=$(yq eval ".targetDirectories[$j]" "${CONFIG_FILE}")
    local target_path="${REPO_ROOT}/${target_dir}/${file}"
    
    # Check if file has changed
    if [[ -f "${target_path}" ]]; then
      if diff -q "${temp_file}" "${target_path}" > /dev/null 2>&1; then
        log_info "  ${target_dir}/${file}: No changes"
      else
        log_info "  ${target_dir}/${file}: Changes detected - updating"
        cp "${temp_file}" "${target_path}"
        changed=true
      fi
    else
      log_info "  ${target_dir}/${file}: New file - creating"
      mkdir -p "$(dirname "${target_path}")"
      cp "${temp_file}" "${target_path}"
      changed=true
    fi
  done
  
  if [[ "${changed}" == "true" ]]; then
    CHANGES_DETECTED=true
  fi
  
  rm -rf "${TEMP_DIR}"
  TEMP_DIR=""
}

sync_resource_types() {
  local resource_type_count
  resource_type_count=$(yq eval '.resourceTypes | length' "${CONFIG_FILE}")
  
  if [[ "${resource_type_count}" == "0" || "${resource_type_count}" == "null" ]]; then
    log_error "No resource types configured in ${CONFIG_FILE}"
    exit 1
  fi
  
  log_info "Syncing ${resource_type_count} resource type(s)..."
  echo "============================================================================"
  
  for ((i=0; i<resource_type_count; i++)); do
    local namespace
    local name
    local file
    local target_file
    
    namespace=$(yq eval ".resourceTypes[$i].namespace" "${CONFIG_FILE}")
    name=$(yq eval ".resourceTypes[$i].name" "${CONFIG_FILE}")
    file=$(yq eval ".resourceTypes[$i].file" "${CONFIG_FILE}")
    
    fetch_resource_type "${namespace}" "${name}" "${file}" || {
      log_error "Failed to sync ${namespace}/${name}"
      exit 1
    }
  done
  
  echo "============================================================================"
  
  if [[ "${CHANGES_DETECTED}" == "true" ]]; then
    log_info "Sync completed with changes"
    exit 0
  else
    log_info "Sync completed - no changes detected"
    exit 0
  fi
}

main() {
  local dry_run=false
  
  while [[ $# -gt 0 ]]; do
    case $1 in
      --dry-run)
        dry_run=true
        shift
        ;;
      --source-repo)
        SOURCE_REPO="$2"
        shift 2
        ;;
      --source-branch)
        SOURCE_BRANCH="$2"
        shift 2
        ;;
      -h|--help)
        usage
        ;;
      *)
        log_error "Unknown option: $1"
        usage
        ;;
    esac
  done
  
  log_info "Starting resource type sync..."
  
  check_dependencies
  fetch_config
  parse_config
  sync_resource_types
  
  log_info "Resource type sync completed successfully"
}

main "$@"
