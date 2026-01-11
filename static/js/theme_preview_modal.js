class ThemePreviewModal {
    constructor() {
        this.modal = document.getElementById('theme-preview-modal');
        this.iframe = document.getElementById('theme-preview-frame');
        this.themeToggle = document.getElementById('preview-theme-toggle');
        this.selectButton = document.getElementById('select-previewed-theme');
        this.currentThemeId = null;
        this.previewTheme = 'light';
        this.lastFocusedElement = null;
        this.init();
    }

    init() {
        if (!this.modal) return;
        
        this.attachEventListeners();
        this.setupFocusTrap();
        this.setupColorChangeListener();
    }

    attachEventListeners() {
        document.addEventListener('theme-preview-requested', (e) => {
            this.open(e.detail.themeId);
        });

        const closeButtons = this.modal.querySelectorAll('.modal-close');
        closeButtons.forEach(btn => {
            btn.addEventListener('click', () => this.close());
        });

        const backdrop = this.modal.querySelector('.modal-backdrop');
        if (backdrop) {
            backdrop.addEventListener('click', () => this.close());
        }

        document.addEventListener('keydown', (e) => {
            if (e.key === 'Escape' && !this.modal.hidden) {
                this.close();
            }
        });

        if (this.themeToggle) {
            this.themeToggle.addEventListener('click', () => {
                this.togglePreviewTheme();
            });
        }

        if (this.selectButton) {
            this.selectButton.addEventListener('click', () => {
                this.selectCurrentTheme();
            });
        }
    }

    open(themeId) {
        this.currentThemeId = themeId;
        this.lastFocusedElement = document.activeElement;
        
        this.loadPreview(themeId);
        
        this.modal.hidden = false;
        document.body.style.overflow = 'hidden';
        
        const firstFocusable = this.modal.querySelector('button, [href], input, select, textarea, [tabindex]:not([tabindex="-1"])');
        if (firstFocusable) {
            firstFocusable.focus();
        }
        
        this.announce('Theme preview opened');
    }

    close() {
        this.modal.hidden = true;
        document.body.style.overflow = '';
        
        if (this.lastFocusedElement) {
            this.lastFocusedElement.focus();
        }
        
        this.announce('Theme preview closed');
    }

    loadPreview(themeId) {
        const formData = this.getEventFormData();
        
        const params = new URLSearchParams({
            theme_id: themeId,
            preview: 'true',
            theme_mode: this.previewTheme,
            ...formData
        });
        
        this.iframe.src = `/api/themes/preview?${params.toString()}`;
    }

    getEventFormData() {
        const form = document.querySelector('form[action*="/events"]');
        if (!form) return {};
        
        const formData = {
            title: form.querySelector('[name="title"]')?.value || 'Sample Event',
            location: form.querySelector('[name="location"]')?.value || 'Sample Location',
            start_time: form.querySelector('[name="start_time"]')?.value || new Date().toISOString(),
            description: form.querySelector('[name="description"]')?.value || 'Sample description'
        };

        const customImageInput = form.querySelector('[name="custom_theme_image_url"]');
        if (customImageInput && customImageInput.value) {
            formData.custom_image_url = customImageInput.value;
        }

        const imagePreview = document.querySelector('.custom-image-preview img');
        if (imagePreview && imagePreview.src && !formData.custom_image_url) {
            formData.custom_image_url = imagePreview.src;
        }

        const customColorInput = document.getElementById('custom-theme-color-value');
        if (customColorInput && customColorInput.value) {
            formData.custom_color = customColorInput.value;
        }

        return formData;
    }

    togglePreviewTheme() {
        this.previewTheme = this.previewTheme === 'light' ? 'dark' : 'light';
        
        const icon = this.themeToggle.querySelector('.theme-icon');
        if (icon) {
            icon.textContent = this.previewTheme === 'dark' ? '☀️' : '🌙';
        }
        
        if (this.currentThemeId) {
            this.loadPreview(this.currentThemeId);
        }
    }

    selectCurrentTheme() {
        if (this.currentThemeId) {
            const event = new CustomEvent('theme-selected', {
                detail: { themeId: this.currentThemeId }
            });
            document.dispatchEvent(event);
            
            this.close();
        }
    }

    setupFocusTrap() {
        this.modal.addEventListener('keydown', (e) => {
            if (e.key === 'Tab') {
                const focusableElements = this.modal.querySelectorAll(
                    'button:not([disabled]), [href], input:not([disabled]), select:not([disabled]), textarea:not([disabled]), [tabindex]:not([tabindex="-1"])'
                );
                
                const firstElement = focusableElements[0];
                const lastElement = focusableElements[focusableElements.length - 1];
                
                if (e.shiftKey && document.activeElement === firstElement) {
                    e.preventDefault();
                    lastElement.focus();
                } else if (!e.shiftKey && document.activeElement === lastElement) {
                    e.preventDefault();
                    firstElement.focus();
                }
            }
        });
    }

    setupColorChangeListener() {
        document.addEventListener('colorChanged', (e) => {
            if (this.currentThemeId && !this.modal.hidden) {
                this.loadPreview(this.currentThemeId);
            }
        });
    }

    announce(message) {
        const announcement = document.createElement('div');
        announcement.setAttribute('role', 'status');
        announcement.setAttribute('aria-live', 'polite');
        announcement.className = 'sr-only';
        announcement.textContent = message;
        document.body.appendChild(announcement);
        setTimeout(() => announcement.remove(), 1000);
    }
}

if (document.readyState === 'loading') {
    document.addEventListener('DOMContentLoaded', () => {
        window.themePreviewModal = new ThemePreviewModal();
    });
} else {
    window.themePreviewModal = new ThemePreviewModal();
}

document.addEventListener('theme-selected', (e) => {
    const themePicker = window.themePicker;
    if (themePicker && themePicker.selectTheme) {
        themePicker.selectTheme(e.detail.themeId);
    }
});
