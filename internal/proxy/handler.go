package proxy

import (
	"blazeginx/internal/httpserver"
	"io"
	"log/slog"
	"net/http"
)

func BuildHandler(path string) func(w http.ResponseWriter, r *http.Request) {
    return func(w http.ResponseWriter, r *http.Request) {
        log := r.Context().Value(httpserver.Logger).(*slog.Logger)
        log.Debug("Received in proxy", "path", path)
        
        req, err := http.NewRequest(r.Method, path, r.Body)
        if httpserver.FailIfErr(w, err, 500) {
            log.Error(
                "Failed to build request for service",
                "service_url", path,
                "error", err, 
            )
            return
        }

        req.Header = r.Header.Clone()

        resp, err := http.DefaultClient.Do(req)
        if httpserver.FailIfErr(w, err, 500) {
            log.Error(
                "Failed to send request to service",
                "service_url", path,
                "error", err, 
            )
            return
        }
        defer func() {
            err := resp.Body.Close()
            if err != nil {
                log.Error(
                    "Failed to close connection",
                    "error", err,
                )
            }
        }()
        
        log.Debug(
            "Received a responce from service",
            "service_url", path,
            "status", resp.Status,
        )

        for key, values := range resp.Header {
            for _, value := range values {
                w.Header().Add(key, value)         
            }
        } 

        w.WriteHeader(resp.StatusCode)
        _, err = io.Copy(w, resp.Body)
        if httpserver.FailIfErr(w, err, 500) {
            log.Error(
                "Failed to send respond",
                "error", err,
            )
        }
    }
}
