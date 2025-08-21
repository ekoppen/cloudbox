# CloudBox Modal Implementation Guide

## Quick Reference

### Use the Modal Component (Recommended)

```svelte
<script>
  import Modal from '$lib/components/ui/modal.svelte';
  let showModal = false;
</script>

<Modal bind:open={showModal} title="My Modal" size="md">
  <!-- Your modal content here -->
  <div slot="footer">
    <Button on:click={() => showModal = false}>Close</Button>
  </div>
</Modal>
```

### Custom Modal Implementation

If you need a custom modal, follow this pattern:

```svelte
{#if showModal}
  <div 
    class="fixed inset-0 modal-backdrop-enhanced flex items-start justify-center p-4 pt-16 sm:pt-20 overflow-y-auto z-50"
    on:click={() => showModal = false}
  >
    <div 
      class="glassmorphism-card max-w-md w-full border-2 shadow-2xl my-auto modal-content-wrapper"
      on:click|stopPropagation
    >
      <!-- Modal content here -->
      <div class="modal-scroll-area">
        <!-- Scrollable content -->
      </div>
    </div>
  </div>
{/if}
```

## Key Features Fixed

1. **No Top Bar Overlap**: Modals start below the header with `pt-16 sm:pt-20`
2. **Proper Scrolling**: Container has `overflow-y-auto` and content uses `modal-scroll-area`
3. **Mobile Responsive**: Automatic width constraints and padding adjustments
4. **Height Management**: Uses `my-auto` and `modal-content-wrapper` class
5. **Click Outside to Close**: Backdrop click handlers with `stopPropagation` on content

## Available CSS Classes

- `modal-backdrop-enhanced`: Styled backdrop with blur
- `modal-container-fixed`: Container with proper positioning
- `modal-content-wrapper`: Content wrapper with height constraints
- `modal-scroll-area`: Scrollable area with custom scrollbars
- `modal-header-fixed`: Fixed header (doesn't scroll)
- `modal-footer-fixed`: Fixed footer (doesn't scroll)

## Size Options (Modal Component)

- `sm`: Small modal (max-width: 28rem)
- `md`: Medium modal (max-width: 32rem) - Default
- `lg`: Large modal (max-width: 42rem)
- `xl`: Extra large modal (max-width: 56rem)
- `2xl`: 2X large modal (max-width: 72rem)
- `3xl`: 3X large modal (max-width: 80rem)