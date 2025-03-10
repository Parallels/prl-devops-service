package aws_s3_bucket

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/Parallels/prl-devops-service/basecontext"
	"github.com/Parallels/prl-devops-service/catalog/common"
	"github.com/Parallels/prl-devops-service/compressor"
	"github.com/Parallels/prl-devops-service/config"
	"github.com/Parallels/prl-devops-service/helpers"
	"github.com/Parallels/prl-devops-service/notifications"
	"github.com/Parallels/prl-devops-service/writers"
	"golang.org/x/sync/errgroup"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
)

type S3Bucket struct {
	Name                         string
	Region                       string
	AccessKey                    string
	SecretKey                    string
	SessionToken                 string
	UseEnvironmentAuthentication string
	ProgressChannel              chan int
}

const providerName = "aws-s3"

type AwsS3BucketProvider struct {
	Bucket          S3Bucket
	ProgressChannel chan int
	FileNameChannel chan string
}

func NewAwsS3Provider() *AwsS3BucketProvider {
	return &AwsS3BucketProvider{}
}

func (s *AwsS3BucketProvider) Name() string {
	return providerName
}

func (s *AwsS3BucketProvider) GetProviderMeta(ctx basecontext.ApiContext) map[string]string {
	return map[string]string{
		common.PROVIDER_VAR_NAME:         providerName,
		"bucket":                         s.Bucket.Name,
		"region":                         s.Bucket.Region,
		"access_key":                     s.Bucket.AccessKey,
		"secret_key":                     s.Bucket.SecretKey,
		"session_token":                  s.Bucket.SessionToken,
		"use_environment_authentication": s.Bucket.UseEnvironmentAuthentication,
	}
}

func (s *AwsS3BucketProvider) GetProviderRootPath(ctx basecontext.ApiContext) string {
	return "/"
}

func (s *AwsS3BucketProvider) CanStream() bool {
	return true
}

func (s *AwsS3BucketProvider) SetProgressChannel(fileNameChannel chan string, progressChannel chan int) {
	s.ProgressChannel = progressChannel
	s.FileNameChannel = fileNameChannel
}

func (s *AwsS3BucketProvider) Check(ctx basecontext.ApiContext, connection string) (bool, error) {
	parts := strings.Split(connection, ";")
	provider := ""
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if strings.Contains(strings.ToLower(part), common.PROVIDER_VAR_NAME+"=") {
			provider = strings.ReplaceAll(part, common.PROVIDER_VAR_NAME+"=", "")
		}
		if strings.Contains(strings.ToLower(part), "bucket=") {
			s.Bucket.Name = strings.ReplaceAll(part, "bucket=", "")
		}
		if strings.Contains(strings.ToLower(part), "region=") {
			s.Bucket.Region = strings.ReplaceAll(part, "region=", "")
		}
		if strings.Contains(strings.ToLower(part), "access_key=") {
			s.Bucket.AccessKey = strings.ReplaceAll(part, "access_key=", "")
		}
		if strings.Contains(strings.ToLower(part), "secret_key=") {
			s.Bucket.SecretKey = strings.ReplaceAll(part, "secret_key=", "")
		}
		if strings.Contains(strings.ToLower(part), "session_token=") {
			s.Bucket.SessionToken = strings.ReplaceAll(part, "session_token=", "")
		}
		if strings.Contains(strings.ToLower(part), "use_environment_authentication=") {
			s.Bucket.UseEnvironmentAuthentication = strings.ReplaceAll(part, "use_environment_authentication=", "")
		}
	}
	if provider == "" || !strings.EqualFold(provider, providerName) {
		ctx.LogDebugf("Provider %s is not %s, skipping", providerName, provider)
		return false, nil
	}

	if s.Bucket.Name == "" {
		return false, fmt.Errorf("missing bucket name")
	}
	if s.Bucket.Region == "" {
		return false, fmt.Errorf("missing bucket region")
	}
	if s.Bucket.AccessKey == "" {
		return false, fmt.Errorf("missing bucket access key")
	}
	if s.Bucket.SecretKey == "" {
		return false, fmt.Errorf("missing bucket secret key")
	}

	return true, nil
}

