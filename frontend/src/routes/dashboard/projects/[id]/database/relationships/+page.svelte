<script lang="ts">
  import { page } from '$app/stores';
  import { onMount } from 'svelte';
  import { auth } from '$lib/stores/auth';
  import { API_ENDPOINTS, createApiRequest } from '$lib/config';
  import Card from '$lib/components/ui/card.svelte';
  import Button from '$lib/components/ui/button.svelte';
  import Badge from '$lib/components/ui/badge.svelte';
  import Icon from '$lib/components/ui/icon.svelte';
  import RelationshipDiagram from '$lib/components/database/relationship-diagram.svelte';

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

  interface Project {
    id: number;
    slug: string;
    name: string;
  }

  let project: Project | null = null;
  let relationships: TableRelationship[] = [];
  let tables: TableInfo[] = [];
  let selectedRelationship: string | null = null;
  let highlightTable: string | null = null;
  let loading = true;
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
        await loadRelationships();
      } else {
        error = 'Project niet gevonden';
      }
    } catch (err) {
      error = 'Fout bij laden van project';
      console.error('Load project error:', err);
    }
  }

  async function loadRelationships() {
    if (!project) return;
    
    loading = true;
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
        const relationshipData: TableRelationship[] = [];
        const tableData: TableInfo[] = [];

        // Process each collection to build relationships
        for (const collection of collections) {
          try {
            // Get documents to analyze structure
            const documentsResponse = await createApiRequest(API_ENDPOINTS.admin.projects.collections.documents.list(project.id.toString(), collection.name), {
              headers: {
                'Authorization': `Bearer ${$auth.token}`,
                'Content-Type': 'application/json',
              },
            });

            let documentCount = 0;
            let foreignKeys: TableInfo['foreign_keys'] = [];

            if (documentsResponse.ok) {
              const documentsData = await documentsResponse.json();
              const documents = documentsData.documents || [];
              documentCount = documents.length;

              // Analyze document structure for relationships
              if (documents.length > 0) {
                const sampleDoc = documents[0];
                const docData = sampleDoc.data || sampleDoc;

                // PhotoPortfolio specific relationship detection
                for (const [key, value] of Object.entries(docData)) {
                  const foreignKey = detectForeignKey(key, value, collections);
                  if (foreignKey) {
                    foreignKeys.push({
                      column: key,
                      references_table: foreignKey.table,
                      references_column: foreignKey.column
                    });

                    // Create relationship record
                    relationshipData.push({
                      id: `${collection.name}_${key}_${foreignKey.table}_${foreignKey.column}`,
                      from_table: collection.name,
                      from_column: key,
                      to_table: foreignKey.table,
                      to_column: foreignKey.column,
                      constraint_name: `fk_${collection.name}_${key}`,
                      relationship_type: Array.isArray(value) ? 'one-to-many' : 'one-to-one',
                      on_delete: 'CASCADE',
                      on_update: 'CASCADE'
                    });
                  }
                }
              }
            }

            // Calculate column count (basic fields + system fields)
            const baseColumns = ['id', 'created_at', 'updated_at'];
            const documentFields = documents.length > 0 ? 
              Object.keys(documents[0].data || documents[0]).filter(k => k !== 'id') : [];
            const columnCount = baseColumns.length + documentFields.length;

            tableData.push({
              name: collection.name,
              primary_keys: ['id'],
              foreign_keys: foreignKeys,
              column_count: columnCount,
              row_count: documentCount
            });

          } catch (err) {
            console.warn(`Failed to analyze relationships for collection ${collection.name}:`, err);
            // Add basic table info for failed collections
            tableData.push({
              name: collection.name,
              primary_keys: ['id'],
              foreign_keys: [],
              column_count: 3, // id, created_at, updated_at
              row_count: 0
            });
          }
        }

        relationships = relationshipData;
        tables = tableData;
        
      } else {
        error = 'Fout bij laden van relaties';
        console.error('Relationships response not OK:', response.status, response.statusText);
      }
    } catch (err) {
      error = 'Fout bij laden van relaties';
      console.error('Load relationships error:', err);
    } finally {
      loading = false;
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
    
    if (fieldName === 'cover_photo' && collectionNames.includes('images')) {
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

  function handleRelationshipClick(event: CustomEvent) {
    const relationship = event.detail.relationship;
    selectedRelationship = selectedRelationship === relationship.id ? null : relationship.id;
  }

  function handleTableClick(event: CustomEvent) {
    highlightTable = highlightTable === event.detail.tableName ? null : event.detail.tableName;
  }

  function exportRelationships() {
    const relationshipsExport = {
      project: project?.name,
      generated_at: new Date().toISOString(),
      tables: tables.length,
      relationships: relationships.length,
      data: {
        tables,
        relationships
      }
    };

    const blob = new Blob([JSON.stringify(relationshipsExport, null, 2)], { type: 'application/json' });
    const url = URL.createObjectURL(blob);
    const a = document.createElement('a');
    a.href = url;
    a.download = `${project?.slug || 'relationships'}-relationships.json`;
    document.body.appendChild(a);
    a.click();
    document.body.removeChild(a);
    URL.revokeObjectURL(url);
  }
</script>

<svelte:head>
  <title>Database Relations - {project?.name || 'CloudBox'}</title>
</svelte:head>

<div class="h-full flex flex-col space-y-6">
  <!-- Header -->
  <div class="flex items-center justify-between">
    <div class="flex items-center space-x-4">
      <div class="w-10 h-10 bg-primary/10 rounded-lg flex items-center justify-center">
        <Icon name="share-2" size={20} className="text-primary" />
      </div>
      <div>
        <h1 class="text-2xl font-bold text-foreground">Database Relations</h1>
        <p class="text-sm text-muted-foreground">
          Visualiseer relaties tussen je database collections
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
        on:click={exportRelationships}
        variant="outline"
        size="md"
        disabled={relationships.length === 0}
        class="flex items-center space-x-2"
      >
        <Icon name="backup" size={16} />
        <span>Relaties Exporteren</span>
      </Button>
      <Button 
        on:click={loadRelationships}
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
        <p class="mt-4 text-muted-foreground">Relaties laden...</p>
      </Card>
    </div>
  {:else if error}
    <div class="flex-1 flex items-center justify-center">
      <Card class="glassmorphism-content p-12 text-center">
        <div class="w-16 h-16 bg-destructive/10 rounded-lg flex items-center justify-center mx-auto mb-4">
          <Icon name="backup" size={32} className="text-destructive" />
        </div>
        <h3 class="text-lg font-medium text-foreground mb-2">Fout bij laden van relaties</h3>
        <p class="text-muted-foreground mb-4">{error}</p>
        <Button on:click={loadRelationships} class="flex items-center space-x-2">
          <Icon name="refresh" size={16} />
          <span>Opnieuw proberen</span>
        </Button>
      </Card>
    </div>
  {:else}
    <!-- Relationships Statistics -->
    <div class="grid grid-cols-1 md:grid-cols-4 gap-4">
      <Card class="glassmorphism-card p-4">
        <div class="flex items-center space-x-3">
          <div class="w-10 h-10 bg-blue-100 dark:bg-gray-800/30 rounded-lg flex items-center justify-center">
            <Icon name="database" size={20} className="text-blue-600 dark:text-blue-400" />
          </div>
          <div>
            <p class="text-2xl font-bold text-foreground">{tables.length}</p>
            <p class="text-sm text-muted-foreground">Tabellen</p>
          </div>
        </div>
      </Card>

      <Card class="glassmorphism-card p-4">
        <div class="flex items-center space-x-3">
          <div class="w-10 h-10 bg-green-100 dark:bg-green-900/30 rounded-lg flex items-center justify-center">
            <Icon name="share-2" size={20} className="text-green-600 dark:text-green-400" />
          </div>
          <div>
            <p class="text-2xl font-bold text-foreground">{relationships.length}</p>
            <p class="text-sm text-muted-foreground">Relaties</p>
          </div>
        </div>
      </Card>

      <Card class="glassmorphism-card p-4">
        <div class="flex items-center space-x-3">
          <div class="w-10 h-10 bg-purple-100 dark:bg-purple-900/30 rounded-lg flex items-center justify-center">
            <Icon name="key" size={20} className="text-purple-600 dark:text-purple-400" />
          </div>
          <div>
            <p class="text-2xl font-bold text-foreground">{tables.reduce((sum, t) => sum + t.foreign_keys.length, 0)}</p>
            <p class="text-sm text-muted-foreground">Foreign Keys</p>
          </div>
        </div>
      </Card>

      <Card class="glassmorphism-card p-4">
        <div class="flex items-center space-x-3">
          <div class="w-10 h-10 bg-yellow-100 dark:bg-yellow-900/30 rounded-lg flex items-center justify-center">
            <Icon name="storage" size={20} className="text-yellow-600 dark:text-yellow-400" />
          </div>
          <div>
            <p class="text-2xl font-bold text-foreground">{tables.reduce((sum, t) => sum + t.row_count, 0).toLocaleString()}</p>
            <p class="text-sm text-muted-foreground">Totaal Records</p>
          </div>
        </div>
      </Card>
    </div>

    <!-- Relationship Diagram -->
    <div class="flex-1 min-h-0">
      {#if relationships.length === 0}
        <Card class="glassmorphism-content p-12 text-center h-full flex items-center justify-center">
          <div>
            <div class="w-16 h-16 bg-muted rounded-lg flex items-center justify-center mx-auto mb-4">
              <Icon name="share-2" size={32} className="text-muted-foreground" />
            </div>
            <h3 class="text-lg font-medium text-foreground mb-2">Geen relaties gevonden</h3>
            <p class="text-muted-foreground mb-4">
              Er zijn nog geen relaties tussen de collections gedetecteerd.<br>
              Voeg foreign key velden toe om relaties te visualiseren.
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
          <RelationshipDiagram
            {relationships}
            {tables}
            {selectedRelationship}
            {highlightTable}
            on:relationshipClick={handleRelationshipClick}
            on:tableClick={handleTableClick}
          />
        </Card>
      {/if}
    </div>

    <!-- Selected Relationship Details -->
    {#if selectedRelationship && relationships.find(r => r.id === selectedRelationship)}
      {@const relationship = relationships.find(r => r.id === selectedRelationship)}
      <Card class="glassmorphism-card">
        <div class="px-6 py-4 border-b border-border">
          <div class="flex items-center justify-between">
            <div class="flex items-center space-x-3">
              <div class="w-8 h-8 bg-primary/10 rounded-lg flex items-center justify-center">
                <Icon name="link" size={16} className="text-primary" />
              </div>
              <div>
                <h3 class="text-lg font-semibold text-foreground">
                  {relationship.from_table} → {relationship.to_table}
                </h3>
                <p class="text-sm text-muted-foreground">
                  {relationship.from_column} verwijst naar {relationship.to_column}
                </p>
              </div>
            </div>
            <Badge variant="outline" class="text-sm px-3 py-1">
              {relationship.relationship_type}
            </Badge>
          </div>
        </div>
        
        <div class="p-6 grid grid-cols-1 md:grid-cols-2 gap-6">
          <div>
            <h4 class="text-sm font-medium text-foreground mb-3">Constraint Details</h4>
            <div class="space-y-2 text-sm">
              <div class="flex justify-between">
                <span class="text-muted-foreground">Constraint Name:</span>
                <span class="text-foreground">{relationship.constraint_name}</span>
              </div>
              <div class="flex justify-between">
                <span class="text-muted-foreground">On Delete:</span>
                <span class="text-foreground">{relationship.on_delete}</span>
              </div>
              <div class="flex justify-between">
                <span class="text-muted-foreground">On Update:</span>
                <span class="text-foreground">{relationship.on_update}</span>
              </div>
            </div>
          </div>
          
          <div>
            <h4 class="text-sm font-medium text-foreground mb-3">Table Information</h4>
            <div class="space-y-2 text-sm">
              {#each [relationship] as rel}
                {@const fromTable = tables.find(t => t.name === rel.from_table)}
                {@const toTable = tables.find(t => t.name === rel.to_table)}
                
                <div class="flex justify-between">
                  <span class="text-muted-foreground">From Table Records:</span>
                  <span class="text-foreground">{fromTable?.row_count?.toLocaleString() || 0}</span>
                </div>
                <div class="flex justify-between">
                  <span class="text-muted-foreground">To Table Records:</span>
                  <span class="text-foreground">{toTable?.row_count?.toLocaleString() || 0}</span>
                </div>
                <div class="flex justify-between">
                  <span class="text-muted-foreground">Relationship Type:</span>
                  <span class="text-foreground capitalize">{rel.relationship_type.replace('-', ' ')}</span>
                </div>
              {/each}
            </div>
          </div>
        </div>
      </Card>
    {/if}
  {/if}
</div>