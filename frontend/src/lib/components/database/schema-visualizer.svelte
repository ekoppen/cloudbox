<script lang="ts">
  import { onMount, createEventDispatcher } from 'svelte';
  import Card from '$lib/components/ui/card.svelte';
  import Button from '$lib/components/ui/button.svelte';
  import Badge from '$lib/components/ui/badge.svelte';
  import Icon from '$lib/components/ui/icon.svelte';

  const dispatch = createEventDispatcher();

  interface TableColumn {
    name: string;
    type: string;
    nullable: boolean;
    primary_key?: boolean;
    foreign_key?: {
      table: string;
      column: string;
    };
    default_value?: string;
  }

  interface TableSchema {
    name: string;
    columns: TableColumn[];
    row_count: number;
    size: string;
    position?: { x: number; y: number };
    relationships?: {
      to_table: string;
      from_column: string;
      to_column: string;
      type: 'one-to-one' | 'one-to-many' | 'many-to-many';
    }[];
  }

  export let tables: TableSchema[] = [];
  export let selectedTable: string | null = null;
  export let showRelationships: boolean = true;
  export let compact: boolean = false;

  let containerElement: HTMLElement;
  let svgElement: SVGElement;
  let isDragging = false;
  let draggedTable = '';
  let dragOffset = { x: 0, y: 0 };

  $: visibleTables = tables.filter(table => table.columns.length > 0);

  onMount(() => {
    // Auto-arrange tables if no positions are set
    if (tables.some(table => !table.position)) {
      autoArrangeTables();
    }
  });

  function autoArrangeTables() {
    const gridSize = 300;
    const padding = 50;
    let row = 0;
    let col = 0;
    const maxCols = Math.ceil(Math.sqrt(visibleTables.length));

    visibleTables.forEach((table, index) => {
      if (!table.position) {
        table.position = {
          x: padding + (col * gridSize),
          y: padding + (row * gridSize)
        };
      }
      
      col++;
      if (col >= maxCols) {
        col = 0;
        row++;
      }
    });
    
    tables = [...tables];
  }

  function getTypeColor(type: string): string {
    const lowerType = type.toLowerCase();
    if (lowerType.includes('int') || lowerType.includes('serial') || lowerType.includes('number')) {
      return 'bg-blue-100 dark:bg-blue-900/30 text-blue-800 dark:text-blue-200 border-blue-200 dark:border-blue-800';
    }
    if (lowerType.includes('varchar') || lowerType.includes('text') || lowerType.includes('string')) {
      return 'bg-green-100 dark:bg-green-900/30 text-green-800 dark:text-green-200 border-green-200 dark:border-green-800';
    }
    if (lowerType.includes('boolean') || lowerType.includes('bool')) {
      return 'bg-purple-100 dark:bg-purple-900/30 text-purple-800 dark:text-purple-200 border-purple-200 dark:border-purple-800';
    }
    if (lowerType.includes('timestamp') || lowerType.includes('date') || lowerType.includes('time')) {
      return 'bg-yellow-100 dark:bg-yellow-900/30 text-yellow-800 dark:text-yellow-200 border-yellow-200 dark:border-yellow-800';
    }
    if (lowerType.includes('json') || lowerType.includes('object')) {
      return 'bg-orange-100 dark:bg-orange-900/30 text-orange-800 dark:text-orange-200 border-orange-200 dark:border-orange-800';
    }
    return 'bg-gray-100 dark:bg-gray-900/30 text-gray-800 dark:text-gray-200 border-gray-200 dark:border-gray-800';
  }

  function handleTableClick(tableName: string) {
    selectedTable = tableName;
    dispatch('tableSelect', { tableName });
  }

  function startDrag(event: MouseEvent, tableName: string) {
    isDragging = true;
    draggedTable = tableName;
    
    const table = tables.find(t => t.name === tableName);
    if (table && table.position) {
      dragOffset = {
        x: event.clientX - table.position.x,
        y: event.clientY - table.position.y
      };
    }
    
    event.preventDefault();
  }

  function handleMouseMove(event: MouseEvent) {
    if (!isDragging || !draggedTable) return;
    
    const table = tables.find(t => t.name === draggedTable);
    if (table) {
      table.position = {
        x: event.clientX - dragOffset.x,
        y: event.clientY - dragOffset.y
      };
      tables = [...tables];
    }
  }

  function handleMouseUp() {
    isDragging = false;
    draggedTable = '';
  }

  function getRelationshipPath(fromTable: TableSchema, toTable: TableSchema, fromColumn: string, toColumn: string): string {
    if (!fromTable.position || !toTable.position) return '';
    
    const fromX = fromTable.position.x + 200; // Table width
    const fromY = fromTable.position.y + 60 + (fromTable.columns.findIndex(col => col.name === fromColumn) * 28);
    const toX = toTable.position.x;
    const toY = toTable.position.y + 60 + (toTable.columns.findIndex(col => col.name === toColumn) * 28);
    
    // Simple curved path
    const midX = fromX + (toX - fromX) / 2;
    return `M ${fromX} ${fromY} C ${midX} ${fromY}, ${midX} ${toY}, ${toX} ${toY}`;
  }

  function zoomToFit() {
    if (!containerElement || visibleTables.length === 0) return;
    
    // Calculate bounds of all tables
    let minX = Infinity, minY = Infinity, maxX = -Infinity, maxY = -Infinity;
    
    visibleTables.forEach(table => {
      if (table.position) {
        minX = Math.min(minX, table.position.x);
        minY = Math.min(minY, table.position.y);
        maxX = Math.max(maxX, table.position.x + 220); // Table width + margin
        maxY = Math.max(maxY, table.position.y + (table.columns.length * 28) + 80); // Header + columns
      }
    });
    
    const width = maxX - minX + 100; // Add padding
    const height = maxY - minY + 100;
    
    // Center the view
    containerElement.scrollTo({
      left: minX - 50,
      top: minY - 50,
      behavior: 'smooth'
    });
  }
