(function() {
    'use strict';

    const toggle = document.querySelector('.app-nav-toggle');
    const menu = document.querySelector('.app-nav-menu');
    const overlay = document.querySelector('.app-nav-overlay');
    const closeBtn = document.querySelector('.app-nav-menu-close');
    
    if (!toggle || !menu || !overlay) {
        return;
    }

    const STORAGE_KEY = 'app-nav-menu-state';

    function openMenu() {
        menu.classList.add('open');
        overlay.classList.add('open');
        toggle.setAttribute('aria-expanded', 'true');
        document.body.style.overflow = 'hidden';
        try {
            localStorage.setItem(STORAGE_KEY, 'open');
        } catch (e) {
            // Ignore localStorage errors
        }
    }

    function closeMenu() {
        menu.classList.remove('open');
        overlay.classList.remove('open');
        toggle.setAttribute('aria-expanded', 'false');
        document.body.style.overflow = '';
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

    if (closeBtn) {
        closeBtn.addEventListener('click', closeMenu);
        
        closeBtn.addEventListener('keydown', function(e) {
            if (e.key === 'Enter' || e.key === ' ') {
                e.preventDefault();
                closeMenu();
            }
        });
    }

    overlay.addEventListener('click', closeMenu);

    function handleResize() {
        if (window.innerWidth >= 768) {
            menu.classList.remove('open');
            overlay.classList.remove('open');
            toggle.setAttribute('aria-expanded', 'false');
            document.body.style.overflow = '';
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

    document.addEventListener('keydown', function(e) {
        if (e.key === 'Escape' && menu.classList.contains('open')) {
            closeMenu();
        }
    });
})();