// uploadFile uploads a file to an S3 bucket
func (s *AwsS3BucketProvider) PushFile(ctx basecontext.ApiContext, rootLocalPath string, path string, filename string) error {
	ctx.LogInfof("Pushing file %s", filename)
	localFilePath := filepath.Join(rootLocalPath, filename)
	remoteFilePath := strings.TrimPrefix(filepath.Join(path, filename), "/")

	// Create a new session using the default region and credentials.
	var err error
	session, err := s.createNewSession()
	if err != nil {
		return err
	}

	uploader := s3manager.NewUploader(session, func(u *s3manager.Uploader) {
		u.PartSize = 10 * 1024 * 1024 // The minimum/default allowed part size is 5MB
		u.Concurrency = 5
	})

	// Open the file for reading.
	file, err := os.Open(filepath.Clean(localFilePath))
	if err != nil {
		return err
	}

	// Get the file info for size
	fileInfo, err := file.Stat()
	if err != nil {
		ctx.LogInfof("ERROR:", err)
		return err
	}

	defer file.Close()

	cr := writers.NewProgressFileReader(file, fileInfo.Size())
	cid := cr.CorrelationId()
	cr.SetPrefix("Reading file parts")

	_, err = uploader.Upload(&s3manager.UploadInput{
		Bucket: aws.String(s.Bucket.Name),
		Key:    aws.String(remoteFilePath),
		Body:   cr,
	})
	if err != nil {
		return err
	}

	ns := notifications.Get()
	msg := fmt.Sprintf("Pushing file %s", filename)
	ns.FinishProgress(cid, msg)
	ns.NotifyInfo(fmt.Sprintf("Finished pushing file %s", filename))
	return nil
}

func (s *AwsS3BucketProvider) PullFile(ctx basecontext.ApiContext, path string, filename string, destination string) error {
	ctx.LogInfof("Pulling file %s", filename)
	startTime := time.Now()
	if s.FileNameChannel != nil {
		s.FileNameChannel <- filename
	}
	remoteFilePath := strings.TrimPrefix(filepath.Join(path, filename), "/")
	destinationFilePath := filepath.Join(destination, filename)

	// Create a new session using the default region and credentials.
	var err error
	session, err := s.createNewSession()
	if err != nil {
		return err
	}

	headObjectOutput, err := s3.New(session).HeadObject(&s3.HeadObjectInput{
		Bucket: aws.String(s.Bucket.Name),
		Key:    aws.String(remoteFilePath),
	})
	if err != nil {
		return err
	}
	fileSize := *headObjectOutput.ContentLength

	downloader := s3manager.NewDownloader(session, func(d *s3manager.Downloader) {
		d.PartSize = 10 * 1024 * 1024 // The minimum/default allowed part size is 5MB
		d.Concurrency = 5             // default is 5
	})

	// Create a file to write the S3 Object contents to.
	f, err := os.Create(filepath.Clean(destinationFilePath))
	if err != nil {
		return err
	}

	cw := writers.NewProgressWriter(f, fileSize)
	cw.SetFilename(filename)
	cw.SetPrefix("Pulling")
	cid := cw.CorrelationId()
	// Write the contents of S3 Object to the file
	_, err = downloader.Download(cw, &s3.GetObjectInput{
		Bucket: aws.String(s.Bucket.Name),
		Key:    aws.String(remoteFilePath),
	})
	if err != nil {
		return err
	}

	ns := notifications.Get()
	msg := fmt.Sprintf("Pulling %s", filename)
	ns.NotifyProgress(cid, msg, 100)
	endTime := time.Now()
	ns.NotifyInfo(fmt.Sprintf("Finished pulling and decompressing file %s, took %s", filename, endTime.Sub(startTime)))
	return nil
}

func (s *AwsS3BucketProvider) PullFileAndDecompress(ctx basecontext.ApiContext, path, filename, destination string) error {
	cfg := config.Get()
	if cfg.IsCanaryEnabled() {
		ctx.LogInfof("\rUsing Canary version of pullFileAndDecompress")
		return s.pullFileAndDecompressStable(ctx, path, filename, destination)
	}
	return s.pullFileAndDecompressStable(ctx, path, filename, destination)
}

