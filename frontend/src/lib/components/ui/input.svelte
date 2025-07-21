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
  }
  
  export let type: $$Props["type"] = "text"
  export let value: $$Props["value"] = ""
  export let placeholder: $$Props["placeholder"] = undefined
  export let disabled: $$Props["disabled"] = false
  export let readonly: $$Props["readonly"] = false
  export let required: $$Props["required"] = false
  
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
    "flex h-10 w-full rounded-md border border-input bg-background px-3 py-2 text-sm ring-offset-background file:border-0 file:bg-transparent file:text-sm file:font-medium placeholder:text-muted-foreground focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-2 disabled:cursor-not-allowed disabled:opacity-50",
    className
  )}
  on:input
  on:focus
  on:blur
  on:keydown
  {...$$restProps}
/>