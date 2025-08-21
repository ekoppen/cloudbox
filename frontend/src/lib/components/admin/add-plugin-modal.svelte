<script lang="ts">
  import { createEventDispatcher } from 'svelte';
  import { fly } from 'svelte/transition';
  import Button from '$lib/components/ui/button.svelte';
  import Input from '$lib/components/ui/input.svelte';
  import Label from '$lib/components/ui/label.svelte';
  import Textarea from '$lib/components/ui/textarea.svelte';
  import Icon from '$lib/components/ui/icon.svelte';
  import { pluginManager } from '$lib/stores/plugins';
  import { addToast } from '$lib/stores/toast';

  export let isOpen = false;

  const dispatch = createEventDispatcher();

  // Form data
  let formData = {
    name: '',
    version: '',
    description: '',
    author: '',
    repository: '',
    license: 'MIT',
    tags: '',
    permissions: '',
    dependencies: '',
    verified: false,
    approved: false,
    type: 'dashboard-plugin',
    main_file: 'index.js'
  };

  let loading = false;
  let errors: Record<string, string> = {};

  function validateForm() {
    errors = {};
    
    if (!formData.name.trim()) {
      errors.name = 'Plugin name is required';
    }
    
    if (!formData.version.trim()) {
      errors.version = 'Version is required';
    }
    
    if (!formData.description.trim()) {
      errors.description = 'Description is required';
    }
    
    if (!formData.author.trim()) {
      errors.author = 'Author is required';
    }
    
    if (!formData.repository.trim()) {
      errors.repository = 'Repository URL is required';
    } else if (!formData.repository.includes('github.com')) {
      errors.repository = 'Must be a GitHub repository URL';
    }
    
    // Validate JSON fields
    if (formData.dependencies.trim()) {
      try {
        JSON.parse(formData.dependencies);
      } catch (e) {
        errors.dependencies = 'Invalid JSON format';
      }
    }
    
    return Object.keys(errors).length === 0;
  }

  async function submitForm() {
    if (!validateForm()) return;
    
    loading = true;
    
    try {
      // Parse optional JSON fields
      let dependencies = {};
      let tags: string[] = [];
      let permissions: string[] = [];
      
      if (formData.dependencies.trim()) {
        dependencies = JSON.parse(formData.dependencies);
      }
      
      if (formData.tags.trim()) {
        tags = formData.tags.split(',').map(tag => tag.trim()).filter(tag => tag);
      }
      
      if (formData.permissions.trim()) {
        permissions = formData.permissions.split(',').map(perm => perm.trim()).filter(perm => perm);
      }
      
      await pluginManager.addPluginToMarketplace({
        name: formData.name.trim(),
        version: formData.version.trim(),
        description: formData.description.trim(),
        author: formData.author.trim(),
        repository: formData.repository.trim(),
        license: formData.license || 'MIT',
        tags,
        permissions,
        dependencies,
        verified: formData.verified,
        approved: formData.approved,
        type: formData.type || 'dashboard-plugin',
        main_file: formData.main_file || 'index.js'
      });
      
      addToast('Plugin added to marketplace successfully!', 'success');
      closeModal();
      
    } catch (error) {
      console.error('Failed to add plugin:', error);
      addToast('Failed to add plugin: ' + (error instanceof Error ? error.message : 'Unknown error'), 'error');
    } finally {
      loading = false;
    }
  }

  function closeModal() {
    // Reset form
    formData = {
      name: '',
      version: '',
      description: '',
      author: '',
      repository: '',
      license: 'MIT',
      tags: '',
      permissions: '',
      dependencies: '',
      verified: false,
      approved: false,
      type: 'dashboard-plugin',
      main_file: 'index.js'
    };
    errors = {};
    dispatch('close');
  }

  // Close on escape key
  function handleKeydown(event: KeyboardEvent) {
    if (event.key === 'Escape' && isOpen) {
      closeModal();
    }
  }
</script>

<svelte:window on:keydown={handleKeydown} />

