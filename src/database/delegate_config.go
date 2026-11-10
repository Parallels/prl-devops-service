package database

// Generated transcription scaffold (PBI #421 task 04, D-1 hybrid delegates):
// every method of interfaces/config delegates to the injected legacy handle.
// Domain tasks 05-16 REPLACE these one-liners with SQL implementations.
import (
	"github.com/Parallels/prl-devops-service/basecontext"
	"github.com/Parallels/prl-devops-service/data/models"
)

func (g *GormDatabase) GetConfiguration(ctx basecontext.ApiContext) (*models.Configuration, error) {
	return g.legacy.GetConfiguration(ctx)
}

func (g *GormDatabase) GetId(ctx basecontext.ApiContext) (string, error) {
	return g.legacy.GetId(ctx)
}

func (g *GormDatabase) SeedId(ctx basecontext.ApiContext) (string, error) {
	return g.legacy.SeedId(ctx)
}

func (g *GormDatabase) SetId(ctx basecontext.ApiContext, id string) (string, error) {
	return g.legacy.SetId(ctx, id)
}
