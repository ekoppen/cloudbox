<script lang="ts">
  import { onMount } from 'svelte';
  import { goto } from '$app/navigation';
  import { auth } from '$lib/stores/auth';
  import Card from '$lib/components/ui/card.svelte';
  import Button from '$lib/components/ui/button.svelte';
  import Input from '$lib/components/ui/input.svelte';
  import Label from '$lib/components/ui/label.svelte';
  import Badge from '$lib/components/ui/badge.svelte';
  import Textarea from '$lib/components/ui/textarea.svelte';
  import Icon from '$lib/components/ui/icon.svelte';
  
  interface Organization {
    id: number;
    name: string;
    color: string;
  }

  interface Project {
    id: number;
    name: string;
    description: string;
    slug: string;
    created_at: string;
    is_active: boolean;
    organization?: Organization;
  }

  let projects: Project[] = [];
  let loading = true;
  let error = '';
  let showCreateModal = false;
  let newProject = { name: '', description: '' };
  let creating = false;
  let searchTerm = '';

  onMount(() => {
    loadProjects();
  });

  async function loadProjects() {
    loading = true;
    error = '';

    try {
      const response = await fetch('http://localhost:8080/api/v1/projects', {
        headers: {
          'Authorization': `Bearer ${$auth.token}`,
          'Content-Type': 'application/json',
        },
      });

      if (response.ok) {
        projects = await response.json();
      } else {
        const data = await response.json();
        error = data.error || 'Fout bij laden van projecten';
      }
    } catch (err) {
      error = 'Netwerkfout bij laden van projecten';
      console.error('Load projects error:', err);
    } finally {
      loading = false;
    }
  }

  async function createProject() {
    if (!newProject.name.trim()) {
      return;
    }

    creating = true;

    try {
      const response = await fetch('http://localhost:8080/api/v1/projects', {
        method: 'POST',
        headers: {
          'Authorization': `Bearer ${$auth.token}`,
          'Content-Type': 'application/json',
        },
        body: JSON.stringify(newProject),
      });

      if (response.ok) {
        const project = await response.json();
        projects = [...projects, project];
        showCreateModal = false;
        newProject = { name: '', description: '' };
      } else {
        const data = await response.json();
        error = data.error || 'Fout bij aanmaken van project';
      }
    } catch (err) {
      error = 'Netwerkfout bij aanmaken van project';
      console.error('Create project error:', err);
    } finally {
      creating = false;
    }
  }

  function formatDate(dateStr: string) {
    return new Date(dateStr).toLocaleDateString('nl-NL', {
      year: 'numeric',
      month: 'long',
      day: 'numeric'
    });
  }

  // Filter projects based on search term
  $: filteredProjects = projects.filter(project =>
    project.name.toLowerCase().includes(searchTerm.toLowerCase()) ||
    project.description.toLowerCase().includes(searchTerm.toLowerCase()) ||
    project.slug.toLowerCase().includes(searchTerm.toLowerCase())
  );
</script>

<svelte:head>
  <title>Projecten - CloudBox</title>
</svelte:head>

