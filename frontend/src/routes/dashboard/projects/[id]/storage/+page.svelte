<script lang="ts">
  import { page } from '$app/stores';
  import { onMount } from 'svelte';
  import { auth } from '$lib/stores/auth';
  import Card from '$lib/components/ui/card.svelte';
  import Button from '$lib/components/ui/button.svelte';
  import Input from '$lib/components/ui/input.svelte';
  import Label from '$lib/components/ui/label.svelte';
  import Badge from '$lib/components/ui/badge.svelte';
  import Icon from '$lib/components/ui/icon.svelte';

  interface StorageFile {
    id: string;
    name: string;
    size: number;
    type: string;
    uploaded_at: string;
    url: string;
    folder?: string;
  }

  interface StorageFolder {
    name: string;
    files: number;
    size: number;
  }

  let files: StorageFile[] = [
    {
      id: '1',
      name: 'profile-avatar.jpg',
      size: 245760,
      type: 'image/jpeg',
      uploaded_at: '2025-01-19T10:30:00Z',
      url: '/storage/profile-avatar.jpg',
      folder: 'avatars'
    },
    {
      id: '2',
      name: 'document.pdf',
      size: 1048576,
      type: 'application/pdf',
      uploaded_at: '2025-01-18T14:20:00Z',
      url: '/storage/document.pdf',
      folder: 'documents'
    },
    {
      id: '3',
      name: 'banner.png',
      size: 512000,
      type: 'image/png',
      uploaded_at: '2025-01-17T09:15:00Z',
      url: '/storage/banner.png',
      folder: 'images'
    }
  ];

  let folders: StorageFolder[] = [
    { name: 'avatars', files: 12, size: 2048000 },
    { name: 'documents', files: 8, size: 15728640 },
    { name: 'images', files: 24, size: 8388608 },
    { name: 'uploads', files: 5, size: 1048576 }
  ];

  let storageStats = {
    total_files: 49,
    total_size: 27262976,
    storage_limit: 1073741824, // 1GB
    bandwidth_used: 157286400, // 150MB
    bandwidth_limit: 10737418240 // 10GB
  };

  let activeTab = 'files';
  let selectedFolder = '';
  let showUpload = false;
  let uploadFiles: FileList | null = null;
  let uploading = false;

  $: projectId = $page.params.id;
  $: filteredFiles = selectedFolder ? files.filter(f => f.folder === selectedFolder) : files;

  function formatFileSize(bytes: number): string {
    if (bytes === 0) return '0 B';
    const k = 1024;
    const sizes = ['B', 'KB', 'MB', 'GB'];
    const i = Math.floor(Math.log(bytes) / Math.log(k));
    return parseFloat((bytes / Math.pow(k, i)).toFixed(2)) + ' ' + sizes[i];
  }

  function getFileIcon(type: string): string {
    if (type.startsWith('image/')) return '🖼️';
    if (type.includes('pdf')) return '📄';
    if (type.includes('video')) return '🎥';
    if (type.includes('audio')) return '🎵';
    if (type.includes('zip') || type.includes('archive')) return '📦';
    if (type.includes('text') || type.includes('document')) return '📝';
    return '📁';
  }

  function formatDate(dateStr: string): string {
    return new Date(dateStr).toLocaleDateString('nl-NL', {
      year: 'numeric',
      month: 'short',
      day: 'numeric',
      hour: '2-digit',
      minute: '2-digit'
    });
  }

  function selectFolder(folderName: string) {
    selectedFolder = selectedFolder === folderName ? '' : folderName;
  }

  function deleteFile(fileId: string) {
    if (confirm('Weet je zeker dat je dit bestand wilt verwijderen?')) {
      files = files.filter(f => f.id !== fileId);
    }
  }

  function handleFileUpload(event: Event) {
    const target = event.target as HTMLInputElement;
    uploadFiles = target.files;
  }

  async function uploadFile() {
    if (!uploadFiles || uploadFiles.length === 0) return;

    uploading = true;
    try {
      // Simulate upload
      await new Promise(resolve => setTimeout(resolve, 2000));
      
      for (let i = 0; i < uploadFiles.length; i++) {
        const file = uploadFiles[i];
        const newFile: StorageFile = {
          id: Date.now().toString() + i,
          name: file.name,
          size: file.size,
          type: file.type,
          uploaded_at: new Date().toISOString(),
          url: `/storage/${file.name}`,
          folder: selectedFolder || 'uploads'
        };
        files = [newFile, ...files];
      }

      showUpload = false;
      uploadFiles = null;
    } catch (err) {
      console.error('Upload error:', err);
    } finally {
      uploading = false;
    }
  }

  function downloadFile(file: StorageFile) {
    // Simulate download
    alert(`Downloaden van ${file.name} gestart...`);
  }

  function copyFileUrl(file: StorageFile) {
    navigator.clipboard.writeText(`https://api.cloudbox.nl${file.url}`);
    alert('URL gekopieerd naar klembord!');
  }
