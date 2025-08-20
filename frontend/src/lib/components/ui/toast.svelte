<script lang="ts">
  import { fly } from 'svelte/transition';
  import { toastStore, type Toast } from '$lib/stores/toast';
  import { X, CheckCircle, XCircle, AlertTriangle, Info } from 'lucide-svelte';

  export let toast: Toast;

  const icons = {
    success: CheckCircle,
    error: XCircle,
    warning: AlertTriangle,
    info: Info,
  };

  const styles = {
    success: 'status-success border',
    error: 'status-error border',
    warning: 'status-warning border',
    info: 'status-info border',
  };

  const iconStyles = {
    success: 'icon-success',
    error: 'icon-error',
    warning: 'icon-warning',
    info: 'icon-info',
  };

  function handleRemove() {
    toastStore.remove(toast.id);
  }
</script>

<div
  class="w-full max-w-sm {styles[toast.type]} border rounded-lg shadow-lg backdrop-blur-sm"
  transition:fly={{ x: 300, duration: 300 }}
>
  <div class="p-4">
    <div class="flex items-start">
      <div class="flex-shrink-0">
        <svelte:component 
          this={icons[toast.type]} 
          class="h-5 w-5 {iconStyles[toast.type]}" 
        />
      </div>
      <div class="ml-3 w-0 flex-1">
        {#if toast.title}
          <p class="text-sm font-medium">
            {toast.title}
          </p>
        {/if}
        <p class="text-sm {toast.title ? 'mt-1' : ''}">
          {toast.message}
        </p>
      </div>
      {#if toast.dismissible}
        <div class="ml-4 flex-shrink-0 flex">
          <button
            class="rounded-md inline-flex text-muted-foreground hover:text-foreground focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-primary"
            on:click={handleRemove}
          >
            <span class="sr-only">Close</span>
            <X class="h-5 w-5" />
          </button>
        </div>
      {/if}
    </div>
  </div>
</div>