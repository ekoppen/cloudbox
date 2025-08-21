<script lang="ts">
  import { onMount } from 'svelte';
  import { page } from '$app/stores';
  import Button from '$lib/components/ui/button.svelte';
  import Card from '$lib/components/ui/card.svelte';
  import Icon from '$lib/components/ui/icon.svelte';
  import { addToast } from '$lib/stores/toast';
  import { API_BASE_URL } from '$lib/config';
  import { auth } from '$lib/stores/auth';

  // Get project ID from URL params
  $: projectId = $page.params.id;

  // State
  let project = null;
  let scripts = [];
  let selectedScripts = [];
  let executionLogs = [];
  let isRunning = false;
  let showEditor = false;
  let showTemplates = false;
  let editingScript = null;
  let filterType = '';
  let filterCategory = '';

  // Templates
  let availableTemplates = [
    {
      name: 'Basic Web Application',
      description: 'Essential database schema voor web applicaties',
      category: 'webapp',
      framework: 'universal'
    },
    {
      name: 'AI Chat Application', 
      description: 'Complete setup voor AI chat applicaties (zoals Aimy)',
      category: 'ai-app',
      framework: 'universal'
    },
    {
      name: 'E-commerce Backend',
      description: 'Database schema voor e-commerce platforms',
      category: 'ecommerce',
      framework: 'universal'
    }
  ];

  // Computed
  $: filteredScripts = scripts.filter(script => {
    const typeMatch = !filterType || script.type === filterType;
    const categoryMatch = !filterCategory || script.category === filterCategory;
    return typeMatch && categoryMatch;
  }).sort((a, b) => (a.run_order || 999) - (b.run_order || 999));

  // Methods
  async function loadProject() {
    if (!projectId) return;
    
    try {
      const response = await fetch(`${API_BASE_URL}/api/v1/projects/${projectId}`, {
        credentials: 'include',
        headers: {
          'Authorization': `Bearer ${$auth.token}`
        }
      });
      if (response.ok) {
        const data = await response.json();
        if (data.success) {
          project = data.project;
        }
      } else {
        throw new Error('Failed to load project');
      }
    } catch (error) {
      console.error('Failed to load project:', error);
      addToast('Failed to load project', 'error');
    }
  }

  async function loadProjectScripts() {
    if (!projectId) return;
    
    try {
      const response = await fetch(`${API_BASE_URL}/api/v1/plugins/script-runner/scripts/${projectId}`, {
        credentials: 'include',
        headers: {
          'Authorization': `Bearer ${$auth.token}`
        }
      });
      if (response.ok) {
        const data = await response.json();
        if (data.success) {
          scripts = data.scripts || [];
        }
      }
    } catch (error) {
      console.error('Failed to load scripts:', error);
      // Don't show error toast if plugin is not installed yet
      if (!error.message.includes('404')) {
        addToast('Failed to load scripts', 'error');
      }
    }
  }

  async function runScript(script) {
    isRunning = true;
    
    try {
      const response = await fetch(`${API_BASE_URL}/api/v1/plugins/script-runner/execute/${projectId}/${script.id}`, {
        method: 'POST',
        headers: { 
          'Content-Type': 'application/json',
          'Authorization': `Bearer ${$auth.token}`
        },
        credentials: 'include',
        body: JSON.stringify({})
      });
      
      if (response.ok) {
        const data = await response.json();
        
        if (data.success) {
          // Update script status
          const scriptIndex = scripts.findIndex(s => s.id === script.id);
          if (scriptIndex !== -1) {
            scripts[scriptIndex].last_status = data.result.success ? 'success' : 'failed';
            scripts[scriptIndex].last_execution = new Date().toISOString();
            scripts = [...scripts]; // Trigger reactivity
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
      } else {
        throw new Error(`HTTP ${response.status}`);
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

  async function applyTemplate(templateName) {
    try {
      const response = await fetch(`${API_BASE_URL}/api/v1/plugins/script-runner/setup-project/${projectId}/${templateName}`, {
        method: 'POST',
        credentials: 'include',
        headers: {
          'Authorization': `Bearer ${$auth.token}`
        }
      });
      
      if (response.ok) {
        const data = await response.json();
        if (data.success) {
          addToast(`Template "${templateName}" applied successfully`, 'success');
          loadProjectScripts(); // Reload scripts
          showTemplates = false;
        } else {
          addToast(`Failed to apply template: ${data.error}`, 'error');
        }
      } else {
        throw new Error(`HTTP ${response.status}`);
      }
    } catch (error) {
      console.error('Template application error:', error);
      addToast('Failed to apply template', 'error');
    }
  }

  function getScriptIcon(type) {
    switch (type) {
      case 'sql': return 'database';
      case 'javascript': return 'code';
      case 'setup': return 'cog';
      case 'migration': return 'arrow-path';
      default: return 'document';
    }
  }

  function getStatusIcon(script) {
    switch (script.last_status) {
      case 'success': return 'check-circle';
      case 'failed': return 'x-circle';
      default: return 'clock';
    }
  }

  function getStatusColor(script) {
    switch (script.last_status) {
      case 'success': return 'text-green-500';
      case 'failed': return 'text-red-500';
      default: return 'text-yellow-500';
    }
  }

  function getTypeClass(type) {
    switch (type) {
      case 'sql': return 'bg-blue-100 text-blue-800 dark:bg-gray-800/20 dark:text-blue-300';
      case 'javascript': return 'bg-yellow-100 text-yellow-800 dark:bg-yellow-900/20 dark:text-yellow-300';
      case 'setup': return 'bg-green-100 text-green-800 dark:bg-green-900/20 dark:text-green-300';
      case 'migration': return 'bg-purple-100 text-purple-800 dark:bg-purple-900/20 dark:text-purple-300';
      default: return 'bg-gray-100 text-gray-800 dark:bg-gray-900/20 dark:text-gray-300';
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
    loadProject();
    loadProjectScripts();
  });

  // Reactive statements - reload when project ID changes
  $: if (projectId) {
    loadProject();
    loadProjectScripts();
  }
</script>

<svelte:head>
  <title>Database Scripts - {project?.name || 'Project'} - CloudBox</title>
</svelte:head>

<div class="space-y-6">
  <!-- Header -->
  <div class="flex items-center justify-between">
    <div>
      <h1 class="text-2xl font-bold text-foreground">Database Scripts</h1>
      <p class="text-muted-foreground mt-1">
        Setup scripts voor {project?.name || 'dit project'} - SQL, functions en automation
      </p>
    </div>
    
    <div class="flex gap-3">
      <Button 
        on:click={() => showEditor = true}
        variant="default"
      >
        <Icon name="plus" size={16} className="mr-2" />
        Nieuw Script
      </Button>
      
      <Button 
        on:click={() => showTemplates = true}
        variant="outline"
      >
        <Icon name="template" size={16} className="mr-2" />
        Templates
      </Button>
    </div>
  </div>

  <!-- Quick Actions -->
  <div class="grid grid-cols-1 md:grid-cols-3 gap-4">
    <Card className="p-4 border-blue-200 bg-blue-50 dark:bg-gray-900/20 dark:border-gray-600">
      <div class="flex items-center">
        <Icon name="database" size={32} className="text-blue-500 mr-3" />
        <div>
          <h3 class="font-semibold text-blue-900 dark:text-blue-100">Database Setup</h3>
          <p class="text-sm text-blue-700 dark:text-blue-300">SQL schemas en migrations</p>
        </div>
      </div>
      <Button 
        on:click={() => showTemplates = true}
        className="mt-3 w-full bg-blue-600 hover:bg-blue-700"
      >
        Database Templates
      </Button>
    </Card>
    
    <Card className="p-4 border-green-200 bg-green-50 dark:bg-green-950/20 dark:border-green-800">
      <div class="flex items-center">
        <Icon name="code" size={32} className="text-green-500 mr-3" />
        <div>
          <h3 class="font-semibold text-green-900 dark:text-green-100">Functions</h3>
          <p class="text-sm text-green-700 dark:text-green-300">Deploy CloudBox functions</p>
        </div>
      </div>
      <Button 
        on:click={() => showTemplates = true}
        className="mt-3 w-full bg-green-600 hover:bg-green-700"
      >
        Function Templates
      </Button>
    </Card>
    
    <Card className="p-4 border-purple-200 bg-purple-50 dark:bg-purple-950/20 dark:border-purple-800">
      <div class="flex items-center">
        <Icon name="chart-bar" size={32} className="text-purple-500 mr-3" />
        <div>
          <h3 class="font-semibold text-purple-900 dark:text-purple-100">Analytics</h3>
          <p class="text-sm text-purple-700 dark:text-purple-300">Script execution stats</p>
        </div>
      </div>
      <Button 
        on:click={() => {}}
        className="mt-3 w-full bg-purple-600 hover:bg-purple-700"
      >
        View Analytics
      </Button>
    </Card>
  </div>

  <!-- Scripts List -->
  <Card>
    <div class="p-6 border-b border-border">
      <div class="flex items-center justify-between mb-4">
        <h2 class="text-lg font-semibold text-foreground">Project Scripts</h2>
        
        <div class="flex gap-2">
          <Button 
            on:click={runAllScripts}
            disabled={isRunning || scripts.length === 0}
            variant="default"
          >
            <Icon name="play" size={16} className="mr-2" />
            Run All Scripts
          </Button>
          
          {#if selectedScripts.length > 0}
            <Button 
              on:click={runSelectedScripts}
              disabled={isRunning}
              className="bg-green-600 hover:bg-green-700"
            >
              <Icon name="play" size={16} className="mr-2" />
              Run Selected ({selectedScripts.length})
            </Button>
          {/if}
        </div>
      </div>
      
      <!-- Filters -->
      <div class="flex gap-4">
        <select bind:value={filterType} class="form-select text-sm border border-border rounded bg-background text-foreground">
          <option value="">All Types</option>
          <option value="sql">SQL Scripts</option>
          <option value="javascript">JavaScript</option>
          <option value="setup">Setup Scripts</option>
          <option value="migration">Migrations</option>
        </select>
        
        <select bind:value={filterCategory} class="form-select text-sm border border-border rounded bg-background text-foreground">
          <option value="">All Categories</option>
          <option value="project-setup">Project Setup</option>
          <option value="custom">Custom Scripts</option>
          <option value="template">Templates</option>
          <option value="migration">Migrations</option>
        </select>
      </div>
    </div>

    <!-- Scripts Table -->
    {#if scripts.length > 0}
      <div class="overflow-x-auto">
        <table class="min-w-full divide-y divide-border">
          <thead class="bg-muted/50">
            <tr>
              <th class="px-6 py-3 text-left text-xs font-medium text-muted-foreground uppercase tracking-wider">
                <input 
                  type="checkbox" 
                  on:change={toggleSelectAll}
                  checked={selectedScripts.length === filteredScripts.length && filteredScripts.length > 0}
                  class="rounded border-border text-primary"
                />
              </th>
              <th class="px-6 py-3 text-left text-xs font-medium text-muted-foreground uppercase tracking-wider">
                Script
              </th>
              <th class="px-6 py-3 text-left text-xs font-medium text-muted-foreground uppercase tracking-wider">
                Type
              </th>
              <th class="px-6 py-3 text-left text-xs font-medium text-muted-foreground uppercase tracking-wider">
                Status
              </th>
              <th class="px-6 py-3 text-left text-xs font-medium text-muted-foreground uppercase tracking-wider">
                Last Run
              </th>
              <th class="px-6 py-3 text-left text-xs font-medium text-muted-foreground uppercase tracking-wider">
                Actions
              </th>
            </tr>
          </thead>
          <tbody class="bg-background divide-y divide-border">
            {#each filteredScripts as script (script.id)}
              <tr class="hover:bg-muted/50">
                <td class="px-6 py-4 whitespace-nowrap">
                  <input 
                    type="checkbox" 
                    checked={selectedScripts.includes(script.id)}
                    on:change={() => toggleScriptSelection(script.id)}
                    class="rounded border-border text-primary"
                  />
                </td>
                
                <td class="px-6 py-4">
                  <div class="flex items-center">
                    <div class="flex-shrink-0 mr-3">
                      <Icon name={getScriptIcon(script.type)} size={20} />
                    </div>
                    <div>
                      <div class="text-sm font-medium text-foreground">{script.name}</div>
                      <div class="text-sm text-muted-foreground">{script.description}</div>
                      {#if script.dependencies && script.dependencies.length > 0}
                        <div class="text-xs text-muted-foreground mt-1">
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
                    <Icon name={getStatusIcon(script)} size={16} className="{getStatusColor(script)} mr-2" />
                    <span class="text-sm text-foreground">
                      {script.last_status || 'never_run'}
                    </span>
                  </div>
                </td>
                
                <td class="px-6 py-4 whitespace-nowrap text-sm text-muted-foreground">
                  {#if script.last_execution}
                    {formatDate(script.last_execution)}
                  {:else}
                    <span class="text-muted-foreground">Never run</span>
                  {/if}
                </td>
                
                <td class="px-6 py-4 whitespace-nowrap text-sm font-medium">
                  <div class="flex gap-2">
                    <button 
                      on:click={() => runScript(script)}
                      disabled={isRunning}
                      class="text-primary hover:text-primary/80 disabled:opacity-50"
                    >
                      Run
                    </button>
                    <button 
                      on:click={() => { editingScript = script; showEditor = true; }}
                      class="text-primary hover:text-primary/80"
                    >
                      Edit
                    </button>
                    <button 
                      on:click={() => {}}
                      class="text-muted-foreground hover:text-foreground"
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
    {:else}
      <div class="p-12 text-center">
        <Icon name="database" size={48} className="mx-auto text-muted-foreground mb-4" />
        <h3 class="text-lg font-medium text-foreground mb-2">Geen scripts gevonden</h3>
        <p class="text-muted-foreground mb-6">Begin met een template of maak je eerste script aan.</p>
        <div class="flex gap-3 justify-center">
          <Button on:click={() => showTemplates = true}>
            <Icon name="template" size={16} className="mr-2" />
            Browse Templates
          </Button>
          <Button on:click={() => showEditor = true} variant="outline">
            <Icon name="plus" size={16} className="mr-2" />
            Nieuw Script
          </Button>
        </div>
      </div>
    {/if}
  </Card>

  <!-- Execution Logs -->
  {#if executionLogs.length > 0}
    <Card>
      <div class="p-6 border-b border-border">
        <h2 class="text-lg font-semibold text-foreground">Recent Executions</h2>
      </div>
      
      <div class="max-h-64 overflow-y-auto">
        {#each executionLogs as log (log.id)}
          <div class="p-4 border-b border-border last:border-b-0">
            <div class="flex items-center justify-between">
              <div class="flex items-center">
                <Icon name={log.status === 'success' ? 'check-circle' : 'x-circle'} 
                      size={16} 
                      className="{log.status === 'success' ? 'text-green-500' : 'text-red-500'} mr-2" />
                <span class="font-medium text-foreground">{log.script_name}</span>
                <span class="ml-2 text-sm text-muted-foreground">({log.duration_ms}ms)</span>
              </div>
              <span class="text-xs text-muted-foreground">{formatDate(log.started_at)}</span>
            </div>
            
            {#if log.output}
              <div class="mt-2 text-sm bg-muted p-2 rounded font-mono">
                {log.output}
              </div>
            {/if}
            
            {#if log.error_message}
              <div class="mt-2 text-sm bg-destructive/10 text-destructive p-2 rounded">
                Error: {log.error_message}
              </div>
            {/if}
          </div>
        {/each}
      </div>
    </Card>
  {/if}

  <!-- Templates Modal -->
  {#if showTemplates}
    <!-- svelte-ignore a11y-click-events-have-key-events -->
    <!-- svelte-ignore a11y-no-noninteractive-element-interactions -->
    <div class="fixed inset-0 bg-black/50 flex items-center justify-center z-50" 
         role="dialog" 
         aria-modal="true" 
         aria-labelledby="templates-modal-title"
         tabindex="-1"
         on:click={() => showTemplates = false}
         on:keydown={(e) => e.key === 'Escape' && (showTemplates = false)}>
      <!-- svelte-ignore a11y-click-events-have-key-events -->
      <!-- svelte-ignore a11y-no-noninteractive-element-interactions -->
      <div class="max-w-4xl w-full mx-4 max-h-[80vh] overflow-y-auto bg-card border border-border rounded-lg shadow-lg" 
           role="document"
           on:click|stopPropagation
           on:keydown|stopPropagation>
        <div class="p-6 border-b border-border">
          <h2 id="templates-modal-title" class="text-xl font-semibold text-foreground">Project Templates</h2>
          <p class="text-muted-foreground mt-1">Kies een template om {project?.name || 'dit project'} mee op te zetten</p>
        </div>
        
        <div class="p-6">
          <div class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
            {#each availableTemplates as template}
              <Card className="p-4 border-2 border-dashed border-border hover:border-primary cursor-pointer transition-colors">
                <div class="text-center">
                  <Icon name="template" size={32} className="mx-auto text-primary mb-2" />
                  <h3 class="font-semibold text-foreground">{template.name}</h3>
                  <p class="text-sm text-muted-foreground mt-1">{template.description}</p>
                  <div class="mt-3">
                    <span class="inline-flex items-center px-2 py-1 rounded-full text-xs bg-muted text-muted-foreground">
                      {template.category}
                    </span>
                  </div>
                  <Button 
                    className="mt-4 w-full" 
                    variant="outline"
                    on:click={() => applyTemplate(template.name)}
                  >
                    Use Template
                  </Button>
                </div>
              </Card>
            {/each}
          </div>
        </div>
        
        <div class="p-6 border-t border-border flex justify-end">
          <Button on:click={() => showTemplates = false} variant="outline">
            Close
          </Button>
        </div>
      </div>
    </div>
  {/if}

  <!-- Info Box -->
  <Card className="p-4 bg-blue-50 border-blue-200 dark:bg-gray-900/20 dark:border-gray-600">
    <h3 class="font-semibold text-blue-900 dark:text-blue-100 mb-2">💡 Database Scripts voor {project?.name || 'dit project'}:</h3>
    <ul class="text-sm text-blue-800 dark:text-blue-200 space-y-1">
      <li>• <strong>SQL Scripts:</strong> Database schemas en migrations (zoals Supabase SQL editor)</li>
      <li>• <strong>JavaScript Scripts:</strong> CloudBox functions voor dit project</li>
      <li>• <strong>Setup Scripts:</strong> Project configuratie en deployment automation</li>
      <li>• <strong>Dependencies:</strong> Scripts worden automatisch in de juiste volgorde uitgevoerd</li>
      <li>• <strong>Templates:</strong> Pre-built setups voor verschillende project types</li>
    </ul>
  </Card>
</div>

<style>
  .form-select {
    @apply p-2;
  }
</style>