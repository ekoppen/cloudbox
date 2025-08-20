<script lang="ts">
  import { page } from '$app/stores';
  import { onMount } from 'svelte';
  import { auth } from '$lib/stores/auth';
  import { toast } from '$lib/stores/toast';
  import { API_ENDPOINTS, createApiRequest } from '$lib/config';
  
  // Import our new data visualization components
  import SchemaVisualizer from '$lib/components/database/schema-visualizer.svelte';
  import DataTable from '$lib/components/database/data-table.svelte';
  import DatabaseNavigation from '$lib/components/database/navigation.svelte';
  import StorageBucketInterface from '$lib/components/storage/bucket-interface.svelte';
  import RelationshipDiagram from '$lib/components/database/relationship-diagram.svelte';
  
  // Import UI components
  import Card from '$lib/components/ui/card.svelte';
  import Button from '$lib/components/ui/button.svelte';
  import Badge from '$lib/components/ui/badge.svelte';
  import Icon from '$lib/components/ui/icon.svelte';

  // Sample data for demonstration
  let activeView = 'schema';
  let selectedTable = 'users';
  let loading = false;

  // Schema visualization data
  let sampleTables = [
    {
      name: 'users',
      columns: [
        { name: 'id', type: 'SERIAL', nullable: false, primary_key: true },
        { name: 'email', type: 'VARCHAR(255)', nullable: false },
        { name: 'username', type: 'VARCHAR(100)', nullable: false },
        { name: 'created_at', type: 'TIMESTAMP', nullable: false },
        { name: 'updated_at', type: 'TIMESTAMP', nullable: true },
        { name: 'profile_id', type: 'INTEGER', nullable: true, foreign_key: { table: 'profiles', column: 'id' } }
      ],
      row_count: 1250,
      size: '45 KB',
      position: { x: 100, y: 100 },
      relationships: [
        { to_table: 'profiles', from_column: 'profile_id', to_column: 'id', type: 'one-to-one' as const },
        { to_table: 'posts', from_column: 'id', to_column: 'user_id', type: 'one-to-many' as const }
      ]
    },
    {
      name: 'profiles',
      columns: [
        { name: 'id', type: 'SERIAL', nullable: false, primary_key: true },
        { name: 'user_id', type: 'INTEGER', nullable: false, foreign_key: { table: 'users', column: 'id' } },
        { name: 'first_name', type: 'VARCHAR(100)', nullable: true },
        { name: 'last_name', type: 'VARCHAR(100)', nullable: true },
        { name: 'bio', type: 'TEXT', nullable: true },
        { name: 'avatar_url', type: 'VARCHAR(500)', nullable: true },
        { name: 'created_at', type: 'TIMESTAMP', nullable: false }
      ],
      row_count: 980,
      size: '78 KB',
      position: { x: 400, y: 150 }
    },
    {
      name: 'posts',
      columns: [
        { name: 'id', type: 'SERIAL', nullable: false, primary_key: true },
        { name: 'user_id', type: 'INTEGER', nullable: false, foreign_key: { table: 'users', column: 'id' } },
        { name: 'title', type: 'VARCHAR(200)', nullable: false },
        { name: 'content', type: 'TEXT', nullable: false },
        { name: 'status', type: 'VARCHAR(20)', nullable: false },
        { name: 'published_at', type: 'TIMESTAMP', nullable: true },
        { name: 'created_at', type: 'TIMESTAMP', nullable: false }
      ],
      row_count: 3450,
      size: '890 KB',
      position: { x: 100, y: 350 }
    },
    {
      name: 'comments',
      columns: [
        { name: 'id', type: 'SERIAL', nullable: false, primary_key: true },
        { name: 'post_id', type: 'INTEGER', nullable: false, foreign_key: { table: 'posts', column: 'id' } },
        { name: 'user_id', type: 'INTEGER', nullable: false, foreign_key: { table: 'users', column: 'id' } },
        { name: 'content', type: 'TEXT', nullable: false },
        { name: 'created_at', type: 'TIMESTAMP', nullable: false }
      ],
      row_count: 8790,
      size: '1.2 MB',
      position: { x: 400, y: 400 }
    }
  ];

  // Sample table data
  let sampleTableData = {
    columns: [
      { name: 'id', type: 'SERIAL', nullable: false, primary_key: true, sortable: true },
      { name: 'email', type: 'VARCHAR(255)', nullable: false, sortable: true },
      { name: 'username', type: 'VARCHAR(100)', nullable: false, sortable: true },
      { name: 'created_at', type: 'TIMESTAMP', nullable: false, sortable: true },
      { name: 'is_active', type: 'BOOLEAN', nullable: false, sortable: true }
    ],
    rows: [
      { id: 1, email: 'alice@example.com', username: 'alice_wonder', created_at: '2024-01-15T10:30:00Z', is_active: true },
      { id: 2, email: 'bob@example.com', username: 'bob_builder', created_at: '2024-01-16T14:20:00Z', is_active: true },
      { id: 3, email: 'charlie@example.com', username: 'charlie_dev', created_at: '2024-01-17T09:15:00Z', is_active: false },
      { id: 4, email: 'diana@example.com', username: 'diana_designer', created_at: '2024-01-18T16:45:00Z', is_active: true },
      { id: 5, email: 'eve@example.com', username: 'eve_analyst', created_at: '2024-01-19T11:30:00Z', is_active: true }
    ],
    total_count: 5
  };

  // Sample storage buckets
  let sampleBuckets = [
    {
      id: 'bucket1',
      name: 'user-uploads',
      description: 'User uploaded files and documents',
      is_public: false,
      max_file_size: 50 * 1024 * 1024,
      file_count: 245,
      total_size: 125 * 1024 * 1024,
      created_at: '2024-01-10T08:00:00Z'
    },
    {
      id: 'bucket2',
      name: 'static-assets',
      description: 'Public static assets for the application',
      is_public: true,
      max_file_size: 10 * 1024 * 1024,
      file_count: 89,
      total_size: 34 * 1024 * 1024,
      created_at: '2024-01-05T12:30:00Z'
    },
    {
      id: 'bucket3',
      name: 'backups',
      description: 'Database backups and system snapshots',
      is_public: false,
      max_file_size: 500 * 1024 * 1024,
      file_count: 12,
      total_size: 1.2 * 1024 * 1024 * 1024,
      created_at: '2024-01-01T00:00:00Z'
    }
  ];

  // Sample relationships
  let sampleRelationships = [
    {
      id: 'rel1',
      from_table: 'users',
      from_column: 'profile_id',
      to_table: 'profiles',
      to_column: 'id',
      constraint_name: 'fk_users_profile',
      relationship_type: 'one-to-one' as const,
      on_delete: 'CASCADE' as const,
      on_update: 'CASCADE' as const
    },
    {
      id: 'rel2',
      from_table: 'posts',
      from_column: 'user_id',
      to_table: 'users',
      to_column: 'id',
      constraint_name: 'fk_posts_user',
      relationship_type: 'one-to-many' as const,
      on_delete: 'CASCADE' as const,
      on_update: 'CASCADE' as const
    },
    {
      id: 'rel3',
      from_table: 'comments',
      from_column: 'post_id',
      to_table: 'posts',
      to_column: 'id',
      constraint_name: 'fk_comments_post',
      relationship_type: 'one-to-many' as const,
      on_delete: 'CASCADE' as const,
      on_update: 'CASCADE' as const
    }
  ];

  // Sample table info for relationships
  let sampleTableInfo = [
    {
      name: 'users',
      primary_keys: ['id'],
      foreign_keys: [{ column: 'profile_id', references_table: 'profiles', references_column: 'id' }],
      column_count: 6,
      row_count: 1250
    },
    {
      name: 'profiles',
      primary_keys: ['id'],
      foreign_keys: [{ column: 'user_id', references_table: 'users', references_column: 'id' }],
      column_count: 7,
      row_count: 980
    },
    {
      name: 'posts',
      primary_keys: ['id'],
      foreign_keys: [{ column: 'user_id', references_table: 'users', references_column: 'id' }],
      column_count: 7,
      row_count: 3450
    },
    {
      name: 'comments',
      primary_keys: ['id'],
      foreign_keys: [
        { column: 'post_id', references_table: 'posts', references_column: 'id' },
        { column: 'user_id', references_table: 'users', references_column: 'id' }
      ],
      column_count: 5,
      row_count: 8790
    }
  ];

  // Sample database sections
  let databaseSections = [
    { id: 'tables', name: 'Tables', icon: 'table', count: 4 },
    { id: 'views', name: 'Views', icon: 'eye', count: 2 },
    { id: 'functions', name: 'Functions', icon: 'code', count: 8 },
    { id: 'indexes', name: 'Indexes', icon: 'zap', count: 12 },
    { id: 'triggers', name: 'Triggers', icon: 'play', count: 3 },
    { id: 'extensions', name: 'Extensions', icon: 'puzzle', count: 5, badge: 'new' }
  ];

  $: projectId = $page.params.id;

  function handleTableSelect(event) {
    selectedTable = event.detail.tableName;
    toast.success(`Selected table: ${selectedTable}`);
  }

  function handleSectionSelect(event) {
    const section = event.detail.section;
    toast.success(`Navigated to: ${section.name}`);
  }

  function handleBucketSelect(event) {
    const bucket = event.detail.bucket;
    toast.success(`Selected bucket: ${bucket.name}`);
  }

  function handleCellEdit(event) {
    const { rowId, column, oldValue, newValue } = event.detail;
    toast.success(`Updated ${column} from ${oldValue} to ${newValue} for row ${rowId}`);
  }

  function handleRelationshipClick(event) {
    const relationship = event.detail.relationship;
    toast.success(`Selected relationship: ${relationship.from_table} → ${relationship.to_table}`);
  }

  function refreshData() {
    loading = true;
    setTimeout(() => {
      loading = false;
      toast.success('Data refreshed successfully');
    }, 1000);
  }
