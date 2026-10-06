#!/usr/bin/env bash
# Install or upgrade the Prism OAuth browser adapter (prism-adapter/) as a
# systemd sidecar of an existing Sub2API install.
#
# Re-running the script upgrades in place: the new adapter is built in a
# staging directory, swapped in, health-checked, and rolled back to the
# previous copy if it does not come up. OAuth sessions, pending journals and
# tool state under PRISM_STATE_DIR are never copied, moved or cleared.
#
# Only pinned wheels and Playwright's prebuilt Chromium are installed;
# nothing is compiled on the host.
set -Eeuo pipefail
umask 022

PRISM_SOURCE_DIR=${PRISM_SOURCE_DIR:-}
PRISM_INSTALL_DIR=${PRISM_INSTALL_DIR:-/opt/sub2api/prism-adapter}
PRISM_STATE_DIR=${PRISM_STATE_DIR:-/var/lib/sub2api-prism}
PRISM_ENV_FILE=${PRISM_ENV_FILE:-/etc/sub2api-prism.env}
PRISM_SERVICE_USER=${PRISM_SERVICE_USER:-sub2api}
PRISM_GATEWAY_SERVICE=${PRISM_GATEWAY_SERVICE:-sub2api}
PRISM_PYTHON=${PRISM_PYTHON:-python3.12}
# The port is written when given explicitly or when the env file has none.
PRISM_PORT_EXPLICIT=${PRISM_PORT+x}
PRISM_PORT=${PRISM_PORT:-8319}
# Initial GATEWAY_PRISM_BROWSER_ENABLED; an existing value is kept.
PRISM_GATEWAY_ENABLED=${PRISM_GATEWAY_ENABLED:-true}
# Use an already installed Chromium instead of Playwright's download.
PRISM_CHROME=${PRISM_CHROME:-}
PRISM_CHROME_SANDBOX=${PRISM_CHROME_SANDBOX:-}
# auto: run "playwright install-deps" only when apt-get is available.
PRISM_INSTALL_DEPS=${PRISM_INSTALL_DEPS:-auto}
# Restarting the gateway needs a deployment window and a binary rollback
# copy, so it only happens when explicitly requested.
PRISM_RESTART_GATEWAY=${PRISM_RESTART_GATEWAY:-no}
PRISM_KEEP_BACKUPS=${PRISM_KEEP_BACKUPS:-2}

die() { echo "ERROR: $*" >&2; exit 1; }
warn() { echo "WARNING: $*" >&2; }
info() { echo "==> $*"; }

# Paths are substituted into the systemd templates with sed.
bad_path_re='[#|&[:space:]\\]'
for path_var in PRISM_INSTALL_DIR PRISM_STATE_DIR PRISM_ENV_FILE; do
  [[ ${!path_var} == /?* ]] || die "$path_var must be an absolute path"
  [[ ! ${!path_var} =~ $bad_path_re ]] || die "$path_var contains unsupported characters"
done
PRISM_INSTALL_DIR=${PRISM_INSTALL_DIR%/}
PRISM_STATE_DIR=${PRISM_STATE_DIR%/}
[[ $PRISM_PORT =~ ^[0-9]+$ ]] && [ "$PRISM_PORT" -ge 1024 ] && [ "$PRISM_PORT" -le 65535 ] \
  || die "PRISM_PORT must be 1024..65535"
[[ $PRISM_KEEP_BACKUPS =~ ^[0-9]+$ ]] || die "PRISM_KEEP_BACKUPS must be a non-negative integer"
case "$PRISM_GATEWAY_ENABLED" in true|false) ;; *) die "PRISM_GATEWAY_ENABLED must be true or false" ;; esac
case "$PRISM_RESTART_GATEWAY" in yes|no) ;; *) die "PRISM_RESTART_GATEWAY must be yes or no" ;; esac
case "$PRISM_INSTALL_DEPS" in auto|yes|no) ;; *) die "PRISM_INSTALL_DEPS must be auto, yes or no" ;; esac

[ "$(id -u)" -eq 0 ] || die "run as root"
[ "$(uname -s)" = Linux ] || die "the Prism adapter only supports Linux"
for cmd in systemctl flock tar find sort cmp sed od; do
  command -v "$cmd" >/dev/null 2>&1 || die "$cmd is required"
