package mappers

import (
	"github.com/Parallels/prl-devops-service/config"
	data_models "github.com/Parallels/prl-devops-service/data/models"
	"github.com/Parallels/prl-devops-service/helpers"
	"github.com/Parallels/prl-devops-service/models"
	"github.com/Parallels/prl-devops-service/security"
)

func ToCatalogManagerResponse(mgr *data_models.CatalogManager, includeCredentials bool) *models.CatalogManager {
	if mgr == nil {
		return nil
	}

	res := &models.CatalogManager{
		ID:                   mgr.ID,
		Name:                 mgr.Name,
		URL:                  mgr.URL,
		Internal:             mgr.Internal,
		Active:               mgr.Active,
		AuthenticationMethod: mgr.AuthenticationMethod,
		Global:               mgr.Global,
		RequiredClaims:       mgr.RequiredClaims,
		OwnerID:              mgr.OwnerID,
		CreatedAt:            mgr.CreatedAt,
		UpdatedAt:            mgr.UpdatedAt,
	}

	if includeCredentials {
		cfg := config.Get()
		res.Username = mgr.Username
		if mgr.Password != "" && cfg.EncryptionPrivateKey() != "" {
			if decrypted, err := security.DecryptString(cfg.EncryptionPrivateKey(), []byte(mgr.Password)); err == nil {
				res.Password = decrypted
			}
		} else {
			res.Password = mgr.Password
		}

		if mgr.ApiKey != "" && cfg.EncryptionPrivateKey() != "" {
			if decrypted, err := security.DecryptString(cfg.EncryptionPrivateKey(), []byte(mgr.ApiKey)); err == nil {
				res.ApiKey = decrypted
			}
		} else {
			res.ApiKey = mgr.ApiKey
		}
	} else {
		// Just a visual indicator that it exists
		if mgr.Username != "" {
			res.Username = "********"
		}
		if mgr.Password != "" {
			res.Password = "********"
		}
		if mgr.ApiKey != "" {
			res.ApiKey = "********"
		}
	}

	return res
}

func ToCatalogManagerResponseList(mgrs []data_models.CatalogManager, includeCredentials bool) []models.CatalogManager {
	res := make([]models.CatalogManager, 0, len(mgrs))
	for _, m := range mgrs {
		res = append(res, *ToCatalogManagerResponse(&m, includeCredentials))
	}
	return res
}

func FromCatalogManagerRequest(req *models.CatalogManagerRequest) *data_models.CatalogManager {
	if req == nil {
		return nil
	}

	res := &data_models.CatalogManager{
		ID:                   helpers.GenerateId(),
		Name:                 req.Name,
		URL:                  req.URL,
		Internal:             req.Internal,
		Active:               req.Active,
		AuthenticationMethod: req.AuthenticationMethod,
		Username:             req.Username,
		Global:               req.Global,
		RequiredClaims:       req.RequiredClaims,
	}

	cfg := config.Get()

	// Encrypt the credentials if present
	if req.Password != "" {
		if cfg.EncryptionPrivateKey() != "" {
			if encrypted, err := security.EncryptString(cfg.EncryptionPrivateKey(), req.Password); err == nil {
				res.Password = string(encrypted)
			}
		} else {
			res.Password = req.Password
		}
	}

	if req.ApiKey != "" {
		if cfg.EncryptionPrivateKey() != "" {
			if encrypted, err := security.EncryptString(cfg.EncryptionPrivateKey(), req.ApiKey); err == nil {
				res.ApiKey = string(encrypted)
			}
		} else {
			res.ApiKey = req.ApiKey
		}
	}

	return res
}

func UpdateCatalogManagerFromRequest(mgr *data_models.CatalogManager, req *models.CatalogManagerRequest) {
	if mgr == nil || req == nil {
		return
	}

	mgr.Name = req.Name
	mgr.URL = req.URL
	mgr.Internal = req.Internal
	mgr.Active = req.Active
	mgr.AuthenticationMethod = req.AuthenticationMethod
	mgr.Username = req.Username
	mgr.Global = req.Global
	mgr.RequiredClaims = req.RequiredClaims

	cfg := config.Get()

	// Only update credentials if they are provided (prevent wiping them out on normal updates)
	if req.Password != "" {
		if cfg.EncryptionPrivateKey() != "" {
			if encrypted, err := security.EncryptString(cfg.EncryptionPrivateKey(), req.Password); err == nil {
				mgr.Password = string(encrypted)
			}
		} else {
			mgr.Password = req.Password
		}
	}

	if req.ApiKey != "" {
		if cfg.EncryptionPrivateKey() != "" {
			if encrypted, err := security.EncryptString(cfg.EncryptionPrivateKey(), req.ApiKey); err == nil {
				mgr.ApiKey = string(encrypted)
			}
		} else {
			mgr.ApiKey = req.ApiKey
		}
	}
}
