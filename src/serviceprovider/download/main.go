package download

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"time"
)

var globalDownloadService *DownloadService

type DownloadHttpMethod string

const (
	DownloadHttpMethodGet  DownloadHttpMethod = "GET"
	DownloadHttpMethodPost DownloadHttpMethod = "POST"
)

type DownloadService struct {
	ChunkSize int
	Retries   int
}

func NewDownloadService() *DownloadService {
	if globalDownloadService == nil {
		globalDownloadService = &DownloadService{
			ChunkSize: 5 * 1024 * 1024,
			Retries:   5,
		}
	}

	return globalDownloadService
}

func (s *DownloadService) DownloadFile(url string, headers map[string]string, destination string) error {
	file, err := os.Create(destination)
	if err != nil {
		return err
	}
	defer file.Close()

	start := 0
	for {
		request, err := http.NewRequest("GET", url, nil)
		if err != nil {
			return err
		}

		for key, value := range headers {
			request.Header.Add(key, value)
		}
		request.Header.Set("Range", fmt.Sprintf("bytes=%d-%d", start, start+s.ChunkSize-1))

		var res *http.Response
		for i := 0; i < s.Retries; i++ {
			res, err = http.DefaultClient.Do(request)
			if err == nil && (res.StatusCode == http.StatusOK || res.StatusCode == http.StatusPartialContent) {
				break
			}
			time.Sleep(time.Second * time.Duration(i*i))
		}

		if err != nil {
			return err
		}

		if res.StatusCode != http.StatusOK && res.StatusCode != http.StatusPartialContent {
			return fmt.Errorf("HTTP request failed with status code %d", res.StatusCode)
		}
		_, err = io.Copy(file, res.Body)
		res.Body.Close()
		if err != nil {
			return err
		}

		if res.ContentLength < int64(s.ChunkSize) {
			break
		}

		start += s.ChunkSize
	}

	return nil
}
