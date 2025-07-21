import { writable } from 'svelte/store';
import { browser } from '$app/environment';
import { goto } from '$app/navigation';

export interface NavigationState {
  history: string[];
  currentIndex: number;
  canGoBack: boolean;
  canGoForward: boolean;
}

// Initial state
const initialState: NavigationState = {
  history: [],
  currentIndex: -1,
  canGoBack: false,
  canGoForward: false,
};

// Create the navigation store
function createNavigationStore() {
  const { subscribe, set, update } = writable<NavigationState>(initialState);

  return {
    subscribe,
    
    // Initialize navigation from localStorage
    init() {
      if (browser) {
        const savedHistory = localStorage.getItem('cloudbox_nav_history');
        const savedIndex = localStorage.getItem('cloudbox_nav_index');
        
        if (savedHistory && savedIndex) {
          try {
            const history = JSON.parse(savedHistory);
            const currentIndex = parseInt(savedIndex, 10);
            
            set({
              history,
              currentIndex,
              canGoBack: currentIndex > 0,
              canGoForward: currentIndex < history.length - 1,
            });
          } catch (err) {
            console.error('Failed to parse navigation data:', err);
            this.clear();
          }
        }
      }
    },

    // Add new page to history
    navigate(path: string) {
      if (browser) {
        update(state => {
          // Remove any forward history when navigating to a new page
          const newHistory = [...state.history.slice(0, state.currentIndex + 1), path];
          const newIndex = newHistory.length - 1;
          
          const newState = {
            history: newHistory,
            currentIndex: newIndex,
            canGoBack: newIndex > 0,
            canGoForward: false,
          };
          
          // Save to localStorage
          localStorage.setItem('cloudbox_nav_history', JSON.stringify(newHistory));
          localStorage.setItem('cloudbox_nav_index', newIndex.toString());
          
          return newState;
        });
      }
    },

    // Go back in history
    goBack() {
      let targetPath = '';
      
      update(state => {
        if (state.canGoBack) {
          const newIndex = state.currentIndex - 1;
          targetPath = state.history[newIndex];
          
          const newState = {
            ...state,
            currentIndex: newIndex,
            canGoBack: newIndex > 0,
            canGoForward: true,
          };
          
          if (browser) {
            localStorage.setItem('cloudbox_nav_index', newIndex.toString());
          }
          
          return newState;
        }
        return state;
      });
      
      if (targetPath && browser) {
        goto(targetPath);
      }
    },

    // Go forward in history
    goForward() {
      let targetPath = '';
      
      update(state => {
        if (state.canGoForward) {
          const newIndex = state.currentIndex + 1;
          targetPath = state.history[newIndex];
          
          const newState = {
            ...state,
            currentIndex: newIndex,
            canGoBack: true,
            canGoForward: newIndex < state.history.length - 1,
          };
          
          if (browser) {
            localStorage.setItem('cloudbox_nav_index', newIndex.toString());
          }
          
          return newState;
        }
        return state;
      });
      
      if (targetPath && browser) {
        goto(targetPath);
      }
    },

    // Clear navigation history
    clear() {
      if (browser) {
        localStorage.removeItem('cloudbox_nav_history');
        localStorage.removeItem('cloudbox_nav_index');
      }
      
      set(initialState);
    },

    // Get current path
    getCurrentPath() {
      let currentPath = '';
      const unsubscribe = subscribe(state => {
        if (state.currentIndex >= 0 && state.history[state.currentIndex]) {
          currentPath = state.history[state.currentIndex];
        }
      });
      unsubscribe();
      return currentPath;
    },
  };
}

export const navigation = createNavigationStore();

// Initialize on app start
if (browser) {
  navigation.init();
}