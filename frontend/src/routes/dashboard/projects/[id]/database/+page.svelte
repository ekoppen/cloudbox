<script lang="ts">
  import { page } from '$app/stores';
  import { onMount } from 'svelte';
  import { auth } from '$lib/stores/auth';
  import Card from '$lib/components/ui/card.svelte';
  import Button from '$lib/components/ui/button.svelte';
  import Input from '$lib/components/ui/input.svelte';
  import Label from '$lib/components/ui/label.svelte';
  import Badge from '$lib/components/ui/badge.svelte';
  import Textarea from '$lib/components/ui/textarea.svelte';
  import Icon from '$lib/components/ui/icon.svelte';

  interface Table {
    name: string;
    rows: number;
    size: string;
    created_at: string;
  }

  interface Column {
    name: string;
    type: string;
    nullable: boolean;
    default_value?: string;
  }

  interface TableData {
    columns: Column[];
    rows: any[];
    total_count: number;
  }

  let tables: Table[] = [
    { name: 'users', rows: 127, size: '2.4 MB', created_at: '2025-01-15' },
    { name: 'posts', rows: 1843, size: '8.7 MB', created_at: '2025-01-16' },
    { name: 'comments', rows: 5621, size: '12.3 MB', created_at: '2025-01-16' },
    { name: 'categories', rows: 23, size: '156 KB', created_at: '2025-01-17' },
    { name: 'tags', rows: 89, size: '234 KB', created_at: '2025-01-17' }
  ];

  let selectedTable: string = '';
  let tableData: TableData | null = null;
  let loading = false;
  let showCreateTable = false;
  let showAddRow = false;
  let newTableName = '';
  let newRowData: Record<string, any> = {};

  $: projectId = $page.params.id;

  // Empty function - will be replaced with real API calls
  function generateMockTableData(tableName: string): TableData {
    return { columns: [], rows: [], total_count: 0 };
  }

  function selectTable(tableName: string) {
    selectedTable = tableName;
    loading = true;
    
    // Simulate API call
    setTimeout(() => {
      tableData = generateMockTableData(tableName);
      loading = false;
    }, 500);
  }

  function formatValue(value: any): string {
    if (value === null || value === undefined) return '—';
    if (typeof value === 'boolean') return value ? '✅' : '❌';
    if (typeof value === 'string' && value.length > 50) return value.substring(0, 50) + '...';
    return String(value);
  }

  function getTypeColor(type: string): string {
    if (type.includes('INTEGER') || type.includes('SERIAL')) return 'bg-blue-100 dark:bg-blue-900 text-blue-800 dark:text-blue-200';
    if (type.includes('VARCHAR') || type.includes('TEXT')) return 'bg-green-100 dark:bg-green-900 text-green-800 dark:text-green-200';
    if (type.includes('BOOLEAN')) return 'bg-purple-100 dark:bg-purple-900 text-purple-800 dark:text-purple-200';
    if (type.includes('TIMESTAMP') || type.includes('DATE')) return 'bg-yellow-100 dark:bg-yellow-900 text-yellow-800 dark:text-yellow-200';
    return 'bg-muted text-muted-foreground';
  }

  function createTable() {
    if (!newTableName.trim()) return;
    
    tables = [...tables, {
      name: newTableName,
      rows: 0,
      size: '0 KB',
      created_at: new Date().toISOString().split('T')[0]
    }];
    
    showCreateTable = false;
    newTableName = '';
  }

  function initNewRow() {
    if (tableData) {
      newRowData = {};
      tableData.columns.forEach(col => {
        if (col.default_value) {
          newRowData[col.name] = col.type.includes('BOOLEAN') ? 
            col.default_value === 'true' : col.default_value;
        } else {
          newRowData[col.name] = '';
        }
      });
    }
  }

  function addRow() {
    if (tableData && Object.keys(newRowData).length > 0) {
      // Generate new ID
      const maxId = Math.max(...tableData.rows.map(row => row.id || 0));
      newRowData.id = maxId + 1;
      newRowData.created_at = new Date().toISOString().replace('T', ' ').substring(0, 19);
      
      tableData.rows = [newRowData, ...tableData.rows];
      tableData.total_count += 1;
      
      showAddRow = false;
      newRowData = {};
    }
  }

  function deleteRow(index: number) {
    if (tableData) {
      tableData.rows = tableData.rows.filter((_, i) => i !== index);
      tableData.total_count -= 1;
    }
  }
