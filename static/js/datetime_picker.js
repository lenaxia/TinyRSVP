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
     * Configuration via data attributes on the trigger button:
     * - data-datetime-picker: Marks the trigger for picker initialization
     * - data-mode: 'datetime-range' | 'datetime-single' | 'date-only'
     * - data-start-input: ID of the hidden input to read/write the start value
     * - data-end-input: ID of the hidden input to read/write the end value (datetime-range only)
     * - data-timezone-input: ID of the hidden timezone input
     * - data-show-timezone: 'true' | 'false'
     * - data-title: Custom panel title
     *
     * Architecture: all trigger buttons on the page share a single panel DOM
     * element. A module-level `activeInstance` pointer tracks which
     * DateTimePicker instance opened the panel. Panel-level buttons (save,
     * cancel, close, nav, timezone) are bound once at module init and delegate
     * to `activeInstance`. Per-instance state (selected dates/times, config)
     * is stored on each instance.
     */

    // -------------------------------------------------------------------------
    // Module-level shared panel references (bound once)
    // -------------------------------------------------------------------------
    let activeInstance = null;

    const panel        = document.querySelector('.datetime-picker-panel');
    const overlay      = document.querySelector('.datetime-picker-overlay');
    const closeBtn     = document.querySelector('.datetime-picker-close');
    const saveBtn      = document.querySelector('.datetime-picker-save');
    const cancelBtn    = document.querySelector('.datetime-picker-cancel');
    const tzPanel      = document.querySelector('.timezone-picker-panel');
    const tzOverlay    = document.querySelector('.timezone-picker-overlay');
    const tzCloseBtn   = document.querySelector('.timezone-picker-close');

    function bindPanelListeners() {
        if (!panel || !overlay) return;

        closeBtn && closeBtn.addEventListener('click', () => activeInstance && activeInstance.closePanel());
        overlay.addEventListener('click', () => activeInstance && activeInstance.closePanel());
        cancelBtn && cancelBtn.addEventListener('click', () => activeInstance && activeInstance.closePanel());
        saveBtn && saveBtn.addEventListener('click', () => activeInstance && activeInstance.saveDateTime());

        document.querySelectorAll('.datetime-toggle-btn').forEach(btn => {
            btn.addEventListener('click', () => activeInstance && activeInstance.switchMode(btn.dataset.mode));
        });

        document.querySelectorAll('.calendar-nav-btn').forEach(btn => {
            btn.addEventListener('click', () => activeInstance && activeInstance.navigateMonth(btn.dataset.nav));
        });

        document.querySelectorAll('.timezone-change-btn').forEach(btn => {
            btn.addEventListener('click', () => activeInstance && activeInstance.openTimezonePanel());
        });

        if (tzPanel) {
            tzCloseBtn && tzCloseBtn.addEventListener('click', () => activeInstance && activeInstance.closeTimezonePanel());
            tzOverlay && tzOverlay.addEventListener('click', () => activeInstance && activeInstance.closeTimezonePanel());
            document.querySelectorAll('.timezone-option').forEach(opt => {
                opt.addEventListener('click', () => activeInstance && activeInstance.selectTimezone(opt.dataset.timezone));
            });
        }

        document.addEventListener('keydown', (e) => {
            if (!activeInstance) return;
            if (e.key === 'Escape') {
                if (tzPanel && tzPanel.classList.contains('open')) {
                    e.stopPropagation();
                    activeInstance.closeTimezonePanel();
                } else if (panel.classList.contains('open')) {
                    activeInstance.closePanel();
                }
            }
        });
    }

    // -------------------------------------------------------------------------
    // DateTimePicker instance — owns per-trigger state only
    // -------------------------------------------------------------------------
    class DateTimePicker {
        constructor(triggerEl) {
            const mode = triggerEl.dataset.mode || 'datetime-single';
            this.config = {
                mode,
                showTimezone:    triggerEl.dataset.showTimezone !== 'false',
                showEndTime:     mode === 'datetime-range',
                triggerEl,
                startInputId:    triggerEl.dataset.startInput  || triggerEl.id,
                endInputId:      triggerEl.dataset.endInput     || null,
                timezoneInputId: triggerEl.dataset.timezoneInput || null,
                title:           triggerEl.dataset.title        || 'Select Date & Time',
                defaultTimezone: triggerEl.dataset.defaultTimezone || 'America/Los_Angeles',
            };

            // Per-instance selection state
            this.currentMode       = 'start';
            this.selectedStartDate = null;
            this.selectedStartTime = null;
            this.selectedEndDate   = null;
            this.selectedEndTime   = null;
            this.selectedTimezone  = this.config.defaultTimezone;
            this.currentMonth      = new Date();
            this.today             = new Date();
            this.today.setHours(0, 0, 0, 0);

            if (!panel || !overlay) return;

            this._bindTrigger();
            this._loadExistingValues();
        }

        // ---- private --------------------------------------------------------

        _bindTrigger() {
            this.config.triggerEl.addEventListener('click', (e) => {
                e.preventDefault();
                this.currentMode = 'start';
                activeInstance = this;
                this.openPanel();
            });
        }

        _loadExistingValues() {
            const startInput = document.getElementById(this.config.startInputId);
            const endInput   = this.config.endInputId   ? document.getElementById(this.config.endInputId)   : null;
            const tzInput    = this.config.timezoneInputId ? document.getElementById(this.config.timezoneInputId) : null;

            if (startInput && startInput.value) {
                const d = new Date(startInput.value);
                if (!isNaN(d.getTime())) {
                    this.selectedStartDate = d;
                    if (this.config.mode !== 'date-only') this.selectedStartTime = this._formatTime(d);
                }
            }

            if (endInput && endInput.value) {
                const d = new Date(endInput.value);
                if (!isNaN(d.getTime())) {
                    this.selectedEndDate = d;
                    if (this.config.mode !== 'date-only') this.selectedEndTime = this._formatTime(d);
                }
            }

            if (tzInput && tzInput.value) {
                this.selectedTimezone = tzInput.value;
            }
        }

        // ---- public API (called by panel-level listeners via activeInstance) -

        openPanel() {
            this._updatePanelTitle();
            this._updateUIForMode();
            panel.classList.add('open');
            overlay.classList.add('open');
            document.body.style.overflow = 'hidden';
            // Always activate the correct content pane. Without this, a prior
            // picker that left the panel in 'end' mode would leave [data-content="end"]
            // visible when a single-mode picker (e.g. RSVP deadline) opens next,
            // causing the user to interact with stale content from the previous instance.
            this.switchMode(this.currentMode);
            if (this.config.showTimezone)  this._updateTimezoneDisplay();
            if (this.config.showEndTime)   this._updateToggleDisplays();
        }

        closePanel() {
            panel.classList.remove('open');
            overlay.classList.remove('open');
            document.body.style.overflow = '';
        }

        openTimezonePanel() {
            if (tzPanel) {
                tzPanel.classList.add('open');
                tzOverlay && tzOverlay.classList.add('open');
            }
        }

        closeTimezonePanel() {
            if (tzPanel) {
                tzPanel.classList.remove('open');
                tzOverlay && tzOverlay.classList.remove('open');
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
            if (this.config.mode !== 'date-only') this.renderTimePicker();
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
            const calendarGrid  = content.querySelector('.calendar-grid');
            const prevBtn       = content.querySelector('[data-nav="prev"]');

            const year  = this.currentMonth.getFullYear();
            const month = this.currentMonth.getMonth();

            monthYearSpan.textContent = this.currentMonth.toLocaleDateString('en-US', { month: 'long', year: 'numeric' });

            const currentMonthStart = new Date(year, month, 1);
            currentMonthStart.setHours(0, 0, 0, 0);
            prevBtn.disabled = currentMonthStart <= this.today;

            calendarGrid.innerHTML = '';

            ['S','M','T','W','T','F','S'].forEach(d => {
                const h = document.createElement('div');
                h.className = 'calendar-day-header';
                h.textContent = d;
                calendarGrid.appendChild(h);
            });

            const firstDay      = new Date(year, month, 1).getDay();
            const daysInMonth   = new Date(year, month + 1, 0).getDate();
            const daysInPrevMon = new Date(year, month, 0).getDate();

            for (let i = firstDay - 1; i >= 0; i--) {
                calendarGrid.appendChild(this._createDayElement(daysInPrevMon - i, true, year, month - 1));
            }
            for (let i = 1; i <= daysInMonth; i++) {
                calendarGrid.appendChild(this._createDayElement(i, false, year, month));
            }
            const totalCells     = calendarGrid.children.length - 7;
            const remainingCells = 42 - totalCells;
            for (let i = 1; i <= remainingCells; i++) {
                calendarGrid.appendChild(this._createDayElement(i, true, year, month + 1));
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
                    const timeStr   = this._formatTimeOption(hour, minute);
                    const timeOption = document.createElement('div');
                    timeOption.className  = 'time-option';
                    timeOption.textContent = timeStr;

                    const selectedTime = this.currentMode === 'start' ? this.selectedStartTime : this.selectedEndTime;
                    if (selectedTime === timeStr) {
                        timeOption.classList.add('selected');
                        setTimeout(() => timeOption.scrollIntoView({ block: 'center', behavior: 'smooth' }), 100);
                    }

                    timeOption.addEventListener('click', (e) => {
                        e.preventDefault();
                        e.stopPropagation();
                        this._selectTime(timeStr);
                    });
                    timeScroll.appendChild(timeOption);
                }
            }
        }

        saveDateTime() {
            const startInput = document.getElementById(this.config.startInputId);
            const endInput   = this.config.endInputId   ? document.getElementById(this.config.endInputId)   : null;
            const tzInput    = this.config.timezoneInputId ? document.getElementById(this.config.timezoneInputId) : null;

            if (this.selectedStartDate) {
                if (this.config.mode === 'date-only') {
                    if (startInput) startInput.value = this._formatDateForInput(this.selectedStartDate);
                } else if (this.selectedStartTime) {
                    const dt = this._combineDateAndTime(this.selectedStartDate, this.selectedStartTime);
                    if (startInput) startInput.value = this._formatForInput(dt);
                }
            }

            if (endInput && this.selectedEndDate) {
                if (this.config.mode === 'date-only') {
                    endInput.value = this._formatDateForInput(this.selectedEndDate);
                } else if (this.selectedEndTime) {
                    const dt = this._combineDateAndTime(this.selectedEndDate, this.selectedEndTime);
                    endInput.value = this._formatForInput(dt);
                }
            }

            if (tzInput && this.config.showTimezone) {
                tzInput.value = this.selectedTimezone;
            }

            this._updateTriggerDisplay();
            this.closePanel();
        }

        selectTimezone(timezone) {
            this.selectedTimezone = timezone;
            this._updateTimezoneDisplay();
            document.querySelectorAll('.timezone-option').forEach(opt => {
                opt.classList.toggle('selected', opt.dataset.timezone === timezone);
            });
            this.closeTimezonePanel();
        }

        // ---- private helpers ------------------------------------------------

        _updatePanelTitle() {
            const titleEl = panel.querySelector('.datetime-picker-title');
            if (titleEl) titleEl.textContent = this.config.title;
        }

        _updateUIForMode() {
            const toggleGroup = panel.querySelector('.datetime-toggle-group');
            if (toggleGroup) {
                toggleGroup.style.display = this.config.showEndTime ? 'flex' : 'none';
            }

            panel.querySelectorAll('.timezone-display').forEach(el => {
                el.style.display = this.config.showTimezone ? '' : 'none';
            });

            panel.querySelectorAll('.time-picker-container').forEach(el => {
                el.style.display = this.config.mode === 'date-only' ? 'none' : '';
            });
        }

        _createDayElement(dayNum, isOtherMonth, year, month) {
            const day     = document.createElement('div');
            day.className = 'calendar-day';
            day.textContent = dayNum;

            const dayDate = new Date(year, month, dayNum);
            dayDate.setHours(0, 0, 0, 0);

            if (isOtherMonth) day.classList.add('other-month');
            if (dayDate < this.today) day.classList.add('disabled');
            if (dayDate.getTime() === this.today.getTime()) day.classList.add('today');

            const selectedDate = this.currentMode === 'start' ? this.selectedStartDate : this.selectedEndDate;
            if (selectedDate) {
                const sel = new Date(selectedDate);
                sel.setHours(0, 0, 0, 0);
                if (dayDate.getTime() === sel.getTime()) day.classList.add('selected');
            }

            if (!isOtherMonth && dayDate >= this.today) {
                day.style.cursor = 'pointer';
                day.addEventListener('click', (e) => {
                    e.preventDefault();
                    e.stopPropagation();
                    this._selectDate(new Date(dayDate));
                });
            }
            return day;
        }

        _selectDate(date) {
            const newDate = new Date(date);
            newDate.setHours(0, 0, 0, 0);
            if (this.currentMode === 'start') {
                this.selectedStartDate = newDate;
                if (!this.selectedStartTime && this.config.mode !== 'date-only') this.selectedStartTime = '12:00 PM';
            } else {
                this.selectedEndDate = newDate;
                if (!this.selectedEndTime && this.config.mode !== 'date-only') this.selectedEndTime = '12:00 PM';
            }
            this.renderCalendar();            if (this.config.mode !== 'date-only') this.renderTimePicker();
            if (this.config.showEndTime) this._updateToggleDisplays();
        }

        _selectTime(timeStr) {
            if (this.currentMode === 'start') {
                this.selectedStartTime = timeStr;
            } else {
                this.selectedEndTime = timeStr;
            }
            this.renderTimePicker();
            if (this.config.showEndTime) this._updateToggleDisplays();
        }

        _updateTimezoneDisplay() {
            const tzName = this._getTimezoneName(this.selectedTimezone);
            panel.querySelectorAll('.timezone-value').forEach(el => { el.textContent = tzName; });
        }

        _getTimezoneName(value) {
            try {
                // Use Intl to get the real DST-aware abbreviation (e.g. PDT vs PST)
                const now = new Date();
                const parts = new Intl.DateTimeFormat('en-US', {
                    timeZone: value,
                    timeZoneName: 'short',
                }).formatToParts(now);
                const tzPart = parts.find(p => p.type === 'timeZoneName');
                return tzPart ? tzPart.value : value;
            } catch (e) {
                return value;
            }
        }

        _updateTriggerDisplay() {
            const triggerEl = this.config.triggerEl;
            if (triggerEl.tagName !== 'BUTTON') return;

            const displayEl = triggerEl.querySelector('.datetime-trigger-text');
            if (!displayEl) return;

            if (this.selectedStartDate && this.selectedStartTime) {
                const startStr = this.selectedStartDate.toLocaleDateString('en-US', {
                    month: 'short', day: 'numeric', year: 'numeric'
                }) + ' at ' + this.selectedStartTime;

                let displayText = startStr;

                if (this.config.showEndTime && this.selectedEndDate && this.selectedEndTime) {
                    const fmt = { year: 'numeric', month: '2-digit', day: '2-digit' };
                    if (this.selectedStartDate.toLocaleDateString('en-US', fmt) !==
                        this.selectedEndDate.toLocaleDateString('en-US', fmt)) {
                        displayText += ' - ' + this.selectedEndDate.toLocaleDateString('en-US', {
                            month: 'short', day: 'numeric', year: 'numeric'
                        }) + ' at ' + this.selectedEndTime;
                    } else {
                        displayText += ' - ' + this.selectedEndTime;
                    }
                }

                if (this.config.showTimezone && this.selectedTimezone) {
                    displayText += ' (' + this._getTimezoneName(this.selectedTimezone) + ')';
                }

                displayEl.textContent = displayText;
            } else {
                displayEl.textContent = 'Click to select date, time, and timezone';
            }
        }

        _updateToggleDisplays() {
            const startDisplay = document.getElementById('start-time-display');
            const endDisplay   = document.getElementById('end-time-display');

            if (startDisplay) {
                startDisplay.textContent = (this.selectedStartDate && this.selectedStartTime)
                    ? this.selectedStartDate.toLocaleDateString('en-US', { month: 'short', day: 'numeric', year: 'numeric' })
                      + ' at ' + this.selectedStartTime
                    : 'Not set';
            }
            if (endDisplay) {
                endDisplay.textContent = (this.selectedEndDate && this.selectedEndTime)
                    ? this.selectedEndDate.toLocaleDateString('en-US', { month: 'short', day: 'numeric', year: 'numeric' })
                      + ' at ' + this.selectedEndTime
                    : 'Not set';
            }
        }

        _combineDateAndTime(date, timeStr) {
            const [time, period] = timeStr.split(' ');
            const [hourStr, minuteStr] = time.split(':');
            let hour = parseInt(hourStr);
            const minute = parseInt(minuteStr);
            if (period === 'PM' && hour !== 12) hour += 12;
            else if (period === 'AM' && hour === 12) hour = 0;
            const result = new Date(date);
            result.setHours(hour, minute, 0, 0);
            return result;
        }

        _formatForInput(date) {
            const year   = date.getFullYear();
            const month  = (date.getMonth() + 1).toString().padStart(2, '0');
            const day    = date.getDate().toString().padStart(2, '0');
            const hour   = date.getHours().toString().padStart(2, '0');
            const minute = date.getMinutes().toString().padStart(2, '0');
            return `${year}-${month}-${day}T${hour}:${minute}`;
        }

        _formatDateForInput(date) {
            const year  = date.getFullYear();
            const month = (date.getMonth() + 1).toString().padStart(2, '0');
            const day   = date.getDate().toString().padStart(2, '0');
            return `${year}-${month}-${day}`;
        }

        _formatTimeOption(hour, minute) {
            const period      = hour >= 12 ? 'PM' : 'AM';
            const displayHour = hour === 0 ? 12 : hour > 12 ? hour - 12 : hour;
            return `${displayHour}:${minute.toString().padStart(2, '0')} ${period}`;
        }

        _formatTime(date) {
            return this._formatTimeOption(date.getHours(), date.getMinutes());
        }
    }

    // -------------------------------------------------------------------------
    // Init
    // -------------------------------------------------------------------------
    function initDateTimePickers() {
        bindPanelListeners();

        document.querySelectorAll('[data-datetime-picker]').forEach(triggerEl => {
            new DateTimePicker(triggerEl);
        });
    }

    if (document.readyState === 'loading') {
        document.addEventListener('DOMContentLoaded', initDateTimePickers);
    } else {
        initDateTimePickers();
    }
})();
