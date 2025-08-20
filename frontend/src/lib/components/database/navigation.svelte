<script lang="ts">
  import { createEventDispatcher } from 'svelte';
  import Button from '$lib/components/ui/button.svelte';
  import Badge from '$lib/components/ui/badge.svelte';
  import Icon from '$lib/components/ui/icon.svelte';

  const dispatch = createEventDispatcher();

  interface DatabaseSection {
    id: string;
    name: string;
    icon: string;
    count?: number;
    badge?: string;
    active?: boolean;
    children?: DatabaseSection[];
  }

  export let sections: DatabaseSection[] = [];
  export let activeSection: string = '';
  export let collapsible: boolean = true;
  export let showCounts: boolean = true;

  let expandedSections: Set<string> = new Set();

  // Default sections if none provided
  $: defaultSections = sections.length > 0 ? sections : [
    {
      id: 'tables',
      name: 'Tables',
      icon: 'table',
      count: 0
    },
    {
      id: 'views',
      name: 'Views',
      icon: 'eye',
      count: 0
    },
    {
      id: 'functions',
      name: 'Functions',
      icon: 'code',
      count: 0
    },
    {
      id: 'indexes',
      name: 'Indexes',
      icon: 'zap',
      count: 0
    },
    {
      id: 'triggers',
      name: 'Triggers',
      icon: 'play',
      count: 0
    },
    {
      id: 'extensions',
      name: 'Extensions',
      icon: 'puzzle',
      count: 0
    }
  ];

  function handleSectionClick(section: DatabaseSection) {
    if (section.children && collapsible) {
      toggleSection(section.id);
    } else {
      dispatch('sectionSelect', { section });
    }
  }

  function toggleSection(sectionId: string) {
    if (expandedSections.has(sectionId)) {
      expandedSections.delete(sectionId);
    } else {
      expandedSections.add(sectionId);
    }
    expandedSections = new Set(expandedSections);
  }

  function isExpanded(sectionId: string): boolean {
    return expandedSections.has(sectionId);
  }

  function getBadgeVariant(badge: string): 'default' | 'secondary' | 'destructive' | 'outline' {
    switch (badge.toLowerCase()) {
      case 'new':
        return 'default';
      case 'error':
        return 'destructive';
      case 'warning':
        return 'outline';
      default:
        return 'secondary';
    }
  }

  function getCountColor(count: number): string {
    if (count === 0) return 'text-muted-foreground';
    if (count < 10) return 'text-blue-600 dark:text-blue-400';
    if (count < 50) return 'text-green-600 dark:text-green-400';
    return 'text-orange-600 dark:text-orange-400';
  }
</script>

