<script lang="ts">
  import { page } from '$app/stores';
  import Icon from '$lib/components/ui/icon.svelte';
</script>

<div class="p-6 space-y-8">
  <!-- Page Header -->
  <div class="flex items-center justify-between">
    <div>
      <h1 class="text-3xl font-bold text-foreground">Admin</h1>
      <p class="text-muted-foreground mt-2">CloudBox administratie en systeembeheer</p>
    </div>
  </div>

  <!-- Admin Navigation -->
  <div class="flex items-center space-x-1 border-b border-border">
    <a 
      href="/dashboard/admin" 
      class="flex items-center space-x-2 px-4 py-3 text-sm font-medium border-b-2 transition-colors"
      class:border-primary={$page.url.pathname === '/dashboard/admin'}
      class:text-primary={$page.url.pathname === '/dashboard/admin'}
      class:border-transparent={$page.url.pathname !== '/dashboard/admin'}
      class:text-muted-foreground={$page.url.pathname !== '/dashboard/admin'}
      class:hover:text-foreground={$page.url.pathname !== '/dashboard/admin'}
      class:hover:border-border={$page.url.pathname !== '/dashboard/admin'}
    >
      <Icon name="settings" size={16} />
      <span>Systeeminstellingen</span>
    </a>
    <a 
      href="/dashboard/admin/users" 
      class="flex items-center space-x-2 px-4 py-3 text-sm font-medium border-b-2 transition-colors"
      class:border-primary={$page.url.pathname === '/dashboard/admin/users'}
      class:text-primary={$page.url.pathname === '/dashboard/admin/users'}
      class:border-transparent={$page.url.pathname !== '/dashboard/admin/users'}
      class:text-muted-foreground={$page.url.pathname !== '/dashboard/admin/users'}
      class:hover:text-foreground={$page.url.pathname !== '/dashboard/admin/users'}
      class:hover:border-border={$page.url.pathname !== '/dashboard/admin/users'}
    >
      <Icon name="user" size={16} />
      <span>Gebruikersbeheer</span>
    </a>
    <a 
      href="/dashboard/admin/plugins" 
      class="flex items-center space-x-2 px-4 py-3 text-sm font-medium border-b-2 transition-colors"
      class:border-primary={$page.url.pathname === '/dashboard/admin/plugins'}
      class:text-primary={$page.url.pathname === '/dashboard/admin/plugins'}
      class:border-transparent={$page.url.pathname !== '/dashboard/admin/plugins'}
      class:text-muted-foreground={$page.url.pathname !== '/dashboard/admin/plugins'}
      class:hover:text-foreground={$page.url.pathname !== '/dashboard/admin/plugins'}
      class:hover:border-border={$page.url.pathname !== '/dashboard/admin/plugins'}
    >
      <Icon name="puzzle" size={16} />
      <span>Plugin Management</span>
    </a>
  </div>

  <!-- Page Content -->
  <div class="space-y-8">
    <slot />
  </div>
</div>