</script>

<svelte:head>
  <title>Database - CloudBox</title>
</svelte:head>

<div class="space-y-6">
  <!-- Header -->
  <div class="flex items-center justify-between">
    <div class="flex items-center space-x-4">
      <div class="w-10 h-10 bg-primary/10 rounded-lg flex items-center justify-center">
        <Icon name="database" size={20} className="text-primary" />
      </div>
      <div>
        <h1 class="text-2xl font-bold text-foreground">Database</h1>
        <p class="text-sm text-muted-foreground">
          Beheer je database tabellen en data
        </p>
      </div>
    </div>
    <Button 
      on:click={() => showCreateTable = true}
      class="flex items-center space-x-2"
    >
      <Icon name="package" size={16} />
      <span>Nieuwe Tabel</span>
    </Button>
  </div>

  <div class="grid grid-cols-1 lg:grid-cols-4 gap-6">
    <!-- Tables Sidebar -->
    <div class="lg:col-span-1">
      <Card>
        <div class="px-4 py-3 border-b border-border">
          <div class="flex items-center space-x-2">
            <Icon name="database" size={16} className="text-primary" />
            <h3 class="text-sm font-medium text-foreground">Tabellen ({tables.length})</h3>
          </div>
        </div>
        <div class="divide-y divide-border">
          {#each tables as table}
            <button
              on:click={() => selectTable(table.name)}
              class="w-full px-4 py-3 text-left hover:bg-muted transition-colors {
                selectedTable === table.name ? 'bg-primary/10 border-r-2 border-primary' : ''
              }"
            >
              <div class="flex items-center justify-between">
                <div>
                  <p class="text-sm font-medium text-foreground">{table.name}</p>
                  <p class="text-xs text-muted-foreground">{table.rows.toLocaleString()} rijen</p>
                </div>
                <div class="text-right">
                  <p class="text-xs text-muted-foreground">{table.size}</p>
                </div>
              </div>
            </button>
          {/each}
        </div>
      </Card>
    </div>

    <!-- Main Content -->
    <div class="lg:col-span-3">
      {#if !selectedTable}
        <Card class="p-12 text-center">
          <div class="w-16 h-16 bg-muted rounded-lg flex items-center justify-center mx-auto mb-4">
            <Icon name="database" size={32} className="text-muted-foreground" />
          </div>
          <h3 class="text-lg font-medium text-foreground mb-2">Selecteer een tabel</h3>
          <p class="text-muted-foreground">Kies een tabel uit de linkerkolom om de data te bekijken</p>
        </Card>
      {:else if loading}
        <Card class="p-12 text-center">
          <div class="animate-spin rounded-full h-8 w-8 border-b-2 border-primary mx-auto"></div>
          <p class="mt-4 text-muted-foreground">Tabel data laden...</p>
        </Card>
      {:else if tableData}
        <div class="space-y-4">
          <!-- Table Info & Actions -->
          <Card class="p-4">
            <div class="flex justify-between items-center">
              <div class="flex items-center space-x-3">
                <div class="w-8 h-8 bg-primary/10 rounded-lg flex items-center justify-center">
                  <Icon name="database" size={16} className="text-primary" />
                </div>
                <div>
                  <h2 class="text-lg font-semibold text-foreground">{selectedTable}</h2>
                  <p class="text-sm text-muted-foreground">
                    {tableData.total_count.toLocaleString()} totale rijen • {tableData.rows.length} getoond
                  </p>
                </div>
              </div>
              <div class="flex space-x-2">
                <Button 
                  on:click={() => { initNewRow(); showAddRow = true; }}
                  variant="outline"
                  size="sm"
                  class="flex items-center space-x-2"
                >
                  <Icon name="package" size={16} />
                  <span>Rij Toevoegen</span>
                </Button>
                <Button variant="outline" size="sm" class="flex items-center space-x-2">
                  <Icon name="backup" size={16} />
                  <span>Exporteren</span>
                </Button>
                <Button variant="outline" size="sm" class="flex items-center space-x-2">
                  <Icon name="settings" size={16} />
                  <span>Schema</span>
                </Button>
              </div>
            </div>
          </Card>

          <!-- Table Schema -->
          <Card>
            <div class="px-4 py-3 border-b border-border">
              <div class="flex items-center space-x-2">
                <Icon name="settings" size={16} className="text-primary" />
                <h3 class="text-sm font-medium text-foreground">Schema</h3>
              </div>
            </div>
            <div class="overflow-x-auto">
              <table class="min-w-full divide-y divide-border">
                <thead class="bg-muted/30">
                  <tr>
                    <th class="px-4 py-2 text-left text-xs font-medium text-muted-foreground uppercase">Kolom</th>
                    <th class="px-4 py-2 text-left text-xs font-medium text-muted-foreground uppercase">Type</th>
                    <th class="px-4 py-2 text-left text-xs font-medium text-muted-foreground uppercase">Nullable</th>
                    <th class="px-4 py-2 text-left text-xs font-medium text-muted-foreground uppercase">Standaard</th>
                  </tr>
                </thead>
                <tbody class="bg-card divide-y divide-border">
                  {#each tableData.columns as column}
                    <tr>
                      <td class="px-4 py-2 text-sm font-medium text-foreground">{column.name}</td>
                      <td class="px-4 py-2">
                        <span class="inline-flex px-2 py-1 text-xs font-medium rounded-full {getTypeColor(column.type)}">
                          {column.type}
                        </span>
                      </td>
                      <td class="px-4 py-2 text-sm text-muted-foreground">
                        {column.nullable ? '✅' : '❌'}
                      </td>
                      <td class="px-4 py-2 text-sm text-muted-foreground">
                        {column.default_value || '—'}
                      </td>
                    </tr>
                  {/each}
                </tbody>
              </table>
            </div>
          </Card>

          <!-- Table Data -->
          <Card>
            <div class="px-4 py-3 border-b border-border">
              <div class="flex items-center space-x-2">
                <Icon name="storage" size={16} className="text-primary" />
                <h3 class="text-sm font-medium text-foreground">Data</h3>
              </div>
            </div>
            <div class="overflow-x-auto">
              <table class="min-w-full divide-y divide-border">
                <thead class="bg-muted/30">
                  <tr>
                    {#each tableData.columns as column}
                      <th class="px-4 py-2 text-left text-xs font-medium text-muted-foreground uppercase tracking-wider">
                        {column.name}
                      </th>
                    {/each}
                    <th class="px-4 py-2 text-right text-xs font-medium text-muted-foreground uppercase tracking-wider">
                      Acties
                    </th>
                  </tr>
                </thead>
                <tbody class="bg-card divide-y divide-border">
                  {#each tableData.rows as row, index}
                    <tr class="hover:bg-muted/30">
                      {#each tableData.columns as column}
                        <td class="px-4 py-2 text-sm text-foreground">
                          {formatValue(row[column.name])}
                        </td>
                      {/each}
                      <td class="px-4 py-2 text-right">
                        <div class="flex justify-end space-x-1">
                          <Button variant="ghost" size="sm" class="h-8 w-8 p-0">
                            <Icon name="settings" size={14} />
                          </Button>
                          <Button 
                            variant="ghost" 
                            size="sm" 
                            class="h-8 w-8 p-0 text-destructive hover:text-destructive"
                            on:click={() => deleteRow(index)}
                          >
                            <Icon name="backup" size={14} />
                          </Button>
                        </div>
                      </td>
                    </tr>
                  {/each}
                </tbody>
              </table>
            </div>
          </Card>
        </div>
      {/if}
    </div>
  </div>
</div>

<!-- Create Table Modal -->
{#if showCreateTable}
  <div class="fixed inset-0 bg-black/80 backdrop-blur-sm flex items-center justify-center p-4 z-50">
    <Card class="max-w-md w-full p-6 border-2 shadow-2xl">
      <div class="flex items-center space-x-3 mb-4">
        <div class="w-10 h-10 bg-primary/10 rounded-lg flex items-center justify-center">
          <Icon name="database" size={20} className="text-primary" />
        </div>
        <h2 class="text-xl font-bold text-foreground">Nieuwe Tabel Aanmaken</h2>
      </div>
      
      <form on:submit|preventDefault={createTable} class="space-y-4">
        <div>
          <Label for="table-name">Tabel naam</Label>
          <Input
            id="table-name"
            type="text"
            bind:value={newTableName}
            required
            class="mt-1"
            placeholder="bijv. orders, products"
          />
        </div>
        
        <div class="flex space-x-3 pt-4">
          <Button
            type="button"
            variant="outline"
            on:click={() => { showCreateTable = false; newTableName = ''; }}
            class="flex-1"
          >
            Annuleren
          </Button>
          <Button
            type="submit"
            disabled={!newTableName.trim()}
            class="flex-1"
          >
            Aanmaken
          </Button>
        </div>
      </form>
    </Card>
  </div>
{/if}

<!-- Add Row Modal -->
{#if showAddRow && tableData}
  <div class="fixed inset-0 bg-black/80 backdrop-blur-sm flex items-center justify-center p-4 z-50">
    <Card class="max-w-2xl w-full p-6 max-h-96 overflow-y-auto border-2 shadow-2xl">
      <div class="flex items-center space-x-3 mb-4">
        <div class="w-10 h-10 bg-primary/10 rounded-lg flex items-center justify-center">
          <Icon name="package" size={20} className="text-primary" />
        </div>
        <h2 class="text-xl font-bold text-foreground">Nieuwe Rij Toevoegen</h2>
      </div>
      
      <form on:submit|preventDefault={addRow} class="space-y-4">
        {#each tableData.columns as column}
          {#if column.name !== 'id' && column.name !== 'created_at'}
            <div>
              <Label for="field-{column.name}">
                {column.name}
                <span class="text-xs text-muted-foreground">({column.type})</span>
              </Label>
              
              {#if column.type.includes('BOOLEAN')}
                <select
                  id="field-{column.name}"
                  bind:value={newRowData[column.name]}
                  class="w-full p-2 border border-border rounded-md bg-background text-foreground mt-1 focus:ring-2 focus:ring-primary"
                >
                  <option value={true}>Waar</option>
                  <option value={false}>Onwaar</option>
                </select>
              {:else if column.type.includes('TEXT')}
                <Textarea
                  id="field-{column.name}"
                  bind:value={newRowData[column.name]}
                  class="mt-1"
                  rows={3}
                  placeholder="Voer {column.name} in..."
                />
              {:else}
                <Input
                  id="field-{column.name}"
                  type={column.type.includes('INTEGER') ? 'number' : 'text'}
                  bind:value={newRowData[column.name]}
                  required={!column.nullable}
                  class="mt-1"
                  placeholder="Voer {column.name} in..."
                />
              {/if}
            </div>
          {/if}
        {/each}
        
        <div class="flex space-x-3 pt-4">
          <Button
            type="button"
            variant="outline"
            on:click={() => { showAddRow = false; newRowData = {}; }}
            class="flex-1"
          >
            Annuleren
          </Button>
          <Button
            type="submit"
            class="flex-1"
          >
            Toevoegen
          </Button>
        </div>
      </form>
    </Card>
  </div>
{/if}