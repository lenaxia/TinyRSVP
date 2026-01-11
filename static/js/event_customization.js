/**
 * Event Customization Manager
 * Handles per-event template customization with real-time preview
 */

class EventCustomization {
    constructor(eventID, apiBaseURL = '/api') {
        this.eventID = eventID;
        this.apiBaseURL = apiBaseURL;
        this.selectedComponentID = null;
        this.customizationData = null;
        this.isDirty = false;
        this.previewFrame = null;
        
        this.init();
    }

    async init() {
        try {
            await this.loadCustomization();
            this.setupEventListeners();
            this.renderComponentList();
            this.renderPreview();
        } catch (error) {
            console.error('Failed to initialize customization:', error);
            this.showError('Failed to load customization data');
        }
    }

    async loadCustomization() {
        const response = await fetch(`${this.apiBaseURL}/events/${this.eventID}/template/customization`, {
            headers: {
                'Accept': 'application/json'
            }
        });

        if (!response.ok) {
            throw new Error(`Failed to load customization: ${response.statusText}`);
        }

        this.customizationData = await response.json();
    }

    setupEventListeners() {
        document.addEventListener('click', (e) => {
            if (e.target.closest('.component-item')) {
                const item = e.target.closest('.component-item');
                this.selectComponent(item.dataset.componentId);
            }
        });

        document.getElementById('btn-save')?.addEventListener('click', () => this.saveCustomization());
        document.getElementById('btn-reset')?.addEventListener('click', () => this.resetCustomization());
        document.getElementById('btn-cancel')?.addEventListener('click', () => this.cancel());

        window.addEventListener('beforeunload', (e) => {
            if (this.isDirty) {
                e.preventDefault();
                e.returnValue = '';
            }
        });
    }

    renderComponentList() {
        const container = document.getElementById('component-list');
        if (!container) return;

        container.innerHTML = '';

        const components = this.customizationData.mergedConfig?.components || [];
        
        components.forEach(component => {
            const item = this.createComponentListItem(component);
            container.appendChild(item);
        });
    }

    createComponentListItem(component) {
        const item = document.createElement('div');
        item.className = 'component-item';
        item.dataset.componentId = component.id;

        const isCustomized = this.isComponentCustomized(component.id);
        if (isCustomized) {
            item.classList.add('customized');
        }

        if (this.selectedComponentID === component.id) {
            item.classList.add('selected');
        }

        item.innerHTML = `
            <div class="component-item-header">
                <span class="component-item-title">${this.formatComponentName(component.id)}</span>
                <span class="component-item-type">${component.type}</span>
            </div>
            <div class="component-item-id">${component.id}</div>
            ${isCustomized ? '<span class="override-indicator">✓ Customized</span>' : ''}
        `;

        return item;
    }

    formatComponentName(id) {
        return id.split('-').map(word => 
            word.charAt(0).toUpperCase() + word.slice(1)
        ).join(' ');
    }

    isComponentCustomized(componentID) {
        if (!this.customizationData.eventOverrides) return false;

        const hasOverride = this.customizationData.eventOverrides.overrides?.some(
            o => o.id === componentID
        );
        const isAddition = this.customizationData.eventOverrides.additions?.some(
            a => a.id === componentID
        );
        const isRemoval = this.customizationData.eventOverrides.removals?.includes(componentID);

        return hasOverride || isAddition || isRemoval;
    }

    selectComponent(componentID) {
        this.selectedComponentID = componentID;
        
        document.querySelectorAll('.component-item').forEach(item => {
            item.classList.toggle('selected', item.dataset.componentId === componentID);
        });

        this.highlightComponentInPreview(componentID);
        this.renderPropertiesPanel(componentID);
    }

    highlightComponentInPreview(componentID) {
        if (!this.previewFrame) return;

        const previewDoc = this.previewFrame.contentDocument;
        if (!previewDoc) return;

        previewDoc.querySelectorAll('.component-highlight').forEach(el => {
            el.classList.remove('component-highlight');
        });

        const element = previewDoc.querySelector(`[data-component-id="${componentID}"]`);
        if (element) {
            element.classList.add('component-highlight');
            element.scrollIntoView({ behavior: 'smooth', block: 'center' });
        }
    }

