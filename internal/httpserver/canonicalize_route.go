package httpserver

import (
	"strings"
)

// This func leads route to chi-readable: /something/*
func CanonicalRoute(route string) string {
	if route == "" {
		panic("Route cannot be empty")
	}

	route = strings.TrimLeft(route, " /")
	route = strings.TrimRight(route, " /")

	route = "/" + route
	route += "/*"
	return route
}
