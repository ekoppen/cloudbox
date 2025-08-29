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

  // State for selected relationship
  let selectedRelationship: { from: string, to: string, fromColumn: string, toColumn: string, type?: string } | null = null;
  let showRelationshipDetails = false;

  let containerElement: HTMLElement;
  let svgElement: SVGElement;
  let isDragging = false;
  let draggedTable = '';
  let dragOffset = { x: 0, y: 0 };

  $: visibleTables = tables.filter(table => {
    const hasColumns = table.columns && table.columns.length > 0;
    console.log(`Filtering table ${table.name}: columns=${table.columns?.length || 0}, hasColumns=${hasColumns}`, table.columns);
    return hasColumns;
  });

  // Temporary: Force show tables even without proper columns for debugging
  $: debugTables = tables.length > 0 ? tables.map(table => ({
    ...table,
    columns: table.columns || [
      { name: 'id', type: 'STRING', nullable: false, primary_key: true },
      { name: 'created_at', type: 'TIMESTAMP', nullable: false }
    ]
  })) : [];

  // Show debug tables if no visible tables but we have data
  $: actualVisibleTables = visibleTables.length > 0 ? visibleTables : debugTables;

  // Debug reactive statements
  $: {
    console.log('SchemaVisualizer: tables prop updated:', tables.length, 'tables');
    console.log('SchemaVisualizer: visibleTables:', visibleTables.length, 'visible');
    console.log('SchemaVisualizer: actualVisibleTables:', actualVisibleTables.length, 'actual visible');
    console.log('SchemaVisualizer: tables structure:', tables.map(t => ({ 
      name: t.name, 
      columns: t.columns?.length || 0,
      hasPosition: !!t.position 
    })));
    if (visibleTables.length === 0 && tables.length > 0) {
      console.log('WARNING: No visible tables but tables exist. Column lengths:', 
        tables.map(t => `${t.name}: ${t.columns?.length || 0} columns`)
      );
      console.log('Using debug tables instead');
    }
  }

  onMount(() => {
    // Auto-arrange tables if no positions are set
    if (tables.some(table => !table.position)) {
      autoArrangeTables();
    }

    // Add global mouse event listeners for dragging
    const handleGlobalMouseMove = (event: MouseEvent) => {
      handleMouseMove(event);
    };

    const handleGlobalMouseUp = () => {
      handleMouseUp();
    };

    document.addEventListener('mousemove', handleGlobalMouseMove);
    document.addEventListener('mouseup', handleGlobalMouseUp);

    return () => {
      document.removeEventListener('mousemove', handleGlobalMouseMove);
      document.removeEventListener('mouseup', handleGlobalMouseUp);
    };
  });

  // Watch for actualVisibleTables changes and arrange if needed
  $: if (actualVisibleTables.length > 0 && actualVisibleTables.some(table => !table.position)) {
    console.log('Arranging tables because actualVisibleTables changed and some have no position');
    autoArrangeActualTables();
  }

  function autoArrangeTables() {
    console.log('AutoArrangeTables: Starting arrangement for', visibleTables.length, 'visible tables');
    const gridSize = 300;
    const padding = 50;
    let row = 0;
    let col = 0;
    const maxCols = Math.ceil(Math.sqrt(visibleTables.length));
    
    console.log('AutoArrangeTables: Grid config - size:', gridSize, 'padding:', padding, 'maxCols:', maxCols);

    visibleTables.forEach((table, index) => {
      if (!table.position) {
        table.position = {
          x: padding + (col * gridSize),
          y: padding + (row * gridSize)
        };
        console.log(`AutoArrangeTables: Set position for ${table.name}:`, table.position);
      } else {
        console.log(`AutoArrangeTables: ${table.name} already has position:`, table.position);
      }
      
      col++;
      if (col >= maxCols) {
        col = 0;
        row++;
      }
    });
    
    console.log('AutoArrangeTables: Final table positions:', visibleTables.map(t => ({ name: t.name, position: t.position })));
    tables = [...tables];
    console.log('AutoArrangeTables: Triggered reactivity update');
  }

  function autoArrangeActualTables() {
    console.log('AutoArrangeActualTables: Starting arrangement for', actualVisibleTables.length, 'actual visible tables');
    const gridSize = 300;
    const padding = 50;
    let row = 0;
    let col = 0;
    const maxCols = Math.ceil(Math.sqrt(actualVisibleTables.length));
    
    console.log('AutoArrangeActualTables: Grid config - size:', gridSize, 'padding:', padding, 'maxCols:', maxCols);

    actualVisibleTables.forEach((table, index) => {
      if (!table.position) {
        table.position = {
          x: padding + (col * gridSize),
          y: padding + (row * gridSize)
        };
        console.log(`AutoArrangeActualTables: Set position for ${table.name}:`, table.position);
      } else {
        console.log(`AutoArrangeActualTables: ${table.name} already has position:`, table.position);
      }
      
      col++;
      if (col >= maxCols) {
        col = 0;
        row++;
      }
    });
    
    console.log('AutoArrangeActualTables: Final table positions:', actualVisibleTables.map(t => ({ name: t.name, position: t.position })));
    // Force reactivity update
    tables = [...tables];
    console.log('AutoArrangeActualTables: Triggered reactivity update');
  }

  function getTypeColor(type: string): string {
    const lowerType = type.toLowerCase();
    if (lowerType.includes('int') || lowerType.includes('serial') || lowerType.includes('number')) {
      return 'bg-blue-50 dark:bg-blue-950/30 text-blue-700 dark:text-blue-300 border-blue-200 dark:border-blue-800';
    }
    if (lowerType.includes('varchar') || lowerType.includes('text') || lowerType.includes('string')) {
      return 'bg-green-50 dark:bg-green-950/30 text-green-700 dark:text-green-300 border-green-200 dark:border-green-800';
    }
    if (lowerType.includes('boolean') || lowerType.includes('bool')) {
      return 'bg-purple-50 dark:bg-purple-950/30 text-purple-700 dark:text-purple-300 border-purple-200 dark:border-purple-800';
    }
    if (lowerType.includes('timestamp') || lowerType.includes('date') || lowerType.includes('time')) {
      return 'bg-amber-50 dark:bg-amber-950/30 text-amber-700 dark:text-amber-300 border-amber-200 dark:border-amber-800';
    }
    if (lowerType.includes('json') || lowerType.includes('object')) {
      return 'bg-orange-50 dark:bg-orange-950/30 text-orange-700 dark:text-orange-300 border-orange-200 dark:border-orange-800';
    }
    if (lowerType.includes('array')) {
      return 'bg-cyan-50 dark:bg-cyan-950/30 text-cyan-700 dark:text-cyan-300 border-cyan-200 dark:border-cyan-800';
    }
    return 'bg-gray-50 dark:bg-gray-950/30 text-gray-700 dark:text-gray-300 border-gray-200 dark:border-gray-800';
  }

  // Dark mode detection and SVG colors
  $: isDarkMode = typeof window !== 'undefined' && 
    (document.documentElement.classList.contains('dark') || 
     document.documentElement.classList.contains('cloudbox-dark'));

  // Reactive colors for SVG elements
  $: svgColors = {
    relationship: {
      default: isDarkMode ? '#6b7280' : '#94a3b8', // gray-500 : slate-400
      selected: '#f59e0b' // amber-500 (same for both modes)
    },
    label: {
      background: isDarkMode ? '#0891b2' : '#22d3ee', // cyan-600 : cyan-400
      backgroundSelected: '#f59e0b', // amber-500
      border: isDarkMode ? '#0e7490' : '#0891b2', // cyan-700 : cyan-600
      borderSelected: '#d97706', // amber-600
      text: isDarkMode ? '#f0f9ff' : '#0c4a6e', // sky-50 : sky-900
      textSelected: '#fff' // white
    }
  };

  function handleTableClick(tableName: string) {
    selectedTable = tableName;
    selectedRelationship = null; // Clear relationship selection when selecting table
    dispatch('tableSelect', { tableName });
  }

  function handleRelationshipClick(fromTable: string, toTable: string, fromColumn: string, toColumn: string, type?: string) {
    selectedRelationship = { from: fromTable, to: toTable, fromColumn, toColumn, type };
    selectedTable = null; // Clear table selection when selecting relationship
    showRelationshipDetails = true;
    dispatch('relationshipSelect', { fromTable, toTable, fromColumn, toColumn, type });
    console.log('Selected relationship:', fromTable, fromColumn, '->', toTable, toColumn, 'type:', type);
  }

  function startDrag(event: MouseEvent, tableName: string) {
    console.log('Starting drag for table:', tableName);
    isDragging = true;
    draggedTable = tableName;
    
    const table = actualVisibleTables.find(t => t.name === tableName);
    if (table && table.position) {
      const rect = containerElement.getBoundingClientRect();
      dragOffset = {
        x: event.clientX - rect.left - table.position.x,
        y: event.clientY - rect.top - table.position.y
      };
      console.log('Drag offset calculated:', dragOffset);
    }
    
    event.preventDefault();
    event.stopPropagation();
  }

  function handleMouseMove(event: MouseEvent) {
    if (!isDragging || !draggedTable || !containerElement) return;
    
    console.log('Mouse move during drag');
    const rect = containerElement.getBoundingClientRect();
    const table = actualVisibleTables.find(t => t.name === draggedTable);
    if (table) {
      table.position = {
        x: Math.max(10, event.clientX - rect.left - dragOffset.x),
        y: Math.max(10, event.clientY - rect.top - dragOffset.y)
      };
      console.log('Updated position for', draggedTable, ':', table.position);
      tables = [...tables];
    }
  }

  function handleMouseUp() {
    if (isDragging) {
      console.log('Ending drag for table:', draggedTable);
    }
    isDragging = false;
    draggedTable = '';
  }

  function getRelationshipPath(fromTable: TableSchema, toTable: TableSchema, fromColumn: string, toColumn: string): string {
    if (!fromTable.position || !toTable.position) return '';
    
    const tableWidth = 256; // Updated table width (w-64 = 16rem = 256px)
    const headerHeight = 40; // Updated header height
    const rowHeight = 32; // Updated row height for better spacing
    
    // Find column positions
    const fromColumnIndex = fromTable.columns.findIndex(col => col.name === fromColumn);
    const toColumnIndex = toTable.columns.findIndex(col => col.name === toColumn);
    
    // Calculate exact Y positions based on column index
    const fromY = fromTable.position.y + headerHeight + (fromColumnIndex * rowHeight) + (rowHeight / 2);
    const toY = toTable.position.y + headerHeight + (toColumnIndex * rowHeight) + (rowHeight / 2);
    
    // Determine connection side based on table positions
    const fromX = fromTable.position.x + tableWidth;
    const toX = toTable.position.x;
    
    // Create smooth curved path
    const dx = toX - fromX;
    const dy = toY - fromY;
    const distance = Math.sqrt(dx * dx + dy * dy);
    const curvature = Math.min(80, distance / 4);
    
    // Use cubic bezier for smoother curves
    const controlX1 = fromX + curvature;
    const controlX2 = toX - curvature;
    
    return `M ${fromX} ${fromY} C ${controlX1} ${fromY}, ${controlX2} ${toY}, ${toX} ${toY}`;
  }

  function zoomToFit() {
    if (!containerElement || actualVisibleTables.length === 0) return;
    
    // Calculate bounds of all tables
    let minX = Infinity, minY = Infinity, maxX = -Infinity, maxY = -Infinity;
    
    actualVisibleTables.forEach(table => {
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
  <div class="flex items-center justify-between p-4 border-b border-gray-200 dark:border-gray-700 bg-white dark:bg-gray-800">
    <div class="flex items-center space-x-4">
      <div class="flex items-center space-x-2">
        <Icon name="database" size={16} className="text-blue-600 dark:text-blue-400" />
        <h3 class="text-sm font-semibold text-gray-900 dark:text-gray-100">Schema Visualizer</h3>
      </div>
      
      {#if actualVisibleTables.length > 0}
        <div class="px-2 py-1 bg-gray-100 dark:bg-gray-700 text-gray-700 dark:text-gray-300 rounded text-xs">
          {actualVisibleTables.length} {actualVisibleTables.length === 1 ? 'table' : 'tables'}
        </div>
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
    class="relative bg-gray-50 dark:bg-gray-900"
    style="height: 800px; background-image: radial-gradient(circle, #e2e8f0 1px, transparent 1px); background-size: 20px 20px;"
    role="application"
    aria-label="Database schema visualization"
    bind:this={containerElement}
  >
    {#if actualVisibleTables.length === 0}
      <div class="absolute inset-0 flex items-center justify-center">
        <div class="text-center">
          <div class="w-16 h-16 bg-gray-100 dark:bg-gray-800 rounded-lg flex items-center justify-center mx-auto mb-4">
            <Icon name="database" size={32} className="text-gray-500 dark:text-gray-400" />
          </div>
          <h3 class="text-lg font-medium text-gray-900 dark:text-gray-100 mb-2">No Tables Found</h3>
          <p class="text-gray-600 dark:text-gray-400">
            {tables.length === 0 ? 'Import your database schema to visualize table relationships' : 
             `${tables.length} tables loaded but none have columns to display`}
          </p>
          {#if tables.length > 0}
            <div class="mt-4 p-4 bg-white dark:bg-gray-800 border border-gray-200 dark:border-gray-700 rounded-lg text-left">
              <h4 class="font-medium mb-2 text-gray-900 dark:text-gray-100">Debug Info:</h4>
              <pre class="text-sm text-gray-600 dark:text-gray-400 whitespace-pre-wrap">{JSON.stringify(tables.map(t => ({ 
                name: t.name, 
                columns: t.columns?.length || 0,
                columnsData: t.columns?.slice(0, 2)
              })), null, 2)}</pre>
            </div>
          {/if}
        </div>
      </div>
    {:else}
      {console.log('Rendering', actualVisibleTables.length, 'visible tables')}
      <!-- SVG for relationship lines -->
      {#if showRelationships}
        <svg
          bind:this={svgElement}
          class="absolute inset-0 pointer-events-none"
          style="width: 100%; height: 100%; z-index: 1;"
        >
          <!-- Arrowhead marker definitions -->
          <defs>
            <marker id="arrowhead" markerWidth="8" markerHeight="8" refX="8" refY="4" orient="auto">
              <circle cx="4" cy="4" r="2" fill="#94a3b8" class="dark:fill-gray-400" />
            </marker>
            <marker id="arrowhead-dark" markerWidth="8" markerHeight="8" refX="8" refY="4" orient="auto">
              <circle cx="4" cy="4" r="2" fill="#9ca3af" />
            </marker>
            <marker id="arrowhead-many" markerWidth="12" markerHeight="8" refX="11" refY="4" orient="auto">
              <line x1="2" y1="2" x2="8" y2="4" stroke="#94a3b8" stroke-width="1.5" class="dark:stroke-gray-400" />
              <line x1="2" y1="6" x2="8" y2="4" stroke="#94a3b8" stroke-width="1.5" class="dark:stroke-gray-400" />
              <circle cx="10" cy="4" r="2" fill="#94a3b8" class="dark:fill-gray-400" />
            </marker>
            <marker id="arrowhead-many-dark" markerWidth="12" markerHeight="8" refX="11" refY="4" orient="auto">
              <line x1="2" y1="2" x2="8" y2="4" stroke="#9ca3af" stroke-width="1.5" />
              <line x1="2" y1="6" x2="8" y2="4" stroke="#9ca3af" stroke-width="1.5" />
              <circle cx="10" cy="4" r="2" fill="#9ca3af" />
            </marker>
          </defs>
          
          {#each actualVisibleTables as table}
            {#if table.relationships}
              {#each table.relationships as rel}
                {@const toTable = actualVisibleTables.find(t => t.name === rel.to_table)}
                {@const isSelected = selectedRelationship && selectedRelationship.from === table.name && selectedRelationship.to === rel.to_table && selectedRelationship.fromColumn === rel.from_column && selectedRelationship.toColumn === rel.to_column}
                {#if toTable}
                  <!-- Invisible wider path for easier clicking -->
                  <path
                    d={getRelationshipPath(table, toTable, rel.from_column, rel.to_column)}
                    stroke="transparent"
                    stroke-width="12"
                    fill="none"
                    style="cursor: pointer;"
                    on:click={() => handleRelationshipClick(table.name, rel.to_table, rel.from_column, rel.to_column, rel.type)}
                  />
                  <!-- Visible relationship line -->
                  <path
                    d={getRelationshipPath(table, toTable, rel.from_column, rel.to_column)}
                    stroke={isSelected ? svgColors.relationship.selected : svgColors.relationship.default}
                    stroke-width={isSelected ? "3" : "2"}
                    fill="none"
                    stroke-opacity="0.8"
                    stroke-dasharray={rel.type === 'many-to-many' ? '5,3' : ''}
                    marker-end={rel.type === 'one-to-many' || rel.type === 'many-to-one' ? `url(#arrowhead-many${isDarkMode ? '-dark' : ''})` : `url(#arrowhead${isDarkMode ? '-dark' : ''})`}
                    style="cursor: pointer; pointer-events: none;"
                  />
                  
                  <!-- Relationship label - simplified calculation -->
                  {#if table.position && toTable.position}
                    {@const midX = (table.position.x + 220 + toTable.position.x) / 2}
                    {@const midY = (table.position.y + 80 + toTable.position.y + 80) / 2}
                    <rect
                      x={midX - 18}
                      y={midY - 10}
                      width="36"
                      height="20"
                      rx="10"
                      fill={isSelected ? svgColors.label.backgroundSelected : svgColors.label.background}
                      stroke={isSelected ? svgColors.label.borderSelected : svgColors.label.border}
                      stroke-width="2"
                      opacity="0.95"
                      style="cursor: pointer;"
                      on:click={() => handleRelationshipClick(table.name, rel.to_table, rel.from_column, rel.to_column, rel.type)}
                    />
                    <text
                      x={midX}
                      y={midY + 3}
                      text-anchor="middle"
                      class="text-xs font-bold"
                      fill={isSelected ? svgColors.label.textSelected : svgColors.label.text}
                      style="cursor: pointer; pointer-events: none;"
                    >
                      {rel.type === 'one-to-one' ? '1:1' : rel.type === 'one-to-many' ? '1:N' : 'N:N'}
                    </text>
                  {/if}
                {/if}
              {/each}
            {/if}
          {/each}
        </svg>
      {/if}

      <!-- Table Cards -->
      {#each actualVisibleTables as table}
        <!-- Debug: Log each table being rendered -->
        {console.log('Rendering table:', table.name, 'at position:', table.position, 'with', table.columns.length, 'columns')}
        <div
          class="absolute z-10"
          style="left: {table.position?.x || 0}px; top: {table.position?.y || 0}px;"
        >
          <div class="w-64 cursor-move {selectedTable === table.name ? 'ring-2 ring-blue-500' : ''} bg-white dark:bg-gray-800 border border-gray-200 dark:border-gray-700 rounded-lg shadow-sm dark:shadow-gray-900/10"
                on:mousedown={(e) => startDrag(e, table.name)}
                on:click={() => handleTableClick(table.name)}
          >
            <!-- Table Header -->
            <div class="px-4 py-2 border-b border-gray-200 dark:border-gray-700 bg-gray-50 dark:bg-gray-700 rounded-t-lg" style="height: 40px; display: flex; align-items: center;">
              <div class="flex items-center justify-between w-full">
                <div class="flex items-center space-x-2">
                  <Icon name="table" size={12} className="text-gray-600 dark:text-gray-400" />
                  <h4 class="text-sm font-semibold text-gray-800 dark:text-gray-200">{table.name}</h4>
                </div>
                <span class="text-xs text-gray-500 dark:text-gray-400">
                  {table.row_count} rows
                </span>
              </div>
            </div>

            <!-- Columns -->
            <div class="max-h-64 overflow-y-auto">
              {#each table.columns.slice(0, compact ? 5 : table.columns.length) as column}
                {@const isRelatedField = selectedRelationship && ((selectedRelationship.from === table.name && selectedRelationship.fromColumn === column.name) || (selectedRelationship.to === table.name && selectedRelationship.toColumn === column.name))}
                {@const hasRelationship = column.foreign_key || table.relationships?.some(rel => rel.from_column === column.name) || actualVisibleTables.some(t => t.relationships?.some(rel => rel.to_table === table.name && rel.to_column === column.name))}
                {@const relationshipType = table.relationships?.find(rel => rel.from_column === column.name)?.type || actualVisibleTables.find(t => t.relationships?.some(rel => rel.to_table === table.name && rel.to_column === column.name))?.relationships?.find(rel => rel.to_table === table.name && rel.to_column === column.name)?.type}
                <div class="px-4 py-2 border-b border-gray-200 dark:border-gray-700 hover:bg-gray-50 dark:hover:bg-gray-700 transition-colors {isRelatedField ? 'bg-amber-50 dark:bg-amber-900/20 border-amber-200 dark:border-amber-800' : ''}" style="min-height: 32px;">
                  <div class="flex items-center justify-between">
                    <div class="flex items-center space-x-2 min-w-0 flex-1">
                      {#if column.primary_key}
                        <Icon name="key" size={12} className="text-amber-600 dark:text-amber-400 flex-shrink-0" />
                      {:else if column.foreign_key}
                        <div class="relative">
                          <Icon name="link" size={12} className="text-blue-600 dark:text-blue-400 flex-shrink-0" />
                          {#if hasRelationship}
                            <div class="absolute -top-1 -right-1 w-2 h-2 bg-cyan-400 dark:bg-cyan-300 rounded-full"></div>
                          {/if}
                        </div>
                      {:else if hasRelationship}
                        <div class="relative">
                          <Icon name="share-2" size={12} className="text-cyan-600 dark:text-cyan-400 flex-shrink-0" />
                          <div class="absolute -top-1 -right-1 w-2 h-2 bg-cyan-400 dark:bg-cyan-300 rounded-full"></div>
                          {#if relationshipType === 'one-to-many'}
                            <div class="absolute -bottom-2 -right-2 text-xs font-bold text-cyan-700 dark:text-cyan-300" title="One-to-Many">1:N</div>
                          {:else if relationshipType === 'many-to-many'}
                            <div class="absolute -bottom-2 -right-2 text-xs font-bold text-purple-700 dark:text-purple-300" title="Many-to-Many">N:N</div>
                          {:else if relationshipType === 'one-to-one'}
                            <div class="absolute -bottom-2 -right-2 text-xs font-bold text-green-700 dark:text-green-300" title="One-to-One">1:1</div>
                          {/if}
                        </div>
                      {:else}
                        <div class="w-3"></div>
                      {/if}
                      <span class="text-xs font-semibold text-gray-800 dark:text-gray-200 truncate {isRelatedField ? 'text-amber-800 dark:text-amber-200' : ''}">{column.name}</span>
                    </div>
                    <div class="flex items-center space-x-2">
                      {#if !column.nullable}
                        <div class="w-1.5 h-1.5 bg-red-500 dark:bg-red-400 rounded-full" title="NOT NULL"></div>
                      {/if}
                      {#if isRelatedField}
                        <div class="w-2 h-2 bg-amber-500 dark:bg-amber-400 rounded-full animate-pulse" title="Related field"></div>
                      {/if}
                    </div>
                  </div>
                  <div class="mt-1">
                    <span class="text-xs px-2 py-1 rounded {getTypeColor(column.type)} font-mono">
                      {column.type}
                    </span>
                  </div>
                  {#if column.foreign_key}
                    <div class="text-xs text-blue-600 dark:text-blue-400 mt-1 font-medium">
                      → {column.foreign_key.table}.{column.foreign_key.column}
                    </div>
                  {/if}
                </div>
              {/each}
              
              {#if compact && table.columns.length > 5}
                <div class="px-4 py-2 text-xs text-gray-500 dark:text-gray-400 text-center border-b border-gray-200 dark:border-gray-700">
                  +{table.columns.length - 5} more columns
                </div>
              {/if}
            </div>
          </div>
        </div>
      {/each}
    {/if}
    
    <!-- Relationship Details Panel -->
    {#if showRelationshipDetails && selectedRelationship}
      <div class="absolute top-4 right-4 z-20 w-80 bg-white/95 dark:bg-gray-800/95 border-2 border-blue-500 dark:border-blue-400 rounded-lg shadow-xl backdrop-blur-sm">
        <div class="px-4 py-3 bg-blue-600 dark:bg-blue-700 text-white rounded-t-lg">
          <div class="flex items-center justify-between">
            <div class="flex items-center space-x-2">
              <Icon name="link" size={16} className="text-white" />
              <h4 class="font-semibold text-sm">Relationship Details</h4>
            </div>
            <button 
              on:click={() => { showRelationshipDetails = false; selectedRelationship = null; }}
              class="text-white hover:text-gray-200 transition-colors"
            >
              <Icon name="x" size={16} />
            </button>
          </div>
        </div>
        
        <div class="p-4 space-y-4">
          <!-- Connection Info -->
          <div class="space-y-3">
            <div class="flex items-center space-x-2">
              <div class="w-3 h-3 bg-green-500 dark:bg-green-400 rounded-full"></div>
              <span class="text-sm font-semibold text-gray-800 dark:text-gray-200">Source</span>
            </div>
            <div class="ml-5 p-3 bg-green-50 dark:bg-green-950/30 border border-green-200 dark:border-green-800 rounded">
              <div class="flex items-center space-x-2 mb-1">
                <Icon name="table" size={14} className="text-green-600 dark:text-green-400" />
                <span class="font-semibold text-green-800 dark:text-green-200">{selectedRelationship.from}</span>
              </div>
              <div class="flex items-center space-x-2">
                <Icon name="columns" size={12} className="text-green-600 dark:text-green-400" />
                <span class="text-sm font-mono bg-white dark:bg-gray-700 text-gray-800 dark:text-gray-200 px-2 py-1 border border-gray-300 dark:border-gray-600 rounded">{selectedRelationship.fromColumn}</span>
              </div>
            </div>
            
            <div class="flex items-center justify-center">
              <div class="flex items-center space-x-2 px-3 py-1 bg-blue-100 dark:bg-blue-950/30 border border-blue-300 dark:border-blue-700 rounded-full">
                <Icon name="arrow-right" size={14} className="text-blue-600 dark:text-blue-400" />
                <span class="text-xs font-bold text-blue-800 dark:text-blue-200">
                  {selectedRelationship.type === 'one-to-one' ? '1:1' : 
                   selectedRelationship.type === 'one-to-many' ? '1:N' : 
                   selectedRelationship.type === 'many-to-many' ? 'N:N' : 'REL'}
                </span>
                <Icon name="arrow-right" size={14} className="text-blue-600 dark:text-blue-400" />
              </div>
            </div>
            
            <div class="flex items-center space-x-2">
              <div class="w-3 h-3 bg-blue-500 dark:bg-blue-400 rounded-full"></div>
              <span class="text-sm font-semibold text-gray-800 dark:text-gray-200">Target</span>
            </div>
            <div class="ml-5 p-3 bg-blue-50 dark:bg-blue-950/30 border border-blue-200 dark:border-blue-800 rounded">
              <div class="flex items-center space-x-2 mb-1">
                <Icon name="table" size={14} className="text-blue-600 dark:text-blue-400" />
                <span class="font-semibold text-blue-800 dark:text-blue-200">{selectedRelationship.to}</span>
              </div>
              <div class="flex items-center space-x-2">
                <Icon name="columns" size={12} className="text-blue-600 dark:text-blue-400" />
                <span class="text-sm font-mono bg-white dark:bg-gray-700 text-gray-800 dark:text-gray-200 px-2 py-1 border border-gray-300 dark:border-gray-600 rounded">{selectedRelationship.toColumn}</span>
              </div>
            </div>
          </div>
          
          <!-- Relationship Type Description -->
          <div class="p-3 bg-gray-50 dark:bg-gray-800 border border-gray-200 dark:border-gray-700 rounded">
            <div class="flex items-center space-x-2 mb-2">
              <Icon name="info" size={14} className="text-gray-600 dark:text-gray-400" />
              <span class="text-sm font-semibold text-gray-800 dark:text-gray-200">Relationship Type</span>
            </div>
            <p class="text-sm text-gray-600 dark:text-gray-300">
              {#if selectedRelationship.type === 'one-to-one'}
                Each record in <strong>{selectedRelationship.from}</strong> relates to exactly one record in <strong>{selectedRelationship.to}</strong>.
              {:else if selectedRelationship.type === 'one-to-many'}
                Each record in <strong>{selectedRelationship.from}</strong> can relate to multiple records in <strong>{selectedRelationship.to}</strong>.
              {:else if selectedRelationship.type === 'many-to-many'}
                Multiple records in <strong>{selectedRelationship.from}</strong> can relate to multiple records in <strong>{selectedRelationship.to}</strong>.
              {:else}
                A relationship exists between these tables through the specified columns.
              {/if}
            </p>
          </div>
        </div>
      </div>
    {/if}
  </div>
</div>

