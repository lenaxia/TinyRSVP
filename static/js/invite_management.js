(function() {
    'use strict';

    let importModal = null;
    let createModal = null;
    let currentEventId = null;

    function getEventIdFromURL() {
        const match = window.location.pathname.match(/\/events\/(\d+)/);
        return match ? match[1] : null;
    }

    function getCookie(name) {
        const value = `; ${document.cookie}`;
        const parts = value.split(`; ${name}=`);
        if (parts.length === 2) {
            return parts.pop().split(';').shift();
        }
        return '';
    }

    function getCSRFToken() {
        return getCookie('csrf_token');
    }

    function showFeedback(message, type = 'success') {
        let feedback = document.getElementById('feedback');
        if (!feedback) {
            feedback = document.createElement('div');
            feedback.id = 'feedback';
            feedback.className = 'feedback';
            document.body.appendChild(feedback);
        }

        feedback.textContent = message;
        feedback.className = `feedback feedback-${type}`;
        feedback.style.display = 'block';

        setTimeout(() => {
            feedback.style.display = 'none';
        }, 5000);
    }

    function handleImportSubmit(e) {
        e.preventDefault();

        const form = e.target;
        const fileInput = form.querySelector('input[type="file"]');
        const submitBtn = form.querySelector('button[type="submit"]');

        if (!fileInput.files || fileInput.files.length === 0) {
            showFeedback('Please select a CSV file to import', 'error');
            return;
        }

        const file = fileInput.files[0];
        if (!file.name.endsWith('.csv')) {
            showFeedback('Please select a valid CSV file', 'error');
            return;
        }

        const formData = new FormData();
        formData.append('file', file);

        submitBtn.disabled = true;
        submitBtn.textContent = 'Importing...';

        fetch(`/api/events/${currentEventId}/invites/import`, {
            method: 'POST',
            headers: {
                'X-CSRF-Token': getCSRFToken(),
            },
            body: formData,
        })
        .then(response => {
            if (!response.ok) {
                return response.json().then(data => {
                    throw new Error(data.error || 'Import failed');
                });
            }
            return response.json();
        })
        .then(data => {
            showFeedback(`Successfully imported ${data.imported || 0} invites`, 'success');
            form.reset();
            if (importModal) {
                importModal.close();
            }
            setTimeout(() => {
                window.location.reload();
            }, 1000);
        })
        .catch(error => {
            showFeedback(error.message || 'Failed to import invites', 'error');
        })
        .finally(() => {
            submitBtn.disabled = false;
            submitBtn.textContent = 'Import';
        });
    }

    function handleCreateSubmit(e) {
        e.preventDefault();

        const form = e.target;
        const nameInput = form.querySelector('input[name="name"]');
        const maxPlusOnesInput = form.querySelector('input[name="max_plus_ones"]');
        const submitBtn = form.querySelector('button[type="submit"]');

        const name = nameInput.value.trim();
        const maxPlusOnes = parseInt(maxPlusOnesInput.value, 10) || 0;

        if (maxPlusOnes < 0 || maxPlusOnes > 10) {
            showFeedback('Plus ones must be between 0 and 10', 'error');
            return;
        }

        const requestBody = {
            max_plus_ones: maxPlusOnes,
        };

        if (name) {
            requestBody.name = name;
        }

        submitBtn.disabled = true;
        submitBtn.textContent = 'Creating...';

        fetch(`/api/events/${currentEventId}/invites/manual`, {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json',
                'X-CSRF-Token': getCSRFToken(),
            },
            body: JSON.stringify(requestBody),
        })
        .then(response => {
            if (!response.ok) {
                return response.json().then(data => {
                    throw new Error(data.error || 'Creation failed');
                });
            }
            return response.json();
        })
        .then(data => {
            const guestName = data.invite?.name || 'Guest';
            showFeedback(`Successfully created invite for ${guestName}`, 'success');
            form.reset();
            if (createModal) {
                createModal.close();
            }
            setTimeout(() => {
                window.location.reload();
            }, 1000);
        })
        .catch(error => {
            showFeedback(error.message || 'Failed to create invite', 'error');
        })
        .finally(() => {
            submitBtn.disabled = false;
            submitBtn.textContent = 'Create';
        });
    }

    function initInviteManagement() {
        currentEventId = getEventIdFromURL();
        if (!currentEventId) {
            return;
        }

        const importModalEl = document.getElementById('import-modal');
        const createModalEl = document.getElementById('create-modal');

        if (importModalEl) {
            importModal = new window.Modal('import-modal');
            const importForm = document.getElementById('import-form');
            if (importForm) {
                importForm.addEventListener('submit', handleImportSubmit);
            }
        }

        if (createModalEl) {
            createModal = new window.Modal('create-modal');
            const createForm = document.getElementById('create-form');
            if (createForm) {
                createForm.addEventListener('submit', handleCreateSubmit);
            }
        }
    }

    if (document.readyState === 'loading') {
        document.addEventListener('DOMContentLoaded', initInviteManagement);
    } else {
        initInviteManagement();
    }

    window.InviteManagement = {
        init: initInviteManagement,
        handleImportSubmit: handleImportSubmit,
        handleCreateSubmit: handleCreateSubmit,
    };
})();
