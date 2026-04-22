package httpserver

import "net/http"


func FailIfErr(w http.ResponseWriter, err error, status int) bool {
	if err == nil {
		return false
	}

	http.Error(w, http.StatusText(status), status)
	return true
}