<div class="h-full flex flex-col">
  <!-- Header -->
  <div class="p-4 border-b border-border">
    <div class="flex items-center space-x-2">
      <Icon name="database" size={16} className="text-primary" />
      <h3 class="text-sm font-semibold text-foreground">Database Schema</h3>
    </div>
  </div>

  <!-- Navigation List -->
  <div class="flex-1 overflow-auto">
    <nav class="p-2 space-y-1">
      {#each defaultSections as section}
        <div class="space-y-1">
          <!-- Main Section -->
          <Button
            variant="ghost"
            class="w-full justify-start h-auto py-2 px-3 {activeSection === section.id ? 'bg-primary/10 text-primary border border-primary/20' : 'text-foreground hover:bg-muted'}"
            on:click={() => handleSectionClick(section)}
          >
            <div class="flex items-center justify-between w-full">
              <div class="flex items-center space-x-3">
                <!-- Expand/Collapse Icon -->
                {#if section.children && collapsible}
                  <Icon 
                    name={isExpanded(section.id) ? 'chevron-down' : 'chevron-right'} 
                    size={12} 
                    className="text-muted-foreground" 
                  />
                {:else}
                  <div class="w-3"></div>
                {/if}
                
                <!-- Section Icon -->
                <div class="w-5 h-5 flex items-center justify-center">
                  <Icon name={section.icon} size={16} />
                </div>
                
                <!-- Section Name -->
                <span class="text-sm font-medium">{section.name}</span>
              </div>
              
              <!-- Count and Badge -->
              <div class="flex items-center space-x-2">
                {#if section.badge}
                  <Badge variant={getBadgeVariant(section.badge)} class="text-xs px-1.5 py-0.5">
                    {section.badge}
                  </Badge>
                {/if}
                
                {#if showCounts && section.count !== undefined}
                  <span class="text-xs {getCountColor(section.count)} font-medium">
                    {section.count}
                  </span>
                {/if}
              </div>
            </div>
          </Button>
          
          <!-- Child Sections -->
          {#if section.children && isExpanded(section.id)}
            <div class="ml-6 space-y-1 border-l border-border pl-2">
              {#each section.children as child}
                <Button
                  variant="ghost"
                  size="sm"
                  class="w-full justify-start h-auto py-1.5 px-2 text-xs {activeSection === child.id ? 'bg-primary/10 text-primary' : 'text-muted-foreground hover:text-foreground hover:bg-muted/50'}"
                  on:click={() => dispatch('sectionSelect', { section: child })}
                >
                  <div class="flex items-center justify-between w-full">
                    <div class="flex items-center space-x-2">
                      <div class="w-4 h-4 flex items-center justify-center">
                        <Icon name={child.icon} size={12} />
                      </div>
                      <span>{child.name}</span>
                    </div>
                    
                    {#if showCounts && child.count !== undefined}
                      <span class="text-xs {getCountColor(child.count)}">
                        {child.count}
                      </span>
                    {/if}
                  </div>
                </Button>
              {/each}
            </div>
          {/if}
        </div>
      {/each}
    </nav>
  </div>

  <!-- Quick Actions Footer -->
  <div class="border-t border-border p-4 space-y-2">
    <div class="flex gap-2">
      <Button
        variant="ghost"
        size="icon"
        on:click={() => dispatch('quickAction', { action: 'refresh' })}
        class="h-8 w-8 hover:rotate-180 transition-transform duration-300"
        title="Vernieuw database data"
      >
        <Icon name="refresh-cw" size={16} />
      </Button>
      
      <Button
        variant="outline"
        size="sm"
        on:click={() => dispatch('quickAction', { action: 'sql' })}
        class="flex items-center space-x-2 justify-center flex-1"
      >
        <Icon name="code" size={12} />
        <span class="text-xs">SQL</span>
      </Button>
    </div>
    
    <Button
      variant="default"
      size="sm"
      on:click={() => dispatch('quickAction', { action: 'create' })}
      class="w-full flex items-center space-x-2 justify-center"
    >
      <Icon name="plus" size={12} />
      <span class="text-xs">New Table</span>
    </Button>
  </div>

  <!-- Database Info -->
  <div class="border-t border-border bg-muted/20 p-3">
    <div class="text-xs text-muted-foreground space-y-1">
      <div class="flex items-center justify-between">
        <span>Connected</span>
        <div class="flex items-center space-x-1">
          <div class="w-2 h-2 bg-green-500 rounded-full"></div>
          <span>Active</span>
        </div>
      </div>
      
      {#if showCounts}
        {@const totalItems = defaultSections.reduce((sum, section) => sum + (section.count || 0), 0)}
        <div class="flex items-center justify-between">
          <span>Total Objects</span>
          <span class="font-medium">{totalItems}</span>
        </div>
      {/if}
      
      <div class="flex items-center justify-between">
        <span>Schema</span>
        <span class="font-medium">public</span>
      </div>
    </div>
  </div>
</div>

<style>
  /* Custom scrollbar for navigation */
  nav::-webkit-scrollbar {
    width: 4px;
  }
  
  nav::-webkit-scrollbar-track {
    background: transparent;
  }
  
  nav::-webkit-scrollbar-thumb {
    background: hsl(var(--border));
    border-radius: 2px;
  }
  
  nav::-webkit-scrollbar-thumb:hover {
    background: hsl(var(--muted-foreground));
  }
</style>