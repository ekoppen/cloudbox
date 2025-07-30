<script lang="ts">
  import { page } from '$app/stores';
  import { onMount } from 'svelte';
  import { auth } from '$lib/stores/auth';
  import { toast } from '$lib/stores/toast';
  import { API_ENDPOINTS, createApiRequest } from '$lib/config';
  import Card from '$lib/components/ui/card.svelte';
  import Button from '$lib/components/ui/button.svelte';
  import Input from '$lib/components/ui/input.svelte';
  import Label from '$lib/components/ui/label.svelte';
  import Badge from '$lib/components/ui/badge.svelte';
  import Icon from '$lib/components/ui/icon.svelte';

  interface Project {
    id: number;
    name: string;
    description: string;
    slug: string;
    created_at: string;
    is_active: boolean;
  }

  interface StorageBucket {
    id: string;
    name: string;
    description: string;
    is_public: boolean;
    max_file_size: number;
    allowed_types: string[];
    file_count: number;
    total_size: number;
    created_at: string;
  }

  interface StorageFile {
    id: string;
    original_name: string;
    file_name: string;
    file_path: string;
    folder_path?: string;
    mime_type: string;
    size: number;
    bucket_name: string;
    is_public: boolean;
    public_url?: string;
    private_url?: string;
    created_at: string;
  }

  interface TreeNode {
    name: string;
    path: string;
    type: 'folder' | 'file';
    children?: TreeNode[];
    file?: StorageFile;
    expanded?: boolean;
  }

  interface BreadcrumbItem {
    name: string;
    path: string;
  }

  let buckets: StorageBucket[] = [];
  let files: StorageFile[] = [];
  let folders: any[] = [];
  let currentBucket: StorageBucket | null = null;
  let currentPath = '';
  let breadcrumbs: BreadcrumbItem[] = [];
  let loading = true;
  let loadingFiles = false;
  let backendAvailable = true;
  let selectedFiles: string[] = [];
  let showNewBucketModal = false;
  let showUploadModal = false;
  let showNewFolderModal = false;
  let newFolderName = '';
  let uploadFiles: FileList | null = null;
  let uploading = false;
  let sortBy: 'name' | 'size' | 'date' = 'name';
  let sortOrder: 'asc' | 'desc' = 'asc';
  let selectedFolder = '';

  let newBucket = {
    name: '',
    description: '',
    is_public: false,
    max_file_size: 50 * 1024 * 1024, // 50MB
    allowed_types: ['image/jpeg', 'image/png', 'image/gif', 'application/pdf', 'text/plain']
  };

  // Settings for current bucket
  let bucketSettings = {
    description: '',
    max_file_size: 50,
    allowed_types: 'image/jpeg,image/png,image/gif,application/pdf,text/plain',
    is_public: false
  };
  let savingSettings = false;

  let activeTab = 'files';
  let treeData: TreeNode[] = [];
  let totalFilesInTree = 0;
  let draggedFile: StorageFile | null = null;
  let dropTarget: string | null = null;
  
  // Get project data from parent layout context (for display purposes)
  let project: Project | null = null;
  let projectLoading = true;

  $: projectId = $page.params.id;
  $: sortedFiles = sortFiles(files, sortBy, sortOrder);
  $: filteredFiles = filterFiles(sortedFiles, currentPath);
  $: if (currentBucket) {
    buildTreeView();
  }

  onMount(() => {
    // Load project data for display and buckets in parallel
    loadProject();
    loadBuckets();
  });

  async function loadProject() {
    projectLoading = true;
    try {
      const response = await createApiRequest(API_ENDPOINTS.admin.projects.get(projectId), {
        headers: {
          'Authorization': `Bearer ${$auth.token}`,
        },
      });

      if (response.ok) {
        project = await response.json();
      } else {
        console.error('Failed to load project:', response.status);
        toast.error('Fout bij laden van project gegevens');
      }
    } catch (error) {
      console.error('Error loading project:', error);
      toast.error('Netwerkfout bij laden van project gegevens');
    } finally {
      projectLoading = false;
    }
  }

  async function loadBuckets() {
    loading = true;
    try {
      const bucketsUrl = API_ENDPOINTS.admin.projects.storage.buckets.list(projectId);
      console.log('🔍 Loading buckets URL:', bucketsUrl);
      const response = await createApiRequest(bucketsUrl, {
        headers: {
          'Authorization': `Bearer ${$auth.token}`,
          'Content-Type': 'application/json',
        },
      });

      if (response.ok) {
        buckets = await response.json();
        if (buckets.length > 0 && !currentBucket) {
          await selectBucket(buckets[0]);
        }
      } else {
        console.error('Failed to load buckets:', response.status);
        if (response.status === 404 || response.status === 500) {
          backendAvailable = false;
        }
      }
    } catch (error) {
      console.error('Error loading buckets:', error);
      if (error.message.includes('Failed to fetch')) {
        backendAvailable = false;
      }
    } finally {
      loading = false;
    }
  }

  async function loadFiles(bucketName: string, path: string = '') {
    if (!currentBucket) return;
    
    loadingFiles = true;
    try {
      const fileUrl = API_ENDPOINTS.admin.projects.storage.files.list(projectId, bucketName);
      console.log('🔍 Loading files URL:', fileUrl);
      const response = await createApiRequest(`${fileUrl}?path=${encodeURIComponent(path)}`, {
        headers: {
          'Authorization': `Bearer ${$auth.token}`,
          'Content-Type': 'application/json',
        },
      });

      if (response.ok) {
        const data = await response.json();
        files = data.files || [];
        updateBreadcrumbs(path);
      } else {
        console.error('Failed to load files:', response.status);
        files = [];
      }
    } catch (error) {
      console.error('Error loading files:', error);
      files = [];
    } finally {
      loadingFiles = false;
    }
  }

  async function loadFolders(path: string = '') {
    if (!currentBucket) return;
    
    try {
      const folderUrl = API_ENDPOINTS.admin.projects.storage.folders.list(projectId, currentBucket.name);
      console.log('🔍 Loading folders URL:', folderUrl);
      const response = await createApiRequest(`${folderUrl}?path=${encodeURIComponent(path)}`, {
        headers: {
          'Authorization': `Bearer ${$auth.token}`,
          'Content-Type': 'application/json',
        },
      });

      if (response.ok) {
        const data = await response.json();
        folders = data.folders || [];
      } else {
        console.error('Failed to load folders:', response.status);
        folders = [];
      }
    } catch (error) {
      console.error('Error loading folders:', error);
      folders = [];
    }
  }

  async function selectBucket(bucket: StorageBucket) {
    currentBucket = bucket;
    currentPath = '';
    await Promise.all([
      loadFiles(bucket.name),
      loadFolders()
    ]);
    loadBucketSettings(bucket);
  }
  
  function loadBucketSettings(bucket: StorageBucket) {
    if (!bucket) return;
    
    bucketSettings = {
      description: bucket.description || '',
      max_file_size: Math.round(bucket.max_file_size / (1024 * 1024)), // Convert to MB
      allowed_types: Array.isArray(bucket.allowed_types) ? bucket.allowed_types.join(',') : '',
      is_public: bucket.is_public || false
    };
  }
  
  async function saveBucketSettings() {
    if (!currentBucket) return;
    
    savingSettings = true;
    try {
      const allowedTypesArray = bucketSettings.allowed_types
        .split(',')
        .map(type => type.trim())
        .filter(type => type.length > 0);
        
      const response = await createApiRequest(API_ENDPOINTS.admin.projects.storage.buckets.update(projectId, currentBucket.name), {
        method: 'PUT',
        headers: {
          'Authorization': `Bearer ${$auth.token}`,
          'Content-Type': 'application/json',
        },
        body: JSON.stringify({
          description: bucketSettings.description,
          max_file_size: bucketSettings.max_file_size * 1024 * 1024, // Convert MB to bytes
          allowed_types: allowedTypesArray,
          is_public: bucketSettings.is_public
        }),
      });
      
      if (response.ok) {
        const updatedBucket = await response.json();
        
        // Update the bucket in our buckets array
        buckets = buckets.map(b => b.name === updatedBucket.name ? updatedBucket : b);
        currentBucket = updatedBucket;
        
        toast.success('Bucket instellingen opgeslagen');
      } else {
        const error = await response.json();
        toast.error(error.error || 'Fout bij opslaan instellingen');
      }
    } catch (error) {
      console.error('Save settings error:', error);
      toast.error('Netwerkfout bij opslaan instellingen');
    } finally {
      savingSettings = false;
    }
  }

  function updateBreadcrumbs(path: string) {
    breadcrumbs = [{ name: currentBucket?.name || 'Root', path: '' }];
    
    if (path) {
      const pathParts = path.split('/').filter(Boolean);
      let currentCrumbPath = '';
      
      for (const part of pathParts) {
        currentCrumbPath += (currentCrumbPath ? '/' : '') + part;
        breadcrumbs.push({
          name: part,
          path: currentCrumbPath
        });
      }
    }
  }

  async function navigateToPath(path: string) {
    if (!currentBucket) return;
    currentPath = path;
    await Promise.all([
      loadFiles(currentBucket.name, path),
      loadFolders(path)
    ]);
  }

  function formatFileSize(bytes: number): string {
    if (bytes === 0) return '0 B';
    const k = 1024;
    const sizes = ['B', 'KB', 'MB', 'GB'];
    const i = Math.floor(Math.log(bytes) / Math.log(k));
    return parseFloat((bytes / Math.pow(k, i)).toFixed(2)) + ' ' + sizes[i];
  }

  function getFileIcon(mimeType: string): string {
    if (mimeType.startsWith('image/')) return 'image';
    if (mimeType.includes('pdf')) return 'file-text';
    if (mimeType.startsWith('video/')) return 'video';
    if (mimeType.startsWith('audio/')) return 'music';
    if (mimeType.includes('zip') || mimeType.includes('archive')) return 'archive';
    if (mimeType.includes('text') || mimeType.includes('document')) return 'file-text';
    return 'file';
  }

  function formatDate(dateStr: string): string {
    return new Date(dateStr).toLocaleDateString('nl-NL', {
      year: 'numeric',
      month: 'short',
      day: 'numeric',
      hour: '2-digit',
      minute: '2-digit'
    });
  }

  function sortFiles(files: StorageFile[], sortBy: string, sortOrder: string): StorageFile[] {
    // Ensure files is always an array to prevent "not iterable" errors
    if (!files || !Array.isArray(files)) {
      return [];
    }
    
    return [...files].sort((a, b) => {
      let aValue: string | number;
      let bValue: string | number;

      switch (sortBy) {
        case 'name':
          aValue = a.original_name.toLowerCase();
          bValue = b.original_name.toLowerCase();
          break;
        case 'size':
          aValue = a.size;
          bValue = b.size;
          break;
        case 'date':
          aValue = new Date(a.created_at).getTime();
          bValue = new Date(b.created_at).getTime();
          break;
        default:
          return 0;
      }

      if (aValue < bValue) return sortOrder === 'asc' ? -1 : 1;
      if (aValue > bValue) return sortOrder === 'asc' ? 1 : -1;
      return 0;
    });
  }

  function filterFiles(files: StorageFile[], path: string): StorageFile[] {
    // Ensure files is always an array to prevent errors
    if (!files || !Array.isArray(files)) {
      return [];
    }
    
    // For now, just return all files
    // In a real implementation, you'd filter based on the current folder path
    return files;
  }

  function selectFolder(folderPath: string) {
    currentPath = folderPath;
    updateBreadcrumbs(folderPath);
    if (currentBucket) {
      loadFiles(currentBucket.name, folderPath);
    }
  }

  function toggleFileSelection(fileId: string) {
    if (selectedFiles.includes(fileId)) {
      selectedFiles = selectedFiles.filter(id => id !== fileId);
    } else {
      selectedFiles = [...selectedFiles, fileId];
    }
  }

  function selectAllFiles() {
    selectedFiles = files.map(f => f.id);
  }

  function clearSelection() {
    selectedFiles = [];
  }

  async function deleteSelectedFiles() {
    if (selectedFiles.length === 0) return;
    
    const confirmMessage = selectedFiles.length === 1 
      ? 'Weet je zeker dat je dit bestand wilt verwijderen?'
      : `Weet je zeker dat je ${selectedFiles.length} bestanden wilt verwijderen?`;
    
    if (!confirm(confirmMessage)) return;

    try {
      for (const fileId of selectedFiles) {
        const response = await createApiRequest(API_ENDPOINTS.admin.projects.storage.files.delete(projectId, currentBucket?.name || '', fileId), {
          method: 'DELETE',
          headers: {
            'Authorization': `Bearer ${$auth.token}`,
          },
        });

        if (!response.ok) {
          toast.error('Fout bij verwijderen van bestand');
          return;
        }
      }

      toast.success(`${selectedFiles.length} bestand(en) verwijderd`);
      clearSelection();
      if (currentBucket) {
        await loadFiles(currentBucket.name, currentPath);
      }
    } catch (error) {
      console.error('Delete error:', error);
      toast.error('Fout bij verwijderen van bestanden');
    }
  }

  async function createNewBucket() {
    if (!newBucket.name.trim()) {
      toast.error('Bucket naam is verplicht');
      return;
    }

    try {
      const createUrl = API_ENDPOINTS.admin.projects.storage.buckets.create(projectId);
      console.log('🔍 Creating bucket URL:', createUrl);
      const response = await createApiRequest(createUrl, {
        method: 'POST',
        headers: {
          'Authorization': `Bearer ${$auth.token}`,
          'Content-Type': 'application/json',
        },
        body: JSON.stringify(newBucket),
      });

      if (response.ok) {
        const bucket = await response.json();
        buckets = [...buckets, bucket];
        newBucket = {
          name: '',
          description: '',
          is_public: false,
          max_file_size: 50 * 1024 * 1024,
          allowed_types: ['image/jpeg', 'image/png', 'image/gif', 'application/pdf', 'text/plain']
        };
        showNewBucketModal = false;
        toast.success('Bucket succesvol aangemaakt');
        await selectBucket(bucket);
      } else {
        const error = await response.json();
        toast.error(error.error || 'Fout bij aanmaken bucket');
      }
    } catch (error) {
      console.error('Create bucket error:', error);
      toast.error('Netwerkfout bij aanmaken bucket');
    }
  }

  async function deleteBucket(bucket: StorageBucket) {
    const confirmMessage = `Weet je zeker dat je bucket "${bucket.name}" wilt verwijderen?\n\nDit verwijdert alle bestanden en mappen in deze bucket permanent.`;
    
    if (!confirm(confirmMessage)) return;

    try {
      const response = await createApiRequest(API_ENDPOINTS.admin.projects.storage.buckets.delete(projectId, bucket.name), {
        method: 'DELETE',
        headers: {
          'Authorization': `Bearer ${$auth.token}`,
        },
      });

      if (response.ok) {
        buckets = buckets.filter(b => b.id !== bucket.id);
        
        // If we deleted the current bucket, select another one or clear
        if (currentBucket?.id === bucket.id) {
          if (buckets.length > 0) {
            await selectBucket(buckets[0]);
          } else {
            currentBucket = null;
            files = [];
            treeData = [];
          }
        }
        
        toast.success(`Bucket "${bucket.name}" succesvol verwijderd`);
      } else {
        const error = await response.json();
        toast.error(error.error || 'Fout bij verwijderen bucket');
      }
    } catch (error) {
      console.error('Delete bucket error:', error);
      toast.error('Netwerkfout bij verwijderen bucket');
    }
  }

  async function createNewFolder() {
    if (!newFolderName.trim() || !currentBucket) {
      toast.error('Map naam is verplicht');
      return;
    }

    try {
      const createFolderUrl = API_ENDPOINTS.admin.projects.storage.folders.create(projectId, currentBucket.name);
      console.log('🔍 Creating folder URL:', createFolderUrl);
      const response = await createApiRequest(createFolderUrl, {
        method: 'POST',
        headers: {
          'Authorization': `Bearer ${$auth.token}`,
          'Content-Type': 'application/json',
        },
        body: JSON.stringify({
          name: newFolderName.trim(),
          path: currentPath || ''
        }),
      });

      if (response.ok) {
        const result = await response.json();
        toast.success(`Map "${newFolderName}" succesvol aangemaakt`);
        newFolderName = '';
        showNewFolderModal = false;
        
        // Reload folders and files
        await loadFolders();
        if (currentBucket) {
          await loadFiles(currentBucket.name, currentPath);
        }
      } else {
        const error = await response.json();
        toast.error(error.error || 'Fout bij aanmaken map');
      }
    } catch (error) {
      console.error('Create folder error:', error);
      toast.error('Netwerkfout bij aanmaken map');
    }
  }

  function handleFileUpload(event: Event) {
    const target = event.target as HTMLInputElement;
    uploadFiles = target.files;
  }

  async function uploadFile() {
    if (!uploadFiles || uploadFiles.length === 0 || !currentBucket) return;

    uploading = true;
    let successCount = 0;
    let errorCount = 0;
    
    try {
      // Upload files one by one since backend expects single file uploads
      for (let i = 0; i < uploadFiles.length; i++) {
        const formData = new FormData();
        formData.append('file', uploadFiles[i]); // Backend expects 'file', not 'files'
        
        if (currentPath) {
          formData.append('path', currentPath);
        }
        
        try {
          const response = await createApiRequest(API_ENDPOINTS.admin.projects.storage.files.upload(projectId, currentBucket.name), {
            method: 'POST',
            headers: {
              'Authorization': `Bearer ${$auth.token}`,
            },
            body: formData,
          });
          
          if (response.ok) {
            successCount++;
          } else {
            errorCount++;
            const error = await response.json();
            console.error(`Failed to upload ${uploadFiles[i].name}:`, error.error);
          }
        } catch (fileError) {
          errorCount++;
          console.error(`Upload error for ${uploadFiles[i].name}:`, fileError);
        }
      }
      
      // Show results
      if (successCount > 0) {
        toast.success(`${successCount} bestand(en) succesvol geüpload`);
      }
      if (errorCount > 0) {
        toast.error(`${errorCount} bestand(en) gefaald bij uploaden`);
      }
      
      showUploadModal = false;
      uploadFiles = null;
      await loadFiles(currentBucket.name, currentPath);
      
    } catch (error) {
      console.error('Upload error:', error);
      toast.error('Netwerkfout bij uploaden bestanden');
    } finally {
      uploading = false;
    }
  }

  async function downloadFile(file: StorageFile) {
    try {
      const url = file.public_url || file.private_url;
      if (url) {
        const link = document.createElement('a');
        link.href = url;
        link.download = file.original_name;
        link.click();
        toast.success(`Download van ${file.original_name} gestart`);
      } else {
        toast.error('Download URL niet beschikbaar');
      }
    } catch (error) {
      console.error('Download error:', error);
      toast.error('Fout bij downloaden bestand');
    }
  }

  function copyFileUrl(file: StorageFile) {
    const url = file.public_url || file.private_url;
    if (url) {
      navigator.clipboard.writeText(url);
      toast.success('URL gekopieerd naar klembord');
    } else {
      toast.error('URL niet beschikbaar');
    }
  }

  async function loadAllFiles() {
    if (!currentBucket) return [];
    
    try {
      const response = await createApiRequest(`${API_ENDPOINTS.admin.projects.storage.files.list(projectId, currentBucket.name)}?limit=1000`, {
        headers: {
          'Authorization': `Bearer ${$auth.token}`,
          'Content-Type': 'application/json',
        },
      });

      if (response.ok) {
        const data = await response.json();
        return data.files || [];
      }
      return [];
    } catch (error) {
      console.error('Error loading all files:', error);
      return [];
    }
  }

  async function loadAllFolders() {
    if (!currentBucket) return [];
    
    try {
      const response = await createApiRequest(API_ENDPOINTS.admin.projects.storage.folders.list(projectId, currentBucket.name), {
        headers: {
          'Authorization': `Bearer ${$auth.token}`,
          'Content-Type': 'application/json',
        },
      });

      if (response.ok) {
        const data = await response.json();
        return data.folders || [];
      }
      return [];
    } catch (error) {
      console.error('Error loading all folders:', error);
      return [];
    }
  }

  async function buildTreeView() {
    if (!currentBucket) return;
    
    const [allFiles, allFolders] = await Promise.all([
      loadAllFiles(),
      loadAllFolders()
    ]);

    const tree: TreeNode[] = [];
    const pathMap = new Map<string, TreeNode>();

    // Add root node
    const rootNode: TreeNode = {
      name: currentBucket.name,
      path: '',
      type: 'folder',
      children: [],
      expanded: true
    };
    tree.push(rootNode);
    pathMap.set('', rootNode);

    // Add all folders to tree
    allFolders.forEach(folder => {
      const node: TreeNode = {
        name: folder.name,
        path: folder.path,
        type: 'folder',
        children: [],
        expanded: false
      };
      
      const parentPath = folder.path.includes('/') 
        ? folder.path.substring(0, folder.path.lastIndexOf('/'))
        : '';
      
      const parent = pathMap.get(parentPath) || rootNode;
      parent.children.push(node);
      pathMap.set(folder.path, node);
    });

    // Add all files to tree
    totalFilesInTree = 0;
    allFiles.forEach(file => {
      const fileNode: TreeNode = {
        name: file.original_name,
        path: file.folder_path || '',
        type: 'file',
        file: file
      };
      
      const parent = pathMap.get(file.folder_path || '') || rootNode;
      parent.children.push(fileNode);
      totalFilesInTree++;
    });

    // Sort children (folders first, then files)
    function sortChildren(node: TreeNode) {
      if (node.children) {
        node.children.sort((a, b) => {
          if (a.type !== b.type) {
            return a.type === 'folder' ? -1 : 1;
          }
          return a.name.localeCompare(b.name);
        });
        node.children.forEach(sortChildren);
      }
    }
    
    tree.forEach(sortChildren);
    treeData = tree;
  }

  function toggleTreeNode(node: TreeNode) {
    if (node.type === 'folder') {
      node.expanded = !node.expanded;
      treeData = [...treeData]; // Trigger reactivity
    } else if (node.file) {
      // Handle file click
      downloadFile(node.file);
    }
  }

  function selectTreePath(path: string) {
    currentPath = path;
    navigateToPath(path);
  }

  // Drag and drop functions
  function handleDragStart(event: DragEvent, file: StorageFile) {
    if (event.dataTransfer) {
      draggedFile = file;
      event.dataTransfer.effectAllowed = 'move';
      event.dataTransfer.setData('text/plain', file.id);
    }
  }

  function handleDragEnd() {
    draggedFile = null;
    dropTarget = null;
  }

  function handleDragOver(event: DragEvent, folderPath: string) {
    event.preventDefault();
    if (event.dataTransfer) {
      event.dataTransfer.dropEffect = 'move';
    }
    dropTarget = folderPath;
  }

  function handleDragLeave() {
    dropTarget = null;
  }

  async function handleDrop(event: DragEvent, targetFolderPath: string) {
    event.preventDefault();
    dropTarget = null;

    if (!draggedFile || !currentBucket) return;

    // Don't move if already in the target folder
    if (draggedFile.folder_path === targetFolderPath) {
      draggedFile = null;
      return;
    }

    try {
      const response = await createApiRequest(API_ENDPOINTS.admin.projects.storage.files.move(projectId, currentBucket.name, draggedFile.id), {
        method: 'PUT',
        headers: {
          'Authorization': `Bearer ${$auth.token}`,
          'Content-Type': 'application/json',
        },
        body: JSON.stringify({
          new_folder_path: targetFolderPath
        }),
      });

      if (response.ok) {
        const updatedFile = await response.json();
        toast.success(`Bestand "${draggedFile.original_name}" verplaatst naar ${targetFolderPath || 'root'}`);
        
        // Refresh the tree view to show the updated file structure
        await buildTreeView();
        
        // Refresh files in current path if needed
        if (currentBucket) {
          await loadFiles(currentBucket.name, currentPath);
        }
      } else {
        const error = await response.json();
        toast.error(error.error || 'Fout bij verplaatsen bestand');
      }
    } catch (error) {
      console.error('Move file error:', error);
      toast.error('Netwerkfout bij verplaatsen bestand');
    } finally {
      draggedFile = null;
    }
  }
