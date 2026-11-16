package mappers

import (
	"github.com/Parallels/prl-devops-service/data/models"
	apimodels "github.com/Parallels/prl-devops-service/models"
)

// SettingDefinitionToResponse maps one catalogued key definition.
func SettingDefinitionToResponse(def models.SettingDefinition) apimodels.SettingDefinitionResponse {
	return apimodels.SettingDefinitionResponse{
		Key:             def.Key,
		Type:            string(def.Type),
		DefaultValue:    def.DefaultValue,
		ValidValues:     def.ValidValues,
		Description:     def.Description,
		AppliesTo:       def.AppliesTo,
		RequiresRestart: def.RequiresRestart,
	}
}

// SettingDefinitionsToResponse maps the definitions catalog.
func SettingDefinitionsToResponse(defs []models.SettingDefinition) []apimodels.SettingDefinitionResponse {
	out := make([]apimodels.SettingDefinitionResponse, 0, len(defs))
	for _, def := range defs {
		out = append(out, SettingDefinitionToResponse(def))
	}
	return out
}

// SettingValueToResponse maps one stored scoped override.
func SettingValueToResponse(value models.SettingValue) apimodels.SettingValueResponse {
	return apimodels.SettingValueResponse{
		ID:        value.ID,
		Scope:     value.Scope,
		Key:       value.Key,
		Value:     value.Value,
		UpdatedBy: value.UpdatedBy,
		CreatedAt: value.CreatedAt,
		UpdatedAt: value.UpdatedAt,
	}
}

// SettingValuesToResponse maps scoped override rows.
func SettingValuesToResponse(values []models.SettingValue) []apimodels.SettingValueResponse {
	out := make([]apimodels.SettingValueResponse, 0, len(values))
	for _, value := range values {
		out = append(out, SettingValueToResponse(value))
	}
	return out
}

// SettingSyncStatesToResponse maps per-host drift rows.
func SettingSyncStatesToResponse(states []models.SettingSyncState) []apimodels.SettingSyncStateResponse {
	out := make([]apimodels.SettingSyncStateResponse, 0, len(states))
	for _, state := range states {
		out = append(out, SettingSyncStateToResponse(state, ""))
	}
	return out
}

// SettingSyncStateToResponse maps one drift row; hostName may be filled by
// the caller joining the hosts table.
func SettingSyncStateToResponse(state models.SettingSyncState, hostName string) apimodels.SettingSyncStateResponse {
	status := state.Status
	if status == "" {
		status = models.SettingSyncStatusNever
	}
	return apimodels.SettingSyncStateResponse{
		HostID:       state.HostID,
		HostName:     hostName,
		Revision:     state.Revision,
		ContentHash:  state.ContentHash,
		Status:       status,
		LastError:    state.LastError,
		LastSyncedAt: state.LastSyncedAt,
	}
}
