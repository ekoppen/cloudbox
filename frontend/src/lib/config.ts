import { PUBLIC_API_URL } from '$env/static/public';

// API Configuration - use relative URL for proxy in development, fallback to full URL
export const API_BASE_URL = PUBLIC_API_URL ? PUBLIC_API_URL : '';

// API Endpoints
export const API_ENDPOINTS = {
  auth: {
    login: `${API_BASE_URL}/api/v1/auth/login`,
    logout: `${API_BASE_URL}/api/v1/auth/logout`,
    refresh: `${API_BASE_URL}/api/v1/auth/refresh`,
    me: `${API_BASE_URL}/api/v1/auth/me`,
    changePassword: `${API_BASE_URL}/api/v1/auth/change-password`,
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
    stats: (id: string) => `${API_BASE_URL}/api/v1/projects/${id}/stats`,
    github: {
      config: (id: string) => `${API_BASE_URL}/api/v1/projects/${id}/github/config`,
      test: (id: string) => `${API_BASE_URL}/api/v1/projects/${id}/github/config/test`,
      instructions: (id: string) => `${API_BASE_URL}/api/v1/projects/${id}/github/instructions`,
    },
  },
  organizations: {
    list: `${API_BASE_URL}/api/v1/organizations`,
    create: `${API_BASE_URL}/api/v1/organizations`,
    get: (id: string) => `${API_BASE_URL}/api/v1/organizations/${id}`,
    update: (id: string) => `${API_BASE_URL}/api/v1/organizations/${id}`,
    delete: (id: string) => `${API_BASE_URL}/api/v1/organizations/${id}`,
    projects: (id: string) => `${API_BASE_URL}/api/v1/organizations/${id}/projects`,
  },
  admin: {
    stats: {
      system: `${API_BASE_URL}/api/v1/admin/stats/system`,
      userGrowth: `${API_BASE_URL}/api/v1/admin/stats/user-growth`,
      projectActivity: `${API_BASE_URL}/api/v1/admin/stats/project-activity`,
      functionExecutions: `${API_BASE_URL}/api/v1/admin/stats/function-executions`,
      deploymentStats: `${API_BASE_URL}/api/v1/admin/stats/deployment-stats`,
      storageStats: `${API_BASE_URL}/api/v1/admin/stats/storage-stats`,
      systemHealth: `${API_BASE_URL}/api/v1/admin/stats/system-health`,
    },
    users: {
      list: `${API_BASE_URL}/api/v1/admin/users`,
      create: `${API_BASE_URL}/api/v1/admin/users`,
      get: (id: string) => `${API_BASE_URL}/api/v1/admin/users/${id}`,
      update: (id: string) => `${API_BASE_URL}/api/v1/admin/users/${id}`,
      delete: (id: string) => `${API_BASE_URL}/api/v1/admin/users/${id}`,
    },
    organizationAdmins: {
      list: `${API_BASE_URL}/api/v1/admin/organization-admins`,
      assign: `${API_BASE_URL}/api/v1/admin/organization-admins`,
      revoke: (userId: string, orgId: string) => `${API_BASE_URL}/api/v1/admin/organization-admins/${userId}/${orgId}`,
    },
    projects: {
      list: `${API_BASE_URL}/api/v1/admin/projects`,
      get: (id: string) => `${API_BASE_URL}/api/v1/admin/projects/${id}`,
      update: (id: string) => `${API_BASE_URL}/api/v1/admin/projects/${id}`,
      delete: (id: string) => `${API_BASE_URL}/api/v1/admin/projects/${id}`,
      // Storage endpoints for project admin - consistent with project ID approach
      storage: {
        buckets: {
          list: (projectId: string) => `${API_BASE_URL}/api/v1/projects/${projectId}/storage/buckets`,
          create: (projectId: string) => `${API_BASE_URL}/api/v1/projects/${projectId}/storage/buckets`,
          get: (projectId: string, bucketName: string) => `${API_BASE_URL}/api/v1/projects/${projectId}/storage/buckets/${bucketName}`,
          update: (projectId: string, bucketName: string) => `${API_BASE_URL}/api/v1/projects/${projectId}/storage/buckets/${bucketName}`,
          delete: (projectId: string, bucketName: string) => `${API_BASE_URL}/api/v1/projects/${projectId}/storage/buckets/${bucketName}`,
        },
        files: {
          list: (projectId: string, bucketName: string) => `${API_BASE_URL}/api/v1/projects/${projectId}/storage/buckets/${bucketName}/files`,
          upload: (projectId: string, bucketName: string) => `${API_BASE_URL}/api/v1/projects/${projectId}/storage/buckets/${bucketName}/files`,
          get: (projectId: string, bucketName: string, fileId: string) => `${API_BASE_URL}/api/v1/projects/${projectId}/storage/buckets/${bucketName}/files/${fileId}`,
          delete: (projectId: string, bucketName: string, fileId: string) => `${API_BASE_URL}/api/v1/projects/${projectId}/storage/buckets/${bucketName}/files/${fileId}`,
          move: (projectId: string, bucketName: string, fileId: string) => `${API_BASE_URL}/api/v1/projects/${projectId}/storage/buckets/${bucketName}/files/${fileId}/move`,
        },
        folders: {
          list: (projectId: string, bucketName: string) => `${API_BASE_URL}/api/v1/projects/${projectId}/storage/buckets/${bucketName}/folders`,
          create: (projectId: string, bucketName: string) => `${API_BASE_URL}/api/v1/projects/${projectId}/storage/buckets/${bucketName}/folders`,
          delete: (projectId: string, bucketName: string) => `${API_BASE_URL}/api/v1/projects/${projectId}/storage/buckets/${bucketName}/folders`,
        }
      },
      // Public access management for admin
      visibility: {
        set: (projectId: string, bucketName: string) => `${API_BASE_URL}/api/v1/projects/${projectId}/storage/buckets/${bucketName}/visibility`,
        publicUrl: (projectId: string, bucketName: string, fileId: string) => `${API_BASE_URL}/api/v1/projects/${projectId}/storage/buckets/${bucketName}/files/${fileId}/public-url`,
        publicBuckets: (projectId: string) => `${API_BASE_URL}/api/v1/projects/${projectId}/storage/public-buckets`,
      },
      // Collections endpoints for project admin
      collections: {
        list: (projectId: string) => `${API_BASE_URL}/api/v1/projects/${projectId}/collections`,
        create: (projectId: string) => `${API_BASE_URL}/api/v1/projects/${projectId}/collections`,
        get: (projectId: string, collection: string) => `${API_BASE_URL}/api/v1/projects/${projectId}/collections/${collection}`,
        delete: (projectId: string, collection: string) => `${API_BASE_URL}/api/v1/projects/${projectId}/collections/${collection}`,
        documents: {
          list: (projectId: string, collection: string) => `${API_BASE_URL}/api/v1/projects/${projectId}/collections/${collection}/documents`,
          create: (projectId: string, collection: string) => `${API_BASE_URL}/api/v1/projects/${projectId}/collections/${collection}/documents`,
          get: (projectId: string, collection: string, id: string) => `${API_BASE_URL}/api/v1/projects/${projectId}/collections/${collection}/documents/${id}`,
          update: (projectId: string, collection: string, id: string) => `${API_BASE_URL}/api/v1/projects/${projectId}/collections/${collection}/documents/${id}`,
          delete: (projectId: string, collection: string, id: string) => `${API_BASE_URL}/api/v1/projects/${projectId}/collections/${collection}/documents/${id}`,
        }
      }
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

// Helper function to check if an error is a connection error (backend not available)
export function isConnectionError(error: Error): boolean {
  return error.message.includes('Failed to fetch') || 
         error.message.includes('NetworkError') ||
         error.message.includes('ERR_CONNECTION_RESET');
}

// Helper function to handle API errors gracefully
export function handleApiError(error: Error, showToast: (message: string) => void = () => {}): boolean {
  console.error('API Error:', error);
  
  if (isConnectionError(error)) {
    // Backend is not available - this is expected during development
    return false; // Don't show error toast
  } else {
    // Show error toast for actual network issues
    showToast('Netwerkfout bij communicatie met server');
    return true;
  }
}

// Helper function to generate public URLs for files
export function generatePublicFileUrl(projectSlug: string, bucketName: string, filePath: string): string {
  return `${API_BASE_URL}/public/${projectSlug}/${bucketName}/${filePath}`;
}

// Helper function to check if a bucket is public (for UI state)
export function isBucketPublic(bucket: any): boolean {
  return bucket && bucket.is_public === true;
}