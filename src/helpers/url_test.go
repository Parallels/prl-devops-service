package helpers

import (
	"net"
	"testing"
)

func TestValidateUrl(t *testing.T) {
	tests := []struct {
		name        string
		url         string
		expectError bool
	}{
		{name: "valid http url", url: "http://example.com", expectError: false},
		{name: "valid https url", url: "https://example.com", expectError: false},
		{name: "valid url with path", url: "https://example.com/api/v1", expectError: false},
		{name: "valid url with port", url: "https://example.com:8080", expectError: false},
		{name: "empty url", url: "", expectError: true},
		{name: "invalid url format", url: "not a url", expectError: true},
		{name: "unsupported protocol", url: "ftp://example.com", expectError: true},
		{name: "url without hostname", url: "http://", expectError: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateUrl(tt.url)
			if tt.expectError && err == nil {
				t.Errorf("expected error, got nil")
			}
			if !tt.expectError && err != nil {
				t.Errorf("unexpected error: %v", err)
			}
		})
	}
}

func TestIsUrlAllowed(t *testing.T) {
	tests := []struct {
		name        string
		url         string
		expectError bool
		needDNS     bool
	}{
		{name: "valid public url", url: "https://catalog.example.com", expectError: false, needDNS: true},
		{name: "localhost url", url: "http://127.0.0.1", expectError: true, needDNS: false},
		{name: "localhost hostname", url: "http://localhost", expectError: true, needDNS: false},
		{name: "private ip 10.x.x.x", url: "http://10.0.0.1", expectError: true, needDNS: false},
		{name: "private ip 172.16.x.x", url: "http://172.16.0.1", expectError: true, needDNS: false},
		{name: "private ip 192.168.x.x", url: "http://192.168.1.1", expectError: true, needDNS: false},
		{name: "link local ip", url: "http://169.254.169.254", expectError: true, needDNS: false},
		{name: "cloud metadata google", url: "http://metadata.google.internal", expectError: true, needDNS: false},
		{name: "url without scheme", url: "example.com", expectError: true, needDNS: false},
	}

	for _, tt := range tests {
		if tt.needDNS {
			continue
		}
		t.Run(tt.name, func(t *testing.T) {
			err := IsUrlAllowed(tt.url)
			if tt.expectError && err == nil {
				t.Errorf("expected error, got nil")
			}
			if !tt.expectError && err != nil {
				t.Errorf("unexpected error: %v", err)
			}
		})
	}
}

func TestIsForbiddenIP(t *testing.T) {
	tests := []struct {
		name string
		ip   string
		want bool
	}{
		{name: "loopback", ip: "127.0.0.1", want: true},
		{name: "private 10.x", ip: "10.0.0.1", want: true},
		{name: "private 172.16.x", ip: "172.16.0.1", want: true},
		{name: "private 192.168.x", ip: "192.168.1.1", want: true},
		{name: "link local", ip: "169.254.169.254", want: true},
		{name: "public ip", ip: "8.8.8.8", want: false},
		{name: "cloud metadata", ip: "169.254.169.254", want: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ip := net.ParseIP(tt.ip)
			got := isForbiddenIP(ip)
			if got != tt.want {
				t.Errorf("isForbiddenIP(%s) = %v, want %v", tt.ip, got, tt.want)
			}
		})
	}
}

func TestIsForbiddenHostname(t *testing.T) {
	tests := []struct {
		name     string
		hostname string
		want     bool
	}{
		{name: "localhost", hostname: "localhost", want: true},
		{name: "LOCALHOST uppercase", hostname: "LOCALHOST", want: true},
		{name: "127.0.0.1", hostname: "127.0.0.1", want: true},
		{name: "public host", hostname: "example.com", want: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := isForbiddenHostname(tt.hostname)
			if got != tt.want {
				t.Errorf("isForbiddenHostname(%s) = %v, want %v", tt.hostname, got, tt.want)
			}
		})
	}
}

func TestGetResolvedIPs(t *testing.T) {
	tests := []struct {
		name        string
		hostname    string
		expectError bool
	}{
		{name: "localhost", hostname: "localhost", expectError: false},
		{name: "empty hostname", hostname: "", expectError: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ips, err := GetResolvedIPs(tt.hostname)
			if tt.expectError && err == nil {
				t.Errorf("expected error, got nil")
			}
			if !tt.expectError && err != nil {
				t.Errorf("unexpected error: %v", err)
			}
			if !tt.expectError && len(ips) == 0 {
				t.Errorf("expected at least one IP, got none")
			}
		})
	}
}

func TestIsWhitelisted(t *testing.T) {
	whitelistedDomains = []string{"local-build.co", "example.com"}

	tests := []struct {
		name     string
		hostname string
		want     bool
	}{
		{name: "exact match local-build.co", hostname: "local-build.co", want: true},
		{name: "exact match example.com", hostname: "example.com", want: true},
		{name: "subdomain local-build.co", hostname: "devops-catalog.local-build.co", want: true},
		{name: "subdomain example.com", hostname: "api.example.com", want: true},
		{name: "different domain", hostname: "other.com", want: false},
		{name: "uppercase match", hostname: "LOCAL-BUILD.CO", want: true},
		{name: "partial match", hostname: "notlocal-build.co", want: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := isWhitelisted(tt.hostname)
			if got != tt.want {
				t.Errorf("isWhitelisted(%s) = %v, want %v", tt.hostname, got, tt.want)
			}
		})
	}

	whitelistedDomains = []string{}
}

func TestIsUrlAllowed_WithWhitelist(t *testing.T) {
	whitelistedDomains = []string{"local-build.co"}

	tests := []struct {
		name        string
		url         string
		expectError bool
	}{
		{name: "whitelisted domain should pass", url: "https://devops-catalog.local-build.co", expectError: false},
		{name: "non-whitelisted private IP should fail", url: "http://192.168.1.1", expectError: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := IsUrlAllowed(tt.url)
			if tt.expectError && err == nil {
				t.Errorf("expected error, got nil")
			}
			if !tt.expectError && err != nil {
				t.Errorf("unexpected error: %v", err)
			}
		})
	}

	whitelistedDomains = []string{}
}
