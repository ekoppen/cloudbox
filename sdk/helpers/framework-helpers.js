/**
 * CloudBox Framework-Specific Helper Functions
 * 
 * Provides enhanced error handling, automatic troubleshooting,
 * and framework-specific optimizations for all major JavaScript frameworks.
 */

/**
 * CORS Error Detection and User Guidance System
 */
export class CORSErrorHandler {
    static detectCORSError(error) {
        const corsPatterns = [
            /Access to fetch.*blocked by CORS/,
            /Cross-Origin Request Blocked/,
            /No 'Access-Control-Allow-Origin' header/,
            /CORS error/,
            /Failed to fetch/,
            /Network Error/,
            /TypeError.*Failed to fetch/
        ];

        const message = error.message || error.toString();
        if (corsPatterns.some(pattern => pattern.test(message))) {
            return {
                type: 'cors',
                origin: typeof window !== 'undefined' ? window.location.origin : 'unknown',
                endpoint: error.config?.url || 'unknown',
                suggestions: this.generateCORSSuggestions(error),
                quickFixes: this.generateQuickFixes(error)
            };
        }

        return null;
    }

    static generateCORSSuggestions(error) {
        const suggestions = [];
        const origin = typeof window !== 'undefined' ? window.location.origin : 'localhost:3000';
        
        suggestions.push('🔧 **CORS Configuration Issue Detected**');
        suggestions.push(`Origin: ${origin}`);
        suggestions.push('');
        
        suggestions.push('💡 **Quick Fix Options:**');
        suggestions.push(`1. Run: \`node scripts/setup-universal-cors.js --origin="${origin}"\``);
        suggestions.push('2. Add to CloudBox .env: `CORS_ORIGINS=' + origin + ',http://localhost:*`');
        suggestions.push('3. Restart CloudBox backend after .env changes');
        suggestions.push('');
        
        if (origin.includes('localhost')) {
            suggestions.push('🏠 **Development Setup:**');
            suggestions.push('For localhost development, add wildcard support:');
            suggestions.push('`CORS_ORIGINS=http://localhost:*,https://localhost:*`');
            suggestions.push('');
        }
        
        suggestions.push('📚 **More Help:**');
        suggestions.push('- Documentation: https://docs.cloudbox.dev/cors-setup');
        suggestions.push('- Discord Support: https://discord.gg/cloudbox');
        
        return suggestions;
    }

    static generateQuickFixes(error) {
        const origin = typeof window !== 'undefined' ? window.location.origin : 'localhost:3000';
        
        return [
            {
                title: 'Auto-fix CORS (Recommended)',
                command: `node scripts/setup-universal-cors.js --origin="${origin}" --create-helpers`,
                description: 'Automatically configure CORS and create framework helpers'
            },
            {
                title: 'Manual Environment Fix',
                command: `echo "CORS_ORIGINS=${origin},http://localhost:*" >> .env`,
                description: 'Add wildcard localhost support to CloudBox'
            },
            {
                title: 'Restart Backend',
                command: 'docker-compose restart backend',
                description: 'Restart CloudBox backend to apply new CORS settings'
            }
        ];
    }

    static createUserFriendlyError(corsInfo) {
        const message = `
🚫 CloudBox Connection Blocked

Your app (${corsInfo.origin}) cannot connect to CloudBox API.

${corsInfo.suggestions.join('\n')}

Quick fixes available:
${corsInfo.quickFixes.map(fix => `• ${fix.title}: ${fix.command}`).join('\n')}
        `;
        
        const error = new Error(message);
        error.name = 'CORSConfigurationError';
        error.corsInfo = corsInfo;
        return error;
    }
}

/**
 * Framework-Specific Client Configurations
 */
