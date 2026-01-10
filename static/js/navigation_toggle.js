(function() {
    'use strict';

    const toggle = document.querySelector('.app-nav-toggle');
    const menu = document.querySelector('.app-nav-menu');
    
    if (!toggle || !menu) {
        return;
    }

    const STORAGE_KEY = 'app-nav-menu-state';

    function openMenu() {
        menu.classList.add('open');
        toggle.setAttribute('aria-expanded', 'true');
        try {
            localStorage.setItem(STORAGE_KEY, 'open');
        } catch (e) {
            // Ignore localStorage errors
        }
    }

    function closeMenu() {
        menu.classList.remove('open');
        toggle.setAttribute('aria-expanded', 'false');
        try {
            localStorage.setItem(STORAGE_KEY, 'closed');
        } catch (e) {
            // Ignore localStorage errors
        }
    }

    function toggleMenu() {
        if (menu.classList.contains('open')) {
            closeMenu();
        } else {
            openMenu();
        }
    }

    toggle.addEventListener('click', toggleMenu);

    toggle.addEventListener('keydown', function(e) {
        if (e.key === 'Enter' || e.key === ' ') {
            e.preventDefault();
            toggleMenu();
        }
    });

    function handleResize() {
        if (window.innerWidth >= 768) {
            menu.classList.remove('open');
            toggle.setAttribute('aria-expanded', 'false');
        } else {
            try {
                const savedState = localStorage.getItem(STORAGE_KEY);
                if (savedState === 'open') {
                    openMenu();
                }
            } catch (e) {
                // Ignore localStorage errors
            }
        }
    }

    window.addEventListener('resize', handleResize);
    handleResize();

    document.addEventListener('click', function(e) {
        if (window.innerWidth < 768 && 
            menu.classList.contains('open') && 
            !menu.contains(e.target) && 
            !toggle.contains(e.target)) {
            closeMenu();
        }
    });
})();
