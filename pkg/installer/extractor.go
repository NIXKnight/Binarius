package installer

import (
	"archive/tar"
	"archive/zip"
	"compress/gzip"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/nixknight/binarius/internal/utils"
)

// ExtractZip extracts a ZIP archive to the specified destination directory.
// Sets executable permissions (0755) on extracted files.
// Prevents path traversal attacks by validating all extraction paths.
//
// Parameters:
//   - zipPath: Path to the ZIP archive
//   - destDir: Destination directory for extraction
func ExtractZip(zipPath, destDir string) error {
	// Open the ZIP archive
	reader, err := zip.OpenReader(zipPath)
	if err != nil {
		return utils.NewUserError(
			fmt.Sprintf("Failed to open ZIP archive: %s", zipPath),
			err.Error(),
			"Ensure the file is a valid ZIP archive and is not corrupted",
		)
	}
	defer reader.Close()

	// Create destination directory if it doesn't exist
	if err := os.MkdirAll(destDir, 0755); err != nil {
		return utils.NewUserError(
			fmt.Sprintf("Failed to create destination directory: %s", destDir),
			err.Error(),
			"Ensure you have write permissions for the directory",
		)
	}

	// Extract each file in the archive
	for _, file := range reader.File {
		if err := extractZipFile(file, destDir); err != nil {
			return err
		}
	}

	return nil
}

// extractZipFile extracts a single file from a ZIP archive.
func extractZipFile(file *zip.File, destDir string) error {
	// Construct target path
	targetPath := filepath.Join(destDir, file.Name)

	// Validate path to prevent traversal attacks
	if err := validateExtractPath(destDir, targetPath); err != nil {
		return err
	}

	// Check if it's a directory
	if file.FileInfo().IsDir() {
		return os.MkdirAll(targetPath, 0755)
	}

	// Create parent directories
	if err := os.MkdirAll(filepath.Dir(targetPath), 0755); err != nil {
		return fmt.Errorf("failed to create parent directory for %s: %w", targetPath, err)
	}

	// Open source file in archive
	srcFile, err := file.Open()
	if err != nil {
		return fmt.Errorf("failed to open file in archive %s: %w", file.Name, err)
	}
	defer srcFile.Close()

	// Create destination file
	destFile, err := os.OpenFile(targetPath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0755)
	if err != nil {
		return fmt.Errorf("failed to create destination file %s: %w", targetPath, err)
	}
	defer destFile.Close()

	// Copy content
	if _, err := io.Copy(destFile, srcFile); err != nil {
		return fmt.Errorf("failed to extract file %s: %w", file.Name, err)
	}

	return nil
}

// ExtractTarGz extracts a tar.gz archive to the specified destination directory.
// Sets executable permissions (0755) on extracted files.
// Prevents path traversal attacks by validating all extraction paths.
//
// Parameters:
//   - tarGzPath: Path to the tar.gz archive
//   - destDir: Destination directory for extraction
func ExtractTarGz(tarGzPath, destDir string) error {
	// Open the tar.gz file
	file, err := os.Open(tarGzPath)
	if err != nil {
		return utils.NewUserError(
			fmt.Sprintf("Failed to open tar.gz archive: %s", tarGzPath),
			err.Error(),
			"Ensure the file exists and is readable",
		)
	}
	defer file.Close()

	// Create gzip reader
	gzipReader, err := gzip.NewReader(file)
	if err != nil {
		return utils.NewUserError(
			fmt.Sprintf("Failed to decompress gzip archive: %s", tarGzPath),
			err.Error(),
			"Ensure the file is a valid gzip archive",
		)
	}
	defer gzipReader.Close()

	// Create tar reader
	tarReader := tar.NewReader(gzipReader)

	// Create destination directory if it doesn't exist
	if err := os.MkdirAll(destDir, 0755); err != nil {
		return utils.NewUserError(
			fmt.Sprintf("Failed to create destination directory: %s", destDir),
			err.Error(),
			"Ensure you have write permissions for the directory",
		)
	}

	// Extract each file in the archive
	for {
		header, err := tarReader.Next()
		if err == io.EOF {
			break // End of archive
		}
		if err != nil {
			return fmt.Errorf("failed to read tar archive: %w", err)
		}

		if err := extractTarFile(tarReader, header, destDir); err != nil {
			return err
		}
	}

	return nil
}

// extractTarFile extracts a single file from a tar archive.
func extractTarFile(tarReader *tar.Reader, header *tar.Header, destDir string) error {
	// Construct target path
	targetPath := filepath.Join(destDir, header.Name)

	// Validate path to prevent traversal attacks
	if err := validateExtractPath(destDir, targetPath); err != nil {
		return err
	}

	// Handle different file types
	switch header.Typeflag {
	case tar.TypeDir:
		// Create directory
		return os.MkdirAll(targetPath, 0755)

	case tar.TypeReg:
		// Create parent directories
		if err := os.MkdirAll(filepath.Dir(targetPath), 0755); err != nil {
			return fmt.Errorf("failed to create parent directory for %s: %w", targetPath, err)
		}

		// Create destination file
		destFile, err := os.OpenFile(targetPath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0755)
		if err != nil {
			return fmt.Errorf("failed to create destination file %s: %w", targetPath, err)
		}
		defer destFile.Close()

		// Copy content
		if _, err := io.Copy(destFile, tarReader); err != nil {
			return fmt.Errorf("failed to extract file %s: %w", header.Name, err)
		}

	default:
		// Skip other file types (symlinks, devices, etc.)
		// Log or ignore based on requirements
	}

	return nil
}

// validateExtractPath validates that an extraction path is safe and doesn't
// attempt path traversal (e.g., using .. sequences).
//
// Parameters:
//   - destDir: The intended destination directory
//   - targetPath: The full path where a file will be extracted
//
// Returns an error if the targetPath is outside destDir (path traversal attempt).
func validateExtractPath(destDir, targetPath string) error {
	// Get absolute paths
	absDestDir, err := filepath.Abs(destDir)
	if err != nil {
		return fmt.Errorf("failed to get absolute path for destination directory: %w", err)
	}

	absTargetPath, err := filepath.Abs(targetPath)
	if err != nil {
		return fmt.Errorf("failed to get absolute path for target: %w", err)
	}

	// Check for path traversal
	if !strings.HasPrefix(absTargetPath, absDestDir+string(filepath.Separator)) &&
		absTargetPath != absDestDir {
		return utils.NewUserError(
			"Path traversal attempt detected",
			fmt.Sprintf("Archive contains unsafe path: %s", targetPath),
			"This archive may be malicious. Do not install tools from untrusted sources.",
		)
	}

	return nil
}
