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
    sm: 'glassmorphism-modal w-11/12 max-w-md p-6 rounded-2xl',
    md: 'glassmorphism-modal w-11/12 max-w-lg p-6 rounded-2xl',
    lg: 'glassmorphism-modal w-11/12 max-w-2xl p-6 rounded-2xl',
    xl: 'glassmorphism-modal w-11/12 max-w-4xl p-6 rounded-2xl',
    '2xl': 'glassmorphism-modal w-11/12 max-w-6xl p-6 rounded-2xl',
    '3xl': 'glassmorphism-modal w-11/12 max-w-7xl p-6 rounded-2xl'
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

  // Better body scroll management
  function manageBodyScroll(isOpen: boolean) {
    if (typeof document !== 'undefined') {
      if (isOpen) {
        document.body.style.overflow = 'hidden';
        document.body.style.paddingRight = '0px'; // Prevent layout shift
      } else {
        document.body.style.overflow = '';
        document.body.style.paddingRight = '';
      }
    }
  }

  onMount(() => {
    manageBodyScroll(open);
    
    return () => {
      manageBodyScroll(false);
    };
  });

  $: manageBodyScroll(open);
</script>

<svelte:window on:keydown={handleKeydown} />

{#if open}
  <!-- Enhanced Modal with improved scrolling and positioning -->
  <div 
    class="fixed inset-0 z-50 flex items-center justify-center p-4 overflow-hidden" 
    transition:fade={{ duration: 200 }}
  >
    <!-- Full screen backdrop with proper scroll blocking -->
    <div 
      class="absolute inset-0 bg-black/50 backdrop-blur-md dark:bg-black/70 dark:backdrop-blur-lg" 
      on:click={closeOnClickOutside ? close : undefined}
      role="presentation"
    ></div>
    
    <!-- Modal container with proper overflow handling -->
    <div 
      class="relative z-10 w-full max-w-[90vw] max-h-[90vh] flex flex-col"
      style="max-width: {size === 'sm' ? '24rem' : size === 'md' ? '32rem' : size === 'lg' ? '42rem' : size === 'xl' ? '56rem' : size === '2xl' ? '72rem' : '80rem'};"
    >
      <!-- Modal content with scrollable body -->
      <div
        bind:this={modalElement}
        class="{sizeClasses[size]} w-full max-h-full flex flex-col"
        role="dialog"
        aria-modal="true"
        aria-labelledby={title ? 'modal-title' : undefined}
        transition:scale={{ duration: 200 }}
        on:click|stopPropagation
      >
        <!-- Header - Fixed at top -->
        {#if title || $$slots.header}
          <div class="flex items-center justify-between mb-4 flex-shrink-0">
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

        <!-- Content with proper scrolling -->
        <div class="modal-content flex-1 min-h-0 overflow-y-auto">
          <slot />
        </div>

        <!-- Footer - Fixed at bottom -->
        {#if $$slots.footer}
          <div class="modal-action flex-shrink-0">
            <slot name="footer" />
          </div>
        {/if}
      </div>
    </div>
  </div>
{/if}

<style>
  /* Improved modal scrolling and positioning */
  :global(.modal-content) {
    /* Remove max-height restriction - let container handle it */
    overflow-y: auto;
    /* Add custom scrollbar styling */
    scrollbar-width: thin;
    scrollbar-color: hsl(var(--border)) transparent;
  }
  
  :global(.modal-content)::-webkit-scrollbar {
    width: 6px;
  }
  
  :global(.modal-content)::-webkit-scrollbar-track {
    background: transparent;
  }
  
  :global(.modal-content)::-webkit-scrollbar-thumb {
    background-color: hsl(var(--border));
    border-radius: 3px;
  }
  
  :global(.modal-content)::-webkit-scrollbar-thumb:hover {
    background-color: hsl(var(--primary) / 0.6);
  }
  
  /* Ensure modal doesn't overflow on mobile */
  @media (max-width: 640px) {
    :global(.glassmorphism-modal) {
      width: calc(100vw - 2rem) !important;
      max-width: calc(100vw - 2rem) !important;
      margin: 0;
    }
  }
</style>