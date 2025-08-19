import { writable } from 'svelte/store';
import { API_ENDPOINTS, createApiRequest } from '$lib/config';
import { auth } from './auth';
import { get } from 'svelte/store';

export interface PluginUIConfig {
  dashboard_tab?: {
    title: string;
    icon: string;
    path: string;
    order?: number;
  };
  project_menu?: {
    title: string;
    icon: string;
    path: string;
  };
}

export interface Plugin {
  name: string;
  version: string;
  description: string;
  author: string;
  type: string;
  status: 'installed' | 'enabled' | 'disabled' | 'error';
  installed_at: string;
  path: string;
  ui?: PluginUIConfig;
  permissions: string[];
  dependencies?: Record<string, string>;
}

export interface MarketplacePlugin {
  name: string;
  version: string;
  description: string;
  author: string;
  repository: string;
  category?: string;
  stars?: number;
  downloads?: number;
  license?: string;
  tags: string[];
  screenshots?: string[];
  readme?: string;
  permissions: string[];
  dependencies?: Record<string, string>;
  verified: boolean;
  official?: boolean;
  rating?: number;
  last_updated: string;
}

// Plugin stores
export const plugins = writable<Plugin[]>([]);
export const pluginsLoaded = writable(false);
export const pluginsLoading = writable(false);

// Marketplace stores
export const marketplacePlugins = writable<MarketplacePlugin[]>([]);
export const marketplaceLoading = writable(false);
export const searchQuery = writable('');
export const selectedTags = writable<string[]>([]);

// Installation progress
export const installationProgress = writable<{
  pluginName: string;
  status: 'downloading' | 'installing' | 'configuring' | 'complete' | 'error';
  progress: number;
  message: string;
} | null>(null);

// Dynamic navigation items from plugins
export const dynamicNavItems = writable<any[]>([]);
export const dynamicProjectMenuItems = writable<any[]>([]);

class PluginManager {
  private loadedPlugins: Map<string, Plugin> = new Map();

  async loadProjectPlugins(projectId: string): Promise<Plugin[]> {
    pluginsLoading.set(true);
    try {
      const authState = get(auth);
      const response = await createApiRequest(`/api/v1/projects/${projectId}/plugins/installed`, {
        credentials: 'include',
        headers: {
          ...(authState.token ? { Authorization: `Bearer ${authState.token}` } : {})
        }
      });
      
      if (response.ok) {
        const data = await response.json();
        if (data.success || data.plugins) {
          const pluginList = data.plugins || [];
          
          // Only include enabled plugins for navigation
          const enabledPlugins = pluginList.filter((plugin: any) => plugin.status === 'enabled');
          
          plugins.set(enabledPlugins);
          pluginsLoaded.set(true);
          this.updateNavigationFromProjectPlugins(enabledPlugins, projectId);
          return enabledPlugins;
        } else {
          throw new Error(data.error || 'Failed to load project plugins');
        }
      } else {
        console.warn(`Failed to load project plugins: ${response.status}`);
        // Set empty arrays but don't throw to allow graceful degradation
        plugins.set([]);
        dynamicNavItems.set([]);
        dynamicProjectMenuItems.set([]);
        pluginsLoaded.set(true);
        return [];
      }
    } catch (error) {
      console.error('Failed to load project plugins:', error);
      // Set empty arrays but don't throw error to allow graceful degradation
      plugins.set([]);
      dynamicNavItems.set([]);
      dynamicProjectMenuItems.set([]);
      pluginsLoaded.set(true);
      return [];
    } finally {
      pluginsLoading.set(false);
    }
  }

  async loadPlugins(): Promise<Plugin[]> {
    pluginsLoading.set(true);
    try {
      const authState = get(auth);
      const response = await createApiRequest(API_ENDPOINTS.admin.plugins.list, {
        credentials: 'include',
        headers: {
          ...(authState.token ? { Authorization: `Bearer ${authState.token}` } : {})
        }
      });
      
      if (response.ok) {
        const data = await response.json();
        if (data.success) {
          const pluginList = data.plugins || [];
          plugins.set(pluginList);
          pluginsLoaded.set(true);
          this.updateNavigation(pluginList);
          return pluginList;
        } else {
          throw new Error(data.error || 'Failed to load plugins');
        }
      } else {
        // Enhanced error handling for different HTTP status codes
        let errorMessage = `HTTP ${response.status}: ${response.statusText}`;
        try {
          const errorData = await response.json();
          if (errorData.error) {
            errorMessage = errorData.error;
          }
        } catch (e) {
          // Keep the original error message if JSON parsing fails
        }
        throw new Error(errorMessage);
      }
    } catch (error) {
      console.error('Failed to load plugins:', error);
      // Set empty array but don't throw error to allow graceful degradation
      plugins.set([]);
      pluginsLoaded.set(true); // Mark as loaded even if failed to prevent infinite loading
      return [];
    } finally {
      pluginsLoading.set(false);
    }
  }

