<script lang="ts">
  import { cn } from "$lib/utils"
  
  interface $$Props {
    variant?: "primary" | "secondary" | "ghost" | "destructive" | "success" | "warning" | "link"
    size?: "sm" | "md" | "lg" | "xl" | "icon"
    class?: string
    href?: string
    type?: "button" | "submit" | "reset"
    disabled?: boolean
    loading?: boolean
  }
  
  export let variant: $$Props["variant"] = "primary"
  export let size: $$Props["size"] = "md" 
  export let href: $$Props["href"] = undefined
  export let type: $$Props["type"] = "button"
  export let disabled: $$Props["disabled"] = false
  export let loading: $$Props["loading"] = false
  
  let className: $$Props["class"] = undefined
  export { className as class }
  
  // Modern button variants with NO BORDERS and beautiful effects
  const variants = {
    // Primary - Beautiful gradient with theme colors and glow effect
    primary: "btn-primary bg-gradient-to-r from-[hsl(var(--primary))] to-[hsl(var(--primary)/0.8)] text-[hsl(var(--primary-foreground))] shadow-lg shadow-[hsl(var(--primary)/0.3)] hover:shadow-xl hover:shadow-[hsl(var(--primary)/0.4)] hover:scale-[1.02] active:scale-[0.98] transform transition-all duration-200",
    
    // Secondary - Glassmorphism effect
    secondary: "btn-secondary bg-white/10 backdrop-blur-md text-foreground shadow-lg shadow-black/5 hover:bg-white/20 hover:shadow-xl hover:shadow-black/10 hover:scale-[1.02] active:scale-[0.98] border-0 transform transition-all duration-200",
    
    // Ghost - Ultra minimal with subtle hover
    ghost: "btn-ghost text-foreground hover:bg-accent/10 hover:text-accent-foreground transition-all duration-200 hover:scale-[1.02] active:scale-[0.98] transform",
    
    // Destructive - Red gradient with glow
    destructive: "btn-destructive bg-gradient-to-r from-red-500 to-red-600 text-white shadow-lg shadow-red-500/30 hover:shadow-xl hover:shadow-red-500/40 hover:scale-[1.02] active:scale-[0.98] transform transition-all duration-200",
    
    // Success - Green gradient with glow
    success: "btn-success bg-gradient-to-r from-[hsl(var(--success))] to-[hsl(var(--success)/0.8)] text-white shadow-lg shadow-[hsl(var(--success)/0.3)] hover:shadow-xl hover:shadow-[hsl(var(--success)/0.4)] hover:scale-[1.02] active:scale-[0.98] transform transition-all duration-200",
    
    // Warning - Orange gradient with glow
    warning: "btn-warning bg-gradient-to-r from-[hsl(var(--warning))] to-[hsl(var(--warning)/0.8)] text-white shadow-lg shadow-[hsl(var(--warning)/0.3)] hover:shadow-xl hover:shadow-[hsl(var(--warning)/0.4)] hover:scale-[1.02] active:scale-[0.98] transform transition-all duration-200",
    
    // Link - Minimal text button
    link: "btn-link text-[hsl(var(--primary))] hover:text-[hsl(var(--primary)/0.8)] underline-offset-4 hover:underline transition-all duration-200 hover:scale-[1.02] transform"
  }
  
  const sizes = {
    sm: "h-8 px-3 text-sm rounded-lg font-medium",
    md: "h-10 px-4 text-sm rounded-xl font-medium",
    lg: "h-12 px-6 text-base rounded-xl font-semibold",
    xl: "h-14 px-8 text-lg rounded-2xl font-semibold",
    icon: "h-10 w-10 rounded-xl"
  }
</script>

