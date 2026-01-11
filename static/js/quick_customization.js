/**
 * Quick Customization Controls
 * Provides one-click customization options for common changes
 */

class QuickCustomization {
    constructor(eventCustomization) {
        this.customization = eventCustomization;
        this.init();
    }

    init() {
        this.setupQuickActions();
    }

    setupQuickActions() {
        document.getElementById('quick-change-header-image')?.addEventListener('click', () => {
            this.quickChangeHeaderImage();
        });

        document.getElementById('quick-change-colors')?.addEventListener('click', () => {
            this.quickChangeColors();
        });

        document.getElementById('quick-change-title')?.addEventListener('click', () => {
            this.quickChangeTitle();
        });

        document.getElementById('quick-add-subtitle')?.addEventListener('click', () => {
            this.quickAddSubtitle();
        });

        document.getElementById('quick-add-photo')?.addEventListener('click', () => {
            this.quickAddPhoto();
        });
    }

    quickChangeHeaderImage() {
        const input = document.createElement('input');
        input.type = 'file';
        input.accept = 'image/*';
        
        input.addEventListener('change', async (e) => {
            const file = e.target.files[0];
            if (!file) return;

            try {
                const url = await this.uploadImage(file);
                this.updateHeaderImage(url);
                this.customization.showSuccess('Header image updated');
            } catch (error) {
                this.customization.showError('Failed to upload image');
            }
        });

        input.click();
    }

    async uploadImage(file) {
        const formData = new FormData();
        formData.append('image', file);
        formData.append('eventID', this.customization.eventID);

        const response = await fetch('/api/upload/image', {
            method: 'POST',
            body: formData
        });

        if (!response.ok) {
            throw new Error('Upload failed');
        }

        const data = await response.json();
        return data.url;
    }

    updateHeaderImage(url) {
        const headerComponent = this.findComponentByType('Image');
        if (headerComponent) {
            this.customization.updateComponentProperty(headerComponent.id, 'image-src', url);
            this.customization.isDirty = true;
            this.customization.renderPreview();
        }
    }

    quickChangeColors() {
        const modal = this.createColorPickerModal();
        document.body.appendChild(modal);
    }

    createColorPickerModal() {
        const modal = document.createElement('div');
        modal.className = 'modal-overlay';
        modal.innerHTML = `
            <div class="modal-content">
                <div class="modal-header">
                    <h3>Quick Color Change</h3>
                    <button type="button" class="modal-close">&times;</button>
                </div>
                <div class="modal-body">
                    <div class="property-field">
                        <label>Primary Color</label>
                        <input type="color" id="quick-primary-color" value="#2c3e50">
                    </div>
                    <div class="property-field">
                        <label>Accent Color</label>
                        <input type="color" id="quick-accent-color" value="#3498db">
                    </div>
                    <div class="property-field">
                        <label>Text Color</label>
                        <input type="color" id="quick-text-color" value="#333333">
                    </div>
                </div>
                <div class="modal-footer">
                    <button type="button" class="btn-cancel" id="color-cancel">Cancel</button>
                    <button type="button" class="btn-save" id="color-apply">Apply</button>
                </div>
            </div>
        `;

        modal.querySelector('.modal-close').addEventListener('click', () => modal.remove());
        modal.querySelector('#color-cancel').addEventListener('click', () => modal.remove());
        modal.querySelector('#color-apply').addEventListener('click', () => {
            this.applyColorScheme(
                modal.querySelector('#quick-primary-color').value,
                modal.querySelector('#quick-accent-color').value,
                modal.querySelector('#quick-text-color').value
            );
            modal.remove();
        });

        return modal;
    }

    applyColorScheme(primaryColor, accentColor, textColor) {
        const textComponents = this.findComponentsByType('TextBox');
        
        textComponents.forEach(component => {
            if (component.id.includes('title')) {
                this.customization.updateComponentProperty(component.id, 'text-color', primaryColor);
            } else {
                this.customization.updateComponentProperty(component.id, 'text-color', textColor);
            }
        });

        this.customization.isDirty = true;
        this.customization.renderPreview();
        this.customization.showSuccess('Colors updated');
    }

