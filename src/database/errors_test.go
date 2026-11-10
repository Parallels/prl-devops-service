package database

import (
	"fmt"
	"testing"

	datainterfaces "github.com/Parallels/prl-devops-service/data/interfaces"
	"github.com/Parallels/prl-devops-service/errors"
	"gorm.io/gorm"
)

func TestMapGormError(t *testing.T) {
	// nil → nil
	if err := mapGormError(nil, datainterfaces.ErrUserNotFound, datainterfaces.ErrUserAlreadyExists); err != nil {
		t.Errorf("nil: want nil, got %v", err)
	}

	// record-not-found → legacy sentinel IDENTITY (==, not equality)
	got := mapGormError(gorm.ErrRecordNotFound, datainterfaces.ErrUserNotFound, nil)
	if got != datainterfaces.ErrUserNotFound {
		t.Errorf("not-found sentinel identity broken: %p vs %p", got, datainterfaces.ErrUserNotFound)
	}

	// not-found with no domain sentinel → generic 500 wrap
	wrapped := mapGormError(gorm.ErrRecordNotFound, nil, nil)
	sysErr, ok := wrapped.(*errors.SystemError)
	if !ok || sysErr.Code() != 500 {
		t.Errorf("want *SystemError code 500, got %#v", wrapped)
	}

	// unique violation (sqlite flavour) → dup sentinel identity
	sqliteDup := fmt.Errorf("could not execute: UNIQUE constraint failed: users.id")
	got = mapGormError(sqliteDup, nil, datainterfaces.ErrUserAlreadyExists)
	if got != datainterfaces.ErrUserAlreadyExists {
		t.Errorf("dup sentinel identity broken: %v", got)
	}

	// unique violation (gorm translated flavour)
	got = mapGormError(gorm.ErrDuplicatedKey, nil, datainterfaces.ErrUserAlreadyExists)
	if got != datainterfaces.ErrUserAlreadyExists {
		t.Errorf("gorm.ErrDuplicatedKey mapping broken: %v", got)
	}

	// unknown driver failure → 500 wrap
	unknown := mapGormError(fmt.Errorf("disk on fire"), datainterfaces.ErrUserNotFound, datainterfaces.ErrUserAlreadyExists)
	sysErr, ok = unknown.(*errors.SystemError)
	if !ok || sysErr.Code() != 500 {
		t.Errorf("unknown: want *SystemError code 500, got %#v", unknown)
	}
}