export class FrameworkClientFactory {
    static createReactClient(config = {}) {
        const { CloudBoxClient } = require('@ekoppen/cloudbox-sdk');
        
        const client = new CloudBoxClient({
            projectId: process.env.REACT_APP_PROJECT_ID || config.projectId,
            apiKey: process.env.REACT_APP_API_KEY || config.apiKey,
            endpoint: process.env.REACT_APP_CLOUDBOX_ENDPOINT || config.endpoint || 'http://localhost:8080',
            authMode: config.authMode || 'project'
        });

        // React-specific error handling
        client.interceptors = {
            response: {
                onError: (error) => {
                    const corsInfo = CORSErrorHandler.detectCORSError(error);
                    if (corsInfo) {
                        console.group('🚫 CloudBox CORS Error');
                        console.error('Configuration issue detected:', corsInfo);
                        console.log('💡 Run this command to fix:', corsInfo.quickFixes[0].command);
                        console.groupEnd();
                        
                        // For development, show user-friendly notification
                        if (process.env.NODE_ENV === 'development') {
                            this.showReactNotification(corsInfo);
                        }
                        
                        throw CORSErrorHandler.createUserFriendlyError(corsInfo);
                    }
                    throw error;
                }
            }
        };

        return client;
    }

    static createVueClient(config = {}) {
        const { CloudBoxClient } = require('@ekoppen/cloudbox-sdk');
        
        const client = new CloudBoxClient({
            projectId: process.env.VUE_APP_PROJECT_ID || config.projectId,
            apiKey: process.env.VUE_APP_API_KEY || config.apiKey,
            endpoint: process.env.VUE_APP_CLOUDBOX_ENDPOINT || config.endpoint || 'http://localhost:8080',
            authMode: config.authMode || 'project'
        });

        // Vue-specific error handling with reactive notifications
        client.interceptors = {
            response: {
                onError: (error) => {
                    const corsInfo = CORSErrorHandler.detectCORSError(error);
                    if (corsInfo) {
                        console.group('🚫 CloudBox CORS Error');
                        console.error('Configuration issue detected:', corsInfo);
                        console.groupEnd();
                        
                        // Vue notification integration
                        this.showVueNotification(corsInfo);
                        
                        throw CORSErrorHandler.createUserFriendlyError(corsInfo);
                    }
                    throw error;
                }
            }
        };

        return client;
    }

    static createAngularClient(config = {}) {
        const { CloudBoxClient } = require('@ekoppen/cloudbox-sdk');
        
        // Angular environment integration
        const environment = config.environment || {};
        
        const client = new CloudBoxClient({
            projectId: environment.projectId || config.projectId,
            apiKey: environment.apiKey || config.apiKey,
            endpoint: environment.cloudboxEndpoint || config.endpoint || 'http://localhost:8080',
            authMode: config.authMode || 'project'
        });

        // Angular-specific error handling with service integration
        client.interceptors = {
            response: {
                onError: (error) => {
                    const corsInfo = CORSErrorHandler.detectCORSError(error);
                    if (corsInfo) {
                        console.group('🚫 CloudBox CORS Error');
                        console.error('Configuration issue detected:', corsInfo);
                        console.groupEnd();
                        
                        // Angular service notification
                        this.showAngularNotification(corsInfo);
                        
                        throw CORSErrorHandler.createUserFriendlyError(corsInfo);
                    }
                    throw error;
                }
            }
        };

        return client;
    }

    static createSvelteClient(config = {}) {
        const { CloudBoxClient } = require('@ekoppen/cloudbox-sdk');
        
        const client = new CloudBoxClient({
            projectId: import.meta?.env?.VITE_PROJECT_ID || config.projectId,
            apiKey: import.meta?.env?.VITE_API_KEY || config.apiKey,
            endpoint: import.meta?.env?.VITE_CLOUDBOX_ENDPOINT || config.endpoint || 'http://localhost:8080',
            authMode: config.authMode || 'project'
        });

        // Svelte-specific error handling with store integration
        client.interceptors = {
            response: {
                onError: (error) => {
                    const corsInfo = CORSErrorHandler.detectCORSError(error);
                    if (corsInfo) {
                        console.group('🚫 CloudBox CORS Error');
                        console.error('Configuration issue detected:', corsInfo);
                        console.groupEnd();
                        
                        // Svelte store notification
                        this.showSvelteNotification(corsInfo);
                        
                        throw CORSErrorHandler.createUserFriendlyError(corsInfo);
                    }
                    throw error;
                }
            }
        };

        return client;
    }

