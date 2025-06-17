package storage

import (
	"context"
	"fmt"
	"mime"
	"os"
	"path/filepath"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

type S3Storage struct {
	Client     *s3.Client
	Bucket     string
	URLPreview string
}

func NewS3Storage(client *s3.Client, bucket string, urlPreview string) *S3Storage {
	return &S3Storage{
		Client:     client,
		Bucket:     bucket,
		URLPreview: urlPreview,
	}
}

func (s *S3Storage) ListBuckets() ([]string, error) {
	result, err := s.Client.ListBuckets(context.TODO(), &s3.ListBucketsInput{})
	if err != nil {
		return nil, err
	}

	var bucketName []string
	if len(result.Buckets) != 0 {
		for _, bucket := range result.Buckets {
			bucketName = append(bucketName, *bucket.Name)
		}
	}

	return bucketName, nil
}

func (s *S3Storage) ReadFile(filePath string) (*string, error) {
	_, err := s.Client.GetObject(context.TODO(), &s3.GetObjectInput{
		Key:    aws.String(filePath),
		Bucket: aws.String(s.Bucket),
	})
	if err != nil {
		return nil, err
	}

	// Generate public link using path-style URL
	publicURL := fmt.Sprintf("%s:%s/%s", s.URLPreview, s.Bucket, filePath)

	return &publicURL, nil
}

func (s *S3Storage) UploadFile(filePath string) error {

	file, err := os.Open(filePath)
	if err != nil {
		return err
	}
	defer file.Close()

	// Guess the MIME type based on file extension
	contentType := mime.TypeByExtension(filepath.Ext(filePath))
	if contentType == "" {
		contentType = "application/octet-stream"
	}

	_, err = s.Client.PutObject(context.TODO(), &s3.PutObjectInput{
		Bucket:      aws.String(s.Bucket),
		Key:         aws.String(filepath.Base(filePath)),
		Body:        file,
		ContentType: aws.String(contentType),
	})

	return err
}

func (s *S3Storage) DeleteFile(filePath string) error {
	_, err := s.Client.DeleteObject(context.TODO(), &s3.DeleteObjectInput{
		Key:    aws.String(filePath), // The key is the file name in the bucket
		Bucket: aws.String(s.Bucket),
	})

	return err
}
