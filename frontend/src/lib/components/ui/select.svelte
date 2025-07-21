<script lang="ts">
  import { createEventDispatcher } from 'svelte';

  export let value: string | number = '';
  export let placeholder: string = 'Selecteer...';
  export let disabled = false;
  export let required = false;
  export let name = '';
  export let id = '';
  export let className = '';

  const dispatch = createEventDispatcher();

  function handleChange(event: Event) {
    const target = event.target as HTMLSelectElement;
    value = target.value;
    dispatch('change', { value });
  }
</script>

<select
  bind:value
  on:change={handleChange}
  {disabled}
  {required}
  {name}
  {id}
  class="flex h-10 w-full rounded-md border border-input bg-background px-3 py-2 text-sm ring-offset-background file:border-0 file:bg-transparent file:text-sm file:font-medium placeholder:text-muted-foreground focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-2 disabled:cursor-not-allowed disabled:opacity-50 {className}"
>
  {#if placeholder}
    <option value="" disabled selected={!value}>{placeholder}</option>
  {/if}
  <slot />
</select>