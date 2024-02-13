package password

import (
	"testing"

	"github.com/Parallels/prl-devops-service/basecontext"
	"github.com/stretchr/testify/assert"
)

func TestSetOptions(t *testing.T) {
	ctx := basecontext.NewRootBaseContext()
	svc := New(ctx)
	svc.SetOptions(&PasswordComplexityOptions{
		ctx:                      ctx,
		minLength:                20,
		maxLength:                40,
		requireLowercase:         true,
		requireUppercase:         true,
		requireNumbers:           true,
		requireSpecialCharacters: true,
		saltPassword:             false,
	})

	testSvc := Get()

	assert.Equal(t, 20, svc.GetOptions().MinLength())
	assert.Equal(t, 40, svc.GetOptions().MaxLength())
	assert.Equal(t, true, svc.GetOptions().RequireLowercase())
	assert.Equal(t, true, svc.GetOptions().RequireUppercase())
	assert.Equal(t, true, svc.GetOptions().RequireNumbers())
	assert.Equal(t, true, svc.GetOptions().RequireSpecialCharacters())

	assert.Equal(t, testSvc.GetOptions().MinLength(), svc.GetOptions().MinLength())
	assert.Equal(t, testSvc.GetOptions().MaxLength(), svc.GetOptions().MaxLength())
	assert.Equal(t, testSvc.GetOptions().RequireLowercase(), svc.GetOptions().RequireLowercase())
	assert.Equal(t, testSvc.GetOptions().RequireUppercase(), svc.GetOptions().RequireUppercase())
	assert.Equal(t, testSvc.GetOptions().RequireNumbers(), svc.GetOptions().RequireNumbers())
	assert.Equal(t, testSvc.GetOptions().RequireSpecialCharacters(), svc.GetOptions().RequireSpecialCharacters())
	assert.Equal(t, testSvc.GetOptions().SaltPassword(), svc.GetOptions().SaltPassword())
}
