<script lang="ts">
  import { onMount } from 'svelte';
  import { API_ENDPOINTS, createApiRequest } from '$lib/config';
  import { auth } from '$lib/stores/auth';
  import Card from '$lib/components/ui/card.svelte';
  import Button from '$lib/components/ui/button.svelte';
  import Input from '$lib/components/ui/input.svelte';
  import Label from '$lib/components/ui/label.svelte';
  import Badge from '$lib/components/ui/badge.svelte';
  import Textarea from '$lib/components/ui/textarea.svelte';
  import Icon from '$lib/components/ui/icon.svelte';
  import ProjectCard from '$lib/components/ui/project-card.svelte';
  
  interface Organization {
    id: number;
    name: string;
    color: string;
  }

  interface Project {
    id: number;
    name: string;
    description: string;
    slug: string;
    created_at: string;
    is_active: boolean;
    organization?: Organization;
  }

  let projects: Project[] = [];
  let loading = true;
  let error = '';
  
  // Admin stats
  let adminStats = null;
  let loadingAdminStats = false;
  
  $: isSuperAdmin = $auth.user?.role === 'superadmin';

  onMount(() => {
    loadProjects();
    if (isSuperAdmin) {
      loadAdminStats();
    }
  });

  async function loadProjects() {
    loading = true;
    error = '';

    try {
      const response = await createApiRequest(API_ENDPOINTS.projects.list, {
        headers: {
          'Authorization': `Bearer ${$auth.token}`,
        },
      });

      if (response.ok) {
        projects = await response.json();
      } else {
        const data = await response.json();
        error = data.error || 'Fout bij laden van projecten';
      }
    } catch (err) {
      error = 'Netwerkfout bij laden van projecten';
      console.error('Load projects error:', err);
    } finally {
      loading = false;
    }
  }


  async function loadAdminStats() {
    if (!isSuperAdmin) return;
    
    loadingAdminStats = true;
    try {
      const response = await createApiRequest(API_ENDPOINTS.admin.stats.system, {
        headers: {
          'Authorization': `Bearer ${$auth.token}`,
        },
      });
      
      if (response.ok) {
        adminStats = await response.json();
      } else {
        console.error('Failed to load admin stats');
      }
    } catch (err) {
      console.error('Admin stats error:', err);
    } finally {
      loadingAdminStats = false;
    }
  }
  
  function formatDate(dateStr: string) {
    return new Date(dateStr).toLocaleDateString('nl-NL', {
      year: 'numeric',
      month: 'long',
      day: 'numeric'
    });
  }
  
  function formatBytes(bytes: number) {
    if (bytes === 0) return '0 B';
    const k = 1024;
    const sizes = ['B', 'KB', 'MB', 'GB', 'TB'];
    const i = Math.floor(Math.log(bytes) / Math.log(k));
    return parseFloat((bytes / Math.pow(k, i)).toFixed(2)) + ' ' + sizes[i];
  }
</script>

