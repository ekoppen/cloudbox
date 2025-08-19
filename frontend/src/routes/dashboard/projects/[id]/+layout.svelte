<script lang="ts">
  import { page } from '$app/stores';
  import { onMount } from 'svelte';
  import { goto } from '$app/navigation';
  import { API_ENDPOINTS, createApiRequest } from '$lib/config';
  import { auth } from '$lib/stores/auth';
  import { theme } from '$lib/stores/theme';
  import { pluginManager, dynamicProjectMenuItems } from '$lib/stores/plugins';
  import Sidebar from '$lib/components/navigation/sidebar.svelte';
  import Icon from '$lib/components/ui/icon.svelte';
  import CloudBoxLogo from '$lib/components/ui/cloudbox-logo.svelte';
  
  function handleLogout() {
    auth.logout();
    goto('/');
  }

  function getInitials(name: string): string {
    return name
      .split(' ')
      .map(word => word.charAt(0).toUpperCase())
      .slice(0, 2)
      .join('');
  }
  
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

  // Debug project ID changes in layout
  $: {
    console.log('Layout - projectId changed:', projectId);
    console.log('Layout - page params:', $page.params);
    console.log('Layout - page URL:', $page.url.pathname);
  }

  onMount(() => {
    loadProject();
    
    // Load project-specific plugins for dynamic menu items
    if (projectId) {
      pluginManager.loadProjectPlugins(projectId);
    }
  });

  // Watch for projectId changes and reload project + plugins if needed
  $: if (projectId && projectId !== project?.id?.toString()) {
    console.log('Project ID changed, reloading project and plugins:', projectId);
    loadProject();
    pluginManager.loadProjectPlugins(projectId);
  }

  async function loadProject() {
    loading = true;
    error = '';

    try {
      const response = await createApiRequest(API_ENDPOINTS.projects.get(projectId), {
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

  // Navigation items Appwrite-style - make reactive so they update when projectId changes
  $: staticNavItems = [
    { id: 'overview', name: 'Overzicht', icon: 'dashboard', href: `/dashboard/projects/${projectId}` },
    { id: 'database', name: 'Database', icon: 'database', href: `/dashboard/projects/${projectId}/database` },
    { id: 'auth', name: 'Authenticatie', icon: 'auth', href: `/dashboard/projects/${projectId}/auth` },
    { id: 'storage', name: 'Opslag', icon: 'storage', href: `/dashboard/projects/${projectId}/storage` },
    { id: 'functions', name: 'Functies', icon: 'functions', href: `/dashboard/projects/${projectId}/functions` },
    { id: 'api', name: 'API', icon: 'settings', href: `/dashboard/projects/${projectId}/api` },
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
    { 
      id: 'settings', 
      name: 'Instellingen', 
      icon: 'settings', 
      href: `/dashboard/projects/${projectId}/settings`
    },
  ];

  // Combine static nav items with dynamic plugin items
  $: navItems = [
    ...staticNavItems.slice(0, 2), // overview, database
    ...($dynamicProjectMenuItems || []).map(plugin => ({
      ...plugin,
      href: plugin.href.replace('{projectId}', projectId)
    })), // Plugin items after database
    ...staticNavItems.slice(2) // rest of static items
  ];

  $: currentPath = $page.url.pathname;

  // Handle clicking outside dropdown to close it
  function handleOutsideClick(event: MouseEvent) {
    const dropdown = document.getElementById('user-dropdown');
    const button = event.target as Element;
    
    if (dropdown && !dropdown.contains(button) && !button.closest('button')) {
      dropdown.classList.add('hidden');
    }
  }

  onMount(() => {
    document.addEventListener('click', handleOutsideClick);
    return () => {
      document.removeEventListener('click', handleOutsideClick);
    };
  });
</script>

{#if loading}
  <div class="min-h-screen bg-background flex items-center justify-center">
    <div class="text-center">
      <div class="animate-spin rounded-full h-12 w-12 border-b-2 border-primary mx-auto"></div>
      <p class="mt-4 text-muted-foreground">Project laden...</p>
    </div>
  </div>
{:else if error}
  <div class="min-h-screen bg-background flex items-center justify-center">
    <div class="max-w-2xl mx-auto text-center">
      <div class="bg-destructive/10 border border-destructive/20 text-destructive px-6 py-8 rounded-lg">
        <Icon name="alert-triangle" size={48} className="text-destructive mx-auto mb-4" />
        <h2 class="text-xl font-semibold text-foreground mb-2">Project Niet Gevonden</h2>
        <p class="text-destructive mb-6">{error}</p>
        <div class="flex items-center justify-center space-x-4">
          <a 
            href="/dashboard" 
            class="inline-flex items-center px-4 py-2 rounded-md bg-primary text-primary-foreground hover:bg-primary/90 transition-colors"
          >
            <Icon name="back" size={16} className="mr-2" />
            Terug naar Dashboard
          </a>
          <button 
            on:click={loadProject}
            class="inline-flex items-center px-4 py-2 rounded-md border border-border bg-background text-foreground hover:bg-muted transition-colors"
          >
            <Icon name="refresh-cw" size={16} className="mr-2" />
            Opnieuw Proberen
          </button>
        </div>
      </div>
    </div>
  </div>
{:else if project}
  <div class="min-h-screen bg-background">
    <!-- Supabase-style Project Sidebar -->
    <Sidebar 
      context="project" 
      projectId={projectId}
      projectName={project.name}
    />

    <!-- Main content with sidebar offset -->
    <div class="transition-all duration-200 ease-in-out min-h-screen ml-sidebar-collapsed">
      <!-- Unified Project Header with proper spacing and alignment -->
      <header class="sticky top-0 z-40 bg-background/95 backdrop-blur supports-[backdrop-filter]:bg-background/60 border-b border-border">
        <div class="flex h-16 items-center justify-between px-6">
          <!-- Left side: Back navigation + Breadcrumbs -->
          <div class="flex items-center space-x-4">
            <a href="/dashboard" class="flex items-center space-x-2 text-muted-foreground hover:text-foreground transition-colors icon-contrast">
              <Icon name="arrow-left" size={16} className="icon-contrast" />
              <span class="text-sm font-medium">Dashboard</span>
            </a>
            <div class="h-4 border-l border-border"></div>
            <div class="flex items-center space-x-3">
              <CloudBoxLogo size="md" showText={false} />
              <div>
                <h1 class="text-lg font-semibold text-foreground leading-tight">{project.name}</h1>
                <p class="text-xs text-muted-foreground">ID: {project.id} • {project.slug}</p>
              </div>
            </div>
          </div>
          
          <!-- Right side: Status + Theme Toggle + User menu -->
          <div class="flex items-center space-x-4">
            <!-- Project Status -->
            <span class="inline-flex items-center px-3 py-1.5 rounded-full text-xs font-medium border {project.is_active ? 'bg-success/10 text-success border-success/20' : 'bg-destructive/10 text-destructive border-destructive/20'}">
              <div class="w-1.5 h-1.5 rounded-full mr-2 {project.is_active ? 'bg-success' : 'bg-destructive'}"></div>
              {project.is_active ? 'Actief' : 'Inactief'}
            </span>
            
            <!-- Theme Toggle -->
            <button
              on:click={() => theme.toggleTheme()}
              class="theme-toggle-btn flex items-center justify-center w-9 h-9 rounded-lg shadow-sm focus-visible"
              title="Schakel tussen licht en donker thema"
            >
              {#if $theme.theme === 'cloudbox-dark'}
                <Icon name="sun" size={16} className="text-foreground" />
              {:else}
                <Icon name="moon" size={16} className="text-foreground" />
              {/if}
            </button>
            
            <!-- User Dropdown -->
            <div class="relative">
              <button
                class="flex items-center space-x-2 px-3 py-2 rounded-lg border border-border bg-background hover:bg-muted transition-all duration-200 hover:scale-105 shadow-sm"
                on:click={() => {
                  const dropdown = document.getElementById('user-dropdown');
                  dropdown?.classList.toggle('hidden');
                }}
              >
                <div class="w-7 h-7 rounded-full bg-primary text-primary-foreground text-xs font-semibold flex items-center justify-center">
                  <span>{getInitials($auth.user?.name || 'User')}</span>
                </div>
                <Icon name="chevron-down" size={14} className="text-muted-foreground" />
              </button>
              
              <div id="user-dropdown" class="user-dropdown-menu absolute right-0 top-full mt-2 w-64 rounded-xl py-2 z-50 hidden">
                <div class="px-4 py-3 border-b border-border">
                  <div class="font-semibold text-card-foreground">{$auth.user?.name}</div>
                  <div class="text-sm text-muted-foreground">{$auth.user?.email}</div>
                </div>
                
                <div class="py-2">
                  <a href="/dashboard/settings" class="flex items-center space-x-3 px-4 py-2 text-sm text-card-foreground hover:bg-muted transition-colors">
                    <Icon name="user" size={16} />
                    <span>Profiel</span>
                  </a>
                  
                  <a href="/dashboard/projects/{projectId}/settings" class="flex items-center space-x-3 px-4 py-2 text-sm text-card-foreground hover:bg-muted transition-colors">
                    <Icon name="settings" size={16} />
                    <span>Project Instellingen</span>
                  </a>
                  
                  {#if $auth.user?.role === 'superadmin'}
                    <a href="/dashboard/admin" class="flex items-center space-x-3 px-4 py-2 text-sm text-card-foreground hover:bg-muted transition-colors">
                      <Icon name="shield-check" size={16} />
                      <span>Admin Panel</span>
                    </a>
                  {/if}
                  
                  <div class="border-t border-border my-2"></div>
                  
                  <button 
                    on:click={handleLogout} 
                    class="flex items-center space-x-3 px-4 py-2 text-sm text-destructive hover:bg-destructive/10 transition-colors w-full text-left"
                  >
                    <Icon name="log-out" size={16} />
                    <span>Uitloggen</span>
                  </button>
                </div>
              </div>
            </div>
          </div>
        </div>
      </header>

      <!-- Page content -->
      <main class="p-6">
        <slot />
      </main>
    </div>
  </div>
{/if}