package pcie

import "fmt"

// LaneCounts tracks the standard lane widths we report.
var LaneCounts = []int{1, 2, 4, 8, 16}

// VersionBandwidth defines theoretical single-direction PCIe throughput.
// Values account for line encoding only and do not include higher-layer protocol overhead.
type VersionBandwidth struct {
	Version          string
	TransferRateGTps float64
	ThroughputGBps   map[int]float64
}

// VersionBandwidthMap contains PCIe generation capabilities used for expected throughput checks.
// The table is intentionally limited to Gen1-Gen5, which aligns with current target systems.
var VersionBandwidthMap = map[string]VersionBandwidth{
	"1.0": buildVersionBandwidth("1.0", 2.5, 0.250000),
	"2.0": buildVersionBandwidth("2.0", 5.0, 0.500000),
	"3.0": buildVersionBandwidth("3.0", 8.0, 0.984615),
	"4.0": buildVersionBandwidth("4.0", 16.0, 1.969231),
	"5.0": buildVersionBandwidth("5.0", 32.0, 3.938462),
}

func buildVersionBandwidth(version string, transferRateGTps, x1GBps float64) VersionBandwidth {
	return VersionBandwidth{
		Version:          version,
		TransferRateGTps: transferRateGTps,
		ThroughputGBps: map[int]float64{
			1:  x1GBps,
			2:  x1GBps * 2,
			4:  x1GBps * 4,
			8:  x1GBps * 8,
			16: x1GBps * 16,
		},
	}
}

// ThroughputGBps returns the theoretical single-direction throughput for a version/lane combination.
func ThroughputGBps(version string, lanes int) (float64, error) {
	entry, ok := VersionBandwidthMap[version]
	if !ok {
		return 0, fmt.Errorf("unsupported PCIe version %q", version)
	}

	throughput, ok := entry.ThroughputGBps[lanes]
	if !ok {
		return 0, fmt.Errorf("unsupported lane count %d", lanes)
	}

	return throughput, nil
}