func (s *AwsS3BucketProvider) PullFileToMemory(ctx basecontext.ApiContext, path string, filename string) ([]byte, error) {
	ctx.LogInfof("Pulling file %s", filename)
	maxFileSize := 0.5 * 1024 * 1024 // 0.5MB

	if s.FileNameChannel != nil {
		s.FileNameChannel <- filename
	}
	remoteFilePath := strings.TrimPrefix(filepath.Join(path, filename), "/")

	// Create a new session using the default region and credentials.
	var err error
	session, err := s.createNewSession()
	if err != nil {
		return nil, err
	}

	headObjectOutput, err := s3.New(session).HeadObject(&s3.HeadObjectInput{
		Bucket: aws.String(s.Bucket.Name),
		Key:    aws.String(remoteFilePath),
	})
	if err != nil {
		return nil, err
	}
	fileSize := *headObjectOutput.ContentLength

	if fileSize > int64(maxFileSize) {
		return nil, fmt.Errorf("file size is too large to pull to memory")
	}

	downloader := s3manager.NewDownloader(session, func(d *s3manager.Downloader) {
		d.PartSize = 10 * 1024 * 1024 // The minimum/default allowed part size is 5MB
		d.Concurrency = 5             // default is 5
	})

	cw := writers.NewByteSliceWriterAt(fileSize)

	// Write the contents of S3 Object to the file
	_, err = downloader.Download(cw, &s3.GetObjectInput{
		Bucket: aws.String(s.Bucket.Name),
		Key:    aws.String(remoteFilePath),
	})
	if err != nil {
		return nil, err
	}

	return cw.Bytes(), nil
}

func (s *AwsS3BucketProvider) DeleteFile(ctx basecontext.ApiContext, path string, fileName string) error {
	remoteFilePath := strings.TrimPrefix(filepath.Join(path, fileName), "/")

	// Create a new AWS session
	session, err := s.createNewSession()
	if err != nil {
		return err
	}

	// Create a new S3 client
	svc := s3.New(session)

	_, err = svc.DeleteObject(&s3.DeleteObjectInput{
		Bucket: aws.String(s.Bucket.Name),
		Key:    aws.String(remoteFilePath),
	})
	if err != nil {
		return err
	}

	return nil
}

func (s *AwsS3BucketProvider) FileChecksum(ctx basecontext.ApiContext, path string, fileName string) (string, error) {
	// Create a new AWS session
	session, err := s.createNewSession()
	if err != nil {
		return "", err
	}

	// Create a new S3 client
	svc := s3.New(session)

	fullPath := filepath.Join(path, fileName)
	resp, err := svc.HeadObject(&s3.HeadObjectInput{
		Bucket: aws.String(s.Bucket.Name),
		Key:    aws.String(fullPath),
	})
	if err != nil {
		return "", err
	}

	// The ETag is enclosed in double quotes, so we remove them
	checksum := strings.Trim(*resp.ETag, "\"")

	return checksum, nil
}

func (s *AwsS3BucketProvider) FileExists(ctx basecontext.ApiContext, path string, fileName string) (bool, error) {
	fullPath := filepath.Join(path, fileName)
	// Create a new AWS session
	session, err := s.createNewSession()
	if err != nil {
		return false, err
	}

	// Create a new S3 client
	svc := s3.New(session)

	// Check if the file exists
	_, err = svc.HeadObject(&s3.HeadObjectInput{
		Bucket: aws.String(s.Bucket.Name),
		Key:    aws.String(fullPath),
	})
	if err != nil {
		if strings.Contains(err.Error(), "NotFound") {
			return false, nil
		}
		return false, err
	}

	return true, nil
}

