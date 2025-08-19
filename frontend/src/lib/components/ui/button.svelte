<script lang="ts">
  import { cn } from "$lib/utils"
  
  interface $$Props {
    variant?: "default" | "destructive" | "outline" | "secondary" | "ghost" | "link"
    size?: "default" | "sm" | "lg" | "icon"
    class?: string
    href?: string
    type?: "button" | "submit" | "reset"
    disabled?: boolean
  }
  
  export let variant: $$Props["variant"] = "default"
  export let size: $$Props["size"] = "default" 
  export let href: $$Props["href"] = undefined
  export let type: $$Props["type"] = "button"
  export let disabled: $$Props["disabled"] = false
  
  let className: $$Props["class"] = undefined
  export { className as class }
  
  const variants = {
    default: "bg-primary text-primary-foreground hover:bg-primary/90 transition-colors duration-200",
    destructive: "bg-destructive text-destructive-foreground hover:bg-destructive/90 transition-colors duration-200",
    outline: "border border-input bg-background hover:bg-accent hover:text-accent-foreground transition-colors duration-200",
    secondary: "bg-secondary text-secondary-foreground hover:bg-secondary/80 transition-colors duration-200",
    ghost: "hover:bg-accent hover:text-accent-foreground transition-colors duration-200",
    link: "text-primary underline-offset-4 hover:underline transition-colors duration-200"
  }
  
  const sizes = {
    default: "h-10 px-4 py-2 rounded-md",
    sm: "h-9 rounded-md px-3",
    lg: "h-11 rounded-md px-8",
    icon: "h-10 w-10 rounded-md"
  }
</script>

{#if href}
  <a
    {href}
    class={cn(
      "btn font-medium inline-flex items-center justify-center whitespace-nowrap focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-2 disabled:pointer-events-none disabled:opacity-50",
      variants[variant || "default"],
      sizes[size || "default"],
      className
    )}
    on:click
    {...$$restProps}
  >
    <slot />
  </a>
{:else}
  <button
    {type}
    {disabled}
    class={cn(
      "btn font-medium inline-flex items-center justify-center whitespace-nowrap focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-2 disabled:pointer-events-none disabled:opacity-50",
      variants[variant || "default"],
      sizes[size || "default"],
      className
    )}
    on:click
    {...$$restProps}
  >
    <slot />
  </button>
{/if}