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
	"sync/atomic"
	"time"

	"github.com/Parallels/prl-devops-service/basecontext"
	"github.com/Parallels/prl-devops-service/catalog/common"
	"github.com/Parallels/prl-devops-service/compressor"
	"github.com/Parallels/prl-devops-service/helpers"
	"github.com/Parallels/prl-devops-service/notifications"
	"github.com/Parallels/prl-devops-service/writers"

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

func (s *AwsS3BucketProvider) PullFileAndDecompress(ctx basecontext.ApiContext, path string, filename string, destination string) error {
	ctx.LogInfof("Pulling file %s", filename)
	startTime := time.Now()
	remoteFilePath := strings.TrimPrefix(filepath.Join(path, filename), "/")
	ns := notifications.Get()

	// Create a new session
	session, err := s.createNewSession()
	if err != nil {
		return err
	}

	svc := s3.New(session)

	headObjectOutput, err := svc.HeadObject(&s3.HeadObjectInput{
		Bucket: aws.String(s.Bucket.Name),
		Key:    aws.String(remoteFilePath),
	})
	if err != nil {
		return err
	}
	totalSize := *headObjectOutput.ContentLength
	var start int64 = 0
	var totalDownloaded int64 = 0

	// We use a larger chunk size for faster downloads
	const chunkSize int64 = 500 * 1024 * 1024 // 2GB
	msgPrefix := fmt.Sprintf("Pulling %s", filename)
	cid := helpers.GenerateId()

	// Create a pipe to feed decompression
	r, w := io.Pipe()

	// Channel to communicate downloaded chunk files
	// Buffer of 1 allows one chunk to be queued while another is being processed
	chunkFilesChan := make(chan string, 1)
	errChan := make(chan error, 1)
	ctxBck := context.Background()
	ctxChunk, cancel := context.WithTimeout(ctxBck, 5*time.Hour)
	defer cancel()

	// Downloader goroutine: downloads chunks into temp files and sends their paths over channel
	go func() {
		defer close(chunkFilesChan)
		buf := make([]byte, 2*1024*1024) // 2MB buffer for reading from S3

		for start < totalSize {
			end := start + chunkSize - 1
			if end >= totalSize {
				end = totalSize - 1
			}

			rangeHeader := fmt.Sprintf("bytes=%d-%d", start, end)
			resp, err := svc.GetObjectWithContext(ctxChunk, &s3.GetObjectInput{
				Bucket: aws.String(s.Bucket.Name),
				Key:    aws.String(remoteFilePath),
				Range:  aws.String(rangeHeader),
			})
			if err != nil {
				errChan <- err
				return
			}

			// Create a temporary file to store this chunk
			tmpFile, err := os.CreateTemp("", "s3_chunk_")
			if err != nil {
				resp.Body.Close()
				errChan <- err
				return
			}

			// Download the entire chunk into tmpFile
			var chunkDownloaded int64
			for {
				n, readErr := resp.Body.Read(buf)
				if n > 0 {
					if _, writeErr := tmpFile.Write(buf[:n]); writeErr != nil {
						tmpFile.Close()
						os.Remove(tmpFile.Name())
						resp.Body.Close()
						errChan <- writeErr
						return
					}
					chunkDownloaded += int64(n)
					atomic.AddInt64(&totalDownloaded, int64(n))
					if ns != nil && totalSize > 0 {
						percent := int((float64(totalDownloaded) / float64(totalSize)) * 100)
						msg := notifications.NewProgressNotificationMessage(cid, msgPrefix, percent).
							SetCurrentSize(totalDownloaded).
							SetTotalSize(totalSize)
						ns.Notify(msg)
					}
				}

				if readErr != nil {
					resp.Body.Close()
					if readErr == io.EOF {
						// Entire chunk downloaded
						break
					} else {
						tmpFile.Close()
						os.Remove(tmpFile.Name())
						errChan <- readErr
						return
					}
				}
			}

			// Close and rewind the temp file
			if _, err := tmpFile.Seek(0, io.SeekStart); err != nil {
				tmpFile.Close()
				os.Remove(tmpFile.Name())
				errChan <- err
				return
			}

			resp.Body.Close()
			tmpFileName := tmpFile.Name()
			tmpFile.Close()

			// Send this chunk file path to the channel
			chunkFilesChan <- tmpFileName

			// Move to the next chunk
			start = end + 1
		}

		// No more chunks
	}()

	// Streamer goroutine: reads chunk file paths, streams them to 'w', and cleans up
	go func() {
		defer w.Close()

		for chunkFileName := range chunkFilesChan {
			// Stream this chunk to w
			chunkFile, err := os.Open(chunkFileName)
			if err != nil {
				errChan <- err
				return
			}
			_, copyErr := io.Copy(w, chunkFile)
			chunkFile.Close()
			os.Remove(chunkFileName) // remove after streaming
			if copyErr != nil {
				errChan <- copyErr
				return
			}
		}

		// All chunks processed
		errChan <- nil
	}()

	// Decompress in the main goroutine
	decompressErr := compressor.DecompressFromReader(ctx, r, destination)

	// Wait for any errors from download/stream goroutines
	pipeErr := <-errChan

	if decompressErr != nil {
		return decompressErr
	}
	if pipeErr != nil {
		return pipeErr
	}

	msg := fmt.Sprintf("Pulling %s", filename)
	ns.NotifyProgress(cid, msg, 100)
	ns.NotifyInfo(fmt.Sprintf("Finished pulling and decompressing file %s, took %v", filename, time.Since(startTime)))
	return nil
}

