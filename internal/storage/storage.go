package storage

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"time"

	"cloud.google.com/go/storage"
	excelize "github.com/xuri/excelize/v2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/option"
)

func GenerateV4ReadObjectSignedURL(w io.Writer, base64EncodedKey string) (string, error) {
	bucket := "rpis"
	object := "reconcile"
	ctx := context.Background()

	decodedKey, err := base64.StdEncoding.DecodeString(base64EncodedKey)
	if err != nil {
		return "", fmt.Errorf("error decoding service account key: %w", err)
	}

	creds, err := google.CredentialsFromJSON(context.TODO(), decodedKey, storage.ScopeReadWrite)

	if err != nil {
		// handle error
		return "", fmt.Errorf("error parsing service account credentials key: %w", err)
	}

	client, err := storage.NewClient(ctx, option.WithCredentials(creds))
	if err != nil {
		return "", fmt.Errorf("error creating storage client: %w", err)
	}
	defer client.Close()

	opts := &storage.SignedURLOptions{
		Scheme:  storage.SigningSchemeV4,
		Method:  "Read",
		Expires: time.Now().Add(15 * time.Minute),
	}

	signedURL, err := client.Bucket(bucket).SignedURL(object, opts)
	if err != nil {
		return "", fmt.Errorf("error generating signed URL: %w", err)
	}

	return signedURL, nil
}

func UploadExcel(file *excelize.File) (string, error) {
	key := os.Getenv("GCP_SERVICE_ACCOUNT")
	bucket := os.Getenv("GCP_STORAGE_BUCKET")

	ctx := context.Background()

	decodedKey, err := base64.StdEncoding.DecodeString(key)
	if err != nil {
		return "", fmt.Errorf("error decoding service account key: %w", err)
	}

	creds, err := google.CredentialsFromJSON(context.TODO(), decodedKey, storage.ScopeReadWrite)

	if err != nil {
		// handle error
		return "", fmt.Errorf("error parsing service account credentials key: %w", err)
	}

	client, err := storage.NewClient(ctx, option.WithCredentials(creds))

	if err != nil {
		return "", fmt.Errorf("error creating storage client: %w", err)
	}

	defer client.Close()

	buff, err := file.WriteToBuffer()

	fmt.Println(len(buff.Bytes()))

	if err != nil {
		return "", err
	}

	ctx, cancel := context.WithTimeout(ctx, time.Second*60)
	defer cancel()

	filename, err := generateRandomFilename("xlsx")

	if err != nil {
		return "", fmt.Errorf("failed to generate a filename: %s", err.Error())
	}

	b := client.Bucket(bucket)

	o := b.Object(filename)
	o = o.If(storage.Conditions{DoesNotExist: true})
	wc := o.NewWriter(ctx)
	wc.ContentType = "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet"

	if _, err = io.Copy(wc, buff); err != nil {
		return "", fmt.Errorf("io.Copy: %w", err)
	}

	if err = wc.Close(); err != nil {
		return "", fmt.Errorf("Writer.Close: %w", err)
	}

	opts := &storage.SignedURLOptions{
		Scheme:  storage.SigningSchemeV4,
		Method:  "GET",
		Expires: time.Now().Add(15 * time.Minute),
	}

	url, err := b.SignedURL(filename, opts)

	return url, err
}

func generateRandomFilename(extension string) (string, error) {
	// Ensure extension does not include the period
	if len(extension) > 0 && extension[0] == '.' {
		extension = extension[1:]
	}

	// Generate random 16 bytes (32 hex characters)
	randomBytes := make([]byte, 16)
	if _, err := rand.Read(randomBytes); err != nil {
		return "", fmt.Errorf("failed to generate random bytes: %v", err)
	}

	// Convert random bytes to hex string
	randomHex := hex.EncodeToString(randomBytes)

	// Get the current date
	currentDate := time.Now().Format("2006-01")

	// Format the filename
	filename := fmt.Sprintf("expobj/%s/%s.%s", currentDate, randomHex, extension)
	return filename, nil
}
