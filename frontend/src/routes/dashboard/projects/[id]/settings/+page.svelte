<script lang="ts">
  import { page } from '$app/stores';
  import { onMount } from 'svelte';
  import { API_BASE_URL, API_ENDPOINTS, createApiRequest } from '$lib/config';
  import { auth } from '$lib/stores/auth';
  import { toastStore } from '$lib/stores/toast';
  import { goto } from '$app/navigation';
  import Card from '$lib/components/ui/card.svelte';
  import Button from '$lib/components/ui/button.svelte';
  import Input from '$lib/components/ui/input.svelte';
  import Label from '$lib/components/ui/label.svelte';
  import Badge from '$lib/components/ui/badge.svelte';
  import Textarea from '$lib/components/ui/textarea.svelte';
  import Icon from '$lib/components/ui/icon.svelte';

  interface APIKey {
    id: number;
    name: string;
    key: string;
    created_at: string;
    last_used_at?: string;
    permissions: string[];
    is_active: boolean;
    is_new_key?: boolean; // Flag to indicate if this is a newly created key with full key visible
  }

  interface CORSConfig {
    allowed_origins: string[];
    allowed_methods: string[];
    allowed_headers: string[];
    allow_credentials: boolean;
    max_age: number;
  }


  let apiKeys: APIKey[] = [];
  let apiKeysLoaded = false;

  let corsConfig: CORSConfig = {
    allowed_origins: ['https://place.holder'],
    allowed_methods: ['GET', 'POST', 'PUT', 'DELETE', 'OPTIONS'],
    allowed_headers: ['Content-Type', 'Authorization', 'X-API-Key'],
    allow_credentials: true,
    max_age: 3600
  };


  let showCreateKey = false;
  let showKeyDetails: APIKey | null = null;
  let newKeyName = '';
  let newKeyPermissions: string[] = [];
  let loading = false;
  let error = '';
  let activeTab = 'api-keys';
  let showDeleteConfirm = false;

  // CORS form fields
  let newOrigin = '';
  let corsFormData = { ...corsConfig };
  let allowedHeadersString = corsConfig.allowed_headers.join(', ');
  
  // Project notes
  let projectNotes = '';
  let savingNotes = false;

  $: projectId = $page.params.id;
  
  // Fallback: extract project ID from URL if params.id is undefined
  $: {
    if (!projectId || projectId === 'undefined') {
      const pathParts = $page.url.pathname.split('/');
      const projectIndex = pathParts.indexOf('projects');
      if (projectIndex !== -1 && pathParts[projectIndex + 1] && pathParts[projectIndex + 1] !== 'undefined') {
        projectId = pathParts[projectIndex + 1];
        console.log('Fallback project ID from URL:', projectId);
      } else {
        // Als laatste redmiddel, vraag gebruiker om handmatig te navigeren
        console.error('Could not determine valid project ID from URL:', $page.url.pathname);
      }
    }
  }
  
  // Debug reactive project ID changes
  $: {
    console.log('Reactive update - projectId:', projectId);
    console.log('Reactive update - page params:', $page.params);
    console.log('Reactive update - page url:', $page.url.pathname);
  }

  // Load API keys, CORS config when projectId changes
  $: {
    if (projectId && projectId !== 'undefined' && !apiKeysLoaded) {
      loadAPIKeys();
      loadCORSConfig();
      loadProjectNotes();
    }
  }

  async function loadAPIKeys() {
    if (!projectId || projectId === 'undefined') return;
    
    try {
      console.log('Loading API keys for project:', projectId);
      const response = await createApiRequest(`${API_BASE_URL}/api/v1/projects/${projectId}/api-keys`, {
        method: 'GET',
        headers: {
          'Authorization': `Bearer ${$auth.token}`,
        },
      });

      if (response.ok) {
        const data = await response.json();
        console.log('Loaded API keys:', data);
        apiKeys = data.map(key => ({
          id: key.id,
          name: key.name,
          key: key.key, // This will be masked on the backend for security
          created_at: key.created_at,
          last_used_at: key.last_used_at,
          permissions: key.permissions || [],
          is_active: key.is_active
        }));
        apiKeysLoaded = true;
      } else {
        console.error('Failed to load API keys:', response.status, response.statusText);
        const errorData = await response.json();
        console.error('Error data:', errorData);
        // Set loaded to true even if no keys exist, to stop spinner
        apiKeysLoaded = true;
      }
    } catch (err) {
      console.error('Error loading API keys:', err);
      // Set loaded to true even on error, to stop spinner
      apiKeysLoaded = true;
    }
  }

  async function loadCORSConfig() {
    if (!projectId || projectId === 'undefined') return;
    
    try {
      console.log('Loading CORS config for project:', projectId);
      const response = await createApiRequest(`${API_BASE_URL}/api/v1/projects/${projectId}/cors`, {
        method: 'GET',
        headers: {
          'Authorization': `Bearer ${$auth.token}`,
        },
      });

      if (response.ok) {
        const data = await response.json();
        console.log('Loaded CORS config:', data);
        corsConfig = {
          allowed_origins: data.allowed_origins || ['*'],
          allowed_methods: data.allowed_methods || ['GET', 'POST', 'PUT', 'DELETE', 'OPTIONS'],
          allowed_headers: data.allowed_headers || ['Content-Type', 'Authorization', 'X-API-Key'],
          allow_credentials: data.allow_credentials || false,
          max_age: data.max_age || 3600
        };
        corsFormData = { ...corsConfig };
        allowedHeadersString = corsConfig.allowed_headers.join(', ');
      } else {
        console.error('Failed to load CORS config:', response.status, response.statusText);
        // Use defaults if no CORS config exists yet
        console.log('Using default CORS config');
      }
    } catch (err) {
      console.error('Error loading CORS config:', err);
    }
  }

  function generateAPIKey(): string {
    const chars = 'abcdefghijklmnopqrstuvwxyz0123456789';
    let result = 'cb_live_';
    for (let i = 0; i < 24; i++) {
      result += chars.charAt(Math.floor(Math.random() * chars.length));
    }
    return result;
  }

  async function createAPIKey() {
    if (!newKeyName.trim()) return;

    loading = true;
    try {
      console.log('Creating API key for project:', projectId);
      const response = await createApiRequest(`${API_BASE_URL}/api/v1/projects/${projectId}/api-keys`, {
        method: 'POST',
        headers: {
          'Authorization': `Bearer ${$auth.token}`,
          'Content-Type': 'application/json',
        },
        body: JSON.stringify({
          name: newKeyName,
          permissions: newKeyPermissions.length > 0 ? newKeyPermissions : ['read', 'write']
        }),
      });

      if (response.ok) {
        const newKey = await response.json();
        console.log('Created API key:', newKey);
        
        apiKeys = [...apiKeys, {
          id: newKey.id,
          name: newKey.name,
          key: newKey.key,
          created_at: newKey.created_at,
          last_used_at: newKey.last_used_at,
          permissions: newKey.permissions || [],
          is_active: newKey.is_active,
          is_new_key: true // Mark as newly created key
        }];
        
        showCreateKey = false;
        newKeyName = '';
        newKeyPermissions = [];
        showKeyDetails = apiKeys[apiKeys.length - 1]; // Show the new key
        toastStore.success('API key succesvol aangemaakt');
      } else {
        const errorData = await response.json();
        console.error('Failed to create API key:', errorData);
        error = errorData.error || 'Fout bij aanmaken van API key';
        toastStore.error(error);
      }
    } catch (err) {
      console.error('Error creating API key:', err);
      error = 'Netwerkfout bij aanmaken van API key';
      toastStore.error(error);
    } finally {
      loading = false;
    }
  }

  function toggleKeyStatus(keyId: number) {
    apiKeys = apiKeys.map(key => 
      key.id === keyId ? { ...key, is_active: !key.is_active } : key
    );
  }

  async function deleteKey(keyId: number) {
    if (confirm('Weet je zeker dat je deze API key wilt verwijderen?')) {
      try {
        console.log('Deleting API key:', keyId, 'for project:', projectId);
        const response = await createApiRequest(`${API_BASE_URL}/api/v1/projects/${projectId}/api-keys/${keyId}`, {
          method: 'DELETE',
          headers: {
            'Authorization': `Bearer ${$auth.token}`,
          },
        });

        if (response.ok) {
          apiKeys = apiKeys.filter(key => key.id !== keyId);
          toastStore.success('API key succesvol verwijderd');
        } else {
          const data = await response.json();
          toastStore.error(data.error || 'Fout bij verwijderen van API key');
        }
      } catch (err) {
        console.error('Error deleting API key:', err);
        toastStore.error('Netwerkfout bij verwijderen van API key');
      }
    }
  }

  function showDeleteProjectDialog() {
    showDeleteConfirm = true;
  }

  async function confirmDeleteProject() {
    console.log('confirmDeleteProject called');
    console.log('Raw projectId:', projectId);
    console.log('projectId type:', typeof projectId);
    console.log('projectId length:', projectId?.length);
    console.log('URL params:', $page.params);
    console.log('Full URL:', $page.url.pathname);
    
    // Safety check for undefined project ID
    if (!projectId || projectId === 'undefined') {
      console.error('Project ID is undefined or invalid');
      toastStore.error('Ongeldige project ID - probeer de pagina te vernieuwen');
      showDeleteConfirm = false;
      return;
    }
    
    showDeleteConfirm = false;
    loading = true;
    
    try {
      const deleteUrl = API_ENDPOINTS.projects.delete(projectId);
      console.log('Delete URL:', deleteUrl);
      console.log('Auth token exists:', !!$auth.token);
      console.log('Auth token length:', $auth.token?.length || 0);
      console.log('Auth store:', $auth);
      console.log('User role:', $auth.user?.role);
      console.log('User:', $auth.user);
      
      const response = await createApiRequest(deleteUrl, {
        method: 'DELETE',
        headers: {
          'Authorization': `Bearer ${$auth.token}`,
          'Content-Type': 'application/json',
        },
      });

      console.log('Delete response status:', response.status);
      console.log('Delete response statusText:', response.statusText);
      console.log('Delete response headers:', Object.fromEntries(response.headers.entries()));

      if (response.ok) {
        toastStore.success('Project succesvol verwijderd');
        console.log('Project deleted successfully, redirecting...');
        setTimeout(() => {
          goto('/dashboard');
        }, 1500);
      } else {
        let errorMessage = 'Fout bij verwijderen van project';
        try {
          const data = await response.json();
          console.error('Delete failed with data:', data);
          errorMessage = data.error || data.message || errorMessage;
        } catch (parseError) {
          console.error('Could not parse error response:', parseError);
          console.error('Raw response text:', await response.text());
        }
        toastStore.error(errorMessage);
      }
    } catch (err) {
      console.error('Delete error (full object):', err);
      console.error('Delete error message:', err.message);
      console.error('Delete error stack:', err.stack);
      toastStore.error(`Netwerkfout bij verwijderen van project: ${err.message}`);
    } finally {
      loading = false;
    }
  }

  function copyToClipboard(text: string, isNewKey: boolean = false) {
    if (!isNewKey && text.includes('...')) {
      toastStore.error('Deze API key is gemaskeerd en kan niet gekopieerd worden. Alleen nieuwe keys kunnen gekopieerd worden.');
      return;
    }
    navigator.clipboard.writeText(text);
    toastStore.success('API key gekopieerd naar klembord!');
  }

  function addOrigin() {
    if (newOrigin.trim() && !corsFormData.allowed_origins.includes(newOrigin.trim())) {
      corsFormData.allowed_origins = [...corsFormData.allowed_origins, newOrigin.trim()];
      newOrigin = '';
    }
  }

  function removeOrigin(origin: string) {
    corsFormData.allowed_origins = corsFormData.allowed_origins.filter(o => o !== origin);
  }

  async function saveCORSConfig() {
    if (!projectId || projectId === 'undefined') {
      toastStore.error('Ongeldige project ID');
      return;
    }

    loading = true;
    try {
      console.log('Saving CORS config for project:', projectId);
      const response = await createApiRequest(`${API_BASE_URL}/api/v1/projects/${projectId}/cors`, {
        method: 'PUT',
        headers: {
          'Authorization': `Bearer ${$auth.token}`,
          'Content-Type': 'application/json',
        },
        body: JSON.stringify({
          allowed_origins: corsFormData.allowed_origins,
          allowed_methods: corsFormData.allowed_methods,
          allowed_headers: allowedHeadersString.split(',').map(h => h.trim()).filter(h => h.length > 0),
          allow_credentials: corsFormData.allow_credentials,
          max_age: corsFormData.max_age
        }),
      });

      if (response.ok) {
        corsConfig = { 
          ...corsFormData, 
          allowed_headers: allowedHeadersString.split(',').map(h => h.trim()).filter(h => h.length > 0)
        };
        console.log('CORS config saved successfully');
        toastStore.success('CORS configuratie succesvol opgeslagen!');
      } else {
        const errorData = await response.json();
        console.error('Failed to save CORS config:', errorData);
        toastStore.error(errorData.error || 'Fout bij opslaan van CORS configuratie');
      }
    } catch (err) {
      console.error('Error saving CORS config:', err);
      toastStore.error('Netwerkfout bij opslaan van CORS configuratie');
    } finally {
      loading = false;
    }
  }

  function resetCORSConfig() {
    corsFormData = { ...corsConfig };
    allowedHeadersString = corsConfig.allowed_headers.join(', ');
  }

  // Project notes functions
  async function loadProjectNotes() {
    if (!projectId || projectId === 'undefined') return;
    
    try {
      console.log('Loading project notes for project:', projectId);
      const response = await createApiRequest(`${API_BASE_URL}/api/v1/projects/${projectId}/notes`, {
        method: 'GET',
        headers: {
          'Authorization': `Bearer ${$auth.token}`,
        },
      });

      if (response.ok) {
        const data = await response.json();
        projectNotes = data.notes || '';
        console.log('Loaded project notes:', projectNotes);
      } else {
        console.error('Failed to load project notes:', response.status, response.statusText);
        // Use empty string as fallback if no notes exist yet
        projectNotes = '';
      }
    } catch (err) {
      console.error('Error loading project notes:', err);
      projectNotes = '';
    }
  }

  async function saveProjectNotes() {
    if (!projectId || projectId === 'undefined') {
      toastStore.error('Ongeldige project ID');
      return;
    }

    savingNotes = true;
    try {
      console.log('Saving project notes for project:', projectId);
      const response = await createApiRequest(`${API_BASE_URL}/api/v1/projects/${projectId}/notes`, {
        method: 'PUT',
        headers: {
          'Authorization': `Bearer ${$auth.token}`,
          'Content-Type': 'application/json',
        },
        body: JSON.stringify({
          notes: projectNotes
        }),
      });

      if (response.ok) {
        console.log('Project notes saved successfully');
        toastStore.success('Project notities opgeslagen');
      } else {
        const errorData = await response.json();
        console.error('Failed to save project notes:', errorData);
        toastStore.error(errorData.error || 'Fout bij opslaan van notities');
      }
    } catch (err) {
      console.error('Error saving project notes:', err);
      toastStore.error('Netwerkfout bij opslaan van notities');
    } finally {
      savingNotes = false;
    }
  }

  function formatDate(dateStr: string): string {
    return new Date(dateStr).toLocaleDateString('nl-NL', {
      year: 'numeric',
      month: 'short',
      day: 'numeric',
      hour: '2-digit',
      minute: '2-digit'
    });
  }

  function getPermissionColor(permission: string): string {
    switch (permission) {
      case 'read': return 'bg-green-100 dark:bg-green-900 text-green-800 dark:text-green-200';
      case 'write': return 'bg-blue-100 dark:bg-blue-900 text-blue-800 dark:text-blue-200';
      case 'delete': return 'bg-red-100 dark:bg-red-900 text-red-800 dark:text-red-200';
      default: return 'bg-muted text-muted-foreground';
    }
  }

