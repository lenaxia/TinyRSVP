package templates

import (
	"context"
	"embed"
	"fmt"

	"github.com/lenaxia/tinyrsvp/internal/db/repositories"
	"github.com/lenaxia/tinyrsvp/internal/models"
)

//go:embed defaults/*.html defaults/*.txt
var defaultTemplates embed.FS

type Seeder struct {
	repo      repositories.TemplateRepository
	createdBy int64
}

func NewSeeder(repo repositories.TemplateRepository, createdBy int64) *Seeder {
	return &Seeder{
		repo:      repo,
		createdBy: createdBy,
	}
}

func (s *Seeder) SeedDefaults(ctx context.Context) error {
	if s.createdBy == 0 {
		return &models.ValidationError{
			Field:   "created_by",
			Message: "created_by must be non-zero for seeding templates",
		}
	}

	templates := []struct {
		name     string
		typ      models.TemplateType
		htmlFile string
		textFile string
	}{
		{
			name:     "Default Invite Email",
			typ:      models.TemplateTypeInviteEmail,
			htmlFile: "defaults/invite_email.html",
			textFile: "defaults/invite_email.txt",
		},
		{
			name:     "Default RSVP Page",
			typ:      models.TemplateTypeRSVPPage,
			htmlFile: "defaults/rsvp_page.html",
			textFile: "",
		},
		{
			name:     "Default Confirmation Page",
			typ:      models.TemplateTypeConfirmationPage,
			htmlFile: "defaults/confirmation_page.html",
			textFile: "",
		},
	}

	for _, tmpl := range templates {
		if err := ctx.Err(); err != nil {
			return fmt.Errorf("context cancelled: %w", err)
		}

		existing, err := s.repo.GetDefaultByType(ctx, tmpl.typ)
		if err == nil && existing != nil {
			continue
		}

		var notFoundErr *models.NotFoundError
		if err != nil {
			switch e := err.(type) {
			case *models.NotFoundError:
				notFoundErr = e
			default:
				return fmt.Errorf("failed to check for existing template %s: %w", tmpl.typ, err)
			}
		}

		if notFoundErr == nil && existing != nil {
			continue
		}

		htmlContent, err := defaultTemplates.ReadFile(tmpl.htmlFile)
		if err != nil {
			return fmt.Errorf("failed to read %s: %w", tmpl.htmlFile, err)
		}

		var textContent *string
		if tmpl.textFile != "" {
			textBytes, err := defaultTemplates.ReadFile(tmpl.textFile)
			if err != nil {
				return fmt.Errorf("failed to read %s: %w", tmpl.textFile, err)
			}
			text := string(textBytes)
			textContent = &text
		}

		template := &models.Template{
			Name:        tmpl.name,
			Type:        tmpl.typ,
			HTMLContent: string(htmlContent),
			TextContent: textContent,
			IsDefault:   true,
			IsActive:    true,
			Version:     1,
			CreatedBy:   s.createdBy,
			Category:    models.CategoryPlain,
		}

		if err := s.repo.Create(ctx, template); err != nil {
			return fmt.Errorf("failed to create default template %s: %w", tmpl.name, err)
		}
	}

	return nil
}