</script>

<svelte:head>
  <title>Opslag - CloudBox</title>
</svelte:head>

<div class="space-y-6">
  <!-- Header -->
  <div class="flex items-center justify-between">
    <div class="flex items-center space-x-4">
      <div class="w-10 h-10 bg-primary/10 rounded-lg flex items-center justify-center">
        <Icon name="storage" size={20} className="text-primary" />
      </div>
      <div>
        <h1 class="text-2xl font-bold text-foreground">Opslag</h1>
        <p class="text-sm text-muted-foreground">
          Beheer bestanden en opslag instellingen
        </p>
      </div>
    </div>
    <Button
      on:click={() => showUpload = true}
      class="flex items-center space-x-2"
    >
      <Icon name="storage" size={16} />
      <span>Bestand Uploaden</span>
    </Button>
  </div>

  <!-- Storage Stats -->
  <div class="grid grid-cols-1 md:grid-cols-3 gap-4">
    <Card class="p-6">
      <div class="flex items-center justify-between">
        <div>
          <p class="text-sm font-medium text-muted-foreground">Opslag Gebruikt</p>
          <p class="text-2xl font-bold text-foreground">{formatFileSize(storageStats.total_size)}</p>
          <p class="text-xs text-muted-foreground">van {formatFileSize(storageStats.storage_limit)}</p>
        </div>
        <div class="w-10 h-10 bg-blue-100 dark:bg-blue-900 rounded-lg flex items-center justify-center">
          <Icon name="storage" size={20} className="text-blue-600 dark:text-blue-400" />
        </div>
      </div>
      <div class="mt-4 bg-muted rounded-full h-2">
        <div 
          class="bg-blue-500 h-2 rounded-full" 
          style="width: {(storageStats.total_size / storageStats.storage_limit) * 100}%"
        ></div>
      </div>
    </Card>

    <Card class="p-6">
      <div class="flex items-center justify-between">
        <div>
          <p class="text-sm font-medium text-muted-foreground">Bandbreedte</p>
          <p class="text-2xl font-bold text-foreground">{formatFileSize(storageStats.bandwidth_used)}</p>
          <p class="text-xs text-muted-foreground">van {formatFileSize(storageStats.bandwidth_limit)}</p>
        </div>
        <div class="w-10 h-10 bg-green-100 dark:bg-green-900 rounded-lg flex items-center justify-center">
          <Icon name="cloud" size={20} className="text-green-600 dark:text-green-400" />
        </div>
      </div>
      <div class="mt-4 bg-muted rounded-full h-2">
        <div 
          class="bg-green-500 h-2 rounded-full" 
          style="width: {(storageStats.bandwidth_used / storageStats.bandwidth_limit) * 100}%"
        ></div>
      </div>
    </Card>

    <Card class="p-6">
      <div class="flex items-center justify-between">
        <div>
          <p class="text-sm font-medium text-muted-foreground">Totaal Bestanden</p>
          <p class="text-2xl font-bold text-foreground">{storageStats.total_files}</p>
          <p class="text-xs text-muted-foreground">in {folders.length} mappen</p>
        </div>
        <div class="w-10 h-10 bg-purple-100 dark:bg-purple-900 rounded-lg flex items-center justify-center">
          <Icon name="package" size={20} className="text-purple-600 dark:text-purple-400" />
        </div>
      </div>
    </Card>
  </div>

  <!-- Tabs -->
  <div class="border-b border-border">
    <nav class="flex space-x-8">
      <button
        on:click={() => activeTab = 'files'}
        class="flex items-center space-x-2 py-2 px-1 border-b-2 text-sm font-medium transition-colors {
          activeTab === 'files' 
            ? 'border-primary text-primary' 
            : 'border-transparent text-muted-foreground hover:text-foreground hover:border-border'
        }"
      >
        <Icon name="storage" size={16} />
        <span>Bestanden</span>
      </button>
      <button
        on:click={() => activeTab = 'folders'}
        class="flex items-center space-x-2 py-2 px-1 border-b-2 text-sm font-medium transition-colors {
          activeTab === 'folders' 
            ? 'border-primary text-primary' 
            : 'border-transparent text-muted-foreground hover:text-foreground hover:border-border'
        }"
      >
        <Icon name="package" size={16} />
        <span>Mappen</span>
      </button>
      <button
        on:click={() => activeTab = 'settings'}
        class="flex items-center space-x-2 py-2 px-1 border-b-2 text-sm font-medium transition-colors {
          activeTab === 'settings' 
            ? 'border-primary text-primary' 
            : 'border-transparent text-muted-foreground hover:text-foreground hover:border-border'
        }"
      >
        <Icon name="settings" size={16} />
        <span>Instellingen</span>
      </button>
    </nav>
  </div>

  <!-- Files Tab -->
  {#if activeTab === 'files'}
    <div class="space-y-4">
      <!-- Folder Filter -->
      {#if folders.length > 0}
        <div class="flex flex-wrap gap-2">
          <Button
            variant={selectedFolder === '' ? 'default' : 'outline'}
            size="sm"
            on:click={() => selectFolder('')}
          >
            Alle bestanden
          </Button>
          {#each folders as folder}
            <Button
              variant={selectedFolder === folder.name ? 'default' : 'outline'}
              size="sm"
              on:click={() => selectFolder(folder.name)}
              class="flex items-center space-x-1"
            >
              <Icon name="package" size={14} />
              <span>{folder.name} ({folder.files})</span>
            </Button>
          {/each}
        </div>
      {/if}

      <!-- Files List -->
      <Card>
        {#if filteredFiles.length === 0}
          <div class="p-12 text-center">
            <div class="w-16 h-16 bg-muted rounded-lg flex items-center justify-center mx-auto mb-4">
              <Icon name="storage" size={32} className="text-muted-foreground" />
            </div>
            <h3 class="text-lg font-medium text-foreground mb-2">Geen bestanden gevonden</h3>
            <p class="text-muted-foreground">Upload bestanden om ze hier te zien</p>
          </div>
        {:else}
          <div class="overflow-x-auto">
            <table class="min-w-full divide-y divide-border">
              <thead class="bg-muted/30">
                <tr>
                  <th class="px-6 py-3 text-left text-xs font-medium text-muted-foreground uppercase tracking-wider">Bestand</th>
                  <th class="px-6 py-3 text-left text-xs font-medium text-muted-foreground uppercase tracking-wider">Grootte</th>
                  <th class="px-6 py-3 text-left text-xs font-medium text-muted-foreground uppercase tracking-wider">Type</th>
                  <th class="px-6 py-3 text-left text-xs font-medium text-muted-foreground uppercase tracking-wider">Geüpload</th>
                  <th class="px-6 py-3 text-right text-xs font-medium text-muted-foreground uppercase tracking-wider">Acties</th>
                </tr>
              </thead>
              <tbody class="bg-card divide-y divide-border">
                {#each filteredFiles as file}
                  <tr class="hover:bg-muted/30">
                    <td class="px-6 py-4">
                      <div class="flex items-center">
                        <div class="w-8 h-8 bg-primary/10 rounded-lg flex items-center justify-center mr-3">
                          <Icon name="storage" size={16} className="text-primary" />
                        </div>
                        <div>
                          <div class="text-sm font-medium text-foreground">{file.name}</div>
                          {#if file.folder}
                            <div class="text-sm text-muted-foreground flex items-center">
                              <Icon name="package" size={12} className="mr-1" />
                              {file.folder}
                            </div>
                          {/if}
                        </div>
                      </div>
                    </td>
                    <td class="px-6 py-4 text-sm text-muted-foreground">
                      {formatFileSize(file.size)}
                    </td>
                    <td class="px-6 py-4 text-sm text-muted-foreground">
                      {file.type}
                    </td>
                    <td class="px-6 py-4 text-sm text-muted-foreground">
                      {formatDate(file.uploaded_at)}
                    </td>
                    <td class="px-6 py-4 text-right">
                      <div class="flex justify-end space-x-2">
                        <Button
                          variant="ghost"
                          size="sm"
                          on:click={() => downloadFile(file)}
                        >
                          Download
                        </Button>
                        <Button
                          variant="ghost"
                          size="sm"
                          on:click={() => copyFileUrl(file)}
                        >
                          URL
                        </Button>
                        <Button
                          variant="ghost"
                          size="sm"
                          class="text-destructive hover:text-destructive"
                          on:click={() => deleteFile(file.id)}
                        >
                          Verwijder
                        </Button>
                      </div>
                    </td>
                  </tr>
                {/each}
              </tbody>
            </table>
          </div>
        {/if}
      </Card>
    </div>
  {/if}

  <!-- Folders Tab -->
  {#if activeTab === 'folders'}
    <div class="space-y-4">
      <div class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
        {#each folders as folder}
          <Card class="p-6 hover:shadow-md transition-shadow cursor-pointer" on:click={() => selectFolder(folder.name)}>
            <div class="flex items-center justify-between">
              <div class="flex items-center space-x-3">
                <div class="w-12 h-12 bg-primary/10 rounded-lg flex items-center justify-center">
                  <Icon name="package" size={24} className="text-primary" />
                </div>
                <div>
                  <h3 class="text-lg font-medium text-foreground">{folder.name}</h3>
                  <p class="text-sm text-muted-foreground">{folder.files} bestanden</p>
                </div>
              </div>
              <div class="text-right">
                <p class="text-sm font-medium text-foreground">{formatFileSize(folder.size)}</p>
              </div>
            </div>
          </Card>
        {/each}
      </div>
    </div>
  {/if}

  <!-- Settings Tab -->
  {#if activeTab === 'settings'}
    <div class="space-y-6">
      <div>
        <h2 class="text-lg font-medium text-foreground">Opslag Instellingen</h2>
        <p class="text-sm text-muted-foreground">Configureer opslag en upload instellingen</p>
      </div>

      <Card class="p-6">
        <div class="space-y-6">
          <div>
            <h3 class="text-sm font-medium text-foreground mb-3">Upload Limieten</h3>
            <div class="grid grid-cols-1 md:grid-cols-2 gap-4">
              <div>
                <Label>Max bestandsgrootte (MB)</Label>
                <Input type="number" value="50" class="mt-1" />
              </div>
              <div>
                <Label>Toegestane bestandstypen</Label>
                <Input type="text" value="jpg,png,gif,pdf,doc,docx" class="mt-1" />
              </div>
            </div>
          </div>

          <div>
            <h3 class="text-sm font-medium text-foreground mb-3">CDN Instellingen</h3>
            <div class="space-y-4">
              <label class="flex items-center">
                <input type="checkbox" checked class="rounded border-border text-primary focus:ring-primary" />
                <span class="ml-2 text-sm text-foreground">CDN inschakelen voor snellere downloads</span>
              </label>
              <label class="flex items-center">
                <input type="checkbox" class="rounded border-border text-primary focus:ring-primary" />
                <span class="ml-2 text-sm text-foreground">Automatische image optimalisatie</span>
              </label>
            </div>
          </div>

          <div class="flex justify-end">
            <Button>
              Instellingen Opslaan
            </Button>
          </div>
        </div>
      </Card>
    </div>
  {/if}
</div>

<!-- Upload Modal -->
{#if showUpload}
  <div class="fixed inset-0 bg-black/80 backdrop-blur-sm flex items-center justify-center p-4 z-50">
    <Card class="max-w-md w-full p-6 border-2 shadow-2xl">
      <div class="flex items-center space-x-3 mb-4">
        <div class="w-10 h-10 bg-primary/10 rounded-lg flex items-center justify-center">
          <Icon name="storage" size={20} className="text-primary" />
        </div>
        <h2 class="text-xl font-bold text-foreground">Bestanden Uploaden</h2>
      </div>
      
      <form on:submit|preventDefault={uploadFile} class="space-y-4">
        <div>
          <Label>Selecteer bestanden</Label>
          <input
            type="file"
            multiple
            on:change={handleFileUpload}
            class="mt-1 block w-full text-sm text-muted-foreground file:mr-4 file:py-2 file:px-4 file:rounded-full file:border-0 file:text-sm file:font-semibold file:bg-primary/10 file:text-primary hover:file:bg-primary/20"
          />
        </div>

        {#if selectedFolder}
          <div>
            <Label>Map</Label>
            <Input
              type="text"
              value={selectedFolder}
              readonly
              class="mt-1 bg-muted"
            />
          </div>
        {/if}
        
        <div class="flex space-x-3 pt-4">
          <Button
            type="button"
            variant="outline"
            on:click={() => { showUpload = false; uploadFiles = null; }}
            class="flex-1"
          >
            Annuleren
          </Button>
          <Button
            type="submit"
            disabled={uploading || !uploadFiles || uploadFiles.length === 0}
            class="flex-1"
          >
            {uploading ? 'Uploaden...' : 'Upload Bestanden'}
          </Button>
        </div>
      </form>
    </Card>
  </div>
{/if}