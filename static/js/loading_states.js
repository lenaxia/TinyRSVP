const LoadingStates = (() => {
    const loadingStates = new Map();
    const originalButtonTexts = new Map();

    function showButtonLoading(button, options = {}) {
        if (!button) return;

        const buttonElement = typeof button === 'string' ? document.querySelector(button) : button;
        if (!buttonElement) return;

        const buttonId = buttonElement.id || `btn-${Date.now()}`;
        if (!buttonElement.id) buttonElement.id = buttonId;

        if (loadingStates.has(buttonId)) return;

        originalButtonTexts.set(buttonId, buttonElement.textContent);

        buttonElement.classList.add('loading');
        buttonElement.disabled = true;
        buttonElement.setAttribute('aria-busy', 'true');

        loadingStates.set(buttonId, {
            element: buttonElement,
            type: 'button',
            startTime: Date.now()
        });

        if (options.timeout) {
            setTimeout(() => {
                hideButtonLoading(buttonElement);
            }, options.timeout);
        }
    }

    function hideButtonLoading(button) {
        if (!button) return;

        const buttonElement = typeof button === 'string' ? document.querySelector(button) : button;
        if (!buttonElement) return;

        const buttonId = buttonElement.id;
        if (!buttonId || !loadingStates.has(buttonId)) return;

        buttonElement.classList.remove('loading');
        buttonElement.disabled = false;
        buttonElement.removeAttribute('aria-busy');

        if (originalButtonTexts.has(buttonId)) {
            buttonElement.textContent = originalButtonTexts.get(buttonId);
            originalButtonTexts.delete(buttonId);
        }

        loadingStates.delete(buttonId);
    }

    function showSpinner(container, options = {}) {
        if (!container) return null;

        const containerElement = typeof container === 'string' ? document.querySelector(container) : container;
        if (!containerElement) return null;

        const spinner = document.createElement('div');
        spinner.className = options.size ? `spinner spinner-${options.size}` : 'spinner';
        spinner.setAttribute('role', 'status');
        spinner.setAttribute('aria-live', 'polite');
        spinner.setAttribute('aria-label', options.label || 'Loading');

        if (options.inline) {
            spinner.classList.add('spinner-inline');
        }

        containerElement.appendChild(spinner);

        const spinnerId = `spinner-${Date.now()}`;
        loadingStates.set(spinnerId, {
            element: spinner,
            container: containerElement,
            type: 'spinner'
        });

        return spinner;
    }

    function hideSpinner(spinner) {
        if (!spinner) return;

        const spinnerElement = typeof spinner === 'string' ? document.querySelector(spinner) : spinner;
        if (!spinnerElement) return;

        spinnerElement.remove();

        for (const [key, value] of loadingStates.entries()) {
            if (value.element === spinnerElement) {
                loadingStates.delete(key);
                break;
            }
        }
    }

    function showOverlay(options = {}) {
        let overlay = document.querySelector('.loading-overlay');
        
        if (overlay) return overlay;

        overlay = document.createElement('div');
        overlay.className = 'loading-overlay';
        overlay.setAttribute('role', 'status');
        overlay.setAttribute('aria-live', 'polite');
        overlay.setAttribute('aria-label', options.label || 'Loading');

        if (options.dark) {
            overlay.classList.add('loading-overlay-dark');
        }

        const spinner = document.createElement('div');
        spinner.className = 'spinner spinner-lg';
        overlay.appendChild(spinner);

        document.body.appendChild(overlay);
        document.body.style.overflow = 'hidden';

        loadingStates.set('overlay', {
            element: overlay,
            type: 'overlay'
        });

        if (options.timeout) {
            setTimeout(() => {
                hideOverlay();
            }, options.timeout);
        }

        return overlay;
    }

    function hideOverlay() {
        const overlay = document.querySelector('.loading-overlay');
        if (!overlay) return;

        overlay.remove();
        document.body.style.overflow = '';
        loadingStates.delete('overlay');
    }

    function updateProgress(progressBar, percentage) {
        if (!progressBar) return;

        const progressElement = typeof progressBar === 'string' ? document.querySelector(progressBar) : progressBar;
        if (!progressElement) return;

        const progressBarFill = progressElement.querySelector('.progress-bar');
        if (!progressBarFill) return;

        const clampedPercentage = Math.max(0, Math.min(100, percentage));
        progressBarFill.style.width = `${clampedPercentage}%`;
        progressBarFill.setAttribute('aria-valuenow', clampedPercentage);

        if (clampedPercentage >= 100) {
            progressBarFill.setAttribute('aria-label', 'Loading complete');
        }
    }

    function setLoadingState(element, loading = true) {
        if (!element) return;

        const targetElement = typeof element === 'string' ? document.querySelector(element) : element;
        if (!targetElement) return;

        if (loading) {
            targetElement.setAttribute('aria-busy', 'true');
            
            const interactiveElements = targetElement.querySelectorAll('button, input, select, textarea, a');
            interactiveElements.forEach(el => {
                if (!el.disabled) {
                    el.disabled = true;
                    el.dataset.wasEnabled = 'true';
                }
            });

            const elementId = targetElement.id || `element-${Date.now()}`;
            if (!targetElement.id) targetElement.id = elementId;

            loadingStates.set(elementId, {
                element: targetElement,
                type: 'element'
            });
        } else {
            clearLoadingState(targetElement);
        }
    }

    function clearLoadingState(element) {
        if (!element) return;

        const targetElement = typeof element === 'string' ? document.querySelector(element) : element;
        if (!targetElement) return;

        targetElement.removeAttribute('aria-busy');

        const interactiveElements = targetElement.querySelectorAll('[data-was-enabled]');
        interactiveElements.forEach(el => {
            el.disabled = false;
            delete el.dataset.wasEnabled;
        });

        if (targetElement.id) {
            loadingStates.delete(targetElement.id);
        }
    }

    function clearAll() {
        for (const [key, state] of loadingStates.entries()) {
            if (state.type === 'button') {
                hideButtonLoading(state.element);
            } else if (state.type === 'spinner') {
                hideSpinner(state.element);
            } else if (state.type === 'overlay') {
                hideOverlay();
            } else if (state.type === 'element') {
                clearLoadingState(state.element);
            }
        }
        loadingStates.clear();
        originalButtonTexts.clear();
    }

    function getActiveStates() {
        return Array.from(loadingStates.entries()).map(([id, state]) => ({
            id,
            type: state.type,
            duration: Date.now() - (state.startTime || 0)
        }));
    }

    return {
        showButtonLoading,
        hideButtonLoading,
        showSpinner,
        hideSpinner,
        showOverlay,
        hideOverlay,
        updateProgress,
        setLoadingState,
        clearLoadingState,
        clearAll,
        getActiveStates
    };
})();

if (typeof module !== 'undefined' && module.exports) {
    module.exports = LoadingStates;
}
