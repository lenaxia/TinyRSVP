(function() {
    'use strict';

    class DateTimePicker {
        constructor(options = {}) {
            this.options = {
                startInputId: options.startInputId || 'start_time',
                endInputId: options.endInputId || 'end_time',
                timezoneInputId: options.timezoneInputId || 'timezone',
                modalId: options.modalId || 'datetime-picker-modal',
                timezoneModalId: options.timezoneModalId || 'timezone-modal',
                ...options
            };

            this.currentMode = 'start';
            this.selectedStartDate = null;
            this.selectedStartTime = null;
            this.selectedEndDate = null;
            this.selectedEndTime = null;
            this.selectedTimezone = 'America/Los_Angeles';
            this.currentMonth = new Date();
            this.today = new Date();
            this.today.setHours(0, 0, 0, 0);

            this.init();
        }

        init() {
            this.startInput = document.getElementById(this.options.startInputId);
            this.endInput = document.getElementById(this.options.endInputId);
            this.timezoneInput = document.getElementById(this.options.timezoneInputId);

            if (!this.startInput) {
                console.error('Start time input not found');
                return;
            }

            this.createModal();
            this.createTimezoneModal();
            this.attachEventListeners();
            this.loadExistingValues();
        }

        loadExistingValues() {
            if (this.startInput && this.startInput.value) {
                const date = new Date(this.startInput.value);
                if (!isNaN(date.getTime())) {
                    this.selectedStartDate = date;
                    this.selectedStartTime = this.formatTime(date);
                }
            }

            if (this.endInput && this.endInput.value) {
                const date = new Date(this.endInput.value);
                if (!isNaN(date.getTime())) {
                    this.selectedEndDate = date;
                    this.selectedEndTime = this.formatTime(date);
                }
            }

            if (this.timezoneInput && this.timezoneInput.value) {
                this.selectedTimezone = this.timezoneInput.value;
            }

            this.updateDisplayField();
        }

        createModal() {
            const modal = document.createElement('div');
            modal.id = this.options.modalId;
            modal.className = 'modal-container';
            modal.innerHTML = `
                <div class="modal-overlay" aria-hidden="true"></div>
                <div class="modal-center modal-md" role="dialog" aria-modal="true" aria-labelledby="datetime-modal-title" aria-hidden="true">
                    <div class="modal-header">
                        <h3 class="modal-title" id="datetime-modal-title">Select Date & Time</h3>
                        <button type="button" class="modal-close" data-modal-close aria-label="Close modal">&times;</button>
                    </div>
                    <div class="modal-body">
                        <div class="datetime-toggle-group">
                            <button type="button" class="datetime-toggle-btn active" data-mode="start">Start Time</button>
                            <button type="button" class="datetime-toggle-btn" data-mode="end">End Time (Optional)</button>
                        </div>

                        <div class="datetime-picker-content active" data-content="start">
                            <div class="datetime-picker-layout">
                                <div class="calendar-container">
                                    <div class="calendar-header">
                                        <button type="button" class="calendar-nav-btn" data-nav="prev" aria-label="Previous month">&lsaquo;</button>
                                        <span class="calendar-month-year"></span>
                                        <button type="button" class="calendar-nav-btn" data-nav="next" aria-label="Next month">&rsaquo;</button>
                                    </div>
                                    <div class="calendar-grid"></div>
                                </div>
                                <div class="time-picker-container">
                                    <label class="time-picker-label">Time</label>
                                    <div class="time-picker-scroll"></div>
                                </div>
                            </div>
                            <div class="timezone-display">
                                <div>
                                    <div class="timezone-label">Timezone</div>
                                    <div class="timezone-value"></div>
                                </div>
                                <button type="button" class="timezone-change-btn">Change</button>
                            </div>
                        </div>

                        <div class="datetime-picker-content" data-content="end">
                            <div class="datetime-picker-layout">
                                <div class="calendar-container">
                                    <div class="calendar-header">
                                        <button type="button" class="calendar-nav-btn" data-nav="prev" aria-label="Previous month">&lsaquo;</button>
                                        <span class="calendar-month-year"></span>
                                        <button type="button" class="calendar-nav-btn" data-nav="next" aria-label="Next month">&rsaquo;</button>
                                    </div>
                                    <div class="calendar-grid"></div>
                                </div>
                                <div class="time-picker-container">
                                    <label class="time-picker-label">Time</label>
                                    <div class="time-picker-scroll"></div>
                                </div>
                            </div>
                            <div class="timezone-display">
                                <div>
                                    <div class="timezone-label">Timezone</div>
                                    <div class="timezone-value"></div>
                                </div>
                                <button type="button" class="timezone-change-btn">Change</button>
                            </div>
                        </div>
                    </div>
                    <div class="modal-footer">
                        <button type="button" class="btn btn-secondary" data-modal-close>Cancel</button>
                        <button type="button" class="btn btn-primary" data-action="save">Save</button>
                    </div>
                </div>
            `;
            document.body.appendChild(modal);

            this.modal = new Modal(this.options.modalId);
            this.modalElement = modal;
        }

        createTimezoneModal() {
            const timezones = [
                { name: 'Pacific Time (PT)', value: 'America/Los_Angeles', offset: 'UTC-8' },
                { name: 'Mountain Time (MT)', value: 'America/Denver', offset: 'UTC-7' },
                { name: 'Central Time (CT)', value: 'America/Chicago', offset: 'UTC-6' },
                { name: 'Eastern Time (ET)', value: 'America/New_York', offset: 'UTC-5' },
                { name: 'UTC', value: 'UTC', offset: 'UTC+0' },
                { name: 'London (GMT)', value: 'Europe/London', offset: 'UTC+0' },
                { name: 'Paris (CET)', value: 'Europe/Paris', offset: 'UTC+1' },
                { name: 'Tokyo (JST)', value: 'Asia/Tokyo', offset: 'UTC+9' },
                { name: 'Sydney (AEDT)', value: 'Australia/Sydney', offset: 'UTC+11' }
            ];

            const modal = document.createElement('div');
            modal.id = this.options.timezoneModalId;
            modal.className = 'modal-container';
            
            const timezoneOptions = timezones.map(tz => `
                <div class="timezone-option" data-timezone="${tz.value}">
                    <div class="timezone-option-name">${tz.name}</div>
                    <div class="timezone-option-offset">${tz.offset}</div>
                </div>
            `).join('');

            modal.innerHTML = `
                <div class="modal-overlay" style="z-index: 1060;" aria-hidden="true"></div>
                <div class="modal-center modal-sm" style="z-index: 1070;" role="dialog" aria-modal="true" aria-labelledby="timezone-modal-title" aria-hidden="true">
                    <div class="modal-header">
                        <h3 class="modal-title" id="timezone-modal-title">Select Timezone</h3>
                        <button type="button" class="modal-close" data-modal-close aria-label="Close modal">&times;</button>
                    </div>
                    <div class="modal-body">
                        <div class="timezone-list">
                            ${timezoneOptions}
                        </div>
                    </div>
                </div>
            `;
            document.body.appendChild(modal);

            this.timezoneModal = new Modal(this.options.timezoneModalId);
        }

        attachEventListeners() {
            const displayInput = document.getElementById('datetime_display');
            console.log('DateTimePicker: Looking for datetime_display element:', displayInput);
            if (displayInput) {
                console.log('DateTimePicker: Attaching click listener to datetime_display');
                displayInput.addEventListener('click', (e) => {
                    console.log('DateTimePicker: Display input clicked!');
                    e.preventDefault();
                    this.openModal();
                });
                displayInput.addEventListener('keydown', (e) => {
                    if (e.key === 'Enter' || e.key === ' ') {
                        e.preventDefault();
                        console.log('DateTimePicker: Display input keyboard activated!');
                        this.openModal();
                    }
                });
            } else {
                console.error('DateTimePicker: datetime_display element not found!');
            }

            this.modalElement.querySelectorAll('.datetime-toggle-btn').forEach(btn => {
                btn.addEventListener('click', () => {
                    this.switchMode(btn.dataset.mode);
                });
            });

            this.modalElement.querySelectorAll('[data-nav]').forEach(btn => {
                btn.addEventListener('click', () => {
                    this.navigateMonth(btn.dataset.nav);
                });
            });

            this.modalElement.querySelectorAll('.timezone-change-btn').forEach(btn => {
                btn.addEventListener('click', () => {
                    this.timezoneModal.open();
                });
            });

            this.modalElement.querySelector('[data-action="save"]').addEventListener('click', () => {
                this.saveDateTime();
            });

            document.getElementById(this.options.timezoneModalId).querySelectorAll('.timezone-option').forEach(option => {
                option.addEventListener('click', () => {
                    this.selectTimezone(option.dataset.timezone);
                });
            });
        }

        openModal() {
            console.log('DateTimePicker: openModal called');
            if (!this.selectedStartDate && !this.selectedStartTime) {
                this.currentMode = 'start';
            }
            this.switchMode(this.currentMode);
            this.renderCalendar();
            this.renderTimePicker();
            this.updateTimezoneDisplay();
            console.log('DateTimePicker: About to open modal, modal object:', this.modal);
            this.modal.open();
        }

        switchMode(mode) {
            this.currentMode = mode;

            this.modalElement.querySelectorAll('.datetime-toggle-btn').forEach(btn => {
                btn.classList.toggle('active', btn.dataset.mode === mode);
            });

            this.modalElement.querySelectorAll('.datetime-picker-content').forEach(content => {
                content.classList.toggle('active', content.dataset.content === mode);
            });

            this.renderCalendar();
            this.renderTimePicker();
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
            const content = this.modalElement.querySelector(`.datetime-picker-content[data-content="${this.currentMode}"]`);
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
                day.addEventListener('click', () => {
                    this.selectDate(dayDate);
                });
            }

            return day;
        }

        selectDate(date) {
            if (this.currentMode === 'start') {
                this.selectedStartDate = date;
            } else {
                this.selectedEndDate = date;
            }
            this.renderCalendar();
        }

        renderTimePicker() {
            const content = this.modalElement.querySelector(`.datetime-picker-content[data-content="${this.currentMode}"]`);
            const timeScroll = content.querySelector('.time-picker-scroll');

            timeScroll.innerHTML = '';

            for (let hour = 0; hour < 24; hour++) {
                for (let minute = 0; minute < 60; minute += 15) {
                    const timeOption = document.createElement('div');
                    timeOption.className = 'time-option';
                    const timeStr = this.formatTimeOption(hour, minute);
                    timeOption.textContent = timeStr;
                    timeOption.dataset.time = timeStr;

                    const selectedTime = this.currentMode === 'start' ? this.selectedStartTime : this.selectedEndTime;
                    if (selectedTime === timeStr) {
                        timeOption.classList.add('selected');
                        setTimeout(() => timeOption.scrollIntoView({ block: 'center' }), 100);
                    }

                    timeOption.addEventListener('click', () => {
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
        }

        updateTimezoneDisplay() {
            const timezoneValues = this.modalElement.querySelectorAll('.timezone-value');
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
            
            const options = document.querySelectorAll('.timezone-option');
            options.forEach(opt => {
                opt.classList.toggle('selected', opt.dataset.timezone === timezone);
            });

            this.timezoneModal.close();
        }

        saveDateTime() {
            if (this.selectedStartDate && this.selectedStartTime) {
                const dateTime = this.combineDateAndTime(this.selectedStartDate, this.selectedStartTime);
                this.startInput.value = this.formatForInput(dateTime);
            }

            if (this.selectedEndDate && this.selectedEndTime) {
                const dateTime = this.combineDateAndTime(this.selectedEndDate, this.selectedEndTime);
                this.endInput.value = this.formatForInput(dateTime);
            } else {
                this.endInput.value = '';
            }

            if (this.timezoneInput) {
                this.timezoneInput.value = this.selectedTimezone;
            }

            this.updateDisplayField();
            this.modal.close();
        }

        updateDisplayField() {
            const displayInput = document.getElementById('datetime_display');
            if (!displayInput) return;

            if (!this.selectedStartDate || !this.selectedStartTime) {
                displayInput.textContent = 'Click to select date and time';
                displayInput.classList.add('empty');
                displayInput.classList.remove('has-value');
                return;
            }

            const startDateTime = this.combineDateAndTime(this.selectedStartDate, this.selectedStartTime);
            const tzName = this.getTimezoneName(this.selectedTimezone);
            
            let displayText = this.formatDisplayDateTime(startDateTime);

            if (this.selectedEndDate && this.selectedEndTime) {
                const endDateTime = this.combineDateAndTime(this.selectedEndDate, this.selectedEndTime);
                displayText += ' - ' + this.formatDisplayDateTime(endDateTime);
            }

            displayText += ' (' + tzName + ')';

            displayInput.textContent = displayText;
            displayInput.classList.remove('empty');
            displayInput.classList.add('has-value');
        }

        formatDisplayDateTime(date) {
            const options = { 
                weekday: 'short', 
                month: 'short', 
                day: 'numeric', 
                year: 'numeric',
                hour: 'numeric',
                minute: '2-digit'
            };
            return date.toLocaleString('en-US', options);
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
    }

    function initDateTimePicker() {
        if (document.getElementById('start_time')) {
            new DateTimePicker();
        }
    }

    if (document.readyState === 'loading') {
        document.addEventListener('DOMContentLoaded', initDateTimePicker);
    } else {
        initDateTimePicker();
    }

    window.DateTimePicker = DateTimePicker;
})();