{#if isOpen}
  <!-- Modal backdrop with proper positioning -->
  <div 
    class="fixed inset-0 bg-black/50 flex items-start justify-center z-50 p-4 pt-16 sm:pt-20 overflow-y-auto"
    role="dialog"
    aria-modal="true"
    aria-labelledby="add-plugin-title"
    tabindex="0"
    transition:fly={{ y: 50, duration: 200 }}
    on:click={closeModal}
    on:keydown={(e) => e.key === 'Escape' && closeModal()}
  >
    <!-- Modal content with better height management -->
    <div 
      class="max-w-2xl w-full bg-card border border-border rounded-lg shadow-lg my-auto min-h-0 flex flex-col"
      style="max-height: calc(100vh - 8rem);"
      role="document"
      on:click|stopPropagation
    >
      <!-- Header -->
      <div class="p-6 border-b border-border">
        <div class="flex items-center justify-between">
          <div class="flex items-center space-x-3">
            <Icon name="plus" size={24} className="text-primary" />
            <div>
              <h2 id="add-plugin-title" class="text-xl font-semibold text-foreground">Add Plugin to Marketplace</h2>
              <p class="text-sm text-muted-foreground">Add a new plugin to the CloudBox marketplace</p>
            </div>
          </div>
          <Button on:click={closeModal} variant="outline" size="sm">
            <Icon name="x" size={16} />
          </Button>
        </div>
      </div>

      <!-- Form with proper scrolling -->
      <div class="p-6 overflow-y-auto flex-1 min-h-0 space-y-6">
        <form on:submit|preventDefault={submitForm} class="space-y-4">
          <!-- Basic Information -->
          <div class="grid grid-cols-1 md:grid-cols-2 gap-4">
            <div>
              <Label for="plugin-name">Plugin Name *</Label>
              <Input 
                id="plugin-name"
                bind:value={formData.name}
                placeholder="my-awesome-plugin"
                class={errors.name ? 'border-red-500' : ''}
              />
              {#if errors.name}
                <p class="text-red-500 text-sm mt-1">{errors.name}</p>
              {/if}
            </div>

            <div>
              <Label for="plugin-version">Version *</Label>
              <Input 
                id="plugin-version"
                bind:value={formData.version}
                placeholder="1.0.0"
                class={errors.version ? 'border-red-500' : ''}
              />
              {#if errors.version}
                <p class="text-red-500 text-sm mt-1">{errors.version}</p>
              {/if}
            </div>
          </div>

          <div>
            <Label for="plugin-description">Description *</Label>
            <Textarea 
              id="plugin-description"
              bind:value={formData.description}
              placeholder="A brief description of what this plugin does..."
              rows="3"
              class={errors.description ? 'border-red-500' : ''}
            />
            {#if errors.description}
              <p class="text-red-500 text-sm mt-1">{errors.description}</p>
            {/if}
          </div>

          <div class="grid grid-cols-1 md:grid-cols-2 gap-4">
            <div>
              <Label for="plugin-author">Author *</Label>
              <Input 
                id="plugin-author"
                bind:value={formData.author}
                placeholder="John Doe"
                class={errors.author ? 'border-red-500' : ''}
              />
              {#if errors.author}
                <p class="text-red-500 text-sm mt-1">{errors.author}</p>
              {/if}
            </div>

            <div>
              <Label for="plugin-license">License</Label>
              <Input 
                id="plugin-license"
                bind:value={formData.license}
                placeholder="MIT"
              />
            </div>
          </div>

          <div>
            <Label for="plugin-repository">GitHub Repository *</Label>
            <Input 
              id="plugin-repository"
              bind:value={formData.repository}
              placeholder="https://github.com/username/my-plugin"
              class={errors.repository ? 'border-red-500' : ''}
            />
            {#if errors.repository}
              <p class="text-red-500 text-sm mt-1">{errors.repository}</p>
            {/if}
          </div>

          <!-- Plugin Configuration -->
          <div class="grid grid-cols-1 md:grid-cols-2 gap-4">
            <div>
              <Label for="plugin-type">Plugin Type</Label>
              <select 
                id="plugin-type"
                bind:value={formData.type}
                class="w-full px-3 py-2 border border-border rounded-md bg-background text-foreground"
              >
                <option value="dashboard-plugin">Dashboard Plugin</option>
                <option value="project-plugin">Project Plugin</option>
                <option value="system-plugin">System Plugin</option>
              </select>
            </div>

            <div>
              <Label for="plugin-main-file">Main File</Label>
              <Input 
                id="plugin-main-file"
                bind:value={formData.main_file}
                placeholder="index.js"
              />
            </div>
          </div>

          <div>
            <Label for="plugin-tags">Tags (comma-separated)</Label>
            <Input 
              id="plugin-tags"
              bind:value={formData.tags}
              placeholder="utility, database, automation"
            />
            <p class="text-xs text-muted-foreground mt-1">Separate multiple tags with commas</p>
          </div>

          <div>
            <Label for="plugin-permissions">Permissions (comma-separated)</Label>
            <Textarea 
              id="plugin-permissions"
              bind:value={formData.permissions}
              placeholder="database:read, database:write, functions:deploy"
              rows="2"
            />
            <p class="text-xs text-muted-foreground mt-1">Separate multiple permissions with commas</p>
          </div>

          <div>
            <Label for="plugin-dependencies">Dependencies (JSON)</Label>
            <Textarea 
              id="plugin-dependencies"
              bind:value={formData.dependencies}
              placeholder="JSON format: cloudbox-sdk version, express version"
              rows="3"
              class={errors.dependencies ? 'border-red-500' : ''}
            />
            {#if errors.dependencies}
              <p class="text-red-500 text-sm mt-1">{errors.dependencies}</p>
            {/if}
            <p class="text-xs text-muted-foreground mt-1">JSON object of dependencies</p>
          </div>

          <!-- Status Checkboxes -->
          <div class="flex items-center space-x-6">
            <div class="flex items-center space-x-2">
              <input 
                type="checkbox" 
                id="plugin-verified"
                bind:checked={formData.verified}
                class="rounded border-border"
              />
              <Label for="plugin-verified">Verified Plugin</Label>
            </div>

            <div class="flex items-center space-x-2">
              <input 
                type="checkbox" 
                id="plugin-approved"
                bind:checked={formData.approved}
                class="rounded border-border"
              />
              <Label for="plugin-approved">Approved Plugin</Label>
            </div>
          </div>
        </form>
      </div>

      <!-- Footer -->
      <div class="p-6 border-t border-border flex justify-between">
        <Button on:click={closeModal} variant="outline">
          Cancel
        </Button>
        
        <Button 
          on:click={submitForm}
          disabled={loading}
        >
          {#if loading}
            <div class="animate-spin rounded-full h-4 w-4 border-b-2 border-white mr-2"></div>
            Adding Plugin...
          {:else}
            <Icon name="plus" size={16} className="mr-2" />
            Add to Marketplace
          {/if}
        </Button>
      </div>
    </div>
  </div>
{/if}