package templates

import (
	"context"
	"testing"

	"github.com/lenaxia/tinyrsvp/internal/models"
)

func TestCoverage_EngineFuncs(t *testing.T) {
	if result := iterate(5); len(result) != 5 {
		t.Errorf("iterate(5) returned %d items, want 5", len(result))
	}
	if result := iterate(0); len(result) != 0 {
		t.Errorf("iterate(0) returned %d items, want 0", len(result))
	}
	if result := add(2, 3); result != 5 {
		t.Errorf("add(2,3) = %d, want 5", result)
	}
	if result := add(-1, 1); result != 0 {
		t.Errorf("add(-1,1) = %d, want 0", result)
	}
}

func TestCoverage_ServiceGetDefaultTemplate(t *testing.T) {
	mockRepo := &mockServiceTemplateRepository{
		GetDefaultByTypeFunc: func(ctx context.Context, tt models.TemplateType) (*models.Template, error) {
			return &models.Template{ID: 1, Type: tt}, nil
		},
	}
	svc := NewService(mockRepo, &mockServiceValidator{})

	tmpl, err := svc.GetDefaultTemplate(context.Background(), models.TemplateTypeInviteEmail)
	if err != nil {
		t.Fatalf("GetDefaultTemplate: %v", err)
	}
	if tmpl.ID != 1 {
		t.Errorf("template ID = %d, want 1", tmpl.ID)
	}
}

func TestCoverage_ServiceGetComponentRenderer(t *testing.T) {
	svc := NewService(&mockServiceTemplateRepository{}, &mockServiceValidator{})
	if svc.GetComponentRenderer() == nil {
		t.Error("GetComponentRenderer() returned nil")
	}
}

func TestCoverage_ServiceRenderEmailTemplate(t *testing.T) {
	htmlContent := "<html><body>Hello {{.Name}}</body></html>"
	textContent := "Hello {{.Name}}"
	mockRepo := &mockServiceTemplateRepository{
		GetByEventAndTypeFunc: func(ctx context.Context, eventID int64, tt models.TemplateType) (*models.Template, error) {
			return &models.Template{
				ID:           1,
				Type:         tt,
				IsActive:     true,
				HTMLContent:  htmlContent,
				TextContent:  &textContent,
			}, nil
		},
	}
	svc := NewService(mockRepo, &mockServiceValidator{})

	data := struct{ Name string }{Name: "World"}
	htmlBody, textBody, err := svc.RenderEmailTemplate(context.Background(), 1, models.TemplateTypeInviteEmail, data)
	if err != nil {
		t.Fatalf("RenderEmailTemplate: %v", err)
	}
	if htmlBody == "" {
		t.Error("expected non-empty htmlBody")
	}
	if textBody == "" {
		t.Error("expected non-empty textBody")
	}
}

func TestCoverage_ComponentRenderer_DeepMergeDimensions(t *testing.T) {
	r := NewComponentRenderer(NewEngine())
	dim := &models.Dimensions{}
	r.deepMergeDimensions(dim, map[string]interface{}{
		"width":  "100px",
		"height": "200px",
	})
	if dim.Width != "100px" {
		t.Errorf("Width = %q, want 100px", dim.Width)
	}
	if dim.Height != "200px" {
		t.Errorf("Height = %q, want 200px", dim.Height)
	}
}

func TestCoverage_ComponentRenderer_MergeBackgroundContent(t *testing.T) {
	r := NewComponentRenderer(NewEngine())
	content := &models.BackgroundContent{}
	r.mergeBackgroundContent(content, map[string]interface{}{
		"type":              "gradient",
		"color":             "#fff",
		"gradient":          "linear-gradient(...)",
		"image":             "bg.jpg",
		"backgroundSize":    "cover",
		"backgroundPosition": "center",
	})
	if content.Type != "gradient" {
		t.Errorf("Type = %q", content.Type)
	}
	if content.Color != "#fff" {
		t.Errorf("Color = %q", content.Color)
	}
}

func TestCoverage_ComponentRenderer_MergeOverlayContent(t *testing.T) {
	r := NewComponentRenderer(NewEngine())
	content := &models.OverlayContent{}
	r.mergeOverlayContent(content, map[string]interface{}{
		"backgroundColor":  "rgba(0,0,0,0.5)",
		"backgroundImage":  "overlay.png",
		"backgroundSize":   "cover",
		"borderRadius":     "8px",
		"border":           "1px solid #ccc",
		"boxShadow":        "0 2px 4px rgba(0,0,0,0.1)",
		"clipPath":         "circle(50%)",
		"placeholder":      true,
	})
	if content.BackgroundColor != "rgba(0,0,0,0.5)" {
		t.Errorf("BackgroundColor = %q", content.BackgroundColor)
	}
	if !content.Placeholder {
		t.Error("Placeholder should be true")
	}
}

