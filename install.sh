#!/usr/bin/env bash
set -euo pipefail

APP_BIN="${APP_BIN:-spectre-updated}"
APP_PORT="${PORT:-8619}"
LOG_DIR="${LOG_DIR:-./logs}"
DATA_DIR="${DATA_DIR:-./data}"

if [ -t 1 ]; then
  COLOR_RED="$(printf '\033[31m')"
  COLOR_GREEN="$(printf '\033[32m')"
  COLOR_YELLOW="$(printf '\033[33m')"
  COLOR_BLUE="$(printf '\033[34m')"
  COLOR_BOLD="$(printf '\033[1m')"
  COLOR_RESET="$(printf '\033[0m')"
else
  COLOR_RED=""
  COLOR_GREEN=""
  COLOR_YELLOW=""
  COLOR_BLUE=""
  COLOR_BOLD=""
  COLOR_RESET=""
fi

log_info() {
  echo "${COLOR_BLUE}==>${COLOR_RESET} $*"
}

log_ok() {
  echo "${COLOR_GREEN}[OK]${COLOR_RESET} $*"
}

log_warn() {
  echo "${COLOR_YELLOW}[WARN]${COLOR_RESET} $*"
}

log_error() {
  echo "${COLOR_RED}[ERROR]${COLOR_RESET} $*"
}

show_help() {
  cat <<EOF
${COLOR_BOLD}Optional native dev helper for Spectre/Ghostbin${COLOR_RESET}

Usage:
  ./install.sh [command]

Commands:
  check         Verify required native dev tools are available
  setup         Prepare local project for native development
  assets        Install/build frontend assets only
  run           Run app in dev mode
  build         Build binary only
  help          Show this help

Examples:
  ./install.sh check
  ./install.sh setup
  ./install.sh assets
  ./install.sh run
  ./install.sh build

Notes:
- This script does NOT install Go, Node, npm, Python, or system packages.
- Install toolchains yourself using your distro package manager.
- Frontend assets are optional and only needed when frontend files changed.
- 'grunt' is checked via local npx resolution first, then global fallback.
EOF
}

have_cmd() {
  command -v "$1" >/dev/null 2>&1
}

print_install_hint() {
  local tool="$1"

  case "$tool" in
    go)
      log_error "Go is not installed or not available in PATH."
      cat <<'EOF'
What you should do:
- Install Go using your distro package manager or from https://go.dev/dl/
- Then open a new shell and re-run: ./install.sh check

Common package names:
- Debian/Ubuntu: golang-go
- Arch: go
EOF
      ;;
    node)
      log_error "Node.js is not installed or not available in PATH."
      cat <<'EOF'
What you should do:
- Install Node.js using your distro package manager or from https://nodejs.org/
- Then open a new shell and re-run: ./install.sh check

Common package names:
- Debian/Ubuntu: nodejs
- Arch: nodejs
EOF
      ;;
    npm)
      log_error "npm is not installed or not available in PATH."
      cat <<'EOF'
What you should do:
- Install npm using your distro package manager.
- Then open a new shell and re-run: ./install.sh check

Common package names:
- Debian/Ubuntu: npm
- Arch: npm
EOF
      ;;
    npx)
      log_warn "npx is not available in PATH."
      cat <<'EOF'
What this means:
- Frontend asset rebuild may not work from this script.
- Usually npx comes with npm on modern Node.js installs.

What you should do:
- Check your Node.js/npm installation
- Then re-run: ./install.sh check
EOF
      ;;
  esac
}

setup_dirs() {
  mkdir -p "$LOG_DIR" "$DATA_DIR"
}

check_toolchain() {
  local blocking_missing=0

  log_info "Checking native dev toolchain"
  echo

  if have_cmd go; then
    log_ok "go    -> $(go version)"
  else
    print_install_hint go
    echo
    blocking_missing=1
  fi

  if have_cmd node; then
    log_ok "node  -> $(node --version)"
  else
    print_install_hint node
    echo
    blocking_missing=1
  fi

  if have_cmd npm; then
    log_ok "npm   -> $(npm --version)"
  else
    print_install_hint npm
    echo
    blocking_missing=1
  fi

  if have_cmd npx; then
    log_ok "npx   -> available"
  else
    print_install_hint npx
    echo
  fi

  if have_cmd npx; then
    if npx --yes grunt --version >/dev/null 2>&1; then
      log_ok "grunt -> available via npx"
    elif have_cmd grunt; then
      log_ok "grunt -> available globally"
    else
      log_warn "grunt is not currently available"
      echo "This is only needed if you run: ./install.sh assets"
      echo "It may become available automatically after: npm install"
    fi
  elif have_cmd grunt; then
    log_ok "grunt -> available globally"
  else
    log_warn "grunt could not be checked yet"
    echo "This is only needed if you run: ./install.sh assets"
  fi

  echo

  if [ "$blocking_missing" -ne 0 ]; then
    log_error "Toolchain check failed."
    echo "Install the missing required tools above, then re-run: ./install.sh check"
    exit 1
  fi

  log_ok "All required core tools are available."
}

setup_native() {
  check_toolchain
  setup_dirs

  log_info "Preparing local directories"
  echo "LOG_DIR:  $LOG_DIR"
  echo "DATA_DIR: $DATA_DIR"

  log_info "Downloading Go modules"
  go mod download

  echo
  log_ok "Native dev setup complete."
  echo
  echo "Next steps:"
  echo "  - Run './install.sh assets' only if frontend assets changed"
  echo "  - Run './install.sh run' to start in dev mode"
  echo "  - Run './install.sh build' to build the binary"
}

build_assets() {
  if ! have_cmd npm; then
    print_install_hint npm
    exit 1
  fi

  if ! have_cmd npx; then
    print_install_hint npx
    echo
    log_error "Cannot build assets without npx."
    exit 1
  fi

  log_info "Installing frontend dependencies"
  npm install

  log_info "Building frontend assets"
  npx grunt

  log_ok "Frontend assets built."
}

run_dev() {
  if ! have_cmd go; then
    print_install_hint go
    exit 1
  fi

  setup_dirs

  log_info "Starting app in dev mode"
  echo "Port: $APP_PORT"
  echo "Logs: $LOG_DIR"
  echo "Data: $DATA_DIR"

  PORT="$APP_PORT" go run . -addr="0.0.0.0:$APP_PORT" -log_dir="$LOG_DIR" -root="$DATA_DIR" --logtostderr=1
}

build_bin() {
  if ! have_cmd go; then
    print_install_hint go
    exit 1
  fi

  setup_dirs

  log_info "Building binary: $APP_BIN"
  go build -o "$APP_BIN" .
  log_ok "Built: ./$APP_BIN"
}

CMD="${1:-help}"

case "$CMD" in
  check)
    check_toolchain
    ;;
  setup)
    setup_native
    ;;
  assets)
    build_assets
    ;;
  run)
    run_dev
    ;;
  build)
    build_bin
    ;;
  help|-h|--help)
    show_help
    ;;
  *)
    log_error "Unknown command: $CMD"
    echo
    show_help
    exit 1
    ;;
esac