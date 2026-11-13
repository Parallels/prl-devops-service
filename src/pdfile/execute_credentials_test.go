package pdfile

import (
	"strings"
	"testing"

	"github.com/Parallels/prl-devops-service/basecontext"
	"github.com/Parallels/prl-devops-service/pdfile/models"
	"github.com/Parallels/prl-devops-service/pdfile/processors"
)

// EXECUTE_USER/EXECUTE_PASSWORD carry the guest credentials used for every
// EXECUTE line, mirroring prlctl exec --user/--password.

func TestExecuteUserProcessor(t *testing.T) {
	dest := models.NewPdFile()
	ctx := basecontext.NewRootBaseContext()
	processed, diag := processors.ExecuteUserCommandProcessor{}.Process(ctx, "EXECUTE_USER build", dest)
	if !processed {
		t.Fatalf("line must be processed")
	}
	if diag.HasErrors() {
		t.Fatalf("unexpected errors: %v", diag.Errors())
	}
	if dest.ExecuteUser != "build" {
		t.Fatalf("expected ExecuteUser build, got %q", dest.ExecuteUser)
	}

	other, _ := processors.ExecuteUserCommandProcessor{}.Process(ctx, "EXECUTE curl -o /tmp/x", dest)
	if other {
		t.Fatalf("EXECUTE lines must not be handled by the user processor")
	}
}

// Guest passwords may contain spaces; the line parser keeps the whole tail.
func TestExecutePasswordProcessorKeepsSpaces(t *testing.T) {
	dest := models.NewPdFile()
	ctx := basecontext.NewRootBaseContext()
	processed, diag := processors.ExecutePasswordCommandProcessor{}.Process(ctx, "EXECUTE_PASSWORD sup3r secret !", dest)
	if !processed {
		t.Fatalf("line must be processed")
	}
	if diag.HasErrors() {
		t.Fatalf("unexpected errors: %v", diag.Errors())
	}
	if dest.ExecutePassword != "sup3r secret !" {
		t.Fatalf("expected password with spaces preserved, got %q", dest.ExecutePassword)
	}
}

func containsExecuteUserError(diag interface {
	Errors() []error
}, want string) bool {
	for _, err := range diag.Errors() {
		if strings.Contains(err.Error(), want) {
			return true
		}
	}
	return false
}

func TestValidateExecutePasswordRequiresExecuteUser(t *testing.T) {
	ctx := basecontext.NewRootBaseContext()

	noUser := &models.PDFile{Raw: []string{"FROM local", "EXECUTE curl -o /tmp/x", "EXECUTE_PASSWORD sup3r", "PULL"}}
	if !containsExecuteUserError(NewPDFileService(ctx, noUser).Validate(), "execute password was found but execute user was not found") {
		t.Fatalf("expected execute-user pairing error, got %v", NewPDFileService(ctx, noUser).Validate().Errors())
	}

	withUser := &models.PDFile{Raw: []string{"FROM local", "EXECUTE curl -o /tmp/x", "EXECUTE_USER build", "EXECUTE_PASSWORD sup3r", "PULL"}}
	if containsExecuteUserError(NewPDFileService(ctx, withUser).Validate(), "execute password was found but execute user was not found") {
		t.Fatalf("paired credentials must not raise the pairing error")
	}

	none := &models.PDFile{Raw: []string{"FROM local", "EXECUTE curl -o /tmp/x", "PULL"}}
	if containsExecuteUserError(NewPDFileService(ctx, none).Validate(), "execute password was found but execute user was not found") {
		t.Fatalf("absent credentials must not raise the pairing error")
	}
}

// Existing pdfiles without the new fields keep validating as before.
func TestValidateUnaffectedWithoutExecuteCredentials(t *testing.T) {
	ctx := basecontext.NewRootBaseContext()
	pd := &models.PDFile{Raw: []string{"FROM local", "PULL"}, MachineName: "m", Destination: "/tmp", Owner: "o"}
	diag := NewPDFileService(ctx, pd).Validate()
	if containsExecuteUserError(diag, "execute") {
		t.Fatalf("no execute-related errors expected, got %v", diag.Errors())
	}
}
