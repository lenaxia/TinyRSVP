class VisualCanvas {
    constructor(canvasElement, options = {}) {
        this.canvas = canvasElement;
        this.options = {
            gridSize: options.gridSize || 10,
            snapToGrid: options.snapToGrid !== false,
            zoom: options.zoom || 1,
            mode: options.mode || 'desktop',
            ...options
        };
        
        this.components = [];
        this.selectedComponent = null;
        this.dragState = null;
        this.resizeState = null;
        this.history = [];
        this.historyIndex = -1;
        
        this.init();
    }

    init() {
        if (!this.canvas) return;
        
        this.canvas.classList.add('visual-canvas');
        this.canvas.setAttribute('role', 'application');
        this.canvas.setAttribute('aria-label', 'Template editor canvas');
        
        this.attachEventListeners();
        this.initializeKeyboardShortcuts();
    }

    attachEventListeners() {
        this.canvas.addEventListener('click', (e) => this.handleClick(e));
        this.canvas.addEventListener('mousedown', (e) => this.handleMouseDown(e));
        this.canvas.addEventListener('mousemove', (e) => this.handleMouseMove(e));
        this.canvas.addEventListener('mouseup', (e) => this.handleMouseUp(e));
        this.canvas.addEventListener('dragover', (e) => this.handleDragOver(e));
        this.canvas.addEventListener('drop', (e) => this.handleDrop(e));
    }

    initializeKeyboardShortcuts() {
        document.addEventListener('keydown', (e) => {
            if (!this.selectedComponent) return;
            
            const step = e.shiftKey ? this.options.gridSize : 1;
            
            switch (e.key) {
                case 'ArrowUp':
                    e.preventDefault();
                    this.moveComponent(this.selectedComponent.id, 0, -step);
                    break;
                case 'ArrowDown':
                    e.preventDefault();
                    this.moveComponent(this.selectedComponent.id, 0, step);
                    break;
                case 'ArrowLeft':
                    e.preventDefault();
                    this.moveComponent(this.selectedComponent.id, -step, 0);
                    break;
                case 'ArrowRight':
                    e.preventDefault();
                    this.moveComponent(this.selectedComponent.id, step, 0);
                    break;
                case 'Delete':
                case 'Backspace':
                    e.preventDefault();
                    this.removeComponent(this.selectedComponent.id);
                    break;
                case 'z':
                    if (e.ctrlKey || e.metaKey) {
                        e.preventDefault();
                        if (e.shiftKey) {
                            this.redo();
                        } else {
                            this.undo();
                        }
                    }
                    break;
            }
        });
    }

    handleClick(e) {
        const componentEl = e.target.closest('[data-component-id]');
        if (componentEl) {
            const componentId = componentEl.dataset.componentId;
            this.selectComponent(componentId);
        } else {
            this.deselectComponent();
        }
    }

    handleMouseDown(e) {
        const componentEl = e.target.closest('[data-component-id]');
        if (!componentEl) return;
        
        const resizeHandle = e.target.closest('.resize-handle');
        if (resizeHandle) {
            this.startResize(componentEl, e, resizeHandle.dataset.direction);
            return;
        }
        
        this.startDrag(componentEl, e);
    }

    handleMouseMove(e) {
        if (this.dragState) {
            this.updateDrag(e);
        } else if (this.resizeState) {
            this.updateResize(e);
        }
    }

    handleMouseUp(e) {
        if (this.dragState) {
            this.endDrag(e);
        } else if (this.resizeState) {
            this.endResize(e);
        }
    }

    handleDragOver(e) {
        e.preventDefault();
        e.dataTransfer.dropEffect = 'copy';
    }

    handleDrop(e) {
        e.preventDefault();
        
        const componentType = e.dataTransfer.getData('component-type');
        if (!componentType) return;
        
        const rect = this.canvas.getBoundingClientRect();
        const x = e.clientX - rect.left;
        const y = e.clientY - rect.top;
        
        const component = this.createDefaultComponent(componentType, x, y);
        this.addComponent(component);
        
        this.dispatchEvent('component-added', { component });
    }

    startDrag(componentEl, e) {
        const componentId = componentEl.dataset.componentId;
        const component = this.components.find(c => c.id === componentId);
        if (!component) return;
        
        const rect = componentEl.getBoundingClientRect();
        const canvasRect = this.canvas.getBoundingClientRect();
        
        this.dragState = {
            componentId,
            startX: e.clientX,
            startY: e.clientY,
            offsetX: e.clientX - rect.left,
            offsetY: e.clientY - rect.top,
            originalX: this.parsePosition(component.position.x),
            originalY: this.parsePosition(component.position.y)
        };
        
        componentEl.classList.add('dragging');
        this.canvas.style.cursor = 'move';
    }

    updateDrag(e) {
        if (!this.dragState) return;
        
        const deltaX = e.clientX - this.dragState.startX;
        const deltaY = e.clientY - this.dragState.startY;
        
        let newX = this.dragState.originalX + deltaX;
        let newY = this.dragState.originalY + deltaY;
        
        if (this.options.snapToGrid) {
            newX = Math.round(newX / this.options.gridSize) * this.options.gridSize;
            newY = Math.round(newY / this.options.gridSize) * this.options.gridSize;
        }
        
        const component = this.components.find(c => c.id === this.dragState.componentId);
        if (component) {
            component.position.x = `${newX}px`;
            component.position.y = `${newY}px`;
            this.renderComponent(component);
        }
    }

    endDrag(e) {
        if (!this.dragState) return;
        
        const componentEl = this.canvas.querySelector(`[data-component-id="${this.dragState.componentId}"]`);
        if (componentEl) {
            componentEl.classList.remove('dragging');
        }
        
        this.saveState();
        this.dispatchEvent('component-moved', { 
            componentId: this.dragState.componentId 
        });
        
        this.dragState = null;
        this.canvas.style.cursor = '';
    }

    startResize(componentEl, e, direction) {
        const componentId = componentEl.dataset.componentId;
        const component = this.components.find(c => c.id === componentId);
        if (!component) return;
        
        const rect = componentEl.getBoundingClientRect();
        
        this.resizeState = {
            componentId,
            direction,
            startX: e.clientX,
            startY: e.clientY,
            originalWidth: rect.width,
            originalHeight: rect.height
        };
        
        componentEl.classList.add('resizing');
    }

    updateResize(e) {
        if (!this.resizeState) return;
        
        const deltaX = e.clientX - this.resizeState.startX;
        const deltaY = e.clientY - this.resizeState.startY;
        
        const component = this.components.find(c => c.id === this.resizeState.componentId);
        if (!component) return;
        
        let newWidth = this.resizeState.originalWidth;
        let newHeight = this.resizeState.originalHeight;
        
        const direction = this.resizeState.direction;
        if (direction.includes('e')) newWidth += deltaX;
        if (direction.includes('w')) newWidth -= deltaX;
        if (direction.includes('s')) newHeight += deltaY;
        if (direction.includes('n')) newHeight -= deltaY;
        
        if (this.options.snapToGrid) {
            newWidth = Math.round(newWidth / this.options.gridSize) * this.options.gridSize;
            newHeight = Math.round(newHeight / this.options.gridSize) * this.options.gridSize;
        }
        
        newWidth = Math.max(20, newWidth);
        newHeight = Math.max(20, newHeight);
        
        component.dimensions.width = `${newWidth}px`;
        component.dimensions.height = `${newHeight}px`;
        this.renderComponent(component);
    }

    endResize(e) {
        if (!this.resizeState) return;
        
        const componentEl = this.canvas.querySelector(`[data-component-id="${this.resizeState.componentId}"]`);
        if (componentEl) {
            componentEl.classList.remove('resizing');
        }
        
        this.saveState();
        this.dispatchEvent('component-resized', { 
            componentId: this.resizeState.componentId 
        });
        
        this.resizeState = null;
    }

    addComponent(component) {
        if (!component.id) {
            component.id = this.generateComponentId();
        }
        
        this.components.push(component);
        this.renderComponent(component);
        this.saveState();
        this.selectComponent(component.id);
    }

    removeComponent(componentId) {
        const index = this.components.findIndex(c => c.id === componentId);
        if (index === -1) return;
        
        this.components.splice(index, 1);
        const componentEl = this.canvas.querySelector(`[data-component-id="${componentId}"]`);
        if (componentEl) {
            componentEl.remove();
        }
        
        this.saveState();
        this.deselectComponent();
        this.dispatchEvent('component-removed', { componentId });
    }

    moveComponent(componentId, deltaX, deltaY) {
        const component = this.components.find(c => c.id === componentId);
        if (!component) return;
        
        const currentX = this.parsePosition(component.position.x);
        const currentY = this.parsePosition(component.position.y);
        
        component.position.x = `${currentX + deltaX}px`;
        component.position.y = `${currentY + deltaY}px`;
        
        this.renderComponent(component);
        this.saveState();
        this.dispatchEvent('component-moved', { componentId });
    }

    updateComponent(componentId, updates) {
        const component = this.components.find(c => c.id === componentId);
        if (!component) return;
        
        Object.assign(component, updates);
        this.renderComponent(component);
        this.saveState();
        this.dispatchEvent('component-updated', { componentId, updates });
    }

    selectComponent(componentId) {
        this.deselectComponent();
        
        const component = this.components.find(c => c.id === componentId);
        if (!component) return;
        
        this.selectedComponent = component;
        
        const componentEl = this.canvas.querySelector(`[data-component-id="${componentId}"]`);
        if (componentEl) {
            componentEl.classList.add('selected');
            componentEl.setAttribute('aria-selected', 'true');
            this.addResizeHandles(componentEl);
        }
        
        this.dispatchEvent('component-selected', { component });
    }

    deselectComponent() {
        if (!this.selectedComponent) return;
        
        const componentEl = this.canvas.querySelector(`[data-component-id="${this.selectedComponent.id}"]`);
        if (componentEl) {
            componentEl.classList.remove('selected');
            componentEl.setAttribute('aria-selected', 'false');
            this.removeResizeHandles(componentEl);
        }
        
        this.selectedComponent = null;
        this.dispatchEvent('component-deselected', {});
    }

    addResizeHandles(componentEl) {
        const directions = ['nw', 'n', 'ne', 'e', 'se', 's', 'sw', 'w'];
        directions.forEach(dir => {
            const handle = document.createElement('div');
            handle.className = `resize-handle resize-handle-${dir}`;
            handle.dataset.direction = dir;
            handle.setAttribute('role', 'button');
            handle.setAttribute('aria-label', `Resize ${dir}`);
            componentEl.appendChild(handle);
        });
    }

    removeResizeHandles(componentEl) {
        const handles = componentEl.querySelectorAll('.resize-handle');
        handles.forEach(handle => handle.remove());
    }

    renderComponent(component) {
        let componentEl = this.canvas.querySelector(`[data-component-id="${component.id}"]`);
        
        if (!componentEl) {
            componentEl = document.createElement('div');
            componentEl.dataset.componentId = component.id;
            componentEl.className = 'canvas-component';
            componentEl.setAttribute('role', 'button');
            componentEl.setAttribute('tabindex', '0');
            this.canvas.appendChild(componentEl);
        }
        
        componentEl.className = `canvas-component canvas-component-${component.type.toLowerCase()}`;
        componentEl.style.position = 'absolute';
        componentEl.style.left = component.position.x;
        componentEl.style.top = component.position.y;
        componentEl.style.width = component.dimensions.width;
        componentEl.style.height = component.dimensions.height;
        componentEl.style.zIndex = component.zIndex || 1;
        componentEl.style.display = component.visible === false ? 'none' : '';
        
        this.renderComponentContent(componentEl, component);
        
        if (this.selectedComponent && this.selectedComponent.id === component.id) {
            componentEl.classList.add('selected');
        }
    }

    renderComponentContent(componentEl, component) {
        switch (component.type) {
            case 'TextBox':
                this.renderTextBox(componentEl, component);
                break;
            case 'Image':
                this.renderImage(componentEl, component);
                break;
            case 'Background':
                this.renderBackground(componentEl, component);
                break;
            case 'Overlay':
                this.renderOverlay(componentEl, component);
                break;
            case 'Container':
                this.renderContainer(componentEl, component);
                break;
            case 'Divider':
                this.renderDivider(componentEl, component);
                break;
        }
    }

    renderTextBox(componentEl, component) {
        componentEl.innerHTML = '';
        const textEl = document.createElement('div');
        textEl.className = 'component-text';
        textEl.textContent = component.content?.text || 'Text Box';
        
        if (component.content) {
            Object.assign(textEl.style, {
                textAlign: component.content.textAlign,
                fontFamily: component.content.fontFamily,
                fontSize: component.content.fontSize,
                fontWeight: component.content.fontWeight,
                color: component.content.color,
                lineHeight: component.content.lineHeight,
                letterSpacing: component.content.letterSpacing,
                textTransform: component.content.textTransform,
                textShadow: component.content.textShadow
            });
        }
        
        componentEl.appendChild(textEl);
    }

    renderImage(componentEl, component) {
        componentEl.innerHTML = '';
        const imgEl = document.createElement('img');
        imgEl.src = component.content?.src || '/static/images/placeholder.png';
        imgEl.alt = component.content?.alt || 'Image component';
        imgEl.style.width = '100%';
        imgEl.style.height = '100%';
        imgEl.style.objectFit = component.content?.objectFit || 'cover';
        imgEl.style.objectPosition = component.content?.objectPosition || 'center';
        imgEl.style.opacity = component.content?.opacity || 1;
        imgEl.style.filter = component.content?.filter || 'none';
        componentEl.appendChild(imgEl);
    }

    renderBackground(componentEl, component) {
        componentEl.innerHTML = '';
        const content = component.content || {};
        
        if (content.type === 'color') {
            componentEl.style.backgroundColor = content.color || '#f8f9fa';
        } else if (content.type === 'gradient') {
            componentEl.style.background = content.gradient;
        } else if (content.type === 'image' && content.image) {
            componentEl.style.backgroundImage = `url(${content.image.src})`;
            componentEl.style.backgroundRepeat = content.image.repeat || 'no-repeat';
            componentEl.style.backgroundSize = content.image.size || 'cover';
            componentEl.style.backgroundPosition = content.image.position || 'center';
        }
        
        const label = document.createElement('div');
        label.className = 'component-label';
        label.textContent = 'Background';
        componentEl.appendChild(label);
    }

    renderOverlay(componentEl, component) {
        componentEl.innerHTML = '';
        const content = component.content || {};
        
        Object.assign(componentEl.style, {
            backgroundColor: content.backgroundColor || 'transparent',
            borderRadius: content.borderRadius,
            border: content.border,
            boxShadow: content.boxShadow,
            clipPath: content.clipPath
        });
        
        if (content.placeholder?.show) {
            const placeholder = document.createElement('div');
            placeholder.className = 'component-placeholder';
            placeholder.textContent = content.placeholder.text || 'Overlay';
            componentEl.appendChild(placeholder);
        }
    }

    renderContainer(componentEl, component) {
        componentEl.innerHTML = '';
        const layout = component.layout || {};
        
        Object.assign(componentEl.style, {
            display: layout.display || 'flex',
            flexDirection: layout.flexDirection,
            alignItems: layout.alignItems,
            justifyContent: layout.justifyContent,
            gap: layout.gap,
            padding: layout.padding
        });
        
        if (component.style) {
            Object.assign(componentEl.style, component.style);
        }
        
        const label = document.createElement('div');
        label.className = 'component-label';
        label.textContent = 'Container';
        componentEl.appendChild(label);
    }

    renderDivider(componentEl, component) {
        componentEl.innerHTML = '';
        if (component.style) {
            Object.assign(componentEl.style, component.style);
        }
    }

    render(components) {
        this.components = components || [];
        this.canvas.innerHTML = '';
        
        this.components
            .sort((a, b) => (a.zIndex || 0) - (b.zIndex || 0))
            .forEach(component => this.renderComponent(component));
    }

    setZoom(zoom) {
        this.options.zoom = zoom;
        this.canvas.style.transform = `scale(${zoom})`;
        this.canvas.style.transformOrigin = 'top left';
    }

    setMode(mode) {
        this.options.mode = mode;
        this.canvas.classList.remove('mode-desktop', 'mode-tablet', 'mode-mobile');
        this.canvas.classList.add(`mode-${mode}`);
        
        const widths = {
            desktop: '1200px',
            tablet: '768px',
            mobile: '375px'
        };
        
        this.canvas.style.maxWidth = widths[mode] || widths.desktop;
    }

    toggleGrid() {
        this.canvas.classList.toggle('show-grid');
    }

    toggleSnapToGrid() {
        this.options.snapToGrid = !this.options.snapToGrid;
    }

    createDefaultComponent(type, x, y) {
        const defaults = {
            TextBox: {
                type: 'TextBox',
                position: { mode: 'absolute', x: `${x}px`, y: `${y}px` },
                dimensions: { width: '200px', height: 'auto' },
                zIndex: 10,
                visible: true,
                content: {
                    text: 'New Text',
                    textAlign: 'center',
                    fontSize: '16px',
                    color: '#000000'
                }
            },
            Image: {
                type: 'Image',
                position: { mode: 'absolute', x: `${x}px`, y: `${y}px` },
                dimensions: { width: '200px', height: '200px' },
                zIndex: 5,
                visible: true,
                content: {
                    src: '/static/images/placeholder.png',
                    objectFit: 'cover'
                }
            },
            Background: {
                type: 'Background',
                position: { mode: 'absolute', x: '0', y: '0' },
                dimensions: { width: '100%', height: '100%' },
                zIndex: 0,
                visible: true,
                content: {
                    type: 'color',
                    color: '#f8f9fa'
                }
            },
            Overlay: {
                type: 'Overlay',
                position: { mode: 'absolute', x: `${x}px`, y: `${y}px` },
                dimensions: { width: '150px', height: '150px' },
                zIndex: 20,
                visible: true,
                content: {
                    backgroundColor: 'rgba(255, 255, 255, 0.9)',
                    borderRadius: '8px'
                }
            },
            Container: {
                type: 'Container',
                position: { mode: 'absolute', x: `${x}px`, y: `${y}px` },
                dimensions: { width: '300px', height: 'auto' },
                zIndex: 5,
                visible: true,
                layout: {
                    display: 'flex',
                    flexDirection: 'column',
                    gap: '10px',
                    padding: '20px'
                }
            },
            Divider: {
                type: 'Divider',
                position: { mode: 'absolute', x: `${x}px`, y: `${y}px` },
                dimensions: { width: '200px', height: '2px' },
                zIndex: 5,
                visible: true,
                style: {
                    backgroundColor: '#e5e7eb'
                }
            }
        };
        
        return defaults[type] || defaults.TextBox;
    }

    generateComponentId() {
        return `component-${Date.now()}-${Math.random().toString(36).substr(2, 9)}`;
    }

    parsePosition(value) {
        if (typeof value === 'number') return value;
        if (typeof value === 'string') {
            return parseInt(value) || 0;
        }
        return 0;
    }

    saveState() {
        const state = JSON.parse(JSON.stringify(this.components));
        
        if (this.historyIndex < this.history.length - 1) {
            this.history = this.history.slice(0, this.historyIndex + 1);
        }
        
        this.history.push(state);
        this.historyIndex++;
        
        if (this.history.length > 50) {
            this.history.shift();
            this.historyIndex--;
        }
    }

    undo() {
        if (this.historyIndex <= 0) return;
        
        this.historyIndex--;
        this.components = JSON.parse(JSON.stringify(this.history[this.historyIndex]));
        this.render(this.components);
        this.dispatchEvent('state-changed', { action: 'undo' });
    }

    redo() {
        if (this.historyIndex >= this.history.length - 1) return;
        
        this.historyIndex++;
        this.components = JSON.parse(JSON.stringify(this.history[this.historyIndex]));
        this.render(this.components);
        this.dispatchEvent('state-changed', { action: 'redo' });
    }

    getComponents() {
        return JSON.parse(JSON.stringify(this.components));
    }

    dispatchEvent(eventName, detail) {
        const event = new CustomEvent(eventName, { detail });
        this.canvas.dispatchEvent(event);
        document.dispatchEvent(event);
    }
}

if (typeof module !== 'undefined' && module.exports) {
    module.exports = VisualCanvas;
}
