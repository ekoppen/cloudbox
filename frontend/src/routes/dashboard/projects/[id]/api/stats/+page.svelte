<script lang="ts">
  import { page } from '$app/stores';
  import { onMount } from 'svelte';
  import { auth } from '$lib/stores/auth';
  import { API_ENDPOINTS, createApiRequest } from '$lib/config';
  import Card from '$lib/components/ui/card.svelte';
  import Button from '$lib/components/ui/button.svelte';
  import Badge from '$lib/components/ui/badge.svelte';
  import Icon from '$lib/components/ui/icon.svelte';

  interface APIRouteStatsItem {
    method: string;
    endpoint: string;
    total_requests: number;
    success_requests: number;
    error_requests: number;
    success_rate: number;
    avg_response_time_ms: number;
    total_data_transfer_bytes: number;
    last_used: string;
  }

  interface APIStatsSummary {
    total_requests: number;
    total_endpoints: number;
    overall_success_rate: number;
    avg_response_time_ms: number;
    total_data_transfer_bytes: number;
    top_endpoint: string;
    top_endpoint_count: number;
  }

  interface APIStatsTimelineItem {
    date: string;
    total_requests: number;
    success_requests: number;
    error_requests: number;
    avg_response_time_ms: number;
  }

  interface APIStatsResponse {
    project_id: number;
    routes: APIRouteStatsItem[];
    summary: APIStatsSummary;
    timeline: APIStatsTimelineItem[];
  }

  let apiStats: APIStatsResponse | null = null;
  let loading = true;
  let error = '';
  let selectedPeriod = '30';

  $: projectId = $page.params.id;

  onMount(() => {
    loadAPIStats();
  });

  async function loadAPIStats() {
    if (!projectId) return;
    
    loading = true;
    error = '';
    try {
      const response = await createApiRequest(`${API_ENDPOINTS.projects.apiStats(projectId)}?days=${selectedPeriod}`, {
        headers: {
          'Authorization': `Bearer ${$auth.token}`,
        },
      });

      if (response.ok) {
        apiStats = await response.json();
      } else {
        error = 'Fout bij laden van API statistieken';
      }
    } catch (err) {
      console.error('API stats error:', err);
      error = 'Fout bij laden van API statistieken';
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

  function formatBytes(bytes: number): string {
    if (bytes === 0) return '0 B';
    const k = 1024;
    const sizes = ['B', 'KB', 'MB', 'GB'];
    const i = Math.floor(Math.log(bytes) / Math.log(k));
    return parseFloat((bytes / Math.pow(k, i)).toFixed(2)) + ' ' + sizes[i];
  }

  function formatDateTime(dateStr: string): string {
    if (!dateStr) return 'Nooit';
    const date = new Date(dateStr);
    return date.toLocaleString('nl-NL');
  }

  function getSuccessRateColor(rate: number): string {
    if (rate >= 98) return 'text-green-600 dark:text-green-400';
    if (rate >= 95) return 'text-yellow-600 dark:text-yellow-400';
    return 'text-red-600 dark:text-red-400';
  }

  $: chartMaxRequests = apiStats ? Math.max(...apiStats.timeline.map(t => t.total_requests), 1) : 1;
</script>

<svelte:head>
  <title>API Statistics - CloudBox</title>
</svelte:head>

<div class="space-y-6">
  <!-- Period Selector -->
  <div class="flex items-center justify-end space-x-2">
    <span class="text-sm font-medium text-muted-foreground">Periode:</span>
      <select 
        bind:value={selectedPeriod} 
        on:change={loadAPIStats}
        class="bg-background border border-input px-3 py-1 rounded-md text-sm focus:outline-none focus:ring-2 focus:ring-ring"
      >
        <option value="7">7 dagen</option>
        <option value="30">30 dagen</option>
        <option value="90">90 dagen</option>
        <option value="365">1 jaar</option>
      </select>
    </div>

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
      <Button on:click={loadAPIStats} variant="outline">
        <Icon name="refresh-cw" size={16} className="mr-2" />
        Opnieuw proberen
      </Button>
    </Card>
  {:else if apiStats}
    <!-- Summary Cards -->
    <div class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-4">
      <Card class="glassmorphism-card p-6">
        <div class="flex items-center justify-between">
          <div>
            <p class="text-sm font-medium text-muted-foreground">Totaal Verzoeken</p>
            <p class="text-2xl font-bold text-foreground">{apiStats.summary.total_requests.toLocaleString()}</p>
          </div>
          <div class="w-10 h-10 bg-blue-100 dark:bg-blue-900 rounded-lg flex items-center justify-center">
            <Icon name="functions" size={20} className="text-blue-600 dark:text-blue-400" />
          </div>
        </div>
      </Card>

      <Card class="glassmorphism-card p-6">
        <div class="flex items-center justify-between">
          <div>
            <p class="text-sm font-medium text-muted-foreground">Actieve Endpoints</p>
            <p class="text-2xl font-bold text-foreground">{apiStats.summary.total_endpoints}</p>
          </div>
          <div class="w-10 h-10 bg-green-100 dark:bg-green-900 rounded-lg flex items-center justify-center">
            <Icon name="settings" size={20} className="text-green-600 dark:text-green-400" />
          </div>
        </div>
      </Card>

      <Card class="glassmorphism-card p-6">
        <div class="flex items-center justify-between">
          <div>
            <p class="text-sm font-medium text-muted-foreground">Success Rate</p>
            <p class="text-2xl font-bold {getSuccessRateColor(apiStats.summary.overall_success_rate)}">
              {apiStats.summary.overall_success_rate.toFixed(1)}%
            </p>
          </div>
          <div class="w-10 h-10 bg-yellow-100 dark:bg-yellow-900 rounded-lg flex items-center justify-center">
            <Icon name="shield" size={20} className="text-yellow-600 dark:text-yellow-400" />
          </div>
        </div>
      </Card>

      <Card class="glassmorphism-card p-6">
        <div class="flex items-center justify-between">
          <div>
            <p class="text-sm font-medium text-muted-foreground">Gem. Response Tijd</p>
            <p class="text-2xl font-bold text-foreground">{apiStats.summary.avg_response_time_ms.toFixed(0)}ms</p>
          </div>
          <div class="w-10 h-10 bg-purple-100 dark:bg-purple-900 rounded-lg flex items-center justify-center">
            <Icon name="zap" size={20} className="text-purple-600 dark:text-purple-400" />
          </div>
        </div>
      </Card>
    </div>

    <!-- Charts Section -->
    <div class="grid grid-cols-1 lg:grid-cols-2 gap-6">
      <!-- Timeline Chart -->
      <Card class="glassmorphism-card p-6">
        <div class="flex items-center space-x-3 mb-4">
          <Icon name="functions" size={20} className="text-primary" />
          <h3 class="text-lg font-medium text-foreground">Verzoeken per Dag</h3>
        </div>
        <div class="space-y-3">
          {#each apiStats.timeline as day}
            <div class="flex items-center justify-between">
              <span class="text-sm font-medium text-muted-foreground w-20">{new Date(day.date).toLocaleDateString('nl-NL', { day: '2-digit', month: '2-digit' })}</span>
              <div class="flex-1 mx-4">
                <div class="bg-muted rounded-full h-3">
                  <div 
                    class="bg-primary h-3 rounded-full" 
                    style="width: {(day.total_requests / chartMaxRequests) * 100}%"
                  ></div>
                </div>
              </div>
              <div class="flex items-center space-x-2">
                <span class="text-sm font-bold text-foreground w-12 text-right">{day.total_requests}</span>
                {#if day.error_requests > 0}
                  <Badge variant="destructive" class="text-xs px-1 py-0">{day.error_requests} errors</Badge>
                {/if}
              </div>
            </div>
          {/each}
        </div>
      </Card>

      <!-- Top Endpoints -->
      <Card class="glassmorphism-card p-6">
        <div class="flex items-center space-x-3 mb-4">
          <Icon name="trending-up" size={20} className="text-primary" />
          <h3 class="text-lg font-medium text-foreground">Top Endpoints</h3>
        </div>
        <div class="space-y-3">
          {#each apiStats.routes.slice(0, 5) as route}
            <div class="flex items-center justify-between">
              <div class="flex items-center space-x-3 flex-1">
                <Badge variant="secondary" class={getMethodColor(route.method)}>
                  {route.method}
                </Badge>
                <span class="text-sm font-mono text-foreground truncate">{route.endpoint}</span>
              </div>
              <div class="text-sm font-bold text-foreground">{route.total_requests}</div>
            </div>
          {/each}
        </div>
      </Card>
    </div>

    <!-- Detailed Routes Table -->
    <Card class="glassmorphism-card">
      <div class="px-6 py-4 border-b border-border">
        <div class="flex items-center space-x-3">
          <Icon name="database" size={20} className="text-primary" />
          <h3 class="text-lg font-medium text-foreground">API Route Details</h3>
        </div>
      </div>
      <div class="overflow-x-auto">
        <table class="w-full">
          <thead class="bg-muted/30">
            <tr>
              <th class="text-left p-4 text-sm font-medium text-muted-foreground">Endpoint</th>
              <th class="text-center p-4 text-sm font-medium text-muted-foreground">Verzoeken</th>
              <th class="text-center p-4 text-sm font-medium text-muted-foreground">Success Rate</th>
              <th class="text-center p-4 text-sm font-medium text-muted-foreground">Avg. Tijd</th>
              <th class="text-center p-4 text-sm font-medium text-muted-foreground">Data Transfer</th>
              <th class="text-center p-4 text-sm font-medium text-muted-foreground">Laatst Gebruikt</th>
            </tr>
          </thead>
          <tbody class="divide-y divide-border">
            {#each apiStats.routes as route}
              <tr class="hover:bg-muted/30 transition-colors">
                <td class="p-4">
                  <div class="flex items-center space-x-3">
                    <Badge variant="secondary" class={getMethodColor(route.method)}>
                      {route.method}
                    </Badge>
                    <span class="font-mono text-sm text-foreground">{route.endpoint}</span>
                  </div>
                </td>
                <td class="text-center p-4">
                  <div class="text-sm">
                    <div class="font-bold text-foreground">{route.total_requests.toLocaleString()}</div>
                    {#if route.error_requests > 0}
                      <div class="text-xs text-destructive">{route.error_requests} errors</div>
                    {/if}
                  </div>
                </td>
                <td class="text-center p-4">
                  <span class="text-sm font-medium {getSuccessRateColor(route.success_rate)}">
                    {route.success_rate.toFixed(1)}%
                  </span>
                </td>
                <td class="text-center p-4">
                  <span class="text-sm font-medium text-foreground">{route.avg_response_time_ms.toFixed(0)}ms</span>
                </td>
                <td class="text-center p-4">
                  <span class="text-sm font-medium text-foreground">{formatBytes(route.total_data_transfer_bytes)}</span>
                </td>
                <td class="text-center p-4">
                  <span class="text-xs text-muted-foreground">{formatDateTime(route.last_used)}</span>
                </td>
              </tr>
            {/each}
          </tbody>
        </table>
      </div>
    </Card>

    {#if apiStats.routes.length === 0}
      <Card class="glassmorphism-content p-12 text-center">
        <div class="w-16 h-16 bg-muted rounded-lg flex items-center justify-center mx-auto mb-4">
          <Icon name="functions" size={32} className="text-muted-foreground" />
        </div>
        <h3 class="text-lg font-medium text-foreground mb-2">Geen API data gevonden</h3>
        <p class="text-muted-foreground mb-4">
          Er zijn nog geen API verzoeken geregistreerd voor de geselecteerde periode.
        </p>
        <p class="text-sm text-muted-foreground">
          API logging is automatisch geactiveerd. Maak gebruik van je API endpoints om statistieken te zien.
        </p>
      </Card>
    {/if}
  {/if}
</div>