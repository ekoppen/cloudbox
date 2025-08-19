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

{#if isLoading}
  <div class="min-h-screen bg-base-100 flex items-center justify-center">
    <div class="text-center">
      <div class="loading loading-spinner loading-lg text-primary"></div>
      <p class="mt-4 text-base-content text-sm">Laden...</p>
    </div>
  </div>
{:else if $auth.isAuthenticated && user}
  <div class="min-h-screen bg-base-100">
    <!-- Supabase-style Collapsible Sidebar -->
    <Sidebar context="dashboard" />

    <!-- Main content with dynamic sidebar offset -->
    <div class="transition-all duration-200 ease-in-out min-h-screen ml-sidebar-collapsed">
      <!-- Header -->
      <Header />

      <!-- Page content -->
      <main class="p-6">
        <slot />
      </main>
    </div>
  </div>
{/if}