<script lang="ts">
  import { cn } from "$lib/utils"
  
  interface $$Props {
    class?: string
    for?: string
  }
  
  export let for_: $$Props["for"] = undefined
  
  let className: $$Props["class"] = undefined
  export { className as class }
</script>

<label
  for={for_}
  class={cn(
    "text-sm font-medium leading-none peer-disabled:cursor-not-allowed peer-disabled:opacity-70",
    className
  )}
  {...$$restProps}
>
  <slot />
</label>