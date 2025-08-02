<script lang="ts">
  import { onMount } from 'svelte';
  import { page } from '$app/stores';
  import { goto } from '$app/navigation';
  import Button from '$lib/components/ui/button/Button.svelte';
  import Card from '$lib/components/ui/card/Card.svelte';
  import Input from '$lib/components/ui/input/Input.svelte';
  import Select from '$lib/components/ui/select/Select.svelte';
  import Textarea from '$lib/components/ui/textarea/Textarea.svelte';
  import Label from '$lib/components/ui/label/Label.svelte';
  import Icon from '$lib/components/Icon.svelte';
  import { auth } from '$lib/stores/auth';
  import { toast } from '$lib/stores/toast';
  import { API_BASE_URL } from '$lib/config';

  let projectId: string;
  let repoId: string | null = null;
  let repoName: string = '';
  let installOption: string = '';
  
  let deployment = {
    name: '',
    description: '',
    repository_url: '',
    branch: 'main',
    install_option: '',
    ssh_key_id: null,
    web_server_id: null,
    deployment_path: '/var/www/',
    port: 3000,
    environment: {}
  };

  let sshKeys: any[] = [];
  let webServers: any[] = [];
  let installOptions: any[] = [];
  let repositoryAnalysis: any = null;
  let loading = false;

  $: projectId = $page.params.id;

  onMount(async () => {
    // Get URL parameters
    const urlParams = $page.url.searchParams;
    repoId = urlParams.get('repo_id');
    repoName = urlParams.get('repo_name') || '';
    installOption = urlParams.get('install_option') || '';

    if (repoId) {
      deployment.name = `${repoName} Deployment`;
      await loadRepositoryAnalysis();
    }

    await loadSSHKeys();
    await loadWebServers();
  });

  async function loadRepositoryAnalysis() {
    if (!repoId) return;
    
    try {
      const response = await fetch(`${API_BASE_URL}/api/v1/projects/${projectId}/github-repositories/${repoId}/analysis`, {
        headers: auth.getAuthHeader()
      });

      if (response.ok) {
        const data = await response.json();
        repositoryAnalysis = data.analysis;
        installOptions = data.analysis.install_options || [];
        
        if (installOption && installOptions.length > 0) {
          const selectedOption = installOptions.find(opt => opt.name === installOption);
          if (selectedOption) {
            deployment.install_option = selectedOption.name;
            deployment.port = selectedOption.port;
            deployment.environment = selectedOption.environment || {};
          }
        }
      }
    } catch (error) {
      console.error('Error loading repository analysis:', error);
    }
  }

  async function loadSSHKeys() {
    try {
      const response = await fetch(`${API_BASE_URL}/api/v1/projects/${projectId}/ssh-keys`, {
        headers: auth.getAuthHeader()
      });

      if (response.ok) {
        const data = await response.json();
        sshKeys = data.ssh_keys || [];
      }
    } catch (error) {
      console.error('Error loading SSH keys:', error);
    }
  }

  async function loadWebServers() {
    try {
      const response = await fetch(`${API_BASE_URL}/api/v1/projects/${projectId}/web-servers`, {
        headers: auth.getAuthHeader()
      });

      if (response.ok) {
        const data = await response.json();
        webServers = data.web_servers || [];
      }
    } catch (error) {
      console.error('Error loading web servers:', error);
    }
  }

  async function createDeployment() {
    loading = true;
    
    try {
      const response = await fetch(`${API_BASE_URL}/api/v1/projects/${projectId}/deployments`, {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
          ...auth.getAuthHeader()
        },
        body: JSON.stringify({
          ...deployment,
          github_repository_id: repoId ? parseInt(repoId) : null
        })
      });

      if (response.ok) {
        toast.success('Deployment succesvol aangemaakt!');
        goto(`/dashboard/projects/${projectId}/deployments`);
      } else {
        const error = await response.json();
        toast.error(error.error || 'Fout bij aanmaken deployment');
      }
    } catch (error) {
      console.error('Error creating deployment:', error);
      toast.error('Netwerkfout bij aanmaken deployment');
    } finally {
      loading = false;
    }
  }

  function goBack() {
    goto(`/dashboard/projects/${projectId}/deployments`);
  }
</script>

<svelte:head>
  <title>Nieuwe Deployment - CloudBox</title>
</svelte:head>

