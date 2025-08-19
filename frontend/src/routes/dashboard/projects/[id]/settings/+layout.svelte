<script lang="ts">
  import { page } from '$app/stores';
  import Icon from '$lib/components/ui/icon.svelte';

  $: projectId = $page.params.id;
</script>

<div class="space-y-6">
  <!-- Header -->
  <div class="flex items-center space-x-4">
    <div class="w-10 h-10 bg-primary/10 rounded-lg flex items-center justify-center">
      <Icon name="settings" size={20} className="text-primary" />
    </div>
    <div>
      <h1 class="text-2xl font-bold text-foreground">Project Instellingen</h1>
      <p class="text-sm text-muted-foreground">
        Beheer API keys, CORS configuratie en project instellingen
      </p>
    </div>
  </div>

  <!-- Tabs -->
  <div class="border-b border-border">
    <nav class="flex space-x-8">
      <a
        href="/dashboard/projects/{projectId}/settings"
        class="flex items-center space-x-2 py-2 px-1 border-b-2 text-sm font-medium transition-colors {
          $page.url.pathname === `/dashboard/projects/${projectId}/settings`
            ? 'border-primary text-primary' 
            : 'border-transparent text-muted-foreground hover:text-foreground hover:border-border'
        }"
      >
        <Icon name="settings" size={16} />
        <span>Algemeen</span>
      </a>
      <a
        href="/dashboard/projects/{projectId}/settings/plugins"
        class="flex items-center space-x-2 py-2 px-1 border-b-2 text-sm font-medium transition-colors {
          $page.url.pathname.includes('/settings/plugins') 
            ? 'border-primary text-primary' 
            : 'border-transparent text-muted-foreground hover:text-foreground hover:border-border'
        }"
      >
        <Icon name="package" size={16} />
        <span>Plugins</span>
      </a>
    </nav>
  </div>

  <!-- Content -->
  <slot />
</div>