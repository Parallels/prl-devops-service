package interfaces

// LifecycleService mirrors main backup method signatures verbatim (PBI #421 D-2).
// Legacy implementor: data.JsonDatabase; SQL implementor: database.GormDatabase (#421 task 04+).
import (
	"github.com/Parallels/prl-devops-service/basecontext"
)

type LifecycleService interface {
	Connect(ctx basecontext.ApiContext) error
	Load(ctx basecontext.ApiContext) error
	Disconnect(ctx basecontext.ApiContext) error
	Filename() string
	IsConnected() bool
	SaveAs(ctx basecontext.ApiContext, filename string) error
	SaveAsync(ctx basecontext.ApiContext) error
	SaveNow(ctx basecontext.ApiContext) error
	SaveNowForced(ctx basecontext.ApiContext, reason string) error
	ProcessSaveQueue(ctx basecontext.ApiContext)
	IsDataFileEmpty(ctx basecontext.ApiContext) (bool, error)
	RunBackup(ctx basecontext.ApiContext)
	Backup(ctx basecontext.ApiContext) error
}
