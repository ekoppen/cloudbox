<script lang="ts">
  import { page } from '$app/stores';
  import { goto } from '$app/navigation';
  import { onMount } from 'svelte';
  import { API_BASE_URL, createApiRequest } from '$lib/config';
  import { auth } from '$lib/stores/auth';
  import { toast } from '$lib/stores/toast';
  import Card from '$lib/components/ui/card.svelte';
  import Button from '$lib/components/ui/button.svelte';
  import Modal from '$lib/components/ui/modal.svelte';
  import Input from '$lib/components/ui/input.svelte';
  import Label from '$lib/components/ui/label.svelte';
  import Select from '$lib/components/ui/select.svelte';
  import Textarea from '$lib/components/ui/textarea.svelte';
  import Icon from '$lib/components/ui/icon.svelte';
  import InstallOptions from '$lib/components/ui/install-options.svelte';

  let projectId = $page.params.id;
  let repositories = [];
  let loading = true;
  let showCreateModal = false;
  let showWebhookModal = false;
  let showEditModal = false;
  let showAnalysisModal = false;
  let selectedRepo = null;
  let editingRepo = null;
  let repoAnalysis = null;
  let analyzingRepo = null;
  let testingRepo = null;
  let editingToken = null; // Repository ID being edited for PAT

  // Form data voor nieuwe repository
  let repoForm = {
    name: '',
    full_name: '',
    clone_url: '',
    branch: 'main',
    is_private: false,
    description: '',
    sdk_version: '1.0.0',
    app_port: 3000,
    build_command: 'npm run build',
    start_command: 'npm start',
    environment: {}
  };

  // Form data voor bewerken repository
  let editForm = {
    name: '',
    full_name: '',
    clone_url: '',
    branch: 'main',
    is_private: false,
    description: '',
    sdk_version: '1.0.0',
    app_port: 3000,
    build_command: 'npm run build',
    start_command: 'npm start',
    environment: {}
  };

  onMount(async () => {
    await loadRepositories();
  });

  async function loadRepositories() {
    loading = true;
    try {
      const response = await fetch(`${API_BASE_URL}/api/v1/projects/${projectId}/github-repositories`, {
        headers: auth.getAuthHeader()
      });

      if (response.ok) {
        repositories = await response.json();
      } else {
        toast.error('Fout bij laden GitHub repositories');
      }
    } catch (error) {
      console.error('Error loading repositories:', error);
      toast.error('Netwerkfout bij laden repositories');
    } finally {
      loading = false;
    }
  }

  async function createRepository() {
    try {
      const response = await fetch(`${API_BASE_URL}/api/v1/projects/${projectId}/github-repositories`, {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
          ...auth.getAuthHeader()
        },
        body: JSON.stringify(repoForm)
      });

      if (response.ok) {
        toast.success('GitHub repository toegevoegd');
        showCreateModal = false;
        await loadRepositories();
        resetForm();
      } else {
        const error = await response.json();
        toast.error(error.error || 'Fout bij toevoegen repository');
      }
    } catch (error) {
      console.error('Error creating repository:', error);
      toast.error('Netwerkfout bij toevoegen repository');
    }
  }

  async function syncRepository(repoId: number) {
    try {
      const response = await fetch(`${API_BASE_URL}/api/v1/projects/${projectId}/github-repositories/${repoId}/sync`, {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
          ...auth.getAuthHeader()
        }
      });

      if (response.ok) {
        toast.success('Repository gesynchroniseerd');
        await loadRepositories();
      } else {
        const error = await response.json();
        toast.error(error.error || 'Fout bij synchroniseren repository');
      }
    } catch (error) {
      console.error('Error syncing repository:', error);
      toast.error('Netwerkfout bij synchroniseren repository');
    }
  }

  async function updateRepository() {
    try {
      const response = await fetch(`${API_BASE_URL}/api/v1/projects/${projectId}/github-repositories/${editingRepo.id}`, {
        method: 'PUT',
        headers: {
          'Content-Type': 'application/json',
          ...auth.getAuthHeader()
        },
        body: JSON.stringify(editForm)
      });

      if (response.ok) {
        toast.success('Repository bijgewerkt');
        showEditModal = false;
        await loadRepositories();
        resetEditForm();
      } else {
        const error = await response.json();
        toast.error(error.error || 'Fout bij bijwerken repository');
      }
    } catch (error) {
      console.error('Error updating repository:', error);
      toast.error('Netwerkfout bij bijwerken repository');
    }
  }

  async function deleteRepository(repoId: number) {
    if (!confirm('Weet je zeker dat je deze repository wilt verwijderen?')) return;

    try {
      const response = await fetch(`${API_BASE_URL}/api/v1/projects/${projectId}/github-repositories/${repoId}`, {
        method: 'DELETE',
        headers: auth.getAuthHeader()
      });

      if (response.ok) {
        toast.success('Repository verwijderd');
        await loadRepositories();
      } else {
        const error = await response.json();
        toast.error(error.error || 'Fout bij verwijderen repository');
      }
    } catch (error) {
      console.error('Error deleting repository:', error);
      toast.error('Netwerkfout bij verwijderen repository');
    }
  }

  async function showWebhookInfo(repo) {
    try {
      const response = await fetch(`${API_BASE_URL}/api/v1/projects/${projectId}/github-repositories/${repo.id}/webhook`, {
        headers: auth.getAuthHeader()
      });

      if (response.ok) {
        const webhookData = await response.json();
        selectedRepo = { ...repo, webhookData };
        showWebhookModal = true;
      } else {
        toast.error('Fout bij laden webhook informatie');
      }
    } catch (error) {
      console.error('Error loading webhook info:', error);
      toast.error('Netwerkfout bij laden webhook informatie');
    }
  }

  function copyToClipboard(text: string) {
    navigator.clipboard.writeText(text).then(() => {
      toast.success('Gekopieerd naar klembord');
    }).catch(() => {
      toast.error('Kon niet kopiëren naar klembord');
    });
  }

  function editRepository(repo) {
    editingRepo = repo;
    editForm = {
      name: repo.name,
      full_name: repo.full_name,
      clone_url: repo.clone_url,
      branch: repo.branch,
      is_private: repo.is_private,
      description: repo.description || '',
      sdk_version: repo.sdk_version || '1.0.0',
      app_port: repo.app_port,
      build_command: repo.build_command,
      start_command: repo.start_command,
      environment: repo.environment || {}
    };
    showEditModal = true;
  }

  function resetForm() {
    repoForm = {
      name: '',
      full_name: '',
      clone_url: '',
      branch: 'main',
      is_private: false,
      description: '',
      sdk_version: '1.0.0',
      app_port: 3000,
      build_command: 'npm run build',
      start_command: 'npm start',
      environment: {}
    };
  }

  function resetEditForm() {
    editForm = {
      name: '',
      full_name: '',
      clone_url: '',
      branch: 'main',
      is_private: false,
      description: '',
      sdk_version: '1.0.0',
      app_port: 3000,
      build_command: 'npm run build',
      start_command: 'npm start',
      environment: {}
    };
    editingRepo = null;
  }

  function getRepoIcon(isPrivate: boolean) {
    return isPrivate ? '🔒' : '📂';
  }

  async function getRepositoryAnalysis(repo) {
    try {
      const response = await fetch(`${API_BASE_URL}/api/v1/projects/${projectId}/github-repositories/${repo.id}/analysis`, {
        headers: auth.getAuthHeader()
      });

      if (response.ok) {
        const analysis = await response.json();
        selectedRepo = repo;
        repoAnalysis = analysis;
        showAnalysisModal = true;
      } else if (response.status === 404) {
        // No analysis found, suggest to analyze
        toast.info('Geen analyse gevonden. Klik op "Analyseren" om de repository te analyseren.');
      } else {
        toast.error('Fout bij laden repository analyse');
      }
    } catch (error) {
      console.error('Error loading repository analysis:', error);
      toast.error('Netwerkfout bij laden repository analyse');
    }
  }

  async function checkCIPCompliance(repo) {
    analyzingRepo = repo.id;
    try {
      const response = await fetch(`${API_BASE_URL}/api/v1/projects/${projectId}/github-repositories/${repo.id}/cip-check`, {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
          ...auth.getAuthHeader()
        },
        body: JSON.stringify({
          repo_url: repo.clone_url,
          branch: repo.branch
        })
      });

      if (response.ok) {
        const result = await response.json();
        if (result.is_cip_compliant) {
          toast.success(`✅ CIP Compliant - ${result.message || 'Repository heeft cloudbox.json en vereiste scripts'}`);
        } else {
          toast.error(`❌ Niet CIP Compliant - ${result.message || 'Geen cloudbox.json gevonden'}`);
        }
        
        // Reload repositories to get updated CIP status
        await loadRepositories();
      } else {
        const error = await response.json();
        toast.error(error.error || 'Fout bij controleren CIP compliance');
      }
    } catch (error) {
      console.error('Error checking CIP compliance:', error);
      toast.error('Netwerkfout bij controleren CIP compliance');
    } finally {
      analyzingRepo = null;
    }
  }

  function getProjectTypeIcon(projectType: string) {
    const icons = {
      'react': '⚛️',
      'vue': '💚',
      'angular': '🅰️',
      'nextjs': '🔼',
      'nuxt': '💚',
      'svelte': '🔶',
      'nodejs': '💚',
      'photoportfolio': '📸',
      'unknown': '❓'
    };
    return icons[projectType] || icons.unknown;
  }

  function getComplexityColor(complexity: number) {
    if (complexity <= 3) return 'bg-green-100 text-green-800 border-green-200';
    if (complexity <= 6) return 'bg-yellow-100 text-yellow-800 border-yellow-200';
    return 'bg-red-100 text-red-800 border-red-200';
  }

  // GitHub Access functions
  async function testRepositoryAccess(repo) {
    testingRepo = repo.id;
    try {
      const response = await fetch(`${API_BASE_URL}/api/v1/projects/${projectId}/github-repositories/${repo.id}/test-access`, {
        headers: {
          'Content-Type': 'application/json',
          ...auth.getAuthHeader()
        }
      });

      if (response.ok) {
        const result = await response.json();
        
        if (result.access_test === 'success') {
          toast.success(`✅ Repository toegang werkt! (${result.authorized_by})`);
        } else {
          if (result.needs_auth) {
            toast.error('❌ Geen toegang - voeg een Personal Access Token toe');
          } else {
            toast.error(`❌ Toegang gefaald: ${result.error}`);
          }
        }
        console.log('Access test result:', result);
      } else {
        const error = await response.json();
        toast.error(error.error || 'Fout bij testen van repository toegang');
      }
    } catch (error) {
      console.error('Error testing repository access:', error);
      toast.error('Netwerkfout bij testen repository toegang');
    } finally {
      testingRepo = null;
    }
  }

  // Personal Access Token functions
  async function updateRepositoryToken(repo, token) {
    try {
      const response = await fetch(`${API_BASE_URL}/api/v1/projects/${projectId}/github-repositories/${repo.id}/token`, {
        method: 'PUT',
        headers: {
          'Content-Type': 'application/json',
          ...auth.getAuthHeader()
        },
        body: JSON.stringify({ access_token: token })
      });

      if (response.ok) {
        toast.success('✅ Personal Access Token opgeslagen!');
        editingToken = null;
        await loadRepositories(); // Reload to get updated repository data
      } else {
        const error = await response.json();
        toast.error(error.error || 'Fout bij opslaan Personal Access Token');
      }
    } catch (error) {
      console.error('Error updating repository token:', error);
      toast.error('Netwerkfout bij opslaan token');
    }
  }

  function toggleTokenEdit(repoId) {
    editingToken = editingToken === repoId ? null : repoId;
  }

  function getComplexityLabel(complexity: number) {
    if (complexity <= 3) return 'Eenvoudig';
    if (complexity <= 6) return 'Gemiddeld';
    return 'Complex';
  }

  async function createDeployment(repo) {
    // Navigate directly to deployments page where user can create deployment
    // Pre-fill repo_id in URL so deployment modal can pre-select the repository
    goto(`/dashboard/projects/${projectId}/deployments?repo_id=${repo.id}`);
  }
