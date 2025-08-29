<script lang="ts">
  import { goto } from '$app/navigation';
  import { page } from '$app/stores';
  import { onMount } from 'svelte';

  $: projectId = $page.params.id;

  onMount(() => {
    // Redirect to endpoints by default
    goto(`/dashboard/projects/${projectId}/api/endpoints`, { replaceState: true });
  });
</script>

<!-- Loading state while redirecting -->
<div class="flex items-center justify-center min-h-64">
  <div class="animate-spin rounded-full h-8 w-8 border-b-2 border-primary"></div>
</div>