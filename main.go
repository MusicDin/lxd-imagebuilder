package main

import (
	"fmt"
	"log/slog"
	"net/http"
	"text/template"

	"github.com/canonical/lxd-imagebuilder/shared"
	"github.com/canonical/lxd-imagebuilder/simplestream-maintainer/stream"
	"github.com/canonical/lxd-imagebuilder/simplestream-maintainer/webpage"
)

func main() {
	addr := "127.0.0.1:8080"
	h := http.NewServeMux()
	h.HandleFunc("/", handleWebpage)

	slog.Info("Starting server", "addr", addr)
	err := http.ListenAndServe(addr, h)
	if err != nil {
		slog.Error(fmt.Sprintf("Error: %v", err))
	}
}

func handleWebpage(w http.ResponseWriter, r *http.Request) {
	catalog, err := shared.ReadJSONFile("embed/templates/catalog.json", &stream.ProductCatalog{})
	if err != nil {
		writeError(w, err)
		return
	}

	t, err := template.ParseFiles("embed/templates/index.html")
	if err != nil {
		writeError(w, err)
		return
	}

	err = t.Execute(w, webpage.NewWebPage(*catalog))
	if err != nil {
		writeError(w, err)
		return
	}
}

func writeError(w http.ResponseWriter, err error) {
	msg := fmt.Sprintf("Error: %v", err)
	slog.Error(msg)
	_, err = w.Write([]byte(msg))
	if err != nil {
		slog.Error(fmt.Sprintf("PANIC: %v", err))
	}
}
