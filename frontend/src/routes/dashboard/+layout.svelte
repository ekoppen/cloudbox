<script lang="ts">
  import { onMount } from 'svelte';
  import { goto } from '$app/navigation';
  import { page } from '$app/stores';
  import { auth, type User } from '$lib/stores/auth';
  import { theme } from '$lib/stores/theme';
  import { navigation } from '$lib/stores/navigation';
  import Icon from '$lib/components/ui/icon.svelte';
  
  let user: User | null = null;
  let isLoading = true;
  let authInitialized = false;

  // Subscribe to auth state
  $: $auth, handleAuthChange();

  function handleAuthChange() {
    user = $auth.user;
    isLoading = $auth.isLoading;
    
    // Only redirect after auth has been initialized
    if (authInitialized && !isLoading && !$auth.isAuthenticated) {
      console.log('Dashboard: redirecting to login - not authenticated');
      goto('/login');
    }
  }

  onMount(async () => {
    console.log('Dashboard: initializing auth...');
    await auth.init();
    authInitialized = true;
    console.log('Dashboard: auth initialized, isAuthenticated:', $auth.isAuthenticated);
    
    theme.init();
    navigation.init();
  });

  function handleLogout() {
    auth.logout();
    goto('/');
  }

  function handleNavigation(path: string) {
    navigation.navigate(path);
    goto(path);
  }

  // Watch for page changes to update navigation
  $: if ($page.url.pathname) {
    navigation.navigate($page.url.pathname);
  }
</script>

{#if isLoading}
  <div class="min-h-screen bg-background flex items-center justify-center">
    <div class="text-center">
      <div class="animate-spin rounded-full h-12 w-12 border-b-2 border-primary mx-auto"></div>
      <p class="mt-4 text-muted-foreground">Laden...</p>
    </div>
  </div>
{:else if $auth.isAuthenticated && user}
  <div class="min-h-screen bg-background">
    <!-- Navigation -->
    <nav class="bg-card shadow-sm border-b border-border">
      <div class="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
        <div class="flex justify-between h-16">
          <!-- Logo & Navigation Controls -->
          <div class="flex items-center space-x-4">
            <div class="flex items-center space-x-2">
              <Icon name="cloud" size={24} />
              <h1 class="text-xl font-bold text-foreground">CloudBox</h1>
            </div>
            
            <!-- Navigation controls -->
            <div class="flex items-center space-x-2">
              <button
                on:click={() => navigation.goBack()}
                disabled={!$navigation.canGoBack}
                class="p-2 text-muted-foreground hover:text-foreground disabled:opacity-50 disabled:cursor-not-allowed"
                title="Terug"
              >
                <Icon name="back" size={16} />
              </button>
              <button
                on:click={() => navigation.goForward()}
                disabled={!$navigation.canGoForward}
                class="p-2 text-muted-foreground hover:text-foreground disabled:opacity-50 disabled:cursor-not-allowed"
                title="Vooruit"
              >
                <Icon name="forward" size={16} />
              </button>
            </div>
          </div>

          <!-- Navigation links -->
          <div class="hidden sm:flex sm:items-center sm:space-x-8">
            <a 
              href="/dashboard" 
              class="flex items-center space-x-2 text-foreground hover:text-primary px-3 py-2 text-sm font-medium"
              class:text-primary={$page.url.pathname === '/dashboard'}
            >
              <Icon name="dashboard" size={16} />
              <span>Dashboard</span>
            </a>
            <a 
              href="/dashboard/organizations" 
              class="flex items-center space-x-2 text-muted-foreground hover:text-primary px-3 py-2 text-sm font-medium"
              class:text-primary={$page.url.pathname.startsWith('/dashboard/organizations')}
            >
              <Icon name="package" size={16} />
              <span>Organizations</span>
            </a>
            <a 
              href="/dashboard/projects" 
              class="flex items-center space-x-2 text-muted-foreground hover:text-primary px-3 py-2 text-sm font-medium"
              class:text-primary={$page.url.pathname.startsWith('/dashboard/projects')}
            >
              <Icon name="projects" size={16} />
              <span>Projecten</span>
            </a>
            <a 
              href="/dashboard/settings" 
              class="flex items-center space-x-2 text-muted-foreground hover:text-primary px-3 py-2 text-sm font-medium"
              class:text-primary={$page.url.pathname === '/dashboard/settings'}
            >
              <Icon name="settings" size={16} />
              <span>Instellingen</span>
            </a>
          </div>

          <!-- User menu -->
          <div class="flex items-center space-x-4">
            <!-- Theme toggle -->
            <button
              on:click={() => theme.toggleTheme()}
              class="p-2 text-muted-foreground hover:text-foreground"
              title="Donkere/lichte modus"
            >
              {#if $theme.theme === 'dark'}
                <Icon name="sun" size={16} />
              {:else}
                <Icon name="moon" size={16} />
              {/if}
            </button>
            
            <div class="flex items-center space-x-2">
              <Icon name="user" size={16} />
              <span class="text-sm text-foreground">Welkom, {user.name}</span>
            </div>
            
            <button
              on:click={() => goto('/admin')}
              class="flex items-center space-x-2 text-muted-foreground hover:text-foreground"
              title="Admin Dashboard"
            >
              <Icon name="shield-check" size={16} />
              <span class="text-sm">Admin</span>
            </button>

            <button
              on:click={handleLogout}
              class="flex items-center space-x-2 text-muted-foreground hover:text-foreground"
              title="Uitloggen"
            >
              <Icon name="logout" size={16} />
              <span class="text-sm">Uitloggen</span>
            </button>
          </div>
        </div>
      </div>
    </nav>

    <!-- Page content -->
    <main class="max-w-7xl mx-auto py-6 px-4 sm:px-6 lg:px-8">
      <slot />
    </main>
  </div>
{/if}