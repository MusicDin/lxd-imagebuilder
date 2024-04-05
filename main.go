package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"path/filepath"
	"text/template"

	"github.com/tdewolff/minify/v2"
	"github.com/tdewolff/minify/v2/css"
	"github.com/tdewolff/minify/v2/html"
	"github.com/tdewolff/minify/v2/js"
	"github.com/tdewolff/minify/v2/svg"

	"github.com/canonical/lxd-imagebuilder/embed"
	"github.com/canonical/lxd-imagebuilder/shared"
	"github.com/canonical/lxd-imagebuilder/simplestream-maintainer/stream"
	"github.com/canonical/lxd-imagebuilder/simplestream-maintainer/testutils"
	"github.com/canonical/lxd-imagebuilder/simplestream-maintainer/webpage"
)

func main() {
	addr := "0.0.0.0:8080"
	h := http.NewServeMux()

	// Serve static files from ./tmp/images for /images path
	h.Handle("/images/", http.StripPrefix("/images/", http.FileServer(http.Dir("./tmp/images"))))

	// Handle other routes
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

	rootDir, err := filepath.Abs("tmp")
	if err != nil {
		writeError(w, err)
		return
	}

	err = mockImageServerContent(rootDir, catalog)
	if err != nil {
		writeError(w, err)
		return
	}

	tplName := "templates/index.html"

	customFuncs := template.FuncMap{
		// json returns a JSON representation of the given value.
		"json": func(v any) (string, error) {
			json, err := json.Marshal(v)
			if err != nil {
				return "", err
			}

			return string(json), nil
		},
	}

	t, err := template.New(filepath.Base(tplName)).Funcs(customFuncs).ParseFS(embed.GetTemplates(), tplName)
	if err != nil {
		writeError(w, err)
		return
	}

	webpage, err := webpage.NewWebPage(rootDir, *catalog)
	if err != nil {
		writeError(w, err)
		return
	}

	var buf bytes.Buffer

	err = t.Execute(&buf, webpage)
	if err != nil {
		writeError(w, err)
		return
	}

	m := minify.New()
	m.AddFunc("text/html", html.Minify)
	m.AddFunc("text/css", css.Minify)
	m.AddFunc("image/svg+xml", svg.Minify)
	m.AddFunc("text/x-handlebars-template", html.Minify)
	m.AddFunc("application/javascript", js.Minify)

	err = m.Minify("text/html", w, &buf)
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

func mockImageServerContent(basePath string, catalog *stream.ProductCatalog) error {
	basePath = filepath.Join(basePath, "images")

	err := os.MkdirAll(basePath, 0755)
	if err != nil {
		return err
	}

	for _, p := range catalog.Products {
		mp := testutils.MockProduct(p.RelPath())

		for vid := range p.Versions {
			v := testutils.MockVersion(vid)

			for iid := range p.Versions[vid].Items {
				v = v.WithFiles(iid)
			}

			// Add some generic files.
			v = v.WithFiles("SHA256SUMS", "image.yaml", "serial")
			mp = mp.AddVersions(v)
		}

		err := mp.CreateOnPath(basePath)
		if err != nil {
			return err
		}
	}

	return nil
}
