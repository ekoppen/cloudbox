<script lang="ts">
  import { onMount } from 'svelte';
  import { fly, scale } from 'svelte/transition';
  import Button from '$lib/components/ui/button.svelte';
  import Card from '$lib/components/ui/card.svelte';
  import Icon from '$lib/components/ui/icon.svelte';
  import Badge from '$lib/components/ui/badge.svelte';
  import Input from '$lib/components/ui/input.svelte';
  import AddPluginModal from '$lib/components/admin/add-plugin-modal.svelte';
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
  import { auth } from '$lib/stores/auth';

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
        if (plugin.tags && Array.isArray(plugin.tags)) {
          plugin.tags.forEach(tag => tags.add(tag));
        }
      });
      availableTags = Array.from(tags).sort();
      
      // Show helpful message if no plugins are available
      if (plugins.length === 0) {
        console.warn('No marketplace plugins available - this may be due to API connectivity issues');
      }
    } catch (error) {
      console.error('Failed to load marketplace:', error);
      // Show a more user-friendly error message
      const errorMessage = error instanceof Error ? error.message : 'Unknown error';
      addToast(`Plugin marketplace temporarily unavailable: ${errorMessage}`, 'warning');
      // Keep the modal open but show empty state
    }
  }

  async function performSearch() {
    if ($searchQuery.trim() || $selectedTags.length > 0) {
      try {
        const results = await pluginManager.searchMarketplace($searchQuery.trim(), $selectedTags);
        // Results are already set in the store by the manager, but we can check if empty
        if (results.length === 0 && ($searchQuery.trim() || $selectedTags.length > 0)) {
          console.log('Search returned no results');
        }
      } catch (error) {
        console.error('Search failed:', error);
        // Show warning instead of error for better UX
        addToast('Search temporarily unavailable. Please try again later.', 'warning');
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

  function formatNumber(num: number | undefined): string {
    if (!num && num !== 0) return '0';
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
  <!-- Enhanced Modal Backdrop with proper coverage -->
  <div 
    class="fixed inset-0 bg-black/50 backdrop-blur-sm flex items-center justify-center z-50 p-4 overflow-hidden"
    role="dialog"
    aria-modal="true"
    aria-labelledby="marketplace-title"
    tabindex="0"
    style="width: 100vw; height: 100vh; top: 0; left: 0;"
    transition:fly={{ y: 50, duration: 200 }}
    on:click={closeModal}
    on:keydown={(e) => e.key === 'Escape' && closeModal()}
  >
    <!-- Main Modal -->
    <div 
      class="max-w-7xl w-full max-h-[90vh] bg-background border border-border rounded-xl shadow-2xl flex flex-col overflow-hidden relative z-10"
      role="document"
      on:click|stopPropagation
    >
      <!-- Header -->
      <div class="px-8 py-6 border-b border-border flex-shrink-0 bg-background">
        <div class="flex items-center justify-between">
          <div class="flex items-center space-x-4">
            <div class="w-12 h-12 bg-primary/10 rounded-xl flex items-center justify-center">
              <Icon name="store" size={24} className="text-primary" />
            </div>
            <div>
              <h2 id="marketplace-title" class="text-2xl font-bold text-foreground tracking-tight font-sans antialiased">Plugin Marketplace</h2>
              <p class="text-sm text-muted-foreground mt-1 font-medium">Discover and install CloudBox plugins from the community</p>
            </div>
          </div>
          <Button on:click={closeModal} variant="ghost" size="sm" className="h-8 w-8 p-0 hover:bg-muted">
            <Icon name="x" size={18} />
          </Button>
        </div>

        <!-- Search and Filters -->
        <div class="mt-6 space-y-4">
          <!-- Search Input -->
          <div class="relative">
            <Icon name="search" size={18} className="absolute left-4 top-1/2 transform -translate-y-1/2 text-muted-foreground" />
            <Input
              bind:value={searchInput}
              placeholder="Search plugins by name, description, or author..."
              className="pl-12 h-12 text-base border-border focus:border-primary focus:ring-1 focus:ring-primary"
            />
          </div>

          <!-- Tag Filters -->
          {#if availableTags.length > 0}
            <div class="flex flex-wrap items-center gap-3">
              <span class="text-sm text-muted-foreground font-medium">Filter by tags:</span>
              <div class="flex flex-wrap gap-2">
                {#each availableTags as tag}
                  <Badge
                    variant={$selectedTags.includes(tag) ? 'default' : 'outline'}
                    className="cursor-pointer hover:scale-105 transition-all duration-200 px-3 py-1"
                    on:click={() => toggleTag(tag)}
                  >
                    {tag}
                  </Badge>
                {/each}
              </div>
            </div>
          {/if}
        </div>
      </div>

      <!-- Content -->
      <div class="flex-1 overflow-hidden bg-muted/30">
        {#if $marketplaceLoading}
          <div class="p-16 text-center">
            <div class="w-12 h-12 bg-primary/10 rounded-xl flex items-center justify-center mx-auto mb-4">
              <div class="animate-spin rounded-full h-6 w-6 border-b-2 border-primary"></div>
            </div>
            <p class="text-foreground font-medium">Loading marketplace...</p>
            <p class="text-sm text-muted-foreground mt-1">Discovering available plugins</p>
          </div>
        {:else if $marketplacePlugins.length === 0}
          <div class="p-16 text-center">
            <div class="w-16 h-16 bg-muted rounded-2xl flex items-center justify-center mx-auto mb-6">
              <Icon name="package" size={32} className="text-muted-foreground" />
            </div>
            <h3 class="text-lg font-semibold text-foreground mb-2">No plugins found</h3>
            <p class="text-muted-foreground mb-6 max-w-md mx-auto">
              {#if $searchQuery || $selectedTags.length > 0}
                Try adjusting your search criteria or clearing filters to see all available plugins.
              {:else}
                The plugin marketplace appears to be empty or temporarily unavailable.
              {/if}
            </p>
            {#if !$searchQuery && $selectedTags.length === 0}
              <div class="bg-amber-50 border border-amber-200 rounded-xl p-6 max-w-lg mx-auto">
                <div class="flex items-start space-x-3">
                  <div class="w-8 h-8 bg-amber-100 rounded-lg flex items-center justify-center flex-shrink-0">
                    <Icon name="info" size={16} className="text-amber-600" />
                  </div>
                  <div class="text-sm text-amber-800 text-left">
                    <p class="font-medium mb-2">Marketplace Status</p>
                    <p class="mb-4 leading-relaxed">The plugin marketplace API may be temporarily unavailable or no plugins have been added yet.</p>
                    <Button 
                      on:click={loadMarketplace} 
                      variant="outline" 
                      size="sm"
                      className="bg-amber-100 border-amber-300 text-amber-800 hover:bg-amber-200"
                    >
                      <Icon name="refresh-cw" size={14} className="mr-2" />
                      Try Again
                    </Button>
                  </div>
                </div>
              </div>
            {/if}
          </div>
        {:else}
          <div class="p-8 overflow-y-auto" style="max-height: calc(90vh - 280px);">
            <div class="grid grid-cols-1 lg:grid-cols-2 xl:grid-cols-3 2xl:grid-cols-4 gap-6">
            {#each $marketplacePlugins as plugin (plugin.repository)}
              <div 
                class="group relative bg-background border border-border rounded-xl hover:border-primary/30 transition-all duration-300 hover:shadow-lg hover:shadow-primary/10 cursor-pointer flex flex-col overflow-hidden" 
                role="button"
                tabindex="0"
                on:click={() => showDetails(plugin)}
                on:keydown={(e) => (e.key === 'Enter' || e.key === ' ') && showDetails(plugin)}
              >
                <!-- Plugin Header -->
                <div class="p-6 space-y-4 flex-1">
                  <div class="flex items-start justify-between">
                    <div class="flex items-center space-x-3 flex-1 min-w-0">
                      <div class="w-12 h-12 bg-gradient-to-br from-primary/10 to-accent/10 rounded-xl flex items-center justify-center flex-shrink-0 border border-border/50 shadow-sm">
                        <Icon name="package" size={20} className="text-primary group-hover:text-primary/80 transition-colors" />
                      </div>
                      <div class="flex-1 min-w-0">
                        <div class="flex items-center space-x-2 mb-1">
                          <h3 class="font-bold text-foreground truncate text-base group-hover:text-primary transition-colors tracking-tight font-sans">{plugin.name}</h3>
                          {#if plugin.verified}
                            <div class="w-5 h-5 bg-success/20 rounded-full flex items-center justify-center border border-success/30">
                              <Icon name="shield-check" size={12} className="text-success" title="Verified plugin" />
                            </div>
                          {/if}
                        </div>
                        <div class="flex items-center space-x-2 text-sm text-muted-foreground">
                          <span>v{plugin.version}</span>
                          <span class="text-muted-foreground/50">•</span>
                          <span>{plugin.author}</span>
                        </div>
                      </div>
                    </div>
                  </div>

                  <!-- Description -->
                  <p class="text-sm text-muted-foreground line-clamp-3 leading-relaxed min-h-[3.75rem]">{plugin.description}</p>

                  <!-- Tags -->
                  {#if plugin.tags && plugin.tags.length > 0}
                    <div class="flex flex-wrap gap-2">
                      {#each plugin.tags.slice(0, 3) as tag}
                        <Badge variant="secondary" className="text-xs px-2 py-1 bg-muted/60 text-muted-foreground border border-border/50">
                          {tag}
                        </Badge>
                      {/each}
                      {#if plugin.tags.length > 3}
                        <Badge variant="secondary" className="text-xs px-2 py-1 bg-muted/60 text-muted-foreground border border-border/50">
                          +{plugin.tags.length - 3} meer
                        </Badge>
                      {/if}
                    </div>
                  {/if}

                  <!-- Stats -->
                  <div class="flex items-center justify-between text-xs text-muted-foreground pt-3 border-t border-border/50">
                    <div class="flex items-center space-x-4">
                      <div class="flex items-center space-x-1">
                        <Icon name="star" size={12} className="text-warning" />
                        <span class="font-medium">{formatNumber(plugin.stars)}</span>
                      </div>
                      <div class="flex items-center space-x-1">
                        <Icon name="download" size={12} className="text-primary" />
                        <span class="font-medium">{formatNumber(plugin.downloads)}</span>
                      </div>
                    </div>
                    <span>{formatDate(plugin.last_updated)}</span>
                  </div>
                </div>

                <!-- Install Button -->
                <div class="p-6 pt-0">
                  <Button 
                    on:click={(e) => { e.stopPropagation(); installPlugin(plugin); }}
                    className="w-full bg-primary/10 hover:bg-primary text-primary hover:text-primary-foreground border border-primary/30 hover:border-primary transition-all duration-200 group-hover:shadow-md"
                    variant="outline"
                    size="sm"
                    disabled={$installationProgress?.pluginName === plugin.name.split('/').pop()}
                  >
                    {#if $installationProgress?.pluginName === plugin.name.split('/').pop()}
                      <div class="animate-spin rounded-full h-4 w-4 border-b-2 border-current mr-2"></div>
                      Installeren...
                    {:else}
                      <Icon name="download" size={16} className="mr-2" />
                      <span class="font-medium">Installeren</span>
                    {/if}
                  </Button>
                </div>
                
                <!-- Subtle gradient overlay on hover -->
                <div class="absolute inset-0 bg-gradient-to-br from-transparent via-transparent to-primary/5 opacity-0 group-hover:opacity-100 transition-opacity duration-300 pointer-events-none"></div>
              </div>
            {/each}
            </div>
          </div>
        {/if}
      </div>
    </div>
  </div>

  <!-- Plugin Details Modal -->
  {#if showPluginDetails && selectedPlugin}
    <div 
      class="fixed inset-0 bg-black/70 backdrop-blur-md flex items-center justify-center z-60 p-4 overflow-hidden"
      role="dialog"
      aria-modal="true"
      aria-labelledby="plugin-details-title"
      tabindex="0"
      style="width: 100vw; height: 100vh; top: 0; left: 0;"
      transition:scale={{ start: 0.95, duration: 200 }}
      on:click={closeDetails}
      on:keydown={(e) => e.key === 'Escape' && closeDetails()}
    >
      <div 
        class="max-w-4xl w-full max-h-[85vh] bg-background border border-border rounded-2xl shadow-2xl flex flex-col overflow-hidden relative z-10"
        role="document"
        on:click|stopPropagation
      >
        <!-- Plugin Details Header -->
        <div class="px-8 py-6 border-b border-border flex-shrink-0 bg-background">
          <div class="flex items-center justify-between">
            <div class="flex items-center space-x-4">
              <div class="w-12 h-12 bg-primary/10 rounded-xl flex items-center justify-center">
                <Icon name="package" size={24} className="text-primary" />
              </div>
              <div>
                <div class="flex items-center space-x-3 mb-1">
                  <h2 id="plugin-details-title" class="text-2xl font-bold text-foreground tracking-tight font-sans antialiased">{selectedPlugin.name}</h2>
                  {#if selectedPlugin.verified}
                    <div class="w-6 h-6 bg-green-100 rounded-full flex items-center justify-center">
                      <Icon name="shield-check" size={14} className="text-green-600" title="Verified plugin" />
                    </div>
                  {/if}
                </div>
                <p class="text-sm text-muted-foreground">
                  v{selectedPlugin.version} by {selectedPlugin.author} • {selectedPlugin.license}
                </p>
              </div>
            </div>
            <Button on:click={closeDetails} variant="ghost" size="sm" className="h-8 w-8 p-0 hover:bg-muted">
              <Icon name="x" size={18} />
            </Button>
          </div>
        </div>

        <!-- Plugin Details Content -->
        <div class="px-8 py-6 overflow-y-auto flex-1 space-y-8 bg-muted/20">
          <!-- Description -->
          <div>
            <h3 class="font-semibold text-foreground mb-3 text-lg">About this plugin</h3>
            <p class="text-muted-foreground leading-relaxed">{selectedPlugin.description}</p>
          </div>

          <!-- Repository -->
          <div>
            <h3 class="font-semibold text-foreground mb-3 text-lg">Repository</h3>
            <div class="bg-background border border-border rounded-lg p-4">
              <a 
                href="https://github.com/{selectedPlugin.repository}" 
                target="_blank" 
                rel="noopener noreferrer"
                class="text-primary hover:text-primary/80 font-mono text-sm flex items-center space-x-2 transition-colors"
              >
                <Icon name="external-link" size={16} />
                <span>{selectedPlugin.repository}</span>
              </a>
            </div>
          </div>

          <!-- Stats -->
          <div class="grid grid-cols-2 md:grid-cols-4 gap-4">
            <div class="text-center p-4 bg-background border border-border rounded-xl">
              <div class="w-10 h-10 bg-yellow-100 rounded-lg flex items-center justify-center mx-auto mb-2">
                <Icon name="star" size={20} className="text-yellow-600" />
              </div>
              <div class="text-lg font-semibold text-foreground">{formatNumber(selectedPlugin.stars)}</div>
              <div class="text-sm text-muted-foreground">Stars</div>
            </div>
            <div class="text-center p-4 bg-background border border-border rounded-xl">
              <div class="w-10 h-10 bg-blue-100 rounded-lg flex items-center justify-center mx-auto mb-2">
                <Icon name="download" size={20} className="text-blue-600" />
              </div>
              <div class="text-lg font-semibold text-foreground">{formatNumber(selectedPlugin.downloads)}</div>
              <div class="text-sm text-muted-foreground">Downloads</div>
            </div>
            <div class="text-center p-4 bg-background border border-border rounded-xl">
              <div class="w-10 h-10 bg-green-100 rounded-lg flex items-center justify-center mx-auto mb-2">
                <Icon name="calendar" size={20} className="text-green-600" />
              </div>
              <div class="text-lg font-semibold text-foreground">{formatDate(selectedPlugin.last_updated)}</div>
              <div class="text-sm text-muted-foreground">Last updated</div>
            </div>
            <div class="text-center p-4 bg-background border border-border rounded-xl">
              <div class="w-10 h-10 bg-purple-100 rounded-lg flex items-center justify-center mx-auto mb-2">
                <Icon name="file-text" size={20} className="text-purple-600" />
              </div>
              <div class="text-lg font-semibold text-foreground">{selectedPlugin.license}</div>
              <div class="text-sm text-muted-foreground">License</div>
            </div>
          </div>

          <!-- Tags -->
          {#if selectedPlugin.tags && selectedPlugin.tags.length > 0}
            <div>
              <h3 class="font-semibold text-foreground mb-3 text-lg">Tags</h3>
              <div class="flex flex-wrap gap-2">
                {#each selectedPlugin.tags as tag}
                  <Badge variant="secondary" className="px-3 py-1">{tag}</Badge>
                {/each}
              </div>
            </div>
          {/if}

          <!-- Permissions -->
          {#if selectedPlugin.permissions.length > 0}
            <div>
              <h3 class="font-semibold text-foreground mb-3 text-lg">Required Permissions</h3>
              <div class="bg-amber-50 border border-amber-200 rounded-xl p-4">
                <div class="flex items-start space-x-3">
                  <div class="w-8 h-8 bg-amber-100 rounded-lg flex items-center justify-center flex-shrink-0">
                    <Icon name="alert-triangle" size={16} className="text-amber-600" />
                  </div>
                  <div class="space-y-2">
                    <p class="text-sm text-amber-800 font-medium">
                      This plugin requires the following permissions:
                    </p>
                    <ul class="space-y-1">
                      {#each selectedPlugin.permissions as permission}
                        <li class="text-sm text-amber-700 font-mono bg-amber-100 px-2 py-1 rounded">• {permission}</li>
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
              <h3 class="font-semibold text-foreground mb-3 text-lg">Dependencies</h3>
              <div class="bg-background border border-border rounded-xl p-4 space-y-2">
                {#each Object.entries(selectedPlugin.dependencies) as [name, version]}
                  <div class="flex items-center justify-between text-sm bg-muted px-3 py-2 rounded-lg">
                    <span class="font-mono">{name}</span>
                    <span class="text-muted-foreground">{version}</span>
                  </div>
                {/each}
              </div>
            </div>
          {/if}
        </div>

        <!-- Plugin Details Footer -->
        <div class="px-8 py-6 border-t border-border flex justify-between items-center flex-shrink-0 bg-background">
          <Button on:click={closeDetails} variant="outline" className="flex items-center space-x-2">
            <Icon name="arrow-left" size={16} />
            <span>Back to Marketplace</span>
          </Button>
          
          <Button 
            on:click={() => installPlugin(selectedPlugin)}
            disabled={$installationProgress?.pluginName === selectedPlugin.name.split('/').pop()}
            size="lg"
            className="px-6"
          >
            {#if $installationProgress?.pluginName === selectedPlugin.name.split('/').pop()}
              <div class="animate-spin rounded-full h-4 w-4 border-b-2 border-white mr-2"></div>
              Installing...
            {:else}
              <Icon name="download" size={18} className="mr-2" />
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

