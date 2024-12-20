package compressor

import (
	"archive/tar"
	"bufio"
	"bytes"
	"compress/gzip"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/Parallels/prl-devops-service/basecontext"
	"github.com/Parallels/prl-devops-service/helpers"
	"github.com/Parallels/prl-devops-service/notifications"
)

// func DecompressFromReader(ctx basecontext.ApiContext, reader io.Reader, destination string) error {
// 	staringTime := time.Now()
// 	var fileReader io.Reader

// 	headerBuf := make([]byte, 512)
// 	n, err := io.ReadFull(reader, headerBuf)
// 	if err != nil && err != io.EOF {
// 		return fmt.Errorf("failed to read header: %w", err)
// 	}

// 	fileType, err := detectFileType(headerBuf)
// 	if err != nil {
// 		return err
// 	}

// 	reader = io.MultiReader(bytes.NewReader(headerBuf[:n]), reader)

// 	switch fileType {
// 	case "tar":
// 		fileReader = reader
// 	case "gzip":
// 		// Create a gzip reader
// 		bufferReader := bufio.NewReader(reader)
// 		gzipReader, err := gzip.NewReader(bufferReader)
// 		if err != nil {
// 			return err
// 		}
// 		defer gzipReader.Close()
// 		fileReader = gzipReader
// 	case "tar.gz":
// 		// Create a gzip reader
// 		bufferReader := bufio.NewReader(reader)
// 		gzipReader, err := gzip.NewReader(bufferReader)
// 		if err != nil {
// 			return err
// 		}
// 		defer gzipReader.Close()
// 		fileReader = gzipReader
// 	}

// 	// Creating the basedir if it does not exist
// 	if _, err := os.Stat(destination); os.IsNotExist(err) {
// 		if err := os.MkdirAll(destination, 0o750); err != nil {
// 			return err
// 		}
// 	}

// 	tarReader := tar.NewReader(fileReader)
// 	if err := processTarFile(ctx, tarReader, destination); err != nil {
// 		return err
// 	}

// 	endingTime := time.Now()
// 	ctx.LogInfof("Finished decompressing machine from stream to %s, in %v", destination, endingTime.Sub(staringTime))
// 	return nil
// }

func DecompressFromReader(ctx basecontext.ApiContext, reader io.Reader, destination string) error {
	startingTime := time.Now()

	// Read initial 512 bytes to determine file type
	headerBuf := make([]byte, 512)
	n, err := io.ReadFull(reader, headerBuf)
	if err != nil && err != io.EOF && n == 0 {
		return fmt.Errorf("failed to read header: %w", err)
	}
	// If file is smaller than 512 bytes, n < 512 is fine.

	fileType, err := detectFileType(headerBuf[:n])
	if err != nil {
		return err
	}

	// Put the initial bytes back into the reader stream
	reader = io.MultiReader(bytes.NewReader(headerBuf[:n]), reader)

	var fileReader io.Reader
	switch fileType {
	case "tar":
		fileReader = reader
	case "tar.gz":
		gzReader, err := gzip.NewReader(reader)
		if err != nil {
			return err
		}
		defer gzReader.Close()
		fileReader = gzReader
	case "gzip":
		// If you ever have pure gzip (non-tar), handle that here.
		gzReader, err := gzip.NewReader(reader)
		if err != nil {
			return err
		}
		defer gzReader.Close()
		fileReader = gzReader
		// If it's pure gzip (not tar), you'd handle differently, but let's assume tar.gz is primary.
	default:
		return fmt.Errorf("unsupported file type: %s", fileType)
	}

	// Ensure the destination directory exists
	if _, err := os.Stat(destination); os.IsNotExist(err) {
		if err := os.MkdirAll(destination, 0o750); err != nil {
			return err
		}
	}

	tarReader := tar.NewReader(fileReader)
	if err := processTarFile(ctx, tarReader, destination); err != nil {
		return err
	}

	endingTime := time.Now()
	ctx.LogInfof("Finished decompressing from stream to %s in %v", destination, endingTime.Sub(startingTime))
	return nil
}

