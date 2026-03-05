package chunkmanagerservice

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/Parallels/prl-devops-service/basecontext"
	"github.com/Parallels/prl-devops-service/compressor"
	"github.com/Parallels/prl-devops-service/constants"
	"github.com/Parallels/prl-devops-service/jobs/tracker"
	"golang.org/x/sync/errgroup"
)

// ChunkManagerService handles downloading and processing large files in chunks
type ChunkManagerService struct {
	downloader      ChunkDownloader
	workerCount     int
	maxChunksOnDisk int
	totalDownloaded int64
}

// NewChunkManagerService creates a new instance of ChunkManagerService
func NewChunkManagerService(downloader ChunkDownloader, workerCount, maxChunksOnDisk int) *ChunkManagerService {
	if workerCount <= 0 {
		workerCount = 6 // default value from original code
	}
	if maxChunksOnDisk <= 0 {
		maxChunksOnDisk = 40 // default value from original code
	}

	return &ChunkManagerService{
		downloader:      downloader,
		workerCount:     workerCount,
		maxChunksOnDisk: maxChunksOnDisk,
	}
}

// DownloadAndDecompress downloads a file in chunks and decompresses it
func (s *ChunkManagerService) DownloadAndDecompress(ctx basecontext.ApiContext, request DownloadRequest) error {
	// Reset the download counter at the start of each download
	atomic.StoreInt64(&s.totalDownloaded, 0)

	startTime := time.Now()

	var logger interface {
		LogInfof(format string, args ...interface{})
		LogErrorf(format string, args ...interface{})
		LogDebugf(format string, args ...interface{})
	} = ctx

	if request.Action == "" {
		request.Action = constants.ActionDownloadingPackFile
	}
	action := request.Action

	if request.NotificationService != nil && request.JobId != "" {
		logger = request.NotificationService.WithJob(request.JobId, action)
	}

	logger.LogInfof("[Catalog] Starting download for %s", request.Filename)

	// Use ctx.Context() for the provider calls
	totalSize, err := s.downloader.GetFileSize(ctx.Context(), filepath.Join(request.Path, request.Filename))
	if err != nil {
		return fmt.Errorf("[Catalog] failed to get file size for %s: %w", request.Filename, err)
	}
	logger.LogInfof("[Catalog] Remote file %s size: %d bytes", request.Filename, totalSize)

	// Calculate total chunks
	chunkSize := request.ChunkSize
	if chunkSize <= 0 {
		chunkSize = 100 * 1024 * 1024 // default 100MB chunks
	}
	totalChunks := (totalSize + chunkSize - 1) / chunkSize
	logger.LogInfof("[Catalog] Will download %d chunks, chunkSize=%d, workerCount=%d",
		totalChunks, chunkSize, s.workerCount)

	// Create new channels for each download
	chunkFilesChan := make(chan string, s.maxChunksOnDisk)
	defer close(chunkFilesChan)

	// Create new error group for each download
	rootCtx := context.Background()
	ctxChunk, cancel := context.WithCancel(rootCtx)
	defer cancel()
	group, groupCtx := errgroup.WithContext(ctxChunk)

	// Create new pipe for each download
	r, w := io.Pipe()
	defer r.Close()
	defer w.Close()

	// Initialize shared state
	st := &sharedState{
		chunkInfos:  make([]chunkInfo, totalChunks),
		onDisk:      0,
		nextToWrite: 0,
	}

	for i := 0; i < int(totalChunks); i++ {
		st.chunkInfos[i] = chunkInfo{
			index: i,
		}
	}

	mu := sync.Mutex{}
	cond := sync.NewCond(&mu)

	// Helper to set a global error once, then cancel
	setGlobalError := func(e error) {
		st.errOnce.Do(func() {
			st.globalErr = e
			cancel()
		})
	}

	cleanupChunks := func() {
		for i := range st.chunkInfos {
			if st.chunkInfos[i].filePath != "" {
				_ = os.Remove(st.chunkInfos[i].filePath)
			}
		}
	}

	// Start all goroutines
	s.runManagerGoroutine(logger, ctx, group, groupCtx, st, totalChunks, totalSize, startTime, request, &mu, cond, setGlobalError)
	s.runStreamerGoroutine(logger, ctx, group, st, totalChunks, w, &mu, cond, setGlobalError)
	s.runDecompressorGoroutine(logger, ctx, group, r, request, setGlobalError)

	// Wait for all goroutines
	if err := group.Wait(); err != nil {
		logger.LogInfof("[Catalog] [DownloadAndDecompress] error from goroutines: %v", err)
		cleanupChunks()
		return err
	}

	// Handle global error if set
	if st.globalErr != nil {
		logger.LogInfof("[Catalog] [DownloadAndDecompress] global error set: %v", st.globalErr)
		cleanupChunks()
		return st.globalErr
	}

	// Send final notification and cleanup
	if request.NotificationService != nil {
		prefix := request.MessagePrefix
		if request.JobId != "" && !strings.HasPrefix(prefix, "["+request.JobId+"]") {
			prefix = fmt.Sprintf("[%s] %s", request.JobId, prefix)
		}
		finalMsg := tracker.NewJobProgressMessage(
			request.CorrelationID,
			prefix,
			100,
		).
			SetCurrentSize(s.totalDownloaded).
			SetTotalSize(totalSize).
			SetStartingTime(startTime).
			SetJobId(request.JobId).
			SetCurrentAction(action)
		request.NotificationService.Notify(finalMsg)

		// Clean up notifications after we're done
		request.NotificationService.CleanupNotifications(request.CorrelationID)
	}

	logger.LogInfof("[Catalog] Successfully downloaded and decompressed %s in %v", request.Filename, time.Since(startTime))
	return nil
}