done
id "$PRISM_SERVICE_USER" >/dev/null 2>&1 \
  || die "service user $PRISM_SERVICE_USER does not exist; install Sub2API first"
service_group=$(id -gn "$PRISM_SERVICE_USER")
command -v "$PRISM_PYTHON" >/dev/null 2>&1 || die "$PRISM_PYTHON is required (Python 3.12)"
"$PRISM_PYTHON" -c 'import sys; raise SystemExit(sys.version_info[:2] != (3, 12))' \
  || die "$PRISM_PYTHON is not Python 3.12"
"$PRISM_PYTHON" -c 'import venv, ensurepip' 2>/dev/null \
  || die "$PRISM_PYTHON lacks venv/ensurepip (Debian/Ubuntu: apt-get install python3.12-venv)"

if [ -z "$PRISM_SOURCE_DIR" ]; then
  script_dir=$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)
  # Repository checkout (deploy/..) or release package (next to the script).
  for candidate in "$script_dir/../prism-adapter" "$script_dir/prism-adapter"; do
    if [ -f "$candidate/server.py" ]; then
      PRISM_SOURCE_DIR=$(cd "$candidate" && pwd)
      break
    fi
  done
fi
[ -n "$PRISM_SOURCE_DIR" ] || die "prism-adapter source not found; set PRISM_SOURCE_DIR"
for required in server.py requirements.txt sub2api-prism-adapter.service sub2api-prism.conf; do
  [ -f "$PRISM_SOURCE_DIR/$required" ] || die "missing $PRISM_SOURCE_DIR/$required"
done
if [ -d "$PRISM_INSTALL_DIR" ] \
  && [ "$(cd "$PRISM_SOURCE_DIR" && pwd -P)" = "$(cd "$PRISM_INSTALL_DIR" && pwd -P)" ]; then
  die "PRISM_SOURCE_DIR must not be the install directory"
fi

unit_name=sub2api-prism-adapter.service
unit_path=/etc/systemd/system/$unit_name
dropin_dir=/etc/systemd/system/${PRISM_GATEWAY_SERVICE}.service.d
dropin_path=$dropin_dir/sub2api-prism.conf
install_parent=$(dirname "$PRISM_INSTALL_DIR")
install_base=$(basename "$PRISM_INSTALL_DIR")
staging_dir="$install_parent/.${install_base}.staging"
backup_root="$install_parent/.${install_base}.backups"
backup_dir=""
had_previous=0

install -d -m 0755 "$install_parent"
exec 9>"$install_parent/.${install_base}.lock"
flock -n 9 || die "another Prism adapter install is running"

cleanup() { rm -rf -- "$staging_dir"; }
trap cleanup EXIT

# Read KEY=value from an env file without echoing secrets.
env_get() {
  [ -f "$2" ] || return 0
  sed -n "s/^$1=//p" "$2" | tail -n 1
}

# Replace or append KEY=value, keeping the file root-owned and 0600.
env_set() {
  local key=$1 value=$2 file=$3 tmp
  tmp=$(mktemp "$file.XXXXXX")
  if [ -f "$file" ]; then
    grep -v "^$key=" "$file" >"$tmp" || true
  fi
  printf '%s=%s\n' "$key" "$value" >>"$tmp"
  chown root:root "$tmp"
  chmod 0600 "$tmp"
  mv -f "$tmp" "$file"
}

# Fill in the systemd templates for the configured paths and service user.
render_template() {
  sed -e "s|/opt/sub2api/prism-adapter|$PRISM_INSTALL_DIR|g" \
      -e "s|/var/lib/sub2api-prism|$PRISM_STATE_DIR|g" \
      -e "s|/etc/sub2api-prism.env|$PRISM_ENV_FILE|g" \
      -e "s|^User=.*|User=$PRISM_SERVICE_USER|" \
      -e "s|^Group=.*|Group=$service_group|" \
      "$1"
}

adapter_port() {
  local port
  port=$(env_get PRISM_ADAPTER_PORT "$PRISM_ENV_FILE")
  echo "${port:-8319}"
}

