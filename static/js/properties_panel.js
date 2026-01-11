class PropertiesPanel {
    constructor(panelElement) {
        this.panel = panelElement;
        this.currentComponent = null;
        this.init();
    }

    init() {
        if (!this.panel) return;
        
        this.panel.classList.add('properties-panel');
        this.panel.setAttribute('role', 'region');
        this.panel.setAttribute('aria-label', 'Component properties');
        
        this.showEmptyState();
    }

    showEmptyState() {
        this.panel.innerHTML = `
            <div class="properties-empty">
                <div class="properties-empty-icon">🎯</div>
                <h3 class="properties-empty-title">No Component Selected</h3>
                <p class="properties-empty-description">
                    Select a component on the canvas to edit its properties
                </p>
            </div>
        `;
    }

    showComponent(component) {
        this.currentComponent = component;
        this.panel.innerHTML = '';
        
        const header = document.createElement('div');
        header.className = 'properties-header';
        header.innerHTML = `
            <h2 class="properties-title">${component.type} Properties</h2>
            <button 
                type="button" 
                class="btn-icon btn-delete-component" 
                aria-label="Delete component"
                title="Delete component"
            >
                🗑️
            </button>
        `;
        this.panel.appendChild(header);
        
        const form = document.createElement('form');
        form.className = 'properties-form';
        form.setAttribute('aria-label', 'Component properties form');
        
        this.addBasicProperties(form, component);
        this.addPositionProperties(form, component);
        this.addDimensionProperties(form, component);
        this.addStyleProperties(form, component);
        this.addContentProperties(form, component);
        
        this.panel.appendChild(form);
        this.attachEventListeners();
    }

    addBasicProperties(form, component) {
        const section = this.createSection('Basic', [
            {
                type: 'text',
                id: 'prop-id',
                label: 'Component ID',
                value: component.id,
                readonly: true
            },
            {
                type: 'number',
                id: 'prop-zindex',
                label: 'Z-Index',
                value: component.zIndex || 1,
                min: 0,
                max: 100
            },
            {
                type: 'checkbox',
                id: 'prop-visible',
                label: 'Visible',
                checked: component.visible !== false
            }
        ]);
        form.appendChild(section);
    }

    addPositionProperties(form, component) {
        const fields = [
            {
                type: 'select',
                id: 'prop-position-mode',
                label: 'Position Mode',
                value: component.position?.mode || 'absolute',
                options: [
                    { value: 'absolute', label: 'Absolute' },
                    { value: 'relative', label: 'Relative' },
                    { value: 'flex', label: 'Flex' }
                ]
            }
        ];
        
        if (component.position?.mode !== 'flex') {
            fields.push(
                {
                    type: 'text',
                    id: 'prop-position-x',
                    label: 'X Position',
                    value: component.position?.x || '0',
                    placeholder: 'e.g., 50%, 100px'
                },
                {
                    type: 'text',
                    id: 'prop-position-y',
                    label: 'Y Position',
                    value: component.position?.y || '0',
                    placeholder: 'e.g., 50%, 100px'
                }
            );
        } else {
            fields.push({
                type: 'number',
                id: 'prop-position-order',
                label: 'Flex Order',
                value: component.position?.order || 0
            });
        }
        
        const section = this.createSection('Position', fields);
        form.appendChild(section);
    }

    addDimensionProperties(form, component) {
        const section = this.createSection('Dimensions', [
            {
                type: 'text',
                id: 'prop-width',
                label: 'Width',
                value: component.dimensions?.width || 'auto',
                placeholder: 'e.g., 100%, 200px, auto'
            },
            {
                type: 'text',
                id: 'prop-height',
                label: 'Height',
                value: component.dimensions?.height || 'auto',
                placeholder: 'e.g., 100%, 200px, auto'
            }
        ]);
        form.appendChild(section);
    }

    addStyleProperties(form, component) {
        const fields = [];
        
        if (component.style) {
            if (component.style.backgroundColor !== undefined) {
                fields.push({
                    type: 'color',
                    id: 'prop-style-bgcolor',
                    label: 'Background Color',
                    value: component.style.backgroundColor || '#ffffff'
                });
            }
            
            if (component.style.borderRadius !== undefined) {
                fields.push({
                    type: 'text',
                    id: 'prop-style-border-radius',
                    label: 'Border Radius',
                    value: component.style.borderRadius || '0'
                });
            }
            
            if (component.style.border !== undefined) {
                fields.push({
                    type: 'text',
                    id: 'prop-style-border',
                    label: 'Border',
                    value: component.style.border || 'none'
                });
            }
            
            if (component.style.boxShadow !== undefined) {
                fields.push({
                    type: 'text',
                    id: 'prop-style-shadow',
                    label: 'Box Shadow',
                    value: component.style.boxShadow || 'none'
                });
            }
        }
        
        if (fields.length > 0) {
            const section = this.createSection('Style', fields);
            form.appendChild(section);
        }
    }

    addContentProperties(form, component) {
        const fields = [];
        
        switch (component.type) {
            case 'TextBox':
                fields.push(
                    {
                        type: 'textarea',
                        id: 'prop-content-text',
                        label: 'Text',
                        value: component.content?.text || '',
                        rows: 3
                    },
                    {
                        type: 'select',
                        id: 'prop-content-align',
                        label: 'Text Align',
                        value: component.content?.textAlign || 'left',
                        options: [
                            { value: 'left', label: 'Left' },
                            { value: 'center', label: 'Center' },
                            { value: 'right', label: 'Right' },
                            { value: 'justify', label: 'Justify' }
                        ]
                    },
                    {
                        type: 'text',
                        id: 'prop-content-font-family',
                        label: 'Font Family',
                        value: component.content?.fontFamily || 'inherit'
                    },
                    {
                        type: 'text',
                        id: 'prop-content-font-size',
                        label: 'Font Size',
                        value: component.content?.fontSize || '16px'
                    },
                    {
                        type: 'select',
                        id: 'prop-content-font-weight',
                        label: 'Font Weight',
                        value: component.content?.fontWeight || '400',
                        options: [
                            { value: '300', label: 'Light' },
                            { value: '400', label: 'Normal' },
                            { value: '600', label: 'Semi-Bold' },
                            { value: '700', label: 'Bold' }
                        ]
                    },
                    {
                        type: 'color',
                        id: 'prop-content-color',
                        label: 'Text Color',
                        value: component.content?.color || '#000000'
                    }
                );
                break;
                
            case 'Image':
                fields.push(
                    {
                        type: 'text',
                        id: 'prop-content-src',
                        label: 'Image URL',
                        value: component.content?.src || ''
                    },
                    {
                        type: 'text',
                        id: 'prop-content-alt',
                        label: 'Alt Text',
                        value: component.content?.alt || ''
                    },
                    {
                        type: 'select',
                        id: 'prop-content-object-fit',
                        label: 'Object Fit',
                        value: component.content?.objectFit || 'cover',
                        options: [
                            { value: 'cover', label: 'Cover' },
                            { value: 'contain', label: 'Contain' },
                            { value: 'fill', label: 'Fill' },
                            { value: 'none', label: 'None' }
                        ]
                    },
                    {
                        type: 'range',
                        id: 'prop-content-opacity',
                        label: 'Opacity',
                        value: component.content?.opacity || 1,
                        min: 0,
                        max: 1,
                        step: 0.1
                    }
                );
                break;
                
            case 'Background':
                fields.push(
                    {
                        type: 'select',
                        id: 'prop-content-bg-type',
                        label: 'Background Type',
                        value: component.content?.type || 'color',
                        options: [
                            { value: 'color', label: 'Solid Color' },
                            { value: 'gradient', label: 'Gradient' },
                            { value: 'image', label: 'Image' }
                        ]
                    },
                    {
                        type: 'color',
                        id: 'prop-content-bg-color',
                        label: 'Color',
                        value: component.content?.color || '#f8f9fa'
                    },
                    {
                        type: 'text',
                        id: 'prop-content-bg-gradient',
                        label: 'Gradient',
                        value: component.content?.gradient || ''
                    }
                );
                break;
                
            case 'Container':
                fields.push(
                    {
                        type: 'select',
                        id: 'prop-layout-direction',
                        label: 'Flex Direction',
                        value: component.layout?.flexDirection || 'column',
                        options: [
                            { value: 'row', label: 'Row' },
                            { value: 'column', label: 'Column' }
                        ]
                    },
                    {
                        type: 'select',
                        id: 'prop-layout-align',
                        label: 'Align Items',
                        value: component.layout?.alignItems || 'flex-start',
                        options: [
                            { value: 'flex-start', label: 'Start' },
                            { value: 'center', label: 'Center' },
                            { value: 'flex-end', label: 'End' },
                            { value: 'stretch', label: 'Stretch' }
                        ]
                    },
                    {
                        type: 'text',
                        id: 'prop-layout-gap',
                        label: 'Gap',
                        value: component.layout?.gap || '0'
                    },
                    {
                        type: 'text',
                        id: 'prop-layout-padding',
                        label: 'Padding',
                        value: component.layout?.padding || '0'
                    }
                );
                break;
        }
        
        if (fields.length > 0) {
            const section = this.createSection('Content', fields);
            form.appendChild(section);
        }
    }

    createSection(title, fields) {
        const section = document.createElement('div');
        section.className = 'properties-section';
        
        const header = document.createElement('h3');
        header.className = 'properties-section-title';
        header.textContent = title;
        section.appendChild(header);
        
        fields.forEach(field => {
            const fieldEl = this.createField(field);
            section.appendChild(fieldEl);
        });
        
        return section;
    }

    createField(field) {
        const fieldWrapper = document.createElement('div');
        fieldWrapper.className = 'properties-field';
        
        const label = document.createElement('label');
        label.className = 'properties-label';
        label.htmlFor = field.id;
        label.textContent = field.label;
        fieldWrapper.appendChild(label);
        
        let input;
        
        switch (field.type) {
            case 'select':
                input = document.createElement('select');
                input.className = 'properties-input properties-select';
                field.options.forEach(opt => {
                    const option = document.createElement('option');
                    option.value = opt.value;
                    option.textContent = opt.label;
                    option.selected = opt.value === field.value;
                    input.appendChild(option);
                });
                break;
                
            case 'textarea':
                input = document.createElement('textarea');
                input.className = 'properties-input properties-textarea';
                input.rows = field.rows || 3;
                input.value = field.value || '';
                break;
                
            case 'checkbox':
                input = document.createElement('input');
                input.type = 'checkbox';
                input.className = 'properties-checkbox';
                input.checked = field.checked || false;
                break;
                
            case 'range':
                const rangeWrapper = document.createElement('div');
                rangeWrapper.className = 'properties-range-wrapper';
                
                input = document.createElement('input');
                input.type = 'range';
                input.className = 'properties-range';
                input.min = field.min;
                input.max = field.max;
                input.step = field.step || 1;
                input.value = field.value || field.min;
                
                const valueDisplay = document.createElement('span');
                valueDisplay.className = 'properties-range-value';
                valueDisplay.textContent = input.value;
                
                input.addEventListener('input', () => {
                    valueDisplay.textContent = input.value;
                });
                
                rangeWrapper.appendChild(input);
                rangeWrapper.appendChild(valueDisplay);
                fieldWrapper.appendChild(rangeWrapper);
                break;
                
            default:
                input = document.createElement('input');
                input.type = field.type;
                input.className = 'properties-input';
                input.value = field.value || '';
                
                if (field.placeholder) input.placeholder = field.placeholder;
                if (field.min !== undefined) input.min = field.min;
                if (field.max !== undefined) input.max = field.max;
                if (field.readonly) input.readOnly = true;
        }
        
        if (input && field.type !== 'range') {
            input.id = field.id;
            input.name = field.id;
            input.dataset.property = field.id.replace('prop-', '');
            fieldWrapper.appendChild(input);
        }
        
        return fieldWrapper;
    }

    attachEventListeners() {
        const form = this.panel.querySelector('.properties-form');
        if (!form) return;
        
        form.addEventListener('input', (e) => {
            if (e.target.matches('.properties-input, .properties-select, .properties-textarea, .properties-checkbox, .properties-range')) {
                this.handlePropertyChange(e.target);
            }
        });
        
        form.addEventListener('change', (e) => {
            if (e.target.matches('.properties-input, .properties-select, .properties-textarea, .properties-checkbox')) {
                this.handlePropertyChange(e.target);
            }
        });
        
        const deleteBtn = this.panel.querySelector('.btn-delete-component');
        if (deleteBtn) {
            deleteBtn.addEventListener('click', () => {
                if (this.currentComponent) {
                    this.dispatchEvent('component-delete-requested', { 
                        componentId: this.currentComponent.id 
                    });
                }
            });
        }
    }

    handlePropertyChange(input) {
        if (!this.currentComponent) return;
        
        const property = input.dataset.property;
        const value = input.type === 'checkbox' ? input.checked : input.value;
        
        const updates = this.parsePropertyPath(property, value);
        
        this.dispatchEvent('property-changed', {
            componentId: this.currentComponent.id,
            property,
            value,
            updates
        });
    }

    parsePropertyPath(property, value) {
        const parts = property.split('-');
        const updates = {};
        
        if (parts[0] === 'zindex') {
            updates.zIndex = parseInt(value);
        } else if (parts[0] === 'visible') {
            updates.visible = value;
        } else if (parts[0] === 'position') {
            if (!updates.position) updates.position = {};
            if (parts[1] === 'mode') updates.position.mode = value;
            else if (parts[1] === 'x') updates.position.x = value;
            else if (parts[1] === 'y') updates.position.y = value;
            else if (parts[1] === 'order') updates.position.order = parseInt(value);
        } else if (parts[0] === 'width' || parts[0] === 'height') {
            if (!updates.dimensions) updates.dimensions = {};
            updates.dimensions[parts[0]] = value;
        } else if (parts[0] === 'style') {
            if (!updates.style) updates.style = {};
            const styleProp = parts.slice(1).join('-');
            if (styleProp === 'bgcolor') updates.style.backgroundColor = value;
            else if (styleProp === 'border-radius') updates.style.borderRadius = value;
            else if (styleProp === 'border') updates.style.border = value;
            else if (styleProp === 'shadow') updates.style.boxShadow = value;
        } else if (parts[0] === 'content') {
            if (!updates.content) updates.content = {};
            const contentProp = parts.slice(1).join('-');
            if (contentProp === 'text') updates.content.text = value;
            else if (contentProp === 'align') updates.content.textAlign = value;
            else if (contentProp === 'font-family') updates.content.fontFamily = value;
            else if (contentProp === 'font-size') updates.content.fontSize = value;
            else if (contentProp === 'font-weight') updates.content.fontWeight = value;
            else if (contentProp === 'color') updates.content.color = value;
            else if (contentProp === 'src') updates.content.src = value;
            else if (contentProp === 'alt') updates.content.alt = value;
            else if (contentProp === 'object-fit') updates.content.objectFit = value;
            else if (contentProp === 'opacity') updates.content.opacity = parseFloat(value);
            else if (contentProp === 'bg-type') updates.content.type = value;
            else if (contentProp === 'bg-color') updates.content.color = value;
            else if (contentProp === 'bg-gradient') updates.content.gradient = value;
        } else if (parts[0] === 'layout') {
            if (!updates.layout) updates.layout = {};
            const layoutProp = parts.slice(1).join('-');
            if (layoutProp === 'direction') updates.layout.flexDirection = value;
            else if (layoutProp === 'align') updates.layout.alignItems = value;
            else if (layoutProp === 'gap') updates.layout.gap = value;
            else if (layoutProp === 'padding') updates.layout.padding = value;
        }
        
        return updates;
    }

    dispatchEvent(eventName, detail) {
        const event = new CustomEvent(eventName, { detail });
        this.panel.dispatchEvent(event);
        document.dispatchEvent(event);
    }
}

if (typeof module !== 'undefined' && module.exports) {
    module.exports = PropertiesPanel;
}