    quickChangeTitle() {
        const titleComponent = this.findComponentByID('title-text') || 
                              this.findComponentByID('event-title');
        
        if (!titleComponent) {
            this.customization.showError('No title component found');
            return;
        }

        this.customization.selectComponent(titleComponent.id);
        
        const textInput = document.getElementById('text-content');
        if (textInput) {
            textInput.focus();
            textInput.select();
        }
    }

    quickAddSubtitle() {
        const subtitleID = 'custom-subtitle-' + Date.now();
        
        const subtitle = {
            id: subtitleID,
            type: 'TextBox',
            position: {
                mode: 'absolute',
                x: '50%',
                y: '350px'
            },
            dimensions: {
                width: '70%',
                height: 'auto'
            },
            zIndex: 10,
            visible: true,
            content: {
                text: 'Add your subtitle here',
                textAlign: 'center',
                fontFamily: 'Lato, sans-serif',
                fontSize: '20px',
                fontWeight: '300',
                color: '#666666',
                fontStyle: 'italic'
            }
        };

        if (!this.customization.customizationData.eventOverrides) {
            this.customization.customizationData.eventOverrides = {
                version: '1.0',
                overrides: [],
                additions: [],
                removals: []
            };
        }

        this.customization.customizationData.eventOverrides.additions.push(subtitle);
        this.customization.isDirty = true;
        this.customization.renderComponentList();
        this.customization.renderPreview();
        this.customization.selectComponent(subtitleID);
        this.customization.showSuccess('Subtitle added');
    }

    quickAddPhoto() {
        const input = document.createElement('input');
        input.type = 'file';
        input.accept = 'image/*';
        
        input.addEventListener('change', async (e) => {
            const file = e.target.files[0];
            if (!file) return;

            try {
                const url = await this.uploadImage(file);
                this.addPhotoOverlay(url);
                this.customization.showSuccess('Photo added');
            } catch (error) {
                this.customization.showError('Failed to add photo');
            }
        });

        input.click();
    }

    addPhotoOverlay(imageURL) {
        const overlayID = 'custom-photo-' + Date.now();
        
        const overlay = {
            id: overlayID,
            type: 'Overlay',
            position: {
                mode: 'absolute',
                x: '50%',
                y: '150px'
            },
            dimensions: {
                width: '200px',
                height: '200px'
            },
            zIndex: 20,
            visible: true,
            content: {
                backgroundColor: 'transparent',
                borderRadius: '50%',
                border: '4px solid white',
                boxShadow: '0 8px 16px rgba(0,0,0,0.2)',
                backgroundImage: `url(${imageURL})`,
                backgroundSize: 'cover',
                backgroundPosition: 'center'
            }
        };

        if (!this.customization.customizationData.eventOverrides) {
            this.customization.customizationData.eventOverrides = {
                version: '1.0',
                overrides: [],
                additions: [],
                removals: []
            };
        }

        this.customization.customizationData.eventOverrides.additions.push(overlay);
        this.customization.isDirty = true;
        this.customization.renderComponentList();
        this.customization.renderPreview();
        this.customization.selectComponent(overlayID);
    }

    findComponentByType(type) {
        return this.customization.customizationData.mergedConfig?.components?.find(
            c => c.type === type
        );
    }

    findComponentsByType(type) {
        return this.customization.customizationData.mergedConfig?.components?.filter(
            c => c.type === type
        ) || [];
    }

    findComponentByID(id) {
        return this.customization.customizationData.mergedConfig?.components?.find(
            c => c.id === id
        );
    }
}

if (typeof module !== 'undefined' && module.exports) {
    module.exports = QuickCustomization;
}
