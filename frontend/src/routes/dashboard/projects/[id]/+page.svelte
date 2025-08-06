<script lang="ts">
  import { page } from '$app/stores';
  import { onMount } from 'svelte';
  import { auth } from '$lib/stores/auth';
  import { API_ENDPOINTS, createApiRequest } from '$lib/config';
  import Card from '$lib/components/ui/card.svelte';
  import Button from '$lib/components/ui/button.svelte';
  import Icon from '$lib/components/ui/icon.svelte';

  interface ProjectStats {
    requests_today: number;
    requests_week: number;
    requests_month: number;
    api_keys_count: number;
    database_tables: number;
    storage_used: number;
    users_count: number;
    deployments_count: number;
  }

  interface RecentActivity {
    id: number;
    type: string;
    message: string;
    timestamp: string;
    status: 'success' | 'error' | 'warning';
  }

  let stats: ProjectStats = {
    requests_today: 0,
    requests_week: 0,
    requests_month: 0,
    api_keys_count: 0,
    database_tables: 0,
    storage_used: 0,
    users_count: 0,
    deployments_count: 0
  };

  let recentActivity: RecentActivity[] = [];
  let loadingStats = true;

  $: projectId = $page.params.id;

  // Chart data - will be loaded from API
  let chartData: Array<{day: string, requests: number}> = [];

  onMount(() => {
    loadProjectStats();
  });

  async function loadProjectStats() {
    if (!projectId) return;
    
    loadingStats = true;
    try {
      const response = await createApiRequest(API_ENDPOINTS.projects.stats(projectId), {
        headers: {
          'Authorization': `Bearer ${$auth.token}`,
        },
      });

      if (response.ok) {
        const data = await response.json();
        stats = data;
        chartData = data.activity_data || [];
      } else {
        console.error('Failed to load project stats');
      }
    } catch (err) {
      console.error('Project stats error:', err);
    } finally {
      loadingStats = false;
    }
  }

  function getStatusColor(status: string) {
    switch (status) {
      case 'success': return 'text-green-600 dark:text-green-400 bg-green-100 dark:bg-green-900';
      case 'error': return 'text-red-600 dark:text-red-400 bg-red-100 dark:bg-red-900';
      case 'warning': return 'text-yellow-600 dark:text-yellow-400 bg-yellow-100 dark:bg-yellow-900';
      default: return 'text-muted-foreground bg-muted';
    }
  }

  function getStatusIcon(status: string) {
    switch (status) {
      case 'success': return '✅';
      case 'error': return '❌';
      case 'warning': return '⚠️';
      default: return 'ℹ️';
    }
  }
</script>

<svelte:head>
  <title>Project Overzicht - CloudBox</title>
</svelte:head>

