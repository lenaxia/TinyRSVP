const ScreenReader = {
    liveRegions: {
        polite: null,
        assertive: null
    },

    createLiveRegion(type) {
        const region = document.createElement('div');
        region.setAttribute('role', type === 'assertive' ? 'alert' : 'status');
        region.setAttribute('aria-live', type);
        region.setAttribute('aria-atomic', 'true');
        region.className = 'sr-only';
        region.style.position = 'absolute';
        region.style.left = '-10000px';
        region.style.width = '1px';
        region.style.height = '1px';
        region.style.overflow = 'hidden';
        document.body.appendChild(region);
        return region;
    },

    announce(message, priority = 'polite') {
        if (!message) return;

        const type = priority === 'assertive' ? 'assertive' : 'polite';
        
        if (!this.liveRegions[type]) {
            this.liveRegions[type] = this.createLiveRegion(type);
        }

        const region = this.liveRegions[type];
        region.textContent = '';
        
        setTimeout(() => {
            region.textContent = message;
        }, 100);
    },

    setLiveRegion(element, type = 'polite', atomic = true) {
        if (!element) return;
        
        const validTypes = ['polite', 'assertive', 'off'];
        const liveType = validTypes.includes(type) ? type : 'polite';
        
        element.setAttribute('aria-live', liveType);
        if (atomic) {
            element.setAttribute('aria-atomic', 'true');
        }
        
        if (liveType === 'assertive') {
            element.setAttribute('role', 'alert');
        } else if (liveType === 'polite') {
            element.setAttribute('role', 'status');
        } else if (liveType === 'off') {
            element.setAttribute('role', 'region');
        } else {
            element.setAttribute('role', 'region');
        }
    },

    addLandmark(element, role, label) {
        if (!element) return;
        
        const validRoles = ['banner', 'navigation', 'main', 'contentinfo', 'complementary', 'region', 'search', 'form'];
        if (!validRoles.includes(role)) {
            return;
        }

        element.setAttribute('role', role);
        
        if (label) {
            element.setAttribute('aria-label', label);
        }
    },

    addLabel(element, label, labelledBy) {
        if (!element) return;

        if (labelledBy) {
            element.setAttribute('aria-labelledby', labelledBy);
        } else if (label) {
            element.setAttribute('aria-label', label);
        }
    },

    addDescription(element, description, describedBy) {
        if (!element) return;

        if (describedBy) {
            element.setAttribute('aria-describedby', describedBy);
        } else if (description) {
            const descId = element.id ? `${element.id}-desc` : `desc-${Date.now()}`;
            const descElement = document.createElement('span');
            descElement.id = descId;
            descElement.className = 'sr-only';
            descElement.textContent = description;
            descElement.style.position = 'absolute';
            descElement.style.left = '-10000px';
            descElement.style.width = '1px';
            descElement.style.height = '1px';
            descElement.style.overflow = 'hidden';
            
            element.parentNode.insertBefore(descElement, element.nextSibling);
            element.setAttribute('aria-describedby', descId);
        }
    },

    hideFromScreenReader(element) {
        if (!element) return;
        element.setAttribute('aria-hidden', 'true');
    },

    showToScreenReader(element) {
        if (!element) return;
        element.removeAttribute('aria-hidden');
    },

    setExpanded(element, expanded) {
        if (!element) return;
        element.setAttribute('aria-expanded', expanded ? 'true' : 'false');
    },

    setPressed(element, pressed) {
        if (!element) return;
        element.setAttribute('aria-pressed', pressed ? 'true' : 'false');
    },

    setChecked(element, checked) {
        if (!element) return;
        element.setAttribute('aria-checked', checked ? 'true' : 'false');
    },

    setCurrent(element, current) {
        if (!element) return;
        
        const validValues = ['page', 'step', 'location', 'date', 'time', 'true', 'false'];
        if (validValues.includes(current)) {
            element.setAttribute('aria-current', current);
        }
    },

    addImageAlt(img, altText) {
        if (!img || img.tagName !== 'IMG') return;
        
        if (altText) {
            img.setAttribute('alt', altText);
        } else if (!img.hasAttribute('alt')) {
            img.setAttribute('alt', '');
        }
    },

    ensureFormLabels() {
        const inputs = document.querySelectorAll('input:not([type="hidden"]), select, textarea');
        
        inputs.forEach(input => {
            if (!input.id) {
                input.id = `input-${Date.now()}-${Math.random().toString(36).substr(2, 9)}`;
            }

            const label = document.querySelector(`label[for="${input.id}"]`);
            if (!label && !input.getAttribute('aria-label') && !input.getAttribute('aria-labelledby')) {
                const placeholder = input.getAttribute('placeholder');
                if (placeholder) {
                    input.setAttribute('aria-label', placeholder);
                }
            }
        });
    },

    ensureButtonLabels() {
        const buttons = document.querySelectorAll('button, [role="button"]');
        
        buttons.forEach(button => {
            const hasText = button.textContent.trim().length > 0;
            const hasLabel = button.getAttribute('aria-label');
            const hasLabelledBy = button.getAttribute('aria-labelledby');
            
            if (!hasText && !hasLabel && !hasLabelledBy) {
                const title = button.getAttribute('title');
                if (title) {
                    button.setAttribute('aria-label', title);
                }
            }
        });
    },

    ensureLinkPurpose() {
        const links = document.querySelectorAll('a[href]');
        
        links.forEach(link => {
            const hasText = link.textContent.trim().length > 0;
            const hasLabel = link.getAttribute('aria-label');
            const hasLabelledBy = link.getAttribute('aria-labelledby');
            
            if (!hasText && !hasLabel && !hasLabelledBy) {
                const title = link.getAttribute('title');
                if (title) {
                    link.setAttribute('aria-label', title);
                }
            }
        });
    },

    init() {
        this.ensureFormLabels();
        this.ensureButtonLabels();
        this.ensureLinkPurpose();

        const header = document.querySelector('header');
        if (header && !header.getAttribute('role')) {
            this.addLandmark(header, 'banner');
        }

        const nav = document.querySelector('nav');
        if (nav && !nav.getAttribute('role')) {
            this.addLandmark(nav, 'navigation', 'Main navigation');
        }

        const main = document.querySelector('main');
        if (main && !main.getAttribute('role')) {
            this.addLandmark(main, 'main');
        }

        const footer = document.querySelector('footer');
        if (footer && !footer.getAttribute('role')) {
            this.addLandmark(footer, 'contentinfo');
        }

        const images = document.querySelectorAll('img:not([alt])');
        images.forEach(img => {
            this.addImageAlt(img, '');
        });
    }
};

if (typeof document !== 'undefined') {
    document.addEventListener('DOMContentLoaded', () => {
        ScreenReader.init();
    });
}

if (typeof module !== 'undefined' && module.exports) {
    module.exports = ScreenReader;
}
