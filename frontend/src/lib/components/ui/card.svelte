<script lang="ts">
  import { cn } from "$lib/utils"
  
  interface $$Props {
    class?: string
    compact?: boolean
    side?: boolean
    bordered?: boolean
  }
  
  let className: $$Props["class"] = undefined
  export { className as class }
  export let compact: $$Props["compact"] = false
  export let side: $$Props["side"] = false
  export let bordered: $$Props["bordered"] = true
</script>

<div
  class={cn(
    "card bg-card text-card-foreground shadow-sm border border-border",
    compact && "card-compact",
    side && "card-side",
    bordered && "border",
    className
  )}
  {...$$restProps}
>
  <slot />
</div>