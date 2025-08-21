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

<style>
  .glassmorphism-card {
    background: rgba(255, 255, 255, 0.85);
    backdrop-filter: blur(20px);
    border: 1px solid rgba(255, 255, 255, 0.2);
    border-radius: 16px;
    padding: 24px;
    box-shadow: 
      0 8px 25px -8px rgba(0, 0, 0, 0.1),
      0 4px 12px -4px rgba(0, 0, 0, 0.08),
      0 0 0 1px rgba(255, 255, 255, 0.05) inset;
    transition: all 0.3s cubic-bezier(0.4, 0.0, 0.2, 1);
    position: relative;
    overflow: hidden;
  }

  .glassmorphism-card:hover {
    transform: translateY(-2px);
    box-shadow: 
      0 12px 35px -12px rgba(0, 0, 0, 0.15),
      0 8px 20px -8px rgba(0, 0, 0, 0.12),
      0 0 0 1px rgba(255, 255, 255, 0.1) inset;
  }

  .glassmorphism-card::before {
    content: '';
    position: absolute;
    top: 0;
    left: 0;
    right: 0;
    bottom: 0;
    background: linear-gradient(
      135deg,
      rgba(255, 255, 255, 0.1) 0%,
      rgba(255, 255, 255, 0) 50%,
      rgba(0, 0, 0, 0.05) 100%
    );
    pointer-events: none;
    z-index: 1;
  }

  .glassmorphism-card > * {
    position: relative;
    z-index: 2;
  }

  .glassmorphism-icon {
    width: 40px;
    height: 40px;
    border-radius: 10px;
    display: flex;
    align-items: center;
    justify-content: center;
    backdrop-filter: blur(10px);
    border: 1px solid rgba(255, 255, 255, 0.2);
    transition: all 0.2s ease;
  }

  .glassmorphism-icon:hover {
    transform: scale(1.05);
  }

  .glassmorphism-modal {
    background: rgba(255, 255, 255, 0.9);
    backdrop-filter: blur(30px);
    border: 1px solid rgba(255, 255, 255, 0.3);
    box-shadow: 
      0 20px 40px -12px rgba(0, 0, 0, 0.2),
      0 8px 20px -8px rgba(0, 0, 0, 0.15),
      0 0 0 1px rgba(255, 255, 255, 0.1) inset;
  }

  .glassmorphism-badge {
    background: rgba(255, 255, 255, 0.7);
    backdrop-filter: blur(8px);
    border: 1px solid rgba(255, 255, 255, 0.3);
    transition: all 0.2s ease;
  }

  .glassmorphism-badge:hover {
    background: rgba(255, 255, 255, 0.9);
    transform: translateY(-1px);
    box-shadow: 0 4px 8px -2px rgba(0, 0, 0, 0.1);
  }

  .glassmorphism-search {
    background: rgba(255, 255, 255, 0.8);
    backdrop-filter: blur(15px);
    border: 1px solid rgba(255, 255, 255, 0.3);
    transition: all 0.3s ease;
  }

  .glassmorphism-search:focus {
    background: rgba(255, 255, 255, 0.95);
    border-color: rgba(59, 130, 246, 0.5);
    box-shadow: 0 0 0 3px rgba(59, 130, 246, 0.1);
  }

  /* Dark mode support - CloudBox theme system */
  :global(.cloudbox-dark) .glassmorphism-card {
    background: rgba(26, 26, 26, 0.7);
    border: 1px solid rgba(255, 255, 255, 0.1);
    box-shadow: 
      0 8px 25px -8px rgba(0, 0, 0, 0.6),
      0 4px 12px -4px rgba(0, 0, 0, 0.4),
      0 0 0 1px rgba(255, 255, 255, 0.05) inset;
    backdrop-filter: blur(25px);
  }
  
  :global(.cloudbox-dark) .glassmorphism-card:hover {
    background: rgba(33, 33, 33, 0.8);
    box-shadow: 
      0 12px 35px -12px rgba(0, 0, 0, 0.7),
      0 8px 20px -8px rgba(0, 0, 0, 0.5),
      0 0 0 1px rgba(255, 255, 255, 0.1) inset;
    backdrop-filter: blur(30px);
  }
  
  :global(.cloudbox-dark) .glassmorphism-card::before {
    background: linear-gradient(
      135deg,
      rgba(255, 255, 255, 0.03) 0%,
      rgba(255, 255, 255, 0) 50%,
      rgba(0, 0, 0, 0.15) 100%
    );
  }
  
  :global(.cloudbox-dark) .glassmorphism-icon {
    border: 1px solid rgba(255, 255, 255, 0.15);
    background: rgba(255, 255, 255, 0.05);
  }

  :global(.cloudbox-dark) .glassmorphism-modal {
    background: rgba(26, 26, 26, 0.85);
    border: 1px solid rgba(255, 255, 255, 0.1);
    backdrop-filter: blur(40px);
    box-shadow: 
      0 20px 40px -12px rgba(0, 0, 0, 0.6),
      0 8px 20px -8px rgba(0, 0, 0, 0.4),
      0 0 0 1px rgba(255, 255, 255, 0.05) inset;
  }

  :global(.cloudbox-dark) .glassmorphism-badge {
    background: rgba(30, 41, 59, 0.7);
    border: 1px solid rgba(255, 255, 255, 0.15);
    backdrop-filter: blur(10px);
  }

  :global(.cloudbox-dark) .glassmorphism-badge:hover {
    background: rgba(30, 41, 59, 0.9);
    border: 1px solid rgba(255, 255, 255, 0.2);
  }

  :global(.cloudbox-dark) .glassmorphism-search {
    background: rgba(30, 41, 59, 0.8);
    border: 1px solid rgba(255, 255, 255, 0.1);
    backdrop-filter: blur(15px);
  }

  :global(.cloudbox-dark) .glassmorphism-search:focus {
    background: rgba(30, 41, 59, 0.9);
    border-color: rgba(102, 126, 234, 0.5);
    box-shadow: 0 0 0 3px rgba(102, 126, 234, 0.15);
  }

  /* Mobile responsiveness */
  @media (max-width: 640px) {
    .glassmorphism-card {
      padding: 20px;
      border-radius: 12px;
    }
    
    .glassmorphism-icon {
      width: 36px;
      height: 36px;
    }
  }

  /* Reduce motion for accessibility */
  @media (prefers-reduced-motion: reduce) {
    .glassmorphism-card,
    .glassmorphism-icon,
    .glassmorphism-badge {
      transition: none;
    }
    
    .glassmorphism-card:hover {
      transform: none;
    }
  }
