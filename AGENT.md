# AGENT.md

## Scope
- This repository builds a Prometheus exporter in Go `1.26+`.
- Target platforms are Linux kernel `2.6+` on `amd64` and `arm64`.
- Primary domain focus is PCIe devices, especially whether negotiated link performance is correct.

## Build and CI
- Use GitHub Actions for CI and automation.
- Keep CI simple and reproducible: lint, build, and test on supported Linux architectures where practical.

## Dependency Policy
- Prefer the Go standard library.
- Avoid external dependencies unless they have a small, shallow dependency graph and clear maintenance history.
- Ask for approval before adding any new dependency.

## Exporter Behavior
- Read PCIe data from sysfs.
- Always support overriding the sysfs root path (for containerized environments and testing), with a sensible default of `/sys`.
- Treat negotiated PCIe performance checks as a first-class metric/output concern.

## Testing Policy
- Use `hammy` assertions from `https://github.com/gogunit/gunit/tree/main/hammy`.
- Avoid mocks in tests.
- If system-level inputs are needed, ask for stub data captured from a live system and use that as test fixtures.

## Engineering Guidelines
- Keep implementations explicit and straightforward over clever abstractions.
- Favor deterministic tests and stable text output for metrics.
- Document assumptions about kernel/sysfs behavior near the relevant code paths.
