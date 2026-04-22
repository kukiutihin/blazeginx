package httpserver

import "strings"

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
