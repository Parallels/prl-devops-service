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
)
