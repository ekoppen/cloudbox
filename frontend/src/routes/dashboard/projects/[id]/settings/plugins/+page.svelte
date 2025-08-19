<script lang="ts">
  import { page } from '$app/stores';
  import { onMount } from 'svelte';
  import { API_ENDPOINTS, createApiRequest } from '$lib/config';
  import { auth } from '$lib/stores/auth';
  import { toastStore } from '$lib/stores/toast';
  import Card from '$lib/components/ui/card.svelte';
  import Button from '$lib/components/ui/button.svelte';
  import Badge from '$lib/components/ui/badge.svelte';
  import Icon from '$lib/components/ui/icon.svelte';
  import Modal from '$lib/components/ui/modal.svelte';

  interface Plugin {
    name: string;
    version: string;
    description: string;
    author: string;
    license: string;
    type: string;
    permissions: string[];
    dependencies?: any;
    ui_config?: any;
    is_installed: boolean;
    is_enabled: boolean;
    status: string;
    installed_at?: string;
    installation_status?: string;
    config?: any;
  }

  let availablePlugins: Plugin[] = [];
  let loading = true;
  let error = '';
  
  // Modal states
  let showConfigModal = false;
  let showInstallConfirm = false;
  let showUninstallConfirm = false;
  let selectedPlugin: Plugin | null = null;
  
  // Operations
  let operationInProgress = false;
  let pluginConfig: any = {};

  $: projectId = $page.params.id;

  console.log('🚀 Project plugins page component loaded');

  onMount(async () => {
    console.log('🔄 OnMount called');
    // Wait for auth to be ready
    if (!$auth.token) {
      console.log('Waiting for auth token...');
      await auth.init();
    }
    console.log('Auth token ready:', !!$auth.token);
    loadPlugins();
  });

  async function loadPlugins() {
    loading = true;
    error = '';
    
    try {
      console.log('Loading available plugins for project:', projectId);
      console.log('Auth token exists:', !!$auth.token);
      
      const response = await createApiRequest(API_ENDPOINTS.plugins.projects.available(projectId), {
        headers: {
          'Authorization': `Bearer ${$auth.token}`,
        },
      });

      console.log('Available plugins response status:', response.status);
      
      if (response.ok) {
        const data = await response.json();
        console.log('Available plugins data received:', data);
        availablePlugins = data.plugins || [];
        console.log('Plugins loaded:', availablePlugins.length);
      } else {
        const errorText = await response.text();
        console.error('HTTP error response:', errorText);
        throw new Error(`Failed to load plugins: ${response.status} ${errorText}`);
      }
    } catch (err) {
      console.error('Error loading plugins:', err);
      error = 'Fout bij laden van plugins';
    } finally {
      loading = false;
    }
  }

  async function togglePlugin(plugin: Plugin) {
    if (operationInProgress) return;
    
    operationInProgress = true;
    try {
      if (plugin.is_installed && plugin.is_enabled) {
        // Disable plugin
        const response = await createApiRequest(
          API_ENDPOINTS.plugins.projects.disable(projectId, plugin.name), 
          {
            method: 'POST',
            headers: {
              'Authorization': `Bearer ${$auth.token}`,
            },
          }
        );

        if (response.ok) {
          toastStore.success(`${plugin.name} uitgeschakeld`);
          await loadPlugins();
        } else {
          throw new Error('Uitschakelen mislukt');
        }
      } else if (plugin.is_installed && !plugin.is_enabled) {
        // Enable plugin
        const response = await createApiRequest(
          API_ENDPOINTS.plugins.projects.enable(projectId, plugin.name), 
          {
            method: 'POST',
            headers: {
              'Authorization': `Bearer ${$auth.token}`,
            },
          }
        );

        if (response.ok) {
          toastStore.success(`${plugin.name} ingeschakeld`);
          await loadPlugins();
        } else {
          throw new Error('Inschakelen mislukt');
        }
      } else {
        // Install plugin first
        openInstallConfirm(plugin);
      }
    } catch (err) {
      console.error('Error toggling plugin:', err);
      toastStore.error(`Fout: ${err.message}`);
    } finally {
      operationInProgress = false;
    }
  }

  function openInstallConfirm(plugin: Plugin) {
    selectedPlugin = plugin;
    showInstallConfirm = true;
  }

  async function installPlugin() {
    if (!selectedPlugin || operationInProgress) return;
    
    operationInProgress = true;
    try {
      const response = await createApiRequest(
        API_ENDPOINTS.plugins.projects.install(projectId, selectedPlugin.name),
        {
          method: 'POST',
          headers: {
            'Authorization': `Bearer ${$auth.token}`,
            'Content-Type': 'application/json',
          },
          body: JSON.stringify({
            config: {}
          }),
        }
      );

      if (response.ok) {
        toastStore.success(`${selectedPlugin.name} geïnstalleerd en ingeschakeld`);
        showInstallConfirm = false;
        selectedPlugin = null;
        await loadPlugins();
      } else {
        const errorData = await response.json();
        throw new Error(errorData.error || 'Installatie mislukt');
      }
    } catch (err) {
      console.error('Error installing plugin:', err);
      toastStore.error(`Installatie fout: ${err.message}`);
    } finally {
      operationInProgress = false;
    }
  }

  function openConfigModal(plugin: Plugin) {
    selectedPlugin = plugin;
    pluginConfig = plugin.config || {};
    showConfigModal = true;
  }

  async function updatePluginConfig() {
    if (!selectedPlugin || operationInProgress) return;
    
    operationInProgress = true;
    try {
      const response = await createApiRequest(
        API_ENDPOINTS.plugins.projects.config(projectId, selectedPlugin.name),
        {
          method: 'PUT',
          headers: {
            'Authorization': `Bearer ${$auth.token}`,
            'Content-Type': 'application/json',
          },
          body: JSON.stringify(pluginConfig),
        }
      );

      if (response.ok) {
        toastStore.success('Plugin configuratie bijgewerkt');
        showConfigModal = false;
        selectedPlugin = null;
        await loadPlugins();
      } else {
        const errorData = await response.json();
        throw new Error(errorData.error || 'Configuratie bijwerken mislukt');
      }
    } catch (err) {
      console.error('Error updating plugin config:', err);
      toastStore.error(`Configuratie fout: ${err.message}`);
    } finally {
      operationInProgress = false;
    }
  }

  function openUninstallConfirm(plugin: Plugin) {
    selectedPlugin = plugin;
    showUninstallConfirm = true;
  }

  async function uninstallPlugin() {
    if (!selectedPlugin || operationInProgress) return;
    
    operationInProgress = true;
    try {
      const response = await createApiRequest(
        API_ENDPOINTS.plugins.projects.uninstall(projectId, selectedPlugin.name),
        {
          method: 'DELETE',
          headers: {
            'Authorization': `Bearer ${$auth.token}`,
          },
        }
      );

      if (response.ok) {
        toastStore.success(`${selectedPlugin.name} verwijderd`);
        showUninstallConfirm = false;
        selectedPlugin = null;
        await loadPlugins();
      } else {
        const errorData = await response.json();
        throw new Error(errorData.error || 'Verwijderen mislukt');
      }
    } catch (err) {
      console.error('Error uninstalling plugin:', err);
      toastStore.error(`Verwijderen fout: ${err.message}`);
    } finally {
      operationInProgress = false;
    }
  }

  function getStatusBadgeVariant(plugin: Plugin): "default" | "secondary" | "destructive" | "outline" {
    if (!plugin.is_installed) return "outline";
    if (plugin.is_enabled) return "default";
    return "secondary";
  }

  function getStatusText(plugin: Plugin): string {
    if (!plugin.is_installed) return "Niet geïnstalleerd";
    if (plugin.is_enabled) return "Ingeschakeld";
    return "Uitgeschakeld";
  }

  function getCategoryIcon(type: string): string {
    switch (type) {
      case 'dashboard-plugin': return 'dashboard';
      case 'service-plugin': return 'settings';
      case 'auth-plugin': return 'auth';
      case 'storage-plugin': return 'storage';
      default: return 'package';
    }
  }
