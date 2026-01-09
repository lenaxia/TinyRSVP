package templates

import (
	"fmt"
	"html/template"
	"regexp"
	"text/template/parse"
	"time"

	"github.com/lenaxia/tinyrsvp/internal/models"
)

type Validator interface {
	ValidateTemplate(tmpl *models.Template) error
	ValidateSyntax(content string, templateType models.TemplateType) error
	ValidateVariables(content string, allowedVars []string) error
	ValidateSize(content string, maxBytes int) error
}

type validator struct {
	engine *Engine
}

func NewValidator(engine *Engine) Validator {
	return &validator{
		engine: engine,
	}
}

func (v *validator) ValidateTemplate(tmpl *models.Template) error {
	if err := tmpl.Validate(); err != nil {
		return err
	}

	if err := v.ValidateSize(tmpl.HTMLContent, 100*1024); err != nil {
		return err
	}

	allowedVars := getAllowedVariables(tmpl.Type)
	if err := v.ValidateVariables(tmpl.HTMLContent, allowedVars); err != nil {
		return err
	}

	if err := v.ValidateSyntax(tmpl.HTMLContent, tmpl.Type); err != nil {
		return err
	}

	if tmpl.TextContent != nil {
		if err := v.ValidateSize(*tmpl.TextContent, 50*1024); err != nil {
			return err
		}

		if err := v.ValidateSyntax(*tmpl.TextContent, tmpl.Type); err != nil {
			return err
		}

		if err := v.ValidateVariables(*tmpl.TextContent, allowedVars); err != nil {
			return err
		}
	}

	if tmpl.CSSContent != nil {
		if err := v.ValidateSize(*tmpl.CSSContent, 50*1024); err != nil {
			return err
		}
	}

	return nil
}

func (v *validator) ValidateSize(content string, maxBytes int) error {
	if len(content) > maxBytes {
		return &models.ValidationError{
			Field:   "template_content",
			Message: fmt.Sprintf("Template size exceeds %d bytes", maxBytes),
		}
	}
	return nil
}

func (v *validator) ValidateSyntax(content string, templateType models.TemplateType) error {
	testData := createTestData(templateType)

	tmpl, err := v.engine.Parse(content)
	if err != nil {
		return &models.ValidationError{
			Field:   "template_content",
			Message: fmt.Sprintf("Template syntax error: %v", err),
		}
	}

	result, err := v.engine.ExecuteToString(tmpl, testData)
	if err != nil {
		return &models.ValidationError{
			Field:   "template_content",
			Message: fmt.Sprintf("Template execution error: %v", err),
		}
	}

	if result == "" && content != "" {
		return &models.ValidationError{
			Field:   "template_content",
			Message: "Template produced empty output",
		}
	}

	return nil
}

func (v *validator) ValidateVariables(content string, allowedVars []string) error {
	tmpl, err := template.New("check").Funcs(v.engine.funcMap).Parse(content)
	if err != nil {
		return &models.ValidationError{
			Field:   "template_content",
			Message: fmt.Sprintf("Failed to parse template for variable validation: %v", err),
		}
	}

	usedVars := extractVariables(tmpl.Tree.Root)

	allowedMap := make(map[string]bool)
	for _, v := range allowedVars {
		allowedMap[v] = true
	}

	for _, usedVar := range usedVars {
		if !allowedMap[usedVar] {
			return &models.ValidationError{
				Field:   "template_content",
				Message: fmt.Sprintf("Undefined variable: {{.%s}}", usedVar),
			}
		}
	}

	return nil
}

func getAllowedVariables(templateType models.TemplateType) []string {
	common := []string{
		"Event.Title",
		"Event.Description",
		"Event.StartTime",
		"Event.EndTime",
		"Event.Timezone",
		"Event.Location",
		"Event.RSVPDeadline",
	}

	switch templateType {
	case models.TemplateTypeInviteEmail:
		return append(common, []string{
			"Invite.Name",
			"Invite.Email",
			"RSVPURL",
			"MaxPlusOnes",
		}...)
	case models.TemplateTypeRSVPPage:
		return append(common, []string{
			"RSVP.Response",
			"RSVP.PlusOnes",
			"Questions",
		}...)
	case models.TemplateTypeConfirmationPage:
		return append(common, []string{
			"RSVP.Response",
			"RSVP.PlusOnes",
			"Answers",
		}...)
	default:
		return common
	}
}

func extractVariables(node parse.Node) []string {
	vars := make(map[string]bool)
	extractVariablesRecursive(node, vars)

	result := make([]string, 0, len(vars))
	for v := range vars {
		result = append(result, v)
	}
	return result
}

