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
  import Textarea from '$lib/components/ui/textarea.svelte';
  import Icon from '$lib/components/ui/icon.svelte';
  import UpdateBadge from '$lib/components/ui/update-badge.svelte';
  import DeploymentConsole from '$lib/components/deployment/console.svelte';

  let projectId = $page.params.id;
  let deployments = [];
  let githubRepos = [];
  let webServers = [];
  let loading = true;
  let showCreateModal = false;
  let showEditModal = false;
  let editingDeployment = null;
  
  // Console state
  let activeDeploymentId = null;
  let showConsole = false;
  
  function closeConsole() {
    showConsole = false;
    activeDeploymentId = null;
  }

  // Form data voor nieuwe deployment
  let deploymentForm = {
    name: '',
    description: '',
    github_repository_id: '',
    web_server_id: '',
    domain: '',
    subdomain: '',
    port: 3000,
    environment: {},
    build_command: '',
    start_command: '',
    branch: 'main',
    is_auto_deploy_enabled: false
  };

  // Form data voor bewerken deployment  
  let editForm = {
    name: '',
    description: '',
    github_repository_id: '',
    web_server_id: '',
    domain: '',
    subdomain: '',
    port: 3000,
    build_command: '',
    start_command: '',
    branch: 'main',
    is_auto_deploy_enabled: false
  };

  onMount(async () => {
    await loadData();
  });

  async function loadData() {
    loading = true;
    try {
      const authHeaders = auth.getAuthHeader();
      console.log('Deployment page: auth headers:', authHeaders);
      console.log('Deployment page: auth state:', $auth);
      
      const headers = {
        'Content-Type': 'application/json',
        ...authHeaders
      };

      // Load deployments
      const deploymentsRes = await fetch(`${API_BASE_URL}/api/v1/projects/${projectId}/deployments`, { headers });
      if (deploymentsRes.ok) {
        deployments = await deploymentsRes.json();
      } else {
        const errorData = await deploymentsRes.text();
        console.error('Failed to load deployments:', deploymentsRes.status, errorData);
        if (deploymentsRes.status === 401) {
          toast.error('Niet geautoriseerd - log opnieuw in');
        } else if (deploymentsRes.status === 500) {
          toast.error('Server fout bij laden deployments');
        }
      }

      // Load GitHub repositories
      const reposRes = await fetch(`${API_BASE_URL}/api/v1/projects/${projectId}/github-repositories`, { headers });
      if (reposRes.ok) {
        githubRepos = await reposRes.json();
      } else {
        const errorData = await reposRes.text();
        console.error('Failed to load GitHub repos:', reposRes.status, errorData);
        if (reposRes.status === 401) {
          toast.error('Niet geautoriseerd - log opnieuw in');
        }
      }

      // Load web servers
      const serversRes = await fetch(`${API_BASE_URL}/api/v1/projects/${projectId}/web-servers`, { headers });
      if (serversRes.ok) {
        webServers = await serversRes.json();
      } else {
        const errorData = await serversRes.text();
        console.error('Failed to load web servers:', serversRes.status, errorData);
        if (serversRes.status === 401) {
          toast.error('Niet geautoriseerd - log opnieuw in');
        }
      }

    } catch (error) {
      console.error('Error loading data:', error);
      toast.error('Fout bij laden van deployment gegevens');
    } finally {
      loading = false;
    }
  }

  async function createDeployment() {
    try {
      // Convert string IDs to numbers before sending to backend
      const payload = {
        ...deploymentForm,
        github_repository_id: parseInt(deploymentForm.github_repository_id),
        web_server_id: parseInt(deploymentForm.web_server_id)
      };

      const response = await fetch(`${API_BASE_URL}/api/v1/projects/${projectId}/deployments`, {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
          ...auth.getAuthHeader()
        },
        body: JSON.stringify(payload)
      });

      if (response.ok) {
        toast.success('Deployment configuratie aangemaakt');
        showCreateModal = false;
        await loadData();
        resetForm();
      } else {
        const error = await response.json();
        toast.error(error.error || 'Fout bij aanmaken deployment');
      }
    } catch (error) {
      console.error('Error creating deployment:', error);
      toast.error('Netwerkfout bij aanmaken deployment');
    }
  }

  async function deploy(deploymentId: number) {
    try {
      // Show console and set active deployment
      activeDeploymentId = deploymentId.toString();
      showConsole = true;
      
      const response = await fetch(`${API_BASE_URL}/api/v1/projects/${projectId}/deployments/${deploymentId}/deploy`, {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
          ...auth.getAuthHeader()
        },
        body: JSON.stringify({
          branch: 'main',
          commit_hash: 'latest'
        })
      });

      if (response.ok) {
        toast.success('Deployment gestart - Console logs worden getoond');
        await loadData();
      } else {
        const error = await response.json();
        toast.error(error.error || 'Fout bij starten deployment');
        // Keep console open to show error logs
      }
    } catch (error) {
      console.error('Error deploying:', error);
      toast.error('Netwerkfout bij deployment');
    }
  }

  async function updateDeployment() {
    try {
      // Convert string IDs to numbers before sending to backend (if they exist and are strings)
      const payload = {
        ...editForm
      };
      
      // Only convert IDs if they are being sent (not empty strings)
      if (editForm.github_repository_id && typeof editForm.github_repository_id === 'string') {
        payload.github_repository_id = parseInt(editForm.github_repository_id);
      }
      if (editForm.web_server_id && typeof editForm.web_server_id === 'string') {
        payload.web_server_id = parseInt(editForm.web_server_id);
      }

      const response = await fetch(`${API_BASE_URL}/api/v1/projects/${projectId}/deployments/${editingDeployment.id}`, {
        method: 'PUT',
        headers: {
          'Content-Type': 'application/json',
          ...auth.getAuthHeader()
        },
        body: JSON.stringify(payload)
      });

      if (response.ok) {
        toast.success('Deployment bijgewerkt');
        showEditModal = false;
        await loadData();
        resetEditForm();
      } else {
        const error = await response.json();
        toast.error(error.error || 'Fout bij bijwerken deployment');
      }
    } catch (error) {
      console.error('Error updating deployment:', error);
      toast.error('Netwerkfout bij bijwerken deployment');
    }
  }

  async function deleteDeployment(deploymentId: number) {
    if (!confirm('Weet je zeker dat je deze deployment wilt verwijderen?')) return;

    try {
      const response = await fetch(`${API_BASE_URL}/api/v1/projects/${projectId}/deployments/${deploymentId}`, {
        method: 'DELETE',
        headers: auth.getAuthHeader()
      });

      if (response.ok) {
        toast.success('Deployment verwijderd');
        await loadData();
      } else {
        const error = await response.json();
        toast.error(error.error || 'Fout bij verwijderen deployment');
      }
    } catch (error) {
      console.error('Error deleting deployment:', error);
      toast.error('Netwerkfout bij verwijderen deployment');
    }
  }

  async function deployPendingUpdate(repoId: number, repoName: string) {
    if (!confirm(`Weet je zeker dat je de pending update voor ${repoName} wilt deployen?`)) return;

    try {
      const response = await fetch(`${API_BASE_URL}/api/v1/projects/${projectId}/github-repositories/${repoId}/deploy-pending`, {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
          ...auth.getAuthHeader()
        }
      });

      if (response.ok) {
        const result = await response.json();
        toast.success(`Deployment gestart voor ${result.deployments_started} environment(s)`);
        await loadData(); // Reload to update badges and status
      } else {
        const error = await response.json();
        toast.error(error.error || 'Fout bij starten pending deployment');
      }
    } catch (error) {
      console.error('Error deploying pending update:', error);
      toast.error('Netwerkfout bij pending deployment');
    }
  }

  function resetForm() {
    deploymentForm = {
      name: '',
      description: '',
      github_repository_id: '',
      web_server_id: '',
      domain: '',
      subdomain: '',
      port: 3000,
      environment: {},
      build_command: '',
      start_command: '',
      branch: 'main',
      is_auto_deploy_enabled: false
    };
  }

  function editDeployment(deployment) {
    editingDeployment = deployment;
    editForm = {
      name: deployment.name,
      description: deployment.description || '',
      github_repository_id: deployment.github_repository_id,
      web_server_id: deployment.web_server_id,
      domain: deployment.domain || '',
      subdomain: deployment.subdomain || '',
      port: deployment.port,
      build_command: deployment.build_command || '',
      start_command: deployment.start_command || '',
      branch: deployment.branch,
      is_auto_deploy_enabled: deployment.is_auto_deploy_enabled
    };
    showEditModal = true;
  }

  function resetEditForm() {
    editForm = {
      name: '',
      description: '',
      github_repository_id: '',
      web_server_id: '',
      domain: '',
      subdomain: '',
      port: 3000,
      build_command: '',
      start_command: '',
      branch: 'main',
      is_auto_deploy_enabled: false
    };
    editingDeployment = null;
  }

  function getStatusColor(status: string) {
    switch (status) {
      case 'deployed': return 'text-green-700 dark:text-green-300 bg-green-100 dark:bg-green-900 border-green-200 dark:border-green-800';
      case 'deploying': return 'text-blue-700 dark:text-blue-300 bg-blue-100 dark:bg-blue-900 border-blue-200 dark:border-blue-800';
      case 'building': return 'text-yellow-700 dark:text-yellow-300 bg-yellow-100 dark:bg-yellow-900 border-yellow-200 dark:border-yellow-800';
      case 'failed': return 'text-red-700 dark:text-red-300 bg-red-100 dark:bg-red-900 border-red-200 dark:border-red-800';
      case 'pending': return 'text-gray-700 dark:text-gray-300 bg-gray-100 dark:bg-gray-900 border-gray-200 dark:border-gray-800';
      default: return 'text-gray-700 dark:text-gray-300 bg-gray-100 dark:bg-gray-900 border-gray-200 dark:border-gray-800';
    }
  }
