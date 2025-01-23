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
		return s.pullFileAndDecompressUnstable(ctx, path, filename, destination)
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

func (s *AwsS3BucketProvider) pullFileAndDecompressStable(ctx basecontext.ApiContext, path, filename, destination string) error {
	ctx.LogInfof("Pulling file %s", filename)
	startTime := time.Now()
	remoteFilePath := strings.TrimPrefix(filepath.Join(path, filename), "/")
	ns := notifications.Get()

	session, err := s.createNewSession()
	if err != nil {
		return err
	}
	svc := s3.New(session)

	// Getting the total size (for progress notifications)
	headObjectOutput, err := svc.HeadObject(&s3.HeadObjectInput{
		Bucket: aws.String(s.Bucket.Name),
		Key:    aws.String(remoteFilePath),
	})
	if err != nil {
		return fmt.Errorf("failed HeadObject: %w", err)
	}
	totalSize := *headObjectOutput.ContentLength

	// Setting up the chunked download and decompression
	const chunkSize int64 = 100 * 1024 * 1024 // 500MB chunk size
	var start int64 = 0
	var totalDownloaded int64 = 0
	msgPrefix := fmt.Sprintf("Pulling %s", filename)
	cid := helpers.GenerateId()

	r, w := io.Pipe()
	chunkFilesChan := make(chan string, 4)

	// Setting up an errgroup to run all goroutines and capture errors
	ctxBck := context.Background()
	ctxChunk, cancel := context.WithTimeout(ctxBck, 5*time.Hour)
	group, groupCtx := errgroup.WithContext(ctxChunk)
	defer cancel()

	// Download goroutine: download chunks into temp files and send their paths over channel
	group.Go(func() error {
		defer close(chunkFilesChan) // Signal no more chunks

		buf := make([]byte, 8*1024*1024) // 2MB buffer for reading from S3

		for start < totalSize {
			// Honor cancellations
			select {
			case <-groupCtx.Done():
				return groupCtx.Err()
			default:
			}

			end := start + chunkSize - 1
			if end >= totalSize {
				end = totalSize - 1
			}
			rangeHeader := fmt.Sprintf("bytes=%d-%d", start, end)

			// Get one chunk from S3
			resp, err := svc.GetObjectWithContext(groupCtx, &s3.GetObjectInput{
				Bucket: aws.String(s.Bucket.Name),
				Key:    aws.String(remoteFilePath),
				Range:  aws.String(rangeHeader),
			})
			if err != nil {
				return fmt.Errorf("failed to get S3 range %s: %w", rangeHeader, err)
			}

			// Create a temporary file for this chunk
			tmpFile, err := os.CreateTemp("", "s3_chunk_")
			if err != nil {
				resp.Body.Close()
				return fmt.Errorf("failed to create temp file: %w", err)
			}

			// Copy the chunk from S3 response to the temp file
			var chunkDownloaded int64
			for {
				n, readErr := resp.Body.Read(buf)
				if n > 0 {
					if _, writeErr := tmpFile.Write(buf[:n]); writeErr != nil {
						tmpFile.Close()
						os.Remove(tmpFile.Name())
						resp.Body.Close()
						return fmt.Errorf("failed to write chunk to temp file: %w", writeErr)
					}
					chunkDownloaded += int64(n)
					atomic.AddInt64(&totalDownloaded, int64(n))

					// Send a progress notification
					if ns != nil && totalSize > 0 {
						percent := float64(float64(totalDownloaded)/float64(totalSize)) * 100 * 10
						msg := notifications.
							NewProgressNotificationMessage(cid, msgPrefix, percent).
							SetCurrentSize(totalDownloaded).
							SetTotalSize(totalSize).
							SetStartingTime(startTime)
						ns.Notify(msg)
					}
				}
				if readErr != nil {
					resp.Body.Close()
					if readErr != io.EOF {
						// Real error
						tmpFile.Close()
						os.Remove(tmpFile.Name())
						return fmt.Errorf("failed reading from S3 chunk: %w", readErr)
					}
					// readErr == io.EOF => done with this chunk
					break
				}
			}

			// Reset temp file offset to the beginning, close the S3 response
			if _, err := tmpFile.Seek(0, io.SeekStart); err != nil {
				tmpFile.Close()
				os.Remove(tmpFile.Name())
				return fmt.Errorf("failed to seek in temp file: %w", err)
			}
			resp.Body.Close()

			// Send temp file name to the streamer
			tmpFileName := tmpFile.Name()
			tmpFile.Close() // Let streamer reopen it
			chunkFilesChan <- tmpFileName

			// Move to the next chunk
			start = end + 1
		}
		return nil // done successfully
	})

	// Streamer goroutine: read chunk file paths, stream them to the decompressor, and clean up
	group.Go(func() error {
		defer w.Close() // signal EOF to the decompressor

		for chunkFileName := range chunkFilesChan {
			// Honor cancellations
			select {
			case <-groupCtx.Done():
				return groupCtx.Err()
			default:
			}

			chunkFile, err := os.Open(chunkFileName)
			if err != nil {
				return fmt.Errorf("failed to open temp chunk file: %w", err)
			}

			// Stream the chunk into the decompressor pipe
			if _, copyErr := io.Copy(w, chunkFile); copyErr != nil {
				chunkFile.Close()
				os.Remove(chunkFileName)
				return fmt.Errorf("failed to copy chunk to pipe: %w", copyErr)
			}
			chunkFile.Close()
			os.Remove(chunkFileName) // clean up
		}
		return nil
	})

	// Decompressor goroutine: decompress from the pipe to the destination
	group.Go(func() error {
		// Decompress from the read-end of the pipe
		decompressErr := compressor.DecompressFromReader(ctx, r, destination)
		if decompressErr != nil {
			// If decompression fails, close the pipe with error (so streamer sees it)
			_ = w.CloseWithError(decompressErr)
			return fmt.Errorf("decompression failed: %w", decompressErr)
		}
		return nil
	})

	// Wait for all goroutines to finish and check for errors
	if err := group.Wait(); err != nil {
		return err
	}

	// Everything finished successfully send a final progress notification
	finalMsg := fmt.Sprintf("Pulling %s", filename)
	ns.NotifyProgress(cid, finalMsg, 100)
	ns.NotifyInfo(fmt.Sprintf("Finished pulling and decompressing file %s, took %v",
		filename, time.Since(startTime)))

	return nil
}