</script>

<svelte:head>
  <title>Bestandsverkenner - CloudBox</title>
</svelte:head>

{#if loading}
  <div class="flex items-center justify-center h-64">
    <div class="text-center">
      <div class="animate-spin rounded-full h-12 w-12 border-b-2 border-primary mx-auto"></div>
      <p class="mt-4 text-muted-foreground">Bestandsverkenner laden...</p>
    </div>
  </div>
{:else}
  <div class="h-full flex flex-col space-y-4">
    <!-- Project Error Notice -->
    {#if !projectLoading && !project}
      <Card class="bg-yellow-50 dark:bg-yellow-900/20 border-yellow-200 dark:border-yellow-800 p-4">
        <div class="flex items-center space-x-3">
          <Icon name="info" size={20} className="text-yellow-600 dark:text-yellow-400" />
          <div>
            <h3 class="text-sm font-medium text-yellow-800 dark:text-yellow-200">Project Info Niet Beschikbaar</h3>
            <p class="text-xs text-yellow-700 dark:text-yellow-300 mt-1">
              Storage functionaliteit werkt nog steeds, maar project details kunnen niet worden getoond.
            </p>
          </div>
        </div>
      </Card>
    {/if}

    <!-- Backend Status Notice -->
    {#if !backendAvailable}
      <Card class="bg-yellow-50 dark:bg-yellow-900/20 border-yellow-200 dark:border-yellow-800 p-4">
        <div class="flex items-center space-x-3">
          <Icon name="warning" size={20} className="text-yellow-600 dark:text-yellow-400" />
          <div>
            <h3 class="text-sm font-medium text-yellow-800 dark:text-yellow-200">Backend Server Niet Beschikbaar</h3>
            <p class="text-xs text-yellow-700 dark:text-yellow-300 mt-1">
              De backend server draait niet of de storage API endpoints zijn nog niet geïmplementeerd. Start de backend server om bestanden te beheren.
            </p>
          </div>
        </div>
      </Card>
    {/if}

    <!-- Header -->
    <div class="flex items-center space-x-4">
      <div class="w-10 h-10 bg-primary/10 rounded-lg flex items-center justify-center">
        <Icon name="folder" size={20} className="text-primary" />
      </div>
      <div>
        <h1 class="text-2xl font-bold text-foreground">Bestandsverkenner</h1>
        <p class="text-sm text-muted-foreground">
          Beheer bestanden en mappen in je buckets
        </p>
      </div>
    </div>

    <!-- Combined Bucket Selector and Navigation -->
    {#if buckets.length > 0}
      <Card class="p-4">
        <div class="flex items-center justify-between">
          <!-- Left: Bucket Selector + Breadcrumbs -->
          <div class="flex items-center space-x-4">
            <div class="flex items-center space-x-2">
              <Label class="text-sm font-medium">Bucket:</Label>
              {#each buckets as bucket}
                <div class="flex items-center space-x-1">
                  <Button
                    variant={currentBucket?.id === bucket.id ? 'default' : 'outline'}
                    size="sm"
                    on:click={() => selectBucket(bucket)}
                    class="flex items-center space-x-2"
                  >
                    <Icon name="folder" size={14} />
                    <span>{bucket.name}</span>
                    <Badge variant="secondary" class="ml-2" title={`${totalFilesInTree} bestanden zichtbaar van ${bucket.file_count} totaal`}>
                      {currentBucket?.id === bucket.id ? totalFilesInTree : bucket.file_count}
                    </Badge>
                  </Button>
                  
                  <!-- Delete bucket button (only show if more than 1 bucket) -->
                  {#if buckets.length > 1}
                    <Button
                      variant="outline"
                      size="sm"
                      on:click={() => deleteBucket(bucket)}
                      class="flex items-center justify-center w-8 h-8 p-0 text-red-600 hover:text-red-700 hover:bg-red-50 dark:hover:bg-red-900/20"
                      title={`Bucket "${bucket.name}" verwijderen`}
                    >
                      <Icon name="trash" size={14} />
                    </Button>
                  {/if}
                </div>
              {/each}
            </div>
            
            <!-- Breadcrumbs -->
            {#if currentBucket && breadcrumbs.length > 1}
              <div class="flex items-center space-x-2 border-l border-border pl-4">
                {#each breadcrumbs as crumb, index}
                  {#if index > 0}
                    <button
                      on:click={() => navigateToPath(crumb.path)}
                      class="text-sm font-medium hover:text-primary transition-colors"
                      class:text-primary={index === breadcrumbs.length - 1}
                      class:text-muted-foreground={index !== breadcrumbs.length - 1}
                    >
                      {crumb.name}
                    </button>
                    {#if index < breadcrumbs.length - 1}
                      <Icon name="arrow-right" size={14} className="text-muted-foreground" />
                    {/if}
                  {/if}
                {/each}
              </div>
            {/if}
          </div>

          <!-- Right: Action Buttons -->
          <div class="flex items-center space-x-3">
            {#if currentBucket}
              <Button
                variant="outline"
                size="sm"
                on:click={() => showNewFolderModal = true}
                class="flex items-center space-x-2"
              >
                <Icon name="folder" size={16} />
                <span>Nieuwe Map</span>
              </Button>
              
              <Button
                size="sm"
                on:click={() => showUploadModal = true}
                class="flex items-center space-x-2"
              >
                <Icon name="upload" size={16} />
                <span>Upload</span>
              </Button>
            {/if}
            
            <Button
              variant="outline"
              size="sm"
              on:click={() => showNewBucketModal = true}
              class="flex items-center space-x-2"
            >
              <Icon name="folder" size={16} />
              <span>Nieuwe Bucket</span>
            </Button>
          </div>
        </div>
      </Card>
    {:else}
      <Card class="p-12">
        <div class="text-center">
          <div class="w-16 h-16 bg-muted rounded-lg flex items-center justify-center mx-auto mb-4">
            <Icon name="folder" size={32} className="text-muted-foreground" />
          </div>
          <h3 class="text-lg font-medium text-foreground mb-2">Geen buckets gevonden</h3>
          <p class="text-muted-foreground mb-4">
            {#if !backendAvailable}
              Backend server niet beschikbaar - start de server om buckets te beheren
            {:else}
              Maak je eerste bucket aan om bestanden te uploaden
            {/if}
          </p>
          {#if backendAvailable}
            <Button on:click={() => showNewBucketModal = true}>
              <Icon name="folder" size={16} class="mr-2" />
              Eerste Bucket Aanmaken
            </Button>
          {/if}
        </div>
      </Card>
    {/if}
  </div>
{/if}

  <!-- Tabs -->
  <div class="border-b border-border">
    <nav class="flex space-x-8">
      <button
        on:click={() => activeTab = 'files'}
        class="flex items-center space-x-2 py-2 px-1 border-b-2 text-sm font-medium transition-colors {
          activeTab === 'files' 
            ? 'border-primary text-primary' 
            : 'border-transparent text-muted-foreground hover:text-foreground hover:border-border'
        }"
      >
        <Icon name="storage" size={16} />
        <span>Bestanden</span>
      </button>
      <button
        on:click={() => activeTab = 'settings'}
        class="flex items-center space-x-2 py-2 px-1 border-b-2 text-sm font-medium transition-colors {
          activeTab === 'settings' 
            ? 'border-primary text-primary' 
            : 'border-transparent text-muted-foreground hover:text-foreground hover:border-border'
        }"
      >
        <Icon name="settings" size={16} />
        <span>Instellingen</span>
      </button>
    </nav>
  </div>

  <!-- Files Tab -->
  {#if activeTab === 'files'}
    <div class="space-y-4">
      <!-- Tree View -->
      <Card class="flex-1 min-h-0">
        <div class="p-4">
          <h3 class="text-sm font-medium text-foreground mb-3">Map Structuur</h3>
          {#if treeData.length === 0}
            <div class="text-center py-8">
              <div class="animate-spin rounded-full h-8 w-8 border-b-2 border-primary mx-auto"></div>
              <p class="mt-2 text-sm text-muted-foreground">Tree laden...</p>
            </div>
          {:else}
            <div class="space-y-1">
              {#each treeData as node}
                <div class="tree-node">
                  <div 
                    class="flex items-center space-x-2 p-2 hover:bg-muted/50 rounded cursor-pointer"
                    on:click={() => toggleTreeNode(node)}
                  >
                    {#if node.type === 'folder'}
                      <Icon name={node.expanded ? 'arrow-down' : 'arrow-right'} size={12} />
                      <Icon name="folder" size={16} className="text-primary" />
                    {:else}
                      <div class="w-3"></div>
                      <Icon name={getFileIcon(node.file?.mime_type || '')} size={16} className="text-muted-foreground" />
                    {/if}
                    <span class="text-sm font-medium">{node.name}</span>
                  </div>
                  
                  {#if node.type === 'folder' && node.expanded && node.children}
                    <div class="ml-6 border-l border-border pl-4">
                      {#each node.children as child}
                        <div class="tree-node">
                          <div 
                            class="flex items-center space-x-2 p-1 hover:bg-muted/30 rounded cursor-pointer"
                            on:click={() => child.type === 'folder' ? selectTreePath(child.path) : toggleTreeNode(child)}
                          >
                            {#if child.type === 'folder'}
                              <Icon name="folder" size={14} className="text-primary" />
                            {:else}
                              <Icon name={getFileIcon(child.file?.mime_type || '')} size={14} className="text-muted-foreground" />
                            {/if}
                            <span class="text-xs">{child.name}</span>
                            {#if child.type === 'file' && child.file}
                              <span class="text-xs text-muted-foreground ml-auto">{formatFileSize(child.file.size)}</span>
                            {/if}
                          </div>
                        </div>
                      {/each}
                    </div>
                  {/if}
                </div>
              {/each}
            </div>
          {/if}
        </div>
      </Card>
    </div>
  {/if}

  <!-- Settings Tab -->
  {#if activeTab === 'settings'}
    <div class="space-y-6">
      <div>
        <h2 class="text-lg font-medium text-foreground">Bucket Instellingen</h2>
        <p class="text-sm text-muted-foreground">
          {#if currentBucket}
            Configureer instellingen voor bucket "{currentBucket.name}"
          {:else}
            Selecteer een bucket om instellingen te beheren
          {/if}
        </p>
      </div>

      {#if !currentBucket}
        <Card class="p-8 text-center">
          <div class="w-16 h-16 bg-muted rounded-lg flex items-center justify-center mx-auto mb-4">
            <Icon name="settings" size={32} className="text-muted-foreground" />
          </div>
          <h3 class="text-lg font-medium text-foreground mb-2">Geen bucket geselecteerd</h3>
          <p class="text-muted-foreground">Selecteer een bucket om de instellingen te beheren</p>
        </Card>
      {:else}
        <form on:submit|preventDefault={saveBucketSettings}>
          <Card class="p-6">
            <div class="space-y-6">
              <div>
                <h3 class="text-sm font-medium text-foreground mb-3">Algemene Instellingen</h3>
                <div class="space-y-4">
                  <div>
                    <Label>Beschrijving</Label>
                    <Input 
                      type="text" 
                      bind:value={bucketSettings.description}
                      placeholder="Optionele beschrijving voor deze bucket"
                      class="mt-1" 
                    />
                  </div>
                  <div>
                    <label class="flex items-center">
                      <input 
                        type="checkbox" 
                        bind:checked={bucketSettings.is_public}
                        class="rounded border-border text-primary focus:ring-primary" 
                      />
                      <span class="ml-2 text-sm text-foreground">Publiek toegankelijk maken</span>
                    </label>
                    <p class="text-xs text-muted-foreground mt-1">
                      Wanneer ingeschakeld kunnen bestanden zonder authenticatie gedownload worden
                    </p>
                  </div>
                </div>
              </div>
              
              <div>
                <h3 class="text-sm font-medium text-foreground mb-3">Upload Limieten</h3>
                <div class="grid grid-cols-1 md:grid-cols-2 gap-4">
                  <div>
                    <Label>Max bestandsgrootte (MB)</Label>
                    <Input 
                      type="number" 
                      bind:value={bucketSettings.max_file_size}
                      min="1"
                      max="1000"
                      class="mt-1" 
                    />
                    <p class="text-xs text-muted-foreground mt-1">Maximum 1000MB (1GB)</p>
                  </div>
                  <div>
                    <Label>Toegestane bestandstypen</Label>
                    <textarea
                      bind:value={bucketSettings.allowed_types}
                      placeholder="image/jpeg,image/png,application/pdf,text/plain,application/msword"
                      rows="3"
                      class="mt-1 w-full px-3 py-2 border border-border rounded-md text-sm bg-background text-foreground placeholder:text-muted-foreground focus:outline-none focus:ring-2 focus:ring-primary focus:border-transparent"
                    ></textarea>
                    <p class="text-xs text-muted-foreground mt-1">
                      Komma-gescheiden lijst van MIME types. Leeg laten voor alle types.
                    </p>
                  </div>
                </div>
              </div>
              
              <div class="flex justify-end">
                <Button type="submit" disabled={savingSettings}>
                  {savingSettings ? 'Bezig met opslaan...' : 'Instellingen Opslaan'}
                </Button>
              </div>
            </div>
          </Card>
        </form>
      {/if}
    </div>
  {/if}

<!-- Upload Modal -->
{#if showUploadModal}
  <div class="fixed inset-0 bg-black/80 backdrop-blur-sm flex items-center justify-center p-4 z-50">
    <Card class="max-w-md w-full p-6 border-2 shadow-2xl">
      <div class="flex items-center space-x-3 mb-4">
        <div class="w-10 h-10 bg-primary/10 rounded-lg flex items-center justify-center">
          <Icon name="storage" size={20} className="text-primary" />
        </div>
        <h2 class="text-xl font-bold text-foreground">Bestanden Uploaden</h2>
      </div>
      
      <form on:submit|preventDefault={uploadFile} class="space-y-4">
        <div>
          <Label>Selecteer bestanden</Label>
          <input
            type="file"
            multiple
            on:change={handleFileUpload}
            class="mt-1 block w-full text-sm text-muted-foreground file:mr-4 file:py-2 file:px-4 file:rounded-full file:border-0 file:text-sm file:font-semibold file:bg-primary/10 file:text-primary hover:file:bg-primary/20"
          />
        </div>

        <div>
          <Label>Upload naar map</Label>
          <select
            bind:value={currentPath}
            class="mt-1 w-full px-3 py-2 border border-border rounded-md text-sm bg-background text-foreground focus:outline-none focus:ring-2 focus:ring-primary focus:border-transparent"
          >
            <option value="">/ (root van bucket)</option>
            {#each folders as folder}
              <option value={folder.path}>/{folder.path}</option>
            {/each}
          </select>
        </div>

        {#if currentPath}
          <div class="text-xs text-muted-foreground">
            Bestanden worden geüpload naar: /{currentPath}/
          </div>
        {/if}
        
        <div class="flex space-x-3 pt-4">
          <Button
            type="button"
            variant="outline"
            on:click={() => { showUploadModal = false; uploadFiles = null; }}
            class="flex-1"
          >
            Annuleren
          </Button>
          <Button
            type="submit"
            disabled={uploading || !uploadFiles || uploadFiles.length === 0}
            class="flex-1"
          >
            {uploading ? 'Uploaden...' : 'Upload Bestanden'}
          </Button>
        </div>
      </form>
    </Card>
  </div>
{/if}

<!-- New Bucket Modal -->
{#if showNewBucketModal}
  <div class="fixed inset-0 bg-black/80 backdrop-blur-sm flex items-center justify-center p-4 z-50">
    <Card class="max-w-md w-full p-6 border-2 shadow-2xl">
      <div class="flex items-center space-x-3 mb-4">
        <div class="w-10 h-10 bg-primary/10 rounded-lg flex items-center justify-center">
          <Icon name="folder" size={20} className="text-primary" />
        </div>
        <h2 class="text-xl font-bold text-foreground">Nieuwe Bucket Aanmaken</h2>
      </div>
      
      <form on:submit|preventDefault={createNewBucket} class="space-y-4">
        <div>
          <Label for="bucket-name">Bucket Naam</Label>
          <Input
            id="bucket-name"
            type="text"
            bind:value={newBucket.name}
            required
            class="mt-1"
            placeholder="bijv. mijn-bucket"
          />
        </div>

        <div>
          <Label for="bucket-description">Beschrijving (optioneel)</Label>
          <Input
            id="bucket-description"
            type="text"
            bind:value={newBucket.description}
            class="mt-1"
            placeholder="Een korte beschrijving van deze bucket"
          />
        </div>

        <div>
          <label class="flex items-center">
            <input
              type="checkbox"
              bind:checked={newBucket.is_public}
              class="rounded border-border text-primary focus:ring-primary"
            />
            <span class="ml-2 text-sm text-foreground">Publiek toegankelijk maken</span>
          </label>
        </div>

        <div>
          <Label for="max-file-size">Maximale bestandsgrootte (MB)</Label>
          <Input
            id="max-file-size"
            type="number"
            value={Math.round(newBucket.max_file_size / (1024 * 1024))}
            min="1"
            max="1000"
            class="mt-1"
            on:input={(e) => newBucket.max_file_size = parseInt(e.target.value) * 1024 * 1024}
          />
        </div>
        
        <div class="flex space-x-3 pt-4">
          <Button
            type="button"
            variant="outline"
            on:click={() => { showNewBucketModal = false; newBucket = { name: '', description: '', is_public: false, max_file_size: 50 * 1024 * 1024, allowed_types: ['image/jpeg', 'image/png', 'image/gif', 'application/pdf', 'text/plain'] }; }}
            class="flex-1"
          >
            Annuleren
          </Button>
          <Button
            type="submit"
            disabled={!newBucket.name.trim()}
            class="flex-1"
          >
            Bucket Aanmaken
          </Button>
        </div>
      </form>
    </Card>
  </div>
{/if}

<!-- New Folder Modal -->
{#if showNewFolderModal}
  <div class="fixed inset-0 bg-black/80 backdrop-blur-sm flex items-center justify-center p-4 z-50">
    <Card class="max-w-md w-full p-6 border-2 shadow-2xl">
      <div class="flex items-center space-x-3 mb-4">
        <div class="w-10 h-10 bg-primary/10 rounded-lg flex items-center justify-center">
          <Icon name="folder" size={20} className="text-primary" />
        </div>
        <h2 class="text-xl font-bold text-foreground">Nieuwe Map Aanmaken</h2>
      </div>
      
      <form on:submit|preventDefault={createNewFolder} class="space-y-4">
        <div>
          <Label for="folder-name">Map Naam</Label>
          <Input
            id="folder-name"
            type="text"
            bind:value={newFolderName}
            required
            class="mt-1"
            placeholder="bijv. documenten"
          />
        </div>

        {#if currentPath}
          <div>
            <Label>Locatie</Label>
            <div class="mt-1 p-2 bg-muted rounded text-sm text-muted-foreground">
              /{currentPath}/
            </div>
          </div>
        {:else}
          <div>
            <Label>Locatie</Label>
            <div class="mt-1 p-2 bg-muted rounded text-sm text-muted-foreground">
              / (root van bucket)
            </div>
          </div>
        {/if}
        
        <div class="flex space-x-3 pt-4">
          <Button
            type="button"
            variant="outline"
            on:click={() => { showNewFolderModal = false; newFolderName = ''; }}
            class="flex-1"
          >
            Annuleren
          </Button>
          <Button
            type="submit"
            disabled={!newFolderName.trim()}
            class="flex-1"
          >
            Map Aanmaken
          </Button>
        </div>
      </form>
    </Card>
  </div>
{/if}