</script>

<svelte:head>
  <title>Instellingen - CloudBox</title>
</svelte:head>

<!-- Page Header -->
<div class="mb-8">
  <h1 class="text-3xl font-bold text-foreground font-['Inter'] mb-2">Project Settings</h1>
  <p class="text-muted-foreground text-base">Configure your project's API keys, CORS settings, and general preferences</p>
</div>

<!-- Sub-tabs within settings -->
<div class="border-b border-border mb-8">
  <nav class="flex space-x-8">
    <button
      on:click={() => activeTab = 'api-keys'}
      class="flex items-center space-x-2 py-3 px-1 border-b-2 text-sm font-semibold transition-all duration-200 {
        activeTab === 'api-keys' 
          ? 'border-primary text-primary' 
          : 'border-transparent text-muted-foreground hover:text-foreground hover:border-muted'
      }"
    >
      <Icon name="key" size={18} />
      <span>API Keys</span>
    </button>
    <button
      on:click={() => activeTab = 'cors'}
      class="flex items-center space-x-2 py-3 px-1 border-b-2 text-sm font-semibold transition-all duration-200 {
        activeTab === 'cors' 
          ? 'border-primary text-primary' 
          : 'border-transparent text-muted-foreground hover:text-foreground hover:border-muted'
      }"
    >
      <Icon name="globe" size={18} />
      <span>CORS</span>
    </button>
    <button
      on:click={() => activeTab = 'general'}
      class="flex items-center space-x-2 py-3 px-1 border-b-2 text-sm font-semibold transition-all duration-200 {
        activeTab === 'general' 
          ? 'border-primary text-primary' 
          : 'border-transparent text-muted-foreground hover:text-foreground hover:border-muted'
      }"
    >
      <Icon name="settings" size={18} />
      <span>General</span>
    </button>
  </nav>
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

  <!-- API Keys Tab -->
  {#if activeTab === 'api-keys'}
    <div class="space-y-8">
      <div class="flex justify-between items-center">
        <div>
          <h2 class="text-2xl font-semibold text-foreground">API Keys</h2>
          <p class="text-muted-foreground mt-1">Manage access to your project API</p>
        </div>
        <Button on:click={() => showCreateKey = true} className="h-10 px-6">
          <Icon name="plus" size={16} className="mr-2" />
          <span>Create API Key</span>
        </Button>
      </div>

      <div class="bg-background border border-border rounded-2xl overflow-hidden">
        {#if !apiKeysLoaded}
          <div class="p-16 text-center">
            <div class="w-12 h-12 bg-primary/10 rounded-xl flex items-center justify-center mx-auto mb-4">
              <div class="animate-spin rounded-full h-6 w-6 border-b-2 border-primary"></div>
            </div>
            <p class="text-foreground font-medium">Loading API keys...</p>
            <p class="text-sm text-muted-foreground mt-1">Please wait while we fetch your keys</p>
          </div>
        {:else if apiKeys.length === 0}
          <div class="p-16 text-center">
            <div class="w-16 h-16 bg-muted rounded-2xl flex items-center justify-center mx-auto mb-6">
              <Icon name="key" size={32} className="text-muted-foreground" />
            </div>
            <h3 class="text-lg font-semibold text-foreground mb-2">No API keys created</h3>
            <p class="text-muted-foreground mb-6 max-w-sm mx-auto">
              Create your first API key to start accessing your project programmatically.
            </p>
            <Button on:click={() => showCreateKey = true} className="px-8">
              <Icon name="plus" size={16} className="mr-2" />
              Create API Key
            </Button>
          </div>
        {:else}
          <div class="divide-y divide-border">
            {#each apiKeys as apiKey}
              <div class="px-8 py-6 hover:bg-muted/30 transition-colors">
                <div class="flex items-center justify-between">
                  <div class="flex items-center space-x-4 flex-1">
                    <div class="w-12 h-12 bg-primary/10 rounded-xl flex items-center justify-center">
                      <Icon name="key" size={20} className="text-primary" />
                    </div>
                    <div class="flex-1">
                      <div class="flex items-center space-x-3 mb-2">
                        <h3 class="text-lg font-semibold text-foreground">{apiKey.name}</h3>
                        <Badge variant={apiKey.is_active ? "default" : "secondary"} className="font-medium">
                          {apiKey.is_active ? 'Active' : 'Inactive'}
                        </Badge>
                      </div>
                      
                      <div class="flex items-center space-x-2 mb-3">
                        {#each apiKey.permissions as permission}
                          <span class="inline-flex px-2 py-1 text-xs font-medium rounded-full {getPermissionColor(permission)}">
                            {permission}
                          </span>
                        {/each}
                      </div>
                      
                      <div class="flex items-center space-x-6 text-sm text-muted-foreground">
                        <div class="flex items-center space-x-1">
                          <Icon name="calendar" size={12} />
                          <span>Created {formatDate(apiKey.created_at)}</span>
                        </div>
                        {#if apiKey.last_used_at}
                          <div class="flex items-center space-x-1">
                            <Icon name="clock" size={12} />
                            <span>Last used {formatDate(apiKey.last_used_at)}</span>
                          </div>
                        {:else}
                          <div class="flex items-center space-x-1">
                            <Icon name="clock" size={12} />
                            <span>Never used</span>
                          </div>
                        {/if}
                      </div>
                    </div>
                  </div>

                  <div class="flex items-center space-x-3">
                    <Button
                      variant="ghost"
                      size="sm"
                      on:click={() => showKeyDetails = apiKey}
                      className="text-muted-foreground hover:text-foreground"
                    >
                      <Icon name="eye" size={16} className="mr-2" />
                      View
                    </Button>
                    <Button
                      variant="outline"
                      size="sm"
                      on:click={() => toggleKeyStatus(apiKey.id)}
                    >
                      {apiKey.is_active ? 'Disable' : 'Enable'}
                    </Button>
                    <Button
                      variant="ghost"
                      size="sm"
                      on:click={() => deleteKey(apiKey.id)}
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
    </div>
  {/if}

  <!-- CORS Tab -->
  {#if activeTab === 'cors'}
    <div class="space-y-6">
      <div>
        <h2 class="text-lg font-medium text-foreground">CORS Configuratie</h2>
        <p class="text-sm text-muted-foreground">Configureer Cross-Origin Resource Sharing voor je API</p>
      </div>

      <Card class="p-6">
        <form on:submit|preventDefault={saveCORSConfig} class="space-y-6">
          <!-- Allowed Origins -->
          <div>
            <Label>Toegestane Origins</Label>
            <div class="mt-2 space-y-2">
              {#each corsFormData.allowed_origins as origin}
                <div class="flex items-center justify-between bg-muted px-3 py-2 rounded">
                  <code class="text-sm text-foreground">{origin}</code>
                  <Button
                    variant="ghost"
                    size="sm"
                    on:click={() => removeOrigin(origin)}
                    class="text-destructive hover:text-destructive"
                  >
                    ×
                  </Button>
                </div>
              {/each}
              
              <div class="flex space-x-2">
                <Input
                  bind:value={newOrigin}
                  placeholder="https://mijnapp.nl"
                  class="flex-1"
                />
                <Button
                  type="button"
                  variant="outline"
                  on:click={addOrigin}
                >
                  Toevoegen
                </Button>
              </div>
            </div>
          </div>

          <!-- Allowed Methods -->
          <div>
            <Label>Toegestane Methods</Label>
            <div class="mt-2 grid grid-cols-3 gap-3">
              {#each ['GET', 'POST', 'PUT', 'DELETE', 'OPTIONS', 'HEAD'] as method}
                <label class="flex items-center">
                  <input
                    type="checkbox"
                    bind:group={corsFormData.allowed_methods}
                    value={method}
                    class="rounded border-border text-primary focus:ring-primary"
                  />
                  <span class="ml-2 text-sm font-medium text-foreground">{method}</span>
                </label>
              {/each}
            </div>
          </div>

          <!-- Allowed Headers -->
          <div>
            <Label>Toegestane Headers</Label>
            <div class="mt-2">
              <Textarea
                bind:value={allowedHeadersString}
                rows={3}
                placeholder="Content-Type, Authorization, X-API-Key"
              />
              <p class="mt-1 text-xs text-muted-foreground">Gescheiden door komma's</p>
            </div>
          </div>

          <!-- Other Settings -->
          <div class="grid grid-cols-1 md:grid-cols-2 gap-6">
            <div>
              <label class="flex items-center">
                <input
                  type="checkbox"
                  bind:checked={corsFormData.allow_credentials}
                  class="rounded border-border text-primary focus:ring-primary"
                />
                <span class="ml-2 text-sm font-medium text-foreground">Credentials toestaan</span>
              </label>
            </div>

            <div>
              <Label>Max Age (seconden)</Label>
              <Input
                type="number"
                bind:value={corsFormData.max_age}
                min="0"
                class="mt-1"
              />
            </div>
          </div>

          <div class="flex justify-end space-x-3">
            <Button
              type="button"
              variant="outline"
              on:click={resetCORSConfig}
            >
              Reset
            </Button>
            <Button type="submit">
              Opslaan
            </Button>
          </div>
        </form>
      </Card>
    </div>
  {/if}


  <!-- General Tab -->
  {#if activeTab === 'general'}
    <div class="space-y-6">
      <div>
        <h2 class="text-lg font-medium text-foreground">Algemene Instellingen</h2>
        <p class="text-sm text-muted-foreground">Project notities en andere algemene instellingen</p>
      </div>

      <Card class="p-6">
        <form on:submit|preventDefault={saveProjectNotes} class="space-y-6">
          <div>
            <Label for="project-notes">Project Notities</Label>
            <Textarea
              id="project-notes"
              bind:value={projectNotes}
              class="mt-1"
              rows={8}
              placeholder="Typ hier je notities over dit project..."
            />
            <p class="mt-1 text-xs text-muted-foreground">
              Deze notities worden opgeslagen in de database en zijn zichtbaar voor alle projectbeheerders.
            </p>
          </div>

          <div class="flex justify-end">
            <Button 
              type="submit" 
              disabled={savingNotes}
              class="flex items-center space-x-2"
            >
              {#if savingNotes}
                <div class="animate-spin rounded-full h-4 w-4 border-b-2 border-current"></div>
              {:else}
                <Icon name="backup" size={16} />
              {/if}
              <span>{savingNotes ? 'Opslaan...' : 'Opslaan'}</span>
            </Button>
          </div>
        </form>
      </Card>

      <!-- Danger Zone -->
      <Card class="border-destructive/20">
        <div class="px-6 py-4 border-b border-destructive/20 bg-destructive/10">
          <h3 class="text-lg font-medium text-destructive flex items-center space-x-2">
            <Icon name="backup" size={20} />
            <span>Gevaarzone</span>
          </h3>
        </div>
        <div class="p-6">
          <div class="space-y-4">
            <div>
              <h4 class="text-sm font-medium text-foreground">Project Verwijderen</h4>
              <p class="text-sm text-muted-foreground">
                Verwijder dit project permanent. Deze actie kan niet ongedaan gemaakt worden.
              </p>
              <Button variant="destructive" class="mt-2" on:click={showDeleteProjectDialog} disabled={loading}>
                {loading ? 'Verwijderen...' : 'Project Verwijderen'}
              </Button>
            </div>
          </div>
        </div>
      </Card>
    </div>
  {/if}
</div>

<!-- Create API Key Modal -->
{#if showCreateKey}
  <div class="fixed inset-0 bg-black/80 backdrop-blur-sm flex items-center justify-center p-4 z-50">
    <Card class="max-w-md w-full p-6 border-2 shadow-2xl">
      <div class="flex items-center space-x-3 mb-4">
        <div class="w-10 h-10 bg-primary/10 rounded-lg flex items-center justify-center">
          <Icon name="auth" size={20} className="text-primary" />
        </div>
        <h2 class="text-xl font-bold text-foreground">Nieuwe API Key</h2>
      </div>
      
      <form on:submit|preventDefault={createAPIKey} class="space-y-4">
        <div>
          <Label for="key-name">Naam</Label>
          <Input
            id="key-name"
            type="text"
            bind:value={newKeyName}
            required
            class="mt-1"
            placeholder="bijv. Frontend App, Mobile App"
          />
        </div>

        <div>
          <Label>Permissions</Label>
          <div class="mt-2 space-y-2">
            {#each ['read', 'write', 'delete'] as permission}
              <label class="flex items-center">
                <input
                  type="checkbox"
                  bind:group={newKeyPermissions}
                  value={permission}
                  class="rounded border-border text-primary focus:ring-primary"
                />
                <span class="ml-2 text-sm font-medium text-foreground capitalize">{permission}</span>
              </label>
            {/each}
          </div>
        </div>
        
        <div class="flex space-x-3 pt-4">
          <Button
            type="button"
            variant="outline"
            on:click={() => { showCreateKey = false; newKeyName = ''; newKeyPermissions = []; }}
            class="flex-1"
          >
            Annuleren
          </Button>
          <Button
            type="submit"
            disabled={loading || !newKeyName.trim()}
            class="flex-1"
          >
            {loading ? 'Aanmaken...' : 'Aanmaken'}
          </Button>
        </div>
      </form>
    </Card>
  </div>
{/if}

<!-- API Key Details Modal -->
{#if showKeyDetails}
  <div class="fixed inset-0 bg-black/80 backdrop-blur-sm flex items-center justify-center p-4 z-50">
    <Card class="max-w-lg w-full p-6 border-2 shadow-2xl">
      <div class="flex justify-between items-center mb-4">
        <div class="flex items-center space-x-3">
          <div class="w-10 h-10 bg-primary/10 rounded-lg flex items-center justify-center">
            <Icon name="auth" size={20} className="text-primary" />
          </div>
          <h2 class="text-xl font-bold text-foreground">{showKeyDetails.name}</h2>
        </div>
        <Button
          variant="ghost"
          size="sm"
          on:click={() => showKeyDetails = null}
        >
          ×
        </Button>
      </div>
      
      <div class="space-y-4">
        <div>
          <Label>API Key</Label>
          {#if showKeyDetails.is_new_key}
            <div class="mt-1 mb-3 bg-yellow-50 dark:bg-yellow-900/20 border border-yellow-200 dark:border-yellow-800 rounded-lg p-3">
              <div class="flex items-center space-x-2">
                <Icon name="backup" size={16} className="text-yellow-600 dark:text-yellow-400" />
                <span class="text-sm font-medium text-yellow-800 dark:text-yellow-200">Belangrijk:</span>
              </div>
              <p class="text-sm text-yellow-700 dark:text-yellow-300 mt-1">
                Dit is de enige keer dat je deze API key volledig kunt zien. Bewaar deze nu - je kunt deze later niet meer ophalen!
              </p>
            </div>
          {/if}
          <div class="mt-1 flex">
            <Input
              type="text"
              value={showKeyDetails.key}
              readonly
              class="flex-1 font-mono text-sm {showKeyDetails.is_new_key ? 'bg-green-50 dark:bg-green-900/20' : ''}"
            />
            <Button
              variant="outline"
              size="sm"
              class="ml-2"
              on:click={() => copyToClipboard(showKeyDetails.key, showKeyDetails.is_new_key)}
              disabled={!showKeyDetails.is_new_key && showKeyDetails.key.includes('...')}
              title={!showKeyDetails.is_new_key && showKeyDetails.key.includes('...') ? 'Deze key is gemaskeerd en kan niet gekopieerd worden' : 'Kopieer API key'}
            >
              <Icon name="backup" size={16} />
            </Button>
          </div>
          {#if !showKeyDetails.is_new_key && showKeyDetails.key.includes('...')}
            <p class="mt-2 text-xs text-muted-foreground">
              Deze API key wordt om veiligheidsredenen gemaskeerd weergegeven. Alleen nieuwe keys kunnen gekopieerd worden.
            </p>
          {/if}
        </div>

        <div>
          <Label>Voorbeeld gebruik</Label>
          <div class="mt-1 bg-muted p-3 rounded text-sm font-mono">
            <div class="text-muted-foreground">curl -H "X-API-Key: {showKeyDetails.key}" \</div>
            <div class="ml-4 text-muted-foreground">${API_BASE_URL}/p/project-slug/api/data/users</div>
          </div>
        </div>

        <div class="grid grid-cols-2 gap-4 text-sm">
          <div>
            <span class="text-muted-foreground">Aangemaakt:</span>
            <div class="font-medium text-foreground">{formatDate(showKeyDetails.created_at)}</div>
          </div>
          {#if showKeyDetails.last_used_at}
            <div>
              <span class="text-muted-foreground">Laatst gebruikt:</span>
              <div class="font-medium text-foreground">{formatDate(showKeyDetails.last_used_at)}</div>
            </div>
          {/if}
        </div>
      </div>
    </Card>
  </div>
{/if}


<!-- Delete Project Confirmation Modal -->
{#if showDeleteConfirm}
  <div class="fixed inset-0 bg-black/80 backdrop-blur-sm flex items-center justify-center p-4 z-50">
    <Card class="max-w-md w-full p-6 border-2 border-destructive shadow-2xl">
      <div class="flex items-center space-x-3 mb-4">
        <div class="w-10 h-10 bg-destructive/10 rounded-lg flex items-center justify-center">
          <Icon name="backup" size={20} className="text-destructive" />
        </div>
        <h2 class="text-xl font-bold text-destructive">Project Verwijderen</h2>
      </div>
      
      <div class="space-y-4">
        <p class="text-sm text-foreground">
          Weet je <strong>ABSOLUUT ZEKER</strong> dat je dit project wilt verwijderen?
        </p>
        
        <div class="bg-destructive/10 border border-destructive/20 rounded-lg p-4">
          <p class="text-sm font-medium text-destructive mb-2">Dit zal ALLE data permanent verwijderen:</p>
          <ul class="text-xs text-destructive space-y-1">
            <li>• Alle collections en documenten</li>
            <li>• Alle bestanden</li>
            <li>• Alle gebruikers</li>
            <li>• Alle berichten</li>
            <li>• Alle API keys</li>
          </ul>
        </div>
        
        <p class="text-xs text-muted-foreground font-medium">
          Deze actie kan NIET ongedaan gemaakt worden!
        </p>
        
        <div class="flex space-x-3 pt-4">
          <Button
            type="button"
            variant="outline"
            on:click={() => showDeleteConfirm = false}
            class="flex-1"
            disabled={loading}
          >
            Annuleren
          </Button>
          <Button
            type="button"
            variant="destructive"
            on:click={confirmDeleteProject}
            class="flex-1"
            disabled={loading}
          >
            {loading ? 'Verwijderen...' : 'Ja, Verwijderen'}
          </Button>
        </div>
      </div>
    </Card>
  </div>
{/if}