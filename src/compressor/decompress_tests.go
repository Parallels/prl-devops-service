package compressor

import (
	"bytes"
	"errors"
	"testing"
)

func TestDetectFileType(t *testing.T) {
	testCases := []struct {
		name         string
		header       []byte
		data         []byte
		expectedType string
		expectedErr  error
	}{
		{
			name:         "gzip file",
			header:       []byte{0x1F, 0x8B, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0xff},
			expectedType: "gzip",
			expectedErr:  nil,
		},
		{
			name:         "truncated gzip file",
			header:       []byte{0x1F},
			expectedType: "unknown",
			expectedErr:  errors.New("could not read file header: EOF"),
		},
		{
			name: "tar file",
			header: func() []byte {
				tarHeader := make([]byte, 512)
				copy(tarHeader[257:], []byte("ustar\x00"))
				return tarHeader
			}(),
			expectedType: "tar",
			expectedErr:  nil,
		},
		{
			name: "tar file with some leading data",
			header: func() []byte {
				tarHeader := make([]byte, 512)
				copy(tarHeader[257:], []byte("ustar\x00"))
				return tarHeader
			}(),
			data:         bytes.Repeat([]byte{0x00}, 2000),
			expectedType: "tar",
			expectedErr:  nil,
		},
		{
			name:         "truncated tar file",
			header:       []byte("ustar"),
			expectedType: "unknown",
			expectedErr:  errors.New("could not read file header: EOF"),
		},
		{
			name:         "unknown file",
			header:       []byte{0x01, 0x02, 0x03, 0x04},
			expectedType: "unknown",
			expectedErr:  errors.New("file format not recognized as gzip or tar"),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			fileType, err := detectFileType(tc.header)

			if fileType != tc.expectedType {
				t.Errorf("Expected file type %q, but got %q", tc.expectedType, fileType)
			}

			if tc.expectedErr != nil {
				if err == nil {
					t.Errorf("Expected error %q, but got nil", tc.expectedErr)
				} else if err.Error() != tc.expectedErr.Error() {
					t.Errorf("Expected error %q, but got %q", tc.expectedErr, err)
				}
			} else if err != nil {
				t.Errorf("Expected nil error, but got: %q", err)
			}
		})
	}
}
