package httpserver

import (
	"blazeginx/internal/config"
	"blazeginx/internal/logger"
	"context"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"strconv"
	"time"
)

type healthUpstreamResponse struct {
	Name           string `json:"name"`
	Status         int    `json:"status,omitempty"`
	ResponseTimeMs int64  `json:"response_time_ms,omitempty"`
	Body           string `json:"response,omitempty"`
	Error          string `json:"error,omitempty"`
}

type healthResponse struct {
	Message           string                   `json:"message"`
	UpstreamResponses []healthUpstreamResponse `json:"upstream_responses"`
}

func NewHealthzHandler(routes []config.Route, timeout time.Duration) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log := logger.GetServiceLogger(r.Context().Value(Logger).(*slog.Logger), "healthz")

		if r.Method != http.MethodGet {
			http.Error(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
			return
		}

		errors := 0
		resps := make([]healthUpstreamResponse, 0, len(routes))

		for _, route := range routes {
			url, err := CreateFullUrl(route, route.HealthPath, "")
			if err != nil {
				log.Error("Failed to create request to upstream",
					"upstream_name", route.Name,
					"error", err,
				)

				errors++
				resps = append(resps, healthUpstreamResponse{
					Name:  route.Name,
					Error: err.Error(),
				})
				continue
			}

			upstreamCtx, cancel := context.WithTimeout(r.Context(), timeout)
			req, err := http.NewRequestWithContext(upstreamCtx, http.MethodGet, url, nil)
			if err != nil {
				log.Error("Failed to create request to upstream",
					"upstream_name", route.Name,
					"error", err,
				)
				cancel()
				errors++
				resps = append(resps, healthUpstreamResponse{
					Name:  route.Name,
					Error: err.Error(),
				})
				continue
			}

			start := time.Now()
			resp, err := http.DefaultClient.Do(req)
			elapsed := time.Since(start)
			cancel()

			upstreamResponse := healthUpstreamResponse{
				Name:           route.Name,
				ResponseTimeMs: elapsed.Milliseconds(),
			}

			if err != nil {
				log.Error("Failed to get response from upstream",
					"upstream_name", route.Name,
					"error", err,
				)
				if upstreamCtx.Err() == context.DeadlineExceeded {
					upstreamResponse.Error = "timeout"
					upstreamResponse.ResponseTimeMs = 0
				} else {
					upstreamResponse.Error = err.Error()
				}

				errors++
				resps = append(resps, upstreamResponse)
				continue
			}

			if resp.StatusCode != http.StatusOK {
				errors++
			}

			body, err := io.ReadAll(resp.Body)
			if err != nil {
				log.Error("Failed to read upstream response body",
					"upstream_name", route.Name,
					"error", err,
				)
				if resp.StatusCode == http.StatusOK {
					errors++
				}
				upstreamResponse.Error = err.Error()
			}
			defer func() {
				_ = resp.Body.Close()
			}()

			upstreamResponse.Status = resp.StatusCode
			upstreamResponse.Body = string(body)

			resps = append(resps, upstreamResponse)
		}

		status := http.StatusOK
		message := "OK"

		if errors > 0 {
			status = http.StatusInternalServerError
			message = strconv.Itoa(errors) + " upstream requests failed"
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(status)

		err := json.NewEncoder(w).Encode(healthResponse{
			Message:           message,
			UpstreamResponses: resps,
		})
		if err != nil {
			log.Error("Failed to encode health response",
				"error", err,
			)
		}
	})
}
