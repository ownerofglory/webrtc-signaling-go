package handler

import "net/http"

const GetVersionPath = basePath + "/version"

var AppVersion = "default"

func HandleGetVersion(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(AppVersion))
}
