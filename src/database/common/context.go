package common

import (
	"time"

	"github.com/Parallels/prl-devops-service/basecontext"
	"github.com/google/uuid"
)

func GetBaseModelFromContext(ctx *basecontext.BaseContext, baseModel *BaseModel) *BaseModel {
	if baseModel == nil {
		baseModel = &BaseModel{}
	}
	baseModel.ID = uuid.New().String()
	baseModel.CreatedAt = time.Now()
	baseModel.UpdatedAt = time.Now()
	userID := ""
	if ctx != nil && ctx.User.ID != "" {
		userID = ctx.User.ID
	}
	baseModel.CreatedBy = userID
	baseModel.UpdatedBy = userID
	return baseModel
}

func GetTenantBaseModelFromContext(ctx *basecontext.BaseContext, baseModel *BaseModelWithTenant) *BaseModelWithTenant {
	if baseModel == nil {
		baseModel = &BaseModelWithTenant{}
	}
	baseModel.ID = uuid.New().String()
	baseModel.CreatedAt = time.Now()
	baseModel.UpdatedAt = time.Now()
	userID := ""
	tenantID := ""
	if ctx != nil && ctx.User.ID != "" {
		userID = ctx.User.ID
		// Tenant ID not directly available in prl-devops BaseContext
		// This can be extended when multi-tenancy is implemented
	}
	baseModel.TenantID = tenantID
	baseModel.CreatedBy = userID
	baseModel.UpdatedBy = userID
	return baseModel
}
