package data

import (
	"strings"

	"github.com/Parallels/prl-devops-service/basecontext"
	"github.com/Parallels/prl-devops-service/data/models"
	"github.com/Parallels/prl-devops-service/errors"
	"github.com/Parallels/prl-devops-service/helpers"
)

var (
	ErrUserConfigNotFound      = errors.NewWithCode("user config not found", 404)
	ErrUserConfigAlreadyExists = errors.NewWithCode("user config already exists", 500)
)

func (j *JsonDatabase) GetUserConfigs(ctx basecontext.ApiContext, userID string, filter string) ([]models.UserConfig, error) {
	if !j.IsConnected() {
		return nil, ErrDatabaseNotConnected
	}

	dbFilter, err := ParseFilter(filter)
	if err != nil {
		return nil, err
	}

	filtered := make([]models.UserConfig, 0)
	for _, cfg := range j.data.UserConfigs {
		if strings.EqualFold(cfg.UserID, userID) {
			filtered = append(filtered, cfg)
		}
	}

	filteredData, err := FilterByProperty(filtered, dbFilter)
	if err != nil {
		return nil, err
	}

	return filteredData, nil
}

func (j *JsonDatabase) GetUserConfig(ctx basecontext.ApiContext, userID string, idOrSlug string) (*models.UserConfig, error) {
	if !j.IsConnected() {
		return nil, ErrDatabaseNotConnected
	}

	for _, cfg := range j.data.UserConfigs {
		if strings.EqualFold(cfg.UserID, userID) &&
			(strings.EqualFold(cfg.ID, idOrSlug) || strings.EqualFold(cfg.Slug, idOrSlug)) {
			return &cfg, nil
		}
	}

	return nil, ErrUserConfigNotFound
}

func (j *JsonDatabase) CreateUserConfig(ctx basecontext.ApiContext, cfg models.UserConfig) (*models.UserConfig, error) {
	if !j.IsConnected() {
		return nil, ErrDatabaseNotConnected
	}

	if cfg.ID == "" {
		cfg.ID = helpers.GenerateId()
	}

	if existing, _ := j.GetUserConfig(ctx, cfg.UserID, cfg.Slug); existing != nil {
		return nil, ErrUserConfigAlreadyExists
	}

	cfg.CreatedAt = helpers.GetUtcCurrentDateTime()
	cfg.UpdatedAt = helpers.GetUtcCurrentDateTime()
	j.data.UserConfigs = append(j.data.UserConfigs, cfg)

	j.SaveNow(ctx)

	return &cfg, nil
}

func (j *JsonDatabase) UpsertUserConfig(ctx basecontext.ApiContext, cfg models.UserConfig) (*models.UserConfig, error) {
	if !j.IsConnected() {
		return nil, ErrDatabaseNotConnected
	}

	existing, _ := j.GetUserConfig(ctx, cfg.UserID, cfg.Slug)
	if existing == nil {
		return j.CreateUserConfig(ctx, cfg)
	}

	cfg.ID = existing.ID
	return j.UpdateUserConfig(ctx, cfg)
}

func (j *JsonDatabase) UpdateUserConfig(ctx basecontext.ApiContext, cfg models.UserConfig) (*models.UserConfig, error) {
	if !j.IsConnected() {
		return nil, ErrDatabaseNotConnected
	}

	for i, existing := range j.data.UserConfigs {
		if !strings.EqualFold(existing.ID, cfg.ID) || !strings.EqualFold(existing.UserID, cfg.UserID) {
			continue
		}

		for {
			if IsRecordLocked(j.data.UserConfigs[i].DbRecord) {
				continue
			}
			LockRecord(ctx, j.data.UserConfigs[i].DbRecord)
			j.data.UserConfigs[i].Name = cfg.Name
			j.data.UserConfigs[i].Type = cfg.Type
			j.data.UserConfigs[i].Value = cfg.Value
			j.data.UserConfigs[i].UpdatedAt = helpers.GetUtcCurrentDateTime()
			UnlockRecord(ctx, j.data.UserConfigs[i].DbRecord)
			break
		}

		result := j.data.UserConfigs[i]
		j.SaveNow(ctx)
		return &result, nil
	}

	return nil, ErrUserConfigNotFound
}

func (j *JsonDatabase) DeleteUserConfig(ctx basecontext.ApiContext, userID string, idOrSlug string) error {
	if !j.IsConnected() {
		return ErrDatabaseNotConnected
	}

	for i, cfg := range j.data.UserConfigs {
		if strings.EqualFold(cfg.UserID, userID) &&
			(strings.EqualFold(cfg.ID, idOrSlug) || strings.EqualFold(cfg.Slug, idOrSlug)) {
			for {
				if IsRecordLocked(j.data.UserConfigs[i].DbRecord) {
					continue
				}
				LockRecord(ctx, j.data.UserConfigs[i].DbRecord)
				j.data.UserConfigs = append(j.data.UserConfigs[:i], j.data.UserConfigs[i+1:]...)
				break
			}
			return nil
		}
	}

	return ErrUserConfigNotFound
}
