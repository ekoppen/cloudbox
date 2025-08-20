<script lang="ts">
  import { onMount } from 'svelte';
  import { page } from '$app/stores';
  import { goto } from '$app/navigation';
  import Button from '$lib/components/ui/button.svelte';
  import Card from '$lib/components/ui/card.svelte';
  import Input from '$lib/components/ui/input.svelte';
  import Textarea from '$lib/components/ui/textarea.svelte';
  import Label from '$lib/components/ui/label.svelte';
  import Icon from '$lib/components/ui/icon.svelte';
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
    ssh_key_id: '',
    web_server_id: '',
    deployment_path: '~/deploys/',
    port: 3000,
    environment: {},
    port_configuration: {}
  };

  let sshKeys: any[] = [];
  let webServers: any[] = [];
  let installOptions: any[] = [];
  let repositoryAnalysis: any = null;
  let loading = false;
  
  // Port configuration
  let selectedInstallOption: any = null;
  let portRequirements: any[] = [];
  let portAvailability: {[key: number]: boolean} = {};
  let checkingPorts = false;

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
        sshKeys = Array.isArray(data) ? data : (data.ssh_keys || []);
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
        webServers = Array.isArray(data) ? data : (data.web_servers || []);
      }
    } catch (error) {
      console.error('Error loading web servers:', error);
    }
  }

  async function createDeployment() {
    loading = true;
    
    try {
      const deploymentData = {
        ...deployment,
        ssh_key_id: deployment.ssh_key_id ? parseInt(deployment.ssh_key_id) : null,
        web_server_id: deployment.web_server_id ? parseInt(deployment.web_server_id) : null,
        github_repository_id: repoId ? parseInt(repoId) : null,
        port_configuration: deployment.port_configuration || {}
      };

      const response = await fetch(`${API_BASE_URL}/api/v1/projects/${projectId}/deployments`, {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
          ...auth.getAuthHeader()
        },
        body: JSON.stringify(deploymentData)
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

  // Watch for install option changes to update port requirements
  $: if (deployment.install_option && installOptions.length > 0) {
    selectedInstallOption = installOptions.find(opt => opt.name === deployment.install_option);
    if (selectedInstallOption?.port_requirements) {
      portRequirements = selectedInstallOption.port_requirements;
      // Initialize port configuration with defaults
      portRequirements.forEach(req => {
        if (!deployment.port_configuration[req.variable]) {
          deployment.port_configuration[req.variable] = req.port;
        }
      });
    } else {
      portRequirements = [];
    }
  }

  // Check port availability when server changes
  $: if (deployment.web_server_id && portRequirements.length > 0) {
    checkPortAvailability();
  }

  // Debounced port checking
  let portCheckTimeout: NodeJS.Timeout;
  function debouncedPortCheck() {
    if (portCheckTimeout) clearTimeout(portCheckTimeout);
    portCheckTimeout = setTimeout(() => {
      checkPortAvailability();
    }, 1000); // Wait 1 second after user stops typing
  }

  async function checkPortAvailability() {
    if (!deployment.web_server_id || portRequirements.length === 0) return;
    
    checkingPorts = true;
    const ports = portRequirements.map(req => deployment.port_configuration[req.variable] || req.port);
    
    try {
      const response = await fetch(`${API_BASE_URL}/api/v1/projects/${projectId}/deployments/check-ports`, {
        method: 'POST',
        headers: {
          ...auth.getAuthHeader(),
          'Content-Type': 'application/json',
        },
        body: JSON.stringify({
          web_server_id: parseInt(deployment.web_server_id),
          ports: ports
        })
      });

      if (response.ok) {
        const result = await response.json();
        portAvailability = result.port_status;
      } else {
        console.error('Failed to check port availability');
        portAvailability = {};
      }
    } catch (error) {
      console.error('Error checking port availability:', error);
      portAvailability = {};
    } finally {
      checkingPorts = false;
    }
  }

  function getPortStatus(port: number): 'available' | 'unavailable' | 'unknown' {
    if (checkingPorts) return 'unknown';
    if (portAvailability.hasOwnProperty(port)) {
      return portAvailability[port] ? 'available' : 'unavailable';
    }
    return 'unknown';
  }

  function hasPortConflicts(): boolean {
    return portRequirements.some(req => {
      const currentPort = deployment.port_configuration[req.variable] || req.port;
      return getPortStatus(currentPort) === 'unavailable';
    });
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
            <select 
              id="install_option" 
              bind:value={deployment.install_option}
              class="flex h-10 w-full rounded-md border border-input bg-background px-3 py-2 text-sm ring-offset-background focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-2 disabled:cursor-not-allowed disabled:opacity-50"
            >
              <option value="">Selecteer deployment methode...</option>
              {#each installOptions as option}
                <option value={option.name} selected={option.is_recommended}>
                  {option.name} {option.is_recommended ? '(aanbevolen)' : ''} - {option.description}
                </option>
              {/each}
            </select>
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
          <select 
            id="web_server" 
            bind:value={deployment.web_server_id}
            class="flex h-10 w-full rounded-md border border-input bg-background px-3 py-2 text-sm ring-offset-background focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-2 disabled:cursor-not-allowed disabled:opacity-50"
          >
            <option value="">Selecteer een server...</option>
            {#each webServers as server}
              <option value={server.id}>
                {server.name}{#if server.host} ({server.host}){/if}
              </option>
            {/each}
          </select>
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
          <select 
            id="ssh_key" 
            bind:value={deployment.ssh_key_id}
            class="flex h-10 w-full rounded-md border border-input bg-background px-3 py-2 text-sm ring-offset-background focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-2 disabled:cursor-not-allowed disabled:opacity-50"
          >
            <option value="">Selecteer een SSH key...</option>
            {#each sshKeys as key}
              <option value={key.id}>
                {key.name}
              </option>
            {/each}
          </select>
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

    <!-- Port Configuratie -->
    {#if portRequirements.length > 0}
      <Card class="p-6">
        <h2 class="text-xl font-semibold mb-4">Port Configuratie</h2>
        <div class="grid gap-4">
          <p class="text-sm text-muted-foreground">
            Deze deployment heeft de volgende poorten nodig. Controleer of deze beschikbaar zijn op je server.
          </p>
          
          {#each portRequirements as req, index}
            <div class="border rounded-lg p-4">
              <div class="flex items-center justify-between mb-2">
                <div>
                  <Label class="font-medium">{req.name}</Label>
                  {#if req.description}
                    <p class="text-sm text-muted-foreground">{req.description}</p>
                  {/if}
                </div>
                <div class="flex items-center gap-2">
                  {#if checkingPorts}
                    <div class="animate-spin rounded-full h-4 w-4 border-b-2 border-blue-500"></div>
                    <span class="text-sm text-muted-foreground">Checking...</span>
                  {:else}
                    {@const currentPort = deployment.port_configuration[req.variable] || req.port}
                    {@const status = getPortStatus(currentPort)}
                    {#if status === 'available'}
                      <Icon name="check-circle" size={16} class="text-green-500" />
                      <span class="text-sm text-green-600">Beschikbaar</span>
                    {:else if status === 'unavailable'}
                      <Icon name="x-circle" size={16} class="text-red-500" />
                      <span class="text-sm text-red-600">In gebruik</span>
                    {:else}
                      <Icon name="help-circle" size={16} class="text-gray-400" />
                      <span class="text-sm text-muted-foreground">Onbekend</span>
                    {/if}
                  {/if}
                </div>
              </div>
              
              <div class="flex items-center gap-2">
                <Label for="port_{req.variable}" class="text-sm">Port:</Label>
                <Input 
                  id="port_{req.variable}"
                  type="number" 
                  bind:value={deployment.port_configuration[req.variable]}
                  placeholder={req.port.toString()}
                  class="w-24"
                  on:input={debouncedPortCheck}
                />
                <span class="text-sm text-muted-foreground">
                  (default: {req.port})
                </span>
                {#if req.required}
                  <span class="text-xs bg-red-100 text-red-800 px-2 py-1 rounded">Verplicht</span>
                {/if}
              </div>
              
              {#if req.variable}
                <p class="text-xs text-muted-foreground mt-1">
                  Environment variabele: {req.variable}
                </p>
              {/if}
            </div>
          {/each}
          
          {#if deployment.web_server_id && !checkingPorts}
            <div class="flex items-center gap-2 mt-2">
              <Button 
                variant="ghost"
                size="icon"
                on:click={checkPortAvailability}
                disabled={checkingPorts}
                class="hover:rotate-180 transition-transform duration-300"
                title="Hercontroleer poorten beschikbaarheid"
              >
                <Icon name="refresh-cw" size={16} class={checkingPorts ? "animate-spin" : ""} />
              </Button>
              <span class="text-sm text-muted-foreground">
                Laatste check: {new Date().toLocaleTimeString()}
              </span>
            </div>
            
            {#if hasPortConflicts()}
              <div class="bg-red-50 border border-red-200 rounded-lg p-3 mt-2">
                <div class="flex items-center gap-2">
                  <Icon name="alert-circle" size={16} class="text-red-600" />
                  <span class="text-sm text-red-800">
                    <strong>Port conflict gedetecteerd!</strong> Sommige poorten zijn al in gebruik. 
                    Pas de port nummers aan of stop de services die deze poorten gebruiken.
                  </span>
                </div>
              </div>
            {/if}
          {:else if !deployment.web_server_id}
            <div class="bg-yellow-50 border border-yellow-200 rounded-lg p-3">
              <div class="flex items-center gap-2">
                <Icon name="alert-triangle" size={16} class="text-yellow-600" />
                <span class="text-sm text-yellow-800">
                  Selecteer eerst een server om port beschikbaarheid te controleren
                </span>
              </div>
            </div>
          {/if}
        </div>
      </Card>
    {/if}

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
      
      {#if hasPortConflicts()}
        <div class="flex items-center gap-2 text-sm text-orange-600">
          <Icon name="alert-triangle" size={16} />
          <span>Waarschuwing: Port conflicten gedetecteerd</span>
        </div>
      {/if}
      
      <Button variant="outline" on:click={goBack}>
        Annuleren
      </Button>
    </div>
  </div>
</div>