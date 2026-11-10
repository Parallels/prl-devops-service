package interfaces

// DatabaseService is the frozen 158-method data-access surface of #421 (R-1 guard:
// inventory captured from data/*.go at task-01 time; drift is a compile error).
// Composed per-domain; implementors: data.JsonDatabase (legacy) and
// database.GormDatabase (SQL). See #421 plan D-2.
type DatabaseService interface {
	LifecycleService
	UsersService
	RolesService
	ClaimsService
	ApiKeysService
	CatalogService
	CatalogManagersService
	OrchestratorService
	SnapshotsService
	JobsService
	ReverseProxyService
	UserConfigsService
	ConfigService
	SchemaService
	PackerTemplatesService
	EnrollmentTokensService
}
