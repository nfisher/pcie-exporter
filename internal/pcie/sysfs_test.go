package pcie

import (
	"math"
	"path/filepath"
	"testing"

	"github.com/gogunit/gunit/hammy"
)

func TestReadDevicesFromFixture(t *testing.T) {
	h := hammy.New(t)

	root := filepath.Join("testdata", "sysfs")
	devices, err := ReadDevices(root)
	h.Is(hammy.NilError(err))
	h.Is(hammy.Number(len(devices)).EqualTo(2))

	h.Is(hammy.String(devices[0].Address).EqualTo("0000:01:00.0"))
	h.Is(hammy.True(devices[0].NegotiatedOK))
	h.Is(hammy.Number(devices[0].SpeedRatio).Within(1.0, 0.00001))
	h.Is(hammy.Number(devices[0].WidthRatio).Within(1.0, 0.00001))

	h.Is(hammy.String(devices[1].Address).EqualTo("0000:02:00.0"))
	h.Is(hammy.False(devices[1].NegotiatedOK))
	h.Is(hammy.Number(devices[1].SpeedRatio).Within(0.5, 0.00001))
	h.Is(hammy.Number(devices[1].WidthRatio).Within(0.5, 0.00001))
}

func TestCompareFallbackBehavior(t *testing.T) {
	h := hammy.New(t)

	ratio, ok := compareSpeed("Unknown", "Unknown")
	h.Is(hammy.True(ok))
	h.Is(hammy.Number(ratio).Within(1.0, 0.00001))

	ratio, ok = compareWidth("x8", "x16")
	h.Is(hammy.False(ok))
	h.Is(hammy.Number(ratio).Within(0.5, 0.00001))

	ratio, ok = compareSpeed("n/a", "16 GT/s")
	h.Is(hammy.False(ok))
	h.Is(hammy.True(math.IsNaN(ratio)))
}
