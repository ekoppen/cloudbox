<script lang="ts">
  import { onMount } from 'svelte';
  import Button from '$lib/components/ui/button.svelte';
  import Card from '$lib/components/ui/card.svelte';
  import Icon from '$lib/components/ui/icon.svelte';
  import Badge from '$lib/components/ui/badge.svelte';
  import PluginMarketplace from '$lib/components/plugin-marketplace.svelte';
  import AddPluginModal from '$lib/components/admin/add-plugin-modal.svelte';
  import { addToast } from '$lib/stores/toast';
  import { API_ENDPOINTS, createApiRequest } from '$lib/config';
  import { auth } from '$lib/stores/auth';
  import { pluginManager } from '$lib/stores/plugins';

  interface Plugin {
    name: string;
    version: string;
    description: string;
    author: string;
    type: string;
    status: 'installed' | 'enabled' | 'disabled' | 'error';
    installed_at: string;
    path: string;
    permissions: string[];
    ui?: {
      dashboard_tab?: any;
      project_menu?: any;
    };
    dependencies?: Record<string, string>;
  }

  let plugins: Plugin[] = [];
  let loading = true;
  let selectedPlugin: Plugin | null = null;
  let showPluginDetails = false;
  let showMarketplace = false;
  let showAddPlugin = false;

  async function loadPlugins() {
    loading = true;
    try {
      const loadedPlugins = await pluginManager.loadPlugins();
      plugins = loadedPlugins || [];
      
      // Show informational message if no plugins loaded due to API issues
      if (plugins.length === 0) {
        console.warn('No plugins loaded - this may be due to API connectivity issues');
      }
    } catch (error) {
      console.error('Failed to load plugins:', error);
      // Show a more user-friendly error message
      const errorMessage = error instanceof Error ? error.message : 'Unknown error occurred';
      addToast(`Failed to load plugins: ${errorMessage}. The plugin system may not be fully available.`, 'warning');
      plugins = [];
    } finally {
      loading = false;
    }
  }

  async function togglePlugin(plugin: Plugin) {
    try {
      if (plugin.status === 'enabled') {
        await pluginManager.disablePlugin(plugin.name);
        addToast('Plugin disabled successfully', 'success');
      } else {
        await pluginManager.enablePlugin(plugin.name);
        addToast('Plugin enabled successfully', 'success');
      }
      // Reload plugins to update the list
      await loadPlugins();
    } catch (error) {
      console.error(`Failed to ${plugin.status === 'enabled' ? 'disable' : 'enable'} plugin:`, error);
      addToast(`Failed to ${plugin.status === 'enabled' ? 'disable' : 'enable'} plugin: ${error.message}`, 'error');
    }
  }

  async function uninstallPlugin(plugin: Plugin) {
    if (!confirm(`Are you sure you want to uninstall ${plugin.name}? This action cannot be undone.`)) {
      return;
    }

    try {
      await pluginManager.uninstallPlugin(plugin.name);
      addToast('Plugin uninstalled successfully', 'success');
      if (selectedPlugin?.name === plugin.name) {
        showPluginDetails = false;
        selectedPlugin = null;
      }
      // Reload plugins to update the list
      await loadPlugins();
    } catch (error) {
      console.error('Failed to uninstall plugin:', error);
      addToast('Failed to uninstall plugin: ' + error.message, 'error');
    }
  }

  async function reloadPlugins() {
    try {
      await pluginManager.reloadPlugins();
      addToast('Plugins reloaded successfully', 'success');
      await loadPlugins();
    } catch (error) {
      console.error('Failed to reload plugins:', error);
      addToast('Failed to reload plugins: ' + error.message, 'error');
    }
  }

  function getStatusColor(status: string) {
    switch (status) {
      case 'enabled': return 'bg-green-100 text-green-800 dark:bg-green-900/20 dark:text-green-300';
      case 'disabled': return 'bg-gray-100 text-gray-800 dark:bg-gray-900/20 dark:text-gray-300';
      case 'error': return 'bg-red-100 text-red-800 dark:bg-red-900/20 dark:text-red-300';
      default: return 'bg-blue-100 text-blue-800 dark:bg-blue-900/20 dark:text-blue-300';
    }
  }

  function getTypeIcon(type: string) {
    switch (type) {
      case 'dashboard-plugin': return 'template';
      case 'project-plugin': return 'folder';
      case 'system-plugin': return 'cog';
      default: return 'puzzle';
    }
  }

  function showDetails(plugin: Plugin) {
    selectedPlugin = plugin;
    showPluginDetails = true;
  }

  onMount(() => {
    loadPlugins();
  });
