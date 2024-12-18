package services

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"time"

	"github.com/minio/minio-go"

	"github.com/sd0ric4/book-reader-backend/app/config"
)

type S3Service struct {
	minioClient *minio.Client
	bucketName  string
}

func NewS3Service(client *minio.Client, bucketName string) *S3Service {
	return &S3Service{
		minioClient: client,
		bucketName:  bucketName,
	}
}
func NewMinioClient(cfg config.S3) (*minio.Client, error) {
	minioClient, err := minio.New(cfg.Endpoint, cfg.AccessKeyID, cfg.SecretAccessKey, false)
	if err != nil {
		return nil, err
	}
	// 尝试列出存储桶以测试连接
	_, err = minioClient.ListBuckets()
	if err != nil {
		return nil, fmt.Errorf("failed to list buckets: %v", err)
	}

	return minioClient, nil
}

// GetFileURL 获取文件的预签名URL
func (s *S3Service) GetFileURL(objectName string) (string, error) {
	// 生成预签名URL，设置过期时间为1小时
	url, err := s.minioClient.PresignedGetObject(s.bucketName, objectName, time.Hour, nil)
	if err != nil {
		return "", fmt.Errorf("failed to generate presigned URL: %v", err)
	}

	return url.String(), nil
}

func CheckFileList(client *minio.Client, bucketName string) error {
	// 列出文件
	doneCh := make(chan struct{})
	defer close(doneCh)

	objectCh := client.ListObjectsV2(bucketName, "", true, doneCh)
	log.Printf("Files in %s:\n", bucketName)
	for object := range objectCh {
		if object.Err != nil {
			return object.Err
		}
		log.Println(object.Key)
	}

	return nil
}

func UploadFile(client *minio.Client, bucketName, objectName string, fileData []byte) error {
	// 上传文件
	reader := bytes.NewReader(fileData)
	_, err := client.PutObject(bucketName, objectName, reader, reader.Size(), minio.PutObjectOptions{})
	if err != nil {
		return err
	}

	log.Printf("Successfully uploaded %s to %s\n", objectName, bucketName)
	return nil
}

func DownloadFile(client *minio.Client, bucketName, objectName string) ([]byte, error) {
	// 下载文件
	object, err := client.GetObject(bucketName, objectName, minio.GetObjectOptions{})
	if err != nil {
		return nil, err
	}
	defer object.Close()

	fileData, err := io.ReadAll(object)
	if err != nil {
		return nil, err
	}

	log.Printf("Successfully downloaded %s from %s\n", objectName, bucketName)
	return fileData, nil
}
