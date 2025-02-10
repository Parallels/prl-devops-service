package chunkmanagerservice

import (
	"sync"

	"github.com/Parallels/prl-devops-service/notifications"
)

// DownloadRequest contains all the parameters needed for a download operation
type DownloadRequest struct {
	// Path to the file in the storage system
	Path string
	// Name of the file to download
	Filename string
	// Local path where the decompressed file should be saved
	Destination string
	// Size of each chunk in bytes. If <= 0, a default of 100MB will be used
	ChunkSize int64
	// Notification related fields
	NotificationService *notifications.NotificationService
	// Prefix for notification messages
	MessagePrefix string
	// Unique ID for correlating notifications
	CorrelationID string
}

// chunkInfo tracks the state of an individual chunk during download
type chunkInfo struct {
	index     int    // chunk's position in the file
	filePath  string // temporary file path where chunk is stored
	err       error  // any error that occurred during download
	completed bool   // whether the chunk has been downloaded successfully
}

// sharedState maintains the shared state between goroutines
type sharedState struct {
	chunkInfos    []chunkInfo // information about all chunks
	onDisk        int         // how many chunk files are currently on disk
	nextToWrite   int         // next chunk index streamer must write to the pipe
	globalErr     error       // record a single global error
	errOnce       sync.Once   // ensure we set globalErr only once
	activeWorkers int         // number of workers currently downloading
}
