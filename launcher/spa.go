package launcher

import (
	"log"
	"net/http"
	"os"
	"path"
	"path/filepath"

	"github.com/gorilla/mux"
)

type spaHandler struct {
	staticPath string
	indexPath  string
}

func (h spaHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	urlPath := r.URL.Path
	var err error
	log.Println("path", urlPath)
	if !path.IsAbs(r.URL.Path) {
		http.Error(w, "Bad path", http.StatusBadRequest)
		return
	}

	// prepend the path with the path to the static directory
	path := filepath.Join(h.staticPath, urlPath)
	log.Println("path to resource", path)
	// check whether a file exists at the given path
	_, err = os.Stat(path)
	if os.IsNotExist(err) {
		// file does not exist, serve index.html
		http.ServeFile(w, r, filepath.Join(h.staticPath, h.indexPath))
		return
	} else if err != nil {
		// if we got an error (that wasn't that the file doesn't exist) stating the
		// file, return a 500 internal server error and stop
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// otherwise, use http.FileServer to serve the static dir
	http.FileServer(http.Dir(h.staticPath)).ServeHTTP(w, r)
}

func addSpaHandler(router *mux.Router) {
	relativePath := `../../launcher-frontend/dist/launcher-frontend`
	absPath, err := filepath.Abs(relativePath)
	if err != nil {
		log.Printf("failed to make abc path: %v", err)
		os.Exit(1)
	}
	log.Printf("abs path is %s\n", absPath)
	spa := spaHandler{staticPath: absPath, indexPath: "index.html"}
	router.PathPrefix("/").Handler(spa)
}
