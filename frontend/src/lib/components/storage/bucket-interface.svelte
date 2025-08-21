<script lang="ts">
  import { createEventDispatcher } from 'svelte';
  import Card from '$lib/components/ui/card.svelte';
  import Button from '$lib/components/ui/button.svelte';
  import Badge from '$lib/components/ui/badge.svelte';
  import Icon from '$lib/components/ui/icon.svelte';

  const dispatch = createEventDispatcher();

  interface StorageBucket {
    id: string;
    name: string;
    description?: string;
    is_public: boolean;
    max_file_size: number;
    file_count: number;
    total_size: number;
    created_at: string;
    updated_at?: string;
  }

  export let buckets: StorageBucket[] = [];
  export let selectedBucket: StorageBucket | null = null;
  export let loading: boolean = false;
  export let showCreateButton: boolean = true;
  export let showSettings: boolean = true;

  function formatFileSize(bytes: number): string {
    if (bytes === 0) return '0 B';
    const k = 1024;
    const sizes = ['B', 'KB', 'MB', 'GB', 'TB'];
    const i = Math.floor(Math.log(bytes) / Math.log(k));
    return parseFloat((bytes / Math.pow(k, i)).toFixed(1)) + ' ' + sizes[i];
  }

  function formatDate(dateStr: string): string {
    return new Date(dateStr).toLocaleDateString('en-US', {
      month: 'short',
      day: 'numeric',
      year: 'numeric'
    });
  }

  function getBucketStatusColor(bucket: StorageBucket): string {
    if (bucket.file_count === 0) return 'bg-gray-100 dark:bg-gray-900 text-gray-600 dark:text-gray-400';
    if (bucket.is_public) return 'bg-green-100 dark:bg-green-900 text-green-600 dark:text-green-400';
    return 'bg-blue-100 dark:bg-gray-800 text-blue-600 dark:text-blue-400';
  }

  function handleBucketSelect(bucket: StorageBucket) {
    selectedBucket = bucket;
    dispatch('bucketSelect', { bucket });
  }

  function handleCreateBucket() {
    dispatch('createBucket');
  }

  function handleBucketSettings(bucket: StorageBucket) {
    dispatch('bucketSettings', { bucket });
  }

  function handleBucketDelete(bucket: StorageBucket) {
    dispatch('bucketDelete', { bucket });
  }

  function getUsagePercentage(bucket: StorageBucket): number {
    if (bucket.max_file_size === 0) return 0;
    return Math.min(100, (bucket.total_size / bucket.max_file_size) * 100);
  }
</script>

