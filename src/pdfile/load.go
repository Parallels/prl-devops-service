package pdfile

import (
	"bufio"
	"os"
	"strings"
)

func Load(filepath string) (*PDFile, *PDFileDiagnostics) {
	lines := []string{}
	diag := NewPDFileDiagnostics()

	file, err := os.Open(filepath)
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

	result, diag := Process(strings.Join(lines, "\n"))
	return result, diag
}
