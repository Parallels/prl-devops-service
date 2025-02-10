package chunkmanagerservice

import (
	"context"
	"fmt"
	"io"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/Parallels/prl-devops-service/basecontext"
	"github.com/Parallels/prl-devops-service/notifications"
)

// MockChunkDownloader implements ChunkDownloader for testing
type MockChunkDownloader struct {
	fileSize      int64
	content       []byte
	getSizeError  error
	downloadError error
	downloadFunc  func(ctx context.Context, path string, start, end int64) (io.ReadCloser, error)
}

func (m *MockChunkDownloader) GetFileSize(ctx context.Context, path string) (int64, error) {
	if m.getSizeError != nil {
		return 0, m.getSizeError
	}
	return m.fileSize, nil
}

func (m *MockChunkDownloader) DownloadChunk(ctx context.Context, path string, start, end int64) (io.ReadCloser, error) {
	if m.downloadFunc != nil {
		return m.downloadFunc(ctx, path, start, end)
	}
	if m.downloadError != nil {
		return nil, m.downloadError
	}
	if start >= int64(len(m.content)) {
		return nil, fmt.Errorf("invalid range: start %d >= content length %d", start, len(m.content))
	}
	endPos := end + 1
	if endPos > int64(len(m.content)) {
		endPos = int64(len(m.content))
	}
	return io.NopCloser(strings.NewReader(string(m.content[start:endPos]))), nil
}

func TestChunkManagerService_DownloadAndDecompress(t *testing.T) {
	tests := []struct {
		name          string
		content       []byte
		chunkSize     int64
		workerCount   int
		maxChunks     int
		getSizeError  error
		downloadError error
		wantErr       bool
	}{
		{
			name:        "successful small file download",
			content:     []byte("Hello, World!"),
			chunkSize:   5,
			workerCount: 2,
			maxChunks:   2,
			wantErr:     false,
		},
		{
			name:         "get size error",
			content:      []byte("Hello, World!"),
			getSizeError: fmt.Errorf("failed to get size"),
			wantErr:      true,
		},
		{
			name:          "download error",
			content:       []byte("Hello, World!"),
			downloadError: fmt.Errorf("failed to download"),
			wantErr:       true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create temp dir for test
			tmpDir, err := os.MkdirTemp("", "chunk_test_*")
			if err != nil {
				t.Fatal(err)
			}
			defer os.RemoveAll(tmpDir)

			destFile := fmt.Sprintf("%s/output.txt", tmpDir)

			mock := &MockChunkDownloader{
				fileSize:      int64(len(tt.content)),
				content:       tt.content,
				getSizeError:  tt.getSizeError,
				downloadError: tt.downloadError,
			}

			service := NewChunkManagerService(mock, tt.workerCount, tt.maxChunks)

			ctx := basecontext.NewBaseContext()
			request := DownloadRequest{
				Path:                "test/path",
				Filename:            "test.txt",
				Destination:         destFile,
				ChunkSize:           tt.chunkSize,
				NotificationService: notifications.Get(),
				MessagePrefix:       "Test download",
				CorrelationID:       "test-123",
			}

			err = service.DownloadAndDecompress(ctx, request)

			if (err != nil) != tt.wantErr {
				t.Errorf("DownloadAndDecompress() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				// Verify the content was downloaded correctly
				content, err := os.ReadFile(destFile)
				if err != nil {
					t.Errorf("Failed to read output file: %v", err)
					return
				}

				if string(content) != string(tt.content) {
					t.Errorf("Content mismatch\nGot:  %s\nWant: %s", string(content), string(tt.content))
				}
			}
		})
	}
}

func TestChunkManagerService_ConcurrencyAndErrors(t *testing.T) {
	// Create a larger test file to test concurrent downloads
	content := make([]byte, 1024*1024) // 1MB
	for i := range content {
		content[i] = byte(i % 256)
	}

	tests := []struct {
		name        string
		chunkSize   int64
		workerCount int
		maxChunks   int
		addDelay    bool
		injectError bool
		wantErr     bool
	}{
		{
			name:        "concurrent download success",
			chunkSize:   1024 * 64, // 64KB chunks
			workerCount: 4,
			maxChunks:   8,
			addDelay:    true,
			wantErr:     false,
		},
		{
			name:        "error propagation",
			chunkSize:   1024 * 64,
			workerCount: 4,
			maxChunks:   8,
			injectError: true,
			wantErr:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir, err := os.MkdirTemp("", "chunk_test_*")
			if err != nil {
				t.Fatal(err)
			}
			defer os.RemoveAll(tmpDir)

			destFile := fmt.Sprintf("%s/output.txt", tmpDir)

			mock := &MockChunkDownloader{
				fileSize: int64(len(content)),
				content:  content,
			}

			if tt.addDelay {
				originalDownloadChunk := mock.DownloadChunk
				mock.downloadFunc = func(ctx context.Context, path string, start, end int64) (io.ReadCloser, error) {
					time.Sleep(10 * time.Millisecond)
					return originalDownloadChunk(ctx, path, start, end)
				}
			}

			if tt.injectError {
				mock.downloadError = fmt.Errorf("simulated error")
			}

			service := NewChunkManagerService(mock, tt.workerCount, tt.maxChunks)

			ctx := basecontext.NewBaseContext()
			request := DownloadRequest{
				Path:                "test/path",
				Filename:            "test.txt",
				Destination:         destFile,
				ChunkSize:           tt.chunkSize,
				NotificationService: notifications.Get(),
				MessagePrefix:       "Test download",
				CorrelationID:       "test-123",
			}

			err = service.DownloadAndDecompress(ctx, request)

			if (err != nil) != tt.wantErr {
				t.Errorf("DownloadAndDecompress() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				downloadedContent, err := os.ReadFile(destFile)
				if err != nil {
					t.Errorf("Failed to read output file: %v", err)
					return
				}

				if len(downloadedContent) != len(content) {
					t.Errorf("Content length mismatch\nGot:  %d\nWant: %d", len(downloadedContent), len(content))
				}
			}
		})
	}
}
