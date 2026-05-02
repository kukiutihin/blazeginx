package httpserver

import (
	"strings"
)

// This func leads any route to /something/*
func NormalizeProxyRoute(route string) string {
	if route == "" {
		panic("Route cannot be empty")
	}

	route = strings.Trim(route, " /")

	route = "/" + route
	route += "/*"
	return route
}
