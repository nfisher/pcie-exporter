package pcie

import (
	"errors"
	"fmt"
	"io/fs"
	"math"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
)

// Device contains the minimum PCIe link information needed for exported metrics.
type Device struct {
	Address          string
	VendorID         string
	DeviceID         string
	Class            string
	CurrentLinkSpeed string
	MaxLinkSpeed     string
	CurrentLinkWidth string
	MaxLinkWidth     string
	NegotiatedOK     bool
	SpeedRatio       float64
	WidthRatio       float64
}

// ReadDevices enumerates PCIe devices from sysfsRoot/bus/pci/devices.
func ReadDevices(sysfsRoot string) ([]Device, error) {
	devicesPath := filepath.Join(sysfsRoot, "bus", "pci", "devices")
	entries, err := os.ReadDir(devicesPath)
	if err != nil {
		return nil, fmt.Errorf("read pci devices from %s: %w", devicesPath, err)
	}

	devices := make([]Device, 0, len(entries))
	for _, entry := range entries {
		address := entry.Name()
		devicePath := filepath.Join(devicesPath, address)
		device, ok, err := readDevice(devicePath, address)
		if err != nil {
			return nil, err
		}
		if ok {
			devices = append(devices, device)
		}
	}

	sort.Slice(devices, func(i, j int) bool {
		return devices[i].Address < devices[j].Address
	})

	return devices, nil
}

func readDevice(devicePath, address string) (Device, bool, error) {
	currentSpeed, hasCurrentSpeed, err := readOptionalTrim(filepath.Join(devicePath, "current_link_speed"))
	if err != nil {
		return Device{}, false, fmt.Errorf("read current_link_speed for %s: %w", address, err)
	}
	maxSpeed, hasMaxSpeed, err := readOptionalTrim(filepath.Join(devicePath, "max_link_speed"))
	if err != nil {
		return Device{}, false, fmt.Errorf("read max_link_speed for %s: %w", address, err)
	}
	currentWidth, hasCurrentWidth, err := readOptionalTrim(filepath.Join(devicePath, "current_link_width"))
	if err != nil {
		return Device{}, false, fmt.Errorf("read current_link_width for %s: %w", address, err)
	}
	maxWidth, hasMaxWidth, err := readOptionalTrim(filepath.Join(devicePath, "max_link_width"))
	if err != nil {
		return Device{}, false, fmt.Errorf("read max_link_width for %s: %w", address, err)
	}

	// Skip entries that do not provide link negotiation info.
	if !(hasCurrentSpeed && hasMaxSpeed && hasCurrentWidth && hasMaxWidth) {
		return Device{}, false, nil
	}

	vendorID, _, err := readOptionalTrim(filepath.Join(devicePath, "vendor"))
	if err != nil {
		return Device{}, false, fmt.Errorf("read vendor for %s: %w", address, err)
	}
	deviceID, _, err := readOptionalTrim(filepath.Join(devicePath, "device"))
	if err != nil {
		return Device{}, false, fmt.Errorf("read device for %s: %w", address, err)
	}
	class, _, err := readOptionalTrim(filepath.Join(devicePath, "class"))
	if err != nil {
		return Device{}, false, fmt.Errorf("read class for %s: %w", address, err)
	}

	speedRatio, speedOK := compareSpeed(currentSpeed, maxSpeed)
	widthRatio, widthOK := compareWidth(currentWidth, maxWidth)

	return Device{
		Address:          address,
		VendorID:         vendorID,
		DeviceID:         deviceID,
		Class:            class,
		CurrentLinkSpeed: currentSpeed,
		MaxLinkSpeed:     maxSpeed,
		CurrentLinkWidth: currentWidth,
		MaxLinkWidth:     maxWidth,
		NegotiatedOK:     speedOK && widthOK,
		SpeedRatio:       speedRatio,
		WidthRatio:       widthRatio,
	}, true, nil
}

func readOptionalTrim(path string) (value string, ok bool, err error) {
	buf, err := os.ReadFile(path)
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return "", false, nil
		}
		return "", false, err
	}
	return strings.TrimSpace(string(buf)), true, nil
}

func compareSpeed(current, max string) (ratio float64, ok bool) {
	currentValue, currentParsed := parseLeadingFloat(current)
	maxValue, maxParsed := parseLeadingFloat(max)
	if currentParsed && maxParsed && maxValue > 0 {
		return currentValue / maxValue, currentValue+1e-9 >= maxValue
	}

	if current != "" && current == max {
		return 1.0, true
	}

	return math.NaN(), false
}

func compareWidth(current, max string) (ratio float64, ok bool) {
	currentValue, currentParsed := parseFirstInt(current)
	maxValue, maxParsed := parseFirstInt(max)
	if currentParsed && maxParsed && maxValue > 0 {
		return float64(currentValue) / float64(maxValue), currentValue >= maxValue
	}

	if current != "" && current == max {
		return 1.0, true
	}

	return math.NaN(), false
}

func parseLeadingFloat(s string) (float64, bool) {
	fields := strings.Fields(strings.TrimSpace(s))
	if len(fields) == 0 {
		return 0, false
	}
	v, err := strconv.ParseFloat(fields[0], 64)
	if err != nil {
		return 0, false
	}
	return v, true
}

func parseFirstInt(s string) (int, bool) {
	s = strings.TrimSpace(s)
	if s == "" {
		return 0, false
	}

	start := -1
	for i := 0; i < len(s); i++ {
		if s[i] >= '0' && s[i] <= '9' {
			start = i
			break
		}
	}
	if start < 0 {
		return 0, false
	}

	end := start
	for end < len(s) && s[end] >= '0' && s[end] <= '9' {
		end++
	}

	v, err := strconv.Atoi(s[start:end])
	if err != nil {
		return 0, false
	}
	return v, true
}
