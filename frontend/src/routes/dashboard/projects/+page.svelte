<style>
  .glassmorphism-card {
    background: rgba(255, 255, 255, 0.85);
    backdrop-filter: blur(20px);
    border: 1px solid rgba(255, 255, 255, 0.2);
    border-radius: 16px;
    padding: 24px;
    box-shadow: 
      0 8px 25px -8px rgba(0, 0, 0, 0.1),
      0 4px 12px -4px rgba(0, 0, 0, 0.08),
      0 0 0 1px rgba(255, 255, 255, 0.05) inset;
    transition: all 0.3s cubic-bezier(0.4, 0.0, 0.2, 1);
    position: relative;
    overflow: hidden;
  }

  .glassmorphism-card:hover {
    transform: translateY(-2px);
    box-shadow: 
      0 12px 35px -12px rgba(0, 0, 0, 0.15),
      0 8px 20px -8px rgba(0, 0, 0, 0.12),
      0 0 0 1px rgba(255, 255, 255, 0.1) inset;
  }

  .glassmorphism-card::before {
    content: '';
    position: absolute;
    top: 0;
    left: 0;
    right: 0;
    bottom: 0;
    background: linear-gradient(
      135deg,
      rgba(255, 255, 255, 0.1) 0%,
      rgba(255, 255, 255, 0) 50%,
      rgba(0, 0, 0, 0.05) 100%
    );
    pointer-events: none;
    z-index: 1;
  }

  .glassmorphism-card > * {
    position: relative;
    z-index: 2;
  }

  .glassmorphism-icon {
    width: 40px;
    height: 40px;
    border-radius: 10px;
    display: flex;
    align-items: center;
    justify-content: center;
    backdrop-filter: blur(10px);
    border: 1px solid rgba(255, 255, 255, 0.2);
    transition: all 0.2s ease;
  }

  .glassmorphism-icon:hover {
    transform: scale(1.05);
  }

  .glassmorphism-search {
    background: rgba(255, 255, 255, 0.9);
    backdrop-filter: blur(15px);
    border: 1px solid rgba(255, 255, 255, 0.3);
    border-radius: 12px;
    transition: all 0.3s ease;
  }

  .glassmorphism-search:focus-within {
    background: rgba(255, 255, 255, 0.95);
    border-color: rgba(59, 130, 246, 0.4);
    box-shadow: 0 4px 20px -4px rgba(59, 130, 246, 0.2);
  }

  .glassmorphism-badge {
    background: rgba(255, 255, 255, 0.8);
    backdrop-filter: blur(10px);
    border: 1px solid rgba(255, 255, 255, 0.3);
    border-radius: 20px;
    padding: 4px 12px;
    font-size: 0.75rem;
    font-weight: 500;
  }

  /* Dark mode support - CloudBox theme system */
  :global(.cloudbox-dark) .glassmorphism-card {
    background: rgba(26, 26, 26, 0.7);
    border: 1px solid rgba(255, 255, 255, 0.1);
    box-shadow: 
      0 8px 25px -8px rgba(0, 0, 0, 0.6),
      0 4px 12px -4px rgba(0, 0, 0, 0.4),
      0 0 0 1px rgba(255, 255, 255, 0.05) inset;
    backdrop-filter: blur(25px);
  }
  
  :global(.cloudbox-dark) .glassmorphism-card:hover {
    background: rgba(33, 33, 33, 0.8);
    box-shadow: 
      0 12px 35px -12px rgba(0, 0, 0, 0.7),
      0 8px 20px -8px rgba(0, 0, 0, 0.5),
      0 0 0 1px rgba(255, 255, 255, 0.1) inset;
    backdrop-filter: blur(30px);
  }
  
  :global(.cloudbox-dark) .glassmorphism-card::before {
    background: linear-gradient(
      135deg,
      rgba(255, 255, 255, 0.03) 0%,
      rgba(255, 255, 255, 0) 50%,
      rgba(0, 0, 0, 0.15) 100%
    );
  }
  
  :global(.cloudbox-dark) .glassmorphism-icon {
    border: 1px solid rgba(255, 255, 255, 0.15);
    background: rgba(255, 255, 255, 0.05);
  }

  :global(.cloudbox-dark) .glassmorphism-search {
    background: rgba(30, 41, 59, 0.8);
    border: 1px solid rgba(255, 255, 255, 0.1);
    backdrop-filter: blur(20px);
  }

  :global(.cloudbox-dark) .glassmorphism-search:focus-within {
    background: rgba(30, 41, 59, 0.9);
    border-color: rgba(59, 130, 246, 0.5);
    box-shadow: 0 4px 20px -4px rgba(59, 130, 246, 0.3);
  }

  :global(.cloudbox-dark) .glassmorphism-badge {
    background: rgba(30, 41, 59, 0.7);
    border: 1px solid rgba(255, 255, 255, 0.15);
    backdrop-filter: blur(10px);
  }

  /* Mobile responsiveness */
  @media (max-width: 640px) {
    .glassmorphism-card {
      padding: 20px;
      border-radius: 12px;
    }
    
    .glassmorphism-icon {
      width: 36px;
      height: 36px;
    }
  }

  /* Reduce motion for accessibility */
  @media (prefers-reduced-motion: reduce) {
    .glassmorphism-card,
    .glassmorphism-icon {
      transition: none;
    }
    
    .glassmorphism-card:hover {
      transform: none;
    }
  }
