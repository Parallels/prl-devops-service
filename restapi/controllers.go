package restapi

import "net/http"

type Controller func(w http.ResponseWriter, r *http.Request)
