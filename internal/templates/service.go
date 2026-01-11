package templates

import (
	"context"
	"fmt"
	"io"
	"time"

	"github.com/lenaxia/tinyrsvp/internal/auth"
	"github.com/lenaxia/tinyrsvp/internal/db/repositories"
	"github.com/lenaxia/tinyrsvp/internal/models"
)

type Service interface {
	CreateTemplate(ctx context.Context, template *models.Template) error
	GetTemplate(ctx context.Context, id int64) (*models.Template, error)
	GetTemplateForEvent(ctx context.Context, eventID int64, templateType models.TemplateType) (*models.Template, error)
	GetDefaultTemplate(ctx context.Context, templateType models.TemplateType) (*models.Template, error)
	UpdateTemplate(ctx context.Context, template *models.Template) error
	DeleteTemplate(ctx context.Context, id int64) error
	SetActive(ctx context.Context, id int64, active bool) error
	SetDefault(ctx context.Context, id int64) error
	ListTemplates(ctx context.Context, filters *repositories.TemplateFilters) ([]*models.Template, error)
	PreviewTemplate(ctx context.Context, req *PreviewRequest) (*PreviewResponse, error)
	GetComponentRenderer() *ComponentRenderer
	RenderRSVPPage(w io.Writer, event *models.Event, template *models.Template) error
}

type service struct {
	repo              repositories.TemplateRepository
	validator         Validator
	componentRenderer *ComponentRenderer
}

func NewService(repo repositories.TemplateRepository, validator Validator) Service {
	engine := NewEngine()
	return &service{
		repo:              repo,
		validator:         validator,
		componentRenderer: NewComponentRenderer(engine),
	}
}

func (s *service) GetComponentRenderer() *ComponentRenderer {
	return s.componentRenderer
}

func (s *service) CreateTemplate(ctx context.Context, template *models.Template) error {
	user, ok := auth.UserFromContext(ctx)
	if !ok || user.ID == 0 {
		return &models.UnauthorizedError{Message: "Authentication required"}
	}

	template.CreatedBy = user.ID
	template.CreatedAt = time.Now()
	template.UpdatedAt = time.Now()
	template.Version = 1
	template.IsActive = true

	if err := s.validator.ValidateTemplate(template); err != nil {
		return err
	}

	if err := s.repo.Create(ctx, template); err != nil {
		return fmt.Errorf("failed to create template: %w", err)
	}

	return nil
}

func (s *service) GetTemplate(ctx context.Context, id int64) (*models.Template, error) {
	return s.repo.GetByID(ctx, id)
}

func (s *service) GetTemplateForEvent(ctx context.Context, eventID int64, templateType models.TemplateType) (*models.Template, error) {
	template, err := s.repo.GetByEventAndType(ctx, eventID, templateType)
	if err == nil && template.IsActive {
		return template, nil
	}

	return s.repo.GetDefaultByType(ctx, templateType)
}

func (s *service) GetDefaultTemplate(ctx context.Context, templateType models.TemplateType) (*models.Template, error) {
	return s.repo.GetDefaultByType(ctx, templateType)
}

func (s *service) UpdateTemplate(ctx context.Context, template *models.Template) error {
	existing, err := s.repo.GetByID(ctx, template.ID)
	if err != nil {
		return err
	}

	user, ok := auth.UserFromContext(ctx)
	if !ok || user.ID == 0 {
		return &models.UnauthorizedError{Message: "Authentication required"}
	}

	if user.Role != models.RoleAdmin && existing.CreatedBy != user.ID {
		return &models.ForbiddenError{Message: "You can only edit your own templates"}
	}

	if existing.IsDefault && user.Role != models.RoleAdmin {
		return &models.ForbiddenError{Message: "Only admins can edit default templates"}
	}

	if err := s.validator.ValidateTemplate(template); err != nil {
		return err
	}

	template.UpdatedAt = time.Now()

	if err := s.repo.Update(ctx, template); err != nil {
		return fmt.Errorf("failed to update template: %w", err)
	}

	return nil
}

func (s *service) DeleteTemplate(ctx context.Context, id int64) error {
	template, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return err
	}

	user, ok := auth.UserFromContext(ctx)
	if !ok || user.ID == 0 {
		return &models.UnauthorizedError{Message: "Authentication required"}
	}

	if user.Role != models.RoleAdmin && template.CreatedBy != user.ID {
		return &models.ForbiddenError{Message: "You can only delete your own templates"}
	}

	if template.IsDefault {
		return &models.ValidationError{
			Field:   "template",
			Message: "Cannot delete default system templates",
		}
	}

	inUse, err := s.repo.IsTemplateInUse(ctx, id)
	if err != nil {
		return err
	}

	if inUse {
		return &models.ValidationError{
			Field:   "template",
			Message: "Cannot delete template that is in use by events",
		}
	}

	if err := s.repo.Delete(ctx, id); err != nil {
		return fmt.Errorf("failed to delete template: %w", err)
	}

	return nil
}

