package models

import (
	"github.com/Parallels/prl-devops-service/errors"
)

type UserConfigValueType string

const (
	UserConfigValueTypeString UserConfigValueType = "string"
	UserConfigValueTypeBool   UserConfigValueType = "bool"
	UserConfigValueTypeInt    UserConfigValueType = "int"
	UserConfigValueTypeJson   UserConfigValueType = "json"
)

type UserConfigRequest struct {
	Slug  string              `json:"slug"`
	Name  string              `json:"name"`
	Type  UserConfigValueType `json:"type"`
	Value string              `json:"value"`
}

func (r *UserConfigRequest) Validate() error {
	if r.Slug == "" {
		return errors.NewWithCode("slug is required", 400)
	}
	if r.Name == "" {
		return errors.NewWithCode("name is required", 400)
	}
	if r.Type == "" {
		r.Type = UserConfigValueTypeString
	}

	switch r.Type {
	case UserConfigValueTypeString, UserConfigValueTypeBool, UserConfigValueTypeInt, UserConfigValueTypeJson:
	default:
		return errors.NewWithCode("type must be one of: string, bool, int, json", 400)
	}

	return nil
}

type UserConfigUpdateRequest struct {
	Name  string              `json:"name,omitempty"`
	Type  UserConfigValueType `json:"type,omitempty"`
	Value string              `json:"value,omitempty"`
}

func (r *UserConfigUpdateRequest) Validate() error {
	if r.Type != "" {
		switch r.Type {
		case UserConfigValueTypeString, UserConfigValueTypeBool, UserConfigValueTypeInt, UserConfigValueTypeJson:
		default:
			return errors.NewWithCode("type must be one of: string, bool, int, json", 400)
		}
	}

	return nil
}

type UserConfigResponse struct {
	ID        string              `json:"id"`
	UserID    string              `json:"user_id"`
	Slug      string              `json:"slug"`
	Name      string              `json:"name"`
	Type      UserConfigValueType `json:"type"`
	Value     string              `json:"value"`
	CreatedAt string              `json:"created_at,omitempty"`
	UpdatedAt string              `json:"updated_at,omitempty"`
}
