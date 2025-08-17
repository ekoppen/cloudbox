<script>
  import { onMount } from 'svelte';
  import { page } from '$app/stores';
  import Button from '$lib/components/ui/button.svelte';
  import Card from '$lib/components/ui/card.svelte';
  import Modal from '$lib/components/ui/modal.svelte';
  import { addToast } from '$lib/stores/toast';

  // State
  let projects = [];
  let selectedProject = '';
  let scripts = [];
  let selectedScripts = [];
  let executionLogs = [];
  let isRunning = false;
  let showEditor = false;
  let showTemplates = false;
  let editingScript = null;
  let filterType = '';
  let filterCategory = '';

  // Computed
  $: filteredScripts = scripts.filter(script => {
    const typeMatch = !filterType || script.type === filterType;
    const categoryMatch = !filterCategory || script.category === filterCategory;
    return typeMatch && categoryMatch;
  }).sort((a, b) => (a.run_order || 999) - (b.run_order || 999));

  // Methods
  async function loadProjects() {
    try {
      const response = await fetch('/api/projects');
      const data = await response.json();
      if (data.success) {
        projects = data.projects;
      }
    } catch (error) {
      console.error('Failed to load projects:', error);
      addToast('Failed to load projects', 'error');
    }
  }

  async function loadProjectScripts() {
    if (!selectedProject) return;
    
    try {
      const response = await fetch(`/api/plugins/script-runner/scripts/${selectedProject}`);
      const data = await response.json();
      if (data.success) {
        scripts = data.scripts;
      }
    } catch (error) {
      console.error('Failed to load scripts:', error);
      addToast('Failed to load scripts', 'error');
    }
  }

  async function runScript(script) {
    isRunning = true;
    
    try {
      const response = await fetch(`/api/plugins/script-runner/execute/${selectedProject}/${script.id}`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({})
      });
      
      const data = await response.json();
      
      if (data.success) {
        // Update script status
        const scriptIndex = scripts.findIndex(s => s.id === script.id);
        if (scriptIndex !== -1) {
          scripts[scriptIndex].last_status = data.result.success ? 'success' : 'failed';
          scripts[scriptIndex].last_execution = new Date().toISOString();
        }
        
        // Add to execution logs
        executionLogs = [{
          id: data.execution_id,
          script_name: script.name,
          status: data.result.success ? 'success' : 'failed',
          output: data.result.output,
          error_message: data.result.error,
          duration_ms: data.result.duration,
          started_at: new Date().toISOString()
        }, ...executionLogs];
        
        addToast(`Script "${script.name}" executed successfully`, 'success');
      } else {
        addToast(`Script execution failed: ${data.error}`, 'error');
      }
    } catch (error) {
      console.error('Script execution error:', error);
      addToast('Failed to execute script', 'error');
    }
    
    isRunning = false;
  }

  async function runAllScripts() {
    isRunning = true;
    
    const sortedScripts = filteredScripts.sort((a, b) => (a.run_order || 999) - (b.run_order || 999));
    
    for (const script of sortedScripts) {
      await runScript(script);
    }
    
    isRunning = false;
  }

  async function runSelectedScripts() {
    isRunning = true;
    
    const sortedSelected = scripts
      .filter(s => selectedScripts.includes(s.id))
      .sort((a, b) => (a.run_order || 999) - (b.run_order || 999));
    
    for (const script of sortedSelected) {
      await runScript(script);
    }
    
    isRunning = false;
  }

  function getScriptIcon(type) {
    switch (type) {
      case 'sql': return '🗄️';
      case 'javascript': return '⚡';
      case 'setup': return '⚙️';
      case 'migration': return '🔄';
      default: return '📄';
    }
  }

  function getStatusIcon(script) {
    switch (script.last_status) {
      case 'success': return '✅';
      case 'failed': return '❌';
      default: return '⏳';
    }
  }

  function getTypeClass(type) {
    switch (type) {
      case 'sql': return 'bg-blue-100 text-blue-800';
      case 'javascript': return 'bg-yellow-100 text-yellow-800';
      case 'setup': return 'bg-green-100 text-green-800';
      case 'migration': return 'bg-purple-100 text-purple-800';
      default: return 'bg-gray-100 text-gray-800';
    }
  }

  function toggleScriptSelection(scriptId) {
    if (selectedScripts.includes(scriptId)) {
      selectedScripts = selectedScripts.filter(id => id !== scriptId);
    } else {
      selectedScripts = [...selectedScripts, scriptId];
    }
  }

  function toggleSelectAll() {
    if (selectedScripts.length === filteredScripts.length) {
      selectedScripts = [];
    } else {
      selectedScripts = filteredScripts.map(s => s.id);
    }
  }

  function formatDate(dateString) {
    return new Date(dateString).toLocaleString('nl-NL');
  }

  onMount(() => {
    loadProjects();
  });

  // Reactive statements
  $: if (selectedProject) {
    loadProjectScripts();
  }
</script>

