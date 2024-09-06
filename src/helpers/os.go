package helpers

import (
	"archive/tar"
	"bytes"
	"context"
	"crypto/md5"  // #nosec G501 This is not a cryptographic function, it is used to calculate a file checksum
	"crypto/sha1" // #nosec G505 This is not a cryptographic function, it is used to calculate a file checksum
	"encoding/hex"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/briandowns/spinner"
)

var GlobalSpinner = spinner.New(spinner.CharSets[9], 500*time.Millisecond)

type Command struct {
	Command          string
	WorkingDirectory string
	Args             []string
}

const executionTimeout = 1 * time.Minute

func (c *Command) String() string {
	return fmt.Sprintf("%s %s", c.Command, strings.Join(c.Args, " "))
}

func CreateDirIfNotExist(path string) error {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		err = os.MkdirAll(path, 0o750)
		if err != nil {
			return fmt.Errorf("failed to create directory: %v", err)
		}
	}

	return nil
}

func ExecuteWithNoOutput(ctx context.Context, command Command) (string, error) {
	var executionContext context.Context
	var cancel context.CancelFunc
	if ctx != nil {
		executionContext, cancel = context.WithTimeout(ctx, executionTimeout)
	} else {
		ctx = context.TODO()
		executionContext, cancel = context.WithTimeout(ctx, executionTimeout)
	}

	defer cancel()

	cmd := exec.CommandContext(executionContext, command.Command, command.Args...) // #nosec G204 This is not a user input, it is a helper function
	if command.WorkingDirectory != "" {
		cmd.Dir = command.WorkingDirectory
	}

	var stdOut, stdIn, stderr bytes.Buffer

	cmd.Stdout = &stdOut
	cmd.Stderr = &stderr
	cmd.Stdin = &stdIn

	if err := cmd.Run(); err != nil {
		if stderr.String() != "" {
			return stdOut.String(), fmt.Errorf("%v, err: %v", stderr.String(), err.Error())
		} else {
			return stdOut.String(), fmt.Errorf("empty output, err: %v", err.Error())
		}
	}

	return stdOut.String(), nil
}

func ExecuteWithOutput(ctx context.Context, command Command) (stdout string, stderr string, exitCode int, err error) {
	var executionContext context.Context
	var cancel context.CancelFunc
	if ctx != nil {
		executionContext, cancel = context.WithTimeout(ctx, executionTimeout)
	} else {
		ctx = context.TODO()
		executionContext, cancel = context.WithTimeout(ctx, executionTimeout)
	}

	defer cancel()

	cmd := exec.CommandContext(executionContext, command.Command, command.Args...) // #nosec G204 This is not a user input, it is a helper function
	if command.WorkingDirectory != "" {
		cmd.Dir = command.WorkingDirectory
	}

	var stdOut, stdIn, stdErr bytes.Buffer

	cmd.Stdout = &stdOut
	cmd.Stderr = &stdErr
	cmd.Stdin = &stdIn

	if err := cmd.Run(); err != nil {
		if stdErr.String() != "" {
			stderr = strings.TrimSuffix(stdErr.String(), "\n")
			stdout = strings.TrimSuffix(stdOut.String(), "\n")
			return stdout, stderr, cmd.ProcessState.ExitCode(), fmt.Errorf("%v, err: %v", stdErr.String(), err.Error())
		} else {
			stderr = ""
			stdout = strings.TrimSuffix(stdOut.String(), "\n")
			return stdout, stderr, cmd.ProcessState.ExitCode(), fmt.Errorf("%v, err: %v", stdErr.String(), err.Error())
		}
	}

	stderr = ""
	stdout = strings.TrimSuffix(stdOut.String(), "\n")
	return stdout, stderr, cmd.ProcessState.ExitCode(), nil
}

func ExecuteAndWatch(command Command) (stdout string, stderr string, exitCode int, err error) {
	ctx, cancel := context.WithCancel(context.TODO())
	defer cancel()

	cmd := exec.CommandContext(ctx, command.Command, command.Args...) // #nosec G204 This is not a user input, it is a helper function
	if command.WorkingDirectory != "" {
		cmd.Dir = command.WorkingDirectory
	}
	var stdOut, stdIn, stdErr bytes.Buffer

	cmd.Stdout = io.MultiWriter(os.Stdout, &stdOut)
	cmd.Stderr = io.MultiWriter(os.Stderr, &stdErr)
	cmd.Stdin = &stdIn

	if err := cmd.Start(); err != nil {
		return stdOut.String(), stdErr.String(), cmd.ProcessState.ExitCode(), err
	}

	go func() {
		<-ctx.Done()
	}()

	if err := cmd.Wait(); err != nil {
		return stdOut.String(), stdErr.String(), cmd.ProcessState.ExitCode(), err
	}

	return stdOut.String(), stdErr.String(), cmd.ProcessState.ExitCode(), err
}

