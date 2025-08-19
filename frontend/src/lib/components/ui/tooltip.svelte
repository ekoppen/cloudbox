<script lang="ts">
  export let text: string;
  export let position: 'top' | 'right' | 'bottom' | 'left' = 'right';
  export let visible = false;
  export let delay = 100;

  let showTooltip = false;
  let timeout: number;

  function handleMouseEnter() {
    clearTimeout(timeout);
    timeout = window.setTimeout(() => {
      showTooltip = true;
    }, delay);
  }

  function handleMouseLeave() {
    clearTimeout(timeout);
    showTooltip = false;
  }

  $: positionClasses = {
    top: 'bottom-full left-1/2 -translate-x-1/2 mb-2',
    right: 'left-full top-1/2 -translate-y-1/2 ml-2',
    bottom: 'top-full left-1/2 -translate-x-1/2 mt-2',
    left: 'right-full top-1/2 -translate-y-1/2 mr-2'
  };

  $: arrowClasses = {
    top: 'top-full left-1/2 -translate-x-1/2 border-l-4 border-r-4 border-t-4 border-l-transparent border-r-transparent border-t-gray-900',
    right: 'right-full top-1/2 -translate-y-1/2 border-t-4 border-b-4 border-r-4 border-t-transparent border-b-transparent border-r-gray-900',
    bottom: 'bottom-full left-1/2 -translate-x-1/2 border-l-4 border-r-4 border-b-4 border-l-transparent border-r-transparent border-b-gray-900',
    left: 'left-full top-1/2 -translate-y-1/2 border-t-4 border-b-4 border-l-4 border-t-transparent border-b-transparent border-l-gray-900'
  };
</script>

<div 
  class="relative inline-block"
  on:mouseenter={handleMouseEnter}
  on:mouseleave={handleMouseLeave}
  role="tooltip"
>
  <slot />
  
  {#if (showTooltip || visible) && text}
    <div 
      class="absolute z-50 px-2 py-1 text-xs text-white bg-gray-900 rounded-md shadow-lg whitespace-nowrap pointer-events-none transition-opacity duration-200 {positionClasses[position]}"
      class:opacity-100={showTooltip || visible}
      class:opacity-0={!showTooltip && !visible}
    >
      {text}
      <div class="absolute h-0 w-0 {arrowClasses[position]}"></div>
    </div>
  {/if}
</div>