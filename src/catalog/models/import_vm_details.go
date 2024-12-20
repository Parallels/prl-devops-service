package models

type ImportVmManifestDetails struct {
	HasMetaFile      bool
	FilePath         string
	MetadataFilename string
	HasPackFile      bool
	MachineFilename  string
	MachineFileSize  int64
}
