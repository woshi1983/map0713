#!/bin/bash
set -euo pipefail

entrypoint=${1:?entrypoint path is required}
expected_sha=c47d1ff11d61b92b1647971bba42df8bd270676b1141784d8e5fc51bc2ee7c5b
actual_sha=$(sha256sum "$entrypoint" | cut -d ' ' -f 1)
if [[ "$actual_sha" != "$expected_sha" ]]; then
  echo "Unexpected cluster-initiator entrypoint; refusing to patch it." >&2
  exit 1
fi

patched=$(mktemp)
awk '
  /^VELERO_PROJECT_NAME=/ || /^VELERO_NAME=/ || /^velero_flag_name=/ { next }
  /^function install_velero\(\)/ { install_found++; skipping=1; next }
  skipping && /^function check_svc_up\(\)/ { skipping=0 }
  /^function check_velero_up\(\)/ { check_found++; skipping=1; next }
  skipping && /^# install metrics server$/ { skipping=0 }
  /^# install velero$/ {
    tail_found++
    print "# MAP does not install Velero during cluster initialization."
    print "echo \"[INFO] Cluster bereit!\""
    print "exit 0"
    exit
  }
  !skipping { print }
  END {
    if (install_found != 1 || check_found != 1 || tail_found != 1) exit 1
  }
' "$entrypoint" > "$patched" || { rm -f "$patched"; exit 1; }

bash -n "$patched"
chmod 755 "$patched"
mv "$patched" "$entrypoint"
