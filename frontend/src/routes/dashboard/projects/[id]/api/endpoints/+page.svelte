<script lang="ts">
  import { page } from '$app/stores';
  import { onMount } from 'svelte';
  import { auth } from '$lib/stores/auth';
  import { API_ENDPOINTS, createApiRequest, API_BASE_URL } from '$lib/config';
  import Card from '$lib/components/ui/card.svelte';
  import Button from '$lib/components/ui/button.svelte';
  import Input from '$lib/components/ui/input.svelte';
  import Badge from '$lib/components/ui/badge.svelte';
  import Icon from '$lib/components/ui/icon.svelte';
  import { toast } from '$lib/stores/toast';

  interface Project {
    id: number;
    slug: string;
    name: string;
  }

  interface APIEndpoint {
    method: string;
    path: string;
    description: string;
    category: string;
    requiresAuth: boolean;
    parameters?: APIParameter[];
    example?: APIExample;
    source?: string; // "database", "template", "core"
  }

  interface APIParameter {
    name: string;
    type: string;
    required: boolean;
    description: string;
  }

  interface APIExample {
    curl: string;
    javascript: string;
    response?: string;
  }

  interface APIDiscoveryResponse {
    baseURL: string;
    routes: APIEndpoint[];
    categories: string[];
    schema?: any;
  }

  let project: Project | null = null;
  let apiDiscovery: APIDiscoveryResponse | null = null;
  let apiEndpoints: APIEndpoint[] = [];
  let loading = true;
  let refreshing = false;
  let discoveryLoading = false;
  let selectedCategory = 'all';
  let searchQuery = '';
  let error = '';

  $: projectId = $page.params.id;
  $: baseURL = apiDiscovery?.baseURL || (project ? `${API_BASE_URL}/p/${project.id}/api` : '');
  $: apiEndpoints = apiDiscovery?.routes || [];
  $: availableCategories = apiDiscovery?.categories ? ['all', ...apiDiscovery.categories] : ['all'];

  $: filteredEndpoints = apiEndpoints.filter(endpoint => {
    const matchesCategory = selectedCategory === 'all' || endpoint.category === selectedCategory;
    const matchesSearch = searchQuery === '' || 
      endpoint.path.toLowerCase().includes(searchQuery.toLowerCase()) ||
      endpoint.description.toLowerCase().includes(searchQuery.toLowerCase());
    return matchesCategory && matchesSearch;
  });

  onMount(() => {
    loadProjectAndAPI();
  });

  async function loadProjectAndAPI() {
    try {
      // Load project info
      const projectResponse = await createApiRequest(API_ENDPOINTS.projects.get(projectId), {
        headers: {
          'Authorization': `Bearer ${$auth.token}`,
          'Content-Type': 'application/json',
        },
      });

      if (projectResponse.ok) {
        project = await projectResponse.json();
        
        // Load API discovery
        await loadAPIDiscovery();
      } else {
        error = 'Project niet gevonden';
      }
    } catch (err) {
      error = 'Fout bij laden van project';
      console.error('Load project error:', err);
    } finally {
      loading = false;
    }
  }

  async function loadAPIDiscovery() {
    if (!project) return;

    discoveryLoading = true;
    try {
      // Use discovery endpoint (requires auth or will fallback to basic routes)
      const discoveryURL = `${API_BASE_URL}/p/${project.id}/api/discovery/routes`;
      const response = await fetch(discoveryURL, {
        headers: {
          'Content-Type': 'application/json',
          'Authorization': `Bearer ${$auth.token}`,
        },
      });

      if (response.ok) {
        apiDiscovery = await response.json();
      } else {
        console.warn('Could not load API discovery, falling back to basic routes');
        // Fallback to basic routes if discovery fails
        apiDiscovery = {
          baseURL: `${API_BASE_URL}/p/${project.id}/api`,
          routes: getBasicRoutes(),
          categories: ['System', 'Authentication', 'Storage'],
        };
      }
    } catch (err) {
      console.error('API Discovery error:', err);
      // Fallback to basic routes
      apiDiscovery = {
        baseURL: `${API_BASE_URL}/p/${project.id}/api`,
        routes: getBasicRoutes(),
        categories: ['System', 'Authentication', 'Storage'],
      };
    } finally {
      discoveryLoading = false;
    }
  }

  // Fallback basic routes if discovery fails
  function getBasicRoutes(): APIEndpoint[] {
    return [
      {
        method: 'GET',
        path: '/health',
        description: 'API status controleren',
        category: 'System',
        requiresAuth: false,
        source: 'core',
        example: {
          curl: 'curl "{{baseURL}}/health"',
          javascript: 'fetch("{{baseURL}}/health")',
        },
      },
      {
        method: 'POST',
        path: '/users/register',
        description: 'Nieuwe gebruiker registreren',
        category: 'Authentication',
        requiresAuth: false,
        source: 'core',
        example: {
          curl: 'curl -X POST -H "Content-Type: application/json" -d \'{"email":"user@example.com","password":"password"}\' "{{baseURL}}/users/register"',
          javascript: 'fetch("{{baseURL}}/users/register", { method: "POST", headers: { "Content-Type": "application/json" }, body: JSON.stringify({ email: "user@example.com", password: "password" }) })',
        },
      },
      {
        method: 'GET',
        path: '/storage/buckets',
        description: 'Alle buckets ophalen',
        category: 'Storage',
        requiresAuth: true,
        source: 'core',
        example: {
          curl: 'curl -H "X-API-Key: your-api-key" "{{baseURL}}/storage/buckets"',
          javascript: 'fetch("{{baseURL}}/storage/buckets", { headers: { "X-API-Key": "your-api-key" } })',
        },
      },
    ];
  }

  function getMethodColor(method: string): string {
    switch (method) {
      case 'GET': return 'bg-green-100 dark:bg-green-900 text-green-800 dark:text-green-200';
      case 'POST': return 'bg-blue-100 dark:bg-gray-800 text-blue-800 dark:text-blue-200';
      case 'PUT': return 'bg-yellow-100 dark:bg-yellow-900 text-yellow-800 dark:text-yellow-200';
      case 'DELETE': return 'bg-red-100 dark:bg-red-900 text-red-800 dark:text-red-200';
      case 'PATCH': return 'bg-purple-100 dark:bg-purple-900 text-purple-800 dark:text-purple-200';
      default: return 'bg-muted text-muted-foreground';
    }
  }

  function copyToClipboard(text: string) {
    navigator.clipboard.writeText(text).then(() => {
      // Could add toast notification here
    });
  }

  async function refreshAPIDiscovery() {
    if (!project) return;

    refreshing = true;
    error = ''; // Clear any previous errors
    
    try {
      // Call the refresh endpoint directly instead of just reloading
      const refreshURL = `${API_BASE_URL}/p/${project.id}/api/discovery/refresh`;
      const refreshResponse = await fetch(refreshURL, {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
          'Authorization': `Bearer ${$auth.token}`,
        },
        body: JSON.stringify({
          reason: 'Manual refresh from dashboard',
          source: 'CloudBox Dashboard UI',
          forceRescan: true
        })
      });

      if (refreshResponse.ok) {
        const refreshResult = await refreshResponse.json();
        
        // Reload the discovery data
        await loadAPIDiscovery();
        
        // Show success toast notification
        toast.success(
          `${refreshResult.routeCount} API routes gevonden in ${refreshResult.categories?.length || 0} categorieën`,
          'API Discovery Voltooid'
        );
        
        // Show inbox notification toast
        toast.info(
          'Een gedetailleerd scanrapport is verzonden naar je project inbox',
          'Scanrapport Beschikbaar'
        );
      } else {
        const errorData = await refreshResponse.json().catch(() => null);
        throw new Error(errorData?.error || 'Failed to refresh API discovery');
      }
    } catch (err) {
      console.error('Failed to refresh API discovery:', err);
      const errorMessage = err instanceof Error ? err.message : 'Onbekende fout opgetreden';
      error = 'Fout bij vernieuwen van API discovery';
      
      // Show error toast notification
      toast.error(
        errorMessage.includes('Failed to refresh') ? 'Kon API discovery niet vernieuwen. Probeer het later opnieuw.' : errorMessage,
        'API Discovery Fout'
      );
    } finally {
      refreshing = false;
    }
  }

  function generateExample(endpoint: APIEndpoint, type: 'curl' | 'javascript' = 'curl'): string {
    // Use pre-generated examples if available
    if (endpoint.example) {
      const example = type === 'curl' ? endpoint.example.curl : endpoint.example.javascript;
      return example.replace(/{{baseURL}}/g, baseURL);
    }

    // Fallback to basic generation
    const url = `${baseURL}${endpoint.path}`;
    const authHeader = endpoint.requiresAuth ? (type === 'curl' ? `-H "X-API-Key: your-api-key" ` : `{ 'X-API-Key': 'your-api-key' }`) : '';
    
    if (type === 'curl') {
      switch (endpoint.method) {
        case 'GET':
          return `curl ${authHeader}"${url}"`;
        case 'POST':
          return `curl -X POST ${authHeader}-H "Content-Type: application/json" -d '{}' "${url}"`;
        case 'PUT':
          return `curl -X PUT ${authHeader}-H "Content-Type: application/json" -d '{}' "${url}"`;
        case 'DELETE':
          return `curl -X DELETE ${authHeader}"${url}"`;
        default:
          return `curl ${authHeader}"${url}"`;
      }
    } else {
      const headers = endpoint.requiresAuth ? `{ 'Content-Type': 'application/json', 'X-API-Key': 'your-api-key' }` : `{ 'Content-Type': 'application/json' }`;
      switch (endpoint.method) {
        case 'GET':
          return `fetch('${url}', { headers: ${authHeader || '{}' } })`;
        case 'POST':
          return `fetch('${url}', { method: 'POST', headers: ${headers}, body: JSON.stringify({}) })`;
        case 'PUT':
          return `fetch('${url}', { method: 'PUT', headers: ${headers}, body: JSON.stringify({}) })`;
        case 'DELETE':
          return `fetch('${url}', { method: 'DELETE', headers: ${authHeader || '{}'} })`;
        default:
          return `fetch('${url}', { headers: ${authHeader || '{}'} })`;
      }
    }
  }