  private updateNavigationFromProjectPlugins(enabledPlugins: any[], projectId: string) {
    console.log('📋 Updating navigation with project plugins:', enabledPlugins);
    
    // Extract project menu items from enabled plugins
    // Handle both ui and ui_config structures for backward compatibility
    const projectMenuItems = enabledPlugins
      .filter(plugin => {
        const uiConfig = plugin.ui || plugin.ui_config;
        return uiConfig?.project_menu;
      })
      .map(plugin => {
        const uiConfig = plugin.ui || plugin.ui_config;
        return {
          id: plugin.plugin_name || plugin.name,
          name: uiConfig.project_menu.title,
          icon: uiConfig.project_menu.icon,
          href: uiConfig.project_menu.path.replace('{projectId}', projectId),
          plugin: true
        };
      });

    console.log('📁 Project menu items from plugins:', projectMenuItems);

    // For project context, we only update project menu items
    dynamicProjectMenuItems.set(projectMenuItems);
    
    console.log('✨ Project navigation updated');
  }

  private updateNavigation(activePlugins: Plugin[]) {
    console.log('📋 Updating navigation with plugins:', activePlugins);
    
    // Extract dashboard navigation items
    const dashboardItems = activePlugins
      .filter(plugin => plugin.status === 'enabled' && plugin.ui?.dashboard_tab)
      .map(plugin => ({
        id: plugin.name,
        name: plugin.ui!.dashboard_tab!.title,
        icon: plugin.ui!.dashboard_tab!.icon,
        href: plugin.ui!.dashboard_tab!.path,
        order: plugin.ui!.dashboard_tab!.order || 999,
        plugin: true
      }))
      .sort((a, b) => a.order - b.order);

    // Extract project menu items
    const projectMenuItems = activePlugins
      .filter(plugin => plugin.status === 'enabled' && plugin.ui?.project_menu)
      .map(plugin => ({
        id: plugin.name,
        name: plugin.ui!.project_menu!.title,
        icon: plugin.ui!.project_menu!.icon,
        href: plugin.ui!.project_menu!.path,
        plugin: true
      }));

    console.log('🎯 Dashboard items:', dashboardItems);
    console.log('📁 Project menu items:', projectMenuItems);

    // Update stores
    dynamicNavItems.set(dashboardItems);
    dynamicProjectMenuItems.set(projectMenuItems);
    
    console.log('✨ Navigation stores updated');
  }

  async enablePlugin(pluginName: string): Promise<void> {
    const authState = get(auth);
    const response = await createApiRequest(API_ENDPOINTS.admin.plugins.enable(pluginName), {
      method: 'POST',
      credentials: 'include',
      headers: {
        ...(authState.token ? { Authorization: `Bearer ${authState.token}` } : {})
      }
    });
    
    if (response.ok) {
      const data = await response.json();
      if (data.success) {
        // Update plugin status in store
        plugins.update(pluginList => {
          const index = pluginList.findIndex(p => p.name === pluginName);
          if (index !== -1) {
            pluginList[index].status = 'enabled';
          }
          return [...pluginList];
        });
        // Reload navigation
        await this.loadPlugins();
      } else {
        throw new Error(data.error || 'Failed to enable plugin');
      }
    } else {
      throw new Error(`HTTP ${response.status}: ${response.statusText}`);
    }
  }

