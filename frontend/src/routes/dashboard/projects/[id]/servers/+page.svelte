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

  let projectId = $page.params.id;
  let webServers = [];
  let sshKeys = [];
  let loading = true;
  let showCreateModal = false;
  let showEditModal = false;
  let showDistributeKeyModal = false;
  let editingServer = null;
  let distributingServer = null;
  let testingConnection = {};
  let distributingKey = {};

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

  // Form data voor bewerken webserver
  let editForm = {
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

  async function updateWebServer() {
    try {
      const response = await fetch(`${API_BASE_URL}/api/v1/projects/${projectId}/web-servers/${editingServer.id}`, {
        method: 'PUT',
        headers: {
          'Content-Type': 'application/json',
          ...auth.getAuthHeader()
        },
        body: JSON.stringify({
          ...editForm,
          ssh_key_id: parseInt(editForm.ssh_key_id)
        })
      });

      if (response.ok) {
        toast.success('Webserver bijgewerkt');
        showEditModal = false;
        await loadData();
        resetEditForm();
      } else {
        const error = await response.json();
        toast.error(error.error || 'Fout bij bijwerken webserver');
      }
    } catch (error) {
      console.error('Error updating webserver:', error);
      toast.error('Netwerkfout bij bijwerken webserver');
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

  function editWebServer(server) {
    editingServer = server;
    editForm = {
      name: server.name,
      hostname: server.hostname,
      port: server.port,
      username: server.username,
      description: server.description || '',
      server_type: server.server_type,
      os: server.os,
      docker_enabled: server.docker_enabled,
      nginx_enabled: server.nginx_enabled,
      deploy_path: server.deploy_path,
      backup_path: server.backup_path,
      log_path: server.log_path,
      ssh_key_id: server.ssh_key_id.toString()
    };
    showEditModal = true;
  }

  function resetEditForm() {
    editForm = {
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
    editingServer = null;
  }

  // Key distribution variables
  let keyDistributionForm = {
    username: '',
    password: ''
  };

  function openDistributeKeyModal(server) {
    distributingServer = server;
    keyDistributionForm = {
      username: server.username || 'root',
      password: ''
    };
    showDistributeKeyModal = true;
  }

  async function distributePublicKey() {
    if (!distributingServer || !keyDistributionForm.username || !keyDistributionForm.password) {
      toast.error('Vul alle velden in');
      return;
    }

    distributingKey[distributingServer.id] = true;
    try {
      const response = await fetch(`${API_BASE_URL}/api/v1/projects/${projectId}/web-servers/${distributingServer.id}/distribute-key`, {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
          ...auth.getAuthHeader()
        },
        body: JSON.stringify({
          username: keyDistributionForm.username,
          password: keyDistributionForm.password
        })
      });

      const result = await response.json();
      
      if (response.ok) {
        toast.success('Public key succesvol gedistribueerd naar server');
        showDistributeKeyModal = false;
        keyDistributionForm = { username: '', password: '' };
        distributingServer = null;
      } else {
        toast.error(result.error || 'Fout bij distribueren van key');
      }
    } catch (error) {
      console.error('Error distributing key:', error);
      toast.error('Netwerkfout bij distribueren van key');
    } finally {
      distributingKey[distributingServer.id] = false;
    }
  }

  function getConnectionStatusColor(status: string) {
    switch (status) {
      case 'connected': return 'text-green-700 dark:text-green-300 bg-green-100 dark:bg-green-900 border-green-200 dark:border-green-800';
      case 'disconnected': return 'text-red-700 dark:text-red-300 bg-red-100 dark:bg-red-900 border-red-200 dark:border-red-800';
      case 'error': return 'text-orange-700 dark:text-orange-300 bg-orange-100 dark:bg-orange-900 border-orange-200 dark:border-orange-800';
      default: return 'text-gray-700 dark:text-gray-300 bg-gray-100 dark:bg-gray-900 border-gray-200 dark:border-gray-800';
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
      <Icon name="server" size={16} className="mr-2" />
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
          <Icon name="server" size={16} className="mr-2" />
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
                    <span class="px-2 py-1 text-xs font-medium rounded-full bg-blue-100 dark:bg-blue-900 border border-blue-200 dark:border-blue-800 text-blue-700 dark:text-blue-300">
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
                  class="border-blue-300 text-blue-600 hover:bg-blue-50 hover:border-blue-400"
                  title="Verbinding testen"
                >
                  <Icon name={testingConnection[server.id] ? "refresh" : "zap"} size={16} />
                </Button>
                <Button
                  on:click={() => openDistributeKeyModal(server)}
                  size="sm"
                  variant="outline"
                  disabled={distributingKey[server.id]}
                  class="border-purple-300 text-purple-600 hover:bg-purple-50 hover:border-purple-400"
                  title="Public key distribueren"
                >
                  <Icon name={distributingKey[server.id] ? "refresh" : "shield"} size={16} />
                </Button>
                <Button
                  on:click={() => editWebServer(server)}
                  size="sm"
                  variant="outline"
                  class="border-green-300 text-green-600 hover:bg-green-50 hover:border-green-400"
                  title="Server bewerken"
                >
                  <Icon name="edit" size={16} />
                </Button>
                <Button
                  on:click={() => deleteWebServer(server.id)}
                  size="sm"
                  variant="outline"
                  class="border-red-300 text-red-600 hover:bg-red-50 hover:border-red-400"
                  title="Server verwijderen"
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

<!-- Create WebServer Modal -->
{#if showCreateModal}
  <Modal open={showCreateModal} on:close={() => showCreateModal = false} size="2xl">
    <div class="p-8 max-h-[80vh] overflow-y-auto">
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

        <div class="grid grid-cols-1 md:grid-cols-3 gap-4">
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
            <select id="ssh_key" bind:value={serverForm.ssh_key_id} required class="w-full px-3 py-2 border border-border rounded-md bg-background text-foreground focus:outline-none focus:ring-2 focus:ring-primary focus:border-transparent">
              <option value="">Selecteer SSH key...</option>
              {#each sshKeys as key}
                <option value={key.id.toString()}>{key.name}</option>
              {/each}
            </select>
          </div>
        </div>

        <div class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
          <div>
            <Label for="server_type">Server Type</Label>
            <select id="server_type" bind:value={serverForm.server_type} class="w-full px-3 py-2 border border-border rounded-md bg-background text-foreground focus:outline-none focus:ring-2 focus:ring-primary focus:border-transparent">
              <option value="vps">VPS</option>
              <option value="dedicated">Dedicated</option>
              <option value="cloud">Cloud</option>
            </select>
          </div>
          <div>
            <Label for="os">Operating System</Label>
            <select id="os" bind:value={serverForm.os} class="w-full px-3 py-2 border border-border rounded-md bg-background text-foreground focus:outline-none focus:ring-2 focus:ring-primary focus:border-transparent">
              <option value="ubuntu">Ubuntu</option>
              <option value="debian">Debian</option>
              <option value="centos">CentOS</option>
              <option value="fedora">Fedora</option>
            </select>
          </div>
          <div class="flex flex-col space-y-2 col-span-full md:col-span-1">
            <Label>Services</Label>
            <div class="flex flex-wrap gap-4">
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

        <div class="grid grid-cols-1 md:grid-cols-3 gap-4">
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
            <Icon name="x" size={16} className="mr-2" />
            Annuleren
          </Button>
          <Button type="submit" class="bg-primary text-primary-foreground">
            <Icon name="server" size={16} className="mr-2" />
            Server Toevoegen
          </Button>
        </div>
      </form>
    </div>
  </Modal>
{/if}

<!-- Edit WebServer Modal -->
{#if showEditModal && editingServer}
  <Modal open={showEditModal} on:close={() => showEditModal = false} size="2xl">
    <div class="p-8 max-h-[80vh] overflow-y-auto">
      <h2 class="text-xl font-semibold mb-4">Webserver Bewerken: {editingServer.name}</h2>
      
      <form on:submit|preventDefault={updateWebServer} class="space-y-4">
        <div class="grid grid-cols-2 gap-4">
          <div>
            <Label for="edit-name">Naam</Label>
            <Input
              id="edit-name"
              bind:value={editForm.name}
              placeholder="Productie Server"
              required
            />
          </div>
          <div>
            <Label for="edit-hostname">Hostname/IP</Label>
            <Input
              id="edit-hostname"
              bind:value={editForm.hostname}
              placeholder="192.168.1.100 of server.example.com"
              required
            />
          </div>
        </div>

        <div>
          <Label for="edit-description">Beschrijving</Label>
          <Textarea
            id="edit-description"
            bind:value={editForm.description}
            placeholder="Productie server voor mijn applicatie"
            rows={2}
          />
        </div>

        <div class="grid grid-cols-1 md:grid-cols-3 gap-4">
          <div>
            <Label for="edit-port">SSH Port</Label>
            <Input
              id="edit-port"
              type="number"
              bind:value={editForm.port}
              min="1"
              max="65535"
            />
          </div>
          <div>
            <Label for="edit-username">SSH Gebruiker</Label>
            <Input
              id="edit-username"
              bind:value={editForm.username}
              placeholder="root"
              required
            />
          </div>
          <div>
            <Label for="edit-ssh_key">SSH Key</Label>
            <select id="edit-ssh_key" bind:value={editForm.ssh_key_id} required class="w-full px-3 py-2 border border-border rounded-md bg-background text-foreground focus:outline-none focus:ring-2 focus:ring-primary focus:border-transparent">
              <option value="">Selecteer SSH key...</option>
              {#each sshKeys as key}
                <option value={key.id.toString()}>{key.name}</option>
              {/each}
            </select>
          </div>
        </div>

        <div class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
          <div>
            <Label for="edit-server_type">Server Type</Label>
            <select id="edit-server_type" bind:value={editForm.server_type} class="w-full px-3 py-2 border border-border rounded-md bg-background text-foreground focus:outline-none focus:ring-2 focus:ring-primary focus:border-transparent">
              <option value="vps">VPS</option>
              <option value="dedicated">Dedicated</option>
              <option value="cloud">Cloud</option>
            </select>
          </div>
          <div>
            <Label for="edit-os">Operating System</Label>
            <select id="edit-os" bind:value={editForm.os} class="w-full px-3 py-2 border border-border rounded-md bg-background text-foreground focus:outline-none focus:ring-2 focus:ring-primary focus:border-transparent">
              <option value="ubuntu">Ubuntu</option>
              <option value="debian">Debian</option>
              <option value="centos">CentOS</option>
              <option value="fedora">Fedora</option>
            </select>
          </div>
          <div class="flex flex-col space-y-2 col-span-full md:col-span-1">
            <Label>Services</Label>
            <div class="flex flex-wrap gap-4">
              <label class="flex items-center space-x-2">
                <input
                  type="checkbox"
                  bind:checked={editForm.docker_enabled}
                  class="rounded border-border text-primary focus:ring-primary"
                />
                <span class="text-sm">Docker</span>
              </label>
              <label class="flex items-center space-x-2">
                <input
                  type="checkbox"
                  bind:checked={editForm.nginx_enabled}
                  class="rounded border-border text-primary focus:ring-primary"
                />
                <span class="text-sm">Nginx</span>
              </label>
            </div>
          </div>
        </div>

        <div class="grid grid-cols-1 md:grid-cols-3 gap-4">
          <div>
            <Label for="edit-deploy_path">Deploy Path</Label>
            <Input
              id="edit-deploy_path"
              bind:value={editForm.deploy_path}
              placeholder="/var/www"
            />
          </div>
          <div>
            <Label for="edit-backup_path">Backup Path</Label>
            <Input
              id="edit-backup_path"
              bind:value={editForm.backup_path}
              placeholder="/var/backups"
            />
          </div>
          <div>
            <Label for="edit-log_path">Log Path</Label>
            <Input
              id="edit-log_path"
              bind:value={editForm.log_path}
              placeholder="/var/log/deployments"
            />
          </div>
        </div>

        <div class="flex justify-end space-x-2 pt-4">
          <Button type="button" variant="outline" on:click={() => showEditModal = false}>
            <Icon name="x" size={16} className="mr-2" />
            Annuleren
          </Button>
          <Button type="submit" class="bg-primary text-primary-foreground">
            <Icon name="save" size={16} className="mr-2" />
            Server Bijwerken
          </Button>
        </div>
      </form>
    </div>
  </Modal>
{/if}

<!-- Distribute Public Key Modal -->
{#if showDistributeKeyModal && distributingServer}
  <Modal open={showDistributeKeyModal} on:close={() => showDistributeKeyModal = false} size="lg">
    <div class="p-8 max-h-[80vh] overflow-y-auto">
      <h2 class="text-xl font-semibold mb-4">Public Key Distribueren naar: {distributingServer.name}</h2>
      
      <form on:submit|preventDefault={distributePublicKey} class="space-y-4">
        <div class="bg-blue-50 dark:bg-blue-900/20 border border-blue-200 dark:border-blue-800 rounded-lg p-4 mb-4">
          <div class="flex items-start space-x-3">
            <Icon name="info" size={20} className="text-blue-600 dark:text-blue-400 mt-0.5" />
            <div>
              <h4 class="font-medium text-blue-900 dark:text-blue-200 mb-1">Public Key Distributie</h4>
              <p class="text-blue-800 dark:text-blue-200 text-sm">
                Deze functie installeert de public key van de geselecteerde SSH key op de server 
                zodat je zonder wachtwoord kunt inloggen. Voer je gebruikersnaam en wachtwoord in 
                voor eenmalige toegang.
              </p>
            </div>
          </div>
        </div>

        <div class="space-y-4">
          <div>
            <Label>Server Details</Label>
            <div class="grid grid-cols-2 gap-4 mt-2 p-3 bg-muted rounded-lg">
              <div>
                <span class="text-sm font-medium text-muted-foreground">Hostname:</span>
                <p class="text-sm font-mono">{distributingServer.hostname}:{distributingServer.port}</p>
              </div>
              <div>
                <span class="text-sm font-medium text-muted-foreground">SSH Key:</span>
                <p class="text-sm">{sshKeys.find(k => k.id === distributingServer.ssh_key_id)?.name || 'Onbekend'}</p>
              </div>
            </div>
          </div>

          <div>
            <Label for="dist-username">Gebruikersnaam</Label>
            <Input
              id="dist-username"
              type="text"
              bind:value={keyDistributionForm.username}
              placeholder="root"
              required
              class="mt-1"
            />
          </div>

          <div>
            <Label for="dist-password">Wachtwoord</Label>
            <Input
              id="dist-password"
              type="password"
              bind:value={keyDistributionForm.password}
              placeholder="Voer je wachtwoord in"
              required
              class="mt-1"
            />
          </div>
        </div>

        <div class="bg-yellow-50 dark:bg-yellow-900/20 border border-yellow-200 dark:border-yellow-800 rounded-lg p-4">
          <div class="flex items-start space-x-3">
            <Icon name="warning" size={20} className="text-yellow-600 dark:text-yellow-400 mt-0.5" />
            <div>
              <h4 class="font-medium text-yellow-900 dark:text-yellow-200 mb-1">Veiligheid</h4>
              <p class="text-yellow-800 dark:text-yellow-200 text-sm">
                Je wachtwoord wordt alleen gebruikt voor deze eenmalige installatie en wordt niet opgeslagen. 
                Na succesvolle installatie kun je inloggen met je SSH key zonder wachtwoord.
              </p>
            </div>
          </div>
        </div>

        <div class="flex justify-end space-x-2 pt-4">
          <Button 
            type="button" 
            variant="outline" 
            on:click={() => { showDistributeKeyModal = false; distributingServer = null; keyDistributionForm = { username: '', password: '' }; }}
          >
            <Icon name="x" size={16} className="mr-2" />
            Annuleren
          </Button>
          <Button 
            type="submit" 
            class="bg-primary text-primary-foreground"
            disabled={distributingKey[distributingServer?.id]}
          >
            <Icon name={distributingKey[distributingServer?.id] ? "refresh" : "shield"} size={16} className="mr-2" />
            {distributingKey[distributingServer?.id] ? 'Bezig...' : 'Key Distribueren'}
          </Button>
        </div>
      </form>
    </div>
  </Modal>
{/if}