</script>

<svelte:head>
  <title>Deployments - CloudBox</title>
</svelte:head>

<div class="p-6">
  <div class="flex justify-between items-center mb-6">
    <div>
      <h1 class="text-3xl font-bold text-foreground">Deployments</h1>
      <p class="text-muted-foreground mt-1">Beheer je app deployments en automatisering</p>
    </div>
    <Button on:click={() => showCreateModal = true} class="bg-primary text-primary-foreground">
      <Icon name="rocket" size={16} className="mr-2" />
      Nieuwe Deployment
    </Button>
  </div>

  <!-- Deployment Console -->
  {#if showConsole && activeDeploymentId}
    <div class="mb-6">
      <DeploymentConsole 
        deploymentId={activeDeploymentId}
        projectId={projectId}
        isVisible={showConsole}
        onClose={closeConsole}
      />
    </div>
  {/if}

  {#if loading}
    <div class="text-center py-8">
      <div class="inline-block animate-spin rounded-full h-8 w-8 border-b-2 border-primary"></div>
      <p class="mt-2 text-muted-foreground">Laden...</p>
    </div>
  {:else}
    {#if deployments.length === 0}
      <Card class="p-8 text-center">
        <div class="text-6xl mb-4">🚀</div>
        <h3 class="text-lg font-semibold mb-2">Nog geen deployments</h3>
        <p class="text-muted-foreground mb-4">Maak je eerste deployment configuratie aan om je app te deployen naar een server.</p>
        
        {#if githubRepos.length === 0 || webServers.length === 0}
          <div class="bg-yellow-50 border border-yellow-200 rounded-lg p-4 mb-4">
            <p class="text-yellow-800 text-sm">
              <strong>Eerst instellen:</strong> Je hebt een GitHub repository en webserver nodig.
            </p>
            <div class="mt-2 space-x-2">
              {#if githubRepos.length === 0}
                <a href="/dashboard/projects/{projectId}/github" class="text-blue-600 hover:text-blue-800 text-sm underline">
                  GitHub Repository Toevoegen
                </a>
              {/if}
              {#if webServers.length === 0}
                <a href="/dashboard/projects/{projectId}/servers" class="text-blue-600 hover:text-blue-800 text-sm underline">
                  Webserver Toevoegen
                </a>
              {/if}
            </div>
          </div>
        {/if}

        <Button on:click={() => showCreateModal = true} disabled={githubRepos.length === 0 || webServers.length === 0}>
          <Icon name="rocket" size={16} className="mr-2" />
          Deployment Aanmaken
        </Button>
      </Card>
    {:else}
      <div class="grid gap-4">
        {#each deployments as deployment}
          <Card class="p-6 hover:shadow-md transition-shadow cursor-pointer group" on:click={() => editDeployment(deployment)}>
            <div class="flex justify-between items-start">
              <div class="flex-1">
                <div class="flex items-center gap-3 mb-2">
                  <h3 class="text-lg font-semibold group-hover:text-primary transition-colors">{deployment.name}</h3>
                  <span class="px-2 py-1 text-xs font-medium rounded-full border {getStatusColor(deployment.status)}">
                    {deployment.status}
                  </span>
                </div>
                <p class="text-muted-foreground text-sm mb-3">{deployment.description || 'Geen beschrijving'}</p>
                
                <div class="grid grid-cols-2 md:grid-cols-4 gap-4 text-sm">
                  <div>
                    <span class="font-medium text-muted-foreground">Repository:</span>
                    <p>{deployment.github_repository?.name || 'Onbekend'}</p>
                  </div>
                  <div>
                    <span class="font-medium text-muted-foreground">Server:</span>
                    <p>{deployment.web_server?.name || 'Onbekend'}</p>
                  </div>
                  <div>
                    <span class="font-medium text-muted-foreground">Branch:</span>
                    <p>{deployment.branch}</p>
                  </div>
                  <div>
                    <span class="font-medium text-muted-foreground">Port:</span>
                    <p>{deployment.port}</p>
                  </div>
                </div>

                {#if deployment.domain}
                  <div class="mt-3">
                    <span class="font-medium text-muted-foreground text-sm">URL:</span>
                    <a href="https://{deployment.domain}" target="_blank" class="text-blue-600 hover:text-blue-800 text-sm ml-1" on:click|stopPropagation>
                      {deployment.domain}
                    </a>
                  </div>
                {/if}
              </div>

              <div class="flex gap-2 ml-4" on:click|stopPropagation>
                <Button
                  on:click={() => deploy(deployment.id)}
                  size="sm"
                  disabled={deployment.status === 'deploying' || deployment.status === 'building'}
                  class="bg-green-600 text-white hover:bg-green-700"
                  title="Deploy uitvoeren"
                >
                  <Icon name={deployment.status === 'deploying' || deployment.status === 'building' ? "refresh" : "rocket"} size={16} />
                </Button>
                <Button
                  on:click={() => editDeployment(deployment)}
                  size="sm"
                  variant="outline"
                  class="border-blue-300 text-blue-600 hover:bg-blue-50 hover:border-blue-400"
                  title="Bewerken"
                >
                  <Icon name="edit" size={16} />
                </Button>
                <Button
                  on:click={() => deleteDeployment(deployment.id)}
                  size="sm"
                  variant="outline"
                  class="border-red-300 text-red-600 hover:bg-red-50 hover:border-red-400"
                  title="Verwijderen"
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

<!-- Create Deployment Modal -->
{#if showCreateModal}
  <Modal open={showCreateModal} on:close={() => showCreateModal = false} size="2xl">
    <div class="p-8 max-h-[80vh] overflow-y-auto">
      <h2 class="text-xl font-semibold mb-4">Nieuwe Deployment</h2>
      
      <form on:submit|preventDefault={createDeployment} class="space-y-4">
        <div class="grid grid-cols-2 gap-4">
          <div>
            <Label for="name">Naam</Label>
            <Input
              id="name"
              bind:value={deploymentForm.name}
              placeholder="Mijn App Productie"
              required
            />
          </div>
          <div>
            <Label for="domain">Domein (optioneel)</Label>
            <Input
              id="domain"
              bind:value={deploymentForm.domain}
              placeholder="mijnapp.com"
            />
          </div>
        </div>

        <div>
          <Label for="description">Beschrijving</Label>
          <Textarea
            id="description"
            bind:value={deploymentForm.description}
            placeholder="Productie deployment voor mijn applicatie"
            rows={2}
          />
        </div>

        <div class="grid grid-cols-2 gap-4">
          <div>
            <Label for="repo">GitHub Repository</Label>
            <select id="repo" bind:value={deploymentForm.github_repository_id} required class="w-full px-3 py-2 border border-border rounded-md bg-background text-foreground focus:outline-none focus:ring-2 focus:ring-primary focus:border-transparent">
              <option value="">Selecteer repository...</option>
              {#each githubRepos as repo}
                <option value={repo.id.toString()}>{repo.name} ({repo.branch})</option>
              {/each}
            </select>
          </div>
          <div>
            <Label for="server">Webserver</Label>
            <select id="server" bind:value={deploymentForm.web_server_id} required class="w-full px-3 py-2 border border-border rounded-md bg-background text-foreground focus:outline-none focus:ring-2 focus:ring-primary focus:border-transparent">
              <option value="">Selecteer server...</option>
              {#each webServers as server}
                <option value={server.id.toString()}>{server.name} ({server.hostname})</option>
              {/each}
            </select>
          </div>
        </div>

        <div class="grid grid-cols-1 md:grid-cols-3 gap-4">
          <div>
            <Label for="branch">Branch</Label>
            <Input
              id="branch"
              bind:value={deploymentForm.branch}
              placeholder="main"
            />
          </div>
          <div>
            <Label for="port">Port</Label>
            <Input
              id="port"
              type="number"
              bind:value={deploymentForm.port}
              min="1"
              max="65535"
            />
          </div>
          <div>
            <Label for="subdomain">Subdomain (optioneel)</Label>
            <Input
              id="subdomain"
              bind:value={deploymentForm.subdomain}
              placeholder="app"
            />
          </div>
        </div>

        <div class="grid grid-cols-2 gap-4">
          <div>
            <Label for="build">Build Command</Label>
            <Input
              id="build"
              bind:value={deploymentForm.build_command}
              placeholder="npm run build"
            />
          </div>
          <div>
            <Label for="start">Start Command</Label>
            <Input
              id="start"
              bind:value={deploymentForm.start_command}
              placeholder="npm start"
            />
          </div>
        </div>

        <div class="flex items-center space-x-2">
          <input
            id="auto-deploy"
            type="checkbox"
            bind:checked={deploymentForm.is_auto_deploy_enabled}
            class="rounded border-border text-primary focus:ring-primary"
          />
          <Label for="auto-deploy" class="text-sm cursor-pointer">
            Automatisch deployen bij push naar branch
          </Label>
        </div>

        <div class="flex justify-end space-x-2 pt-4">
          <Button type="button" variant="outline" on:click={() => showCreateModal = false}>
            <Icon name="x" size={16} className="mr-2" />
            Annuleren
          </Button>
          <Button type="submit" class="bg-primary text-primary-foreground">
            <Icon name="rocket" size={16} className="mr-2" />
            Deployment Aanmaken
          </Button>
        </div>
      </form>
    </div>
  </Modal>
{/if}

<!-- Edit Deployment Modal -->
{#if showEditModal && editingDeployment}
  <Modal open={showEditModal} on:close={() => showEditModal = false} size="2xl">
    <div class="p-8 max-h-[80vh] overflow-y-auto">
      <h2 class="text-xl font-semibold mb-4">Deployment Bewerken: {editingDeployment.name}</h2>
      
      <form on:submit|preventDefault={updateDeployment} class="space-y-4">
        <div class="grid grid-cols-2 gap-4">
          <div>
            <Label for="edit-name">Naam</Label>
            <Input
              id="edit-name"
              bind:value={editForm.name}
              placeholder="Mijn App Productie"
              required
            />
          </div>
          <div>
            <Label for="edit-domain">Domein (optioneel)</Label>
            <Input
              id="edit-domain"
              bind:value={editForm.domain}
              placeholder="mijnapp.com"
            />
          </div>
        </div>

        <div>
          <Label for="edit-description">Beschrijving</Label>
          <Textarea
            id="edit-description"
            bind:value={editForm.description}
            placeholder="Productie deployment voor mijn applicatie"
            rows={2}
          />
        </div>

        <div class="grid grid-cols-2 gap-4">
          <div>
            <Label for="edit-repo">GitHub Repository</Label>
            <select id="edit-repo" bind:value={editForm.github_repository_id} required class="w-full px-3 py-2 border border-border rounded-md bg-background text-foreground focus:outline-none focus:ring-2 focus:ring-primary focus:border-transparent">
              <option value="">Selecteer repository...</option>
              {#each githubRepos as repo}
                <option value={repo.id.toString()}>{repo.name} ({repo.branch})</option>
              {/each}
            </select>
          </div>
          <div>
            <Label for="edit-server">Webserver</Label>
            <select id="edit-server" bind:value={editForm.web_server_id} required class="w-full px-3 py-2 border border-border rounded-md bg-background text-foreground focus:outline-none focus:ring-2 focus:ring-primary focus:border-transparent">
              <option value="">Selecteer server...</option>
              {#each webServers as server}
                <option value={server.id.toString()}>{server.name} ({server.hostname})</option>
              {/each}
            </select>
          </div>
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
            <Label for="edit-port">Port</Label>
            <Input
              id="edit-port"
              type="number"
              bind:value={editForm.port}
              min="1"
              max="65535"
            />
          </div>
          <div>
            <Label for="edit-subdomain">Subdomain (optioneel)</Label>
            <Input
              id="edit-subdomain"
              bind:value={editForm.subdomain}
              placeholder="app"
            />
          </div>
        </div>

        <div class="grid grid-cols-2 gap-4">
          <div>
            <Label for="edit-build">Build Command</Label>
            <Input
              id="edit-build"
              bind:value={editForm.build_command}
              placeholder="npm run build"
            />
          </div>
          <div>
            <Label for="edit-start">Start Command</Label>
            <Input
              id="edit-start"
              bind:value={editForm.start_command}
              placeholder="npm start"
            />
          </div>
        </div>

        <div class="flex items-center space-x-2">
          <input
            id="edit-auto-deploy"
            type="checkbox"
            bind:checked={editForm.is_auto_deploy_enabled}
            class="rounded border-border text-primary focus:ring-primary"
          />
          <Label for="edit-auto-deploy" class="text-sm cursor-pointer">
            Automatisch deployen bij push naar branch
          </Label>
        </div>

        <div class="flex justify-end space-x-2 pt-4">
          <Button type="button" variant="outline" on:click={() => showEditModal = false}>
            <Icon name="x" size={16} className="mr-2" />
            Annuleren
          </Button>
          <Button type="submit" class="bg-primary text-primary-foreground">
            <Icon name="save" size={16} className="mr-2" />
            Deployment Bijwerken
          </Button>
        </div>
      </form>
    </div>
  </Modal>
{/if}