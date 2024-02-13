package tester

import (
	"fmt"
	"path/filepath"

	"github.com/Parallels/pd-api-service/basecontext"
	"github.com/Parallels/pd-api-service/catalog"
	"github.com/Parallels/pd-api-service/helpers"
	"github.com/cjlapao/common-go/helper"
)

const (
	TESTING_REMOTE_FOLDER_NAME = "TESTING"
	TESTING_FILENAME           = "testing.txt"
	TESTING_FILE_CONTENT       = "----Test File Content----"
)

type TestProvider struct {
	ctx        basecontext.ApiContext
	service    *catalog.CatalogManifestService
	connection string
	checksum   string
}

func NewTestProvider(ctx basecontext.ApiContext, connection string) *TestProvider {
	return &TestProvider{
		ctx:        ctx,
		service:    catalog.NewManifestService(ctx),
		connection: connection,
	}
}

func (s *TestProvider) Test() error {
	if err := s.Check(); err != nil {
		s.ctx.LogErrorf("Error checking remote service: %v", err)
		return err
	}
	if err := s.testCreateFolder(); err != nil {
		s.ctx.LogErrorf("Error testing create folder: %v", err)
		if err := s.Clean(); err != nil {
			s.ctx.LogErrorf("Error cleaning: %v", err)
		}
		return err
	}
	if err := s.testFolderExists(); err != nil {
		s.ctx.LogErrorf("Error testing folder exists: %v", err)
		if err := s.Clean(); err != nil {
			s.ctx.LogErrorf("Error cleaning: %v", err)
		}

		return err
	}
	if err := s.testPushFile(); err != nil {
		s.ctx.LogErrorf("Error testing push file: %v", err)
		if err := s.Clean(); err != nil {
			s.ctx.LogErrorf("Error cleaning: %v", err)
		}

		return err
	}
	if err := s.testFileExists(); err != nil {
		s.ctx.LogErrorf("Error testing file exists: %v", err)
		if err := s.Clean(); err != nil {
			s.ctx.LogErrorf("Error cleaning: %v", err)
		}

		return err
	}
	if err := s.testFileChecksum(); err != nil {
		s.ctx.LogErrorf("Error testing file checksum: %v", err)
		if err := s.Clean(); err != nil {
			s.ctx.LogErrorf("Error cleaning: %v", err)
		}

		return err
	}
	if err := s.testPullFile(); err != nil {
		s.ctx.LogErrorf("Error testing pull file: %v", err)
		if err := s.Clean(); err != nil {
			s.ctx.LogErrorf("Error cleaning: %v", err)
		}

		return err
	}
	if err := s.testDeleteFile(); err != nil {
		s.ctx.LogErrorf("Error testing delete file: %v", err)
		if err := s.Clean(); err != nil {
			s.ctx.LogErrorf("Error cleaning: %v", err)
		}

		return err
	}
	if err := s.testDeleteFolder(); err != nil {
		s.ctx.LogErrorf("Error testing delete folder: %v", err)
		if err := s.Clean(); err != nil {
			s.ctx.LogErrorf("Error cleaning: %v", err)
		}

		return err
	}

	if err := s.Clean(); err != nil {
		s.ctx.LogErrorf("Error cleaning: %v", err)
	}

	return nil
}

func (s *TestProvider) Clean() error {
	execPath, err := helpers.GetCurrentDirectory()
	if err != nil {
		s.ctx.LogErrorf("Error getting current directory: %v", err)
		return err
	}
	tempFilePath := filepath.Join(execPath, TESTING_FILENAME)
	if err := helper.DeleteFile(tempFilePath); err != nil {
		return err
	}

	return nil
}

func (s *TestProvider) Check() error {
	found := false
	for _, rs := range s.service.GetProviders(s.ctx) {
		check, checkErr := rs.Check(s.ctx, s.connection)
		if checkErr != nil {
			s.ctx.LogErrorf("Error checking remote service %v: %v", rs.Name(), checkErr)
			return checkErr
		}

		if check {
			found = true
			break
		}
	}

	if !found {
		return fmt.Errorf("no remote service found")
	}

	return nil
}

