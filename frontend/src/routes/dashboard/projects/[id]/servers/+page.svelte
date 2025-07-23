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
  import { API_BASE_URL, createApiRequest } from '$lib/config';
  import Select from '$lib/components/ui/select.svelte';
  import { API_BASE_URL, createApiRequest } from '$lib/config';
  import Textarea from '$lib/components/ui/textarea.svelte';
  import { API_BASE_URL, createApiRequest } from '$lib/config';

  let projectId = $page.params.id;
  let webServers = [];
  let sshKeys = [];
  let loading = true;
  let showCreateModal = false;
  let testingConnection = {};

  // Form data voor nieuwe webserver
  let serverForm = {
    name: '',
    hostname: '',
    port: 22,
    username: 'root',
    description: '',
    server_type: 'vps',
    os: 'ubuntu',
    docker_enabled: true,
    nginx_enabled: true,
    deploy_path: '/var/www',
    backup_path: '/var/backups',
    log_path: '/var/log/deployments',
    ssh_key_id: ''
  };

  onMount(async () => {
    await loadData();
  });

  async function loadData() {
    loading = true;
    try {
      const headers = auth.getAuthHeader();

      // Load web servers
      const serversRes = await fetch(`${API_BASE_URL}/api/v1/projects/${projectId}/web-servers`, { headers });
      if (serversRes.ok) {
        webServers = await serversRes.json();
      }

      // Load SSH keys
      const keysRes = await fetch(`${API_BASE_URL}/api/v1/projects/${projectId}/ssh-keys`, { headers });
      if (keysRes.ok) {
        sshKeys = await keysRes.json();
      }

    } catch (error) {
      console.error('Error loading data:', error);
      toast.error('Fout bij laden van server gegevens');
    } finally {
      loading = false;
    }
  }

  async function createWebServer() {
    try {
      const response = await fetch(`${API_BASE_URL}/api/v1/projects/${projectId}/web-servers`, {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
          ...auth.getAuthHeader()
        },
        body: JSON.stringify({
          ...serverForm,
          ssh_key_id: parseInt(serverForm.ssh_key_id)
        })
      });

      if (response.ok) {
        toast.success('Webserver toegevoegd');
        showCreateModal = false;
        await loadData();
        resetForm();
      } else {
        const error = await response.json();
        toast.error(error.error || 'Fout bij toevoegen webserver');
      }
    } catch (error) {
      console.error('Error creating webserver:', error);
      toast.error('Netwerkfout bij toevoegen webserver');
    }
  }

  async function testConnection(serverId: number) {
    testingConnection[serverId] = true;
    try {
      const response = await fetch(`${API_BASE_URL}/api/v1/projects/${projectId}/web-servers/${serverId}/test`, {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
          ...auth.getAuthHeader()
        }
      });

      const result = await response.json();
      
      if (response.ok && result.connected) {
        toast.success('Verbinding geslaagd!');
      } else {
        toast.error(result.error || 'Verbinding mislukt');
      }
      
      // Reload data to get updated connection status
      await loadData();
    } catch (error) {
      console.error('Error testing connection:', error);
      toast.error('Netwerkfout bij testen verbinding');
    } finally {
      testingConnection[serverId] = false;
    }
  }

  async function deleteWebServer(serverId: number) {
    if (!confirm('Weet je zeker dat je deze webserver wilt verwijderen?')) return;

    try {
      const response = await fetch(`${API_BASE_URL}/api/v1/projects/${projectId}/web-servers/${serverId}`, {
        method: 'DELETE',
        headers: auth.getAuthHeader()
      });

      if (response.ok) {
        toast.success('Webserver verwijderd');
        await loadData();
      } else {
        const error = await response.json();
        toast.error(error.error || 'Fout bij verwijderen webserver');
      }
    } catch (error) {
      console.error('Error deleting webserver:', error);
      toast.error('Netwerkfout bij verwijderen webserver');
    }
  }

  function resetForm() {
    serverForm = {
      name: '',
      hostname: '',
      port: 22,
      username: 'root',
      description: '',
      server_type: 'vps',
      os: 'ubuntu',
      docker_enabled: true,
      nginx_enabled: true,
      deploy_path: '/var/www',
      backup_path: '/var/backups',
      log_path: '/var/log/deployments',
      ssh_key_id: ''
    };
  }

  function getConnectionStatusColor(status: string) {
    switch (status) {
      case 'connected': return 'text-green-600 bg-green-50 border-green-200';
      case 'disconnected': return 'text-red-600 bg-red-50 border-red-200';
      case 'error': return 'text-orange-600 bg-orange-50 border-orange-200';
      default: return 'text-gray-600 bg-gray-50 border-gray-200';
    }
  }

  function getServerTypeIcon(type: string) {
    switch (type) {
      case 'vps': return '🖥️';
      case 'dedicated': return '🏢';
      case 'cloud': return '☁️';
      default: return '💻';
    }
  }
