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
    success: 'bg-green-50 border-green-200 text-green-800',
    error: 'bg-red-50 border-red-200 text-red-800',
    warning: 'bg-yellow-50 border-yellow-200 text-yellow-800',
    info: 'bg-blue-50 border-blue-200 text-blue-800',
  };

  const iconStyles = {
    success: 'text-green-400',
    error: 'text-red-400',
    warning: 'text-yellow-400',
    info: 'text-blue-400',
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
            class="rounded-md inline-flex text-gray-400 hover:text-gray-500 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-indigo-500"
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