/**
*
* The idea is, that the client can be downloaded directly from the homepage of the server
* TODO: there should also be an update function that will update the client via an update command
*
**/

package client

import (
	"net/http"
	"os"
	"path/filepath"
)

func GetClient(w http.ResponseWriter, r *http.Request){
	clientPath := "client/gochat"

	file, err := os.Open(clientPath)
	if err != nil {
		http.Error(w, "Client not found", http.StatusNotFound)
		return
	}
	defer file.Close()

	fileInfo, err := file.Stat()
	if err != nil {
		http.Error(w, "Could not get file info", http.StatusInternalServerError)
	}

	w.Header().Set("Content-Disposition", "attachment; filename="+filepath.Base(clientPath))
    w.Header().Set("Content-Type", "application/octet-stream")
    w.Header().Set("Content-Length", string(rune(fileInfo.Size())))

	http.ServeContent(w, r, clientPath, fileInfo.ModTime(), file)
}
