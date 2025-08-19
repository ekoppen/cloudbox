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
    success: 'bg-green-50 dark:bg-green-900/20 border-green-200 dark:border-green-800 text-green-800 dark:text-green-200',
    error: 'bg-red-50 dark:bg-red-900/20 border-red-200 dark:border-red-800 text-red-800 dark:text-red-200',
    warning: 'bg-yellow-50 dark:bg-yellow-900/20 border-yellow-200 dark:border-yellow-800 text-yellow-800 dark:text-yellow-200',
    info: 'bg-blue-50 dark:bg-blue-900/20 border-blue-200 dark:border-blue-800 text-blue-800 dark:text-blue-200',
  };

  const iconStyles = {
    success: 'text-green-400 dark:text-green-300',
    error: 'text-red-400 dark:text-red-300',
    warning: 'text-yellow-400 dark:text-yellow-300',
    info: 'text-blue-400 dark:text-blue-300',
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
            class="rounded-md inline-flex text-gray-400 dark:text-gray-500 hover:text-gray-500 dark:hover:text-gray-400 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-indigo-500 dark:focus:ring-indigo-400"
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