<style>
  .glassmorphism-card {
    background: rgba(255, 255, 255, 0.85);
    backdrop-filter: blur(20px);
    border: 1px solid rgba(255, 255, 255, 0.2);
    border-radius: 16px;
    padding: 24px;
    box-shadow: 
      0 8px 25px -8px rgba(0, 0, 0, 0.1),
      0 4px 12px -4px rgba(0, 0, 0, 0.08),
      0 0 0 1px rgba(255, 255, 255, 0.05) inset;
    transition: all 0.3s cubic-bezier(0.4, 0.0, 0.2, 1);
    position: relative;
    overflow: hidden;
  }

  .glassmorphism-card:hover {
    transform: translateY(-2px);
    box-shadow: 
      0 12px 35px -12px rgba(0, 0, 0, 0.15),
      0 8px 20px -8px rgba(0, 0, 0, 0.12),
      0 0 0 1px rgba(255, 255, 255, 0.1) inset;
  }

  .glassmorphism-card::before {
    content: '';
    position: absolute;
    top: 0;
    left: 0;
    right: 0;
    bottom: 0;
    background: linear-gradient(
      135deg,
      rgba(255, 255, 255, 0.1) 0%,
      rgba(255, 255, 255, 0) 50%,
      rgba(0, 0, 0, 0.05) 100%
    );
    pointer-events: none;
    z-index: 1;
  }

  .glassmorphism-card > * {
    position: relative;
    z-index: 2;
  }

  .glassmorphism-icon {
    width: 40px;
    height: 40px;
    border-radius: 10px;
    display: flex;
    align-items: center;
    justify-content: center;
    backdrop-filter: blur(10px);
    border: 1px solid rgba(255, 255, 255, 0.2);
    transition: all 0.2s ease;
  }

  .glassmorphism-icon:hover {
    transform: scale(1.05);
  }

  /* Colorful Glassmorphism Variants */
  .glassmorphism-card-primary {
    background: linear-gradient(
      135deg,
      rgba(59, 130, 246, 0.15) 0%,
      rgba(147, 197, 253, 0.1) 50%,
      rgba(219, 234, 254, 0.05) 100%
    );
    border: 1px solid rgba(59, 130, 246, 0.2);
    box-shadow: 
      0 8px 25px -8px rgba(59, 130, 246, 0.3),
      0 4px 12px -4px rgba(59, 130, 246, 0.2),
      0 0 0 1px rgba(147, 197, 253, 0.1) inset;
  }

  .glassmorphism-card-primary:hover {
    transform: translateY(-3px) scale(1.02);
    box-shadow: 
      0 12px 40px -12px rgba(59, 130, 246, 0.4),
      0 8px 25px -8px rgba(59, 130, 246, 0.3),
      0 0 0 1px rgba(147, 197, 253, 0.15) inset;
  }

  .glassmorphism-card-success {
    background: linear-gradient(
      135deg,
      rgba(34, 197, 94, 0.15) 0%,
      rgba(134, 239, 172, 0.1) 50%,
      rgba(220, 252, 231, 0.05) 100%
    );
    border: 1px solid rgba(34, 197, 94, 0.2);
    box-shadow: 
      0 8px 25px -8px rgba(34, 197, 94, 0.3),
      0 4px 12px -4px rgba(34, 197, 94, 0.2),
      0 0 0 1px rgba(134, 239, 172, 0.1) inset;
  }

  .glassmorphism-card-success:hover {
    transform: translateY(-3px) scale(1.02);
    box-shadow: 
      0 12px 40px -12px rgba(34, 197, 94, 0.4),
      0 8px 25px -8px rgba(34, 197, 94, 0.3),
      0 0 0 1px rgba(134, 239, 172, 0.15) inset;
  }

  .glassmorphism-card-purple {
    background: linear-gradient(
      135deg,
      rgba(168, 85, 247, 0.15) 0%,
      rgba(196, 181, 253, 0.1) 50%,
      rgba(237, 233, 254, 0.05) 100%
    );
    border: 1px solid rgba(168, 85, 247, 0.2);
    box-shadow: 
      0 8px 25px -8px rgba(168, 85, 247, 0.3),
      0 4px 12px -4px rgba(168, 85, 247, 0.2),
      0 0 0 1px rgba(196, 181, 253, 0.1) inset;
  }

  .glassmorphism-card-purple:hover {
    transform: translateY(-3px) scale(1.02);
    box-shadow: 
      0 12px 40px -12px rgba(168, 85, 247, 0.4),
      0 8px 25px -8px rgba(168, 85, 247, 0.3),
      0 0 0 1px rgba(196, 181, 253, 0.15) inset;
  }

  .glassmorphism-card-warning {
    background: linear-gradient(
      135deg,
      rgba(245, 158, 11, 0.15) 0%,
      rgba(252, 211, 77, 0.1) 50%,
      rgba(254, 243, 199, 0.05) 100%
    );
    border: 1px solid rgba(245, 158, 11, 0.2);
    box-shadow: 
      0 8px 25px -8px rgba(245, 158, 11, 0.3),
      0 4px 12px -4px rgba(245, 158, 11, 0.2),
      0 0 0 1px rgba(252, 211, 77, 0.1) inset;
  }

  .glassmorphism-card-warning:hover {
    transform: translateY(-3px) scale(1.02);
    box-shadow: 
      0 12px 40px -12px rgba(245, 158, 11, 0.4),
      0 8px 25px -8px rgba(245, 158, 11, 0.3),
      0 0 0 1px rgba(252, 211, 77, 0.15) inset;
  }

  /* Dark mode support - CloudBox theme system */
  :global(.cloudbox-dark) .glassmorphism-card {
    background: rgba(26, 26, 26, 0.7);
    border: 1px solid rgba(255, 255, 255, 0.1);
    box-shadow: 
      0 8px 25px -8px rgba(0, 0, 0, 0.6),
      0 4px 12px -4px rgba(0, 0, 0, 0.4),
      0 0 0 1px rgba(255, 255, 255, 0.05) inset;
    backdrop-filter: blur(25px);
  }
  
  :global(.cloudbox-dark) .glassmorphism-card:hover {
    background: rgba(33, 33, 33, 0.8);
    box-shadow: 
      0 12px 35px -12px rgba(0, 0, 0, 0.7),
      0 8px 20px -8px rgba(0, 0, 0, 0.5),
      0 0 0 1px rgba(255, 255, 255, 0.1) inset;
    backdrop-filter: blur(30px);
  }
  
  :global(.cloudbox-dark) .glassmorphism-card::before {
    background: linear-gradient(
      135deg,
      rgba(255, 255, 255, 0.03) 0%,
      rgba(255, 255, 255, 0) 50%,
      rgba(0, 0, 0, 0.15) 100%
    );
  }
  
  :global(.cloudbox-dark) .glassmorphism-icon {
    border: 1px solid rgba(255, 255, 255, 0.15);
    background: rgba(255, 255, 255, 0.05);
  }

  /* Dark mode colorful variants */
  :global(.cloudbox-dark) .glassmorphism-card-primary {
    background: linear-gradient(
      135deg,
      rgba(59, 130, 246, 0.2) 0%,
      rgba(30, 58, 138, 0.15) 50%,
      rgba(33, 33, 33, 0.8) 100%
    );
  }

  :global(.cloudbox-dark) .glassmorphism-card-success {
    background: linear-gradient(
      135deg,
      rgba(34, 197, 94, 0.2) 0%,
      rgba(20, 83, 45, 0.15) 50%,
      rgba(33, 33, 33, 0.8) 100%
    );
  }

  :global(.cloudbox-dark) .glassmorphism-card-purple {
    background: linear-gradient(
      135deg,
      rgba(168, 85, 247, 0.2) 0%,
      rgba(88, 28, 135, 0.15) 50%,
      rgba(33, 33, 33, 0.8) 100%
    );
  }

  :global(.cloudbox-dark) .glassmorphism-card-warning {
    background: linear-gradient(
      135deg,
      rgba(245, 158, 11, 0.2) 0%,
      rgba(146, 64, 14, 0.15) 50%,
      rgba(33, 33, 33, 0.8) 100%
    );
  }

  /* Mobile responsiveness */
  @media (max-width: 640px) {
    .glassmorphism-card {
      padding: 20px;
      border-radius: 12px;
    }
    
    .glassmorphism-icon {
      width: 36px;
      height: 36px;
    }
  }

  /* Reduce motion for accessibility */
  @media (prefers-reduced-motion: reduce) {
    .glassmorphism-card,
    .glassmorphism-icon {
      transition: none;
    }
    
    .glassmorphism-card:hover {
      transform: none;
    }
  }