    renderPropertiesPanel(componentID) {
        const panel = document.getElementById('properties-panel-content');
        if (!panel) return;

        const component = this.getComponent(componentID);
        if (!component) {
            panel.innerHTML = '<div class="properties-empty">Select a component to edit</div>';
            return;
        }

        panel.innerHTML = this.generatePropertiesHTML(component);
        this.attachPropertyEventListeners(componentID);
    }

    getComponent(componentID) {
        return this.customizationData.mergedConfig?.components?.find(
            c => c.id === componentID
        );
    }

    generatePropertiesHTML(component) {
        let html = `
            <div class="property-section">
                <h3>Basic Properties</h3>
                <div class="property-group">
                    <div class="property-field">
                        <label>Component ID</label>
                        <input type="text" value="${component.id}" disabled>
                    </div>
                    <div class="property-field">
                        <label>Type</label>
                        <input type="text" value="${component.type}" disabled>
                    </div>
                    <div class="visibility-toggle">
                        <input type="checkbox" id="visible-toggle" ${component.visible ? 'checked' : ''}>
                        <label for="visible-toggle">Visible</label>
                    </div>
                </div>
            </div>
        `;

        if (component.type === 'TextBox' && component.content) {
            html += this.generateTextPropertiesHTML(component.content);
        } else if (component.type === 'Image' && component.content) {
            html += this.generateImagePropertiesHTML(component.content);
        } else if (component.type === 'Background' && component.content) {
            html += this.generateBackgroundPropertiesHTML(component.content);
        }

        html += `
            <div class="property-section">
                <h3>Position & Size</h3>
                <div class="property-group">
                    <div class="property-row">
                        <div class="property-field">
                            <label>X Position</label>
                            <input type="text" id="pos-x" value="${component.position?.x || ''}">
                        </div>
                        <div class="property-field">
                            <label>Y Position</label>
                            <input type="text" id="pos-y" value="${component.position?.y || ''}">
                        </div>
                    </div>
                    <div class="property-row">
                        <div class="property-field">
                            <label>Width</label>
                            <input type="text" id="dim-width" value="${component.dimensions?.width || ''}">
                        </div>
                        <div class="property-field">
                            <label>Height</label>
                            <input type="text" id="dim-height" value="${component.dimensions?.height || ''}">
                        </div>
                    </div>
                    <div class="property-field">
                        <label>Z-Index</label>
                        <div class="zindex-control">
                            <button type="button" id="zindex-down">−</button>
                            <input type="number" id="zindex" value="${component.zIndex || 0}">
                            <button type="button" id="zindex-up">+</button>
                        </div>
                    </div>
                </div>
            </div>
        `;

        html += `
            <div class="component-actions">
                <button type="button" class="btn-component-action" id="btn-reset-component">
                    Reset to Default
                </button>
                <button type="button" class="btn-component-action danger" id="btn-remove-component">
                    Remove
                </button>
            </div>
        `;

        return html;
    }

    generateTextPropertiesHTML(content) {
        return `
            <div class="property-section">
                <h3>Text Content</h3>
                <div class="property-group">
                    <div class="property-field">
                        <label>Text</label>
                        <textarea id="text-content">${content.text || ''}</textarea>
                    </div>
                    <div class="property-field">
                        <label>Font Family</label>
                        <input type="text" id="font-family" value="${content.fontFamily || ''}">
                    </div>
                    <div class="property-row">
                        <div class="property-field">
                            <label>Font Size</label>
                            <input type="text" id="font-size" value="${content.fontSize || ''}">
                        </div>
                        <div class="property-field">
                            <label>Font Weight</label>
                            <select id="font-weight">
                                <option value="300" ${content.fontWeight === '300' ? 'selected' : ''}>Light</option>
                                <option value="400" ${content.fontWeight === '400' ? 'selected' : ''}>Normal</option>
                                <option value="600" ${content.fontWeight === '600' ? 'selected' : ''}>Semi-Bold</option>
                                <option value="700" ${content.fontWeight === '700' ? 'selected' : ''}>Bold</option>
                            </select>
                        </div>
                    </div>
                    <div class="property-field">
                        <label>Color</label>
                        <div class="color-picker-wrapper">
                            <div class="color-preview" style="background-color: ${content.color || '#000000'}"></div>
                            <input type="color" id="text-color" class="color-picker-input" value="${content.color || '#000000'}">
                        </div>
                    </div>
                    <div class="property-field">
                        <label>Text Align</label>
                        <select id="text-align">
                            <option value="left" ${content.textAlign === 'left' ? 'selected' : ''}>Left</option>
                            <option value="center" ${content.textAlign === 'center' ? 'selected' : ''}>Center</option>
                            <option value="right" ${content.textAlign === 'right' ? 'selected' : ''}>Right</option>
                        </select>
                    </div>
                </div>
            </div>
        `;
    }

