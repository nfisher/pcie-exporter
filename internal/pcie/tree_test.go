package pcie

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/gogunit/gunit/hammy"
)

func TestReadTreeBuildsHierarchyWithLinkData(t *testing.T) {
	h := hammy.New(t)

	sysfsRoot := t.TempDir()
	busDevices := filepath.Join(sysfsRoot, "bus", "pci", "devices")
	h.Is(hammy.NilError(os.MkdirAll(busDevices, 0o755)))

	bridgePath := filepath.Join(sysfsRoot, "devices", "pci0000:00", "0000:00:01.0")
	gpuPath := filepath.Join(bridgePath, "0000:01:00.0")
	nicPath := filepath.Join(sysfsRoot, "devices", "pci0000:00", "0000:02:00.0")

	mustMkdirAll(t, bridgePath)
	mustMkdirAll(t, gpuPath)
	mustMkdirAll(t, nicPath)

	mustWriteFile(t, filepath.Join(bridgePath, "vendor"), "0x8086\n")
	mustWriteFile(t, filepath.Join(bridgePath, "device"), "0x1234\n")
	mustWriteFile(t, filepath.Join(bridgePath, "max_link_speed"), "16 GT/s PCIe\n")
	mustWriteFile(t, filepath.Join(bridgePath, "max_link_width"), "16\n")
	mustWriteFile(t, filepath.Join(bridgePath, "current_link_speed"), "16 GT/s PCIe\n")
	mustWriteFile(t, filepath.Join(bridgePath, "current_link_width"), "16\n")

	mustWriteFile(t, filepath.Join(gpuPath, "label"), "NVIDIA H100\n")
	mustWriteFile(t, filepath.Join(gpuPath, "vendor"), "0x10de\n")
	mustWriteFile(t, filepath.Join(gpuPath, "device"), "0x2331\n")
	mustWriteFile(t, filepath.Join(gpuPath, "max_link_speed"), "32 GT/s PCIe\n")
	mustWriteFile(t, filepath.Join(gpuPath, "max_link_width"), "16\n")
	mustWriteFile(t, filepath.Join(gpuPath, "current_link_speed"), "16 GT/s PCIe\n")
	mustWriteFile(t, filepath.Join(gpuPath, "current_link_width"), "8\n")

	mustWriteFile(t, filepath.Join(nicPath, "vendor"), "0x15b3\n")
	mustWriteFile(t, filepath.Join(nicPath, "device"), "0x1017\n")

	mustSymlink(t, bridgePath, filepath.Join(busDevices, "0000:00:01.0"))
	mustSymlink(t, gpuPath, filepath.Join(busDevices, "0000:01:00.0"))
	mustSymlink(t, nicPath, filepath.Join(busDevices, "0000:02:00.0"))

	tree, err := ReadTree(sysfsRoot)
	h.Is(hammy.NilError(err))
	h.Is(hammy.Number(len(tree)).EqualTo(2))

	rootBridge := tree[0]
	h.Is(hammy.String(rootBridge.BusID).EqualTo("0000:00:01.0"))
	h.Is(hammy.String(rootBridge.Name).EqualTo("8086:1234"))
	h.Is(hammy.String(rootBridge.LinkCapacity).EqualTo("16 GT/s PCIe x16"))
	h.Is(hammy.String(rootBridge.LinkStatus).EqualTo("16 GT/s PCIe x16"))
	h.Is(hammy.Number(len(rootBridge.Children)).EqualTo(1))

	gpu := rootBridge.Children[0]
	h.Is(hammy.String(gpu.BusID).EqualTo("0000:01:00.0"))
	h.Is(hammy.String(gpu.Name).EqualTo("NVIDIA H100"))
	h.Is(hammy.String(gpu.LinkCapacity).EqualTo("32 GT/s PCIe x16"))
	h.Is(hammy.String(gpu.LinkStatus).EqualTo("16 GT/s PCIe x8"))

	nic := tree[1]
	h.Is(hammy.String(nic.BusID).EqualTo("0000:02:00.0"))
	h.Is(hammy.String(nic.Name).EqualTo("15b3:1017"))
	h.Is(hammy.String(nic.LinkCapacity).EqualTo("unknown"))
	h.Is(hammy.String(nic.LinkStatus).EqualTo("unknown"))
}

func mustMkdirAll(t *testing.T, path string) {
	t.Helper()
	if err := os.MkdirAll(path, 0o755); err != nil {
		t.Fatalf("mkdir %s: %v", path, err)
	}
}

func mustWriteFile(t *testing.T, path, content string) {
	t.Helper()
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("write %s: %v", path, err)
	}
}

func mustSymlink(t *testing.T, target, link string) {
	t.Helper()
	if err := os.Symlink(target, link); err != nil {
		t.Fatalf("symlink %s -> %s: %v", link, target, err)
	}
}
