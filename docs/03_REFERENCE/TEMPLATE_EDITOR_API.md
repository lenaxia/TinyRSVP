# Template Editor API Reference

## Quick Start

The Template Editor API provides endpoints for managing component-based template configurations. All endpoints require authentication and admin privileges.

## Base URL

```
/api/templates/:id/components
```

## Endpoints

### 1. Get Component Configuration

**Endpoint:** `GET /api/templates/:id/components`

**Description:** Retrieve the component configuration for a template.

**Authentication:** Required (Admin or template owner)

**Response:**
```json
{
  "template": {
    "id": 1,
    "name": "Wedding Elegance",
    "type": "rsvp_page",
    "category": "card"
  },
  "component_config": {
    "version": "1.0",
    "metadata": {
      "name": "Wedding Elegance",
      "category": "card",
      "description": "Elegant wedding invitation"
    },
    "layout": {
      "mode": "card",
      "cardWidth": "800px",
      "backgroundColor": "#ffffff"
    },
    "components": [
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
          "text": "{{.Event.Title}}",
          "textAlign": "center",
          "fontSize": "48px",
          "color": "#2c3e50"
        }
      }
    ]
  }
}
```

### 2. Update Component Configuration

**Endpoint:** `PUT /api/templates/:id/components`

**Description:** Replace the entire component array for a template.

**Authentication:** Required (Admin only)

**Request Body:**
```json
{
  "components": [
    {
      "id": "title-text",
      "type": "TextBox",
      "position": {
        "mode": "absolute",
        "x": "50%",
        "y": "250px"
      },
      "dimensions": {
        "width": "80%",
        "height": "auto"
      },
      "zIndex": 10,
      "visible": true,
      "content": {
        "text": "{{.Event.Title}}",
        "textAlign": "center",
        "fontSize": "56px",
        "color": "#8b4789"
      }
    },
    {
      "id": "subtitle-text",
      "type": "TextBox",
      "position": {
        "mode": "absolute",
        "x": "50%",
        "y": "320px"
      },
      "dimensions": {
        "width": "70%",
        "height": "auto"
      },
      "zIndex": 9,
      "visible": true,
      "content": {
        "text": "Join us for a celebration",
        "textAlign": "center",
        "fontSize": "24px",
        "color": "#666666"
      }
    }
  ]
}
```

**Response:**
```json
{
  "message": "components updated successfully"
}
```

### 3. Preview Component Changes

**Endpoint:** `POST /api/templates/:id/components/preview`

**Description:** Generate a preview of component changes without saving to the database.

**Authentication:** Required (Admin or template owner)

**Request Body:**
```json
{
  "updates": [
    {
      "component_id": "title-text",
      "property": "zIndex",
      "value": 20
    },
    {
      "component_id": "title-text",
      "property": "content",
      "value": {
        "fontSize": "64px",
        "color": "#ff0000"
      }
    }
  ],
  "additions": [
    {
      "id": "new-component",
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
        "src": "/static/images/header.jpg",
        "objectFit": "cover"
      }
    }
  ],
  "removals": ["old-component-id"]
}
```

**Response:**
```json
{
  "preview": {
    "version": "1.0",
    "metadata": {...},
    "layout": {...},
    "components": [
      {
        "id": "title-text",
        "zIndex": 20,
        "content": {
          "fontSize": "64px",
          "color": "#ff0000"
        }
      },
      {
        "id": "new-component",
        "type": "Image",
        "zIndex": 1
      }
    ]
  }
}
```

### 4. Validate Component Configuration

**Endpoint:** `GET /api/templates/:id/components/validate`

**Description:** Validate the current component configuration for a template.

**Authentication:** Required (Admin or template owner)

**Response (Valid):**
```json
{
  "valid": true,
  "errors": []
}
```

**Response (Invalid):**
```json
{
  "valid": false,
  "errors": [
    "component[0]: ID is required",
    "component[2]: duplicate ID title-text",
    "component[3]: invalid type InvalidType",
    "component[4]: invalid position mode invalid"
  ]
}
```

## Error Responses

All endpoints return standard error responses:

### 400 Bad Request
```json
{
  "error": "invalid template ID",
  "code": "BAD_REQUEST"
}
```