func (s *service) SetActive(ctx context.Context, id int64, active bool) error {
	user, ok := auth.UserFromContext(ctx)
	if !ok || user.ID == 0 {
		return &models.UnauthorizedError{Message: "Authentication required"}
	}

	template, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return err
	}

	if user.Role != models.RoleAdmin && template.CreatedBy != user.ID {
		return &models.ForbiddenError{Message: "You can only modify your own templates"}
	}

	return s.repo.SetActive(ctx, id, active)
}

func (s *service) SetDefault(ctx context.Context, id int64) error {
	user, ok := auth.UserFromContext(ctx)
	if !ok || user.ID == 0 {
		return &models.UnauthorizedError{Message: "Authentication required"}
	}

	if user.Role != models.RoleAdmin {
		return &models.ForbiddenError{Message: "Only admins can set default templates"}
	}

	return s.repo.SetDefault(ctx, id)
}

func (s *service) ListTemplates(ctx context.Context, filters *repositories.TemplateFilters) ([]*models.Template, error) {
	return s.repo.List(ctx, filters)
}

func (s *service) PreviewTemplate(ctx context.Context, req *PreviewRequest) (*PreviewResponse, error) {
	if req == nil {
		return nil, fmt.Errorf("preview request cannot be nil")
	}

	if req.HTMLContent == "" {
		return nil, &models.ValidationError{
			Field:   "html_content",
			Message: "HTML content is required",
		}
	}

	if !req.Type.IsValid() {
		return nil, &models.ValidationError{
			Field:   "type",
			Message: "Invalid template type",
		}
	}

	if err := s.validator.ValidateSize(req.HTMLContent, 100*1024); err != nil {
		return nil, err
	}

	if err := s.validator.ValidateSyntax(req.HTMLContent, req.Type); err != nil {
		validationErr, ok := err.(*models.ValidationError)
		if ok {
			validationErr.Field = "html_content"
		}
		return nil, err
	}

	allowedVars := getAllowedVariables(req.Type)
	if err := s.validator.ValidateVariables(req.HTMLContent, allowedVars); err != nil {
		validationErr, ok := err.(*models.ValidationError)
		if ok {
			validationErr.Field = "html_content"
		}
		return nil, err
	}

	testData := CreateTestData(req.Type)
	if testData == nil {
		return nil, &models.ValidationError{
			Field:   "type",
			Message: "Invalid template type",
		}
	}

	engine := NewEngine()

	htmlTemplate, err := engine.Parse(req.HTMLContent)
	if err != nil {
		return nil, &models.ValidationError{
			Field:   "html_content",
			Message: fmt.Sprintf("Failed to parse HTML template: %v", err),
		}
	}

	htmlPreview, err := engine.ExecuteToString(htmlTemplate, testData)
	if err != nil {
		return nil, &models.ValidationError{
			Field:   "html_content",
			Message: fmt.Sprintf("Failed to render HTML template: %v", err),
		}
	}

	response := &PreviewResponse{
		HTMLPreview: htmlPreview,
	}

	if req.TextContent != nil && *req.TextContent != "" {
		if err := s.validator.ValidateSize(*req.TextContent, 50*1024); err != nil {
			return nil, err
		}

		if err := s.validator.ValidateSyntax(*req.TextContent, req.Type); err != nil {
			validationErr, ok := err.(*models.ValidationError)
			if ok {
				validationErr.Field = "text_content"
			}
			return nil, err
		}

		if err := s.validator.ValidateVariables(*req.TextContent, allowedVars); err != nil {
			validationErr, ok := err.(*models.ValidationError)
			if ok {
				validationErr.Field = "text_content"
			}
			return nil, err
		}

		textTemplate, err := engine.Parse(*req.TextContent)
		if err != nil {
			return nil, &models.ValidationError{
				Field:   "text_content",
				Message: fmt.Sprintf("Failed to parse text template: %v", err),
			}
		}

		textPreview, err := engine.ExecuteToString(textTemplate, testData)
		if err != nil {
			return nil, &models.ValidationError{
				Field:   "text_content",
				Message: fmt.Sprintf("Failed to render text template: %v", err),
			}
		}

		response.TextPreview = textPreview
	}

	return response, nil
}

func (s *service) RenderRSVPPage(w io.Writer, event *models.Event, template *models.Template) error {
	if template == nil {
		return fmt.Errorf("template cannot be nil")
	}

	if event == nil {
		return fmt.Errorf("event cannot be nil")
	}

	if template.ComponentConfig != nil && *template.ComponentConfig != "" {
		return s.componentRenderer.Render(w, event, template)
	}

	return s.renderLegacyHTML(w, event, template)
}

func (s *service) renderLegacyHTML(w io.Writer, event *models.Event, template *models.Template) error {
	data := map[string]interface{}{
		"Event":    event,
		"Template": template,
	}

	tmpl, err := NewEngine().Parse(template.HTMLContent)
	if err != nil {
		return fmt.Errorf("failed to parse legacy template: %w", err)
	}

	engine := NewEngine()
	if err := engine.Execute(w, tmpl, data); err != nil {
		return fmt.Errorf("failed to execute legacy template: %w", err)
	}

	return nil
}
