package aws_s3_bucket

// func (s *AwsS3BucketProvider) PullFileAndDecompress2(ctx basecontext.ApiContext, path string, filename string, destination string) error {
// 	ctx.LogInfof("Pulling file %s", filename)
// 	startTime := time.Now()
// 	remoteFilePath := strings.TrimPrefix(filepath.Join(path, filename), "/")
// 	ns := notifications.Get()

// 	// Create a new session
// 	session, err := s.createNewSession()
// 	if err != nil {
// 		return err
// 	}

// 	svc := s3.New(session)

// 	headObjectOutput, err := svc.HeadObject(&s3.HeadObjectInput{
// 		Bucket: aws.String(s.Bucket.Name),
// 		Key:    aws.String(remoteFilePath),
// 	})
// 	if err != nil {
// 		return err
// 	}
// 	totalSize := *headObjectOutput.ContentLength
// 	var start int64 = 0
// 	var totalDownloaded int64 = 0

// 	// We use a larger chunk size for faster downloads
// 	const chunkSize int64 = 500 * 1024 * 1024 // 2GB
// 	msgPrefix := fmt.Sprintf("Pulling %s", filename)
// 	cid := helpers.GenerateId()

// 	// Create a pipe to feed decompression
// 	r, w := io.Pipe()

// 	// Channel to communicate downloaded chunk files
// 	// Buffer of 1 allows one chunk to be queued while another is being processed
// 	chunkFilesChan := make(chan string, 1)
// 	// errChan := make(chan error, 1)
// 	ctxBck := context.Background()
// 	ctxChunk, cancel := context.WithTimeout(ctxBck, 5*time.Hour)
// 	g, grpCtx := errgroup.WithContext(ctxChunk)
// 	defer cancel()

// 	// Downloader goroutine: downloads chunks into temp files and sends their paths over channel
// 	g.Go(func() error {
// 		defer close(chunkFilesChan)
// 		buf := make([]byte, 2*1024*1024) // 2MB buffer for reading from S3

// 		for start < totalSize {
// 			end := start + chunkSize - 1
// 			if end >= totalSize {
// 				end = totalSize - 1
// 			}

// 			rangeHeader := fmt.Sprintf("bytes=%d-%d", start, end)
// 			resp, err := svc.GetObjectWithContext(grpCtx, &s3.GetObjectInput{
// 				Bucket: aws.String(s.Bucket.Name),
// 				Key:    aws.String(remoteFilePath),
// 				Range:  aws.String(rangeHeader),
// 			})
// 			if err != nil {
// 				return err
// 			}

// 			// Create a temporary file to store this chunk
// 			tmpFile, err := os.CreateTemp("", "s3_chunk_")
// 			if err != nil {
// 				resp.Body.Close()
// 				return err
// 			}

// 			// Download the entire chunk into tmpFile
// 			var chunkDownloaded int64
// 			for {
// 				n, readErr := resp.Body.Read(buf)
// 				if n > 0 {
// 					if _, writeErr := tmpFile.Write(buf[:n]); writeErr != nil {
// 						tmpFile.Close()
// 						os.Remove(tmpFile.Name())
// 						resp.Body.Close()
// 						return err
// 					}
// 					chunkDownloaded += int64(n)
// 					atomic.AddInt64(&totalDownloaded, int64(n))
// 					if ns != nil && totalSize > 0 {
// 						percent := int((float64(totalDownloaded) / float64(totalSize)) * 100)
// 						msg := notifications.NewProgressNotificationMessage(cid, msgPrefix, percent).
// 							SetCurrentSize(totalDownloaded).
// 							SetTotalSize(totalSize)
// 						ns.Notify(msg)
// 					}
// 				}

// 				if readErr != nil {
// 					resp.Body.Close()
// 					if readErr == io.EOF {
// 						// Entire chunk downloaded
// 						break
// 					} else {
// 						tmpFile.Close()
// 						os.Remove(tmpFile.Name())
// 						return err
// 					}
// 				}
// 			}