</style>

{#if isOpen}
  <!-- Enhanced Modal Backdrop with proper full viewport coverage and scrolling -->
  <div 
    class="fixed inset-0 bg-black/60 backdrop-blur-sm flex items-start justify-center z-50 p-4 pt-16 sm:pt-20 overflow-y-auto"
    role="dialog"
    aria-modal="true"
    aria-labelledby="marketplace-title"
    tabindex="0"
    style="width: 100vw; height: 100vh; top: 0; left: 0; position: fixed;"
    transition:fly={{ y: 50, duration: 200 }}
    on:click={closeModal}
    on:keydown={(e) => e.key === 'Escape' && closeModal()}
  >
    <!-- Main Modal with better height management -->
    <div 
      class="max-w-7xl w-full glassmorphism-modal rounded-xl flex flex-col my-auto min-h-0"
      style="max-height: calc(100vh - 8rem);"
      role="document"
      on:click|stopPropagation
    >
      <!-- Header -->
      <div class="px-8 py-6 border-b border-white/10 flex-shrink-0 bg-white/5">
        <div class="flex items-center justify-between">
          <div class="flex items-center space-x-4">
            <div class="glassmorphism-icon bg-primary/20">
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
              className="pl-12 h-12 text-base glassmorphism-search"
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
                    className="cursor-pointer glassmorphism-badge px-3 py-1 {$selectedTags.includes(tag) ? 'bg-primary/30 border-primary/50' : ''}"
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
      <div class="flex-1 flex flex-col overflow-hidden bg-white/5">
        {#if $marketplaceLoading}
          <div class="p-16 text-center">
            <div class="glassmorphism-icon bg-primary/20 mx-auto mb-4">
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
              <div class="status-warning border rounded-xl p-6 max-w-lg mx-auto">
                <div class="flex items-start space-x-3">
                  <div class="w-8 h-8 bg-warning-light rounded-lg flex items-center justify-center flex-shrink-0">
                    <Icon name="info" size={16} className="icon-warning" />
                  </div>
                  <div class="text-sm text-left">
                    <p class="font-medium mb-2">Marketplace Status</p>
                    <p class="mb-4 leading-relaxed">The plugin marketplace API may be temporarily unavailable or no plugins have been added yet.</p>
                    <Button 
                      on:click={loadMarketplace} 
                      variant="ghost"
                      size="icon"
                      className="hover:rotate-180 transition-transform duration-300"
                      title="Vernieuw marketplace"
                    >
                      <Icon name="refresh-cw" size={16} />
                    </Button>
                  </div>
                </div>
              </div>
            {/if}
          </div>
        {:else}
          <div class="flex-1 overflow-y-auto p-8 min-h-0">
            <div class="grid grid-cols-1 lg:grid-cols-2 xl:grid-cols-3 2xl:grid-cols-4 gap-6">
            {#each $marketplacePlugins as plugin (plugin.repository)}
              <div 
                class="group relative glassmorphism-card cursor-pointer flex flex-col overflow-hidden" 
                role="button"
                tabindex="0"
                on:click={() => showDetails(plugin)}
                on:keydown={(e) => (e.key === 'Enter' || e.key === ' ') && showDetails(plugin)}
              >
                <!-- Plugin Header -->
                <div class="p-6 space-y-4 flex-1">
                  <div class="flex items-start justify-between">
                    <div class="flex items-center space-x-3 flex-1 min-w-0">
                      <div class="glassmorphism-icon bg-gradient-to-br from-primary/20 to-accent/20 flex-shrink-0">
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
                        <Badge variant="secondary" className="text-xs px-2 py-1 glassmorphism-badge">
                          {tag}
                        </Badge>
                      {/each}
                      {#if plugin.tags.length > 3}
                        <Badge variant="secondary" className="text-xs px-2 py-1 glassmorphism-badge">
                          +{plugin.tags.length - 3} meer
                        </Badge>
                      {/if}
                    </div>
                  {/if}

                  <!-- Stats -->
                  <div class="flex items-center justify-between text-xs text-muted-foreground pt-3 border-t border-white/10">
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
                    className="w-full glassmorphism-card bg-primary/20 hover:bg-primary text-primary hover:text-primary-foreground border-primary/30 hover:border-primary"
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
                
                <!-- Enhanced gradient overlay on hover -->
                <div class="absolute inset-0 bg-gradient-to-br from-transparent via-transparent to-primary/8 opacity-0 group-hover:opacity-100 transition-opacity duration-300 pointer-events-none rounded-xl"></div>
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
      class="fixed inset-0 bg-black/60 backdrop-blur-lg flex items-center justify-center z-60 p-4"
      role="dialog"
      aria-modal="true"
      aria-labelledby="plugin-details-title"
      tabindex="0"
      style="width: 100vw; height: 100vh; top: 0; left: 0; position: fixed;"
      transition:scale={{ start: 0.95, duration: 200 }}
      on:click={closeDetails}
      on:keydown={(e) => e.key === 'Escape' && closeDetails()}
    >
      <div 
        class="max-w-4xl w-full max-h-[90vh] glassmorphism-modal rounded-2xl flex flex-col overflow-hidden relative z-10"
        role="document"
        on:click|stopPropagation
      >
        <!-- Plugin Details Header -->
        <div class="px-8 py-6 border-b border-white/10 flex-shrink-0 bg-white/5">
          <div class="flex items-center justify-between">
            <div class="flex items-center space-x-4">
              <div class="glassmorphism-icon bg-primary/20">
                <Icon name="package" size={24} className="text-primary" />
              </div>
              <div>
                <div class="flex items-center space-x-3 mb-1">
                  <h2 id="plugin-details-title" class="text-2xl font-bold text-foreground tracking-tight font-sans antialiased">{selectedPlugin.name}</h2>
                  {#if selectedPlugin.verified}
                    <div class="w-6 h-6 bg-success-light rounded-full flex items-center justify-center">
                      <Icon name="shield-check" size={14} className="icon-success" title="Verified plugin" />
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
        <div class="flex-1 overflow-y-auto px-8 py-6 space-y-8 bg-white/5" style="max-height: calc(90vh - 160px);">
          <!-- Description -->
          <div>
            <h3 class="font-semibold text-foreground mb-3 text-lg">About this plugin</h3>
            <p class="text-muted-foreground leading-relaxed">{selectedPlugin.description}</p>
          </div>

          <!-- Repository -->
          <div>
            <h3 class="font-semibold text-foreground mb-3 text-lg">Repository</h3>
            <div class="glassmorphism-card rounded-lg p-4">
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
            <div class="text-center glassmorphism-card">
              <div class="glassmorphism-icon bg-warning/20 mx-auto mb-2">
                <Icon name="star" size={20} className="text-warning" />
              </div>
              <div class="text-lg font-semibold text-foreground">{formatNumber(selectedPlugin.stars)}</div>
              <div class="text-sm text-muted-foreground">Stars</div>
            </div>
            <div class="text-center glassmorphism-card">
              <div class="glassmorphism-icon bg-info/20 mx-auto mb-2">
                <Icon name="download" size={20} className="text-info" />
              </div>
              <div class="text-lg font-semibold text-foreground">{formatNumber(selectedPlugin.downloads)}</div>
              <div class="text-sm text-muted-foreground">Downloads</div>
            </div>
            <div class="text-center glassmorphism-card">
              <div class="glassmorphism-icon bg-success/20 mx-auto mb-2">
                <Icon name="calendar" size={20} className="text-success" />
              </div>
              <div class="text-lg font-semibold text-foreground">{formatDate(selectedPlugin.last_updated)}</div>
              <div class="text-sm text-muted-foreground">Last updated</div>
            </div>
            <div class="text-center glassmorphism-card">
              <div class="glassmorphism-icon bg-info/20 mx-auto mb-2">
                <Icon name="file-text" size={20} className="text-info" />
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
                  <Badge variant="secondary" className="px-3 py-1 glassmorphism-badge">{tag}</Badge>
                {/each}
              </div>
            </div>
          {/if}

          <!-- Permissions -->
          {#if selectedPlugin.permissions.length > 0}
            <div>
              <h3 class="font-semibold text-foreground mb-3 text-lg">Required Permissions</h3>
              <div class="status-warning border rounded-xl p-4">
                <div class="flex items-start space-x-3">
                  <div class="w-8 h-8 bg-warning-light rounded-lg flex items-center justify-center flex-shrink-0">
                    <Icon name="alert-triangle" size={16} className="icon-warning" />
                  </div>
                  <div class="space-y-2">
                    <p class="text-sm font-medium">
                      This plugin requires the following permissions:
                    </p>
                    <ul class="space-y-1">
                      {#each selectedPlugin.permissions as permission}
                        <li class="text-sm font-mono bg-warning-light px-2 py-1 rounded">• {permission}</li>
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
              <div class="glassmorphism-card rounded-xl space-y-2">
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
        <div class="px-8 py-6 border-t border-white/10 flex justify-between items-center flex-shrink-0 bg-white/5">
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
    class="fixed bottom-4 right-4 z-50 glassmorphism-card rounded-lg p-4 min-w-80"
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
          <div class="mt-2 bg-white/20 rounded-full h-2 overflow-hidden">
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

