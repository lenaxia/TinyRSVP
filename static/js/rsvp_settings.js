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
            
            this.updateTriggerButtonText();
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

    function initRSVPSettings() {
        const trigger = document.querySelector('[data-rsvp-settings-trigger]');
        if (trigger) {
            new RSVPSettingsPanel(trigger);
        }
    }

    if (document.readyState === 'loading') {
        document.addEventListener('DOMContentLoaded', initRSVPSettings);
    } else {
        initRSVPSettings();
    }
})();
