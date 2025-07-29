<script lang="ts">
  export let hasPendingUpdate: boolean = false;
  export let size: 'sm' | 'md' | 'lg' = 'md';
  export let pulse: boolean = true;
  
  const sizeClasses = {
    sm: 'h-2 w-2',
    md: 'h-3 w-3', 
    lg: 'h-4 w-4'
  };
</script>

{#if hasPendingUpdate}
  <div class="relative inline-flex">
    <span class="absolute -top-1 -right-1 flex {sizeClasses[size]}">
      <span class="animate-ping absolute inline-flex h-full w-full rounded-full bg-orange-400 opacity-75" class:hidden={!pulse}></span>
      <span class="relative inline-flex rounded-full {sizeClasses[size]} bg-orange-500"></span>
    </span>
    
    <!-- Tooltip -->
    <div class="invisible absolute -top-8 left-1/2 transform -translate-x-1/2 px-2 py-1 text-xs text-white bg-gray-900 rounded opacity-0 transition-opacity group-hover:visible group-hover:opacity-100 whitespace-nowrap">
      Update beschikbaar
    </div>
  </div>
{/if}

<style>
  .group:hover .invisible {
    visibility: visible;
  }
</style>