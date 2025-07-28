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
  import Label from '$lib/components/ui/label.svelte';
  import Select from '$lib/components/ui/select.svelte';
  import Textarea from '$lib/components/ui/textarea.svelte';
  import Icon from '$lib/components/ui/icon.svelte';

  let projectId = $page.params.id;
  let repositories = [];
  let loading = true;
  let showCreateModal = false;
  let showWebhookModal = false;
  let showEditModal = false;
  let selectedRepo = null;
  let editingRepo = null;

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
                  <span class="px-2 py-1 text-xs font-medium rounded-full bg-blue-100 dark:bg-blue-900 border border-blue-200 dark:border-blue-800 text-blue-700 dark:text-blue-300">
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

              <div class="flex gap-2 ml-4 flex-col">
                <Button
                  on:click={() => syncRepository(repo.id)}
                  size="sm"
                  variant="outline"
                  class="border-blue-300 text-blue-600 hover:bg-blue-50 hover:border-blue-400"
                >
                  <Icon name="refresh" size={16} />
                </Button>
                <Button
                  on:click={() => editRepository(repo)}
                  size="sm"
                  variant="outline"
                  class="border-purple-300 text-purple-600 hover:bg-purple-50 hover:border-purple-400"
                >
                  <Icon name="edit" size={16} />
                </Button>
                <Button
                  on:click={() => showWebhookInfo(repo)}
                  size="sm"
                  variant="outline"
                  class="border-green-300 text-green-600 hover:bg-green-50 hover:border-green-400"
                >
                  <Icon name="link" size={16} />
                </Button>
                <Button
                  on:click={() => deleteRepository(repo.id)}
                  size="sm"
                  variant="outline"
                  class="border-red-300 text-red-600 hover:bg-red-50 hover:border-red-400"
                >
                  <Icon name="trash" size={16} />
                </Button>
              </div>
            </div>
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