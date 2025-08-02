<script lang="ts">
  import { onMount } from 'svelte';
  import { page } from '$app/stores';
  import { goto } from '$app/navigation';
  import { auth } from '$lib/stores/auth';
  import { toastStore } from '$lib/stores/toast';
  import { API_ENDPOINTS, createApiRequest } from '$lib/config';
  import Card from '$lib/components/ui/card.svelte';
  import Button from '$lib/components/ui/button.svelte';
  import Badge from '$lib/components/ui/badge.svelte';
  import Icon from '$lib/components/ui/icon.svelte';

  interface Organization {
    id: number;
    name: string;
    description: string;
    color: string;
    is_active: boolean;
    project_count: number;
    created_at: string;
  }

  interface Project {
    id: number;
    name: string;
    description: string;
    slug: string;
    created_at: string;
    is_active: boolean;
  }

  let orgId: string;
  let organization: Organization | null = null;
  let projects: Project[] = [];
  let loading = true;
  let loadingProjects = true;

  $: orgId = $page.params.id;

  onMount(async () => {
    if (orgId) {
      await Promise.all([
        loadOrganization(),
        loadOrganizationProjects()
      ]);
    }
  });

  async function loadOrganization() {
    try {
      const response = await createApiRequest(API_ENDPOINTS.organizations.get(orgId), {
        headers: {
          'Authorization': `Bearer ${$auth.token}`,
        },
      });

      if (response.ok) {
        organization = await response.json();
      } else {
        toastStore.error('Organization niet gevonden');
        goto('/dashboard/organizations');
      }
    } catch (err) {
      toastStore.error('Fout bij laden van organization');
      goto('/dashboard/organizations');
    } finally {
      loading = false;
    }
  }

  async function loadOrganizationProjects() {
    try {
      const response = await createApiRequest(API_ENDPOINTS.organizations.projects(orgId), {
        headers: {
          'Authorization': `Bearer ${$auth.token}`,
        },
      });

      if (response.ok) {
        projects = await response.json();
      } else {
        console.error('Failed to load organization projects');
      }
    } catch (err) {
      console.error('Error loading organization projects:', err);
    } finally {
      loadingProjects = false;
    }
  }

  function formatDate(dateStr: string): string {
    return new Date(dateStr).toLocaleDateString('nl-NL', {
      year: 'numeric',
      month: 'long',
      day: 'numeric'
    });
  }
</script>

<svelte:head>
  <title>{organization?.name || 'Organization'} - CloudBox</title>
</svelte:head>

