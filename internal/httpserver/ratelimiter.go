package httpserver

import (
	"blazeginx/internal/config"
	"blazeginx/internal/logger"
	"blazeginx/pkg/storage"
	"log"
	"log/slog"
	"net/http"
	"time"
)

type userInfo struct {
	lastRequestTime time.Time
	tokensRemain    uint
}

// userInfoStorage is storage.Storage wrap for specific type: userInfo.
// it will panic if type is not a userInfo
type userInfoStorage struct {
	storage *storage.Storage
}

func (s *userInfoStorage) Add(key string, value userInfo) {
	s.storage.Add(key, value)
}

func (s *userInfoStorage) Get(key string) (*userInfo, bool) {
	valueAny, ok := s.storage.Get(key)
	if !ok {
		return nil, false
	}
	value, ok := valueAny.(userInfo)
	if !ok {
		log.Fatalf("Incorrect value in storage")
	}
	return &value, true
}

type TokenBucket struct {
	data       *userInfoStorage
	maxTokens  uint
	refillRate time.Duration
}

func NewTokenBucket(cfg config.RateLimit) TokenBucket {
	return TokenBucket{
		data: &userInfoStorage{storage.New(
			cfg.DefaultExpiration,
			cfg.CleanupInterval,
		)},
		maxTokens:  cfg.MaxTokens,
		refillRate: cfg.RefillRate,
	}
}

// isAllow waits addr only without port
func (b *TokenBucket) isAllow(addr string) bool {
	info, ok := b.data.Get(addr)
	var newTokensRemain uint

	if ok {
		tokensAdd := time.Since(info.lastRequestTime).Nanoseconds() / b.refillRate.Nanoseconds()
		newTokensRemain = min(
			b.maxTokens,
			info.tokensRemain+uint(tokensAdd),
		)
	} else {
		newTokensRemain = b.maxTokens
	}

	var result bool
	if newTokensRemain > 0 {
		result = true
		newTokensRemain--
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
			log := logger.GetServiceLogger(r.Context().Value(Logger).(*slog.Logger), "ratelimiter")
			addr := r.URL.Hostname()
			if !b.isAllow(addr) {
				http.Error(w, http.StatusText(http.StatusTooManyRequests), http.StatusTooManyRequests)
				log.Warn("Request rate limit exceeded, request was canceled",
					"addr", addr,
				)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}
