import { PUBLIC_API_URL } from '$env/static/public';

// API Configuration
export const API_BASE_URL = PUBLIC_API_URL || 'http://localhost:8080';

// API Endpoints
export const API_ENDPOINTS = {
  auth: {
    login: `${API_BASE_URL}/api/v1/auth/login`,
    logout: `${API_BASE_URL}/api/v1/auth/logout`,
    refresh: `${API_BASE_URL}/api/v1/auth/refresh`,
    me: `${API_BASE_URL}/api/v1/auth/me`,
  },
  users: {
    list: `${API_BASE_URL}/api/v1/users`,
    create: `${API_BASE_URL}/api/v1/users`,
    update: (id: string) => `${API_BASE_URL}/api/v1/users/${id}`,
    delete: (id: string) => `${API_BASE_URL}/api/v1/users/${id}`,
  },
  projects: {
    list: `${API_BASE_URL}/api/v1/projects`,
    create: `${API_BASE_URL}/api/v1/projects`,
    get: (id: string) => `${API_BASE_URL}/api/v1/projects/${id}`,
    update: (id: string) => `${API_BASE_URL}/api/v1/projects/${id}`,
    delete: (id: string) => `${API_BASE_URL}/api/v1/projects/${id}`,
  },
  organizations: {
    list: `${API_BASE_URL}/api/v1/organizations`,
    create: `${API_BASE_URL}/api/v1/organizations`,
    get: (id: string) => `${API_BASE_URL}/api/v1/organizations/${id}`,
    update: (id: string) => `${API_BASE_URL}/api/v1/organizations/${id}`,
    delete: (id: string) => `${API_BASE_URL}/api/v1/organizations/${id}`,
  },
  admin: {
    stats: {
      overview: `${API_BASE_URL}/api/v1/admin/stats/overview`,
      userGrowth: `${API_BASE_URL}/api/v1/admin/stats/user-growth`,
      projectActivity: `${API_BASE_URL}/api/v1/admin/stats/project-activity`,
      systemHealth: `${API_BASE_URL}/api/v1/admin/stats/system-health`,
    },
    users: {
      list: `${API_BASE_URL}/api/v1/admin/users`,
      get: (id: string) => `${API_BASE_URL}/api/v1/admin/users/${id}`,
      update: (id: string) => `${API_BASE_URL}/api/v1/admin/users/${id}`,
      delete: (id: string) => `${API_BASE_URL}/api/v1/admin/users/${id}`,
    },
    projects: {
      list: `${API_BASE_URL}/api/v1/admin/projects`,
      get: (id: string) => `${API_BASE_URL}/api/v1/admin/projects/${id}`,
      update: (id: string) => `${API_BASE_URL}/api/v1/admin/projects/${id}`,
      delete: (id: string) => `${API_BASE_URL}/api/v1/admin/projects/${id}`,
    },
    system: {
      info: `${API_BASE_URL}/api/v1/admin/system/info`,
      settings: `${API_BASE_URL}/api/v1/admin/system/settings`,
      restart: `${API_BASE_URL}/api/v1/admin/system/restart`,
      clearCache: `${API_BASE_URL}/api/v1/admin/system/clear-cache`,
      backup: `${API_BASE_URL}/api/v1/admin/system/backup`,
    }
  }
} as const;

// Helper function to create API fetch with default options
export function createApiRequest(url: string, options: RequestInit = {}): Promise<Response> {
  const defaultOptions: RequestInit = {
    headers: {
      'Content-Type': 'application/json',
      ...options.headers,
    },
    ...options,
  };

  return fetch(url, defaultOptions);
}