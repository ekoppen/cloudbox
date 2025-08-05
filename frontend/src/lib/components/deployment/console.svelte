<script lang="ts">
  import { onMount, onDestroy } from 'svelte';
  import { API_BASE_URL } from '$lib/config';
  import { auth } from '$lib/stores/auth';
  import Icon from '$lib/components/ui/icon.svelte';
  import Button from '$lib/components/ui/button.svelte';

  export let deploymentId: string;
  export let isVisible: boolean = false;
  export let projectId: string;
  export let onClose: (() => void) | null = null;

  let logs: any[] = [];
  let isCollapsed: boolean = false;
  let isConnected: boolean = false;
  let consoleElement: HTMLElement;
  let pollingInterval: number | null = null;
  let statusInterval: number | null = null;
  let lastLogCount: number = 0;
  let status: string = 'pending';
  let progress: number = 0;
  let deploymentStats: any = null;

  // Auto scroll to bottom
  function scrollToBottom() {
    if (consoleElement && !isCollapsed) {
      setTimeout(() => {
        consoleElement.scrollTop = consoleElement.scrollHeight;
      }, 10);
    }
  }

  // Poll for new logs every 2 seconds when deployment is active
  async function pollLogs() {
    if (!deploymentId || !isVisible) return;
    
    try {
      const response = await fetch(`${API_BASE_URL}/api/v1/projects/${projectId}/deployments/${deploymentId}/logs`, {
        headers: {
          'Content-Type': 'application/json',
          ...auth.getAuthHeader()
        }
      });

      if (response.ok) {
        const data = await response.json();
        if (data.logs && Array.isArray(data.logs)) {
          // Update logs if changed
          if (data.logs.length !== lastLogCount) {
            logs = data.logs;
            lastLogCount = data.logs.length;
            scrollToBottom();
          }
        }
        
        // Update status from logs response
        if (data.status) {
          status = data.status;
        }
      }
    } catch (error) {
      console.error('Failed to fetch deployment logs:', error);
    }
  }

  // Poll for deployment status and progress every 3 seconds
  async function pollStatus() {
    if (!deploymentId || !isVisible) return;
    
    try {
      const response = await fetch(`${API_BASE_URL}/api/v1/projects/${projectId}/deployments/${deploymentId}/status`, {
        headers: {
          'Content-Type': 'application/json',
          ...auth.getAuthHeader()
        }
      });

      if (response.ok) {
        const data = await response.json();
        status = data.status || status;
        progress = data.progress || 0;
        deploymentStats = {
          build_time: data.build_time,
          deploy_time: data.deploy_time,
          file_count: data.file_count,
          total_size: data.total_size,
          deployed_at: data.deployed_at
        };
        
        // Keep polling after deployment completion for status updates
        // User can manually close console when done
      }
    } catch (error) {
      console.error('Failed to fetch deployment status:', error);
    }
  }

  // Start polling when component becomes visible
  function startPolling() {
    if (pollingInterval) return;
    
    isConnected = true;
    pollLogs(); // Initial fetch
    pollStatus(); // Initial status fetch
    pollingInterval = setInterval(pollLogs, 2000);
    statusInterval = setInterval(pollStatus, 3000);
  }

  // Stop polling
  function stopPolling() {
    if (pollingInterval) {
      clearInterval(pollingInterval);
      pollingInterval = null;
    }
    if (statusInterval) {
      clearInterval(statusInterval);
      statusInterval = null;
    }
    isConnected = false;
  }

  // Toggle collapsed state
  function toggleCollapse() {
    isCollapsed = !isCollapsed;
    if (!isCollapsed) {
      scrollToBottom();
    }
  }

  // Clear logs
  function clearLogs() {
    logs = [];
    lastLogCount = 0;
  }

  // Copy logs to clipboard
  async function copyLogs() {
    const logsText = logs.map(log => `[${log.timestamp}] [${log.level.toUpperCase()}] [${log.phase}] ${log.message}`).join('\n');
    try {
      await navigator.clipboard.writeText(logsText);
      // Could add toast notification here
    } catch (err) {
      console.error('Failed to copy logs:', err);
    }
  }

  // Get status color for progress bar
  function getStatusColor(status: string) {
    switch (status) {
      case 'pending': return 'bg-yellow-500';
      case 'building': return 'bg-blue-500';
      case 'deploying': return 'bg-purple-500';
      case 'deployed': return 'bg-green-500';
      case 'failed': return 'bg-red-500';
      default: return 'bg-gray-500';
    }
  }

  // Get log level color
  function getLogLevelColor(level: string) {
    switch (level) {
      case 'error': return 'text-red-400';
      case 'warn': return 'text-yellow-400';
      case 'info': return 'text-green-400';
      default: return 'text-gray-400';
    }
  }

  // Get phase badge color
  function getPhaseBadgeColor(phase: string) {
    switch (phase) {
      case 'build': return 'bg-blue-600';
      case 'deploy': return 'bg-purple-600';
      case 'error': return 'bg-red-600';
      case 'status': return 'bg-gray-600';
      default: return 'bg-gray-600';
    }
  }

  // Format file size
  function formatFileSize(bytes: number) {
    if (!bytes) return '0 B';
    const sizes = ['B', 'KB', 'MB', 'GB'];
    const i = Math.floor(Math.log(bytes) / Math.log(1024));
    return `${(bytes / Math.pow(1024, i)).toFixed(1)} ${sizes[i]}`;
  }

  // Format duration
  function formatDuration(ms: number) {
    if (!ms) return '0s';
    return `${(ms / 1000).toFixed(1)}s`;
  }

  // Watch for visibility changes
  $: if (isVisible && deploymentId) {
    startPolling();
  } else {
    stopPolling();
  }

  onDestroy(() => {
    stopPolling();
  });