### 401 Unauthorized
```json
{
  "error": "authentication required",
  "code": "UNAUTHORIZED"
}
```

### 403 Forbidden
```json
{
  "error": "only admins can edit templates",
  "code": "PERMISSION_DENIED"
}
```

### 404 Not Found
```json
{
  "error": "Template not found: 123",
  "code": "NOT_FOUND"
}
```

### 422 Validation Error
```json
{
  "error": "duplicate component ID: title-text",
  "code": "VALIDATION_ERROR",
  "field": "components[1].id"
}
```

## Validation Rules

### Component Configuration
- ✅ Version is required
- ✅ Maximum 50 components per template
- ✅ Component IDs must be unique
- ✅ Component IDs cannot be empty
- ✅ Component types must be valid (TextBox, Image, Background, Overlay, Container, Divider)
- ✅ Position modes must be valid (absolute, relative, flex)

### Component Properties
- **id** (string, required): Unique identifier for the component
- **type** (string, required): Component type
- **position** (object, required): Position configuration
  - **mode** (string, required): Position mode
  - **x** (string, optional): X coordinate (e.g., "50%", "100px")
  - **y** (string, optional): Y coordinate (e.g., "200px", "10%")
  - **order** (number, optional): Flex order
- **dimensions** (object, required): Size configuration
  - **width** (string, required): Width (e.g., "100%", "800px")
  - **height** (string, required): Height (e.g., "auto", "400px")
- **zIndex** (number, required): Stacking order
- **visible** (boolean, required): Visibility flag
- **className** (string, optional): CSS class name
- **content** (object, optional): Component-specific content
- **layout** (object, optional): Layout properties (for Container)
- **style** (object, optional): Additional styles
- **children** (array, optional): Child component IDs (for Container)
- **responsive** (object, optional): Responsive overrides

## Usage Examples

### Example 1: Get Template Components

```bash
curl -X GET \
  http://localhost:8080/api/templates/1/components \
  -H 'Authorization: Bearer YOUR_TOKEN' \
  -H 'Accept: application/json'
```

### Example 2: Update Components

```bash
curl -X PUT \
  http://localhost:8080/api/templates/1/components \
  -H 'Authorization: Bearer YOUR_TOKEN' \
  -H 'Content-Type: application/json' \
  -d '{
    "components": [
      {
        "id": "title-text",
        "type": "TextBox",
        "position": {"mode": "absolute", "x": "50%", "y": "200px"},
        "dimensions": {"width": "80%", "height": "auto"},
        "zIndex": 10,
        "visible": true,
        "content": {
          "text": "{{.Event.Title}}",
          "fontSize": "48px"
        }
      }
    ]
  }'
```

### Example 3: Preview Changes

```bash
curl -X POST \
  http://localhost:8080/api/templates/1/components/preview \
  -H 'Authorization: Bearer YOUR_TOKEN' \
  -H 'Content-Type: application/json' \
  -d '{
    "updates": [
      {
        "component_id": "title-text",
        "property": "zIndex",
        "value": 20
      }
    ],
    "additions": [],
    "removals": []
  }'
```

### Example 4: Validate Configuration

```bash
curl -X GET \
  http://localhost:8080/api/templates/1/components/validate \
  -H 'Authorization: Bearer YOUR_TOKEN' \
  -H 'Accept: application/json'
```

## Integration with Editor Service

The handlers use the EditorService which provides additional methods not directly exposed via HTTP:

```go
// In your Go code
editorService := templates.NewEditorService(templateRepo)

// Add a single component
err := editorService.AddComponent(ctx, templateID, component)

// Remove a component
err := editorService.RemoveComponent(ctx, templateID, "component-id")

// Update a single property
err := editorService.UpdateComponentProperty(ctx, templateID, "component-id", "zIndex", 20)

// Reorder components
err := editorService.ReorderComponents(ctx, templateID, []string{"comp-1", "comp-2", "comp-3"})
```

## Next Steps

This backend API is ready for frontend integration in Phase 3, Part 2, which will implement:
- Visual component editor UI
- Drag-and-drop positioning
- Property editing panels
- Real-time preview
- Component library palette