func (s *AwsS3BucketProvider) pullFileAndDecompressUnstable1(ctx basecontext.ApiContext, path, filename, destination string) error {
	ctx.LogInfof("Pulling file %s", filename)
	startTime := time.Now()
	remoteFilePath := strings.TrimPrefix(filepath.Join(path, filename), "/")

	// Notification bits
	ns := notifications.Get()
	cid := helpers.GenerateId()
	msgPrefix := fmt.Sprintf("Pulling %s", filename)

	session, err := s.createNewSession()
	if err != nil {
		return err
	}
	svc := s3.New(session)

	// Determine total size
	headResp, err := svc.HeadObjectWithContext(context.Background(), &s3.HeadObjectInput{
		Bucket: aws.String(s.Bucket.Name),
		Key:    aws.String(remoteFilePath),
	})
	if err != nil {
		return fmt.Errorf("failed HeadObject: %w", err)
	}
	totalSize := *headResp.ContentLength

	const chunkSize int64 = 500 * 1024 * 1024 // e.g. 500 MiB
	var offset int64 = 0
	index := 0

	// Identify each chunk’s start and end
	var chunks []struct {
		index int
		start int64
		end   int64
	}
	for offset < totalSize {
		end := offset + chunkSize - 1
		if end >= totalSize {
			end = totalSize - 1
		}
		chunks = append(chunks, struct {
			index int
			start int64
			end   int64
		}{
			index: index,
			start: offset,
			end:   end,
		})
		offset = end + 1
		index++
	}

	// Prepare to stream data to the decompressor
	r, w := io.Pipe()

	// For concurrency
	const maxWorkers = 5
	chunkTasks := make(chan struct {
		index int
		start int64
		end   int64
	})
	chunkResults := make(chan struct {
		index       int
		tmpFilePath string
	})

	// Track progress with an atomic
	var totalDownloaded int64

	// Build an errgroup and context
	ctxBck := context.Background()
	ctxChunk, cancel := context.WithTimeout(ctxBck, 5*time.Hour)
	group, groupCtx := errgroup.WithContext(ctxChunk)
	defer cancel()

	// Worker goroutines to fetch chunks concurrently
	for wkr := 0; wkr < maxWorkers; wkr++ {
		group.Go(func() error {
			buf := make([]byte, 2*1024*1024) // 2 MiB buffer
			for task := range chunkTasks {
				// Honour cancellations
				select {
				case <-groupCtx.Done():
					return groupCtx.Err()
				default:
				}

				rangeHeader := fmt.Sprintf("bytes=%d-%d", task.start, task.end)
				getObjResp, err := svc.GetObjectWithContext(groupCtx, &s3.GetObjectInput{
					Bucket: aws.String(s.Bucket.Name),
					Key:    aws.String(remoteFilePath),
					Range:  aws.String(rangeHeader),
				})
				if err != nil {
					return fmt.Errorf("failed to fetch S3 range %s: %w", rangeHeader, err)
				}

				// Create a temporary file for this chunk
				tmpFile, err := os.CreateTemp("", "s3_chunk_")
				if err != nil {
					getObjResp.Body.Close()
					return fmt.Errorf("failed to create temp file: %w", err)
				}

				// Copy chunk data
				var chunkDownloaded int64
				for {
					n, readErr := getObjResp.Body.Read(buf)
					if n > 0 {
						if _, writeErr := tmpFile.Write(buf[:n]); writeErr != nil {
							tmpFile.Close()
							os.Remove(tmpFile.Name())
							getObjResp.Body.Close()
							return fmt.Errorf("failed writing chunk to temp file: %w", writeErr)
						}
						atomic.AddInt64(&totalDownloaded, int64(n))
						chunkDownloaded += int64(n)

						// Progress update
						if ns != nil && totalSize > 0 {
							percent := float64(float64(atomic.LoadInt64(&totalDownloaded)) / float64(totalSize) * 100)
							msg := notifications.NewProgressNotificationMessage(cid, msgPrefix, percent).
								SetCurrentSize(atomic.LoadInt64(&totalDownloaded)).
								SetTotalSize(totalSize)
							ns.Notify(msg)
						}
					}
					if readErr != nil {
						getObjResp.Body.Close()
						if readErr != io.EOF {
							tmpFile.Close()
							os.Remove(tmpFile.Name())
							return fmt.Errorf("failed reading from S3 chunk: %w", readErr)
						}
						// EOF => done with this chunk
						break
					}
				}

				if _, err := tmpFile.Seek(0, io.SeekStart); err != nil {
					tmpFile.Close()
					os.Remove(tmpFile.Name())
					return fmt.Errorf("failed to seek in temp file: %w", err)
				}
				getObjResp.Body.Close()
				tmpName := tmpFile.Name()
				tmpFile.Close()

				// Send result
				chunkResults <- struct {
					index       int
					tmpFilePath string
				}{
					index:       task.index,
					tmpFilePath: tmpName,
				}
			}
			return nil
		})
	}

	// Feeder goroutine: queue up all chunks to be fetched
	group.Go(func() error {
		defer close(chunkTasks)
		for _, c := range chunks {
			select {
			case <-groupCtx.Done():
				return groupCtx.Err()
			case chunkTasks <- c:
			}
		}
		return nil
	})

	// Streamer goroutine: receive chunk results in *any* order, but write them
	// to the pipe in ascending order as soon as the next chunk is available.
	group.Go(func() error {
		defer w.Close()

		results := make([]*struct {
			index       int
			tmpFilePath string
		}, len(chunks))

		nextToWrite := 0
		receivedCount := 0
		totalChunks := len(chunks)

	outer:
		for {
			// If we've written all chunks, we're finished
			if nextToWrite >= totalChunks {
				break outer
			}

			select {
			case <-groupCtx.Done():
				return groupCtx.Err()

			case res, ok := <-chunkResults:
				if !ok {
					// If channel closed unexpectedly and we still have chunks to write, it’s an error
					if nextToWrite < totalChunks {
						return fmt.Errorf("chunkResults channel closed too soon")
					}
					break outer
				}
				// Store the result
				results[res.index] = &res
				receivedCount++

				// Now see if we can write any newly-available chunks in sequence
				for nextToWrite < totalChunks && results[nextToWrite] != nil {
					rres := results[nextToWrite]

					// Stream this chunk to the decompressor pipe
					f, err := os.Open(rres.tmpFilePath)
					if err != nil {
						return fmt.Errorf("failed to open temp file: %w", err)
					}
					if _, copyErr := io.Copy(w, f); copyErr != nil {
						f.Close()
						os.Remove(rres.tmpFilePath)
						return fmt.Errorf("failed to copy chunk to pipe: %w", copyErr)
					}
					f.Close()
					os.Remove(rres.tmpFilePath)

					nextToWrite++
				}
			}
		}
		return nil
	})

	// Decompressor goroutine: decompress from the pipe as soon as data becomes available
	group.Go(func() error {
		if err := compressor.DecompressFromReader(ctx, r, destination); err != nil {
			// If decompression fails, close the pipe
			_ = w.CloseWithError(err)
			return fmt.Errorf("decompression failed: %w", err)
		}
		return nil
	})

	// Wait for everything to complete
	if err := group.Wait(); err != nil {
		return err
	}

	// Final notification
	finalMsg := fmt.Sprintf("Pulling %s", filename)
	ns.NotifyProgress(cid, finalMsg, 100)
	ns.NotifyInfo(fmt.Sprintf("Finished pulling and decompressing file %s, took %v",
		filename, time.Since(startTime)))

	return nil
}

