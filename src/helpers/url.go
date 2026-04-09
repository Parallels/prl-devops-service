package helpers

import (
	"net"
	"net/url"
	"os"
	"strings"

	"github.com/Parallels/prl-devops-service/constants"
	"github.com/Parallels/prl-devops-service/errors"
)

var (
	forbiddenIPs = []string{
		"127.0.0.1",
		"::1",
		"0.0.0.0",
		"::",
		"169.254.169.254",
		"metadata.google.internal",
		"metadata.azure.com",
		"169.254.169.253",
	}

	forbiddenCIDRs = []*net.IPNet{}
)

func init() {
	for _, cidr := range []string{
		"10.0.0.0/8", "172.16.0.0/12", "192.168.0.0/16", "100.64.0.0/10", "169.254.0.0/16",
		"192.0.0.0/24", "192.0.2.0/24", "198.51.100.0/24", "203.0.113.0/24", "224.0.0.0/4",
		"240.0.0.0/4", "198.18.0.0/15", "192.88.99.0/24", "fc00::/7", "fe80::/10", "ff00::/8",
	} {
		_, network, err := net.ParseCIDR(cidr)
		if err == nil {
			forbiddenCIDRs = append(forbiddenCIDRs, network)
		}
	}
}

func ValidateUrl(urlStr string) error {
	if urlStr == "" {
		return errors.New("url is required")
	}

	parsedURL, err := url.Parse(urlStr)
	if err != nil {
		return errors.New("invalid url format")
	}

	if parsedURL.Scheme != "http" && parsedURL.Scheme != "https" {
		return errors.New("only http and https protocols are allowed")
	}

	if parsedURL.Host == "" {
		return errors.New("url must contain a hostname")
	}

	return nil
}

func IsUrlAllowed(urlStr string) error {
	if err := ValidateUrl(urlStr); err != nil {
		return err
	}

	parsedURL, err := url.Parse(urlStr)
	if err != nil {
		return errors.New("invalid url format")
	}

	host := parsedURL.Hostname()

	if isForbiddenHostname(host) {
		return errors.New("access to this url is not allowed")
	}

	ips, err := net.LookupIP(host)
	if err != nil {
		return errors.New("unable to resolve hostname")
	}

	for _, ip := range ips {
		if isForbiddenIP(ip) {
			return errors.New("access to this url is not allowed")
		}
	}

	return nil
}

func isForbiddenHostname(hostname string) bool {
	hostname = strings.ToLower(strings.TrimSpace(hostname))

	for _, forbidden := range forbiddenIPs {
		if strings.EqualFold(hostname, forbidden) {
			return true
		}
	}

	if strings.EqualFold(hostname, "localhost") {
		return true
	}

	return false
}

func isForbiddenIP(ip net.IP) bool {
	if ip == nil {
		return true
	}

	if ip.IsLoopback() {
		return true
	}

	if ip.IsLinkLocalUnicast() || ip.IsLinkLocalMulticast() {
		return true
	}

	if ip.IsUnspecified() {
		return true
	}

	ip4 := ip.To4()
	if ip4 != nil {
		if ip4[0] == 127 {
			return true
		}
		if ip4[0] == 10 {
			return true
		}
		if ip4[0] == 172 && ip4[1] >= 16 && ip4[1] <= 31 {
			return true
		}
		if ip4[0] == 192 && ip4[1] == 168 {
			return true
		}
		if ip4[0] == 169 && ip4[1] == 254 {
			return true
		}
		if ip4.IsMulticast() {
			return true
		}

		for _, cidr := range forbiddenCIDRs {
			if cidr.IP.To4() != nil && cidr.Contains(ip) {
				return true
			}
		}
	} else {
		for _, cidr := range forbiddenCIDRs {
			if cidr.IP.To4() == nil && cidr.Contains(ip) {
				return true
			}
		}
	}

	return false
}

func GetResolvedIPs(hostname string) ([]net.IP, error) {
	if hostname == "" {
		return nil, errors.New("hostname is required")
	}

	ips, err := net.LookupIP(hostname)
	if err != nil {
		return nil, err
	}

	if len(ips) == 0 {
		return nil, errors.New("no ip addresses found for hostname")
	}

	return ips, nil
}

func ValidateCatalogManagerUrl(managerURL string) error {
	if isUrlValidationDisabled() {
		return nil
	}

	if err := IsUrlAllowed(managerURL); err != nil {
		return err
	}

	return nil
}

func isUrlValidationDisabled() bool {
	disable := os.Getenv(constants.DISABLE_URL_VALIDATION_ENV_VAR)
	return strings.ToLower(strings.TrimSpace(disable)) == "true"
}
