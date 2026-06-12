package storage

import (
	"context"
	"io"
	"time"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

func NewMinioClient(endpoint, accessKey, secretKey string, useSSL bool) (*minio.Client, error) {
	minioClient, err := minio.New(endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(accessKey, secretKey, ""),
		Secure: useSSL,
	})
	if err != nil {
		return nil, err
	}
	return minioClient, nil
}

type StorageClient interface {
	Upload(ctx context.Context, bucket, objectName string, reader io.Reader, size int64) error
	Download(ctx context.Context, bucket, objectName string) (io.ReadCloser, error)
	Delete(ctx context.Context, bucket, objectName string) error
	GetURL(ctx context.Context, bucket, objectName string, expires time.Duration) (string, error)
}

type MinioStorage struct {
	client *minio.Client
}

func NewMinioStorage(client *minio.Client) *MinioStorage {
	return &MinioStorage{client: client}
}

func (m *MinioStorage) Upload(ctx context.Context, bucket, objectName string, reader io.Reader, size int64) error {
	_, err := m.client.PutObject(ctx, bucket, objectName, reader, size, minio.PutObjectOptions{})
	return err
}

func (m *MinioStorage) Download(ctx context.Context, bucket, objectName string) (io.ReadCloser, error) {
	return m.client.GetObject(ctx, bucket, objectName, minio.GetObjectOptions{})
}

func (m *MinioStorage) Delete(ctx context.Context, bucket, objectName string) error {
	return m.client.RemoveObject(ctx, bucket, objectName, minio.RemoveObjectOptions{})
}

func (m *MinioStorage) GetURL(ctx context.Context, bucket, objectName string, expires time.Duration) (string, error) {
	url, err := m.client.PresignedGetObject(ctx, bucket, objectName, expires, nil)
	if err != nil {
		return "", err
	}
	return url.String(), nil
}
