<script lang="ts">
  import { createEventDispatcher } from 'svelte';
  import Card from '$lib/components/ui/card.svelte';
  import Button from '$lib/components/ui/button.svelte';
  import Input from '$lib/components/ui/input.svelte';
  import Badge from '$lib/components/ui/badge.svelte';
  import Icon from '$lib/components/ui/icon.svelte';
  import { toast } from '$lib/stores/toast';

  const dispatch = createEventDispatcher();

  interface Column {
    name: string;
    type: string;
    nullable: boolean;
    primary_key?: boolean;
    sortable?: boolean;
    width?: string;
  }

  interface TableData {
    columns: Column[];
    rows: Record<string, any>[];
    total_count: number;
  }

  export let data: TableData;
  export let tableName: string = '';
  export let loading: boolean = false;
  export let editable: boolean = false;
  export let selectable: boolean = false;
  export let searchable: boolean = true;
  export let paginated: boolean = true;
  export let pageSize: number = 50;

  let searchQuery = '';
  let currentPage = 0;
  let sortColumn: string | null = null;
  let sortDirection: 'asc' | 'desc' = 'asc';
  let selectedRows: Set<string> = new Set();
  let editingCell: { row: number; column: string } | null = null;
  let editValue = '';

  $: filteredRows = searchQuery 
    ? data.rows.filter(row => 
        Object.values(row).some(value => 
          String(value).toLowerCase().includes(searchQuery.toLowerCase())
        )
      )
    : data.rows;

  $: sortedRows = sortColumn 
    ? [...filteredRows].sort((a, b) => {
        const aVal = a[sortColumn];
        const bVal = b[sortColumn];
        
        if (aVal === null || aVal === undefined) return sortDirection === 'asc' ? 1 : -1;
        if (bVal === null || bVal === undefined) return sortDirection === 'asc' ? -1 : 1;
        
        if (typeof aVal === 'number' && typeof bVal === 'number') {
          return sortDirection === 'asc' ? aVal - bVal : bVal - aVal;
        }
        
        const aStr = String(aVal).toLowerCase();
        const bStr = String(bVal).toLowerCase();
        
        if (sortDirection === 'asc') {
          return aStr < bStr ? -1 : aStr > bStr ? 1 : 0;
        } else {
          return aStr > bStr ? -1 : aStr < bStr ? 1 : 0;
        }
      })
    : filteredRows;

  $: paginatedRows = paginated 
    ? sortedRows.slice(currentPage * pageSize, (currentPage + 1) * pageSize)
    : sortedRows;

  $: totalPages = paginated ? Math.ceil(sortedRows.length / pageSize) : 1;

  function getTypeColor(type: string): string {
    const lowerType = type.toLowerCase();
    if (lowerType.includes('int') || lowerType.includes('serial') || lowerType.includes('number')) {
      return 'bg-blue-100 dark:bg-blue-900/30 text-blue-800 dark:text-blue-200';
    }
    if (lowerType.includes('varchar') || lowerType.includes('text') || lowerType.includes('string')) {
      return 'bg-green-100 dark:bg-green-900/30 text-green-800 dark:text-green-200';
    }
    if (lowerType.includes('boolean') || lowerType.includes('bool')) {
      return 'bg-purple-100 dark:bg-purple-900/30 text-purple-800 dark:text-purple-200';
    }
    if (lowerType.includes('timestamp') || lowerType.includes('date')) {
      return 'bg-yellow-100 dark:bg-yellow-900/30 text-yellow-800 dark:text-yellow-200';
    }
    return 'bg-gray-100 dark:bg-gray-900/30 text-gray-800 dark:text-gray-200';
  }

  function formatValue(value: any, column: Column): string {
    if (value === null || value === undefined) return '—';
    if (column.type.toLowerCase().includes('boolean')) {
      return value ? '✅' : '❌';
    }
    if (column.type.toLowerCase().includes('date') || column.type.toLowerCase().includes('timestamp')) {
      try {
        return new Date(value).toLocaleString();
      } catch {
        return String(value);
      }
    }
    if (typeof value === 'string' && value.length > 100) {
      return value.substring(0, 100) + '...';
    }
    return String(value);
  }

  function handleSort(columnName: string) {
    const column = data.columns.find(col => col.name === columnName);
    if (!column?.sortable) return;
    
    if (sortColumn === columnName) {
      sortDirection = sortDirection === 'asc' ? 'desc' : 'asc';
    } else {
      sortColumn = columnName;
      sortDirection = 'asc';
    }
  }

  function toggleRowSelection(rowIndex: number) {
    const row = paginatedRows[rowIndex];
    const rowId = row.id || String(rowIndex);
    
    if (selectedRows.has(rowId)) {
      selectedRows.delete(rowId);
    } else {
      selectedRows.add(rowId);
    }
    selectedRows = new Set(selectedRows);
  }

  function selectAllRows() {
    if (selectedRows.size === paginatedRows.length) {
      selectedRows.clear();
    } else {
      selectedRows = new Set(paginatedRows.map((row, index) => row.id || String(index)));
    }
  }

  function startEdit(rowIndex: number, columnName: string) {
    if (!editable) return;
    
    const row = paginatedRows[rowIndex];
    editingCell = { row: rowIndex, column: columnName };
    editValue = String(row[columnName] || '');
  }

  function cancelEdit() {
    editingCell = null;
    editValue = '';
  }

  function saveEdit() {
    if (!editingCell) return;
    
    const row = paginatedRows[editingCell.row];
    const column = data.columns.find(col => col.name === editingCell.column);
    
    if (!column) return;
    
    let parsedValue: any = editValue;
    
    // Type conversion
    if (column.type.toLowerCase().includes('int') || column.type.toLowerCase().includes('number')) {
      parsedValue = editValue === '' ? null : Number(editValue);
      if (isNaN(parsedValue)) {
        toast.error('Invalid number format');
        return;
      }
    } else if (column.type.toLowerCase().includes('boolean')) {
      parsedValue = editValue.toLowerCase() === 'true' || editValue === '1';
    }
    
    dispatch('cellEdit', {
      rowId: row.id,
      column: editingCell.column,
      oldValue: row[editingCell.column],
      newValue: parsedValue
    });
    
    // Update local data
    row[editingCell.column] = parsedValue;
    data = { ...data };
    
    cancelEdit();
    toast.success('Cell updated successfully');
  }

  function handleKeydown(event: KeyboardEvent) {
    if (!editingCell) return;
    
    if (event.key === 'Enter') {
      saveEdit();
    } else if (event.key === 'Escape') {
      cancelEdit();
    }
  }

  function exportData() {
    const headers = data.columns.map(col => col.name).join(',');
    const rows = sortedRows.map(row => 
      data.columns.map(col => {
        const value = row[col.name];
        return value === null || value === undefined ? '' : `"${String(value).replace(/"/g, '""')}"`;
      }).join(',')
    ).join('\n');
    
    const csv = headers + '\n' + rows;
    const blob = new Blob([csv], { type: 'text/csv' });
    const url = URL.createObjectURL(blob);
    
    const link = document.createElement('a');
    link.href = url;
    link.download = `${tableName || 'data'}.csv`;
    link.click();
    
    URL.revokeObjectURL(url);
  }

  function deleteSelectedRows() {
    if (selectedRows.size === 0) return;
    
    if (!confirm(`Delete ${selectedRows.size} selected row(s)?`)) return;
    
    dispatch('deleteRows', {
      rowIds: Array.from(selectedRows)
    });
    
    selectedRows.clear();
    selectedRows = new Set();
  }
