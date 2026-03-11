package service

import (
	"context"
	"fmt"
	"io"
	"time"

	"github.com/Azure/azure-sdk-for-go/sdk/storage/azblob"
	"github.com/Azure/azure-sdk-for-go/sdk/storage/azblob/blob"
	"github.com/chanombude/twitter-go-api/internal/config"
	"github.com/google/uuid"
	"github.com/rs/zerolog/log"
)

type StorageService interface {
	UploadFile(ctx context.Context, fileStream io.Reader, originalFilename string, contentType string) (string, error)
	DeleteFile(ctx context.Context, fileUrl string) error
}

type AzureStorageService struct {
	client        *azblob.Client
	containerName string
}

func NewAzureStorageService(config config.Config) (StorageService, error) {
	// If the connection string is just the placeholder, don't fail, but log a warning.
	if config.AzureStorageConnString == "" || config.AzureStorageConnString == "your-azure-connection-string" {
		log.Warn().Msg("AZURE_STORAGE_CONNECTION_STRING is not set or is using the default placeholder")
		return &AzureStorageService{
			client:        nil,
			containerName: config.AzureStorageContainer,
		}, nil
	}

	client, err := azblob.NewClientFromConnectionString(config.AzureStorageConnString, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create azure blob client: %w", err)
	}

	// Ensure the container exists (creates it if it doesn't)
	// We ignore the error here because the most common error is that it already exists.
	// If it fails for another reason (e.g. auth), the subsequent uploads will still catch the error.
	createCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	_, _ = client.CreateContainer(createCtx, config.AzureStorageContainer, nil)

	return &AzureStorageService{
		client:        client,
		containerName: config.AzureStorageContainer,
	}, nil
}

func (s *AzureStorageService) UploadFile(ctx context.Context, fileStream io.Reader, originalFilename string, contentType string) (string, error) {
	if s.client == nil {
		return "", fmt.Errorf("azure storage client is not configured")
	}

	filename := fmt.Sprintf("%s_%s", uuid.New().String(), originalFilename)

	opts := &azblob.UploadStreamOptions{
		HTTPHeaders: &blob.HTTPHeaders{
			BlobContentType: &contentType,
		},
	}

	_, err := s.client.UploadStream(ctx, s.containerName, filename, fileStream, opts)
	if err != nil {
		return "", fmt.Errorf("failed to upload blob to azure: %w", err)
	}

	// azure blob url format: https://<account_name>.blob.core.windows.net/<container_name>/<blob_name>
	// s.client.URL() returns the base service URL. We can construct the final URL string from that.
	blobUrl := fmt.Sprintf("%s%s/%s", s.client.ServiceClient().URL(), s.containerName, filename)

	return blobUrl, nil
}

func (s *AzureStorageService) DeleteFile(ctx context.Context, fileUrl string) error {
	if s.client == nil || fileUrl == "" {
		return nil // silently ignore if no client or no url
	}

	// Basic extraction of filename from URL (assumes filename is everything after the last '/')
	// e.g. https://account.blob.core.windows.net/container/uuid_filename.jpg -> uuid_filename.jpg
	var filename string
	for i := len(fileUrl) - 1; i >= 0; i-- {
		if fileUrl[i] == '/' {
			filename = fileUrl[i+1:]
			break
		}
	}

	if filename == "" {
		filename = fileUrl // fallback
	}

	_, err := s.client.DeleteBlob(ctx, s.containerName, filename, nil)
	if err != nil {
		// Log but don't fail the parent transaction (e.g. deleting a tweet shouldn't fail if the image is already gone)
		log.Warn().Err(err).Msg("Failed to delete blob from Azure")
	}

	return nil
}
