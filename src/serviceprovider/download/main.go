package download

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/Parallels/prl-devops-service/writers"
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

func (s *DownloadService) DownloadFile(url string, headers map[string]string, destination string, progressReporter *writers.ProgressReporter) error {
	file, err := os.Create(filepath.Clean(destination))
	if err != nil {
		return err
	}
	defer file.Close()
	var progressWriter io.Writer
	if progressReporter != nil {
		progressWriter = writers.NewProgressWriter(file, progressReporter.Size)
	} else {
		progressWriter = file
	}
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

		if _, err = io.Copy(progressWriter, res.Body); err != nil {
			return err
		}

		if err := res.Body.Close(); err != nil {
			return err
		}

		if res.ContentLength < int64(s.ChunkSize) {
			break
		}

		start += s.ChunkSize
	}

	return nil
}

func (s *DownloadService) DownloadFileToBytes(url string, headers map[string]string, progressReporter *writers.ProgressReporter) ([]byte, error) {
	var data []byte
	start := 0
	for {
		request, err := http.NewRequest("GET", url, nil)
		if err != nil {
			return nil, err
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
			return nil, err
		}

		if res.StatusCode != http.StatusOK && res.StatusCode != http.StatusPartialContent {
			return nil, fmt.Errorf("HTTP request failed with status code %d", res.StatusCode)
		}

		chunk, err := io.ReadAll(res.Body)
		if err != nil {
			return nil, err
		}

		data = append(data, chunk...)

		if err := res.Body.Close(); err != nil {
			return nil, err
		}

		if res.ContentLength < int64(s.ChunkSize) {
			break
		}

		start += s.ChunkSize
	}

	return data, nil
}
