package storage

import (
	"fmt"
	"github.com/lietu/stream-manager/config"
	"log"
	"net/http"
	"os"
)

func ensureDirectory(path string) {
	os.MkdirAll(path, 0750)
}

func maxAgeHandler(seconds int, h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("Cache-Control", fmt.Sprintf("max-age=%d, public, must-revalidate, proxy-revalidate", seconds))
		h.ServeHTTP(w, r)
	})
}

func ConfigureOverlayHTTP(config *config.Config) *http.ServeMux {
	// TODO: These don't belong here
	ensureDirectory(config.CustomFilesPath)
	ensureDirectory(config.MongoHosts)

	s := http.NewServeMux()
	s.Handle("/core/", maxAgeHandler(0, http.StripPrefix("/core/", http.FileServer(http.Dir(config.OverlayCorePath)))))
	s.Handle("/", maxAgeHandler(0, http.FileServer(http.Dir(config.CustomFilesPath))))

	log.Printf("Serving overlay files for / from %s", config.CustomFilesPath)
	log.Printf("Serving overlay files for /core from %s", config.OverlayCorePath)

	return s
}