// 			// Close and rewind the temp file
// 			if _, err := tmpFile.Seek(0, io.SeekStart); err != nil {
// 				tmpFile.Close()
// 				os.Remove(tmpFile.Name())
// 				return err
// 			}

// 			resp.Body.Close()
// 			tmpFileName := tmpFile.Name()
// 			tmpFile.Close()

// 			// Send this chunk file path to the channel
// 			chunkFilesChan <- tmpFileName

// 			// Move to the next chunk
// 			start = end + 1
// 		}

// 		// No more chunks
// 		return nil
// 	})

// 	// Streamer goroutine: reads chunk file paths, streams them to 'w', and cleans up
// 	g.Go(func() error {
// 		defer w.Close()

// 		for chunkFileName := range chunkFilesChan {
// 			// Stream this chunk to w
// 			chunkFile, err := os.Open(chunkFileName)
// 			if err != nil {
// 				return err
// 			}
// 			_, copyErr := io.Copy(w, chunkFile)
// 			chunkFile.Close()
// 			os.Remove(chunkFileName) // remove after streaming
// 			if copyErr != nil {
// 				return err
// 			}
// 		}

// 		// All chunks processed
// 		return nil
// 	})

// 	// Decompress in the main goroutine
// 	decompressErr := compressor.DecompressFromReader(ctx, r, destination)

// 	// Wait for downloader & streamer
// 	if err := g.Wait(); err != nil {
// 		return err
// 	}

// 	if decompressErr != nil {
// 		return decompressErr
// 	}

// 	msg := fmt.Sprintf("Pulling %s", filename)
// 	ns.NotifyProgress(cid, msg, 100)
// 	ns.NotifyInfo(fmt.Sprintf("Finished pulling and decompressing file %s, took %v", filename, time.Since(startTime)))
// 	return nil
// }

// func (s *AwsS3BucketProvider) PullFileAndDecompress3(ctx basecontext.ApiContext, path string, filename string, destination string) error {
// 	ctx.LogInfof("Pulling file %s", filename)
// 	remoteFilePath := strings.TrimPrefix(filepath.Join(path, filename), "/")
// 	ns := notifications.Get()
// 	// Create a new session
// 	session, err := s.createNewSession()
// 	if err != nil {
// 		return err
// 	}

// 	svc := s3.New(session)

// 	headObjectOutput, err := svc.HeadObject(&s3.HeadObjectInput{
// 		Bucket: aws.String(s.Bucket.Name),
// 		Key:    aws.String(remoteFilePath),
// 	})
// 	if err != nil {
// 		return err
// 	}
// 	totalSize := *headObjectOutput.ContentLength
// 	var start int64 = 0
// 	var totalDownloaded int64 = 0

// 	// Initialize decompression once if it can handle streaming from multiple chunks.
// 	// Otherwise, you can chain decompression by feeding chunks sequentially.
// 	// For a tar.gz, you need continuous data, so just feed each chunk in order.
// 	const chunkSize int64 = 2000 * 1024 * 1024 // 2GB
// 	msgPrefix := fmt.Sprintf("Pulling %s", filename)
// 	cid := helpers.GenerateId()

// 	// Create a pipe to feed decompression
// 	r, w := io.Pipe()

// 	go func() {
// 		defer w.Close()
// 		buf := make([]byte, 2*1024*1024) // buffer for reading each chunk of 2MB

// 		for start < totalSize {
// 			end := start + chunkSize - 1
// 			if end >= totalSize {
// 				end = totalSize - 1
// 			}

// 			// Create a new session for this chunk request
// 			chunkSession, err := s.createNewSession()
// 			if err != nil {
// 				w.CloseWithError(err)
// 				return
// 			}

// 			chunkSvc := s3.New(chunkSession)
// 			rangeHeader := fmt.Sprintf("bytes=%d-%d", start, end)
// 			ctxBck := context.Background()
// 			ctxChunk, cancel := context.WithTimeout(ctxBck, 5*time.Hour)
// 			defer cancel()

