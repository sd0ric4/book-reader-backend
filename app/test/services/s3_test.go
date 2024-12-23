package services_test

import (
	"os"
	"testing"

	"github.com/sd0ric4/book-reader-backend/app/config"
	"github.com/sd0ric4/book-reader-backend/app/services"
)

func TestCheckFileList(t *testing.T) {
	// Mock configuration
	config.LoadConfig("../../../config/config.yaml")

	// Initialize Minio client
	client, err := services.NewMinioClient(config.Config.S3)
	if err != nil {
		t.Fatalf("Error initializing Minio client: %s", err)
	}

	// Check file list
	err = services.CheckFileList(client, config.Config.S3.BucketName)
	if err != nil {
		t.Fatalf("Error checking file list: %s", err)
	}
}

func TestUploadFile(t *testing.T) {
	config.LoadConfig("../../../config/config.yaml")

	client, err := services.NewMinioClient(config.Config.S3)

	if err != nil {
		t.Fatalf("Error initializing Minio client: %s", err)
	}
	uploadFileName := "config.example.yaml"
	uploadFilePath := "../../../config/config.example.yaml"
	uploadFileData, err := os.ReadFile(uploadFilePath)
	if err != nil {
		t.Fatalf("Error reading file: %s", err)
	}
	err = services.UploadFile(client, config.Config.S3.BucketName, uploadFileName, uploadFileData)
	if err != nil {
		t.Fatalf("Error uploading file: %s", err)
	}
}

func TestDownloadFile(t *testing.T) {
	config.LoadConfig("../../../config/config.yaml")

	client, err := services.NewMinioClient(config.Config.S3)

	if err != nil {
		t.Fatalf("Error initializing Minio client: %s", err)
	}
	downloadFileName := "config.example.yaml"
	downloadFilePath := "../../../config/config.example.yaml"
	downloadFileData, err := services.DownloadFile(client, config.Config.S3.BucketName, downloadFileName)
	if err != nil {
		t.Fatalf("Error downloading file: %s", err)
	}
	err = os.WriteFile(downloadFilePath, downloadFileData, 0644)
	if err != nil {
		t.Fatalf("Error writing file: %s", err)
	}
}

func TestGetFileURL(t *testing.T) {
	t.Parallel()
	config.LoadConfig("../../../config/config.yaml")

	client, err := services.NewMinioClient(config.Config.S3)
	if err != nil {
		t.Fatalf("Error initializing Minio client: %s", err)
	}

	s3Service := services.NewS3Service(client, config.Config.S3.BucketName)

	fileName := "Pride and Prejudice (Austen Jane) (Z-Library).epub"
	url, err := s3Service.GetFileURL(fileName)
	if err != nil {
		t.Fatalf("Error getting file URL: %s", err)
	}
	t.Logf("File URL: %s", url)
}