    generateImagePropertiesHTML(content) {
        return `
            <div class="property-section">
                <h3>Image Properties</h3>
                <div class="property-group">
                    <div class="property-field">
                        <label>Image URL</label>
                        <input type="text" id="image-src" value="${content.src || ''}">
                    </div>
                    <div class="property-field">
                        <label>Upload New Image</label>
                        <div class="image-upload-area" id="image-upload">
                            <p>Click or drag image here</p>
                            <input type="file" id="image-file" accept="image/*" style="display: none;">
                        </div>
                        ${content.src ? `<img src="${content.src}" class="image-preview" alt="Current image">` : ''}
                    </div>
                    <div class="property-field">
                        <label>Object Fit</label>
                        <select id="object-fit">
                            <option value="cover" ${content.objectFit === 'cover' ? 'selected' : ''}>Cover</option>
                            <option value="contain" ${content.objectFit === 'contain' ? 'selected' : ''}>Contain</option>
                            <option value="fill" ${content.objectFit === 'fill' ? 'selected' : ''}>Fill</option>
                            <option value="none" ${content.objectFit === 'none' ? 'selected' : ''}>None</option>
                        </select>
                    </div>
                    <div class="property-field">
                        <label>Opacity</label>
                        <input type="range" id="image-opacity" min="0" max="1" step="0.1" value="${content.opacity || 1}">
                        <span id="opacity-value">${content.opacity || 1}</span>
                    </div>
                </div>
            </div>
        `;
    }

    generateBackgroundPropertiesHTML(content) {
        return `
            <div class="property-section">
                <h3>Background Properties</h3>
                <div class="property-group">
                    <div class="property-field">
                        <label>Type</label>
                        <select id="bg-type">
                            <option value="color" ${content.type === 'color' ? 'selected' : ''}>Solid Color</option>
                            <option value="gradient" ${content.type === 'gradient' ? 'selected' : ''}>Gradient</option>
                            <option value="image" ${content.type === 'image' ? 'selected' : ''}>Image</option>
                        </select>
                    </div>
                    <div class="property-field" id="bg-color-field">
                        <label>Color</label>
                        <div class="color-picker-wrapper">
                            <div class="color-preview" style="background-color: ${content.color || '#ffffff'}"></div>
                            <input type="color" id="bg-color" class="color-picker-input" value="${content.color || '#ffffff'}">
                        </div>
                    </div>
                </div>
            </div>
        `;
    }

    attachPropertyEventListeners(componentID) {
        const inputs = document.querySelectorAll('#properties-panel-content input, #properties-panel-content select, #properties-panel-content textarea');
        
        inputs.forEach(input => {
            input.addEventListener('change', () => {
                this.handlePropertyChange(componentID, input);
            });

            if (input.type !== 'file') {
                input.addEventListener('input', () => {
                    this.handlePropertyChange(componentID, input);
                });
            }
        });

        document.getElementById('zindex-up')?.addEventListener('click', () => {
            const input = document.getElementById('zindex');
            if (input) {
                input.value = parseInt(input.value || 0) + 1;
                this.handlePropertyChange(componentID, input);
            }
        });

        document.getElementById('zindex-down')?.addEventListener('click', () => {
            const input = document.getElementById('zindex');
            if (input) {
                input.value = Math.max(0, parseInt(input.value || 0) - 1);
                this.handlePropertyChange(componentID, input);
            }
        });

        const colorPreviews = document.querySelectorAll('.color-preview');
        colorPreviews.forEach(preview => {
            preview.addEventListener('click', () => {
                const input = preview.nextElementSibling;
                if (input) input.click();
            });
        });

        const colorInputs = document.querySelectorAll('.color-picker-input');
        colorInputs.forEach(input => {
            input.addEventListener('change', (e) => {
                const preview = e.target.previousElementSibling;
                if (preview) {
                    preview.style.backgroundColor = e.target.value;
                }
            });
        });

        document.getElementById('image-upload')?.addEventListener('click', () => {
            document.getElementById('image-file')?.click();
        });

        document.getElementById('image-file')?.addEventListener('change', (e) => {
            this.handleImageUpload(componentID, e.target.files[0]);
        });

        document.getElementById('btn-reset-component')?.addEventListener('click', () => {
            this.resetComponent(componentID);
        });

        document.getElementById('btn-remove-component')?.addEventListener('click', () => {
            this.removeComponent(componentID);
        });
    }

