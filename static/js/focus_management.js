const FocusManager = {
    previousFocus: null,
    activeTrap: null,
    focusableSelector: 'a[href]:not([disabled]), button:not([disabled]), textarea:not([disabled]), input:not([disabled]), select:not([disabled]), [tabindex]:not([tabindex="-1"])',

    getFocusableElements(container) {
        if (!container) return [];
        return Array.from(container.querySelectorAll(this.focusableSelector));
    },

    getFirstFocusable(container) {
        const elements = this.getFocusableElements(container);
        return elements.length > 0 ? elements[0] : null;
    },

    getLastFocusable(container) {
        const elements = this.getFocusableElements(container);
        return elements.length > 0 ? elements[elements.length - 1] : null;
    },

    saveFocus() {
        this.previousFocus = document.activeElement;
    },

    restoreFocus() {
        if (this.previousFocus && this.previousFocus.focus) {
            try {
                this.previousFocus.focus();
            } catch (e) {
            }
            this.previousFocus = null;
        }
    },

    moveFocusTo(element, options = {}) {
        if (!element || !element.focus) return false;

        try {
            element.focus();
            
            if (options.scroll !== false && element.scrollIntoView) {
                element.scrollIntoView({ 
                    behavior: options.smooth ? 'smooth' : 'auto', 
                    block: options.block || 'nearest',
                    inline: 'nearest'
                });
            }
            
            return true;
        } catch (e) {
            return false;
        }
    },

    trapFocus(container) {
        if (!container) return null;

        const focusableElements = this.getFocusableElements(container);
        if (focusableElements.length === 0) return null;

        const firstFocusable = focusableElements[0];
        const lastFocusable = focusableElements[focusableElements.length - 1];

        const handleKeyDown = (event) => {
            if (event.key !== 'Tab') return;

            if (event.shiftKey) {
                if (document.activeElement === firstFocusable) {
                    event.preventDefault();
                    lastFocusable.focus();
                }
            } else {
                if (document.activeElement === lastFocusable) {
                    event.preventDefault();
                    firstFocusable.focus();
                }
            }
        };

        container.addEventListener('keydown', handleKeyDown);

        this.activeTrap = {
            container: container,
            handler: handleKeyDown
        };

        return () => this.releaseFocusTrap();
    },

    releaseFocusTrap() {
        if (this.activeTrap) {
            this.activeTrap.container.removeEventListener('keydown', this.activeTrap.handler);
            this.activeTrap = null;
        }
    },

    focusFirstInContainer(container) {
        const firstFocusable = this.getFirstFocusable(container);
        if (firstFocusable) {
            this.moveFocusTo(firstFocusable);
            return true;
        }
        return false;
    },

    focusLastInContainer(container) {
        const lastFocusable = this.getLastFocusable(container);
        if (lastFocusable) {
            this.moveFocusTo(lastFocusable);
            return true;
        }
        return false;
    },

    manageFocusForModal(modal, options = {}) {
        if (!modal) return null;

        this.saveFocus();

        const cleanup = this.trapFocus(modal);
        
        if (options.focusFirst !== false) {
            this.focusFirstInContainer(modal);
        }

        return () => {
            if (cleanup) cleanup();
            if (options.restoreFocus !== false) {
                this.restoreFocus();
            }
        };
    },

    ensureFocusVisible() {
        const style = document.createElement('style');
        style.textContent = `
            *:focus:not(:focus-visible) { outline: none; }
            *:focus-visible { outline: 2px solid var(--color-border-focus, #3b82f6); outline-offset: 2px; }
        `;
        document.head.appendChild(style);
    },

    init() {
        if (typeof window !== 'undefined' && !window.CSS?.supports?.('selector(:focus-visible)')) {
            this.ensureFocusVisible();
        }
    }
};

if (typeof module !== 'undefined' && module.exports) {
    module.exports = FocusManager;
}