func extractVariablesRecursive(node parse.Node, vars map[string]bool) {
	if node == nil {
		return
	}

	switch n := node.(type) {
	case *parse.ActionNode:
		if n.Pipe != nil {
			for _, cmd := range n.Pipe.Cmds {
				for _, arg := range cmd.Args {
					if fieldNode, ok := arg.(*parse.FieldNode); ok {
						if len(fieldNode.Ident) > 0 {
							varPath := ""
							for i, ident := range fieldNode.Ident {
								if i > 0 {
									varPath += "."
								}
								varPath += ident
							}
							vars[varPath] = true
						}
					}
				}
			}
		}
	case *parse.ListNode:
		if n.Nodes != nil {
			for _, child := range n.Nodes {
				extractVariablesRecursive(child, vars)
			}
		}
	case *parse.IfNode:
		if n.Pipe != nil {
			for _, cmd := range n.Pipe.Cmds {
				for _, arg := range cmd.Args {
					if fieldNode, ok := arg.(*parse.FieldNode); ok {
						if len(fieldNode.Ident) > 0 {
							varPath := ""
							for i, ident := range fieldNode.Ident {
								if i > 0 {
									varPath += "."
								}
								varPath += ident
							}
							vars[varPath] = true
						}
					}
				}
			}
		}
		if n.List != nil {
			extractVariablesRecursive(n.List, vars)
		}
		if n.ElseList != nil {
			extractVariablesRecursive(n.ElseList, vars)
		}
	case *parse.WithNode:
		if n.Pipe != nil {
			for _, cmd := range n.Pipe.Cmds {
				for _, arg := range cmd.Args {
					if fieldNode, ok := arg.(*parse.FieldNode); ok {
						if len(fieldNode.Ident) > 0 {
							varPath := ""
							for i, ident := range fieldNode.Ident {
								if i > 0 {
									varPath += "."
								}
								varPath += ident
							}
							vars[varPath] = true
						}
					}
				}
			}
		}
		if n.List != nil {
			extractVariablesRecursive(n.List, vars)
		}
		if n.ElseList != nil {
			extractVariablesRecursive(n.ElseList, vars)
		}
	case *parse.RangeNode:
		if n.Pipe != nil {
			for _, cmd := range n.Pipe.Cmds {
				for _, arg := range cmd.Args {
					if fieldNode, ok := arg.(*parse.FieldNode); ok {
						if len(fieldNode.Ident) > 0 {
							varPath := ""
							for i, ident := range fieldNode.Ident {
								if i > 0 {
									varPath += "."
								}
								varPath += ident
							}
							vars[varPath] = true
						}
					}
				}
			}
		}
		if n.List != nil {
			extractVariablesRecursive(n.List, vars)
		}
		if n.ElseList != nil {
			extractVariablesRecursive(n.ElseList, vars)
		}
	}
}

var dangerousPatterns = []*regexp.Regexp{
	regexp.MustCompile(`(?i)<script[^>]*>.*?</script>`),
	regexp.MustCompile(`(?i)javascript:`),
	regexp.MustCompile(`(?i)on\w+\s*=`),
	regexp.MustCompile(`(?i)eval\s*\(`),
	regexp.MustCompile(`(?i)Function\s*\(`),
	regexp.MustCompile(`(?i)setTimeout\s*\(`),
	regexp.MustCompile(`(?i)setInterval\s*\(`),
}

func createTestData(templateType models.TemplateType) interface{} {
	type Event struct {
		Title        string
		Description  string
		StartTime    interface{}
		EndTime      interface{}
		Timezone     string
		Location     string
		RSVPDeadline interface{}
	}

	type Invite struct {
		Name  string
		Email string
	}

	type RSVP struct {
		Response string
		PlusOnes int
	}

	type Question struct {
		Text string
	}

	type Answer struct {
		Question string
		Answer   string
	}

	startTime, _ := time.Parse("2006-01-02 15:04", "2026-01-01 10:00")
	endTime, _ := time.Parse("2006-01-02 15:04", "2026-01-01 12:00")
	deadline, _ := time.Parse("2006-01-02", "2025-12-31")

	event := Event{
		Title:        "Test Event",
		Description:  "Test Description",
		StartTime:    startTime,
		EndTime:      endTime,
		Timezone:     "America/Los_Angeles",
		Location:     "Test Location",
		RSVPDeadline: deadline,
	}

	switch templateType {
	case models.TemplateTypeInviteEmail:
		return struct {
			Event       Event
			Invite      Invite
			RSVPURL     string
			MaxPlusOnes int
		}{
			Event: event,
			Invite: Invite{
				Name:  "Test User",
				Email: "test@example.com",
			},
			RSVPURL:     "https://example.com/rsvp/token",
			MaxPlusOnes: 2,
		}
	case models.TemplateTypeRSVPPage:
		return struct {
			Event     Event
			RSVP      RSVP
			Questions []Question
		}{
			Event: event,
			RSVP: RSVP{
				Response: "yes",
				PlusOnes: 1,
			},
			Questions: []Question{
				{Text: "Dietary restrictions?"},
			},
		}
	case models.TemplateTypeConfirmationPage:
		return struct {
			Event   Event
			RSVP    RSVP
			Answers []Answer
		}{
			Event: event,
			RSVP: RSVP{
				Response: "yes",
				PlusOnes: 1,
			},
			Answers: []Answer{
				{Question: "Dietary restrictions?", Answer: "None"},
			},
		}
	default:
		return struct {
			Event Event
		}{
			Event: event,
		}
	}
}
