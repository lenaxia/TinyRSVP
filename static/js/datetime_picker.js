(function() {
    'use strict';

    class DateTimePicker {
        constructor() {
            this.panel = document.querySelector('.datetime-picker-panel');
            this.overlay = document.querySelector('.datetime-picker-overlay');
            this.closeBtn = document.querySelector('.datetime-picker-close');
            this.startInput = document.getElementById('start_time');
            this.endInput = document.getElementById('end_time');
            this.timezoneInput = document.getElementById('timezone');
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
            this.selectedTimezone = 'America/Los_Angeles';
            this.currentMonth = new Date();
            this.today = new Date();
            this.today.setHours(0, 0, 0, 0);

            this.init();
        }

        init() {
            this.loadExistingValues();
            this.attachEventListeners();
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
        }

        attachEventListeners() {
            this.startInput.addEventListener('click', (e) => {
                e.preventDefault();
                this.currentMode = 'start';
                this.openPanel();
            });
            this.startInput.addEventListener('focus', (e) => {
                e.preventDefault();
                this.startInput.blur();
            });

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
            
            document.querySelector('.datetime-picker-cancel').addEventListener('click', () => {
                this.closePanel();
            });

            document.querySelectorAll('.datetime-toggle-btn').forEach(btn => {
                btn.addEventListener('click', () => {
                    this.switchMode(btn.dataset.mode);
                });
            });

            document.querySelectorAll('.calendar-nav-btn').forEach(btn => {
                btn.addEventListener('click', () => {
                    this.navigateMonth(btn.dataset.nav);
                });
            });

            document.querySelectorAll('.timezone-change-btn').forEach(btn => {
                btn.addEventListener('click', () => {
                    this.openTimezonePanel();
                });
            });

            document.querySelector('.datetime-picker-save').addEventListener('click', () => {
                this.saveDateTime();
            });

            if (this.timezonePanel) {
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
                        this.closeTimezonePanel();
                    } else if (this.panel.classList.contains('open')) {
                        this.closePanel();
                    }
                }
            });
        }

        openPanel() {
            this.panel.classList.add('open');
            this.overlay.classList.add('open');
            document.body.style.overflow = 'hidden';
            this.switchMode(this.currentMode);
            this.renderCalendar();
            this.renderTimePicker();
            this.updateTimezoneDisplay();
        }

        closePanel() {
            this.panel.classList.remove('open');
            this.overlay.classList.remove('open');
            document.body.style.overflow = '';
        }

        openTimezonePanel() {
            this.timezonePanel.classList.add('open');
            this.timezoneOverlay.classList.add('open');
        }

        closeTimezonePanel() {
            this.timezonePanel.classList.remove('open');
            this.timezoneOverlay.classList.remove('open');
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
            const content = document.querySelector(`.datetime-picker-content[data-content="${this.currentMode}"]`);
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
                if (!this.selectedStartTime) {
                    this.selectedStartTime = '12:00 PM';
                }
            } else {
                this.selectedEndDate = newDate;
                if (!this.selectedEndTime) {
                    this.selectedEndTime = '12:00 PM';
                }
            }
            this.renderCalendar();
            this.renderTimePicker();
        }

        renderTimePicker() {
            const content = document.querySelector(`.datetime-picker-content[data-content="${this.currentMode}"]`);
            const timeScroll = content.querySelector('.time-picker-scroll');

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
            
            // Re-render to update selected state
            this.renderTimePicker();
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
            if (this.selectedStartDate && this.selectedStartTime) {
                const dateTime = this.combineDateAndTime(this.selectedStartDate, this.selectedStartTime);
                this.startInput.value = this.formatForInput(dateTime);
            }

            if (this.selectedEndDate && this.selectedEndTime) {
                const dateTime = this.combineDateAndTime(this.selectedEndDate, this.selectedEndTime);
                this.endInput.value = this.formatForInput(dateTime);
            }

            if (this.timezoneInput) {
                this.timezoneInput.value = this.selectedTimezone;
            }

            this.closePanel();
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
})();
