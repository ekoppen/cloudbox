<script lang="ts">
  import { page } from '$app/stores';
  import { onMount } from 'svelte';
  import { auth } from '$lib/stores/auth';
  import { API_ENDPOINTS, createApiRequest } from '$lib/config';
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

  interface Project {
    id: number;
    slug: string;
    name: string;
  }

  let tables: Table[] = [];
  let project: Project | null = null;
  let selectedTable: string = '';
  let tableData: TableData | null = null;
  let loading = false;
  let loadingTables = true;
  let showCreateTable = false;
  let showAddRow = false;
  let newTableName = '';
  let newRowData: Record<string, any> = {};
  let error = '';

  $: projectId = $page.params.id;

  onMount(() => {
    loadProject();
  });

  async function loadProject() {
    try {
      const response = await createApiRequest(API_ENDPOINTS.projects.get(projectId), {
        headers: {
          'Authorization': `Bearer ${$auth.token}`,
          'Content-Type': 'application/json',
        },
      });

      if (response.ok) {
        project = await response.json();
        await loadCollections();
      } else {
        error = 'Project niet gevonden';
      }
    } catch (err) {
      error = 'Fout bij laden van project';
      console.error('Load project error:', err);
    }
  }

  async function loadCollections() {
    if (!project) return;
    
    loadingTables = true;
    error = '';

    try {
      // Load collections using admin API
      const response = await createApiRequest(API_ENDPOINTS.admin.projects.collections.list(project.id.toString()), {
        headers: {
          'Authorization': `Bearer ${$auth.token}`,
          'Content-Type': 'application/json',
        },
      });

      if (response.ok) {
        const collections = await response.json();
        const collectionData = [];

        // Get document count for each collection
        for (const collection of collections) {
          try {
            const documentsResponse = await createApiRequest(API_ENDPOINTS.admin.projects.collections.documents.list(project.id.toString(), collection.name), {
              headers: {
                'Authorization': `Bearer ${$auth.token}`,
                'Content-Type': 'application/json',
              },
            });

            let documentCount = 0;
            if (documentsResponse.ok) {
              const documentsData = await documentsResponse.json();
              documentCount = documentsData.documents ? documentsData.documents.length : (documentsData.total || 0);
            }

            collectionData.push({
              name: collection.name,
              rows: documentCount,
              size: `${Math.round((JSON.stringify(collection).length / 1024) * 10) / 10} KB`,
              created_at: new Date(collection.created_at).toLocaleDateString('nl-NL')
            });
          } catch (err) {
            console.warn(`Failed to load documents for collection ${collection.name}:`, err);
            collectionData.push({
              name: collection.name,
              rows: 0,
              size: '0 KB',
              created_at: new Date(collection.created_at).toLocaleDateString('nl-NL')
            });
          }
        }

        tables = collectionData;
      } else {
        error = 'Fout bij laden van collections';
        console.error('Collections response not OK:', response.status, response.statusText);
      }
    } catch (err) {
      error = 'Fout bij laden van collections';
      console.error('Load collections error:', err);
    } finally {
      loadingTables = false;
    }
  }


  async function selectTable(tableName: string) {
    if (!project) return;
    
    selectedTable = tableName;
    loading = true;
    
    try {
      const response = await createApiRequest(API_ENDPOINTS.admin.projects.collections.documents.list(project.id.toString(), tableName), {
        headers: {
          'Authorization': `Bearer ${$auth.token}`,
          'Content-Type': 'application/json',
        },
      });

      if (response.ok) {
        const data = await response.json();
        
        // Generate columns based on data structure
        let columns: Column[] = [];
        const documents = data.documents || [];
        if (Array.isArray(documents) && documents.length > 0) {
          const firstItem = documents[0].data || documents[0];
          columns = Object.keys(firstItem).map(key => ({
            name: key,
            type: typeof firstItem[key] === 'number' ? 'INTEGER' : 
                  typeof firstItem[key] === 'boolean' ? 'BOOLEAN' :
                  key.includes('_at') ? 'TIMESTAMP' : 'VARCHAR',
            nullable: true,
            default_value: undefined
          }));
        }

        tableData = {
          columns,
          rows: documents.map((doc: any) => ({ id: doc.id, ...doc.data, created_at: doc.created_at, updated_at: doc.updated_at })),
          total_count: documents.length
        };
      } else {
        console.error('Table response not OK:', response.status, response.statusText);
        tableData = { columns: [], rows: [], total_count: 0 };
      }
    } catch (err) {
      console.error(`Error loading table ${tableName}:`, err);
      tableData = { columns: [], rows: [], total_count: 0 };
    } finally {
      loading = false;
    }
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
          Beheer je project collections en data
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
      <Card class="glassmorphism-sidebar">
        <div class="px-4 py-3 border-b border-border">
          <div class="flex items-center space-x-2">
            <Icon name="database" size={16} className="text-primary" />
            <h3 class="text-sm font-medium text-foreground">Collections ({tables.length})</h3>
          </div>
        </div>
        <div class="divide-y divide-border">
          {#if loadingTables}
            <div class="px-4 py-8 text-center">
              <div class="animate-spin rounded-full h-6 w-6 border-b-2 border-primary mx-auto"></div>
              <p class="mt-2 text-xs text-muted-foreground">Collections laden...</p>
            </div>
          {:else if error}
            <div class="px-4 py-4 text-center">
              <Icon name="backup" size={20} className="text-destructive mx-auto mb-2" />
              <p class="text-xs text-destructive">{error}</p>
              <button 
                on:click={loadCollections}
                class="text-xs text-primary hover:underline mt-1"
              >
                Opnieuw proberen
              </button>
            </div>
          {:else if tables.length === 0}
            <div class="px-4 py-8 text-center">
              <Icon name="database" size={20} className="text-muted-foreground mx-auto mb-2" />
              <p class="text-xs text-muted-foreground">Geen collections gevonden</p>
            </div>
          {:else}
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
                    <p class="text-xs text-muted-foreground">{table.rows.toLocaleString()} records</p>
                  </div>
                  <div class="text-right">
                    <p class="text-xs text-muted-foreground">{table.size}</p>
                  </div>
                </div>
              </button>
            {/each}
          {/if}
        </div>
      </Card>
    </div>

    <!-- Main Content -->
    <div class="lg:col-span-3">
      {#if !selectedTable}
        <Card class="glassmorphism-content p-12 text-center">
          <div class="w-16 h-16 bg-muted rounded-lg flex items-center justify-center mx-auto mb-4">
            <Icon name="database" size={32} className="text-muted-foreground" />
          </div>
          <h3 class="text-lg font-medium text-foreground mb-2">Selecteer een collection</h3>
          <p class="text-muted-foreground">Kies een collection uit de linkerkolom om de data te bekijken</p>
        </Card>
      {:else if loading}
        <Card class="glassmorphism-content p-12 text-center">
          <div class="animate-spin rounded-full h-8 w-8 border-b-2 border-primary mx-auto"></div>
          <p class="mt-4 text-muted-foreground">Collection data laden...</p>
        </Card>
      {:else if tableData}
        <div class="space-y-4">
          <!-- Table Info & Actions -->
          <Card class="glassmorphism-card p-4">
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
          <Card class="glassmorphism-table">
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
          <Card class="glassmorphism-table">
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
    <Card class="glassmorphism-modal max-w-md w-full p-6 border-2 shadow-2xl">
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
    <Card class="glassmorphism-modal max-w-2xl w-full p-6 max-h-96 overflow-y-auto border-2 shadow-2xl">
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