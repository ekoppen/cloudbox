<script lang="ts">
  import { page } from '$app/stores';
  import { onMount } from 'svelte';
  import { auth } from '$lib/stores/auth';
  import { API_ENDPOINTS, createApiRequest } from '$lib/config';
  
  // Import UI components
  import Card from '$lib/components/ui/card.svelte';
  import Button from '$lib/components/ui/button.svelte';
  import Badge from '$lib/components/ui/badge.svelte';
  import Icon from '$lib/components/ui/icon.svelte';
  import SchemaVisualizer from '$lib/components/database/schema-visualizer.svelte';
  import DataTable from '$lib/components/database/data-table.svelte';
  import DatabaseNavigation from '$lib/components/database/navigation.svelte';
  import StorageBucketInterface from '$lib/components/storage/bucket-interface.svelte';
  import RelationshipDiagram from '$lib/components/database/relationship-diagram.svelte';

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

  // State management
  let activeView = 'schema';
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
  
  // Table Editor state
  let selectedDataTable: string | null = null;
  let tableData: any[] = [];
  let loadingData = false;
  let currentPage = 1;
  let itemsPerPage = 50;
  let totalPages = 1;
  let availableColumns: string[] = [];
  let visibleColumns: string[] = [];
  let showColumnFilter = false;
  let editingCell: {row: number, column: string} | null = null;
  let editingValue = '';
  let columnWidths: {[key: string]: number} = {};
  let isResizing = false;
  let resizingColumn: string | null = null;
  
  // SQL Editor state
  let sqlQuery = `SELECT 
  *
FROM 
  pages
LIMIT 
  100;`;
  let sqlResults: any[] = [];
  let sqlError = '';
  let executingSql = false;
  let savedQueries: {id: string, name: string, query: string, created_at: string}[] = [
    {
      id: '1',
      name: 'Admin Role Assignment', 
      query: `SELECT 
  u.email,
  u.created_at,
  ur.role
FROM auth.users u
LEFT JOIN public.user_roles ur ON u.id = ur.user_id
ORDER BY u.created_at DESC;`,
      created_at: '2024-01-15'
    },
    {
      id: '2', 
      name: 'User Accounts with Roles',
      query: `SELECT 
  u.email,
  u.created_at,
  ur.role
FROM auth.users u
LEFT JOIN public.user_roles ur ON u.id = ur.user_id
ORDER BY u.created_at DESC;`,
      created_at: '2024-01-10'
    }
  ];
  let showSaveQueryModal = false;
  let saveQueryName = '';
  let selectedQueryTab = 'main';
  
  // Storage view state
  let storageBuckets: any[] = [];
  let loadingStorage = false;

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
    
    return relationships;
  }

  function handleTableSelect(event: CustomEvent) {
    selectedTable = event.detail.tableName;
  }
  
  async function loadTableData(tableName: string) {
    if (!project) return;
    
    loadingData = true;
    selectedDataTable = tableName;
    
    try {
      const response = await createApiRequest(
        API_ENDPOINTS.admin.projects.collections.documents.list(project.id.toString(), tableName),
        {
          headers: {
            'Authorization': `Bearer ${$auth.token}`,
            'Content-Type': 'application/json',
          },
        }
      );

      if (response.ok) {
        const data = await response.json();
        tableData = data.documents || [];
        totalPages = Math.ceil(tableData.length / itemsPerPage);
        currentPage = 1;
        
        // Set up columns for filtering
        if (tableData.length > 0) {
          const firstRow = tableData[0].data || tableData[0];
          availableColumns = Object.keys(firstRow);
          visibleColumns = [...availableColumns]; // Show all columns by default
        }
      } else {
        console.error('Failed to load table data');
        tableData = [];
        totalPages = 1;
        currentPage = 1;
        availableColumns = [];
        visibleColumns = [];
      }
    } catch (err) {
      console.error('Error loading table data:', err);
      tableData = [];
    } finally {
      loadingData = false;
    }
  }
  
  async function loadStorageBuckets() {
    if (!project) return;
    
    loadingStorage = true;
    
    try {
      const response = await createApiRequest(
        API_ENDPOINTS.admin.projects.storage.buckets.list(project.id.toString()),
        {
          headers: {
            'Authorization': `Bearer ${$auth.token}`,
            'Content-Type': 'application/json',
          },
        }
      );

      if (response.ok) {
        storageBuckets = await response.json();
      } else {
        console.error('Failed to load storage buckets');
        storageBuckets = [];
      }
    } catch (err) {
      console.error('Error loading storage buckets:', err);
      storageBuckets = [];
    } finally {
      loadingStorage = false;
    }
  }
  
  // Load data when switching tabs
  $: if (activeView === 'table-editor' && tables.length > 0 && !selectedDataTable) {
    loadTableData(tables[0].name);
  }
  
  $: if (activeView === 'storage' && storageBuckets.length === 0) {
    loadStorageBuckets();
  }
  
  async function executeSqlQuery() {
    if (!project || !sqlQuery.trim()) return;
    
    executingSql = true;
    sqlError = '';
    
    try {
      // Parse the SQL query to extract table name
      const selectMatch = sqlQuery.match(/FROM\s+(\w+)/i);
      if (selectMatch && selectMatch[1]) {
        const tableName = selectMatch[1];
        
        // Check if it's a simple SELECT query
        if (sqlQuery.toLowerCase().includes('select')) {
          // Load data from the table
          const response = await createApiRequest(
            API_ENDPOINTS.admin.projects.collections.documents.list(project.id.toString(), tableName),
            {
              headers: {
                'Authorization': `Bearer ${$auth.token}`,
                'Content-Type': 'application/json',
              },
            }
          );

          if (response.ok) {
            const data = await response.json();
            const documents = data.documents || [];
            
            // Extract limit from query
            const limitMatch = sqlQuery.match(/LIMIT\s+(\d+)/i);
            const limit = limitMatch ? parseInt(limitMatch[1]) : 100;
            
            // Format results for display
            sqlResults = documents.slice(0, limit).map((doc: any) => doc.data || doc);
            
            if (sqlResults.length === 0) {
              sqlError = `No data found in table '${tableName}'`;
            }
          } else {
            sqlError = `Table '${tableName}' not found`;
            sqlResults = [];
          }
        } else {
          sqlError = 'Only SELECT queries are currently supported';
          sqlResults = [];
        }
      } else {
        sqlError = 'Invalid SQL query format';
        sqlResults = [];
      }
    } catch (err) {
      sqlError = 'Error executing SQL query: ' + err.message;
      sqlResults = [];
    } finally {
      executingSql = false;
    }
  }
  
  // Calculate paginated data
  $: paginatedData = tableData.slice(
    (currentPage - 1) * itemsPerPage,
    currentPage * itemsPerPage
  );

  // Column resizing functions
  function startColumnResize(event: MouseEvent, columnName: string) {
    event.preventDefault();
    event.stopPropagation();
    isResizing = true;
    resizingColumn = columnName;
    
    const startX = event.clientX;
    const startWidth = columnWidths[columnName] || 120;
    
    function onMouseMove(e: MouseEvent) {
      if (!isResizing || !resizingColumn) return;
      const diff = e.clientX - startX;
      const newWidth = Math.max(80, startWidth + diff);
      columnWidths[resizingColumn] = newWidth;
      // Force reactivity update
      columnWidths = { ...columnWidths };
    }
    
    function onMouseUp() {
      isResizing = false;
      resizingColumn = null;
      document.removeEventListener('mousemove', onMouseMove);
      document.removeEventListener('mouseup', onMouseUp);
    }
    
    document.addEventListener('mousemove', onMouseMove);
    document.addEventListener('mouseup', onMouseUp);
  }

  // Handle cell editing
  function saveCellEdit() {
    if (editingCell && editingValue !== '') {
      // TODO: Implement save to API
      console.log('Saving cell edit:', editingCell, editingValue);
    }
    editingCell = null;
    editingValue = '';
  }

  // SQL Editor functions
  function saveQuery() {
    if (saveQueryName && sqlQuery.trim()) {
      const newQuery = {
        id: Date.now().toString(),
        name: saveQueryName,
        query: sqlQuery,
        created_at: new Date().toISOString().split('T')[0]
      };
      savedQueries = [newQuery, ...savedQueries];
      showSaveQueryModal = false;
      saveQueryName = '';
    }
  }

  function loadSavedQuery(query: string) {
    sqlQuery = query;
  }

  function deleteQuery(queryId: string) {
    savedQueries = savedQueries.filter(q => q.id !== queryId);
  }

  // Syntax highlighting function (basic implementation)
  function highlightSQL(sql: string): string {
    const keywords = ['SELECT', 'FROM', 'WHERE', 'JOIN', 'LEFT', 'RIGHT', 'INNER', 'OUTER', 'ON', 'ORDER', 'BY', 'GROUP', 'HAVING', 'INSERT', 'UPDATE', 'DELETE', 'CREATE', 'ALTER', 'DROP', 'DESC', 'ASC', 'LIMIT', 'OFFSET', 'AND', 'OR', 'NOT', 'NULL', 'IS', 'IN', 'LIKE', 'BETWEEN', 'EXISTS'];
    
    let highlighted = sql;
    
    // Highlight keywords
    keywords.forEach(keyword => {
      const regex = new RegExp(`\\b${keyword}\\b`, 'gi');
      highlighted = highlighted.replace(regex, `<span class="text-blue-600 font-semibold">${keyword}</span>`);
    });
    
    // Highlight strings
    highlighted = highlighted.replace(/'([^']*)'/g, '<span class="text-green-600">\'$1\'</span>');
    
    // Highlight comments
    highlighted = highlighted.replace(/--.*$/gm, '<span class="text-gray-500 italic">$&</span>');
    
    return highlighted;
  }

  // Sample data for other views
  let sampleSections = [
    { id: 'collections', name: 'Collections', icon: 'table', count: schemaStats.totalTables },
    { id: 'documents', name: 'Documents', icon: 'file', count: schemaStats.totalRecords },
    { id: 'storage', name: 'Storage', icon: 'hard-drive', count: 0 },
    { id: 'functions', name: 'Functions', icon: 'code', count: 0 }
  ];

  $: sampleSections = [
    { id: 'collections', name: 'Collections', icon: 'table', count: schemaStats.totalTables },
    { id: 'documents', name: 'Documents', icon: 'file', count: schemaStats.totalRecords },
    { id: 'storage', name: 'Storage', icon: 'hard-drive', count: 0 },
    { id: 'functions', name: 'Functions', icon: 'code', count: 0 }
  ];
