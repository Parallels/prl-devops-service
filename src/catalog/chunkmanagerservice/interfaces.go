package chunkmanagerservice

import (
	"context"
	"io"
)

// ChunkDownloader defines the interface for downloading file chunks.
// Implementations should handle the specifics of downloading chunks from different storage providers
// (e.g., S3, Azure Blob Storage, local filesystem).
type ChunkDownloader interface {
	// GetFileSize returns the total size of the remote file in bytes.
	// path should be the full path to the file in the storage system.
	GetFileSize(ctx context.Context, path string) (int64, error)

	// DownloadChunk downloads a specific byte range of the file.
	// path should be the full path to the file in the storage system.
	// start is the starting byte position (inclusive).
	// end is the ending byte position (inclusive).
	// Returns an io.ReadCloser that must be closed by the caller.
	DownloadChunk(ctx context.Context, path string, start, end int64) (io.ReadCloser, error)
}