    // Framework-specific notification methods
    static showReactNotification(corsInfo) {
        // React Toast/Notification integration
        if (typeof window !== 'undefined' && window.ReactDOM) {
            const notification = document.createElement('div');
            notification.style.cssText = `
                position: fixed; top: 20px; right: 20px; z-index: 10000;
                background: #f8d7da; border: 1px solid #f5c6cb; color: #721c24;
                padding: 15px; border-radius: 8px; max-width: 400px;
                font-family: system-ui; font-size: 14px;
            `;
            notification.innerHTML = `
                <strong>🚫 CloudBox CORS Error</strong><br>
                <p>${corsInfo.suggestions[0]}</p>
                <button onclick="this.parentElement.remove()" style="
                    background: #721c24; color: white; border: none; 
                    padding: 5px 10px; border-radius: 4px; cursor: pointer;
                ">Dismiss</button>
            `;
            document.body.appendChild(notification);
        }
    }

    static showVueNotification(corsInfo) {
        // Vue.js notification integration
        if (typeof window !== 'undefined' && window.Vue) {
            console.log('Vue CORS notification:', corsInfo.quickFixes[0].command);
        }
    }

    static showAngularNotification(corsInfo) {
        // Angular notification service integration
        if (typeof window !== 'undefined') {
            console.log('Angular CORS notification:', corsInfo.quickFixes[0].command);
        }
    }

    static showSvelteNotification(corsInfo) {
        // Svelte store notification integration
        if (typeof window !== 'undefined') {
            console.log('Svelte CORS notification:', corsInfo.quickFixes[0].command);
        }
    }
}

/**
 * Authentication State Management Helpers
 */
export class AuthStateManager {
    static createReactAuthHook(client) {
        // React hook for authentication state
        const { useState, useEffect } = require('react');
        
        return function useCloudBoxAuth() {
            const [user, setUser] = useState(null);
            const [loading, setLoading] = useState(true);
            const [error, setError] = useState(null);

            useEffect(() => {
                const checkAuth = async () => {
                    try {
                        const currentUser = await client.auth.me();
                        setUser(currentUser);
                    } catch (err) {
                        setError(err);
                    } finally {
                        setLoading(false);
                    }
                };

                if (client.getAuthToken()) {
                    checkAuth();
                } else {
                    setLoading(false);
                }
            }, []);

            const login = async (credentials) => {
                try {
                    setLoading(true);
                    setError(null);
                    const response = await client.auth.login(credentials);
                    client.setAuthToken(response.token);
                    setUser(response.user);
                    return response;
                } catch (err) {
                    setError(err);
                    throw err;
                } finally {
                    setLoading(false);
                }
            };

            const logout = async () => {
                try {
                    await client.auth.logout();
                } finally {
                    client.clearAuthToken();
                    setUser(null);
                }
            };

            return { user, loading, error, login, logout };
        };
    }

    static createVueAuthComposable(client) {
        // Vue 3 Composition API composable
        const { ref, computed, onMounted } = require('vue');
        
        return function useCloudBoxAuth() {
            const user = ref(null);
            const loading = ref(true);
            const error = ref(null);

            const isAuthenticated = computed(() => !!user.value);

            const checkAuth = async () => {
                try {
                    if (client.getAuthToken()) {
                        const currentUser = await client.auth.me();
                        user.value = currentUser;
                    }
                } catch (err) {
                    error.value = err;
                } finally {
                    loading.value = false;
                }
            };

            const login = async (credentials) => {
                try {
                    loading.value = true;
                    error.value = null;
                    const response = await client.auth.login(credentials);
                    client.setAuthToken(response.token);
                    user.value = response.user;
                    return response;
                } catch (err) {
                    error.value = err;
                    throw err;
                } finally {
                    loading.value = false;
                }
            };

            const logout = async () => {
                try {
                    await client.auth.logout();
                } finally {
                    client.clearAuthToken();
                    user.value = null;
                }
            };

            onMounted(checkAuth);

            return {
                user,
                loading,
                error,
                isAuthenticated,
                login,
                logout
            };
        };
    }