func (s *AwsS3BucketProvider) CreateFolder(ctx basecontext.ApiContext, folderPath string, folderName string) error {
	fullPath := filepath.Join(folderPath, folderName)
	// Create a new session using the default region and credentials.
	var err error
	// Create a new AWS session
	session, err := s.createNewSession()
	if err != nil {
		return err
	}

	uploader := s3manager.NewUploader(session, func(u *s3manager.Uploader) {
		u.PartSize = 10 * 1024 * 1024 // The minimum/default allowed part size is 5MB
		u.Concurrency = 5             // default is 5
	})

	if !strings.HasSuffix(fullPath, "/") {
		fullPath = fullPath + "/"
	}

	exists, err := s.FolderExists(ctx, folderPath, folderName)
	if err != nil {
		return err
	}
	if exists {
		return nil
	}

	_, err = uploader.Upload(&s3manager.UploadInput{
		Bucket: aws.String(s.Bucket.Name),
		Key:    aws.String(fullPath),
		Body:   bytes.NewReader([]byte{}),
	})
	if err != nil {
		return err
	}

	return nil
}

func (s *AwsS3BucketProvider) DeleteFolder(ctx basecontext.ApiContext, folderPath string, folderName string) error {
	fullPath := filepath.Join(folderPath, folderName)
	fullPath = strings.TrimPrefix(fullPath, "/")
	// Create a new AWS session
	session, err := s.createNewSession()
	if err != nil {
		return err
	}

	// Create a new S3 client
	svc := s3.New(session)

	resp, err := svc.ListObjectsV2(&s3.ListObjectsV2Input{
		Bucket: aws.String(s.Bucket.Name),
		Prefix: aws.String(fullPath),
	})
	if err != nil {
		return err
	}

	for _, obj := range resp.Contents {
		_, err := svc.DeleteObject(&s3.DeleteObjectInput{
			Bucket: aws.String(s.Bucket.Name),
			Key:    obj.Key,
		})
		if err != nil {
			return err
		}
	}

	return nil
}

func (s *AwsS3BucketProvider) FolderExists(ctx basecontext.ApiContext, folderPath string, folderName string) (bool, error) {
	fullPath := filepath.Join(folderPath, folderName)
	fullPath = strings.TrimPrefix(fullPath, "/")

	// Create a new AWS session
	session, err := s.createNewSession()
	if err != nil {
		return false, err
	}

	// Create a new S3 client
	svc := s3.New(session)

	// Check if the folder exists
	resp, err := svc.ListObjectsV2(&s3.ListObjectsV2Input{
		Bucket:    aws.String(s.Bucket.Name),
		Prefix:    aws.String(fullPath),
		Delimiter: aws.String("/"),
		MaxKeys:   aws.Int64(1),
	})
	if err != nil {
		return false, err
	}

	// If the folder exists, return true
	if len(resp.CommonPrefixes) > 0 {
		return true, nil
	}

	// If the folder does not exist, return false
	return false, nil
}

func (s *AwsS3BucketProvider) FileSize(ctx basecontext.ApiContext, path string, filename string) (int64, error) {
	ctx.LogInfof("Checking file %s size", filename)
	remoteFilePath := strings.TrimPrefix(filepath.Join(path, filename), "/")

	// Create a new session using the default region and credentials.
	var err error
	session, err := s.createNewSession()
	if err != nil {
		return -1, err
	}

	headObjectOutput, err := s3.New(session).HeadObject(&s3.HeadObjectInput{
		Bucket: aws.String(s.Bucket.Name),
		Key:    aws.String(remoteFilePath),
	})
	if err != nil {
		return -1, err
	}
	fileSize := *headObjectOutput.ContentLength

	return fileSize, nil
}

func (s *AwsS3BucketProvider) createNewSession() (*session.Session, error) {
	// Create a new session using the default region and credentials.
	var creds *credentials.Credentials
	var err error

	if s.Bucket.UseEnvironmentAuthentication == "true" {
		creds = credentials.NewEnvCredentials()
	} else {
		creds = credentials.NewStaticCredentials(s.Bucket.AccessKey, s.Bucket.SecretKey, s.Bucket.SessionToken)
	}

	cfg := s.generateNewCfg()
	cfg.Credentials = creds
	cfg.MaxRetries = aws.Int(10)
	cfg.Region = &s.Bucket.Region

	session := session.Must(session.NewSession(cfg))

	return session, err
}

