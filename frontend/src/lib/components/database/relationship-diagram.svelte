<script lang="ts">
  import { onMount, createEventDispatcher } from 'svelte';
  import Card from '$lib/components/ui/card.svelte';
  import Button from '$lib/components/ui/button.svelte';
  import Badge from '$lib/components/ui/badge.svelte';
  import Icon from '$lib/components/ui/icon.svelte';

  const dispatch = createEventDispatcher();

  interface TableRelationship {
    id: string;
    from_table: string;
    from_column: string;
    to_table: string;
    to_column: string;
    constraint_name: string;
    relationship_type: 'one-to-one' | 'one-to-many' | 'many-to-many';
    on_delete: 'CASCADE' | 'RESTRICT' | 'SET NULL' | 'NO ACTION';
    on_update: 'CASCADE' | 'RESTRICT' | 'SET NULL' | 'NO ACTION';
  }

  interface TableInfo {
    name: string;
    primary_keys: string[];
    foreign_keys: {
      column: string;
      references_table: string;
      references_column: string;
    }[];
    column_count: number;
    row_count: number;
  }

  export let relationships: TableRelationship[] = [];
  export let tables: TableInfo[] = [];
  export let selectedRelationship: string | null = null;
  export let highlightTable: string | null = null;
  export let showLabels: boolean = true;
  export let compactView: boolean = false;

  let svgElement: SVGElement;
  let containerElement: HTMLElement;
  let viewBox = { x: 0, y: 0, width: 1000, height: 600 };

  $: tablePositions = calculateTablePositions(tables, relationships);
  $: relationshipPaths = calculateRelationshipPaths(relationships, tablePositions);

  onMount(() => {
    if (tables.length > 0) {
      fitToContent();
    }
  });

  function calculateTablePositions(tables: TableInfo[], relationships: TableRelationship[]): Map<string, { x: number, y: number }> {
    const positions = new Map();
    
    if (tables.length === 0) return positions;

    // Simple circular layout for demo
    const centerX = 400;
    const centerY = 300;
    const radius = Math.min(200, Math.max(100, tables.length * 30));
    
    tables.forEach((table, index) => {
      const angle = (index / tables.length) * 2 * Math.PI;
      const x = centerX + radius * Math.cos(angle);
      const y = centerY + radius * Math.sin(angle);
      positions.set(table.name, { x, y });
    });

    return positions;
  }

  function calculateRelationshipPaths(relationships: TableRelationship[], positions: Map<string, { x: number, y: number }>): Array<{
    id: string;
    path: string;
    relationship: TableRelationship;
    fromPoint: { x: number, y: number };
    toPoint: { x: number, y: number };
    midPoint: { x: number, y: number };
  }> {
    return relationships.map(rel => {
      const fromPos = positions.get(rel.from_table);
      const toPos = positions.get(rel.to_table);
      
      if (!fromPos || !toPos) {
        return {
          id: rel.id,
          path: '',
          relationship: rel,
          fromPoint: { x: 0, y: 0 },
          toPoint: { x: 0, y: 0 },
          midPoint: { x: 0, y: 0 }
        };
      }

      const fromPoint = {
        x: fromPos.x + 80, // Table width/2
        y: fromPos.y + 20  // Table height/2
      };
      
      const toPoint = {
        x: toPos.x + 80,
        y: toPos.y + 20
      };
      
      const midPoint = {
        x: (fromPoint.x + toPoint.x) / 2,
        y: (fromPoint.y + toPoint.y) / 2
      };

      // Create curved path
      const dx = toPoint.x - fromPoint.x;
      const dy = toPoint.y - fromPoint.y;
      const dist = Math.sqrt(dx * dx + dy * dy);
      const curvature = Math.min(50, dist / 4);
      
      const path = `M ${fromPoint.x} ${fromPoint.y} Q ${midPoint.x + curvature} ${midPoint.y - curvature} ${toPoint.x} ${toPoint.y}`;

      return {
        id: rel.id,
        path,
        relationship: rel,
        fromPoint,
        toPoint,
        midPoint
      };
    });
  }

  function getRelationshipColor(type: string): string {
    switch (type) {
      case 'one-to-one':
        return 'stroke-blue-500';
      case 'one-to-many':
        return 'stroke-green-500';
      case 'many-to-many':
        return 'stroke-purple-500';
      default:
        return 'stroke-gray-400';
    }
  }

  function getRelationshipIcon(type: string): string {
    switch (type) {
      case 'one-to-one':
        return 'minus';
      case 'one-to-many':
        return 'arrow-right';
      case 'many-to-many':
        return 'shuffle';
      default:
        return 'link';
    }
  }

  function handleTableClick(tableName: string) {
    dispatch('tableClick', { tableName });
  }

  function handleRelationshipClick(relationship: TableRelationship) {
    selectedRelationship = selectedRelationship === relationship.id ? null : relationship.id;
    dispatch('relationshipClick', { relationship });
  }

  function fitToContent() {
    if (!svgElement || tablePositions.size === 0) return;
    
    let minX = Infinity, minY = Infinity, maxX = -Infinity, maxY = -Infinity;
    
    for (const [_, pos] of tablePositions) {
      minX = Math.min(minX, pos.x - 80);
      minY = Math.min(minY, pos.y - 40);
      maxX = Math.max(maxX, pos.x + 160);
      maxY = Math.max(maxY, pos.y + 80);
    }
    
    const padding = 50;
    viewBox = {
      x: minX - padding,
      y: minY - padding,
      width: (maxX - minX) + (padding * 2),
      height: (maxY - minY) + (padding * 2)
    };
  }

  function exportDiagram() {
    if (!svgElement) return;
    
    const svgData = new XMLSerializer().serializeToString(svgElement);
    const svgBlob = new Blob([svgData], { type: 'image/svg+xml;charset=utf-8' });
    const url = URL.createObjectURL(svgBlob);
    
    const link = document.createElement('a');
    link.href = url;
    link.download = 'database-relationships.svg';
    link.click();
    
    URL.revokeObjectURL(url);
  }
