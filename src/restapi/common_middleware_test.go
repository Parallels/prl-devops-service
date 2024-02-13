package restapi

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Parallels/prl-devops-service/constants"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRequestIdMiddlewareAdapter(t *testing.T) {
	var finalRequest *http.Request
	var requestId string
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestId = r.Context().Value(constants.REQUEST_ID_KEY).(string)
		assert.NotEmpty(t, requestId)
		finalRequest = r
		w.WriteHeader(http.StatusOK)
	})

	adapter := RequestIdMiddlewareAdapter()
	wrappedHandler := adapter(handler)

	req, err := http.NewRequest(http.MethodGet, "/", nil)
	require.NoError(t, err)

	rr := httptest.NewRecorder()
	wrappedHandler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
	assert.NotEmpty(t, finalRequest.Header.Get("X-Request-Id"))
	assert.Equal(t, finalRequest.Header.Get("X-Request-Id"), requestId)
}
