#!/usr/bin/env bash
set -euo pipefail

usage() {
  cat <<'USAGE'
Usage: capture-pci-sysfs.sh [-s SYSFS_ROOT] [-o OUTPUT_TAR_GZ]

Capture PCI device details from sysfs into a gzipped tar archive.

Options:
  -s SYSFS_ROOT    Sysfs root path (default: /sys)
  -o OUTPUT        Output archive path (default: ./pcie-sysfs-<timestamp>.tar.gz)
  -h               Show help
USAGE
}

sysfs_root="/sys"
output=""

while getopts ":s:o:h" opt; do
  case "$opt" in
    s) sysfs_root="$OPTARG" ;;
    o) output="$OPTARG" ;;
    h)
      usage
      exit 0
      ;;
    :)
      echo "error: option -$OPTARG requires an argument" >&2
      usage >&2
      exit 2
      ;;
    \?)
      echo "error: invalid option -$OPTARG" >&2
      usage >&2
      exit 2
      ;;
  esac
done

if [[ -z "$output" ]]; then
  timestamp="$(date +%Y%m%d-%H%M%S)"
  output="pcie-sysfs-${timestamp}.tar.gz"
fi

devices_dir="${sysfs_root%/}/bus/pci/devices"
if [[ ! -d "$devices_dir" ]]; then
  echo "error: PCI devices directory not found: $devices_dir" >&2
  exit 1
fi

tmpdir="$(mktemp -d)"
snapshot_root="$tmpdir/sysfs_snapshot"
cleanup() {
  rm -rf "$tmpdir"
}
trap cleanup EXIT

mkdir -p "$snapshot_root/bus/pci/devices"

copy_file_if_exists() {
  local src="$1"
  local dst="$2"
  if [[ -r "$src" && -f "$src" ]]; then
    cp "$src" "$dst"
  fi
}

write_link_target() {
  local src="$1"
  local dst="$2"
  local name="$3"

  if [[ -L "$src" ]]; then
    local target
    target="$(readlink "$src")"
    printf '%s=%s\n' "$name" "$target" >> "$dst"
  fi
}

selected_files=(
  class
  vendor
  device
  subsystem_vendor
  subsystem_device
  revision
  modalias
  numa_node
  local_cpulist
  irq
  enable
  current_link_speed
  current_link_width
  max_link_speed
  max_link_width
  current_link_speed_raw
  max_link_speed_raw
  broken_parity_status
  ari_enabled
  d3cold_allowed
  uevent
  resource
)

snapshot_info="$snapshot_root/snapshot-info.txt"
{
  echo "captured_at_utc=$(date -u +%Y-%m-%dT%H:%M:%SZ)"
  echo "hostname=$(hostname 2>/dev/null || true)"
  echo "kernel=$(uname -srmo 2>/dev/null || true)"
  echo "sysfs_root=$sysfs_root"
} > "$snapshot_info"

device_list="$snapshot_root/devices.txt"
: > "$device_list"

shopt -s nullglob
for device_path in "$devices_dir"/*; do
  device_id="$(basename "$device_path")"
  if [[ ! "$device_id" =~ ^[0-9a-fA-F]{4}:[0-9a-fA-F]{2}:[0-9a-fA-F]{2}\.[0-7]$ ]]; then
    continue
  fi

  echo "$device_id" >> "$device_list"

  out_dir="$snapshot_root/bus/pci/devices/$device_id"
  mkdir -p "$out_dir"

  for f in "${selected_files[@]}"; do
    copy_file_if_exists "$device_path/$f" "$out_dir/$f"
  done

  links_file="$out_dir/links.txt"
  : > "$links_file"
  write_link_target "$device_path/driver" "$links_file" "driver"
  write_link_target "$device_path/iommu_group" "$links_file" "iommu_group"
  write_link_target "$device_path/physfn" "$links_file" "physfn"
  write_link_target "$device_path/subsystem" "$links_file" "subsystem"

  if [[ ! -s "$links_file" ]]; then
    rm -f "$links_file"
  fi

  resolved_file="$out_dir/resolved-path.txt"
  if resolved="$(readlink -f "$device_path" 2>/dev/null)"; then
    printf '%s\n' "$resolved" > "$resolved_file"
  elif resolved="$(realpath "$device_path" 2>/dev/null)"; then
    printf '%s\n' "$resolved" > "$resolved_file"
  fi

  power_dir="$device_path/power"
  if [[ -d "$power_dir" ]]; then
    mkdir -p "$out_dir/power"
    for p in "$power_dir"/*; do
      if [[ -r "$p" && -f "$p" ]]; then
        cp "$p" "$out_dir/power/$(basename "$p")"
      fi
    done
  fi

done
shopt -u nullglob

sort -u "$device_list" -o "$device_list"

tar -czf "$output" -C "$tmpdir" "sysfs_snapshot"

echo "wrote $output"