{#if href}
  <a
    {href}
    class={cn(
      "btn inline-flex items-center justify-center whitespace-nowrap relative overflow-hidden border-0 focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-primary/50 focus-visible:ring-offset-0 disabled:pointer-events-none disabled:opacity-50 disabled:transform-none",
      variants[variant],
      sizes[size],
      (disabled || loading) && "cursor-not-allowed opacity-50 transform-none hover:transform-none",
      className
    )}
    on:click
    {...$$restProps}
  >
    {#if loading}
      <div class="animate-spin rounded-full h-4 w-4 border-2 border-current border-t-transparent mr-2"></div>
    {/if}
    <slot />
    <!-- Ripple effect element -->
    <div class="absolute inset-0 overflow-hidden rounded-[inherit] pointer-events-none">
      <div class="ripple absolute bg-white/20 rounded-full transform scale-0 transition-transform duration-300"></div>
    </div>
  </a>
{:else}
  <button
    {type}
    {disabled}
    class={cn(
      "btn inline-flex items-center justify-center whitespace-nowrap relative overflow-hidden border-0 focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-primary/50 focus-visible:ring-offset-0 disabled:pointer-events-none disabled:opacity-50 disabled:transform-none",
      variants[variant],
      sizes[size],
      (disabled || loading) && "cursor-not-allowed opacity-50 transform-none hover:transform-none",
      className
    )}
    on:click
    {...$$restProps}
  >
    {#if loading}
      <div class="animate-spin rounded-full h-4 w-4 border-2 border-current border-t-transparent mr-2"></div>
    {/if}
    <slot />
    <!-- Ripple effect element -->
    <div class="absolute inset-0 overflow-hidden rounded-[inherit] pointer-events-none">
      <div class="ripple absolute bg-white/20 rounded-full transform scale-0 transition-transform duration-300"></div>
    </div>
  </button>
{/if}

<style>
  /* Modern button styles with NO BORDERS */
  :global(.btn) {
    border: none !important;
    position: relative;
    font-weight: 500;
    letter-spacing: 0.025em;
  }

  /* Primary button - Gradient with glow */
  :global(.btn-primary) {
    background: linear-gradient(135deg, hsl(var(--primary-400)), hsl(var(--primary-600)));
    box-shadow: 0 4px 15px rgba(var(--primary) / 0.3), 0 2px 4px rgba(0, 0, 0, 0.1);
  }

  :global(.btn-primary:hover) {
    background: linear-gradient(135deg, hsl(var(--primary-300)), hsl(var(--primary-500)));
    box-shadow: 0 8px 25px rgba(var(--primary) / 0.4), 0 4px 8px rgba(0, 0, 0, 0.15);
  }

  /* Secondary button - Glassmorphism */
  :global(.btn-secondary) {
    background: rgba(255, 255, 255, 0.1);
    backdrop-filter: blur(10px);
    -webkit-backdrop-filter: blur(10px);
    box-shadow: 0 4px 15px rgba(0, 0, 0, 0.1), inset 0 1px 0 rgba(255, 255, 255, 0.1);
  }

  :global(.btn-secondary:hover) {
    background: rgba(255, 255, 255, 0.2);
    box-shadow: 0 8px 25px rgba(0, 0, 0, 0.15), inset 0 1px 0 rgba(255, 255, 255, 0.2);
  }

  /* Dark mode glassmorphism */
  :global(.dark .btn-secondary) {
    background: rgba(255, 255, 255, 0.05);
    box-shadow: 0 4px 15px rgba(0, 0, 0, 0.3), inset 0 1px 0 rgba(255, 255, 255, 0.05);
  }

  :global(.dark .btn-secondary:hover) {
    background: rgba(255, 255, 255, 0.1);
    box-shadow: 0 8px 25px rgba(0, 0, 0, 0.4), inset 0 1px 0 rgba(255, 255, 255, 0.1);
  }

  /* Ghost button - Ultra minimal */
  :global(.btn-ghost) {
    background: transparent;
    box-shadow: none;
  }

  :global(.btn-ghost:hover) {
    background: rgba(var(--accent) / 0.1);
  }

  /* Status buttons with gradients */
  :global(.btn-destructive) {
    background: linear-gradient(135deg, #ef4444, #dc2626);
    box-shadow: 0 4px 15px rgba(239, 68, 68, 0.3), 0 2px 4px rgba(0, 0, 0, 0.1);
  }

  :global(.btn-success) {
    background: linear-gradient(135deg, #10b981, #059669);
    box-shadow: 0 4px 15px rgba(16, 185, 129, 0.3), 0 2px 4px rgba(0, 0, 0, 0.1);
  }

  :global(.btn-warning) {
    background: linear-gradient(135deg, #f59e0b, #d97706);
    box-shadow: 0 4px 15px rgba(245, 158, 11, 0.3), 0 2px 4px rgba(0, 0, 0, 0.1);
  }

  /* Link button */
  :global(.btn-link) {
    background: transparent;
    box-shadow: none;
    text-decoration: underline;
    text-underline-offset: 4px;
  }

  /* Disabled state */
  :global(.btn:disabled) {
    opacity: 0.5;
    transform: none !important;
    box-shadow: none !important;
    cursor: not-allowed;
  }

  /* Focus states with NO BORDER rings */
  :global(.btn:focus-visible) {
    outline: none;
    ring: 2px solid rgba(var(--primary) / 0.5);
    ring-offset: 0;
  }

  /* Ripple effect */
  @keyframes ripple {
    to {
      transform: scale(4);
      opacity: 0;
    }
  }

  :global(.btn:active .ripple) {
    animation: ripple 0.3s linear;
  }

  /* Ensure no default button borders anywhere */
  :global(.btn),
  :global(.btn:hover),
  :global(.btn:focus),
  :global(.btn:active),
  :global(.btn:disabled) {
    border: none !important;
    outline: none;
  }

  /* Loading spinner */
  :global(.btn .loading-spinner) {
    margin-right: 0.5rem;
  }
</style>