func RemoveFolder(path string) error {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return nil
	}

	err := os.RemoveAll(path)
	if err != nil {
		return fmt.Errorf("failed to remove folder: %v", err)
	}

	return nil
}

func MoveFolder(source string, destination string) error {
	// if !helper.DirectoryExists(destination) {
	// 	return fmt.Errorf("destination folder does not exist")
	// }

	// if !helper.DirectoryExists(source) {
	// 	return fmt.Errorf("source folder does not exist")
	// }

	err := os.Rename(source, destination)
	if err != nil {
		return err
	}
	return nil
}

func ToHCL(m map[string]interface{}, indent int) string {
	var lines []string
	for k, v := range m {
		switch v := v.(type) {
		case string:
			lines = append(lines, fmt.Sprintf("%s = \"%s\"", k, v))
		case bool:
			lines = append(lines, fmt.Sprintf("%s = %t", k, v))
		case []string:
			lines = append(lines, fmt.Sprintf("%s = [", k))
			for idx, item := range v {
				line := strings.Repeat(" ", 2*(indent+1))
				line = fmt.Sprintf("%v\"%s\"", line, item)
				if idx >= 0 && idx < len(v)-1 {
					line = fmt.Sprintf("%s,", line)
				}
				lines = append(lines, line)
			}
			if indent > 0 {
				lines = append(lines, fmt.Sprintf("%s%s", strings.Repeat(" ", 2*(indent)), "]"))
			} else {
				lines = append(lines, "]")
			}

		case []interface{}:
			lines = append(lines, fmt.Sprintf("%s = [", k))
			for idx, item := range v {
				line := ""
				switch item := item.(type) {
				case string:
					line = fmt.Sprintf("\"%s\"", item)
				case bool:
					line = fmt.Sprintf("%t", item)
				case map[string]interface{}:
					line = ToHCL(item, indent+1)
				default:
					line = fmt.Sprintf("%v", item)
				}
				if idx >= 0 && idx < len(v)-1 {
					line = fmt.Sprintf("%s,", line)
				}
				lines = append(lines, fmt.Sprintf("%s%s", strings.Repeat(" ", 2*(indent+1)), line))
			}
			if indent > 0 {
				lines = append(lines, fmt.Sprintf("%s%s", strings.Repeat(" ", 2*(indent)), "],"))
			} else {
				lines = append(lines, "]")
			}
		case map[string]interface{}:
			lines = append(lines, fmt.Sprintf("%s = {", k))
			count := 0
			for k2, v2 := range v {
				line := strings.Repeat(" ", 2*(indent+1))
				line = fmt.Sprintf("%v%v", line, ToHCL(map[string]interface{}{k2: v2}, indent+1))
				if count >= 0 && count < len(v)-1 {
					line = fmt.Sprintf("%s,", line)
				}
				lines = append(lines, line)
				count = count + 1
			}
			if indent > 0 {
				lines = append(lines, fmt.Sprintf("%s%s", strings.Repeat(" ", 2*(indent)), "}"))
			} else {
				lines = append(lines, "}")
			}
		default:
			lines = append(lines, fmt.Sprintf("%s = %v", k, v))
		}
	}
	result := strings.Join(lines, "\n")
	result = strings.TrimSuffix(result, ",")

	return result
}

func IsDirectory(path string) (bool, error) {
	fileInfo, err := os.Stat(path)
	if err != nil {
		return false, err
	}
	return fileInfo.IsDir(), nil
}

func GetFileChecksum(path string) (string, error) {
	file, err := os.Open(filepath.Clean(path))
	if err != nil {
		return "", err
	}
	defer file.Close()

	hash := sha1.New() // #nosec G401 This is not a cryptographic function, it is used to calculate a file checksum
	if _, err := io.Copy(hash, file); err != nil {
		return "", err
	}

	checksum := hex.EncodeToString(hash.Sum(nil))
	return checksum, nil
}

func GetCurrentDirectory() (string, error) {
	dir, err := os.Getwd()
	if err != nil {
		return "", err
	}

	return dir, nil
}

func GetFileMD5Checksum(path string) (string, error) {
	file, err := os.Open(filepath.Clean(path))
	if err != nil {
		return "", err
	}
	defer file.Close()

	hash := md5.New() // #nosec G401 This is not a cryptographic function, it is used to calculate a file checksum
	if _, err := io.Copy(hash, file); err != nil {
		return "", err
	}

	checksum := hex.EncodeToString(hash.Sum(nil))
	return checksum, nil
}

