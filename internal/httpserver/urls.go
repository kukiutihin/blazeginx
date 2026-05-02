package httpserver

import (
	"blazeginx/internal/config"
	"net/url"
	"strings"
)

func stripRoutePrefix(path string, pref string) string {
	if path == "" || pref == "" {
		return path
	}

	path = "/" + strings.TrimPrefix(path, "/")
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
func CreateFullUrl(route config.Route, path string, query string) (string, error) {
	if route.StripPrefix {
		path = stripRoutePrefix(path, route.Path)
	}

	u, err := url.Parse(route.Url)
	if err != nil {
		return "", err
	}

	u.Path = path
	u.RawQuery = query

	return u.String(), nil
}
