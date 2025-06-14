package storage

import (
	"context"
	"fmt"
	"log"
	"mime"
	"os"
	"path/filepath"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

func InitStorage() *s3.Client {

	cfg := aws.Config{
		Region: os.Getenv("REGION"),
		Credentials: aws.NewCredentialsCache(
			credentials.NewStaticCredentialsProvider(os.Getenv("ACCESS_KEY"), os.Getenv("SECRET_KEY"), ""),
		),
		EndpointResolverWithOptions: aws.EndpointResolverWithOptionsFunc(
			func(service, region string, options ...any) (aws.Endpoint, error) {
				return aws.Endpoint{
					URL:           os.Getenv("ENDPOINT"),
					SigningRegion: region,
				}, nil
			},
		),
	}

	return s3.NewFromConfig(cfg, func(o *s3.Options) {
		o.UsePathStyle = true // Use path-style URLs for S3
	})
}

func ListBuckets(s3Client *s3.Client) {
	result, err := s3Client.ListBuckets(context.TODO(), &s3.ListBucketsInput{})
	if err != nil {
		log.Fatalf("Failed to list buckets: %v", err)
	}

	if len(result.Buckets) == 0 {
		log.Fatal("No buckets found!")
	} else {
		for _, bucket := range result.Buckets {
			log.Println("✅ Bucket:", *bucket.Name)
		}
	}
}

func ReadFile(s3Client *s3.Client, filePath string) string {
	_, err := s3Client.GetObject(context.TODO(), &s3.GetObjectInput{
		Key:    aws.String(filePath),
		Bucket: aws.String(os.Getenv("BUCKET_NAME")),
	})
	if err != nil {
		log.Fatalf("❌ Failed to read file: %v", err)
	}

	// Generate public link using path-style URL
	publicURL := fmt.Sprintf("%s:%s/%s", os.Getenv("URL_FILE"), os.Getenv("BUCKET_NAME"), filePath)
	log.Println("✅ File is accessible at:", publicURL)

	return publicURL
}

func UploadFile(s3Client *s3.Client, filePath string) {

	file, err := os.Open(filePath)
	if err != nil {
		log.Fatalf("Failed to open file: %v", err)
		return
	}
	defer file.Close()

	// Guess the MIME type based on file extension
	contentType := mime.TypeByExtension(filepath.Ext(filePath))
	if contentType == "" {
		contentType = "application/octet-stream"
	}

	_, err = s3Client.PutObject(context.TODO(), &s3.PutObjectInput{
		Bucket:      aws.String(os.Getenv("BUCKET_NAME")),
		Key:         aws.String(filepath.Base(filePath)),
		Body:        file,
		ContentType: aws.String(contentType),
	})
	if err != nil {
		log.Fatalf("Failed to upload file: %v", err)
		return
	}

	log.Println("✅ File uploaded successfully!")
}

func DeleteFile(s3Client *s3.Client, filePath string) {
	_, err := s3Client.DeleteObject(context.TODO(), &s3.DeleteObjectInput{
		Key:    aws.String(filePath), // The key is the file name in the bucket
		Bucket: aws.String(os.Getenv("BUCKET_NAME")),
	})
	if err != nil {
		log.Fatalf("Failed to upload file: %v", err)
	}

	log.Println("✅ File deleted successfully!")
}
