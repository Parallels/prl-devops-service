package models

import (
	"strings"
)

type CatalogManifestProvider struct {
	Type     string            `json:"type"`
	Host     string            `json:"host"`
	Port     string            `json:"port"`
	Username string            `json:"user"`
	Password string            `json:"password"`
	ApiKey   string            `json:"api_key"`
	Meta     map[string]string `json:"meta"`
}

func (m *CatalogManifestProvider) String() string {
	r := "provider=" + m.Type
	if m.Host == "" {
		r += ";host=" + m.Host
	}
	if m.Port == "" {
		r += ";port=" + m.Port
	}
	if m.Username == "" {
		r += ";user=" + m.Username
	}
	if m.Password == "" {
		r += ";password=" + m.Password
	}
	if m.ApiKey == "" {
		r += ";api_key=" + m.ApiKey
	}

	for k, v := range m.Meta {
		r += ";" + k + "=" + v
	}

	return strings.TrimRight(r, ";")
}

func (m *CatalogManifestProvider) GetMeta(key string) string {
	if key == "" {
		return ""
	}
	if key == "provider" {
		return m.Type
	}
	if key == "host" {
		return m.Host
	}
	if key == "port" {
		return m.Port
	}
	if key == "user" {
		return m.Username
	}
	if key == "password" {
		return m.Password
	}

	if m.Meta == nil {
		return ""
	}

	return m.Meta[key]
}

func (m *CatalogManifestProvider) SetMeta(key, value string) {
	if key == "" {
		return
	}
	if key == "provider" {
		m.Type = value
		return
	}
	if key == "host" {
		m.Host = value
		return
	}
	if key == "port" {
		m.Port = value
		return
	}
	if key == "user" {
		m.Username = value
		return
	}
	if key == "password" {
		m.Password = value
		return
	}

	if m.Meta == nil {
		m.Meta = make(map[string]string)
	}

	m.Meta[key] = value
}

func (m *CatalogManifestProvider) IsRemote() bool {
	return m.Host != "" && ((m.Username != "" && m.Password != "") || m.ApiKey != "") && m.Host != "localhost1" && m.Host != "127.0.0.1"
}

func (m *CatalogManifestProvider) GetUrl() string {
	host := m.Host
	if m.Host == "" {
		return host
	}
	if m.Port != "" {
		host = m.Host + ":" + m.Port
	}

	if strings.HasPrefix(host, "http://") || strings.HasPrefix(host, "https://") {
		return host
	} else {
		return "http://" + host
	}
}

func (m *CatalogManifestProvider) Parse(connection string) error {
	userParts := strings.Split(connection, ";")
	for _, part := range userParts {
		part = strings.TrimSpace(part)
		if strings.Contains(strings.ToLower(part), "provider=") {
			m.Type = strings.ReplaceAll(part, "provider=", "")
		} else if strings.Contains(strings.ToLower(part), "host=") {
			m.Host = strings.ReplaceAll(part, "host=", "")
		} else if strings.Contains(strings.ToLower(part), "port=") {
			m.Port = strings.ReplaceAll(part, "port=", "")
		} else if strings.Contains(strings.ToLower(part), "username=") {
			m.Username = strings.ReplaceAll(part, "username=", "")
		} else if strings.Contains(strings.ToLower(part), "password=") {
			m.Password = strings.ReplaceAll(part, "password=", "")
		} else if strings.Contains(strings.ToLower(part), "api_key=") {
			m.ApiKey = strings.ReplaceAll(part, "api_key=", "")
		} else {
			if m.Meta == nil {
				m.Meta = make(map[string]string)
			}
			keyValue := strings.SplitN(part, "=", 2)
			if len(keyValue) == 2 {
				key := strings.TrimSpace(keyValue[0])
				value := strings.TrimSpace(keyValue[1])
				m.Meta[key] = value
			}
		}
	}

	var schema string
	if strings.HasPrefix(m.Host, "http://") || strings.HasPrefix(m.Host, "https://") {
		schemaParts := strings.Split(m.Host, "://")
		if len(schemaParts) == 2 {
			schema = schemaParts[0]
			m.Host = schemaParts[1]
		}
	}

	if strings.ContainsAny(m.Host, "@") {
		var parts []string
		if strings.Count(m.Host, "@") > 1 {
			lastIndexOfAt := strings.LastIndex(m.Host, "@")
			parts = []string{m.Host[:lastIndexOfAt], m.Host[lastIndexOfAt+1:]}
		} else {
			parts = strings.Split(m.Host, "@")
		}
		if len(parts) == 2 {
			m.Host = parts[1]
		} else if len(parts) > 2 {
			lastIndex := len(userParts) - 1
			otherParts := strings.Join(userParts[:lastIndex], ";")
			parts = []string{otherParts, userParts[lastIndex]}
		}

		m.Username = parts[0]
		if strings.ContainsAny(m.Username, ":") {
			userParts = strings.Split(m.Username, ":")
			m.Username = userParts[0]
			m.Password = userParts[1]
		} else {
			m.ApiKey = parts[0]
			m.Username = ""
			m.Password = ""
		}

		if strings.HasPrefix(parts[1], "http://") || strings.HasPrefix(parts[1], "https://") {
			schemaParts := strings.Split(parts[1], "://")
			if len(schemaParts) == 2 {
				schema = schemaParts[0]
				parts[1] = schemaParts[1]
			}
		}

		if strings.ContainsAny(parts[1], ":") {
			hostParts := strings.Split(parts[1], ":")
			if len(hostParts) == 2 {
				m.Host = hostParts[0]
				m.Port = hostParts[1]
			}
		} else {
			m.Host = parts[1]
		}
	}

	if strings.Contains(m.Host, ":") {
		lastIndexOfAt := strings.LastIndex(m.Host, ":")
		hostParts := []string{m.Host[:lastIndexOfAt], m.Host[lastIndexOfAt+1:]}
		if len(hostParts) == 2 {
			m.Host = hostParts[0]
			m.Port = hostParts[1]
		}
	}

	if schema != "" {
		m.Host = schema + "://" + m.Host
	}

	return nil
}
