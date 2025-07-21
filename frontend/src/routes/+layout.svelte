<script lang="ts">
  import '../app.css';
  import { onMount } from 'svelte';
  import { theme } from '$lib/stores/theme';
  import { navigation } from '$lib/stores/navigation';
  import { page } from '$app/stores';
  import ToastContainer from '$lib/components/ui/toast-container.svelte';

  onMount(async () => {
    // Initialize auth first
    await theme.init();
    navigation.init();
  });

  // Track page changes for navigation
  $: if ($page.url.pathname) {
    navigation.navigate($page.url.pathname);
  }
</script>

<main class="min-h-screen">
  <slot />
</main>

<ToastContainer />