    static createSvelteAuthStore(client) {
        // Svelte store for authentication
        const { writable, derived } = require('svelte/store');
        
        const user = writable(null);
        const loading = writable(true);
        const error = writable(null);

        const isAuthenticated = derived(user, $user => !!$user);

        const checkAuth = async () => {
            try {
                if (client.getAuthToken()) {
                    const currentUser = await client.auth.me();
                    user.set(currentUser);
                }
            } catch (err) {
                error.set(err);
            } finally {
                loading.set(false);
            }
        };

        const login = async (credentials) => {
            try {
                loading.set(true);
                error.set(null);
                const response = await client.auth.login(credentials);
                client.setAuthToken(response.token);
                user.set(response.user);
                return response;
            } catch (err) {
                error.set(err);
                throw err;
            } finally {
                loading.set(false);
            }
        };

        const logout = async () => {
            try {
                await client.auth.logout();
            } finally {
                client.clearAuthToken();
                user.set(null);
            }
        };

        // Initialize auth check
        checkAuth();

        return {
            user,
            loading,
            error,
            isAuthenticated,
            login,
            logout
        };
    }
}

/**
 * Development and Debug Utilities
 */
export class DevUtils {
    static enableDebugMode(client) {
        if (process.env.NODE_ENV !== 'development') return;

        console.log('🔧 CloudBox Debug Mode Enabled');
        
        // Log all requests
        const originalRequest = client.request.bind(client);
        client.request = async function(path, options = {}) {
            console.group(`📡 CloudBox API Call: ${options.method || 'GET'} ${path}`);
            console.log('Options:', options);
            
            try {
                const result = await originalRequest(path, options);
                console.log('✅ Success:', result);
                console.groupEnd();
                return result;
            } catch (error) {
                console.log('❌ Error:', error);
                console.groupEnd();
                throw error;
            }
        };

        // Add debug commands to window
        if (typeof window !== 'undefined') {
            window.cloudboxDebug = {
                client,
                testAuth: () => client.testAuthHeaders(),
                testConnection: () => client.testConnection(),
                corsInfo: () => {
                    console.log('CORS Debug Info:');
                    console.log('Origin:', window.location.origin);
                    console.log('Auth Strategies:', client.getAuthStrategies());
                    console.log('Current Token:', client.getAuthToken() ? 'Set' : 'None');
                }
            };
            console.log('💡 Debug tools available at window.cloudboxDebug');
        }
    }

    static createHealthCheck(client) {
        return async function healthCheck() {
            const results = {
                connection: false,
                authentication: false,
                cors: false,
                timestamp: new Date().toISOString()
            };

            try {
                // Test basic connection
                const connectionResult = await client.testConnection();
                results.connection = connectionResult.success;

                // Test authentication if token exists
                if (client.getAuthToken()) {
                    try {
                        await client.auth.me();
                        results.authentication = true;
                    } catch (error) {
                        results.authenticationError = error.message;
                    }
                }

                // Test CORS
                try {
                    await client.request('/collections', { method: 'GET' });
                    results.cors = true;
                } catch (error) {
                    const corsInfo = CORSErrorHandler.detectCORSError(error);
                    if (corsInfo) {
                        results.corsError = corsInfo;
                        results.corsQuickFix = corsInfo.quickFixes[0].command;
                    }
                }

            } catch (error) {
                results.generalError = error.message;
            }

            return results;
        };
    }
}