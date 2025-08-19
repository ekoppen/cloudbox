<script lang="ts">
  import { createEventDispatcher } from 'svelte';
  import Card from '$lib/components/ui/card.svelte';
  import Button from '$lib/components/ui/button.svelte';
  import Badge from '$lib/components/ui/badge.svelte';
  import Icon from '$lib/components/ui/icon.svelte';

  export let plugin: any;
  export let variant: 'marketplace' | 'installed' = 'marketplace';
  export let projectPlugin: any = null;

  const dispatch = createEventDispatcher();

  function getCategoryIcon(category: string): string {
    switch (category) {
      case 'authentication': return 'auth';
      case 'database': return 'database';
      case 'storage': return 'storage';
      case 'messaging': return 'messaging';
      case 'analytics': return 'dashboard';
      case 'deployment': return 'deployments';
      case 'development': return 'settings';
      default: return 'package';
    }
  }

  function getStatusColor(status: string): string {
    switch (status) {
      case 'enabled': return 'bg-green-100 dark:bg-green-900 text-green-800 dark:text-green-200';
      case 'disabled': return 'bg-gray-100 dark:bg-gray-800 text-gray-600 dark:text-gray-400';
      case 'error': return 'bg-red-100 dark:bg-red-900 text-red-800 dark:text-red-200';
      default: return 'bg-muted text-muted-foreground';
    }
  }

  function formatDate(dateStr: string): string {
    return new Date(dateStr).toLocaleDateString('nl-NL', {
      year: 'numeric',
      month: 'short',
      day: 'numeric',
      hour: '2-digit',
      minute: '2-digit'
    });
  }
</script>

