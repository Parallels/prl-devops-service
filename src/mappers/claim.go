package mappers

import (
	"strings"

	"github.com/Parallels/prl-devops-service/constants"
	data_models "github.com/Parallels/prl-devops-service/data/models"
	"github.com/Parallels/prl-devops-service/helpers"
	"github.com/Parallels/prl-devops-service/models"
)

func DtoClaimToApi(model data_models.Claim) models.ClaimResponse {
	claim := models.ClaimResponse{
		ID:          model.ID,
		Name:        model.Name,
		Description: model.Description,
		Internal:    model.Internal,
		Group:       model.Group,
		Resource:    model.Resource,
		Action:      model.Action,
	}

	for _, user := range model.Users {
		claim.Users = append(claim.Users, models.ApiUser{
			ID:       user.ID,
			Username: user.Username,
			Name:     user.Name,
			Email:    user.Email,
		})
	}

	return claim
}

func DtoClaimsToApi(m []data_models.Claim) []models.ClaimResponse {
	var claims []models.ClaimResponse
	for _, model := range m {
		claims = append(claims, DtoClaimToApi(model))
	}

	return claims
}

func ApiClaimToDto(model models.ClaimRequest) data_models.Claim {
	return data_models.Claim{
		ID:   strings.ToUpper(helpers.NormalizeString(model.Name)),
		Name: model.Name,
	}
}

func ApiClaimsToDto(m []models.ClaimRequest) []data_models.Claim {
	var claims []data_models.Claim
	for _, model := range m {
		claims = append(claims, ApiClaimToDto(model))
	}

	return claims
}

// DtoClaimsToGrouped builds the matrix-oriented response from a flat claim list.
// Groups are emitted in the order defined by constants.ClaimGroupOrder; any
// group not in that list is appended alphabetically afterwards.
func DtoClaimsToGrouped(dtos []data_models.Claim) []models.ClaimGroupResponse {
	// group name → resource name → claims
	type resourceMap map[string][]models.ClaimResponse
	groupMap := map[string]resourceMap{}

	for _, c := range dtos {
		grp := c.Group
		if grp == "" {
			grp = constants.ClaimGroupCustom
		}
		res := c.Resource
		if res == "" {
			res = c.Name
		}
		if groupMap[grp] == nil {
			groupMap[grp] = resourceMap{}
		}
		groupMap[grp][res] = append(groupMap[grp][res], DtoClaimToApi(c))
	}

	// Collect groups in canonical order, then any extras alphabetically.
	seen := map[string]bool{}
	var orderedGroups []string
	for _, g := range constants.ClaimGroupOrder {
		if _, ok := groupMap[g]; ok {
			orderedGroups = append(orderedGroups, g)
			seen[g] = true
		}
	}
	for g := range groupMap {
		if !seen[g] {
			orderedGroups = append(orderedGroups, g)
		}
	}

	result := make([]models.ClaimGroupResponse, 0, len(orderedGroups))
	for _, grp := range orderedGroups {
		resMap := groupMap[grp]
		resources := make([]models.ClaimGroupResourceResponse, 0, len(resMap))
		for res, claims := range resMap {
			resources = append(resources, models.ClaimGroupResourceResponse{
				Resource: res,
				Claims:   claims,
			})
		}
		result = append(result, models.ClaimGroupResponse{
			Group:     grp,
			Resources: resources,
		})
	}
	return result
}
