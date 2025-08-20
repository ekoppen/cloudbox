<script lang="ts">
  import { page } from '$app/stores';
  import { auth } from '$lib/stores/auth';
  import { sidebarStore } from '$lib/stores/sidebar';
  import Icon from '$lib/components/ui/icon.svelte';
  import Tooltip from '$lib/components/ui/tooltip.svelte';
  import CloudBoxLogo from '$lib/components/ui/cloudbox-logo.svelte';
  import { createEventDispatcher, onMount } from 'svelte';
  import './sidebar.css';

  interface NavigationItem {
    label: string;
    href: string;
    icon: string;
    exact?: boolean;
    adminOnly?: boolean;
    children?: NavigationItem[];
  }

  export let context: 'dashboard' | 'project' | 'admin' = 'dashboard';
  export let projectId: string | undefined = undefined;
  export let projectName: string | undefined = undefined;

  const dispatch = createEventDispatcher();

  let expandTimeout: number;
  let collapseTimeout: number;

  // Subscribe to sidebar state
  $: sidebarState = $sidebarStore;
  $: isHovered = sidebarState.isHovered;

  onMount(() => {
    // Initialize sidebar context
    sidebarStore.setContext(context, projectId, projectName);
  });

  // Update context when props change
  $: {
    sidebarStore.setContext(context, projectId, projectName);
  }

  // Dashboard navigation items
  const dashboardItems: NavigationItem[] = [
    {
      label: 'Dashboard',
      href: '/dashboard',
      icon: 'home',
      exact: true
    },
    {
      label: 'Projecten',
      href: '/dashboard/projects',
      icon: 'package'
    },
    {
      label: 'Organizations',
      href: '/dashboard/organizations',
      icon: 'building'
    },
    {
      label: 'Instellingen',
      href: '/dashboard/settings',
      icon: 'settings'
    },
    {
      label: 'Admin',
      href: '/dashboard/admin',
      icon: 'shield-check',
      adminOnly: true
    }
  ];

  // Project navigation items (dynamic based on projectId)
  function getProjectItems(projectId: string | undefined): NavigationItem[] {
    if (!projectId) return [];
    
    return [
      { 
        label: 'Overzicht', 
        href: `/dashboard/projects/${projectId}`, 
        icon: 'dashboard',
        exact: true 
      },
      { 
        label: 'Database', 
        href: `/dashboard/projects/${projectId}/database`, 
        icon: 'database' 
      },
      { 
        label: 'Authenticatie', 
        href: `/dashboard/projects/${projectId}/auth`, 
        icon: 'auth' 
      },
      { 
        label: 'Opslag', 
        href: `/dashboard/projects/${projectId}/storage`, 
        icon: 'storage' 
      },
      { 
        label: 'Functies', 
        href: `/dashboard/projects/${projectId}/functions`, 
        icon: 'functions' 
      },
      { 
        label: 'API', 
        href: `/dashboard/projects/${projectId}/api`, 
        icon: 'settings' 
      },
      { 
        label: 'Berichten', 
        href: `/dashboard/projects/${projectId}/messaging`, 
        icon: 'messaging' 
      },
      {
        label: 'Deployments',
        href: `/dashboard/projects/${projectId}/deployments`,
        icon: 'deployments',
        children: [
          { label: 'Deployments', href: `/dashboard/projects/${projectId}/deployments`, icon: 'deployments' },
          { label: 'SSH Keys', href: `/dashboard/projects/${projectId}/ssh-keys`, icon: 'key' },
          { label: 'Servers', href: `/dashboard/projects/${projectId}/servers`, icon: 'server' },
          { label: 'GitHub', href: `/dashboard/projects/${projectId}/github`, icon: 'github' }
        ]
      },
      { 
        label: 'Instellingen', 
        href: `/dashboard/projects/${projectId}/settings`, 
        icon: 'settings' 
      }
    ];
  }

  $: projectItems = getProjectItems(projectId);

  // Admin navigation items
  const adminItems: NavigationItem[] = [
    {
      label: 'Plugins',
      href: '/dashboard/admin/plugins',
      icon: 'plugin'
    },
    {
      label: 'Users',
      href: '/dashboard/admin/users',
      icon: 'users'
    },
    {
      label: 'System',
      href: '/dashboard/admin/system',
      icon: 'settings'
    }
  ];

  // Get appropriate navigation items based on context
  $: navigationItems = (() => {
    switch (context) {
      case 'project':
        return projectItems;
      case 'admin':
        return adminItems;
      default:
        return dashboardItems;
    }
  })();

  function isActive(item: NavigationItem, currentPath: string): boolean {
    if (item.exact) {
      return currentPath === item.href;
    }
    
    // For project routes, be more specific with matching
    if (item.href.includes('/projects/')) {
      // Exact match for main sections
      if (currentPath === item.href) {
        return true;
      }
      
      // Check if we're in a subsection of this item
      if (currentPath.startsWith(item.href + '/')) {
        return true;
      }
      
      // Check children if they exist
      if (item.children && item.children.some(child => 
        currentPath === child.href || currentPath.startsWith(child.href + '/')
      )) {
        return true;
      }
      
      return false;
    }
    
    // Default behavior for non-project routes
    return currentPath.startsWith(item.href) || 
           (item.children && item.children.some(child => currentPath.startsWith(child.href)));
  }

  function handleMouseEnter() {
    clearTimeout(collapseTimeout);
    expandTimeout = window.setTimeout(() => {
      sidebarStore.setHovered(true);
    }, 100);
  }

  function handleMouseLeave() {
    clearTimeout(expandTimeout);
    collapseTimeout = window.setTimeout(() => {
      sidebarStore.setHovered(false);
    }, 300);
  }

  $: currentPath = $page.url.pathname;
  $: user = $auth.user;
  $: isAdmin = user?.role === 'superadmin';

  // Filter items based on admin permissions
  $: visibleItems = navigationItems.filter(item => 
    !item.adminOnly || (item.adminOnly && isAdmin)
  );

  // Ensure reactivity for project context
  $: if (context === 'project' && projectId) {
    // Force recalculation when project context is active
    navigationItems; // Reference to trigger reactivity
  }
