package exporter

import (
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"testing"

	"github.com/gogunit/gunit/hammy"
)

func TestTreeHandlerServesJSONTree(t *testing.T) {
	h := hammy.New(t)

	sysfsRoot := filepath.Join("..", "pcie", "testdata", "sysfs")
	handler := NewTreeHandler(sysfsRoot)

	req := httptest.NewRequest(http.MethodGet, "/pcie-tree", nil)
	resp := httptest.NewRecorder()
	handler.ServeHTTP(resp, req)

	h.Is(hammy.Number(resp.Code).EqualTo(http.StatusOK))
	h.Is(hammy.String(resp.Header().Get("Content-Type")).Contains("application/json"))

	body := resp.Body.String()
	h.Is(hammy.String(body).Contains(`"bus_id":"0000:01:00.0"`))
	h.Is(hammy.String(body).Contains(`"link_capacity":"16 GT/s PCIe x16"`))
	h.Is(hammy.String(body).Contains(`"link_status":"8 GT/s PCIe x8"`))
}
