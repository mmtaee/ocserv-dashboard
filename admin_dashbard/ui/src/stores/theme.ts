import { defineStore } from 'pinia';
import vuetify, { DARK_THEME, LIGHT_THEME, THEME_STORAGE_KEY } from '@/plugins/vuetify';

type ThemeName = typeof LIGHT_THEME | typeof DARK_THEME;

interface ThemeState {
    current: ThemeName;
}

function readStoredTheme(): ThemeName {
    const stored = localStorage.getItem(THEME_STORAGE_KEY);
    if (stored === LIGHT_THEME || stored === DARK_THEME) {
        return stored;
    }
    if (typeof window !== 'undefined' && window.matchMedia?.('(prefers-color-scheme: dark)').matches) {
        return DARK_THEME;
    }
    return LIGHT_THEME;
}

function writeVuetifyTheme(name: ThemeName) {
    // Access the vuetify singleton directly so this works from any context
    // (event handlers, lifecycle hooks, etc.) without relying on inject().
    vuetify.theme.global.name.value = name;
    if (typeof document !== 'undefined') {
        // Mirror the active theme on <html data-theme="..."> so non-Vue styles
        // (preloader, raw CSS) can react to the same toggle.
        document.documentElement.setAttribute('data-theme', name);
    }
}

export const useThemeStore = defineStore('theme', {
    state: (): ThemeState => ({
        current: readStoredTheme()
    }),
    getters: {
        isDark: (state): boolean => state.current === DARK_THEME
    },
    actions: {
        apply(name: ThemeName) {
            this.current = name;
            localStorage.setItem(THEME_STORAGE_KEY, name);
            writeVuetifyTheme(name);
        },
        toggle() {
            this.apply(this.current === DARK_THEME ? LIGHT_THEME : DARK_THEME);
        },
        sync() {
            writeVuetifyTheme(this.current);
        }
    }
});
