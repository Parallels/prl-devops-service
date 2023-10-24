package aws

import (
	"bytes"
	"os"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
)

type S3Bucket struct {
	Name      string
	Region    string
	AccessKey string
	SecretKey string
}

type AwsS3Service struct {
	Bucket S3Bucket
}

func NewAwsS3Service(bucket S3Bucket) *AwsS3Service {
	return &AwsS3Service{
		Bucket: bucket,
	}
}

// UploadFile uploads a file to an S3 bucket
func (s *AwsS3Service) UploadFile(filePath string) error {
	// Create a new session using the default region and credentials.
	var err error
	session := session.Must(session.NewSession(&aws.Config{
		Region:      &s.Bucket.Region,
		Credentials: credentials.NewStaticCredentials(s.Bucket.AccessKey, s.Bucket.SecretKey, ""),
	}))

	uploader := s3manager.NewUploader(session, func(u *s3manager.Uploader) {
		u.PartSize = 10 * 1024 * 1024 // The minimum/default allowed part size is 5MB
		u.Concurrency = 5             // default is 5
	})

	// Open the file for reading.
	file, err := os.Open(filePath)
	if err != nil {
		return err
	}
	defer file.Close()

	_, err = uploader.Upload(&s3manager.UploadInput{
		Bucket: aws.String(s.Bucket.Name),
		Key:    aws.String("my-folder/"),
		Body:   bytes.NewReader([]byte{}),
	})

	_, err = uploader.Upload(&s3manager.UploadInput{
		Bucket: aws.String(s.Bucket.Name),
		Key:    aws.String("test.dd"),
		Body:   file,
	})

	if err != nil {
		return err
	}

	return nil
}