</script>

{#if isVisible}
  <div class="deployment-console border rounded-lg bg-gray-900 text-green-400 font-mono text-sm">
    <!-- Console Header -->
    <div class="p-3 border-b border-gray-700 bg-gray-800">
      <div class="flex items-center justify-between mb-2">
        <div class="flex items-center gap-2">
          <Icon name="terminal" size={16} className="text-green-400" />
          <span class="font-medium text-gray-200">Deployment Console</span>
          <span class="px-2 py-1 text-xs font-medium rounded-full {status === 'deployed' ? 'bg-green-600' : status === 'failed' ? 'bg-red-600' : status === 'deploying' ? 'bg-purple-600' : status === 'building' ? 'bg-blue-600' : 'bg-yellow-600'} text-white">
            {status}
          </span>
          {#if isConnected}
            <div class="flex items-center gap-1 text-xs text-green-400">
              <div class="w-2 h-2 bg-green-400 rounded-full animate-pulse"></div>
              Live
            </div>
          {:else}
            <div class="flex items-center gap-1 text-xs text-gray-500">
              <div class="w-2 h-2 bg-gray-500 rounded-full"></div>
              Disconnected
            </div>
          {/if}
        </div>
        
        <div class="flex items-center gap-2 text-xs text-gray-400">
          {#if deploymentStats}
            {#if deploymentStats.build_time}
              <span>Build: {formatDuration(deploymentStats.build_time)}</span>
            {/if}
            {#if deploymentStats.deploy_time}
              <span>Deploy: {formatDuration(deploymentStats.deploy_time)}</span>
            {/if}
            {#if deploymentStats.file_count}
              <span>{deploymentStats.file_count} files</span>
            {/if}
            {#if deploymentStats.total_size}
              <span>{formatFileSize(deploymentStats.total_size)}</span>
            {/if}
          {/if}
        </div>
      </div>
      
      <!-- Progress Bar -->
      <div class="w-full bg-gray-700 rounded-full h-2 mb-3">
        <div class="h-2 rounded-full transition-all duration-500 {getStatusColor(status)}" style="width: {progress}%"></div>
      </div>
      
      <div class="flex items-center justify-between">
        <div class="text-xs text-gray-400">
          Progress: {progress}%
        </div>
        
        <div class="flex items-center gap-2">
          <Button
            size="sm"
            variant="ghost"
            on:click={clearLogs}
            title="Clear logs"
            class="text-gray-400 hover:text-gray-200 p-1"
          >
            <Icon name="trash" size={14} />
          </Button>
          
          <Button
            size="sm"
            variant="ghost"
            on:click={copyLogs}
            title="Copy logs"
            class="text-gray-400 hover:text-gray-200 p-1"
          >
            <Icon name="copy" size={14} />
          </Button>
          
          <Button
            size="sm"
            variant="ghost"
            on:click={toggleCollapse}
            title={isCollapsed ? 'Expand' : 'Collapse'}
            class="text-gray-400 hover:text-gray-200 p-1"
          >
            <Icon name={isCollapsed ? 'chevron-down' : 'chevron-up'} size={14} />
          </Button>
          
          {#if onClose}
            <Button
              size="sm"
              variant="ghost"
              on:click={onClose}
              title="Close console"
              class="text-gray-400 hover:text-gray-200 p-1"
            >
              <Icon name="x" size={14} />
            </Button>
          {/if}
        </div>
      </div>
    </div>

    <!-- Console Content -->
    {#if !isCollapsed}
      <div 
        bind:this={consoleElement}
        class="console-content h-64 overflow-y-auto p-4 bg-gray-900"
        style="scroll-behavior: smooth;"
      >
        {#if logs.length === 0}
          <div class="text-gray-500 italic">
            {isConnected ? 'Waiting for deployment logs...' : 'No logs available'}
          </div>
        {:else}
          {#each logs as log}
            <div class="console-line flex items-start gap-2 mb-1 leading-5">
              <span class="text-xs text-gray-500 font-mono shrink-0">{log.timestamp}</span>
              <span class="px-1.5 py-0.5 text-xs font-medium rounded {getPhaseBadgeColor(log.phase)} text-white shrink-0">
                {log.phase}
              </span>
              <span class="text-xs {getLogLevelColor(log.level)} font-mono shrink-0 uppercase">
                {log.level}
              </span>
              <span class="text-sm text-gray-300 font-mono break-all">
                {log.message}
              </span>
            </div>
          {/each}
        {/if}
      </div>
    {/if}
  </div>
{/if}

<style>
  .console-content::-webkit-scrollbar {
    width: 6px;
  }
  
  .console-content::-webkit-scrollbar-track {
    background: #374151;
  }
  
  .console-content::-webkit-scrollbar-thumb {
    background: #6b7280;
    border-radius: 3px;
  }
  
  .console-content::-webkit-scrollbar-thumb:hover {
    background: #9ca3af;
  }

  .console-line {
    font-family: 'Monaco', 'Menlo', 'Ubuntu Mono', monospace;
    font-size: 12px;
    line-height: 1.4;
  }
</style>