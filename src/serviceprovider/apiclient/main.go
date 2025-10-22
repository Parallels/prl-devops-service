package apiclient

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"time"

	"github.com/Parallels/prl-devops-service/basecontext"
	"github.com/Parallels/prl-devops-service/config"
	"github.com/Parallels/prl-devops-service/constants"
	"github.com/Parallels/prl-devops-service/helpers"
	"github.com/Parallels/prl-devops-service/models"
)

const (
	DEFAULT_API_LOGIN_URL = "api/v1/auth/token"
	defaultTimeout        = 5 * time.Minute
	defaultConnTimeout    = 5 * time.Minute
)

type HttpClientService struct {
	timeout              time.Duration
	connTimeout          time.Duration
	context              context.Context
	ctx                  basecontext.ApiContext
	headers              map[string]string
	authorization        *HttpClientServiceAuthorization
	authorizer           *HttpClientServiceAuthorizer
	disableTlsValidation bool
}

type HttpClientServiceResponse struct {
	StatusCode int
	Data       interface{}
	ApiError   *models.ApiErrorResponse
}

func NewHttpClient(ctx basecontext.ApiContext) *HttpClientService {
	client := &HttpClientService{
		ctx:           ctx,
		headers:       make(map[string]string, 0),
		authorizer:    nil,
		authorization: nil,
	}

	// Set default headers
	client.headers["User-Agent"] = "PrlDevOpsService/ApiClient"
	client.headers[constants.INTERNAL_API_CLIENT] = "true"

	cfg := config.Get()
	if cfg != nil {
		client.disableTlsValidation = cfg.DisableTlsValidation()
	}
	return client
}

func (c *HttpClientService) WithContext(ctx context.Context) *HttpClientService {
	c.context = ctx
	return c
}

func (c *HttpClientService) WithHeader(key, value string) *HttpClientService {
	c.headers[key] = value
	return c
}

func (c *HttpClientService) WithHeaders(headers map[string]string) *HttpClientService {
	for k, v := range headers {
		c.headers[k] = v
	}
	return c
}

func (c *HttpClientService) WithTimeout(duration time.Duration) *HttpClientService {
	c.ctx.LogInfof("[Api Client] Setting timeout to %v", duration)
	context, _ := context.WithTimeout(context.Background(), duration)

	c.context = context
	c.timeout = duration
	c.connTimeout = duration // Use the same timeout for connections
	c.ctx.LogDebugf("[Api Client] Using connection timeout of %v", c.connTimeout)
	return c
}

func (c *HttpClientService) AuthorizeWithUsernameAndPassword(username, password string) *HttpClientService {
	c.authorization = &HttpClientServiceAuthorization{
		Username: username,
		Password: password,
	}

	return c
}

func (c *HttpClientService) AuthorizeWithApiKey(apiKey string) *HttpClientService {
	c.authorization = &HttpClientServiceAuthorization{
		ApiKey: apiKey,
	}

	return c
}

func (c *HttpClientService) SetAuthorization(authorization HttpClientServiceAuthorization) *HttpClientService {
	c.authorization = &authorization
	return c
}

func (c *HttpClientService) Get(url string, destination interface{}) (*HttpClientServiceResponse, error) {
	return c.RequestData(HttpClientServiceVerbGet, url, nil, destination)
}

func (c *HttpClientService) Post(url string, data interface{}, destination interface{}) (*HttpClientServiceResponse, error) {
	return c.RequestData(HttpClientServiceVerbPost, url, data, destination)
}

func (c *HttpClientService) Put(url string, data interface{}, destination interface{}) (*HttpClientServiceResponse, error) {
	return c.RequestData(HttpClientServiceVerbPut, url, data, destination)
}

func (c *HttpClientService) Delete(url string, destination interface{}) (*HttpClientServiceResponse, error) {
	return c.RequestData(HttpClientServiceVerbDelete, url, nil, destination)
}