# /health only proves the HTTP process is up, not OAuth or model access.
health_ok() {
  "$PRISM_INSTALL_DIR/venv/bin/python" - "$(adapter_port)" <<'PY' >/dev/null 2>&1
import sys, urllib.request
opener = urllib.request.build_opener(urllib.request.ProxyHandler({}))
with opener.open(f"http://127.0.0.1:{sys.argv[1]}/health", timeout=3) as resp:
    raise SystemExit(resp.status != 200)
PY
}

wait_healthy() {
  local attempt
  for attempt in $(seq 1 30); do
    if systemctl is-active --quiet "$unit_name" && health_ok; then
      return 0
    fi
    sleep 2
  done
  return 1
}

backup_file() {
  if [ -f "$1" ]; then
    cp -p "$1" "$backup_dir/$2"
  fi
}

restore_file() {
  if [ -f "$backup_dir/$2" ]; then
    cp -p "$backup_dir/$2" "$1"
  fi
}

prune_backups() {
  local old
  [ -d "$backup_root" ] || return 0
  find "$backup_root" -mindepth 1 -maxdepth 1 -type d -name '20*' | sort -r \
    | tail -n "+$((PRISM_KEEP_BACKUPS + 1))" | while read -r old; do
      rm -rf -- "$old"
    done
}

info "Staging adapter from $PRISM_SOURCE_DIR"
rm -rf -- "$staging_dir"
install -d -m 0755 "$staging_dir"
tar -C "$PRISM_SOURCE_DIR" \
  --exclude=venv --exclude=browsers --exclude=__pycache__ --exclude='*.pyc' \
  --exclude='.env*' --exclude=pending --exclude=tools \
  -cf - . | tar -C "$staging_dir" -xf -

"$PRISM_PYTHON" -m venv "$staging_dir/venv"
staged_python="$staging_dir/venv/bin/python"
# Wheels only: fail rather than fall back to a source build.
"$staged_python" -m pip install --disable-pip-version-check --no-input \
  --only-binary=:all: -r "$staging_dir/requirements.txt"
# Same startup dependency check as the adapter (tool bridge needs both).
"$staged_python" -c 'import playwright, jsonschema, lark'

pw() { "$staged_python" -m playwright "$@"; }

if [ "$PRISM_INSTALL_DEPS" = yes ] \
  || { [ "$PRISM_INSTALL_DEPS" = auto ] && command -v apt-get >/dev/null 2>&1; }; then
  info "Installing Chromium system libraries"
  pw install-deps chromium \
    || warn "playwright install-deps failed; install Chromium runtime libraries manually"
fi

find_staged() {
  find "$staging_dir/browsers" -type f "$@" 2>/dev/null | sort -V | tail -n 1
}

if [ -n "$PRISM_CHROME" ]; then
  [ -x "$PRISM_CHROME" ] || die "PRISM_CHROME is not executable: $PRISM_CHROME"
  chrome_path=$PRISM_CHROME
  sandbox_path=$PRISM_CHROME_SANDBOX
  if [ -z "$sandbox_path" ]; then
    for candidate in "$(dirname "$PRISM_CHROME")/chrome-sandbox" "$(dirname "$PRISM_CHROME")/chrome_sandbox"; do
      if [ -f "$candidate" ]; then sandbox_path=$candidate; break; fi
    done
  fi
  [ -n "$sandbox_path" ] || die "Chromium sandbox helper not found; set PRISM_CHROME_SANDBOX"
  owner_mode=$(stat -c '%U %a' "$sandbox_path")
  [ "$owner_mode" = "root 4755" ] \
    || die "$sandbox_path must be owned by root with mode 4755 (found: $owner_mode)"
else
  info "Installing Playwright Chromium headless shell"
  # Reuse the previous download when the pinned Playwright revision is unchanged.
  if [ -d "$PRISM_INSTALL_DIR/browsers" ]; then
    cp -a "$PRISM_INSTALL_DIR/browsers" "$staging_dir/browsers"
  fi
  export PLAYWRIGHT_BROWSERS_PATH="$staging_dir/browsers"
  pw install --only-shell chromium
  chrome_path=$(find_staged \( -name chrome-headless-shell -o -name headless_shell \) -perm -u+x)
  [ -n "$chrome_path" ] || die "Chromium headless shell not found after install"
  sandbox_path=$(find_staged \( -name chrome-sandbox -o -name chrome_sandbox \))
  if [ -z "$sandbox_path" ]; then
    # Some headless shell builds ship without the SUID helper; full Chromium has it.
    pw install chromium
    sandbox_path=$(find_staged \( -name chrome-sandbox -o -name chrome_sandbox \))
  fi
  [ -n "$sandbox_path" ] || die "Chromium sandbox helper not found; set PRISM_CHROME/PRISM_CHROME_SANDBOX"
