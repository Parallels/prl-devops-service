package restapi

import (
	"Parallels/pd-api-service/common"
	"context"
	"net/http"
	"regexp"

	"github.com/google/uuid"
)

const (
	REQUEST_ID_KEY = "requestId"
)

func RequestIdMiddlewareAdapter() Adapter {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			id := uuid.New().String()
			r.Header.Add("X-Request-Id", id)
			ctx := context.WithValue(r.Context(), REQUEST_ID_KEY, id)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func JsonContentMiddlewareAdapter() Adapter {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Add("Content-Type", "application/json")
			next.ServeHTTP(w, r)
		})
	}
}

func LoggerMiddlewareAdapter(logHealthCheck bool) Adapter {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			shouldLog := true
			rMatch := regexp.MustCompile("health")
			if rMatch.MatchString(r.URL.Path) && !logHealthCheck {
				shouldLog = false
			}

			if shouldLog {
				id := GetRequestId(r)
				common.Logger.Info("[%s] [%v] %v from %v", id, r.Method, r.URL.Path, r.Host)
			}

			next.ServeHTTP(w, r)
		})
	}
}