    handlePropertyChange(componentID, input) {
        this.isDirty = true;
        this.updateComponentProperty(componentID, input.id, input.value);
        this.debouncedPreview();
    }

    updateComponentProperty(componentID, propertyID, value) {
        if (!this.customizationData.eventOverrides) {
            this.customizationData.eventOverrides = {
                version: '1.0',
                overrides: [],
                additions: [],
                removals: []
            };
        }

        let override = this.customizationData.eventOverrides.overrides.find(o => o.id === componentID);
        if (!override) {
            override = { id: componentID, updates: {} };
            this.customizationData.eventOverrides.overrides.push(override);
        }

        const propertyPath = this.getPropertyPath(propertyID);
        this.setNestedProperty(override.updates, propertyPath, value);
    }

    getPropertyPath(propertyID) {
        const pathMap = {
            'text-content': ['content', 'text'],
            'font-family': ['content', 'fontFamily'],
            'font-size': ['content', 'fontSize'],
            'font-weight': ['content', 'fontWeight'],
            'text-color': ['content', 'color'],
            'text-align': ['content', 'textAlign'],
            'image-src': ['content', 'src'],
            'object-fit': ['content', 'objectFit'],
            'image-opacity': ['content', 'opacity'],
            'bg-type': ['content', 'type'],
            'bg-color': ['content', 'color'],
            'pos-x': ['position', 'x'],
            'pos-y': ['position', 'y'],
            'dim-width': ['dimensions', 'width'],
            'dim-height': ['dimensions', 'height'],
            'zindex': ['zIndex'],
            'visible-toggle': ['visible']
        };

        return pathMap[propertyID] || [propertyID];
    }

    setNestedProperty(obj, path, value) {
        for (let i = 0; i < path.length - 1; i++) {
            if (!obj[path[i]]) {
                obj[path[i]] = {};
            }
            obj = obj[path[i]];
        }
        
        const lastKey = path[path.length - 1];
        if (lastKey === 'visible') {
            obj[lastKey] = value === 'on' || value === true;
        } else if (lastKey === 'zIndex') {
            obj[lastKey] = parseInt(value);
        } else if (lastKey === 'opacity') {
            obj[lastKey] = parseFloat(value);
        } else {
            obj[lastKey] = value;
        }
    }

    debouncedPreview = this.debounce(() => {
        this.renderPreview();
    }, 500);

    debounce(func, wait) {
        let timeout;
        return function executedFunction(...args) {
            const later = () => {
                clearTimeout(timeout);
                func(...args);
            };
            clearTimeout(timeout);
            timeout = setTimeout(later, wait);
        };
    }

