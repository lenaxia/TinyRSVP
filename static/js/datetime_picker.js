(function() {
    'use strict';

    /**
     * DateTimePicker - Reusable datetime picker component
     * 
     * Supports three modes:
     * - datetime-range: Start and end time selection with timezone
     * - datetime-single: Single datetime selection
     * - date-only: Date selection without time picker
     * 
     * Configuration via data attributes:
     * - data-datetime-picker: Marks input for picker initialization
     * - data-mode: 'datetime-range' | 'datetime-single' | 'date-only'
     * - data-end-input: ID of end time input (datetime-range only)
     * - data-timezone-input: ID of timezone input
     * - data-show-timezone: 'true' | 'false'
     * - data-title: Custom panel title
     */
    class DateTimePicker {
        constructor(config) {
            const mode = config.mode || 'datetime-single';
            this.config = {
                mode: mode,
                showTimezone: config.showTimezone !== false,
                showEndTime: mode === 'datetime-range',
                inputId: config.inputId,
                endInputId: config.endInputId,
                timezoneInputId: config.timezoneInputId,
                title: config.title || 'Select Date & Time',
                defaultTimezone: config.defaultTimezone || 'America/Los_Angeles'
            };

            this.panel = document.querySelector('.datetime-picker-panel');
            this.overlay = document.querySelector('.datetime-picker-overlay');
            this.closeBtn = document.querySelector('.datetime-picker-close');
            this.startInput = document.getElementById(this.config.inputId);
            this.endInput = this.config.endInputId ? document.getElementById(this.config.endInputId) : null;
            this.timezoneInput = this.config.timezoneInputId ? document.getElementById(this.config.timezoneInputId) : null;
            this.timezonePanel = document.querySelector('.timezone-picker-panel');
            this.timezoneOverlay = document.querySelector('.timezone-picker-overlay');
            this.timezoneCloseBtn = document.querySelector('.timezone-picker-close');

            if (!this.panel || !this.overlay || !this.startInput) {
                return;
            }

            this.currentMode = 'start';
            this.selectedStartDate = null;
            this.selectedStartTime = null;
            this.selectedEndDate = null;
            this.selectedEndTime = null;
            this.selectedTimezone = this.config.defaultTimezone;
            this.currentMonth = new Date();
            this.today = new Date();
            this.today.setHours(0, 0, 0, 0);

            this.init();
        }

        init() {
            this.updatePanelTitle();
            this.loadExistingValues();
            this.attachEventListeners();
        }

        updatePanelTitle() {
            const titleElement = this.panel.querySelector('.datetime-picker-title');
            if (titleElement) {
                titleElement.textContent = this.config.title;
            }
        }

        updateUIForMode() {
            const toggleGroup = this.panel.querySelector('.datetime-toggle-group');
            const timezoneDisplays = this.panel.querySelectorAll('.timezone-display');
            
            if (toggleGroup) {
                if (this.config.mode === 'datetime-single' || this.config.mode === 'date-only') {
                    toggleGroup.style.setProperty('display', 'none', 'important');
                } else if (this.config.mode === 'datetime-range') {
                    toggleGroup.style.setProperty('display', 'flex', 'important');
                }
            }

            timezoneDisplays.forEach(display => {
                if (this.config.showTimezone) {
                    display.style.removeProperty('display');
                } else {
                    display.style.setProperty('display', 'none', 'important');
                }
            });

            const timeContainers = this.panel.querySelectorAll('.time-picker-container');
            timeContainers.forEach(container => {
                if (this.config.mode === 'date-only') {
                    container.style.setProperty('display', 'none', 'important');
                } else {
                    container.style.removeProperty('display');
                }
            });
        }

        loadExistingValues() {
            // For button triggers, load from hidden fields
            const hiddenStartInput = this.startInput.tagName === 'BUTTON' ? 
                document.getElementById('start_time') : this.startInput;
            const hiddenEndInput = this.startInput.tagName === 'BUTTON' ? 
                document.getElementById('end_time') : this.endInput;
            const hiddenTimezoneInput = this.startInput.tagName === 'BUTTON' ? 
                document.getElementById('timezone') : this.timezoneInput;

            if (hiddenStartInput && hiddenStartInput.value) {
                const date = new Date(hiddenStartInput.value);
                if (!isNaN(date.getTime())) {
                    this.selectedStartDate = date;
                    if (this.config.mode !== 'date-only') {
                        this.selectedStartTime = this.formatTime(date);
                    }
                }
            }

            if (hiddenEndInput && hiddenEndInput.value) {
                const date = new Date(hiddenEndInput.value);
                if (!isNaN(date.getTime())) {
                    this.selectedEndDate = date;
                    if (this.config.mode !== 'date-only') {
                        this.selectedEndTime = this.formatTime(date);
                    }
                }
            }

            if (hiddenTimezoneInput && hiddenTimezoneInput.value) {
                this.selectedTimezone = hiddenTimezoneInput.value;
            }
        }

        attachEventListeners() {
            const isButton = this.startInput.tagName === 'BUTTON';
            
            this.startInput.addEventListener('click', (e) => {
                e.preventDefault();
                this.currentMode = 'start';
                this.openPanel();
            });
            
            if (!isButton) {
                this.startInput.addEventListener('focus', (e) => {
                    e.preventDefault();
                    this.startInput.blur();
                });
            }

            if (this.endInput) {
                this.endInput.addEventListener('click', (e) => {
                    e.preventDefault();
                    this.currentMode = 'end';
                    this.openPanel();
                });
                this.endInput.addEventListener('focus', (e) => {
                    e.preventDefault();
                    this.endInput.blur();
                });
            }

            this.closeBtn.addEventListener('click', () => this.closePanel());
            this.overlay.addEventListener('click', () => this.closePanel());
            
            const cancelBtn = document.querySelector('.datetime-picker-cancel');
            if (cancelBtn) {
                cancelBtn.addEventListener('click', () => {
                    this.closePanel();
                });
            }

            if (this.config.showEndTime) {
                document.querySelectorAll('.datetime-toggle-btn').forEach(btn => {
                    btn.addEventListener('click', () => {
                        this.switchMode(btn.dataset.mode);
                    });
                });
            }

            document.querySelectorAll('.calendar-nav-btn').forEach(btn => {
                btn.addEventListener('click', () => {
                    this.navigateMonth(btn.dataset.nav);
                });
            });

            if (this.config.showTimezone) {
                document.querySelectorAll('.timezone-change-btn').forEach(btn => {
                    btn.addEventListener('click', () => {
                        this.openTimezonePanel();
                    });
                });
            }

            const saveBtn = document.querySelector('.datetime-picker-save');
            if (saveBtn) {
                saveBtn.addEventListener('click', () => {
                    this.saveDateTime();
                });
            }

            if (this.timezonePanel && this.config.showTimezone) {
                this.timezoneCloseBtn.addEventListener('click', () => this.closeTimezonePanel());
                this.timezoneOverlay.addEventListener('click', () => this.closeTimezonePanel());

                document.querySelectorAll('.timezone-option').forEach(option => {
                    option.addEventListener('click', () => {
                        this.selectTimezone(option.dataset.timezone);
                    });
                });
            }

            document.addEventListener('keydown', (e) => {
                if (e.key === 'Escape') {
                    if (this.timezonePanel && this.timezonePanel.classList.contains('open')) {
                        e.stopPropagation();
                        this.closeTimezonePanel();
                    } else if (this.panel.classList.contains('open')) {
                        this.closePanel();
                    }
                }
            });
        }

        openPanel() {
            this.updateUIForMode();
            this.panel.classList.add('open');
            this.overlay.classList.add('open');
            document.body.style.overflow = 'hidden';
            this.switchMode(this.currentMode);
            this.renderCalendar();
            if (this.config.mode !== 'date-only') {
                this.renderTimePicker();
            }
            if (this.config.showTimezone) {
                this.updateTimezoneDisplay();
            }
            if (this.config.showEndTime) {
                this.updateToggleDisplays();
            }
        }

        closePanel() {
            this.panel.classList.remove('open');
            this.overlay.classList.remove('open');
            document.body.style.overflow = '';
        }

        openTimezonePanel() {
            if (this.timezonePanel) {
                this.timezonePanel.classList.add('open');
                this.timezoneOverlay.classList.add('open');
            }
        }

        closeTimezonePanel() {
            if (this.timezonePanel) {
                this.timezonePanel.classList.remove('open');
                this.timezoneOverlay.classList.remove('open');
            }
        }

        switchMode(mode) {
            this.currentMode = mode;

            document.querySelectorAll('.datetime-toggle-btn').forEach(btn => {
                btn.classList.toggle('active', btn.dataset.mode === mode);
            });

            document.querySelectorAll('.datetime-picker-content').forEach(content => {
                content.classList.toggle('active', content.dataset.content === mode);
            });

            this.renderCalendar();
            if (this.config.mode !== 'date-only') {
                this.renderTimePicker();
            }
        }

        navigateMonth(direction) {
            if (direction === 'prev') {
                this.currentMonth.setMonth(this.currentMonth.getMonth() - 1);
            } else {
                this.currentMonth.setMonth(this.currentMonth.getMonth() + 1);
            }
            this.renderCalendar();
        }

        renderCalendar() {
            const content = document.querySelector(`.datetime-picker-content[data-content="${this.currentMode}"]`);
            if (!content) return;

            const monthYearSpan = content.querySelector('.calendar-month-year');
            const calendarGrid = content.querySelector('.calendar-grid');
            const prevBtn = content.querySelector('[data-nav="prev"]');

            const year = this.currentMonth.getFullYear();
            const month = this.currentMonth.getMonth();

            monthYearSpan.textContent = this.currentMonth.toLocaleDateString('en-US', { month: 'long', year: 'numeric' });

            const currentMonthStart = new Date(year, month, 1);
            currentMonthStart.setHours(0, 0, 0, 0);
            prevBtn.disabled = currentMonthStart <= this.today;

            calendarGrid.innerHTML = '';

            const dayHeaders = ['S', 'M', 'T', 'W', 'T', 'F', 'S'];
            dayHeaders.forEach(day => {
                const header = document.createElement('div');
                header.className = 'calendar-day-header';
                header.textContent = day;
                calendarGrid.appendChild(header);
            });

            const firstDay = new Date(year, month, 1).getDay();
            const daysInMonth = new Date(year, month + 1, 0).getDate();
            const daysInPrevMonth = new Date(year, month, 0).getDate();

            for (let i = firstDay - 1; i >= 0; i--) {
                const day = this.createDayElement(daysInPrevMonth - i, true, year, month - 1);
                calendarGrid.appendChild(day);
            }

            for (let i = 1; i <= daysInMonth; i++) {
                const day = this.createDayElement(i, false, year, month);
                calendarGrid.appendChild(day);
            }

            const totalCells = calendarGrid.children.length - 7;
            const remainingCells = 42 - totalCells;
            for (let i = 1; i <= remainingCells; i++) {
                const day = this.createDayElement(i, true, year, month + 1);
                calendarGrid.appendChild(day);
            }
        }

        createDayElement(dayNum, isOtherMonth, year, month) {
            const day = document.createElement('div');
            day.className = 'calendar-day';
            day.textContent = dayNum;

            const dayDate = new Date(year, month, dayNum);
            dayDate.setHours(0, 0, 0, 0);

            if (isOtherMonth) {
                day.classList.add('other-month');
            }

            if (dayDate < this.today) {
                day.classList.add('disabled');
            }

            if (dayDate.getTime() === this.today.getTime()) {
                day.classList.add('today');
            }

            const selectedDate = this.currentMode === 'start' ? this.selectedStartDate : this.selectedEndDate;
            if (selectedDate) {
                const selected = new Date(selectedDate);
                selected.setHours(0, 0, 0, 0);
                if (dayDate.getTime() === selected.getTime()) {
                    day.classList.add('selected');
                }
            }

            if (!isOtherMonth && dayDate >= this.today) {
                day.style.cursor = 'pointer';
                day.addEventListener('click', (e) => {
                    e.preventDefault();
                    e.stopPropagation();
                    this.selectDate(new Date(dayDate));
                });
            }

            return day;
        }

        selectDate(date) {
            const newDate = new Date(date);
            newDate.setHours(0, 0, 0, 0);
            
            if (this.currentMode === 'start') {
                this.selectedStartDate = newDate;
                if (!this.selectedStartTime && this.config.mode !== 'date-only') {
                    this.selectedStartTime = '12:00 PM';
                }
            } else {
                this.selectedEndDate = newDate;
                if (!this.selectedEndTime && this.config.mode !== 'date-only') {
                    this.selectedEndTime = '12:00 PM';
                }
            }
            this.renderCalendar();
            if (this.config.mode !== 'date-only') {
                this.renderTimePicker();
            }
            if (this.config.showEndTime) {
                this.updateToggleDisplays();
            }
        }

        renderTimePicker() {
            const content = document.querySelector(`.datetime-picker-content[data-content="${this.currentMode}"]`);
            if (!content) return;

            const timeScroll = content.querySelector('.time-picker-scroll');
            if (!timeScroll) return;

            timeScroll.innerHTML = '';

            for (let hour = 0; hour < 24; hour++) {
                for (let minute = 0; minute < 60; minute += 15) {
                    const timeOption = document.createElement('div');
                    timeOption.className = 'time-option';
                    const timeStr = this.formatTimeOption(hour, minute);
                    timeOption.textContent = timeStr;

                    const selectedTime = this.currentMode === 'start' ? this.selectedStartTime : this.selectedEndTime;
                    if (selectedTime === timeStr) {
                        timeOption.classList.add('selected');
                        setTimeout(() => {
                            timeOption.scrollIntoView({ block: 'center', behavior: 'smooth' });
                        }, 100);
                    }

                    timeOption.addEventListener('click', (e) => {
                        e.preventDefault();
                        e.stopPropagation();
                        this.selectTime(timeStr);
                    });

                    timeScroll.appendChild(timeOption);
                }
            }
        }

        formatTimeOption(hour, minute) {
            const period = hour >= 12 ? 'PM' : 'AM';
            const displayHour = hour === 0 ? 12 : hour > 12 ? hour - 12 : hour;
            const displayMinute = minute.toString().padStart(2, '0');
            return `${displayHour}:${displayMinute} ${period}`;
        }

        formatTime(date) {
            const hour = date.getHours();
            const minute = date.getMinutes();
            return this.formatTimeOption(hour, minute);
        }

        selectTime(timeStr) {
            if (this.currentMode === 'start') {
                this.selectedStartTime = timeStr;
            } else {
                this.selectedEndTime = timeStr;
            }
            
            this.renderTimePicker();
            if (this.config.showEndTime) {
                this.updateToggleDisplays();
            }
        }

        updateTimezoneDisplay() {
            const timezoneValues = document.querySelectorAll('.timezone-value');
            const tzName = this.getTimezoneName(this.selectedTimezone);
            timezoneValues.forEach(el => {
                el.textContent = tzName;
            });
        }

        getTimezoneName(value) {
            const tzMap = {
                'America/Los_Angeles': 'Pacific Time (PT)',
                'America/Denver': 'Mountain Time (MT)',
                'America/Chicago': 'Central Time (CT)',
                'America/New_York': 'Eastern Time (ET)',
                'UTC': 'UTC',
                'Europe/London': 'London (GMT)',
                'Europe/Paris': 'Paris (CET)',
                'Asia/Tokyo': 'Tokyo (JST)',
                'Australia/Sydney': 'Sydney (AEDT)'
            };
            return tzMap[value] || value;
        }

        selectTimezone(timezone) {
            this.selectedTimezone = timezone;
            this.updateTimezoneDisplay();
            
            document.querySelectorAll('.timezone-option').forEach(opt => {
                opt.classList.toggle('selected', opt.dataset.timezone === timezone);
            });

            this.closeTimezonePanel();
        }

        saveDateTime() {
            const hiddenStartInput = document.getElementById(this.config.inputId === 'event_datetime_trigger' ? 'start_time' : this.config.inputId);
            const hiddenEndInput = this.config.endInputId ? document.getElementById(this.config.endInputId) : this.endInput;
            const hiddenTimezoneInput = this.config.timezoneInputId ? document.getElementById(this.config.timezoneInputId) : this.timezoneInput;

            if (this.selectedStartDate) {
                if (this.config.mode === 'date-only') {
                    const value = this.formatDateForInput(this.selectedStartDate);
                    if (hiddenStartInput) hiddenStartInput.value = value;
                } else if (this.selectedStartTime) {
                    const dateTime = this.combineDateAndTime(this.selectedStartDate, this.selectedStartTime);
                    const value = this.formatForInput(dateTime);
                    if (hiddenStartInput) hiddenStartInput.value = value;
                }
            }

            if (hiddenEndInput && this.selectedEndDate) {
                if (this.config.mode === 'date-only') {
                    hiddenEndInput.value = this.formatDateForInput(this.selectedEndDate);
                } else if (this.selectedEndTime) {
                    const dateTime = this.combineDateAndTime(this.selectedEndDate, this.selectedEndTime);
                    hiddenEndInput.value = this.formatForInput(dateTime);
                }
            }

            if (hiddenTimezoneInput && this.config.showTimezone) {
                hiddenTimezoneInput.value = this.selectedTimezone;
            }

            this.updateTriggerDisplay();
            this.closePanel();
        }

        updateTriggerDisplay() {
            if (this.startInput.tagName !== 'BUTTON') return;

            const displayElement = this.startInput.querySelector('.datetime-trigger-text');
            if (!displayElement) return;

            if (this.selectedStartDate && this.selectedStartTime) {
                const startStr = this.selectedStartDate.toLocaleDateString('en-US', {
                    month: 'short',
                    day: 'numeric',
                    year: 'numeric'
                }) + ' at ' + this.selectedStartTime;

                let displayText = startStr;

                if (this.selectedEndDate && this.selectedEndTime) {
                    const startDateStr = this.selectedStartDate.toLocaleDateString('en-US', {
                        year: 'numeric',
                        month: '2-digit',
                        day: '2-digit'
                    });
                    const endDateStr = this.selectedEndDate.toLocaleDateString('en-US', {
                        year: 'numeric',
                        month: '2-digit',
                        day: '2-digit'
                    });
                    
                    if (startDateStr !== endDateStr) {
                        const endStr = this.selectedEndDate.toLocaleDateString('en-US', {
                            month: 'short',
                            day: 'numeric',
                            year: 'numeric'
                        }) + ' at ' + this.selectedEndTime;
                        displayText += ' - ' + endStr;
                    } else {
                        displayText += ' - ' + this.selectedEndTime;
                    }
                }

                if (this.config.showTimezone && this.selectedTimezone) {
                    const tzName = this.getTimezoneName(this.selectedTimezone);
                    displayText += ' (' + tzName + ')';
                }

                displayElement.textContent = displayText;
            } else {
                displayElement.textContent = 'Click to select date, time, and timezone';
            }
        }

        combineDateAndTime(date, timeStr) {
            const [time, period] = timeStr.split(' ');
            const [hourStr, minuteStr] = time.split(':');
            let hour = parseInt(hourStr);
            const minute = parseInt(minuteStr);

            if (period === 'PM' && hour !== 12) {
                hour += 12;
            } else if (period === 'AM' && hour === 12) {
                hour = 0;
            }

            const result = new Date(date);
            result.setHours(hour, minute, 0, 0);
            return result;
        }

        formatForInput(date) {
            const year = date.getFullYear();
            const month = (date.getMonth() + 1).toString().padStart(2, '0');
            const day = date.getDate().toString().padStart(2, '0');
            const hour = date.getHours().toString().padStart(2, '0');
            const minute = date.getMinutes().toString().padStart(2, '0');
            return `${year}-${month}-${day}T${hour}:${minute}`;
        }

        formatDateForInput(date) {
            const year = date.getFullYear();
            const month = (date.getMonth() + 1).toString().padStart(2, '0');
            const day = date.getDate().toString().padStart(2, '0');
            return `${year}-${month}-${day}`;
        }

        updateToggleDisplays() {
            const startDisplay = document.getElementById('start-time-display');
            const endDisplay = document.getElementById('end-time-display');

            if (startDisplay) {
                if (this.selectedStartDate && this.selectedStartTime) {
                    const dateStr = this.selectedStartDate.toLocaleDateString('en-US', { 
                        month: 'short', 
                        day: 'numeric',
                        year: 'numeric'
                    });
                    startDisplay.textContent = `${dateStr} at ${this.selectedStartTime}`;
                } else {
                    startDisplay.textContent = 'Not set';
                }
            }

            if (endDisplay) {
                if (this.selectedEndDate && this.selectedEndTime) {
                    const dateStr = this.selectedEndDate.toLocaleDateString('en-US', { 
                        month: 'short', 
                        day: 'numeric',
                        year: 'numeric'
                    });
                    endDisplay.textContent = `${dateStr} at ${this.selectedEndTime}`;
                } else {
                    endDisplay.textContent = 'Not set';
                }
            }
        }
    }

    function initDateTimePickers() {
        const inputs = document.querySelectorAll('[data-datetime-picker]');
        inputs.forEach(input => {
            const config = {
                mode: input.dataset.mode || 'datetime-single',
                showTimezone: input.dataset.showTimezone !== 'false',
                inputId: input.id,
                endInputId: input.dataset.endInput,
                timezoneInputId: input.dataset.timezoneInput,
                title: input.dataset.title || 'Select Date & Time',
                defaultTimezone: input.dataset.defaultTimezone || 'America/Los_Angeles'
            };
            new DateTimePicker(config);
        });
    }

    if (document.readyState === 'loading') {
        document.addEventListener('DOMContentLoaded', initDateTimePickers);
    } else {
        initDateTimePickers();
    }
})();
