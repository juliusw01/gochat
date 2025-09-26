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

func GetClient(w http.ResponseWriter, r *http.Request) {

	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var clientPath string

	operatingSystem := r.PathValue("os")
	switch operatingSystem {
	case "linux":
		clientPath = "bin/clients/linux/gochat"
	case "mac":
		clientPath = "bin/clients/mac/gochat"
	case "windows":
		clientPath = "bin/clients/windows/gochat"
	default:
		http.Error(w, "Operating System not supported. Supported systems are 'linux', 'mac', and 'windows'.", http.StatusBadRequest)
		return
	}

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
	http.Redirect(w, r, "http://raspberrypi.fritz.box:8080/", http.StatusFound)
}
