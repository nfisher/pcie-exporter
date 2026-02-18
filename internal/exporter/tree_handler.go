package exporter

import (
	"encoding/json"
	"net/http"

	"github.com/nfisher/pcie-exporter/internal/pcie"
)

// TreeHandler serves PCIe topology in JSON format.
type TreeHandler struct {
	sysfsRoot string
}

func NewTreeHandler(sysfsRoot string) *TreeHandler {
	return &TreeHandler{sysfsRoot: sysfsRoot}
}

func (h *TreeHandler) ServeHTTP(w http.ResponseWriter, _ *http.Request) {
	tree, err := pcie.ReadTree(h.sysfsRoot)
	if err != nil {
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		w.WriteHeader(http.StatusInternalServerError)
		_ = json.NewEncoder(w).Encode(map[string]string{
			"error": err.Error(),
		})
		return
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(tree)
}
