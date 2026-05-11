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
//
// (' ', ?, #) inside path -> error
//
// multiple slashes inside path -> error
func NormalizeRoutePath(path string) (string, error) {
	path = strings.TrimSpace(path)
	if path == "" {
		return "", errors.New("route path canno`t be empty")
	}

	if strings.Contains(path, " ") {
		return "", errors.New("route path canno`t contains spaces")
	}

	if strings.Contains(path, "#") {
		return "", errors.New("route path canno`t contains #")
	}

	if strings.Contains(path, "?") {
		return "", errors.New("route path canno`t contains ?")
	}

	hasTrailing := path[len(path)-1] == '/'

	path = strings.Trim(path, "/")
	if path == "" {
		return "/", nil
	}

	if strings.Contains(path, "//") {
		return "", errors.New("route path canno`t contains //")
	}

	path = "/" + path
	if hasTrailing {
		path = path + "/"
	}
	return path, nil
}

// NormalizeUrl strips slashes and space.
// Remove path, query, fragment.
// Check host is not empty.
// Allows only http or https (for now).
func NormalizeUrl(raw string) (*url.URL, error) {
	raw = strings.ToLower(raw)
	raw = strings.Trim(raw, " /")
	if raw == "" {
		return nil, errors.New("url canno`t be empty")
	}
	u, err := url.Parse(raw)
	if err != nil {
		return nil, errors.New("failed to parse url: " + err.Error())
	}
	switch u.Scheme {
	case "http", "https":
	default:
		return nil, errors.New("unsupported url scheme: " + u.Scheme)
	}

	if u.Host == "" {
		return nil, errors.New("host canno`t be empty")
	}

	u.Path = ""
	u.RawPath = ""
	u.RawQuery = ""
	u.Fragment = ""

	return u, nil
}

// StripPrefix removes prefix from path.
// It waits already normalized path and prefix(/some or /some/).
// Also it doesnt check that path contains prefix.
//
// (/some, /) -> /some
//
// (/some, /some) -> /
//
// (/some/other, /some) -> /other
//
// (/some/other/, /some) -> /other/
func StripPrefix(path, prefix string) string {
	if path == prefix {
		return "/"
	}

	if prefix == "/" {
		return path
	}

	if prefix[len(prefix)-1] != '/' {
		prefix = prefix + "/"
	}
	path = strings.TrimPrefix(path, prefix)

	// path len always > 0 due to first check
	if path[0] != '/' {
		path = "/" + path
	}
	return path
}