// 			resp, err := chunkSvc.GetObjectWithContext(ctxChunk, &s3.GetObjectInput{
// 				Bucket: aws.String(s.Bucket.Name),
// 				Key:    aws.String(remoteFilePath),
// 				Range:  aws.String(rangeHeader),
// 			})
// 			if err != nil {
// 				w.CloseWithError(err)
// 				return
// 			}

// 			// Create a temporary file to store this chunk
// 			tmpFile, err := os.CreateTemp("", "s3_chunk_")
// 			if err != nil {
// 				resp.Body.Close()
// 				w.CloseWithError(err)
// 				return
// 			}

// 			// Download the chunk to tmpFile
// 			var chunkDownloaded int64
// 			for {
// 				n, readErr := resp.Body.Read(buf)
// 				if n > 0 {
// 					if _, writeErr := tmpFile.Write(buf[:n]); writeErr != nil {
// 						tmpFile.Close()
// 						os.Remove(tmpFile.Name())
// 						resp.Body.Close()
// 						w.CloseWithError(writeErr)
// 						return
// 					}

// 					chunkDownloaded += int64(n)
// 					atomic.AddInt64(&totalDownloaded, int64(n))
// 					if ns != nil && totalSize > 0 {
// 						percent := int((float64(totalDownloaded) / float64(totalSize)) * 100)
// 						msg := notifications.NewProgressNotificationMessage(cid, msgPrefix, percent).
// 							SetCurrentSize(totalDownloaded).
// 							SetTotalSize(totalSize)
// 						ns.Notify(msg)
// 					}
// 				}

// 				if readErr != nil {
// 					resp.Body.Close()
// 					if readErr == io.EOF {
// 						// Entire chunk downloaded to tmpFile
// 						break
// 					} else {
// 						tmpFile.Close()
// 						os.Remove(tmpFile.Name())
// 						w.CloseWithError(readErr)
// 						return
// 					}
// 				}
// 			}

// 			// Finished downloading this chunk to the tmpFile
// 			tmpFile.Seek(0, io.SeekStart) // rewind to the start of the file
// 			resp.Body.Close()

// 			// Now stream the chunk from tmpFile to the pipe
// 			if _, err := io.Copy(w, tmpFile); err != nil {
// 				tmpFile.Close()
// 				os.Remove(tmpFile.Name())
// 				w.CloseWithError(err)
// 				return
// 			}

// 			tmpFile.Close()
// 			os.Remove(tmpFile.Name())

// 			// Move to the next chunk
// 			start = end + 1
// 		}
// 	}()

// 	if err := compressor.DecompressFromReader(ctx, r, destination); err != nil {
// 		return err
// 	}

// 	msg := fmt.Sprintf("Pulling %s", filename)
// 	ns.NotifyProgress(cid, msg, 100)
// 	ns.NotifyInfo(fmt.Sprintf("Finished pulling and decompressing file %s", filename))
// 	return nil
// }

// func (s *AwsS3BucketProvider) PullFileAndDecompress1(ctx basecontext.ApiContext, path string, filename string, destination string) error {
// 	ctx.LogInfof("Pulling file %s", filename)
// 	remoteFilePath := strings.TrimPrefix(filepath.Join(path, filename), "/")
// 	ns := notifications.Get()
// 	// Create a new session
// 	session, err := s.createNewSession()
// 	if err != nil {
// 		return err
// 	}

// 	svc := s3.New(session)

// 	headObjectOutput, err := svc.HeadObject(&s3.HeadObjectInput{
// 		Bucket: aws.String(s.Bucket.Name),
// 		Key:    aws.String(remoteFilePath),
// 	})
// 	if err != nil {
// 		return err
// 	}
// 	totalSize := *headObjectOutput.ContentLength
// 	var start int64 = 0
// 	var totalDownloaded int64 = 0

// 	// Get the object from S3 as a stream
// 	objOutput, err := svc.GetObject(&s3.GetObjectInput{
// 		Bucket: aws.String(s.Bucket.Name),
// 		Key:    aws.String(remoteFilePath),
// 	})
// 	if err != nil {
// 		return err
// 	}

