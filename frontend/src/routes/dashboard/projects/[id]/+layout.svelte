<script lang="ts">
  import { page } from '$app/stores';
  import { onMount } from 'svelte';
  import { goto } from '$app/navigation';
  import { API_ENDPOINTS, createApiRequest } from '$lib/config';
  import { auth } from '$lib/stores/auth';
  import { pluginManager, dynamicProjectMenuItems } from '$lib/stores/plugins';
  import { sidebarStore } from '$lib/stores/sidebar';
  import Button from '$lib/components/ui/button.svelte';
  import Icon from '$lib/components/ui/icon.svelte';
  import CloudBoxLogo from '$lib/components/ui/cloudbox-logo.svelte';
  
  
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
    
    // Load project-specific plugins for dynamic menu items
    if (projectId) {
      pluginManager.loadProjectPlugins(projectId);
    }
  });

  // Watch for projectId changes and reload project + plugins if needed
  $: if (projectId && projectId !== project?.id?.toString()) {
    loadProject();
    pluginManager.loadProjectPlugins(projectId);
  }

  // Update sidebar context when project loads
  $: if (project && projectId) {
    sidebarStore.setContext('project', projectId, project.name);
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

  $: currentPath = $page.url.pathname;

</script>

<style>
  .glassmorphism-header {
    background: rgba(255, 255, 255, 0.9);
    backdrop-filter: blur(20px);
    -webkit-backdrop-filter: blur(20px);
    border-bottom: 1px solid rgba(255, 255, 255, 0.2);
  }

  :global(.cloudbox-dark) .glassmorphism-header {
    background: rgba(38, 38, 38, 0.9);
    border-bottom: 1px solid rgba(255, 255, 255, 0.1);
  }
</style>

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
          <Button 
            on:click={loadProject}
            variant="ghost"
            size="icon"
            class="hover:rotate-180 transition-transform duration-300"
            title="Herlaad project"
          >
            <Icon name="refresh-cw" size={16} />
          </Button>
        </div>
      </div>
    </div>
  </div>
{:else if project}
  <div class="min-h-screen bg-background group">
    <!-- Main content with sidebar offset - sidebar is handled by parent dashboard layout -->
    <div class="transition-all duration-200 ease-in-out min-h-screen">
      <!-- Enhanced Project Header with Glassmorphism -->
      <header class="sticky top-0 z-40 glassmorphism-header backdrop-blur-xl shadow-lg">
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
          
          <!-- Right side: Project Status only -->
          <div class="flex items-center">
            <!-- Project Status -->
            <span class="inline-flex items-center px-3 py-1.5 rounded-full text-xs font-medium border {project.is_active ? 'bg-success/10 text-success border-success/20' : 'bg-destructive/10 text-destructive border-destructive/20'}">
              <div class="w-1.5 h-1.5 rounded-full mr-2 {project.is_active ? 'bg-success' : 'bg-destructive'}"></div>
              {project.is_active ? 'Actief' : 'Inactief'}
            </span>
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