func (s *AwsS3BucketProvider) PullFileAndDecompress2(ctx basecontext.ApiContext, path string, filename string, destination string) error {
	ctx.LogInfof("Pulling file %s", filename)
	remoteFilePath := strings.TrimPrefix(filepath.Join(path, filename), "/")
	ns := notifications.Get()
	// Create a new session
	session, err := s.createNewSession()
	if err != nil {
		return err
	}

	svc := s3.New(session)

	headObjectOutput, err := svc.HeadObject(&s3.HeadObjectInput{
		Bucket: aws.String(s.Bucket.Name),
		Key:    aws.String(remoteFilePath),
	})
	if err != nil {
		return err
	}
	totalSize := *headObjectOutput.ContentLength
	var start int64 = 0
	var totalDownloaded int64 = 0

	// Initialize decompression once if it can handle streaming from multiple chunks.
	// Otherwise, you can chain decompression by feeding chunks sequentially.
	// For a tar.gz, you need continuous data, so just feed each chunk in order.
	const chunkSize int64 = 2000 * 1024 * 1024 // 2GB
	msgPrefix := fmt.Sprintf("Pulling %s", filename)
	cid := helpers.GenerateId()

	// Create a pipe to feed decompression
	r, w := io.Pipe()

	go func() {
		defer w.Close()
		buf := make([]byte, 2*1024*1024) // buffer for reading each chunk of 2MB

		for start < totalSize {
			end := start + chunkSize - 1
			if end >= totalSize {
				end = totalSize - 1
			}

			// Create a new session for this chunk request
			chunkSession, err := s.createNewSession()
			if err != nil {
				w.CloseWithError(err)
				return
			}

			chunkSvc := s3.New(chunkSession)
			rangeHeader := fmt.Sprintf("bytes=%d-%d", start, end)
			ctxBck := context.Background()
			ctxChunk, cancel := context.WithTimeout(ctxBck, 5*time.Hour)
			defer cancel()

			resp, err := chunkSvc.GetObjectWithContext(ctxChunk, &s3.GetObjectInput{
				Bucket: aws.String(s.Bucket.Name),
				Key:    aws.String(remoteFilePath),
				Range:  aws.String(rangeHeader),
			})
			if err != nil {
				w.CloseWithError(err)
				return
			}

			// Create a temporary file to store this chunk
			tmpFile, err := os.CreateTemp("", "s3_chunk_")
			if err != nil {
				resp.Body.Close()
				w.CloseWithError(err)
				return
			}

			// Download the chunk to tmpFile
			var chunkDownloaded int64
			for {
				n, readErr := resp.Body.Read(buf)
				if n > 0 {
					if _, writeErr := tmpFile.Write(buf[:n]); writeErr != nil {
						tmpFile.Close()
						os.Remove(tmpFile.Name())
						resp.Body.Close()
						w.CloseWithError(writeErr)
						return
					}

					chunkDownloaded += int64(n)
					atomic.AddInt64(&totalDownloaded, int64(n))
					if ns != nil && totalSize > 0 {
						percent := int((float64(totalDownloaded) / float64(totalSize)) * 100)
						msg := notifications.NewProgressNotificationMessage(cid, msgPrefix, percent).
							SetCurrentSize(totalDownloaded).
							SetTotalSize(totalSize)
						ns.Notify(msg)
					}
				}

				if readErr != nil {
					resp.Body.Close()
					if readErr == io.EOF {
						// Entire chunk downloaded to tmpFile
						break
					} else {
						tmpFile.Close()
						os.Remove(tmpFile.Name())
						w.CloseWithError(readErr)
						return
					}
				}
			}

			// Finished downloading this chunk to the tmpFile
			tmpFile.Seek(0, io.SeekStart) // rewind to the start of the file
			resp.Body.Close()

			// Now stream the chunk from tmpFile to the pipe
			if _, err := io.Copy(w, tmpFile); err != nil {
				tmpFile.Close()
				os.Remove(tmpFile.Name())
				w.CloseWithError(err)
				return
			}

			tmpFile.Close()
			os.Remove(tmpFile.Name())

			// Move to the next chunk
			start = end + 1
		}
	}()

	if err := compressor.DecompressFromReader(ctx, r, destination); err != nil {
		return err
	}

	msg := fmt.Sprintf("Pulling %s", filename)
	ns.NotifyProgress(cid, msg, 100)
	ns.NotifyInfo(fmt.Sprintf("Finished pulling and decompressing file %s", filename))
	return nil
}

