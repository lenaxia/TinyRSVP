class ThemeController {
    constructor() {
        this.STORAGE_KEY = 'tinyrsvp-theme';
        this.THEMES = { LIGHT: 'light', DARK: 'dark' };
        this.init();
    }

    init() {
        const savedTheme = this.getSavedTheme();
        const systemTheme = this.getSystemTheme();
        const theme = savedTheme || systemTheme || this.THEMES.LIGHT;
        this.setTheme(theme);
        this.attachEventListeners();
    }

    getSavedTheme() {
        return localStorage.getItem(this.STORAGE_KEY);
    }

    getSystemTheme() {
        if (window.matchMedia('(prefers-color-scheme: dark)').matches) {
            return this.THEMES.DARK;
        }
        return this.THEMES.LIGHT;
    }

    setTheme(theme) {
        document.documentElement.setAttribute('data-theme', theme);
        localStorage.setItem(this.STORAGE_KEY, theme);
        this.updateToggleButton(theme);
    }

    toggleTheme() {
        const currentTheme = document.documentElement.getAttribute('data-theme');
        const newTheme = currentTheme === this.THEMES.DARK 
            ? this.THEMES.LIGHT 
            : this.THEMES.DARK;
        this.setTheme(newTheme);
    }

    updateToggleButton(theme) {
        const button = document.getElementById('theme-toggle');
        if (!button) return;
        
        const icon = button.querySelector('.theme-icon');
        const label = button.querySelector('.sr-only');
        
        if (theme === this.THEMES.DARK) {
            icon.textContent = '☀️';
            label.textContent = 'Switch to light mode';
            button.setAttribute('aria-label', 'Switch to light mode');
        } else {
            icon.textContent = '🌙';
            label.textContent = 'Switch to dark mode';
            button.setAttribute('aria-label', 'Switch to dark mode');
        }
    }

    attachEventListeners() {
        const button = document.getElementById('theme-toggle');
        if (button) {
            button.addEventListener('click', () => this.toggleTheme());
        }
    }
}

if (document.readyState === 'loading') {
    document.addEventListener('DOMContentLoaded', () => new ThemeController());
} else {
    new ThemeController();
}
