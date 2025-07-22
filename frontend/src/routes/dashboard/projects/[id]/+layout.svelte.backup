<script lang="ts">
  import { page } from '$app/stores';
  import { onMount } from 'svelte';
  import { auth } from '$lib/stores/auth';
  import Icon from '$lib/components/ui/icon.svelte';
  
  interface Project {
    id: number;
    name: string;
    description: string;
    slug: string;
    created_at: string;
    is_active: boolean;
  }

  let project: Project | null = null;
  let loading = true;
  let error = '';

  $: projectId = $page.params.id;

  onMount(() => {
    loadProject();
  });

  async function loadProject() {
    loading = true;
    error = '';

    try {
      const response = await fetch(`http://localhost:8080/api/v1/projects/${projectId}`, {
        headers: {
          'Authorization': `Bearer ${$auth.token}`,
          'Content-Type': 'application/json',
        },
      });

      if (response.ok) {
        project = await response.json();
      } else {
        const data = await response.json();
        error = data.error || 'Project niet gevonden';
      }
    } catch (err) {
      error = 'Fout bij laden van project';
      console.error('Load project error:', err);
    } finally {
      loading = false;
    }
  }

  // Navigation items Appwrite-style
  const navItems = [
    { id: 'overview', name: 'Overzicht', icon: 'dashboard', href: `/dashboard/projects/${projectId}` },
    { id: 'database', name: 'Database', icon: 'database', href: `/dashboard/projects/${projectId}/database` },
    { id: 'auth', name: 'Authenticatie', icon: 'auth', href: `/dashboard/projects/${projectId}/auth` },
    { id: 'storage', name: 'Opslag', icon: 'storage', href: `/dashboard/projects/${projectId}/storage` },
    { id: 'functions', name: 'Functies', icon: 'functions', href: `/dashboard/projects/${projectId}/functions` },
    { id: 'messaging', name: 'Berichten', icon: 'messaging', href: `/dashboard/projects/${projectId}/messaging` },
    { 
      id: 'deployments', 
      name: 'Deployments', 
      icon: 'deployments', 
      href: `/dashboard/projects/${projectId}/deployments`,
      subItems: [
        { id: 'deployments', name: 'Deployments', href: `/dashboard/projects/${projectId}/deployments` },
        { id: 'ssh-keys', name: 'SSH Keys', href: `/dashboard/projects/${projectId}/ssh-keys` },
        { id: 'servers', name: 'Servers', href: `/dashboard/projects/${projectId}/servers` },
        { id: 'github', name: 'GitHub', href: `/dashboard/projects/${projectId}/github` },
      ]
    },
    { id: 'settings', name: 'Instellingen', icon: 'settings', href: `/dashboard/projects/${projectId}/settings` },
  ];

  $: currentPath = $page.url.pathname;
</script>

{#if loading}
  <div class="flex items-center justify-center min-h-64">
    <div class="text-center">
      <div class="animate-spin rounded-full h-8 w-8 border-b-2 border-primary mx-auto"></div>
      <p class="mt-4 text-muted-foreground">Project laden...</p>
    </div>
  </div>
{:else if error}
  <div class="max-w-2xl mx-auto mt-8">
    <div class="bg-destructive/10 border border-destructive/20 text-destructive px-4 py-3 rounded">
      {error}
    </div>
  </div>
{:else if project}
  <div class="min-h-screen bg-background">
    <!-- Project Header -->
    <div class="bg-card border-b border-border">
      <div class="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
        <div class="flex items-center justify-between h-16">
          <div class="flex items-center space-x-4">
            <a href="/dashboard" class="flex items-center space-x-2 text-muted-foreground hover:text-foreground">
              <Icon name="back" size={16} />
              <span>Terug naar projecten</span>
            </a>
            <div class="h-6 border-l border-border"></div>
            <div class="flex items-center space-x-3">
              <Icon name="package" size={20} />
              <div>
                <h1 class="text-xl font-semibold text-foreground">{project.name}</h1>
                <p class="text-sm text-muted-foreground">Project ID: {project.id}</p>
              </div>
            </div>
          </div>
          
          <div class="flex items-center space-x-3">
            <span class="inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-medium bg-green-100 text-green-800 dark:bg-green-900 dark:text-green-300">
              Actief
            </span>
            <button class="btn-secondary text-sm">
              Project Instellingen
            </button>
          </div>
        </div>
      </div>
    </div>

    <div class="max-w-7xl mx-auto flex">
      <!-- Sidebar Navigation - Appwrite Style -->
      <div class="w-64 bg-card border-r border-border min-h-screen">
        <nav class="mt-6 px-3">
          <div class="space-y-1">
            {#each navItems as item}
              <a
                href={item.href}
                class="group flex items-center px-3 py-2 text-sm font-medium rounded-md transition-colors duration-150 {
                  currentPath === item.href 
                    ? 'bg-primary/10 text-primary border-r-2 border-primary' 
                    : 'text-muted-foreground hover:text-foreground hover:bg-muted'
                }"
              >
                <Icon name={item.icon} size={16} className="mr-3" />
                {item.name}
              </a>
            {/each}
          </div>
        </nav>

        <!-- Project Info Sidebar -->
        <div class="mt-8 px-3">
          <div class="bg-muted rounded-lg p-4">
            <h3 class="text-sm font-medium text-foreground mb-3">Project Info</h3>
            <div class="space-y-2 text-xs">
              <div>
                <span class="text-muted-foreground">Slug:</span>
                <code class="ml-1 bg-background px-1 rounded text-foreground">{project.slug}</code>
              </div>
              <div>
                <span class="text-muted-foreground">Aangemaakt:</span>
                <span class="ml-1 text-foreground">{new Date(project.created_at).toLocaleDateString('nl-NL')}</span>
              </div>
              <div>
                <span class="text-muted-foreground">API Endpoint:</span>
                <code class="ml-1 bg-background px-1 rounded text-xs text-foreground">
                  /p/{project.slug}/api
                </code>
              </div>
            </div>
          </div>
        </div>
      </div>

      <!-- Main Content Area -->
      <div class="flex-1 bg-background">
        <main class="p-6">
          <slot />
        </main>
      </div>
    </div>
  </div>
{/if}