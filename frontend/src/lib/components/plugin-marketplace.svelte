<script lang="ts">
  import { onMount } from 'svelte';
  import { fly, scale } from 'svelte/transition';
  import Button from '$lib/components/ui/button.svelte';
  import Card from '$lib/components/ui/card.svelte';
  import Icon from '$lib/components/ui/icon.svelte';
  import Badge from '$lib/components/ui/badge.svelte';
  import Input from '$lib/components/ui/input.svelte';
  import { 
    pluginManager, 
    marketplacePlugins, 
    marketplaceLoading, 
    searchQuery, 
    selectedTags,
    installationProgress,
    type MarketplacePlugin 
  } from '$lib/stores/plugins';
  import { addToast } from '$lib/stores/toast';

  export let isOpen = false;
  export let onClose = () => {};

  let searchInput = '';
  let availableTags: string[] = [];
  let selectedPlugin: MarketplacePlugin | null = null;
  let showPluginDetails = false;

  // Reactive search
  $: if (searchInput !== $searchQuery) {
    searchQuery.set(searchInput);
    performSearch();
  }

  // Reactive tag filtering
  $: if ($selectedTags.length > 0 || $searchQuery) {
    performSearch();
  }

  async function loadMarketplace() {
    try {
      const plugins = await pluginManager.loadMarketplace();
      // Extract unique tags
      const tags = new Set<string>();
      plugins.forEach(plugin => {
        plugin.tags.forEach(tag => tags.add(tag));
      });
      availableTags = Array.from(tags).sort();
    } catch (error) {
      console.error('Failed to load marketplace:', error);
      addToast('Failed to load plugin marketplace: ' + (error instanceof Error ? error.message : 'Unknown error'), 'error');
    }
  }

  async function performSearch() {
    if ($searchQuery.trim() || $selectedTags.length > 0) {
      try {
        await pluginManager.searchMarketplace($searchQuery.trim(), $selectedTags);
      } catch (error) {
        console.error('Search failed:', error);
        addToast('Search failed: ' + (error instanceof Error ? error.message : 'Unknown error'), 'error');
      }
    } else {
      // Load all plugins when no search query
      await loadMarketplace();
    }
  }

  function toggleTag(tag: string) {
    selectedTags.update(tags => {
      if (tags.includes(tag)) {
        return tags.filter(t => t !== tag);
      } else {
        return [...tags, tag];
      }
    });
  }

  function formatNumber(num: number): string {
    if (num >= 1000000) return (num / 1000000).toFixed(1) + 'M';
    if (num >= 1000) return (num / 1000).toFixed(1) + 'K';
    return num.toString();
  }

  function formatDate(dateString: string): string {
    return new Date(dateString).toLocaleDateString('nl-NL');
  }

  async function installPlugin(plugin: MarketplacePlugin) {
    try {
      await pluginManager.installPlugin(plugin.repository, plugin.version);
      addToast(`Plugin ${plugin.name} installed successfully!`, 'success');
      showPluginDetails = false;
    } catch (error) {
      console.error('Installation failed:', error);
      addToast('Installation failed: ' + (error instanceof Error ? error.message : 'Unknown error'), 'error');
    }
  }

  function showDetails(plugin: MarketplacePlugin) {
    selectedPlugin = plugin;
    showPluginDetails = true;
  }

  function closeDetails() {
    showPluginDetails = false;
    selectedPlugin = null;
  }

  function closeModal() {
    closeDetails();
    onClose();
  }

  // Load marketplace when modal opens
  $: if (isOpen) {
    loadMarketplace();
  }

  // Reset search when modal closes
  $: if (!isOpen) {
    searchInput = '';
    searchQuery.set('');
    selectedTags.set([]);
    closeDetails();
  }
</script>

