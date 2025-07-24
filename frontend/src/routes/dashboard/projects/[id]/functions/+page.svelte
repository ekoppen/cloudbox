<script lang="ts">
  import { page } from '$app/stores';
  import { onMount } from 'svelte';
  import { API_BASE_URL, createApiRequest } from '$lib/config';
  import { auth } from '$lib/stores/auth';
  import { toast } from '$lib/stores/toast';
  import Card from '$lib/components/ui/card.svelte';
  import Button from '$lib/components/ui/button.svelte';
  import Modal from '$lib/components/ui/modal.svelte';
  import Input from '$lib/components/ui/input.svelte';
  import { API_BASE_URL, createApiRequest } from '$lib/config';
  import Label from '$lib/components/ui/label.svelte';
  import Select from '$lib/components/ui/select.svelte';
  import Textarea from '$lib/components/ui/textarea.svelte';

  let projectId = $page.params.id;
  let functions = [];
  let loading = true;
  let showCreateModal = false;
  let showCodeModal = false;
  let showLogsModal = false;
  let selectedFunction = null;
  let selectedLogs = [];
  let deployingFunctions = {};
  let testingFunctions = {};

  // Form data for new function
  let functionForm = {
    name: '',
    description: '',
    runtime: 'nodejs18',
    language: 'javascript',
    code: getDefaultCode('javascript'),
    entry_point: 'index.handler',
    timeout: 30,
    memory: 128,
    environment: {},
    is_public: false
  };

  function getDefaultCode(language: string) {
    const templates = {
      javascript: `// CloudBox Function - JavaScript
exports.handler = async (event, context) => {
  console.log('Function invoked with event:', event);
  
  return {
    statusCode: 200,
    body: {
      message: 'Hello from CloudBox Function!',
      timestamp: new Date().toISOString(),
      input: event
    }
  };
};`,
      python: `# CloudBox Function - Python
def handler(event, context):
    print(f'Function invoked with event: {event}')
    
    return {
        'statusCode': 200,
        'body': {
            'message': 'Hello from CloudBox Function!',
            'timestamp': datetime.now().isoformat(),
            'input': event
        }
    }`,
      go: `package main

import (
    "context"
    "encoding/json"
    "fmt"
    "log"
    "time"
)

type Event map[string]interface{}
type Response struct {
    StatusCode int         \`json:"statusCode"\`
    Body       interface{} \`json:"body"\`
}

func Handler(ctx context.Context, event Event) (Response, error) {
    log.Printf("Function invoked with event: %+v", event)
    
    return Response{
        StatusCode: 200,
        Body: map[string]interface{}{
            "message":   "Hello from CloudBox Function!",
            "timestamp": time.Now().Format(time.RFC3339),
            "input":     event,
        },
    }, nil
}`
    };
    return templates[language] || templates.javascript;
  }

  onMount(async () => {
    await loadFunctions();
  });

  async function loadFunctions() {
    loading = true;
    try {
      const response = await fetch(`${API_BASE_URL}/api/v1/projects/${projectId}/functions`, {
        headers: auth.getAuthHeader()
      });

      if (response.ok) {
        functions = await response.json();
      } else {
        toast.error('Fout bij laden functies');
      }
    } catch (error) {
      console.error('Error loading functions:', error);
      toast.error('Netwerkfout bij laden functies');
    } finally {
      loading = false;
    }
  }

  async function createFunction() {
    try {
      // Parse environment variables from string format
      let environmentObj = {};
      if (functionForm.environment && typeof functionForm.environment === 'string') {
        try {
          environmentObj = JSON.parse(functionForm.environment);
        } catch {
          environmentObj = {};
        }
      }

      const response = await fetch(`${API_BASE_URL}/api/v1/projects/${projectId}/functions`, {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
          ...auth.getAuthHeader()
        },
        body: JSON.stringify({
          ...functionForm,
          environment: environmentObj
        })
      });

      if (response.ok) {
        toast.success('Functie aangemaakt');
        showCreateModal = false;
        await loadFunctions();
        resetForm();
      } else {
        const error = await response.json();
        toast.error(error.error || 'Fout bij aanmaken functie');
      }
    } catch (error) {
      console.error('Error creating function:', error);
      toast.error('Netwerkfout bij aanmaken functie');
    }
  }

  async function deployFunction(functionId: number) {
    deployingFunctions[functionId] = true;
    try {
      const response = await fetch(`${API_BASE_URL}/api/v1/projects/${projectId}/functions/${functionId}/deploy`, {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
          ...auth.getAuthHeader()
        }
      });

      if (response.ok) {
        toast.success('Deployment gestart');
        // Reload functions to get updated status
        setTimeout(() => loadFunctions(), 1000);
        setTimeout(() => loadFunctions(), 5000); // Check again after build
      } else {
        const error = await response.json();
        toast.error(error.error || 'Fout bij deployen functie');
      }
    } catch (error) {
      console.error('Error deploying function:', error);
      toast.error('Netwerkfout bij deployen functie');
    } finally {
      deployingFunctions[functionId] = false;
    }
  }

  async function testFunction(functionId: number) {
    testingFunctions[functionId] = true;
    try {
      const response = await fetch(`${API_BASE_URL}/api/v1/projects/${projectId}/functions/${functionId}/execute`, {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
          ...auth.getAuthHeader()
        },
        body: JSON.stringify({
          data: { test: true, message: 'Test execution from CloudBox dashboard' }
        })
      });

      if (response.ok) {
        const result = await response.json();
        toast.success(`Functie uitgevoerd in ${result.execution_time}ms`);
        console.log('Function result:', result);
      } else {
        const error = await response.json();
        toast.error(error.error || 'Fout bij uitvoeren functie');
      }
    } catch (error) {
      console.error('Error testing function:', error);
      toast.error('Netwerkfout bij testen functie');
    } finally {
      testingFunctions[functionId] = false;
    }
  }

  async function showFunctionLogs(func) {
    try {
      const response = await fetch(`${API_BASE_URL}/api/v1/projects/${projectId}/functions/${func.id}/logs`, {
        headers: auth.getAuthHeader()
      });

      if (response.ok) {
        const logs = await response.json();
        selectedFunction = func;
        selectedLogs = logs;
        showLogsModal = true;
      } else {
        toast.error('Fout bij laden logs');
      }
    } catch (error) {
      console.error('Error loading logs:', error);
      toast.error('Netwerkfout bij laden logs');
    }
  }

  async function deleteFunction(functionId: number) {
    if (!confirm('Weet je zeker dat je deze functie wilt verwijderen?')) return;

    try {
      const response = await fetch(`${API_BASE_URL}/api/v1/projects/${projectId}/functions/${functionId}`, {
        method: 'DELETE',
        headers: auth.getAuthHeader()
      });

      if (response.ok) {
        toast.success('Functie verwijderd');
        await loadFunctions();
      } else {
        const error = await response.json();
        toast.error(error.error || 'Fout bij verwijderen functie');
      }
    } catch (error) {
      console.error('Error deleting function:', error);
      toast.error('Netwerkfout bij verwijderen functie');
    }
  }

  function showFunctionCode(func) {
    selectedFunction = func;
    showCodeModal = true;
  }

  function copyToClipboard(text: string) {
    navigator.clipboard.writeText(text).then(() => {
      toast.success('Gekopieerd naar klembord');
    }).catch(() => {
      toast.error('Kon niet kopiëren naar klembord');
    });
  }

  function resetForm() {
    functionForm = {
      name: '',
      description: '',
      runtime: 'nodejs18',
      language: 'javascript',
      code: getDefaultCode('javascript'),
      entry_point: 'index.handler',
      timeout: 30,
      memory: 128,
      environment: {},
      is_public: false
    };
  }

  function onLanguageChange() {
    functionForm.code = getDefaultCode(functionForm.language);
    // Update runtime and entry point based on language
    const runtimeMap = {
      javascript: { runtime: 'nodejs18', entry_point: 'index.handler' },
      python: { runtime: 'python3.9', entry_point: 'main.handler' },
      go: { runtime: 'go1.19', entry_point: 'main.Handler' }
    };
    
    const config = runtimeMap[functionForm.language];
    if (config) {
      functionForm.runtime = config.runtime;
      functionForm.entry_point = config.entry_point;
    }
  }

  function getStatusColor(status: string) {
    const colors = {
      draft: 'text-gray-600 bg-gray-50 border-gray-200',
      building: 'text-blue-600 bg-blue-50 border-blue-200',
      deployed: 'text-green-600 bg-green-50 border-green-200',
      error: 'text-red-600 bg-red-50 border-red-200'
    };
    return colors[status] || colors.draft;
  }

  function getFunctionIcon(language: string) {
    const icons = {
      javascript: '🟨',
      python: '🐍',
      go: '🔷',
    };
    return icons[language] || '⚡';
  }
