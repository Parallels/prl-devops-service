package pdfile

import (
	"bufio"
	"os"
	"path/filepath"
	"strings"

	"github.com/Parallels/prl-devops-service/basecontext"
	"github.com/Parallels/prl-devops-service/pdfile/diagnostics"
	"github.com/Parallels/prl-devops-service/pdfile/models"
)

func Load(ctx basecontext.ApiContext, pdFilepath string) (*models.PDFile, *diagnostics.PDFileDiagnostics) {
	lines := []string{}
	diag := diagnostics.NewPDFileDiagnostics()

	file, err := os.Open(filepath.Clean(pdFilepath))
	if err != nil {
		diag.AddError(err)
		return nil, diag
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}
	if err := scanner.Err(); err != nil {
		diag.AddError(err)
		return nil, diag
	}

	result, diag := Process(ctx, strings.Join(lines, "\n"))
	return result, diag
}
