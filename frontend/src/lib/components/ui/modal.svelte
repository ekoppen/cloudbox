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
    sm: 'modal-box w-11/12 max-w-md bg-card text-card-foreground border border-border',
    md: 'modal-box w-11/12 max-w-lg bg-card text-card-foreground border border-border',
    lg: 'modal-box w-11/12 max-w-2xl bg-card text-card-foreground border border-border',
    xl: 'modal-box w-11/12 max-w-4xl bg-card text-card-foreground border border-border',
    '2xl': 'modal-box w-11/12 max-w-6xl bg-card text-card-foreground border border-border',
    '3xl': 'modal-box w-11/12 max-w-7xl bg-card text-card-foreground border border-border'
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
  <!-- Enhanced Modal with proper backdrop coverage -->
  <div class="fixed inset-0 z-50 flex items-center justify-center p-4" transition:fade={{ duration: 200 }}>
    <!-- Full screen backdrop -->
    <div 
      class="absolute inset-0 bg-black/50 backdrop-blur-sm" 
      on:click={closeOnClickOutside ? close : undefined}
      role="presentation"
    ></div>
    
    <!-- Modal content -->
    <div
      bind:this={modalElement}
      class="{sizeClasses[size]} relative z-10"
      role="dialog"
      aria-modal="true"
      aria-labelledby={title ? 'modal-title' : undefined}
      transition:scale={{ duration: 200 }}
      on:click|stopPropagation
    >
      <!-- Header -->
      {#if title || $$slots.header}
        <div class="flex items-center justify-between mb-4">
          {#if $$slots.header}
            <slot name="header" />
          {:else}
            <h2 id="modal-title" class="text-lg font-semibold">
              {title}
            </h2>
          {/if}
          
          <Button
            variant="ghost"
            size="sm"
            on:click={close}
            class="btn-sm btn-square"
            aria-label="Sluiten"
          >
            <Icon name="x" size={16} className="icon-contrast" />
          </Button>
        </div>
      {/if}

      <!-- Content -->
      <div class="modal-content">
        <slot />
      </div>

      <!-- Footer -->
      {#if $$slots.footer}
        <div class="modal-action">
          <slot name="footer" />
        </div>
      {/if}
    </div>
  </div>
{/if}

<style>
  :global(.modal-content) {
    max-height: 70vh;
    overflow-y: auto;
  }
</style>