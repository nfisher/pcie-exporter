package pcie

import (
	"testing"

	"github.com/gogunit/gunit/hammy"
)

func TestThroughputMapValues(t *testing.T) {
	h := hammy.New(t)

	v5, ok := VersionBandwidthMap["5.0"]
	h.Is(hammy.True(ok))
	h.Is(hammy.Number(v5.TransferRateGTps).Within(32.0, 0.000001))
	h.Is(hammy.Number(v5.ThroughputGBps[1]).Within(3.938462, 0.000001))
	h.Is(hammy.Number(v5.ThroughputGBps[16]).Within(63.015392, 0.000001))

	v3, ok := VersionBandwidthMap["3.0"]
	h.Is(hammy.True(ok))
	h.Is(hammy.Number(v3.ThroughputGBps[8]).Within(7.876920, 0.000001))
}

func TestThroughputGBpsLookup(t *testing.T) {
	h := hammy.New(t)

	value, err := ThroughputGBps("4.0", 16)
	h.Is(hammy.NilError(err))
	h.Is(hammy.Number(value).Within(31.507696, 0.000001))

	_, err = ThroughputGBps("6.0", 16)
	h.Is(hammy.String(err.Error()).Contains("unsupported PCIe version"))

	_, err = ThroughputGBps("4.0", 12)
	h.Is(hammy.String(err.Error()).Contains("unsupported lane count"))
}