{#if isOpen}
  <!-- Modal Backdrop -->
  <div 
    class="fixed inset-0 bg-black/50 flex items-center justify-center z-50 p-4"
    role="dialog"
    aria-modal="true"
    aria-labelledby="marketplace-title"
    transition:fly={{ y: 50, duration: 200 }}
    on:click={closeModal}
    on:keydown={(e) => e.key === 'Escape' && closeModal()}
  >
    <!-- Main Modal -->
    <div 
      class="max-w-6xl w-full max-h-[90vh] bg-card border border-border rounded-lg shadow-lg flex flex-col"
      role="document"
      on:click|stopPropagation
      on:keydown|stopPropagation
    >
      <!-- Header -->
      <div class="p-6 border-b border-border flex-shrink-0">
        <div class="flex items-center justify-between">
          <div class="flex items-center space-x-3">
            <Icon name="store" size={24} className="text-primary" />
            <div>
              <h2 id="marketplace-title" class="text-xl font-semibold text-foreground">Plugin Marketplace</h2>
              <p class="text-sm text-muted-foreground">Discover and install CloudBox plugins from GitHub</p>
            </div>
          </div>
          <Button on:click={closeModal} variant="outline" size="sm">
            <Icon name="x" size={16} />
          </Button>
        </div>

        <!-- Search and Filters -->
        <div class="mt-4 space-y-4">
          <!-- Search Input -->
          <div class="relative">
            <Icon name="search" size={16} className="absolute left-3 top-1/2 transform -translate-y-1/2 text-muted-foreground" />
            <Input
              bind:value={searchInput}
              placeholder="Search plugins..."
              className="pl-10"
            />
          </div>

          <!-- Tag Filters -->
          {#if availableTags.length > 0}
            <div class="flex flex-wrap gap-2">
              <span class="text-sm text-muted-foreground font-medium">Tags:</span>
              {#each availableTags as tag}
                <Badge
                  variant={$selectedTags.includes(tag) ? 'default' : 'outline'}
                  className="cursor-pointer hover:bg-primary/20 transition-colors"
                  on:click={() => toggleTag(tag)}
                >
                  {tag}
                </Badge>
              {/each}
            </div>
          {/if}
        </div>
      </div>

      <!-- Content -->
      <div class="flex-1 overflow-hidden">
        {#if $marketplaceLoading}
          <div class="p-12 text-center">
            <div class="animate-spin rounded-full h-8 w-8 border-b-2 border-primary mx-auto"></div>
            <p class="mt-4 text-muted-foreground">Loading marketplace...</p>
          </div>
        {:else if $marketplacePlugins.length === 0}
          <div class="p-12 text-center">
            <Icon name="package" size={48} className="mx-auto text-muted-foreground mb-4" />
            <h3 class="text-lg font-medium text-foreground mb-2">No plugins found</h3>
            <p class="text-muted-foreground">
              {#if $searchQuery || $selectedTags.length > 0}
                Try adjusting your search criteria
              {:else}
                No plugins are available in the marketplace
              {/if}
            </p>
          </div>
        {:else}
          <div class="p-6 grid grid-cols-1 lg:grid-cols-2 xl:grid-cols-3 gap-4 overflow-y-auto max-h-full">
            {#each $marketplacePlugins as plugin (plugin.repository)}
              <Card className="p-4 hover:shadow-md transition-shadow cursor-pointer h-fit" on:click={() => showDetails(plugin)}>
                <div class="space-y-3">
                  <!-- Plugin Header -->
                  <div class="flex items-start justify-between">
                    <div class="flex-1 min-w-0">
                      <div class="flex items-center space-x-2">
                        <h3 class="font-semibold text-foreground truncate">{plugin.name}</h3>
                        {#if plugin.verified}
                          <Icon name="shield-check" size={16} className="text-green-500 flex-shrink-0" title="Verified plugin" />
                        {/if}
                      </div>
                      <p class="text-sm text-muted-foreground">v{plugin.version} by {plugin.author}</p>
                    </div>
                  </div>

                  <!-- Description -->
                  <p class="text-sm text-muted-foreground line-clamp-2">{plugin.description}</p>

                  <!-- Tags -->
                  {#if plugin.tags.length > 0}
                    <div class="flex flex-wrap gap-1">
                      {#each plugin.tags.slice(0, 3) as tag}
                        <Badge variant="outline" className="text-xs">
                          {tag}
                        </Badge>
                      {/each}
                      {#if plugin.tags.length > 3}
                        <Badge variant="outline" className="text-xs">
                          +{plugin.tags.length - 3}
                        </Badge>
                      {/if}
                    </div>
                  {/if}

                  <!-- Stats -->
                  <div class="flex items-center justify-between text-xs text-muted-foreground">
                    <div class="flex items-center space-x-4">
                      <div class="flex items-center space-x-1">
                        <Icon name="star" size={12} />
                        <span>{formatNumber(plugin.stars)}</span>
                      </div>
                      <div class="flex items-center space-x-1">
                        <Icon name="download" size={12} />
                        <span>{formatNumber(plugin.downloads)}</span>
                      </div>
                    </div>
                    <span>Updated {formatDate(plugin.last_updated)}</span>
                  </div>

                  <!-- Install Button -->
                  <Button 
                    on:click={() => installPlugin(plugin)}
                    className="w-full"
                    size="sm"
                    disabled={$installationProgress?.pluginName === plugin.name.split('/').pop()}
                  >
                    {#if $installationProgress?.pluginName === plugin.name.split('/').pop()}
                      <div class="animate-spin rounded-full h-4 w-4 border-b-2 border-white mr-2"></div>
                      Installing...
                    {:else}
                      <Icon name="download" size={16} className="mr-2" />
                      Install
                    {/if}
                  </Button>
                </div>
              </Card>
            {/each}
          </div>
        {/if}
      </div>
    </div>
  </div>

  <!-- Plugin Details Modal -->
  {#if showPluginDetails && selectedPlugin}
    <div 
      class="fixed inset-0 bg-black/75 flex items-center justify-center z-60 p-4"
      role="dialog"
      aria-modal="true"
      aria-labelledby="plugin-details-title"
      transition:scale={{ start: 0.95, duration: 200 }}
      on:click={closeDetails}
      on:keydown={(e) => e.key === 'Escape' && closeDetails()}
    >
      <div 
        class="max-w-4xl w-full max-h-[80vh] bg-card border border-border rounded-lg shadow-lg overflow-hidden"
        role="document"
        on:click|stopPropagation
        on:keydown|stopPropagation
      >
        <!-- Plugin Details Header -->
        <div class="p-6 border-b border-border">
          <div class="flex items-center justify-between">
            <div class="flex items-center space-x-3">
              <Icon name="package" size={24} className="text-primary" />
              <div>
                <div class="flex items-center space-x-2">
                  <h2 id="plugin-details-title" class="text-xl font-semibold text-foreground">{selectedPlugin.name}</h2>
                  {#if selectedPlugin.verified}
                    <Icon name="shield-check" size={20} className="text-green-500" title="Verified plugin" />
                  {/if}
                </div>
                <p class="text-sm text-muted-foreground">
                  v{selectedPlugin.version} by {selectedPlugin.author} • {selectedPlugin.license}
                </p>
              </div>
            </div>
            <Button on:click={closeDetails} variant="outline" size="sm">
              <Icon name="x" size={16} />
            </Button>
          </div>
        </div>

        <!-- Plugin Details Content -->
        <div class="p-6 overflow-y-auto max-h-[60vh] space-y-6">
          <!-- Description -->
          <div>
            <h3 class="font-semibold text-foreground mb-2">Description</h3>
            <p class="text-muted-foreground">{selectedPlugin.description}</p>
          </div>

          <!-- Repository -->
          <div>
            <h3 class="font-semibold text-foreground mb-2">Repository</h3>
            <a 
              href="https://github.com/{selectedPlugin.repository}" 
              target="_blank" 
              rel="noopener noreferrer"
              class="text-primary hover:underline font-mono text-sm"
            >
              {selectedPlugin.repository}
            </a>
          </div>

          <!-- Stats -->
          <div class="grid grid-cols-2 md:grid-cols-4 gap-4">
            <div class="text-center p-3 bg-muted rounded-lg">
              <Icon name="star" size={20} className="mx-auto mb-1 text-yellow-500" />
              <div class="text-sm font-medium">{formatNumber(selectedPlugin.stars)}</div>
              <div class="text-xs text-muted-foreground">Stars</div>
            </div>
            <div class="text-center p-3 bg-muted rounded-lg">
              <Icon name="download" size={20} className="mx-auto mb-1 text-blue-500" />
              <div class="text-sm font-medium">{formatNumber(selectedPlugin.downloads)}</div>
              <div class="text-xs text-muted-foreground">Downloads</div>
            </div>
            <div class="text-center p-3 bg-muted rounded-lg">
              <Icon name="calendar" size={20} className="mx-auto mb-1 text-green-500" />
              <div class="text-sm font-medium">{formatDate(selectedPlugin.last_updated)}</div>
              <div class="text-xs text-muted-foreground">Updated</div>
            </div>
            <div class="text-center p-3 bg-muted rounded-lg">
              <Icon name="file-text" size={20} className="mx-auto mb-1 text-purple-500" />
              <div class="text-sm font-medium">{selectedPlugin.license}</div>
              <div class="text-xs text-muted-foreground">License</div>
            </div>
          </div>

          <!-- Tags -->
          {#if selectedPlugin.tags.length > 0}
            <div>
              <h3 class="font-semibold text-foreground mb-2">Tags</h3>
              <div class="flex flex-wrap gap-2">
                {#each selectedPlugin.tags as tag}
                  <Badge variant="outline">{tag}</Badge>
                {/each}
              </div>
            </div>
          {/if}

          <!-- Permissions -->
          {#if selectedPlugin.permissions.length > 0}
            <div>
              <h3 class="font-semibold text-foreground mb-2">Required Permissions</h3>
              <div class="bg-yellow-50 dark:bg-yellow-900/20 border border-yellow-200 dark:border-yellow-800 rounded-lg p-3">
                <div class="flex items-start space-x-2">
                  <Icon name="alert-triangle" size={16} className="text-yellow-600 dark:text-yellow-400 mt-0.5 flex-shrink-0" />
                  <div class="space-y-1">
                    <p class="text-sm text-yellow-800 dark:text-yellow-200 font-medium">
                      This plugin requires the following permissions:
                    </p>
                    <ul class="space-y-1">
                      {#each selectedPlugin.permissions as permission}
                        <li class="text-sm text-yellow-700 dark:text-yellow-300 font-mono">• {permission}</li>
                      {/each}
                    </ul>
                  </div>
                </div>
              </div>
            </div>
          {/if}

          <!-- Dependencies -->
          {#if selectedPlugin.dependencies && Object.keys(selectedPlugin.dependencies).length > 0}
            <div>
              <h3 class="font-semibold text-foreground mb-2">Dependencies</h3>
              <div class="space-y-1">
                {#each Object.entries(selectedPlugin.dependencies) as [name, version]}
                  <div class="text-sm font-mono bg-muted p-2 rounded">
                    {name}: {version}
                  </div>
                {/each}
              </div>
            </div>
          {/if}
        </div>

        <!-- Plugin Details Footer -->
        <div class="p-6 border-t border-border flex justify-between">
          <Button on:click={closeDetails} variant="outline">
            Back to Marketplace
          </Button>
          
          <Button 
            on:click={() => installPlugin(selectedPlugin)}
            disabled={$installationProgress?.pluginName === selectedPlugin.name.split('/').pop()}
          >
            {#if $installationProgress?.pluginName === selectedPlugin.name.split('/').pop()}
              <div class="animate-spin rounded-full h-4 w-4 border-b-2 border-white mr-2"></div>
              Installing...
            {:else}
              <Icon name="download" size={16} className="mr-2" />
              Install Plugin
            {/if}
          </Button>
        </div>
      </div>
    </div>
  {/if}
{/if}

<!-- Installation Progress Toast -->
{#if $installationProgress}
  <div 
    class="fixed bottom-4 right-4 z-50 bg-card border border-border rounded-lg shadow-lg p-4 min-w-80"
    transition:fly={{ x: 300, duration: 200 }}
  >
    <div class="flex items-center space-x-3">
      {#if $installationProgress.status === 'error'}
        <Icon name="x-circle" size={20} className="text-red-500 flex-shrink-0" />
      {:else if $installationProgress.status === 'complete'}
        <Icon name="check-circle" size={20} className="text-green-500 flex-shrink-0" />
      {:else}
        <div class="animate-spin rounded-full h-5 w-5 border-b-2 border-primary flex-shrink-0"></div>
      {/if}
      
      <div class="flex-1 min-w-0">
        <div class="font-medium text-foreground">
          {$installationProgress.pluginName}
        </div>
        <div class="text-sm text-muted-foreground">
          {$installationProgress.message}
        </div>
        
        {#if $installationProgress.status !== 'error' && $installationProgress.status !== 'complete'}
          <div class="mt-2 bg-muted rounded-full h-2 overflow-hidden">
            <div 
              class="bg-primary h-full transition-all duration-300"
              style="width: {$installationProgress.progress}%"
            ></div>
          </div>
        {/if}
      </div>
    </div>
  </div>
{/if}

<style>
  .line-clamp-2 {
    display: -webkit-box;
    -webkit-line-clamp: 2;
    -webkit-box-orient: vertical;
    overflow: hidden;
  }
</style>