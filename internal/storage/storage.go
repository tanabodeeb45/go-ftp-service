package storage

import (
	"bytes"
	"context"
	"encoding/base64"
	"fmt"
	"io"
	"os"
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

	year := time.Now().Year()
	month := time.Now().Month()

	objPath := fmt.Sprintf("reconcile/%d/%d/%s", year, month, filename)

	writer := client.Bucket(bucket).Object(objPath).NewWriter(ctx)
	writer.ContentType = "text/csv"

	if _, err := io.Copy(writer, fileContent); err != nil {
		return "", fmt.Errorf("error uploading file to GCS: %w", err)
	}

	if err := writer.Close(); err != nil {
		return "", fmt.Errorf("error finalizing GCS upload: %w", err)
	}

	url, err := client.Bucket(bucket).SignedURL(objPath, &storage.SignedURLOptions{
		Scheme:  storage.SigningSchemeV4,
		Method:  "GET",
		Expires: time.Now().Add(15 * time.Minute),
	})
	if err != nil {
		return "", fmt.Errorf("error generating signed URL: %w", err)
	}

	return url, nil
}
