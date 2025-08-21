<script lang="ts">
  import { onMount } from 'svelte';
  import { API_BASE_URL, API_ENDPOINTS, createApiRequest } from '$lib/config';
  import { auth } from '$lib/stores/auth';
  import { toast } from '$lib/stores/toast';
  import Card from '$lib/components/ui/card.svelte';
  import Button from '$lib/components/ui/button.svelte';
  import Icon from '$lib/components/ui/icon.svelte';

  // Statistics data
  let systemStats = {
    totalUsers: 0,
    totalProjects: 0,
    totalDeployments: 0,
    totalFunctions: 0,
    totalDocuments: 0,
    totalStorageFiles: 0,
    totalStorageSize: 0,
    activeUsers: 0,
    inactiveUsers: 0
  };

  let userGrowth = [];
  let projectActivity = [];
  let functionExecutions = [];
  let deploymentStats = [];
  let storageStats = [];
  let systemHealth = {
    cpuUsage: 0,
    memoryUsage: 0,
    diskUsage: 0,
    apiResponseTime: 0,
    activeConnections: 0
  };

  let loading = true;
  let selectedPeriod = '7d';

  const periods = [
    { value: '24h', label: '24 uur' },
    { value: '7d', label: '7 dagen' },
    { value: '30d', label: '30 dagen' },
    { value: '90d', label: '90 dagen' }
  ];

  onMount(async () => {
    await loadAllStats();
  });

  async function loadAllStats() {
    loading = true;
    try {
      const headers = {
        'Content-Type': 'application/json',
        ...auth.getAuthHeader()
      };

      // Load general system statistics
      await Promise.all([
        loadSystemStats(headers),
        loadUserGrowth(headers),
        loadProjectActivity(headers),
        loadFunctionExecutions(headers),
        loadDeploymentStats(headers),
        loadStorageStats(headers),
        loadSystemHealth(headers)
      ]);

    } catch (error) {
      console.error('Error loading admin statistics:', error);
      toast.error('Fout bij laden van admin statistieken');
    } finally {
      loading = false;
    }
  }

  async function loadSystemStats(headers) {
    try {
      const response = await createApiRequest(API_ENDPOINTS.admin.stats.system, { headers });
      if (response.ok) {
        const stats = await response.json();
        systemStats = {
          totalUsers: stats.total_users,
          totalProjects: stats.total_projects,
          totalDeployments: stats.total_deployments,
          totalFunctions: stats.total_functions,
          totalDocuments: stats.total_documents,
          totalStorageFiles: stats.total_files,
          totalStorageSize: stats.total_storage_size,
          activeUsers: stats.active_users,
          inactiveUsers: stats.inactive_users
        };
      } else {
        console.error('Failed to load system stats:', response.status, await response.text());
      }
    } catch (error) {
      console.error('Error loading system stats:', error);
    }
  }

  async function loadUserGrowth(headers) {
    try {
      const response = await createApiRequest(API_ENDPOINTS.admin.stats.userGrowth + '?days=30', { headers });
      if (response.ok) {
        userGrowth = await response.json();
      } else {
        console.error('Failed to load user growth:', response.status);
        userGrowth = [];
      }
    } catch (error) {
      console.error('Error loading user growth:', error);
      userGrowth = [];
    }
  }

  async function loadProjectActivity(headers) {
    try {
      const response = await createApiRequest(API_ENDPOINTS.admin.stats.projectActivity + '?days=7', { headers });
      if (response.ok) {
        projectActivity = await response.json();
      } else {
        console.error('Failed to load project activity:', response.status);
        projectActivity = [];
      }
    } catch (error) {
      console.error('Error loading project activity:', error);
      projectActivity = [];
    }
  }

  async function loadFunctionExecutions(headers) {
    try {
      const response = await createApiRequest(API_ENDPOINTS.admin.stats.functionExecutions + '?hours=24', { headers });
      if (response.ok) {
        functionExecutions = await response.json();
      } else {
        console.error('Failed to load function executions:', response.status);
        functionExecutions = [];
      }
    } catch (error) {
      console.error('Error loading function executions:', error);
      functionExecutions = [];
    }
  }

  async function loadDeploymentStats(headers) {
    try {
      const response = await createApiRequest(API_ENDPOINTS.admin.stats.deploymentStats, { headers });
      if (response.ok) {
        deploymentStats = await response.json();
      } else {
        console.error('Failed to load deployment stats:', response.status);
        deploymentStats = [];
      }
    } catch (error) {
      console.error('Error loading deployment stats:', error);
    }
  }

  async function loadStorageStats(headers) {
    try {
      const response = await createApiRequest(API_ENDPOINTS.admin.stats.storageStats, { headers });
      if (response.ok) {
        storageStats = await response.json();
      } else {
        console.error('Failed to load storage stats:', response.status);
        storageStats = [];
      }
    } catch (error) {
      console.error('Error loading storage stats:', error);
    }
  }

  async function loadSystemHealth(headers) {
    try {
      const response = await createApiRequest(API_ENDPOINTS.admin.stats.systemHealth, { headers });
      if (response.ok) {
        const health = await response.json();
        systemHealth = {
          cpuUsage: health.cpu_usage,
          memoryUsage: health.memory_usage,
          diskUsage: health.disk_usage,
          apiResponseTime: health.api_response_time,
          activeConnections: health.active_connections
        };
      } else {
        console.error('Failed to load system health:', response.status);
        systemHealth = {
          cpuUsage: 0,
          memoryUsage: 0,
          diskUsage: 0,
          apiResponseTime: 0,
          activeConnections: 0
        };
      }
    } catch (error) {
      console.error('Error loading system health:', error);
    }
  }

  function formatBytes(bytes) {
    if (bytes === 0) return '0 B';
    const k = 1024;
    const sizes = ['B', 'KB', 'MB', 'GB', 'TB'];
    const i = Math.floor(Math.log(bytes) / Math.log(k));
    return parseFloat((bytes / Math.pow(k, i)).toFixed(2)) + ' ' + sizes[i];
  }

  function formatNumber(num) {
    return new Intl.NumberFormat('nl-NL').format(num);
  }

  async function refreshStats() {
    await loadAllStats();
    toast.success('Statistieken vernieuwd');
  }

  async function exportStats() {
    try {
      const data = {
        systemStats,
        userGrowth,
        projectActivity,
        functionExecutions,
        deploymentStats,
        storageStats,
        systemHealth,
        exportedAt: new Date().toISOString()
      };

      const blob = new Blob([JSON.stringify(data, null, 2)], { 
        type: 'application/json' 
      });
      
      const url = URL.createObjectURL(blob);
      const a = document.createElement('a');
      a.href = url;
      a.download = `cloudbox-admin-stats-${new Date().toISOString().split('T')[0]}.json`;
      a.click();
      URL.revokeObjectURL(url);
      
      toast.success('Statistieken geëxporteerd');
    } catch (error) {
      console.error('Export failed:', error);
      toast.error('Export mislukt');
    }
  }
