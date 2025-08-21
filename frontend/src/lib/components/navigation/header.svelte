<script lang="ts">
  import { auth } from '$lib/stores/auth';
  import { theme } from '$lib/stores/theme';
  import { goto } from '$app/navigation';
  import Icon from '$lib/components/ui/icon.svelte';
  import Breadcrumbs from './breadcrumbs.svelte';

  let showUserDropdown = false;

  function handleLogout() {
    auth.logout();
    goto('/');
  }

  function getInitials(name: string): string {
    return name
      .split(' ')
      .map(word => word.charAt(0).toUpperCase())
      .slice(0, 2)
      .join('');
  }

  // Close dropdown when clicking outside
  function handleClickOutside(event: MouseEvent) {
    if (!event.target || !(event.target as Element).closest('.user-dropdown')) {
      showUserDropdown = false;
    }
  }

  $: user = $auth.user;
</script>

<svelte:document on:click={handleClickOutside} />

<!-- Header -->
<header class="sticky top-0 z-40 bg-card/95 backdrop-blur-xl supports-[backdrop-filter]:bg-card/85 shadow-sm">
  <div class="flex h-16 items-center justify-between px-6">
    <!-- Breadcrumbs -->
    <div class="flex items-center space-x-4">
      <Breadcrumbs />
    </div>

    <!-- Right side actions -->
    <div class="flex items-center space-x-3">
      <!-- Theme toggle switch -->
      <div class="theme-toggle-container">
        <label class="theme-toggle-switch" title="Schakel tussen licht en donker thema">
          <input 
            type="checkbox" 
            checked={$theme.theme === 'cloudbox-dark'}
            on:change={() => theme.toggleTheme()}
            class="theme-toggle-input"
          />
          <span class="theme-toggle-slider">
            <Icon name="sun" size={12} className="theme-toggle-icon theme-toggle-sun" />
            <Icon name="moon" size={12} className="theme-toggle-icon theme-toggle-moon" />
          </span>
        </label>
      </div>
      
      <!-- User dropdown -->
      <div class="relative user-dropdown">
        <button
          on:click={() => showUserDropdown = !showUserDropdown}
          class="flex items-center space-x-3 rounded-lg px-3 py-2 text-sm font-medium transition-all duration-200 hover:bg-accent/10 focus:bg-accent/10 focus:outline-none text-foreground border border-transparent hover:border-border focus:border-primary"
          title="Gebruikersmenu"
        >
          <div class="flex items-center space-x-2">
            <div class="avatar placeholder">
              <div class="w-8 h-8 rounded-full bg-primary text-primary-content text-sm font-semibold shadow-sm">
                <span>{getInitials(user?.name || 'User')}</span>
              </div>
            </div>
            <div class="hidden sm:block text-left min-w-0">
              <div class="text-sm font-medium text-foreground truncate max-w-32">
                {user?.name || 'User'}
              </div>
              <div class="text-xs text-muted-foreground truncate max-w-32">
                {user?.email || ''}
              </div>
            </div>
          </div>
          <Icon name="chevron-down" size={14} className="transition-transform duration-200 icon-contrast {showUserDropdown ? 'rotate-180' : ''}" />
        </button>

        {#if showUserDropdown}
          <div class="dropdown-content menu bg-card rounded-lg z-[1] w-64 p-3 shadow-xl border border-border absolute right-0 mt-2">
            <div class="px-3 py-2 border-b border-border mb-2">
              <div class="flex items-center space-x-3">
                <div class="avatar placeholder">
                  <div class="w-10 h-10 rounded-full bg-primary text-primary-content text-sm font-semibold">
                    <span>{getInitials(user?.name || 'User')}</span>
                  </div>
                </div>
                <div class="min-w-0">
                  <div class="text-sm font-semibold text-foreground truncate">{user?.name}</div>
                  <div class="text-xs text-muted-foreground truncate">{user?.email}</div>
                  {#if user?.role === 'superadmin'}
                    <div class="text-xs font-medium text-primary mt-1">Administrator</div>
                  {/if}
                </div>
              </div>
            </div>
            <div class="space-y-1">
              <button
                on:click={() => { 
                  showUserDropdown = false; 
                  goto('/dashboard/settings'); 
                }}
                class="flex items-center space-x-3 w-full px-3 py-2 text-sm text-foreground rounded-md hover:bg-accent/10 transition-colors"
              >
                <Icon name="user" size={16} className="icon-contrast" />
                <span>Profiel</span>
              </button>
              {#if user?.role === 'superadmin'}
                <button
                  on:click={() => { 
                    showUserDropdown = false; 
                    goto('/dashboard/admin'); 
                  }}
                  class="flex items-center space-x-3 w-full px-3 py-2 text-sm text-foreground rounded-md hover:bg-accent/10 transition-colors"
                >
                  <Icon name="shield-check" size={16} className="icon-contrast" />
                  <span>Admin</span>
                </button>
              {/if}
              <hr class="my-2 border-border">
              <button
                on:click={() => { 
                  showUserDropdown = false; 
                  handleLogout(); 
                }}
                class="flex items-center space-x-3 w-full px-3 py-2 text-sm text-destructive rounded-md hover:bg-destructive/10 transition-colors"
              >
                <Icon name="log-out" size={16} className="icon-contrast" />
                <span>Uitloggen</span>
              </button>
            </div>
          </div>
        {/if}
      </div>
    </div>
  </div>
</header>

<style>
  .theme-toggle-container {
    display: flex;
    align-items: center;
  }

  .theme-toggle-switch {
    position: relative;
    display: inline-block;
    width: 52px;
    height: 28px;
    cursor: pointer;
  }

  .theme-toggle-input {
    opacity: 0;
    width: 0;
    height: 0;
    position: absolute;
  }

  .theme-toggle-slider {
    position: absolute;
    top: 0;
    left: 0;
    right: 0;
    bottom: 0;
    background-color: hsl(var(--muted));
    border: 2px solid hsl(var(--border));
    transition: all 0.3s ease;
    border-radius: 24px;
    display: flex;
    align-items: center;
    justify-content: space-between;
    padding: 0 6px;
    box-shadow: inset 0 1px 3px rgba(0, 0, 0, 0.1);
  }

  .theme-toggle-slider:before {
    position: absolute;
    content: "";
    height: 18px;
    width: 18px;
    left: 3px;
    bottom: 3px;
    background-color: hsl(var(--background));
    border: 1px solid hsl(var(--border));
    transition: all 0.3s cubic-bezier(0.4, 0.0, 0.2, 1);
    border-radius: 50%;
    box-shadow: 0 2px 4px rgba(0, 0, 0, 0.2);
    z-index: 2;
  }

  .theme-toggle-input:checked + .theme-toggle-slider {
    background-color: hsl(var(--primary) / 0.2);
    border-color: hsl(var(--primary));
  }

  .theme-toggle-input:checked + .theme-toggle-slider:before {
    transform: translateX(24px);
    background-color: hsl(var(--primary));
    border-color: hsl(var(--primary));
    box-shadow: 0 2px 8px rgba(0, 0, 0, 0.3);
  }

  .theme-toggle-icon {
    transition: all 0.3s ease;
    z-index: 1;
    position: relative;
  }

  .theme-toggle-sun {
    color: hsl(var(--warning));
    opacity: 1;
  }

  .theme-toggle-moon {
    color: hsl(var(--primary));
    opacity: 0.4;
  }

  .theme-toggle-input:checked + .theme-toggle-slider .theme-toggle-sun {
    opacity: 0.4;
  }

  .theme-toggle-input:checked + .theme-toggle-slider .theme-toggle-moon {
    opacity: 1;
    color: hsl(var(--primary-foreground));
  }

  .theme-toggle-switch:hover .theme-toggle-slider {
    box-shadow: inset 0 1px 3px rgba(0, 0, 0, 0.1), 0 0 0 2px hsl(var(--ring) / 0.2);
  }

  .theme-toggle-switch:focus-within .theme-toggle-slider {
    box-shadow: inset 0 1px 3px rgba(0, 0, 0, 0.1), 0 0 0 2px hsl(var(--ring));
  }

  /* Enhanced styles for dark mode */
  :global(.cloudbox-dark) .theme-toggle-slider {
    background-color: hsl(var(--muted));
    border-color: hsl(var(--border));
    box-shadow: inset 0 1px 3px rgba(255, 255, 255, 0.1);
  }

  :global(.cloudbox-dark) .theme-toggle-input:checked + .theme-toggle-slider {
    background-color: hsl(var(--primary) / 0.3);
    border-color: hsl(var(--primary));
  }

  :global(.cloudbox-dark) .theme-toggle-slider:before {
    background-color: hsl(var(--card));
    border-color: hsl(var(--border));
    box-shadow: 0 2px 4px rgba(0, 0, 0, 0.4);
  }

  :global(.cloudbox-dark) .theme-toggle-input:checked + .theme-toggle-slider:before {
    background-color: hsl(var(--primary));
    border-color: hsl(var(--primary));
    box-shadow: 0 2px 8px rgba(0, 0, 0, 0.5);
  }
</style>