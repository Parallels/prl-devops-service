#!/bin/bash
set -euo pipefail

# ---------------------------------------------------------------------------
# ssh_install.sh — remote wrapper for install.sh
#
# Downloads install.sh from GitHub and executes it on a remote host via SSH.
# Usage:
#   curl -sSL https://raw.githubusercontent.com/Parallels/prl-devops-service/main/scripts/ssh_install.sh \
#     | bash -s -- --host 192.168.1.100 --ssh-user admin --ssh-key ~/.ssh/id_rsa --install
# ---------------------------------------------------------------------------

INSTALL_SH_URL="https://raw.githubusercontent.com/Parallels/prl-devops-service/main/scripts/install.sh"

# ---------------------------------------------------------------------------
# SSH-specific parameters (consumed locally)
# ---------------------------------------------------------------------------
SSH_HOST=""
SSH_USER="${USER:-root}"
SSH_PORT="22"
SSH_KEY=""
SSH_PASSWORD=""
SUDO_PASSWORD=""
NO_STRICT_HOST_KEY_CHECKING="false"

# ---------------------------------------------------------------------------
# Forwarded parameters (passed through to install.sh on the remote)
# ---------------------------------------------------------------------------
FORWARD_ARGS=()

# ---------------------------------------------------------------------------
# Temp file — cleaned up on exit
# ---------------------------------------------------------------------------
TEMP_SCRIPT="$(mktemp /tmp/prl-ssh-install-$$.XXXXXX.sh)"
chmod 600 "$TEMP_SCRIPT"

cleanup() {
  rm -f "$TEMP_SCRIPT"
}
trap cleanup EXIT

# ---------------------------------------------------------------------------
# Helpers
# ---------------------------------------------------------------------------
usage() {
  cat >&2 <<EOF
Usage: $0 [SSH OPTIONS] [INSTALL OPTIONS]

SSH options (consumed locally):
  -H, --host <host>                 Remote hostname or IP (required)
  -l, --ssh-user <user>             SSH login username (default: \$USER)
  -P, --ssh-port <port>             SSH port (default: 22)
  -k, --ssh-key <path>              Path to private key file
      --ssh-password <pass>         SSH password (requires sshpass; prefer key auth)
      --sudo-password <pass>        sudo password on the remote host
      --no-strict-host-key-checking Skip SSH host key verification

Install options (forwarded to install.sh on the remote):
  -i, --install                     Install prldevops (default)
  -u, --uninstall                   Uninstall prldevops
  -U, --update                      Update prldevops
  -p, --path <path>                 Installation destination path
  -v, --version <version>           Specific version to install
      --no-service                  Do not install/manage the system service
      --std-user                    Install without sudo (standard user)
      --pre-release                 Allow pre-release versions
      --modules <modules>           Comma-separated list of modules
      --api-port <port>             API port for the prldevops service
  -r, --root-password <pass>        Root password for the prldevops service
EOF
  exit 1
}

err() {
  echo "ERROR: $*" >&2
}

warn() {
  echo "WARNING: $*" >&2
}

# ---------------------------------------------------------------------------
# Argument parsing
# ---------------------------------------------------------------------------
while [[ $# -gt 0 ]]; do
  case $1 in
  # SSH-specific
  -H | --host)
    SSH_HOST="$2"
    shift 2
    ;;
  -l | --ssh-user)
    SSH_USER="$2"
    shift 2
    ;;
  -P | --ssh-port)
    SSH_PORT="$2"
    shift 2
    ;;
  -k | --ssh-key)
    SSH_KEY="$2"
    shift 2
    ;;
  --ssh-password)
    SSH_PASSWORD="$2"
    shift 2
    ;;
  --sudo-password)
    SUDO_PASSWORD="$2"
    shift 2
    ;;
  --no-strict-host-key-checking)
    NO_STRICT_HOST_KEY_CHECKING="true"
    shift
    ;;
  # Forwarded flags (no value)
  -i | --install | -u | --uninstall | -U | --update | --no-service | --std-user | --pre-release)
    FORWARD_ARGS+=("$1")
    shift
    ;;
  # Forwarded flags (with value)
  -p | --path | -v | --version | --modules | --api-port | -r | --root-password)
    FORWARD_ARGS+=("$1" "$2")
    shift 2
    ;;
  -h | --help)
    usage
    ;;
  *)
    err "Unknown option: $1"
    usage
    ;;
  esac