func CopyTarChunks(file *os.File, reader *tar.Reader) error {
	for {
		_, err := io.CopyN(file, reader, 1024)
		if err != nil {
			if err == io.EOF {
				break
			}
			return err
		}
	}

	return nil
}

// FileExists Checks if a file/directory exists
func FileExists(path string) bool {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return false
	}

	return true
}

// CopyDir recursively copies a directory tree, attempting to preserve permissions.
// Source directory must exist, destination directory must *not* exist.
// Symlinks are ignored and skipped.
func CopyDir(src string, dst string) (err error) {
	src = filepath.Clean(src)
	dst = filepath.Clean(dst)

	si, err := os.Stat(src)
	if err != nil {
		return err
	}
	if !si.IsDir() {
		return fmt.Errorf("source is not a directory")
	}

	// _, err = os.Stat(dst)
	// if err != nil && !os.IsNotExist(err) {
	// 	return
	// }
	// if err == nil {
	// 	return fmt.Errorf("destination already exists")
	// }

	// err = os.MkdirAll(dst, si.Mode())
	// if err != nil {
	// 	return
	// }

	if runtime.GOOS == "darwin" {
		if FileExists(dst) {
			os.RemoveAll(dst)
		}

		// fmt.Printf("Copying folder with macos clone %s, %s\n", src, dst)
		cmd := Command{
			Command: "cp",
			Args:    []string{"-c", "-r", src, dst},
		}
		// if the destination is a mounted volume, we cannot use the clone command
		if strings.HasPrefix(dst, "/Volumes") {
			cmd = Command{
				Command: "cp",
				Args:    []string{"-r", src, dst},
			}
		}

		if _, err = ExecuteWithNoOutput(context.TODO(), cmd); err != nil {
			return err
		}
		return
	}

	if FileExists(src) {
		err = os.MkdirAll(dst, si.Mode())
		if err != nil {
			return
		}
	}

	entries, err := ioutil.ReadDir(src)
	if err != nil {
		return
	}

	for _, entry := range entries {
		srcPath := filepath.Join(src, entry.Name())
		dstPath := filepath.Join(dst, entry.Name())

		if entry.IsDir() {
			err = CopyDir(srcPath, dstPath)
			if err != nil {
				return
			}
		} else {
			// Skip symlinks.
			if entry.Mode()&os.ModeSymlink != 0 {
				continue
			}

			err = CopyFile(srcPath, dstPath)
			if err != nil {
				return
			}
		}
	}

	return
}

// CopyFile copies a file from src to dst. If src and dst files exist, and are
// the same, then return success. Otherise, attempt to create a hard link
// between the two files. If that fail, copy the file contents from src to dst.
func CopyFile(src, dst string) (err error) {
	sfi, err := os.Stat(src)
	if err != nil {
		return
	}
	if !sfi.Mode().IsRegular() {
		// cannot copy non-regular files (e.g., directories,
		// symlinks, devices, etc.)
		return fmt.Errorf("CopyFile: non-regular source file %s (%q)", sfi.Name(), sfi.Mode().String())
	}
	dfi, err := os.Stat(dst)
	if err != nil {
		if !os.IsNotExist(err) {
			return
		}
	} else {
		if !(dfi.Mode().IsRegular()) {
			return fmt.Errorf("CopyFile: non-regular destination file %s (%q)", dfi.Name(), dfi.Mode().String())
		}
		if os.SameFile(sfi, dfi) {
			return
		}
	}
	if runtime.GOOS == "darwin" {
		fmt.Printf("Copying with macos clone %s, %s\n", src, dst)
		cmd := Command{
			Command: "cp",
			Args:    []string{"-c", src, dst},
		}

		if _, err := ExecuteWithNoOutput(context.Background(), cmd); err != nil {
			return err
		}

		return
	}

	// if err = os.Link(src, dst); err == nil {
	// 	return
	// }
	err = copyFileContents(src, dst)
	return
}

// copyFileContents copies the contents of the file named src to the file named
// by dst. The file will be created if it does not already exist. If the
// destination file exists, all it's contents will be replaced by the contents
// of the source file.
func copyFileContents(src, dst string) (err error) {
	in, err := os.Open(src)
	if err != nil {
		return
	}
	defer in.Close()
	out, err := os.Create(dst)
	if err != nil {
		return
	}
	defer func() {
		cerr := out.Close()
		if err == nil {
			err = cerr
		}
	}()
	if _, err = io.Copy(out, in); err != nil {
		return
	}
	err = out.Sync()
	return
}