<div class="space-y-6">
  {#if loading}
    <Card class="p-12">
      <div class="text-center">
        <div class="w-8 h-8 border-4 border-primary border-t-transparent rounded-full animate-spin mx-auto mb-4"></div>
        <p class="text-muted-foreground">Organization laden...</p>
      </div>
    </Card>
  {:else if organization}
    <!-- Header -->
    <div class="flex items-center justify-between">
      <div class="flex items-center space-x-4">
        <Button
          href="/dashboard/organizations"
          variant="outline"
          size="sm"
          class="flex items-center space-x-2"
        >
          <Icon name="arrow-left" size={16} />
          <span>Terug naar Organizations</span>
        </Button>
      </div>
    </div>

    <!-- Organization Details -->
    <Card class="p-6">
      <div class="flex items-start justify-between mb-6">
        <div class="flex items-center space-x-4">
          <div 
            class="w-16 h-16 rounded-xl flex items-center justify-center text-white"
            style="background-color: {organization.color}"
          >
            <Icon name="building" size={32} />
          </div>
          <div>
            <h1 class="text-3xl font-bold text-foreground">{organization.name}</h1>
            {#if organization.description}
              <p class="text-muted-foreground mt-2 max-w-2xl">
                {organization.description}
              </p>
            {/if}
            <div class="flex items-center space-x-4 mt-3">
              <Badge variant={organization.is_active ? "default" : "secondary"} class="flex items-center space-x-1">
                <div class="w-2 h-2 rounded-full {organization.is_active ? 'bg-green-500' : 'bg-gray-400'}"></div>
                <span>{organization.is_active ? 'Actief' : 'Inactief'}</span>
              </Badge>
              <span class="text-sm text-muted-foreground">
                Aangemaakt op {formatDate(organization.created_at)}
              </span>
            </div>
          </div>
        </div>
      </div>

      <!-- Stats -->
      <div class="grid grid-cols-1 md:grid-cols-3 gap-4 mb-6">
        <div class="bg-muted/50 rounded-lg p-4">
          <div class="flex items-center space-x-2 text-sm font-medium text-foreground mb-1">
            <Icon name="package" size={16} className="text-primary" />
            <span>Totaal Projecten</span>
          </div>
          <p class="text-2xl font-bold text-foreground">{projects.length}</p>
        </div>
        <div class="bg-muted/50 rounded-lg p-4">
          <div class="flex items-center space-x-2 text-sm font-medium text-foreground mb-1">
            <Icon name="check-circle" size={16} className="text-green-600" />
            <span>Actieve Projecten</span>
          </div>
          <p class="text-2xl font-bold text-foreground">{projects.filter(p => p.is_active).length}</p>
        </div>
        <div class="bg-muted/50 rounded-lg p-4">
          <div class="flex items-center space-x-2 text-sm font-medium text-foreground mb-1">
            <Icon name="pause-circle" size={16} className="text-gray-500" />
            <span>Inactieve Projecten</span>
          </div>
          <p class="text-2xl font-bold text-foreground">{projects.filter(p => !p.is_active).length}</p>
        </div>
      </div>
    </Card>

    <!-- Projects Section -->
    <div>
      <div class="flex items-center justify-between mb-4">
        <h2 class="text-xl font-semibold text-foreground">Projecten in deze Organization</h2>
        <Button
          href="/dashboard/projects"
          size="sm"
          class="flex items-center space-x-2"
        >
          <Icon name="plus" size={14} />
          <span>Nieuw Project</span>
        </Button>
      </div>

      {#if loadingProjects}
        <Card class="p-8">
          <div class="text-center">
            <div class="w-6 h-6 border-4 border-primary border-t-transparent rounded-full animate-spin mx-auto mb-4"></div>
            <p class="text-muted-foreground">Projecten laden...</p>
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
              Deze organization heeft nog geen projecten. Maak je eerste project aan.
            </p>
            <Button
              href="/dashboard/projects"
              size="lg"
              class="flex items-center space-x-2"
            >
              <Icon name="plus" size={16} />
              <span>Eerste Project Aanmaken</span>
            </Button>
          </div>
        </Card>
      {:else}
        <div class="grid gap-4">
          {#each projects as project}
            <Card class="group p-4 hover:shadow-md hover:shadow-primary/5 transition-all duration-200 border hover:border-primary/20">
              <div class="flex items-center justify-between">
                <div class="flex items-center space-x-3">
                  <div class="w-10 h-10 bg-primary/10 rounded-lg flex items-center justify-center">
                    <Icon name="package" size={20} className="text-primary" />
                  </div>
                  <div>
                    <h3 class="font-semibold text-foreground group-hover:text-primary transition-colors">
                      {project.name}
                    </h3>
                    {#if project.description}
                      <p class="text-sm text-muted-foreground line-clamp-1">
                        {project.description}
                      </p>
                    {/if}
                    <div class="flex items-center space-x-3 text-xs text-muted-foreground mt-1">
                      <span>Slug: <code class="bg-muted px-1 rounded font-mono">{project.slug}</code></span>
                      <span>•</span>
                      <span>{formatDate(project.created_at)}</span>
                    </div>
                  </div>
                </div>
                <div class="flex items-center space-x-3">
                  <Badge variant={project.is_active ? "default" : "secondary"} class="flex items-center space-x-1">
                    <div class="w-2 h-2 rounded-full {project.is_active ? 'bg-green-500' : 'bg-gray-400'}"></div>
                    <span>{project.is_active ? 'Actief' : 'Inactief'}</span>
                  </Badge>
                  <Button
                    href="/dashboard/projects/{project.id}"
                    size="sm"
                    variant="outline"
                    class="flex items-center space-x-1"
                  >
                    <Icon name="settings" size={14} />
                    <span>Beheren</span>
                  </Button>
                </div>
              </div>
            </Card>
          {/each}
        </div>
      {/if}
    </div>
  {/if}
</div>