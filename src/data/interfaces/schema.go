package interfaces

// SchemaService mirrors schema method signatures verbatim (PBI #421 D-2).
// Legacy implementor: data.JsonDatabase; SQL implementor: database.GormDatabase (#421 task 04+).
import (
	"github.com/Parallels/prl-devops-service/basecontext"
)

type SchemaService interface {
	GetSchemaVersion(ctx basecontext.ApiContext) (string, error)
	UpdateSchemaVersion(ctx basecontext.ApiContext, version string) error
}