func (c *HttpClientService) RequestData(verb HttpClientServiceVerb, url string, data interface{}, destination interface{}) (*HttpClientServiceResponse, error) {
	c.ctx.LogInfof("[Api Client] Starting %v request to %s with timeout %v", verb, url, c.timeout)
	var err error
	var req *http.Request
	apiResponse := HttpClientServiceResponse{
		StatusCode: 0,
		Data:       nil,
	}

	if destination != nil {
		destType := reflect.TypeOf(destination)
		if destType.Kind() != reflect.Ptr {
			return &apiResponse, errors.New("dest must be a pointer type")
		}
	}

	if url == "" {
		return &apiResponse, errors.New("url cannot be empty")
	}

	// Ensure timeout is set
	if c.timeout == 0 {
		c.timeout = defaultTimeout
	}

	// Log if using non-default timeout
	if c.timeout != defaultTimeout {
		c.ctx.LogInfof("[Api Client] Using custom timeout of %v for request to %s (default is %v)",
			c.timeout, url, defaultTimeout)
	}

	// Use the same timeout for everything when custom timeout is set
	connTimeout := defaultConnTimeout
	if c.connTimeout > 0 {
		connTimeout = c.connTimeout
	}
	c.ctx.LogDebugf("[Api Client] Using connection timeout %v for request to %s", connTimeout, url)

	// Use appropriate timeouts for the transport
	transport := &http.Transport{
		TLSHandshakeTimeout:   30 * time.Second,
		ResponseHeaderTimeout: c.timeout,
		ExpectContinueTimeout: 30 * time.Second,
		IdleConnTimeout:       c.timeout,
		DisableKeepAlives:     false,
		MaxIdleConns:          100,
		MaxIdleConnsPerHost:   10,
		DialContext: (&net.Dialer{
			Timeout:   30 * time.Second,
			KeepAlive: c.timeout,
		}).DialContext,
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: c.disableTlsValidation,
		},
	}

	client := &http.Client{
		Transport: transport,
		Timeout:   c.timeout,
	}

	// Use the passed context if available, otherwise create new one
	var ctx context.Context
	var cancel context.CancelFunc
	if c.context != nil {
		ctx = c.context
	} else {
		ctx, cancel = context.WithTimeout(context.Background(), c.timeout)
		defer cancel()
	}

	c.context = ctx

	if data != nil {
		reqBody, err := json.MarshalIndent(data, "", "  ")
		if err != nil {
			return &apiResponse, fmt.Errorf("error marshalling data, err: %v", err)
		}
		req, err = http.NewRequestWithContext(ctx, verb.String(), url, bytes.NewBuffer(reqBody))
		if err != nil {
			return &apiResponse, fmt.Errorf("error creating request, err: %v", err)
		}
	} else {
		req, err = http.NewRequestWithContext(ctx, verb.String(), url, nil)
		if err != nil {
			return &apiResponse, fmt.Errorf("error creating request, err: %v", err)
		}
	}

	if req == nil {
		return &apiResponse, fmt.Errorf("request is nil")
	}

	if c.authorization != nil {
		c.authorizer = nil
		if c.authorization.ApiKey != "" {
			c.authorizer = &HttpClientServiceAuthorizer{
				ApiKey: c.authorization.ApiKey,
			}
		}
		if c.authorization.Username != "" && c.authorization.Password != "" {
			c.ctx.LogDebugf("[Api Client] Getting Client Authorization with username %s ", c.authorization.Username)
			token, err := getJwtToken(c.ctx, 1*time.Minute, url, c.authorization.Username, c.authorization.Password)
			if err != nil {
				apiResponse.StatusCode = 401
				return &apiResponse, err
			}
			c.authorizer = &HttpClientServiceAuthorizer{
				BearerToken: token,
			}
		}
	}

	if c.authorizer != nil {
		if c.authorizer.BearerToken != "" {
			c.ctx.LogDebugf("[Api Client] Setting Authorization header to Bearer %s", helpers.ObfuscateString(c.authorizer.BearerToken))
			req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.authorizer.BearerToken))
		} else if c.authorizer.ApiKey != "" {
			c.ctx.LogDebugf("[Api Client] Setting Authorization header to X-Api-Key %s", helpers.ObfuscateString(c.authorizer.ApiKey))
			req.Header.Set("X-Api-Key", c.authorizer.ApiKey)
		}
	}

	if req.Header.Get("Content-Type") == "" && data != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	if len(c.headers) > 0 {
		for k, v := range c.headers {
			req.Header.Set(k, v)
		}
	}

	// Add correlation ID for tracing
	correlationId := helpers.GenerateId()
	c.ctx.LogInfof("[Api Client][%s] Starting request: %s %s (timeout: %v)", correlationId, verb, url, c.timeout)

	// Add Istio trace headers
	req.Header.Set("x-b3-sampled", "1")
	req.Header.Set("x-request-id", correlationId)
	req.Header.Set("x-envoy-upstream-rq-timeout-ms", fmt.Sprintf("%d", c.timeout.Milliseconds()))
	req.Header.Set("x-envoy-expected-rq-timeout-ms", fmt.Sprintf("%d", c.timeout.Milliseconds()))

	if strings.Contains(url, "/machines") {
		req.Header.Set("x-long-running-op", "true")
	}

	response, err := client.Do(req)
	if err != nil {
		if os.IsTimeout(err) {
			c.ctx.LogErrorf("[Api Client][%s] Timeout occurred during request to %s: %v", correlationId, url, err)
		} else if netErr, ok := err.(net.Error); ok {
			c.ctx.LogErrorf("[Api Client][%s] Network error during request to %s: %v (timeout: %v)", correlationId, url, err, netErr.Timeout())
			c.ctx.LogErrorf("[Api Client][%s] Network error type: %T", correlationId, netErr)
		} else {
			c.ctx.LogErrorf("[Api Client][%s] Error during request to %s: %v (type: %T)", correlationId, url, err, err)
		}
		return &apiResponse, fmt.Errorf("error %s data on %s, err: %v", verb, url, err)
	}

	defer response.Body.Close()

	apiResponse.StatusCode = response.StatusCode
	if response.StatusCode < 200 || response.StatusCode >= 300 {
		var errMsg models.ApiErrorResponse
		body, bodyErr := io.ReadAll(response.Body)
		if bodyErr == nil {
			if err := json.Unmarshal(body, &errMsg); err == nil {
				apiResponse.ApiError = &errMsg
			}
		}

		if apiResponse.ApiError != nil && apiResponse.ApiError.Message != "" {
			return &apiResponse, fmt.Errorf("error on %s data from %s, err: %v message: %v", verb, url, apiResponse.ApiError.Code, apiResponse.ApiError.Message)
		} else {
			return &apiResponse, fmt.Errorf("error on %s data from %s, status code: %d", verb, url, response.StatusCode)
		}
	}

	if response.Body != http.NoBody {
		body, err := io.ReadAll(response.Body)
		if err != nil {
			return &apiResponse, fmt.Errorf("error reading response body from %s, err: %v", url, err)
		}
		if destination != nil {
			if response.Header.Get("Content-Type") == "application/json" {
				err = json.Unmarshal(body, destination)
				if err != nil {
					return &apiResponse, fmt.Errorf("error unmarshalling body from %s, err: %v ", url, err)
				}
			} else {
				if strPtr, ok := destination.(*string); ok {
					*strPtr = string(body)
					destination = strPtr
				} else {
					return &apiResponse, fmt.Errorf("destination must be a string pointer for non-JSON responses")
				}
			}

			c.ctx.LogTracef("[Api Client] Response body: \n%s", string(body))
			apiResponse.Data = destination
		} else {
			if len(body) == 0 {
				return &apiResponse, nil
			}

			var bodyData map[string]interface{}
			if response.Header.Get("Content-Type") == "application/json" {
				err = json.Unmarshal(body, &bodyData)
				if err != nil {
					return &apiResponse, fmt.Errorf("error unmarshalling body from %s, err: %v ", url, err)
				}
			} else {
				if strPtr, ok := destination.(*string); ok {
					*strPtr = string(body)
					destination = strPtr
				} else {
					return &apiResponse, fmt.Errorf("destination must be a string pointer for non-JSON responses")
				}
			}

			apiResponse.Data = bodyData
		}
	}
	return &apiResponse, nil
}

