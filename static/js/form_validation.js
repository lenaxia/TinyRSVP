const FormValidator = {
    validateEmail(email) {
        const re = /^[^\s@]+@[^\s@]+\.[^\s@]+$/;
        return re.test(String(email).toLowerCase());
    },

    validateRequired(value) {
        if (typeof value === 'string') {
            return value.trim().length > 0;
        }
        return value !== null && value !== undefined && value !== '';
    },

    validateDateTime(value) {
        if (!value) return false;
        const date = new Date(value);
        return date instanceof Date && !isNaN(date);
    },

    validateDateTimeRange(startValue, endValue) {
        if (!startValue || !endValue) return true;
        const start = new Date(startValue);
        const end = new Date(endValue);
        return end > start;
    },

    validateNumber(value, min, max) {
        const num = parseFloat(value);
        if (isNaN(num)) return false;
        if (min !== undefined && num < min) return false;
        if (max !== undefined && num > max) return false;
        return true;
    },

    showError(field, message) {
        this.clearSuccess(field);
        field.classList.add('error');
        field.classList.remove('success');
        field.setAttribute('aria-invalid', 'true');

        let errorEl = field.parentElement.querySelector('.form-error');
        if (!errorEl) {
            errorEl = document.createElement('span');
            errorEl.className = 'form-error';
            errorEl.setAttribute('role', 'alert');
            
            const helpText = field.parentElement.querySelector('.form-help-text');
            if (helpText) {
                helpText.parentNode.insertBefore(errorEl, helpText.nextSibling);
            } else {
                field.parentNode.appendChild(errorEl);
            }
        }
        errorEl.textContent = message;
        
        const errorId = field.id + '-error';
        errorEl.id = errorId;
        
        const describedBy = field.getAttribute('aria-describedby') || '';
        if (!describedBy.includes(errorId)) {
            field.setAttribute('aria-describedby', describedBy ? `${describedBy} ${errorId}` : errorId);
        }
    },

    clearError(field) {
        field.classList.remove('error');
        field.removeAttribute('aria-invalid');
        
        const errorEl = field.parentElement.querySelector('.form-error');
        if (errorEl) {
            errorEl.textContent = '';
            errorEl.remove();
        }
    },

    showSuccess(field) {
        this.clearError(field);
        field.classList.add('success');
        field.classList.remove('error');
        field.removeAttribute('aria-invalid');
    },

    clearSuccess(field) {
        field.classList.remove('success');
    },

    getCustomErrorMessage(field, validationType) {
        const customMessage = field.getAttribute('data-error-message');
        if (customMessage) return customMessage;

        const fieldName = field.getAttribute('aria-label') || 
                         field.previousElementSibling?.textContent?.replace('*', '').trim() || 
                         field.name || 
                         'This field';

        const messages = {
            required: `${fieldName} is required`,
            email: 'Please enter a valid email address',
            datetime: 'Please enter a valid date and time',
            number: 'Please enter a valid number',
            min: `Value must be at least ${field.min}`,
            max: `Value must be no more than ${field.max}`,
            daterange: 'End date must be after start date'
        };

        return messages[validationType] || `${fieldName} is invalid`;
    },

    validateField(field) {
        if (field.disabled || field.readOnly) {
            return true;
        }

        const value = field.value;
        const type = field.type;
        const isRequired = field.hasAttribute('required');

        if (isRequired && !this.validateRequired(value)) {
            this.showError(field, this.getCustomErrorMessage(field, 'required'));
            return false;
        }

        if (!value) {
            this.clearError(field);
            this.clearSuccess(field);
            return true;
        }

        if (type === 'email' && !this.validateEmail(value)) {
            this.showError(field, this.getCustomErrorMessage(field, 'email'));
            return false;
        }

        if (type === 'datetime-local' && !this.validateDateTime(value)) {
            this.showError(field, this.getCustomErrorMessage(field, 'datetime'));
            return false;
        }

        if (type === 'number') {
            const min = field.hasAttribute('min') ? parseFloat(field.min) : undefined;
            const max = field.hasAttribute('max') ? parseFloat(field.max) : undefined;
            
            if (!this.validateNumber(value, min, max)) {
                if (min !== undefined && parseFloat(value) < min) {
                    this.showError(field, this.getCustomErrorMessage(field, 'min'));
                } else if (max !== undefined && parseFloat(value) > max) {
                    this.showError(field, this.getCustomErrorMessage(field, 'max'));
                } else {
                    this.showError(field, this.getCustomErrorMessage(field, 'number'));
                }
                return false;
            }
        }

        if (field.hasAttribute('maxlength')) {
            const maxLength = parseInt(field.getAttribute('maxlength'));
            if (value.length > maxLength) {
                this.showError(field, `Maximum ${maxLength} characters allowed`);
                return false;
            }
        }

        this.showSuccess(field);
        return true;
    },

    validateDateRange(form) {
        const startField = form.querySelector('#start_time');
        const endField = form.querySelector('#end_time');

        if (!startField || !endField) return true;
        if (!startField.value || !endField.value) return true;

        if (!this.validateDateTimeRange(startField.value, endField.value)) {
            this.showError(endField, this.getCustomErrorMessage(endField, 'daterange'));
            return false;
        }

        return true;
    },

    validateRSVPDeadline(form) {
        const startField = form.querySelector('#start_time');
        const deadlineField = form.querySelector('#rsvp_deadline');

        if (!startField || !deadlineField) return true;
        if (!startField.value || !deadlineField.value) return true;

        if (!this.validateDateTimeRange(deadlineField.value, startField.value)) {
            this.showError(deadlineField, 'RSVP deadline must be before event start time');
            return false;
        }

        return true;
    },

    validateForm(form) {
        let isValid = true;
        const fields = form.querySelectorAll('input:not([type="hidden"]):not([type="radio"]):not([type="checkbox"]), select, textarea');

        fields.forEach(field => {
            if (!this.validateField(field)) {
                isValid = false;
            }
        });

        const radioGroups = {};
        form.querySelectorAll('input[type="radio"][required]').forEach(radio => {
            if (!radioGroups[radio.name]) {
                radioGroups[radio.name] = [];
            }
            radioGroups[radio.name].push(radio);
        });

        Object.keys(radioGroups).forEach(groupName => {
            const radios = radioGroups[groupName];
            const isChecked = radios.some(radio => radio.checked);
            
            if (!isChecked) {
                const firstRadio = radios[0];
                this.showError(firstRadio, this.getCustomErrorMessage(firstRadio, 'required'));
                isValid = false;
            } else {
                radios.forEach(radio => this.clearError(radio));
            }
        });

        if (!this.validateDateRange(form)) {
            isValid = false;
        }

        if (!this.validateRSVPDeadline(form)) {
            isValid = false;
        }

        return isValid;
    },

    init(formSelector = 'form[novalidate]') {
        const forms = document.querySelectorAll(formSelector);

        forms.forEach(form => {
            const fields = form.querySelectorAll('input:not([type="hidden"]):not([type="submit"]):not([type="button"]), select, textarea');

            fields.forEach(field => {
                field.addEventListener('blur', () => {
                    this.validateField(field);
                });

                field.addEventListener('input', () => {
                    if (field.classList.contains('error')) {
                        this.validateField(field);
                    }
                });
            });

            const endTimeField = form.querySelector('#end_time');
            if (endTimeField) {
                const startTimeField = form.querySelector('#start_time');
                if (startTimeField) {
                    endTimeField.addEventListener('blur', () => {
                        this.validateDateRange(form);
                    });
                    startTimeField.addEventListener('blur', () => {
                        if (endTimeField.value) {
                            this.validateDateRange(form);
                        }
                    });
                }
            }

            const rsvpDeadlineField = form.querySelector('#rsvp_deadline');
            if (rsvpDeadlineField) {
                const startTimeField = form.querySelector('#start_time');
                if (startTimeField) {
                    rsvpDeadlineField.addEventListener('blur', () => {
                        this.validateRSVPDeadline(form);
                    });
                    startTimeField.addEventListener('blur', () => {
                        if (rsvpDeadlineField.value) {
                            this.validateRSVPDeadline(form);
                        }
                    });
                }
            }

            form.addEventListener('submit', (e) => {
                if (!this.validateForm(form)) {
                    e.preventDefault();
                    
                    const firstError = form.querySelector('.error');
                    if (firstError) {
                        firstError.focus();
                        firstError.scrollIntoView({ behavior: 'smooth', block: 'center' });
                    }
                }
            });
        });
    }
};

if (typeof document !== 'undefined') {
    document.addEventListener('DOMContentLoaded', () => {
        FormValidator.init();
    });
}

if (typeof module !== 'undefined' && module.exports) {
    module.exports = FormValidator;
}
