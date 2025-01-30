package main

import (
	"bytes"
	"io"
	"log"
	"os"
	"path/filepath"
	"time"

	"ftp-docker-local/internal/service"
	"ftp-docker-local/internal/storage"

	"github.com/jlaffaye/ftp"
	"github.com/joho/godotenv"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Println("Note: No .env file found, using system environment variables")
	}

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

	entries, err := service.ListValidFiles(client, dir)
	if err != nil {
		log.Fatalf("File listing error: %v", err)
	}

	if len(entries) == 0 {
		log.Fatal("No files found in the directory")
	}

	entry := entries[0]

	//Download the file from the FTP server
	remotePath := filepath.Join(dir, entry)
	resp, err := client.Retr(remotePath)
	if err != nil {
		log.Fatalf("Failed to download %s: %v", remotePath, err)
	}
	defer resp.Close()

	//Read the file content into a byte slice
	content, err := io.ReadAll(resp)
	if err != nil {
		log.Fatalf("Failed to read %s: %v", remotePath, err)
	}

	url, err := storage.UploadCSV(bytes.NewBuffer(content), entry)
	if err != nil {
		log.Fatalf("Failed to upload %s to GCS: %v", entry, err)
	}

	log.Printf("File uploaded to GCS: %s", url)
}
