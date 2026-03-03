package data

import (
	"errors"

	"github.com/Parallels/prl-devops-service/basecontext"
	"github.com/Parallels/prl-devops-service/data/models"
)

var (
	ErrCatalogManagerNotFound       = errors.New("catalog manager not found")
	ErrRemoveInternalCatalogManager = errors.New("catalog manager is internal and cannot be removed")
	ErrUpdateInternalCatalogManager = errors.New("catalog manager is internal and cannot be updated")
)

func (j *JsonDatabase) GetCatalogManagers() ([]models.CatalogManager, error) {
	j.dataMutex.RLock()
	defer j.dataMutex.RUnlock()

	return j.data.CatalogManagers, nil
}

func (j *JsonDatabase) GetCatalogManager(id string) (*models.CatalogManager, error) {
	j.dataMutex.RLock()
	defer j.dataMutex.RUnlock()

	for _, mgr := range j.data.CatalogManagers {
		if mgr.ID == id {
			return &mgr, nil
		}
	}

	return nil, ErrCatalogManagerNotFound
}

func (j *JsonDatabase) AddCatalogManager(ctx basecontext.ApiContext, mgr models.CatalogManager) error {
	j.dataMutex.Lock()
	if j.data.CatalogManagers == nil {
		j.data.CatalogManagers = make([]models.CatalogManager, 0)
	}
	j.data.CatalogManagers = append(j.data.CatalogManagers, mgr)
	j.dataMutex.Unlock()

	if err := j.SaveAsync(ctx); err != nil {
		return err
	}

	return nil
}

func (j *JsonDatabase) UpdateCatalogManager(ctx basecontext.ApiContext, mgr models.CatalogManager) error {
	j.dataMutex.Lock()
	defer j.dataMutex.Unlock()

	for i, m := range j.data.CatalogManagers {
		if m.ID == mgr.ID {
			if m.Internal {
				return ErrUpdateInternalCatalogManager
			}

			j.data.CatalogManagers[i] = mgr

			if err := j.SaveAsync(ctx); err != nil {
				return err
			}

			return nil
		}
	}

	return ErrCatalogManagerNotFound
}

func (j *JsonDatabase) DeleteCatalogManager(ctx basecontext.ApiContext, id string) error {
	j.dataMutex.Lock()
	defer j.dataMutex.Unlock()

	found := false
	for i, m := range j.data.CatalogManagers {
		if m.ID == id {
			if m.Internal && !IsRootUser(ctx) {
				return ErrRemoveInternalCatalogManager
			}
			j.data.CatalogManagers = append(j.data.CatalogManagers[:i], j.data.CatalogManagers[i+1:]...)
			found = true
			break
		}
	}

	if !found {
		return ErrCatalogManagerNotFound
	}

	if err := j.SaveAsync(ctx); err != nil {
		return err
	}

	return nil
}
