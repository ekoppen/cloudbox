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
  let showUserDropdown = false;

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

    // Close dropdown when clicking outside
    const handleClickOutside = (event) => {
      if (!event.target.closest('.relative')) {
        showUserDropdown = false;
      }
    };
    document.addEventListener('click', handleClickOutside);
    
    return () => {
      document.removeEventListener('click', handleClickOutside);
    };
  });

  function handleLogout() {
    auth.logout();
    goto('/');
  }

  function handleNavigation(path: string) {
    navigation.navigate(path);
    goto(path);
  }

  function getInitials(name: string): string {
    return name
      .split(' ')
      .map(word => word.charAt(0).toUpperCase())
      .slice(0, 2)
      .join('');
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
          </div>

          <!-- User menu -->
          <div class="flex items-center space-x-4">
            <!-- Theme toggle -->
            <button
              on:click={() => theme.toggleTheme()}
              class="p-2 text-muted-foreground hover:text-foreground rounded-lg hover:bg-muted transition-colors"
              title="Donkere/lichte modus"
            >
              {#if $theme.theme === 'dark'}
                <Icon name="sun" size={16} />
              {:else}
                <Icon name="moon" size={16} />
              {/if}
            </button>
            
            <!-- User Dropdown -->
            <div class="relative">
              <button
                on:click={() => showUserDropdown = !showUserDropdown}
                class="flex items-center space-x-3 bg-primary/10 hover:bg-primary/20 text-primary px-3 py-2 rounded-lg transition-colors"
                title="User menu"
              >
                <div class="w-8 h-8 bg-primary text-primary-foreground rounded-full flex items-center justify-center text-sm font-semibold">
                  {getInitials(user.name || 'User')}
                </div>
                <span class="text-sm font-medium">{user.name}</span>
                <Icon name="chevron-down" size={14} className="transition-transform {showUserDropdown ? 'rotate-180' : ''}" />
              </button>

              {#if showUserDropdown}
                <div class="absolute right-0 mt-2 w-48 bg-card border border-border rounded-lg shadow-lg z-50">
                  <div class="py-1">
                    <button
                      on:click={() => { showUserDropdown = false; goto('/dashboard/settings'); }}
                      class="w-full flex items-center space-x-3 px-4 py-2 text-sm text-foreground hover:bg-muted transition-colors"
                    >
                      <Icon name="user" size={16} />
                      <span>Profiel</span>
                    </button>
                    
                    {#if user.role === 'superadmin'}
                      <button
                        on:click={() => { showUserDropdown = false; goto('/dashboard/admin'); }}
                        class="w-full flex items-center space-x-3 px-4 py-2 text-sm text-foreground hover:bg-muted transition-colors"
                      >
                        <Icon name="shield-check" size={16} />
                        <span>Admin</span>
                      </button>
                    {/if}
                    
                    <div class="border-t border-border my-1"></div>
                    
                    <button
                      on:click={() => { showUserDropdown = false; handleLogout(); }}
                      class="w-full flex items-center space-x-3 px-4 py-2 text-sm text-red-600 hover:bg-red-50 dark:hover:bg-red-900/10 transition-colors"
                    >
                      <Icon name="logout" size={16} />
                      <span>Uitloggen</span>
                    </button>
                  </div>
                </div>
              {/if}
            </div>
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