</script>

<!-- Supabase-style Collapsible Sidebar -->
<aside 
  class="sidebar fixed inset-y-0 left-0 z-50 bg-sidebar border-r border-sidebar-border group"
  class:w-sidebar-collapsed={!isHovered}
  class:w-sidebar={isHovered}
  on:mouseenter={handleMouseEnter}
  on:mouseleave={handleMouseLeave}
  role="navigation"
  aria-label="Main navigation"
>
  <div class="flex h-full flex-col">
    <!-- Logo/Brand Section -->
    <div class="flex h-16 items-center border-b border-sidebar-border px-4">
      {#if !isHovered}
        <!-- Collapsed state - just logo -->
        <div class="flex items-center justify-center w-full">
          <a href="/dashboard" class="flex-shrink-0">
            <div class="flex h-8 w-8 items-center justify-center rounded-lg bg-primary text-primary-foreground">
              <Icon name="cloud" size={20} className="icon-contrast" />
            </div>
          </a>
        </div>
      {:else if context === 'project' && projectName}
        <!-- Project context header -->
        <div class="flex items-center space-x-3 min-w-0">
          <a href="/dashboard" class="flex-shrink-0">
            <div class="flex h-8 w-8 items-center justify-center rounded-lg bg-primary text-primary-foreground">
              <Icon name="cloud" size={16} className="icon-contrast" />
            </div>
          </a>
          <div class="project-header min-w-0">
            <div class="flex items-center space-x-2">
              <Icon name="package" size={16} className="text-muted-foreground flex-shrink-0 icon-contrast" />
              <span class="text-sm font-heading font-semibold text-foreground truncate">{projectName}</span>
            </div>
          </div>
        </div>
      {:else}
        <!-- Dashboard/Admin context header -->
        <div class="flex items-center space-x-2">
          <CloudBoxLogo size="md" showText={true} />
        </div>
      {/if}
    </div>

    <!-- Navigation -->
    <nav class="flex-1 space-y-1 p-2 overflow-y-auto">
      
      {#each visibleItems as item}
        <!-- Main navigation item -->
        <div class="relative">
          {#if !isHovered}
            <Tooltip text={item.label} position="right">
              <a
                href={item.href}
                class="nav-item group/item flex items-center h-10 rounded-lg text-sidebar-foreground hover:bg-primary/10 hover:text-primary focus:bg-primary/10 focus:text-primary focus:outline-none relative justify-center px-2"
                class:nav-item-active={isActive(item, currentPath)}
              >
                <Icon 
                  name={item.icon} 
                  size={20}
                  className="flex-shrink-0 icon-contrast"
                />
              </a>
            </Tooltip>
          {:else}
            <a
              href={item.href}
              class="nav-item group/item flex items-center h-10 rounded-lg text-sidebar-foreground hover:bg-primary/10 hover:text-primary focus:bg-primary/10 focus:text-primary focus:outline-none relative px-3"
              class:nav-item-active={isActive(item, currentPath)}
            >
              <Icon 
                name={item.icon} 
                size={20}
                className="flex-shrink-0 mr-3 icon-contrast"
              />
              
              <!-- Label (visible when expanded) -->
              <span 
                class="sidebar-text text-sm font-heading font-semibold truncate sidebar-text-visible"
              >
                {item.label}
              </span>
              
              <!-- Expand indicator for items with children -->
              {#if item.children}
                <div class="ml-auto">
                  <Icon 
                    name="chevron-down" 
                    size={16} 
                    className="chevron text-muted-foreground icon-contrast {isActive(item, currentPath) ? 'chevron-expanded' : ''}"
                  />
                </div>
              {/if}
            </a>
          {/if}
        </div>
        
        <!-- Sub-items (only visible when expanded and parent is active) -->
        {#if item.children && isHovered}
          <div 
            class="sub-menu ml-6 space-y-1"
            class:sub-menu-expanded={isActive(item, currentPath)}
          >
            {#each item.children as subItem}
              <a
                href={subItem.href}
                class="nav-item flex items-center h-8 px-3 rounded-md text-sm font-heading {currentPath === subItem.href ? 'bg-primary/15 text-primary font-semibold' : 'text-muted-foreground hover:text-primary hover:bg-primary/5 font-medium'}"
              >
                <Icon name={subItem.icon} size={16} className="mr-2 flex-shrink-0 icon-contrast" />
                {subItem.label}
              </a>
            {/each}
          </div>
        {/if}
      {/each}
    </nav>

    <!-- User section at bottom -->
    <div class="border-t border-sidebar-border p-2">
      {#if !isHovered}
        <Tooltip text={user?.name || 'User'} position="right">
          <div class="nav-item flex items-center h-10 rounded-lg hover:bg-sidebar-hover cursor-pointer justify-center px-2">
            <div class="h-7 w-7 rounded-full bg-primary text-primary-foreground flex items-center justify-center text-xs font-semibold flex-shrink-0">
              {user?.name?.charAt(0)?.toUpperCase() || 'U'}
            </div>
          </div>
        </Tooltip>
      {:else}
        <div class="nav-item flex items-center h-10 rounded-lg hover:bg-sidebar-hover cursor-pointer px-3">
          <div class="h-7 w-7 rounded-full bg-primary text-primary-foreground flex items-center justify-center text-xs font-semibold flex-shrink-0">
            {user?.name?.charAt(0)?.toUpperCase() || 'U'}
          </div>
          
          <div class="user-info ml-3 min-w-0 user-info-visible">
            <div class="text-ui font-medium text-foreground truncate">
              {user?.name || 'User'}
            </div>
            <div class="text-caption text-muted-foreground truncate">
              {user?.email || ''}
            </div>
          </div>
        </div>
      {/if}
    </div>
  </div>
</aside>