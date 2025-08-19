import { writable } from 'svelte/store';
import { browser } from '$app/environment';

export type Theme = 'cloudbox' | 'cloudbox-dark';
export type AccentColor = 'blue' | 'green' | 'purple';

export interface ThemeState {
  theme: Theme;
  accentColor: AccentColor;
}

// Initial state
const initialState: ThemeState = {
  theme: 'cloudbox',
  accentColor: 'blue',
};

// Create the theme store
function createThemeStore() {
  const { subscribe, set, update } = writable<ThemeState>(initialState);

  return {
    subscribe,
    
    // Initialize theme from localStorage
    async init() {
      if (browser) {
        const savedTheme = localStorage.getItem('cloudbox_theme') as Theme;
        const savedAccentColor = localStorage.getItem('cloudbox_accent_color') as AccentColor;
        
        // Handle legacy theme values
        let finalTheme = savedTheme || initialState.theme;
        if (finalTheme === 'light' as any) finalTheme = 'cloudbox';
        if (finalTheme === 'dark' as any) finalTheme = 'cloudbox-dark';
        
        const finalAccentColor = savedAccentColor || initialState.accentColor;
        
        console.log('Theme store: initializing with theme', finalTheme, 'and accent', finalAccentColor);
        
        update(state => ({
          theme: finalTheme,
          accentColor: finalAccentColor,
        }));
        
        // Apply theme to DOM
        this.applyTheme(finalTheme, finalAccentColor);
      }
    },

    // Toggle between light and dark theme
    toggleTheme() {
      const store = this;
      update(state => {
        const newTheme = state.theme === 'cloudbox' ? 'cloudbox-dark' : 'cloudbox';
        store.applyTheme(newTheme, state.accentColor);
        
        if (browser) {
          localStorage.setItem('cloudbox_theme', newTheme);
        }
        
        return { ...state, theme: newTheme };
      });
    },

    // Set specific theme
    setTheme(theme: Theme) {
      const store = this;
      update(state => {
        store.applyTheme(theme, state.accentColor);
        
        if (browser) {
          localStorage.setItem('cloudbox_theme', theme);
        }
        
        return { ...state, theme };
      });
    },

    // Set accent color
    setAccentColor(accentColor: AccentColor) {
      const store = this;
      update(state => {
        store.applyTheme(state.theme, accentColor);
        
        if (browser) {
          localStorage.setItem('cloudbox_accent_color', accentColor);
        }
        
        return { ...state, accentColor };
      });
    },

    // Apply theme to DOM
    applyTheme(theme: Theme, accentColor: AccentColor) {
      if (browser) {
        const root = document.documentElement;
        const body = document.body;
        
        console.log('Theme store: applying theme', theme, 'with accent', accentColor);
        
        // Remove existing theme classes and data attributes
        [root, body].forEach(element => {
          if (element) {
            element.classList.remove('light', 'dark', 'cloudbox', 'cloudbox-dark');
            element.classList.remove('accent-blue', 'accent-green', 'accent-purple');
            element.removeAttribute('data-theme');
          }
        });
        
        // Set DaisyUI data-theme attribute
        root.setAttribute('data-theme', theme);
        
        // Add theme classes for backwards compatibility and accent colors
        const isDark = theme === 'cloudbox-dark';
        [root, body].forEach(element => {
          if (element) {
            element.classList.add(theme);
            element.classList.add(isDark ? 'dark' : 'light');
            element.classList.add(`accent-${accentColor}`);
          }
        });
        
        // Set color-scheme for browser UI
        root.style.setProperty('color-scheme', isDark ? 'dark' : 'light');
        
        console.log('Theme store: applied theme', theme, 'to html with classes:', root.classList.toString());
        console.log('Theme store: data-theme attribute:', root.getAttribute('data-theme'));
        
        // Verify CSS variables are working
        const computedStyle = getComputedStyle(root);
        console.log('Theme store: CSS variables after change:', {
          background: computedStyle.getPropertyValue('--background'),
          primary: computedStyle.getPropertyValue('--primary')
        });
      }
    },
  };
}

export const theme = createThemeStore();