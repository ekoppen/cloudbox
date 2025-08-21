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
    
    // Reset sidebar context when not in a project route
    if (!$page.url.pathname.startsWith('/dashboard/projects/') || $page.url.pathname === '/dashboard/projects') {
      sidebarStore.setContext('dashboard');
    }
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
    background: 
      radial-gradient(circle at 20% 50%, rgba(120, 119, 198, 0.3) 0%, transparent 50%),
      radial-gradient(circle at 80% 20%, rgba(255, 119, 198, 0.3) 0%, transparent 50%),
      radial-gradient(circle at 40% 80%, rgba(102, 126, 234, 0.4) 0%, transparent 50%),
      linear-gradient(
        135deg,
        rgba(102, 126, 234, 0.08) 0%,
        rgba(118, 75, 162, 0.06) 25%,
        rgba(240, 147, 251, 0.08) 50%,
        rgba(245, 87, 108, 0.06) 75%,
        rgba(79, 172, 254, 0.1) 100%
      );
    background-size: 400% 400%, 300% 300%, 500% 500%, 400% 400%;
    animation: vibrantGradientShift 20s ease infinite;
    z-index: 1;
    pointer-events: none;
    filter: blur(40px);
    opacity: 0.7;
  }

  @keyframes vibrantGradientShift {
    0% { 
      background-position: 0% 50%, 100% 0%, 50% 100%, 0% 50%; 
      transform: scale(1);
    }
    33% { 
      background-position: 50% 100%, 0% 50%, 100% 0%, 50% 100%; 
      transform: scale(1.05);
    }
    66% { 
      background-position: 100% 0%, 50% 100%, 0% 50%, 100% 0%; 
      transform: scale(1.02);
    }
    100% { 
      background-position: 0% 50%, 100% 0%, 50% 100%, 0% 50%; 
      transform: scale(1);
    }
  }

  /* Dark mode support - CloudBox theme system - Neutral grey gradient */
  :global(.cloudbox-dark) .dashboard-gradient {
    background: 
      radial-gradient(circle at 20% 50%, rgba(64, 64, 64, 0.15) 0%, transparent 50%),
      radial-gradient(circle at 80% 20%, rgba(96, 96, 96, 0.15) 0%, transparent 50%),
      radial-gradient(circle at 40% 80%, rgba(80, 80, 80, 0.2) 0%, transparent 50%),
      linear-gradient(
        135deg,
        rgba(48, 48, 48, 0.08) 0%,
        rgba(64, 64, 64, 0.06) 25%,
        rgba(96, 96, 96, 0.1) 50%,
        rgba(80, 80, 80, 0.06) 75%,
        rgba(112, 112, 112, 0.12) 100%
      );
    filter: blur(60px);
    opacity: 0.7;
  }

  /* Reduce motion for accessibility */
  @media (prefers-reduced-motion: reduce) {
    .dashboard-gradient {
      animation: none;
      filter: blur(20px);
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