  async disablePlugin(pluginName: string): Promise<void> {
    const authState = get(auth);
    const response = await createApiRequest(API_ENDPOINTS.admin.plugins.disable(pluginName), {
      method: 'POST',
      credentials: 'include',
      headers: {
        ...(authState.token ? { Authorization: `Bearer ${authState.token}` } : {})
      }
    });
    
    if (response.ok) {
      const data = await response.json();
      if (data.success) {
        // Update plugin status in store
        plugins.update(pluginList => {
          const index = pluginList.findIndex(p => p.name === pluginName);
          if (index !== -1) {
            pluginList[index].status = 'disabled';
          }
          return [...pluginList];
        });
        // Reload navigation
        await this.loadPlugins();
      } else {
        throw new Error(data.error || 'Failed to disable plugin');
      }
    } else {
      throw new Error(`HTTP ${response.status}: ${response.statusText}`);
    }
  }

  async uninstallPlugin(pluginName: string): Promise<void> {
    const authState = get(auth);
    const response = await createApiRequest(API_ENDPOINTS.admin.plugins.uninstall(pluginName), {
      method: 'DELETE',
      credentials: 'include',
      headers: {
        ...(authState.token ? { Authorization: `Bearer ${authState.token}` } : {})
      }
    });
    
    if (response.ok) {
      const data = await response.json();
      if (data.success) {
        // Remove plugin from store
        plugins.update(pluginList => pluginList.filter(p => p.name !== pluginName));
        // Reload navigation
        await this.loadPlugins();
      } else {
        throw new Error(data.error || 'Failed to uninstall plugin');
      }
    } else {
      throw new Error(`HTTP ${response.status}: ${response.statusText}`);
    }
  }

  async installPlugin(repository: string, version?: string): Promise<void> {
    const authState = get(auth);
    
    // Set installation progress
    installationProgress.set({
      pluginName: repository.split('/').pop() || repository,
      status: 'downloading',
      progress: 0,
      message: 'Downloading plugin from GitHub...'
    });

    try {
      const response = await createApiRequest(API_ENDPOINTS.admin.plugins.install, {
        method: 'POST',
        credentials: 'include',
        headers: {
          ...(authState.token ? { Authorization: `Bearer ${authState.token}` } : {})
        },
        body: JSON.stringify({ repository, version })
      });
      
      if (response.ok) {
        const data = await response.json();
        if (data.success) {
          // Update progress
          installationProgress.set({
            pluginName: repository.split('/').pop() || repository,
            status: 'complete',
            progress: 100,
            message: 'Plugin installed successfully!'
          });
          
          // Reload plugins
          await this.loadPlugins();
          
          // Clear progress after delay
          setTimeout(() => {
            installationProgress.set(null);
          }, 3000);
        } else {
          throw new Error(data.error || 'Failed to install plugin');
        }
      } else {
        throw new Error(`HTTP ${response.status}: ${response.statusText}`);
      }
    } catch (error) {
      installationProgress.set({
        pluginName: repository.split('/').pop() || repository,
        status: 'error',
        progress: 0,
        message: error instanceof Error ? error.message : 'Installation failed'
      });
      
      // Clear error after delay
      setTimeout(() => {
        installationProgress.set(null);
      }, 5000);
      
      throw error;
    }
  }

  async loadMarketplace(): Promise<MarketplacePlugin[]> {
    marketplaceLoading.set(true);
    try {
      const authState = get(auth);
      const response = await createApiRequest(API_ENDPOINTS.admin.plugins.marketplace, {
        credentials: 'include',
        headers: {
          ...(authState.token ? { Authorization: `Bearer ${authState.token}` } : {})
        }
      });
      
      if (response.ok) {
        const data = await response.json();
        if (data.success) {
          const rawPlugins = data.marketplace || data.plugins || [];
          // Map the backend response to the frontend interface
          const plugins = rawPlugins.map((plugin: any) => ({
            ...plugin,
            // Extract stars from source_metadata if available
            stars: plugin.source_metadata?.stars || plugin.stars || 0,
            // Map installs to downloads if downloads is missing
            downloads: plugin.downloads || plugin.installs || 0,
            // Ensure tags array exists
            tags: plugin.tags || [],
            // Map last_updated from published_at or updated_at
            last_updated: plugin.published_at || plugin.updated_at || plugin.last_updated || new Date().toISOString()
          }));
          marketplacePlugins.set(plugins);
          return plugins;
        } else {
          throw new Error(data.error || 'Failed to load marketplace');
        }
      } else {
        // Enhanced error handling for different HTTP status codes
        let errorMessage = `HTTP ${response.status}: ${response.statusText}`;
        try {
          const errorData = await response.json();
          if (errorData.error) {
            errorMessage = errorData.error;
          }
        } catch (e) {
          // Keep the original error message if JSON parsing fails
        }
        throw new Error(errorMessage);
      }
    } catch (error) {
      console.error('Failed to load marketplace:', error);
      // Set empty array but still throw to show error message
      marketplacePlugins.set([]);
      return []; // Return empty array instead of throwing
    } finally {
      marketplaceLoading.set(false);
    }
  }