func (s *AwsS3BucketProvider) generateNewCfg() *aws.Config {
	cfg := aws.NewConfig().WithHTTPClient(&http.Client{
		Timeout: 0,
		Transport: &http.Transport{
			IdleConnTimeout:       120 * time.Minute,
			TLSHandshakeTimeout:   30 * time.Second,
			ExpectContinueTimeout: 120 * time.Minute,
			ResponseHeaderTimeout: 120 * time.Minute,
			DialContext: func(ctx context.Context, network, addr string) (net.Conn, error) {
				d := net.Dialer{
					Timeout:   30 * time.Second,
					KeepAlive: 30 * time.Second,
				}
				conn, err := d.DialContext(ctx, network, addr)
				return conn, err
			},
		},
	})

	return cfg
}

// We have a small shared state, protected by a mutex + condition variable
type chunkInfo struct {
	index     int
	filePath  string
	err       error
	completed bool
}

type sharedState struct {
	chunkInfos    []chunkInfo
	onDisk        int       // how many chunk files are currently on disk
	nextToWrite   int       // next chunk index streamer must write to the pipe
	globalErr     error     // record a single global error
	errOnce       sync.Once // ensure we set globalErr only once
	activeWorkers int
}

func (s *AwsS3BucketProvider) pullFileAndDecompressStable(ctx basecontext.ApiContext, path, filename, destination string) error {
	ctx.LogInfof("Starting pullFileAndDecompressOrdered for %s", filename)
	startTime := time.Now()
	// firstChunkSize := int64(50 * 1024 * 1024)
	chunkSize := int64(100 * 1024 * 1024)
	workerCount := 6
	keepChunkOnDiskCount := 40
	var totalDownloaded int64
	ns := notifications.Get()
	msgPrefix := fmt.Sprintf("Pulling %s", filename)
	cid := helpers.GenerateId()

	// 1) Create S3 session
	session, err := s.createNewSession()
	if err != nil {
		return fmt.Errorf("failed to create S3 session: %w", err)
	}
	svc := s3.New(session)

	// 2) Get file size to calculate chunk count
	remoteFilePath := strings.TrimPrefix(filepath.Join(path, filename), "/")
	headObjectOutput, err := svc.HeadObject(&s3.HeadObjectInput{
		Bucket: aws.String(s.Bucket.Name),
		Key:    aws.String(remoteFilePath),
	})
	if err != nil {
		return fmt.Errorf("failed HeadObject: %w", err)
	}
	totalSize := *headObjectOutput.ContentLength
	ctx.LogInfof("Remote file %s size: %d bytes", filename, totalSize)

	// 3) Compute total number of chunks
	totalChunks := (totalSize + chunkSize - 1) / chunkSize
	ctx.LogInfof("Will download %d chunks, chunkSize=%d, workerCount=%d",
		totalChunks, chunkSize, workerCount)

	// 4) Prepare an io.Pipe for streaming to the decompressor
	r, w := io.Pipe()

	// We'll use an errgroup to manage concurrency and a context to cancel on error
	rootCtx := context.Background()
	ctxChunk, cancel := context.WithCancel(rootCtx)
	group, groupCtx := errgroup.WithContext(ctxChunk)

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

	// mu.Lock()
	// st.onDisk++
	// mu.Unlock()

	// if e := s.downloadChunkFile(
	// 	groupCtx,
	// 	svc,
	// 	ns,
	// 	startTime,
	// 	remoteFilePath,
	// 	0,
	// 	int(totalChunks),
	// 	totalSize,
	// 	&totalDownloaded,
	// 	firstChunkSize,
	// 	msgPrefix,
	// 	cid,
	// 	st,
	// 	&mu,
	// 	cond,
	// ); e != nil {
	// 	setGlobalError(e)
	// 	cleanupChunks()
	// 	return e
	// }

	// // Mark chunk 0 as completed
	// mu.Lock()
	// // Suppose the local file path was set inside downloadChunkFile:
	// // st.chunkInfos[0].filePath = "/some/tmp/path_for_chunk0"
	// // st.chunkInfos[0].err = nil
	// st.chunkInfos[0].completed = true
	// mu.Unlock()
	// cond.Broadcast() // notify the streamer chunk 0 is ready

	// 5) Manager goroutine — schedules chunks for download (0..totalChunks-1)
	group.Go(func() error {
		defer ctx.LogInfof("Manager goroutine exited")

		for idx := 0; idx < int(totalChunks); idx++ {
			mu.Lock()
			// Wait while we have too many workers or too many chunks on disk
			for (st.activeWorkers >= workerCount || st.onDisk >= keepChunkOnDiskCount) && st.globalErr == nil {
				cond.Wait()
			}
			if st.globalErr != nil || groupCtx.Err() != nil {
				// If there's a global error or the group context is cancelled, bail out
				mu.Unlock()
				return st.globalErr
			}

			// We can start a new download worker now:
			st.activeWorkers++
			st.onDisk++
			mu.Unlock()

			// Start a goroutine to download chunk idx
			go func(chunkIndex int) {
				localErr := s.downloadChunkFile(
					groupCtx,
					svc,
					ns,
					startTime,
					remoteFilePath,
					chunkIndex,
					int(totalChunks),
					totalSize,
					&totalDownloaded,
					chunkSize,
					msgPrefix,
					cid,
					st,
					&mu,
					cond,
				)
				mu.Lock()
				if localErr != nil {
					st.chunkInfos[chunkIndex].err = localErr
				} else {
					// st.chunkInfos[chunkIndex].filePath was set inside downloadChunkFile
					st.chunkInfos[chunkIndex].completed = true
				}
				st.activeWorkers--
				mu.Unlock()
				cond.Broadcast()

				if localErr != nil {
					setGlobalError(localErr)
				}
			}(idx)
		}
		return nil
	})

	// 6) Streamer goroutine
	// Reads chunks in ascending order [0..totalChunks-1] from local temp files,
	// writes them into the pipe 'w' for decompression.
	group.Go(func() error {
		defer func() {
			ctx.LogInfof("Streamer goroutine done, closing pipe writer")
			_ = w.Close()
		}()

		for i := 0; i < int(totalChunks); i++ {
			mu.Lock()
			// Wait for chunk i to be completed or for an error
			for !st.chunkInfos[i].completed && st.globalErr == nil {
				cond.Wait()
			}
			ci := st.chunkInfos[i]
			mu.Unlock()

			if ci.err != nil {
				setGlobalError(ci.err)
				return ci.err
			}
			// At this point, chunk i is definitely downloaded
			chunkFile, err := os.Open(ci.filePath)
			if err != nil {
				setGlobalError(err)
				return fmt.Errorf("streamer failed opening chunk %d: %w", i, err)
			}

			// Copy chunk i to the pipe => piped into decompressor
			_, copyErr := io.Copy(w, chunkFile)
			chunkFile.Close()
			if copyErr != nil {
				setGlobalError(copyErr)
				return fmt.Errorf("streamer failed copying chunk %d: %w", i, copyErr)
			}

			// Remove the chunk from disk
			if rmErr := os.Remove(ci.filePath); rmErr != nil {
				ctx.LogInfof("failed to remove chunk file %s: %v", ci.filePath, rmErr)
			}

			mu.Lock()
			st.chunkInfos[i].filePath = ""
			st.onDisk--
			mu.Unlock()
			cond.Broadcast()
		}
		return nil
	})

	// 7) Decompressor goroutine
	// Reads from 'r' as soon as the streamer writes chunk 0 data
	group.Go(func() error {
		defer ctx.LogInfof("Decompressor goroutine exited")

		decompressErr := compressor.DecompressFromReader(ctx, r, destination)
		if decompressErr != nil {
			// if decompression fails, close the writer with error
			_ = w.CloseWithError(decompressErr)
			return fmt.Errorf("decompression failed: %w", decompressErr)
		}
		return nil
	})

	// 8) Wait for all goroutines
	if err := group.Wait(); err != nil {
		// Something failed
		ctx.LogInfof("pullFileAndDecompressOrdered: error from goroutines: %v", err)
		cleanupChunks()
		return err
	}
	// If manager or workers set a globalErr, handle it
	if st.globalErr != nil {
		ctx.LogInfof("pullFileAndDecompressOrdered: global error set: %v", st.globalErr)
		cleanupChunks()
		return st.globalErr
	}

	ctx.LogInfof("Finished pulling %s in %v", filename, time.Since(startTime))
	return nil
}

