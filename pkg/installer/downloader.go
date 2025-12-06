package installer

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/nixknight/binarius/internal/utils"
)

// Download downloads a file from the specified URL to the destination path.
// Uses streaming to handle large files efficiently.
//
// Parameters:
//   - url: The HTTPS URL to download from
//   - destPath: The local file path where the download should be saved
func Download(url, destPath string) error {
	// Create HTTP client with timeout
	client := &http.Client{
		Timeout: 300 * time.Second, // 5 minutes timeout for large downloads
	}

	// Create HTTP GET request
	resp, err := client.Get(url)
	if err != nil {
		return utils.NewUserError(
			fmt.Sprintf("Failed to download file from %s", url),
			err.Error(),
			"Check your internet connection and ensure the URL is correct",
		)
	}
	defer func() { _ = resp.Body.Close() }()

	// Check HTTP status
	if resp.StatusCode != http.StatusOK {
		return utils.NewUserError(
			fmt.Sprintf("Failed to download file from %s", url),
			fmt.Sprintf("HTTP error: %d %s", resp.StatusCode, resp.Status),
			"The file may not be available. Verify the tool version exists.",
		)
	}

	// Ensure destination directory exists
	destDir := filepath.Dir(destPath)
	if err := os.MkdirAll(destDir, 0755); err != nil {
		return utils.NewUserError(
			fmt.Sprintf("Failed to create destination directory: %s", destDir),
			err.Error(),
			"Ensure you have write permissions for the cache directory",
		)
	}

	// Create destination file
	destFile, err := os.Create(destPath)
	if err != nil {
		return utils.NewUserError(
			fmt.Sprintf("Failed to create destination file: %s", destPath),
			err.Error(),
			"Ensure you have write permissions for the cache directory",
		)
	}

	// Stream download to file
	_, err = io.Copy(destFile, resp.Body)
	if err != nil {
		_ = destFile.Close()
		// Clean up partial download
		_ = os.Remove(destPath)
		return utils.NewUserError(
			"Download interrupted",
			err.Error(),
			"Network connection may have been lost. Please try again.",
		)
	}

	if err := destFile.Close(); err != nil {
		return utils.NewUserError(
			"Failed to close downloaded file",
			err.Error(),
			"Disk may be full or write permissions may have changed",
		)
	}

	return nil
}
