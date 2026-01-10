const Counter = {
    init(selector = '.counter') {
        const counters = document.querySelectorAll(selector);
        
        counters.forEach(counter => {
            this.setupCounter(counter);
        });
    },

    setupCounter(counter) {
        const decrementBtn = counter.querySelector('.counter-decrement');
        const incrementBtn = counter.querySelector('.counter-increment');
        const valueDisplay = counter.querySelector('.counter-value');
        const hiddenInput = counter.querySelector('input[type="hidden"]');
        
        if (!decrementBtn || !incrementBtn || !valueDisplay) {
            return;
        }

        const min = parseInt(counter.getAttribute('data-min') || '0');
        const max = parseInt(counter.getAttribute('data-max') || '999');
        const step = parseInt(counter.getAttribute('data-step') || '1');

        const updateValue = (newValue) => {
            const clampedValue = Math.max(min, Math.min(max, newValue));
            valueDisplay.textContent = clampedValue;
            
            if (hiddenInput) {
                hiddenInput.value = clampedValue;
                hiddenInput.dispatchEvent(new Event('change', { bubbles: true }));
            }

            decrementBtn.disabled = clampedValue <= min;
            incrementBtn.disabled = clampedValue >= max;

            counter.setAttribute('data-value', clampedValue);
            counter.dispatchEvent(new CustomEvent('counterchange', {
                detail: { value: clampedValue },
                bubbles: true
            }));
        };

        decrementBtn.addEventListener('click', (e) => {
            e.preventDefault();
            const currentValue = parseInt(valueDisplay.textContent || '0');
            updateValue(currentValue - step);
        });

        incrementBtn.addEventListener('click', (e) => {
            e.preventDefault();
            const currentValue = parseInt(valueDisplay.textContent || '0');
            updateValue(currentValue + step);
        });

        const initialValue = parseInt(valueDisplay.textContent || '0');
        updateValue(initialValue);
    },

    getValue(counter) {
        const valueDisplay = counter.querySelector('.counter-value');
        return parseInt(valueDisplay?.textContent || '0');
    },

    setValue(counter, value) {
        const valueDisplay = counter.querySelector('.counter-value');
        const hiddenInput = counter.querySelector('input[type="hidden"]');
        
        const min = parseInt(counter.getAttribute('data-min') || '0');
        const max = parseInt(counter.getAttribute('data-max') || '999');
        const clampedValue = Math.max(min, Math.min(max, value));
        
        if (valueDisplay) {
            valueDisplay.textContent = clampedValue;
        }
        
        if (hiddenInput) {
            hiddenInput.value = clampedValue;
        }

        const decrementBtn = counter.querySelector('.counter-decrement');
        const incrementBtn = counter.querySelector('.counter-increment');
        
        if (decrementBtn) {
            decrementBtn.disabled = clampedValue <= min;
        }
        
        if (incrementBtn) {
            incrementBtn.disabled = clampedValue >= max;
        }
    }
};

if (typeof document !== 'undefined') {
    document.addEventListener('DOMContentLoaded', () => {
        Counter.init();
    });
}

if (typeof module !== 'undefined' && module.exports) {
    module.exports = Counter;
}
