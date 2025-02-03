package service

import (
	"errors"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/jlaffaye/ftp"
)

func ListValidFiles(client *ftp.ServerConn, remotePath string) ([]string, error) {
	entries, err := client.List(remotePath)

	if err != nil {
		return nil, fmt.Errorf("ftp list failed: %w", err)
	}

	var validFiles []string
	for _, entry := range entries {
		isValid, err := validateFileName(entry.Name)
		if !isValid {
			fmt.Printf("Skipping invalid file: %s. Reason: %v\n", entry.Name, err)
			continue
		}
		validFiles = append(validFiles, entry.Name)
	}

	if len(validFiles) == 0 {
		return nil, fmt.Errorf("no valid files found in directory")
	}

	fmt.Println("Valid files: ", validFiles)

	return validFiles, nil
}

func validateFileName(name string) (bool, error) {
	//expected format: FundTransferMerchantReconcile_20250130.csv
	expectedParts := 3
	expectedPrefix := "FundTransferMerchantReconcile"
	expectedSuffix := ".csv"
	datePartLength := 8
	fullDateLength := 12

	parts := strings.Split(name, "_")
	now := time.Now().Format("20060102")

	if len(parts) != expectedParts {
		return false, fmt.Errorf("invalid format: expected %d parts, got %d", expectedParts, len(parts))
	}

	if parts[0] != expectedPrefix {
		return false, fmt.Errorf("invalid prefix: expected %s, got %s", expectedPrefix, parts[0])
	}

	if parts[1] == "" {
		return false, errors.New("empty date")
	}

	if len(parts[2]) != fullDateLength || !strings.HasSuffix(parts[2], expectedSuffix) {
		return false, fmt.Errorf("invalid suffix: expected %s, got %s", expectedSuffix, parts[2])
	}

	dateStr := parts[2][:datePartLength]

	if _, err := time.Parse("20060102", dateStr); err != nil {
		return false, fmt.Errorf("invalid date: %s", err)
	}

	if dateStr != now {
		return false, fmt.Errorf("date mismatch: expected %s, got %s", now, dateStr)
	}

	return true, nil
}

func DelFile(client *ftp.ServerConn, filePath string, retries int) error {
	var err error
	for i := 0; i < retries; i++ {
		err = client.Delete(filePath)
		if err == nil {
			return nil
		}
		log.Printf("Attempt %d to delete file failed: %v", i+1, err)
		time.Sleep(time.Second * 2)
	}
	return err
}