</script>

<svelte:head>
  <title>GitHub Repositories - CloudBox</title>
</svelte:head>

<div class="p-6">
  <div class="flex justify-between items-center mb-6">
    <div>
      <h1 class="text-3xl font-bold text-foreground">GitHub Repositories</h1>
      <p class="text-muted-foreground mt-1">Koppel je GitHub repositories voor automatische deployments</p>
    </div>
    <Button on:click={() => showCreateModal = true} class="bg-primary text-primary-foreground">
      <Icon name="code" size={16} className="mr-2" />
      Repository Toevoegen
    </Button>
  </div>

  {#if loading}
    <div class="text-center py-8">
      <div class="inline-block animate-spin rounded-full h-8 w-8 border-b-2 border-primary"></div>
      <p class="mt-2 text-muted-foreground">Laden...</p>
    </div>
  {:else}
    {#if repositories.length === 0}
      <Card class="p-8 text-center">
        <div class="text-6xl mb-4">📂</div>
        <h3 class="text-lg font-semibold mb-2">Nog geen repositories</h3>
        <p class="text-muted-foreground mb-4">Voeg je eerste GitHub repository toe om automatische deployments in te stellen.</p>
        <Button on:click={() => showCreateModal = true}>
          <Icon name="code" size={16} className="mr-2" />
          Repository Toevoegen
        </Button>
      </Card>
    {:else}
      <div class="grid gap-4">
        {#each repositories as repo}
          <Card class="p-6">
            <div class="flex justify-between items-start">
              <div class="flex-1">
                <div class="flex items-center gap-3 mb-2">
                  <span class="text-2xl">{getRepoIcon(repo.is_private)}</span>
                  <h3 class="text-lg font-semibold">{repo.name}</h3>
                  <span class="px-2 py-1 text-xs font-medium rounded-full bg-blue-100 dark:bg-gray-800 border border-blue-200 dark:border-gray-600 text-blue-700 dark:text-blue-300">
                    {repo.branch}
                  </span>
                  {#if repo.is_active}
                    <span class="px-2 py-1 text-xs font-medium rounded-full bg-green-100 dark:bg-green-900 border border-green-200 dark:border-green-800 text-green-700 dark:text-green-300">
                      Actief
                    </span>
                  {/if}
                </div>
                <p class="text-muted-foreground text-sm mb-3">{repo.description || 'Geen beschrijving'}</p>
                
                <div class="grid grid-cols-2 md:grid-cols-4 gap-4 text-sm">
                  <div>
                    <span class="font-medium text-muted-foreground">Full Name:</span>
                    <p class="font-mono text-xs">{repo.full_name}</p>
                  </div>
                  <div>
                    <span class="font-medium text-muted-foreground">SDK Version:</span>
                    <p>{repo.sdk_version || 'Niet ingesteld'}</p>
                  </div>
                  <div>
                    <span class="font-medium text-muted-foreground">App Port:</span>
                    <p>{repo.app_port}</p>
                  </div>
                  <div>
                    <span class="font-medium text-muted-foreground">Type:</span>
                    <p>{repo.is_private ? 'Private' : 'Public'}</p>
                  </div>
                </div>

                <div class="mt-3 grid grid-cols-2 gap-4 text-sm">
                  <div>
                    <span class="font-medium text-muted-foreground">Build Command:</span>
                    <code class="text-xs bg-muted px-1 rounded">{repo.build_command}</code>
                  </div>
                  <div>
                    <span class="font-medium text-muted-foreground">Start Command:</span>
                    <code class="text-xs bg-muted px-1 rounded">{repo.start_command}</code>
                  </div>
                </div>

                {#if repo.last_sync_at}
                  <div class="mt-2 text-sm text-muted-foreground">
                    Laatst gesynchroniseerd: {new Date(repo.last_sync_at).toLocaleString('nl-NL')}
                  </div>
                {/if}
              </div>

              <div class="ml-4">
                <!-- Primaire acties (horizontaal, prominent) -->
                <div class="flex gap-2 mb-3">
                  <Button
                    on:click={() => checkCIPCompliance(repo)}
                    size="sm"
                    variant="default"
                    class="bg-blue-600 hover:bg-blue-700 text-white"
                    disabled={analyzingRepo === repo.id}
                    title="Controleer of repository CloudBox Install Protocol ondersteunt"
                  >
                    {#if analyzingRepo === repo.id}
                      <div class="animate-spin rounded-full h-4 w-4 border-b-2 border-white mr-2"></div>
                      Controleren...
                    {:else}
                      <Icon name="check-circle" size={16} class="mr-2" />
                      CIP Check
                    {/if}
                  </Button>
                  
                  <!-- Deploy knop -->
                  <Button
                    on:click={() => createDeployment(repo)}
                    size="sm" 
                    variant="default"
                    class="bg-green-600 hover:bg-green-700 text-white"
                    title="Nieuwe deployment aanmaken vanuit deze repository"
                  >
                    <Icon name="rocket" size={16} class="mr-2" />
                    Deploy
                  </Button>
                </div>
                
                <!-- Secundaire acties (horizontaal, kleiner) -->
                <div class="flex gap-1 flex-wrap">
                  <Button
                    on:click={() => syncRepository(repo.id)}
                    variant="ghost"
                    size="icon"
                    class="h-8 w-8 hover:rotate-180 transition-transform duration-300"
                    title="Synchroniseer repository"
                  >
                    <Icon name="refresh-cw" size={16} />
                  </Button>
                  
                  <Button
                    on:click={() => testRepositoryAccess(repo)}
                    size="sm"
                    variant="ghost"
                    class="text-gray-600 hover:bg-gray-100 text-xs px-2 py-1"
                    disabled={testingRepo === repo.id}
                  >
                    {#if testingRepo === repo.id}
                      <div class="animate-spin rounded-full h-3 w-3 border-b-2 border-gray-600 mr-1"></div>
                      Test...
                    {:else}
                      <Icon name="shield" size={14} class="mr-1" />
                      Test
                    {/if}
                  </Button>
                  
                  <Button
                    on:click={() => toggleTokenEdit(repo.id)}
                    size="sm"
                    variant="ghost"
                    class="text-gray-600 hover:bg-gray-100 text-xs px-2 py-1"
                    title="Personal Access Token beheren"
                  >
                    <Icon name="key" size={14} class="mr-1" />
                    Token
                  </Button>
                  
                  <Button
                    on:click={() => editRepository(repo)}
                    size="sm"
                    variant="ghost"
                    class="text-gray-600 hover:bg-gray-100 text-xs px-2 py-1"
                    title="Repository instellingen bewerken"
                  >
                    <Icon name="edit" size={14} class="mr-1" />
                    Edit
                  </Button>
                  
                  <Button
                    on:click={() => showWebhookInfo(repo)}
                    size="sm"
                    variant="ghost"
                    class="text-gray-600 hover:bg-gray-100 text-xs px-2 py-1"
                    title="Webhook configuratie bekijken"
                  >
                    <Icon name="link" size={14} class="mr-1" />
                    Webhook
                  </Button>
                  
                  <Button
                    on:click={() => deleteRepository(repo.id)}
                    size="sm"
                    variant="ghost"
                    class="text-red-600 hover:bg-red-100 text-xs px-2 py-1"
                    title="Repository verwijderen"
                  >
                    <Icon name="trash" size={14} class="mr-1" />
                    Delete
                  </Button>
                </div>
              </div>
            </div>

            <!-- Personal Access Token Section -->
            {#if editingToken === repo.id}
              <div class="mt-4 pt-4 border-t border-gray-200">
                <div class="space-y-3">
                  <div class="flex items-center justify-between">
                    <h4 class="text-sm font-medium text-gray-900">Personal Access Token</h4>
                    <a 
                      href="https://github.com/settings/personal-access-tokens/new" 
                      target="_blank" 
                      class="text-xs text-blue-600 hover:text-blue-700 underline"
                    >
                      ↗ GitHub PAT aanmaken
                    </a>
                  </div>
                  
                  <div class="bg-blue-50 border border-blue-200 rounded-md p-3">
                    <div class="text-xs text-blue-800">
                      <strong>Fine-grained Token:</strong> Selecteer alleen deze repository ({repo.full_name}) 
                      voor maximale beveiliging. Geef "Contents" read toegang.
                    </div>
                  </div>

                  <form on:submit|preventDefault={(e) => {
                    const formData = new FormData(e.target);
                    const token = formData.get('token');
                    if (token?.trim()) {
                      updateRepositoryToken(repo, token.trim());
                    }
                  }}>
                    <div class="flex gap-2">
                      <Input
                        name="token"
                        type="password"
                        placeholder="github_pat_11A..."
                        class="flex-1 text-sm"
                        required
                      />
                      <Button type="submit" size="sm" class="bg-green-600 hover:bg-green-700 text-white">
                        Opslaan
                      </Button>
                      <Button 
                        type="button" 
                        size="sm" 
                        variant="outline"
                        on:click={() => editingToken = null}
                      >
                        Annuleren
                      </Button>
                    </div>
                  </form>
                </div>
              </div>
            {/if}
          </Card>
        {/each}
      </div>
    {/if}
  {/if}
</div>

<!-- Create Repository Modal -->
{#if showCreateModal}
  <Modal open={showCreateModal} on:close={() => showCreateModal = false} size="2xl">
    <div class="p-8 max-h-[80vh] overflow-y-auto">
      <h2 class="text-xl font-semibold mb-4">GitHub Repository Toevoegen</h2>
      
      <form on:submit|preventDefault={createRepository} class="space-y-4">
        <div class="grid grid-cols-2 gap-4">
          <div>
            <Label for="name">Repository Naam</Label>
            <Input
              id="name"
              bind:value={repoForm.name}
              placeholder="my-awesome-app"
              required
            />
          </div>
          <div>
            <Label for="full_name">Full Name (owner/repo)</Label>
            <Input
              id="full_name"
              bind:value={repoForm.full_name}
              placeholder="username/my-awesome-app"
              required
            />
          </div>
        </div>

        <div>
          <Label for="clone_url">Clone URL</Label>
          <Input
            id="clone_url"
            bind:value={repoForm.clone_url}
            placeholder="https://github.com/username/my-awesome-app.git"
            required
          />
        </div>

        <div>
          <Label for="description">Beschrijving</Label>
          <Textarea
            id="description"
            bind:value={repoForm.description}
            placeholder="Mijn geweldige applicatie"
            rows={2}
          />
        </div>

        <div class="grid grid-cols-1 md:grid-cols-3 gap-4">
          <div>
            <Label for="branch">Branch</Label>
            <Input
              id="branch"
              bind:value={repoForm.branch}
              placeholder="main"
            />
          </div>
          <div>
            <Label for="sdk_version">SDK Version</Label>
            <Input
              id="sdk_version"
              bind:value={repoForm.sdk_version}
              placeholder="1.0.0"
            />
          </div>
          <div>
            <Label for="app_port">App Port</Label>
            <Input
              id="app_port"
              type="number"
              bind:value={repoForm.app_port}
              min="1"
              max="65535"
            />
          </div>
        </div>

        <div class="grid grid-cols-2 gap-4">
          <div>
            <Label for="build_command">Build Command</Label>
            <Input
              id="build_command"
              bind:value={repoForm.build_command}
              placeholder="npm run build"
            />
          </div>
          <div>
            <Label for="start_command">Start Command</Label>
            <Input
              id="start_command"
              bind:value={repoForm.start_command}
              placeholder="npm start"
            />
          </div>
        </div>

        <div class="flex items-center space-x-2">
          <input
            id="is_private"
            type="checkbox"
            bind:checked={repoForm.is_private}
            class="rounded border-border text-primary focus:ring-primary"
          />
          <Label for="is_private" class="text-sm cursor-pointer">
            Private repository
          </Label>
        </div>

        <div class="bg-blue-50 border border-blue-200 rounded-lg p-4">
          <p class="text-blue-800 text-sm">
            <strong>Tip:</strong> Na het toevoegen kun je webhook informatie ophalen om automatische deployments in te stellen.
          </p>
        </div>

        <div class="flex justify-end space-x-2 pt-4">
          <Button type="button" variant="outline" on:click={() => showCreateModal = false}>
            <Icon name="x" size={16} className="mr-2" />
            Annuleren
          </Button>
          <Button type="submit" class="bg-primary text-primary-foreground">
            <Icon name="code" size={16} className="mr-2" />
            Repository Toevoegen
          </Button>
        </div>
      </form>
    </div>
  </Modal>
{/if}

<!-- Webhook Info Modal -->
{#if showWebhookModal && selectedRepo}
  <Modal open={showWebhookModal} on:close={() => showWebhookModal = false} size="xl">
    <div class="p-8 max-h-[80vh] overflow-y-auto">
      <h2 class="text-xl font-semibold mb-4">Webhook Configuratie: {selectedRepo.name}</h2>
      
      <div class="space-y-4">
        <div>
          <Label>Webhook URL</Label>
          <div class="mt-1 relative">
            <input
              readonly
              class="w-full p-3 border border-border rounded-md font-mono text-sm bg-muted"
              value={selectedRepo.webhookData?.webhook_url || ''}
            />
            <Button
              on:click={() => copyToClipboard(selectedRepo.webhookData?.webhook_url || '')}
              size="sm"
              class="absolute top-2 right-2"
            >
              <Icon name="package" size={16} />
            </Button>
          </div>
        </div>

        <div>
          <Label>Webhook Secret</Label>
          <div class="mt-1 relative">
            <input
              readonly
              class="w-full p-3 border border-border rounded-md font-mono text-sm bg-muted"
              value={selectedRepo.webhookData?.webhook_secret || ''}
            />
            <Button
              on:click={() => copyToClipboard(selectedRepo.webhookData?.webhook_secret || '')}
              size="sm"
              class="absolute top-2 right-2"
            >
              <Icon name="package" size={16} />
            </Button>
          </div>
        </div>

        <div class="bg-green-50 dark:bg-green-900/20 border border-green-200 dark:border-green-800 rounded-lg p-4">
          <h4 class="font-medium text-green-900 dark:text-green-200 mb-2">GitHub Webhook Instellen:</h4>
          <ol class="text-green-800 dark:text-green-200 text-sm space-y-1 list-decimal list-inside">
            <li>Ga naar je GitHub repository</li>
            <li>Klik op "Settings" → "Webhooks"</li>
            <li>Klik "Add webhook"</li>
            <li>Plak de webhook URL hierboven</li>
            <li>Plak het secret</li>
            <li>Selecteer "application/json" als content type</li>
            <li>Selecteer "Just the push event"</li>
            <li>Klik "Add webhook"</li>
          </ol>
        </div>

        <div class="text-sm text-muted-foreground">
          <p><strong>Events:</strong> {selectedRepo.webhookData?.events?.join(', ') || 'push'}</p>
          <p><strong>Content Type:</strong> {selectedRepo.webhookData?.content_type || 'application/json'}</p>
        </div>

        <div class="flex justify-end pt-4">
          <Button on:click={() => showWebhookModal = false}>
            <Icon name="x" size={16} className="mr-2" />
            Sluiten
          </Button>
        </div>
      </div>
    </div>
  </Modal>
{/if}

<!-- Edit Repository Modal -->
{#if showEditModal && editingRepo}
  <Modal open={showEditModal} on:close={() => showEditModal = false} size="2xl">
    <div class="p-8 max-h-[80vh] overflow-y-auto">
      <h2 class="text-xl font-semibold mb-4">Repository Bewerken: {editingRepo.name}</h2>
      
      <form on:submit|preventDefault={updateRepository} class="space-y-4">
        <div class="grid grid-cols-2 gap-4">
          <div>
            <Label for="edit-name">Repository Naam</Label>
            <Input
              id="edit-name"
              bind:value={editForm.name}
              placeholder="my-awesome-app"
              required
            />
          </div>
          <div>
            <Label for="edit-full_name">Full Name (owner/repo)</Label>
            <Input
              id="edit-full_name"
              bind:value={editForm.full_name}
              placeholder="username/my-awesome-app"
              required
            />
          </div>
        </div>

        <div>
          <Label for="edit-clone_url">Clone URL</Label>
          <Input
            id="edit-clone_url"
            bind:value={editForm.clone_url}
            placeholder="https://github.com/username/my-awesome-app.git"
            required
          />
        </div>

        <div>
          <Label for="edit-description">Beschrijving</Label>
          <Textarea
            id="edit-description"
            bind:value={editForm.description}
            placeholder="Mijn geweldige applicatie"
            rows={2}
          />
        </div>

        <div class="grid grid-cols-1 md:grid-cols-3 gap-4">
          <div>
            <Label for="edit-branch">Branch</Label>
            <Input
              id="edit-branch"
              bind:value={editForm.branch}
              placeholder="main"
            />
          </div>
          <div>
            <Label for="edit-sdk_version">SDK Version</Label>
            <Input
              id="edit-sdk_version"
              bind:value={editForm.sdk_version}
              placeholder="1.0.0"
            />
          </div>
          <div>
            <Label for="edit-app_port">App Port</Label>
            <Input
              id="edit-app_port"
              type="number"
              bind:value={editForm.app_port}
              min="1"
              max="65535"
            />
          </div>
        </div>

        <div class="grid grid-cols-2 gap-4">
          <div>
            <Label for="edit-build_command">Build Command</Label>
            <Input
              id="edit-build_command"
              bind:value={editForm.build_command}
              placeholder="npm run build"
            />
          </div>
          <div>
            <Label for="edit-start_command">Start Command</Label>
            <Input
              id="edit-start_command"
              bind:value={editForm.start_command}
              placeholder="npm start"
            />
          </div>
        </div>

        <div class="flex items-center space-x-2">
          <input
            id="edit-is_private"
            type="checkbox"
            bind:checked={editForm.is_private}
            class="rounded border-border text-primary focus:ring-primary"
          />
          <Label for="edit-is_private" class="text-sm cursor-pointer">
            Private repository
          </Label>
        </div>

        <div class="flex justify-end space-x-2 pt-4">
          <Button type="button" variant="outline" on:click={() => showEditModal = false}>
            <Icon name="x" size={16} className="mr-2" />
            Annuleren
          </Button>
          <Button type="submit" class="bg-primary text-primary-foreground">
            <Icon name="save" size={16} className="mr-2" />
            Repository Bijwerken
          </Button>
        </div>
      </form>
    </div>
  </Modal>
{/if}

<!-- Repository Analysis Modal -->
{#if showAnalysisModal && selectedRepo && repoAnalysis}
  <Modal open={showAnalysisModal} on:close={() => showAnalysisModal = false} size="3xl">
    <div class="p-8 max-h-[90vh] overflow-y-auto">
      <div class="flex justify-between items-center mb-6">
        <div>
          <h2 class="text-2xl font-semibold flex items-center gap-3">
            {getProjectTypeIcon(repoAnalysis.project_type)}
            Repository Analyse: {selectedRepo.name}
          </h2>
          <p class="text-muted-foreground mt-1">
            Geanalyseerd op {new Date(repoAnalysis.analyzed_at).toLocaleString('nl-NL')}
          </p>
        </div>
        <Button
          on:click={() => analyzeRepository(selectedRepo, true)}
          variant="ghost"
          size="icon"
          class="text-orange-600 hover:text-orange-700 hover:bg-orange-50 hover:rotate-180 transition-transform duration-300"
          disabled={analyzingRepo === selectedRepo.id}
          title="Heranalyseer repository"
        >
          {#if analyzingRepo === selectedRepo.id}
            <div class="animate-spin rounded-full h-4 w-4 border-b-2 border-orange-600"></div>
          {:else}
            <Icon name="refresh-cw" size={16} />
          {/if}
        </Button>
      </div>

      <div class="grid gap-6">
        <!-- Project Info -->
        <Card class="p-6">
          <h3 class="text-lg font-semibold mb-4 flex items-center gap-2">
            <Icon name="code" size={20} />
            Project Informatie
          </h3>
          <div class="grid grid-cols-2 md:grid-cols-4 gap-4">
            <div>
              <span class="text-sm font-medium text-muted-foreground">Type</span>
              <p class="flex items-center gap-2">
                {getProjectTypeIcon(repoAnalysis.project_type)}
                {repoAnalysis.project_type}
              </p>
            </div>
            <div>
              <span class="text-sm font-medium text-muted-foreground">Framework</span>
              <p>{repoAnalysis.framework}</p>
            </div>
            <div>
              <span class="text-sm font-medium text-muted-foreground">Taal</span>
              <p>{repoAnalysis.language}</p>
            </div>
            <div>
              <span class="text-sm font-medium text-muted-foreground">Package Manager</span>
              <p>{repoAnalysis.package_manager}</p>
            </div>
          </div>
          
          <div class="mt-4 flex items-center gap-4">
            <div>
              <span class="text-sm font-medium text-muted-foreground">Port</span>
              <p class="font-mono text-sm">{repoAnalysis.port}</p>
            </div>
            <div>
              <span class="text-sm font-medium text-muted-foreground">Complexiteit</span>
              <span class="px-2 py-1 text-xs font-medium rounded-full border {getComplexityColor(repoAnalysis.complexity)}">
                {getComplexityLabel(repoAnalysis.complexity)} ({repoAnalysis.complexity}/10)
              </span>
            </div>
            {#if repoAnalysis.has_docker}
              <div>
                <span class="px-2 py-1 text-xs font-medium rounded-full bg-blue-100 text-blue-800 border border-blue-200">
                  🐳 Docker Support
                </span>
              </div>
            {/if}
          </div>
        </Card>

        <!-- Install Options -->
        {#if repoAnalysis.install_options && repoAnalysis.install_options.length > 0}
          <Card class="p-6">
            <h3 class="text-lg font-semibold mb-4 flex items-center gap-2">
              <Icon name="package" size={20} />
              Installatie Opties
            </h3>
            <InstallOptions 
              options={repoAnalysis.install_options} 
              showSelectButton={false}
            />
          </Card>
        {/if}

        <!-- Build Configuration -->
        <Card class="p-6">
          <h3 class="text-lg font-semibold mb-4 flex items-center gap-2">
            <Icon name="settings" size={20} />
            Build Configuratie
          </h3>
          <div class="grid grid-cols-1 md:grid-cols-2 gap-4">
            <div>
              <span class="text-sm font-medium text-muted-foreground">Install Command</span>
              <code class="block bg-muted px-3 py-2 rounded text-sm font-mono mt-1">{repoAnalysis.install_command || 'Niet ingesteld'}</code>
            </div>
            <div>
              <span class="text-sm font-medium text-muted-foreground">Build Command</span>
              <code class="block bg-muted px-3 py-2 rounded text-sm font-mono mt-1">{repoAnalysis.build_command || 'Niet ingesteld'}</code>
            </div>
            <div>
              <span class="text-sm font-medium text-muted-foreground">Start Command</span>
              <code class="block bg-muted px-3 py-2 rounded text-sm font-mono mt-1">{repoAnalysis.start_command || 'Niet ingesteld'}</code>
            </div>
            <div>
              <span class="text-sm font-medium text-muted-foreground">Dev Command</span>
              <code class="block bg-muted px-3 py-2 rounded text-sm font-mono mt-1">{repoAnalysis.dev_command || 'Niet ingesteld'}</code>
            </div>
          </div>
        </Card>

        <!-- Insights & Warnings -->
        {#if (repoAnalysis.insights && repoAnalysis.insights.length > 0) || (repoAnalysis.warnings && repoAnalysis.warnings.length > 0)}
          <Card class="p-6">
            <h3 class="text-lg font-semibold mb-4 flex items-center gap-2">
              <Icon name="lightbulb" size={20} />
              Inzichten & Aanbevelingen
            </h3>
            
            {#if repoAnalysis.insights && repoAnalysis.insights.length > 0}
              <div class="mb-4">
                <h4 class="font-medium text-green-800 mb-2 flex items-center gap-2">
                  💡 Inzichten
                </h4>
                <ul class="space-y-1">
                  {#each repoAnalysis.insights as insight}
                    <li class="text-sm text-green-700 flex items-start gap-2">
                      <Icon name="check" size={16} className="mt-0.5 text-green-600" />
                      {insight}
                    </li>
                  {/each}
                </ul>
              </div>
            {/if}

            {#if repoAnalysis.warnings && repoAnalysis.warnings.length > 0}
              <div>
                <h4 class="font-medium text-orange-800 mb-2 flex items-center gap-2">
                  ⚠️ Waarschuwingen
                </h4>
                <ul class="space-y-1">
                  {#each repoAnalysis.warnings as warning}
                    <li class="text-sm text-orange-700 flex items-start gap-2">
                      <Icon name="alert-triangle" size={16} className="mt-0.5 text-orange-600" />
                      {warning}
                    </li>
                  {/each}
                </ul>
              </div>
            {/if}
          </Card>
        {/if}

        <!-- File Structure -->
        {#if repoAnalysis.important_files && repoAnalysis.important_files.length > 0}
          <Card class="p-6">
            <h3 class="text-lg font-semibold mb-4 flex items-center gap-2">
              <Icon name="file" size={20} />
              Belangrijke Bestanden
            </h3>
            <div class="grid grid-cols-2 md:grid-cols-3 gap-2">
              {#each repoAnalysis.important_files as file}
                <code class="text-xs bg-muted px-2 py-1 rounded font-mono">{file}</code>
              {/each}
            </div>
          </Card>
        {/if}

        <!-- Performance Metrics -->
        {#if repoAnalysis.estimated_build_time || repoAnalysis.estimated_size}
          <Card class="p-6">
            <h3 class="text-lg font-semibold mb-4 flex items-center gap-2">
              <Icon name="activity" size={20} />
              Performance Schattingen
            </h3>
            <div class="grid grid-cols-2 gap-4">
              {#if repoAnalysis.estimated_build_time}
                <div>
                  <span class="text-sm font-medium text-muted-foreground">Geschatte Build Tijd</span>
                  <p class="text-lg font-semibold">{repoAnalysis.estimated_build_time}s</p>
                </div>
              {/if}
              {#if repoAnalysis.estimated_size}
                <div>
                  <span class="text-sm font-medium text-muted-foreground">Geschatte Grootte</span>
                  <p class="text-lg font-semibold">{Math.round(repoAnalysis.estimated_size / 1024 / 1024)}MB</p>
                </div>
              {/if}
            </div>
          </Card>
        {/if}
      </div>

      <div class="flex justify-end gap-2 pt-6 mt-6 border-t">
        <Button
          on:click={() => showAnalysisModal = false}
          variant="outline"
        >
          <Icon name="x" size={16} className="mr-2" />
          Sluiten
        </Button>
        <Button
          on:click={() => {
            showAnalysisModal = false;
            createDeployment(selectedRepo);
          }}
          class="bg-green-600 text-white hover:bg-green-700"
        >
          <Icon name="rocket" size={16} className="mr-2" />
          Deployment Starten
        </Button>
      </div>
    </div>
  </Modal>
{/if}