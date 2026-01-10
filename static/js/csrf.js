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

function addCSRFHeader(headers) {
    const token = getCSRFToken();
    if (token) {
        headers['X-CSRF-Token'] = token;
    }
    return headers;
}

function fetchWithCSRF(url, options = {}) {
    options.headers = options.headers || {};
    options.headers = addCSRFHeader(options.headers);
    return fetch(url, options);
}

function syncCSRFTokenInForms() {
    const currentToken = getCSRFToken();
    if (!currentToken) {
        return;
    }

    const csrfInputs = document.querySelectorAll('input[name="csrf_token"]');
    csrfInputs.forEach(input => {
        if (input.value !== currentToken) {
            input.value = currentToken;
        }
    });
}

if (typeof document !== 'undefined') {
    document.addEventListener('DOMContentLoaded', () => {
        syncCSRFTokenInForms();
    });

    window.addEventListener('pageshow', (event) => {
        if (event.persisted) {
            syncCSRFTokenInForms();
        }
    });
}
