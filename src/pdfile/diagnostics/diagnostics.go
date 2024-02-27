package diagnostics

type PDFileDiagnostics struct {
	errors   []error
	warnings []error
}

func NewPDFileDiagnostics() *PDFileDiagnostics {
	return &PDFileDiagnostics{
		errors: []error{},
	}
}

func (pd *PDFileDiagnostics) AddError(err error) {
	pd.errors = append(pd.errors, err)
}

func (pd *PDFileDiagnostics) AddWarning(err error) {
	pd.warnings = append(pd.warnings, err)
}

func (pd *PDFileDiagnostics) HasErrors() bool {
	return len(pd.errors) > 0
}

func (pd *PDFileDiagnostics) HasWarnings() bool {
	return len(pd.warnings) > 0
}

func (pd *PDFileDiagnostics) Errors() []error {
	return pd.errors
}

func (pd *PDFileDiagnostics) Warnings() []error {
	return pd.warnings
}

func (pd *PDFileDiagnostics) Append(diagnostics *PDFileDiagnostics) {
	pd.errors = append(pd.errors, diagnostics.errors...)
	pd.warnings = append(pd.warnings, diagnostics.warnings...)
}