func DecompressFile(ctx basecontext.ApiContext, filePath string, destination string) error {
	staringTime := time.Now()
	cleanFilePath := filepath.Clean(filePath)
	compressedFile, err := os.Open(cleanFilePath)
	if err != nil {
		return err
	}
	defer compressedFile.Close()

	fileHeader, err := readFileHeader(cleanFilePath)
	if err != nil {
		return err
	}

	fileType, err := detectFileType(fileHeader)
	if err != nil {
		return err
	}

	var fileReader io.Reader

	switch fileType {
	case "tar":
		fileReader = compressedFile
	case "gzip":
		// Create a gzip reader
		bufferReader := bufio.NewReader(compressedFile)
		gzipReader, err := gzip.NewReader(bufferReader)
		if err != nil {
			return err
		}
		defer gzipReader.Close()
		fileReader = gzipReader
	case "tar.gz":
		// Create a gzip reader
		bufferReader := bufio.NewReader(compressedFile)
		gzipReader, err := gzip.NewReader(bufferReader)
		if err != nil {
			return err
		}
		defer gzipReader.Close()
		fileReader = gzipReader
	}

	tarReader := tar.NewReader(fileReader)
	if err := processTarFile(ctx, tarReader, destination); err != nil {
		return err
	}

	endingTime := time.Now()
	ctx.LogInfof("Finished decompressing machine from %s to %s, in %v", filePath, destination, endingTime.Sub(staringTime))
	return nil
}

func processTarFile(ctx basecontext.ApiContext, tarReader *tar.Reader, destination string) error {
	ns := notifications.Get()
	for {
		header, err := tarReader.Next()
		if err != nil {
			if err == io.EOF {
				break
			}

			return err
		}

		destinationFilePath, err := helpers.SanitizeArchivePath(destination, header.Name)
		if err != nil {
			return err
		}

		if ns != nil {
			msg := fmt.Sprintf("Decompressing file %s", destinationFilePath)
			ns.NotifyProgress(destinationFilePath, msg, 0)
		}

		// Creating the basedir if it does not exist
		baseDir := filepath.Dir(destinationFilePath)
		if _, err := os.Stat(baseDir); os.IsNotExist(err) {
			if err := os.MkdirAll(baseDir, 0o750); err != nil {
				return err
			}
		}

		switch header.Typeflag {
		case tar.TypeDir:
			ctx.LogDebugf("Directory type found for file %v (byte %v, rune %v)", destinationFilePath, header.Typeflag, string(header.Typeflag))
			if _, err := os.Stat(destinationFilePath); os.IsNotExist(err) {
				if err := os.MkdirAll(destinationFilePath, os.FileMode(header.Mode)); err != nil {
					return err
				}
			}
		case tar.TypeReg:
			ctx.LogDebugf("HardFile type found for file %v (byte %v, rune %v): size %v", destinationFilePath, header.Typeflag, string(header.Typeflag), header.Size)
			file, err := os.OpenFile(filepath.Clean(destinationFilePath), os.O_CREATE|os.O_TRUNC|os.O_WRONLY, os.FileMode(header.Mode))
			if err != nil {
				return err
			}
			defer file.Close()

			if err := copyTarChunks(file, tarReader, header.Size); err != nil {
				return err
			}
		case tar.TypeGNUSparse:
			ctx.LogDebugf("Sparse File type found for file %v (byte %v, rune %v): size %v", destinationFilePath, header.Typeflag, string(header.Typeflag), header.Size)
			file, err := os.OpenFile(filepath.Clean(destinationFilePath), os.O_CREATE|os.O_TRUNC|os.O_WRONLY, os.FileMode(header.Mode))
			if err != nil {
				return err
			}
			defer file.Close()

			if err := copyTarChunks(file, tarReader, header.Size); err != nil {
				return err
			}
		case tar.TypeSymlink:
			ctx.LogDebugf("Symlink File type found for file %v (byte %v, rune %v)", destinationFilePath, header.Typeflag, string(header.Typeflag))
			os.Symlink(header.Linkname, destinationFilePath)
			realLinkPath, err := filepath.EvalSymlinks(filepath.Join(destination, header.Linkname))
			if err != nil {
				ctx.LogWarnf("Error resolving symlink path: %v", header.Linkname)
				if err := os.Remove(destinationFilePath); err != nil {
					return fmt.Errorf("failed to remove invalid symlink: %v", err)
				}
			} else {
				relLinkPath, err := filepath.Rel(destination, realLinkPath)
				if err != nil || strings.HasPrefix(filepath.Clean(relLinkPath), "..") {
					return fmt.Errorf("invalid symlink path: %v", header.Linkname)
				}
				os.Symlink(realLinkPath, destinationFilePath)
			}
		default:
			ctx.LogWarnf("Unknown type found for file %v, ignoring (byte %v, rune %v)", destinationFilePath, header.Typeflag, string(header.Typeflag))
		}
	}

	return nil
}

