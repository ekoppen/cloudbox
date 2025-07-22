import { writable } from 'svelte/store';
import { browser } from '$app/environment';
import { API_ENDPOINTS, createApiRequest } from '$lib/config';

export interface User {
  id: number;
  email: string;
  name: string;
  created_at: string;
}

export interface AuthState {
  user: User | null;
  token: string | null;
  refreshToken: string | null;
  isLoading: boolean;
  isAuthenticated: boolean;
}

// Initial state
const initialState: AuthState = {
  user: null,
  token: null,
  refreshToken: null,
  isLoading: false,
  isAuthenticated: false,
};

// Create the auth store
function createAuthStore() {
  const { subscribe, set, update } = writable<AuthState>(initialState);

  return {
    subscribe,
    
    // Initialize auth state from localStorage
    async init() {
      console.log('Auth store: initializing...');
      
      if (browser) {
        const token = localStorage.getItem('cloudbox_token');
        const refreshToken = localStorage.getItem('cloudbox_refresh_token');
        const userStr = localStorage.getItem('cloudbox_user');
        
        console.log('Auth store: found token:', !!token);
        console.log('Auth store: found refresh token:', !!refreshToken);
        console.log('Auth store: found user:', !!userStr);
        
        if (token && userStr) {
          try {
            const user = JSON.parse(userStr);
            console.log('Auth store: restoring session for user:', user.email);
            
            set({
              user,
              token,
              refreshToken,
              isLoading: false,
              isAuthenticated: true,
            });
            
            console.log('Auth store: session restored successfully');
          } catch (err) {
            console.error('Failed to parse user data:', err);
            this.logout();
          }
        } else if (refreshToken && userStr) {
          // Try to refresh token if we have refresh token but no access token
          console.log('Auth store: attempting token refresh...');
          try {
            await this.refreshAccessToken();
          } catch (err) {
            console.error('Failed to refresh token:', err);
            this.logout();
          }
        } else {
          console.log('Auth store: no saved session found');
          set({
            user: null,
            token: null,
            refreshToken: null,
            isLoading: false,
            isAuthenticated: false,
          });
        }
      }
      
      // Small delay to ensure state is fully updated
      await new Promise(resolve => setTimeout(resolve, 10));
      console.log('Auth store: initialization complete');
    },

    // Login user
    login(user: User, token: string, refreshToken?: string) {
      console.log('Auth store: logging in user', user.email);
      
      if (browser) {
        localStorage.setItem('cloudbox_token', token);
        localStorage.setItem('cloudbox_user', JSON.stringify(user));
        if (refreshToken) {
          localStorage.setItem('cloudbox_refresh_token', refreshToken);
        }
        console.log('Auth store: saved to localStorage');
      }
      
      set({
        user,
        token,
        refreshToken: refreshToken || null,
        isLoading: false,
        isAuthenticated: true,
      });
    },

    // Logout user
    async logout() {
      const currentState = get(auth);
      
      // Send logout request to invalidate refresh token
      if (currentState.refreshToken) {
        try {
          await createApiRequest(API_ENDPOINTS.auth.logout, {
            method: 'POST',
            headers: {
              ...(currentState.token && { Authorization: `Bearer ${currentState.token}` })
            },
            body: JSON.stringify({ refresh_token: currentState.refreshToken }),
          });
        } catch (err) {
          console.error('Failed to logout on server:', err);
        }
      }
      
      if (browser) {
        localStorage.removeItem('cloudbox_token');
        localStorage.removeItem('cloudbox_refresh_token');
        localStorage.removeItem('cloudbox_user');
      }
      
      set({
        user: null,
        token: null,
        refreshToken: null,
        isLoading: false,
        isAuthenticated: false,
      });
    },

    // Set loading state
    setLoading(loading: boolean) {
      update(state => ({ ...state, isLoading: loading }));
    },

    // Refresh access token using refresh token
    async refreshAccessToken(): Promise<void> {
      const currentState = get(auth);
      
      if (!currentState.refreshToken) {
        throw new Error('No refresh token available');
      }
      
      const response = await createApiRequest(API_ENDPOINTS.auth.refresh, {
        method: 'POST',
        body: JSON.stringify({ refresh_token: currentState.refreshToken }),
      });
      
      if (!response.ok) {
        throw new Error('Failed to refresh token');
      }
      
      const data = await response.json();
      
      // Update state with new token
      update(state => ({
        ...state,
        user: data.user,
        token: data.token,
      }));
      
      // Update localStorage
      if (browser) {
        localStorage.setItem('cloudbox_token', data.token);
        localStorage.setItem('cloudbox_user', JSON.stringify(data.user));
      }
    },

    // Get authorization header
    getAuthHeader(): Record<string, string> {
      const state = get(auth);
      return state.token ? { Authorization: `Bearer ${state.token}` } : {};
    },
  };
}

export const auth = createAuthStore();

// Helper function to get current state
function get<T>(store: { subscribe: (fn: (value: T) => void) => () => void }): T {
  let value: T;
  store.subscribe(v => value = v)();
  return value!;
}