/**
 * Tests for datetime picker toggle group visibility
 */

describe('DateTimePicker Toggle Group Visibility', () => {
    let container;
    let panel;
    let overlay;
    let toggleGroup;

    beforeEach(() => {
        container = document.createElement('div');
        container.innerHTML = `
            <button
                type="button"
                id="event_datetime_trigger"
                class="datetime-trigger-btn"
                data-datetime-picker
                data-mode="datetime-range"
                data-start-input="start_time"
                data-end-input="end_time"
                data-timezone-input="timezone"
                data-show-timezone="true"
                data-title="Select Event Date & Time"
            >
                <span class="datetime-trigger-text">Event DateTime</span>
            </button>
            <input type="hidden" id="start_time" value="">
            <input type="hidden" id="end_time" value="">
            <input type="hidden" id="timezone" value="America/Los_Angeles">

            <button
                type="button"
                id="rsvp_deadline_trigger"
                class="datetime-trigger-btn"
                data-datetime-picker
                data-mode="datetime-single"
                data-start-input="rsvp_deadline_input"
                data-show-timezone="false"
                data-title="Select RSVP Deadline"
            >
                <span class="datetime-trigger-text">RSVP Deadline</span>
            </button>
            <input type="hidden" id="rsvp_deadline_input" value="">

            <div class="datetime-picker-overlay"></div>
            <div class="datetime-picker-panel">
                <div class="datetime-picker-header">
                    <h2 class="datetime-picker-title">Select Date & Time</h2>
                    <button class="datetime-picker-close">×</button>
                </div>
                <div class="datetime-picker-body">
                    <div class="datetime-toggle-group">
                        <button type="button" class="datetime-toggle-btn active" data-mode="start">
                            <span class="datetime-toggle-label">Start Time</span>
                            <span class="datetime-toggle-value" id="start-time-display">Not set</span>
                        </button>
                        <button type="button" class="datetime-toggle-btn" data-mode="end">
                            <span class="datetime-toggle-label">End Time (Optional)</span>
                            <span class="datetime-toggle-value" id="end-time-display">Not set</span>
                        </button>
                    </div>
                    <div class="datetime-picker-content active" data-content="start">
                        <div class="datetime-picker-layout">
                            <div class="calendar-container">
                                <div class="calendar-header">
                                    <button type="button" class="calendar-nav-btn" data-nav="prev">‹</button>
                                    <span class="calendar-month-year"></span>
                                    <button type="button" class="calendar-nav-btn" data-nav="next">›</button>
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
                                    <button type="button" class="calendar-nav-btn" data-nav="prev">‹</button>
                                    <span class="calendar-month-year"></span>
                                    <button type="button" class="calendar-nav-btn" data-nav="next">›</button>
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
                <div class="datetime-picker-footer">
                    <button type="button" class="btn btn-secondary datetime-picker-cancel">Cancel</button>
                    <button type="button" class="btn btn-primary datetime-picker-save">Save</button>
                </div>
            </div>

            <div class="timezone-picker-overlay"></div>
            <div class="timezone-picker-panel datetime-picker-panel">
                <div class="datetime-picker-header">
                    <h2 class="datetime-picker-title">Select Timezone</h2>
                    <button class="timezone-picker-close datetime-picker-close">×</button>
                </div>
                <div class="datetime-picker-body">
                    <div class="timezone-list">
                        <div class="timezone-option" data-timezone="America/Los_Angeles">
                            <div class="timezone-option-name">Pacific Time (PT)</div>
                            <div class="timezone-option-offset">UTC-8</div>
                        </div>
                    </div>
                </div>
            </div>
        `;
        document.body.appendChild(container);

        panel = document.querySelector('.datetime-picker-panel');
        overlay = document.querySelector('.datetime-picker-overlay');
        toggleGroup = document.querySelector('.datetime-toggle-group');
    });

    afterEach(() => {
        if (container && container.parentNode) {
            container.parentNode.removeChild(container);
        }
    });

    test('toggle group should be hidden for datetime-single mode', (done) => {
        const rsvpTrigger = document.getElementById('rsvp_deadline_trigger');
        
        setTimeout(() => {
            rsvpTrigger.click();
            
            setTimeout(() => {
                expect(panel.classList.contains('open')).toBe(true);
                expect(toggleGroup.style.display).toBe('none');
                done();
            }, 100);
        }, 100);
    });

    test('toggle group should be visible for datetime-range mode', (done) => {
        const eventTrigger = document.getElementById('event_datetime_trigger');
        
        setTimeout(() => {
            eventTrigger.click();
            
            setTimeout(() => {
                expect(panel.classList.contains('open')).toBe(true);
                expect(toggleGroup.style.display).toBe('flex');
                done();
            }, 100);
        }, 100);
    });

    test('toggle group should switch correctly when opening different pickers', (done) => {
        const eventTrigger = document.getElementById('event_datetime_trigger');
        const rsvpTrigger = document.getElementById('rsvp_deadline_trigger');
        const closeBtn = document.querySelector('.datetime-picker-close');
        
        setTimeout(() => {
            eventTrigger.click();
            
            setTimeout(() => {
                expect(toggleGroup.style.display).toBe('flex');
                
                closeBtn.click();
                
                setTimeout(() => {
                    rsvpTrigger.click();
                    
                    setTimeout(() => {
                        expect(panel.classList.contains('open')).toBe(true);
                        expect(toggleGroup.style.display).toBe('none');
                        done();
                    }, 100);
                }, 100);
            }, 100);
        }, 100);
    });

    test('timezone display should be hidden when showTimezone is false', (done) => {
        const rsvpTrigger = document.getElementById('rsvp_deadline_trigger');
        
        setTimeout(() => {
            rsvpTrigger.click();
            
            setTimeout(() => {
                const timezoneDisplays = panel.querySelectorAll('.timezone-display');
                timezoneDisplays.forEach(display => {
                    expect(display.style.display).toBe('none');
                });
                done();
            }, 100);
        }, 100);
    });

    test('timezone display should be visible when showTimezone is true', (done) => {
        const eventTrigger = document.getElementById('event_datetime_trigger');
        
        setTimeout(() => {
            eventTrigger.click();
            
            setTimeout(() => {
                const timezoneDisplays = panel.querySelectorAll('.timezone-display');
                timezoneDisplays.forEach(display => {
                    expect(display.style.display).toBe('');
                });
                done();
            }, 100);
        }, 100);
    });

    test('single-date mode should not display end date in trigger text', (done) => {
        const rsvpTrigger = document.getElementById('rsvp_deadline_trigger');
        const rsvpInput = document.getElementById('rsvp_deadline_input');
        const saveBtn = document.querySelector('.datetime-picker-save');
        
        setTimeout(() => {
            rsvpTrigger.click();
            
            setTimeout(() => {
                const calendarDays = panel.querySelectorAll('.calendar-day:not(.other-month):not(.disabled)');
                if (calendarDays.length > 0) {
                    calendarDays[0].click();
                    
                    setTimeout(() => {
                        const timeOptions = panel.querySelectorAll('.time-option');
                        if (timeOptions.length > 0) {
                            timeOptions[0].click();
                            
                            setTimeout(() => {
                                saveBtn.click();
                                
                                setTimeout(() => {
                                    const triggerText = rsvpTrigger.querySelector('.datetime-trigger-text').textContent;
                                    expect(triggerText).not.toContain(' - ');
                                    expect(rsvpInput.value).toBeTruthy();
                                    done();
                                }, 100);
                            }, 100);
                        } else {
                            done();
                        }
                    }, 100);
                } else {
                    done();
                }
            }, 100);
        }, 100);
    });

    test('date-range mode should display both start and end dates in trigger text', (done) => {
        const eventTrigger = document.getElementById('event_datetime_trigger');
        const startInput = document.getElementById('start_time');
        const endInput = document.getElementById('end_time');
        const saveBtn = document.querySelector('.datetime-picker-save');
        const closeBtn = document.querySelector('.datetime-picker-close');
        
        setTimeout(() => {
            eventTrigger.click();
            
            setTimeout(() => {
                const startCalendarDays = panel.querySelectorAll('.calendar-day:not(.other-month):not(.disabled)');
                if (startCalendarDays.length > 1) {
                    startCalendarDays[0].click();
                    
                    setTimeout(() => {
                        const startTimeOptions = panel.querySelectorAll('.time-option');
                        if (startTimeOptions.length > 0) {
                            startTimeOptions[0].click();
                            
                            setTimeout(() => {
                                const endToggleBtn = panel.querySelector('.datetime-toggle-btn[data-mode="end"]');
                                endToggleBtn.click();
                                
                                setTimeout(() => {
                                    const endCalendarDays = panel.querySelectorAll('.calendar-day:not(.other-month):not(.disabled)');
                                    if (endCalendarDays.length > 1) {
                                        endCalendarDays[1].click();
                                        
                                        setTimeout(() => {
                                            const endTimeOptions = panel.querySelectorAll('.time-option');
                                            if (endTimeOptions.length > 0) {
                                                endTimeOptions[0].click();
                                                
                                                setTimeout(() => {
                                                    saveBtn.click();
                                                    
                                                    setTimeout(() => {
                                                        const triggerText = eventTrigger.querySelector('.datetime-trigger-text').textContent;
                                                        expect(triggerText).toContain(' - ');
                                                        expect(startInput.value).toBeTruthy();
                                                        expect(endInput.value).toBeTruthy();
                                                        done();
                                                    }, 100);
                                                }, 100);
                                            } else {
                                                done();
                                            }
                                        }, 100);
                                    } else {
                                        done();
                                    }
                                }, 100);
                            }, 100);
                        } else {
                            done();
                        }
                    }, 100);
                } else {
                    done();
                }
            }, 100);
        }, 100);
    });

    test('single-date mode should not show end date even after using date-range mode', (done) => {
        const eventTrigger = document.getElementById('event_datetime_trigger');
        const rsvpTrigger = document.getElementById('rsvp_deadline_trigger');
        const saveBtn = document.querySelector('.datetime-picker-save');
        const closeBtn = document.querySelector('.datetime-picker-close');
        
        setTimeout(() => {
            eventTrigger.click();
            
            setTimeout(() => {
                const startCalendarDays = panel.querySelectorAll('.calendar-day:not(.other-month):not(.disabled)');
                if (startCalendarDays.length > 1) {
                    startCalendarDays[0].click();
                    
                    setTimeout(() => {
                        const startTimeOptions = panel.querySelectorAll('.time-option');
                        if (startTimeOptions.length > 0) {
                            startTimeOptions[0].click();
                            
                            setTimeout(() => {
                                const endToggleBtn = panel.querySelector('.datetime-toggle-btn[data-mode="end"]');
                                endToggleBtn.click();
                                
                                setTimeout(() => {
                                    const endCalendarDays = panel.querySelectorAll('.calendar-day:not(.other-month):not(.disabled)');
                                    if (endCalendarDays.length > 1) {
                                        endCalendarDays[1].click();
                                        
                                        setTimeout(() => {
                                            const endTimeOptions = panel.querySelectorAll('.time-option');
                                            if (endTimeOptions.length > 0) {
                                                endTimeOptions[0].click();
                                                
                                                setTimeout(() => {
                                                    saveBtn.click();
                                                    
                                                    setTimeout(() => {
                                                        const eventTriggerText = eventTrigger.querySelector('.datetime-trigger-text').textContent;
                                                        expect(eventTriggerText).toContain(' - ');
                                                        
                                                        setTimeout(() => {
                                                            rsvpTrigger.click();
                                                            
                                                            setTimeout(() => {
                                                                const rsvpCalendarDays = panel.querySelectorAll('.calendar-day:not(.other-month):not(.disabled)');
                                                                if (rsvpCalendarDays.length > 0) {
                                                                    rsvpCalendarDays[0].click();
                                                                    
                                                                    setTimeout(() => {
                                                                        const rsvpTimeOptions = panel.querySelectorAll('.time-option');
                                                                        if (rsvpTimeOptions.length > 0) {
                                                                            rsvpTimeOptions[0].click();
                                                                            
                                                                            setTimeout(() => {
                                                                                saveBtn.click();
                                                                                
                                                                                setTimeout(() => {
                                                                                    const rsvpTriggerText = rsvpTrigger.querySelector('.datetime-trigger-text').textContent;
                                                                                    expect(rsvpTriggerText).not.toContain(' - ');
                                                                                    done();
                                                                                }, 100);
                                                                            }, 100);
                                                                        } else {
                                                                            done();
                                                                        }
                                                                    }, 100);
                                                                } else {
                                                                    done();
                                                                }
                                                            }, 100);
                                                        }, 100);
                                                    }, 100);
                                                }, 100);
                                            } else {
                                                done();
                                            }
                                        }, 100);
                                    } else {
                                        done();
                                    }
                                }, 100);
                            }, 100);
                        } else {
                            done();
                        }
                    }, 100);
                } else {
                    done();
                }
            }, 100);
        }, 100);
    });
});

