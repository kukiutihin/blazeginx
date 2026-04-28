package httpserver

import (
	"blazeginx/internal/config"
	"blazeginx/pkg/storage"
	"log/slog"
	"net/http"
	"time"
)

type userInfo struct {
	lastRequestTime time.Time
	tokensRemain    uint
}

type TokenBucket struct {
	data       *storage.Storage
	maxTokens  uint
	refillRate uint
}

func NewTokenBucket(cfg config.Config, store *storage.Storage) TokenBucket {
	return TokenBucket{
		data:       store,
		maxTokens:  cfg.RateLimit.MaxTokens,
		refillRate: cfg.RateLimit.RefillRateInSecs,
	}
}

// isAllow waits addr only without port
func (b *TokenBucket) isAllow(addr string) bool {
	info, ok := b.data.Get(addr)
	var passed uint
	var newTokensRemain uint

	if ok {
		info, ok := info.(userInfo)
		if !ok {
			panic("Incorrect value in storage")
		}

		passed = uint(time.Since(info.lastRequestTime).Seconds())
		newTokensRemain = max(
			b.maxTokens,
			info.tokensRemain+b.refillRate*uint(passed),
		)
	} else {
		newTokensRemain = b.maxTokens
	}

	var result bool
	if newTokensRemain > 0 {
		result = true
	} else {
		result = false
	}

	b.data.Add(addr, userInfo{
		lastRequestTime: time.Now(),
		tokensRemain:    newTokensRemain,
	})
	return result
}

func (b *TokenBucket) BuildRateLimiter() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			log := r.Context().Value(Logger).(*slog.Logger)
			addr := r.URL.Hostname()
			if !b.isAllow(addr) {
				http.Error(w, "Too many requests", http.StatusTooManyRequests)
				log.Warn("Request rate limit exceeded, request is canceled",
					"addr", addr,
				)
			}

			next.ServeHTTP(w, r)
		})
	}
}
