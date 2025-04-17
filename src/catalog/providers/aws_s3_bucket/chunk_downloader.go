package aws_s3_bucket

import (
	"context"
	"fmt"
	"io"
	"path/filepath"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3iface"
)

type S3ChunkDownloader struct {
	bucket     string
	svc        s3iface.S3API
	bucketPath string
}

func NewS3ChunkDownloader(bucket, bucketPath string, s3Service s3iface.S3API) *S3ChunkDownloader {
	return &S3ChunkDownloader{
		bucket:     bucket,
		svc:        s3Service,
		bucketPath: bucketPath,
	}
}

func (d *S3ChunkDownloader) GetFileSize(ctx context.Context, path string) (int64, error) {
	remoteFilePath := d.getRemoteFilePath(path)

	headObjectOutput, err := d.svc.HeadObjectWithContext(ctx, &s3.HeadObjectInput{
		Bucket: aws.String(d.bucket),
		Key:    aws.String(remoteFilePath),
	})
	if err != nil {
		return 0, fmt.Errorf("failed HeadObject: %w", err)
	}

	return *headObjectOutput.ContentLength, nil
}

func (d *S3ChunkDownloader) DownloadChunk(ctx context.Context, path string, start, end int64) (io.ReadCloser, error) {
	remoteFilePath := d.getRemoteFilePath(path)
	rangeHeader := fmt.Sprintf("bytes=%d-%d", start, end)

	resp, err := d.svc.GetObjectWithContext(ctx, &s3.GetObjectInput{
		Bucket: aws.String(d.bucket),
		Key:    aws.String(remoteFilePath),
		Range:  aws.String(rangeHeader),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to download chunk range=%s: %w", rangeHeader, err)
	}

	return resp.Body, nil
}

func (d *S3ChunkDownloader) getRemoteFilePath(path string) string {
	fullPath := path
	if d.bucketPath != "" {
		if !strings.HasPrefix(path, d.bucketPath) {
			fullPath = filepath.Join(d.bucketPath, path)
		}
	}

	fullPath = strings.TrimPrefix(fullPath, "/")

	return fullPath
}