</script>

<div class="h-full flex flex-col">
  <!-- Header -->
  <div class="flex items-center justify-between p-4 border-b border-border bg-card">
    <div class="flex items-center space-x-4">
      <div class="flex items-center space-x-2">
        <Icon name="table" size={16} className="text-primary" />
        <h3 class="text-sm font-semibold text-foreground">
          {tableName || 'Data Table'}
        </h3>
      </div>
      
      <div class="flex items-center space-x-2">
        <Badge variant="secondary" class="text-xs">
          {filteredRows.length.toLocaleString()} rows
        </Badge>
        
        {#if data.total_count > filteredRows.length}
          <Badge variant="outline" class="text-xs">
            of {data.total_count.toLocaleString()} total
          </Badge>
        {/if}
        
        {#if selectedRows.size > 0}
          <Badge variant="default" class="text-xs">
            {selectedRows.size} selected
          </Badge>
        {/if}
      </div>
    </div>
    
    <div class="flex items-center space-x-2">
      {#if selectedRows.size > 0}
        <Button
          variant="destructive"
          size="sm"
          on:click={deleteSelectedRows}
          class="flex items-center space-x-2"
        >
          <Icon name="trash" size={14} />
          <span>Delete ({selectedRows.size})</span>
        </Button>
      {/if}
      
      <Button
        variant="outline"
        size="sm"
        on:click={exportData}
        class="flex items-center space-x-2"
      >
        <Icon name="download" size={14} />
        <span>Export</span>
      </Button>
      
      <Button
        variant="outline"
        size="sm"
        on:click={() => dispatch('refresh')}
        class="flex items-center space-x-2"
      >
        <Icon name="refresh-cw" size={14} />
        <span>Refresh</span>
      </Button>
    </div>
  </div>

  <!-- Search and Filters -->
  {#if searchable}
    <div class="p-4 border-b border-border bg-card">
      <div class="flex items-center space-x-4">
        <div class="flex-1 relative">
          <Icon name="search" size={16} className="absolute left-3 top-1/2 transform -translate-y-1/2 text-muted-foreground" />
          <Input
            type="text"
            placeholder="Search in table data..."
            bind:value={searchQuery}
            class="pl-10"
          />
        </div>
        
        {#if sortColumn}
          <Button
            variant="outline"
            size="sm"
            on:click={() => { sortColumn = null; sortDirection = 'asc'; }}
            class="flex items-center space-x-2"
          >
            <Icon name="x" size={14} />
            <span>Clear Sort</span>
          </Button>
        {/if}
      </div>
    </div>
  {/if}

  <!-- Table -->
  <div class="flex-1 overflow-auto">
    {#if loading}
      <div class="flex items-center justify-center h-64">
        <div class="text-center">
          <div class="animate-spin rounded-full h-8 w-8 border-b-2 border-primary mx-auto"></div>
          <p class="mt-2 text-sm text-muted-foreground">Loading data...</p>
        </div>
      </div>
    {:else if paginatedRows.length === 0}
      <div class="flex items-center justify-center h-64">
        <div class="text-center">
          <div class="w-16 h-16 bg-muted rounded-lg flex items-center justify-center mx-auto mb-4">
            <Icon name="database" size={32} className="text-muted-foreground" />
          </div>
          <h3 class="text-lg font-medium text-foreground mb-2">No data found</h3>
          <p class="text-muted-foreground">
            {searchQuery ? 'No rows match your search criteria' : 'This table contains no data'}
          </p>
        </div>
      </div>
    {:else}
      <table class="min-w-full">
        <thead class="bg-muted/30 sticky top-0 z-10">
          <tr>
            {#if selectable}
              <th class="w-12 px-4 py-3 text-left">
                <input
                  type="checkbox"
                  checked={selectedRows.size === paginatedRows.length && paginatedRows.length > 0}
                  on:change={selectAllRows}
                  class="rounded border-border text-primary focus:ring-primary"
                />
              </th>
            {/if}
            
            {#each data.columns as column}
              <th
                class="px-4 py-3 text-left text-xs font-medium text-muted-foreground uppercase tracking-wider {column.sortable ? 'cursor-pointer hover:bg-muted/50' : ''}"
                style={column.width ? `width: ${column.width}` : ''}
                on:click={() => handleSort(column.name)}
              >
                <div class="flex items-center space-x-2">
                  <span>{column.name}</span>
                  
                  {#if column.primary_key}
                    <Icon name="key" size={12} className="text-yellow-600" />
                  {/if}
                  
                  <Badge variant="outline" class="text-xs px-1 py-0 {getTypeColor(column.type)}">
                    {column.type}
                  </Badge>
                  
                  {#if column.sortable}
                    {#if sortColumn === column.name}
                      <Icon name={sortDirection === 'asc' ? 'arrow-up' : 'arrow-down'} size={12} className="text-primary" />
                    {:else}
                      <Icon name="chevrons-up-down" size={12} className="text-muted-foreground opacity-50" />
                    {/if}
                  {/if}
                </div>
              </th>
            {/each}
            
            {#if editable}
              <th class="w-20 px-4 py-3 text-right text-xs font-medium text-muted-foreground uppercase tracking-wider">
                Actions
              </th>
            {/if}
          </tr>
        </thead>
        
        <tbody class="bg-card divide-y divide-border">
          {#each paginatedRows as row, rowIndex}
            <tr class="hover:bg-muted/30 transition-colors {selectedRows.has(row.id || String(rowIndex)) ? 'bg-primary/5' : ''}">
              {#if selectable}
                <td class="px-4 py-3">
                  <input
                    type="checkbox"
                    checked={selectedRows.has(row.id || String(rowIndex))}
                    on:change={() => toggleRowSelection(rowIndex)}
                    class="rounded border-border text-primary focus:ring-primary"
                  />
                </td>
              {/if}
              
              {#each data.columns as column}
                <td class="px-4 py-3 text-sm">
                  {#if editingCell && editingCell.row === rowIndex && editingCell.column === column.name}
                    <Input
                      type="text"
                      bind:value={editValue}
                      on:keydown={handleKeydown}
                      on:blur={saveEdit}
                      class="h-8 text-xs"
                      autofocus
                    />
                  {:else}
                    {#if editable}
                      <button
                        class="text-foreground cursor-pointer hover:bg-muted/50 rounded px-2 py-1 w-full text-left bg-transparent border-none"
                        on:dblclick={() => startEdit(rowIndex, column.name)}
                        on:keydown={(e) => (e.key === 'Enter' || e.key === ' ') && startEdit(rowIndex, column.name)}
                        title={String(row[column.name] || '')}
                      >
                        {formatValue(row[column.name], column)}
                      </button>
                    {:else}
                      <div
                        class="text-foreground"
                        title={String(row[column.name] || '')}
                      >
                        {formatValue(row[column.name], column)}
                      </div>
                    {/if}
                  {/if}
                </td>
              {/each}
              
              {#if editable}
                <td class="px-4 py-3 text-right">
                  <Button
                    variant="ghost"
                    size="sm"
                    on:click={() => dispatch('editRow', { row, rowIndex })}
                    class="h-8 w-8 p-0"
                  >
                    <Icon name="edit" size={14} />
                  </Button>
                </td>
              {/if}
            </tr>
          {/each}
        </tbody>
      </table>
    {/if}
  </div>

  <!-- Pagination -->
  {#if paginated && totalPages > 1}
    <div class="flex items-center justify-between p-4 border-t border-border bg-card">
      <div class="flex items-center space-x-2 text-sm text-muted-foreground">
        <span>
          Showing {(currentPage * pageSize) + 1} to {Math.min((currentPage + 1) * pageSize, filteredRows.length)}
          of {filteredRows.length.toLocaleString()} rows
        </span>
      </div>
      
      <div class="flex items-center space-x-2">
        <Button
          variant="outline"
          size="sm"
          disabled={currentPage === 0}
          on:click={() => currentPage = 0}
        >
          <Icon name="chevrons-left" size={14} />
        </Button>
        
        <Button
          variant="outline"
          size="sm"
          disabled={currentPage === 0}
          on:click={() => currentPage--}
        >
          <Icon name="chevron-left" size={14} />
        </Button>
        
        <div class="flex items-center space-x-1">
          {#each Array.from({length: Math.min(5, totalPages)}, (_, i) => {
            const page = Math.max(0, Math.min(totalPages - 5, currentPage - 2)) + i;
            return page;
          }) as page}
            <Button
              variant={currentPage === page ? 'default' : 'outline'}
              size="sm"
              on:click={() => currentPage = page}
              class="w-8 h-8 p-0"
            >
              {page + 1}
            </Button>
          {/each}
        </div>
        
        <Button
          variant="outline"
          size="sm"
          disabled={currentPage >= totalPages - 1}
          on:click={() => currentPage++}
        >
          <Icon name="chevron-right" size={14} />
        </Button>
        
        <Button
          variant="outline"
          size="sm"
          disabled={currentPage >= totalPages - 1}
          on:click={() => currentPage = totalPages - 1}
        >
          <Icon name="chevrons-right" size={14} />
        </Button>
      </div>
    </div>
  {/if}
</div>