</script>

<svelte:head>
  <title>Webservers - CloudBox</title>
</svelte:head>

<div class="p-6">
  <div class="flex justify-between items-center mb-6">
    <div>
      <h1 class="text-3xl font-bold text-foreground">Webservers</h1>
      <p class="text-muted-foreground mt-1">Beheer je deployment servers en verbindingen</p>
    </div>
    <Button on:click={() => showCreateModal = true} class="bg-primary text-primary-foreground">
      <span class="mr-2">+</span>
      Server Toevoegen
    </Button>
  </div>

  {#if loading}
    <div class="text-center py-8">
      <div class="inline-block animate-spin rounded-full h-8 w-8 border-b-2 border-primary"></div>
      <p class="mt-2 text-muted-foreground">Laden...</p>
    </div>
  {:else}
    {#if webServers.length === 0}
      <Card class="p-8 text-center">
        <div class="text-6xl mb-4">🖥️</div>
        <h3 class="text-lg font-semibold mb-2">Nog geen webservers</h3>
        <p class="text-muted-foreground mb-4">Voeg je eerste webserver toe om deployments naar uit te kunnen voeren.</p>
        
        {#if sshKeys.length === 0}
          <div class="bg-yellow-50 border border-yellow-200 rounded-lg p-4 mb-4">
            <p class="text-yellow-800 text-sm">
              <strong>Eerst instellen:</strong> Je hebt een SSH key nodig om verbinding te maken met servers.
            </p>
            <div class="mt-2">
              <a href="/dashboard/projects/{projectId}/ssh-keys" class="text-blue-600 hover:text-blue-800 text-sm underline">
                SSH Key Genereren
              </a>
            </div>
          </div>
        {/if}

        <Button on:click={() => showCreateModal = true} disabled={sshKeys.length === 0}>
          Server Toevoegen
        </Button>
      </Card>
    {:else}
      <div class="grid gap-4">
        {#each webServers as server}
          <Card class="p-6">
            <div class="flex justify-between items-start">
              <div class="flex-1">
                <div class="flex items-center gap-3 mb-2">
                  <span class="text-2xl">{getServerTypeIcon(server.server_type)}</span>
                  <h3 class="text-lg font-semibold">{server.name}</h3>
                  <span class="px-2 py-1 text-xs font-medium rounded-full border {getConnectionStatusColor(server.connection_status)}">
                    {server.connection_status || 'unknown'}
                  </span>
                  {#if server.is_active}
                    <span class="px-2 py-1 text-xs font-medium rounded-full bg-blue-50 border border-blue-200 text-blue-600">
                      Actief
                    </span>
                  {/if}
                </div>
                <p class="text-muted-foreground text-sm mb-3">{server.description || 'Geen beschrijving'}</p>
                
                <div class="grid grid-cols-2 md:grid-cols-4 gap-4 text-sm">
                  <div>
                    <span class="font-medium text-muted-foreground">Hostname:</span>
                    <p class="font-mono">{server.hostname}:{server.port}</p>
                  </div>
                  <div>
                    <span class="font-medium text-muted-foreground">Gebruiker:</span>
                    <p>{server.username}</p>
                  </div>
                  <div>
                    <span class="font-medium text-muted-foreground">OS:</span>
                    <p class="capitalize">{server.os}</p>
                  </div>
                  <div>
                    <span class="font-medium text-muted-foreground">Deploy Path:</span>
                    <p class="font-mono text-xs">{server.deploy_path}</p>
                  </div>
                </div>

                <div class="mt-3 flex gap-4 text-sm">
                  {#if server.docker_enabled}
                    <span class="text-blue-600">🐳 Docker</span>
                  {/if}
                  {#if server.nginx_enabled}
                    <span class="text-green-600">🌐 Nginx</span>
                  {/if}
                </div>

                {#if server.last_connected_at}
                  <div class="mt-2 text-sm text-muted-foreground">
                    Laatst verbonden: {new Date(server.last_connected_at).toLocaleString('nl-NL')}
                  </div>
                {/if}
              </div>

              <div class="flex gap-2 ml-4">
                <Button
                  on:click={() => testConnection(server.id)}
                  size="sm"
                  variant="outline"
                  disabled={testingConnection[server.id]}
                  class="border-blue-300 text-blue-600 hover:bg-blue-50"
                >
                  {testingConnection[server.id] ? 'Testen...' : 'Test Verbinding'}
                </Button>
                <Button
                  on:click={() => deleteWebServer(server.id)}
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

<!-- Create WebServer Modal -->
{#if showCreateModal}
  <Modal on:close={() => showCreateModal = false}>
    <div class="p-6 max-h-[80vh] overflow-y-auto">
      <h2 class="text-xl font-semibold mb-4">Webserver Toevoegen</h2>
      
      <form on:submit|preventDefault={createWebServer} class="space-y-4">
        <div class="grid grid-cols-2 gap-4">
          <div>
            <Label for="name">Naam</Label>
            <Input
              id="name"
              bind:value={serverForm.name}
              placeholder="Productie Server"
              required
            />
          </div>
          <div>
            <Label for="hostname">Hostname/IP</Label>
            <Input
              id="hostname"
              bind:value={serverForm.hostname}
              placeholder="192.168.1.100 of server.example.com"
              required
            />
          </div>
        </div>

        <div>
          <Label for="description">Beschrijving</Label>
          <Textarea
            id="description"
            bind:value={serverForm.description}
            placeholder="Productie server voor mijn applicatie"
            rows={2}
          />
        </div>

        <div class="grid grid-cols-3 gap-4">
          <div>
            <Label for="port">SSH Port</Label>
            <Input
              id="port"
              type="number"
              bind:value={serverForm.port}
              min="1"
              max="65535"
            />
          </div>
          <div>
            <Label for="username">SSH Gebruiker</Label>
            <Input
              id="username"
              bind:value={serverForm.username}
              placeholder="root"
              required
            />
          </div>
          <div>
            <Label for="ssh_key">SSH Key</Label>
            <Select id="ssh_key" bind:value={serverForm.ssh_key_id} required>
              <option value="">Selecteer SSH key...</option>
              {#each sshKeys as key}
                <option value={key.id}>{key.name}</option>
              {/each}
            </Select>
          </div>
        </div>

        <div class="grid grid-cols-3 gap-4">
          <div>
            <Label for="server_type">Server Type</Label>
            <Select id="server_type" bind:value={serverForm.server_type}>
              <option value="vps">VPS</option>
              <option value="dedicated">Dedicated</option>
              <option value="cloud">Cloud</option>
            </Select>
          </div>
          <div>
            <Label for="os">Operating System</Label>
            <Select id="os" bind:value={serverForm.os}>
              <option value="ubuntu">Ubuntu</option>
              <option value="debian">Debian</option>
              <option value="centos">CentOS</option>
              <option value="fedora">Fedora</option>
            </Select>
          </div>
          <div class="flex flex-col space-y-2">
            <Label>Services</Label>
            <div class="flex space-x-4">
              <label class="flex items-center space-x-2">
                <input
                  type="checkbox"
                  bind:checked={serverForm.docker_enabled}
                  class="rounded border-border text-primary focus:ring-primary"
                />
                <span class="text-sm">Docker</span>
              </label>
              <label class="flex items-center space-x-2">
                <input
                  type="checkbox"
                  bind:checked={serverForm.nginx_enabled}
                  class="rounded border-border text-primary focus:ring-primary"
                />
                <span class="text-sm">Nginx</span>
              </label>
            </div>
          </div>
        </div>

        <div class="grid grid-cols-3 gap-4">
          <div>
            <Label for="deploy_path">Deploy Path</Label>
            <Input
              id="deploy_path"
              bind:value={serverForm.deploy_path}
              placeholder="/var/www"
            />
          </div>
          <div>
            <Label for="backup_path">Backup Path</Label>
            <Input
              id="backup_path"
              bind:value={serverForm.backup_path}
              placeholder="/var/backups"
            />
          </div>
          <div>
            <Label for="log_path">Log Path</Label>
            <Input
              id="log_path"
              bind:value={serverForm.log_path}
              placeholder="/var/log/deployments"
            />
          </div>
        </div>

        <div class="bg-blue-50 border border-blue-200 rounded-lg p-4">
          <p class="text-blue-800 text-sm">
            <strong>Let op:</strong> Zorg ervoor dat de SSH key is geïnstalleerd op de server voordat je een verbinding test.
          </p>
        </div>

        <div class="flex justify-end space-x-2 pt-4">
          <Button type="button" variant="outline" on:click={() => showCreateModal = false}>
            Annuleren
          </Button>
          <Button type="submit" class="bg-primary text-primary-foreground">
            Server Toevoegen
          </Button>
        </div>
      </form>
    </div>
  </Modal>
{/if}