</script>

<div class="h-full flex flex-col">
  <!-- Header -->
  <div class="flex items-center justify-between p-4 border-b border-border bg-card">
    <div class="flex items-center space-x-4">
      <div class="flex items-center space-x-2">
        <Icon name="share-2" size={16} className="text-primary" />
        <h3 class="text-sm font-semibold text-foreground">Table Relationships</h3>
      </div>
      
      {#if relationships.length > 0}
        <Badge variant="secondary" class="text-xs">
          {relationships.length} relationship{relationships.length === 1 ? '' : 's'}
        </Badge>
      {/if}
    </div>
    
    <div class="flex items-center space-x-2">
      <Button
        variant="outline"
        size="sm"
        on:click={() => showLabels = !showLabels}
        class="flex items-center space-x-2"
      >
        <Icon name={showLabels ? 'eye-off' : 'eye'} size={14} />
        <span>{showLabels ? 'Hide' : 'Show'} Labels</span>
      </Button>
      
      <Button
        variant="outline"
        size="sm"
        on:click={() => compactView = !compactView}
        class="flex items-center space-x-2"
      >
        <Icon name={compactView ? 'maximize' : 'minimize'} size={14} />
        <span>{compactView ? 'Expand' : 'Compact'}</span>
      </Button>
      
      <Button
        variant="outline"
        size="sm"
        on:click={fitToContent}
        class="flex items-center space-x-2"
      >
        <Icon name="zoom-in" size={14} />
        <span>Fit</span>
      </Button>
      
      <Button
        variant="outline"
        size="sm"
        on:click={exportDiagram}
        class="flex items-center space-x-2"
      >
        <Icon name="download" size={14} />
        <span>Export</span>
      </Button>
    </div>
  </div>

  <!-- Diagram -->
  <div class="flex-1 overflow-auto bg-muted/20" bind:this={containerElement}>
    {#if tables.length === 0}
      <div class="absolute inset-0 flex items-center justify-center">
        <div class="text-center">
          <div class="w-16 h-16 bg-muted rounded-lg flex items-center justify-center mx-auto mb-4">
            <Icon name="share-2" size={32} className="text-muted-foreground" />
          </div>
          <h3 class="text-lg font-medium text-foreground mb-2">No Relationships Found</h3>
          <p class="text-muted-foreground">Add foreign key constraints to visualize table relationships</p>
        </div>
      </div>
    {:else}
      <svg
        bind:this={svgElement}
        class="w-full h-full min-h-[400px]"
        viewBox="{viewBox.x} {viewBox.y} {viewBox.width} {viewBox.height}"
      >
        <!-- Background grid -->
        <defs>
          <pattern id="grid" width="20" height="20" patternUnits="userSpaceOnUse">
            <path d="M 20 0 L 0 0 0 20" fill="none" stroke="hsl(var(--border))" stroke-width="0.5" opacity="0.3"/>
          </pattern>
        </defs>
        <rect width="100%" height="100%" fill="url(#grid)" />
        
        <!-- Relationship lines -->
        {#each relationshipPaths as relPath}
          <g class="relationship-group cursor-pointer" 
             role="button"
             tabindex="0"
             on:click={() => handleRelationshipClick(relPath.relationship)}
             on:keydown={(e) => (e.key === 'Enter' || e.key === ' ') && handleRelationshipClick(relPath.relationship)}>
            
            <!-- Main relationship line -->
            <path
              d={relPath.path}
              fill="none"
              stroke-width={selectedRelationship === relPath.id ? "3" : "2"}
              class="{getRelationshipColor(relPath.relationship.relationship_type)} {selectedRelationship === relPath.id ? 'opacity-100' : 'opacity-60'} hover:opacity-100 transition-opacity"
              stroke-dasharray={relPath.relationship.relationship_type === 'many-to-many' ? '5,5' : ''}
            />
            
            <!-- Arrowhead -->
            <polygon
              points="0,-4 8,0 0,4"
              class="{getRelationshipColor(relPath.relationship.relationship_type).replace('stroke-', 'fill-')}"
              transform="translate({relPath.toPoint.x - 8},{relPath.toPoint.y})"
            />
            
            <!-- Relationship type indicator -->
            {#if showLabels}
              <g transform="translate({relPath.midPoint.x},{relPath.midPoint.y})">
                <rect
                  x="-20"
                  y="-12"
                  width="40"
                  height="24"
                  rx="4"
                  fill="hsl(var(--background))"
                  stroke="hsl(var(--border))"
                  stroke-width="1"
                />
                <text
                  x="0"
                  y="4"
                  text-anchor="middle"
                  class="text-xs fill-current text-foreground font-medium"
                >
                  {relPath.relationship.relationship_type.split('-').join(':').toUpperCase()}
                </text>
              </g>
            {/if}
          </g>
        {/each}
        
        <!-- Table nodes -->
        {#each tables as table}
          {@const pos = tablePositions.get(table.name)}
          {#if pos}
            <g transform="translate({pos.x},{pos.y})" class="table-node cursor-pointer"
               role="button"
               tabindex="0"
               on:click={() => handleTableClick(table.name)}
               on:keydown={(e) => (e.key === 'Enter' || e.key === ' ') && handleTableClick(table.name)}>
              
              <!-- Table card -->
              <rect
                x="0"
                y="0"
                width="160"
                height={compactView ? "40" : "60"}
                rx="8"
                fill="hsl(var(--card))"
                stroke="hsl(var(--border))"
                stroke-width={highlightTable === table.name ? "3" : "2"}
                class={highlightTable === table.name ? 'stroke-primary' : 'hover:stroke-primary transition-colors'}
              />
              
              <!-- Table header -->
              <rect
                x="0"
                y="0"
                width="160"
                height="30"
                rx="8"
                fill="hsl(var(--primary))"
                fill-opacity="0.1"
              />
              
              <!-- Table icon -->
              <foreignObject x="8" y="6" width="18" height="18">
                <div class="w-full h-full flex items-center justify-center">
                  <Icon name="table" size={14} className="text-primary" />
                </div>
              </foreignObject>
              
              <!-- Table name -->
              <text
                x="32"
                y="20"
                class="text-sm font-semibold fill-current text-foreground"
              >
                {table.name}
              </text>
              
              {#if !compactView}
                <!-- Table stats -->
                <text
                  x="8"
                  y="45"
                  class="text-xs fill-current text-muted-foreground"
                >
                  {table.column_count} cols • {table.row_count.toLocaleString()} rows
                </text>
                
                <!-- Key indicators -->
                {#if table.primary_keys.length > 0}
                  <foreignObject x="130" y="35" width="12" height="12">
                    <div class="w-full h-full flex items-center justify-center" title="Primary Keys: {table.primary_keys.join(', ')}">
                      <Icon name="key" size={10} className="text-yellow-600" />
                    </div>
                  </foreignObject>
                {/if}
                
                {#if table.foreign_keys.length > 0}
                  <foreignObject x="145" y="35" width="12" height="12">
                    <div class="w-full h-full flex items-center justify-center" title="Foreign Keys: {table.foreign_keys.length}">
                      <Icon name="link" size={10} className="text-blue-600" />
                    </div>
                  </foreignObject>
                {/if}
              {/if}
            </g>
          {/if}
        {/each}
      </svg>
    {/if}
  </div>

  <!-- Legend -->
  {#if relationships.length > 0}
    <div class="border-t border-border bg-card p-4">
      <div class="flex items-center justify-between">
        <div class="flex items-center space-x-6">
          <h4 class="text-xs font-medium text-foreground uppercase tracking-wider">Relationship Types</h4>
          
          <div class="flex items-center space-x-4">
            <div class="flex items-center space-x-2">
              <div class="w-4 h-0.5 bg-blue-500"></div>
              <span class="text-xs text-muted-foreground">One-to-One</span>
            </div>
            
            <div class="flex items-center space-x-2">
              <div class="w-4 h-0.5 bg-green-500"></div>
              <span class="text-xs text-muted-foreground">One-to-Many</span>
            </div>
            
            <div class="flex items-center space-x-2">
              <div class="w-4 h-0.5 bg-purple-500 border-dashed border-t"></div>
              <span class="text-xs text-muted-foreground">Many-to-Many</span>
            </div>
          </div>
        </div>
        
        {#if selectedRelationship}
          {@const selectedRel = relationships.find(r => r.id === selectedRelationship)}
          {#if selectedRel}
            <div class="text-xs text-muted-foreground">
              Selected: {selectedRel.from_table}.{selectedRel.from_column} → {selectedRel.to_table}.{selectedRel.to_column}
            </div>
          {/if}
        {/if}
      </div>
    </div>
  {/if}
</div>

<style>
  .relationship-group:hover path {
    stroke-width: 3;
  }
  
  .table-node:hover rect {
    stroke-width: 3;
  }
</style>