// 	defer objOutput.Body.Close()

// 	// Initialize decompression once if it can handle streaming from multiple chunks.
// 	// Otherwise, you can chain decompression by feeding chunks sequentially.
// 	// For a tar.gz, you need continuous data, so just feed each chunk in order.
// 	const chunkSize int64 = 10 * 1024 * 1024
// 	msgPrefix := fmt.Sprintf("Pulling %s", filename)
// 	cid := helpers.GenerateId()

// 	// Create a pipe to feed decompression
// 	r, w := io.Pipe()
// 	go func() {
// 		defer w.Close()
// 		buf := make([]byte, 64*1024) // buffer for reading each chunk

// 		for start < totalSize {
// 			end := start + chunkSize - 1
// 			if end >= totalSize {
// 				end = totalSize - 1
// 			}

// 			// Create a new session for this chunk request
// 			chunkSession, err := s.createNewSession()
// 			if err != nil {
// 				w.CloseWithError(err)
// 				return
// 			}
// 			cfg := aws.NewConfig().WithHTTPClient(&http.Client{
// 				Timeout: 0,
// 				Transport: &http.Transport{
// 					IdleConnTimeout:       120 * time.Minute,
// 					TLSHandshakeTimeout:   30 * time.Second,
// 					ExpectContinueTimeout: 120 * time.Minute,
// 					ResponseHeaderTimeout: 120 * time.Minute,
// 					DialContext: func(ctx context.Context, network, addr string) (net.Conn, error) {
// 						d := net.Dialer{
// 							Timeout:   30 * time.Second,
// 							KeepAlive: 30 * time.Second, // send keep-alive probes more frequently
// 						}
// 						conn, err := d.DialContext(ctx, network, addr)
// 						return conn, err
// 					},
// 				},
// 			})

// 			chunkSvc := s3.New(chunkSession, cfg)
// 			rangeHeader := fmt.Sprintf("bytes=%d-%d", start, end)
// 			ctxBck := context.Background()
// 			ctx, cancel := context.WithTimeout(ctxBck, 5*time.Hour)
// 			defer cancel()

// 			resp, err := chunkSvc.GetObjectWithContext(ctx, &s3.GetObjectInput{
// 				Bucket: aws.String(s.Bucket.Name),
// 				Key:    aws.String(remoteFilePath),
// 				Range:  aws.String(rangeHeader),
// 			})
// 			if err != nil {
// 				w.CloseWithError(err)
// 				return
// 			}

// 			// Read this chunk in increments and update progress
// 			for {
// 				n, readErr := resp.Body.Read(buf)
// 				if n > 0 {
// 					// Write to the pipe
// 					if _, wErr := w.Write(buf[:n]); wErr != nil {
// 						// Error writing to pipe
// 						w.CloseWithError(wErr)
// 						resp.Body.Close()
// 						return
// 					}

// 					// Update progress
// 					atomic.AddInt64(&totalDownloaded, int64(n))
// 					if ns != nil && totalSize > 0 {
// 						percent := int((float64(totalDownloaded) / float64(totalSize)) * 100)
// 						msg := notifications.NewProgressNotificationMessage(cid, msgPrefix, percent).
// 							SetCurrentSize(totalDownloaded).
// 							SetTotalSize(totalSize)
// 						ns.Notify(msg)
// 					}
// 				}

// 				if readErr != nil {
// 					if readErr == io.EOF {
// 						// End of this chunk
// 						resp.Body.Close()
// 						break
// 					}
// 					// Some other error
// 					w.CloseWithError(readErr)
// 					resp.Body.Close()
// 					return
// 				}
// 			}

// 			start = end + 1
// 		}
// 	}()

// 	if err := compressor.DecompressFromReader(ctx, r, destination); err != nil {
// 		return err
// 	}

// 	msg := fmt.Sprintf("Pulling %s", filename)
// 	ns.NotifyProgress(cid, msg, 100)
// 	ns.NotifyInfo(fmt.Sprintf("Finished pulling and decompressing file %s", filename))
// 	return nil
// }
