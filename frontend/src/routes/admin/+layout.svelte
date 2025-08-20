<script lang="ts">
  import { onMount } from 'svelte';
  import { goto } from '$app/navigation';
  import { page } from '$app/stores';
  import { auth } from '$lib/stores/auth';
  import { toast } from '$lib/stores/toast';
  import Button from '$lib/components/ui/button.svelte';
  import Icon from '$lib/components/ui/icon.svelte';

  let isLoading = true;
  let hasAccess = false;

  onMount(async () => {
    console.log('Admin layout: initializing auth...');
    await auth.init();
    
    // Check if user is authenticated and is admin
    if (!$auth.isAuthenticated) {
      console.log('Admin layout: not authenticated, redirecting to login');
      toast.error('Je moet ingelogd zijn om admin pagina\'s te bezoeken');
      goto('/login');
      return;
    }

    // Check if user has admin privileges
    // For now, we'll allow any authenticated user to access admin
    // In production, you'd check for admin role
    hasAccess = true;
    isLoading = false;

    console.log('Admin layout: access granted');
  });

  function handleBackToDashboard() {
    goto('/dashboard');
  }

  function handleLogout() {
    auth.logout();
    goto('/');
  }
</script>

{#if isLoading}
  <div class="min-h-screen bg-gray-50 dark:bg-gray-900 flex items-center justify-center">
    <div class="text-center">
      <div class="animate-spin rounded-full h-12 w-12 border-b-2 border-blue-500 mx-auto mb-4"></div>
      <p class="text-gray-600 dark:text-gray-400">Admin toegang controleren...</p>
    </div>
  </div>
{:else if !hasAccess}
  <div class="min-h-screen bg-gray-50 dark:bg-gray-900 flex items-center justify-center">
    <div class="text-center max-w-md">
      <Icon name="shield-x" class="w-16 h-16 text-red-500 mx-auto mb-4" />
      <h1 class="text-2xl font-bold text-gray-900 dark:text-white mb-2">
        Geen toegang
      </h1>
      <p class="text-gray-600 dark:text-gray-400 mb-6">
        Je hebt geen admin rechten om deze pagina te bezoeken.
      </p>
      <Button on:click={handleBackToDashboard}>
        Terug naar Dashboard
      </Button>
    </div>
  </div>
{:else}
  <!-- Admin Header -->
  <div class="bg-white dark:bg-gray-800 border-b border-gray-200 dark:border-gray-700">
    <div class="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
      <div class="flex justify-between items-center h-16">
        <div class="flex items-center">
          <div class="flex-shrink-0">
            <Icon name="shield-check" class="w-8 h-8 text-blue-600" />
          </div>
          <div class="ml-3">
            <h1 class="text-lg font-semibold text-gray-900 dark:text-white">
              CloudBox Admin
            </h1>
          </div>
        </div>

        <div class="flex items-center space-x-4">
          <div class="text-sm text-gray-600 dark:text-gray-400">
            Ingelogd als: <span class="font-medium">{$auth.user?.name}</span>
          </div>
          
          <Button variant="secondary" size="sm" on:click={handleBackToDashboard}>
            <Icon name="arrow-left" class="w-4 h-4 mr-2" />
            Dashboard
          </Button>
          
          <Button variant="ghost" size="sm" on:click={handleLogout}>
            <Icon name="log-out" class="w-4 h-4 mr-2" />
            Uitloggen
          </Button>
        </div>
      </div>
    </div>
  </div>

  <!-- Admin Navigation -->
  <nav class="bg-gray-50 dark:bg-gray-900 border-b border-gray-200 dark:border-gray-700">
    <div class="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
      <div class="flex space-x-8 h-12 items-center">
        <a 
          href="/admin" 
          class="flex items-center space-x-2 px-3 py-2 rounded-md text-sm font-medium transition-colors
                 hover:bg-gray-100 dark:hover:bg-gray-800 
                 {$page.url.pathname === '/admin' ? 'text-primary bg-primary/10' : 'text-muted-foreground'}"
        >
          <Icon name="dashboard" size={16} />
          <span>Dashboard</span>
        </a>
        
        <a 
          href="/admin/users" 
          class="flex items-center space-x-2 px-3 py-2 rounded-md text-sm font-medium transition-colors
                 hover:bg-gray-100 dark:hover:bg-gray-800 
                 {$page.url.pathname === '/admin/users' ? 'text-primary bg-primary/10' : 'text-muted-foreground'}"
        >
          <Icon name="users" size={16} />
          <span>Gebruikers</span>
        </a>
        
        <a 
          href="/admin/projects" 
          class="flex items-center space-x-2 px-3 py-2 rounded-md text-sm font-medium transition-colors
                 hover:bg-gray-100 dark:hover:bg-gray-800 
                 {$page.url.pathname === '/admin/projects' ? 'text-primary bg-primary/10' : 'text-muted-foreground'}"
        >
          <Icon name="package" size={16} />
          <span>Projecten</span>
        </a>
        
        <a 
          href="/admin/system" 
          class="flex items-center space-x-2 px-3 py-2 rounded-md text-sm font-medium transition-colors
                 hover:bg-gray-100 dark:hover:bg-gray-800 
                 {$page.url.pathname === '/admin/system' ? 'text-primary bg-primary/10' : 'text-muted-foreground'}"
        >
          <Icon name="settings" size={16} />
          <span>Systeem</span>
        </a>
      </div>
    </div>
  </nav>

  <!-- Admin Content -->
  <main class="admin-dashboard min-h-screen bg-background p-6">
    <div class="max-w-7xl mx-auto">
      <slot />
    </div>
  </main>
{/if}

<style>
  :global(.admin-dashboard) {
    min-height: calc(100vh - 65px); /* Subtract header height */
  }
</style>