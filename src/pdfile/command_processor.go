package pdfile

import (
	"github.com/Parallels/prl-devops-service/basecontext"
	"github.com/Parallels/prl-devops-service/pdfile/diagnostics"
	"github.com/Parallels/prl-devops-service/pdfile/models"
)

type PDFileCommandProcessor interface {
	Process(ctx basecontext.ApiContext, line string, pdFile *models.PDFile) (bool, *diagnostics.PDFileDiagnostics)
}
