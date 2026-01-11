class ComponentPalettePresets {
    constructor(paletteElement) {
        this.palette = paletteElement;
        this.presets = [];
        this.init();
    }

    async init() {
        await this.loadPresets();
    }

    async loadPresets() {
        try {
            const response = await fetch('/api/component-presets');
            if (response.ok) {
                this.presets = await response.json();
            }
        } catch (error) {
            console.error('Failed to load component presets:', error);
            this.presets = this.getDefaultPresets();
        }
    }

    getDefaultPresets() {
        return [
            {
                name: 'hero-section',
                description: 'Hero section with background image and title',
                category: 'header',
                tags: ['hero', 'header', 'banner'],
                thumbnailUrl: '/static/images/presets/hero-section.png'
            },
            {
                name: 'call-to-action',
                description: 'Call-to-action section with button',
                category: 'action',
                tags: ['cta', 'button', 'action'],
                thumbnailUrl: '/static/images/presets/call-to-action.png'
            },
            {
                name: 'image-gallery',
                description: 'Grid-based image gallery',
                category: 'media',
                tags: ['gallery', 'images', 'grid'],
                thumbnailUrl: '/static/images/presets/image-gallery.png'
            },
            {
                name: 'testimonial',
                description: 'Testimonial section with quote and author',
                category: 'content',
                tags: ['testimonial', 'quote', 'review'],
                thumbnailUrl: '/static/images/presets/testimonial.png'
            }
        ];
    }

    renderPresetsSection() {
        const section = document.createElement('div');
        section.className = 'palette-presets-section';
        
        const header = document.createElement('div');
        header.className = 'palette-section-header';
        header.innerHTML = `
            <h3 class="palette-section-title">Component Presets</h3>
            <button 
                type="button" 
                class="btn-icon btn-collapse-section" 
                aria-label="Collapse presets section"
                aria-expanded="true"
            >
                ▼
            </button>
        `;
        section.appendChild(header);
        
        const list = document.createElement('div');
        list.className = 'palette-presets-list';
        list.setAttribute('role', 'list');
        
        this.presets.forEach(preset => {
            const item = this.createPresetItem(preset);
            list.appendChild(item);
        });
        
        section.appendChild(list);
        return section;
    }

    createPresetItem(preset) {
        const item = document.createElement('div');
        item.className = 'palette-preset-item';
        item.setAttribute('draggable', 'true');
        item.setAttribute('role', 'listitem');
        item.setAttribute('tabindex', '0');
        item.dataset.presetName = preset.name;
        
        item.innerHTML = `
            <div class="palette-preset-thumbnail">
                <img src="${preset.thumbnailUrl}" alt="${preset.name}" />
            </div>
            <div class="palette-preset-content">
                <div class="palette-preset-name">${preset.name}</div>
                <div class="palette-preset-description">${preset.description}</div>
                <div class="palette-preset-tags">
                    ${preset.tags.map(tag => `<span class="preset-tag">${tag}</span>`).join('')}
                </div>
            </div>
        `;
        
        item.addEventListener('dragstart', (e) => this.handlePresetDragStart(e, preset));
        item.addEventListener('dragend', (e) => this.handlePresetDragEnd(e));
        
        item.addEventListener('click', () => {
            this.dispatchEvent('preset-selected', { preset });
        });
        
        item.addEventListener('keydown', (e) => {
            if (e.key === 'Enter' || e.key === ' ') {
                e.preventDefault();
                this.dispatchEvent('preset-selected', { preset });
            }
        });
        
        return item;
    }

    handlePresetDragStart(e, preset) {
        e.dataTransfer.effectAllowed = 'copy';
        e.dataTransfer.setData('preset-name', preset.name);
        e.dataTransfer.setData('text/plain', preset.name);
        
        e.target.classList.add('dragging');
        this.dispatchEvent('preset-drag-started', { preset });
    }

    handlePresetDragEnd(e) {
        e.target.classList.remove('dragging');
        this.dispatchEvent('preset-drag-ended', {});
    }

    async getPresetComponents(presetName) {
        try {
            const response = await fetch(`/api/component-presets/${presetName}`);
            if (response.ok) {
                const preset = await response.json();
                return preset.components;
            }
        } catch (error) {
            console.error(`Failed to load preset ${presetName}:`, error);
        }
        return null;
    }

    filterPresets(searchTerm) {
        if (!searchTerm) return this.presets;
        
        const term = searchTerm.toLowerCase();
        return this.presets.filter(preset =>
            preset.name.toLowerCase().includes(term) ||
            preset.description.toLowerCase().includes(term) ||
            preset.category.toLowerCase().includes(term) ||
            preset.tags.some(tag => tag.toLowerCase().includes(term))
        );
    }

    dispatchEvent(eventName, detail) {
        const event = new CustomEvent(eventName, { detail });
        this.palette.dispatchEvent(event);
        document.dispatchEvent(event);
    }
}

if (typeof module !== 'undefined' && module.exports) {
    module.exports = ComponentPalettePresets;
}
