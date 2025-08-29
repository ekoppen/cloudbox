<script lang="ts">
  import { page } from '$app/stores';
  import Button from '$lib/components/ui/button.svelte';
  import Icon from '$lib/components/ui/icon.svelte';

  $: projectId = $page.params.id;
  $: currentPath = $page.url.pathname;

  // Check if current route is active
  function isActive(path: string): boolean {
    if (path === '/endpoints') {
      return currentPath === `/dashboard/projects/${projectId}/api` || currentPath === `/dashboard/projects/${projectId}/api/endpoints`;
    }
    return currentPath.includes(path);
  }

  // Navigation items for API section
  const navItems = [
    {
      label: 'Endpoints',
      href: 'endpoints',
      icon: 'settings',
      description: 'API endpoints en documentatie'
    },
    {
      label: 'Statistics',
      href: 'stats', 
      icon: 'trending-up',
      description: 'API gebruiksstatistieken'
    }
  ];
</script>

<style>
  .glassmorphism-nav {
    background: rgba(255, 255, 255, 0.8);
    backdrop-filter: blur(16px);
    -webkit-backdrop-filter: blur(16px);
    border: 1px solid rgba(255, 255, 255, 0.2);
  }

  :global(.cloudbox-dark) .glassmorphism-nav {
    background: rgba(38, 38, 38, 0.8);
    border: 1px solid rgba(255, 255, 255, 0.1);
  }
</style>

<div class="space-y-6">
  <!-- API Section Header -->
  <div class="flex items-center space-x-4">
    <div class="w-12 h-12 bg-primary rounded-xl flex items-center justify-center">
      <Icon name="zap" size={24} color="white" />
    </div>
    <div>
      <h1 class="text-2xl font-bold text-foreground">API Management</h1>
      <p class="mt-1 text-sm text-muted-foreground">
        Beheer je API endpoints en bekijk gebruiksstatistieken
      </p>
    </div>
  </div>

  <!-- Navigation Tabs -->
  <div class="glassmorphism-nav rounded-lg p-2">
    <nav class="flex space-x-2">
      {#each navItems as item}
        <Button
          href="/dashboard/projects/{projectId}/api/{item.href}"
          variant={isActive(item.href) ? 'default' : 'ghost'}
          size="sm"
          class="flex items-center space-x-2 transition-all duration-200 {isActive(item.href) ? 'shadow-sm' : 'hover:bg-background/50'}"
        >
          <Icon name={item.icon} size={16} />
          <span>{item.label}</span>
        </Button>
      {/each}
    </nav>
  </div>

  <!-- Page Content -->
  <slot />
</div>