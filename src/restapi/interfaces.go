package restapi

import "net/http"

type RestApiController interface {
	Serve() error
}

type Adapter func(http.Handler) http.Handler

type ControllerHandler func(w http.ResponseWriter, r *http.Request)
