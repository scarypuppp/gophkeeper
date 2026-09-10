//go:generate mockgen -destination=mocks/mock_file_storage.go -package=mocks . FileStorage
package filestorage

import (
	"context"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"io"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
)

// FileStorage описывает операции хранилища файлов, используемые FileService.
type FileStorage interface {
	GetFile(ctx context.Context, path string) (io.ReadCloser, error)
	PutFile(ctx context.Context, path string, file io.Reader) (string, error)
	DeleteFile(ctx context.Context, path string) error
}

type S3Storage struct {
	Bucket string
	client *s3.Client
}

func NewS3Client(ctx context.Context, endpoint string, accessKey string, secretKey string, bucket string) (*S3Storage, error) {
	cfg, err := config.LoadDefaultConfig(ctx,
		config.WithRegion("ru"),
		config.WithCredentialsProvider(
			credentials.NewStaticCredentialsProvider(accessKey, secretKey, ""),
		),
	)
	if err != nil {
		return nil, err
	}

	client := s3.NewFromConfig(cfg, func(o *s3.Options) {
		o.BaseEndpoint = aws.String(endpoint)
		o.UsePathStyle = true
	})

	return &S3Storage{Bucket: bucket, client: client}, nil
}

func (s *S3Storage) GetFile(ctx context.Context, path string) (io.ReadCloser, error) {
	output, err := s.client.GetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(s.Bucket),
		Key:    aws.String(path),
	})
	if err != nil {
		return nil, err
	}
	return output.Body, nil
}

// PutFile загружает файл в хранилище и возвращает SHA-256 его содержимого в hex.
// Сумму считает SDK по ходу отправки, дополнительного прохода по файлу не требуется.
func (s *S3Storage) PutFile(ctx context.Context, path string, file io.Reader) (string, error) {
	output, err := s.client.PutObject(ctx, &s3.PutObjectInput{
		Bucket:            aws.String(s.Bucket),
		Key:               aws.String(path),
		Body:              file,
		ChecksumAlgorithm: types.ChecksumAlgorithmSha256,
	})
	if err != nil {
		return "", err
	}
	if output.ChecksumSHA256 == nil {
		return "", nil
	}
	// S3 отдаёт сумму в base64, в БД храним hex.
	raw, err := base64.StdEncoding.DecodeString(*output.ChecksumSHA256)
	if err != nil {
		return "", fmt.Errorf("PutFile: decode checksum: %w", err)
	}
	return hex.EncodeToString(raw), nil
}

func (s *S3Storage) DeleteFile(ctx context.Context, path string) error {
	_, err := s.client.DeleteObject(ctx, &s3.DeleteObjectInput{
		Bucket: aws.String(s.Bucket),
		Key:    aws.String(path),
	})
	if err != nil {
		return err
	}
	return nil
}
