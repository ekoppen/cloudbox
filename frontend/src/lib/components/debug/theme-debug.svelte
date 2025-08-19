<script lang="ts">
  import { theme } from '$lib/stores/theme';
  import { onMount } from 'svelte';

  let themeInfo: any = {};

  function analyzeTheme() {
    if (typeof document === 'undefined') return;

    const html = document.documentElement;
    const computedStyle = getComputedStyle(html);
    
    themeInfo = {
      dataTheme: html.getAttribute('data-theme'),
      htmlClasses: html.classList.toString(),
      bodyClasses: document.body.classList.toString(),
      cssVariables: {
        background: computedStyle.getPropertyValue('--background'),
        foreground: computedStyle.getPropertyValue('--foreground'),
        primary: computedStyle.getPropertyValue('--primary'),
        secondary: computedStyle.getPropertyValue('--secondary'),
        accent: computedStyle.getPropertyValue('--accent'),
        muted: computedStyle.getPropertyValue('--muted'),
        card: computedStyle.getPropertyValue('--card'),
        border: computedStyle.getPropertyValue('--border'),
        sidebar: computedStyle.getPropertyValue('--sidebar'),
        'sidebar-foreground': computedStyle.getPropertyValue('--sidebar-foreground'),
        'sidebar-hover': computedStyle.getPropertyValue('--sidebar-hover')
      },
      daisyUIColors: {
        'base-100': computedStyle.getPropertyValue('--base-100') || 'Not found',
        'base-200': computedStyle.getPropertyValue('--base-200') || 'Not found',
        'base-content': computedStyle.getPropertyValue('--base-content') || 'Not found',
        primary: computedStyle.getPropertyValue('--p') || 'Not found',
        secondary: computedStyle.getPropertyValue('--s') || 'Not found'
      }
    };
  }

  onMount(() => {
    analyzeTheme();
    
    // Reanalyze when theme changes
    const unsubscribe = theme.subscribe(() => {
      setTimeout(analyzeTheme, 100);
    });
    
    return unsubscribe;
  });
</script>

<div class="p-4 bg-card border border-border rounded-lg text-xs font-mono">
  <h3 class="text-sm font-bold text-card-foreground mb-3">🔍 Theme Debug Info</h3>
  
  <div class="space-y-3">
    <div>
      <strong class="text-primary">Current Store:</strong>
      <div class="ml-2">
        Theme: <span class="text-accent">{$theme.theme}</span><br>
        Accent: <span class="text-accent">{$theme.accentColor}</span>
      </div>
    </div>

    <div>
      <strong class="text-primary">HTML Element:</strong>
      <div class="ml-2">
        data-theme: <span class="text-accent">{themeInfo.dataTheme || 'Not set'}</span><br>
        HTML classes: <span class="text-muted-foreground">{themeInfo.htmlClasses || 'None'}</span><br>
        Body classes: <span class="text-muted-foreground">{themeInfo.bodyClasses || 'None'}</span>
      </div>
    </div>

    <div>
      <strong class="text-primary">CSS Variables:</strong>
      <div class="ml-2 grid grid-cols-2 gap-2">
        {#each Object.entries(themeInfo.cssVariables || {}) as [key, value]}
          <div>
            <span class="text-secondary">--{key}:</span>
            <span class="text-muted-foreground">{value || 'Not found'}</span>
          </div>
        {/each}
      </div>
    </div>

    <div>
      <strong class="text-primary">DaisyUI Variables:</strong>
      <div class="ml-2">
        {#each Object.entries(themeInfo.daisyUIColors || {}) as [key, value]}
          <div>
            <span class="text-secondary">--{key}:</span>
            <span class="text-muted-foreground">{value}</span>
          </div>
        {/each}
      </div>
    </div>

    <div class="mt-4 pt-3 border-t border-border">
      <strong class="text-primary">Color Test:</strong>
      <div class="flex space-x-2 mt-2">
        <div class="w-6 h-6 bg-primary rounded border border-border" title="Primary"></div>
        <div class="w-6 h-6 bg-secondary rounded border border-border" title="Secondary"></div>
        <div class="w-6 h-6 bg-accent rounded border border-border" title="Accent"></div>
        <div class="w-6 h-6 bg-muted rounded border border-border" title="Muted"></div>
        <div class="w-6 h-6 bg-card rounded border border-border" title="Card"></div>
      </div>
    </div>

    <button
      on:click={analyzeTheme}
      class="w-full mt-3 bg-primary text-primary-foreground px-2 py-1 rounded text-xs hover:opacity-90"
    >
      🔄 Refresh Analysis
    </button>
  </div>
</div>