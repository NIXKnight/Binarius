package installer

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/nixknight/binarius/internal/utils"
)

// VerifyChecksum verifies that a file's SHA256 checksum matches the expected value.
// This function uses streaming to handle large files efficiently without loading
// them entirely into memory.
//
// Parameters:
//   - filePath: Path to the file to verify
//   - expectedSHA256: Expected SHA256 checksum in hexadecimal format (case-insensitive)
//
// Returns an error if:
//   - The file doesn't exist
//   - The file cannot be read
//   - The checksum doesn't match
func VerifyChecksum(filePath, expectedSHA256 string) error {
	// Open the file
	file, err := os.Open(filePath)
	if err != nil {
		return utils.NewUserError(
			fmt.Sprintf("Failed to open file for checksum verification: %s", filePath),
			err.Error(),
			fmt.Sprintf("Ensure the file exists and is readable: %s", filePath),
		)
	}
	defer func() { _ = file.Close() }()

	// Create SHA256 hasher
	hasher := sha256.New()

	// Stream file content to hasher (memory efficient)
	if _, err := io.Copy(hasher, file); err != nil {
		return utils.NewUserError(
			fmt.Sprintf("Failed to read file for checksum verification: %s", filePath),
			err.Error(),
			"Ensure the file is not corrupted and is readable",
		)
	}

	// Compute the checksum
	actualChecksum := hex.EncodeToString(hasher.Sum(nil))

	// Normalize both checksums to lowercase for comparison
	expectedSHA256 = strings.ToLower(strings.TrimSpace(expectedSHA256))
	actualChecksum = strings.ToLower(actualChecksum)

	// Compare checksums
	if actualChecksum != expectedSHA256 {
		return utils.NewUserError(
			"Checksum verification failed",
			fmt.Sprintf("Downloaded file checksum mismatch. Expected: %s, Got: %s", expectedSHA256, actualChecksum),
			"The downloaded file may be corrupted or tampered with. Please try downloading again.",
		)
	}

	return nil
}
