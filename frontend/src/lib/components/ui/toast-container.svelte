<script lang="ts">
  import { toastStore } from '$lib/stores/toast';
  import Toast from './toast.svelte';
</script>

<!-- Toast Container - Fixed position with better spacing -->
<div class="fixed top-6 right-6 z-[9999] space-y-2 max-w-[400px] w-full pointer-events-none">
  <div class="pointer-events-auto">
    {#each $toastStore as toast (toast.id)}
      <div class="mb-2">
        <Toast {toast} />
      </div>
    {/each}
  </div>
</div>