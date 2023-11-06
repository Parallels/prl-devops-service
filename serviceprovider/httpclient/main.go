package httpclient

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"reflect"
	"strings"

	"github.com/Parallels/pd-api-service/basecontext"
	"github.com/Parallels/pd-api-service/errors"
	"github.com/Parallels/pd-api-service/models"
)

type HttpCaller struct{}

func NewHttpCaller() *HttpCaller {
	return &HttpCaller{}
}

func (c *HttpCaller) Get(ctx basecontext.ApiContext, url string, headers *map[string]string, auth *HttpClientAuthorization, destination interface{}) (*HttpClientResponse, error) {
	return c.requestDataToClient(ctx, HttpCallerVerbGet, url, headers, nil, auth, destination)
}

func (c *HttpCaller) Post(ctx basecontext.ApiContext, url string, headers *map[string]string, data interface{}, auth *HttpClientAuthorization, destination interface{}) (*HttpClientResponse, error) {
	return c.requestDataToClient(ctx, HttpCallerVerbPost, url, headers, data, auth, destination)
}

func (c *HttpCaller) Put(ctx basecontext.ApiContext, url string, headers *map[string]string, data interface{}, auth *HttpClientAuthorization, destination interface{}) (*HttpClientResponse, error) {
	return c.requestDataToClient(ctx, HttpCallerVerbPut, url, headers, data, auth, destination)
}

func (c *HttpCaller) Delete(ctx basecontext.ApiContext, url string, headers *map[string]string, auth *HttpClientAuthorization, destination interface{}) (*HttpClientResponse, error) {
	return c.requestDataToClient(ctx, HttpCallerVerbDelete, url, headers, nil, auth, destination)
}

func (c *HttpCaller) requestDataToClient(ctx basecontext.ApiContext, verb HttpClientVerb, baseUrl string, headers *map[string]string, data interface{}, auth *HttpClientAuthorization, destination interface{}) (*HttpClientResponse, error) {
	ctx.LogInfo("%v data from %s", verb, baseUrl)
	var err error
	clientResponse := HttpClientResponse{
		StatusCode: 500,
		Data:       nil,
	}

	baseUrl = c.getHostWithProtocol(baseUrl)
	_, err = url.Parse(baseUrl)
	if err != nil {
		return &clientResponse, fmt.Errorf("error parsing url %s, err: %v", baseUrl, err)
	}

	if destination != nil {
		var destType = reflect.TypeOf(destination)
		if destType.Kind() != reflect.Ptr {
			return &clientResponse, errors.New("dest must be a pointer type")
		}
	}

	if baseUrl == "" {
		return &clientResponse, errors.New("url cannot be empty")
	}

	client := http.DefaultClient
	var req *http.Request

	if data != nil {
		reqBody, err := json.Marshal(data)
		if err != nil {
			return &clientResponse, fmt.Errorf("error marshalling data, err: %v", err)
		}
		req, err = http.NewRequest(verb.String(), baseUrl, bytes.NewBuffer(reqBody))
		if err != nil {
			return &clientResponse, fmt.Errorf("error creating request, err: %v", err)
		}
	} else {
		req, err = http.NewRequest(verb.String(), baseUrl, nil)
		if err != nil {
			return &clientResponse, fmt.Errorf("error creating request, err: %v", err)
		}
	}

	if req == nil {
		return &clientResponse, fmt.Errorf("request is nil")
	}

	if auth != nil {
		if auth.BearerToken != "" {
			ctx.LogInfo("Setting Authorization header to Bearer %s", auth.BearerToken)
			req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", auth.BearerToken))
		} else if auth.ApiKey != "" {
			req.Header.Set("X-Api-Key", auth.ApiKey)
		}
	}

	req.Header.Set("Content-Type", "application/json")
	if headers != nil && len(*headers) > 0 {
		for k, v := range *headers {
			req.Header.Set(k, v)
		}
	}

	response, err := client.Do(req)
	if err != nil {
		return &clientResponse, errors.Newf("error %s data on %s, err: %v", verb, baseUrl, err)
	}

	clientResponse.StatusCode = response.StatusCode
	if response.StatusCode < 200 || response.StatusCode >= 300 {
		var errMsg models.ApiErrorResponse
		body, bodyErr := io.ReadAll(response.Body)
		if bodyErr == nil {
			if err := json.Unmarshal(body, &errMsg); err == nil {
				clientResponse.ApiError = &errMsg
			}
		}
		return &clientResponse, fmt.Errorf("error %s data from %s, status code: %d", verb, baseUrl, response.StatusCode)
	}

	if destination != nil {
		body, err := io.ReadAll(response.Body)
		if err != nil {
			return &clientResponse, fmt.Errorf("error reading response body from %s, err: %v", baseUrl, err)
		}

		err = json.Unmarshal(body, destination)
		if err != nil {
			return &clientResponse, fmt.Errorf("error unmarshalling body from %s, err: %v ", baseUrl, err)
		}

		clientResponse.Data = destination
	}

	return &clientResponse, nil
}

func (c *HttpCaller) GetJwtToken(ctx basecontext.ApiContext, baseUrl, username, password string) (string, error) {
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

	ctx.LogInfo("Getting token from %s/api/v1/auth/token with username %s and password %s", baseUrl, username, password)

	var tokenResponse models.LoginResponse
	if _, err := c.Post(ctx, baseUrl+"/api/v1/auth/token", nil, tokenRequest, nil, &tokenResponse); err != nil {
		return "", err
	}
	return tokenResponse.Token, nil
}

func (c *HttpCaller) GetFileFromUrl(fileUrl string, destinationPath string) error {
	// Create the file in the tmp folder
	file, err := os.Create(destinationPath)
	if err != nil {
		return err
	}

	defer file.Close()

	// Download the file from the URL
	resp, err := http.Get(fileUrl)
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

func (c *HttpCaller) getHostWithProtocol(host string) string {
	if strings.HasPrefix(host, "http://") || strings.HasPrefix(host, "https://") {
		return host
	}

	return "http://" + host
}

func (c *HttpCaller) CleanUrlSuffixAndPrefix(url string) string {
	url = strings.TrimPrefix(url, "/")
	url = strings.TrimSuffix(url, "/")
	return url
}

func (c *HttpCaller) CleanUrlPrefix(url string) string {
	url = strings.TrimPrefix(url, "/")
	return url
}

func (c *HttpCaller) CleanUrlSuffix(url string) string {
	url = strings.TrimSuffix(url, "/")
	return url
}

func (c *HttpCaller) CleanUrlPrefixAndSuffix(url string) string {
	url = strings.TrimPrefix(url, "/")
	url = strings.TrimSuffix(url, "/")
	return url
}

func (c *HttpCaller) CleanUrlPrefixAndSuffixAndDoubleSlash(url string) string {
	url = strings.TrimPrefix(url, "/")
	url = strings.TrimSuffix(url, "/")
	url = strings.ReplaceAll(url, "//", "/")
	return url
}
