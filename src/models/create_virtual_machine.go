package models

import (
	"os"
	"strings"

	"github.com/Parallels/prl-devops-service/constants"
	"github.com/Parallels/prl-devops-service/errors"
)

type CreateVirtualMachineRequest struct {
	Name          string   `json:"name"`
	Owner         string   `json:"owner,omitempty"`
	Architecture  string   `json:"architecture,omitempty"`
	SelectionTags []string `json:"selection_tags,omitempty"`
	// TargetTags is a hard placement constraint: when present, only hosts that
	// carry ALL the listed host tags (case-insensitive exact match) are
	// eligible for placement and the pre-flight check. Absent or empty means
	// no tag constraint. Unlike SelectionTags (legacy OR semantics: a host
	// matching at least one tag survives), every entry must match.
	TargetTags      []string                            `json:"target_tags,omitempty"`
	PackerTemplate  *CreatePackerVirtualMachineRequest  `json:"packer_template,omitempty"`
	VagrantBox      *CreateVagrantMachineRequest        `json:"vagrant_box,omitempty"`
	CatalogManifest *CreateCatalogVirtualMachineRequest `json:"catalog_manifest,omitempty"`
	StartOnCreate   bool                                `json:"start_on_create,omitempty"`
}

func (r *CreateVirtualMachineRequest) Validate() error {
	if r.Name == "" {
		return errors.New("Name cannot be empty")
	}

	if r.Owner == "" {
		r.Owner = os.Getenv(constants.CURRENT_USER_ENV_VAR)
	}

	if r.Architecture == "" {
		return errors.New("Architecture cannot be empty")
	}

	targetTags, err := NormalizeTagConstraints(r.TargetTags)
	if err != nil {
		return err
	}
	r.TargetTags = targetTags

	if r.PackerTemplate != nil {
		if r.VagrantBox != nil || r.CatalogManifest != nil {
			return errors.New("Only one of packer_template, vagrant_box or catalog_manifest can be specified")
		}
		r.PackerTemplate.Name = r.Name
		r.PackerTemplate.Owner = r.Owner
		r.Architecture = r.PackerTemplate.Architecture
		return r.PackerTemplate.Validate()
	}

	if r.VagrantBox != nil {
		if r.PackerTemplate != nil || r.CatalogManifest != nil {
			return errors.New("Only one of packer_template, vagrant_box or catalog_manifest can be specified")
		}
		r.VagrantBox.Name = r.Name
		r.VagrantBox.Owner = r.Owner
		r.VagrantBox.Architecture = r.Architecture
		return r.VagrantBox.Validate()
	}

	if r.CatalogManifest != nil {
		if r.PackerTemplate != nil || r.VagrantBox != nil {
			return errors.New("Only one of packer_template, vagrant_box or catalog_manifest can be specified")
		}

		r.CatalogManifest.MachineName = r.Name
		r.CatalogManifest.Owner = r.Owner
		r.CatalogManifest.Architecture = r.Architecture
		return r.CatalogManifest.Validate()
	}

	return nil
}

// NormalizeTagConstraints cleans a tag-constraint list (target_tags) coming from
// an API request body: entries are trimmed, an entry that is empty after
// trimming is rejected with a 400, duplicates are dropped case-insensitively
// keeping the first spelling and the original order, and a list left with no
// entries normalizes to nil — so "no constraint" and "match nothing" can never
// be confused, and semantically identical requests stay byte-stable.
func NormalizeTagConstraints(tags []string) ([]string, error) {
	normalized := make([]string, 0, len(tags))
	seen := make(map[string]bool, len(tags))
	for _, tag := range tags {
		tag = strings.TrimSpace(tag)
		if tag == "" {
			return nil, errors.NewWithCode("target_tags cannot contain empty values", 400)
		}
		key := strings.ToLower(tag)
		if seen[key] {
			continue
		}
		seen[key] = true
		normalized = append(normalized, tag)
	}
	if len(normalized) == 0 {
		return nil, nil
	}
	return normalized, nil
}

type CreateVirtualMachineResponse struct {
	ID           string `json:"id"`
	Host         string `json:"host,omitempty"`
	Name         string `json:"name"`
	Owner        string `json:"owner"`
	CurrentState string `json:"current_state"`
}
