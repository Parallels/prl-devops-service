package minio

import (
	"context"
	"fmt"
	"io"
	"strings"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/request"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3iface"
)

// MockMinioClient implements s3iface.S3API for testing
type MockMinioClient struct {
	s3iface.S3API
	content      string
	headErr      error
	getObjectErr error
}

func (m *MockMinioClient) HeadObjectWithContext(ctx context.Context, input *s3.HeadObjectInput, opts ...request.Option) (*s3.HeadObjectOutput, error) {
	if m.headErr != nil {
		return nil, m.headErr
	}
	size := int64(len(m.content))
	return &s3.HeadObjectOutput{
		ContentLength: aws.Int64(size),
	}, nil
}

func (m *MockMinioClient) GetObjectWithContext(ctx context.Context, input *s3.GetObjectInput, opts ...request.Option) (*s3.GetObjectOutput, error) {
	if m.getObjectErr != nil {
		return nil, m.getObjectErr
	}

	rangeStr := aws.StringValue(input.Range)
	start, end := parseRange(rangeStr)
	if end >= int64(len(m.content)) {
		end = int64(len(m.content)) - 1
	}

	body := io.NopCloser(strings.NewReader(m.content[start : end+1]))
	return &s3.GetObjectOutput{
		Body: body,
	}, nil
}

func parseRange(rangeStr string) (start, end int64) {
	// Parse "bytes=start-end"
	parts := strings.Split(strings.TrimPrefix(rangeStr, "bytes="), "-")
	if len(parts) != 2 {
		return 0, 0
	}
	fmt.Sscanf(parts[0], "%d", &start)
	fmt.Sscanf(parts[1], "%d", &end)
	return
}

func TestS3ChunkDownloader(t *testing.T) {
	content := "Hello, this is a test content for S3 mock!"
	mockS3 := &MockMinioClient{content: content}
	downloader := NewMinioChunkDownloader("test-bucket", "test/path", mockS3)

	t.Run("GetFileSize", func(t *testing.T) {
		size, err := downloader.GetFileSize(context.Background(), "test.txt")
		if err != nil {
			t.Errorf("GetFileSize() error = %v", err)
			return
		}
		if size != int64(len(content)) {
			t.Errorf("GetFileSize() = %v, want %v", size, len(content))
		}
	})

	t.Run("DownloadChunk", func(t *testing.T) {
		reader, err := downloader.DownloadChunk(context.Background(), "test.txt", 0, 4)
		if err != nil {
			t.Errorf("DownloadChunk() error = %v", err)
			return
		}
		defer reader.Close()

		data, err := io.ReadAll(reader)
		if err != nil {
			t.Errorf("Failed to read chunk: %v", err)
			return
		}

		if string(data) != "Hello" {
			t.Errorf("DownloadChunk() = %v, want %v", string(data), "Hello")
		}
	})

	t.Run("Error cases", func(t *testing.T) {
		mockS3WithErrors := &MockMinioClient{
			content:      content,
			headErr:      fmt.Errorf("head error"),
			getObjectErr: fmt.Errorf("get error"),
		}
		errorDownloader := NewMinioChunkDownloader("test-bucket", "test/path", mockS3WithErrors)

		_, err := errorDownloader.GetFileSize(context.Background(), "test.txt")
		if err == nil {
			t.Error("GetFileSize() expected error, got nil")
		}

		_, err = errorDownloader.DownloadChunk(context.Background(), "test.txt", 0, 4)
		if err == nil {
			t.Error("DownloadChunk() expected error, got nil")
		}
	})
}