</style>

<script lang="ts">
  import { onMount } from 'svelte';
  import { goto } from '$app/navigation';
  import { auth } from '$lib/stores/auth';
  import { API_ENDPOINTS, API_BASE_URL, createApiRequest } from '$lib/config';
  import Card from '$lib/components/ui/card.svelte';
  import Button from '$lib/components/ui/button.svelte';
  import Input from '$lib/components/ui/input.svelte';
  import Label from '$lib/components/ui/label.svelte';
  import Badge from '$lib/components/ui/badge.svelte';
  import Textarea from '$lib/components/ui/textarea.svelte';
  import Icon from '$lib/components/ui/icon.svelte';
  import ProjectCard from '$lib/components/ui/project-card.svelte';
  
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
  let organizations: Organization[] = [];
  let loading = true;
  let loadingOrganizations = false;
  let error = '';
  let showCreateModal = false;
  let newProject = { name: '', description: '', organization_id: 0 };
  let creating = false;
  let searchTerm = '';

  onMount(() => {
    loadProjects();
    loadOrganizations();
  });

  async function loadProjects() {
    loading = true;
    error = '';

    try {
      const response = await createApiRequest(API_ENDPOINTS.projects.list, {
        headers: {
          'Authorization': `Bearer ${$auth.token}`,
        },
      });

      if (response.ok) {
        projects = await response.json();
        console.log('Loaded projects:', projects);
        console.log('First project:', projects[0]);
        projects.forEach((project, index) => {
          console.log(`Project ${index}:`, {
            id: project.id,
            name: project.name,
            idType: typeof project.id
          });
        });
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

  async function loadOrganizations() {
    loadingOrganizations = true;
    try {
      console.log('Loading organizations...');
      const response = await createApiRequest(API_ENDPOINTS.organizations.list, {
        headers: {
          'Authorization': `Bearer ${$auth.token}`,
        },
      });

      if (response.ok) {
        organizations = await response.json();
        console.log('Loaded organizations:', organizations);
        
        // If no organizations exist, auto-select the first one if available
        if (organizations.length > 0 && newProject.organization_id === 0) {
          newProject.organization_id = organizations[0].id;
        }
      } else {
        console.error('Failed to load organizations, status:', response.status);
        const errorData = await response.json().catch(() => null);
        console.error('Organizations error data:', errorData);
        error = 'Kon organizations niet laden. Controleer je permissies.';
      }
    } catch (err) {
      console.error('Load organizations error:', err);
      error = 'Netwerkfout bij laden van organizations';
    } finally {
      loadingOrganizations = false;
    }
  }

  async function createProject() {
    error = ''; // Reset error
    
    if (!newProject.name.trim()) {
      error = 'Vul een project naam in';
      return;
    }
    
    if (!newProject.organization_id || newProject.organization_id === 0) {
      error = 'Selecteer een organization voor dit project';
      console.error('Invalid organization_id:', newProject.organization_id);
      console.error('Available organizations:', organizations);
      return;
    }

    creating = true;
    console.log('Creating project with data:', newProject);

    try {
      const response = await createApiRequest(API_ENDPOINTS.projects.create, {
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
        newProject = { name: '', description: '', organization_id: 0 };
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
      <div class="w-12 h-12 bg-primary/10 rounded-2xl flex items-center justify-center">
        <Icon name="package" size={24} className="text-primary" />
      </div>
      <div>
        <h1 class="text-2xl font-bold text-foreground">Projecten</h1>
        <p class="text-sm text-muted-foreground">
          Beheer al je CloudBox projecten op één plek
        </p>
      </div>
    </div>
    <Button
      on:click={() => showCreateModal = true}
      variant="floating"
      size="icon-lg"
      iconOnly={true}
      tooltip="Nieuw Project"
    >
      <Icon name="plus" size={20} />
    </Button>
  </div>

  <!-- Search and Stats -->
  <div class="flex flex-col space-y-4 sm:flex-row sm:items-center sm:justify-between sm:space-y-0">
    <!-- Search -->
    <div class="relative max-w-sm">
      <div class="glassmorphism-search relative">
        <Icon name="search" size={16} className="absolute left-3 top-1/2 transform -translate-y-1/2 text-muted-foreground" />
        <Input
          type="text"
          placeholder="Zoek projecten..."
          bind:value={searchTerm}
          class="pl-10 h-10 bg-transparent border-none focus:ring-0 focus:border-none"
        />
      </div>
    </div>

    <!-- Quick Stats -->
    <div class="flex items-center space-x-4 text-sm">
      <div class="glassmorphism-badge flex items-center space-x-2">
        <div class="h-2 w-2 rounded-full bg-primary"></div>
        <span class="text-muted-foreground">{projects.length} projecten</span>
      </div>
      <div class="glassmorphism-badge flex items-center space-x-2">
        <div class="h-2 w-2 rounded-full bg-success"></div>
        <span class="text-muted-foreground">{projects.filter(p => p.is_active).length} actief</span>
      </div>
    </div>
  </div>

  <!-- Error message -->
  {#if error}
    <div class="glassmorphism-card bg-destructive/10 border-destructive/30">
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
    </div>
  {/if}

  <!-- Projects Grid -->
  {#if loading}
    <div class="text-center py-12">
      <div class="animate-spin rounded-full h-8 w-8 border-b-2 border-primary mx-auto"></div>
      <p class="mt-4 text-muted-foreground">Projecten laden...</p>
    </div>
  {:else if filteredProjects.length === 0 && searchTerm}
    <div class="glassmorphism-card text-center py-12">
      <div class="glassmorphism-icon mx-auto mb-4 bg-muted/30">
        <Icon name="search" size={24} className="text-muted-foreground" />
      </div>
      <h3 class="mb-2 text-lg font-medium text-foreground">Geen resultaten</h3>
      <p class="mb-6 text-sm text-muted-foreground max-w-sm mx-auto">
        Geen projecten gevonden voor "{searchTerm}". Probeer een andere zoekterm.
      </p>
      <Button
        variant="secondary"
        on:click={() => searchTerm = ''}
        class="flex items-center space-x-2"
      >
        <Icon name="x" size={16} />
        <span>Zoekterm wissen</span>
      </Button>
    </div>
  {:else if projects.length === 0}
    <div class="glassmorphism-card text-center py-12">
      <div class="glassmorphism-icon mx-auto mb-4 bg-muted/30">
        <Icon name="package" size={24} className="text-muted-foreground" />
      </div>
      <h3 class="mb-2 text-lg font-medium text-foreground">Nog geen projecten</h3>
      <p class="mb-6 text-sm text-muted-foreground max-w-sm mx-auto">
        Maak je eerste CloudBox project aan om te beginnen met je Backend-as-a-Service
      </p>
      <Button
        on:click={() => showCreateModal = true}
        class="flex items-center space-x-2"
      >
        <Icon name="plus" size={16} />
        <span>Eerste project aanmaken</span>
      </Button>
    </div>
  {:else}
    <!-- Projects Grid -->
    <div class="grid gap-6 sm:grid-cols-2 lg:grid-cols-3 xl:grid-cols-4">
      {#each filteredProjects as project}
        <ProjectCard {project} />
      {/each}
    </div>

  {/if}
</div>

<!-- Create Project Modal -->
{#if showCreateModal}
  <div class="fixed inset-0 modal-backdrop-enhanced flex items-center justify-center p-4 z-50 overflow-y-auto">
    <div class="glassmorphism-card max-w-lg w-full border-2 shadow-2xl my-8 max-h-[90vh] overflow-y-auto">
      <div class="flex items-center space-x-3 mb-6">
        <div class="glassmorphism-icon bg-primary/20">
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
            <Icon name="type" size={14} />
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
        
        <div class="space-y-2">
          <Label for="project-organization" class="flex items-center space-x-2">
            <Icon name="building" size={14} />
            <span>Organization *</span>
          </Label>
          
          
          {#if loadingOrganizations}
            <div class="flex items-center space-x-2 p-3 bg-muted rounded-md">
              <div class="animate-spin rounded-full h-4 w-4 border-b-2 border-primary"></div>
              <span class="text-sm text-muted-foreground">Organizations laden...</span>
            </div>
          {:else if organizations.length === 0}
            <div class="p-3 bg-destructive/10 border border-destructive/20 rounded-md">
              <p class="text-sm text-destructive">Geen organizations beschikbaar. Vraag een superadmin om een organization aan te maken.</p>
            </div>
          {:else}
            <select
              id="project-organization"
              bind:value={newProject.organization_id}
              required
              class="w-full px-3 py-2 border border-border rounded-md bg-background text-foreground focus:outline-none focus:ring-2 focus:ring-primary focus:border-transparent"
            >
              <option value={0}>Selecteer een organization</option>
              {#each organizations as org}
                <option value={org.id}>{org.name}</option>
              {/each}
            </select>
          {/if}
          <p class="text-xs text-muted-foreground">
            Elk project moet gekoppeld zijn aan een organization
          </p>
        </div>

        <!-- Project Features Preview -->
        <div class="glassmorphism-card bg-muted/30 p-4 space-y-3">
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
            variant="secondary"
            on:click={() => {
              showCreateModal = false;
              newProject = { name: '', description: '', organization_id: 0 };
            }}
            disabled={creating}
            class="flex-1 flex items-center justify-center space-x-2"
          >
            <Icon name="x" size={14} />
            <span>Annuleren</span>
          </Button>
          <Button
            type="submit"
            variant="primary"
            disabled={creating || !newProject.name.trim() || !newProject.organization_id}
            class="flex-1 flex items-center justify-center space-x-2"
            loading={creating}
          >
            {#if creating}
              <div class="animate-spin rounded-full h-4 w-4 border-b-2 border-white"></div>
              <span>Aanmaken...</span>
            {:else}
              <Icon name="plus" size={14} />
              <span>Project Aanmaken</span>
            {/if}
          </Button>
        </div>
      </form>
    </div>
  </div>
{/if}