func (s *AwsS3BucketProvider) pullFileAndDecompressUnstable2(ctx basecontext.ApiContext, path, filename, destination string) error {
	ctx.LogInfof("START pullFileAndDecompressStable for %s", filename)
	startTime := time.Now()
	remoteFilePath := strings.TrimPrefix(filepath.Join(path, filename), "/")
	ns := notifications.Get()

	// Create S3 session
	session, err := s.createNewSession()
	if err != nil {
		return err
	}
	svc := s3.New(session)

	// Head request to get total size
	headObjectOutput, err := svc.HeadObject(&s3.HeadObjectInput{
		Bucket: aws.String(s.Bucket.Name),
		Key:    aws.String(remoteFilePath),
	})
	if err != nil {
		return fmt.Errorf("failed HeadObject: %w", err)
	}
	totalSize := *headObjectOutput.ContentLength
	ctx.LogInfof("Total size for %s is %d bytes", filename, totalSize)

	// Settings
	const chunkSize int64 = 500 * 1024 * 1024
	var totalDownloaded int64
	msgPrefix := fmt.Sprintf("Pulling %s", filename)
	cid := helpers.GenerateId()

	// Pipe to stream data into decompressor
	r, w := io.Pipe()

	// Channels
	chunkRangesChan := make(chan [2]int64, 8)
	chunkFilesChan := make(chan string, 4)

	// ErrGroup context to manage cancellations & timeouts
	rootCtx := context.Background()
	ctxChunk, cancel := context.WithTimeout(rootCtx, 5*time.Hour)
	group, groupCtx := errgroup.WithContext(ctxChunk)
	defer cancel()

	// 1) Produce chunk ranges
	group.Go(func() error {
		defer close(chunkRangesChan)
		ctx.LogInfof("Chunk range producer started")

		var start int64 = 0
		for start < totalSize {
			end := start + chunkSize - 1
			if end >= totalSize {
				end = totalSize - 1
			}
			ctx.LogInfof("Producing chunk range: %d-%d", start, end)
			chunkRangesChan <- [2]int64{start, end}
			start = end + 1
		}

		ctx.LogInfof("Chunk range producer done")
		return nil
	})

	// 2) Parallel worker pool to download chunks
	workerCount := 4
	var wg sync.WaitGroup
	wg.Add(workerCount)

	for i := 0; i < workerCount; i++ {
		idx := i
		group.Go(func() error {
			defer wg.Done()
			buf := make([]byte, 2*1024*1024) // 2MB read buffer

			ctx.LogInfof("Worker #%d started", idx)
			for {
				select {
				case <-groupCtx.Done():
					ctx.LogInfof("Worker #%d sees groupCtx.Done()", idx)
					return groupCtx.Err()

				case rng, ok := <-chunkRangesChan:
					if !ok {
						ctx.LogInfof("Worker #%d: no more chunks to download", idx)
						return nil
					}
					start, end := rng[0], rng[1]
					rangeHeader := fmt.Sprintf("bytes=%d-%d", start, end)
					ctx.LogInfof("Worker #%d downloading range: %s", idx, rangeHeader)

					resp, err := svc.GetObjectWithContext(groupCtx, &s3.GetObjectInput{
						Bucket: aws.String(s.Bucket.Name),
						Key:    aws.String(remoteFilePath),
						Range:  aws.String(rangeHeader),
					})
					if err != nil {
						ctx.LogInfof("Worker #%d fails GetObject: %v", idx, err)
						return fmt.Errorf("failed to get S3 range %s: %w", rangeHeader, err)
					}

					// Create temp file for chunk
					tmpFile, err := os.CreateTemp("", "s3_chunk_")
					if err != nil {
						resp.Body.Close()
						ctx.LogInfof("Worker #%d fails CreateTemp: %v", idx, err)
						return fmt.Errorf("failed to create temp file: %w", err)
					}

					var chunkDownloaded int64
					for {
						n, readErr := resp.Body.Read(buf)
						if n > 0 {
							if _, writeErr := tmpFile.Write(buf[:n]); writeErr != nil {
								tmpFile.Close()
								os.Remove(tmpFile.Name())
								resp.Body.Close()
								ctx.LogInfof("Worker #%d fails writing chunk: %v", idx, writeErr)
								return fmt.Errorf("failed to write chunk: %w", writeErr)
							}

							chunkDownloaded += int64(n)
							atomic.AddInt64(&totalDownloaded, int64(n))

							// Progress notifications
							if ns != nil && totalSize > 0 {
								percent := float64(totalDownloaded) / float64(totalSize) * 100
								msg := notifications.
									NewProgressNotificationMessage(cid, msgPrefix, percent).
									SetCurrentSize(totalDownloaded).
									SetTotalSize(totalSize)
								ns.Notify(msg)
							}
						}

						if readErr != nil {
							resp.Body.Close()
							if readErr != io.EOF {
								tmpFile.Close()
								os.Remove(tmpFile.Name())
								ctx.LogInfof("Worker #%d read error: %v", idx, readErr)
								return fmt.Errorf("failed reading from S3: %w", readErr)
							}
							// if readErr == io.EOF => chunk done
							break
						}
					}

					// Close the file and response
					if err := tmpFile.Close(); err != nil {
						os.Remove(tmpFile.Name())
						return fmt.Errorf("worker #%d: closing tmpFile: %w", idx, err)
					}
					tmpFileName := tmpFile.Name()

					// Immediately close resp.Body so we don't hold it open
					resp.Body.Close()

					// Seek back to start for the next consumer
					f, err := os.OpenFile(tmpFileName, os.O_RDWR, 0)
					if err != nil {
						os.Remove(tmpFileName)
						ctx.LogInfof("Worker #%d fails re-open temp file: %v", idx, err)
						return fmt.Errorf("failed to open temp file: %w", err)
					}
					if _, err := f.Seek(0, io.SeekStart); err != nil {
						f.Close()
						os.Remove(tmpFileName)
						ctx.LogInfof("Worker #%d fails Seek in temp file: %v", idx, err)
						return fmt.Errorf("failed to seek in temp file: %w", err)
					}
					f.Close()

					// Pass it along
					ctx.LogInfof("Worker #%d finished chunk, sending file %s to chunkFilesChan", idx, tmpFileName)
					chunkFilesChan <- tmpFileName
				}
			}
		})
	}

	// 3) Close chunkFilesChan after all workers done
	group.Go(func() error {
		wg.Wait()
		ctx.LogInfof("All workers done, closing chunkFilesChan")
		close(chunkFilesChan)
		return nil
	})

	// 4) Streamer goroutine
	group.Go(func() error {
		defer func() {
			ctx.LogInfof("Streamer goroutine done, closing pipe writer")
			_ = w.Close()
		}()

		for filePath := range chunkFilesChan {
			select {
			case <-groupCtx.Done():
				ctx.LogInfof("Streamer sees groupCtx.Done()")
				return groupCtx.Err()
			default:
			}

			ctx.LogInfof("Streamer received chunk file: %s", filePath)
			chunkFile, err := os.Open(filePath)
			if err != nil {
				ctx.LogInfof("Streamer fails to open file: %v", err)
				return fmt.Errorf("failed to open temp chunk file: %w", err)
			}

			// Copy data into the pipe
			if _, copyErr := io.Copy(w, chunkFile); copyErr != nil {
				chunkFile.Close()
				os.Remove(filePath)
				ctx.LogInfof("Streamer fails copying chunk to pipe: %v", copyErr)
				return fmt.Errorf("failed to copy chunk to pipe: %w", copyErr)
			}

			chunkFile.Close()
			os.Remove(filePath)
			ctx.LogInfof("Streamer finished chunk file: %s", filePath)
		}

		// All chunk files have been processed
		ctx.LogInfof("Streamer goroutine: no more files in chunkFilesChan")
		return nil
	})

	// 5) Decompressor goroutine
	group.Go(func() error {
		ctx.LogInfof("Decompressor goroutine started")
		decompressErr := compressor.DecompressFromReader(ctx, r, destination)
		if decompressErr != nil {
			// Close the writer with error so streamer sees it
			_ = w.CloseWithError(decompressErr)
			ctx.LogInfof("Decompressor fails: %v", decompressErr)
			return fmt.Errorf("decompression failed: %w", decompressErr)
		}
		ctx.LogInfof("Decompressor goroutine finished successfully")
		return nil
	})

	// Wait for all goroutines
	if err := group.Wait(); err != nil {
		ctx.LogInfof("pullFileAndDecompressStable error: %v", err)
		return err
	}

	// Final notification
	finalMsg := fmt.Sprintf("Pulling %s", filename)
	ns.NotifyProgress(cid, finalMsg, 100)
	ns.NotifyInfo(fmt.Sprintf("Finished pulling and decompressing file %s, took %v",
		filename, time.Since(startTime)))

	ctx.LogInfof("END pullFileAndDecompressStable for %s", filename)
	return nil
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
type chunkInfoProvider interface {
	// for example, your s.downloadChunkFile(...) signature
}

func (s *AwsS3BucketProvider) pullFileAndDecompressUnstable(ctx basecontext.ApiContext, path, filename, destination string) error {
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

func cleanupAllChunks(m map[int]chunkInfo) {
	for _, ci := range m {
		if ci.filePath != "" {
			_ = os.Remove(ci.filePath)
		}
	}
}
