package httpserver

import "context"

type ContextValues string

const (
	Logger ContextValues = "logger"
)

// IsDone just checks context is Done.
func IsDone(ctx context.Context) bool {
	select {
	case <-ctx.Done():
		return true
	default:
		return false
	}
}