</script>

<svelte:head>
  <title>Plugins - CloudBox</title>
</svelte:head>

<!-- Plugin Status Summary -->
<div class="flex justify-between items-center mb-6">
  <div>
    <h2 class="text-lg font-medium text-foreground">Plugin Management</h2>
    <p class="text-sm text-muted-foreground">
      Beheer plugins voor dit project
    </p>
  </div>
  <div class="text-right">
    <p class="text-sm text-muted-foreground">
      {availablePlugins.filter(p => p.is_enabled).length} van {availablePlugins.length} plugins ingeschakeld
    </p>
  </div>
</div>

<div class="space-y-6">

  <!-- Error Message -->
  {#if error}
    <Card class="bg-destructive/10 border-destructive/20 p-4">
      <div class="flex justify-between items-center">
        <p class="text-destructive text-sm">{error}</p>
        <Button variant="ghost" size="sm" on:click={() => error = ''}>×</Button>
      </div>
    </Card>
  {/if}

  <!-- Loading State -->
  {#if loading}
    <div class="flex items-center justify-center min-h-64">
      <div class="text-center">
        <div class="animate-spin rounded-full h-8 w-8 border-b-2 border-primary mx-auto"></div>
        <p class="mt-4 text-muted-foreground">Plugins laden...</p>
      </div>
    </div>
  {:else}
    <!-- Plugins List -->
    {#if availablePlugins.length === 0}
      <Card class="p-8 text-center">
        <Icon name="package" size={48} className="mx-auto text-muted-foreground mb-4" />
        <h3 class="text-lg font-medium text-foreground mb-2">Geen plugins beschikbaar</h3>
        <p class="text-muted-foreground">
          Er zijn geen plugins ingeschakeld door de administrator.
        </p>
      </Card>
    {:else}
      <Card class="divide-y divide-border">
        {#each availablePlugins as plugin}
          <div class="p-6 flex items-center justify-between">
            <!-- Plugin Info -->
            <div class="flex items-center space-x-4 flex-1">
              <div class="w-12 h-12 bg-primary/10 rounded-lg flex items-center justify-center">
                <Icon name={getCategoryIcon(plugin.type)} size={24} className="text-primary" />
              </div>
              <div class="flex-1">
                <div class="flex items-center space-x-3 mb-1">
                  <h3 class="font-medium text-foreground">{plugin.name}</h3>
                  <Badge variant={getStatusBadgeVariant(plugin)} class="text-xs">
                    {getStatusText(plugin)}
                  </Badge>
                </div>
                <p class="text-sm text-muted-foreground mb-2">{plugin.description}</p>
                <div class="flex items-center space-x-4 text-xs text-muted-foreground">
                  <span>v{plugin.version}</span>
                  <span>•</span>
                  <span>{plugin.author}</span>
                  <span>•</span>
                  <span>{plugin.license}</span>
                </div>
              </div>
            </div>

            <!-- Actions -->
            <div class="flex items-center space-x-2">
              {#if plugin.is_installed && plugin.is_enabled}
                <!-- Plugin is enabled -->
                <Button
                  variant="outline" 
                  size="sm"
                  on:click={() => openConfigModal(plugin)}
                  disabled={operationInProgress}
                >
                  <Icon name="settings" size={14} />
                  Config
                </Button>
                <Button
                  variant="outline"
                  size="sm"
                  on:click={() => togglePlugin(plugin)}
                  disabled={operationInProgress}
                  class="text-orange-600 border-orange-600 hover:bg-orange-50"
                >
                  <Icon name="stop" size={14} />
                  Uitschakelen
                </Button>
                <Button
                  variant="outline"
                  size="sm"
                  on:click={() => openUninstallConfirm(plugin)}
                  disabled={operationInProgress}
                  class="text-red-600 border-red-600 hover:bg-red-50"
                >
                  <Icon name="trash" size={14} />
                </Button>
              {:else if plugin.is_installed}
                <!-- Plugin is installed but disabled -->
                <Button
                  variant="outline"
                  size="sm"
                  on:click={() => togglePlugin(plugin)}
                  disabled={operationInProgress}
                  class="text-green-600 border-green-600 hover:bg-green-50"
                >
                  <Icon name="play" size={14} />
                  Inschakelen
                </Button>
                <Button
                  variant="outline"
                  size="sm"
                  on:click={() => openUninstallConfirm(plugin)}
                  disabled={operationInProgress}
                  class="text-red-600 border-red-600 hover:bg-red-50"
                >
                  <Icon name="trash" size={14} />
                </Button>
              {:else}
                <!-- Plugin is not installed -->
                <Button
                  size="sm"
                  on:click={() => togglePlugin(plugin)}
                  disabled={operationInProgress}
                  class="bg-green-600 hover:bg-green-700 text-white"
                >
                  <Icon name="plus" size={14} />
                  Installeren
                </Button>
              {/if}
            </div>
          </div>
        {/each}
      </Card>
    {/if}
  {/if}
</div>

<!-- Install Confirmation Modal -->
{#if showInstallConfirm && selectedPlugin}
  <Modal
    title="Plugin Installeren"
    open={showInstallConfirm}
    on:close={() => { showInstallConfirm = false; selectedPlugin = null; }}
  >
    <div class="space-y-4">
      <div class="flex items-center space-x-4 p-4 bg-muted rounded-lg">
        <div class="w-12 h-12 bg-primary/10 rounded-lg flex items-center justify-center">
          <Icon name={getCategoryIcon(selectedPlugin.type)} size={20} className="text-primary" />
        </div>
        <div>
          <h3 class="font-medium text-foreground">{selectedPlugin.name}</h3>
          <p class="text-sm text-muted-foreground">{selectedPlugin.description}</p>
          <div class="flex items-center space-x-2 mt-1 text-xs text-muted-foreground">
            <span>v{selectedPlugin.version}</span>
            <span>•</span>
            <span>{selectedPlugin.author}</span>
          </div>
        </div>
      </div>

      {#if selectedPlugin.permissions && selectedPlugin.permissions.length > 0}
        <div>
          <h4 class="text-sm font-medium mb-2">Vereiste permissies:</h4>
          <ul class="text-sm text-muted-foreground space-y-1">
            {#each selectedPlugin.permissions as permission}
              <li class="flex items-center space-x-2">
                <Icon name="auth" size={12} />
                <span>{permission}</span>
              </li>
            {/each}
          </ul>
        </div>
      {/if}

      <p class="text-sm text-muted-foreground">
        Deze plugin wordt geïnstalleerd en automatisch ingeschakeld voor dit project.
      </p>
    </div>

    <svelte:fragment slot="footer">
      <Button
        variant="outline"
        on:click={() => { showInstallConfirm = false; selectedPlugin = null; }}
        disabled={operationInProgress}
      >
        Annuleren
      </Button>
      <Button
        on:click={installPlugin}
        disabled={operationInProgress}
        class="flex items-center space-x-2"
      >
        {#if operationInProgress}
          <div class="animate-spin rounded-full h-4 w-4 border-b-2 border-current"></div>
        {:else}
          <Icon name="plus" size={16} />
        {/if}
        <span>{operationInProgress ? 'Installeren...' : 'Installeren'}</span>
      </Button>
    </svelte:fragment>
  </Modal>
{/if}

<!-- Configure Plugin Modal -->
{#if showConfigModal && selectedPlugin}
  <Modal
    title="Plugin Configureren"
    open={showConfigModal}
    on:close={() => { showConfigModal = false; selectedPlugin = null; }}
  >
    <div class="space-y-4">
      <div class="flex items-center space-x-4 p-4 bg-muted rounded-lg">
        <div class="w-12 h-12 bg-primary/10 rounded-lg flex items-center justify-center">
          <Icon name={getCategoryIcon(selectedPlugin.type)} size={20} className="text-primary" />
        </div>
        <div>
          <h3 class="font-medium text-foreground">{selectedPlugin.name}</h3>
          <p class="text-sm text-muted-foreground">Configuratie voor dit project</p>
        </div>
      </div>

      <div>
        <label for="plugin-config" class="block text-sm font-medium mb-2">Configuratie (JSON)</label>
        <textarea
          id="plugin-config"
          bind:value={pluginConfig}
          rows="8"
          class="w-full p-3 text-sm font-mono bg-muted border border-border rounded-lg focus:outline-none focus:ring-2 focus:ring-primary focus:border-transparent"
          placeholder="Plugin configuratie in JSON formaat"
        ></textarea>
        <p class="mt-1 text-xs text-muted-foreground">
          Configureer de plugin instellingen specifiek voor dit project
        </p>
      </div>
    </div>

    <svelte:fragment slot="footer">
      <Button
        variant="outline"
        on:click={() => { showConfigModal = false; selectedPlugin = null; }}
        disabled={operationInProgress}
      >
        Annuleren
      </Button>
      <Button
        on:click={updatePluginConfig}
        disabled={operationInProgress}
        class="flex items-center space-x-2"
      >
        {#if operationInProgress}
          <div class="animate-spin rounded-full h-4 w-4 border-b-2 border-current"></div>
        {:else}
          <Icon name="settings" size={16} />
        {/if}
        <span>{operationInProgress ? 'Opslaan...' : 'Opslaan'}</span>
      </Button>
    </svelte:fragment>
  </Modal>
{/if}

<!-- Uninstall Confirmation Modal -->
{#if showUninstallConfirm && selectedPlugin}
  <Modal
    title="Plugin Verwijderen"
    open={showUninstallConfirm}
    on:close={() => { showUninstallConfirm = false; selectedPlugin = null; }}
  >
    <div class="space-y-4">
      <div class="flex items-center space-x-4 p-4 bg-destructive/10 border border-destructive/20 rounded-lg">
        <Icon name="trash" size={24} className="text-destructive" />
        <div>
          <h3 class="font-medium text-destructive">
            {selectedPlugin.name} verwijderen?
          </h3>
          <p class="text-sm text-muted-foreground">
            Deze actie kan niet ongedaan gemaakt worden.
          </p>
        </div>
      </div>

      <div class="text-sm text-muted-foreground space-y-2">
        <p>Door het verwijderen van deze plugin:</p>
        <ul class="list-disc list-inside space-y-1 ml-4">
          <li>Worden alle plugin-specifieke instellingen verwijderd</li>
          <li>Kan functionaliteit die afhankelijk is van deze plugin stoppen met werken</li>
          <li>Blijven eventuele gegevens die door de plugin zijn aangemaakt behouden</li>
        </ul>
      </div>
    </div>

    <svelte:fragment slot="footer">
      <Button
        variant="outline"
        on:click={() => { showUninstallConfirm = false; selectedPlugin = null; }}
        disabled={operationInProgress}
      >
        Annuleren
      </Button>
      <Button
        variant="destructive"
        on:click={uninstallPlugin}
        disabled={operationInProgress}
        class="flex items-center space-x-2"
      >
        {#if operationInProgress}
          <div class="animate-spin rounded-full h-4 w-4 border-b-2 border-current"></div>
        {:else}
          <Icon name="trash" size={16} />
        {/if}
        <span>{operationInProgress ? 'Verwijderen...' : 'Verwijderen'}</span>
      </Button>
    </svelte:fragment>
  </Modal>
{/if}