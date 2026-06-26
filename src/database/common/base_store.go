package common

import "gorm.io/gorm"

// BaseDataStore provides common functionality for all data stores
type BaseDataStore struct {
	db *gorm.DB
}

// NewBaseDataStore creates a new base data store
func NewBaseDataStore(db *gorm.DB) *BaseDataStore {
	return &BaseDataStore{
		db: db,
	}
}

// GetDB returns the database connection
func (s *BaseDataStore) GetDB() *gorm.DB {
	return s.db
}

// WithTransaction executes the given function within a transaction
func (s *BaseDataStore) WithTransaction(fn func(tx *gorm.DB) error) error {
	return s.db.Transaction(fn)
}