func (s *ChunkManagerService) runManagerGoroutine(
	logger interface {
		LogInfof(format string, args ...interface{})
		LogErrorf(format string, args ...interface{})
		LogDebugf(format string, args ...interface{})
	},
	ctx basecontext.ApiContext,
	group *errgroup.Group,
	groupCtx context.Context,
	st *sharedState,
	totalChunks int64,
	totalSize int64,
	startTime time.Time,
	request DownloadRequest,
	mu *sync.Mutex,
	cond *sync.Cond,
	setGlobalError func(error),
) {
	group.Go(func() error {
		defer logger.LogInfof("[Catalog] Manager goroutine exited")

		for idx := 0; idx < int(totalChunks); idx++ {
			mu.Lock()
			// Wait while we have too many workers or too many chunks on disk
			for (st.activeWorkers >= s.workerCount || st.onDisk >= s.maxChunksOnDisk) && st.globalErr == nil {
				cond.Wait()
			}
			if st.globalErr != nil || groupCtx.Err() != nil {
				mu.Unlock()
				return st.globalErr
			}

			st.activeWorkers++
			st.onDisk++
			mu.Unlock()

			go s.downloadChunk(
				logger,
				ctx,
				groupCtx,
				request,
				idx,
				st,
				totalSize,
				startTime,
				mu,
				cond,
				setGlobalError,
			)
		}
		return nil
	})
}

func (s *ChunkManagerService) runStreamerGoroutine(
	logger interface {
		LogInfof(format string, args ...interface{})
		LogErrorf(format string, args ...interface{})
		LogDebugf(format string, args ...interface{})
	},
	ctx basecontext.ApiContext,
	group *errgroup.Group,
	st *sharedState,
	totalChunks int64,
	w *io.PipeWriter,
	mu *sync.Mutex,
	cond *sync.Cond,
	setGlobalError func(error),
) {
	group.Go(func() error {
		defer func() {
			logger.LogInfof("[Catalog] Streamer goroutine done, closing pipe writer")
			_ = w.Close()
		}()

		for i := 0; i < int(totalChunks); i++ {
			mu.Lock()
			for !st.chunkInfos[i].completed && st.globalErr == nil {
				cond.Wait()
			}
			ci := st.chunkInfos[i]
			mu.Unlock()

			if ci.err != nil {
				setGlobalError(ci.err)
				return ci.err
			}

			if err := s.writeChunkToPipe(logger, ctx, ci, w, st, mu, cond); err != nil {
				setGlobalError(err)
				return err
			}
		}
		return nil
	})
}

func (s *ChunkManagerService) writeChunkToPipe(
	logger interface {
		LogInfof(format string, args ...interface{})
		LogErrorf(format string, args ...interface{})
		LogDebugf(format string, args ...interface{})
	},
	ctx basecontext.ApiContext,
	ci chunkInfo,
	w *io.PipeWriter,
	st *sharedState,
	mu *sync.Mutex,
	cond *sync.Cond,
) error {
	chunkFile, err := os.Open(ci.filePath)
	if err != nil {
		return fmt.Errorf("[Catalog] streamer failed opening chunk %d: %w", ci.index, err)
	}
	defer chunkFile.Close()

	_, copyErr := io.Copy(w, chunkFile)
	if copyErr != nil {
		return fmt.Errorf("[Catalog] streamer failed copying chunk %d: %w", ci.index, copyErr)
	}

	if rmErr := os.Remove(ci.filePath); rmErr != nil {
		logger.LogInfof("[Catalog] failed to remove chunk file %s: %v", ci.filePath, rmErr)
	}

	mu.Lock()
	st.chunkInfos[ci.index].filePath = ""
	st.onDisk--
	mu.Unlock()
	cond.Broadcast()

	return nil
}