func TestCoverage_ComponentRenderer_DeepMergeLayout(t *testing.T) {
	r := NewComponentRenderer(NewEngine())
	comp := &models.Component{Type: models.ComponentTypeContainer}
	r.deepMergeLayout(comp, map[string]interface{}{
		"display":        "flex",
		"flexDirection":  "column",
		"justifyContent": "center",
		"alignItems":     "stretch",
		"gap":            "10px",
		"padding":        "20px",
		"children":       []interface{}{"child1", "child2"},
	})
	if comp.Layout == nil {
		t.Fatal("Layout should be initialized")
	}
	if comp.Layout.Display != "flex" {
		t.Errorf("Display = %q", comp.Layout.Display)
	}
	if len(comp.Layout.Children) != 2 {
		t.Errorf("Children count = %d, want 2", len(comp.Layout.Children))
	}
}

func TestCoverage_ComponentRenderer_DeepMergeLayout_NonContainer(t *testing.T) {
	r := NewComponentRenderer(NewEngine())
	comp := &models.Component{Type: models.ComponentTypeTextBox}
	err := r.deepMergeLayout(comp, map[string]interface{}{"display": "flex"})
	if err != nil {
		t.Errorf("deepMergeLayout on non-container: %v", err)
	}
	if comp.Layout != nil {
		t.Error("Layout should remain nil for non-container")
	}
}

func TestCoverage_ComponentRenderer_DeepMergeStyle(t *testing.T) {
	r := NewComponentRenderer(NewEngine())
	comp := &models.Component{Type: models.ComponentTypeDivider}
	r.deepMergeStyle(comp, map[string]interface{}{
		"backgroundColor": "#ccc",
		"height":          "2px",
		"margin":          "10px 0",
		"borderRadius":    "1px",
	})
	if comp.Style == nil {
		t.Fatal("Style should be initialized")
	}
	if comp.Style.BackgroundColor != "#ccc" {
		t.Errorf("BackgroundColor = %q", comp.Style.BackgroundColor)
	}
}

func TestCoverage_ComponentRenderer_DeepMergeStyle_NonDivider(t *testing.T) {
	r := NewComponentRenderer(NewEngine())
	comp := &models.Component{Type: models.ComponentTypeTextBox}
	err := r.deepMergeStyle(comp, map[string]interface{}{"backgroundColor": "#ccc"})
	if err != nil {
		t.Errorf("deepMergeStyle on non-divider: %v", err)
	}
	if comp.Style != nil {
		t.Error("Style should remain nil for non-divider")
	}
}

func TestCoverage_ComponentRenderer_DeepMergeResponsive(t *testing.T) {
	r := NewComponentRenderer(NewEngine())
	comp := &models.Component{Type: models.ComponentTypeTextBox}
	r.deepMergeResponsive(comp, map[string]interface{}{
		"mobile": map[string]interface{}{
			"width":    "100%",
			"fontSize": "14px",
		},
		"tablet": map[string]interface{}{
			"width": "50%",
		},
		"desktop": map[string]interface{}{
			"width": "33%",
		},
	})
	if comp.Responsive == nil {
		t.Fatal("Responsive should be initialized")
	}
	if comp.Responsive.Mobile == nil || comp.Responsive.Mobile.Width != "100%" {
		t.Error("Mobile responsive not set correctly")
	}
	if comp.Responsive.Tablet == nil || comp.Responsive.Tablet.Width != "50%" {
		t.Error("Tablet responsive not set correctly")
	}
	if comp.Responsive.Desktop == nil || comp.Responsive.Desktop.Width != "33%" {
		t.Error("Desktop responsive not set correctly")
	}
}

func TestCoverage_EditorService_DeepCopyContent(t *testing.T) {
	svc := NewEditorService(nil).(*editorService)
	textBox := models.TextBoxContent{Text: "hello"}
	image := models.ImageContent{Src: "img.png"}
	background := models.BackgroundContent{Color: "#fff"}
	overlay := models.OverlayContent{BackgroundColor: "rgba(0,0,0,0.5)"}

	content := &models.ComponentContent{
		TextBox:    &textBox,
		Image:      &image,
		Background: &background,
		Overlay:    &overlay,
	}

	copied := svc.deepCopyContent(content)
	if copied == nil {
		t.Fatal("expected non-nil copy")
	}
	if copied.TextBox.Text != "hello" {
		t.Error("TextBox not copied")
	}
	if copied.Image.Src != "img.png" {
		t.Error("Image not copied")
	}

	if svc.deepCopyContent(nil) != nil {
		t.Error("nil input should return nil")
	}
}