fi

# Code is root-owned and world-readable; only the state directory is writable.
chown -R root:root "$staging_dir"
chmod -R go-w,a+rX "$staging_dir"
if [[ $sandbox_path == "$staging_dir"/* ]]; then
  chmod 4755 "$sandbox_path"
  if [ "$(basename "$sandbox_path")" = chrome_sandbox ]; then
    ln -sfn chrome_sandbox "$(dirname "$sandbox_path")/chrome-sandbox"
    sandbox_path="$(dirname "$sandbox_path")/chrome-sandbox"
  fi
fi
# Final paths after the staging directory is swapped in.
chrome_path=${chrome_path/#"$staging_dir"/$PRISM_INSTALL_DIR}
sandbox_path=${sandbox_path/#"$staging_dir"/$PRISM_INSTALL_DIR}

backup_dir="$backup_root/$(date +%Y%m%d-%H%M%S)"
install -d -m 0700 "$backup_root" "$backup_dir"
backup_file "$PRISM_ENV_FILE" env
backup_file "$unit_path" unit
backup_file "$dropin_path" dropin

if systemctl is-active --quiet "$unit_name"; then
  info "Stopping $unit_name"
  systemctl stop "$unit_name"
fi
if [ -d "$PRISM_INSTALL_DIR" ]; then
  had_previous=1
  mv "$PRISM_INSTALL_DIR" "$backup_dir/adapter"
fi
mv "$staging_dir" "$PRISM_INSTALL_DIR"
info "Installed adapter to $PRISM_INSTALL_DIR"

# Existing OAuth sessions, pending journals and tool state stay untouched.
install -d -o "$PRISM_SERVICE_USER" -g "$service_group" -m 0700 "$PRISM_STATE_DIR"

if [ ! -f "$PRISM_ENV_FILE" ]; then
  install -o root -g root -m 0600 /dev/null "$PRISM_ENV_FILE"
fi
chown root:root "$PRISM_ENV_FILE"
chmod 0600 "$PRISM_ENV_FILE"
gateway_env_before=$(grep '^GATEWAY_PRISM_BROWSER_' "$PRISM_ENV_FILE" | sort || true)

adapter_key=$(env_get PRISM_ADAPTER_API_KEY "$PRISM_ENV_FILE")
if [ "${#adapter_key}" -lt 32 ]; then
  # Fixed read length keeps pipefail away from SIGPIPE.
  adapter_key=$(od -vAn -N32 -tx1 /dev/urandom | tr -d ' \n')
  env_set PRISM_ADAPTER_API_KEY "$adapter_key" "$PRISM_ENV_FILE"
  info "Generated a new adapter bridge secret"
fi
if [ "$(env_get GATEWAY_PRISM_BROWSER_API_KEY "$PRISM_ENV_FILE")" != "$adapter_key" ]; then
  env_set GATEWAY_PRISM_BROWSER_API_KEY "$adapter_key" "$PRISM_ENV_FILE"
fi
if [ -n "$PRISM_PORT_EXPLICIT" ] || [ -z "$(env_get PRISM_ADAPTER_PORT "$PRISM_ENV_FILE")" ]; then
  env_set PRISM_ADAPTER_PORT "$PRISM_PORT" "$PRISM_ENV_FILE"
fi
base_url="http://127.0.0.1:$(adapter_port)/v1"
if [ "$(env_get GATEWAY_PRISM_BROWSER_BASE_URL "$PRISM_ENV_FILE")" != "$base_url" ]; then
  env_set GATEWAY_PRISM_BROWSER_BASE_URL "$base_url" "$PRISM_ENV_FILE"
fi
if [ -z "$(env_get GATEWAY_PRISM_BROWSER_ENABLED "$PRISM_ENV_FILE")" ]; then
  env_set GATEWAY_PRISM_BROWSER_ENABLED "$PRISM_GATEWAY_ENABLED" "$PRISM_ENV_FILE"
fi
env_set PRISM_ADAPTER_CHROME "$chrome_path" "$PRISM_ENV_FILE"
env_set CHROME_DEVEL_SANDBOX "$sandbox_path" "$PRISM_ENV_FILE"
env_set PRISM_ADAPTER_STATE_DIR "$PRISM_STATE_DIR" "$PRISM_ENV_FILE"
if [ "$(env_get PRISM_ADAPTER_CLIENT_TOOLS_ENABLED "$PRISM_ENV_FILE")" = false ]; then
  warn "PRISM_ADAPTER_CLIENT_TOOLS_ENABLED=false keeps the 6.1 Sol client tool bridge disabled"
fi
gateway_env_after=$(grep '^GATEWAY_PRISM_BROWSER_' "$PRISM_ENV_FILE" | sort || true)

render_template "$PRISM_INSTALL_DIR/sub2api-prism-adapter.service" >"$unit_path.tmp"
chmod 0644 "$unit_path.tmp"
mv -f "$unit_path.tmp" "$unit_path"

install -d -m 0755 "$dropin_dir"
dropin_changed=0
render_template "$PRISM_INSTALL_DIR/sub2api-prism.conf" >"$dropin_path.tmp"
chmod 0644 "$dropin_path.tmp"
if ! cmp -s "$dropin_path.tmp" "$dropin_path" 2>/dev/null; then
  dropin_changed=1
fi
mv -f "$dropin_path.tmp" "$dropin_path"

systemctl daemon-reload
systemctl enable "$unit_name" >/dev/null
info "Starting $unit_name"
systemctl restart "$unit_name" || true

if ! wait_healthy; then
  journalctl -u "$unit_name" -n 50 --no-pager >&2 || true
  if [ "$had_previous" -eq 0 ]; then
    die "adapter did not become healthy; the install is left in place for inspection"
  fi
  warn "adapter did not become healthy; rolling back to the previous install"
  systemctl stop "$unit_name" || true
  failed_dir="$backup_dir/failed-adapter"
  mv "$PRISM_INSTALL_DIR" "$failed_dir"
  mv "$backup_dir/adapter" "$PRISM_INSTALL_DIR"
  restore_file "$PRISM_ENV_FILE" env
  restore_file "$unit_path" unit
  restore_file "$dropin_path" dropin
  systemctl daemon-reload
  systemctl restart "$unit_name" || true
  if wait_healthy; then
    die "rolled back to the previous adapter; failed build kept at $failed_dir"
  fi
  die "rollback did not become healthy either; check journalctl -u $unit_name"
fi
info "Adapter is healthy on 127.0.0.1:$(adapter_port)"

gateway_unit=${PRISM_GATEWAY_SERVICE}.service
if ! systemctl cat "$gateway_unit" >/dev/null 2>&1; then
  warn "$gateway_unit not found; load $PRISM_ENV_FILE into the gateway environment manually"
elif [ "$dropin_changed" -eq 1 ] || [ "$gateway_env_before" != "$gateway_env_after" ]; then
  if [ "$PRISM_RESTART_GATEWAY" = yes ]; then
    info "Restarting $gateway_unit to load the Prism gateway settings"
    systemctl restart "$gateway_unit"
  else
    warn "$gateway_unit must be restarted to load the Prism gateway settings."
    warn "Keep a rollback copy of the gateway binary, then run: systemctl restart $gateway_unit"
  fi
fi

if [ "$had_previous" -eq 0 ] && [ -z "$(ls -A "$backup_dir")" ]; then
  rmdir "$backup_dir"
fi
prune_backups

echo
echo "Prism adapter:  $PRISM_INSTALL_DIR"
echo "State dir:      $PRISM_STATE_DIR"
echo "Env file:       $PRISM_ENV_FILE (root, 0600)"
echo "Chromium:       $chrome_path"
if [ "$had_previous" -eq 1 ]; then
  echo "Rollback copy:  $backup_dir"
fi
echo
echo "/health only proves the HTTP process is up. Enable Prism on an OAuth account"
echo "and run the admin account test before routing real traffic."
