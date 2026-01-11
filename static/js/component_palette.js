class ComponentPalette {
    constructor(paletteElement) {
        this.palette = paletteElement;
        this.componentTypes = [
            {
                type: 'TextBox',
                name: 'Text Box',
                icon: '📝',
                description: 'Add text content with customizable styling'
            },
            {
                type: 'Image',
                name: 'Image',
                icon: '🖼️',
                description: 'Add an image with positioning controls'
            },
            {
                type: 'Background',
                name: 'Background',
                icon: '🎨',
                description: 'Set background color, gradient, or image'
            },
            {
                type: 'Overlay',
                name: 'Overlay',
                icon: '⬜',
                description: 'Add transparent overlay for photos'
            },
            {
                type: 'Container',
                name: 'Container',
                icon: '📦',
                description: 'Group components with flex layout'
            },
            {
                type: 'Divider',
                name: 'Divider',
                icon: '➖',
                description: 'Add horizontal or vertical divider'
            }
        ];
        
        this.searchTerm = '';
        this.init();
    }

    init() {
        if (!this.palette) return;
        
        this.render();
        this.attachEventListeners();
    }

    render() {
        this.palette.innerHTML = '';
        
        const header = document.createElement('div');
        header.className = 'palette-header';
        header.innerHTML = `
            <h2 class="palette-title">Components</h2>
            <div class="palette-search">
                <input 
                    type="text" 
                    id="component-search" 
                    class="palette-search-input"
                    placeholder="Search components..."
                    aria-label="Search components"
                />
            </div>
        `;
        this.palette.appendChild(header);
        
        const list = document.createElement('div');
        list.className = 'palette-list';
        list.setAttribute('role', 'list');
        
        const filteredTypes = this.filterComponents();
        
        if (filteredTypes.length === 0) {
            const empty = document.createElement('div');
            empty.className = 'palette-empty';
            empty.textContent = 'No components found';
            list.appendChild(empty);
        } else {
            filteredTypes.forEach(componentType => {
                const item = this.createPaletteItem(componentType);
                list.appendChild(item);
            });
        }
        
        this.palette.appendChild(list);
    }

    createPaletteItem(componentType) {
        const item = document.createElement('div');
        item.className = 'palette-item';
        item.setAttribute('draggable', 'true');
        item.setAttribute('role', 'listitem');
        item.setAttribute('tabindex', '0');
        item.dataset.componentType = componentType.type;
        
        item.innerHTML = `
            <div class="palette-item-icon" aria-hidden="true">${componentType.icon}</div>
            <div class="palette-item-content">
                <div class="palette-item-name">${componentType.name}</div>
                <div class="palette-item-description">${componentType.description}</div>
            </div>
        `;
        
        item.addEventListener('dragstart', (e) => this.handleDragStart(e, componentType));
        item.addEventListener('dragend', (e) => this.handleDragEnd(e));
        
        item.addEventListener('click', () => {
            this.dispatchEvent('component-selected', { componentType });
        });
        
        item.addEventListener('keydown', (e) => {
            if (e.key === 'Enter' || e.key === ' ') {
                e.preventDefault();
                this.dispatchEvent('component-selected', { componentType });
            }
        });
        
        return item;
    }

    attachEventListeners() {
        const searchInput = this.palette.querySelector('#component-search');
        if (searchInput) {
            searchInput.addEventListener('input', (e) => {
                this.searchTerm = e.target.value.toLowerCase();
                this.render();
            });
        }
    }

    handleDragStart(e, componentType) {
        e.dataTransfer.effectAllowed = 'copy';
        e.dataTransfer.setData('component-type', componentType.type);
        e.dataTransfer.setData('text/plain', componentType.name);
        
        const dragImage = e.target.cloneNode(true);
        dragImage.style.opacity = '0.5';
        dragImage.style.position = 'absolute';
        dragImage.style.top = '-1000px';
        document.body.appendChild(dragImage);
        e.dataTransfer.setDragImage(dragImage, 0, 0);
        
        setTimeout(() => dragImage.remove(), 0);
        
        e.target.classList.add('dragging');
        this.dispatchEvent('drag-started', { componentType });
    }

    handleDragEnd(e) {
        e.target.classList.remove('dragging');
        this.dispatchEvent('drag-ended', {});
    }

    filterComponents() {
        if (!this.searchTerm) {
            return this.componentTypes;
        }
        
        return this.componentTypes.filter(type => 
            type.name.toLowerCase().includes(this.searchTerm) ||
            type.description.toLowerCase().includes(this.searchTerm) ||
            type.type.toLowerCase().includes(this.searchTerm)
        );
    }

    getComponentType(type) {
        return this.componentTypes.find(ct => ct.type === type);
    }

    dispatchEvent(eventName, detail) {
        const event = new CustomEvent(eventName, { detail });
        this.palette.dispatchEvent(event);
        document.dispatchEvent(event);
    }
}

if (typeof module !== 'undefined' && module.exports) {
    module.exports = ComponentPalette;
}
