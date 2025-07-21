import { writable } from 'svelte/store';
import { browser } from '$app/environment';

export type Theme = 'light' | 'dark';
export type AccentColor = 'blue' | 'green' | 'purple' | 'orange' | 'red' | 'pink';

export interface ThemeState {
  theme: Theme;
  accentColor: AccentColor;
}

// Initial state
const initialState: ThemeState = {
  theme: 'dark',
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
        
        const finalTheme = savedTheme || initialState.theme;
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
        const newTheme = state.theme === 'light' ? 'dark' : 'light';
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
        
        // Remove existing theme classes from both html and body
        [root, body].forEach(element => {
          if (element) {
            element.classList.remove('light', 'dark');
            element.classList.remove('accent-blue', 'accent-green', 'accent-purple', 'accent-orange', 'accent-red', 'accent-pink');
          }
        });
        
        // Add new theme classes to both html and body
        [root, body].forEach(element => {
          if (element) {
            element.classList.add(theme);
            element.classList.add(`accent-${accentColor}`);
          }
        });
        
        // Force style recalculation
        root.style.setProperty('color-scheme', theme);
        
        console.log('Theme store: applied classes to html:', root.classList.toString());
        console.log('Theme store: applied classes to body:', body?.classList.toString());
        
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