func (s *TestProvider) testCreateFolder() error {
	for _, rs := range s.service.GetProviders(s.ctx) {
		check, checkErr := rs.Check(s.ctx, s.connection)
		if checkErr != nil {
			s.ctx.LogErrorf("Error checking remote service %v: %v", rs.Name(), checkErr)
			return checkErr
		}

		if check {
			s.ctx.LogInfof("Testing remote service %v create directory capability", rs.Name())
			if err := rs.CreateFolder(s.ctx, "/", TESTING_REMOTE_FOLDER_NAME); err != nil {
				s.ctx.LogErrorf("Error creating folder %v: %v", TESTING_REMOTE_FOLDER_NAME, err)
				return err
			}
		}
	}
	return nil
}

func (s *TestProvider) testFolderExists() error {
	for _, rs := range s.service.GetProviders(s.ctx) {
		check, checkErr := rs.Check(s.ctx, s.connection)
		if checkErr != nil {
			s.ctx.LogErrorf("Error checking remote service %v: %v", rs.Name(), checkErr)
			return checkErr
		}

		if check {
			s.ctx.LogInfof("Testing remote service %v folder exists capability", rs.Name())
			if exists, err := rs.FolderExists(s.ctx, "/", TESTING_REMOTE_FOLDER_NAME); err != nil {
				s.ctx.LogErrorf("Error checking if folder %v exists: %v", TESTING_REMOTE_FOLDER_NAME, err)
				return err
			} else if !exists {
				s.ctx.LogErrorf("Folder %v does not exist", TESTING_REMOTE_FOLDER_NAME)
				return fmt.Errorf("folder %v does not exist", TESTING_REMOTE_FOLDER_NAME)
			}
		}
	}

	return nil
}

func (s *TestProvider) testDeleteFolder() error {
	for _, rs := range s.service.GetProviders(s.ctx) {
		check, checkErr := rs.Check(s.ctx, s.connection)
		if checkErr != nil {
			s.ctx.LogErrorf("Error checking remote service %v: %v", rs.Name(), checkErr)
			return checkErr
		}

		if check {
			s.ctx.LogInfof("Testing remote service %v delete folder", rs.Name())
			if err := rs.DeleteFolder(s.ctx, "/", TESTING_REMOTE_FOLDER_NAME); err != nil {
				s.ctx.LogErrorf("Error deleting folder %v: %v", TESTING_REMOTE_FOLDER_NAME, err)
				return err
			}
		}
	}
	return nil
}

func (s *TestProvider) testPushFile() error {
	for _, rs := range s.service.GetProviders(s.ctx) {
		check, checkErr := rs.Check(s.ctx, s.connection)
		if checkErr != nil {
			s.ctx.LogErrorf("Error checking remote service %v: %v", rs.Name(), checkErr)
			return checkErr
		}

		if check {
			s.ctx.LogInfof("Testing remote service %v push file capability", rs.Name())
			execPath, err := helpers.GetCurrentDirectory()
			if err != nil {
				s.ctx.LogErrorf("Error getting current directory: %v", err)
				return err
			}
			tempFilePath := filepath.Join(execPath, TESTING_FILENAME)
			if err := helper.WriteToFile(TESTING_FILE_CONTENT, tempFilePath); err != nil {
				s.ctx.LogErrorf("Error writing to file %v: %v", tempFilePath, err)
				return err
			}

			checksum, err := helpers.GetFileMD5Checksum(tempFilePath)
			if err != nil {
				s.ctx.LogErrorf("Error getting file %v checksum: %v", tempFilePath, err)
			}

			s.checksum = checksum

			if err := rs.PushFile(s.ctx, execPath, TESTING_REMOTE_FOLDER_NAME, TESTING_FILENAME); err != nil {
				s.ctx.LogErrorf("Error pushing file to remote service %v: %v", rs.Name(), err)
				return err
			}
		}
	}
	return nil
}