{#if variant === 'marketplace'}
  <div class="group relative bg-background border border-border rounded-xl hover:border-primary/30 transition-all duration-300 hover:shadow-lg hover:shadow-primary/10 overflow-hidden">
    <!-- Card header with enhanced styling -->
    <div class="p-6 space-y-4">
      <!-- Plugin icon and category badge -->
      <div class="flex items-start justify-between">
        <div class="flex items-center space-x-3">
          <div class="w-14 h-14 bg-gradient-to-br from-primary/10 to-accent/10 rounded-xl flex items-center justify-center flex-shrink-0 border border-border/50 shadow-sm">
            {#if plugin.demo_url}
              <img src={plugin.demo_url} alt={plugin.display_name || plugin.name} class="w-8 h-8 rounded" />
            {:else}
              <Icon name={getCategoryIcon(plugin.category)} size={24} className="text-primary group-hover:text-primary/80 transition-colors" />
            {/if}
          </div>
          <div class="flex-1 min-w-0">
            <h3 class="text-lg font-semibold text-foreground truncate group-hover:text-primary transition-colors">
              {plugin.display_name || plugin.name}
            </h3>
            <div class="flex items-center space-x-2 mt-1">
              <span class="text-sm text-muted-foreground">v{plugin.version}</span>
              <span class="text-muted-foreground/50">•</span>
              <span class="text-sm text-muted-foreground">{plugin.author}</span>
            </div>
          </div>
        </div>
        <Badge variant="secondary" class="text-xs px-3 py-1 bg-primary/10 text-primary border border-primary/20">
          {plugin.category}
        </Badge>
      </div>
      
      <!-- Description -->
      <p class="text-sm text-muted-foreground leading-relaxed line-clamp-2 min-h-[2.5rem]">
        {plugin.description}
      </p>
      
      <!-- Tags -->
      {#if plugin.tags && plugin.tags.length > 0}
        <div class="flex flex-wrap gap-2">
          {#each plugin.tags.slice(0, 3) as tag}
            <span class="inline-flex px-2 py-1 text-xs bg-muted/60 text-muted-foreground rounded-md border border-border/50 hover:bg-accent/10 hover:text-accent-foreground transition-colors">
              {tag}
            </span>
          {/each}
          {#if plugin.tags.length > 3}
            <span class="inline-flex px-2 py-1 text-xs bg-muted/60 text-muted-foreground rounded-md border border-border/50">
              +{plugin.tags.length - 3} meer
            </span>
          {/if}
        </div>
      {/if}
    </div>
    
    <!-- Card footer with install button -->
    <div class="px-6 pb-6">
      <Button
        size="sm"
        on:click={() => dispatch('install', plugin)}
        class="w-full bg-primary/10 hover:bg-primary text-primary hover:text-primary-foreground border border-primary/30 hover:border-primary transition-all duration-200 group-hover:shadow-md"
        variant="outline"
      >
        <Icon name="download" size={16} className="mr-2" />
        <span class="font-medium">Installeren</span>
      </Button>
    </div>
    
    <!-- Subtle gradient overlay on hover -->
    <div class="absolute inset-0 bg-gradient-to-br from-transparent via-transparent to-primary/5 opacity-0 group-hover:opacity-100 transition-opacity duration-300 pointer-events-none"></div>
  </div>
{:else}
  <!-- Installed plugin card -->
  <div class="group bg-background border border-border rounded-xl hover:border-primary/30 transition-all duration-300 hover:shadow-md hover:shadow-primary/5 overflow-hidden">
    <div class="p-6">
      <div class="flex items-start justify-between">
        <!-- Plugin info section -->
        <div class="flex items-start space-x-4 flex-1">
          <div class="w-14 h-14 bg-gradient-to-br from-primary/10 to-accent/10 rounded-xl flex items-center justify-center flex-shrink-0 border border-border/50 shadow-sm">
            {#if plugin.demo_url}
              <img src={plugin.demo_url} alt={plugin.display_name || plugin.name} class="w-8 h-8 rounded" />
            {:else}
              <Icon name={getCategoryIcon(plugin.category)} size={24} className="text-primary" />
            {/if}
          </div>
          <div class="flex-1 min-w-0">
            <div class="flex items-center space-x-3 mb-2">
              <h3 class="text-lg font-semibold text-foreground group-hover:text-primary transition-colors">
                {plugin.display_name || plugin.name}
              </h3>
              <Badge 
                variant={projectPlugin?.is_enabled ? "default" : "secondary"}
                class={projectPlugin?.is_enabled ? "bg-success/10 text-success border-success/30" : "bg-muted text-muted-foreground"}
              >
                {projectPlugin?.is_enabled ? 'Actief' : 'Inactief'}
              </Badge>
              {#if projectPlugin?.status}
                <span class="inline-flex px-3 py-1 text-xs font-medium rounded-full border {getStatusColor(projectPlugin.status)}">
                  {projectPlugin.status}
                </span>
              {/if}
            </div>
            <p class="text-sm text-muted-foreground leading-relaxed mb-3">
              {plugin.description}
            </p>
            <div class="flex items-center space-x-4 text-xs text-muted-foreground">
              <span class="font-medium">v{plugin.version}</span>
              {#if projectPlugin?.installed_at}
                <span>Geïnstalleerd: {formatDate(projectPlugin.installed_at)}</span>
              {/if}
              {#if projectPlugin?.updated_at && projectPlugin.updated_at !== projectPlugin.installed_at}
                <span>Bijgewerkt: {formatDate(projectPlugin.updated_at)}</span>
              {/if}
            </div>
          </div>
        </div>

        <!-- Action buttons -->
        <div class="flex items-center space-x-2 flex-shrink-0">
          <Button
            variant="outline"
            size="sm"
            on:click={() => dispatch('configure', { plugin, projectPlugin })}
            class="border-border hover:border-primary/50 hover:bg-primary/5"
          >
            <Icon name="settings" size={14} className="mr-1" />
            Configureren
          </Button>
          <Button
            variant={projectPlugin?.is_enabled ? "outline" : "default"}
            size="sm"
            on:click={() => dispatch('toggle', { plugin, projectPlugin })}
            class={projectPlugin?.is_enabled ? "border-warning/50 text-warning hover:bg-warning/10" : "bg-success/10 text-success border-success/30 hover:bg-success/20"}
          >
            {projectPlugin?.is_enabled ? 'Uitschakelen' : 'Inschakelen'}
          </Button>
          <Button
            variant="outline"
            size="sm"
            on:click={() => dispatch('uninstall', { plugin, projectPlugin })}
            class="border-destructive/30 text-destructive hover:bg-destructive/10 hover:border-destructive/50"
          >
            <Icon name="trash-2" size={14} className="mr-1" />
            Verwijderen
          </Button>
        </div>
      </div>
    </div>
  </div>
{/if}