func (s *AwsS3BucketProvider) PullFileAndDecompress1(ctx basecontext.ApiContext, path string, filename string, destination string) error {
	ctx.LogInfof("Pulling file %s", filename)
	remoteFilePath := strings.TrimPrefix(filepath.Join(path, filename), "/")
	ns := notifications.Get()
	// Create a new session
	session, err := s.createNewSession()
	if err != nil {
		return err
	}

	svc := s3.New(session)

	headObjectOutput, err := svc.HeadObject(&s3.HeadObjectInput{
		Bucket: aws.String(s.Bucket.Name),
		Key:    aws.String(remoteFilePath),
	})
	if err != nil {
		return err
	}
	totalSize := *headObjectOutput.ContentLength
	var start int64 = 0
	var totalDownloaded int64 = 0

	// Get the object from S3 as a stream
	objOutput, err := svc.GetObject(&s3.GetObjectInput{
		Bucket: aws.String(s.Bucket.Name),
		Key:    aws.String(remoteFilePath),
	})
	if err != nil {
		return err
	}

	defer objOutput.Body.Close()

	// Initialize decompression once if it can handle streaming from multiple chunks.
	// Otherwise, you can chain decompression by feeding chunks sequentially.
	// For a tar.gz, you need continuous data, so just feed each chunk in order.
	const chunkSize int64 = 10 * 1024 * 1024
	msgPrefix := fmt.Sprintf("Pulling %s", filename)
	cid := helpers.GenerateId()

	// Create a pipe to feed decompression
	r, w := io.Pipe()
	go func() {
		defer w.Close()
		buf := make([]byte, 64*1024) // buffer for reading each chunk

		for start < totalSize {
			end := start + chunkSize - 1
			if end >= totalSize {
				end = totalSize - 1
			}

			// Create a new session for this chunk request
			chunkSession, err := s.createNewSession()
			if err != nil {
				w.CloseWithError(err)
				return
			}
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
							KeepAlive: 30 * time.Second, // send keep-alive probes more frequently
						}
						conn, err := d.DialContext(ctx, network, addr)
						return conn, err
					},
				},
			})

			chunkSvc := s3.New(chunkSession, cfg)
			rangeHeader := fmt.Sprintf("bytes=%d-%d", start, end)
			ctxBck := context.Background()
			ctx, cancel := context.WithTimeout(ctxBck, 5*time.Hour)
			defer cancel()

			resp, err := chunkSvc.GetObjectWithContext(ctx, &s3.GetObjectInput{
				Bucket: aws.String(s.Bucket.Name),
				Key:    aws.String(remoteFilePath),
				Range:  aws.String(rangeHeader),
			})
			if err != nil {
				w.CloseWithError(err)
				return
			}

			// Read this chunk in increments and update progress
			for {
				n, readErr := resp.Body.Read(buf)
				if n > 0 {
					// Write to the pipe
					if _, wErr := w.Write(buf[:n]); wErr != nil {
						// Error writing to pipe
						w.CloseWithError(wErr)
						resp.Body.Close()
						return
					}

					// Update progress
					atomic.AddInt64(&totalDownloaded, int64(n))
					if ns != nil && totalSize > 0 {
						percent := int((float64(totalDownloaded) / float64(totalSize)) * 100)
						msg := notifications.NewProgressNotificationMessage(cid, msgPrefix, percent).
							SetCurrentSize(totalDownloaded).
							SetTotalSize(totalSize)
						ns.Notify(msg)
					}
				}

				if readErr != nil {
					if readErr == io.EOF {
						// End of this chunk
						resp.Body.Close()
						break
					}
					// Some other error
					w.CloseWithError(readErr)
					resp.Body.Close()
					return
				}
			}

			start = end + 1
		}
	}()

	if err := compressor.DecompressFromReader(ctx, r, destination); err != nil {
		return err
	}

	msg := fmt.Sprintf("Pulling %s", filename)
	ns.NotifyProgress(cid, msg, 100)
	ns.NotifyInfo(fmt.Sprintf("Finished pulling and decompressing file %s", filename))
	return nil
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
