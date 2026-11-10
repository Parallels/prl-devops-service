package interfaces

// EnrollmentTokensService mirrors orchestrator_enrollment_token method signatures verbatim (PBI #421 D-2).
// Legacy implementor: data.JsonDatabase; SQL implementor: database.GormDatabase (#421 task 04+).
import (
	"github.com/Parallels/prl-devops-service/basecontext"
	"github.com/Parallels/prl-devops-service/data/models"
)

type EnrollmentTokensService interface {
	CreateEnrollmentToken(ctx basecontext.ApiContext, hostName string, ttlMinutes int) (*models.OrchestratorEnrollmentToken, error)
	GetEnrollmentToken(ctx basecontext.ApiContext, token string) (*models.OrchestratorEnrollmentToken, error)
	ValidateEnrollmentToken(ctx basecontext.ApiContext, token string) (*models.OrchestratorEnrollmentToken, error)
	MarkEnrollmentTokenUsed(ctx basecontext.ApiContext, id string) error
	DeleteExpiredEnrollmentTokens(ctx basecontext.ApiContext) error
}
