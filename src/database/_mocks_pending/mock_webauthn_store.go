package mocks

import (
	"github.com/Parallels/prl-devops-service/basecontext"
	"github.com/Parallels/prl-devops-service/database/entities"
	"github.com/Parallels/prl-devops-service/errors"
)

type MockWebAuthnStore struct {
	BaseMockStore
}

func NewMockWebAuthnStore() *MockWebAuthnStore {
	return &MockWebAuthnStore{
		BaseMockStore: *NewBaseMockStore(),
	}
}

func (m *MockWebAuthnStore) GetCredentialsByUser(ctx basecontext.BaseContext, userID string) ([]entities.WebAuthnCredential, *errors.Diagnostics) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Get(1).(*errors.Diagnostics)
	}
	return args.Get(0).([]entities.WebAuthnCredential), args.Get(1).(*errors.Diagnostics)
}

func (m *MockWebAuthnStore) GetCredentialByCredentialID(ctx basecontext.BaseContext, credentialID []byte) (*entities.WebAuthnCredential, *errors.Diagnostics) {
	args := m.Called(ctx, credentialID)
	if args.Get(0) == nil {
		return nil, args.Get(1).(*errors.Diagnostics)
	}
	return args.Get(0).(*entities.WebAuthnCredential), args.Get(1).(*errors.Diagnostics)
}

func (m *MockWebAuthnStore) SaveCredential(ctx basecontext.BaseContext, cred *entities.WebAuthnCredential) *errors.Diagnostics {
	args := m.Called(ctx, cred)
	if args.Get(0) == nil {
		return nil
	}
	return args.Get(0).(*errors.Diagnostics)
}

func (m *MockWebAuthnStore) UpdateCredential(ctx basecontext.BaseContext, cred *entities.WebAuthnCredential) *errors.Diagnostics {
	args := m.Called(ctx, cred)
	if args.Get(0) == nil {
		return nil
	}
	return args.Get(0).(*errors.Diagnostics)
}
