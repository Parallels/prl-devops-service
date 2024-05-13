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
	"os"
	"os/exec"
	"path/filepath"
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

func ExecuteWithNoOutput(command Command) (string, error) {
	cmd := exec.Command(command.Command, command.Args...) // #nosec G204 This is not a user input, it is a helper function
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

func ExecuteWithOutput(command Command) (stdout string, stderr string, exitCode int, err error) {
	cmd := exec.Command(command.Command, command.Args...) // #nosec G204 This is not a user input, it is a helper function
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
