<script lang="ts">
  import { page } from '$app/stores';
  import { onMount } from 'svelte';
  import { auth } from '$lib/stores/auth';
  import { API_ENDPOINTS, createApiRequest } from '$lib/config';
  import Card from '$lib/components/ui/card.svelte';
  import Button from '$lib/components/ui/button.svelte';
  import Input from '$lib/components/ui/input.svelte';
  import Badge from '$lib/components/ui/badge.svelte';
  import Icon from '$lib/components/ui/icon.svelte';

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
    example?: string;
  }

  let project: Project | null = null;
  let apiEndpoints: APIEndpoint[] = [];
  let loading = true;
  let selectedCategory = 'all';
  let searchQuery = '';

  $: projectId = $page.params.id;
  $: baseURL = project ? `http://localhost:8080/p/${project.id}/api` : '';

  // Define available API endpoints
  $: if (project) {
    apiEndpoints = [
      // Collections/Database endpoints (dynamic based on project)
      { method: 'GET', path: '/pages', description: 'Alle pagina\'s ophalen', category: 'Collections', requiresAuth: true },
      { method: 'POST', path: '/pages', description: 'Nieuwe pagina aanmaken', category: 'Collections', requiresAuth: true },
      { method: 'GET', path: '/pages/{id}', description: 'Specifieke pagina ophalen', category: 'Collections', requiresAuth: true },
      { method: 'PUT', path: '/pages/{id}', description: 'Pagina bijwerken', category: 'Collections', requiresAuth: true },
      { method: 'DELETE', path: '/pages/{id}', description: 'Pagina verwijderen', category: 'Collections', requiresAuth: true },
      
      { method: 'GET', path: '/albums', description: 'Alle albums ophalen', category: 'Collections', requiresAuth: true },
      { method: 'POST', path: '/albums', description: 'Nieuw album aanmaken', category: 'Collections', requiresAuth: true },
      { method: 'GET', path: '/albums/{id}', description: 'Specifiek album ophalen', category: 'Collections', requiresAuth: true },
      { method: 'PUT', path: '/albums/{id}', description: 'Album bijwerken', category: 'Collections', requiresAuth: true },
      { method: 'DELETE', path: '/albums/{id}', description: 'Album verwijderen', category: 'Collections', requiresAuth: true },
      
      { method: 'GET', path: '/images', description: 'Alle afbeeldingen ophalen', category: 'Collections', requiresAuth: true },
      { method: 'POST', path: '/images', description: 'Nieuwe afbeelding metadata aanmaken', category: 'Collections', requiresAuth: true },
      { method: 'GET', path: '/images/{id}', description: 'Specifieke afbeelding ophalen', category: 'Collections', requiresAuth: true },
      { method: 'PUT', path: '/images/{id}', description: 'Afbeelding metadata bijwerken', category: 'Collections', requiresAuth: true },
      { method: 'DELETE', path: '/images/{id}', description: 'Afbeelding verwijderen', category: 'Collections', requiresAuth: true },
      
      { method: 'GET', path: '/settings', description: 'Project instellingen ophalen', category: 'Collections', requiresAuth: true },
      { method: 'PUT', path: '/settings', description: 'Project instellingen bijwerken', category: 'Collections', requiresAuth: true },
      
      { method: 'GET', path: '/branding', description: 'Brand instellingen ophalen', category: 'Collections', requiresAuth: true },
      { method: 'PUT', path: '/branding', description: 'Brand instellingen bijwerken', category: 'Collections', requiresAuth: true },

      // Storage endpoints
      { method: 'GET', path: '/storage/buckets', description: 'Alle buckets ophalen', category: 'Storage', requiresAuth: true },
      { method: 'POST', path: '/storage/buckets', description: 'Nieuwe bucket aanmaken', category: 'Storage', requiresAuth: true },
      { method: 'GET', path: '/storage/buckets/{bucket}/files', description: 'Bestanden in bucket ophalen', category: 'Storage', requiresAuth: true },
      { method: 'POST', path: '/storage/buckets/{bucket}/files', description: 'Bestand uploaden naar bucket', category: 'Storage', requiresAuth: true },
      { method: 'GET', path: '/storage/buckets/{bucket}/files/{file}', description: 'Specifiek bestand ophalen', category: 'Storage', requiresAuth: true },
      { method: 'DELETE', path: '/storage/buckets/{bucket}/files/{file}', description: 'Bestand verwijderen', category: 'Storage', requiresAuth: true },

      // Authentication endpoints
      { method: 'POST', path: '/auth/register', description: 'Nieuwe gebruiker registreren', category: 'Authentication', requiresAuth: false },
      { method: 'POST', path: '/auth/login', description: 'Gebruiker inloggen', category: 'Authentication', requiresAuth: false },
      { method: 'POST', path: '/auth/logout', description: 'Gebruiker uitloggen', category: 'Authentication', requiresAuth: true },
      { method: 'GET', path: '/auth/me', description: 'Huidige gebruiker ophalen', category: 'Authentication', requiresAuth: true },
      { method: 'PUT', path: '/auth/profile', description: 'Gebruikersprofiel bijwerken', category: 'Authentication', requiresAuth: true },

      // Template endpoints
      { method: 'GET', path: '/templates', description: 'Beschikbare templates ophalen', category: 'Templates', requiresAuth: true },
      { method: 'GET', path: '/templates/{template}', description: 'Specifieke template ophalen', category: 'Templates', requiresAuth: true },
      { method: 'POST', path: '/templates/{template}/setup', description: 'Template installeren/configureren', category: 'Templates', requiresAuth: true },

      // System endpoints
      { method: 'GET', path: '/health', description: 'API status controleren', category: 'System', requiresAuth: false },
      { method: 'GET', path: '/info', description: 'Project informatie ophalen', category: 'System', requiresAuth: true },
      { method: 'GET', path: '/cors', description: 'CORS configuratie ophalen', category: 'System', requiresAuth: true },
      { method: 'PUT', path: '/cors', description: 'CORS configuratie bijwerken', category: 'System', requiresAuth: true },
    ];
  }

  $: categories = ['all', ...new Set(apiEndpoints.map(endpoint => endpoint.category))];
  $: filteredEndpoints = apiEndpoints.filter(endpoint => {
    const matchesCategory = selectedCategory === 'all' || endpoint.category === selectedCategory;
    const matchesSearch = searchQuery === '' || 
      endpoint.path.toLowerCase().includes(searchQuery.toLowerCase()) ||
      endpoint.description.toLowerCase().includes(searchQuery.toLowerCase());
    return matchesCategory && matchesSearch;
  });

  onMount(() => {
    loadProject();
  });

  async function loadProject() {
    try {
      const response = await createApiRequest(API_ENDPOINTS.projects.get(projectId), {
        headers: {
          'Authorization': `Bearer ${$auth.token}`,
          'Content-Type': 'application/json',
        },
      });

      if (response.ok) {
        project = await response.json();
      }
    } catch (err) {
      console.error('Load project error:', err);
    } finally {
      loading = false;
    }
  }

  function getMethodColor(method: string): string {
    switch (method) {
      case 'GET': return 'bg-green-100 dark:bg-green-900 text-green-800 dark:text-green-200';
      case 'POST': return 'bg-blue-100 dark:bg-blue-900 text-blue-800 dark:text-blue-200';
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

  function generateExample(endpoint: APIEndpoint): string {
    const url = `${baseURL}${endpoint.path}`;
    const authHeader = endpoint.requiresAuth ? `-H "X-API-Key: your-api-key" ` : '';
    
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
  }
</script>

<svelte:head>
  <title>API Overzicht - CloudBox</title>
</svelte:head>

<div class="space-y-6">
  <!-- Header -->
  <div class="flex items-center justify-between">
    <div class="flex items-center space-x-4">
      <div class="w-10 h-10 bg-primary/10 rounded-lg flex items-center justify-center">
        <Icon name="settings" size={20} className="text-primary" />
      </div>
      <div>
        <h1 class="text-2xl font-bold text-foreground">API Overzicht</h1>
        <p class="text-sm text-muted-foreground">
          Alle beschikbare API endpoints voor jouw project
        </p>
      </div>
    </div>
  </div>

  {#if loading}
    <div class="flex items-center justify-center min-h-64">
      <div class="animate-spin rounded-full h-8 w-8 border-b-2 border-primary"></div>
    </div>
  {:else if project}
    <!-- Base URL Info -->
    <Card class="p-6">
      <div class="flex items-start justify-between">
        <div>
          <h3 class="text-lg font-medium text-foreground mb-2">API Base URL</h3>
          <div class="flex items-center space-x-2">
            <code class="bg-muted px-3 py-2 rounded-md text-foreground font-mono">
              {baseURL}
            </code>
            <Button 
              variant="outline" 
              size="sm" 
              on:click={() => copyToClipboard(baseURL)}
              class="flex items-center space-x-1"
            >
              <Icon name="backup" size={14} />
              <span>Kopiëren</span>
            </Button>
          </div>
          <p class="text-sm text-muted-foreground mt-2">
            Gebruik deze base URL voor alle API requests naar jouw project.
          </p>
        </div>
      </div>
    </Card>

    <!-- Filters -->
    <Card class="p-4">
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
          {#each categories as category}
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
        <Card class="p-6">
          <div class="flex items-start justify-between mb-4">
            <div class="flex items-center space-x-3">
              <Badge variant="secondary" class={getMethodColor(endpoint.method)}>
                {endpoint.method}
              </Badge>
              <code class="text-sm font-mono text-foreground">{endpoint.path}</code>
            </div>
            <div class="flex items-center space-x-2">
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
          
          <!-- cURL Example -->
          <div class="bg-muted rounded-lg p-4">
            <div class="flex items-center justify-between mb-2">
              <span class="text-sm font-medium text-foreground">cURL Voorbeeld:</span>
              <Button 
                variant="ghost" 
                size="sm" 
                on:click={() => copyToClipboard(generateExample(endpoint))}
                class="h-6 px-2 text-xs"
              >
                <Icon name="backup" size={12} className="mr-1" />
                Kopiëren
              </Button>
            </div>
            <code class="text-xs font-mono text-foreground break-all">
              {generateExample(endpoint)}
            </code>
          </div>
        </Card>
      {/each}
    </div>

    {#if filteredEndpoints.length === 0}
      <Card class="p-12 text-center">
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
    <Card class="p-6">
      <div class="flex items-start space-x-4">
        <div class="w-10 h-10 bg-blue-100 dark:bg-blue-900 rounded-lg flex items-center justify-center flex-shrink-0">
          <Icon name="settings" size={20} className="text-blue-600 dark:text-blue-400" />
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
  {/if}
</div>