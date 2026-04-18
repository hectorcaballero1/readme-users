package storage

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"mime/multipart"
	"path/filepath"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

type S3Service struct {
	client *s3.Client
	bucket string
	region string
}

func NewS3Service(accessKey, secretKey, region, bucket string) (*S3Service, error) {
	cfg, err := awsconfig.LoadDefaultConfig(context.Background(),
		awsconfig.WithRegion(region),
		awsconfig.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(accessKey, secretKey, "")),
	)
	if err != nil {
		return nil, err
	}
	return &S3Service{
		client: s3.NewFromConfig(cfg),
		bucket: bucket,
		region: region,
	}, nil
}

func (s *S3Service) UploadProfilePhoto(file multipart.File, header *multipart.FileHeader, userID uint) (string, error) {
	b := make([]byte, 16)
	rand.Read(b)
	key := fmt.Sprintf("profile-photos/%d/%s%s", userID, hex.EncodeToString(b), filepath.Ext(header.Filename))

	_, err := s.client.PutObject(context.Background(), &s3.PutObjectInput{
		Bucket:      aws.String(s.bucket),
		Key:         aws.String(key),
		Body:        file,
		ContentType: aws.String(header.Header.Get("Content-Type")),
	})
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("https://%s.s3.%s.amazonaws.com/%s", s.bucket, s.region, key), nil
}

func (s *S3Service) DeleteProfilePhoto(photoURL string) error {
	prefix := fmt.Sprintf("https://%s.s3.%s.amazonaws.com/", s.bucket, s.region)
	if !strings.HasPrefix(photoURL, prefix) {
		return nil // URL externa, no es nuestro bucket
	}
	key := strings.TrimPrefix(photoURL, prefix)

	_, err := s.client.DeleteObject(context.Background(), &s3.DeleteObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(key),
	})
	return err
}
