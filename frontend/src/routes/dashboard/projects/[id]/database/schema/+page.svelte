<script lang="ts">
  import { page } from '$app/stores';
  import { onMount } from 'svelte';
  import { auth } from '$lib/stores/auth';
  import { API_ENDPOINTS, createApiRequest } from '$lib/config';
  import Card from '$lib/components/ui/card.svelte';
  import Button from '$lib/components/ui/button.svelte';
  import Badge from '$lib/components/ui/badge.svelte';
  import Icon from '$lib/components/ui/icon.svelte';
  import SchemaVisualizer from '$lib/components/database/schema-visualizer.svelte';

  interface TableSchema {
    name: string;
    columns: {
      name: string;
      type: string;
      nullable: boolean;
      primary_key?: boolean;
      foreign_key?: {
        table: string;
        column: string;
      };
      default_value?: string;
    }[];
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

  interface Project {
    id: number;
    slug: string;
    name: string;
  }

  let project: Project | null = null;
  let tables: TableSchema[] = [];
  let selectedTable: string | null = null;
  let loading = true;
  let error = '';
  let schemaStats = {
    totalTables: 0,
    totalColumns: 0,
    totalRecords: 0,
    averageTableSize: '0 KB'
  };

  $: projectId = $page.params.id;

  onMount(() => {
    loadProject();
  });

  async function loadProject() {
    try {
      console.log('Loading project with ID:', projectId);
      const response = await createApiRequest(API_ENDPOINTS.projects.get(projectId), {
        headers: {
          'Authorization': `Bearer ${$auth.token}`,
          'Content-Type': 'application/json',
        },
      });

      if (response.ok) {
        project = await response.json();
        console.log('Loaded project:', project);
        await loadSchema();
      } else {
        error = 'Project niet gevonden';
        console.error('Project response not OK:', response.status, response.statusText);
      }
    } catch (err) {
      error = 'Fout bij laden van project';
      console.error('Load project error:', err);
    }
  }

  async function loadSchema() {
    if (!project) return;
    
    loading = true;
    error = '';

    try {
      // Load collections using admin API
      console.log('Loading collections for project:', project.id);
      const collectionsUrl = API_ENDPOINTS.admin.projects.collections.list(project.id.toString());
      console.log('Collections API URL:', collectionsUrl);
      
      const response = await createApiRequest(collectionsUrl, {
        headers: {
          'Authorization': `Bearer ${$auth.token}`,
          'Content-Type': 'application/json',
        },
      });

      console.log('Collections response status:', response.status);
      if (response.ok) {
        const collections = await response.json();
        console.log('Loaded collections:', collections);
        
        if (!collections || !Array.isArray(collections)) {
          console.warn('No collections found or invalid response format');
          tables = [];
          schemaStats = {
            totalTables: 0,
            totalColumns: 0,
            totalRecords: 0,
            averageTableSize: '0 KB'
          };
          return;
        }
        
        const schemaData: TableSchema[] = [];
        let totalRecords = 0;
        let totalColumns = 0;

        // Process each collection to build schema
        for (const collection of collections) {
          try {
            // Get documents to analyze structure with timeout
            const controller = new AbortController();
            const timeoutId = setTimeout(() => controller.abort(), 10000); // 10 second timeout
            
            const documentsResponse = await createApiRequest(API_ENDPOINTS.admin.projects.collections.documents.list(project.id.toString(), collection.name), {
              headers: {
                'Authorization': `Bearer ${$auth.token}`,
                'Content-Type': 'application/json',
              },
              signal: controller.signal
            });
            
            clearTimeout(timeoutId);

            let columns: TableSchema['columns'] = [];
            let documentCount = 0;

            if (documentsResponse.ok) {
              const documentsData = await documentsResponse.json();
              const documents = documentsData.documents || [];
              documentCount = documents.length;
              totalRecords += documentCount;
              console.log(`Collection ${collection.name}: ${documentCount} documents found`);

              // Analyze document structure to generate schema
              if (documents.length > 0) {
                console.log(`Analyzing documents for collection ${collection.name}:`, documents);
                const sampleDoc = documents[0];
                const docData = sampleDoc.data || sampleDoc;
                console.log(`Sample document data for ${collection.name}:`, docData);
                
                // Add default columns
                columns.push({
                  name: 'id',
                  type: 'STRING',
                  nullable: false,
                  primary_key: true
                });

                // Analyze document fields
                for (const [key, value] of Object.entries(docData)) {
                  if (key !== 'id') {
                    const foreignKey = detectForeignKey(key, value, collections);
                    columns.push({
                      name: key,
                      type: getFieldType(value),
                      nullable: true,
                      default_value: undefined,
                      foreign_key: foreignKey
                    });
                  }
                }

                // Add timestamp columns
                columns.push(
                  {
                    name: 'created_at',
                    type: 'TIMESTAMP',
                    nullable: false
                  },
                  {
                    name: 'updated_at',
                    type: 'TIMESTAMP',
                    nullable: false
                  }
                );

                totalColumns += columns.length;
              } else {
                // No documents found, create basic schema
                console.log(`No documents found for collection ${collection.name}, creating basic schema`);
                columns = [
                  { name: 'id', type: 'STRING', nullable: false, primary_key: true },
                  { name: 'created_at', type: 'TIMESTAMP', nullable: false },
                  { name: 'updated_at', type: 'TIMESTAMP', nullable: false }
                ];
                totalColumns += columns.length;
              }
            } else {
              console.warn(`Failed to load documents for collection ${collection.name}:`, documentsResponse.status);
              // Create basic schema even if documents can't be loaded
              columns = [
                { name: 'id', type: 'STRING', nullable: false, primary_key: true },
                { name: 'created_at', type: 'TIMESTAMP', nullable: false },
                { name: 'updated_at', type: 'TIMESTAMP', nullable: false }
              ];
              totalColumns += columns.length;
            }

            schemaData.push({
              name: collection.name,
              columns,
              row_count: documentCount,
              size: `${Math.round((JSON.stringify(collection).length / 1024) * 10) / 10} KB`,
              relationships: detectRelationships(collection, collections)
            });

          } catch (err) {
            console.warn(`Failed to analyze schema for collection ${collection.name}:`, err);
            // Add basic schema for failed collections
            schemaData.push({
              name: collection.name,
              columns: [
                { name: 'id', type: 'STRING', nullable: false, primary_key: true },
                { name: 'created_at', type: 'TIMESTAMP', nullable: false },
                { name: 'updated_at', type: 'TIMESTAMP', nullable: false }
              ],
              row_count: 0,
              size: '0 KB',
              relationships: detectRelationships(collection, collections)
            });
          }
        }

        tables = schemaData;
        
        console.log('Final tables array for visualization:', tables);
        console.log('Number of tables with columns:', tables.filter(t => t.columns.length > 0).length);
        console.log('Tables data structure:', tables.map(t => ({ 
          name: t.name, 
          columnCount: t.columns.length, 
          hasPosition: !!t.position,
          relationships: t.relationships?.length || 0
        })));
        
        // Calculate statistics
        schemaStats = {
          totalTables: tables.length,
          totalColumns,
          totalRecords,
          averageTableSize: tables.length > 0 ? 
            `${Math.round((tables.reduce((sum, table) => sum + parseFloat(table.size), 0) / tables.length) * 10) / 10} KB` : 
            '0 KB'
        };
        
      } else {
        error = 'Fout bij laden van schema';
        console.error('Schema response not OK:', response.status, response.statusText);
      }
    } catch (err) {
      error = 'Fout bij laden van schema';
      console.error('Load schema error:', err);
      // Ensure we always stop loading even on error
      tables = [];
      schemaStats = {
        totalTables: 0,
        totalColumns: 0,
        totalRecords: 0,
        averageTableSize: '0 KB'
      };
    } finally {
      loading = false;
    }
  }

  function getFieldType(value: any): string {
    if (value === null || value === undefined) return 'STRING';
    
    switch (typeof value) {
      case 'number':
        return Number.isInteger(value) ? 'INTEGER' : 'DECIMAL';
      case 'boolean':
        return 'BOOLEAN';
      case 'object':
        if (Array.isArray(value)) return 'ARRAY';
        if (value instanceof Date) return 'TIMESTAMP';
        return 'JSON';
      case 'string':
      default:
        // Check if it's a date string
        if (value.match(/^\d{4}-\d{2}-\d{2}T\d{2}:\d{2}:\d{2}/)) {
          return 'TIMESTAMP';
        }
        // Check if it's a long text (>255 chars)
        if (value.length > 255) {
          return 'TEXT';
        }
        return 'VARCHAR';
    }
  }

  function detectForeignKey(fieldName: string, value: any, collections: any[]): { table: string; column: string } | undefined {
    const collectionNames = collections.map(c => c.name);
    
    // PhotoPortfolio specific foreign keys
    if (fieldName === 'featured_image' && collectionNames.includes('images')) {
      return { table: 'images', column: 'id' };
    }
    
    if (fieldName === 'photos' && Array.isArray(value) && collectionNames.includes('images')) {
      return { table: 'images', column: 'id' };
    }
    
    // Generic foreign key detection based on naming patterns
    for (const collectionName of collectionNames) {
      const patterns = [
        `${collectionName}_id`,
        `${collectionName.slice(0, -1)}_id`, // singular form
      ];
      
      if (patterns.some(pattern => fieldName === pattern)) {
        return { table: collectionName, column: 'id' };
      }
    }
    
    return undefined;
  }

  function detectRelationships(collection: any, allCollections: any[]): TableSchema['relationships'] {
    const relationships: TableSchema['relationships'] = [];
    const collectionNames = allCollections.map(c => c.name);
    
    // PhotoPortfolio specific relationships
    if (collection.name === 'albums') {
      // albums -> images (one-to-many via photos array)
      if (collectionNames.includes('images')) {
        relationships.push({
          to_table: 'images',
          from_column: 'photos',
          to_column: 'id',
          type: 'one-to-many'
        });
      }
    }
    
    if (collection.name === 'images') {
      // images -> albums (many-to-one, reverse of albums->images)
      if (collectionNames.includes('albums')) {
        relationships.push({
          to_table: 'albums',
          from_column: 'id',
          to_column: 'photos',
          type: 'many-to-one'
        });
      }
    }
    
    if (collection.name === 'pages') {
      // pages -> images (one-to-one via featured_image)
      if (collectionNames.includes('images')) {
        relationships.push({
          to_table: 'images',
          from_column: 'featured_image',
          to_column: 'id',
          type: 'one-to-one'
        });
      }
    }
    
    // Generic relationship detection based on naming conventions
    for (const otherCollection of allCollections) {
      if (otherCollection.name === collection.name) continue;
      
      // Look for foreign key patterns in field names
      const foreignKeyPatterns = [
        `${otherCollection.name}_id`,
        `${otherCollection.name.slice(0, -1)}_id`, // singular form
        `${otherCollection.name}`
      ];
      
      // Check if any fields match foreign key patterns
      // This would require analyzing actual document structure
    }
    
    return relationships;
  }

  function handleTableSelect(event: CustomEvent) {
    selectedTable = event.detail.tableName;
  }

  function exportSchema() {
    const schemaExport = {
      project: project?.name,
      generated_at: new Date().toISOString(),
      statistics: schemaStats,
      tables: tables.map(table => ({
        name: table.name,
        columns: table.columns,
        row_count: table.row_count,
        size: table.size
      }))
    };

    const blob = new Blob([JSON.stringify(schemaExport, null, 2)], { type: 'application/json' });
    const url = URL.createObjectURL(blob);
    const a = document.createElement('a');
    a.href = url;
    a.download = `${project?.slug || 'schema'}-schema.json`;
    document.body.appendChild(a);
    a.click();
    document.body.removeChild(a);
    URL.revokeObjectURL(url);
  }
</script>

<svelte:head>
  <title>Database Schema - {project?.name || 'CloudBox'}</title>
</svelte:head>

<div class="h-full flex flex-col space-y-6">
  <!-- Header -->
  <div class="flex items-center justify-between">
    <div class="flex items-center space-x-4">
      <div class="w-10 h-10 bg-primary/10 rounded-lg flex items-center justify-center">
        <Icon name="github" size={20} className="text-primary" />
      </div>
      <div>
        <h1 class="text-2xl font-bold text-foreground">Database Schema</h1>
        <p class="text-sm text-muted-foreground">
          Visualiseer je database structuur en relaties
        </p>
      </div>
    </div>
    <div class="flex items-center space-x-3">
      <Button
        href={`/dashboard/projects/${projectId}/database`}
        variant="secondary"
        size="md"
        class="flex items-center space-x-2"
      >
        <Icon name="database" size={16} />
        <span>Terug naar Database</span>
      </Button>
      <Button 
        on:click={exportSchema}
        variant="outline"
        size="md"
        disabled={tables.length === 0}
        class="flex items-center space-x-2"
      >
        <Icon name="backup" size={16} />
        <span>Schema Exporteren</span>
      </Button>
      <Button 
        on:click={loadSchema}
        size="md"
        class="flex items-center space-x-2"
      >
        <Icon name="refresh" size={16} />
        <span>Vernieuwen</span>
      </Button>
    </div>
  </div>

  {#if loading}
    <div class="flex-1 flex items-center justify-center">
      <Card class="glassmorphism-content p-12 text-center">
        <div class="animate-spin rounded-full h-8 w-8 border-b-2 border-primary mx-auto"></div>
        <p class="mt-4 text-muted-foreground">Schema laden...</p>
      </Card>
    </div>
  {:else if error}
    <div class="flex-1 flex items-center justify-center">
      <Card class="glassmorphism-content p-12 text-center">
        <div class="w-16 h-16 bg-destructive/10 rounded-lg flex items-center justify-center mx-auto mb-4">
          <Icon name="backup" size={32} className="text-destructive" />
        </div>
        <h3 class="text-lg font-medium text-foreground mb-2">Fout bij laden van schema</h3>
        <p class="text-muted-foreground mb-4">{error}</p>
        <Button on:click={loadSchema} class="flex items-center space-x-2">
          <Icon name="refresh" size={16} />
          <span>Opnieuw proberen</span>
        </Button>
      </Card>
    </div>
  {:else}
    <!-- Schema Statistics -->
    <div class="grid grid-cols-1 md:grid-cols-4 gap-4">
      <Card class="glassmorphism-card p-4">
        <div class="flex items-center space-x-3">
          <div class="w-10 h-10 bg-blue-100 dark:bg-gray-800/30 rounded-lg flex items-center justify-center">
            <Icon name="database" size={20} className="text-blue-600 dark:text-blue-400" />
          </div>
          <div>
            <p class="text-2xl font-bold text-foreground">{schemaStats.totalTables}</p>
            <p class="text-sm text-muted-foreground">Tabellen</p>
          </div>
        </div>
      </Card>

      <Card class="glassmorphism-card p-4">
        <div class="flex items-center space-x-3">
          <div class="w-10 h-10 bg-green-100 dark:bg-green-900/30 rounded-lg flex items-center justify-center">
            <Icon name="grid" size={20} className="text-green-600 dark:text-green-400" />
          </div>
          <div>
            <p class="text-2xl font-bold text-foreground">{schemaStats.totalColumns}</p>
            <p class="text-sm text-muted-foreground">Kolommen</p>
          </div>
        </div>
      </Card>

      <Card class="glassmorphism-card p-4">
        <div class="flex items-center space-x-3">
          <div class="w-10 h-10 bg-purple-100 dark:bg-purple-900/30 rounded-lg flex items-center justify-center">
            <Icon name="storage" size={20} className="text-purple-600 dark:text-purple-400" />
          </div>
          <div>
            <p class="text-2xl font-bold text-foreground">{schemaStats.totalRecords.toLocaleString()}</p>
            <p class="text-sm text-muted-foreground">Records</p>
          </div>
        </div>
      </Card>

      <Card class="glassmorphism-card p-4">
        <div class="flex items-center space-x-3">
          <div class="w-10 h-10 bg-yellow-100 dark:bg-yellow-900/30 rounded-lg flex items-center justify-center">
            <Icon name="storage" size={20} className="text-yellow-600 dark:text-yellow-400" />
          </div>
          <div>
            <p class="text-2xl font-bold text-foreground">{schemaStats.averageTableSize}</p>
            <p class="text-sm text-muted-foreground">Gem. Grootte</p>
          </div>
        </div>
      </Card>
    </div>

    <!-- Schema Visualizer -->
    <div class="flex-1 min-h-0">
      {#if tables.length === 0}
        <Card class="glassmorphism-content p-12 text-center h-full flex items-center justify-center">
          <div>
            <div class="w-16 h-16 bg-muted rounded-lg flex items-center justify-center mx-auto mb-4">
              <Icon name="database" size={32} className="text-muted-foreground" />
            </div>
            <h3 class="text-lg font-medium text-foreground mb-2">Geen tabellen gevonden</h3>
            <p class="text-muted-foreground mb-4">
              Er zijn nog geen collections in dit project om een schema van te genereren.
            </p>
            <Button
              href={`/dashboard/projects/${projectId}/database`}
              class="flex items-center space-x-2"
            >
              <Icon name="database" size={16} />
              <span>Ga naar Database</span>
            </Button>
          </div>
        </Card>
      {:else}
        <Card class="glassmorphism-table h-full p-0 overflow-hidden">
          <SchemaVisualizer
            {tables}
            {selectedTable}
            on:tableSelect={handleTableSelect}
          />
        </Card>
      {/if}
    </div>

    <!-- Selected Table Details -->
    {#if selectedTable && tables.find(t => t.name === selectedTable)}
      {@const table = tables.find(t => t.name === selectedTable)}
      <Card class="glassmorphism-card">
        <div class="px-6 py-4 border-b border-border">
          <div class="flex items-center justify-between">
            <div class="flex items-center space-x-3">
              <div class="w-8 h-8 bg-primary/10 rounded-lg flex items-center justify-center">
                <Icon name="layers" size={16} className="text-primary" />
              </div>
              <div>
                <h3 class="text-lg font-semibold text-foreground">{table.name}</h3>
                <p class="text-sm text-muted-foreground">
                  {table.row_count.toLocaleString()} records • {table.columns.length} kolommen
                </p>
              </div>
            </div>
            <Badge variant="outline" class="text-sm px-3 py-1">
              {table.size}
            </Badge>
          </div>
        </div>
        
        <div class="p-6">
          <div class="grid grid-cols-1 lg:grid-cols-2 xl:grid-cols-3 gap-4">
            {#each table.columns as column}
              <div class="flex items-center space-x-3 p-3 rounded-lg border border-border bg-card">
                <div class="flex-shrink-0">
                  {#if column.primary_key}
                    <div class="w-6 h-6 bg-yellow-100 dark:bg-yellow-900/30 rounded flex items-center justify-center">
                      <Icon name="key" size={12} className="text-yellow-600 dark:text-yellow-400" />
                    </div>
                  {:else if column.foreign_key}
                    <div class="w-6 h-6 bg-blue-100 dark:bg-gray-800/30 rounded flex items-center justify-center">
                      <Icon name="link" size={12} className="text-blue-600 dark:text-blue-400" />
                    </div>
                  {:else}
                    <div class="w-6 h-6 bg-muted rounded flex items-center justify-center">
                      <Icon name="database" size={12} className="text-muted-foreground" />
                    </div>
                  {/if}
                </div>
                <div class="flex-1 min-w-0">
                  <div class="flex items-center space-x-2">
                    <p class="text-sm font-medium text-foreground truncate">{column.name}</p>
                    {#if !column.nullable}
                      <div class="w-1.5 h-1.5 bg-red-500 rounded-full" title="NOT NULL"></div>
                    {/if}
                  </div>
                  <p class="text-xs text-muted-foreground">{column.type}</p>
                  {#if column.foreign_key}
                    <p class="text-xs text-blue-600 dark:text-blue-400">
                      → {column.foreign_key.table}.{column.foreign_key.column}
                    </p>
                  {/if}
                </div>
              </div>
            {/each}
          </div>
        </div>
      </Card>
    {/if}
  {/if}
</div>