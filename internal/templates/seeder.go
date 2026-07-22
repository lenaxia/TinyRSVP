package templates

import (
	"context"
	"embed"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"

	"github.com/lenaxia/tinyrsvp/internal/db/repositories"
	"github.com/lenaxia/tinyrsvp/internal/models"
)

//go:embed defaults/*.html defaults/*.txt defaults/component_configs/*.json
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

func (s *Seeder) SeedThemes(ctx context.Context) error {
	themes := s.getDefaultThemes()

	for _, theme := range themes {
		if err := ctx.Err(); err != nil {
			return fmt.Errorf("context cancelled: %w", err)
		}

		if err := s.seedTheme(ctx, theme); err != nil {
			slog.Warn("Failed to seed theme", "theme", theme.Name, "error", err)
		}
	}

	return nil
}

func (s *Seeder) seedTheme(ctx context.Context, theme *models.Template) error {
	existing, err := s.repo.GetByNameAndType(ctx, theme.Name, theme.Type)
	if err == nil && existing != nil {
		theme.ID = existing.ID
		theme.Version = existing.Version
		theme.CreatedAt = existing.CreatedAt
		theme.CreatedBy = existing.CreatedBy
		return s.repo.Update(ctx, theme)
	}

	var notFoundErr *models.NotFoundError
	if err != nil {
		switch e := err.(type) {
		case *models.NotFoundError:
			notFoundErr = e
		default:
			return fmt.Errorf("failed to check for existing theme: %w", err)
		}
	}

	if notFoundErr == nil && existing != nil {
		return nil
	}

	return s.repo.Create(ctx, theme)
}

