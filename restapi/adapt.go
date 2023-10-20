package restapi

import "net/http"

func Adapt(handler http.Handler, adapters ...Adapter) http.Handler {
	for i := len(adapters); i > 0; i-- {
		handler = adapters[i-1](handler)
	}

	return handler
}
