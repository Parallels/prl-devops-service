package errors

import (
	"fmt"
	"strings"
)

type Diagnostics struct {
	errors   []error
	warnings []error
}

func NewDiagnostics() *Diagnostics {
	return &Diagnostics{
		errors:   []error{},
		warnings: []error{},
	}
}

func (d *Diagnostics) AddError(err error) {
	for _, e := range d.errors {
		if e.Error() == err.Error() {
			return
		}
	}

	d.errors = append(d.errors, err)
}

func (d *Diagnostics) AddWarning(err error) {
	for _, e := range d.warnings {
		if e.Error() == err.Error() {
			return
		}
	}
	d.warnings = append(d.warnings, err)
}

func (d *Diagnostics) HasErrors() bool {
	return len(d.errors) > 0
}

func (d *Diagnostics) HasWarnings() bool {
	return len(d.warnings) > 0
}

func (d *Diagnostics) Errors() []error {
	return d.errors
}

func (d *Diagnostics) Warnings() []error {
	return d.warnings
}

func (d *Diagnostics) Append(diagnostics *Diagnostics) {
	d.errors = append(d.errors, diagnostics.errors...)
	d.warnings = append(d.warnings, diagnostics.warnings...)
}

func (d *Diagnostics) Error() string {
	msg := ""
	if len(d.errors) > 0 {
		if len(d.errors) == 1 {
			return d.errors[0].Error()
		} else {
			msg = "errors:\n"
			for _, err := range d.errors {
				errMsg := strings.ReplaceAll(err.Error(), "error: ", "")
				msg = fmt.Sprintf("%v\t%v\n", msg, errMsg)
			}
		}
	}

	if len(d.warnings) > 0 {
		if len(d.warnings) == 1 {
			return d.warnings[0].Error()
		} else {
			msg = "warnings:\n"
			for _, err := range d.errors {
				errMsg := strings.ReplaceAll(err.Error(), "error: ", "")
				msg = fmt.Sprintf("%v\t%v\n", msg, errMsg)
			}
		}
	}

	return msg
}
