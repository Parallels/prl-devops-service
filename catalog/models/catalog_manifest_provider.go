package models

import "strings"

type CatalogManifestProvider struct {
	Type string            `json:"type"`
	Meta map[string]string `json:"meta"`
}

func (m *CatalogManifestProvider) String() string {
	r := "provider=" + m.Type

	for k, v := range m.Meta {
		r += ";" + k + "=" + v
	}

	return strings.TrimRight(r, ";")
}
