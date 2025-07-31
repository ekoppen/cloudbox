<script lang="ts">
  import Card from './card.svelte';
  import Button from './button.svelte';
  import Icon from './icon.svelte';

  export let options = [];
  export let selectedOption = null;
  export let onSelect = null;
  export let showSelectButton = false;
  export let selectButtonText = 'Selecteren';

  function selectOption(option) {
    selectedOption = option;
    if (onSelect) {
      onSelect(option);
    }
  }

  function getOptionIcon(name: string) {
    const icons = {
      'docker': '🐳',
      'npm': '📦',
      'yarn': '🧶',
      'pnpm': '⚡',
      'custom': '⚙️'
    };
    return icons[name] || '📦';
  }
</script>

{#if options && options.length > 0}
  <div class="grid gap-4">
    {#each options as option}
      <div 
        class="border rounded-lg p-4 transition-all duration-200 hover:shadow-md cursor-pointer
        {option.is_recommended ? 'border-green-300 bg-green-50 hover:bg-green-100' : 'border-border hover:border-primary/50'}
        {selectedOption?.name === option.name ? 'ring-2 ring-primary ring-opacity-50 bg-primary/5' : ''}"
        on:click={() => selectOption(option)}
        on:keydown={(e) => e.key === 'Enter' && selectOption(option)}
        role="button"
        tabindex="0"
      >
        <div class="flex justify-between items-start mb-2">
          <div class="flex items-center gap-3">
            <span class="text-2xl">{getOptionIcon(option.name)}</span>
            <div>
              <h4 class="font-semibold capitalize flex items-center gap-2">
                {option.name}
                {#if option.is_recommended}
                  <span class="px-2 py-1 text-xs font-medium rounded-full bg-green-100 text-green-800 border border-green-200">
                    Aanbevolen
                  </span>
                {/if}
                {#if selectedOption?.name === option.name}
                  <span class="px-2 py-1 text-xs font-medium rounded-full bg-primary text-primary-foreground">
                    Geselecteerd
                  </span>
                {/if}
              </h4>
            </div>
          </div>
          {#if showSelectButton}
            <Button
              size="sm"
              variant={selectedOption?.name === option.name ? "default" : "outline"}
              on:click={(e) => {
                e.stopPropagation();
                selectOption(option);
              }}
            >
              {#if selectedOption?.name === option.name}
                <Icon name="check" size={16} className="mr-1" />
                Geselecteerd
              {:else}
                {selectButtonText}
              {/if}
            </Button>
          {/if}
        </div>
        
        <p class="text-sm text-muted-foreground mb-3">{option.description}</p>
        
        <div class="grid grid-cols-1 md:grid-cols-2 gap-3 text-sm">
          <div>
            <span class="font-medium text-muted-foreground">Install:</span>
            <code class="block bg-muted px-2 py-1 rounded text-xs font-mono mt-1 border">{option.command}</code>
          </div>
          {#if option.build_command}
            <div>
              <span class="font-medium text-muted-foreground">Build:</span>
              <code class="block bg-muted px-2 py-1 rounded text-xs font-mono mt-1 border">{option.build_command}</code>
            </div>
          {/if}
          {#if option.start_command}
            <div>
              <span class="font-medium text-muted-foreground">Start:</span>
              <code class="block bg-muted px-2 py-1 rounded text-xs font-mono mt-1 border">{option.start_command}</code>
            </div>
          {/if}
          {#if option.dev_command}
            <div>
              <span class="font-medium text-muted-foreground">Dev:</span>
              <code class="block bg-muted px-2 py-1 rounded text-xs font-mono mt-1 border">{option.dev_command}</code>
            </div>
          {/if}
        </div>

        {#if option.port && option.port !== 3000}
          <div class="mt-3 flex items-center gap-2">
            <span class="text-xs font-medium text-muted-foreground">Port:</span>
            <code class="text-xs bg-muted px-1 rounded font-mono border">{option.port}</code>
          </div>
        {/if}
      </div>
    {/each}
  </div>
{:else}
  <div class="text-center py-8 text-muted-foreground">
    <div class="text-4xl mb-2">📦</div>
    <p>Geen installatie opties beschikbaar</p>
  </div>
{/if}