<div class="container mx-auto px-6 py-8">
  <div class="flex items-center gap-4 mb-6">
    <Button variant="ghost" on:click={goBack} class="p-2">
      <Icon name="arrow-left" size={20} />
    </Button>
    <div>
      <h1 class="text-3xl font-bold">Nieuwe Deployment</h1>
      <p class="text-muted-foreground">Maak een nieuwe deployment aan voor {repoName || 'je repository'}</p>
    </div>
  </div>

  <div class="grid gap-6 max-w-4xl">
    <!-- Basis informatie -->
    <Card class="p-6">
      <h2 class="text-xl font-semibold mb-4">Basis Informatie</h2>
      <div class="grid gap-4">
        <div>
          <Label for="name">Deployment Naam</Label>
          <Input 
            id="name" 
            bind:value={deployment.name} 
            placeholder="Bijv. My App Production"
            required
          />
        </div>
        
        <div>
          <Label for="description">Beschrijving (optioneel)</Label>
          <Textarea 
            id="description" 
            bind:value={deployment.description} 
            placeholder="Beschrijf deze deployment..."
            rows={3}
          />
        </div>
      </div>
    </Card>

    <!-- Repository informatie -->
    {#if repositoryAnalysis}
      <Card class="p-6">
        <h2 class="text-xl font-semibold mb-4">Repository & Deployment</h2>
        <div class="grid gap-4">
          <div>
            <Label>Gedetecteerd Project Type</Label>
            <div class="flex gap-2 items-center text-sm text-muted-foreground">
              <span class="font-medium">{repositoryAnalysis.project_type}</span>
              {#if repositoryAnalysis.framework}
                <span>• {repositoryAnalysis.framework}</span>
              {/if}
              {#if repositoryAnalysis.language}
                <span>• {repositoryAnalysis.language}</span>
              {/if}
            </div>
          </div>

          <div>
            <Label for="install_option">Deployment Methode</Label>
            <Select id="install_option" bind:value={deployment.install_option}>
              <option value="">Selecteer deployment methode...</option>
              {#each installOptions as option}
                <option value={option.name} selected={option.is_recommended}>
                  {option.name} {option.is_recommended ? '(aanbevolen)' : ''} - {option.description}
                </option>
              {/each}
            </Select>
          </div>

          <div>
            <Label for="branch">Branch</Label>
            <Input 
              id="branch" 
              bind:value={deployment.branch} 
              placeholder="main"
            />
          </div>

          <div>
            <Label for="port">Applicatie Port</Label>
            <Input 
              id="port" 
              type="number" 
              bind:value={deployment.port} 
              placeholder="3000"
            />
          </div>
        </div>
      </Card>
    {/if}

    <!-- Server configuratie -->
    <Card class="p-6">
      <h2 class="text-xl font-semibold mb-4">Server Configuratie</h2>
      <div class="grid gap-4">
        <div>
          <Label for="web_server">Doelserver</Label>
          <Select id="web_server" bind:value={deployment.web_server_id}>
            <option value={null}>Selecteer een server...</option>
            {#each webServers as server}
              <option value={server.id}>
                {server.name} ({server.host})
              </option>
            {/each}
          </Select>
          {#if webServers.length === 0}
            <p class="text-sm text-muted-foreground mt-1">
              Geen servers beschikbaar. 
              <a href="/dashboard/projects/{projectId}/servers" class="text-blue-600 hover:underline">
                Voeg eerst een server toe
              </a>
            </p>
          {/if}
        </div>

        <div>
          <Label for="ssh_key">SSH Key voor toegang</Label>
          <Select id="ssh_key" bind:value={deployment.ssh_key_id}>
            <option value={null}>Selecteer een SSH key...</option>
            {#each sshKeys as key}
              <option value={key.id}>
                {key.name}
              </option>
            {/each}
          </Select>
          {#if sshKeys.length === 0}
            <p class="text-sm text-muted-foreground mt-1">
              Geen SSH keys beschikbaar. 
              <a href="/dashboard/projects/{projectId}/ssh-keys" class="text-blue-600 hover:underline">
                Voeg eerst een SSH key toe
              </a>
            </p>
          {/if}
        </div>

        <div>
          <Label for="deployment_path">Deployment Pad op Server</Label>
          <Input 
            id="deployment_path" 
            bind:value={deployment.deployment_path} 
            placeholder="/var/www/myapp"
          />
        </div>
      </div>
    </Card>

    <!-- Acties -->
    <div class="flex gap-4">
      <Button on:click={createDeployment} disabled={loading} class="min-w-[120px]">
        {#if loading}
          <div class="animate-spin rounded-full h-4 w-4 border-b-2 border-white mr-2"></div>
          Aanmaken...
        {:else}
          <Icon name="rocket" size={16} class="mr-2" />
          Deployment Aanmaken
        {/if}
      </Button>
      
      <Button variant="outline" on:click={goBack}>
        Annuleren
      </Button>
    </div>
  </div>
</div>