</script>

<svelte:head>
  <title>Functions - CloudBox</title>
</svelte:head>

<div class="p-6">
  <div class="flex justify-between items-center mb-6">
    <div>
      <h1 class="text-3xl font-bold text-foreground">Functions</h1>
      <p class="text-muted-foreground mt-1">Serverless functies voor je applicatie</p>
    </div>
    <Button on:click={() => showCreateModal = true} class="bg-primary text-primary-foreground">
      <span class="mr-2">+</span>
      Functie Maken
    </Button>
  </div>

  {#if loading}
    <div class="text-center py-8">
      <div class="inline-block animate-spin rounded-full h-8 w-8 border-b-2 border-primary"></div>
      <p class="mt-2 text-muted-foreground">Laden...</p>
    </div>
  {:else}
    {#if functions.length === 0}
      <Card class="p-8 text-center">
        <div class="text-6xl mb-4">⚡</div>
        <h3 class="text-lg font-semibold mb-2">Nog geen functies</h3>
        <p class="text-muted-foreground mb-4">Maak je eerste serverless functie om API endpoints en automatiseringen te bouwen.</p>
        <Button on:click={() => showCreateModal = true}>
          Eerste Functie Maken
        </Button>
      </Card>
    {:else}
      <div class="grid gap-4">
        {#each functions as func}
          <Card class="p-6">
            <div class="flex justify-between items-start">
              <div class="flex-1">
                <div class="flex items-center gap-3 mb-2">
                  <span class="text-2xl">{getFunctionIcon(func.language)}</span>
                  <h3 class="text-lg font-semibold">{func.name}</h3>
                  <span class="px-2 py-1 text-xs font-medium rounded-full border {getStatusColor(func.status)}">
                    {func.status}
                  </span>
                  <span class="px-2 py-1 text-xs font-medium rounded-full bg-blue-50 border border-blue-200 text-blue-600">
                    {func.runtime}
                  </span>
                  {#if func.is_public}
                    <span class="px-2 py-1 text-xs font-medium rounded-full bg-green-50 border border-green-200 text-green-600">
                      Public
                    </span>
                  {/if}
                </div>
                <p class="text-muted-foreground text-sm mb-3">{func.description || 'Geen beschrijving'}</p>
                
                <div class="grid grid-cols-2 md:grid-cols-4 gap-4 text-sm mb-3">
                  <div>
                    <span class="font-medium text-muted-foreground">Memory:</span>
                    <p>{func.memory} MB</p>
                  </div>
                  <div>
                    <span class="font-medium text-muted-foreground">Timeout:</span>
                    <p>{func.timeout}s</p>
                  </div>
                  <div>
                    <span class="font-medium text-muted-foreground">Version:</span>
                    <p>v{func.version}</p>
                  </div>
                  <div>
                    <span class="font-medium text-muted-foreground">Language:</span>
                    <p class="capitalize">{func.language}</p>
                  </div>
                </div>

                {#if func.function_url}
                  <div class="mb-3">
                    <span class="font-medium text-muted-foreground text-sm">Function URL:</span>
                    <code class="ml-1 bg-gray-100 px-2 py-1 rounded text-xs font-mono break-all">{func.function_url}</code>
                  </div>
                {/if}

                {#if func.last_deployed_at}
                  <div class="text-sm text-muted-foreground">
                    Laatst gedeployed: {new Date(func.last_deployed_at).toLocaleString('nl-NL')}
                  </div>
                {/if}
              </div>

              <div class="flex gap-2 ml-4 flex-col">
                {#if func.status === 'deployed'}
                  <Button
                    on:click={() => testFunction(func.id)}
                    size="sm"
                    variant="outline"
                    disabled={testingFunctions[func.id]}
                    class="border-green-300 text-green-600 hover:bg-green-50"
                  >
                    {testingFunctions[func.id] ? 'Testen...' : 'Test'}
                  </Button>
                {/if}
                
                <Button
                  on:click={() => deployFunction(func.id)}
                  size="sm"
                  variant="outline"
                  disabled={deployingFunctions[func.id] || func.status === 'building'}
                  class="border-blue-300 text-blue-600 hover:bg-blue-50"
                >
                  {deployingFunctions[func.id] || func.status === 'building' ? 'Deploying...' : 'Deploy'}
                </Button>
                
                <Button
                  on:click={() => showFunctionCode(func)}
                  size="sm"
                  variant="outline"
                  class="border-purple-300 text-purple-600 hover:bg-purple-50"
                >
                  Code
                </Button>
                
                <Button
                  on:click={() => showFunctionLogs(func)}
                  size="sm"
                  variant="outline"
                  class="border-gray-300 text-gray-600 hover:bg-gray-50"
                >
                  Logs
                </Button>
                
                <Button
                  on:click={() => deleteFunction(func.id)}
                  size="sm"
                  variant="outline"
                  class="border-red-300 text-red-600 hover:bg-red-50"
                >
                  Verwijder
                </Button>
              </div>
            </div>
          </Card>
        {/each}
      </div>
    {/if}
  {/if}
</div>

<!-- Create Function Modal -->
{#if showCreateModal}
  <Modal on:close={() => showCreateModal = false}>
    <div class="p-6 max-h-[80vh] overflow-y-auto">
      <h2 class="text-xl font-semibold mb-4">Nieuwe Functie Maken</h2>
      
      <form on:submit|preventDefault={createFunction} class="space-y-4">
        <div class="grid grid-cols-2 gap-4">
          <div>
            <Label for="name">Functie Naam</Label>
            <Input
              id="name"
              bind:value={functionForm.name}
              placeholder="my-awesome-function"
              required
            />
          </div>
          <div>
            <Label for="language">Programmeertaal</Label>
            <Select id="language" bind:value={functionForm.language} on:change={onLanguageChange}>
              <option value="javascript">JavaScript</option>
              <option value="python">Python</option>
              <option value="go">Go</option>
            </Select>
          </div>
        </div>

        <div>
          <Label for="description">Beschrijving</Label>
          <Textarea
            id="description"
            bind:value={functionForm.description}
            placeholder="Beschrijf wat deze functie doet"
            rows={2}
          />
        </div>

        <div class="grid grid-cols-3 gap-4">
          <div>
            <Label for="runtime">Runtime</Label>
            <Input
              id="runtime"
              bind:value={functionForm.runtime}
              placeholder="nodejs18"
              readonly
            />
          </div>
          <div>
            <Label for="timeout">Timeout (sec)</Label>
            <Input
              id="timeout"
              type="number"
              bind:value={functionForm.timeout}
              min="1"
              max="900"
            />
          </div>
          <div>
            <Label for="memory">Memory (MB)</Label>
            <Select id="memory" bind:value={functionForm.memory}>
              <option value="128">128 MB</option>
              <option value="256">256 MB</option>
              <option value="512">512 MB</option>
              <option value="1024">1024 MB</option>
            </Select>
          </div>
        </div>

        <div>
          <Label for="entry_point">Entry Point</Label>
          <Input
            id="entry_point"
            bind:value={functionForm.entry_point}
            placeholder="index.handler"
          />
        </div>

        <div>
          <Label for="code">Functie Code</Label>
          <Textarea
            id="code"
            bind:value={functionForm.code}
            rows={15}
            class="font-mono text-sm"
            required
          />
        </div>

        <div class="flex items-center space-x-2">
          <input
            id="is_public"
            type="checkbox"
            bind:checked={functionForm.is_public}
            class="rounded border-border text-primary focus:ring-primary"
          />
          <Label for="is_public" class="text-sm cursor-pointer">
            Publiekelijk toegankelijk maken
          </Label>
        </div>

        <div class="bg-blue-50 border border-blue-200 rounded-lg p-4">
          <p class="text-blue-800 text-sm">
            <strong>Tip:</strong> Na het maken kun je de functie deployen om deze beschikbaar te maken via HTTP endpoints.
          </p>
        </div>

        <div class="flex justify-end space-x-2 pt-4">
          <Button type="button" variant="outline" on:click={() => showCreateModal = false}>
            Annuleren
          </Button>
          <Button type="submit" class="bg-primary text-primary-foreground">
            Functie Maken
          </Button>
        </div>
      </form>
    </div>
  </Modal>
{/if}

<!-- Function Code Modal -->
{#if showCodeModal && selectedFunction}
  <Modal on:close={() => showCodeModal = false}>
    <div class="p-6 max-w-4xl">
      <h2 class="text-xl font-semibold mb-4">Code: {selectedFunction.name}</h2>
      
      <div class="space-y-4">
        <div class="bg-gray-50 rounded-lg p-4">
          <div class="flex justify-between items-center mb-2">
            <span class="text-sm font-medium text-gray-700">Function Code</span>
            <Button
              on:click={() => copyToClipboard(selectedFunction.code)}
              size="sm"
            >
              Kopiëren
            </Button>
          </div>
          <pre class="text-xs font-mono bg-white p-4 rounded border overflow-x-auto max-h-96">{selectedFunction.code}</pre>
        </div>

        <div class="grid grid-cols-2 gap-4 text-sm">
          <div>
            <span class="font-medium text-muted-foreground">Runtime:</span>
            <p>{selectedFunction.runtime}</p>
          </div>
          <div>
            <span class="font-medium text-muted-foreground">Entry Point:</span>
            <p>{selectedFunction.entry_point}</p>
          </div>
          <div>
            <span class="font-medium text-muted-foreground">Memory:</span>
            <p>{selectedFunction.memory} MB</p>
          </div>
          <div>
            <span class="font-medium text-muted-foreground">Timeout:</span>
            <p>{selectedFunction.timeout} seconds</p>
          </div>
        </div>

        <div class="flex justify-end pt-4">
          <Button on:click={() => showCodeModal = false}>
            Sluiten
          </Button>
        </div>
      </div>
    </div>
  </Modal>
{/if}

<!-- Function Logs Modal -->
{#if showLogsModal && selectedFunction}
  <Modal on:close={() => showLogsModal = false}>
    <div class="p-6 max-w-4xl">
      <h2 class="text-xl font-semibold mb-4">Execution Logs: {selectedFunction.name}</h2>
      
      <div class="space-y-4">
        {#if selectedLogs.length === 0}
          <div class="text-center py-8 text-muted-foreground">
            Nog geen execution logs beschikbaar
          </div>
        {:else}
          <div class="space-y-2 max-h-96 overflow-y-auto">
            {#each selectedLogs as log}
              <div class="bg-gray-50 rounded-lg p-3 text-sm">
                <div class="flex justify-between items-start mb-2">
                  <span class="font-mono text-xs text-gray-500">{log.execution_id}</span>
                  <div class="flex gap-2">
                    <span class="px-2 py-1 text-xs font-medium rounded-full {log.status === 'success' ? 'bg-green-50 text-green-600' : 'bg-red-50 text-red-600'}">
                      {log.status}
                    </span>
                    <span class="text-xs text-gray-500">{log.execution_time}ms</span>
                  </div>
                </div>
                <div class="text-xs text-gray-600">
                  <strong>Method:</strong> {log.method} | 
                  <strong>IP:</strong> {log.client_ip} | 
                  <strong>Time:</strong> {new Date(log.created_at).toLocaleString('nl-NL')}
                </div>
                {#if log.logs}
                  <pre class="text-xs mt-2 bg-white p-2 rounded border">{log.logs}</pre>
                {/if}
                {#if log.error_message}
                  <pre class="text-xs mt-2 bg-red-50 text-red-800 p-2 rounded border border-red-200">{log.error_message}</pre>
                {/if}
              </div>
            {/each}
          </div>
        {/if}

        <div class="flex justify-end pt-4">
          <Button on:click={() => showLogsModal = false}>
            Sluiten
          </Button>
        </div>
      </div>
    </div>
  </Modal>
{/if}