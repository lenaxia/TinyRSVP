(function() {
    'use strict';

    class SlidePanel {
        constructor(panelSelector, options = {}) {
            this.panelSelector = panelSelector;
            this.options = {
                onOpen: options.onOpen || null,
                onClose: options.onClose || null,
                onSave: options.onSave || null,
                onCancel: options.onCancel || null,
                ...options
            };

            this.panel = document.querySelector(panelSelector);
            this.overlay = document.querySelector(`${panelSelector.replace('-panel', '-overlay')}`);
            
            const baseClass = panelSelector.replace('.', '').replace('-panel', '');
            this.closeBtn = this.panel?.querySelector(`.${baseClass}-close`);
            this.cancelBtn = this.panel?.querySelector(`.${baseClass}-cancel`);
            this.saveBtn = this.panel?.querySelector(`.${baseClass}-save`);

            if (!this.panel || !this.overlay) {
                console.error(`SlidePanel: Could not find panel or overlay for selector ${panelSelector}`);
                return;
            }

            this.init();
        }

        init() {
            this.attachEventListeners();
        }

        attachEventListeners() {
            if (this.closeBtn) {
                this.closeBtn.addEventListener('click', () => this.close());
            }

            if (this.overlay) {
                this.overlay.addEventListener('click', () => this.close());
            }

            if (this.cancelBtn) {
                this.cancelBtn.addEventListener('click', () => this.cancel());
            }

            if (this.saveBtn) {
                this.saveBtn.addEventListener('click', () => this.save());
            }

            document.addEventListener('keydown', (e) => {
                if (e.key === 'Escape' && this.isOpen()) {
                    this.close();
                }
            });
        }

        open() {
            if (this.options.onOpen) {
                this.options.onOpen();
            }

            this.panel.classList.add('open');
            this.overlay.classList.add('open');
            document.body.style.overflow = 'hidden';
        }

        close() {
            this.panel.classList.remove('open');
            this.overlay.classList.remove('open');
            document.body.style.overflow = '';

            if (this.options.onClose) {
                this.options.onClose();
            }
        }

        cancel() {
            if (this.options.onCancel) {
                this.options.onCancel();
            }
            this.close();
        }

        save() {
            if (this.options.onSave) {
                this.options.onSave();
            }
            this.close();
        }

        isOpen() {
            return this.panel.classList.contains('open');
        }
    }

    window.SlidePanel = SlidePanel;
})();
