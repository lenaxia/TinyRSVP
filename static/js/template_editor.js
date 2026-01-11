class TemplateEditor {
    constructor(options = {}) {
        this.templateId = options.templateId;
        this.apiBaseUrl = options.apiBaseUrl || '/api/templates';
        
        this.canvas = null;
        this.palette = null;
        this.propertiesPanel = null;
        
        this.templateData = null;
        this.isDirty = false;
        this.isSaving = false;
        
        this.init();
    }

    async init() {
        this.initializeElements();
        this.initializeModules();
        this.attachEventListeners();
        this.attachGlobalListeners();
        
        if (this.templateId) {
            await this.loadTemplate();
        }
    }

    initializeElements() {
        this.elements = {
            canvas: document.getElementById('template-canvas'),
            palette: document.getElementById('component-palette'),
            properties: document.getElementById('properties-panel'),
            toolbar: document.getElementById('editor-toolbar'),
            saveBtn: document.getElementById('btn-save'),
            previewBtn: document.getElementById('btn-preview'),
            undoBtn: document.getElementById('btn-undo'),
            redoBtn: document.getElementById('btn-redo'),
            zoomIn: document.getElementById('btn-zoom-in'),
            zoomOut: document.getElementById('btn-zoom-out'),
            zoomReset: document.getElementById('btn-zoom-reset'),
            modeDesktop: document.getElementById('btn-mode-desktop'),
            modeTablet: document.getElementById('btn-mode-tablet'),
            modeMobile: document.getElementById('btn-mode-mobile'),
            toggleGrid: document.getElementById('btn-toggle-grid'),
            toggleSnap: document.getElementById('btn-toggle-snap'),
            statusText: document.getElementById('editor-status')
        };
    }

    initializeModules() {
        if (this.elements.canvas && typeof VisualCanvas !== 'undefined') {
            this.canvas = new VisualCanvas(this.elements.canvas, {
                gridSize: 10,
                snapToGrid: true,
                zoom: 1,
                mode: 'desktop'
            });
        }
        
        if (this.elements.palette && typeof ComponentPalette !== 'undefined') {
            this.palette = new ComponentPalette(this.elements.palette);
        }
        
        if (this.elements.properties && typeof PropertiesPanel !== 'undefined') {
            this.propertiesPanel = new PropertiesPanel(this.elements.properties);
        }
    }

    attachEventListeners() {
        if (this.elements.saveBtn) {
            this.elements.saveBtn.addEventListener('click', () => this.save());
        }
        
        if (this.elements.previewBtn) {
            this.elements.previewBtn.addEventListener('click', () => this.preview());
        }
        
        if (this.elements.undoBtn) {
            this.elements.undoBtn.addEventListener('click', () => this.undo());
        }
        
        if (this.elements.redoBtn) {
            this.elements.redoBtn.addEventListener('click', () => this.redo());
        }
        
        if (this.elements.zoomIn) {
            this.elements.zoomIn.addEventListener('click', () => this.zoomIn());
        }
        
        if (this.elements.zoomOut) {
            this.elements.zoomOut.addEventListener('click', () => this.zoomOut());
        }
        
        if (this.elements.zoomReset) {
            this.elements.zoomReset.addEventListener('click', () => this.zoomReset());
        }
        
        if (this.elements.modeDesktop) {
            this.elements.modeDesktop.addEventListener('click', () => this.setMode('desktop'));
        }
        
        if (this.elements.modeTablet) {
            this.elements.modeTablet.addEventListener('click', () => this.setMode('tablet'));
        }
        
        if (this.elements.modeMobile) {
            this.elements.modeMobile.addEventListener('click', () => this.setMode('mobile'));
        }
        
        if (this.elements.toggleGrid) {
            this.elements.toggleGrid.addEventListener('click', () => this.toggleGrid());
        }
        
        if (this.elements.toggleSnap) {
            this.elements.toggleSnap.addEventListener('click', () => this.toggleSnap());
        }
    }

    attachGlobalListeners() {
        document.addEventListener('component-selected', (e) => {
            if (e.detail.component && this.propertiesPanel) {
                this.propertiesPanel.showComponent(e.detail.component);
            }
        });
        
        document.addEventListener('component-deselected', () => {
            if (this.propertiesPanel) {
                this.propertiesPanel.showEmptyState();
            }
        });
        
        document.addEventListener('component-added', () => {
            this.markDirty();
        });
        
        document.addEventListener('component-removed', () => {
            this.markDirty();
        });
        
        document.addEventListener('component-moved', () => {
            this.markDirty();
        });
        
        document.addEventListener('component-resized', () => {
            this.markDirty();
        });
        
        document.addEventListener('component-updated', () => {
            this.markDirty();
        });
        
        document.addEventListener('property-changed', (e) => {
            if (this.canvas && e.detail.componentId && e.detail.updates) {
                this.canvas.updateComponent(e.detail.componentId, e.detail.updates);
            }
        });
        
        document.addEventListener('component-delete-requested', (e) => {
            if (this.canvas && e.detail.componentId) {
                if (confirm('Are you sure you want to delete this component?')) {
                    this.canvas.removeComponent(e.detail.componentId);
                }
            }
        });
        
        window.addEventListener('beforeunload', (e) => {
            if (this.isDirty) {
                e.preventDefault();
                e.returnValue = 'You have unsaved changes. Are you sure you want to leave?';
                return e.returnValue;
            }
        });
        
        document.addEventListener('keydown', (e) => {
            if ((e.ctrlKey || e.metaKey) && e.key === 's') {
                e.preventDefault();
                this.save();
            }
        });
    }

    async loadTemplate() {
        try {
            this.updateStatus('Loading template...');
            
            const response = await fetch(`${this.apiBaseUrl}/${this.templateId}/components`, {
                headers: {
                    'Accept': 'application/json'
                }
            });
            
            if (!response.ok) {
                throw new Error(`Failed to load template: ${response.statusText}`);
            }
            
            const data = await response.json();
            this.templateData = data;
            
            if (data.component_config && data.component_config.components) {
                if (this.canvas) {
                    this.canvas.render(data.component_config.components);
                }
            }
            
            this.updateStatus('Template loaded');
            this.isDirty = false;
            
        } catch (error) {
            console.error('Error loading template:', error);
            this.updateStatus('Error loading template', 'error');
            this.showError('Failed to load template: ' + error.message);
        }
    }

    async save() {
        if (this.isSaving) return;
        
        try {
            this.isSaving = true;
            this.updateStatus('Saving...');
            
            if (this.elements.saveBtn) {
                this.elements.saveBtn.disabled = true;
            }
            
            const components = this.canvas ? this.canvas.getComponents() : [];
            
            const response = await fetch(`${this.apiBaseUrl}/${this.templateId}/components`, {
                method: 'PUT',
                headers: {
                    'Content-Type': 'application/json',
                    'Accept': 'application/json'
                },
                body: JSON.stringify({ components })
            });
            
            if (!response.ok) {
                const errorData = await response.json().catch(() => ({}));
                throw new Error(errorData.error || `Save failed: ${response.statusText}`);
            }
            
            this.isDirty = false;
            this.updateStatus('Saved successfully', 'success');
            
            setTimeout(() => {
                this.updateStatus('Ready');
            }, 2000);
            
        } catch (error) {
            console.error('Error saving template:', error);
            this.updateStatus('Save failed', 'error');
            this.showError('Failed to save: ' + error.message);
        } finally {
            this.isSaving = false;
            if (this.elements.saveBtn) {
                this.elements.saveBtn.disabled = false;
            }
        }
    }

    async preview() {
        try {
            this.updateStatus('Generating preview...');
            
            const components = this.canvas ? this.canvas.getComponents() : [];
            
            const response = await fetch(`${this.apiBaseUrl}/${this.templateId}/components/preview`, {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json',
                    'Accept': 'application/json'
                },
                body: JSON.stringify({
                    updates: [],
                    additions: components,
                    removals: []
                })
            });
            
            if (!response.ok) {
                throw new Error(`Preview failed: ${response.statusText}`);
            }
            
            const data = await response.json();
            
            const previewWindow = window.open('', 'Template Preview', 'width=800,height=600');
            if (previewWindow) {
                previewWindow.document.write('<html><head><title>Template Preview</title>');
                previewWindow.document.write('<link rel="stylesheet" href="/static/css/component_renderer.css">');
                previewWindow.document.write('</head><body>');
                previewWindow.document.write('<div class="preview-container">');
                previewWindow.document.write(JSON.stringify(data.preview, null, 2));
                previewWindow.document.write('</div>');
                previewWindow.document.write('</body></html>');
                previewWindow.document.close();
            }
            
            this.updateStatus('Preview opened');
            
        } catch (error) {
            console.error('Error generating preview:', error);
            this.updateStatus('Preview failed', 'error');
            this.showError('Failed to generate preview: ' + error.message);
        }
    }

    undo() {
        if (this.canvas) {
            this.canvas.undo();
            this.updateStatus('Undo');
        }
    }

    redo() {
        if (this.canvas) {
            this.canvas.redo();
            this.updateStatus('Redo');
        }
    }

    zoomIn() {
        if (this.canvas) {
            const newZoom = Math.min(2, this.canvas.options.zoom + 0.1);
            this.canvas.setZoom(newZoom);
            this.updateStatus(`Zoom: ${Math.round(newZoom * 100)}%`);
        }
    }

    zoomOut() {
        if (this.canvas) {
            const newZoom = Math.max(0.5, this.canvas.options.zoom - 0.1);
            this.canvas.setZoom(newZoom);
            this.updateStatus(`Zoom: ${Math.round(newZoom * 100)}%`);
        }
    }

    zoomReset() {
        if (this.canvas) {
            this.canvas.setZoom(1);
            this.updateStatus('Zoom: 100%');
        }
    }

    setMode(mode) {
        if (this.canvas) {
            this.canvas.setMode(mode);
            this.updateStatus(`Mode: ${mode}`);
            
            [this.elements.modeDesktop, this.elements.modeTablet, this.elements.modeMobile].forEach(btn => {
                if (btn) btn.classList.remove('active');
            });
            
            const activeBtn = this.elements[`mode${mode.charAt(0).toUpperCase() + mode.slice(1)}`];
            if (activeBtn) activeBtn.classList.add('active');
        }
    }

    toggleGrid() {
        if (this.canvas) {
            this.canvas.toggleGrid();
            const isVisible = this.elements.canvas.classList.contains('show-grid');
            this.updateStatus(`Grid ${isVisible ? 'shown' : 'hidden'}`);
            
            if (this.elements.toggleGrid) {
                this.elements.toggleGrid.classList.toggle('active');
            }
        }
    }

    toggleSnap() {
        if (this.canvas) {
            this.canvas.toggleSnapToGrid();
            this.updateStatus(`Snap to grid ${this.canvas.options.snapToGrid ? 'enabled' : 'disabled'}`);
            
            if (this.elements.toggleSnap) {
                this.elements.toggleSnap.classList.toggle('active', this.canvas.options.snapToGrid);
            }
        }
    }

    markDirty() {
        this.isDirty = true;
        this.updateStatus('Modified');
    }

    updateStatus(message, type = 'info') {
        if (this.elements.statusText) {
            this.elements.statusText.textContent = message;
            this.elements.statusText.className = `editor-status editor-status-${type}`;
        }
    }

    showError(message) {
        const errorDiv = document.createElement('div');
        errorDiv.className = 'editor-error';
        errorDiv.setAttribute('role', 'alert');
        errorDiv.innerHTML = `
            <div class="editor-error-content">
                <span class="editor-error-icon">⚠️</span>
                <span class="editor-error-message">${message}</span>
                <button class="editor-error-close" aria-label="Close error">×</button>
            </div>
        `;
        
        const closeBtn = errorDiv.querySelector('.editor-error-close');
        closeBtn.addEventListener('click', () => errorDiv.remove());
        
        document.body.appendChild(errorDiv);
        
        setTimeout(() => errorDiv.remove(), 5000);
    }
}

if (document.readyState === 'loading') {
    document.addEventListener('DOMContentLoaded', () => {
        const templateId = document.body.dataset.templateId;
        if (templateId) {
            new TemplateEditor({ templateId });
        }
    });
} else {
    const templateId = document.body.dataset.templateId;
    if (templateId) {
        new TemplateEditor({ templateId });
    }
}

if (typeof module !== 'undefined' && module.exports) {
    module.exports = TemplateEditor;
}
