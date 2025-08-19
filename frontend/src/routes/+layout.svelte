<script lang="ts">
  import '../app.css';
  import { onMount } from 'svelte';
  import { theme } from '$lib/stores/theme';
  import { navigation } from '$lib/stores/navigation';
  import { page } from '$app/stores';
  import ToastContainer from '$lib/components/ui/toast-container.svelte';
  import { browser } from '$app/environment';

  // Apply default theme immediately to prevent flash (client-side only)
  if (browser) {
    const savedTheme = localStorage.getItem('cloudbox_theme') || 'cloudbox';
    const savedAccent = localStorage.getItem('cloudbox_accent_color') || 'blue';
    
    // Set data-theme immediately
    document.documentElement.setAttribute('data-theme', savedTheme);
    
    // Add theme classes
    document.documentElement.classList.add(savedTheme);
    document.documentElement.classList.add(savedTheme === 'cloudbox-dark' ? 'dark' : 'light');
    document.documentElement.classList.add(`accent-${savedAccent}`);
    
    document.body.classList.add(savedTheme);
    document.body.classList.add(savedTheme === 'cloudbox-dark' ? 'dark' : 'light');
    document.body.classList.add(`accent-${savedAccent}`);
  }

  onMount(async () => {
    // Initialize theme first to avoid flash
    await theme.init();
    navigation.init();
  });

  // Subscribe to theme changes to ensure they're applied immediately
  $: if ($theme) {
    // Force theme application on any theme changes
    theme.applyTheme($theme.theme, $theme.accentColor);
  }

  // Track page changes for navigation
  $: if ($page.url.pathname) {
    navigation.navigate($page.url.pathname);
  }
</script>

<main class="min-h-screen">
  <slot />
</main>

<ToastContainer />