<div class="space-y-6">
  <!-- Welcome Section -->
  <div class="flex items-center space-x-4">
    <div class="w-12 h-12 bg-primary rounded-xl flex items-center justify-center">
      <Icon name="dashboard" size={24} color="white" />
    </div>
    <div>
      <h1 class="text-2xl font-bold text-foreground">Project Overzicht</h1>
      <p class="mt-1 text-sm text-muted-foreground">
        Beheer en monitor je CloudBox project
      </p>
    </div>
  </div>

  <!-- Quick Stats Grid -->
  {#if loadingStats}
    <div class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-4">
      {#each Array(4) as _}
        <Card class="p-6">
          <div class="animate-pulse">
            <div class="h-4 bg-muted rounded w-3/4 mb-2"></div>
            <div class="h-8 bg-muted rounded w-1/2"></div>
          </div>
        </Card>
      {/each}
    </div>
  {:else}
    <div class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-4">
      <Card class="p-6">
        <div class="flex items-center justify-between">
          <div>
            <p class="text-sm font-medium text-muted-foreground">Verzoeken vandaag</p>
            <p class="text-2xl font-bold text-foreground">{stats.requests_today.toLocaleString()}</p>
          </div>
          <div class="w-10 h-10 bg-blue-100 dark:bg-blue-900 rounded-lg flex items-center justify-center">
            <Icon name="functions" size={20} className="text-blue-600 dark:text-blue-400" />
          </div>
        </div>
      </Card>

    <Card class="p-6">
      <div class="flex items-center justify-between">
        <div>
          <p class="text-sm font-medium text-muted-foreground">Gebruikers</p>
          <p class="text-2xl font-bold text-foreground">{stats.users_count}</p>
        </div>
        <div class="w-10 h-10 bg-green-100 dark:bg-green-900 rounded-lg flex items-center justify-center">
          <Icon name="user" size={20} className="text-green-600 dark:text-green-400" />
        </div>
      </div>
    </Card>

    <Card class="p-6">
      <div class="flex items-center justify-between">
        <div>
          <p class="text-sm font-medium text-muted-foreground">Database Tabellen</p>
          <p class="text-2xl font-bold text-foreground">{stats.database_tables}</p>
        </div>
        <div class="w-10 h-10 bg-yellow-100 dark:bg-yellow-900 rounded-lg flex items-center justify-center">
          <Icon name="database" size={20} className="text-yellow-600 dark:text-yellow-400" />
        </div>
      </div>
    </Card>

    <Card class="p-6">
      <div class="flex items-center justify-between">
        <div>
          <p class="text-sm font-medium text-muted-foreground">API Keys</p>
          <p class="text-2xl font-bold text-foreground">{stats.api_keys_count}</p>
        </div>
        <div class="w-10 h-10 bg-purple-100 dark:bg-purple-900 rounded-lg flex items-center justify-center">
          <Icon name="auth" size={20} className="text-purple-600 dark:text-purple-400" />
        </div>
      </div>
    </Card>
    </div>
  {/if}

  <!-- Charts Section -->
  <div class="grid grid-cols-1 lg:grid-cols-2 gap-6">
    <!-- API Requests Chart -->
    <Card class="p-6">
      <div class="flex items-center space-x-3 mb-4">
        <Icon name="functions" size={20} className="text-primary" />
        <h3 class="text-lg font-medium text-foreground">API Verzoeken (Deze Week)</h3>
      </div>
      <div class="space-y-3">
        {#each chartData as day}
          <div class="flex items-center justify-between">
            <span class="text-sm font-medium text-muted-foreground">{day.day}</span>
            <div class="flex-1 mx-4">
              <div class="bg-muted rounded-full h-2">
                <div 
                  class="bg-primary h-2 rounded-full" 
                  style="width: {(day.requests / 2000) * 100}%"
                ></div>
              </div>
            </div>
            <span class="text-sm font-bold text-foreground">{day.requests}</span>
          </div>
        {/each}
      </div>
    </Card>

    <!-- Usage Stats -->
    <Card class="p-6">
      <div class="flex items-center space-x-3 mb-4">
        <Icon name="storage" size={20} className="text-primary" />
        <h3 class="text-lg font-medium text-foreground">Gebruik Statistieken</h3>
      </div>
      <div class="space-y-4">
        <div>
          <div class="flex justify-between text-sm">
            <span class="text-muted-foreground">Opslag gebruikt</span>
            <span class="font-medium text-foreground">{stats.storage_used} MB / 1000 MB</span>
          </div>
          <div class="mt-1 bg-muted rounded-full h-2">
            <div 
              class="bg-green-500 h-2 rounded-full" 
              style="width: {(stats.storage_used / 1000) * 100}%"
            ></div>
          </div>
        </div>

        <div>
          <div class="flex justify-between text-sm">
            <span class="text-muted-foreground">API Limiet</span>
            <span class="font-medium text-foreground">{stats.requests_month.toLocaleString()} / 100,000</span>
          </div>
          <div class="mt-1 bg-muted rounded-full h-2">
            <div 
              class="bg-blue-500 h-2 rounded-full" 
              style="width: {(stats.requests_month / 100000) * 100}%"
            ></div>
          </div>
        </div>

        <div class="grid grid-cols-2 gap-4 pt-4">
          <div class="text-center">
            <p class="text-2xl font-bold text-foreground">{stats.deployments_count}</p>
            <p class="text-sm text-muted-foreground">Deployments</p>
          </div>
          <div class="text-center">
            <p class="text-2xl font-bold text-foreground">99.9%</p>
            <p class="text-sm text-muted-foreground">Uptime</p>
          </div>
        </div>
      </div>
    </Card>
  </div>

  <!-- Recent Activity -->
  <Card>
    <div class="px-6 py-4 border-b border-border">
      <div class="flex items-center space-x-3">
        <Icon name="backup" size={20} className="text-primary" />
        <h3 class="text-lg font-medium text-foreground">Recente Activiteit</h3>
      </div>
    </div>
    <div class="divide-y divide-border">
      {#each recentActivity as activity}
        <div class="px-6 py-4 flex items-center space-x-4">
          <div class="flex-shrink-0">
            <span class="inline-flex items-center justify-center h-8 w-8 rounded-full {getStatusColor(activity.status)}">
              {getStatusIcon(activity.status)}
            </span>
          </div>
          <div class="flex-1 min-w-0">
            <p class="text-sm font-medium text-foreground">{activity.message}</p>
            <p class="text-sm text-muted-foreground">{activity.type} • {activity.timestamp}</p>
          </div>
          <div class="flex-shrink-0">
            <Button variant="ghost" size="sm" class="text-muted-foreground hover:text-foreground">
              Meer info
            </Button>
          </div>
        </div>
      {/each}
    </div>
    <div class="px-6 py-3 bg-muted/30 text-center">
      <Button variant="ghost" size="sm" class="text-primary hover:text-primary/80 font-medium">
        Alle activiteit bekijken
      </Button>
    </div>
  </Card>

  <!-- Quick Actions -->
  <Card class="p-6">
    <div class="flex items-center space-x-3 mb-4">
      <Icon name="settings" size={20} className="text-primary" />
      <h3 class="text-lg font-medium text-foreground">Snelle Acties</h3>
    </div>
    <div class="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-4">
      <Button href="/dashboard/projects/{projectId}/database" variant="outline" class="flex items-center justify-center space-x-2 h-12">
        <Icon name="database" size={16} />
        <span>Database Beheren</span>
      </Button>
      <Button href="/dashboard/projects/{projectId}/settings" variant="outline" class="flex items-center justify-center space-x-2 h-12">
        <Icon name="auth" size={16} />
        <span>API Keys</span>
      </Button>
      <Button href="/dashboard/projects/{projectId}/settings" variant="outline" class="flex items-center justify-center space-x-2 h-12">
        <Icon name="settings" size={16} />
        <span>CORS Instellen</span>
      </Button>
      <Button class="flex items-center justify-center space-x-2 h-12">
        <Icon name="backup" size={16} />
        <span>Backup Maken</span>
      </Button>
    </div>
  </Card>
</div>