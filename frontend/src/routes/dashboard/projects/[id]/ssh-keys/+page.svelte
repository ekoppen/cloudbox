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
  let sshKeys = [];
  let loading = true;
  let showCreateModal = false;
  let showKeyModal = false;
  let selectedKey = null;

  // Form data voor nieuwe SSH key
  let keyForm = {
    name: '',
    description: '',
    key_type: 'rsa',
    key_size: 2048
  };

  onMount(async () => {
    await loadSSHKeys();
  });

  async function loadSSHKeys() {
    loading = true;
    try {
      const response = await fetch(`${API_BASE_URL}/api/v1/projects/${projectId}/ssh-keys`, {
        headers: auth.getAuthHeader()
      });

      if (response.ok) {
        sshKeys = await response.json();
      } else {
        toast.error('Fout bij laden SSH keys');
      }
    } catch (error) {
      console.error('Error loading SSH keys:', error);
      toast.error('Netwerkfout bij laden SSH keys');
    } finally {
      loading = false;
    }
  }

  async function createSSHKey() {
    try {
      const response = await fetch(`${API_BASE_URL}/api/v1/projects/${projectId}/ssh-keys`, {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
          ...auth.getAuthHeader()
        },
        body: JSON.stringify(keyForm)
      });

      if (response.ok) {
        const newKey = await response.json();
        toast.success('SSH key aangemaakt');
        showCreateModal = false;
        selectedKey = newKey;
        showKeyModal = true;
        await loadSSHKeys();
        resetForm();
      } else {
        const error = await response.json();
        toast.error(error.error || 'Fout bij aanmaken SSH key');
      }
    } catch (error) {
      console.error('Error creating SSH key:', error);
      toast.error('Netwerkfout bij aanmaken SSH key');
    }
  }

  async function deleteSSHKey(keyId: number) {
    if (!confirm('Weet je zeker dat je deze SSH key wilt verwijderen?')) return;

    try {
      const response = await fetch(`${API_BASE_URL}/api/v1/projects/${projectId}/ssh-keys/${keyId}`, {
        method: 'DELETE',
        headers: auth.getAuthHeader()
      });

      if (response.ok) {
        toast.success('SSH key verwijderd');
        await loadSSHKeys();
      } else {
        const error = await response.json();
        toast.error(error.error || 'Fout bij verwijderen SSH key');
      }
    } catch (error) {
      console.error('Error deleting SSH key:', error);
      toast.error('Netwerkfout bij verwijderen SSH key');
    }
  }

  function showPublicKey(key) {
    selectedKey = key;
    showKeyModal = true;
  }

  function copyToClipboard(text: string) {
    navigator.clipboard.writeText(text).then(() => {
      toast.success('Gekopieerd naar klembord');
    }).catch(() => {
      toast.error('Kon niet kopiëren naar klembord');
    });
  }

  function resetForm() {
    keyForm = {
      name: '',
      description: '',
      key_type: 'rsa',
      key_size: 2048
    };
  }

  function formatFingerprint(fingerprint: string) {
    return fingerprint.replace(/(.{2})/g, '$1:').slice(0, -1);
  }
</script>

<svelte:head>
  <title>SSH Keys - CloudBox</title>
</svelte:head>

