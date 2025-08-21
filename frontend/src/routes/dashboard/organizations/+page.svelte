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

  .glassmorphism-org-card {
    background: rgba(255, 255, 255, 0.9);
    backdrop-filter: blur(20px);
    border: 1px solid rgba(255, 255, 255, 0.3);
    border-radius: 16px;
    padding: 24px;
    transition: all 0.3s cubic-bezier(0.4, 0.0, 0.2, 1);
    position: relative;
    overflow: hidden;
    cursor: pointer;
  }

  .glassmorphism-org-card:hover {
    transform: translateY(-4px);
    box-shadow: 
      0 15px 40px -12px rgba(0, 0, 0, 0.2),
      0 8px 25px -8px rgba(0, 0, 0, 0.15);
  }

  .glassmorphism-org-card::before {
    content: '';
    position: absolute;
    top: 0;
    left: 0;
    right: 0;
    bottom: 0;
    background: linear-gradient(
      135deg,
      rgba(255, 255, 255, 0.15) 0%,
      rgba(255, 255, 255, 0) 50%,
      rgba(0, 0, 0, 0.05) 100%
    );
    pointer-events: none;
    z-index: 1;
  }

  .glassmorphism-org-card > * {
    position: relative;
    z-index: 2;
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

  :global(.cloudbox-dark) .glassmorphism-org-card {
    background: rgba(33, 33, 33, 0.8);
    border: 1px solid rgba(255, 255, 255, 0.15);
    backdrop-filter: blur(30px);
  }

  :global(.cloudbox-dark) .glassmorphism-org-card::before {
    background: linear-gradient(
      135deg,
      rgba(255, 255, 255, 0.05) 0%,
      rgba(255, 255, 255, 0) 50%,
      rgba(0, 0, 0, 0.15) 100%
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

    .glassmorphism-org-card {
      padding: 20px;
      border-radius: 12px;
    }
  }

  /* Reduce motion for accessibility */
  @media (prefers-reduced-motion: reduce) {
    .glassmorphism-card,
    .glassmorphism-icon,
    .glassmorphism-org-card {
      transition: none;
    }
    
    .glassmorphism-card:hover,
    .glassmorphism-org-card:hover {
      transform: none;
    }
  }
</style>

<script lang="ts">
  import { onMount } from 'svelte';
  import { auth } from '$lib/stores/auth';
  import { toastStore } from '$lib/stores/toast';
  import { API_ENDPOINTS, createApiRequest } from '$lib/config';
  import Card from '$lib/components/ui/card.svelte';
  import Button from '$lib/components/ui/button.svelte';
  import Input from '$lib/components/ui/input.svelte';
  import Label from '$lib/components/ui/label.svelte';
  import Textarea from '$lib/components/ui/textarea.svelte';
  import Icon from '$lib/components/ui/icon.svelte';

  interface Organization {
    id: number;
    name: string;
    description: string;
    color: string;
    is_active: boolean;
    project_count: number;
    created_at: string;
  }

  let organizations: Organization[] = [];
  let loading = true;
  let showCreateModal = false;
  let newOrgName = '';
  let newOrgDescription = '';
  let newOrgColor = '#3B82F6';

  const colorOptions = [
    '#3B82F6', // Blue
    '#10B981', // Green  
    '#F59E0B', // Yellow
    '#EF4444', // Red
    '#8B5CF6', // Purple
    '#F97316', // Orange
    '#06B6D4', // Cyan
    '#84CC16', // Lime
  ];

  onMount(async () => {
    await loadOrganizations();
  });

  async function loadOrganizations() {
    try {
      const response = await createApiRequest(API_ENDPOINTS.organizations.list, {
        headers: {
          'Authorization': `Bearer ${$auth.token}`,
        },
      });

      if (response.ok) {
        organizations = await response.json();
      } else {
        toastStore.error('Fout bij laden van organizations');
      }
    } catch (err) {
      toastStore.error('Netwerkfout bij laden van organizations');
    } finally {
      loading = false;
    }
  }

  async function createOrganization() {
    if (!newOrgName.trim()) {
      toastStore.error('Naam is verplicht');
      return;
    }

    try {
      const response = await createApiRequest(API_ENDPOINTS.organizations.create, {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
          'Authorization': `Bearer ${$auth.token}`,
        },
        body: JSON.stringify({
          name: newOrgName.trim(),
          description: newOrgDescription.trim(),
          color: newOrgColor,
        }),
      });

      if (response.ok) {
        const newOrg = await response.json();
        organizations = [...organizations, newOrg];
        toastStore.success('Organization succesvol aangemaakt');
        
        // Reset form
        newOrgName = '';
        newOrgDescription = '';
        newOrgColor = '#3B82F6';
        showCreateModal = false;
      } else {
        const data = await response.json();
        toastStore.error(data.error || 'Fout bij aanmaken van organization');
      }
    } catch (err) {
      toastStore.error('Netwerkfout bij aanmaken van organization');
    }
  }

  async function deleteOrganization(orgId: number) {
    if (confirm('Weet je zeker dat je deze organization wilt verwijderen?\n\nAlle projecten in deze organization worden verplaatst naar "Geen Organization".')) {
      try {
        const response = await createApiRequest(API_ENDPOINTS.organizations.delete(orgId.toString()), {
          method: 'DELETE',
          headers: {
            'Authorization': `Bearer ${$auth.token}`,
          },
        });

        if (response.ok) {
          organizations = organizations.filter(org => org.id !== orgId);
          toastStore.success('Organization succesvol verwijderd');
        } else {
          const data = await response.json();
          toastStore.error(data.error || 'Fout bij verwijderen van organization');
        }
      } catch (err) {
        toastStore.error('Netwerkfout bij verwijderen van organization');
      }
    }
  }

  function formatDate(dateStr: string): string {
    return new Date(dateStr).toLocaleDateString('nl-NL', {
      year: 'numeric',
      month: 'short',
      day: 'numeric'
    });
  }
</script>

<svelte:head>
  <title>Organizations - CloudBox</title>
</svelte:head>

<div class="space-y-6">
  <!-- Header -->
  <div class="flex items-center justify-between">
    <div class="flex items-center space-x-4">
      <div class="glassmorphism-icon bg-primary/20">
        <Icon name="building" size={20} className="text-primary" />
      </div>
      <div>
        <h1 class="text-2xl font-bold text-foreground">Organizations</h1>
        <p class="text-sm text-muted-foreground">
          Organiseer je projecten in groepen voor beter overzicht
        </p>
      </div>
    </div>
    <Button 
      on:click={() => showCreateModal = true}
      variant="floating"
      size="icon-lg"
      iconOnly={true}
      tooltip="Nieuwe Organization"
    >
      <Icon name="plus" size={20} />
    </Button>
  </div>

  <!-- Organizations Grid -->
  <div class="space-y-6">
    {#if loading}
      <div class="glassmorphism-card glassmorphism-card-primary p-12">
        <div class="text-center">
          <div class="w-8 h-8 border-4 border-primary border-t-transparent rounded-full animate-spin mx-auto mb-4"></div>
          <p class="text-muted-foreground">Laden...</p>
        </div>
      </div>
    {:else if organizations.length === 0}
      <div class="glassmorphism-card glassmorphism-card-success p-12">
        <div class="text-center">
          <div class="glassmorphism-icon w-16 h-16 bg-muted/30 mx-auto mb-4">
            <Icon name="building" size={32} className="text-muted-foreground" />
          </div>
          <h3 class="text-lg font-medium text-foreground mb-2">Nog geen organizations</h3>
          <p class="text-muted-foreground mb-6 max-w-sm mx-auto">
            Maak je eerste organization aan om projecten beter te organiseren
          </p>
          <Button
            on:click={() => showCreateModal = true}
            variant="floating"
            size="icon-lg"
            iconOnly={true}
            tooltip="Eerste Organization Aanmaken"
          >
            <Icon name="plus" size={20} />
          </Button>
        </div>
      </div>
    {:else}
      <div class="grid gap-6 md:grid-cols-2 lg:grid-cols-3">
        {#each organizations as org}
          <div class="glassmorphism-org-card group">
            <div class="flex items-start justify-between mb-4">
              <div class="flex items-center space-x-3">
                <div 
                  class="glassmorphism-icon text-white"
                  style="background-color: {org.color}"
                >
                  <Icon name="building" size={20} />
                </div>
                <div>
                  <h3 class="font-semibold text-foreground group-hover:text-primary transition-colors">
                    {org.name}
                  </h3>
                  <p class="text-xs text-muted-foreground">
                    {org.project_count || 0} project{(org.project_count || 0) !== 1 ? 'en' : ''}
                  </p>
                </div>
              </div>
              <div class="opacity-0 group-hover:opacity-100 transition-opacity">
                <Button
                  variant="floating"
                  size="icon-sm"
                  iconOnly={true}
                  tooltip="Organization Verwijderen"
                  on:click={() => deleteOrganization(org.id)}
                  class="text-destructive hover:text-destructive bg-destructive/10 hover:bg-destructive/20"
                >
                  <Icon name="trash" size={14} />
                </Button>
              </div>
            </div>

            <div class="flex-1 mb-4">
              {#if org.description}
                <p class="text-sm text-muted-foreground line-clamp-2">
                  {org.description}
                </p>
              {:else}
                <div class="h-10"></div>
              {/if}
            </div>

            <div class="flex items-center justify-between">
              <div class="flex space-x-2">
                <Button
                  href="/dashboard/organizations/{org.id}"
                  variant="secondary"
                  size="sm"
                  class="flex items-center space-x-2"
                >
                  <Icon name="eye" size={14} />
                  <span>Bekijken</span>
                </Button>
              </div>
              <div class="text-right">
                <div class="flex items-center justify-end space-x-1 text-xs text-muted-foreground mb-1">
                  <div 
                    class="w-2 h-2 rounded-full"
                    style="background-color: {org.color}"
                  ></div>
                  <span class="capitalize">
                    {org.is_active ? 'Actief' : 'Inactief'}
                  </span>
                </div>
                <span class="text-xs text-muted-foreground">
                  {formatDate(org.created_at)}
                </span>
              </div>
            </div>
          </div>
        {/each}
      </div>
    {/if}
  </div>
</div>

<!-- Create Organization Modal -->
{#if showCreateModal}
  <div class="fixed inset-0 modal-backdrop-enhanced flex items-start justify-center p-4 pt-16 sm:pt-20 overflow-y-auto z-50"
       on:click={() => showCreateModal = false}>
    <div class="glassmorphism-card max-w-md w-full border-2 shadow-2xl my-auto modal-content-wrapper"
         on:click|stopPropagation>
      <div class="flex items-center space-x-3 mb-6">
        <div class="glassmorphism-icon bg-primary/20">
          <Icon name="building" size={20} className="text-primary" />
        </div>
        <h2 class="text-xl font-bold text-foreground">Nieuwe Organization</h2>
      </div>
      
      <form on:submit|preventDefault={createOrganization} class="space-y-6">
        <div class="space-y-2">
          <Label for="org-name" class="flex items-center space-x-2">
            <Icon name="type" size={14} />
            <span>Organization naam</span>
          </Label>
          <Input
            id="org-name"
            type="text"
            bind:value={newOrgName}
            required
            placeholder="Bijv. Mijn Bedrijf"
            class="pl-4"
          />
          <p class="text-xs text-muted-foreground">
            Deze naam wordt gebruikt voor je organization identificatie
          </p>
        </div>

        <div class="space-y-2">
          <Label for="org-description" class="flex items-center space-x-2">
            <Icon name="edit" size={14} />
            <span>Beschrijving (optioneel)</span>
          </Label>
          <Textarea
            id="org-description"
            bind:value={newOrgDescription}
            placeholder="Een korte beschrijving van je organization..."
            class="pl-4"
            rows={3}
          />
        </div>

        <div class="space-y-2">
          <Label class="flex items-center space-x-2">
            <Icon name="palette" size={14} />
            <span>Kleur thema</span>
          </Label>
          <div class="flex flex-wrap gap-2 mt-2">
            {#each colorOptions as color}
              <button
                type="button"
                class="w-8 h-8 rounded-full border-2 transition-all {newOrgColor === color ? 'border-foreground scale-110' : 'border-transparent hover:scale-105'}"
                style="background-color: {color}"
                aria-label="Select color {color}"
                on:click={() => newOrgColor = color}
              ></button>
            {/each}
          </div>
          <p class="text-xs text-muted-foreground">
            Kies een kleur om je organization visueel te onderscheiden
          </p>
        </div>

        <!-- Organization Features Preview -->
        <div class="glassmorphism-card bg-muted/30 p-4 space-y-3">
          <div class="flex items-center space-x-2 text-sm font-medium text-foreground">
            <Icon name="building" size={16} className="text-primary" />
            <span>Jouw organization krijgt:</span>
          </div>
          <div class="grid grid-cols-2 gap-2 text-xs">
            <div class="flex items-center space-x-2">
              <Icon name="package" size={12} className="text-blue-600" />
              <span class="text-muted-foreground">Project organisatie</span>
            </div>
            <div class="flex items-center space-x-2">
              <Icon name="users" size={12} className="text-green-600" />
              <span class="text-muted-foreground">Team samenwerking</span>
            </div>
            <div class="flex items-center space-x-2">
              <Icon name="settings" size={12} className="text-purple-600" />
              <span class="text-muted-foreground">Gedeelde instellingen</span>
            </div>
            <div class="flex items-center space-x-2">
              <Icon name="shield-check" size={12} className="text-orange-600" />
              <span class="text-muted-foreground">Toegangsbeheer</span>
            </div>
          </div>
        </div>

        <div class="flex space-x-3 pt-4">
          <Button
            type="button"
            variant="secondary"
            on:click={() => {
              showCreateModal = false;
              newOrgName = '';
              newOrgDescription = '';
              newOrgColor = '#3B82F6';
            }}
            class="flex-1 flex items-center justify-center space-x-2"
          >
            <Icon name="x" size={14} />
            <span>Annuleren</span>
          </Button>
          <Button 
            type="submit"
            variant="primary"
            disabled={!newOrgName.trim()}
            class="flex-1 flex items-center justify-center space-x-2"
          >
            <Icon name="plus" size={14} />
            <span>Organization Aanmaken</span>
          </Button>
        </div>
      </form>
    </div>
  </div>
{/if}