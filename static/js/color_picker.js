class ColorPicker {
    constructor() {
        this.colorInput = null;
        this.hexInput = null;
        this.hiddenInput = null;
        this.preview = null;
        this.resetBtn = null;
        this.defaultColor = '#007bff';
    }

    init() {
        this.colorInput = document.getElementById('custom-theme-color');
        this.hexInput = document.getElementById('custom-theme-color-hex');
        this.hiddenInput = document.getElementById('custom-theme-color-value');
        this.preview = document.querySelector('.color-preview');
        this.resetBtn = document.getElementById('reset-color-btn');

        if (!this.colorInput || !this.hexInput || !this.hiddenInput) {
            return;
        }

        this.setupEventListeners();
        this.syncInitialValues();
    }

    setupEventListeners() {
        this.colorInput.addEventListener('input', (e) => {
            this.handleColorInputChange(e.target.value);
        });

        this.colorInput.addEventListener('change', (e) => {
            this.handleColorInputChange(e.target.value);
        });

        this.hexInput.addEventListener('input', (e) => {
            this.handleHexInputChange(e.target.value);
        });

        this.hexInput.addEventListener('keydown', (e) => {
            if (e.key === 'Enter') {
                e.preventDefault();
                this.handleHexInputChange(e.target.value);
            }
        });

        this.hexInput.addEventListener('blur', (e) => {
            this.handleHexInputChange(e.target.value);
        });

        if (this.resetBtn) {
            this.resetBtn.addEventListener('click', () => {
                this.reset();
            });
        }
    }

    syncInitialValues() {
        const initialColor = this.colorInput.value || this.defaultColor;
        this.updateAllInputs(initialColor);
    }

    handleColorInputChange(color) {
        if (this.isValidHex(color)) {
            this.updateAllInputs(color);
            this.clearError();
        }
    }

    handleHexInputChange(value) {
        let hex = value.trim().toUpperCase();
        
        if (!hex.startsWith('#')) {
            hex = '#' + hex;
        }

        if (this.isValidHex(hex)) {
            this.updateAllInputs(hex);
            this.clearError();
        } else {
            this.showError();
        }
    }

    updateAllInputs(color) {
        const upperColor = color.toUpperCase();
        
        if (this.colorInput.value !== color) {
            this.colorInput.value = color;
        }
        
        if (this.hexInput.value !== upperColor) {
            this.hexInput.value = upperColor;
        }
        
        if (this.hiddenInput.value !== color) {
            this.hiddenInput.value = color;
        }
        
        this.updatePreview(color);
        this.triggerPreviewRefresh();
    }

    updatePreview(color) {
        if (this.preview) {
            this.preview.style.backgroundColor = color;
            this.preview.setAttribute('aria-label', `Current color preview: ${color}`);
        }
    }

    isValidHex(color) {
        const hexPattern = /^#[0-9A-Fa-f]{6}$/;
        return hexPattern.test(color);
    }

    showError() {
        if (this.hexInput) {
            this.hexInput.classList.add('error');
            this.hexInput.setAttribute('aria-invalid', 'true');
        }
    }

    clearError() {
        if (this.hexInput) {
            this.hexInput.classList.remove('error');
            this.hexInput.setAttribute('aria-invalid', 'false');
        }
    }

    reset() {
        this.updateAllInputs(this.defaultColor);
        this.clearError();
    }

    triggerPreviewRefresh() {
        const event = new CustomEvent('colorChanged', {
            detail: { color: this.hiddenInput.value },
            bubbles: true
        });
        document.dispatchEvent(event);
    }

    getCurrentColor() {
        return this.hiddenInput ? this.hiddenInput.value : this.defaultColor;
    }
}

if (typeof window !== 'undefined') {
    window.ColorPicker = ColorPicker;

    document.addEventListener('DOMContentLoaded', () => {
        const colorPickerSection = document.querySelector('.color-picker-section');
        if (colorPickerSection) {
            const colorPicker = new ColorPicker();
            colorPicker.init();
            window.colorPickerInstance = colorPicker;
        }
    });
}
