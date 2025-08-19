<script lang="ts">
  import Icon from './icon.svelte';
  import Badge from './badge.svelte';

  export let project: {
    id: number;
    name: string;
    description?: string;
    slug: string;
    created_at: string;
    is_active: boolean;
    organization?: {
      id: number;
      name: string;
      color: string;
    };
  };

  function formatDate(dateStr: string) {
    return new Date(dateStr).toLocaleDateString('nl-NL', {
      year: 'numeric',
      month: 'short',
      day: 'numeric'
    });
  }

  function getProjectUrl(projectId: number) {
    return `/dashboard/projects/${projectId}`;
  }

  function handleProjectClick(event: Event) {
    // Log navigation for debugging
    console.log('Project card clicked, navigating to:', getProjectUrl(project.id));
    console.log('Project data:', project);
  }
</script>

<a 
  href={getProjectUrl(project.id)} 
  on:click={handleProjectClick}
  class="group relative flex flex-col rounded-xl border border-border bg-card p-6 transition-all duration-200 hover:border-primary/30 hover:shadow-brand-lg cursor-pointer block no-underline"
>
  <!-- Status indicator -->
  <div class="absolute right-4 top-4">
    <div class="h-2 w-2 rounded-full {project.is_active ? 'bg-success' : 'bg-gray-300'}"></div>
  </div>

  <!-- Project header -->
  <div class="mb-4">
    <div class="flex items-start justify-between">
      <div class="flex items-center space-x-3">
        <div class="flex h-12 w-12 items-center justify-center rounded-lg bg-primary/10">
          <Icon name="package" size={20} className="text-primary" />
        </div>
        <div>
          <h3 class="text-lg font-semibold text-card-foreground group-hover:text-primary transition-colors">
            {project.name}
          </h3>
          {#if project.organization}
            <div class="flex items-center space-x-1.5 mt-1">
              <div 
                class="h-2 w-2 rounded-full"
                style="background-color: {project.organization.color}"
              ></div>
              <span class="text-sm text-muted-foreground">{project.organization.name}</span>
            </div>
          {/if}
        </div>
      </div>
    </div>
  </div>

  <!-- Description -->
  <div class="flex-1 mb-4">
    {#if project.description}
      <p class="text-sm text-muted-foreground leading-relaxed line-clamp-2">
        {project.description}
      </p>
    {:else}
      <p class="text-sm text-muted-foreground italic">Geen beschrijving</p>
    {/if}
  </div>

  <!-- Project details -->
  <div class="mb-4 space-y-2">
    <div class="flex items-center justify-between text-xs">
      <span class="text-muted-foreground">API Slug</span>
      <code class="rounded bg-muted px-2 py-0.5 font-mono text-xs">{project.slug}</code>
    </div>
    <div class="flex items-center justify-between text-xs">
      <span class="text-muted-foreground">Aangemaakt</span>
      <span class="text-card-foreground">{formatDate(project.created_at)}</span>
    </div>
    <div class="flex items-center justify-between text-xs">
      <span class="text-muted-foreground">Status</span>
      <Badge variant={project.is_active ? "default" : "secondary"} class="text-xs">
        {project.is_active ? 'Actief' : 'Inactief'}
      </Badge>
    </div>
  </div>

  <!-- Actions -->
  <div class="mt-auto pt-4 border-t border-border/50">
    <div class="inline-flex w-full items-center justify-center space-x-2 rounded-lg bg-primary/10 px-4 py-2.5 text-sm font-medium text-primary transition-colors group-hover:bg-primary group-hover:text-primary-foreground">
      <Icon name="arrow-right" size={16} />
      <span>Project openen</span>
    </div>
  </div>

  <!-- Hover indicator -->
  <div class="absolute inset-0 rounded-xl ring-1 ring-transparent transition-all duration-200 group-hover:ring-primary/20"></div>
</a>