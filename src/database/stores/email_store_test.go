package stores_test

import (
	"testing"

	"github.com/Parallels/prl-devops-service/basecontext"
	"github.com/Parallels/prl-devops-service/database/common"
	"github.com/Parallels/prl-devops-service/database/models"
	"github.com/Parallels/prl-devops-service/database/stores"
	"github.com/Parallels/prl-devops-service/database/stores/testhelpers"
	logging "github.com/cjlapao/common-go-logger"
	"github.com/stretchr/testify/assert"
)

func TestEmailDataStore(t *testing.T) {
	logger := logging.Get()
	logger.Info("Starting email store tests")
	db := testhelpers.NewTestDB(t)
	defer testhelpers.CleanupDB(db)

	store := &stores.EmailDataStore{
		BaseDataStore: *common.NewBaseDataStore(db),
	}

	ctx := basecontext.NewBaseContext()

	t.Run("CreateTemplate", func(t *testing.T) {
		template := &models.EmailTemplate{
			Name:    "Welcome Email",
			Slug:    "welcome-email",
			Subject: "Welcome to our platform!",
			Body:    "<h1>Welcome!</h1><p>Thank you for joining us.</p>",
		}

		createdTemplate, diag := store.CreateTemplate(*ctx, template)
		assert.False(t, diag.HasErrors())
		assert.NotNil(t, createdTemplate)
		assert.NotEmpty(t, createdTemplate.ID)
		assert.Equal(t, "welcome-email", createdTemplate.Slug)
		assert.Equal(t, "Welcome to our platform!", createdTemplate.Subject)
	})

	t.Run("GetTemplateBySlug", func(t *testing.T) {
		template := &models.EmailTemplate{
			Name:    "Password Reset",
			Slug:    "password-reset",
			Subject: "Reset your password",
			Body:    "<p>Click here to reset your password.</p>",
		}
		store.CreateTemplate(*ctx, template)

		retrieved, diag := store.GetTemplateBySlug(*ctx, "password-reset")
		assert.False(t, diag.HasErrors())
		assert.NotNil(t, retrieved)
		assert.Equal(t, "password-reset", retrieved.Slug)
		assert.Equal(t, "Reset your password", retrieved.Subject)
	})

	t.Run("GetTemplateBySlug_NotFound", func(t *testing.T) {
		retrieved, diag := store.GetTemplateBySlug(*ctx, "non-existent-template")
		assert.Nil(t, retrieved)
		if diag != nil {
			assert.False(t, diag.HasErrors())
		}
	})

	t.Run("GetTemplatesByTenant", func(t *testing.T) {
		templates, diag := store.GetTemplatesByTenant(*ctx)
		assert.False(t, diag.HasErrors())
		assert.NotNil(t, templates)
		assert.True(t, len(templates) >= 2)
	})

	t.Run("CreateSystemTemplate", func(t *testing.T) {
		template := &models.EmailTemplate{
			Name:     "System Alert",
			Slug:     "system-alert",
			Subject:  "System Alert",
			Body:     "<p>System notification</p>",
			IsSystem: true,
		}

		createdTemplate, diag := store.CreateTemplate(*ctx, template)
		assert.False(t, diag.HasErrors())
		assert.True(t, createdTemplate.IsSystem)
	})

	t.Run("UpdateTemplate", func(t *testing.T) {
		template := &models.EmailTemplate{
			Name:    "Update Test",
			Slug:    "update-test",
			Subject: "Original Subject",
			Body:    "<p>Original Body</p>",
		}
		created, diag := store.CreateTemplate(*ctx, template)
		assert.False(t, diag.HasErrors())

		created.Subject = "Updated Subject"
		created.Body = "<p>Updated Body</p>"
		result := db.Save(created)
		assert.NoError(t, result.Error)

		retrieved, diag := store.GetTemplateBySlug(*ctx, "update-test")
		assert.False(t, diag.HasErrors())
		assert.Equal(t, "Updated Subject", retrieved.Subject)
	})

	t.Run("DeleteTemplate", func(t *testing.T) {
		template := &models.EmailTemplate{
			Name:    "Delete Test",
			Slug:    "delete-test",
			Subject: "Delete Me",
			Body:    "<p>Delete this template</p>",
		}
		created, diag := store.CreateTemplate(*ctx, template)
		assert.False(t, diag.HasErrors())

		result := db.Delete(created)
		assert.NoError(t, result.Error)

		retrieved, diag := store.GetTemplateBySlug(*ctx, "delete-test")
		assert.Nil(t, retrieved)
	})

	t.Run("CreateTemplate_WithHTMLBody", func(t *testing.T) {
		htmlBody := "<html><body><h1>Hello {{.Name}}</h1><p>Code: {{.Code}}</p></body></html>"
		template := &models.EmailTemplate{
			Name:    "Verification Email",
			Slug:    "verification",
			Subject: "Verify your email",
			Body:    htmlBody,
		}

		created, diag := store.CreateTemplate(*ctx, template)
		assert.False(t, diag.HasErrors())
		assert.Contains(t, created.Body, "{{.Name}}")
	})
}