</script>

<div class="space-y-8">
    <!-- Header -->
    <div class="mb-8">
      <div class="flex items-center justify-between">
        <div>
          <h1 class="text-3xl font-bold text-gray-900 dark:text-white">
            Admin Dashboard
          </h1>
          <p class="text-gray-600 dark:text-gray-400 mt-1">
            Overzicht van systeem statistieken en prestaties
          </p>
        </div>
        
        <div class="flex space-x-3">
          <select bind:value={selectedPeriod} on:change={loadAllStats}
                  class="px-3 py-2 border border-gray-300 rounded-md shadow-sm focus:outline-none focus:ring-blue-500 focus:border-blue-500 dark:bg-gray-700 dark:border-gray-600 dark:text-white">
            {#each periods as period}
              <option value={period.value}>{period.label}</option>
            {/each}
          </select>
          
          <Button 
            on:click={refreshStats} 
            disabled={loading}
            variant="ghost"
            size="icon"
            class="hover:rotate-180 transition-transform duration-300"
            title="Vernieuwen"
          >
            <Icon name="refresh-cw" class="w-5 h-5" />
          </Button>
          
          <Button on:click={exportStats} variant="outline">
            <Icon name="download" class="w-4 h-4 mr-2" />
            Exporteren
          </Button>
        </div>
      </div>
    </div>

    {#if loading}
      <div class="flex justify-center items-center h-64">
        <div class="animate-spin rounded-full h-12 w-12 border-b-2 border-blue-500"></div>
      </div>
    {:else}
      <!-- System Overview Cards -->
      <div class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-6 mb-8">
        <Card class="p-6">
          <div class="flex items-center">
            <div class="p-2 bg-blue-100 dark:bg-gray-800 rounded-lg">
              <Icon name="users" class="w-6 h-6 text-blue-600" />
            </div>
            <div class="ml-4">
              <p class="text-sm font-medium text-gray-600 dark:text-gray-400">Totaal Gebruikers</p>
              <p class="text-2xl font-bold text-gray-900 dark:text-white">
                {formatNumber(systemStats.totalUsers)}
              </p>
              <p class="text-xs text-green-600">
                {systemStats.activeUsers} actief
              </p>
            </div>
          </div>
        </Card>

        <Card class="p-6">
          <div class="flex items-center">
            <div class="p-2 bg-green-100 dark:bg-green-900 rounded-lg">
              <Icon name="folder" class="w-6 h-6 text-green-600" />
            </div>
            <div class="ml-4">
              <p class="text-sm font-medium text-gray-600 dark:text-gray-400">Projecten</p>
              <p class="text-2xl font-bold text-gray-900 dark:text-white">
                {formatNumber(systemStats.totalProjects)}
              </p>
            </div>
          </div>
        </Card>

        <Card class="p-6">
          <div class="flex items-center">
            <div class="p-2 bg-purple-100 dark:bg-purple-900 rounded-lg">
              <Icon name="code" class="w-6 h-6 text-purple-600" />
            </div>
            <div class="ml-4">
              <p class="text-sm font-medium text-gray-600 dark:text-gray-400">Functies</p>
              <p class="text-2xl font-bold text-gray-900 dark:text-white">
                {formatNumber(systemStats.totalFunctions)}
              </p>
            </div>
          </div>
        </Card>

        <Card class="p-6">
          <div class="flex items-center">
            <div class="p-2 bg-orange-100 dark:bg-orange-900 rounded-lg">
              <Icon name="server" class="w-6 h-6 text-orange-600" />
            </div>
            <div class="ml-4">
              <p class="text-sm font-medium text-gray-600 dark:text-gray-400">Deployments</p>
              <p class="text-2xl font-bold text-gray-900 dark:text-white">
                {formatNumber(systemStats.totalDeployments)}
              </p>
            </div>
          </div>
        </Card>
      </div>

      <!-- Charts Row 1 -->
      <div class="grid grid-cols-1 lg:grid-cols-2 gap-6 mb-8">
        <!-- User Growth Chart -->
        <Card class="p-6">
          <h3 class="text-lg font-semibold text-gray-900 dark:text-white mb-4">
            Gebruikers Groei (30 dagen)
          </h3>
          <div class="h-64 bg-gray-100 dark:bg-gray-800 rounded-lg flex items-center justify-center">
            <div class="text-center">
              <Icon name="trending-up" class="w-12 h-12 text-gray-400 mx-auto mb-2" />
              <p class="text-gray-500 dark:text-gray-400">
                +{userGrowth.reduce((sum, day) => sum + (day.new_users || day.newUsers || 0), 0)} nieuwe gebruikers
              </p>
              <p class="text-sm text-gray-400">
                {userGrowth[userGrowth.length - 1]?.total_users || userGrowth[userGrowth.length - 1]?.totalUsers || 0} totaal
              </p>
            </div>
          </div>
        </Card>

        <!-- Function Executions -->
        <Card class="p-6">
          <h3 class="text-lg font-semibold text-gray-900 dark:text-white mb-4">
            Function Executions (24u)
          </h3>
          <div class="h-64 bg-gray-100 dark:bg-gray-800 rounded-lg flex items-center justify-center">
            <div class="text-center">
              <Icon name="zap" class="w-12 h-12 text-yellow-500 mx-auto mb-2" />
              <p class="text-2xl font-bold text-gray-900 dark:text-white">
                {formatNumber(functionExecutions.reduce((sum, hour) => sum + hour.executions, 0))}
              </p>
              <p class="text-sm text-gray-500 dark:text-gray-400">
                Totaal executions vandaag
              </p>
              <p class="text-xs text-red-500 mt-1">
                {functionExecutions.reduce((sum, hour) => sum + hour.errors, 0)} errors
              </p>
            </div>
          </div>
        </Card>
      </div>

      <!-- Charts Row 2 -->
      <div class="grid grid-cols-1 lg:grid-cols-3 gap-6 mb-8">
        <!-- Deployment Status -->
        <Card class="p-6">
          <h3 class="text-lg font-semibold text-gray-900 dark:text-white mb-4">
            Deployment Status
          </h3>
          <div class="space-y-3">
            {#each deploymentStats as stat}
              <div class="flex items-center justify-between">
                <div class="flex items-center">
                  <div class="w-3 h-3 rounded-full mr-3" style="background-color: {stat.color}"></div>
                  <span class="text-sm text-gray-600 dark:text-gray-400 capitalize">
                    {stat.status}
                  </span>
                </div>
                <span class="text-sm font-medium text-gray-900 dark:text-white">
                  {stat.count}
                </span>
              </div>
            {/each}
          </div>
        </Card>

        <!-- Storage Stats -->
        <Card class="p-6">
          <h3 class="text-lg font-semibold text-gray-900 dark:text-white mb-4">
            Storage Overzicht
          </h3>
          <div class="space-y-3">
            {#each storageStats as stat}
              <div class="flex items-center justify-between">
                <div class="flex items-center">
                  <div class="w-3 h-3 rounded-full mr-3" style="background-color: {stat.color}"></div>
                  <span class="text-sm text-gray-600 dark:text-gray-400">
                    {stat.type}
                  </span>
                </div>
                <div class="text-right">
                  <div class="text-sm font-medium text-gray-900 dark:text-white">
                    {stat.count} bestanden
                  </div>
                  <div class="text-xs text-gray-500">
                    {formatBytes(stat.size)}
                  </div>
                </div>
              </div>
            {/each}
          </div>
          <div class="mt-4 pt-3 border-t border-gray-200 dark:border-gray-700">
            <div class="flex justify-between text-sm">
              <span class="text-gray-600 dark:text-gray-400">Totaal</span>
              <span class="font-medium text-gray-900 dark:text-white">
                {formatBytes(systemStats.totalStorageSize)}
              </span>
            </div>
          </div>
        </Card>

        <!-- System Health -->
        <Card class="p-6">
          <h3 class="text-lg font-semibold text-gray-900 dark:text-white mb-4">
            Systeem Gezondheid
          </h3>
          <div class="space-y-4">
            <div>
              <div class="flex justify-between text-sm mb-1">
                <span class="text-gray-600 dark:text-gray-400">CPU</span>
                <span class="text-gray-900 dark:text-white">{systemHealth.cpuUsage}%</span>
              </div>
              <div class="w-full bg-gray-200 rounded-full h-2">
                <div class="bg-blue-600 h-2 rounded-full" style="width: {systemHealth.cpuUsage}%"></div>
              </div>
            </div>

            <div>
              <div class="flex justify-between text-sm mb-1">
                <span class="text-gray-600 dark:text-gray-400">Memory</span>
                <span class="text-gray-900 dark:text-white">{systemHealth.memoryUsage}%</span>
              </div>
              <div class="w-full bg-gray-200 rounded-full h-2">
                <div class="bg-green-600 h-2 rounded-full" style="width: {systemHealth.memoryUsage}%"></div>
              </div>
            </div>

            <div>
              <div class="flex justify-between text-sm mb-1">
                <span class="text-gray-600 dark:text-gray-400">Disk</span>
                <span class="text-gray-900 dark:text-white">{systemHealth.diskUsage}%</span>
              </div>
              <div class="w-full bg-gray-200 rounded-full h-2">
                <div class="bg-orange-600 h-2 rounded-full" style="width: {systemHealth.diskUsage}%"></div>
              </div>
            </div>

            <div class="pt-2 border-t border-gray-200 dark:border-gray-700">
              <div class="flex justify-between text-xs text-gray-500 dark:text-gray-400">
                <span>API Response</span>
                <span>{systemHealth.apiResponseTime}ms</span>
              </div>
              <div class="flex justify-between text-xs text-gray-500 dark:text-gray-400 mt-1">
                <span>Connections</span>
                <span>{systemHealth.activeConnections}</span>
              </div>
            </div>
          </div>
        </Card>
      </div>

      <!-- Database & Activity Table -->
      <div class="grid grid-cols-1 lg:grid-cols-2 gap-6">
        <!-- Database Statistics -->
        <Card class="p-6">
          <h3 class="text-lg font-semibold text-gray-900 dark:text-white mb-4">
            Database Statistieken
          </h3>
          <div class="space-y-3">
            <div class="flex justify-between">
              <span class="text-gray-600 dark:text-gray-400">Documenten</span>
              <span class="font-medium text-gray-900 dark:text-white">
                {formatNumber(systemStats.totalDocuments)}
              </span>
            </div>
            <div class="flex justify-between">
              <span class="text-gray-600 dark:text-gray-400">Bestanden</span>
              <span class="font-medium text-gray-900 dark:text-white">
                {formatNumber(systemStats.totalStorageFiles)}
              </span>
            </div>
            <div class="flex justify-between">
              <span class="text-gray-600 dark:text-gray-400">Storage Grootte</span>
              <span class="font-medium text-gray-900 dark:text-white">
                {formatBytes(systemStats.totalStorageSize)}
              </span>
            </div>
            <div class="flex justify-between">
              <span class="text-gray-600 dark:text-gray-400">Function Executions</span>
              <span class="font-medium text-gray-900 dark:text-white">
                {formatNumber(functionExecutions.reduce((sum, hour) => sum + hour.executions, 0))}
              </span>
            </div>
          </div>
        </Card>

        <!-- Recent Activity -->
        <Card class="p-6">
          <h3 class="text-lg font-semibold text-gray-900 dark:text-white mb-4">
            Recente Activiteit (7 dagen)
          </h3>
          <div class="space-y-2">
            {#each projectActivity.slice(-5) as activity}
              <div class="flex justify-between text-sm">
                <span class="text-gray-600 dark:text-gray-400">
                  {activity.date}
                </span>
                <div class="flex space-x-4">
                  <span class="text-green-600">{activity.created} gemaakt</span>
                  <span class="text-blue-600">{activity.deployed} deployed</span>
                  <span class="text-orange-600">{activity.updated} updates</span>
                </div>
              </div>
            {/each}
          </div>
        </Card>
      </div>
    {/if}
</div>

<style>
  /* Custom scrollbar styles */
  :global(.admin-dashboard) {
    scrollbar-width: thin;
    scrollbar-color: #cbd5e0 #f7fafc;
  }

  :global(.admin-dashboard::-webkit-scrollbar) {
    width: 6px;
  }

  :global(.admin-dashboard::-webkit-scrollbar-track) {
    background: #f7fafc;
  }

  :global(.admin-dashboard::-webkit-scrollbar-thumb) {
    background-color: #cbd5e0;
    border-radius: 3px;
  }
</style>