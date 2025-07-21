<script lang="ts">
  import { page } from '$app/stores';
  import { onMount } from 'svelte';
  import { auth } from '$lib/stores/auth';
  import { toast } from '$lib/stores/toast';
  import Card from '$lib/components/ui/card.svelte';
  import Button from '$lib/components/ui/button.svelte';
  import Modal from '$lib/components/ui/modal.svelte';
  import Input from '$lib/components/ui/input.svelte';
  import Label from '$lib/components/ui/label.svelte';
  import Select from '$lib/components/ui/select.svelte';
  import Textarea from '$lib/components/ui/textarea.svelte';

  let projectId = $page.params.id;
  let deployments = [];
  let githubRepos = [];
  let webServers = [];
  let loading = true;
  let showCreateModal = false;

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
      const deploymentsRes = await fetch(`http://localhost:8080/api/v1/projects/${projectId}/deployments`, { headers });
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
      const reposRes = await fetch(`http://localhost:8080/api/v1/projects/${projectId}/github-repositories`, { headers });
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
      const serversRes = await fetch(`http://localhost:8080/api/v1/projects/${projectId}/web-servers`, { headers });
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
      const response = await fetch(`http://localhost:8080/api/v1/projects/${projectId}/deployments`, {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
          ...auth.getAuthHeader()
        },
        body: JSON.stringify(deploymentForm)
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
      const response = await fetch(`http://localhost:8080/api/v1/projects/${projectId}/deployments/${deploymentId}/deploy`, {
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
        toast.success('Deployment gestart');
        await loadData();
      } else {
        const error = await response.json();
        toast.error(error.error || 'Fout bij starten deployment');
      }
    } catch (error) {
      console.error('Error deploying:', error);
      toast.error('Netwerkfout bij deployment');
    }
  }

  async function deleteDeployment(deploymentId: number) {
    if (!confirm('Weet je zeker dat je deze deployment wilt verwijderen?')) return;

    try {
      const response = await fetch(`http://localhost:8080/api/v1/projects/${projectId}/deployments/${deploymentId}`, {
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

  function getStatusColor(status: string) {
    switch (status) {
      case 'deployed': return 'text-green-600 bg-green-50 border-green-200';
      case 'deploying': return 'text-blue-600 bg-blue-50 border-blue-200';
      case 'building': return 'text-yellow-600 bg-yellow-50 border-yellow-200';
      case 'failed': return 'text-red-600 bg-red-50 border-red-200';
      case 'pending': return 'text-gray-600 bg-gray-50 border-gray-200';
      default: return 'text-gray-600 bg-gray-50 border-gray-200';
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
      <span class="mr-2">+</span>
      Nieuwe Deployment
    </Button>
  </div>

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
          Deployment Aanmaken
        </Button>
      </Card>
    {:else}
      <div class="grid gap-4">
        {#each deployments as deployment}
          <Card class="p-6">
            <div class="flex justify-between items-start">
              <div class="flex-1">
                <div class="flex items-center gap-3 mb-2">
                  <h3 class="text-lg font-semibold">{deployment.name}</h3>
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
                    <a href="http://{deployment.domain}" target="_blank" class="text-blue-600 hover:text-blue-800 text-sm ml-1">
                      {deployment.domain}
                    </a>
                  </div>
                {/if}
              </div>

              <div class="flex gap-2 ml-4">
                <Button
                  on:click={() => deploy(deployment.id)}
                  size="sm"
                  disabled={deployment.status === 'deploying' || deployment.status === 'building'}
                  class="bg-green-600 text-white hover:bg-green-700"
                >
                  {deployment.status === 'deploying' || deployment.status === 'building' ? 'Bezig...' : 'Deploy'}
                </Button>
                <Button
                  on:click={() => deleteDeployment(deployment.id)}
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

<!-- Create Deployment Modal -->
{#if showCreateModal}
  <Modal on:close={() => showCreateModal = false}>
    <div class="p-6">
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
            <Select id="repo" bind:value={deploymentForm.github_repository_id} required>
              <option value="">Selecteer repository...</option>
              {#each githubRepos as repo}
                <option value={repo.id}>{repo.name} ({repo.branch})</option>
              {/each}
            </Select>
          </div>
          <div>
            <Label for="server">Webserver</Label>
            <Select id="server" bind:value={deploymentForm.web_server_id} required>
              <option value="">Selecteer server...</option>
              {#each webServers as server}
                <option value={server.id}>{server.name} ({server.hostname})</option>
              {/each}
            </Select>
          </div>
        </div>

        <div class="grid grid-cols-3 gap-4">
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
            Annuleren
          </Button>
          <Button type="submit" class="bg-primary text-primary-foreground">
            Deployment Aanmaken
          </Button>
        </div>
      </form>
    </div>
  </Modal>
{/if}