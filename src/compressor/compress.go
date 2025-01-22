package compressor

import (
	"archive/tar"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/Parallels/prl-devops-service/basecontext"
	"github.com/Parallels/prl-devops-service/notifications"
)

func Compress(ctx basecontext.ApiContext, path string, compressedFilename string, destination string) (string, error) {
	startingTime := time.Now()
	tarFilename := compressedFilename
	tarFilePath := filepath.Join(destination, filepath.Clean(tarFilename))

	tarFile, err := os.Create(filepath.Clean(tarFilePath))
	if err != nil {
		return "", err
	}
	defer tarFile.Close()

	tarWriter := tar.NewWriter(tarFile)
	defer tarWriter.Close()

	countFiles := 0
	if err := filepath.Walk(path, func(filePath string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}
		countFiles += 1
		return nil
	}); err != nil {
		return "", err
	}

	compressed := 1
	err = filepath.Walk(path, func(machineFilePath string, info os.FileInfo, err error) error {
		ctx.LogInfof("[%v/%v] Compressing file %v", compressed, countFiles, machineFilePath)
		compressed += 1
		if err != nil {
			return err
		}

		if info.IsDir() {
			compressed -= 1
			return nil
		}

		f, err := os.Open(filepath.Clean(machineFilePath))
		if err != nil {
			return err
		}
		defer f.Close()

		relPath := strings.TrimPrefix(machineFilePath, path)
		hdr := &tar.Header{
			Name: relPath,
			Mode: int64(info.Mode()),
			Size: info.Size(),
		}
		if err := tarWriter.WriteHeader(hdr); err != nil {
			return err
		}

		n, err := io.Copy(tarWriter, f)
		if err != nil {
			return err
		}
		if info.Size() > 0 {
			ns := notifications.Get()
			percentage := float64(n) * 100 / float64(info.Size())
			if ns != nil {
				prefix := "Compressing file " + machineFilePath
				msg := notifications.NewProgressNotificationMessage(compressedFilename, prefix, percentage)
				ns.Notify(msg)
			}
		}
		return err
	})
	if err != nil {
		return "", err
	}

	endingTime := time.Now()
	ctx.LogInfof("Finished compressing machine from %s to %s in %v", path, tarFilePath, endingTime.Sub(startingTime))
	return tarFilePath, nil
}
