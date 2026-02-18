package pcie

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
)

var pciAddressPattern = regexp.MustCompile(`^[0-9a-fA-F]{4}:[0-9a-fA-F]{2}:[0-9a-fA-F]{2}\.[0-7]$`)

// TreeNode represents one device in the PCIe topology tree.
type TreeNode struct {
	BusID        string      `json:"bus_id"`
	Name         string      `json:"name"`
	LinkCapacity string      `json:"link_capacity"`
	LinkStatus   string      `json:"link_status"`
	Children     []*TreeNode `json:"children,omitempty"`
}

// ReadTree builds a PCIe topology tree from sysfsRoot/bus/pci/devices.
func ReadTree(sysfsRoot string) ([]*TreeNode, error) {
	devicesPath := filepath.Join(sysfsRoot, "bus", "pci", "devices")
	entries, err := os.ReadDir(devicesPath)
	if err != nil {
		return nil, fmt.Errorf("read pci devices from %s: %w", devicesPath, err)
	}

	nodes := make(map[string]*TreeNode, len(entries))
	parents := make(map[string]string, len(entries))

	for _, entry := range entries {
		address := strings.ToLower(entry.Name())
		devicePath := filepath.Join(devicesPath, entry.Name())

		node, err := readTreeNode(devicePath, address)
		if err != nil {
			return nil, err
		}
		nodes[address] = node

		parentAddress, err := resolveParentAddress(devicePath, address)
		if err != nil {
			return nil, err
		}
		if parentAddress != "" {
			parents[address] = parentAddress
		}
	}

	roots := make([]*TreeNode, 0, len(nodes))
	for address, node := range nodes {
		parentAddress, hasParent := parents[address]
		if hasParent {
			if parentNode, ok := nodes[parentAddress]; ok {
				parentNode.Children = append(parentNode.Children, node)
				continue
			}
		}
		roots = append(roots, node)
	}

	sortTree(roots)
	return roots, nil
}

func readTreeNode(devicePath, address string) (*TreeNode, error) {
	name, err := readDeviceName(devicePath, address)
	if err != nil {
		return nil, err
	}

	currentSpeed, _, err := readOptionalTrim(filepath.Join(devicePath, "current_link_speed"))
	if err != nil {
		return nil, fmt.Errorf("read current_link_speed for %s: %w", address, err)
	}
	maxSpeed, _, err := readOptionalTrim(filepath.Join(devicePath, "max_link_speed"))
	if err != nil {
		return nil, fmt.Errorf("read max_link_speed for %s: %w", address, err)
	}
	currentWidth, _, err := readOptionalTrim(filepath.Join(devicePath, "current_link_width"))
	if err != nil {
		return nil, fmt.Errorf("read current_link_width for %s: %w", address, err)
	}
	maxWidth, _, err := readOptionalTrim(filepath.Join(devicePath, "max_link_width"))
	if err != nil {
		return nil, fmt.Errorf("read max_link_width for %s: %w", address, err)
	}

	return &TreeNode{
		BusID:        address,
		Name:         name,
		LinkCapacity: formatLinkSummary(maxSpeed, maxWidth),
		LinkStatus:   formatLinkSummary(currentSpeed, currentWidth),
	}, nil
}

func readDeviceName(devicePath, address string) (string, error) {
	label, hasLabel, err := readOptionalTrim(filepath.Join(devicePath, "label"))
	if err != nil {
		return "", fmt.Errorf("read label for %s: %w", address, err)
	}
	if hasLabel && label != "" {
		return label, nil
	}

	driverName, hasDriver, err := readDriverName(devicePath)
	if err != nil {
		return "", fmt.Errorf("read driver for %s: %w", address, err)
	}
	if hasDriver {
		return driverName, nil
	}

	vendorID, _, err := readOptionalTrim(filepath.Join(devicePath, "vendor"))
	if err != nil {
		return "", fmt.Errorf("read vendor for %s: %w", address, err)
	}
	deviceID, _, err := readOptionalTrim(filepath.Join(devicePath, "device"))
	if err != nil {
		return "", fmt.Errorf("read device for %s: %w", address, err)
	}

	vendorID = trimHexPrefix(vendorID)
	deviceID = trimHexPrefix(deviceID)
	if vendorID != "" || deviceID != "" {
		return vendorID + ":" + deviceID, nil
	}

	return address, nil
}

func readDriverName(devicePath string) (value string, ok bool, err error) {
	driverPath := filepath.Join(devicePath, "driver")
	resolvedPath, err := filepath.EvalSymlinks(driverPath)
	if err != nil {
		if os.IsNotExist(err) {
			return "", false, nil
		}
		return "", false, err
	}
	name := strings.TrimSpace(filepath.Base(resolvedPath))
	if name == "" || name == "." || name == string(filepath.Separator) {
		return "", false, nil
	}
	return name, true, nil
}

func resolveParentAddress(devicePath, address string) (string, error) {
	resolvedPath, err := filepath.EvalSymlinks(devicePath)
	if err != nil {
		return "", fmt.Errorf("resolve symlink for %s: %w", address, err)
	}

	pathParts := strings.Split(filepath.ToSlash(resolvedPath), "/")
	addresses := make([]string, 0, 4)
	for _, part := range pathParts {
		if pciAddressPattern.MatchString(part) {
			addresses = append(addresses, strings.ToLower(part))
		}
	}

	if len(addresses) < 2 {
		return "", nil
	}

	for i := len(addresses) - 1; i >= 0; i-- {
		if addresses[i] != address {
			continue
		}
		if i == 0 {
			return "", nil
		}
		return addresses[i-1], nil
	}

	return "", nil
}

func formatLinkSummary(speed, width string) string {
	speed = strings.TrimSpace(speed)
	width = formatWidth(width)

	if speed == "" && width == "" {
		return "unknown"
	}
	if speed == "" {
		return width
	}
	if width == "" {
		return speed
	}
	return speed + " " + width
}

func formatWidth(width string) string {
	width = strings.TrimSpace(width)
	if width == "" {
		return ""
	}

	if strings.HasPrefix(strings.ToLower(width), "x") {
		return "x" + strings.TrimPrefix(strings.TrimPrefix(width, "x"), "X")
	}

	value, ok := parseFirstInt(width)
	if !ok {
		return width
	}
	return "x" + strconv.Itoa(value)
}

func trimHexPrefix(value string) string {
	value = strings.TrimSpace(value)
	value = strings.TrimPrefix(value, "0x")
	value = strings.TrimPrefix(value, "0X")
	return value
}

func sortTree(nodes []*TreeNode) {
	sort.Slice(nodes, func(i, j int) bool {
		return nodes[i].BusID < nodes[j].BusID
	})
	for _, node := range nodes {
		sortTree(node.Children)
	}
}
