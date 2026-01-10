(function() {
    'use strict';

    class Modal {
        constructor(modalId, options = {}) {
            this.modal = document.getElementById(modalId);
            if (!this.modal) {
                console.error(`Modal with id "${modalId}" not found`);
                return;
            }

            this.options = {
                closeOnOverlayClick: options.closeOnOverlayClick !== false,
                closeOnEscape: options.closeOnEscape !== false,
                preventBodyScroll: options.preventBodyScroll !== false,
                onOpen: options.onOpen || null,
                onClose: options.onClose || null,
                ...options
            };

            this.overlay = this.modal.querySelector('.modal-overlay') || this.createOverlay();
            this.closeButtons = this.modal.querySelectorAll('[data-modal-close]');
            this.isOpen = false;

            this.init();
        }

        init() {
            this.closeButtons.forEach(btn => {
                btn.addEventListener('click', () => this.close());
            });

            if (this.options.closeOnOverlayClick) {
                this.overlay.addEventListener('click', (e) => {
                    if (e.target === this.overlay) {
                        this.close();
                    }
                });
            }

            if (this.options.closeOnEscape) {
                document.addEventListener('keydown', (e) => {
                    if (e.key === 'Escape' && this.isOpen) {
                        this.close();
                    }
                });
            }
        }

        createOverlay() {
            const overlay = document.createElement('div');
            overlay.className = 'modal-overlay';
            overlay.setAttribute('aria-hidden', 'true');
            this.modal.insertBefore(overlay, this.modal.firstChild);
            return overlay;
        }

        open() {
            if (this.isOpen) return;

            this.isOpen = true;

            if (this.options.preventBodyScroll) {
                document.body.classList.add('modal-open');
            }

            this.overlay.classList.add('active');
            this.overlay.setAttribute('aria-hidden', 'false');

            const modalContent = this.modal.querySelector('.modal-center, .modal-slide');
            if (modalContent) {
                modalContent.classList.add('active');
                modalContent.setAttribute('aria-hidden', 'false');
                
                const firstFocusable = modalContent.querySelector(
                    'button, [href], input, select, textarea, [tabindex]:not([tabindex="-1"])'
                );
                if (firstFocusable) {
                    setTimeout(() => firstFocusable.focus(), 100);
                }
            }

            if (this.options.onOpen) {
                this.options.onOpen(this);
            }

            this.modal.dispatchEvent(new CustomEvent('modal:open', { detail: { modal: this } }));
        }

        close() {
            if (!this.isOpen) return;

            this.isOpen = false;

            const modalContent = this.modal.querySelector('.modal-center, .modal-slide');
            if (modalContent) {
                modalContent.classList.remove('active');
                modalContent.setAttribute('aria-hidden', 'true');
            }

            setTimeout(() => {
                this.overlay.classList.remove('active');
                this.overlay.setAttribute('aria-hidden', 'true');

                if (this.options.preventBodyScroll) {
                    document.body.classList.remove('modal-open');
                }
            }, 300);

            if (this.options.onClose) {
                this.options.onClose(this);
            }

            this.modal.dispatchEvent(new CustomEvent('modal:close', { detail: { modal: this } }));
        }

        toggle() {
            if (this.isOpen) {
                this.close();
            } else {
                this.open();
            }
        }

        destroy() {
            this.closeButtons.forEach(btn => {
                btn.removeEventListener('click', () => this.close());
            });
            this.close();
        }
    }

    function initModals() {
        const modalTriggers = document.querySelectorAll('[data-modal-trigger]');
        
        modalTriggers.forEach(trigger => {
            const modalId = trigger.getAttribute('data-modal-trigger');
            const modal = new Modal(modalId);

            trigger.addEventListener('click', (e) => {
                e.preventDefault();
                modal.open();
            });
        });
    }

    if (document.readyState === 'loading') {
        document.addEventListener('DOMContentLoaded', initModals);
    } else {
        initModals();
    }

    window.Modal = Modal;
})();
