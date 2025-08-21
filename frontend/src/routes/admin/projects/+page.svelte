<script lang="ts">
  import { onMount } from 'svelte';
  import { API_ENDPOINTS, API_BASE_URL, createApiRequest } from '$lib/config';
  import { auth } from '$lib/stores/auth';
  import { toast } from '$lib/stores/toast';
  import Card from '$lib/components/ui/card.svelte';
  import Button from '$lib/components/ui/button.svelte';
  import Input from '$lib/components/ui/input.svelte';
  import Label from '$lib/components/ui/label.svelte';
  import Badge from '$lib/components/ui/badge.svelte';
  import Icon from '$lib/components/ui/icon.svelte';

  interface Organization {
    id: number;
    name: string;
    color: string;
  }

  interface User {
    id: number;
    email: string;
    name: string;
  }

  interface Project {
    id: number;
    name: string;
    description: string;
    slug: string;
    created_at: string;
    updated_at: string;
    is_active: boolean;
    organization?: Organization;
    user?: User;
  }

  let projects: Project[] = [];
  let loading = true;
  let error = '';
  let searchTerm = '';
  let selectedStatus = 'all';

  const statusOptions = [
    { value: 'all', label: 'Alle statussen' },
    { value: 'active', label: 'Actief' },
    { value: 'inactive', label: 'Inactief' }
  ];

  onMount(() => {
    loadProjects();
  });

  async function loadProjects() {
    loading = true;
    error = '';

    try {
      const response = await createApiRequest(API_ENDPOINTS.admin.projects.list, {
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

  async function toggleProjectStatus(project: Project) {
    try {
      const response = await fetch(`${API_BASE_URL}/api/v1/admin/projects/${project.id}/status`, {
        method: 'PUT',
        headers: {
          'Authorization': `Bearer ${$auth.token}`,
          'Content-Type': 'application/json',
        },
        body: JSON.stringify({
          is_active: !project.is_active
        }),
      });

      if (response.ok) {
        const updatedProject = await response.json();
        projects = projects.map(p => p.id === project.id ? updatedProject : p);
        toast.success(`Project ${project.is_active ? 'gedeactiveerd' : 'geactiveerd'}`);
      } else {
        const data = await response.json();
        error = data.error || 'Fout bij wijzigen project status';
      }
    } catch (err) {
      error = 'Netwerkfout bij wijzigen project status';
      console.error('Toggle project status error:', err);
    }
  }

  function formatDate(dateStr: string) {
    return new Date(dateStr).toLocaleDateString('nl-NL', {
      year: 'numeric',
      month: 'short',
      day: 'numeric',
      hour: '2-digit',
      minute: '2-digit'
    });
  }

  // Filter projects based on search term and status
  $: filteredProjects = projects.filter(project => {
    const matchesSearch = project.name.toLowerCase().includes(searchTerm.toLowerCase()) ||
                         project.description.toLowerCase().includes(searchTerm.toLowerCase()) ||
                         project.slug.toLowerCase().includes(searchTerm.toLowerCase()) ||
                         project.user?.name.toLowerCase().includes(searchTerm.toLowerCase()) ||
                         project.user?.email.toLowerCase().includes(searchTerm.toLowerCase());
    const matchesStatus = selectedStatus === 'all' || 
                         (selectedStatus === 'active' && project.is_active) ||
                         (selectedStatus === 'inactive' && !project.is_active);
    
    return matchesSearch && matchesStatus;
  });

  $: projectStats = {
    total: projects.length,
    active: projects.filter(p => p.is_active).length,
    inactive: projects.filter(p => !p.is_active).length,
    withOrg: projects.filter(p => p.organization).length
  };
</script>

<svelte:head>
  <title>Projecten Beheer - CloudBox Admin</title>
</svelte:head>

<div class="space-y-8">
  <!-- Header -->
  <div class="flex items-center justify-between">
    <div class="flex items-center space-x-4">
      <div class="w-12 h-12 bg-primary rounded-xl flex items-center justify-center">
        <Icon name="package" size={24} color="white" />
      </div>
      <div>
        <h1 class="text-3xl font-bold text-foreground">Projecten Beheer</h1>
        <p class="text-muted-foreground">
          Beheer alle CloudBox projecten van alle gebruikers
        </p>
      </div>
    </div>
    <div class="flex items-center space-x-3">
      <Button
        variant="outline"
        href="/admin"
        size="lg"
        class="flex items-center space-x-2"
      >
        <Icon name="backup" size={16} />
        <span>Terug naar Dashboard</span>
      </Button>
      <Button
        on:click={loadProjects}
        variant="ghost"
        size="icon"
        class="hover:rotate-180 transition-transform duration-300"
        title="Vernieuw projectenlijst"
      >
        <Icon name="refresh-cw" size={16} />
      </Button>
    </div>
  </div>

  <!-- Project Stats -->
  <div class="grid grid-cols-1 md:grid-cols-4 gap-6">
    <Card class="p-6">
      <div class="flex items-center justify-between">
        <div>
          <p class="text-sm font-medium text-muted-foreground">Totaal Projecten</p>
          <p class="text-2xl font-bold text-foreground">{projectStats.total}</p>
        </div>
        <div class="w-10 h-10 bg-blue-100 dark:bg-gray-800 rounded-lg flex items-center justify-center">
          <Icon name="package" size={20} className="text-blue-600 dark:text-blue-400" />
        </div>
      </div>
    </Card>
    
    <Card class="p-6">
      <div class="flex items-center justify-between">
        <div>
          <p class="text-sm font-medium text-muted-foreground">Actieve Projecten</p>
          <p class="text-2xl font-bold text-green-600">{projectStats.active}</p>
        </div>
        <div class="w-10 h-10 bg-green-100 dark:bg-green-900 rounded-lg flex items-center justify-center">
          <Icon name="check" size={20} className="text-green-600 dark:text-green-400" />
        </div>
      </div>
    </Card>
    
    <Card class="p-6">
      <div class="flex items-center justify-between">
        <div>
          <p class="text-sm font-medium text-muted-foreground">Inactieve Projecten</p>
          <p class="text-2xl font-bold text-gray-600">{projectStats.inactive}</p>
        </div>
        <div class="w-10 h-10 bg-gray-100 dark:bg-gray-900 rounded-lg flex items-center justify-center">
          <Icon name="x" size={20} className="text-gray-600 dark:text-gray-400" />
        </div>
      </div>
    </Card>
    
    <Card class="p-6">
      <div class="flex items-center justify-between">
        <div>
          <p class="text-sm font-medium text-muted-foreground">Met Organization</p>
          <p class="text-2xl font-bold text-purple-600">{projectStats.withOrg}</p>
        </div>
        <div class="w-10 h-10 bg-purple-100 dark:bg-purple-900 rounded-lg flex items-center justify-center">
          <Icon name="building" size={20} className="text-purple-600 dark:text-purple-400" />
        </div>
      </div>
    </Card>
  </div>

  <!-- Filters -->
  <Card class="p-6">
    <div class="grid grid-cols-1 md:grid-cols-3 gap-4">
      <div>
        <Label for="search">Zoeken</Label>
        <div class="relative">
          <Icon name="search" size={16} className="absolute left-3 top-1/2 transform -translate-y-1/2 text-muted-foreground" />
          <Input
            id="search"
            type="text"
            placeholder="Naam, slug, gebruiker..."
            bind:value={searchTerm}
            class="pl-10"
          />
        </div>
      </div>
      <div>
        <Label for="status">Status</Label>
        <select 
          id="status"
          bind:value={selectedStatus}
          class="w-full px-3 py-2 border border-input rounded-md bg-background text-foreground focus:outline-none focus:ring-2 focus:ring-ring focus:border-transparent"
        >
          {#each statusOptions as status}
            <option value={status.value}>{status.label}</option>
          {/each}
        </select>
      </div>
      <div class="flex items-end">
        <Button
          variant="outline"
          on:click={() => {
            searchTerm = '';
            selectedStatus = 'all';
          }}
          class="flex items-center space-x-2"
        >
          <Icon name="x" size={16} />
          <span>Wissen</span>
        </Button>
      </div>
    </div>
  </Card>

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
  {:else if filteredProjects.length === 0}
    <Card class="p-12">
      <div class="text-center">
        <div class="w-16 h-16 bg-muted rounded-full flex items-center justify-center mx-auto mb-4">
          <Icon name="package" size={32} className="text-muted-foreground" />
        </div>
        <h3 class="text-lg font-medium text-foreground mb-2">Geen projecten gevonden</h3>
        <p class="text-muted-foreground mb-6 max-w-sm mx-auto">
          {searchTerm || selectedStatus !== 'all' 
            ? 'Geen projecten voldoen aan de filteropties.' 
            : 'Er zijn nog geen projecten aangemaakt.'}
        </p>
      </div>
    </Card>
  {:else}
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
                <div class="flex items-center space-x-1 text-xs">
                  <Icon name="user" size={12} className="text-muted-foreground" />
                  <span class="text-muted-foreground">{project.user?.name}</span>
                </div>
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
          
          {#if project.organization}
            <div class="flex items-center space-x-2 mb-4">
              <div 
                class="w-3 h-3 rounded-full"
                style="background-color: {project.organization.color}"
              ></div>
              <span class="text-sm text-muted-foreground">{project.organization.name}</span>
            </div>
          {/if}
          
          <div class="bg-muted/50 rounded-lg p-3 mb-4 space-y-2">
            <div class="flex items-center justify-between text-xs">
              <span class="text-muted-foreground">API Slug:</span>
              <code class="bg-background px-2 py-1 rounded text-xs font-mono">{project.slug}</code>
            </div>
            <div class="flex items-center justify-between text-xs">
              <span class="text-muted-foreground">Eigenaar:</span>
              <span class="text-foreground">{project.user?.email}</span>
            </div>
            <div class="flex items-center justify-between text-xs">
              <span class="text-muted-foreground">Aangemaakt:</span>
              <span class="text-foreground">{formatDate(project.created_at)}</span>
            </div>
          </div>
          
          <div class="flex space-x-2">
            <Button
              variant="outline"
              size="sm"
              on:click={() => toggleProjectStatus(project)}
              class="flex-1 flex items-center justify-center space-x-2"
            >
              <Icon name={project.is_active ? 'x' : 'check'} size={14} />
              <span>{project.is_active ? 'Deactiveren' : 'Activeren'}</span>
            </Button>
            <Button
              href="/dashboard/projects/{project.id}"
              size="sm"
              class="flex items-center justify-center space-x-2"
            >
              <Icon name="external-link" size={14} />
              <span>Bekijken</span>
            </Button>
          </div>
        </Card>
      {/each}
    </div>
  {/if}
</div>