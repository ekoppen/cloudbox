<script lang="ts">
  import { createEventDispatcher, onMount } from 'svelte';
  import { fade, scale } from 'svelte/transition';
  import Button from './button.svelte';
  import Icon from './icon.svelte';

  export let open = false;
  export let title = '';
  export let size: 'sm' | 'md' | 'lg' | 'xl' | '2xl' | '3xl' = 'md';
  export let closeOnEscape = true;
  export let closeOnClickOutside = true;

  const dispatch = createEventDispatcher();

  let modalElement: HTMLElement;

  const sizeClasses = {
    sm: 'max-w-md',
    md: 'max-w-lg',
    lg: 'max-w-2xl',
    xl: 'max-w-4xl',
    '2xl': 'max-w-6xl',
    '3xl': 'max-w-7xl'
  };

  function handleKeydown(event: KeyboardEvent) {
    if (event.key === 'Escape' && closeOnEscape && open) {
      close();
    }
  }

  function handleClickOutside(event: MouseEvent) {
    if (closeOnClickOutside && modalElement && !modalElement.contains(event.target as Node)) {
      close();
    }
  }

  function close() {
    open = false;
    dispatch('close');
  }

  onMount(() => {
    if (open) {
      document.body.style.overflow = 'hidden';
    }

    return () => {
      document.body.style.overflow = '';
    };
  });

  $: if (open) {
    document.body.style.overflow = 'hidden';
  } else {
    document.body.style.overflow = '';
  }
</script>

<svelte:window on:keydown={handleKeydown} />

{#if open}
  <!-- Backdrop -->
  <div
    class="fixed inset-0 z-50 bg-black/50 backdrop-blur-sm"
    on:click={handleClickOutside}
    transition:fade={{ duration: 200 }}
    role="presentation"
  >
    <!-- Modal -->
    <div
      class="fixed left-1/2 top-1/2 z-50 w-full {sizeClasses[size]} -translate-x-1/2 -translate-y-1/2 p-6"
      transition:scale={{ duration: 200 }}
    >
      <div
        bind:this={modalElement}
        class="relative bg-card border border-border rounded-lg shadow-lg p-6"
        role="dialog"
        aria-modal="true"
        aria-labelledby={title ? 'modal-title' : undefined}
      >
        <!-- Header -->
        {#if title || $$slots.header}
          <div class="flex items-center justify-between mb-4">
            {#if $$slots.header}
              <slot name="header" />
            {:else}
              <h2 id="modal-title" class="text-lg font-semibold text-foreground">
                {title}
              </h2>
            {/if}
            
            <Button
              variant="ghost"
              size="sm"
              on:click={close}
              class="h-8 w-8 p-0"
              aria-label="Sluiten"
            >
              <Icon name="x" size={16} />
            </Button>
          </div>
        {/if}

        <!-- Content -->
        <div class="modal-content">
          <slot />
        </div>

        <!-- Footer -->
        {#if $$slots.footer}
          <div class="flex items-center justify-end gap-3 mt-6 pt-4 border-t border-border">
            <slot name="footer" />
          </div>
        {/if}
      </div>
    </div>
  </div>
{/if}

<style>
  :global(.modal-content) {
    max-height: 70vh;
    overflow-y: auto;
  }
</style>