</script>

<svelte:head>
  <title>Plugin Management - Admin - CloudBox</title>
</svelte:head>

<div class="space-y-8">
  <!-- Header -->
  <div class="flex items-center justify-between">
    <div>
      <h1 class="text-3xl font-bold text-foreground font-['Inter']">Plugin Management</h1>
      <p class="text-muted-foreground mt-2 text-base">
        Manage installed CloudBox plugins and extensions
      </p>
    </div>
    
    <div class="flex gap-3">
      <Button on:click={reloadPlugins} variant="outline" className="h-10">
        <Icon name="refresh" size={16} className="mr-2" />
        Reload
      </Button>
      
      <Button on:click={() => showAddPlugin = true} variant="outline" className="h-10">
        <Icon name="plus-circle" size={16} className="mr-2" />
        Add Plugin
      </Button>
      
      <Button on:click={() => showMarketplace = true} className="h-10 px-6">
        <Icon name="store" size={16} className="mr-2" />
        Browse Marketplace
      </Button>
    </div>
  </div>

  <!-- Plugin Stats -->
  <div class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-6">
    <div class="bg-background border border-border rounded-xl p-6 hover:shadow-sm transition-shadow">
      <div class="flex items-center justify-between">
        <div>
          <p class="text-sm font-medium text-muted-foreground mb-1">Total Plugins</p>
          <p class="text-3xl font-bold text-foreground">{plugins.length}</p>
        </div>
        <div class="w-12 h-12 bg-blue-100 rounded-xl flex items-center justify-center">
          <Icon name="puzzle" size={24} className="text-blue-600" />
        </div>
      </div>
    </div>
    
    <div class="bg-background border border-border rounded-xl p-6 hover:shadow-sm transition-shadow">
      <div class="flex items-center justify-between">
        <div>
          <p class="text-sm font-medium text-muted-foreground mb-1">Enabled</p>
          <p class="text-3xl font-bold text-foreground">{plugins.filter(p => p.status === 'enabled').length}</p>
        </div>
        <div class="w-12 h-12 bg-green-100 rounded-xl flex items-center justify-center">
          <Icon name="check-circle" size={24} className="text-green-600" />
        </div>
      </div>
    </div>
    
    <div class="bg-background border border-border rounded-xl p-6 hover:shadow-sm transition-shadow">
      <div class="flex items-center justify-between">
        <div>
          <p class="text-sm font-medium text-muted-foreground mb-1">Disabled</p>
          <p class="text-3xl font-bold text-foreground">{plugins.filter(p => p.status === 'disabled').length}</p>
        </div>
        <div class="w-12 h-12 bg-gray-100 rounded-xl flex items-center justify-center">
          <Icon name="pause-circle" size={24} className="text-gray-600" />
        </div>
      </div>
    </div>
    
    <div class="bg-background border border-border rounded-xl p-6 hover:shadow-sm transition-shadow">
      <div class="flex items-center justify-between">
        <div>
          <p class="text-sm font-medium text-muted-foreground mb-1">Errors</p>
          <p class="text-3xl font-bold text-foreground">{plugins.filter(p => p.status === 'error').length}</p>
        </div>
        <div class="w-12 h-12 bg-red-100 rounded-xl flex items-center justify-center">
          <Icon name="x-circle" size={24} className="text-red-600" />
        </div>
      </div>
    </div>
  </div>

  <!-- Plugins List -->
  <div class="bg-background border border-border rounded-2xl overflow-hidden">
    <div class="px-8 py-6 border-b border-border">
      <h2 class="text-xl font-semibold text-foreground">Installed Plugins</h2>
      <p class="text-sm text-muted-foreground mt-1">Manage your installed plugin collection</p>
    </div>

    {#if loading}
      <div class="p-16 text-center">
        <div class="w-12 h-12 bg-primary/10 rounded-xl flex items-center justify-center mx-auto mb-4">
          <div class="animate-spin rounded-full h-6 w-6 border-b-2 border-primary"></div>
        </div>
        <p class="text-foreground font-medium">Loading plugins...</p>
        <p class="text-sm text-muted-foreground mt-1">Please wait while we fetch your plugins</p>
      </div>
    {:else if plugins.length === 0}
      <div class="p-16 text-center">
        <div class="w-20 h-20 bg-muted rounded-2xl flex items-center justify-center mx-auto mb-6">
          <Icon name="puzzle" size={40} className="text-muted-foreground" />
        </div>
        <h3 class="text-xl font-semibold text-foreground mb-3">No plugins installed</h3>
        <p class="text-muted-foreground mb-8 max-w-md mx-auto leading-relaxed">
          Get started by installing your first plugin to extend CloudBox functionality and unlock new features.
        </p>
        <div class="space-y-4">
          <Button on:click={() => showMarketplace = true} size="lg" className="px-8">
            <Icon name="store" size={18} className="mr-2" />
            Browse Plugin Marketplace
          </Button>
          <div class="text-sm text-muted-foreground">
            Plugin system is ready and operational
          </div>
        </div>
      </div>
    {:else}
      <div class="divide-y divide-border">
        {#each plugins as plugin (plugin.name)}
          <div class="px-8 py-6 hover:bg-muted/50 transition-colors">
            <div class="flex items-center justify-between">
              <div class="flex items-center space-x-4 flex-1">
                <div class="w-12 h-12 bg-primary/10 rounded-xl flex items-center justify-center flex-shrink-0">
                  <Icon name={getTypeIcon(plugin.type)} size={24} className="text-primary" />
                </div>
                
                <div class="flex-1 min-w-0">
                  <div class="flex items-center space-x-3 mb-2">
                    <h3 class="text-lg font-semibold text-foreground">{plugin.name}</h3>
                    <Badge className="{getStatusColor(plugin.status)} font-medium">
                      {plugin.status}
                    </Badge>
                    <span class="text-sm text-muted-foreground font-mono">v{plugin.version}</span>
                  </div>
                  
                  <p class="text-sm text-muted-foreground mb-3 leading-relaxed">{plugin.description}</p>
                  
                  <div class="flex items-center space-x-6 text-xs text-muted-foreground">
                    <div class="flex items-center space-x-1">
                      <Icon name="user" size={12} />
                      <span>By {plugin.author}</span>
                    </div>
                    <div class="flex items-center space-x-1">
                      <Icon name="calendar" size={12} />
                      <span>Installed {new Date(plugin.installed_at).toLocaleDateString('en-US')}</span>
                    </div>
                    {#if plugin.ui?.project_menu}
                      <div class="flex items-center space-x-1">
                        <Icon name="link" size={12} />
                        <span>Project Integration</span>
                      </div>
                    {/if}
                  </div>
                </div>
              </div>

              <div class="flex items-center space-x-3">
                <Button 
                  on:click={() => showDetails(plugin)}
                  variant="ghost"
                  size="sm"
                  className="text-muted-foreground hover:text-foreground"
                >
                  <Icon name="eye" size={16} className="mr-1" />
                  Details
                </Button>
                
                <Button 
                  on:click={() => togglePlugin(plugin)}
                  variant={plugin.status === 'enabled' ? 'outline' : 'default'}
                  size="sm"
                >
                  {plugin.status === 'enabled' ? 'Disable' : 'Enable'}
                </Button>
                
                <Button 
                  on:click={() => uninstallPlugin(plugin)}
                  variant="ghost"
                  size="sm"
                  className="text-red-600 hover:text-red-700 hover:bg-red-50"
                >
                  <Icon name="trash" size={16} />
                </Button>
              </div>
            </div>
          </div>
        {/each}
      </div>
    {/if}
  </div>

  <!-- Plugin Details Modal -->
  {#if showPluginDetails && selectedPlugin}
    <!-- svelte-ignore a11y-click-events-have-key-events -->
    <!-- svelte-ignore a11y-no-noninteractive-element-interactions -->
    <div class="fixed inset-0 bg-black/50 flex items-center justify-center z-50" 
         role="dialog" 
         aria-modal="true" 
         aria-labelledby="plugin-modal-title"
         tabindex="-1"
         on:click={() => showPluginDetails = false}
         on:keydown={(e) => e.key === 'Escape' && (showPluginDetails = false)}>
      <!-- svelte-ignore a11y-click-events-have-key-events -->
      <!-- svelte-ignore a11y-no-noninteractive-element-interactions -->
      <div class="max-w-2xl w-full mx-4 max-h-[80vh] overflow-y-auto bg-card border border-border rounded-lg shadow-lg" 
           role="document"
           on:click|stopPropagation
           on:keydown|stopPropagation>
        <div class="p-6 border-b border-border">
          <div class="flex items-center justify-between">
            <div class="flex items-center space-x-3">
              <Icon name={getTypeIcon(selectedPlugin.type)} size={24} className="text-primary" />
              <div>
                <h2 id="plugin-modal-title" class="text-xl font-semibold text-foreground">{selectedPlugin.name}</h2>
                <p class="text-sm text-muted-foreground">v{selectedPlugin.version} by {selectedPlugin.author}</p>
              </div>
            </div>
            <Badge className={getStatusColor(selectedPlugin.status)}>
              {selectedPlugin.status}
            </Badge>
          </div>
        </div>
        
        <div class="p-6 space-y-6">
          <!-- Description -->
          <div>
            <h3 class="font-semibold text-foreground mb-2">Description</h3>
            <p class="text-muted-foreground">{selectedPlugin.description}</p>
          </div>

          <!-- UI Integration -->
          {#if selectedPlugin.ui}
            <div>
              <h3 class="font-semibold text-foreground mb-2">UI Integration</h3>
              <div class="space-y-2">
                {#if selectedPlugin.ui.dashboard_tab}
                  <div class="flex items-center text-sm">
                    <Icon name="layout-dashboard" size={16} className="mr-2 text-blue-500" />
                    <span>Dashboard Tab: {selectedPlugin.ui.dashboard_tab.title}</span>
                  </div>
                {/if}
                {#if selectedPlugin.ui.project_menu}
                  <div class="flex items-center text-sm">
                    <Icon name="folder" size={16} className="mr-2 text-green-500" />
                    <span>Project Menu: {selectedPlugin.ui.project_menu.title}</span>
                  </div>
                {/if}
              </div>
            </div>
          {/if}

          <!-- Permissions -->
          <div>
            <h3 class="font-semibold text-foreground mb-2">Permissions</h3>
            <div class="flex flex-wrap gap-2">
              {#each selectedPlugin.permissions as permission}
                <Badge variant="outline" className="text-xs">
                  {permission}
                </Badge>
              {/each}
            </div>
          </div>

          <!-- Dependencies -->
          {#if selectedPlugin.dependencies && Object.keys(selectedPlugin.dependencies).length > 0}
            <div>
              <h3 class="font-semibold text-foreground mb-2">Dependencies</h3>
              <div class="space-y-1">
                {#each Object.entries(selectedPlugin.dependencies) as [name, version]}
                  <div class="text-sm font-mono bg-muted p-2 rounded">
                    {name}: {version}
                  </div>
                {/each}
              </div>
            </div>
          {/if}

          <!-- Plugin Path -->
          <div>
            <h3 class="font-semibold text-foreground mb-2">Installation Path</h3>
            <code class="text-sm bg-muted p-2 rounded block">{selectedPlugin.path}</code>
          </div>
        </div>
        
        <div class="p-6 border-t border-border flex justify-between">
          <div class="flex gap-2">
            <Button 
              on:click={() => togglePlugin(selectedPlugin)}
              variant={selectedPlugin.status === 'enabled' ? 'outline' : 'default'}
            >
              {selectedPlugin.status === 'enabled' ? 'Disable' : 'Enable'} Plugin
            </Button>
            
            <Button 
              on:click={() => uninstallPlugin(selectedPlugin)}
              variant="outline"
              className="text-red-600 hover:text-red-700"
            >
              Uninstall
            </Button>
          </div>
          
          <Button on:click={() => showPluginDetails = false} variant="outline">
            Close
          </Button>
        </div>
      </div>
    </div>
  {/if}
</div>

<!-- Plugin Marketplace Modal -->
<PluginMarketplace 
  isOpen={showMarketplace} 
  onClose={() => {
    showMarketplace = false;
    loadPlugins(); // Reload plugins after marketplace operations
  }} 
/>

<!-- Add Plugin Modal -->
<AddPluginModal 
  isOpen={showAddPlugin} 
  on:close={() => {
    showAddPlugin = false;
    // Optionally reload marketplace to show the newly added plugin
  }} 
/>