func copyTarChunks(file *os.File, reader *tar.Reader, fileSize int64) error {
	extractedSize := int64(0)
	lastPrintTime := time.Now()
	ns := notifications.Get()
	for {
		_, err := io.CopyN(file, reader, 1024)
		if err != nil {
			if err == io.EOF {
				msg := fmt.Sprintf("Decompressing file %s", file.Name())
				ns.NotifyProgress(file.Name(), msg, 100)
				break
			}
			return err
		}
		if ns != nil {
			extractedSize += 1024
			percentage := float64(extractedSize) / float64(fileSize) * 100
			if time.Since(lastPrintTime) >= 1*time.Second {
				msg := fmt.Sprintf("Decompressing file %s", file.Name())
				ns.NotifyProgress(file.Name(), msg, int(percentage))
				lastPrintTime = time.Now()
			}
		}
	}

	return nil
}

func readFileHeader(filepath string) ([]byte, error) {
	file, err := os.Open(filepath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	header := make([]byte, 512)
	n, err := file.Read(header)
	if err != nil && err != io.EOF {
		return nil, fmt.Errorf("could not read file header: %w", err)
	}

	file.Close()
	return header[:n], nil
}

// func detectFileType(header []byte) (string, error) {
// 	// Read the first 512 bytes
// 	buff := bytes.NewReader(header)

// 	n, err := buff.Read(header)
// 	if err != nil && err != io.EOF {
// 		return "", fmt.Errorf("could not read file header: %w", err)
// 	}
// 	header = header[:n]

// 	// Check for Gzip magic number
// 	if n >= 2 && header[0] == 0x1F && header[1] == 0x8B {
// 		// It's a gzip file, but is it a compressed tar?
// 		gzipReader, err := gzip.NewReader(buff)
// 		if err != nil {
// 			return "gzip", nil
// 		}
// 		defer gzipReader.Close()

// 		// Read the first 512 bytes of the decompressed data
// 		tarHeader := make([]byte, 512)
// 		n, err := gzipReader.Read(tarHeader)
// 		if err != nil && err != io.EOF {
// 			return "gzip", nil // It's a gzip file, but not a tar archive
// 		}
// 		tarHeader = tarHeader[:n]

// 		// Check for tar magic string in decompressed data
// 		if n > 262 {
// 			tarMagic := string(tarHeader[257 : 257+5])
// 			if tarMagic == "ustar" || tarMagic == "ustar\x00" {
// 				return "tar.gz", nil
// 			}
// 		}
// 		return "gzip", nil
// 	}

// 	// Check for Tar magic string at offset 257
// 	if n > 262 {
// 		tarMagic := string(header[257 : 257+5])
// 		if tarMagic == "ustar" || tarMagic == "ustar\x00" {
// 			return "tar", nil
// 		}
// 	}

// 	// If none of the above, return unknown
// 	return "unknown", errors.New("file format not recognized as gzip or tar")
// }

// detectFileType attempts to identify the file type based on its header bytes.
// It checks for gzip and tar archives.
//
// Supported file types:
//   - "gzip": Files starting with the gzip magic number (0x1F 0x8B).
//   - "tar":  Files containing the "ustar\000" sequence at offset 257 (can be plain or gzipped).
//   - "unknown": If no known file type is detected.
//
// Parameters:
//   - header:  A byte slice representing the beginning of the file's contents.
//   - data  :  The entire file data.
//
// Returns:
//   - string:  The identified file type ("gzip", "tar", or "unknown").
//   - error:   An error if the detection process fails, otherwise nil.
//
// Examples:
//
//		// Example using a gzip file header
//		gzipHeader := []byte{0x1F, 0x8B, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0xff}
//		fileType, err := detectFileType(gzipHeader, []byte{})
//		// fileType will be "gzip", err will be nil
//
//		// Example using a tar file header
//		tarHeader := make([]byte, 512)
//	 copy(tarHeader[257:], []byte("ustar\x00"))
//		fileType, err := detectFileType(tarHeader, []byte{})
//		// fileType will be "tar", err will be nil
//
//	 // Example using a non recognizable file header
//		unknownHeader := []byte{0x01, 0x02, 0x03, 0x04}
//		fileType, err := detectFileType(unknownHeader, []byte{})
//		// fileType will be "unknown", err will be non-nil
func detectFileType(header []byte) (string, error) {
	// Check for Gzip magic number
	if len(header) >= 2 && header[0] == 0x1F && header[1] == 0x8B {
		// We have a gzip file. Usually, this is tar.gz for your use-case.
		// If you want to distinguish pure gzip from tar.gz, you'd need to peek into the decompressed data.
		// But that requires another read. To keep it simple, assume tar.gz.
		return "tar.gz", nil
	}

	// Check for Tar magic
	if len(header) > 262 {
		tarMagic := string(header[257 : 257+5])
		if tarMagic == "ustar" || tarMagic == "ustar\x00" {
			return "tar", nil
		}
	}

	return "unknown", errors.New("file format not recognized as gzip or tar")
}
