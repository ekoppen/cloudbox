<script lang="ts">
  import { onMount } from 'svelte';
  import { API_ENDPOINTS, createApiRequest } from '$lib/config';
  import { auth } from '$lib/stores/auth';
  import Card from '$lib/components/ui/card.svelte';
  import Button from '$lib/components/ui/button.svelte';
  import Input from '$lib/components/ui/input.svelte';
  import Label from '$lib/components/ui/label.svelte';
  import Badge from '$lib/components/ui/badge.svelte';
  import Textarea from '$lib/components/ui/textarea.svelte';
  import Icon from '$lib/components/ui/icon.svelte';
  
  interface Organization {
    id: number;
    name: string;
    color: string;
  }

  interface Project {
    id: number;
    name: string;
    description: string;
    slug: string;
    created_at: string;
    is_active: boolean;
    organization?: Organization;
  }

  let projects: Project[] = [];
  let loading = true;
  let error = '';
  
  // Admin stats
  let adminStats = null;
  let loadingAdminStats = false;
  
  $: isSuperAdmin = $auth.user?.role === 'superadmin';

  onMount(() => {
    loadProjects();
    if (isSuperAdmin) {
      loadAdminStats();
    }
  });

  async function loadProjects() {
    loading = true;
    error = '';

    try {
      const response = await createApiRequest(API_ENDPOINTS.projects.list, {
        headers: {
          'Authorization': `Bearer ${$auth.token}`,
        },
      });

      if (response.ok) {
        projects = await response.json();
      } else {
        const data = await response.json();
        error = data.error || 'Fout bij laden van projecten';
      }
    } catch (err) {
      error = 'Netwerkfout bij laden van projecten';
      console.error('Load projects error:', err);
    } finally {
      loading = false;
    }
  }


  async function loadAdminStats() {
    if (!isSuperAdmin) return;
    
    loadingAdminStats = true;
    try {
      const response = await createApiRequest(API_ENDPOINTS.admin.stats.system, {
        headers: {
          'Authorization': `Bearer ${$auth.token}`,
        },
      });
      
      if (response.ok) {
        adminStats = await response.json();
      } else {
        console.error('Failed to load admin stats');
      }
    } catch (err) {
      console.error('Admin stats error:', err);
    } finally {
      loadingAdminStats = false;
    }
  }
  
  function formatDate(dateStr: string) {
    return new Date(dateStr).toLocaleDateString('nl-NL', {
      year: 'numeric',
      month: 'long',
      day: 'numeric'
    });
  }
  
  function formatBytes(bytes: number) {
    if (bytes === 0) return '0 B';
    const k = 1024;
    const sizes = ['B', 'KB', 'MB', 'GB', 'TB'];
    const i = Math.floor(Math.log(bytes) / Math.log(k));
    return parseFloat((bytes / Math.pow(k, i)).toFixed(2)) + ' ' + sizes[i];
  }
</script>

<svelte:head>
  <title>Dashboard - CloudBox</title>
</svelte:head>