</script>

<svelte:head>
  <title>Data Visualization - CloudBox</title>
</svelte:head>

<div class="h-full flex flex-col">
  <!-- Header -->
  <div class="flex items-center justify-between p-6 border-b border-border">
    <div class="flex items-center space-x-4">
      <div class="w-10 h-10 bg-primary/10 rounded-lg flex items-center justify-center">
        <Icon name="database" size={20} className="text-primary" />
      </div>
      <div>
        <h1 class="text-2xl font-bold text-foreground">Data Visualization</h1>
        <p class="text-sm text-muted-foreground">
          Advanced database and storage visualization components
        </p>
      </div>
    </div>
    
    <div class="flex items-center space-x-2">
      <Button
        variant="ghost"
        size="icon"
        on:click={refreshData}
        disabled={loading}
        class="hover:rotate-180 transition-transform duration-300"
        title="Refresh data"
      >
        <Icon name={loading ? 'loader' : 'refresh-cw'} size={18} className={loading ? 'animate-spin' : ''} />
      </Button>
    </div>
  </div>

  <!-- View Tabs -->
  <div class="border-b border-border bg-card">
    <nav class="flex space-x-8 px-6">
      {#each [
        { id: 'schema', name: 'Schema Visualizer', icon: 'database' },
        { id: 'data', name: 'Data Table', icon: 'table' },
        { id: 'relationships', name: 'Relationships', icon: 'share-2' },
        { id: 'storage', name: 'Storage Buckets', icon: 'folder' },
        { id: 'navigation', name: 'Database Navigation', icon: 'menu' }
      ] as view}
        <button
          on:click={() => activeView = view.id}
          class="flex items-center space-x-2 py-4 px-1 border-b-2 text-sm font-medium transition-colors {
            activeView === view.id 
              ? 'border-primary text-primary' 
              : 'border-transparent text-muted-foreground hover:text-foreground hover:border-border'
          }"
        >
          <Icon name={view.icon} size={16} />
          <span>{view.name}</span>
        </button>
      {/each}
    </nav>
  </div>

  <!-- Content Area -->
  <div class="flex-1 p-6 overflow-hidden">
    {#if activeView === 'schema'}
      <Card class="h-full">
        <SchemaVisualizer 
          tables={sampleTables}
          {selectedTable}
          on:tableSelect={handleTableSelect}
        />
      </Card>
    {/if}

    {#if activeView === 'data'}
      <Card class="h-full">
        <DataTable
          data={sampleTableData}
          tableName={selectedTable}
          {loading}
          editable={true}
          selectable={true}
          searchable={true}
          paginated={true}
          pageSize={25}
          on:cellEdit={handleCellEdit}
          on:refresh={refreshData}
        />
      </Card>
    {/if}

    {#if activeView === 'relationships'}
      <Card class="h-full">
        <RelationshipDiagram
          relationships={sampleRelationships}
          tables={sampleTableInfo}
          highlightTable={selectedTable}
          on:relationshipClick={handleRelationshipClick}
          on:tableClick={handleTableSelect}
        />
      </Card>
    {/if}

    {#if activeView === 'storage'}
      <Card class="h-full">
        <StorageBucketInterface
          buckets={sampleBuckets}
          {loading}
          showCreateButton={true}
          showSettings={true}
          on:bucketSelect={handleBucketSelect}
          on:createBucket={() => toast.success('Create bucket modal would open')}
          on:bucketSettings={(e) => toast.success(`Settings for ${e.detail.bucket.name}`)}
          on:bucketDelete={(e) => toast.success(`Delete ${e.detail.bucket.name}?`)}
        />
      </Card>
    {/if}

    {#if activeView === 'navigation'}
      <div class="grid grid-cols-1 lg:grid-cols-3 gap-6 h-full">
        <Card class="lg:col-span-1">
          <DatabaseNavigation
            sections={databaseSections}
            activeSection="tables"
            collapsible={true}
            showCounts={true}
            on:sectionSelect={handleSectionSelect}
            on:quickAction={(e) => toast.success(`Quick action: ${e.detail.action}`)}
          />
        </Card>
        
        <Card class="lg:col-span-2">
          <div class="p-6 h-full flex items-center justify-center">
            <div class="text-center">
              <div class="w-16 h-16 bg-primary/10 rounded-lg flex items-center justify-center mx-auto mb-4">
                <Icon name="layout-grid" size={32} className="text-primary" />
              </div>
              <h3 class="text-lg font-medium text-foreground mb-2">Navigation Demo</h3>
              <p class="text-muted-foreground">
                Click on navigation items in the sidebar to see interactions.<br>
                This area would show the selected content in a real application.
              </p>
            </div>
          </div>
        </Card>
      </div>
    {/if}
  </div>

  <!-- Info Panel -->
  <div class="border-t border-border bg-card p-4">
    <div class="flex items-center justify-between text-sm">
      <div class="flex items-center space-x-6">
        <div class="flex items-center space-x-2">
          <Icon name="layers" size={14} className="text-muted-foreground" />
          <span class="text-muted-foreground">Components:</span>
          <Badge variant="outline" class="text-xs">5 Active</Badge>
        </div>
        
        <div class="flex items-center space-x-2">
          <Icon name="database" size={14} className="text-muted-foreground" />
          <span class="text-muted-foreground">Tables:</span>
          <span class="text-foreground font-medium">{sampleTables.length}</span>
        </div>
        
        <div class="flex items-center space-x-2">
          <Icon name="folder" size={14} className="text-muted-foreground" />
          <span class="text-muted-foreground">Storage Buckets:</span>
          <span class="text-foreground font-medium">{sampleBuckets.length}</span>
        </div>
      </div>
      
      <div class="flex items-center space-x-2 text-muted-foreground">
        <Icon name="zap" size={14} />
        <span>All components loaded and interactive</span>
      </div>
    </div>
  </div>
</div>