package main

import (
	"log"
	"os"
	"time"

	"ftp-docker-local/internal/service"

	"github.com/jlaffaye/ftp"
	"github.com/joho/godotenv"
)

func main() {
	godotenv.Load()

	host := os.Getenv("HOST")
	usr := os.Getenv("USERNAME")
	pwd := os.Getenv("PASSWORD")
	dir := os.Getenv("DIRECTORY")

	client, err := ftp.Dial(host, ftp.DialWithTimeout(5*time.Second))

	if err != nil {
		log.Fatal("Connection error:", err)
	}

	defer client.Quit()

	if err := client.Login(usr, pwd); err != nil {
		log.Fatal("Login failed:", err)
	}

	var remoteDir string

	if dir == "" {
		remoteDir = "reconcile_local"
	} else {
		remoteDir = dir
	}

	entry, err := service.ListValidFiles(client, remoteDir)

	if err != nil {
		log.Fatal("List error:", err)
	}

	for _, file := range entry {
		log.Println("File:", file)
	}
}
