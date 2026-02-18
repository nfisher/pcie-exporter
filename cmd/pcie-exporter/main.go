package main

import (
	"errors"
	"flag"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/nfisher/pcie-exporter/internal/exporter"
)

func main() {
	listenAddress := flag.String("listen-address", ":9808", "HTTP listen address")
	sysfsRootFlag := flag.String("sysfs-root", "", "sysfs root path override (defaults to /sys or PCIE_EXPORTER_SYSFS)")
	flag.Parse()

	sysfsRoot := resolveSysfsRoot(*sysfsRootFlag)

	mux := http.NewServeMux()
	mux.Handle("/metrics", exporter.NewHandler(sysfsRoot))
	mux.HandleFunc("/healthz", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok\n"))
	})

	server := &http.Server{
		Addr:              *listenAddress,
		Handler:           mux,
		ReadHeaderTimeout: 5 * time.Second,
	}

	log.Printf("starting pcie-exporter on %s with sysfs root %s", *listenAddress, sysfsRoot)
	if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		log.Fatal(err)
	}
}

func resolveSysfsRoot(sysfsRootFlag string) string {
	if sysfsRootFlag != "" {
		return sysfsRootFlag
	}
	if fromEnv := os.Getenv("PCIE_EXPORTER_SYSFS"); fromEnv != "" {
		return fromEnv
	}
	return "/sys"
}