</style>

<svelte:head>
  <title>Dashboard - CloudBox</title>
</svelte:head>

<div class="space-y-8">
  <!-- Header -->
  <div class="flex items-center justify-between">
    <div>
      <h1 class="text-heading-1">
        Welkom terug, {$auth.user?.name?.split(' ')[0] || 'User'}
      </h1>
      <p class="mt-1 text-body-sm">
        Hier is een overzicht van je CloudBox projecten
      </p>
    </div>
    <Button
      href="/dashboard/projects"
      variant="floating"
      size="icon-lg"
      iconOnly={true}
      tooltip="Bekijk alle projecten"
    >
      <Icon name="package" size={20} />
    </Button>
  </div>

  <!-- Quick Stats with Colorful Glassmorphism -->
  <div class="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-4">
    <div class="glassmorphism-card glassmorphism-card-primary p-6">
      <div class="flex items-center justify-between">
        <div>
          <p class="text-caption-lg opacity-70">Projecten</p>
          <p class="text-heading-3">{projects.length}</p>
        </div>
        <div class="glassmorphism-icon bg-blue-500/20">
          <Icon name="package" size={20} className="text-blue-600" />
        </div>
      </div>
    </div>
    
    <div class="glassmorphism-card glassmorphism-card-success p-6">
      <div class="flex items-center justify-between">
        <div>
          <p class="text-caption-lg opacity-70">Actief</p>
          <p class="text-heading-3">{projects.filter(p => p.is_active).length}</p>
        </div>
        <div class="glassmorphism-icon bg-green-500/20">
          <Icon name="database" size={20} className="text-green-600" />
        </div>
      </div>
    </div>
    
    <div class="glassmorphism-card glassmorphism-card-purple p-6">
      <div class="flex items-center justify-between">
        <div>
          <p class="text-caption-lg opacity-70">Storage</p>
          <p class="text-heading-3">
            {#if isSuperAdmin && adminStats}
              {formatBytes(adminStats.storage_used || 0)}
            {:else}
              2.4GB
            {/if}
          </p>
        </div>
        <div class="glassmorphism-icon bg-purple-500/20">
          <Icon name="storage" size={20} className="text-purple-600" />
        </div>
      </div>
    </div>
    
    <div class="glassmorphism-card glassmorphism-card-warning p-6">
      <div class="flex items-center justify-between">
        <div>
          <p class="text-caption-lg opacity-70">API Calls</p>
          <p class="text-heading-3">
            {#if isSuperAdmin && adminStats}
              {adminStats.api_calls_24h || 0}
            {:else}
              1,247
            {/if}
          </p>
        </div>
        <div class="glassmorphism-icon bg-amber-500/20">
          <Icon name="zap" size={20} className="text-amber-600" />
        </div>
      </div>
    </div>
  </div>
  
  <!-- Admin Stats for Superadmins -->
  {#if isSuperAdmin}
    <div class="space-y-6">
      <div class="flex items-center justify-between">
        <h2 class="text-heading-2">Systeem Statistieken</h2>
        <Button
          variant="secondary"
          size="icon"
          iconOnly={true}
          on:click={loadAdminStats}
          disabled={loadingAdminStats}
          tooltip="Vernieuwen"
          class="hover:rotate-180 transition-transform duration-300"
        >
          <Icon name="refresh-cw" size={18} className={loadingAdminStats ? 'animate-spin' : ''} />
        </Button>
      </div>
      
      {#if loadingAdminStats}
        <Card class="p-6">
          <div class="text-center">
            <div class="animate-spin rounded-full h-8 w-8 border-b-2 border-primary mx-auto"></div>
            <p class="mt-4 text-body-sm">Admin statistieken laden...</p>
          </div>
        </Card>
      {:else if adminStats}
        <div class="grid grid-cols-1 md:grid-cols-3 gap-6">
          <div class="glassmorphism-card glassmorphism-card-primary p-6">
            <div class="flex items-center justify-between">
              <div>
                <p class="text-caption-lg opacity-70">Totaal Gebruikers</p>
                <p class="text-heading-3">{adminStats.total_users || 0}</p>
              </div>
              <div class="glassmorphism-icon bg-blue-500/20">
                <Icon name="user" size={20} className="text-blue-600" />
              </div>
            </div>
          </div>
          
          <div class="glassmorphism-card glassmorphism-card-success p-6">
            <div class="flex items-center justify-between">
              <div>
                <p class="text-caption-lg opacity-70">Systeem Uptime</p>
                <p class="text-heading-3">{adminStats.uptime || '0d'}</p>
              </div>
              <div class="glassmorphism-icon bg-green-500/20">
                <Icon name="shield-check" size={20} className="text-green-600" />
              </div>
            </div>
          </div>
          
          <div class="glassmorphism-card glassmorphism-card-purple p-6">
            <div class="flex items-center justify-between">
              <div>
                <p class="text-caption-lg opacity-70">Database Queries</p>
                <p class="text-heading-3">{adminStats.database_queries || 0}</p>
              </div>
              <div class="glassmorphism-icon bg-purple-500/20">
                <Icon name="database" size={20} className="text-purple-600" />
              </div>
            </div>
          </div>
        </div>
        
        <div class="grid grid-cols-1 md:grid-cols-2 gap-6">
          <div class="glassmorphism-card glassmorphism-floating p-6">
            <h3 class="text-heading-4 mb-4">Systeem Informatie</h3>
            <div class="space-y-3 text-body-sm">
              <div class="flex justify-between">
                <span class="opacity-70">Server OS:</span>
                <span class="font-medium">{adminStats.os || 'Onbekend'}</span>
              </div>
              <div class="flex justify-between">
                <span class="opacity-70">CPU Usage:</span>
                <span class="font-medium">{adminStats.cpu_usage || '0'}%</span>
              </div>
              <div class="flex justify-between">
                <span class="opacity-70">Memory Usage:</span>
                <span class="font-medium">{adminStats.memory_usage || '0'}%</span>
              </div>
              <div class="flex justify-between">
                <span class="opacity-70">Disk Usage:</span>
                <span class="font-medium">{adminStats.disk_usage || '0'}%</span>
              </div>
            </div>
          </div>
          
          <div class="glassmorphism-card glassmorphism-floating p-6">
            <h3 class="text-heading-4 mb-4">Recente Activiteit</h3>
            <div class="space-y-3 text-body-sm">
              <div class="flex justify-between">
                <span class="opacity-70">Deployments (7d):</span>
                <span class="font-medium">{adminStats.deployments_7d || 0}</span>
              </div>
              <div class="flex justify-between">
                <span class="opacity-70">Functions Executed:</span>
                <span class="font-medium">{adminStats.functions_executed || 0}</span>
              </div>
              <div class="flex justify-between">
                <span class="opacity-70">Active Sessions:</span>
                <span class="font-medium">{adminStats.active_sessions || 0}</span>
              </div>
              <div class="flex justify-between">
                <span class="opacity-70">Error Rate (24h):</span>
                <span class="font-medium">{adminStats.error_rate_24h || '0'}%</span>
              </div>
            </div>
          </div>
        </div>
      {/if}
    </div>
  {/if}

  <!-- Error message -->
  {#if error}
    <Card class="bg-destructive/10 border-destructive/20 p-4">
      <div class="flex justify-between items-center">
        <p class="text-destructive text-sm">{error}</p>
        <Button
          variant="ghost"
          size="sm"
          on:click={() => error = ''}
          class="text-destructive hover:text-destructive/80"
        >
          ×
        </Button>
      </div>
    </Card>
  {/if}

  <!-- Projects Section -->
  <div>
    <div class="flex items-center justify-between mb-6">
      <h2 class="text-heading-3">Recente projecten</h2>
      {#if projects.length > 3}
        <Button
          variant="ghost"
          size="sm"
          href="/dashboard/projects"
          class="flex items-center space-x-2 text-muted-foreground"
        >
          <span>Alle projecten</span>
          <Icon name="arrow-right" size={14} />
        </Button>
      {/if}
    </div>

    {#if loading}
      <div class="text-center py-8">
        <div class="animate-spin rounded-full h-6 w-6 border-b-2 border-primary mx-auto"></div>
        <p class="mt-2 text-body-sm">Projecten laden...</p>
      </div>
    {:else if projects.length === 0}
      <div class="text-center py-8">
        <div class="mx-auto mb-4 flex h-12 w-12 items-center justify-center rounded-full bg-muted">
          <Icon name="package" size={24} className="text-muted-foreground" />
        </div>
        <h3 class="mb-2 text-heading-4">Nog geen projecten</h3>
        <p class="mb-4 text-body-sm max-w-sm mx-auto">
          Maak je eerste CloudBox project aan om te beginnen
        </p>
        <Button
          href="/dashboard/projects"
          class="flex items-center space-x-2"
        >
          <Icon name="plus" size={16} />
          <span>Eerste project aanmaken</span>
        </Button>
      </div>
    {:else}
      <div class="grid gap-4 sm:grid-cols-2 lg:grid-cols-3">
        {#each projects.slice(0, 6) as project}
          <ProjectCard {project} />
        {/each}
      </div>
    {/if}
  </div>
</div>