func (c *HttpClientService) GetFileFromUrl(fileUrl string, destinationPath string) error {
	// Create the file in the tmp folder
	file, err := os.Create(filepath.Clean(destinationPath))
	if err != nil {
		return err
	}

	defer file.Close()
	httpRequest, err := http.NewRequest("GET", fileUrl, nil)
	if err != nil {
		return err
	}

	// Download the file from the URL
	resp, err := http.DefaultClient.Do(httpRequest)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	// Write the file to disk
	_, err = io.Copy(file, resp.Body)
	if err != nil {
		return err
	}

	return nil
}

func (c *HttpClientService) Authorize(ctx basecontext.ApiContext, url string) (*HttpClientServiceAuthorizer, error) {
	if c.authorization != nil {
		c.authorizer = nil
		if c.authorization.ApiKey != "" {
			c.authorizer = &HttpClientServiceAuthorizer{
				ApiKey: c.authorization.ApiKey,
			}
		}
		if c.authorization.Username != "" && c.authorization.Password != "" {
			c.ctx.LogDebugf("[Api Client] Getting Client Authorization with username %s ", c.authorization.Username)
			token, err := getJwtToken(c.ctx, 1*time.Minute, url, c.authorization.Username, c.authorization.Password)
			if err != nil {
				return nil, err
			}
			c.authorizer = &HttpClientServiceAuthorizer{
				BearerToken: token,
			}
		}

		return c.authorizer, nil
	}

	return c.authorizer, errors.New("error authorizing, no authorization method set")
}

func getJwtToken(ctx basecontext.ApiContext, timeout time.Duration, baseUrl, username, password string) (string, error) {
	if username == "" {
		return "", errors.New("username cannot be empty")
	}

	if password == "" {
		return "", errors.New("password cannot be empty")
	}

	tokenRequest := models.LoginRequest{
		Email:    username,
		Password: password,
	}

	h, err := url.Parse(baseUrl)
	if err != nil {
		return "", err
	}

	hostAndPath := fmt.Sprintf("%s://%s/%s", h.Scheme, h.Host, DEFAULT_API_LOGIN_URL)

	// setting the timeout in the get token request
	c := NewHttpClient(ctx)
	if timeout > 0 {
		c.WithTimeout(timeout)
	}

	c.ctx.LogDebugf("[Api Client] Getting token from %s with username %s and password %s", hostAndPath, username, helpers.ObfuscateString(password))

	var tokenResponse models.LoginResponse
	if _, err := c.Post(hostAndPath, tokenRequest, &tokenResponse); err != nil {
		return "", err
	}
	return tokenResponse.Token, nil
}
