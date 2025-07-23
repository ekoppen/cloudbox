<script lang="ts">
  import { page } from '$app/stores';
  import { onMount } from 'svelte';
  import { API_BASE_URL, createApiRequest } from '$lib/config';
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
  }

  interface CORSConfig {
    allowed_origins: string[];
    allowed_methods: string[];
    allowed_headers: string[];
    allow_credentials: boolean;
    max_age: number;
  }

  let apiKeys: APIKey[] = [];

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

  $: projectId = $page.params.id;

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
      const newKey: APIKey = {
        id: Date.now(),
        name: newKeyName,
        key: generateAPIKey(),
        created_at: new Date().toISOString(),
        permissions: [...newKeyPermissions],
        is_active: true
      };

      apiKeys = [...apiKeys, newKey];
      showCreateKey = false;
      newKeyName = '';
      newKeyPermissions = [];
      showKeyDetails = newKey; // Show the new key
    } catch (err) {
      error = 'Fout bij aanmaken van API key';
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
        const response = await createApiRequest(API_ENDPOINTS.projects.delete(projectId), {
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
        toastStore.error('Netwerkfout bij verwijderen van API key');
      }
    }
  }

  function showDeleteProjectDialog() {
    showDeleteConfirm = true;
  }

  async function confirmDeleteProject() {
    console.log('confirmDeleteProject called, projectId:', projectId);
    showDeleteConfirm = false;
    loading = true;
    
    try {
      const deleteUrl = API_ENDPOINTS.projects.delete(projectId);
      console.log('Delete URL:', deleteUrl);
      console.log('Auth token exists:', !!$auth.token);
      console.log('Auth token length:', $auth.token?.length || 0);
      
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

  function copyToClipboard(text: string) {
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

  function saveCORSConfig() {
    corsConfig = { ...corsFormData };
    alert('CORS configuratie opgeslagen!');
  }

  function resetCORSConfig() {
    corsFormData = { ...corsConfig };
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

<div class="space-y-6">
  <!-- Header -->
  <div class="flex items-center space-x-4">
    <div class="w-10 h-10 bg-primary/10 rounded-lg flex items-center justify-center">
      <Icon name="settings" size={20} className="text-primary" />
    </div>
    <div>
      <h1 class="text-2xl font-bold text-foreground">Project Instellingen</h1>
      <p class="text-sm text-muted-foreground">
        Beheer API keys, CORS configuratie en project instellingen
      </p>
    </div>
  </div>

  <!-- Tabs -->
  <div class="border-b border-border">
    <nav class="flex space-x-8">
      <button
        on:click={() => activeTab = 'api-keys'}
        class="flex items-center space-x-2 py-2 px-1 border-b-2 text-sm font-medium transition-colors {
          activeTab === 'api-keys' 
            ? 'border-primary text-primary' 
            : 'border-transparent text-muted-foreground hover:text-foreground hover:border-border'
        }"
      >
        <Icon name="auth" size={16} />
        <span>API Keys</span>
      </button>
      <button
        on:click={() => activeTab = 'cors'}
        class="flex items-center space-x-2 py-2 px-1 border-b-2 text-sm font-medium transition-colors {
          activeTab === 'cors' 
            ? 'border-primary text-primary' 
            : 'border-transparent text-muted-foreground hover:text-foreground hover:border-border'
        }"
      >
        <Icon name="cloud" size={16} />
        <span>CORS</span>
      </button>
      <button
        on:click={() => activeTab = 'general'}
        class="flex items-center space-x-2 py-2 px-1 border-b-2 text-sm font-medium transition-colors {
          activeTab === 'general' 
            ? 'border-primary text-primary' 
            : 'border-transparent text-muted-foreground hover:text-foreground hover:border-border'
        }"
      >
        <Icon name="settings" size={16} />
        <span>Algemeen</span>
      </button>
    </nav>
  </div>

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
    <div class="space-y-6">
      <div class="flex justify-between items-center">
        <div>
          <h2 class="text-lg font-medium text-foreground">API Keys</h2>
          <p class="text-sm text-muted-foreground">Beheer toegang tot je project API</p>
        </div>
        <Button on:click={() => showCreateKey = true} class="flex items-center space-x-2">
          <Icon name="package" size={16} />
          <span>Nieuwe API Key</span>
        </Button>
      </div>

      <Card>
        <div class="divide-y divide-border">
          {#each apiKeys as apiKey}
            <div class="p-6">
              <div class="flex items-center justify-between">
                <div class="flex-1">
                  <div class="flex items-center space-x-3">
                    <div class="w-8 h-8 bg-primary/10 rounded-lg flex items-center justify-center">
                      <Icon name="auth" size={16} className="text-primary" />
                    </div>
                    <div>
                      <h3 class="text-lg font-medium text-foreground">{apiKey.name}</h3>
                      <div class="flex items-center space-x-2 mt-1">
                        <Badge variant={apiKey.is_active ? "default" : "secondary"}>
                          {apiKey.is_active ? 'Actief' : 'Inactief'}
                        </Badge>
                        {#each apiKey.permissions as permission}
                          <span class="inline-flex px-2 py-1 text-xs font-medium rounded-full {getPermissionColor(permission)}">
                            {permission}
                          </span>
                        {/each}
                      </div>
                    </div>
                  </div>
                  
                  <div class="mt-2 flex items-center space-x-4 text-sm text-muted-foreground">
                    <span>Aangemaakt: {formatDate(apiKey.created_at)}</span>
                    {#if apiKey.last_used_at}
                      <span>Laatst gebruikt: {formatDate(apiKey.last_used_at)}</span>
                    {:else}
                      <span>Nog niet gebruikt</span>
                    {/if}
                  </div>
                </div>

                <div class="flex items-center space-x-2">
                  <Button
                    variant="outline"
                    size="sm"
                    on:click={() => showKeyDetails = apiKey}
                  >
                    Details
                  </Button>
                  <Button
                    variant="ghost"
                    size="sm"
                    on:click={() => toggleKeyStatus(apiKey.id)}
                  >
                    {apiKey.is_active ? 'Deactiveren' : 'Activeren'}
                  </Button>
                  <Button
                    variant="ghost"
                    size="sm"
                    on:click={() => deleteKey(apiKey.id)}
                    class="text-destructive hover:text-destructive"
                  >
                    Verwijderen
                  </Button>
                </div>
              </div>
            </div>
          {/each}
        </div>
      </Card>
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
                bind:value={corsFormData.allowed_headers}
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
        <p class="text-sm text-muted-foreground">Project naam, beschrijving en gevaarlijke acties</p>
      </div>

      <Card class="p-6">
        <div class="space-y-6">
          <div>
            <Label>Project Naam</Label>
            <Input
              type="text"
              value="Mijn Eerste Project"
              class="mt-1"
            />
          </div>

          <div>
            <Label>Beschrijving</Label>
            <Textarea
              class="mt-1"
              rows={3}
              placeholder="Project beschrijving..."
              value="Dit is een test project voor CloudBox"
            />
          </div>

          <div class="flex justify-end">
            <Button>
              Opslaan
            </Button>
          </div>
        </div>
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
          <div class="mt-1 flex">
            <Input
              type="text"
              value={showKeyDetails.key}
              readonly
              class="flex-1 font-mono text-sm"
            />
            <Button
              variant="outline"
              size="sm"
              class="ml-2"
              on:click={() => copyToClipboard(showKeyDetails.key)}
            >
              <Icon name="backup" size={16} />
            </Button>
          </div>
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