package services

import (
	"fmt"
	"io"
	"net/http"
	"os"
)

var globalDownloadService *DownloadService

type DownloadService struct {
}

func NewDownloadService() *DownloadService {
	if globalDownloadService == nil {
		globalDownloadService = &DownloadService{}
	}

	return globalDownloadService
}

func (s *DownloadService) DownloadFile(url string, filename string) error {
	chunkSize := 4096 * 1024 // 4MB
	// Create an HTTP client
	// Create an HTTP client
	client := &http.Client{}

	// Make a GET request to the URL
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return err
	}

	// Send the request
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	// Check if the response status code is 200 OK
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected status code %d", resp.StatusCode)
	}

	// Create a new file with the given filename
	out, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer out.Close()

	// Copy the response body to the file in chunks
	var written int64
	for {
		// Read the next chunk from the response body
		buf := make([]byte, chunkSize)
		n, err := resp.Body.Read(buf)
		if err != nil && err != io.EOF {
			return err
		}

		// Write the chunk to the output file
		if n > 0 {
			n, err = out.Write(buf[:n])
			if err != nil {
				return err
			}
			written += int64(n)
		}

		// If we've reached the end of the response body, break out of the loop
		if err == io.EOF {
			break
		}
	}

	return nil
}