func (s *ChunkManagerService) downloadChunk(
	logger interface {
		LogInfof(format string, args ...interface{})
		LogErrorf(format string, args ...interface{})
		LogDebugf(format string, args ...interface{})
	},
	apiCtx basecontext.ApiContext,
	ctx context.Context,
	request DownloadRequest,
	chunkIndex int,
	st *sharedState,
	totalSize int64,
	startTime time.Time,
	mu *sync.Mutex,
	cond *sync.Cond,
	setGlobalError func(error),
) {
	start := int64(chunkIndex) * request.ChunkSize
	end := start + request.ChunkSize - 1

	reader, err := s.downloader.DownloadChunk(ctx, filepath.Join(request.Path, request.Filename), start, end)
	if err != nil {
		mu.Lock()
		st.chunkInfos[chunkIndex].err = err
		st.activeWorkers--
		mu.Unlock()
		cond.Broadcast()
		setGlobalError(err)
		return
	}
	defer reader.Close()

	// Create temp file for the chunk
	tmpFile, err := os.CreateTemp("", fmt.Sprintf("chunk_%d_", chunkIndex))
	if err != nil {
		mu.Lock()
		st.chunkInfos[chunkIndex].err = err
		st.activeWorkers--
		mu.Unlock()
		cond.Broadcast()
		setGlobalError(err)
		return
	}

	// Copy downloaded data to temp file with progress tracking
	buf := make([]byte, 6*1024*1024) // 6MB buffer
	for {
		n, err := reader.Read(buf)
		if n > 0 {
			if _, writeErr := tmpFile.Write(buf[:n]); writeErr != nil {
				mu.Lock()
				st.chunkInfos[chunkIndex].err = writeErr
				st.activeWorkers--
				mu.Unlock()
				cond.Broadcast()
				setGlobalError(writeErr)
				tmpFile.Close()
				_ = os.Remove(tmpFile.Name())
				return
			}

			// Update progress
			downloaded := atomic.AddInt64(&s.totalDownloaded, int64(n))
			if request.NotificationService != nil && totalSize > 0 {
				percent := float64(downloaded) / float64(totalSize) * 100
				if percent > 100 {
					percent = 100 // Cap at 100%
				}
				prefix := request.MessagePrefix
				if request.JobId != "" && !strings.HasPrefix(prefix, "["+request.JobId+"]") {
					prefix = fmt.Sprintf("[%s] %s", request.JobId, prefix)
				}
				msg := tracker.NewJobProgressMessage(
					request.CorrelationID,
					prefix,
					percent,
				).
					SetCurrentSize(downloaded).
					SetTotalSize(totalSize).
					SetStartingTime(startTime).
					SetJobId(request.JobId).
					SetCurrentAction(request.Action)
				request.NotificationService.Notify(msg)
			}
		}
		if err == io.EOF {
			break
		}
		if err != nil {
			mu.Lock()
			st.chunkInfos[chunkIndex].err = err
			st.activeWorkers--
			mu.Unlock()
			cond.Broadcast()
			setGlobalError(err)
			tmpFile.Close()
			_ = os.Remove(tmpFile.Name())
			return
		}
	}

	tmpFile.Close()

	mu.Lock()
	st.chunkInfos[chunkIndex].filePath = tmpFile.Name()
	st.chunkInfos[chunkIndex].completed = true
	st.activeWorkers--
	mu.Unlock()
	cond.Broadcast()
}

func (s *ChunkManagerService) runDecompressorGoroutine(
	logger interface {
		LogInfof(format string, args ...interface{})
		LogErrorf(format string, args ...interface{})
		LogDebugf(format string, args ...interface{})
	},
	ctx basecontext.ApiContext,
	group *errgroup.Group,
	r *io.PipeReader,
	request DownloadRequest,
	setGlobalError func(error),
) {
	group.Go(func() error {
		defer logger.LogInfof("Decompressor goroutine exited")
		defer r.Close()

		logger.LogInfof("Starting decompression process")
		if request.JobId != "" {
			err := compressor.DecompressTarGzStream(ctx, r, "", request.Destination, request.JobId, constants.ActionDecompressingPackFile)
			if err != nil {
				setGlobalError(err)
				return fmt.Errorf("decompression failed: %w", err)
			}
		} else {
			if err := compressor.DecompressFromReader(ctx, r, request.Destination); err != nil {
				setGlobalError(err)
				return fmt.Errorf("decompression failed: %w", err)
			}
		}
		logger.LogInfof("Decompression completed successfully")
		return nil
	})
}