<div class="h-full flex flex-col">
  <!-- Header -->
  <div class="flex items-center justify-between p-4 border-b border-border bg-card">
    <div class="flex items-center space-x-2">
      <Icon name="folder" size={16} className="text-primary" />
      <h3 class="text-sm font-semibold text-foreground">Storage Buckets</h3>
      
      {#if buckets.length > 0}
        <Badge variant="secondary" class="text-xs">
          {buckets.length} bucket{buckets.length === 1 ? '' : 's'}
        </Badge>
      {/if}
    </div>
    
    {#if showCreateButton}
      <Button
        size="sm"
        on:click={handleCreateBucket}
        class="flex items-center space-x-2"
      >
        <Icon name="plus" size={14} />
        <span>New Bucket</span>
      </Button>
    {/if}
  </div>

  <!-- Bucket List -->
  <div class="flex-1 overflow-auto">
    {#if loading}
      <div class="flex items-center justify-center h-32">
        <div class="text-center">
          <div class="animate-spin rounded-full h-8 w-8 border-b-2 border-primary mx-auto"></div>
          <p class="mt-2 text-sm text-muted-foreground">Loading buckets...</p>
        </div>
      </div>
    {:else if buckets.length === 0}
      <!-- Empty State -->
      <div class="flex items-center justify-center h-full min-h-[300px]">
        <div class="text-center max-w-md mx-auto px-4">
          <div class="w-16 h-16 bg-muted rounded-lg flex items-center justify-center mx-auto mb-4">
            <Icon name="folder" size={32} className="text-muted-foreground" />
          </div>
          <h3 class="text-lg font-medium text-foreground mb-2">No storage buckets</h3>
          <p class="text-muted-foreground mb-6 text-sm">
            Create your first storage bucket to organize and manage your files.
            Buckets help you organize files and control access permissions.
          </p>
          {#if showCreateButton}
            <Button on:click={handleCreateBucket} class="flex items-center space-x-2 mx-auto">
              <Icon name="plus" size={16} />
              <span>Create Your First Bucket</span>
            </Button>
          {/if}
        </div>
      </div>
    {:else}
      <!-- Bucket Grid -->
      <div class="p-4 space-y-3">
        {#each buckets as bucket}
          <Card 
            class="p-4 cursor-pointer transition-all hover:shadow-md border-2 {selectedBucket?.id === bucket.id ? 'border-primary ring-2 ring-primary/20 bg-primary/5' : 'hover:border-border'}"
            on:click={() => handleBucketSelect(bucket)}
          >
            <div class="flex items-start justify-between">
              <!-- Bucket Info -->
              <div class="flex items-start space-x-3 flex-1">
                <div class="w-10 h-10 bg-primary/10 rounded-lg flex items-center justify-center flex-shrink-0">
                  <Icon 
                    name={bucket.is_public ? 'globe' : 'folder'} 
                    size={20} 
                    className={bucket.is_public ? 'text-green-600' : 'text-primary'} 
                  />
                </div>
                
                <div class="flex-1 min-w-0">
                  <div class="flex items-center space-x-2 mb-1">
                    <h4 class="text-sm font-semibold text-foreground truncate">{bucket.name}</h4>
                    
                    <Badge 
                      variant="outline" 
                      class="text-xs px-1.5 py-0.5 {getBucketStatusColor(bucket)}"
                    >
                      {bucket.is_public ? 'Public' : 'Private'}
                    </Badge>
                    
                    {#if bucket.file_count === 0}
                      <Badge variant="outline" class="text-xs px-1.5 py-0.5 text-muted-foreground">
                        Empty
                      </Badge>
                    {/if}
                  </div>
                  
                  {#if bucket.description}
                    <p class="text-xs text-muted-foreground mb-2 truncate">{bucket.description}</p>
                  {/if}
                  
                  <div class="flex items-center space-x-4 text-xs text-muted-foreground">
                    <div class="flex items-center space-x-1">
                      <Icon name="file" size={12} />
                      <span>{bucket.file_count.toLocaleString()} files</span>
                    </div>
                    
                    <div class="flex items-center space-x-1">
                      <Icon name="hard-drive" size={12} />
                      <span>{formatFileSize(bucket.total_size)}</span>
                    </div>
                    
                    <div class="flex items-center space-x-1">
                      <Icon name="calendar" size={12} />
                      <span>Created {formatDate(bucket.created_at)}</span>
                    </div>
                  </div>
                  
                  <!-- Usage Bar -->
                  {#if bucket.max_file_size > 0}
                    {@const usagePercent = getUsagePercentage(bucket)}
                    <div class="mt-3">
                      <div class="flex items-center justify-between text-xs text-muted-foreground mb-1">
                        <span>Storage Usage</span>
                        <span>{usagePercent.toFixed(1)}%</span>
                      </div>
                      <div class="w-full bg-gray-200 dark:bg-gray-800 rounded-full h-2">
                        <div 
                          class="h-2 rounded-full transition-all {
                            usagePercent > 90 ? 'bg-red-500' :
                            usagePercent > 75 ? 'bg-yellow-500' :
                            'bg-primary'
                          }"
                          style="width: {usagePercent}%"
                        ></div>
                      </div>
                      <div class="flex justify-between text-xs text-muted-foreground mt-1">
                        <span>{formatFileSize(bucket.total_size)}</span>
                        <span>{formatFileSize(bucket.max_file_size)}</span>
                      </div>
                    </div>
                  {/if}
                </div>
              </div>
              
              <!-- Actions -->
              <div class="flex items-center space-x-1 opacity-0 hover:opacity-100 transition-opacity">
                {#if showSettings}
                  <Button
                    variant="ghost"
                    size="sm"
                    class="h-8 w-8 p-0"
                    on:click={() => handleBucketSettings(bucket)}
                    title="Bucket settings"
                  >
                    <Icon name="settings" size={14} />
                  </Button>
                {/if}
                
                <Button
                  variant="ghost"
                  size="sm"
                  class="h-8 w-8 p-0 text-destructive hover:text-destructive hover:bg-destructive/10"
                  on:click={() => handleBucketDelete(bucket)}
                  title="Delete bucket"
                >
                  <Icon name="trash" size={14} />
                </Button>
              </div>
            </div>
          </Card>
        {/each}
      </div>
    {/if}
  </div>

  <!-- Summary Footer -->
  {#if buckets.length > 0}
    <div class="border-t border-border bg-card px-4 py-3">
      <div class="flex items-center justify-between text-xs text-muted-foreground">
        <div class="flex items-center space-x-4">
          <span>Total: {buckets.reduce((sum, bucket) => sum + bucket.file_count, 0).toLocaleString()} files</span>
          <span>Size: {formatFileSize(buckets.reduce((sum, bucket) => sum + bucket.total_size, 0))}</span>
        </div>
        
        <div class="flex items-center space-x-2">
          {#if buckets.filter(b => b.is_public).length > 0}
            <Badge variant="outline" class="text-xs px-1.5 py-0.5 text-green-600">
              {buckets.filter(b => b.is_public).length} public
            </Badge>
          {/if}
          
          {#if buckets.filter(b => !b.is_public).length > 0}
            <Badge variant="outline" class="text-xs px-1.5 py-0.5 text-blue-600">
              {buckets.filter(b => !b.is_public).length} private
            </Badge>
          {/if}
        </div>
      </div>
    </div>
  {/if}
</div>

