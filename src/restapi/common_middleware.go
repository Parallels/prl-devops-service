package restapi

import (
	"context"
	"net/http"
	"regexp"
	"strings"

	"github.com/Parallels/pd-api-service/common"
	"github.com/Parallels/pd-api-service/config"
	"github.com/Parallels/pd-api-service/constants"
	"github.com/cjlapao/common-go/helper/http_helper"

	"github.com/google/uuid"
)

func SetDefaultVersionMiddlewareAdapter() Adapter {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			conf := config.NewConfig()
			initialUrl := http_helper.JoinUrl(conf.GetApiPrefix(), globalHttpListener.Options.DefaultApiVersion)
			if !strings.HasPrefix(r.URL.Path, initialUrl) {
				r.URL.Path = http_helper.JoinUrl(initialUrl, r.URL.Path)
			}
			// r.URL.Path = "/v1" + r.URL.Path
			ctx := r.Context()
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func RequestIdMiddlewareAdapter() Adapter {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			id := uuid.New().String()
			r.Header.Add("X-Request-Id", id)
			ctx := context.WithValue(r.Context(), constants.REQUEST_ID_KEY, id)
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