<!-- Header -->
<div class="mb-8">
  <div class="flex items-center justify-between">
    <div>
      <h1 class="text-3xl font-bold text-gray-900">Script Runner</h1>
      <p class="text-gray-600 mt-2">
        Database scripts en project setup - Universeel voor alle CloudBox projecten (zoals Supabase)
      </p>
    </div>
    
    <div class="flex gap-3">
      <Button 
        on:click={() => showEditor = true}
        variant="primary"
      >
        ➕ Nieuw Script
      </Button>
      
      <Button 
        on:click={() => showTemplates = true}
        variant="secondary"
      >
        📋 Templates
      </Button>
    </div>
  </div>
</div>

<!-- Project Selector -->
<Card class="mb-6">
  <div class="p-6">
    <label class="block text-sm font-medium text-gray-700 mb-2">
      Project:
    </label>
    <select 
      bind:value={selectedProject} 
      class="w-full p-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500"
    >
      <option value="">Selecteer een project</option>
      {#each projects as project}
        <option value={project.id}>{project.name}</option>
      {/each}
    </select>
  </div>
</Card>

<!-- Quick Actions -->
{#if selectedProject}
  <div class="grid grid-cols-1 md:grid-cols-3 gap-4 mb-6">
    <Card class="bg-blue-50 border-blue-200">
      <div class="p-4">
        <div class="flex items-center">
          <div class="text-2xl mr-3">🗄️</div>
          <div>
            <h3 class="font-semibold text-blue-900">Database Setup</h3>
            <p class="text-sm text-blue-700">SQL schemas en migrations</p>
          </div>
        </div>
        <Button 
          on:click={() => showTemplates = true}
          class="mt-3 w-full bg-blue-600 hover:bg-blue-700"
        >
          Database Templates
        </Button>
      </div>
    </Card>
    
    <Card class="bg-green-50 border-green-200">
      <div class="p-4">
        <div class="flex items-center">
          <div class="text-2xl mr-3">⚡</div>
          <div>
            <h3 class="font-semibold text-green-900">Functions</h3>
            <p class="text-sm text-green-700">Deploy CloudBox functions</p>
          </div>
        </div>
        <Button 
          on:click={() => showTemplates = true}
          class="mt-3 w-full bg-green-600 hover:bg-green-700"
        >
          Function Templates
        </Button>
      </div>
    </Card>
    
    <Card class="bg-purple-50 border-purple-200">
      <div class="p-4">
        <div class="flex items-center">
          <div class="text-2xl mr-3">📊</div>
          <div>
            <h3 class="font-semibold text-purple-900">Analytics</h3>
            <p class="text-sm text-purple-700">Script execution stats</p>
          </div>
        </div>
        <Button 
          on:click={() => {}}
          class="mt-3 w-full bg-purple-600 hover:bg-purple-700"
        >
          View Analytics
        </Button>
      </div>
    </Card>
  </div>
{/if}

<!-- Scripts List -->
{#if selectedProject}
  <Card>
    <div class="p-6 border-b border-gray-200">
      <div class="flex items-center justify-between mb-4">
        <h2 class="text-lg font-semibold text-gray-900">Project Scripts</h2>
        
        <div class="flex gap-2">
          <Button 
            on:click={runAllScripts}
            disabled={isRunning || scripts.length === 0}
            variant="primary"
          >
            ▶️ Run All Scripts
          </Button>
          
          {#if selectedScripts.length > 0}
            <Button 
              on:click={runSelectedScripts}
              disabled={isRunning}
              class="bg-green-600 hover:bg-green-700"
            >
              ▶️ Run Selected ({selectedScripts.length})
            </Button>
          {/if}
        </div>
      </div>
      
      <!-- Filters -->
      <div class="flex gap-4">
        <select bind:value={filterType} class="form-select text-sm">
          <option value="">All Types</option>
          <option value="sql">SQL Scripts</option>
          <option value="javascript">JavaScript</option>
          <option value="setup">Setup Scripts</option>
          <option value="migration">Migrations</option>
        </select>
        
        <select bind:value={filterCategory} class="form-select text-sm">
          <option value="">All Categories</option>
          <option value="project-setup">Project Setup</option>
          <option value="custom">Custom Scripts</option>
          <option value="template">Templates</option>
          <option value="migration">Migrations</option>
        </select>
      </div>
    </div>

    <!-- Scripts Table -->
    <div class="overflow-x-auto">
      <table class="min-w-full divide-y divide-gray-200">
        <thead class="bg-gray-50">
          <tr>
            <th class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
              <input 
                type="checkbox" 
                on:change={toggleSelectAll}
                checked={selectedScripts.length === filteredScripts.length && filteredScripts.length > 0}
                class="rounded border-gray-300 text-blue-600"
              />
            </th>
            <th class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
              Script
            </th>
            <th class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
              Type
            </th>
            <th class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
              Status
            </th>
            <th class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
              Last Run
            </th>
            <th class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
              Actions
            </th>
          </tr>
        </thead>
        <tbody class="bg-white divide-y divide-gray-200">
          {#each filteredScripts as script (script.id)}
            <tr class="hover:bg-gray-50">
              <td class="px-6 py-4 whitespace-nowrap">
                <input 
                  type="checkbox" 
                  checked={selectedScripts.includes(script.id)}
                  on:change={() => toggleScriptSelection(script.id)}
                  class="rounded border-gray-300 text-blue-600"
                />
              </td>
              
              <td class="px-6 py-4">
                <div class="flex items-center">
                  <div class="flex-shrink-0 mr-3 text-xl">
                    {getScriptIcon(script.type)}
                  </div>
                  <div>
                    <div class="text-sm font-medium text-gray-900">{script.name}</div>
                    <div class="text-sm text-gray-500">{script.description}</div>
                    {#if script.dependencies && script.dependencies.length > 0}
                      <div class="text-xs text-gray-400 mt-1">
                        Depends on: {script.dependencies.join(', ')}
                      </div>
                    {/if}
                  </div>
                </div>
              </td>
              
              <td class="px-6 py-4 whitespace-nowrap">
                <span class="inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-medium {getTypeClass(script.type)}">
                  {script.type}
                </span>
              </td>
              
              <td class="px-6 py-4 whitespace-nowrap">
                <div class="flex items-center">
                  <span class="mr-2">{getStatusIcon(script)}</span>
                  <span class="text-sm text-gray-900">
                    {script.last_status || 'never_run'}
                  </span>
                </div>
              </td>
              
              <td class="px-6 py-4 whitespace-nowrap text-sm text-gray-500">
                {#if script.last_execution}
                  {formatDate(script.last_execution)}
                {:else}
                  <span class="text-gray-400">Never run</span>
                {/if}
              </td>
              
              <td class="px-6 py-4 whitespace-nowrap text-sm font-medium">
                <div class="flex gap-2">
                  <button 
                    on:click={() => runScript(script)}
                    disabled={isRunning}
                    class="text-indigo-600 hover:text-indigo-900 disabled:opacity-50"
                  >
                    Run
                  </button>
                  <button 
                    on:click={() => { editingScript = script; showEditor = true; }}
                    class="text-blue-600 hover:text-blue-900"
                  >
                    Edit
                  </button>
                  <button 
                    on:click={() => {}}
                    class="text-gray-600 hover:text-gray-900"
                  >
                    View
                  </button>
                </div>
              </td>
            </tr>
          {/each}
        </tbody>
      </table>
    </div>
  </Card>
{/if}

<!-- Execution Logs -->
{#if selectedProject && executionLogs.length > 0}
  <Card class="mt-6">
    <div class="p-6 border-b border-gray-200">
      <h2 class="text-lg font-semibold text-gray-900">Recent Executions</h2>
    </div>
    
    <div class="max-h-64 overflow-y-auto">
      {#each executionLogs as log (log.id)}
        <div class="p-4 border-b border-gray-100 last:border-b-0">
          <div class="flex items-center justify-between">
            <div class="flex items-center">
              <span class="mr-2">{log.status === 'success' ? '✅' : '❌'}</span>
              <span class="font-medium">{log.script_name}</span>
              <span class="ml-2 text-sm text-gray-500">({log.duration_ms}ms)</span>
            </div>
            <span class="text-xs text-gray-400">{formatDate(log.started_at)}</span>
          </div>
          
          {#if log.output}
            <div class="mt-2 text-sm bg-gray-50 p-2 rounded font-mono">
              {log.output}
            </div>
          {/if}
          
          {#if log.error_message}
            <div class="mt-2 text-sm bg-red-50 text-red-700 p-2 rounded">
              Error: {log.error_message}
            </div>
          {/if}
        </div>
      {/each}
    </div>
  </Card>
{/if}

<!-- Info Box -->
<Card class="mt-6 bg-blue-50 border-blue-200">
  <div class="p-4">
    <h3 class="font-semibold text-blue-900 mb-2">💡 Hoe het werkt:</h3>
    <ul class="text-sm text-blue-800 space-y-1">
      <li>• <strong>SQL Scripts:</strong> Database schemas en migrations (zoals Supabase SQL editor)</li>
      <li>• <strong>JavaScript Scripts:</strong> Serverless functions en automation</li>
      <li>• <strong>Setup Scripts:</strong> Project configuratie en deployment</li>
      <li>• <strong>Dependencies:</strong> Scripts worden automatisch in de juiste volgorde uitgevoerd</li>
      <li>• <strong>Templates:</strong> Pre-built scripts voor verschillende project types</li>
      <li>• <strong>Universeel:</strong> Werkt met alle CloudBox projecten en frameworks</li>
    </ul>
  </div>
</Card>

<style>
  .form-select {
    @apply block border-gray-300 rounded-md shadow-sm focus:ring-blue-500 focus:border-blue-500 p-2;
  }
</style>