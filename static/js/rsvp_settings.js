(function() {
    'use strict';

    class RSVPSettingsPanel extends SlidePanel {
        constructor(triggerBtn) {
            super('.rsvp-settings-panel', {
                onOpen: () => this.handleOpen(),
                onSave: () => this.handleSave(),
                onCancel: () => this.handleCancel()
            });

            this.triggerBtn = triggerBtn;
            this.hasDeadlineCheckbox = document.getElementById('has_rsvp_deadline');
            this.deadlineFields = document.querySelectorAll('.rsvp-deadline-fields');
            this.hasCapacityCheckbox = document.getElementById('has_event_capacity');
            this.capacityFields = document.querySelectorAll('.event-capacity-fields');
            
            this.originalValues = {};
            
            if (!this.panel || !this.overlay || !this.triggerBtn) {
                return;
            }
            
            this.initRSVPSettings();
        }
        
        initRSVPSettings() {
            this.triggerBtn.addEventListener('click', (e) => {
                e.preventDefault();
                this.open();
            });
            
            this.hasDeadlineCheckbox.addEventListener('change', (e) => {
                this.toggleDeadlineFields(e.target.checked);
            });
            
            this.hasCapacityCheckbox.addEventListener('change', (e) => {
                this.toggleCapacityFields(e.target.checked);
            });
        }
        
        handleOpen() {
            this.saveOriginalValues();
        }
        
        handleCancel() {
            this.restoreOriginalValues();
        }
        
        handleSave() {
            if (!this.hasDeadlineCheckbox.checked) {
                document.getElementById('rsvp_deadline_input').value = '';
            }
            
            if (!this.hasCapacityCheckbox.checked) {
                document.getElementById('event_capacity_input').value = '';
            }
            
            this.injectIntoForm();
            this.updateTriggerButtonText();
        }
        
        injectIntoForm() {
            injectRSVPSettingsIntoForm();
        }
        
        saveOriginalValues() {
            this.originalValues = {
                hasDeadline: this.hasDeadlineCheckbox.checked,
                deadline: document.getElementById('rsvp_deadline_input').value,
                allowAfterDeadline: document.getElementById('allow_rsvp_after_deadline').checked,
                allowMaybe: document.getElementById('allow_maybe_rsvp').checked,
                privateList: document.getElementById('private_guest_list').checked,
                maxPlusOnes: document.getElementById('max_plus_ones_input').value,
                familyHeadcount: document.getElementById('family_headcount').checked,
                hasCapacity: this.hasCapacityCheckbox.checked,
                capacity: document.getElementById('event_capacity_input').value
            };
        }
        
        restoreOriginalValues() {
            this.hasDeadlineCheckbox.checked = this.originalValues.hasDeadline;
            document.getElementById('rsvp_deadline_input').value = this.originalValues.deadline;
            document.getElementById('allow_rsvp_after_deadline').checked = this.originalValues.allowAfterDeadline;
            document.getElementById('allow_maybe_rsvp').checked = this.originalValues.allowMaybe;
            document.getElementById('private_guest_list').checked = this.originalValues.privateList;
            document.getElementById('max_plus_ones_input').value = this.originalValues.maxPlusOnes;
            document.getElementById('family_headcount').checked = this.originalValues.familyHeadcount;
            this.hasCapacityCheckbox.checked = this.originalValues.hasCapacity;
            document.getElementById('event_capacity_input').value = this.originalValues.capacity;
            
            this.toggleDeadlineFields(this.originalValues.hasDeadline);
            this.toggleCapacityFields(this.originalValues.hasCapacity);
        }
        
        toggleDeadlineFields(show) {
            this.deadlineFields.forEach(field => {
                field.style.display = show ? 'block' : 'none';
            });
        }
        
        toggleCapacityFields(show) {
            this.capacityFields.forEach(field => {
                field.style.display = show ? 'block' : 'none';
            });
        }
        
        updateTriggerButtonText() {
            const settingsCount = this.getActiveSettingsCount();
            const textElement = this.triggerBtn.querySelector('.rsvp-settings-trigger-text');
            
            if (textElement) {
                if (settingsCount === 0) {
                    textElement.textContent = 'Configure RSVP Settings';
                } else {
                    textElement.textContent = `RSVP Settings (${settingsCount} active)`;
                }
            }
        }
        
        getActiveSettingsCount() {
            let count = 0;
            
            if (this.hasDeadlineCheckbox.checked) count++;
            if (!document.getElementById('allow_maybe_rsvp').checked) count++;
            if (document.getElementById('private_guest_list').checked) count++;
            if (parseInt(document.getElementById('max_plus_ones_input').value) > 0) count++;
            if (document.getElementById('family_headcount').checked) count++;
            if (this.hasCapacityCheckbox.checked) count++;
            
            return count;
        }
    }

    function injectRSVPSettingsIntoForm() {
        const form = document.querySelector('form.event-form');
        if (!form) return;
        
        const setHidden = (name, value) => {
            let input = form.querySelector(`input[name="${name}"][data-rsvp-injected]`);
            if (!input) {
                input = document.createElement('input');
                input.type = 'hidden';
                input.name = name;
                input.setAttribute('data-rsvp-injected', '1');
                form.appendChild(input);
            }
            input.value = value;
        };
        
        // Sentinel: signals to the server that RSVP settings are being submitted
        setHidden('rsvp_settings_saved', '1');
        
        // RSVP deadline
        const deadlineEl = document.getElementById('rsvp_deadline_input');
        setHidden('rsvp_deadline', deadlineEl ? deadlineEl.value : '');
        
        // Checkboxes — submit 'on' when checked, '' when unchecked
        const cb = (id) => { const el = document.getElementById(id); return el && el.checked ? 'on' : ''; };
        setHidden('allow_rsvp_after_deadline', cb('allow_rsvp_after_deadline'));
        setHidden('allow_maybe_rsvp', cb('allow_maybe_rsvp'));
        setHidden('private_guest_list', cb('private_guest_list'));
        setHidden('family_headcount', cb('family_headcount'));
        
        // Max plus ones
        const maxPlusOnesEl = document.getElementById('max_plus_ones_input');
        if (maxPlusOnesEl) setHidden('max_plus_ones', maxPlusOnesEl.value);
        
        // Event capacity
        const capacityEl = document.getElementById('event_capacity_input');
        setHidden('event_capacity', capacityEl ? capacityEl.value : '');
    }

    function initRSVPSettings() {
        const trigger = document.querySelector('[data-rsvp-settings-trigger]');
        if (trigger) {
            new RSVPSettingsPanel(trigger);
        }
        
        // Attach submit listener independently so RSVP panel values are always
        // injected into the form on submit, even if the panel was never opened.
        const form = document.querySelector('form.event-form');
        if (form) {
            form.addEventListener('submit', injectRSVPSettingsIntoForm);
        }
    }

    if (document.readyState === 'loading') {
        document.addEventListener('DOMContentLoaded', initRSVPSettings);
    } else {
        initRSVPSettings();
    }
})();