func (s *AwsS3BucketProvider) downloadChunkFile(
	ctx context.Context,
	svc *s3.S3,
	ns *notifications.NotificationService,
	startingTime time.Time,
	remoteFilePath string,
	chunkIndex int,
	totalChunks int,
	totalSize int64,
	totalDownloaded *int64,
	chunkSize int64,
	msgPrefix string,
	cid string,
	st *sharedState,
	mu *sync.Mutex,
	cond *sync.Cond,
) error {
	// Calculate [start..end]
	start := int64(chunkIndex) * chunkSize
	end := start + chunkSize - 1
	rangeHeader := fmt.Sprintf("bytes=%d-%d", start, end)

	// Fetch object chunk
	resp, err := svc.GetObjectWithContext(ctx, &s3.GetObjectInput{
		Bucket: aws.String(s.Bucket.Name),
		Key:    aws.String(remoteFilePath),
		Range:  aws.String(rangeHeader),
	})
	if err != nil {
		mu.Lock()
		st.chunkInfos[chunkIndex].filePath = ""
		st.chunkInfos[chunkIndex].err = err
		cond.Broadcast()
		mu.Unlock()
		return fmt.Errorf("worker failed chunk %d range=%s: %w", chunkIndex, rangeHeader, err)
	}
	defer resp.Body.Close()

	// Create temp file
	tmpFile, err := os.CreateTemp("", fmt.Sprintf("chunk_%d_", chunkIndex))
	if err != nil {
		mu.Lock()
		st.chunkInfos[chunkIndex].filePath = ""
		st.chunkInfos[chunkIndex].err = err
		cond.Broadcast()
		mu.Unlock()
		return fmt.Errorf("cannot create temp file for chunk %d: %w", chunkIndex, err)
	}

	// Copy from S3 to temp file
	buf := make([]byte, 6*1024*1024) // 2MB buffer
	var chunkDownloaded int64
	for {
		n, readErr := resp.Body.Read(buf)
		if n > 0 {
			if _, writeErr := tmpFile.Write(buf[:n]); writeErr != nil {
				tmpFile.Close()
				mu.Lock()
				st.chunkInfos[chunkIndex].filePath = ""
				st.chunkInfos[chunkIndex].err = writeErr
				cond.Broadcast()
				mu.Unlock()
				return fmt.Errorf("write error chunk %d: %w", chunkIndex, writeErr)
			}

			chunkDownloaded += int64(n)
			atomic.AddInt64(totalDownloaded, int64(n))

			// Progress notifications
			if ns != nil && totalSize > 0 {
				percent := float64(*totalDownloaded) / float64(totalSize) * 100
				msg := notifications.
					NewProgressNotificationMessage(cid, msgPrefix, percent).
					SetCurrentSize(*totalDownloaded).
					SetTotalSize(totalSize).
					SetStartingTime(startingTime)
				ns.Notify(msg)
			}
		}
		if readErr == io.EOF {
			break
		}
		if readErr != nil {
			tmpFile.Close()
			mu.Lock()
			st.chunkInfos[chunkIndex].filePath = ""
			st.chunkInfos[chunkIndex].err = readErr
			cond.Broadcast()
			mu.Unlock()
			return fmt.Errorf("read error chunk %d: %w", chunkIndex, readErr)
		}
	}

	// Seek back to start (not strictly necessary since the streamer is copying from the file anyway,
	// but done in case the streamer logic starts reading it from the beginning).
	if _, err := tmpFile.Seek(0, io.SeekStart); err != nil {
		tmpFile.Close()
		mu.Lock()
		st.chunkInfos[chunkIndex].filePath = ""
		st.chunkInfos[chunkIndex].err = err
		cond.Broadcast()
		mu.Unlock()
		return fmt.Errorf("failed to seek in chunk %d: %w", chunkIndex, err)
	}

	filePath := tmpFile.Name()
	_ = tmpFile.Close()

	// Store the chunk file location in shared map
	mu.Lock()
	st.chunkInfos[chunkIndex].filePath = filePath
	st.chunkInfos[chunkIndex].err = nil
	cond.Broadcast() // let the streamer know chunk is ready
	mu.Unlock()

	return nil
}
