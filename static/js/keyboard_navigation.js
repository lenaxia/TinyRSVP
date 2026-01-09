const KeyboardNav = {
    focusableSelector: 'a[href], button:not([disabled]), textarea:not([disabled]), input:not([disabled]), select:not([disabled]), [tabindex]:not([tabindex="-1"])',

    getFocusableElements(container) {
        return Array.from(container.querySelectorAll(this.focusableSelector));
    },

    handleEscape(event, callback) {
        if (event.key === 'Escape') {
            event.preventDefault();
            if (callback && typeof callback === 'function') {
                callback(event);
            }
        }
    },

    handleTab(event, container) {
        if (event.key !== 'Tab') return;

        const focusableElements = this.getFocusableElements(container);
        if (focusableElements.length === 0) return;

        const firstElement = focusableElements[0];
        const lastElement = focusableElements[focusableElements.length - 1];

        if (event.shiftKey && document.activeElement === firstElement) {
            event.preventDefault();
            lastElement.focus();
        } else if (!event.shiftKey && document.activeElement === lastElement) {
            event.preventDefault();
            firstElement.focus();
        }
    },

    handleArrowKeys(event, items, currentIndex) {
        const key = event.key;
        
        if (!['ArrowUp', 'ArrowDown', 'ArrowLeft', 'ArrowRight'].includes(key)) {
            return currentIndex;
        }

        event.preventDefault();
        
        let newIndex = currentIndex;
        
        if (key === 'ArrowDown' || key === 'ArrowRight') {
            newIndex = (currentIndex + 1) % items.length;
        } else if (key === 'ArrowUp' || key === 'ArrowLeft') {
            newIndex = (currentIndex - 1 + items.length) % items.length;
        }

        if (items[newIndex] && items[newIndex].focus) {
            items[newIndex].focus();
        }

        return newIndex;
    },

    handleEnterSpace(event, callback) {
        if (event.key === 'Enter' || event.key === ' ') {
            event.preventDefault();
            if (callback && typeof callback === 'function') {
                callback(event);
            }
        }
    },

    trapFocus(element) {
        if (!element) return null;

        const handleKeyDown = (event) => {
            this.handleTab(event, element);
        };

        element.addEventListener('keydown', handleKeyDown);

        return () => {
            element.removeEventListener('keydown', handleKeyDown);
        };
    },

    moveFocusTo(element) {
        if (element && element.focus) {
            element.focus();
            
            if (element.scrollIntoView) {
                element.scrollIntoView({ behavior: 'smooth', block: 'nearest' });
            }
        }
    },

    saveFocus() {
        this.previousFocus = document.activeElement;
    },

    restoreFocus() {
        if (this.previousFocus && this.previousFocus.focus) {
            this.previousFocus.focus();
            this.previousFocus = null;
        }
    },

    initModalKeyboardHandling(modal, closeCallback) {
        if (!modal) return null;

        this.saveFocus();

        const handleKeyDown = (event) => {
            this.handleEscape(event, () => {
                if (closeCallback && typeof closeCallback === 'function') {
                    closeCallback();
                }
                this.restoreFocus();
            });
            this.handleTab(event, modal);
        };

        modal.addEventListener('keydown', handleKeyDown);

        const firstFocusable = this.getFocusableElements(modal)[0];
        if (firstFocusable) {
            firstFocusable.focus();
        }

        return () => {
            modal.removeEventListener('keydown', handleKeyDown);
            this.restoreFocus();
        };
    },

    initDropdownKeyboardHandling(trigger, dropdown, items) {
        if (!trigger || !dropdown || !items || items.length === 0) return null;

        let currentIndex = -1;

        const handleTriggerKeyDown = (event) => {
            if (event.key === 'Enter' || event.key === ' ') {
                event.preventDefault();
                trigger.click();
            } else if (event.key === 'ArrowDown') {
                event.preventDefault();
                trigger.click();
                currentIndex = 0;
                if (items[0]) items[0].focus();
            }
        };

        const handleDropdownKeyDown = (event) => {
            if (event.key === 'Escape') {
                event.preventDefault();
                dropdown.style.display = 'none';
                trigger.focus();
                currentIndex = -1;
            } else if (event.key === 'ArrowDown' || event.key === 'ArrowUp' || 
                       event.key === 'ArrowLeft' || event.key === 'ArrowRight') {
                currentIndex = this.handleArrowKeys(event, items, currentIndex);
            } else if (event.key === 'Enter' || event.key === ' ') {
                event.preventDefault();
                if (items[currentIndex]) {
                    items[currentIndex].click();
                }
            }
        };

        trigger.addEventListener('keydown', handleTriggerKeyDown);
        dropdown.addEventListener('keydown', handleDropdownKeyDown);

        return () => {
            trigger.removeEventListener('keydown', handleTriggerKeyDown);
            dropdown.removeEventListener('keydown', handleDropdownKeyDown);
        };
    },

    init() {
        const modals = document.querySelectorAll('[role="dialog"], .modal');
        modals.forEach(modal => {
            const closeButton = modal.querySelector('[data-dismiss="modal"], .modal-close');
            if (closeButton) {
                this.initModalKeyboardHandling(modal, () => {
                    modal.style.display = 'none';
                    modal.setAttribute('aria-hidden', 'true');
                });
            }
        });

        const dropdowns = document.querySelectorAll('.dropdown');
        dropdowns.forEach(dropdown => {
            const trigger = dropdown.querySelector('.dropdown-trigger');
            const menu = dropdown.querySelector('.dropdown-menu');
            const items = menu ? Array.from(menu.querySelectorAll('a, button')) : [];
            
            if (trigger && menu && items.length > 0) {
                this.initDropdownKeyboardHandling(trigger, menu, items);
            }
        });

        const customButtons = document.querySelectorAll('[role="button"]:not(button)');
        customButtons.forEach(button => {
            button.addEventListener('keydown', (event) => {
                this.handleEnterSpace(event, () => {
                    button.click();
                });
            });
        });
    }
};

if (typeof document !== 'undefined') {
    document.addEventListener('DOMContentLoaded', () => {
        KeyboardNav.init();
    });
}

if (typeof module !== 'undefined' && module.exports) {
    module.exports = KeyboardNav;
}
