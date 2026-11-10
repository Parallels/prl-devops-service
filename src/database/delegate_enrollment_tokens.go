package database

// Generated transcription scaffold (PBI #421 task 04, D-1 hybrid delegates):
// every method of interfaces/enrollment_tokens delegates to the injected legacy handle.
// Domain tasks 05-16 REPLACE these one-liners with SQL implementations.
import (
	"github.com/Parallels/prl-devops-service/basecontext"
	"github.com/Parallels/prl-devops-service/data/models"
)

func (g *GormDatabase) CreateEnrollmentToken(ctx basecontext.ApiContext, hostName string, ttlMinutes int) (*models.OrchestratorEnrollmentToken, error) {
	return g.legacy.CreateEnrollmentToken(ctx, hostName, ttlMinutes)
}

func (g *GormDatabase) GetEnrollmentToken(ctx basecontext.ApiContext, token string) (*models.OrchestratorEnrollmentToken, error) {
	return g.legacy.GetEnrollmentToken(ctx, token)
}

func (g *GormDatabase) ValidateEnrollmentToken(ctx basecontext.ApiContext, token string) (*models.OrchestratorEnrollmentToken, error) {
	return g.legacy.ValidateEnrollmentToken(ctx, token)
}

func (g *GormDatabase) MarkEnrollmentTokenUsed(ctx basecontext.ApiContext, id string) error {
	return g.legacy.MarkEnrollmentTokenUsed(ctx, id)
}

func (g *GormDatabase) DeleteExpiredEnrollmentTokens(ctx basecontext.ApiContext) error {
	return g.legacy.DeleteExpiredEnrollmentTokens(ctx)
}
