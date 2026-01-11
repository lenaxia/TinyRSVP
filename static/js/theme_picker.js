class ThemePicker {
    constructor() {
        this.gallery = document.querySelector('.theme-gallery');
        this.filterSelect = document.getElementById('theme-category-filter');
        this.hiddenInput = document.getElementById('selected-theme-id');
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
    document.addEventListener('DOMContentLoaded', () => new ThemePicker());
} else {
    new ThemePicker();
}
