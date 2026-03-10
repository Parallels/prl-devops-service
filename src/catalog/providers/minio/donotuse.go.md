```golang
func (s *MinioBucketProvider) pullFileAndDecompressStable(ctx basecontext.ApiContext, path, filename, destination string) error {
	ctx.LogInfof("Pulling file %s", filename)
	startTime := time.Now()
	remoteFilePath := strings.TrimPrefix(filepath.Join(path, filename), "/")
	ns := tracker.GetProgressService()

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
						percent := float64(totalDownloaded) / float64(totalSize) * 100
						msg := tracker.
							NewJobProgressMessage(cid, msgPrefix, percent).
							WithJob(s.JobId, constants.ActionDownloadingPackFile).
							WithTransfer(totalDownloaded, totalSize).
							SetStartingTime(startTime).
							SetFilename(filename)
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
		decompressErr := compressor.DecompressTarGzStream(ctx, r, "", destination, s.JobId, constants.ActionDecompress)
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
	finalMsg := fmt.Sprintf("Finished pulling %s", filename)
	ns.NotifyProgress(cid, finalMsg, 100)
	ns.NotifyInfo(fmt.Sprintf("Finished pulling and decompressing file %s, took %v",
		filename, time.Since(startTime)))

	return nil
}
```