func (s *Seeder) getDefaultThemes() []*models.Template {
	return []*models.Template{
		{
			Name:            "Simple & Clean",
			Type:            models.TemplateTypeRSVPPage,
			Category:        models.CategoryPlainText,
			Description:     "Minimalist text-based invitation, perfect for accessibility and fast loading",
			HTMLContent:     s.loadThemeTemplate("plain-text.html"),
			CSSContent:      stringPtr(s.loadThemeCSS("plain-text.css")),
			ComponentConfig: stringPtr(s.loadComponentConfig("plain-text.json")),
			ThumbnailURL:    stringPtr("/static/images/themes/plain-text-thumb.svg"),
			ImageURL:        nil,
			Tags:            []string{"accessible", "minimal", "text-only"},
			SortOrder:       0,
			IsDefault:       true,
			IsActive:        true,
			Version:         1,
			CreatedBy:       0,
		},
		{
			Name:            "Wedding Elegance",
			Type:            models.TemplateTypeRSVPPage,
			Category:        models.CategoryWeddingElegance,
			Description:     "Elegant floral design perfect for weddings and formal celebrations",
			HTMLContent:     s.loadThemeTemplate("wedding-elegance.html"),
			CSSContent:      stringPtr(s.loadThemeCSS("wedding-elegance.css")),
			ComponentConfig: stringPtr(s.loadComponentConfig("wedding-elegance.json")),
			ThumbnailURL:    stringPtr("/static/images/themes/wedding-elegance-thumb.svg"),
			ImageURL:        stringPtr("/static/images/themes/wedding-elegance-header.svg"),
			Tags:            []string{"wedding", "formal", "floral", "elegant"},
			SortOrder:       1,
			IsDefault:       false,
			IsActive:        true,
			Version:         1,
			CreatedBy:       0,
		},
		{
			Name:            "Birthday Celebration",
			Type:            models.TemplateTypeRSVPPage,
			Category:        models.CategoryBirthdayCelebration,
			Description:     "Fun and colorful design for birthday parties and celebrations",
			HTMLContent:     s.loadThemeTemplate("birthday-celebration.html"),
			CSSContent:      stringPtr(s.loadThemeCSS("birthday-celebration.css")),
			ComponentConfig: stringPtr(s.loadComponentConfig("birthday-celebration.json")),
			ThumbnailURL:    stringPtr("/static/images/themes/birthday-celebration-thumb.svg"),
			ImageURL:        stringPtr("/static/images/themes/birthday-celebration-header.svg"),
			Tags:            []string{"birthday", "celebration", "fun", "colorful"},
			SortOrder:       2,
			IsDefault:       false,
			IsActive:        true,
			Version:         1,
			CreatedBy:       0,
		},
		{
			Name:            "Corporate Professional",
			Type:            models.TemplateTypeRSVPPage,
			Category:        models.CategoryCorporatePro,
			Description:     "Clean and professional design for business events and meetings",
			HTMLContent:     s.loadThemeTemplate("corporate-professional.html"),
			CSSContent:      stringPtr(s.loadThemeCSS("corporate-professional.css")),
			ComponentConfig: stringPtr(s.loadComponentConfig("corporate-professional.json")),
			ThumbnailURL:    stringPtr("/static/images/themes/corporate-professional-thumb.svg"),
			ImageURL:        stringPtr("/static/images/themes/corporate-professional-header.svg"),
			Tags:            []string{"corporate", "professional", "business", "formal"},
			SortOrder:       3,
			IsDefault:       false,
			IsActive:        true,
			Version:         1,
			CreatedBy:       0,
		},
		{
			Name:            "Holiday Festive",
			Type:            models.TemplateTypeRSVPPage,
			Category:        models.CategoryHolidayFestive,
			Description:     "Warm and festive design for holiday gatherings and seasonal events",
			HTMLContent:     s.loadThemeTemplate("holiday-festive.html"),
			CSSContent:      stringPtr(s.loadThemeCSS("holiday-festive.css")),
			ComponentConfig: stringPtr(s.loadComponentConfig("holiday-festive.json")),
			ThumbnailURL:    stringPtr("/static/images/themes/holiday-festive-thumb.svg"),
			ImageURL:        stringPtr("/static/images/themes/holiday-festive-header.svg"),
			Tags:            []string{"holiday", "festive", "seasonal", "warm"},
			SortOrder:       4,
			IsDefault:       false,
			IsActive:        true,
			Version:         1,
			CreatedBy:       0,
		},
		{
			Name:            "Garden Party",
			Type:            models.TemplateTypeRSVPPage,
			Category:        models.CategoryGardenParty,
			Description:     "Fresh botanical design for outdoor events and garden parties",
			HTMLContent:     s.loadThemeTemplate("garden-party.html"),
			CSSContent:      stringPtr(s.loadThemeCSS("garden-party.css")),
			ComponentConfig: stringPtr(s.loadComponentConfig("garden-party.json")),
			ThumbnailURL:    stringPtr("/static/images/themes/garden-party-thumb.svg"),
			ImageURL:        stringPtr("/static/images/themes/garden-party-header.svg"),
			Tags:            []string{"garden", "nature", "outdoor", "botanical"},
			SortOrder:       5,
			IsDefault:       false,
			IsActive:        true,
			Version:         1,
			CreatedBy:       0,
		},
		{
			Name:            "Modern Minimalist",
			Type:            models.TemplateTypeRSVPPage,
			Category:        models.CategoryModernMinimalist,
			Description:     "Contemporary minimal design with clean lines and bold typography",
			HTMLContent:     s.loadThemeTemplate("modern-minimalist.html"),
			CSSContent:      stringPtr(s.loadThemeCSS("modern-minimalist.css")),
			ComponentConfig: stringPtr(s.loadComponentConfig("modern-minimalist.json")),
			ThumbnailURL:    stringPtr("/static/images/themes/modern-minimalist-thumb.svg"),
			ImageURL:        stringPtr("/static/images/themes/modern-minimalist-header.svg"),
			Tags:            []string{"modern", "minimal", "contemporary", "clean"},
			SortOrder:       6,
			IsDefault:       false,
			IsActive:        true,
			Version:         1,
			CreatedBy:       0,
		},
	}
}

func (s *Seeder) loadThemeTemplate(filename string) string {
	paths := []string{
		filepath.Join("templates/web/rsvp_themes", filename),
		filepath.Join("../../templates/web/rsvp_themes", filename),
	}

	for _, path := range paths {
		content, err := os.ReadFile(path)
		if err == nil {
			return string(content)
		}
	}

	slog.Warn("Failed to load theme template from disk; theme will be empty. Ensure WORKDIR contains templates/web/rsvp_themes/", "filename", filename)
	return ""
}

func (s *Seeder) loadThemeCSS(filename string) string {
	paths := []string{
		filepath.Join("static/css/themes", filename),
		filepath.Join("../../static/css/themes", filename),
	}

	for _, path := range paths {
		content, err := os.ReadFile(path)
		if err == nil {
			return string(content)
		}
	}

	slog.Warn("Failed to load theme CSS from disk; theme styling will be missing. Ensure WORKDIR contains static/css/themes/", "filename", filename)
	return ""
}

func (s *Seeder) loadComponentConfig(filename string) string {
	embeddedPath := filepath.Join("defaults/component_configs", filename)
	content, err := defaultTemplates.ReadFile(embeddedPath)
	if err == nil {
		return string(content)
	}

	paths := []string{
		filepath.Join("internal/templates/defaults/component_configs", filename),
		filepath.Join("defaults/component_configs", filename),
	}

	for _, path := range paths {
		content, err := os.ReadFile(path)
		if err == nil {
			return string(content)
		}
	}

	slog.Warn("Failed to load component config from disk", "filename", filename)
	return ""
}
