package orchestrator

import (
	"errors"
	"time"

	"github.com/Parallels/prl-devops-service/basecontext"
	data_models "github.com/Parallels/prl-devops-service/data/models"
	"github.com/Parallels/prl-devops-service/orchestrator/registry"
	"github.com/Parallels/prl-devops-service/serviceprovider"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// Job links, durable.
//
// A host job is created by whichever replica dispatched it, but every replica
// dials every host's WebSocket — so the completion event routinely arrives at a
// replica that never saw the dispatch. If the only record of the link is one
// process's map, that event is forwarded raw, the orchestrator job is never
// marked terminal, and the capacity its create reserved stays charged until the
// lease expires. Persisting the link is what turns "the other replica will
// handle it" from a hope into a mechanism.
//
// The in-memory map stays authoritative for the hot path (see
// registry.JobLinkRegistry); this is the fallback underneath it.
func (s *OrchestratorService) wireJobLinks(ctx basecontext.ApiContext) {
	db, err := serviceprovider.GetSQLDatabase(ctx)
	if err != nil {
		// Legacy JSON data layer: no table to fall back to, so links stay exactly
		// as process-local as they were before. Logged once, not per event.
		s.ctx.LogWarnf("[Orchestrator] Job links stay in-process without a SQL database: %v", err)
		return
	}

	registry.Get().SetDurableStore(
		func(hostJobID string, link registry.JobLink) error {
			if hostJobID == "" || link.OrchestratorJobID == "" {
				return nil
			}

			// Upsert rather than insert: a retried dispatch re-registers the same
			// host job, and a bookkeeping row must never be the reason a create
			// returns an error.
			return db.Clauses(clause.OnConflict{UpdateAll: true}).
				Create(&data_models.JobLink{
					HostJobID:         hostJobID,
					OrchestratorJobID: link.OrchestratorJobID,
					HostID:            link.HostID,
					CreatedAt:         time.Now().UTC().Format(time.RFC3339),
				}).Error
		},
		func(hostJobID string) (registry.JobLink, bool, error) {
			var row data_models.JobLink
			if err := db.Where("host_job_id = ?", hostJobID).Take(&row).Error; err != nil {
				if errors.Is(err, gorm.ErrRecordNotFound) {
					return registry.JobLink{}, false, nil
				}

				s.ctx.LogWarnf("[Orchestrator] Job link lookup failed for host job %s: %v", hostJobID, err)
				return registry.JobLink{}, false, err
			}

			return registry.JobLink{OrchestratorJobID: row.OrchestratorJobID, HostID: row.HostID}, true, nil
		},
		func(hostJobID string) error {
			return db.Where("host_job_id = ?", hostJobID).Delete(&data_models.JobLink{}).Error
		},
	)
}
