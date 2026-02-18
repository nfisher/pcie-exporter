package exporter

import (
	"fmt"
	"math"
	"net/http"
	"strconv"
	"strings"
	"sync/atomic"
	"time"

	"github.com/nfisher/pcie-exporter/internal/pcie"
)

const contentType = "text/plain; version=0.0.4; charset=utf-8"

// Handler serves Prometheus text exposition for PCIe link metrics.
type Handler struct {
	sysfsRoot  string
	scrapes    atomic.Uint64
	scrapeErrs atomic.Uint64
}

func NewHandler(sysfsRoot string) *Handler {
	return &Handler{sysfsRoot: sysfsRoot}
}

func (h *Handler) ServeHTTP(w http.ResponseWriter, _ *http.Request) {
	start := time.Now()
	devices, err := pcie.ReadDevices(h.sysfsRoot)
	scrapeDuration := time.Since(start).Seconds()

	h.scrapes.Add(1)
	scrapeSuccess := 1
	if err != nil {
		h.scrapeErrs.Add(1)
		scrapeSuccess = 0
	}

	w.Header().Set("Content-Type", contentType)
	w.WriteHeader(http.StatusOK)

	var b strings.Builder
	b.Grow(4096)

	b.WriteString("# HELP pcie_devices_total Number of PCIe devices with link data in sysfs.\n")
	b.WriteString("# TYPE pcie_devices_total gauge\n")
	b.WriteString("pcie_devices_total ")
	b.WriteString(strconv.Itoa(len(devices)))
	b.WriteString("\n")

	b.WriteString("# HELP pcie_link_negotiated_ok Whether negotiated PCIe link speed and width match maximum supported values.\n")
	b.WriteString("# TYPE pcie_link_negotiated_ok gauge\n")

	b.WriteString("# HELP pcie_link_speed_ratio Negotiated link speed divided by max supported link speed.\n")
	b.WriteString("# TYPE pcie_link_speed_ratio gauge\n")

	b.WriteString("# HELP pcie_link_width_ratio Negotiated link width divided by max supported link width.\n")
	b.WriteString("# TYPE pcie_link_width_ratio gauge\n")

	for _, device := range devices {
		labels := metricLabels(device)
		okValue := "0"
		if device.NegotiatedOK {
			okValue = "1"
		}

		b.WriteString("pcie_link_negotiated_ok")
		b.WriteString(labels)
		b.WriteString(" ")
		b.WriteString(okValue)
		b.WriteString("\n")

		if !math.IsNaN(device.SpeedRatio) {
			b.WriteString("pcie_link_speed_ratio")
			b.WriteString(labels)
			b.WriteString(" ")
			b.WriteString(fmt.Sprintf("%.6f", device.SpeedRatio))
			b.WriteString("\n")
		}

		if !math.IsNaN(device.WidthRatio) {
			b.WriteString("pcie_link_width_ratio")
			b.WriteString(labels)
			b.WriteString(" ")
			b.WriteString(fmt.Sprintf("%.6f", device.WidthRatio))
			b.WriteString("\n")
		}
	}

	b.WriteString("# HELP pcie_exporter_scrapes_total Total number of metrics scrapes.\n")
	b.WriteString("# TYPE pcie_exporter_scrapes_total counter\n")
	b.WriteString("pcie_exporter_scrapes_total ")
	b.WriteString(strconv.FormatUint(h.scrapes.Load(), 10))
	b.WriteString("\n")

	b.WriteString("# HELP pcie_exporter_scrape_errors_total Total number of scrape-level errors.\n")
	b.WriteString("# TYPE pcie_exporter_scrape_errors_total counter\n")
	b.WriteString("pcie_exporter_scrape_errors_total ")
	b.WriteString(strconv.FormatUint(h.scrapeErrs.Load(), 10))
	b.WriteString("\n")

	b.WriteString("# HELP pcie_exporter_last_scrape_duration_seconds Duration of the most recent scrape in seconds.\n")
	b.WriteString("# TYPE pcie_exporter_last_scrape_duration_seconds gauge\n")
	b.WriteString("pcie_exporter_last_scrape_duration_seconds ")
	b.WriteString(fmt.Sprintf("%.6f", scrapeDuration))
	b.WriteString("\n")

	b.WriteString("# HELP pcie_exporter_last_scrape_success Whether the most recent scrape succeeded.\n")
	b.WriteString("# TYPE pcie_exporter_last_scrape_success gauge\n")
	b.WriteString("pcie_exporter_last_scrape_success ")
	b.WriteString(strconv.Itoa(scrapeSuccess))
	b.WriteString("\n")

	if err != nil {
		b.WriteString("# pcie_exporter_error ")
		b.WriteString(escapeLabelValue(err.Error()))
		b.WriteString("\n")
	}

	_, _ = w.Write([]byte(b.String()))
}

func metricLabels(device pcie.Device) string {
	return "{" +
		`device="` + escapeLabelValue(device.Address) + `",` +
		`vendor_id="` + escapeLabelValue(device.VendorID) + `",` +
		`device_id="` + escapeLabelValue(device.DeviceID) + `",` +
		`class="` + escapeLabelValue(device.Class) + `",` +
		`current_link_speed="` + escapeLabelValue(device.CurrentLinkSpeed) + `",` +
		`max_link_speed="` + escapeLabelValue(device.MaxLinkSpeed) + `",` +
		`current_link_width="` + escapeLabelValue(device.CurrentLinkWidth) + `",` +
		`max_link_width="` + escapeLabelValue(device.MaxLinkWidth) + `"` +
		"}"
}

func escapeLabelValue(value string) string {
	replacer := strings.NewReplacer(
		`\\`, `\\\\`,
		`"`, `\\"`,
		"\n", `\\n`,
	)
	return replacer.Replace(value)
}
