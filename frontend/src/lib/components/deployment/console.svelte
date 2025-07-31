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

  let logs: string[] = [];
  let isCollapsed: boolean = false;
  let isConnected: boolean = false;
  let consoleElement: HTMLElement;
  let pollingInterval: number | null = null;
  let lastLogId: string | null = null;

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
          // Simple approach: replace all logs (in future we can optimize with incremental updates)
          const newLogs = data.logs.map((log: any) => {
            const timestamp = new Date(log.created_at).toLocaleTimeString();
            const level = log.level || 'info';
            const message = log.message || '';
            return `[${timestamp}] [${level.toUpperCase()}] ${message}`;
          });
          
          if (newLogs.length !== logs.length) {
            logs = newLogs;
            scrollToBottom();
          }
        }
      }
    } catch (error) {
      console.error('Failed to fetch deployment logs:', error);
    }
  }

  // Start polling when component becomes visible
  function startPolling() {
    if (pollingInterval) return;
    
    isConnected = true;
    pollLogs(); // Initial fetch
    pollingInterval = setInterval(pollLogs, 2000);
  }

  // Stop polling
  function stopPolling() {
    if (pollingInterval) {
      clearInterval(pollingInterval);
      pollingInterval = null;
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
  }

  // Copy logs to clipboard
  async function copyLogs() {
    const logsText = logs.join('\n');
    try {
      await navigator.clipboard.writeText(logsText);
      // Could add toast notification here
    } catch (err) {
      console.error('Failed to copy logs:', err);
    }
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
    <div class="flex items-center justify-between p-3 border-b border-gray-700 bg-gray-800">
      <div class="flex items-center gap-2">
        <Icon name="terminal" size={16} className="text-green-400" />
        <span class="font-medium text-gray-200">Deployment Console</span>
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
            <div class="console-line whitespace-pre-wrap mb-1 leading-5">
              {log}
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