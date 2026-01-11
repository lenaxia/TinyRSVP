# Component-Based Template Architecture for TinyRSVP

**Date:** 2026-01-11  
**Author:** Architecture Mode  
**Status:** Design Complete  
**Version:** 1.0

---

## Executive Summary

This document defines a comprehensive **component-based template architecture** that transforms TinyRSVP's invite templates from static HTML/CSS files into flexible, composable layouts where every element (text boxes, images, backgrounds, overlays) can be positioned, styled, and customized at multiple configuration levels.

**Key Innovation:** This is not just CSS theming - it's a **declarative component layout system** where templates are defined as JSON arrays of positioned, styled components that can be overridden per-event.

**Architecture Highlights:**
- Component-based data model with position, dimensions, and styling
- Three-layer configuration: Core Template → Leaf Template → Per-Event Customization
- JSON-based component definitions stored in database
- Merge strategy with component ID-based overrides
- Backward compatible with existing template system

---

## Table of Contents

1. [Current System Analysis](#1-current-system-analysis)
2. [Component Type Definitions](#2-component-type-definitions)
3. [Configuration Schema](#3-configuration-schema)
4. [Database Schema](#4-database-schema)
5. [Rendering Engine Architecture](#5-rendering-engine-architecture)
6. [Migration Strategy](#6-migration-strategy)
7. [Concrete Examples](#7-concrete-examples)
8. [Implementation Roadmap](#8-implementation-roadmap)

---

## 1. Current System Analysis

### 1.1 Current Template System

**What Exists:**
```go
type Template struct {
    ID          int64
    EventID     *int64
    Name        string
    Type        TemplateType  // invite_email, rsvp_page, confirmation_page
    HTMLContent string        // Full HTML template
    CSSContent  *string       // Optional CSS
    Category    TemplateCategory
    ThumbnailURL *string
    ImageURL    *string
    Tags        []string
}

type Event struct {
    TemplateID          *int64
    CustomThemeImageURL *string
    CustomThemeColor    *string
}
```

**Current Approach:**
- Templates are complete HTML documents
- Limited customization via `CustomThemeImageURL` and `CustomThemeColor`
- No component-level control
- No positioning flexibility
- No ability to add/remove elements per event

**Limitations:**
1. Cannot move text boxes to different positions
2. Cannot change individual text box fonts/sizes/colors
3. Cannot add new text boxes per event
4. Cannot layer multiple images (background, foreground, overlays)
5. Cannot remove specific elements
6. Customization limited to entire template replacement

### 1.2 Gap Analysis

**Missing Capabilities:**
- ❌ Component-level positioning (x, y coordinates or flex order)
- ❌ Component-level styling (font, size, color per text box)
- ❌ Component composition (multiple images, text boxes, overlays)
- ❌ Per-event component overrides
- ❌ Component add/remove operations
- ❌ Z-index layering control
- ❌ Transparent overlay support for user photos

---

## 2. Component Type Definitions

### 2.1 Component Base Schema

All components share a common base structure:

```json
{
  "id": "string",           // Unique identifier for override targeting
  "type": "ComponentType",  // Component type discriminator
  "position": {             // Positioning strategy
    "mode": "absolute|relative|flex",
    "x": "number|string",   // Pixels or percentage for absolute
    "y": "number|string",   // Pixels or percentage for absolute
    "order": "number"       // Flex order for relative positioning
  },
  "dimensions": {
    "width": "string",      // CSS width (px, %, auto)
    "height": "string"      // CSS height (px, %, auto)
  },
  "zIndex": "number",       // Layering control
  "visible": "boolean",     // Show/hide toggle
  "className": "string"     // Additional CSS classes
}
```

### 2.2 Component Types

#### TextBox Component

Renders text content with full typography control.

```json
{
  "id": "title-text",
  "type": "TextBox",
  "position": {
    "mode": "absolute",
    "x": "50%",
    "y": "200px"
  },
  "dimensions": {
    "width": "80%",
    "height": "auto"
  },
  "zIndex": 10,
  "visible": true,
  "content": {
    "text": "{{.Event.Title}}",     // Template variable support
    "textAlign": "center",
    "fontFamily": "Playfair Display, serif",
    "fontSize": "48px",
    "fontWeight": "700",
    "color": "#2c3e50",
    "lineHeight": "1.2",
    "letterSpacing": "0.02em",
    "textTransform": "none",      // none, uppercase, lowercase, capitalize
    "textShadow": "2px 2px 4px rgba(0,0,0,0.1)"
  },
  "responsive": {
    "mobile": {
      "fontSize": "32px",
      "width": "90%"
    }
  }
}
```

#### Image Component

Renders images with positioning and styling control.

```json
{
  "id": "header-image",
  "type": "Image",
  "position": {
    "mode": "absolute",
    "x": "0",
    "y": "0"
  },
  "dimensions": {
    "width": "100%",
    "height": "400px"
  },
  "zIndex": 1,
  "visible": true,
  "content": {
    "src": "/static/images/themes/wedding-elegance-header.jpg",
    "alt": "Wedding invitation header",
    "objectFit": "cover",        // cover, contain, fill, none
    "objectPosition": "center",  // CSS object-position
    "opacity": 1.0,
    "filter": "none"             // CSS filter (brightness, contrast, etc.)
  }
}
```

#### Background Component

Special image component for page/card backgrounds.

```json
{
  "id": "page-background",
  "type": "Background",
  "position": {
    "mode": "absolute",
    "x": "0",
    "y": "0"
  },
  "dimensions": {
    "width": "100%",
    "height": "100%"
  },
  "zIndex": 0,
  "visible": true,
  "content": {
    "type": "color|gradient|image",
    "color": "#f8f9fa",
    "gradient": "linear-gradient(135deg, #667eea 0%, #764ba2 100%)",
    "image": {
      "src": "/static/images/backgrounds/texture.jpg",
      "repeat": "no-repeat",
      "size": "cover",
      "position": "center",
      "attachment": "fixed"
    }
  }
}
```

#### Overlay Component

Transparent layer for effects or user photo placement.

```json
{
  "id": "photo-overlay",
  "type": "Overlay",
  "position": {
    "mode": "absolute",
    "x": "50%",
    "y": "100px"
  },
  "dimensions": {
    "width": "200px",
    "height": "200px"
  },
  "zIndex": 20,
  "visible": true,
  "content": {
    "backgroundColor": "transparent",
    "borderRadius": "50%",
    "border": "4px solid white",
    "boxShadow": "0 4px 6px rgba(0,0,0,0.1)",
    "clipPath": "circle(50%)",
    "placeholder": {
      "show": true,
      "text": "Add Photo",
      "icon": "camera"
    }
  }
}
```

#### Container Component

Groups other components with layout control.

```json
{
  "id": "event-details-container",
  "type": "Container",
  "position": {
    "mode": "relative",
    "order": 2
  },
  "dimensions": {
    "width": "100%",
    "height": "auto"
  },
  "zIndex": 5,
  "visible": true,
  "layout": {
    "display": "flex",
    "flexDirection": "column",
    "alignItems": "center",
    "justifyContent": "flex-start",
    "gap": "20px",
    "padding": "40px 20px"
  },
  "style": {
    "backgroundColor": "rgba(255, 255, 255, 0.95)",
    "borderRadius": "8px",
    "boxShadow": "0 2px 8px rgba(0,0,0,0.1)"
  },
  "children": ["date-text", "location-text", "description-text"]
}
```

#### Divider Component

Visual separator between sections.

```json
{
  "id": "section-divider",
  "type": "Divider",
  "position": {
    "mode": "relative",
    "order": 3
  },
  "dimensions": {
    "width": "80%",
    "height": "2px"
  },
  "zIndex": 5,
  "visible": true,
  "style": {
    "backgroundColor": "#e5e7eb",
    "margin": "30px auto",
    "borderRadius": "1px"
  }
}
```

### 2.3 Component Type Enum

```go
type ComponentType string

const (
    ComponentTypeTextBox    ComponentType = "TextBox"
    ComponentTypeImage      ComponentType = "Image"
    ComponentTypeBackground ComponentType = "Background"
    ComponentTypeOverlay    ComponentType = "Overlay"
    ComponentTypeContainer  ComponentType = "Container"
    ComponentTypeDivider    ComponentType = "Divider"
)
```

---

## 3. Configuration Schema

### 3.1 Three-Layer Configuration Model

```
┌─────────────────────────────────────────┐
│  Core Template (Go html/template)      │
│  - Base HTML structure                  │
│  - Component rendering engine           │
│  - Shared across all leaf templates    │
└─────────────────────────────────────────┘
                  ↓
┌─────────────────────────────────────────┐
│  Leaf Template Configuration (JSON)     │
│  - Array of component definitions       │
│  - Default positions and styles         │
│  - Stored in templates.component_config │
└─────────────────────────────────────────┘
                  ↓
┌─────────────────────────────────────────┐
│  Per-Event Customization (JSON)         │
│  - Component overrides by ID            │
│  - New component additions              │
│  - Stored in events.component_overrides │
└─────────────────────────────────────────┘
                  ↓
┌─────────────────────────────────────────┐
│  Final Rendered Template                │
│  - Merged component configuration       │
│  - Applied to core template             │
└─────────────────────────────────────────┘
```

### 3.2 Leaf Template Configuration Schema

```json
{
  "version": "1.0",
  "metadata": {
    "name": "Wedding Elegance",
    "category": "card",
    "description": "Elegant wedding invitation with floral design"
  },
  "layout": {
    "mode": "card",              // card, fullpage, split
    "cardWidth": "800px",
    "cardMaxWidth": "90vw",
    "cardPadding": "0",
    "backgroundColor": "#ffffff"
  },
  "components": [
    {
      "id": "page-background",
      "type": "Background",
      "position": {"mode": "absolute", "x": "0", "y": "0"},
      "dimensions": {"width": "100%", "height": "100%"},
      "zIndex": 0,
      "visible": true,
      "content": {
        "type": "color",
        "color": "#f8f9fa"
      }
    },
    {
      "id": "header-image",
      "type": "Image",
      "position": {"mode": "absolute", "x": "0", "y": "0"},
      "dimensions": {"width": "100%", "height": "400px"},
      "zIndex": 1,
      "visible": true,
      "content": {
        "src": "/static/images/themes/wedding-elegance-header.jpg",
        "alt": "Wedding invitation header",
        "objectFit": "cover",
        "objectPosition": "center"
      }
    },
    {
      "id": "title-text",
      "type": "TextBox",
      "position": {"mode": "absolute", "x": "50%", "y": "450px"},
      "dimensions": {"width": "80%", "height": "auto"},
      "zIndex": 10,
      "visible": true,
      "content": {
        "text": "{{.Event.Title}}",
        "textAlign": "center",
        "fontFamily": "Playfair Display, serif",
        "fontSize": "48px",
        "fontWeight": "700",
        "color": "#2c3e50"
      }
    }
  ]
}
```

### 3.3 Per-Event Override Schema

```json
{
  "version": "1.0",
  "overrides": [
    {
      "id": "title-text",
      "updates": {
        "position": {"y": "500px"},
        "content": {
          "fontSize": "56px",
          "color": "#8b4789"
        }
      }
    },
    {
      "id": "header-image",
      "updates": {
        "content": {
          "src": "/uploads/events/123/custom-header.jpg"
        }
      }
    }
  ],
  "additions": [
    {
      "id": "custom-subtitle",
      "type": "TextBox",
      "position": {"mode": "absolute", "x": "50%", "y": "550px"},
      "dimensions": {"width": "70%", "height": "auto"},
      "zIndex": 10,
      "visible": true,
      "content": {
        "text": "Join us for a celebration",
        "textAlign": "center",
        "fontFamily": "Lato, sans-serif",
        "fontSize": "24px",
        "color": "#666666"
      }
    },
    {
      "id": "couple-photo-overlay",
      "type": "Overlay",
      "position": {"mode": "absolute", "x": "50%", "y": "150px"},
      "dimensions": {"width": "200px", "height": "200px"},
      "zIndex": 20,
      "visible": true,
      "content": {
        "backgroundColor": "transparent",
        "borderRadius": "50%",
        "border": "4px solid white",
        "backgroundImage": "url(/uploads/events/123/couple-photo.jpg)",
        "backgroundSize": "cover"
      }
    }
  ],
  "removals": ["section-divider"]
}
```

### 3.4 Merge Strategy

**Algorithm:**
1. Start with leaf template components array
2. Apply overrides by matching component ID
3. Deep merge override properties (position, content, etc.)
4. Append additions to components array
5. Filter out components in removals array
6. Sort by zIndex for rendering order

**Pseudocode:**
```go
func MergeComponentConfig(leaf, eventOverride ComponentConfig) ComponentConfig {
    result := make([]Component, 0)
    
    // Create map for quick lookup
    overrideMap := make(map[string]ComponentUpdate)
    for _, override := range eventOverride.Overrides {
        overrideMap[override.ID] = override.Updates
    }
    
    removalSet := make(map[string]bool)
    for _, id := range eventOverride.Removals {
        removalSet[id] = true
    }
    
    // Process leaf components
    for _, component := range leaf.Components {
        // Skip if marked for removal
        if removalSet[component.ID] {
            continue
        }
        
        // Apply overrides if present
        if updates, exists := overrideMap[component.ID]; exists {
            component = DeepMerge(component, updates)
        }
        
        result = append(result, component)
    }
    
    // Add new components
    result = append(result, eventOverride.Additions...)
    
    // Sort by zIndex
    sort.Slice(result, func(i, j int) bool {
        return result[i].ZIndex < result[j].ZIndex
    })
    
    return ComponentConfig{Components: result}
}
```

---

## 4. Database Schema

### 4.1 Extended Template Model

```sql
-- Add component configuration column to templates table
ALTER TABLE templates ADD COLUMN component_config TEXT; -- JSON

-- component_config stores the leaf template configuration
-- Example: {"version": "1.0", "layout": {...}, "components": [...]}
```

**Go Model:**
```go
type Template struct {
    ID              int64
    EventID         *int64
    Name            string
    Type            TemplateType
    Description     string
    
    // Legacy fields (maintain for backward compatibility)
    HTMLContent     string
    TextContent     *string
    CSSContent      *string
    
    // New component-based fields
    ComponentConfig *string  // JSON: ComponentConfiguration
    
    IsDefault       bool
    IsActive        bool
    Version         int
    CreatedBy       int64
    CreatedAt       time.Time
    UpdatedAt       time.Time
    
    Category        TemplateCategory
    ThumbnailURL    *string
    ImageURL        *string
    Tags            []string
    SortOrder       int
}

type ComponentConfiguration struct {
    Version    string              `json:"version"`
    Metadata   ConfigMetadata      `json:"metadata"`
    Layout     LayoutConfig        `json:"layout"`
    Components []Component         `json:"components"`
}

type ConfigMetadata struct {
    Name        string `json:"name"`
    Category    string `json:"category"`
    Description string `json:"description"`
}

type LayoutConfig struct {
    Mode            string `json:"mode"`
    CardWidth       string `json:"cardWidth,omitempty"`
    CardMaxWidth    string `json:"cardMaxWidth,omitempty"`
    CardPadding     string `json:"cardPadding,omitempty"`
    BackgroundColor string `json:"backgroundColor,omitempty"`
}
```

### 4.2 Extended Event Model

```sql
-- Add component override column to events table
ALTER TABLE events ADD COLUMN component_overrides TEXT; -- JSON

-- component_overrides stores per-event customizations
-- Example: {"version": "1.0", "overrides": [...], "additions": [...], "removals": [...]}
```

**Go Model:**
```go
type Event struct {
    ID                  int64
    PublicID            *string
    FriendlyName        *string
    Title               string
    Description         *string
    StartTime           time.Time
    EndTime             *time.Time
    Timezone            string
    Location            *string
    Status              EventStatus
    CreatedBy           int64
    Version             int
    ICSSequence         int
    MaxPlusOnes         int
    RSVPDeadline        *time.Time
    TemplateID          *int64
    
    // Legacy customization fields (maintain for backward compatibility)
    CustomThemeImageURL *string
    CustomThemeColor    *string
    
    // New component-based customization
    ComponentOverrides  *string  // JSON: ComponentOverrides
    
    CreatedAt           time.Time
    UpdatedAt           time.Time
}

type ComponentOverrides struct {
    Version   string              `json:"version"`
    Overrides []ComponentOverride `json:"overrides"`
    Additions []Component         `json:"additions"`
    Removals  []string            `json:"removals"`
}

type ComponentOverride struct {
    ID      string                 `json:"id"`
    Updates map[string]interface{} `json:"updates"`
}
```

### 4.3 Component Model

```go
type Component struct {
    ID         string              `json:"id"`
    Type       ComponentType       `json:"type"`
    Position   Position            `json:"position"`
    Dimensions Dimensions          `json:"dimensions"`
    ZIndex     int                 `json:"zIndex"`
    Visible    bool                `json:"visible"`
    ClassName  string              `json:"className,omitempty"`
    
    // Type-specific content (use interface{} for flexibility)
    Content    interface{}         `json:"content"`
    
    // Optional fields for specific component types
    Layout     *LayoutConfig       `json:"layout,omitempty"`
    Style      map[string]string   `json:"style,omitempty"`
    Children   []string            `json:"children,omitempty"`
    Responsive map[string]interface{} `json:"responsive,omitempty"`
}

type Position struct {
    Mode  string      `json:"mode"`  // absolute, relative, flex
    X     interface{} `json:"x,omitempty"`  // string or number
    Y     interface{} `json:"y,omitempty"`  // string or number
    Order int         `json:"order,omitempty"`
}

type Dimensions struct {
    Width  string `json:"width"`
    Height string `json:"height"`
}
```

### 4.4 Migration Script

```sql
-- Migration: 000012_add_component_config.up.sql

-- Add component configuration to templates
ALTER TABLE templates ADD COLUMN component_config TEXT;

-- Add component overrides to events
ALTER TABLE events ADD COLUMN component_overrides TEXT;

-- Create index for faster JSON queries (PostgreSQL example)
-- CREATE INDEX idx_templates_component_config ON templates USING GIN (component_config);
-- CREATE INDEX idx_events_component_overrides ON events USING GIN (component_overrides);

-- For SQLite, we'll rely on full column scans (acceptable for small datasets)

-- Migration: 000012_add_component_config.down.sql
ALTER TABLE templates DROP COLUMN component_config;
ALTER TABLE events DROP COLUMN component_overrides;
```

---

## 5. Rendering Engine Architecture

### 5.1 Core Template Structure

The core template is a single Go html/template that renders components dynamically:

```html
<!DOCTYPE html>
<html lang="en" data-theme="{{.SystemTheme}}">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>{{.Event.Title}}</title>
    <link rel="stylesheet" href="/static/css/variables.css">
    <link rel="stylesheet" href="/static/css/component-renderer.css">
    <script src="/static/js/theme_controller.js"></script>
</head>
<body>
    <div class="component-canvas" style="position: relative;">
        {{range .Components}}
            {{template "component" .}}
        {{end}}
    </div>
    
    {{/* Component rendering templates */}}
    {{define "component"}}
        {{if eq .Type "TextBox"}}
            {{template "textbox" .}}
        {{else if eq .Type "Image"}}
            {{template "image" .}}
        {{else if eq .Type "Background"}}
            {{template "background" .}}
        {{else if eq .Type "Overlay"}}
            {{template "overlay" .}}
        {{else if eq .Type "Container"}}
            {{template "container" .}}
        {{else if eq .Type "Divider"}}
            {{template "divider" .}}
        {{end}}
    {{end}}
    
    {{define "textbox"}}
        <div id="{{.ID}}" 
             class="component component-textbox {{.ClassName}}"
             style="{{template "position-style" .Position}}
                    {{template "dimensions-style" .Dimensions}}
                    z-index: {{.ZIndex}};
                    {{if not .Visible}}display: none;{{end}}
                    text-align: {{.Content.TextAlign}};
                    font-family: {{.Content.FontFamily}};
                    font-size: {{.Content.FontSize}};
                    font-weight: {{.Content.FontWeight}};
                    color: {{.Content.Color}};
                    line-height: {{.Content.LineHeight}};
                    {{if .Content.LetterSpacing}}letter-spacing: {{.Content.LetterSpacing}};{{end}}
                    {{if .Content.TextTransform}}text-transform: {{.Content.TextTransform}};{{end}}
                    {{if .Content.TextShadow}}text-shadow: {{.Content.TextShadow}};{{end}}">
            {{.Content.Text}}
        </div>
    {{end}}
    
    {{define "image"}}
        <img id="{{.ID}}"
             class="component component-image {{.ClassName}}"
             src="{{.Content.Src}}"
             alt="{{.Content.Alt}}"
             style="{{template "position-style" .Position}}
                    {{template "dimensions-style" .Dimensions}}
                    z-index: {{.ZIndex}};
                    {{if not .Visible}}display: none;{{end}}
                    object-fit: {{.Content.ObjectFit}};
                    object-position: {{.Content.ObjectPosition}};
                    opacity: {{.Content.Opacity}};
                    {{if .Content.Filter}}filter: {{.Content.Filter}};{{end}}">
    {{end}}
    
    {{define "position-style"}}
        {{if eq .Mode "absolute"}}
            position: absolute;
            left: {{.X}};
            top: {{.Y}};
            transform: translateX(-50%);
        {{else if eq .Mode "relative"}}
            position: relative;
            order: {{.Order}};
        {{else if eq .Mode "flex"}}
            order: {{.Order}};
        {{end}}
    {{end}}
    
    {{define "dimensions-style"}}
        width: {{.Width}};
        height: {{.Height}};
    {{end}}
</body>
</html>
```

### 5.2 Rendering Service

```go
package templates

type ComponentRenderer struct {
    templateEngine *Engine
    logger         *log.Logger
}

func NewComponentRenderer(engine *Engine) *ComponentRenderer {
    return &ComponentRenderer{
        templateEngine: engine,
        logger:         log.New(os.Stdout, "[ComponentRenderer] ", log.LstdFlags),
    }
}

func (r *ComponentRenderer) Render(w io.Writer, event *models.Event, template *models.Template) error {
    // 1. Parse leaf template component config
    leafConfig, err := r.parseComponentConfig(template.ComponentConfig)
    if err != nil {
        return fmt.Errorf("failed to parse leaf config: %w", err)
    }
    
    // 2. Parse event component overrides
    var eventOverrides *models.ComponentOverrides
    if event.ComponentOverrides != nil {
        eventOverrides, err = r.parseComponentOverrides(*event.ComponentOverrides)
        if err != nil {
            return fmt.Errorf("failed to parse event overrides: %w", err)
        }
    }
    
    // 3. Merge configurations
    finalConfig := r.mergeConfigs(leafConfig, eventOverrides)
    
    // 4. Prepare template data
    data := struct {
        Event        *models.Event
        Components   []models.Component
        SystemTheme  string
    }{
        Event:       event,
        Components:  finalConfig.Components,
        SystemTheme: "light", // Default, will be overridden by JS
    }
    
    // 5. Render core template
    return r.templateEngine.Execute(w, "core-component-template", data)
}

func (r *ComponentRenderer) parseComponentConfig(jsonStr *string) (*models.ComponentConfiguration, error) {
    if jsonStr == nil || *jsonStr == "" {
        return &models.ComponentConfiguration{Components: []models.Component{}}, nil
    }
    
    var config models.ComponentConfiguration
    if err := json.Unmarshal([]byte(*jsonStr), &config); err != nil {
        return nil, err
    }
    
    return &config, nil
}

func (r *ComponentRenderer) parseComponentOverrides(jsonStr string) (*models.ComponentOverrides, error) {
    var overrides models.ComponentOverrides
    if err := json.Unmarshal([]byte(jsonStr), &overrides); err != nil {
        return nil, err
    }
    
    return &overrides, nil
}

func (r *ComponentRenderer) mergeConfigs(leaf *models.ComponentConfiguration, overrides *models.ComponentOverrides) *models.ComponentConfiguration {
    if overrides == nil {
        return leaf
    }
    
    result := &models.ComponentConfiguration{
        Version:    leaf.Version,
        Metadata:   leaf.Metadata,
        Layout:     leaf.Layout,
        Components: make([]models.Component, 0),
    }
    
    // Create override map
    overrideMap := make(map[string]models.ComponentOverride)
    for _, override := range overrides.Overrides {
        overrideMap[override.ID] = override
    }
    
    // Create removal set
    removalSet := make(map[string]bool)
    for _, id := range overrides.Removals {
        removalSet[id] = true
    }
    
    // Process leaf components
    for _, component := range leaf.Components {
        if removalSet[component.ID] {
            continue
        }
        
        if override, exists := overrideMap[component.ID]; exists {
            component = r.applyOverride(component, override)
        }
        
        result.Components = append(result.Components, component)
    }
    
    // Add new components
    result.Components = append(result.Components, overrides.Additions...)
    
    // Sort by zIndex
    sort.Slice(result.Components, func(i, j int) bool {
        return result.Components[i].ZIndex < result.Components[j].ZIndex
    })
    
    return result
}

func (r *ComponentRenderer) applyOverride(component models.Component, override models.ComponentOverride) models.Component {
    // Deep merge using JSON marshal/unmarshal
    componentJSON, _ := json.Marshal(component)
    var componentMap map[string]interface{}
    json.Unmarshal(componentJSON, &componentMap)
    
    // Merge updates
    for key, value := range override.Updates {
        if existingValue, exists := componentMap[key]; exists {
            // Deep merge for nested objects
            if existingMap, ok := existingValue.(map[string]interface{}); ok {
                if updateMap, ok := value.(map[string]interface{}); ok {
                    componentMap[key] = r.deepMerge(existingMap, updateMap)
                    continue
                }
            }
        }
        componentMap[key] = value
    }
    
    // Convert back to Component
    mergedJSON, _ := json.Marshal(componentMap)
    var result models.Component
    json.Unmarshal(mergedJSON, &result)
    
    return result
}

func (r *ComponentRenderer) deepMerge(base, update map[string]interface{}) map[string]interface{} {
    result := make(map[string]interface{})
    
    // Copy base
    for k, v := range base {
        result[k] = v
    }
    
    // Apply updates
    for k, v := range update {
        if existingValue, exists := result