func (s *TestProvider) testFileChecksum() error {
	for _, rs := range s.service.GetProviders(s.ctx) {
		check, checkErr := rs.Check(s.ctx, s.connection)
		if checkErr != nil {
			s.ctx.LogErrorf("Error checking remote service %v: %v", rs.Name(), checkErr)
			return checkErr
		}

		if check {
			s.ctx.LogInfof("Testing remote service %v file checksum capability", rs.Name())

			checksum, err := rs.FileChecksum(s.ctx, TESTING_REMOTE_FOLDER_NAME, TESTING_FILENAME)
			if err != nil {
				s.ctx.LogErrorf("Error getting file %v checksum: %v", TESTING_REMOTE_FOLDER_NAME, err)
			}

			if checksum != s.checksum {
				s.ctx.LogErrorf("File %v checksum is not correct", TESTING_FILENAME)
				return fmt.Errorf("file %v checksum is not correct", TESTING_FILENAME)
			}
		}
	}

	return nil
}

func (s *TestProvider) testPullFile() error {
	for _, rs := range s.service.GetProviders(s.ctx) {
		check, checkErr := rs.Check(s.ctx, s.connection)
		if checkErr != nil {
			s.ctx.LogErrorf("Error checking remote service %v: %v", rs.Name(), checkErr)
			return checkErr
		}

		if check {
			s.ctx.LogInfof("Testing remote service %v pull file capability", rs.Name())
			execPath, err := helpers.GetCurrentDirectory()
			if err != nil {
				s.ctx.LogErrorf("Error getting current directory: %v", err)
			}

			localFilePath := filepath.Join(execPath, TESTING_FILENAME)
			if helper.FileExists(localFilePath) {
				if err := helper.DeleteFile(localFilePath); err != nil {
					s.ctx.LogErrorf("Error deleting file %v: %v", TESTING_FILENAME, err)
					return err
				}
			}

			if err := rs.PullFile(s.ctx, TESTING_REMOTE_FOLDER_NAME, TESTING_FILENAME, execPath); err != nil {
				s.ctx.LogErrorf("Error pulling file from remote service %v: %v", rs.Name(), err)
				return err
			}

			content, err := helper.ReadFromFile(localFilePath)
			if err != nil {
				s.ctx.LogErrorf("Error reading file %v: %v", TESTING_FILENAME, err)
				return err
			}

			if string(content) != TESTING_FILE_CONTENT {
				s.ctx.LogErrorf("File %v content is not correct", TESTING_FILENAME)
				return fmt.Errorf("file %v content is not correct", TESTING_FILENAME)
			}
		}
	}

	return nil
}

func (s *TestProvider) testFileExists() error {
	for _, rs := range s.service.GetProviders(s.ctx) {
		check, checkErr := rs.Check(s.ctx, s.connection)
		if checkErr != nil {
			s.ctx.LogErrorf("Error checking remote service %v: %v", rs.Name(), checkErr)
			return checkErr
		}

		if check {
			s.ctx.LogInfof("Testing remote service %v file exist capability", rs.Name())
			if exists, err := rs.FileExists(s.ctx, TESTING_REMOTE_FOLDER_NAME, TESTING_FILENAME); err != nil {
				s.ctx.LogErrorf("Error checking if file %v exists: %v", TESTING_FILENAME, err)
				return err
			} else if !exists {
				s.ctx.LogErrorf("File %v does not exist", TESTING_FILENAME)
				return fmt.Errorf("file %v does not exist", TESTING_FILENAME)
			}
		}
	}

	return nil
}

func (s *TestProvider) testDeleteFile() error {
	for _, rs := range s.service.GetProviders(s.ctx) {
		check, checkErr := rs.Check(s.ctx, s.connection)
		if checkErr != nil {
			s.ctx.LogErrorf("Error checking remote service %v: %v", rs.Name(), checkErr)
			return checkErr
		}

		if check {
			s.ctx.LogInfof("Testing remote service %v delete file capability", rs.Name())
			if err := rs.DeleteFile(s.ctx, TESTING_REMOTE_FOLDER_NAME, TESTING_FILENAME); err != nil {
				s.ctx.LogErrorf("Error deleting file %v: %v", TESTING_FILENAME, err)
				return err
			}
		}
	}
	return nil
}