</script>

<svelte:head>
  <title>Data Visualization - {project?.name || 'CloudBox'}</title>
</svelte:head>

<div class="h-full flex flex-col space-y-6">
  <!-- Header -->
  <div class="flex items-center justify-between">
    <div class="flex items-center space-x-4">
      <div class="w-10 h-10 bg-primary/10 rounded-lg flex items-center justify-center">
        <Icon name="bar-chart" size={20} className="text-primary" />
      </div>
      <div>
        <h1 class="text-2xl font-bold text-foreground">Data Visualization</h1>
        <p class="text-sm text-muted-foreground">
          Verken en visualiseer je database structuur en gegevens
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
        <span>Database</span>
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
        <p class="mt-4 text-muted-foreground">Data laden...</p>
      </Card>
    </div>
  {:else if error}
    <div class="flex-1 flex items-center justify-center">
      <Card class="glassmorphism-content p-12 text-center">
        <div class="w-16 h-16 bg-destructive/10 rounded-lg flex items-center justify-center mx-auto mb-4">
          <Icon name="alert-triangle" size={32} className="text-destructive" />
        </div>
        <h3 class="text-lg font-medium text-foreground mb-2">Fout bij laden van data</h3>
        <p class="text-muted-foreground mb-4">{error}</p>
        <Button on:click={loadSchema} class="flex items-center space-x-2">
          <Icon name="refresh" size={16} />
          <span>Opnieuw proberen</span>
        </Button>
      </Card>
    </div>
  {:else}
    <!-- Navigation Tabs -->
    <div class="flex space-x-1 bg-gray-100 dark:bg-gray-800 p-1 rounded-lg">
      <button
        class="flex items-center space-x-2 px-4 py-2 rounded-md text-sm font-medium transition-colors {activeView === 'schema' ? 'bg-white dark:bg-gray-700 text-gray-900 dark:text-gray-100 shadow-sm' : 'text-gray-600 dark:text-gray-400 hover:text-gray-900 dark:hover:text-gray-100'}"
        on:click={() => activeView = 'schema'}
      >
        <Icon name="git-branch" size={16} />
        <span>Database Schema</span>
      </button>
      <button
        class="flex items-center space-x-2 px-4 py-2 rounded-md text-sm font-medium transition-colors {activeView === 'table-editor' ? 'bg-white dark:bg-gray-700 text-gray-900 dark:text-gray-100 shadow-sm' : 'text-gray-600 dark:text-gray-400 hover:text-gray-900 dark:hover:text-gray-100'}"
        on:click={() => activeView = 'table-editor'}
      >
        <Icon name="table" size={16} />
        <span>Table Editor</span>
      </button>
      <button
        class="flex items-center space-x-2 px-4 py-2 rounded-md text-sm font-medium transition-colors {activeView === 'sql-editor' ? 'bg-white dark:bg-gray-700 text-gray-900 dark:text-gray-100 shadow-sm' : 'text-gray-600 dark:text-gray-400 hover:text-gray-900 dark:hover:text-gray-100'}"
        on:click={() => activeView = 'sql-editor'}
      >
        <Icon name="terminal" size={16} />
        <span>SQL Editor</span>
      </button>
    </div>


    <!-- Content Area -->
    <div class="flex-1 min-h-0">
      <div class="h-full bg-white dark:bg-gray-800 border border-gray-200 dark:border-gray-700 rounded-lg overflow-hidden">
        {#if activeView === 'schema'}
          <SchemaVisualizer
            {tables}
            {selectedTable}
            on:tableSelect={handleTableSelect}
          />
        {:else if activeView === 'table-editor'}
          <div class="h-full flex bg-gray-50 dark:bg-gray-900">
            <!-- Supabase-style Sidebar -->
            <div class="w-64 bg-white dark:bg-gray-800 border-r border-gray-200 dark:border-gray-700 flex-shrink-0">
              <!-- Schema selector -->
              <div class="p-4 border-b border-gray-200 dark:border-gray-700">
                <select class="w-full text-sm border border-gray-300 dark:border-gray-600 rounded px-2 py-1 bg-white dark:bg-gray-700 text-gray-900 dark:text-gray-100">
                  <option>schema public</option>
                </select>
              </div>
              
              <!-- Table list -->
              <div class="p-2">
                <div class="mb-2">
                  <input 
                    type="text" 
                    placeholder="Search tables..." 
                    class="w-full text-xs px-2 py-1 border border-gray-300 dark:border-gray-600 rounded bg-white dark:bg-gray-700 text-gray-900 dark:text-gray-100 placeholder-gray-500 dark:placeholder-gray-400"
                  />
                </div>
                {#each tables as table}
                  <button
                    class="w-full text-left px-2 py-1 rounded text-sm hover:bg-gray-100 dark:hover:bg-gray-700 transition-colors {selectedDataTable === table.name ? 'bg-blue-50 dark:bg-blue-900/30 text-blue-700 dark:text-blue-400 border-l-2 border-blue-500' : 'text-gray-700 dark:text-gray-300'}"
                    on:click={() => { selectedDataTable = table.name; loadTableData(table.name); }}
                  >
                    <div class="flex items-center space-x-2">
                      <Icon name="table" size={12} className="text-gray-400 dark:text-gray-500" />
                      <span>{table.name}</span>
                    </div>
                  </button>
                {/each}
              </div>
            </div>
            
            <!-- Main Content -->
            <div class="flex-1 flex flex-col">
              {#if selectedDataTable}
                <!-- Supabase-style Header -->
                <div class="bg-white dark:bg-gray-800 border-b border-gray-200 dark:border-gray-700">
                  <div class="flex items-center justify-between px-4 py-3">
                    <div class="flex items-center space-x-4">
                      <!-- Table tabs -->
                      <div class="flex">
                        <button class="px-3 py-1 text-sm bg-gray-100 dark:bg-gray-700 text-gray-700 dark:text-gray-300 border border-gray-300 dark:border-gray-600 rounded-l">
                          <Icon name="table" size={12} className="inline mr-1" />
                          {selectedDataTable}
                        </button>
                        <button class="px-2 py-1 text-sm border-t border-r border-b border-gray-300 dark:border-gray-600 text-gray-400 dark:text-gray-500 hover:text-gray-600 dark:hover:text-gray-300">
                          +
                        </button>
                      </div>
                      <span class="text-xs text-gray-500 dark:text-gray-400">
                        {tableData.length} rows
                      </span>
                    </div>
                    
                    <!-- Toolbar -->
                    <div class="flex items-center space-x-2">
                      <button class="p-1 text-gray-400 dark:text-gray-500 hover:text-gray-600 dark:hover:text-gray-300">
                        <Icon name="filter" size={14} />
                      </button>
                      <button class="p-1 text-gray-400 dark:text-gray-500 hover:text-gray-600 dark:hover:text-gray-300">
                        <Icon name="sort-desc" size={14} />
                      </button>
                      <button class="p-1 text-gray-400 dark:text-gray-500 hover:text-gray-600 dark:hover:text-gray-300">
                        <Icon name="refresh" size={14} />
                      </button>
                    </div>
                  </div>
                </div>
                
                <!-- Supabase-style Data Table -->
                <div class="flex-1 overflow-hidden">
                  {#if loadingData}
                    <div class="flex items-center justify-center h-64">
                      <div class="text-center">
                        <div class="animate-spin rounded-full h-6 w-6 border-2 border-blue-500 border-t-transparent mx-auto"></div>
                        <p class="mt-2 text-sm text-gray-500">Loading data...</p>
                      </div>
                    </div>
                  {:else if tableData.length === 0}
                    <div class="flex items-center justify-center h-64">
                      <div class="text-center">
                        <Icon name="inbox" size={32} className="text-gray-400 mx-auto mb-2" />
                        <p class="text-sm text-gray-500">This table has no data yet.</p>
                      </div>
                    </div>
                  {:else if paginatedData && paginatedData.length > 0}
                    <div class="overflow-auto h-full">
                      <table class="w-full border-collapse">
                        <thead class="bg-gray-50 dark:bg-gray-700 sticky top-0">
                          <tr>
                            <th class="w-8 p-0 border-r border-gray-200 dark:border-gray-600">
                              <div class="h-8 flex items-center justify-center">
                                <input type="checkbox" class="rounded" />
                              </div>
                            </th>
                            {#each availableColumns as column}
                              <th class="border-r border-gray-200 dark:border-gray-600 p-0 text-left font-medium text-xs text-gray-700 dark:text-gray-300 relative" style="width: {columnWidths[column] || 120}px; min-width: 80px;">
                                <div class="h-8 flex items-center px-2 relative group">
                                  <span class="truncate">{column}</span>
                                  <button class="ml-auto opacity-0 group-hover:opacity-100 p-0.5 hover:bg-gray-200 dark:hover:bg-gray-600 rounded">
                                    <Icon name="chevron-down" size={10} />
                                  </button>
                                  <!-- Column resize handle -->
                                  <div 
                                    class="absolute right-0 top-0 h-full w-1 cursor-col-resize hover:bg-blue-500 opacity-0 hover:opacity-100 {resizingColumn === column ? 'bg-blue-500 opacity-100' : ''}"
                                    on:mousedown={(e) => startColumnResize(e, column)}
                                  ></div>
                                </div>
                              </th>
                            {/each}
                          </tr>
                        </thead>
                        <tbody>
                          {#each paginatedData as row, rowIndex}
                            {@const rowData = row.data || row}
                            <tr class="hover:bg-gray-50 dark:hover:bg-gray-700">
                              <td class="w-8 h-8 border-r border-gray-200 dark:border-gray-600 border-b border-gray-100 dark:border-gray-700">
                                <div class="h-full flex items-center justify-center">
                                  <input type="checkbox" class="rounded" />
                                </div>
                              </td>
                              {#each availableColumns as column}
                                {@const value = rowData[column]}
                                {@const isEditing = editingCell && editingCell.row === rowIndex && editingCell.column === column}
                                <td 
                                  class="border-r border-gray-200 dark:border-gray-600 border-b border-gray-100 dark:border-gray-700 p-0 cursor-cell hover:bg-blue-50 dark:hover:bg-blue-900/20" 
                                  style="width: {columnWidths[column] || 120}px; height: 32px;"
                                  on:click={() => {
                                    editingCell = {row: rowIndex, column};
                                    editingValue = value?.toString() || '';
                                  }}
                                >
                                  {#if isEditing}
                                    <input 
                                      type="text" 
                                      bind:value={editingValue}
                                      class="w-full h-full px-2 text-sm border-0 focus:ring-2 focus:ring-blue-500 focus:outline-none bg-white dark:bg-gray-800 text-gray-900 dark:text-gray-100"
                                      on:blur={saveCellEdit}
                                      on:keydown={(e) => {
                                        if (e.key === 'Enter') {
                                          saveCellEdit();
                                        }
                                        if (e.key === 'Escape') {
                                          editingCell = null;
                                          editingValue = '';
                                        }
                                        if (e.key === 'Tab') {
                                          e.preventDefault();
                                          saveCellEdit();
                                          // Move to next cell
                                          const nextColumnIndex = availableColumns.indexOf(column) + 1;
                                          if (nextColumnIndex < availableColumns.length) {
                                            editingCell = {row: rowIndex, column: availableColumns[nextColumnIndex]};
                                            editingValue = rowData[availableColumns[nextColumnIndex]]?.toString() || '';
                                          }
                                        }
                                      }}
                                      autofocus
                                    />
                                  {:else}
                                    <div class="h-full flex items-center px-2 text-sm overflow-hidden">
                                      {#if typeof value === 'object' && value !== null}
                                        <span class="text-xs text-gray-500 dark:text-gray-400 italic truncate w-full" title={JSON.stringify(value)}>{JSON.stringify(value)}</span>
                                      {:else if typeof value === 'boolean'}
                                        <span class="text-xs px-1.5 py-0.5 rounded {value ? 'bg-green-100 dark:bg-green-900/30 text-green-700 dark:text-green-400' : 'bg-red-100 dark:bg-red-900/30 text-red-700 dark:text-red-400'}">
                                          {value}
                                        </span>
                                      {:else if value === null || value === undefined}
                                        <span class="text-gray-400 dark:text-gray-500 italic text-xs">NULL</span>
                                      {:else}
                                        <span class="truncate w-full text-gray-900 dark:text-gray-100" title={value?.toString()}>{value}</span>
                                      {/if}
                                    </div>
                                  {/if}
                                </td>
                              {/each}
                            </tr>
                          {/each}
                        </tbody>
                      </table>
                    </div>
                  {/if}
                </div>
                
                <!-- Supabase-style Footer -->
                <div class="bg-white dark:bg-gray-800 border-t border-gray-200 dark:border-gray-700 px-4 py-2">
                  <div class="flex items-center justify-between text-xs text-gray-500 dark:text-gray-400">
                    <div class="flex items-center space-x-4">
                      <span>Page {currentPage} of {totalPages}</span>
                      <select class="border border-gray-300 dark:border-gray-600 rounded px-2 py-1 text-xs bg-white dark:bg-gray-700 text-gray-900 dark:text-gray-100">
                        <option value="100">100 rows</option>
                        <option value="50">50 rows</option>
                        <option value="25">25 rows</option>
                      </select>
                    </div>
                    <div class="flex items-center space-x-1">
                      <button 
                        class="px-2 py-1 border border-gray-300 dark:border-gray-600 rounded text-xs hover:bg-gray-50 dark:hover:bg-gray-700 disabled:opacity-50 bg-white dark:bg-gray-800 text-gray-700 dark:text-gray-300"
                        disabled={currentPage === 1}
                        on:click={() => currentPage = Math.max(1, currentPage - 1)}
                      >
                        Previous
                      </button>
                      <button 
                        class="px-2 py-1 border border-gray-300 dark:border-gray-600 rounded text-xs hover:bg-gray-50 dark:hover:bg-gray-700 disabled:opacity-50 bg-white dark:bg-gray-800 text-gray-700 dark:text-gray-300"
                        disabled={currentPage === totalPages}
                        on:click={() => currentPage = Math.min(totalPages, currentPage + 1)}
                      >
                        Next
                      </button>
                    </div>
                  </div>
                </div>
              {:else}
                <div class="flex-1 flex items-center justify-center">
                  <div class="text-center">
                    <Icon name="table" size={48} className="text-gray-400 dark:text-gray-600 mx-auto mb-4" />
                    <h3 class="text-lg font-medium mb-2 text-gray-900 dark:text-gray-100">Select a Table</h3>
                    <p class="text-gray-600 dark:text-gray-400">Choose a table from the sidebar to view its data.</p>
                  </div>
                </div>
              {/if}
            </div>
          </div>
        {:else if activeView === 'sql-editor'}
          <div class="h-full flex bg-gray-50 dark:bg-gray-900">
            <!-- Supabase-style Sidebar -->
            <div class="w-64 bg-white dark:bg-gray-800 border-r border-gray-200 dark:border-gray-700 flex-shrink-0 flex flex-col">
              <!-- Header -->
              <div class="p-4 border-b border-gray-200 dark:border-gray-700">
                <h3 class="text-sm font-medium text-gray-900 dark:text-gray-100">SQL Editor</h3>
                <input 
                  type="text" 
                  placeholder="Search queries..." 
                  class="w-full mt-2 text-xs px-2 py-1 border border-gray-300 dark:border-gray-600 rounded bg-white dark:bg-gray-700 text-gray-900 dark:text-gray-100 placeholder-gray-500 dark:placeholder-gray-400"
                />
              </div>
              
              <!-- Categories -->
              <div class="flex-1 overflow-y-auto">
                <!-- Shared Section -->
                <div class="p-2">
                  <button class="w-full flex items-center justify-between px-2 py-1 text-xs text-gray-600 dark:text-gray-400 hover:bg-gray-50 dark:hover:bg-gray-700 rounded">
                    <div class="flex items-center space-x-2">
                      <Icon name="users" size={12} />
                      <span>SHARED</span>
                    </div>
                    <Icon name="chevron-down" size={10} />
                  </button>
                </div>

                <!-- Private Section -->
                <div class="p-2">
                  <button class="w-full flex items-center justify-between px-2 py-1 text-xs text-gray-600 dark:text-gray-400 hover:bg-gray-50 dark:hover:bg-gray-700 rounded">
                    <div class="flex items-center space-x-2">
                      <Icon name="lock" size={12} />
                      <span>PRIVATE ({savedQueries.length})</span>
                    </div>
                    <Icon name="chevron-down" size={10} />
                  </button>
                  
                  <!-- Saved Queries List -->
                  <div class="mt-1 ml-2 space-y-1">
                    {#each savedQueries as query}
                      <div class="group flex items-center justify-between px-2 py-1 text-xs rounded hover:bg-gray-100 dark:hover:bg-gray-700">
                        <button 
                          class="flex-1 text-left truncate text-gray-700 dark:text-gray-300"
                          on:click={() => loadSavedQuery(query.query)}
                          title={query.name}
                        >
                          {query.name}
                        </button>
                        <button 
                          class="opacity-0 group-hover:opacity-100 p-0.5 text-gray-400 dark:text-gray-500 hover:text-red-600 dark:hover:text-red-500"
                          on:click={() => deleteQuery(query.id)}
                        >
                          <Icon name="trash-2" size={10} />
                        </button>
                      </div>
                    {/each}
                  </div>
                </div>

                <!-- Community Section -->
                <div class="p-2">
                  <button class="w-full flex items-center justify-between px-2 py-1 text-xs text-gray-600 dark:text-gray-400 hover:bg-gray-50 dark:hover:bg-gray-700 rounded">
                    <div class="flex items-center space-x-2">
                      <Icon name="globe" size={12} />
                      <span>COMMUNITY</span>
                    </div>
                    <Icon name="chevron-down" size={10} />
                  </button>
                </div>
              </div>
              
              <!-- Footer -->
              <div class="p-2 border-t border-gray-200 dark:border-gray-700">
                <div class="text-xs text-gray-500 dark:text-gray-400">View running queries</div>
              </div>
            </div>
            
            <!-- Main SQL Editor -->
            <div class="flex-1 flex flex-col">
              <!-- Tab Bar -->
              <div class="bg-white dark:bg-gray-800 border-b border-gray-200 dark:border-gray-700 flex items-center">
                <div class="flex">
                  <button 
                    class="px-4 py-2 text-sm border-r border-gray-200 dark:border-gray-700 bg-white dark:bg-gray-800 text-gray-700 dark:text-gray-300 font-medium {selectedQueryTab === 'main' ? 'border-b-2 border-blue-500' : ''}"
                    on:click={() => selectedQueryTab = 'main'}
                  >
                    New query
                  </button>
                </div>
                <div class="flex-1"></div>
                <!-- Toolbar -->
                <div class="flex items-center space-x-2 px-4">
                  <button 
                    class="p-1 text-gray-400 dark:text-gray-500 hover:text-gray-600 dark:hover:text-gray-300"
                    on:click={() => showSaveQueryModal = true}
                    title="Save query"
                  >
                    <Icon name="save" size={14} />
                  </button>
                  <button class="p-1 text-gray-400 dark:text-gray-500 hover:text-gray-600 dark:hover:text-gray-300">
                    <Icon name="download" size={14} />
                  </button>
                  <div class="w-px h-4 bg-gray-300 dark:bg-gray-600"></div>
                  <select class="text-xs border border-gray-300 dark:border-gray-600 rounded px-2 py-1 bg-white dark:bg-gray-700 text-gray-900 dark:text-gray-100">
                    <option>Primary Database</option>
                  </select>
                  <Button
                    on:click={executeSqlQuery}
                    disabled={executingSql}
                    size="sm"
                    class="flex items-center space-x-1 bg-green-600 hover:bg-green-700 text-white px-3 py-1 text-xs"
                  >
                    <Icon name="play" size={12} />
                    <span>Run</span>
                  </Button>
                </div>
              </div>
              
              <!-- Query Editor - Vertical Layout -->
              <div class="flex-1 flex flex-col">
                <!-- SQL Input Area -->
                <div class="h-64 border-b border-gray-200 dark:border-gray-700 flex flex-col">
                  <div class="flex-1 relative">
                    <textarea
                      bind:value={sqlQuery}
                      class="w-full h-full p-4 bg-white dark:bg-gray-800 text-gray-900 dark:text-gray-100 font-mono text-sm resize-none focus:outline-none border-none placeholder-gray-500 dark:placeholder-gray-400"
                      placeholder="-- Start writing your SQL query
SELECT column1, column2
FROM your_table
WHERE condition = 'value';"
                      spellcheck="false"
                      style="line-height: 1.6;"
                    />
                  </div>
                  
                  <!-- SQL Status Bar -->
                  <div class="bg-gray-50 dark:bg-gray-700 border-t border-gray-200 dark:border-gray-600 px-4 py-2">
                    <div class="flex items-center justify-between text-xs text-gray-500 dark:text-gray-400">
                      <div>Lines: {sqlQuery.split('\n').length}</div>
                      <div>Syntax: SQL</div>
                    </div>
                  </div>
                </div>
                
                <!-- Results Area -->
                <div class="flex-1 flex flex-col bg-white dark:bg-gray-800">
                  <!-- Results Header -->
                  <div class="border-b border-gray-200 dark:border-gray-700 p-3">
                    <div class="flex items-center space-x-4">
                      <button class="flex items-center space-x-1 text-xs font-medium text-gray-700 dark:text-gray-300 border-b-2 border-blue-500 pb-1">
                        <Icon name="table" size={12} />
                        <span>Results</span>
                      </button>
                      <button class="flex items-center space-x-1 text-xs text-gray-500 dark:text-gray-400">
                        <Icon name="bar-chart-2" size={12} />
                        <span>Chart</span>
                      </button>
                    </div>
                  </div>
                  
                  <!-- Results Content -->
                  <div class="flex-1 overflow-auto">
                    {#if executingSql}
                      <div class="flex items-center justify-center h-full">
                        <div class="text-center">
                          <div class="animate-spin rounded-full h-6 w-6 border-2 border-blue-500 border-t-transparent mx-auto"></div>
                          <p class="mt-2 text-sm text-gray-500 dark:text-gray-400">Executing query...</p>
                        </div>
                      </div>
                    {:else if sqlError}
                      <div class="p-4">
                        <div class="bg-red-50 dark:bg-red-900/20 border border-red-200 dark:border-red-800 rounded p-3">
                          <div class="flex items-start space-x-2">
                            <Icon name="alert-circle" size={14} className="text-red-500 mt-0.5" />
                            <div>
                              <p class="text-sm font-medium text-red-800 dark:text-red-400">Query Error</p>
                              <p class="text-sm text-red-700 dark:text-red-300 mt-1">{sqlError}</p>
                            </div>
                          </div>
                        </div>
                      </div>
                    {:else if sqlResults.length > 0}
                      <div class="p-4">
                        <div class="mb-2 text-xs text-gray-500 dark:text-gray-400">
                          {sqlResults.length} rows returned
                        </div>
                        <div class="border border-gray-200 dark:border-gray-700 rounded overflow-hidden">
                          <table class="w-full text-sm">
                            <thead class="bg-gray-50 dark:bg-gray-700">
                              <tr>
                                {#each Object.keys(sqlResults[0]) as key}
                                  <th class="text-left px-3 py-2 text-xs font-medium text-gray-700 dark:text-gray-300 border-r border-gray-200 dark:border-gray-600">
                                    {key}
                                  </th>
                                {/each}
                              </tr>
                            </thead>
                            <tbody>
                              {#each sqlResults as row, i}
                                <tr class="border-b border-gray-100 dark:border-gray-700 hover:bg-gray-50 dark:hover:bg-gray-700">
                                  {#each Object.values(row) as value}
                                    <td class="px-3 py-2 border-r border-gray-200 dark:border-gray-600 text-sm text-gray-900 dark:text-gray-100">
                                      {#if value === null}
                                        <span class="text-gray-400 dark:text-gray-500 italic">NULL</span>
                                      {:else}
                                        {value}
                                      {/if}
                                    </td>
                                  {/each}
                                </tr>
                              {/each}
                            </tbody>
                          </table>
                        </div>
                      </div>
                    {:else}
                      <div class="flex items-center justify-center h-full">
                        <div class="text-center text-gray-500 dark:text-gray-400">
                          <Icon name="play-circle" size={48} className="mx-auto mb-4 opacity-30" />
                          <p class="text-sm">Click Run to execute your query</p>
                        </div>
                      </div>
                    {/if}
                  </div>
                </div>
              </div>
            </div>
          </div>
        {/if}
      </div>
    </div>

  {/if}
</div>

<!-- Save Query Modal -->
{#if showSaveQueryModal}
  <div class="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center p-4 z-50">
    <div class="bg-white dark:bg-gray-800 rounded-lg p-6 w-full max-w-md">
      <h3 class="text-lg font-medium mb-4 text-gray-900 dark:text-gray-100">Save Query</h3>
      
      <div class="mb-4">
        <label class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">
          Query Name
        </label>
        <input
          type="text"
          bind:value={saveQueryName}
          placeholder="Enter a name for your query"
          class="w-full px-3 py-2 border border-gray-300 dark:border-gray-600 rounded-md focus:ring-2 focus:ring-blue-500 focus:border-blue-500 bg-white dark:bg-gray-700 text-gray-900 dark:text-gray-100 placeholder-gray-500 dark:placeholder-gray-400"
          autofocus
        />
      </div>
      
      <div class="flex items-center justify-end space-x-3">
        <Button
          variant="secondary"
          size="sm"
          on:click={() => {
            showSaveQueryModal = false;
            saveQueryName = '';
          }}
          class="border-gray-300 dark:border-gray-600 text-gray-700 dark:text-gray-300 bg-white dark:bg-gray-700 hover:bg-gray-50 dark:hover:bg-gray-600"
        >
          Cancel
        </Button>
        <Button
          size="sm"
          on:click={saveQuery}
          disabled={!saveQueryName.trim()}
          class="bg-blue-600 hover:bg-blue-700 text-white"
        >
          Save Query
        </Button>
      </div>
    </div>
  </div>
{/if}