# pcie-exporter

Prometheus exporter for PCIe link negotiation visibility on Linux.

This exporter focuses on whether each PCIe device is running at expected negotiated performance (speed and width) versus its maximum supported link.

## Scope

- Go `1.26+`
- Linux kernel `2.6+`
- Architectures: `amd64`, `arm64`
- Runtime dependencies: Go standard library only

## Why

PCIe mis-negotiation (lower-than-expected link speed or width) can silently degrade host and workload performance. This exporter surfaces those conditions as scrapeable metrics.

## Build

```bash
go build ./cmd/pcie-exporter
```

## Run

```bash
./pcie-exporter \
  -listen-address=:9808 \
  -sysfs-root=/host/sysfs
```

Configuration precedence for sysfs root:

1. `-sysfs-root` flag
2. `PCIE_EXPORTER_SYSFS` environment variable
3. default `/sys`

This allows running in containers where sysfs is mounted at a non-default path.

Container example:

```bash
docker run --rm -p 9808:9808 \
  -v /sys:/host/sysfs:ro \
  ghcr.io/nfisher/pcie-exporter:latest \
  -sysfs-root=/host/sysfs
```

## HTTP Endpoints

- `/metrics`: Prometheus text exposition
- `/pcie-tree`: PCIe topology tree in JSON with `bus_id`, `name`, `link_capacity`, and `link_status`
- `/healthz`: basic health probe (`200 ok`)

Example:

```bash
curl -s http://127.0.0.1:9808/pcie-tree
```

## Exported Metrics

- `pcie_devices_total` gauge: devices with complete PCIe link files
- `pcie_link_negotiated_ok` gauge: `1` if negotiated speed and width match max supported values, else `0`
- `pcie_link_speed_ratio` gauge: negotiated speed / max speed
- `pcie_link_width_ratio` gauge: negotiated width / max width
- `pcie_exporter_scrapes_total` counter
- `pcie_exporter_scrape_errors_total` counter
- `pcie_exporter_last_scrape_duration_seconds` gauge
- `pcie_exporter_last_scrape_success` gauge

## PCIe Throughput Map

The repository includes a version/lane throughput map at `internal/pcie/bandwidth_map.go`.

These are theoretical single-direction values and account for line encoding only.

| PCIe Version | Transfer Rate / lane (GT/s) | x1 (GB/s) | x2 (GB/s) | x4 (GB/s) | x8 (GB/s) | x16 (GB/s) |
| --- | --- | --- | --- | --- | --- | --- |
| 1.0 | 2.5 | 0.250 | 0.500 | 1.000 | 2.000 | 4.000 |
| 2.0 | 5.0 | 0.500 | 1.000 | 2.000 | 4.000 | 8.000 |
| 3.0 | 8.0 | 0.985 | 1.969 | 3.938 | 7.877 | 15.754 |
| 4.0 | 16.0 | 1.969 | 3.938 | 7.877 | 15.754 | 31.508 |
| 5.0 | 32.0 | 3.938 | 7.877 | 15.754 | 31.508 | 63.015 |

## Testing

Tests are fixture-driven and do not use mocks.

- Assertions use `hammy`: <https://github.com/gogunit/gunit/tree/main/hammy>
- Fixture data lives under `internal/pcie/testdata/sysfs`

Run tests:

```bash
go test ./...
```

For best fidelity, replace/add fixtures with data captured from a live target system (host classes you care about most).

Current priority target systems for fixture captures:

- NVIDIA GB200
- NVIDIA GB300
- NVIDIA B300
- NVIDIA H100
- NVIDIA H200

## CI

GitHub Actions workflow is at `.github/workflows/ci.yml` and currently runs:

- `go test ./...` on Linux
- cross-arch build (`linux/amd64`, `linux/arm64`) for `cmd/pcie-exporter`

## Dependency Policy

- Prefer standard library.
- External dependencies should remain small and shallow.
- Ask before adding dependencies.
