const DateDefaults = {
    detectTimezone() {
        try {
            return Intl.DateTimeFormat().resolvedOptions().timeZone;
        } catch (e) {
            return 'America/Los_Angeles';
        }
    },

    getDefaultStartTime() {
        const now = new Date();
        now.setDate(now.getDate() + 7);
        now.setHours(18, 0, 0, 0);
        return this.formatDateTimeLocal(now);
    },

    getDefaultEndTime() {
        const now = new Date();
        now.setDate(now.getDate() + 7);
        now.setHours(21, 0, 0, 0);
        return this.formatDateTimeLocal(now);
    },

    getDefaultRSVPDeadline() {
        const now = new Date();
        now.setDate(now.getDate() + 5);
        now.setHours(23, 59, 0, 0);
        return this.formatDateTimeLocal(now);
    },

    formatDateTimeLocal(date) {
        const year = date.getFullYear();
        const month = String(date.getMonth() + 1).padStart(2, '0');
        const day = String(date.getDate()).padStart(2, '0');
        const hours = String(date.getHours()).padStart(2, '0');
        const minutes = String(date.getMinutes()).padStart(2, '0');
        return `${year}-${month}-${day}T${hours}:${minutes}`;
    },

    setTimezoneDefault() {
        const timezoneSelect = document.getElementById('timezone');
        if (!timezoneSelect || timezoneSelect.value) {
            return;
        }

        const detectedTimezone = this.detectTimezone();
        
        const options = Array.from(timezoneSelect.options);
        const matchingOption = options.find(opt => opt.value === detectedTimezone);
        
        if (matchingOption) {
            timezoneSelect.value = detectedTimezone;
        } else {
            const fallbackMap = {
                'America/Los_Angeles': 'America/Los_Angeles',
                'America/Denver': 'America/Denver',
                'America/Chicago': 'America/Chicago',
                'America/New_York': 'America/New_York',
            };
            
            for (const [pattern, value] of Object.entries(fallbackMap)) {
                if (detectedTimezone.includes(pattern.split('/')[1])) {
                    timezoneSelect.value = value;
                    return;
                }
            }
            
            timezoneSelect.value = 'America/Los_Angeles';
        }
    },

    setDateDefaults() {
        const startTimeInput = document.getElementById('start_time');
        const endTimeInput = document.getElementById('end_time');
        const rsvpDeadlineInput = document.getElementById('rsvp_deadline');

        if (startTimeInput && !startTimeInput.value) {
            startTimeInput.value = this.getDefaultStartTime();
        }

        if (endTimeInput && !endTimeInput.value) {
            endTimeInput.value = this.getDefaultEndTime();
        }

        if (rsvpDeadlineInput && !rsvpDeadlineInput.value) {
            rsvpDeadlineInput.value = this.getDefaultRSVPDeadline();
        }
    },

    init() {
        this.setTimezoneDefault();
        this.setDateDefaults();
    }
};

if (typeof document !== 'undefined') {
    document.addEventListener('DOMContentLoaded', () => {
        DateDefaults.init();
    });
}

if (typeof module !== 'undefined' && module.exports) {
    module.exports = DateDefaults;
}
