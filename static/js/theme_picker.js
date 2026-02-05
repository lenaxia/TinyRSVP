class ThemePicker {
    constructor() {
        this.gallery = document.querySelector('.theme-gallery');
        this.filterSelect = document.getElementById('theme-category-filter');
        this.hiddenInput = document.getElementById('selected-theme-id');
        this.picker = document.querySelector('.theme-picker');
        this.galleryModeBtn = document.getElementById('gallery-mode-btn');
        this.designModeBtn = document.getElementById('design-mode-btn');
        this.galleryContainer = document.getElementById('theme-gallery-container');
        this.designContainer = document.getElementById('design-mode-container');
        this.designThemeSelect = document.getElementById('design-theme-select');
        this.livePreviewFrame = document.getElementById('live-preview-frame');
        this.loadingIndicator = document.querySelector('.live-preview-loading');
        this.errorIndicator = document.querySelector('.live-preview-error');
        this.retryBtn = document.querySelector('.btn-retry-preview');
        this.mobileEditBtn = document.getElementById('mobile-edit-btn');
        this.mobilePreviewBtn = document.getElementById('mobile-preview-btn');
        this.eventForm = document.querySelector('.event-form');
        this.debounceTimer = null;
        this.currentMode = 'gallery';
        this.init();
    }

    init() {
        if (!this.gallery) return;
        
        this.attachEventListeners();
        this.initializeKeyboardNavigation();
    }

    attachEventListeners() {
        if (this.filterSelect) {
            this.filterSelect.addEventListener('change', (e) => {
                this.filterThemes(e.target.value);
            });
        }

        this.gallery.addEventListener('click', (e) => {
            const selectBtn = e.target.closest('.btn-select');
            if (selectBtn) {
                const themeId = selectBtn.dataset.themeId;
                this.selectTheme(themeId);
                return;
            }

            const previewBtn = e.target.closest('.btn-preview');
            if (previewBtn) {
                const themeId = previewBtn.dataset.themeId;
                this.previewTheme(themeId);
                return;
            }

            const card = e.target.closest('.theme-card');
            if (card && !selectBtn && !previewBtn) {
                const themeId = card.dataset.themeId;
                this.selectTheme(themeId);
            }
        });

        if (this.galleryModeBtn) {
            this.galleryModeBtn.addEventListener('click', () => {
                this.switchMode('gallery');
            });
        }

        if (this.designModeBtn) {
            this.designModeBtn.addEventListener('click', () => {
                this.switchMode('design');
            });
        }

        if (this.designThemeSelect) {
            this.designThemeSelect.addEventListener('change', () => {
                const themeId = this.designThemeSelect.value;
                this.selectTheme(themeId);
                this.updatePreview();
            });
        }

        if (this.retryBtn) {
            this.retryBtn.addEventListener('click', () => {
                this.updatePreview();
            });
        }

        if (this.mobileEditBtn) {
            this.mobileEditBtn.addEventListener('click', () => {
                this.toggleMobileView('edit');
            });
        }

        if (this.mobilePreviewBtn) {
            this.mobilePreviewBtn.addEventListener('click', () => {
                this.toggleMobileView('preview');
            });
        }

        this.attachFormWatchers();
    }

    attachFormWatchers() {
        if (!this.eventForm) return;

        const watchedFields = [
            'title',
            'description',
            'location',
            'start_time',
            'end_time',
            'custom_theme_image_url'
        ];

        watchedFields.forEach(fieldName => {
            const field = this.eventForm.querySelector(`[name="${fieldName}"]`);
            if (field) {
                field.addEventListener('input', () => {
                    if (this.currentMode === 'design') {
                        this.debouncedUpdatePreview();
                    }
                });
            }
        });

        const colorInput = document.getElementById('custom-theme-color-value');
        if (colorInput) {
            document.addEventListener('colorChanged', () => {
                if (this.currentMode === 'design') {
                    this.debouncedUpdatePreview();
                }
            });
        }
    }

    debouncedUpdatePreview() {
        if (this.debounceTimer) {
            clearTimeout(this.debounceTimer);
        }
        this.debounceTimer = setTimeout(() => {
            this.updatePreview();
        }, 500);
    }

    switchMode(mode) {
        this.currentMode = mode;

        if (this.debounceTimer) {
            clearTimeout(this.debounceTimer);
            this.debounceTimer = null;
        }

        if (mode === 'gallery') {
            this.galleryModeBtn.setAttribute('aria-selected', 'true');
            this.designModeBtn.setAttribute('aria-selected', 'false');
            this.galleryContainer.hidden = false;
            this.galleryContainer.setAttribute('aria-hidden', 'false');
            this.designContainer.hidden = true;
            this.designContainer.setAttribute('aria-hidden', 'true');
            this.picker.setAttribute('data-mode', 'gallery');
            this.livePreviewFrame.src = 'about:blank';
        } else {
            this.galleryModeBtn.setAttribute('aria-selected', 'false');
            this.designModeBtn.setAttribute('aria-selected', 'true');
            this.galleryContainer.hidden = true;
            this.galleryContainer.setAttribute('aria-hidden', 'true');
            this.designContainer.hidden = false;
            this.designContainer.setAttribute('aria-hidden', 'false');
            this.picker.setAttribute('data-mode', 'design');
            this.updatePreview();
        }
    }

    toggleMobileView(view) {
        if (this.mobileEditBtn && this.mobilePreviewBtn) {
            if (view === 'edit') {
                this.mobileEditBtn.setAttribute('aria-selected', 'true');
                this.mobilePreviewBtn.setAttribute('aria-selected', 'false');
                this.eventForm.setAttribute('data-mobile-view', 'edit');
            } else {
                this.mobileEditBtn.setAttribute('aria-selected', 'false');
                this.mobilePreviewBtn.setAttribute('aria-selected', 'true');
                this.eventForm.setAttribute('data-mobile-view', 'preview');
            }
        }
    }

    updatePreview() {
        if (!this.livePreviewFrame) return;

        this.showLoading();
        this.hideError();

        const previewURL = this.buildPreviewURL();
        
        if (previewURL.length > 2000) {
            this.showError();
            this.hideLoading();
            return;
        }

        this.livePreviewFrame.src = previewURL;

        const loadTimeout = setTimeout(() => {
            this.showError();
            this.hideLoading();
        }, 10000);

        this.livePreviewFrame.onload = () => {
            clearTimeout(loadTimeout);
            this.hideLoading();
            this.hideError();
        };

        this.livePreviewFrame.onerror = () => {
            clearTimeout(loadTimeout);
            this.showError();
            this.hideLoading();
        };
    }

    buildPreviewURL() {
        const params = new URLSearchParams();
        
        const themeId = this.hiddenInput?.value || this.designThemeSelect?.value;
        if (themeId) {
            params.set('theme_id', themeId);
        }

        const fields = {
            title: this.eventForm?.querySelector('[name="title"]')?.value,
            description: this.eventForm?.querySelector('[name="description"]')?.value,
            location: this.eventForm?.querySelector('[name="location"]')?.value,
            start_time: this.eventForm?.querySelector('[name="start_time"]')?.value,
            end_time: this.eventForm?.querySelector('[name="end_time"]')?.value
        };

        Object.entries(fields).forEach(([key, value]) => {
            if (value) {
                params.set(key, value);
            }
        });

        const customImageURL = this.eventForm?.querySelector('[name="custom_theme_image_url"]')?.value;
        if (customImageURL) {
            params.set('custom_image_url', customImageURL);
        }

        const customColor = document.getElementById('custom-theme-color-value')?.value;
        if (customColor) {
            params.set('custom_color', customColor);
        }

        return `/api/themes/preview?${params.toString()}`;
    }

    showLoading() {
        if (this.loadingIndicator) {
            this.loadingIndicator.hidden = false;
        }
    }

    hideLoading() {
        if (this.loadingIndicator) {
            this.loadingIndicator.hidden = true;
        }
    }

    showError() {
        if (this.errorIndicator) {
            this.errorIndicator.hidden = false;
        }
    }

    hideError() {
        if (this.errorIndicator) {
            this.errorIndicator.hidden = true;
        }
    }

    filterThemes(category) {
        const cards = this.gallery.querySelectorAll('.theme-card');
        
        cards.forEach(card => {
            const cardCategory = card.getAttribute('data-category');
            if (!category || cardCategory === category) {
                card.style.display = '';
            } else {
                card.style.display = 'none';
            }
        });
    }

    selectTheme(themeId) {
        const previousSelected = this.gallery.querySelector('.theme-card.selected');
        if (previousSelected) {
            previousSelected.classList.remove('selected');
            previousSelected.setAttribute('aria-checked', 'false');
            previousSelected.setAttribute('tabindex', '-1');
        }

        const card = this.gallery.querySelector(`[data-theme-id="${themeId}"]`);
        if (card) {
            card.classList.add('selected');
            card.setAttribute('aria-checked', 'true');
            card.setAttribute('tabindex', '0');
            card.focus();
        }

        if (this.hiddenInput) {
            this.hiddenInput.value = themeId;
        }

        if (this.designThemeSelect) {
            this.designThemeSelect.value = themeId;
        }

        this.announceSelection(card);
    }

    previewTheme(themeId) {
        const event = new CustomEvent('theme-preview-requested', {
            detail: { themeId }
        });
        document.dispatchEvent(event);
    }

    initializeKeyboardNavigation() {
        this.gallery.addEventListener('keydown', (e) => {
            const card = e.target.closest('.theme-card');
            if (!card) return;

            let nextCard = null;

            switch (e.key) {
                case 'Enter':
                case ' ':
                    e.preventDefault();
                    this.selectTheme(card.dataset.themeId);
                    break;
                case 'ArrowRight':
                case 'ArrowDown':
                    e.preventDefault();
                    nextCard = card.nextElementSibling;
                    while (nextCard && nextCard.style.display === 'none') {
                        nextCard = nextCard.nextElementSibling;
                    }
                    if (nextCard) nextCard.focus();
                    break;
                case 'ArrowLeft':
                case 'ArrowUp':
                    e.preventDefault();
                    nextCard = card.previousElementSibling;
                    while (nextCard && nextCard.style.display === 'none') {
                        nextCard = nextCard.previousElementSibling;
                    }
                    if (nextCard) nextCard.focus();
                    break;
                case 'Home':
                    e.preventDefault();
                    const firstCard = this.gallery.querySelector('.theme-card:not([style*="display: none"])');
                    if (firstCard) firstCard.focus();
                    break;
                case 'End':
                    e.preventDefault();
                    const cards = Array.from(this.gallery.querySelectorAll('.theme-card:not([style*="display: none"])'));
                    const lastCard = cards[cards.length - 1];
                    if (lastCard) lastCard.focus();
                    break;
            }
        });
    }

    announceSelection(card) {
        if (!card) return;
        
        const themeName = card.querySelector('.theme-name')?.textContent;
        const announcement = document.createElement('div');
        announcement.setAttribute('role', 'status');
        announcement.setAttribute('aria-live', 'polite');
        announcement.className = 'sr-only';
        announcement.textContent = `${themeName} theme selected`;
        document.body.appendChild(announcement);
        setTimeout(() => announcement.remove(), 1000);
    }
}

if (document.readyState === 'loading') {
    document.addEventListener('DOMContentLoaded', () => {
        if (!window.themePicker) {
            window.themePicker = new ThemePicker();
        }
    });
} else {
    if (!window.themePicker) {
        window.themePicker = new ThemePicker();
    }
}
