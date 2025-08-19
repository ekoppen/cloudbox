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
  import ProjectCard from '$lib/components/ui/project-card.svelte';
  
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
    <div>
      <h1 class="text-heading-1">
        Welkom terug, {$auth.user?.name?.split(' ')[0] || 'User'}
      </h1>
      <p class="mt-1 text-body-sm">
        Hier is een overzicht van je CloudBox projecten
      </p>
    </div>
    <Button
      href="/dashboard/projects"
      class="flex items-center space-x-2"
    >
      <Icon name="plus" size={16} />
      <span>Nieuw project</span>
    </Button>
  </div>

  <!-- Quick Stats -->
  <div class="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-4">
    <div class="rounded-xl border border-border bg-card p-6">
      <div class="flex items-center justify-between">
        <div>
          <p class="text-caption-lg">Projecten</p>
          <p class="text-heading-3">{projects.length}</p>
        </div>
        <div class="flex h-10 w-10 items-center justify-center rounded-lg bg-primary/10">
          <Icon name="package" size={20} className="text-primary" />
        </div>
      </div>
    </div>
    
    <div class="rounded-xl border border-border bg-card p-6">
      <div class="flex items-center justify-between">
        <div>
          <p class="text-caption-lg">Actief</p>
          <p class="text-heading-3">{projects.filter(p => p.is_active).length}</p>
        </div>
        <div class="flex h-10 w-10 items-center justify-center rounded-lg bg-success/10">
          <Icon name="database" size={20} className="text-success" />
        </div>
      </div>
    </div>
    
    <div class="rounded-xl border border-border bg-card p-6">
      <div class="flex items-center justify-between">
        <div>
          <p class="text-caption-lg">Storage</p>
          <p class="text-heading-3">
            {#if isSuperAdmin && adminStats}
              {formatBytes(adminStats.storage_used || 0)}
            {:else}
              2.4GB
            {/if}
          </p>
        </div>
        <div class="flex h-10 w-10 items-center justify-center rounded-lg bg-gray-500/10">
          <Icon name="storage" size={20} className="text-gray-500" />
        </div>
      </div>
    </div>
    
    <div class="rounded-xl border border-border bg-card p-6">
      <div class="flex items-center justify-between">
        <div>
          <p class="text-caption-lg">API Calls</p>
          <p class="text-heading-3">
            {#if isSuperAdmin && adminStats}
              {adminStats.api_calls_24h || 0}
            {:else}
              1,247
            {/if}
          </p>
        </div>
        <div class="flex h-10 w-10 items-center justify-center rounded-lg bg-warning/10">
          <Icon name="zap" size={20} className="text-warning" />
        </div>
      </div>
    </div>
  </div>
  
  <!-- Admin Stats for Superadmins -->
  {#if isSuperAdmin}
    <div class="space-y-6">
      <div class="flex items-center justify-between">
        <h2 class="text-heading-2">Systeem Statistieken</h2>
        <Button
          variant="outline"
          size="sm"
          on:click={loadAdminStats}
          disabled={loadingAdminStats}
          class="flex items-center space-x-2"
        >
          <Icon name="refresh-cw" size={14} />
          <span>Vernieuwen</span>
        </Button>
      </div>
      
      {#if loadingAdminStats}
        <Card class="p-6">
          <div class="text-center">
            <div class="animate-spin rounded-full h-8 w-8 border-b-2 border-primary mx-auto"></div>
            <p class="mt-4 text-body-sm">Admin statistieken laden...</p>
          </div>
        </Card>
      {:else if adminStats}
        <div class="grid grid-cols-1 md:grid-cols-3 gap-6">
          <Card class="p-6">
            <div class="flex items-center justify-between">
              <div>
                <p class="text-caption-lg">Totaal Gebruikers</p>
                <p class="text-heading-3">{adminStats.total_users || 0}</p>
              </div>
              <div class="w-10 h-10 bg-indigo-100 dark:bg-indigo-900 rounded-lg flex items-center justify-center">
                <Icon name="user" size={20} className="text-indigo-600 dark:text-indigo-400" />
              </div>
            </div>
          </Card>
          
          <Card class="p-6">
            <div class="flex items-center justify-between">
              <div>
                <p class="text-caption-lg">Systeem Uptime</p>
                <p class="text-heading-3">{adminStats.uptime || '0d'}</p>
              </div>
              <div class="w-10 h-10 bg-emerald-100 dark:bg-emerald-900 rounded-lg flex items-center justify-center">
                <Icon name="shield-check" size={20} className="text-emerald-600 dark:text-emerald-400" />
              </div>
            </div>
          </Card>
          
          <Card class="p-6">
            <div class="flex items-center justify-between">
              <div>
                <p class="text-caption-lg">Database Queries</p>
                <p class="text-heading-3">{adminStats.database_queries || 0}</p>
              </div>
              <div class="w-10 h-10 bg-cyan-100 dark:bg-cyan-900 rounded-lg flex items-center justify-center">
                <Icon name="database" size={20} className="text-cyan-600 dark:text-cyan-400" />
              </div>
            </div>
          </Card>
        </div>
        
        <div class="grid grid-cols-1 md:grid-cols-2 gap-6">
          <Card class="p-6">
            <h3 class="text-heading-4 mb-4">Systeem Informatie</h3>
            <div class="space-y-3 text-body-sm">
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
            <h3 class="text-heading-4 mb-4">Recente Activiteit</h3>
            <div class="space-y-3 text-body-sm">
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
      <h2 class="text-heading-3">Recente projecten</h2>
      {#if projects.length > 3}
        <Button
          variant="ghost"
          size="sm"
          href="/dashboard/projects"
          class="flex items-center space-x-2 text-muted-foreground"
        >
          <span>Alle projecten</span>
          <Icon name="arrow-right" size={14} />
        </Button>
      {/if}
    </div>

    {#if loading}
      <div class="text-center py-8">
        <div class="animate-spin rounded-full h-6 w-6 border-b-2 border-primary mx-auto"></div>
        <p class="mt-2 text-body-sm">Projecten laden...</p>
      </div>
    {:else if projects.length === 0}
      <div class="text-center py-8">
        <div class="mx-auto mb-4 flex h-12 w-12 items-center justify-center rounded-full bg-muted">
          <Icon name="package" size={24} className="text-muted-foreground" />
        </div>
        <h3 class="mb-2 text-heading-4">Nog geen projecten</h3>
        <p class="mb-4 text-body-sm max-w-sm mx-auto">
          Maak je eerste CloudBox project aan om te beginnen
        </p>
        <Button
          href="/dashboard/projects"
          class="flex items-center space-x-2"
        >
          <Icon name="plus" size={16} />
          <span>Eerste project aanmaken</span>
        </Button>
      </div>
    {:else}
      <div class="grid gap-4 sm:grid-cols-2 lg:grid-cols-3">
        {#each projects.slice(0, 6) as project}
          <ProjectCard {project} />
        {/each}
      </div>
    {/if}
  </div>
</div>