<div class="p-6">
  <div class="flex justify-between items-center mb-6">
    <div>
      <h1 class="text-3xl font-bold text-foreground">SSH Keys</h1>
      <p class="text-muted-foreground mt-1">Beheer SSH keys voor veilige verbindingen met je servers</p>
    </div>
    <Button on:click={() => showCreateModal = true} class="bg-primary text-primary-foreground">
      <span class="mr-2">+</span>
      SSH Key Genereren
    </Button>
  </div>

  {#if loading}
    <div class="text-center py-8">
      <div class="inline-block animate-spin rounded-full h-8 w-8 border-b-2 border-primary"></div>
      <p class="mt-2 text-muted-foreground">Laden...</p>
    </div>
  {:else}
    {#if sshKeys.length === 0}
      <Card class="p-8 text-center">
        <div class="text-6xl mb-4">🔑</div>
        <h3 class="text-lg font-semibold mb-2">Nog geen SSH keys</h3>
        <p class="text-muted-foreground mb-4">Genereer SSH keys om veilig verbinding te maken met je webservers voor deployments.</p>
        <Button on:click={() => showCreateModal = true}>
          Eerste SSH Key Genereren
        </Button>
      </Card>
    {:else}
      <div class="grid gap-4">
        {#each sshKeys as key}
          <Card class="p-6">
            <div class="flex justify-between items-start">
              <div class="flex-1">
                <div class="flex items-center gap-3 mb-2">
                  <h3 class="text-lg font-semibold">{key.name}</h3>
                  <span class="px-2 py-1 text-xs font-medium rounded-full bg-green-50 border border-green-200 text-green-600">
                    {key.key_type.toUpperCase()} {key.key_size}
                  </span>
                  {#if key.is_active}
                    <span class="px-2 py-1 text-xs font-medium rounded-full bg-blue-50 border border-blue-200 text-blue-600">
                      Actief
                    </span>
                  {/if}
                </div>
                <p class="text-muted-foreground text-sm mb-3">{key.description || 'Geen beschrijving'}</p>
                
                <div class="grid grid-cols-1 md:grid-cols-2 gap-4 text-sm">
                  <div>
                    <span class="font-medium text-muted-foreground">Fingerprint:</span>
                    <p class="font-mono text-xs break-all">{formatFingerprint(key.fingerprint)}</p>
                  </div>
                  <div>
                    <span class="font-medium text-muted-foreground">Aangemaakt:</span>
                    <p>{new Date(key.created_at).toLocaleDateString('nl-NL')}</p>
                  </div>
                </div>

                {#if key.last_used_at}
                  <div class="mt-2 text-sm">
                    <span class="font-medium text-muted-foreground">Laatst gebruikt:</span>
                    <span class="ml-1">{new Date(key.last_used_at).toLocaleDateString('nl-NL')}</span>
                  </div>
                {/if}
              </div>

              <div class="flex gap-2 ml-4">
                <Button
                  on:click={() => showPublicKey(key)}
                  size="sm"
                  variant="outline"
                  class="border-blue-300 text-blue-600 hover:bg-blue-50"
                >
                  Public Key
                </Button>
                <Button
                  on:click={() => deleteSSHKey(key.id)}
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

<!-- Create SSH Key Modal -->
{#if showCreateModal}
  <Modal on:close={() => showCreateModal = false}>
    <div class="p-6">
      <h2 class="text-xl font-semibold mb-4">SSH Key Genereren</h2>
      
      <form on:submit|preventDefault={createSSHKey} class="space-y-4">
        <div>
          <Label for="name">Naam</Label>
          <Input
            id="name"
            bind:value={keyForm.name}
            placeholder="Productie Server Key"
            required
          />
        </div>

        <div>
          <Label for="description">Beschrijving</Label>
          <Textarea
            id="description"
            bind:value={keyForm.description}
            placeholder="SSH key voor toegang tot productie server"
            rows={2}
          />
        </div>

        <div class="grid grid-cols-2 gap-4">
          <div>
            <Label for="key_type">Key Type</Label>
            <Select id="key_type" bind:value={keyForm.key_type}>
              <option value="rsa">RSA</option>
              <!-- <option value="ed25519">Ed25519</option> -->
            </Select>
          </div>
          <div>
            <Label for="key_size">Key Size</Label>
            <Select id="key_size" bind:value={keyForm.key_size}>
              <option value="2048">2048 bits</option>
              <option value="4096">4096 bits</option>
            </Select>
          </div>
        </div>

        <div class="bg-yellow-50 border border-yellow-200 rounded-lg p-4">
          <p class="text-yellow-800 text-sm">
            <strong>Let op:</strong> De private key wordt veilig opgeslagen in CloudBox. 
            De public key moet je handmatig toevoegen aan de authorized_keys van je server.
          </p>
        </div>

        <div class="flex justify-end space-x-2 pt-4">
          <Button type="button" variant="outline" on:click={() => showCreateModal = false}>
            Annuleren
          </Button>
          <Button type="submit" class="bg-primary text-primary-foreground">
            SSH Key Genereren
          </Button>
        </div>
      </form>
    </div>
  </Modal>
{/if}

<!-- Show Public Key Modal -->
{#if showKeyModal && selectedKey}
  <Modal on:close={() => showKeyModal = false}>
    <div class="p-6">
      <h2 class="text-xl font-semibold mb-4">Public Key: {selectedKey.name}</h2>
      
      <div class="space-y-4">
        <div>
          <Label>Public Key</Label>
          <div class="mt-1 relative">
            <textarea
              readonly
              class="w-full h-32 p-3 border border-border rounded-md font-mono text-xs resize-none bg-gray-50"
              value={selectedKey.public_key}
            ></textarea>
            <Button
              on:click={() => copyToClipboard(selectedKey.public_key)}
              size="sm"
              class="absolute top-2 right-2"
            >
              Kopiëren
            </Button>
          </div>
        </div>

        <div class="bg-blue-50 border border-blue-200 rounded-lg p-4">
          <h4 class="font-medium text-blue-900 mb-2">Installatie instructies:</h4>
          <ol class="text-blue-800 text-sm space-y-1 list-decimal list-inside">
            <li>Kopieer de public key hierboven</li>
            <li>Log in op je server via SSH</li>
            <li>Voeg de key toe: <code class="bg-blue-100 px-1 rounded">echo "PASTE_KEY_HERE" >> ~/.ssh/authorized_keys</code></li>
            <li>Zet de juiste permissies: <code class="bg-blue-100 px-1 rounded">chmod 600 ~/.ssh/authorized_keys</code></li>
          </ol>
        </div>

        <div class="text-sm text-muted-foreground">
          <p><strong>Fingerprint:</strong> {formatFingerprint(selectedKey.fingerprint)}</p>
          <p><strong>Type:</strong> {selectedKey.key_type.toUpperCase()} {selectedKey.key_size} bits</p>
        </div>

        <div class="flex justify-end pt-4">
          <Button on:click={() => showKeyModal = false}>
            Sluiten
          </Button>
        </div>
      </div>
    </div>
  </Modal>
{/if}