</script>

<div class="h-full flex flex-col">
  <!-- Toolbar -->
  <div class="flex items-center justify-between p-4 border-b border-border bg-card">
    <div class="flex items-center space-x-4">
      <div class="flex items-center space-x-2">
        <Icon name="database" size={16} className="text-primary" />
        <h3 class="text-sm font-semibold text-foreground">Schema Visualizer</h3>
      </div>
      
      {#if visibleTables.length > 0}
        <Badge variant="secondary" class="text-xs">
          {visibleTables.length} {visibleTables.length === 1 ? 'table' : 'tables'}
        </Badge>
      {/if}
    </div>
    
    <div class="flex items-center space-x-2">
      <Button
        variant="outline"
        size="sm"
        on:click={() => showRelationships = !showRelationships}
        class="flex items-center space-x-2"
      >
        <Icon name={showRelationships ? 'eye-off' : 'eye'} size={14} />
        <span>{showRelationships ? 'Hide' : 'Show'} Relations</span>
      </Button>
      
      <Button
        variant="outline"
        size="sm"
        on:click={() => compact = !compact}
        class="flex items-center space-x-2"
      >
        <Icon name={compact ? 'maximize' : 'minimize'} size={14} />
        <span>{compact ? 'Expand' : 'Compact'}</span>
      </Button>
      
      <Button
        variant="outline"
        size="sm"
        on:click={autoArrangeTables}
        class="flex items-center space-x-2"
      >
        <Icon name="grid" size={14} />
        <span>Auto Layout</span>
      </Button>
      
      <Button
        variant="outline"
        size="sm"
        on:click={zoomToFit}
        class="flex items-center space-x-2"
      >
        <Icon name="zoom-in" size={14} />
        <span>Fit</span>
      </Button>
    </div>
  </div>

  <!-- Schema Canvas -->
  <div 
    class="flex-1 overflow-auto bg-muted/20 relative"
    role="application"
    aria-label="Database schema visualization"
    bind:this={containerElement}
  >
    {#if visibleTables.length === 0}
      <div class="absolute inset-0 flex items-center justify-center">
        <div class="text-center">
          <div class="w-16 h-16 bg-muted rounded-lg flex items-center justify-center mx-auto mb-4">
            <Icon name="database" size={32} className="text-muted-foreground" />
          </div>
          <h3 class="text-lg font-medium text-foreground mb-2">No Tables Found</h3>
          <p class="text-muted-foreground">Import your database schema to visualize table relationships</p>
        </div>
      </div>
    {:else}
      <!-- SVG for relationship lines -->
      {#if showRelationships}
        <svg
          bind:this={svgElement}
          class="absolute inset-0 pointer-events-none"
          style="width: 100%; height: 100%; z-index: 1;"
        >
          {#each visibleTables as table}
            {#if table.relationships}
              {#each table.relationships as rel}
                {@const toTable = visibleTables.find(t => t.name === rel.to_table)}
                {#if toTable}
                  <path
                    d={getRelationshipPath(table, toTable, rel.from_column, rel.to_column)}
                    stroke="hsl(var(--primary))"
                    stroke-width="2"
                    fill="none"
                    stroke-opacity="0.6"
                    stroke-dasharray={rel.type === 'one-to-many' ? '5,5' : ''}
                  />
                  <!-- Relationship type indicator -->
                  <circle
                    cx={toTable.position?.x}
                    cy={toTable.position ? toTable.position.y + 60 + (toTable.columns.findIndex(col => col.name === rel.to_column) * 28) : 0}
                    r="4"
                    fill="hsl(var(--primary))"
                    opacity="0.8"
                  />
                {/if}
              {/each}
            {/if}
          {/each}
        </svg>
      {/if}

      <!-- Table Cards -->
      {#each visibleTables as table}
        <div
          class="absolute z-10"
          style="left: {table.position?.x || 0}px; top: {table.position?.y || 0}px;"
        >
          <Card class="w-56 shadow-lg border-2 cursor-move {selectedTable === table.name ? 'border-primary ring-2 ring-primary/20' : ''}"
                on:mousedown={(e) => startDrag(e, table.name)}
                on:click={() => handleTableClick(table.name)}
                on:mousemove={handleMouseMove}
                on:mouseup={handleMouseUp}
                on:mouseleave={handleMouseUp}
          >
            <!-- Table Header -->
            <div class="px-4 py-3 border-b border-border bg-card">
              <div class="flex items-center justify-between">
                <div class="flex items-center space-x-2">
                  <Icon name="table" size={14} className="text-primary" />
                  <h4 class="text-sm font-semibold text-foreground">{table.name}</h4>
                </div>
                <div class="flex space-x-1">
                  <Badge variant="outline" class="text-xs px-1.5 py-0.5">
                    {table.row_count}
                  </Badge>
                </div>
              </div>
              <p class="text-xs text-muted-foreground mt-1">{table.size}</p>
            </div>

            <!-- Columns -->
            <div class="max-h-64 overflow-y-auto">
              {#each table.columns.slice(0, compact ? 5 : table.columns.length) as column}
                <div class="px-4 py-2 border-b border-border/50 hover:bg-muted/30 transition-colors">
                  <div class="flex items-center justify-between">
                    <div class="flex items-center space-x-2 min-w-0 flex-1">
                      {#if column.primary_key}
                        <Icon name="key" size={12} className="text-yellow-600 flex-shrink-0" />
                      {:else if column.foreign_key}
                        <Icon name="link" size={12} className="text-blue-600 flex-shrink-0" />
                      {:else}
                        <div class="w-3"></div>
                      {/if}
                      <span class="text-xs font-medium text-foreground truncate">{column.name}</span>
                    </div>
                    <div class="flex items-center space-x-2">
                      {#if !column.nullable}
                        <div class="w-1.5 h-1.5 bg-red-500 rounded-full" title="NOT NULL"></div>
                      {/if}
                    </div>
                  </div>
                  <div class="mt-1">
                    <Badge variant="outline" class="text-xs px-1.5 py-0.5 {getTypeColor(column.type)}">
                      {column.type}
                    </Badge>
                  </div>
                  {#if column.foreign_key}
                    <div class="text-xs text-muted-foreground mt-1">
                      → {column.foreign_key.table}.{column.foreign_key.column}
                    </div>
                  {/if}
                </div>
              {/each}
              
              {#if compact && table.columns.length > 5}
                <div class="px-4 py-2 text-xs text-muted-foreground text-center border-b border-border/50">
                  +{table.columns.length - 5} more columns
                </div>
              {/if}
            </div>
          </Card>
        </div>
      {/each}
    {/if}
  </div>
</div>

