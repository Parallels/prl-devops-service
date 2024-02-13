package helpers

import (
	"net/url"
	"strings"

	"github.com/Parallels/prl-devops-service/errors"
)

func JoinUrl(segments []string) (*url.URL, error) {
	if len(segments) == 0 {
		return nil, errors.New("segments cannot be empty")
	}

	var result *url.URL
	address := ""
	for _, segment := range segments {
		segment = strings.TrimSpace(segment)
		if segment == "" {
			continue
		}
		if address == "" {
			address = segment
			continue
		}
		join := "/"
		if strings.HasSuffix(address, "://") ||
			strings.HasSuffix(address, "/") ||
			strings.HasPrefix(segment, "/") ||
			strings.HasPrefix(segment, ":") {
			join = ""
		}

		if strings.HasPrefix(segment, "//") && !strings.HasPrefix(segment, "://") {
			segment = strings.TrimPrefix(segment, "//")
		}
		if strings.HasSuffix(segment, "//") && !strings.HasSuffix(segment, "://") {
			segment = strings.TrimSuffix(segment, "//")
		}

		if strings.HasSuffix(address, "/") && strings.HasPrefix(segment, "/") {
			segment = strings.TrimPrefix(segment, "/")
		}

		address = strings.Join([]string{address, join, segment}, "")
	}

	result, err := url.Parse(address)
	if err != nil {
		return nil, err
	}

	return result, nil
}

func CleanUrlSuffixAndPrefix(url string) string {
	url = strings.TrimPrefix(url, "/")
	url = strings.TrimSuffix(url, "/")

	return url
}
