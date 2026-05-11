package routing

import (
	"errors"
	"net/url"
	"strings"
)

// /some/ -> /some/
//
// /some  -> /some
//
// some   -> /some
//
// some/  -> /some/
//
// /      -> /
//
// ""     -> error
func NormalizeRoutePath(path string) (string, error) {
	path = strings.TrimSpace(path)
	if path == "" {
		return "", errors.New("route path canno`t be empty")
	}

	hasTrailing := path[len(path)-1] == '/'
	path = strings.Trim(path, "/")
	path = "/" + path
	if hasTrailing {
		path = path + "/"
	}
	return path, nil
}

// Strips slashes and space.
// Remove path, query, fragment.
// Check host is not empty.
// Allows only http or https (for now).
func NormalizeUrl(raw string) (string, error) {
	raw = strings.Trim(raw, " /")
	if raw == "" {
		return "", errors.New("url canno`t be empty")
	}
	u, err := url.Parse(raw)
	if err != nil {
		return "", errors.New("failed to parse url: " + err.Error())
	}
	switch u.Scheme {
	case "http", "https":
	default:
		return "", errors.New("unsupported url scheme: " + u.Scheme)
	}

	if u.Host == "" {
		return "", errors.New("host canno`t be empty")
	}

	u.Path = ""
	u.RawPath = ""
	u.RawQuery = ""
	u.Fragment = ""

	return u.String(), nil
}
