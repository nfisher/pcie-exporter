#!/usr/bin/env bash
set -euo pipefail

usage() {
  cat <<'USAGE'
Usage: capture-pci-sysfs.sh [-s SYSFS_ROOT] [-o OUTPUT_TAR_GZ]

Create a small snapshot of PCI device details from sysfs.

Options:
  -s SYSFS_ROOT    Sysfs root path (default: /sys)
  -o OUTPUT        Output archive path (default: pcie-sysfs-<timestamp>.tar.gz)
  -h               Show help
USAGE
}

sysfs_root="/sys"
output=""

while getopts ":s:o:h" opt; do
  case "$opt" in
    s) sysfs_root="$OPTARG" ;;
    o) output="$OPTARG" ;;
    h) usage; exit 0 ;;
    :) echo "error: -$OPTARG needs a value" >&2; exit 2 ;;
    \?) echo "error: invalid option -$OPTARG" >&2; exit 2 ;;
  esac
done

if [[ -z "$output" ]]; then
  output="pcie-sysfs-$(date +%Y%m%d-%H%M%S).tar.gz"
fi

devices_dir="${sysfs_root%/}/bus/pci/devices"
if [[ ! -d "$devices_dir" ]]; then
  echo "error: not found: $devices_dir" >&2
  exit 1
fi

# Keep this list short and human-reviewable.
files=(
  vendor
  device
  class
  subsystem_vendor
  subsystem_device
  current_link_speed
  current_link_width
  max_link_speed
  max_link_width
  modalias
  uevent
)

tmpdir="$(mktemp -d)"
trap 'rm -rf "$tmpdir"' EXIT
snapshot="$tmpdir/sysfs_snapshot"
mkdir -p "$snapshot/bus/pci/devices"

{
  echo "captured_at_utc=$(date -u +%Y-%m-%dT%H:%M:%SZ)"
  echo "sysfs_root=$sysfs_root"
} > "$snapshot/snapshot-info.txt"

: > "$snapshot/devices.txt"

for device_path in "$devices_dir"/*; do
  device_id="$(basename "$device_path")"
  [[ "$device_id" =~ ^[0-9a-fA-F]{4}:[0-9a-fA-F]{2}:[0-9a-fA-F]{2}\.[0-7]$ ]] || continue

  echo "$device_id" >> "$snapshot/devices.txt"

  out_dir="$snapshot/bus/pci/devices/$device_id"
  mkdir -p "$out_dir"

  for f in "${files[@]}"; do
    if [[ -f "$device_path/$f" && -r "$device_path/$f" ]]; then
      cp "$device_path/$f" "$out_dir/$f"
    fi
  done

  if [[ -L "$device_path/driver" ]]; then
    readlink "$device_path/driver" > "$out_dir/driver-link.txt"
  fi
done

sort -u "$snapshot/devices.txt" -o "$snapshot/devices.txt"
tar -czf "$output" -C "$tmpdir" sysfs_snapshot

echo "wrote $output"
