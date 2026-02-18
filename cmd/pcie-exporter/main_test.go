package main

import (
	"testing"

	"github.com/gogunit/gunit/hammy"
)

func TestResolveSysfsRootFlagWins(t *testing.T) {
	h := hammy.New(t)

	t.Setenv("PCIE_EXPORTER_SYSFS", "/from/env")
	resolved := resolveSysfsRoot("/from/flag")
	h.Is(hammy.String(resolved).EqualTo("/from/flag"))
}

func TestResolveSysfsRootUsesEnv(t *testing.T) {
	h := hammy.New(t)

	t.Setenv("PCIE_EXPORTER_SYSFS", "/from/env")
	resolved := resolveSysfsRoot("")
	h.Is(hammy.String(resolved).EqualTo("/from/env"))
}

func TestResolveSysfsRootDefaultsToSys(t *testing.T) {
	h := hammy.New(t)

	t.Setenv("PCIE_EXPORTER_SYSFS", "")
	resolved := resolveSysfsRoot("")
	h.Is(hammy.String(resolved).EqualTo("/sys"))
}
