package pdfile

import (
	"strings"

	"github.com/Parallels/prl-devops-service/basecontext"
	"github.com/Parallels/prl-devops-service/pdfile/diagnostics"
	"github.com/Parallels/prl-devops-service/pdfile/models"
	"github.com/Parallels/prl-devops-service/pdfile/processors"
)

type PDFileService struct {
	ctx        basecontext.ApiContext
	processors []PDFileCommandProcessor
	pdfile     *models.PDFile
}

func NewPDFileService(ctx basecontext.ApiContext, pdFile *models.PDFile) *PDFileService {
	return &PDFileService{
		ctx: ctx,
		processors: []PDFileCommandProcessor{
			&processors.EmptyOrCommentedCommandProcessor{},
			&processors.ArchitectureCommandProcessor{},
			&processors.AuthenticateCommandProcessor{},
			&processors.CatalogIdCommandProcessor{},
			&processors.ClaimCommandProcessor{},
			&processors.DescriptionCommandProcessor{},
			&processors.DestinationCommandProcessor{},
			&processors.ExecuteCommandProcessor{},
			&processors.FromToCommandProcessor{},
			&processors.InsecureCommandProcessor{},
			&processors.LocalPathCommandProcessor{},
			&processors.MachineNameCommandProcessor{},
			&processors.OwnerCommandProcessor{},
			&processors.ProviderCommandProcessor{},
			&processors.RoleCommandProcessor{},
			&processors.CommandCommandProcessor{},
			&processors.StartAfterPullCommandProcessor{},
			&processors.TagCommandProcessor{},
			&processors.VersionCommandProcessor{},
			&processors.CloneCommandProcessor{},
			&processors.ClientCommandProcessor{},
			&processors.MinimumSpecsRequirementsCommandProcessor{},
			&processors.IsCompressedCommandProcessor{},
			&processors.ForceCommandProcessor{},
			&processors.CompressPackCommandProcessor{},
			&processors.VmRemotePathCommandProcessor{},
			&processors.VmSizeCommandProcessor{},
			&processors.VmTypeCommandProcessor{},
			&processors.CompressPackLevelCommandProcessor{},
			&processors.CloneDestinationCommandProcessor{},
		},

		pdfile: pdFile,
	}
}

func (p *PDFileService) Run(ctx basecontext.ApiContext) (interface{}, *diagnostics.PDFileDiagnostics) {
	diag := diagnostics.NewPDFileDiagnostics()

	if strings.EqualFold(p.pdfile.Command, "list") {
		url := p.pdfile.GetHostCatalogUrl()
		out, runDiag := p.runList(ctx, url)
		diag.Append(runDiag)
		return out, diag
	}

	if strings.EqualFold(p.pdfile.Command, "push") {
		out, runDiag := p.runPush(ctx)
		diag.Append(runDiag)
		return out, diag
	}

	if strings.EqualFold(p.pdfile.Command, "pull") {
		out, runDiag := p.runPull(ctx)
		diag.Append(runDiag)
		return out, diag
	}

	if strings.EqualFold(p.pdfile.Command, "import") {
		out, runDiag := p.runImport(ctx)
		diag.Append(runDiag)
		return out, diag
	}

	if strings.EqualFold(p.pdfile.Command, "import-vm") {
		out, runDiag := p.runImportVM(ctx)
		diag.Append(runDiag)
		return out, diag
	}

	return nil, diag
}