func TestCoverage_EditorService_DeepCopyLayout(t *testing.T) {
	svc := NewEditorService(nil).(*editorService)
	layout := &models.ContainerLayout{
		Display:        "flex",
		FlexDirection:  "row",
		JustifyContent: "center",
		AlignItems:     "stretch",
		Gap:            "10px",
		Padding:        "20px",
		Children:       []string{"a", "b", "c"},
	}

	copied := svc.deepCopyLayout(layout)
	if copied == nil {
		t.Fatal("expected non-nil copy")
	}
	if copied.Display != "flex" {
		t.Error("Display not copied")
	}
	if len(copied.Children) != 3 {
		t.Errorf("Children count = %d, want 3", len(copied.Children))
	}

	if svc.deepCopyLayout(nil) != nil {
		t.Error("nil input should return nil")
	}
}

func TestCoverage_EditorService_DeepCopyStyle(t *testing.T) {
	svc := NewEditorService(nil).(*editorService)
	style := &models.DividerStyle{
		BackgroundColor: "#ccc",
		Height:          "2px",
		Margin:          "10px",
		BorderRadius:    "1px",
	}

	copied := svc.deepCopyStyle(style)
	if copied == nil {
		t.Fatal("expected non-nil copy")
	}
	if copied.BackgroundColor != "#ccc" {
		t.Error("BackgroundColor not copied")
	}

	if svc.deepCopyStyle(nil) != nil {
		t.Error("nil input should return nil")
	}
}

func TestCoverage_EditorService_DeepCopyResponsive(t *testing.T) {
	svc := NewEditorService(nil).(*editorService)
	mobile := models.ResponsiveBreakpoint{Width: "100%"}
	tablet := models.ResponsiveBreakpoint{Width: "50%"}
	desktop := models.ResponsiveBreakpoint{Width: "33%"}

	resp := &models.ResponsiveConfig{
		Mobile:  &mobile,
		Tablet:  &tablet,
		Desktop: &desktop,
	}

	copied := svc.deepCopyResponsive(resp)
	if copied == nil {
		t.Fatal("expected non-nil copy")
	}
	if copied.Mobile.Width != "100%" {
		t.Error("Mobile not copied")
	}
	if copied.Tablet.Width != "50%" {
		t.Error("Tablet not copied")
	}
	if copied.Desktop.Width != "33%" {
		t.Error("Desktop not copied")
	}

	if svc.deepCopyResponsive(nil) != nil {
		t.Error("nil input should return nil")
	}
}

func TestCoverage_MockServiceTemplateRepository_Defaults(t *testing.T) {
	m := &mockServiceTemplateRepository{}
	ctx := context.Background()

	if templates, err := m.GetTemplatesByCategory(ctx, models.CategoryPlain); err != nil || len(templates) != 0 {
		t.Errorf("GetTemplatesByCategory default: templates=%v err=%v", templates, err)
	}
	if templates, err := m.ListThemes(ctx, models.TemplateTypeInviteEmail, nil); err != nil || len(templates) != 0 {
		t.Errorf("ListThemes default: templates=%v err=%v", templates, err)
	}
	if config, err := m.GetComponentConfig(ctx, 1); err != nil || config != nil {
		t.Errorf("GetComponentConfig default: config=%v err=%v", config, err)
	}
	if err := m.UpdateComponentConfig(ctx, 1, nil); err != nil {
		t.Errorf("UpdateComponentConfig default: err=%v", err)
	}
	if err := m.ValidateComponentConfig(ctx, nil); err != nil {
		t.Errorf("ValidateComponentConfig default: err=%v", err)
	}
}

func TestCoverage_MockServiceValidator_Defaults(t *testing.T) {
	m := &mockServiceValidator{}
	if err := m.ValidateSyntax("content", models.TemplateTypeInviteEmail); err != nil {
		t.Errorf("ValidateSyntax default: err=%v", err)
	}
	if err := m.ValidateVariables("content", []string{"var1"}); err != nil {
		t.Errorf("ValidateVariables default: err=%v", err)
	}
	if err := m.ValidateSize("content", 1000); err != nil {
		t.Errorf("ValidateSize default: err=%v", err)
	}
}

func TestCoverage_EditorService_DeepCopyMap(t *testing.T) {
	svc := NewEditorService(nil).(*editorService)

	original := map[string]interface{}{
		"key1": "value1",
		"nested": map[string]interface{}{
			"inner": "val",
		},
	}

	copied := svc.deepCopyMap(original)
	if copied == nil {
		t.Fatal("expected non-nil copy")
	}
	if copied["key1"] != "value1" {
		t.Error("key1 not copied")
	}
	nested, ok := copied["nested"].(map[string]interface{})
	if !ok {
		t.Fatal("nested map not copied as map")
	}
	if nested["inner"] != "val" {
		t.Error("nested inner not copied")
	}

	if svc.deepCopyMap(nil) != nil {
		t.Error("nil input should return nil")
	}
}
