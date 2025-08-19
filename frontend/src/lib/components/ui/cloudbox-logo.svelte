<script lang="ts">
  import { cn } from "$lib/utils";
  
  export let size: 'xs' | 'sm' | 'md' | 'lg' | 'xl' = 'md';
  export let className: string = '';
  export let showText: boolean = true;
  
  const sizeClasses = {
    xs: 'h-4 w-4',
    sm: 'h-6 w-6', 
    md: 'h-8 w-8',
    lg: 'h-10 w-10',
    xl: 'h-12 w-12'
  };
  
  const textSizeClasses = {
    xs: 'text-xs',
    sm: 'text-sm',
    md: 'text-base',
    lg: 'text-lg', 
    xl: 'text-xl'
  };
</script>

<div class={cn("flex items-center space-x-2", className)}>
  <!-- Logo Icon - CloudBox themed gradient -->
  <div class={cn(
    "flex items-center justify-center rounded-lg bg-gradient-to-br from-primary to-secondary text-primary-content shadow-sm",
    sizeClasses[size]
  )}>
    <!-- CloudBox Logo SVG -->
    <svg 
      class="w-2/3 h-2/3" 
      viewBox="0 0 24 24" 
      fill="none" 
      xmlns="http://www.w3.org/2000/svg"
    >
      <!-- Cloud shape with box -->
      <path 
        d="M18 10h-1.26A8 8 0 1 0 9 20h9a5 5 0 0 0 0-10z" 
        fill="currentColor" 
        opacity="0.8"
      />
      <!-- Box/container in the cloud -->
      <rect 
        x="9" 
        y="12" 
        width="6" 
        height="4" 
        rx="1" 
        fill="currentColor" 
        opacity="0.9"
      />
      <rect 
        x="10" 
        y="13" 
        width="4" 
        height="2" 
        rx="0.5" 
        fill="white" 
        opacity="0.4"
      />
    </svg>
  </div>
  
  {#if showText}
    <span class={cn("font-semibold text-base-content cloudbox-theme", textSizeClasses[size])}>
      CloudBox
    </span>
  {/if}
</div>