package file

import (
	"context"
	"io"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

type MinioClient struct {
	config *MinioConfig
	client *minio.Client
}

func NewMinioClient(config *MinioConfig) *MinioClient {
	return &MinioClient{
		config: config,
	}
}

func (mc *MinioClient) initClient() error {
	if mc.client != nil {
		return nil
	}
	client, err := minio.New(mc.config.Endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(mc.config.AccessKey, mc.config.SecretKey, ""),
		Secure: mc.config.UseSSL,
	})
	if err != nil {
		return err
	}
	mc.client = client
	return nil
}

func (mc *MinioClient) Start() error {
	if err := mc.initClient(); err != nil {
		return err
	}
	return nil
}

func (mc *MinioClient) Stop() error {
	return nil
}

func (mc *MinioClient) GetFile(filename string) (io.ReadCloser, error) {
	obj, err := mc.client.GetObject(context.Background(), mc.config.Bucket, filename, minio.GetObjectOptions{})
	if err != nil {
		println(err.Error())
		return nil, err
	}
	return obj, nil
}

func (mc *MinioClient) PutFile(filename string, override bool, fileReader io.Reader) (int64, error) {
	info, err := mc.client.PutObject(context.Background(), mc.config.Bucket, filename, fileReader, -1, minio.PutObjectOptions{})
	if err != nil {
		return 0, err
	}
	return info.Size, nil
}

func (mc *MinioClient) DeleteFile(filename string) error {
	err := mc.client.RemoveObject(context.Background(), mc.config.Bucket, filename, minio.RemoveObjectOptions{})
	if err != nil {
		return err
	}
	return nil
}

func (mc *MinioClient) HeadFile(filename string) (*FileInfo, error) {
	info, err := mc.client.StatObject(context.Background(), mc.config.Bucket, filename, minio.GetObjectOptions{})
	if err != nil {
		return nil, err
	}
	return &FileInfo{
		Size:         info.Size,
		LastModified: info.LastModified.UnixMilli(),
		ContentType:  info.ContentType,
	}, nil
}

func (mc *MinioClient) ListFolder(folder string) ([]string, error) {
	objectCh := mc.client.ListObjects(context.Background(), mc.config.Bucket, minio.ListObjectsOptions{
		Prefix:    folder,
		Recursive: false,
	})
	var res []string
	for obj := range objectCh {
		res = append(res, obj.Key)
	}
	return res, nil
}