    async renderPreview() {
        try {
            const response = await fetch(`${this.apiBaseURL}/events/${this.eventID}/template/customization/preview`, {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json',
                    'Accept': 'application/json'
                },
                body: JSON.stringify(this.customizationData.eventOverrides || {
                    version: '1.0',
                    overrides: [],
                    additions: [],
                    removals: []
                })
            });

            if (!response.ok) {
                throw new Error('Preview failed');
            }

            const config = await response.json();
            this.updatePreviewFrame(config);
        } catch (error) {
            console.error('Preview error:', error);
        }
    }

    updatePreviewFrame(config) {
        const frame = document.getElementById('preview-frame');
        if (!frame) return;

        this.previewFrame = frame;
    }

    async handleImageUpload(componentID, file) {
        if (!file) return;

        const formData = new FormData();
        formData.append('image', file);
        formData.append('eventID', this.eventID);

        try {
            const response = await fetch('/api/upload/image', {
                method: 'POST',
                body: formData
            });

            if (!response.ok) {
                throw new Error('Upload failed');
            }

            const data = await response.json();
            this.updateComponentProperty(componentID, 'image-src', data.url);
            this.renderPropertiesPanel(componentID);
            this.renderPreview();
            this.showSuccess('Image uploaded successfully');
        } catch (error) {
            console.error('Upload error:', error);
            this.showError('Failed to upload image');
        }
    }

    async saveCustomization() {
        try {
            const response = await fetch(`${this.apiBaseURL}/events/${this.eventID}/template/customization`, {
                method: 'PUT',
                headers: {
                    'Content-Type': 'application/json',
                    'Accept': 'application/json'
                },
                body: JSON.stringify(this.customizationData.eventOverrides || {
                    version: '1.0',
                    overrides: [],
                    additions: [],
                    removals: []
                })
            });

            if (!response.ok) {
                const error = await response.json();
                throw new Error(error.message || 'Save failed');
            }

            this.isDirty = false;
            this.showSuccess('Customization saved successfully');
        } catch (error) {
            console.error('Save error:', error);
            this.showError(error.message || 'Failed to save customization');
        }
    }

    async resetCustomization() {
        if (!confirm('Are you sure you want to reset all customizations? This cannot be undone.')) {
            return;
        }

        try {
            const response = await fetch(`${this.apiBaseURL}/events/${this.eventID}/template/customization`, {
                method: 'DELETE',
                headers: {
                    'Accept': 'application/json'
                }
            });

            if (!response.ok) {
                throw new Error('Reset failed');
            }

            this.customizationData.eventOverrides = null;
            this.isDirty = false;
            this.renderComponentList();
            this.renderPreview();
            this.showSuccess('Customization reset successfully');
        } catch (error) {
            console.error('Reset error:', error);
            this.showError('Failed to reset customization');
        }
    }

    resetComponent(componentID) {
        if (!this.customizationData.eventOverrides) return;

        this.customizationData.eventOverrides.overrides = 
            this.customizationData.eventOverrides.overrides.filter(o => o.id !== componentID);

        this.isDirty = true;
        this.renderComponentList();
        this.renderPropertiesPanel(componentID);
        this.renderPreview();
    }

    removeComponent(componentID) {
        if (!confirm('Are you sure you want to remove this component?')) {
            return;
        }

        if (!this.customizationData.eventOverrides) {
            this.customizationData.eventOverrides = {
                version: '1.0',
                overrides: [],
                additions: [],
                removals: []
            };
        }

        if (!this.customizationData.eventOverrides.removals.includes(componentID)) {
            this.customizationData.eventOverrides.removals.push(componentID);
        }

        this.isDirty = true;
        this.selectedComponentID = null;
        this.renderComponentList();
        this.renderPropertiesPanel(null);
        this.renderPreview();
    }

    cancel() {
        if (this.isDirty && !confirm('You have unsaved changes. Are you sure you want to leave?')) {
            return;
        }

        window.location.href = `/events/${this.eventID}`;
    }

    showSuccess(message) {
        this.showMessage(message, 'success');
    }

    showError(message) {
        this.showMessage(message, 'error');
    }

    showMessage(message, type) {
        const container = document.getElementById('message-container');
        if (!container) return;

        const banner = document.createElement('div');
        banner.className = `message-banner ${type}`;
        banner.innerHTML = `
            <svg viewBox="0 0 20 20" fill="currentColor">
                ${type === 'success' 
                    ? '<path fill-rule="evenodd" d="M10 18a8 8 0 100-16 8 8 0 000 16zm3.707-9.293a1 1 0 00-1.414-1.414L9 10.586 7.707 9.293a1 1 0 00-1.414 1.414l2 2a1 1 0 001.414 0l4-4z" clip-rule="evenodd"/>'
                    : '<path fill-rule="evenodd" d="M10 18a8 8 0 100-16 8 8 0 000 16zM8.707 7.293a1 1 0 00-1.414 1.414L8.586 10l-1.293 1.293a1 1 0 101.414 1.414L10 11.414l1.293 1.293a1 1 0 001.414-1.414L11.414 10l1.293-1.293a1 1 0 00-1.414-1.414L10 8.586 8.707 7.293z" clip-rule="evenodd"/>'
                }
            </svg>
            <span>${message}</span>
        `;

        container.appendChild(banner);

        setTimeout(() => {
            banner.style.opacity = '0';
            setTimeout(() => banner.remove(), 300);
        }, 5000);
    }
}

if (typeof module !== 'undefined' && module.exports) {
    module.exports = EventCustomization;
}