</script>

<svelte:head>
  <title>API Endpoints - CloudBox</title>
</svelte:head>

<div class="space-y-6">

  {#if loading}
    <div class="flex items-center justify-center min-h-64">
      <div class="animate-spin rounded-full h-8 w-8 border-b-2 border-primary"></div>
    </div>
  {:else if error}
    <Card class="glassmorphism-content p-12 text-center">
      <div class="w-16 h-16 bg-destructive/10 rounded-lg flex items-center justify-center mx-auto mb-4">
        <Icon name="alert-triangle" size={32} className="text-destructive" />
      </div>
      <h3 class="text-lg font-medium text-foreground mb-2">Fout bij laden</h3>
      <p class="text-destructive mb-6">{error}</p>
      <Button on:click={loadProjectAndAPI} variant="outline">
        <Icon name="refresh-cw" size={16} className="mr-2" />
        Opnieuw proberen
      </Button>
    </Card>
  {:else if project}
    <!-- Dynamic API Discovery Info -->
    {#if apiDiscovery && (apiDiscovery.routes.some(r => r.source === 'database') || apiDiscovery.routes.some(r => r.source === 'template'))}
      <Card class="glassmorphism-card p-4 border-blue-200 dark:border-blue-800 bg-blue-50/50 dark:bg-blue-950/30">
        <div class="flex items-start space-x-3">
          <div class="w-8 h-8 bg-blue-100 dark:bg-blue-900 rounded-lg flex items-center justify-center flex-shrink-0">
            <Icon name="zap" size={16} className="text-blue-600 dark:text-blue-400" />
          </div>
          <div class="flex-1">
            <h3 class="text-sm font-medium text-blue-900 dark:text-blue-100 mb-1">🎉 Dynamische API Discovery</h3>
            <p class="text-sm text-blue-700 dark:text-blue-300">
              Deze API routes zijn automatisch gegenereerd op basis van jouw database schema en geïnstalleerde templates. 
              {#if apiDiscovery.routes.some(r => r.source === 'database')}
                <span class="font-medium">Database routes</span> worden real-time bijgewerkt als je tabellen wijzigt.
              {/if}
            </p>
          </div>
        </div>
      </Card>
    {/if}

    <!-- Base URL Info -->
    <Card class="glassmorphism-card p-6">
      <div class="flex items-start justify-between">
        <div class="flex-1">
          <div class="flex items-center space-x-2 mb-2">
            <h3 class="text-lg font-medium text-foreground">API Base URL</h3>
            {#if discoveryLoading}
              <div class="animate-spin rounded-full h-4 w-4 border-b-2 border-primary"></div>
              <span class="text-xs text-muted-foreground">Routes laden...</span>
            {/if}
          </div>
          <div class="flex items-center space-x-2">
            <code class="bg-muted px-3 py-2 rounded-md text-foreground font-mono">
              {baseURL}
            </code>
            <div class="flex items-center space-x-2">
              <Button 
                variant="outline" 
                size="sm" 
                on:click={() => copyToClipboard(baseURL)}
                class="flex items-center space-x-1"
              >
                <Icon name="backup" size={14} />
                <span>Kopiëren</span>
              </Button>
              <Button 
                variant="outline" 
                size="sm" 
                on:click={refreshAPIDiscovery}
                class="flex items-center space-x-1"
                disabled={refreshing || discoveryLoading}
              >
                <Icon name="refresh-cw" size={14} className={refreshing ? 'animate-spin' : ''} />
                <span>{refreshing ? 'Vernieuwen...' : 'API Routes Vernieuwen'}</span>
              </Button>
            </div>
          </div>
          <p class="text-sm text-muted-foreground mt-2">
            Gebruik deze base URL voor alle API requests naar jouw project.
            <br />
            <span class="text-xs">💡 Klik op "API Routes Vernieuwen" om de nieuwste routes te laden na database- of template wijzigingen.</span>
            <br />
            <span class="text-xs">📬 Na het vernieuwen ontvang je automatisch een gedetailleerd scanrapport in je project inbox.</span>
          </p>
        </div>
      </div>
    </Card>

    <!-- Filters -->
    <Card class="glassmorphism-nav p-4">
      <div class="flex flex-col sm:flex-row gap-4">
        <div class="flex-1">
          <Input
            type="text"
            placeholder="Zoek endpoints..."
            bind:value={searchQuery}
            class="w-full"
          />
        </div>
        <div class="flex gap-2 flex-wrap">
          {#each availableCategories as category}
            <Button
              variant={selectedCategory === category ? "default" : "outline"}
              size="sm"
              on:click={() => selectedCategory = category}
              class="capitalize"
            >
              {category === 'all' ? 'Alle' : category}
            </Button>
          {/each}
        </div>
      </div>
    </Card>

    <!-- API Endpoints -->
    <div class="space-y-4">
      {#each filteredEndpoints as endpoint}
        <Card class="glassmorphism-card p-6">
          <div class="flex items-start justify-between mb-4">
            <div class="flex items-center space-x-3">
              <Badge variant="secondary" class={getMethodColor(endpoint.method)}>
                {endpoint.method}
              </Badge>
              <code class="text-sm font-mono text-foreground">{endpoint.path}</code>
            </div>
            <div class="flex items-center space-x-2">
              {#if endpoint.source}
                <Badge variant="outline" class="text-xs {endpoint.source === 'database' ? 'border-blue-300 text-blue-700 dark:border-blue-700 dark:text-blue-300' : endpoint.source === 'template' ? 'border-purple-300 text-purple-700 dark:border-purple-700 dark:text-purple-300' : 'border-gray-300 text-gray-700 dark:border-gray-700 dark:text-gray-300'}">
                  <Icon name={endpoint.source === 'database' ? 'database' : endpoint.source === 'template' ? 'package' : 'settings'} size={12} className="mr-1" />
                  {endpoint.source === 'database' ? 'Database' : endpoint.source === 'template' ? 'Template' : 'Core'}
                </Badge>
              {/if}
              {#if endpoint.requiresAuth}
                <Badge variant="outline" class="text-xs">
                  <Icon name="auth" size={12} className="mr-1" />
                  Auth Required
                </Badge>
              {/if}
              <Badge variant="outline" class="text-xs">{endpoint.category}</Badge>
            </div>
          </div>
          
          <p class="text-muted-foreground mb-4">{endpoint.description}</p>
          
          <!-- Parameters -->
          {#if endpoint.parameters && endpoint.parameters.length > 0}
            <div class="mb-4">
              <h4 class="text-sm font-medium text-foreground mb-2">Parameters:</h4>
              <div class="grid grid-cols-1 md:grid-cols-2 gap-2">
                {#each endpoint.parameters as param}
                  <div class="flex items-center space-x-2 text-xs">
                    <Badge variant={param.required ? "destructive" : "secondary"} class="text-xs">
                      {param.required ? 'Required' : 'Optional'}
                    </Badge>
                    <code class="bg-muted px-2 py-1 rounded">{param.name}</code>
                    <span class="text-muted-foreground">({param.type})</span>
                  </div>
                  <div class="text-xs text-muted-foreground col-span-full">
                    {param.description}
                  </div>
                {/each}
              </div>
            </div>
          {/if}

          <!-- Code Examples -->
          <div class="space-y-4">
            <!-- cURL Example -->
            <div class="bg-muted rounded-lg p-4">
              <div class="flex items-center justify-between mb-2">
                <span class="text-sm font-medium text-foreground">cURL Voorbeeld:</span>
                <Button 
                  variant="ghost" 
                  size="sm" 
                  on:click={() => copyToClipboard(generateExample(endpoint, 'curl'))}
                  class="h-6 px-2 text-xs"
                >
                  <Icon name="backup" size={12} className="mr-1" />
                  Kopiëren
                </Button>
              </div>
              <code class="text-xs font-mono text-foreground break-all">
                {generateExample(endpoint, 'curl')}
              </code>
            </div>

            <!-- JavaScript Example -->
            <div class="bg-muted rounded-lg p-4">
              <div class="flex items-center justify-between mb-2">
                <span class="text-sm font-medium text-foreground">JavaScript Voorbeeld:</span>
                <Button 
                  variant="ghost" 
                  size="sm" 
                  on:click={() => copyToClipboard(generateExample(endpoint, 'javascript'))}
                  class="h-6 px-2 text-xs"
                >
                  <Icon name="backup" size={12} className="mr-1" />
                  Kopiëren
                </Button>
              </div>
              <code class="text-xs font-mono text-foreground break-all">
                {generateExample(endpoint, 'javascript')}
              </code>
            </div>
          </div>
        </Card>
      {/each}
    </div>

    {#if filteredEndpoints.length === 0}
      <Card class="glassmorphism-content p-12 text-center">
        <div class="w-16 h-16 bg-muted rounded-lg flex items-center justify-center mx-auto mb-4">
          <Icon name="settings" size={32} className="text-muted-foreground" />
        </div>
        <h3 class="text-lg font-medium text-foreground mb-2">Geen endpoints gevonden</h3>
        <p class="text-muted-foreground">
          Probeer een andere zoekopdracht of categorie.
        </p>
      </Card>
    {/if}

    <!-- API Documentation -->
    <div class="grid grid-cols-1 lg:grid-cols-2 gap-6">
      <!-- Authentication -->
      <Card class="glassmorphism-form p-6">
        <div class="flex items-start space-x-4">
          <div class="w-10 h-10 bg-blue-100 dark:bg-gray-800 rounded-lg flex items-center justify-center flex-shrink-0">
            <Icon name="auth" size={20} className="text-blue-600 dark:text-blue-400" />
          </div>
          <div>
            <h3 class="text-lg font-medium text-foreground mb-2">Authenticatie</h3>
            <p class="text-muted-foreground mb-4">
              De meeste endpoints vereisen authenticatie via een API key. Voeg de volgende header toe aan je requests:
            </p>
            <code class="bg-muted px-3 py-2 rounded-md text-foreground font-mono text-sm">
              X-API-Key: your-api-key-here
            </code>
            <p class="text-muted-foreground mt-4 text-sm">
              Je kunt API keys beheren in de <a href="/dashboard/projects/{projectId}/settings" class="text-primary hover:underline">project instellingen</a>.
            </p>
          </div>
        </div>
      </Card>

      <!-- External Triggers -->
      <Card class="glassmorphism-form p-6">
        <div class="flex items-start space-x-4">
          <div class="w-10 h-10 bg-purple-100 dark:bg-purple-900 rounded-lg flex items-center justify-center flex-shrink-0">
            <Icon name="webhook" size={20} className="text-purple-600 dark:text-purple-400" />
          </div>
          <div>
            <h3 class="text-lg font-medium text-foreground mb-2">Externe Refresh Trigger</h3>
            <p class="text-muted-foreground mb-4">
              Apps kunnen API discovery programmatisch refreshen via een POST request:
            </p>
            <code class="bg-muted px-3 py-2 rounded-md text-foreground font-mono text-sm break-all">
              POST {baseURL}/discovery/refresh
            </code>
            <p class="text-muted-foreground mt-4 text-sm">
              Handig voor <span class="font-medium">app updates</span>, <span class="font-medium">database migraties</span>, of <span class="font-medium">template wijzigingen</span>.
            </p>
          </div>
        </div>
      </Card>
    </div>
  {/if}
</div>