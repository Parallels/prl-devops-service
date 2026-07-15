package data

import (
	"strings"
	"time"

	"github.com/Parallels/prl-devops-service/basecontext"
	"github.com/Parallels/prl-devops-service/data/models"
	"github.com/Parallels/prl-devops-service/errors"
	"github.com/Parallels/prl-devops-service/helpers"
	"github.com/google/uuid"
)

var (
	ErrEnrollmentTokenNotFound = errors.NewWithCode("enrollment token not found", 404)
	ErrEnrollmentTokenExpired  = errors.NewWithCode("enrollment token has expired", 401)
	ErrEnrollmentTokenUsed     = errors.NewWithCode("enrollment token has already been used", 401)
)

func (j *JsonDatabase) CreateEnrollmentToken(ctx basecontext.ApiContext, hostName string, ttlMinutes int) (*models.OrchestratorEnrollmentToken, error) {
	if !j.IsConnected() {
		return nil, ErrDatabaseNotConnected
	}
	if ttlMinutes <= 0 {
		ttlMinutes = 15
	}

	token := &models.OrchestratorEnrollmentToken{
		ID:        uuid.New().String(),
		Token:     uuid.New().String() + uuid.New().String(), // 72-char random token
		HostName:  hostName,
		Used:      false,
		ExpiresAt: time.Now().UTC().Add(time.Duration(ttlMinutes) * time.Minute).Format(time.RFC3339),
		CreatedAt: helpers.GetUtcCurrentDateTime(),
		DbRecord:  &models.DbRecord{},
	}

	j.dataMutex.Lock()
	j.data.EnrollmentTokens = append(j.data.EnrollmentTokens, *token)
	j.dataMutex.Unlock()

	return token, nil
}

func (j *JsonDatabase) GetEnrollmentToken(ctx basecontext.ApiContext, token string) (*models.OrchestratorEnrollmentToken, error) {
	if !j.IsConnected() {
		return nil, ErrDatabaseNotConnected
	}

	j.dataMutex.RLock()
	defer j.dataMutex.RUnlock()

	for _, t := range j.data.EnrollmentTokens {
		if strings.EqualFold(t.Token, token) {
			copy := t
			return &copy, nil
		}
	}

	return nil, ErrEnrollmentTokenNotFound
}

// ValidateEnrollmentToken checks that the token exists, is not used, and is not expired.
// Returns the token record on success, or an error describing the failure.
func (j *JsonDatabase) ValidateEnrollmentToken(ctx basecontext.ApiContext, token string) (*models.OrchestratorEnrollmentToken, error) {
	if !j.IsConnected() {
		return nil, ErrDatabaseNotConnected
	}

	t, err := j.GetEnrollmentToken(ctx, token)
	if err != nil {
		return nil, err
	}

	if t.Used {
		return nil, ErrEnrollmentTokenUsed
	}

	if t.ExpiresAt != "" {
		exp, err := time.Parse(time.RFC3339, t.ExpiresAt)
		if err == nil && time.Now().UTC().After(exp) {
			return nil, ErrEnrollmentTokenExpired
		}
	}

	return t, nil
}

func (j *JsonDatabase) MarkEnrollmentTokenUsed(ctx basecontext.ApiContext, id string) error {
	if !j.IsConnected() {
		return ErrDatabaseNotConnected
	}

	j.dataMutex.Lock()
	defer j.dataMutex.Unlock()

	for i, t := range j.data.EnrollmentTokens {
		if t.ID == id {
			for {
				if IsRecordLocked(j.data.EnrollmentTokens[i].DbRecord) {
					continue
				}
				LockRecord(ctx, j.data.EnrollmentTokens[i].DbRecord)
				j.data.EnrollmentTokens[i].Used = true
				UnlockRecord(ctx, j.data.EnrollmentTokens[i].DbRecord)
				break
			}
			return nil
		}
	}

	return ErrEnrollmentTokenNotFound
}

// DeleteExpiredEnrollmentTokens removes tokens that have expired or been used.
func (j *JsonDatabase) DeleteExpiredEnrollmentTokens(ctx basecontext.ApiContext) error {
	if !j.IsConnected() {
		return ErrDatabaseNotConnected
	}

	now := time.Now().UTC()

	j.dataMutex.Lock()
	defer j.dataMutex.Unlock()

	active := j.data.EnrollmentTokens[:0]
	for _, t := range j.data.EnrollmentTokens {
		if t.Used {
			continue
		}
		if t.ExpiresAt != "" {
			exp, err := time.Parse(time.RFC3339, t.ExpiresAt)
			if err == nil && now.After(exp) {
				continue
			}
		}
		active = append(active, t)
	}

	j.data.EnrollmentTokens = active
	return nil
}
