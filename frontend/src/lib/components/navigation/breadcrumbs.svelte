<script lang="ts">
  import { page } from '$app/stores';
  import Icon from '$lib/components/ui/icon.svelte';

  interface BreadcrumbItem {
    label: string;
    href?: string;
    icon?: string;
  }

  // Function to generate breadcrumbs from current path
  function generateBreadcrumbs(pathname: string): BreadcrumbItem[] {
    const segments = pathname.split('/').filter(Boolean);
    const breadcrumbs: BreadcrumbItem[] = [];

    // Always start with Dashboard
    if (segments[0] === 'dashboard') {
      breadcrumbs.push({
        label: 'Dashboard',
        href: '/dashboard',
        icon: 'home'
      });

      // Handle specific routes
      if (segments.length > 1) {
        const section = segments[1];
        
        switch (section) {
          case 'projects':
            breadcrumbs.push({
              label: 'Projecten',
              href: segments.length === 2 ? undefined : '/dashboard/projects',
              icon: 'package'
            });
            
            // Specific project
            if (segments.length > 2 && segments[2] !== 'create') {
              // Try to get project name from page data or use ID
              const projectId = segments[2];
              breadcrumbs.push({
                label: `Project ${projectId}`,
                href: segments.length === 3 ? undefined : `/dashboard/projects/${projectId}`,
                icon: 'folder'
              });
              
              // Project subsection
              if (segments.length > 3) {
                const subsection = segments[3];
                const subsectionLabels: Record<string, { label: string; icon: string }> = {
                  'settings': { label: 'Instellingen', icon: 'settings' },
                  'database': { label: 'Database', icon: 'database' },
                  'storage': { label: 'Storage', icon: 'storage' },
                  'auth': { label: 'Authentication', icon: 'user-check' },
                  'api': { label: 'API', icon: 'code' },
                  'functions': { label: 'Functions', icon: 'zap' },
                  'deployments': { label: 'Deployments', icon: 'rocket' },
                  'github': { label: 'GitHub', icon: 'github' },
                  'scripts': { label: 'Scripts', icon: 'terminal' },
                  'messaging': { label: 'Messaging', icon: 'mail' },
                  'servers': { label: 'Servers', icon: 'server' },
                  'ssh-keys': { label: 'SSH Keys', icon: 'key' }
                };
                
                const subsectionInfo = subsectionLabels[subsection] || { 
                  label: subsection.charAt(0).toUpperCase() + subsection.slice(1), 
                  icon: 'folder' 
                };
                
                breadcrumbs.push({
                  label: subsectionInfo.label,
                  icon: subsectionInfo.icon
                });
              }
            }
            break;
            
          case 'organizations':
            breadcrumbs.push({
              label: 'Organizations',
              href: segments.length === 2 ? undefined : '/dashboard/organizations',
              icon: 'building'
            });
            
            if (segments.length > 2) {
              const orgId = segments[2];
              breadcrumbs.push({
                label: `Organization ${orgId}`,
                icon: 'building'
              });
            }
            break;
            
          case 'settings':
            breadcrumbs.push({
              label: 'Instellingen',
              icon: 'settings'
            });
            break;
            
          case 'admin':
            breadcrumbs.push({
              label: 'Admin',
              href: segments.length === 2 ? undefined : '/dashboard/admin',
              icon: 'shield-check'
            });
            
            if (segments.length > 2) {
              const adminSection = segments[2];
              const adminLabels: Record<string, string> = {
                'users': 'Gebruikers',
                'plugins': 'Plugins'
              };
              
              breadcrumbs.push({
                label: adminLabels[adminSection] || adminSection.charAt(0).toUpperCase() + adminSection.slice(1),
                icon: adminSection === 'users' ? 'users' : adminSection === 'plugins' ? 'puzzle' : 'folder'
              });
            }
            break;
        }
      }
    }

    return breadcrumbs;
  }

  $: breadcrumbs = generateBreadcrumbs($page.url.pathname);
</script>

<nav class="flex items-center space-x-2 text-sm text-muted-foreground">
  {#each breadcrumbs as crumb, index}
    {#if index > 0}
      <Icon name="chevron-right" size={14} className="flex-shrink-0" />
    {/if}
    
    {#if crumb.href}
      <a 
        href={crumb.href}
        class="flex items-center space-x-1.5 hover:text-foreground transition-colors"
      >
        {#if crumb.icon}
          <Icon name={crumb.icon} size={14} className="flex-shrink-0" />
        {/if}
        <span class="truncate">{crumb.label}</span>
      </a>
    {:else}
      <div class="flex items-center space-x-1.5 text-foreground">
        {#if crumb.icon}
          <Icon name={crumb.icon} size={14} className="flex-shrink-0" />
        {/if}
        <span class="truncate font-medium">{crumb.label}</span>
      </div>
    {/if}
  {/each}
</nav>