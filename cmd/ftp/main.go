package main

import (
	"bytes"
	"fmt"
	"ftp-docker-local/internal/storage"
	"io"
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/jlaffaye/ftp"
	"github.com/joho/godotenv"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Println("Note: No .env file found, using system environment variables")
	}

	filename := templateFile()

	host := os.Getenv("HOST")
	user := os.Getenv("USERNAME")
	pass := os.Getenv("PASSWORD")
	dir := os.Getenv("DIRECTORY")

	client, err := ftp.Dial(host, ftp.DialWithTimeout(5*time.Second))
	if err != nil {
		log.Fatalf("FTP connection failed: %v", err)
	}
	defer client.Quit()

	if err := client.Login(user, pass); err != nil {
		log.Fatalf("FTP login failed: %v", err)
	}

	filePath := filepath.Join(dir, filename)

	entry, err := client.Retr(filePath)

	if err != nil {
		log.Fatalf("Failed to download %s: %v", filename, err)
	}
	defer entry.Close()

	content, err := io.ReadAll(entry)
	if err != nil {
		log.Fatalf("Failed to read %s: %v", filename, err)
	}

	url, err := storage.UploadCSV(bytes.NewBuffer(content), filename)
	if err != nil {
		log.Fatalf("Failed to upload %s to GCS: %v", filename, err)
	}

	log.Printf("File uploaded to GCS: %s", url)
}

func templateFile() string {
	baseFilename := os.Getenv("FILENAME")
	mId := os.Getenv("MERCHANT_ID")
	date := time.Now().Format("20060102")

	return fmt.Sprintf("%s_%s_%s.csv", baseFilename, mId, date)
}
