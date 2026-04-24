package httpserver

import (
	"blazeginx/internal/config"
	"strings"
)

func stripRoutePrefix(path string, pref string) string {
	if path == "" || pref == "" {
		return path
	}

	path = "/" + strings.Trim(path, "/")
	pref = "/" + strings.Trim(pref, "/")

	if path == pref {
		return "/"
	}

	if strings.HasPrefix(path, pref+"/") {
		return strings.TrimPrefix(path, pref)
	}

	return path
}

// Returns a full url as addr/path, also removes prefix from path if it need
func CreateUrl(route config.Route, addr string, path string) string {
	if route.StripPrefix {
		path = stripRoutePrefix(path, route.Path)
	}
	addr = strings.TrimSuffix(addr, "/")
	return addr + path
}
