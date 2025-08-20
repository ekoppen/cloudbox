<script lang="ts">
  import { onMount } from 'svelte';
  import { goto } from '$app/navigation';
  import { page } from '$app/stores';
  import { auth, type User } from '$lib/stores/auth';
  import { theme } from '$lib/stores/theme';
  import { navigation } from '$lib/stores/navigation';
  import { sidebarStore } from '$lib/stores/sidebar';
  import Sidebar from '$lib/components/navigation/sidebar.svelte';
  import Header from '$lib/components/navigation/header.svelte';
  
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

  // Watch for page changes to update navigation
  $: if ($page.url.pathname) {
    navigation.navigate($page.url.pathname);
  }
</script>

<style>
  .dashboard-background {
    position: relative;
    background: var(--color-base-100);
  }

  .dashboard-gradient {
    position: fixed;
    top: 0;
    left: 0;
    right: 0;
    bottom: 0;
    background: linear-gradient(
      135deg,
      rgba(102, 126, 234, 0.03) 0%,
      rgba(118, 75, 162, 0.02) 25%,
      rgba(240, 147, 251, 0.025) 50%,
      rgba(245, 87, 108, 0.02) 75%,
      rgba(79, 172, 254, 0.03) 100%
    );
    background-size: 400% 400%;
    animation: subtleGradientShift 30s ease infinite;
    z-index: 1;
    pointer-events: none;
  }

  @keyframes subtleGradientShift {
    0% { background-position: 0% 50%; }
    50% { background-position: 100% 50%; }
    100% { background-position: 0% 50%; }
  }

  /* Dark mode support - CloudBox theme system */
  :global(.cloudbox-dark) .dashboard-gradient {
    background: linear-gradient(
      135deg,
      rgba(102, 126, 234, 0.05) 0%,
      rgba(118, 75, 162, 0.04) 25%,
      rgba(240, 147, 251, 0.045) 50%,
      rgba(245, 87, 108, 0.04) 75%,
      rgba(79, 172, 254, 0.05) 100%
    );
  }

  /* Reduce motion for accessibility */
  @media (prefers-reduced-motion: reduce) {
    .dashboard-gradient {
      animation: none;
    }
  }
</style>

{#if isLoading}
  <div class="min-h-screen bg-base-100 flex items-center justify-center">
    <div class="text-center">
      <div class="loading loading-spinner loading-lg text-primary"></div>
      <p class="mt-4 text-base-content text-sm">Laden...</p>
    </div>
  </div>
{:else if $auth.isAuthenticated && user}
  <div class="min-h-screen dashboard-background">
    <!-- Subtle gradient background -->
    <div class="dashboard-gradient"></div>
    
    <!-- Supabase-style Collapsible Sidebar -->
    <Sidebar context="dashboard" />

    <!-- Main content with dynamic sidebar offset -->
    <div class="transition-all duration-200 ease-in-out min-h-screen ml-sidebar-collapsed">
      <!-- Header -->
      <Header />

      <!-- Page content -->
      <main class="p-6 relative z-10">
        <slot />
      </main>
    </div>
  </div>
{/if}