package exporter

import (
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"testing"

	"github.com/gogunit/gunit/hammy"
)

func TestHandlerServesMetrics(t *testing.T) {
	h := hammy.New(t)

	sysfsRoot := filepath.Join("..", "pcie", "testdata", "sysfs")
	handler := NewHandler(sysfsRoot)

	req := httptest.NewRequest(http.MethodGet, "/metrics", nil)
	resp := httptest.NewRecorder()
	handler.ServeHTTP(resp, req)

	h.Is(hammy.Number(resp.Code).EqualTo(http.StatusOK))
	h.Is(hammy.String(resp.Header().Get("Content-Type")).Contains("text/plain"))

	body := resp.Body.String()
	h.Is(hammy.String(body).Contains("pcie_devices_total 2"))
	h.Is(hammy.String(body).Contains("pcie_link_negotiated_ok{device=\"0000:01:00.0\""))
	h.Is(hammy.String(body).Contains("pcie_link_negotiated_ok{device=\"0000:02:00.0\""))
	h.Is(hammy.String(body).Contains("pcie_exporter_last_scrape_success 1"))
}