<div class="space-y-8">
  <!-- Header -->
  <div class="flex items-center justify-between">
    <div class="flex items-center space-x-4">
      <div class="w-12 h-12 bg-primary rounded-xl flex items-center justify-center">
        <Icon name="package" size={24} color="white" />
      </div>
      <div>
        <h1 class="text-3xl font-bold text-foreground">Projecten</h1>
        <p class="text-muted-foreground">
          Beheer al je CloudBox projecten op één plek
        </p>
      </div>
    </div>
    <div class="flex items-center space-x-3">
      <Button
        variant="outline"
        href="/dashboard"
        size="lg"
        class="flex items-center space-x-2"
      >
        <Icon name="backup" size={16} />
        <span>Terug naar Dashboard</span>
      </Button>
      <Button
        on:click={() => showCreateModal = true}
        size="lg"
        class="flex items-center space-x-2"
      >
        <Icon name="package" size={16} />
        <span>Nieuw Project</span>
      </Button>
    </div>
  </div>

  <!-- Stats and Search -->
  <div class="flex flex-col lg:flex-row lg:items-center lg:justify-between gap-6">
    <!-- Quick Stats -->
    <div class="flex items-center space-x-6">
      <div class="text-center">
        <p class="text-2xl font-bold text-foreground">{projects.length}</p>
        <p class="text-sm text-muted-foreground">Totaal</p>
      </div>
      <div class="text-center">
        <p class="text-2xl font-bold text-green-600">{projects.filter(p => p.is_active).length}</p>
        <p class="text-sm text-muted-foreground">Actief</p>
      </div>
      <div class="text-center">
        <p class="text-2xl font-bold text-gray-400">{projects.filter(p => !p.is_active).length}</p>
        <p class="text-sm text-muted-foreground">Inactief</p>
      </div>
    </div>

    <!-- Search -->
    <div class="relative max-w-md">
      <Icon name="search" size={16} className="absolute left-3 top-1/2 transform -translate-y-1/2 text-muted-foreground" />
      <Input
        type="text"
        placeholder="Zoek projecten..."
        bind:value={searchTerm}
        class="pl-10"
      />
    </div>
  </div>

  <!-- Error message -->
  {#if error}
    <Card class="bg-destructive/10 border-destructive/20 p-4">
      <div class="flex justify-between items-center">
        <p class="text-destructive text-sm">{error}</p>
        <Button
          variant="ghost"
          size="sm"
          on:click={() => error = ''}
          class="text-destructive hover:text-destructive/80"
        >
          ×
        </Button>
      </div>
    </Card>
  {/if}

  <!-- Projects Grid -->
  {#if loading}
    <Card class="p-12">
      <div class="text-center">
        <div class="animate-spin rounded-full h-8 w-8 border-b-2 border-primary mx-auto"></div>
        <p class="mt-4 text-muted-foreground">Projecten laden...</p>
      </div>
    </Card>
  {:else if filteredProjects.length === 0 && searchTerm}
    <Card class="p-12">
      <div class="text-center">
        <div class="w-16 h-16 bg-muted rounded-full flex items-center justify-center mx-auto mb-4">
          <Icon name="search" size={32} className="text-muted-foreground" />
        </div>
        <h3 class="text-lg font-medium text-foreground mb-2">Geen resultaten</h3>
        <p class="text-muted-foreground mb-6 max-w-sm mx-auto">
          Geen projecten gevonden voor "{searchTerm}". Probeer een andere zoekterm.
        </p>
        <Button
          variant="outline"
          on:click={() => searchTerm = ''}
          size="lg"
          class="flex items-center space-x-2"
        >
          <Icon name="backup" size={16} />
          <span>Zoekterm Wissen</span>
        </Button>
      </div>
    </Card>
  {:else if projects.length === 0}
    <Card class="p-12">
      <div class="text-center">
        <div class="w-16 h-16 bg-muted rounded-full flex items-center justify-center mx-auto mb-4">
          <Icon name="package" size={32} className="text-muted-foreground" />
        </div>
        <h3 class="text-lg font-medium text-foreground mb-2">Nog geen projecten</h3>
        <p class="text-muted-foreground mb-6 max-w-sm mx-auto">
          Maak je eerste CloudBox project aan om te beginnen met je Backend-as-a-Service
        </p>
        <Button
          on:click={() => showCreateModal = true}
          size="lg"
          class="flex items-center space-x-2"
        >
          <Icon name="package" size={16} />
          <span>Eerste Project Aanmaken</span>
        </Button>
      </div>
    </Card>
  {:else}
    <!-- Projects Grid -->
    <div class="grid gap-6 md:grid-cols-2 lg:grid-cols-3">
      {#each filteredProjects as project}
        <Card class="group p-6 hover:shadow-lg hover:shadow-primary/5 transition-all duration-200 border hover:border-primary/20">
          <div class="flex items-start justify-between mb-4">
            <div class="flex items-center space-x-3">
              <div class="w-10 h-10 bg-primary/10 rounded-lg flex items-center justify-center">
                <Icon name="package" size={20} className="text-primary" />
              </div>
              <div>
                <h3 class="text-lg font-semibold text-card-foreground group-hover:text-primary transition-colors">
                  {project.name}
                </h3>
                {#if project.organization}
                  <div class="flex items-center space-x-1 text-xs">
                    <div 
                      class="w-2 h-2 rounded-full"
                      style="background-color: {project.organization.color}"
                    ></div>
                    <span class="text-muted-foreground">{project.organization.name}</span>
                  </div>
                {:else}
                  <p class="text-xs text-muted-foreground">
                    Persoonlijk project
                  </p>
                {/if}
              </div>
            </div>
            <Badge variant={project.is_active ? "default" : "secondary"} class="flex items-center space-x-1">
              <div class="w-2 h-2 rounded-full {project.is_active ? 'bg-green-500' : 'bg-gray-400'}"></div>
              <span>{project.is_active ? 'Actief' : 'Inactief'}</span>
            </Badge>
          </div>
          
          {#if project.description}
            <p class="text-muted-foreground text-sm mb-4 line-clamp-2">{project.description}</p>
          {/if}
          
          <div class="bg-muted/50 rounded-lg p-3 mb-4 space-y-2">
            <div class="flex items-center justify-between text-xs">
              <span class="text-muted-foreground">API Slug:</span>
              <code class="bg-background px-2 py-1 rounded text-xs font-mono">{project.slug}</code>
            </div>
            <div class="flex items-center justify-between text-xs">
              <span class="text-muted-foreground">Aangemaakt:</span>
              <span class="text-foreground">{formatDate(project.created_at)}</span>
            </div>
          </div>
          
          <div class="flex space-x-2">
            <Button
              href="/dashboard/projects/{project.id}"
              size="sm"
              class="flex-1 flex items-center justify-center space-x-2"
            >
              <Icon name="settings" size={14} />
              <span>Beheren</span>
            </Button>
          </div>
        </Card>
      {/each}
    </div>

    <!-- Load more / Pagination would go here -->
    {#if filteredProjects.length > 0}
      <div class="text-center">
        <Button
          variant="outline"
          on:click={loadProjects}
          size="lg"
          class="flex items-center space-x-2"
        >
          <Icon name="backup" size={16} />
          <span>Vernieuwen</span>
        </Button>
      </div>
    {/if}
  {/if}
</div>

<!-- Create Project Modal -->
{#if showCreateModal}
  <div class="fixed inset-0 bg-black/80 backdrop-blur-sm flex items-center justify-center p-4 z-50">
    <Card class="max-w-lg w-full p-6 border-2 shadow-2xl">
      <div class="flex items-center space-x-3 mb-6">
        <div class="w-10 h-10 bg-primary/10 rounded-lg flex items-center justify-center">
          <Icon name="package" size={20} className="text-primary" />
        </div>
        <div>
          <h2 class="text-xl font-bold text-foreground">Nieuw Project</h2>
          <p class="text-sm text-muted-foreground">Maak een nieuwe CloudBox BaaS project aan</p>
        </div>
      </div>
      
      <form on:submit|preventDefault={createProject} class="space-y-6">
        <div class="space-y-2">
          <Label for="project-name" class="flex items-center space-x-2">
            <Icon name="user" size={14} />
            <span>Project naam</span>
          </Label>
          <Input
            id="project-name"
            type="text"
            bind:value={newProject.name}
            required
            placeholder="Mijn geweldige app"
            class="pl-4"
          />
          <p class="text-xs text-muted-foreground">
            Deze naam wordt gebruikt voor je project identificatie
          </p>
        </div>
        
        <div class="space-y-2">
          <Label for="project-description" class="flex items-center space-x-2">
            <Icon name="edit" size={14} />
            <span>Beschrijving (optioneel)</span>
          </Label>
          <Textarea
            id="project-description"
            bind:value={newProject.description}
            rows={3}
            placeholder="Een korte beschrijving van je project..."
            class="pl-4"
          />
        </div>

        <!-- Project Features Preview -->
        <div class="bg-muted/50 rounded-lg p-4 space-y-3">
          <div class="flex items-center space-x-2 text-sm font-medium text-foreground">
            <Icon name="cloud" size={16} className="text-primary" />
            <span>Jouw project krijgt toegang tot:</span>
          </div>
          <div class="grid grid-cols-2 gap-2 text-xs">
            <div class="flex items-center space-x-2">
              <Icon name="database" size={12} className="text-green-600" />
              <span class="text-muted-foreground">Database API</span>
            </div>
            <div class="flex items-center space-x-2">
              <Icon name="storage" size={12} className="text-purple-600" />
              <span class="text-muted-foreground">File Storage</span>
            </div>
            <div class="flex items-center space-x-2">
              <Icon name="auth" size={12} className="text-blue-600" />
              <span class="text-muted-foreground">Authentication</span>
            </div>
            <div class="flex items-center space-x-2">
              <Icon name="functions" size={12} className="text-orange-600" />
              <span class="text-muted-foreground">Cloud Functions</span>
            </div>
          </div>
        </div>
        
        <div class="flex space-x-3 pt-4">
          <Button
            type="button"
            variant="outline"
            on:click={() => {
              showCreateModal = false;
              newProject = { name: '', description: '' };
            }}
            class="flex-1 flex items-center justify-center space-x-2"
          >
            <Icon name="backup" size={14} />
            <span>Annuleren</span>
          </Button>
          <Button
            type="submit"
            disabled={creating || !newProject.name.trim()}
            class="flex-1 flex items-center justify-center space-x-2"
          >
            {#if creating}
              <div class="animate-spin rounded-full h-4 w-4 border-b-2 border-white"></div>
              <span>Aanmaken...</span>
            {:else}
              <Icon name="package" size={14} />
              <span>Project Aanmaken</span>
            {/if}
          </Button>
        </div>
      </form>
    </Card>
  </div>
{/if}