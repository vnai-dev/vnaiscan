package extractor

import (
	"archive/tar"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

// SafeExtract extracts a tar stream to dstDir with security protections:
// - Prevents path traversal attacks (../)
// - Skips symlinks, hardlinks, device nodes
// - Validates all paths stay within dstDir
func SafeExtract(r io.Reader, dstDir string) error {
	// Ensure dstDir is absolute and clean
	dstDir, err := filepath.Abs(dstDir)
	if err != nil {
		return fmt.Errorf("failed to resolve destination: %w", err)
	}

	tr := tar.NewReader(r)

	for {
		header, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return fmt.Errorf("failed to read tar header: %w", err)
		}

		// Clean and validate the path
		cleanPath, err := sanitizePath(header.Name, dstDir)
		if err != nil {
			// Log and skip dangerous paths
			fmt.Fprintf(os.Stderr, "⚠️  Skipping dangerous path: %s (%v)\n", header.Name, err)
			continue
		}

		// Skip dangerous file types
		switch header.Typeflag {
		case tar.TypeSymlink:
			fmt.Fprintf(os.Stderr, "⚠️  Skipping symlink: %s -> %s\n", header.Name, header.Linkname)
			continue
		case tar.TypeLink:
			fmt.Fprintf(os.Stderr, "⚠️  Skipping hardlink: %s -> %s\n", header.Name, header.Linkname)
			continue
		case tar.TypeChar, tar.TypeBlock:
			fmt.Fprintf(os.Stderr, "⚠️  Skipping device node: %s\n", header.Name)
			continue
		case tar.TypeFifo:
			fmt.Fprintf(os.Stderr, "⚠️  Skipping FIFO: %s\n", header.Name)
			continue
		}

		// Create directories
		if header.Typeflag == tar.TypeDir {
			if err := os.MkdirAll(cleanPath, 0755); err != nil {
				return fmt.Errorf("failed to create directory %s: %w", cleanPath, err)
			}
			continue
		}

		// Create regular files
		if header.Typeflag == tar.TypeReg {
			if err := extractFile(tr, cleanPath, header.Mode); err != nil {
				return fmt.Errorf("failed to extract file %s: %w", cleanPath, err)
			}
		}
	}

	return nil
}

// sanitizePath validates and cleans a tar path to prevent path traversal
func sanitizePath(name string, dstDir string) (string, error) {
	// Clean the path and remove leading slashes
	clean := filepath.Clean(name)
	clean = strings.TrimPrefix(clean, "/")
	clean = strings.TrimPrefix(clean, "./")

	// Check for path traversal attempts
	if strings.Contains(clean, "..") {
		return "", fmt.Errorf("path traversal detected")
	}

	// Build the full destination path
	fullPath := filepath.Join(dstDir, clean)

	// Verify the path is still within dstDir (defense in depth)
	if !strings.HasPrefix(fullPath, dstDir+string(os.PathSeparator)) && fullPath != dstDir {
		return "", fmt.Errorf("path escapes destination directory")
	}

	return fullPath, nil
}

// extractFile safely extracts a single file from the tar stream
func extractFile(tr *tar.Reader, path string, mode int64) error {
	// Ensure parent directory exists
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return err
	}

	// Create the file with restricted permissions initially
	f, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0600)
	if err != nil {
		return err
	}
	defer f.Close()

	// Copy with size limit (10GB max per file)
	const maxFileSize = 10 * 1024 * 1024 * 1024
	limited := io.LimitReader(tr, maxFileSize)

	if _, err := io.Copy(f, limited); err != nil {
		return err
	}

	// Set final permissions (but not executable for safety unless explicitly set)
	finalMode := os.FileMode(mode) & 0755
	if err := os.Chmod(path, finalMode); err != nil {
		return err
	}

	return nil
}
