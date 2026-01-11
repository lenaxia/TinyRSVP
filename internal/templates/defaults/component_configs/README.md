# Component Configurations

This directory contains the default component configurations for all RSVP themes.

## Purpose

These JSON files define the component-based layouts for each theme, which can be:
1. Stored in the database `templates.component_config` field
2. Overridden per-event via `events.component_overrides` field
3. Used by the ComponentRenderer to generate dynamic layouts

## Files

- `plain-text.json` - Simple, clean text-based invitation
- `wedding-elegance.json` - Elegant wedding invitation with floral design
- `birthday-celebration.json` - Fun birthday invitation with balloons
- `corporate-professional.json` - Professional business event invitation
- `holiday-festive.json` - Festive holiday celebration invitation
- `garden-party.json` - Fresh garden party invitation with floral elements
- `modern-minimalist.json` - Clean, modern minimalist invitation design

## Structure

Each configuration follows the ComponentConfiguration schema:

```json
{
  "version": "1.0",
  "metadata": {
    "name": "Theme Name",
    "category": "card|simple",
    "description": "Theme description"
  },
  "layout": {
    "mode": "card",
    "cardWidth": "800px",
    "cardMaxWidth": "90vw",
    "backgroundColor": "#ffffff"
  },
  "components": [
    {
      "id": "unique-component-id",
      "type": "TextBox|Image|Background|Overlay|Container|Divider",
      "position": {
        "mode": "absolute|relative|flex",
        "x": "50%",
        "y": "100px",
        "order": 1
      },
      "dimensions": {
        "width": "100%",
        "height": "auto"
      },
      "zIndex": 10,
      "visible": true,
      "content": {
        "text": "{{.Event.Title}}",
        "textAlign": "center",
        "fontFamily": "Arial, sans-serif",
        "fontSize": "2rem",
        "color": "#000000"
      },
      "responsive": {
        "mobile": {
          "fontSize": "1.5rem"
        }
      }
    }
  ]
}
```

## Usage

These configurations are loaded by the template seeder and stored in the database during initialization.

See `internal/templates/seeder.go` for the seeding logic.