  async searchMarketplace(query: string, tags?: string[]): Promise<MarketplacePlugin[]> {
    marketplaceLoading.set(true);
    try {
      const authState = get(auth);
      const params = new URLSearchParams();
      if (query) params.append('q', query);
      if (tags && tags.length > 0) params.append('tags', tags.join(','));
      
      const response = await createApiRequest(
        `${API_ENDPOINTS.admin.plugins.search}?${params.toString()}`, 
        {
          credentials: 'include',
          headers: {
            ...(authState.token ? { Authorization: `Bearer ${authState.token}` } : {})
          }
        }
      );
      
      if (response.ok) {
        const data = await response.json();
        if (data.success) {
          const rawPlugins = data.results || data.plugins || [];
          // Map the backend response to the frontend interface (same as loadMarketplace)
          const plugins = rawPlugins.map((plugin: any) => ({
            ...plugin,
            // Extract stars from source_metadata if available
            stars: plugin.source_metadata?.stars || plugin.stars || 0,
            // Map installs to downloads if downloads is missing
            downloads: plugin.downloads || plugin.installs || 0,
            // Ensure tags array exists
            tags: plugin.tags || [],
            // Map last_updated from published_at or updated_at
            last_updated: plugin.published_at || plugin.updated_at || plugin.last_updated || new Date().toISOString()
          }));
          marketplacePlugins.set(plugins);
          return plugins;
        } else {
          throw new Error(data.error || 'Failed to search marketplace');
        }
      } else {
        // Enhanced error handling for different HTTP status codes
        let errorMessage = `HTTP ${response.status}: ${response.statusText}`;
        try {
          const errorData = await response.json();
          if (errorData.error) {
            errorMessage = errorData.error;
          }
        } catch (e) {
          // Keep the original error message if JSON parsing fails
        }
        throw new Error(errorMessage);
      }
    } catch (error) {
      console.error('Failed to search marketplace:', error);
      marketplacePlugins.set([]);
      return []; // Return empty array instead of throwing for graceful degradation
    } finally {
      marketplaceLoading.set(false);
    }
  }

  async addPluginToMarketplace(pluginData: {
    name: string;
    version: string;
    description: string;
    author: string;
    repository: string;
    license?: string;
    tags?: string[];
    permissions?: string[];
    dependencies?: Record<string, string>;
    verified?: boolean;
    approved?: boolean;
    type?: string;
    main_file?: string;
  }): Promise<void> {
    const authState = get(auth);
    const response = await createApiRequest(API_ENDPOINTS.admin.plugins.addToMarketplace, {
      method: 'POST',
      credentials: 'include',
      headers: {
        ...(authState.token ? { Authorization: `Bearer ${authState.token}` } : {})
      },
      body: JSON.stringify(pluginData)
    });
    
    if (response.ok) {
      const data = await response.json();
      if (data.success) {
        // Reload marketplace to show the new plugin
        await this.loadMarketplace();
      } else {
        throw new Error(data.error || 'Failed to add plugin to marketplace');
      }
    } else {
      throw new Error(`HTTP ${response.status}: ${response.statusText}`);
    }
  }

  async reloadPlugins(): Promise<Plugin[]> {
    const authState = get(auth);
    const response = await createApiRequest(API_ENDPOINTS.admin.plugins.reload, {
      method: 'POST',
      credentials: 'include',
      headers: {
        ...(authState.token ? { Authorization: `Bearer ${authState.token}` } : {})
      }
    });
    
    if (response.ok) {
      const data = await response.json();
      if (data.success) {
        return await this.loadPlugins();
      } else {
        throw new Error(data.error || 'Failed to reload plugins');
      }
    } else {
      throw new Error(`HTTP ${response.status}: ${response.statusText}`);
    }
  }
}

export const pluginManager = new PluginManager();