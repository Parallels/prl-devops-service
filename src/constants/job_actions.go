package constants

// Action Strings used across the Job Workflow for Pull and Push operations
const (
	ActionValidatingRequest         = "Validating request"
	ActionCheckingLocalCatalog      = "Checking local catalog"
	ActionCheckingRemoteCatalog     = "Checking Remote Providers"
	ActionDownloadingManifest       = "Downloading manifest"
	ActionCreatingDestinationFolder = "Creating destination folder"
	ActionDownloadingPackFile       = "Downloading pack file"
	ActionCachingPackFile           = "Caching pack file"
	ActionCopyingFromCache          = "Copying from cache"
	ActionDecompressingPackFile     = "Decompressing pack file"
	ActionCleaningStructure         = "Cleaning and flattening pulled structure"
	ActionRegisteringMachine        = "Registering machine"
	ActionRenamingMachine           = "Renaming machine"
	ActionStartingMachine           = "Starting machine"
	ActionUploadingPackFile         = "Uploading pack file"

	// New ones
	ActionInitialize   = "Initializing"
	ActionDecompressor = "decompressor"
	ActionDownloader   = "downloader"
	ActionUploader     = "uploader"
	ActionCleaningUp   = "cleaning_up"

	// Pull Actions
	ActionPullValidateStage   = "pull_validate_stage"
	ActionPullCheckCacheStage = "pull_check_cache_stage"
	ActionPullCacheStage      = "pull_cache_stage"
	ActionPullRegisterVm      = "pull_register_stage"
	ActionPullRenameVm        = "pull_rename_stage"
	ActionPullStartVm         = "pull_start_stage"
)