<div class="space-y-8">
  <!-- Header -->
  <div class="flex items-center justify-between">
    <div class="flex items-center space-x-4">
      <div class="w-12 h-12 bg-primary rounded-xl flex items-center justify-center">
        <Icon name="dashboard" size={24} color="white" />
      </div>
      <div>
        <h1 class="text-3xl font-bold text-foreground">Dashboard</h1>
        <p class="text-muted-foreground">
          Beheer je CloudBox projecten en API's
        </p>
      </div>
    </div>
    <Button
      href="/dashboard/projects"
      size="lg"
      class="flex items-center space-x-2"
    >
      <Icon name="package" size={16} />
      <span>Nieuw Project</span>
    </Button>
  </div>

  <!-- Quick Stats -->
  <div class="grid grid-cols-1 md:grid-cols-4 gap-6">
    <Card class="p-6">
      <div class="flex items-center justify-between">
        <div>
          <p class="text-sm font-medium text-muted-foreground">Totaal Projecten</p>
          <p class="text-2xl font-bold text-foreground">{projects.length}</p>
        </div>
        <div class="w-10 h-10 bg-blue-100 dark:bg-blue-900 rounded-lg flex items-center justify-center">
          <Icon name="package" size={20} className="text-blue-600 dark:text-blue-400" />
        </div>
      </div>
    </Card>
    
    <Card class="p-6">
      <div class="flex items-center justify-between">
        <div>
          <p class="text-sm font-medium text-muted-foreground">Actieve API's</p>
          <p class="text-2xl font-bold text-foreground">{projects.filter(p => p.is_active).length}</p>
        </div>
        <div class="w-10 h-10 bg-green-100 dark:bg-green-900 rounded-lg flex items-center justify-center">
          <Icon name="database" size={20} className="text-green-600 dark:text-green-400" />
        </div>
      </div>
    </Card>
    
    <Card class="p-6">
      <div class="flex items-center justify-between">
        <div>
          <p class="text-sm font-medium text-muted-foreground">Storage Gebruikt</p>
          <p class="text-2xl font-bold text-foreground">
            {#if isSuperAdmin && adminStats}
              {formatBytes(adminStats.storage_used || 0)}
            {:else}
              2.4GB
            {/if}
          </p>
        </div>
        <div class="w-10 h-10 bg-purple-100 dark:bg-purple-900 rounded-lg flex items-center justify-center">
          <Icon name="storage" size={20} className="text-purple-600 dark:text-purple-400" />
        </div>
      </div>
    </Card>
    
    <Card class="p-6">
      <div class="flex items-center justify-between">
        <div>
          <p class="text-sm font-medium text-muted-foreground">API Calls (24h)</p>
          <p class="text-2xl font-bold text-foreground">
            {#if isSuperAdmin && adminStats}
              {adminStats.api_calls_24h || 0}
            {:else}
              1,247
            {/if}
          </p>
        </div>
        <div class="w-10 h-10 bg-orange-100 dark:bg-orange-900 rounded-lg flex items-center justify-center">
          <Icon name="functions" size={20} className="text-orange-600 dark:text-orange-400" />
        </div>
      </div>
    </Card>
  </div>
  
  <!-- Admin Stats for Superadmins -->
  {#if isSuperAdmin}
    <div class="space-y-6">
      <div class="flex items-center justify-between">
        <h2 class="text-xl font-semibold text-foreground">Systeem Statistieken</h2>
        <Button
          variant="outline"
          size="sm"
          on:click={loadAdminStats}
          disabled={loadingAdminStats}
          class="flex items-center space-x-2"
        >
          <Icon name="backup" size={14} />
          <span>Vernieuwen</span>
        </Button>
      </div>
      
      {#if loadingAdminStats}
        <Card class="p-6">
          <div class="text-center">
            <div class="animate-spin rounded-full h-8 w-8 border-b-2 border-primary mx-auto"></div>
            <p class="mt-4 text-muted-foreground">Admin statistieken laden...</p>
          </div>
        </Card>
      {:else if adminStats}
        <div class="grid grid-cols-1 md:grid-cols-3 gap-6">
          <Card class="p-6">
            <div class="flex items-center justify-between">
              <div>
                <p class="text-sm font-medium text-muted-foreground">Totaal Gebruikers</p>
                <p class="text-2xl font-bold text-foreground">{adminStats.total_users || 0}</p>
              </div>
              <div class="w-10 h-10 bg-indigo-100 dark:bg-indigo-900 rounded-lg flex items-center justify-center">
                <Icon name="user" size={20} className="text-indigo-600 dark:text-indigo-400" />
              </div>
            </div>
          </Card>
          
          <Card class="p-6">
            <div class="flex items-center justify-between">
              <div>
                <p class="text-sm font-medium text-muted-foreground">Systeem Uptime</p>
                <p class="text-2xl font-bold text-foreground">{adminStats.uptime || '0d'}</p>
              </div>
              <div class="w-10 h-10 bg-emerald-100 dark:bg-emerald-900 rounded-lg flex items-center justify-center">
                <Icon name="shield-check" size={20} className="text-emerald-600 dark:text-emerald-400" />
              </div>
            </div>
          </Card>
          
          <Card class="p-6">
            <div class="flex items-center justify-between">
              <div>
                <p class="text-sm font-medium text-muted-foreground">Database Queries</p>
                <p class="text-2xl font-bold text-foreground">{adminStats.database_queries || 0}</p>
              </div>
              <div class="w-10 h-10 bg-cyan-100 dark:bg-cyan-900 rounded-lg flex items-center justify-center">
                <Icon name="database" size={20} className="text-cyan-600 dark:text-cyan-400" />
              </div>
            </div>
          </Card>
        </div>
        
        <div class="grid grid-cols-1 md:grid-cols-2 gap-6">
          <Card class="p-6">
            <h3 class="text-lg font-semibold text-foreground mb-4">Systeem Informatie</h3>
            <div class="space-y-3 text-sm">
              <div class="flex justify-between">
                <span class="text-muted-foreground">Server OS:</span>
                <span class="text-foreground font-medium">{adminStats.os || 'Onbekend'}</span>
              </div>
              <div class="flex justify-between">
                <span class="text-muted-foreground">CPU Usage:</span>
                <span class="text-foreground font-medium">{adminStats.cpu_usage || '0'}%</span>
              </div>
              <div class="flex justify-between">
                <span class="text-muted-foreground">Memory Usage:</span>
                <span class="text-foreground font-medium">{adminStats.memory_usage || '0'}%</span>
              </div>
              <div class="flex justify-between">
                <span class="text-muted-foreground">Disk Usage:</span>
                <span class="text-foreground font-medium">{adminStats.disk_usage || '0'}%</span>
              </div>
            </div>
          </Card>
          
          <Card class="p-6">
            <h3 class="text-lg font-semibold text-foreground mb-4">Recente Activiteit</h3>
            <div class="space-y-3 text-sm">
              <div class="flex justify-between">
                <span class="text-muted-foreground">Deployments (7d):</span>
                <span class="text-foreground font-medium">{adminStats.deployments_7d || 0}</span>
              </div>
              <div class="flex justify-between">
                <span class="text-muted-foreground">Functions Executed:</span>
                <span class="text-foreground font-medium">{adminStats.functions_executed || 0}</span>
              </div>
              <div class="flex justify-between">
                <span class="text-muted-foreground">Active Sessions:</span>
                <span class="text-foreground font-medium">{adminStats.active_sessions || 0}</span>
              </div>
              <div class="flex justify-between">
                <span class="text-muted-foreground">Error Rate (24h):</span>
                <span class="text-foreground font-medium">{adminStats.error_rate_24h || '0'}%</span>
              </div>
            </div>
          </Card>
        </div>
      {/if}
    </div>
  {/if}

  <!-- Error message -->
  {#if error}
    <Card class="bg-destructive/10 border-destructive/20 p-4">
      <div class="flex justify-between items-center">
        <p class="text-destructive text-sm">{error}</p>
        <Button
          variant="ghost"
          size="sm"
          on:click={() => error = ''}
          class="text-destructive hover:text-destructive/80"
        >
          ×
        </Button>
      </div>
    </Card>
  {/if}

  <!-- Projects Section -->
  <div>
    <div class="flex items-center justify-between mb-6">
      <h2 class="text-xl font-semibold text-foreground">Mijn Projecten</h2>
      {#if projects.length > 0}
        <Button
          variant="outline"
          size="sm"
          on:click={loadProjects}
          class="flex items-center space-x-2"
        >
          <Icon name="backup" size={14} />
          <span>Vernieuwen</span>
        </Button>
      {/if}
    </div>

    {#if loading}
      <Card class="p-12">
        <div class="text-center">
          <div class="animate-spin rounded-full h-8 w-8 border-b-2 border-primary mx-auto"></div>
          <p class="mt-4 text-muted-foreground">Projecten laden...</p>
        </div>
      </Card>
    {:else if projects.length === 0}
      <Card class="p-12">
        <div class="text-center">
          <div class="w-16 h-16 bg-muted rounded-full flex items-center justify-center mx-auto mb-4">
            <Icon name="package" size={32} className="text-muted-foreground" />
          </div>
          <h3 class="text-lg font-medium text-foreground mb-2">Nog geen projecten</h3>
          <p class="text-muted-foreground mb-6 max-w-sm mx-auto">
            Maak je eerste CloudBox project aan om te beginnen met je Backend-as-a-Service
          </p>
          <Button
            href="/dashboard/projects"
            size="lg"
            class="flex items-center space-x-2"
          >
            <Icon name="package" size={16} />
            <span>Eerste Project Aanmaken</span>
          </Button>
        </div>
      </Card>
    {:else}
      <div class="grid gap-6 md:grid-cols-2 lg:grid-cols-3">
        {#each projects as project}
          <Card class="group p-6 hover:shadow-lg hover:shadow-primary/5 transition-all duration-200 border hover:border-primary/20">
            <div class="flex items-start justify-between mb-4">
              <div class="flex items-center space-x-3">
                <div class="w-10 h-10 bg-primary/10 rounded-lg flex items-center justify-center">
                  <Icon name="package" size={20} className="text-primary" />
                </div>
                <div>
                  <h3 class="text-lg font-semibold text-card-foreground group-hover:text-primary transition-colors">
                    {project.name}
                  </h3>
                  {#if project.organization}
                    <div class="flex items-center space-x-1 text-xs">
                      <div 
                        class="w-2 h-2 rounded-full"
                        style="background-color: {project.organization.color}"
                      ></div>
                      <span class="text-muted-foreground">{project.organization.name}</span>
                    </div>
                  {:else}
                    <p class="text-xs text-muted-foreground">
                      Geen organization
                    </p>
                  {/if}
                </div>
              </div>
              <Badge variant={project.is_active ? "default" : "secondary"} class="flex items-center space-x-1">
                <div class="w-2 h-2 rounded-full {project.is_active ? 'bg-green-500' : 'bg-gray-400'}"></div>
                <span>{project.is_active ? 'Actief' : 'Inactief'}</span>
              </Badge>
            </div>
            
            {#if project.description}
              <p class="text-muted-foreground text-sm mb-4 line-clamp-2">{project.description}</p>
            {/if}
            
            <div class="bg-muted/50 rounded-lg p-3 mb-4 space-y-2">
              <div class="flex items-center justify-between text-xs">
                <span class="text-muted-foreground">API Slug:</span>
                <code class="bg-background px-2 py-1 rounded text-xs font-mono">{project.slug}</code>
              </div>
              <div class="flex items-center justify-between text-xs">
                <span class="text-muted-foreground">Aangemaakt:</span>
                <span class="text-foreground">{formatDate(project.created_at)}</span>
              </div>
            </div>
            
            <div class="flex space-x-2">
              <Button
                href="/dashboard/projects/{project.id}"
                size="sm"
                class="flex-1 flex items-center justify-center space-x-2"
              >
                <Icon name="settings" size={14} />
                <span>Beheren</span>
              </Button>
            </div>
          </Card>
        {/each}
      </div>
    {/if}
  </div>
</div>

