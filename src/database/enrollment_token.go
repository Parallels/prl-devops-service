package database

// Enrollment tokens on GORM (PBI #421 task 16). Legacy
// src/data/orchestrator_enrollment_token.go mirrored: Create mints uuid id
// + 72-char uuid-pair token, ttlMinutes<=0 → 15, ExpiresAt in RFC3339
// (SECOND precision — distinct from the Nano CreatedAt); Get matches the
// TOKEN VALUE case-insensitively and the ID never matches; Validate checks
// used BEFORE expiry (:75-84) and an unparseable ExpiresAt counts as valid
// (:81 err-skip); Mark matches the id EXACTLY (:98) and is idempotent
// (re-mark → nil); Sweep prunes used-always (regardless of expiry!) plus
// expired, keeping used=fresh and unparseable stamps (:126-138). The
// legacy DbRecord lock-spin in Mark is an in-memory save-coherency device
// — the durable store has no analogue. Expiry is evaluated on read; the
// sweep cadence is API-side (startup/main.go:130), NOT engine-side —
// nothing self-starts here (recorded for task 18). The clock rides the
// shared nowFn seam introduced in task 15.

import (
	"strings"
	"time"

	"github.com/Parallels/prl-devops-service/basecontext"
	interfaces "github.com/Parallels/prl-devops-service/data/interfaces"
	"github.com/Parallels/prl-devops-service/data/models"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

const enrollmentTokenTable = "orchestrator_enrollment_tokens"

func (g *GormDatabase) enrollmentTokenRows() ([]models.OrchestratorEnrollmentToken, error) {
	var rows []models.OrchestratorEnrollmentToken
	if err := g.gormDB.Table(enrollmentTokenTable).Order("rowid").Find(&rows).Error; err != nil {
		return nil, mapGormError(err, nil, nil)
	}
	return rows, nil
}

func (g *GormDatabase) CreateEnrollmentToken(ctx basecontext.ApiContext, hostName string, ttlMinutes int) (*models.OrchestratorEnrollmentToken, error) {
	if !g.IsConnected() {
		return nil, interfaces.ErrDatabaseNotConnected
	}
	if ttlMinutes <= 0 {
		ttlMinutes = 15
	}
	token := &models.OrchestratorEnrollmentToken{
		ID:        uuid.New().String(),
		Token:     uuid.New().String() + uuid.New().String(), // 72-char random token
		HostName:  hostName,
		Used:      false,
		ExpiresAt: g.jobsNowTime().UTC().Add(time.Duration(ttlMinutes) * time.Minute).Format(time.RFC3339),
		CreatedAt: g.jobsStamp(),
	}
	// No dup pre-check (legacy has none — uuid space); the #419 unique
	// index on token is the backstop, mapped through D-11.
	if err := mapGormError(g.gormDB.Table(enrollmentTokenTable).Create(token).Error, nil, nil); err != nil {
		return nil, err
	}
	return token, nil
}

func (g *GormDatabase) GetEnrollmentToken(ctx basecontext.ApiContext, token string) (*models.OrchestratorEnrollmentToken, error) {
	if !g.IsConnected() {
		return nil, interfaces.ErrDatabaseNotConnected
	}
	rows, err := g.enrollmentTokenRows()
	if err != nil {
		return nil, err
	}
	for _, r := range rows {
		if strings.EqualFold(r.Token, token) { // token VALUE; never the id (:54)
			row := r
			return &row, nil
		}
	}
	return nil, interfaces.ErrEnrollmentTokenNotFound
}

func (g *GormDatabase) ValidateEnrollmentToken(ctx basecontext.ApiContext, token string) (*models.OrchestratorEnrollmentToken, error) {
	if !g.IsConnected() {
		return nil, interfaces.ErrDatabaseNotConnected
	}
	t, err := g.GetEnrollmentToken(ctx, token)
	if err != nil {
		return nil, err
	}
	if t.Used { // used BEFORE expiry (:75)
		return nil, interfaces.ErrEnrollmentTokenUsed
	}
	if t.ExpiresAt != "" {
		exp, perr := time.Parse(time.RFC3339, t.ExpiresAt)
		if perr == nil && g.jobsNowTime().UTC().After(exp) { // strictly after — equal is VALID
			return nil, interfaces.ErrEnrollmentTokenExpired
		}
	}
	return t, nil
}

func (g *GormDatabase) MarkEnrollmentTokenUsed(ctx basecontext.ApiContext, id string) error {
	if !g.IsConnected() {
		return interfaces.ErrDatabaseNotConnected
	}
	res := g.gormDB.Table(enrollmentTokenTable).Where("id = ?", id).Update("used", true) // EXACT id (:98)
	if res.Error != nil {
		return mapGormError(res.Error, nil, nil)
	}
	if res.RowsAffected == 0 {
		return interfaces.ErrEnrollmentTokenNotFound
	}
	return nil // idempotent re-mark (:104 sets true again)
}

func (g *GormDatabase) DeleteExpiredEnrollmentTokens(ctx basecontext.ApiContext) error {
	if !g.IsConnected() {
		return interfaces.ErrDatabaseNotConnected
	}
	now := g.jobsNowTime().UTC() // ONE sample (:121)
	rows, err := g.enrollmentTokenRows()
	if err != nil {
		return err
	}
	var doomed []string
	for _, r := range rows {
		if r.Used { // used tokens ALWAYS go, expired or not (:128)
			doomed = append(doomed, r.ID)
			continue
		}
		if r.ExpiresAt != "" {
			exp, perr := time.Parse(time.RFC3339, r.ExpiresAt)
			if perr == nil && now.After(exp) { // unparseable stays (:133)
				doomed = append(doomed, r.ID)
			}
		}
	}
	if len(doomed) == 0 {
		return nil
	}
	return g.gormDB.Transaction(func(tx *gorm.DB) error { // whole file held one lock — one tx here
		for _, id := range doomed {
			if err := mapGormError(tx.Table(enrollmentTokenTable).Where("id = ?", id).
				Delete(&models.OrchestratorEnrollmentToken{}).Error, nil, nil); err != nil {
				return err
			}
		}
		return nil
	})
}
