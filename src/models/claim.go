package models

import "github.com/Parallels/prl-devops-service/errors"

type ClaimRequest struct {
	Name string `json:"name"`
}

func (r *ClaimRequest) Validate() error {
	if r.Name == "" {
		return errors.NewWithCode("Claim name cannot be empty", 400)
	}

	return nil
}

type ClaimResponse struct {
	ID          string    `json:"id,omitempty"`
	Name        string    `json:"name"`
	Description string    `json:"description,omitempty"`
	Internal    bool      `json:"internal"`
	Group       string    `json:"group,omitempty"`
	Resource    string    `json:"resource,omitempty"`
	Action      string    `json:"action,omitempty"`
	Users       []ApiUser `json:"users,omitempty"`
}

// ClaimGroupResourceResponse represents a single resource row in the matrix,
// containing all claims that belong to it.
type ClaimGroupResourceResponse struct {
	Resource string          `json:"resource"`
	Claims   []ClaimResponse `json:"claims"`
}

// ClaimGroupResponse represents one group (matrix section) with its resources.
type ClaimGroupResponse struct {
	Group     string                       `json:"group"`
	Resources []ClaimGroupResourceResponse `json:"resources"`
}