// -----------------------------------------------------------------------------
// Regression tests for active-instance / shared-panel bugs
// -----------------------------------------------------------------------------
describe('DateTimePicker active-instance isolation', () => {
    let container;

    beforeEach(() => {
        container = document.createElement('div');
        container.innerHTML = `
            <button
                type="button"
                id="event_datetime_trigger"
                class="datetime-trigger-btn"
                data-datetime-picker
                data-mode="datetime-range"
                data-start-input="start_time"
                data-end-input="end_time"
                data-timezone-input="timezone"
                data-show-timezone="true"
                data-title="Select Event Date & Time"
            ><span class="datetime-trigger-text">Event DateTime</span></button>
            <input type="hidden" id="start_time" value="">
            <input type="hidden" id="end_time" value="">
            <input type="hidden" id="timezone" value="America/Los_Angeles">

            <button
                type="button"
                id="rsvp_deadline_trigger"
                class="datetime-trigger-btn"
                data-datetime-picker
                data-mode="datetime-single"
                data-start-input="rsvp_deadline_input"
                data-show-timezone="false"
                data-title="Select RSVP Deadline"
            ><span class="datetime-trigger-text">RSVP Deadline</span></button>
            <input type="hidden" id="rsvp_deadline_input" value="">

            <div class="datetime-picker-overlay"></div>
            <div class="datetime-picker-panel">
                <div class="datetime-picker-header">
                    <h2 class="datetime-picker-title">Select Date & Time</h2>
                    <button class="datetime-picker-close">×</button>
                </div>
                <div class="datetime-picker-body">
                    <div class="datetime-toggle-group">
                        <button type="button" class="datetime-toggle-btn active" data-mode="start">
                            <span class="datetime-toggle-label">Start Time</span>
                            <span class="datetime-toggle-value" id="start-time-display">Not set</span>
                        </button>
                        <button type="button" class="datetime-toggle-btn" data-mode="end">
                            <span class="datetime-toggle-label">End Time (Optional)</span>
                            <span class="datetime-toggle-value" id="end-time-display">Not set</span>
                        </button>
                    </div>
                    <div class="datetime-picker-content active" data-content="start">
                        <div class="datetime-picker-layout">
                            <div class="calendar-container">
                                <div class="calendar-header">
                                    <button type="button" class="calendar-nav-btn" data-nav="prev">‹</button>
                                    <span class="calendar-month-year"></span>
                                    <button type="button" class="calendar-nav-btn" data-nav="next">›</button>
                                </div>
                                <div class="calendar-grid"></div>
                            </div>
                            <div class="time-picker-container">
                                <label class="time-picker-label">Time</label>
                                <div class="time-picker-scroll"></div>
                            </div>
                        </div>
                        <div class="timezone-display">
                            <div><div class="timezone-label">Timezone</div><div class="timezone-value"></div></div>
                            <button type="button" class="timezone-change-btn">Change</button>
                        </div>
                    </div>
                    <div class="datetime-picker-content" data-content="end">
                        <div class="datetime-picker-layout">
                            <div class="calendar-container">
                                <div class="calendar-header">
                                    <button type="button" class="calendar-nav-btn" data-nav="prev">‹</button>
                                    <span class="calendar-month-year"></span>
                                    <button type="button" class="calendar-nav-btn" data-nav="next">›</button>
                                </div>
                                <div class="calendar-grid"></div>
                            </div>
                            <div class="time-picker-container">
                                <label class="time-picker-label">Time</label>
                                <div class="time-picker-scroll"></div>
                            </div>
                        </div>
                        <div class="timezone-display">
                            <div><div class="timezone-label">Timezone</div><div class="timezone-value"></div></div>
                            <button type="button" class="timezone-change-btn">Change</button>
                        </div>
                    </div>
                </div>
                <div class="datetime-picker-footer">
                    <button type="button" class="btn btn-secondary datetime-picker-cancel">Cancel</button>
                    <button type="button" class="btn btn-primary datetime-picker-save">Save</button>
                </div>
            </div>
            <div class="timezone-picker-overlay"></div>
            <div class="timezone-picker-panel datetime-picker-panel">
                <div class="datetime-picker-header">
                    <h2 class="datetime-picker-title">Select Timezone</h2>
                    <button class="timezone-picker-close datetime-picker-close">×</button>
                </div>
                <div class="datetime-picker-body">
                    <div class="timezone-list">
                        <div class="timezone-option" data-timezone="America/Los_Angeles">
                            <div class="timezone-option-name">Pacific Time (PT)</div>
                        </div>
                    </div>
                </div>
            </div>
        `;
        document.body.appendChild(container);
    });

    afterEach(() => {
        if (container && container.parentNode) {
            container.parentNode.removeChild(container);
        }
    });

    test('saving RSVP deadline writes to rsvp_deadline_input, not start_time', (done) => {
        const rsvpTrigger   = document.getElementById('rsvp_deadline_trigger');
        const rsvpInput     = document.getElementById('rsvp_deadline_input');
        const startInput    = document.getElementById('start_time');
        const saveBtn       = document.querySelector('.datetime-picker-save');

        setTimeout(() => {
            rsvpTrigger.click();

            setTimeout(() => {
                const calendarDays = document.querySelectorAll('.calendar-day:not(.other-month):not(.disabled)');
                if (calendarDays.length === 0) { done(); return; }
                calendarDays[0].click();

                setTimeout(() => {
                    const timeOptions = document.querySelectorAll('.time-option');
                    if (timeOptions.length === 0) { done(); return; }
                    timeOptions[0].click();

                    setTimeout(() => {
                        saveBtn.click();

                        setTimeout(() => {
                            // Deadline input must be populated
                            expect(rsvpInput.value).toBeTruthy();
                            // start_time must NOT have been touched
                            expect(startInput.value).toBe('');
                            done();
                        }, 100);
                    }, 100);
                }, 100);
            }, 100);
        }, 100);
    });

    test('saving event datetime does not overwrite rsvp_deadline_input', (done) => {
        const eventTrigger  = document.getElementById('event_datetime_trigger');
        const startInput    = document.getElementById('start_time');
        const rsvpInput     = document.getElementById('rsvp_deadline_input');
        const saveBtn       = document.querySelector('.datetime-picker-save');

        setTimeout(() => {
            eventTrigger.click();

            setTimeout(() => {
                const calendarDays = document.querySelectorAll('.calendar-day:not(.other-month):not(.disabled)');
                if (calendarDays.length === 0) { done(); return; }
                calendarDays[0].click();

                setTimeout(() => {
                    const timeOptions = document.querySelectorAll('.time-option');
                    if (timeOptions.length === 0) { done(); return; }
                    timeOptions[0].click();

                    setTimeout(() => {
                        saveBtn.click();

                        setTimeout(() => {
                            // start_time must be populated
                            expect(startInput.value).toBeTruthy();
                            // rsvp_deadline_input must NOT have been touched
                            expect(rsvpInput.value).toBe('');
                            done();
                        }, 100);
                    }, 100);
                }, 100);
            }, 100);
        }, 100);
    });

    test('save button fires only once per click regardless of picker count', (done) => {
        const rsvpTrigger = document.getElementById('rsvp_deadline_trigger');
        const rsvpInput   = document.getElementById('rsvp_deadline_input');
        const saveBtn     = document.querySelector('.datetime-picker-save');
        let saveCallCount = 0;

        // Intercept value assignments to count actual writes
        const origDescriptor = Object.getOwnPropertyDescriptor(HTMLInputElement.prototype, 'value');
        const orig = origDescriptor.set;
        Object.defineProperty(rsvpInput, 'value', {
            set(v) { if (v) saveCallCount++; orig.call(this, v); },
            get() { return orig ? rsvpInput.getAttribute('value') || '' : ''; },
            configurable: true,
        });

        setTimeout(() => {
            rsvpTrigger.click();

            setTimeout(() => {
                const calendarDays = document.querySelectorAll('.calendar-day:not(.other-month):not(.disabled)');
                if (calendarDays.length === 0) { done(); return; }
                calendarDays[0].click();

                setTimeout(() => {
                    const timeOptions = document.querySelectorAll('.time-option');
                    if (timeOptions.length === 0) { done(); return; }
                    timeOptions[0].click();

                    setTimeout(() => {
                        saveCallCount = 0; // reset before the save
                        saveBtn.click();

                        setTimeout(() => {
                            expect(saveCallCount).toBe(1);
                            done();
                        }, 100);
                    }, 100);
                }, 100);
            }, 100);
        }, 100);
    });
});
