<script lang="ts">
  import { cn } from "$lib/utils"
  
  interface $$Props {
    class?: string
    type?: string
    value?: string
    placeholder?: string
    disabled?: boolean
    readonly?: boolean
    required?: boolean
    size?: "xs" | "sm" | "md" | "lg"
    bordered?: boolean
    ghost?: boolean
  }
  
  export let type: $$Props["type"] = "text"
  export let value: $$Props["value"] = ""
  export let placeholder: $$Props["placeholder"] = undefined
  export let disabled: $$Props["disabled"] = false
  export let readonly: $$Props["readonly"] = false
  export let required: $$Props["required"] = false
  export let size: $$Props["size"] = "md"
  export let bordered: $$Props["bordered"] = true
  export let ghost: $$Props["ghost"] = false
  
  let className: $$Props["class"] = undefined
  export { className as class }
</script>

<input
  {type}
  bind:value
  {placeholder}
  {disabled}
  {readonly}
  {required}
  class={cn(
    "input w-full bg-background text-foreground border border-input focus:border-primary focus:ring-2 focus:ring-primary/20 focus:outline-none transition-colors duration-200",
    size === "xs" && "input-xs",
    size === "sm" && "input-sm",
    size === "lg" && "input-lg",
    bordered && "border",
    ghost && "input-ghost border-0",
    className
  )}
  on:input
  on:focus
  on:blur
  on:keydown
  {...$$restProps}
/>