done

# ---------------------------------------------------------------------------
# Validation
# ---------------------------------------------------------------------------
if [[ -z "$SSH_HOST" ]]; then
  err "--host is required"
  usage
fi

if [[ -n "$SSH_PASSWORD" ]]; then
  warn "SSH password auth is less secure than key-based auth. Prefer --ssh-key."
  if ! command -v sshpass &>/dev/null; then
    err "sshpass is required for --ssh-password but was not found."
    err "Install it with:"
    err "  macOS:  brew install hudochenkov/sshpass/sshpass"
    err "  Ubuntu: sudo apt-get install sshpass"
    err "  RHEL:   sudo yum install sshpass"
    exit 1
  fi
fi

if [[ -n "$SUDO_PASSWORD" ]]; then
  warn "Passing --sudo-password embeds the password in the remote command string."
  warn "Passwordless sudo on the remote host is strongly preferred."
fi

# ---------------------------------------------------------------------------
# Download install.sh
# ---------------------------------------------------------------------------
echo "Downloading install.sh from GitHub..."
HTTP_STATUS=$(curl -sSL -w "%{http_code}" "$INSTALL_SH_URL" -o "$TEMP_SCRIPT")

if [[ "$HTTP_STATUS" != "200" ]]; then
  err "Failed to download install.sh (HTTP $HTTP_STATUS)"
  exit 1
fi

if [[ ! -s "$TEMP_SCRIPT" ]]; then
  err "Downloaded install.sh is empty"
  exit 1
fi

# Basic sanity check — file should start with a bash shebang
FIRST_LINE=$(head -n 1 "$TEMP_SCRIPT")
if [[ "$FIRST_LINE" != "#!/bin/bash" ]]; then
  err "Downloaded file does not look like a bash script (first line: $FIRST_LINE)"
  exit 1
fi

echo "install.sh downloaded and verified."

# ---------------------------------------------------------------------------
# Build SSH options
# ---------------------------------------------------------------------------
SSH_OPTS=(-p "$SSH_PORT")

if [[ -n "$SSH_KEY" ]]; then
  SSH_OPTS+=(-i "$SSH_KEY")
fi

if [[ "$NO_STRICT_HOST_KEY_CHECKING" = "true" ]]; then
  SSH_OPTS+=(-o StrictHostKeyChecking=no -o UserKnownHostsFile=/dev/null)
fi

if [[ -z "$SSH_PASSWORD" ]]; then
  # Non-interactive: fail immediately if no key/agent auth available
  SSH_OPTS+=(-o BatchMode=yes)
fi

# ---------------------------------------------------------------------------
# Build the remote command
# ---------------------------------------------------------------------------
# Shell-quote each forwarded argument so values with spaces survive the
# double trip through the shell (local → SSH → remote bash -s).
REMOTE_ARGS=""
for arg in "${FORWARD_ARGS[@]}"; do
  REMOTE_ARGS="${REMOTE_ARGS} $(printf '%q' "$arg")"
done

if [[ -n "$SUDO_PASSWORD" ]]; then
  # Pre-populate the sudo credential cache so install.sh's sudo calls work.
  # The echo pipes only to 'sudo -S true'; bash -s reads from SSH stdin (the
  # redirected script file), not from this pipeline.
  REMOTE_CMD="echo $(printf '%q' "$SUDO_PASSWORD") | sudo -S true 2>/dev/null && bash -s --${REMOTE_ARGS}"
else
  REMOTE_CMD="bash -s --${REMOTE_ARGS}"
fi

# ---------------------------------------------------------------------------
# Execute via SSH
# ---------------------------------------------------------------------------
echo "Connecting to ${SSH_USER}@${SSH_HOST}:${SSH_PORT}..."

if [[ -n "$SSH_PASSWORD" ]]; then
  sshpass -p "$SSH_PASSWORD" \
    ssh "${SSH_OPTS[@]}" "${SSH_USER}@${SSH_HOST}" "$REMOTE_CMD" \
    < "$TEMP_SCRIPT"
else
  ssh "${SSH_OPTS[@]}" "${SSH_USER}@${SSH_HOST}" "$REMOTE_CMD" \
    < "$TEMP_SCRIPT"
fi

EXIT_CODE=$?

if [[ $EXIT_CODE -ne 0 ]]; then
  err "Remote install.sh exited with code $EXIT_CODE"
fi

exit $EXIT_CODE
