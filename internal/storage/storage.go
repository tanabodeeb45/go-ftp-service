package storage

import (
	"bytes"
	"context"
	"encoding/base64"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"

	"cloud.google.com/go/storage"
	"google.golang.org/api/option"
)

func UploadCSV(fileContent *bytes.Buffer, filename string) (string, error) {
	key := os.Getenv("GCP_SERVICE_ACCOUNT")
	if key == "" {
		return "", fmt.Errorf("GCP_SERVICE_ACCOUNT environment variable not set")
	}

	bucket := os.Getenv("GCP_STORAGE_BUCKET")
	if bucket == "" {
		return "", fmt.Errorf("GCP_STORAGE_BUCKET environment variable not set")
	}

	decodedKey, err := base64.StdEncoding.DecodeString(key)
	if err != nil {
		return "", fmt.Errorf("error decoding service account key: %w", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	client, err := storage.NewClient(ctx, option.WithCredentialsJSON(decodedKey))
	if err != nil {
		return "", fmt.Errorf("error creating GCS client: %w", err)
	}
	defer client.Close()

	objectPath := filepath.Join("reconcile", filename)

	writer := client.Bucket(bucket).Object(objectPath).NewWriter(ctx)
	writer.ContentType = "text/csv"

	if _, err := io.Copy(writer, fileContent); err != nil {
		return "", fmt.Errorf("error uploading file to GCS: %w", err)
	}

	if err := writer.Close(); err != nil {
		return "", fmt.Errorf("error finalizing GCS upload: %w", err)
	}

	url, err := client.Bucket(bucket).SignedURL(objectPath, &storage.SignedURLOptions{
		Scheme:  storage.SigningSchemeV4,
		Method:  "GET",
		Expires: time.Now().Add(15 * time.Minute),
	})
	if err != nil {
		return "", fmt.Errorf("error generating signed URL: %w", err)
	}

	return url, nil
}
