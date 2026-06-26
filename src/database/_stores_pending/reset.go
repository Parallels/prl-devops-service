package stores

import "sync"

// Reset clears all store singletons (for testing)
func Reset() {
	// Tenant
	tenantDataStoreInstance = nil
	tenantDataStoreOnce = sync.Once{}

	// User
	userDataStoreInstance = nil
	userDataStoreOnce = sync.Once{}

	// Role
	roleDataStoreInstance = nil
	roleDataStoreOnce = sync.Once{}

	// Claim
	claimDataStoreInstance = nil
	claimDataStoreOnce = sync.Once{}

	// Activity
	activityDataStoreInstance = nil
	activityDataStoreOnce = sync.Once{}

	// ApiKey (Auth Data Store)
	authDataStoreInstance = nil
	authDataStoreOnce = sync.Once{}

	// Email
	emailDataStoreInstance = nil
	emailDataStoreOnce = sync.Once{}

	// Message
	messageDataStoreInstance = nil
	messageDataStoreOnce = sync.Once{}

	// Configuration
	configurationDataStoreInstance = nil
	configurationDataStoreOnce = sync.Once{}

	// WebAuthn
	webAuthnDataStoreInstance = nil
	webAuthnDataStoreOnce = sync.Once{}

	// Notification
	notificationDataStoreInstance = nil
	notificationDataStoreOnce = sync.Once{}

	// IpBan
	ipBanDataStoreInstance = nil
	ipBanDataStoreOnce = sync.Once{}
}
