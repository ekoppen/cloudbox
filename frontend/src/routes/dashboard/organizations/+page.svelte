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
      <div class="w-10 h-10 bg-primary/10 rounded-lg flex items-center justify-center">
        <Icon name="package" size={20} className="text-primary" />
      </div>
      <div>
        <h1 class="text-2xl font-bold text-foreground">Organizations</h1>
        <p class="text-sm text-muted-foreground">
          Organiseer je projecten in groepen voor beter overzicht
        </p>
      </div>
    </div>
    <Button on:click={() => showCreateModal = true} class="flex items-center space-x-2">
      <Icon name="package" size={16} />
      <span>Nieuwe Organization</span>
    </Button>
  </div>

  <!-- Organizations Grid -->
  <div class="space-y-6">
    {#if loading}
      <Card class="p-12">
        <div class="text-center">
          <div class="w-8 h-8 border-4 border-primary border-t-transparent rounded-full animate-spin mx-auto mb-4"></div>
          <p class="text-muted-foreground">Laden...</p>
        </div>
      </Card>
    {:else if organizations.length === 0}
      <Card class="p-12">
        <div class="text-center">
          <div class="w-16 h-16 bg-muted rounded-full flex items-center justify-center mx-auto mb-4">
            <Icon name="package" size={32} className="text-muted-foreground" />
          </div>
          <h3 class="text-lg font-medium text-foreground mb-2">Nog geen organizations</h3>
          <p class="text-muted-foreground mb-6 max-w-sm mx-auto">
            Maak je eerste organization aan om projecten beter te organiseren
          </p>
          <Button
            on:click={() => showCreateModal = true}
            size="lg"
            class="flex items-center space-x-2"
          >
            <Icon name="package" size={16} />
            <span>Eerste Organization Aanmaken</span>
          </Button>
        </div>
      </Card>
    {:else}
      <div class="grid gap-6 md:grid-cols-2 lg:grid-cols-3">
        {#each organizations as org}
          <Card class="group p-6 hover:shadow-lg hover:shadow-primary/5 transition-all duration-200 border hover:border-primary/20">
            <div class="flex items-start justify-between mb-4">
              <div class="flex items-center space-x-3">
                <div 
                  class="w-10 h-10 rounded-lg flex items-center justify-center text-white"
                  style="background-color: {org.color}"
                >
                  <Icon name="package" size={20} />
                </div>
                <div>
                  <h3 class="font-semibold text-foreground group-hover:text-primary transition-colors">
                    {org.name}
                  </h3>
                  <p class="text-xs text-muted-foreground">
                    {org.project_count} project{org.project_count !== 1 ? 'en' : ''}
                  </p>
                </div>
              </div>
              <div class="opacity-0 group-hover:opacity-100 transition-opacity">
                <Button
                  variant="ghost"
                  size="sm"
                  on:click={() => deleteOrganization(org.id)}
                  class="text-destructive hover:text-destructive"
                >
                  <Icon name="backup" size={14} />
                </Button>
              </div>
            </div>

            {#if org.description}
              <p class="text-sm text-muted-foreground mb-4 line-clamp-2">
                {org.description}
              </p>
            {/if}

            <div class="flex items-center justify-between text-xs text-muted-foreground">
              <span>Aangemaakt {formatDate(org.created_at)}</span>
              <div class="flex items-center space-x-1">
                <div 
                  class="w-2 h-2 rounded-full"
                  style="background-color: {org.color}"
                ></div>
                <span class="capitalize">
                  {org.is_active ? 'Actief' : 'Inactief'}
                </span>
              </div>
            </div>
          </Card>
        {/each}
      </div>
    {/if}
  </div>
</div>

<!-- Create Organization Modal -->
{#if showCreateModal}
  <div class="fixed inset-0 bg-black/80 backdrop-blur-sm flex items-center justify-center p-4 z-50">
    <Card class="max-w-md w-full p-6 border-2 shadow-2xl">
      <div class="flex items-center space-x-3 mb-6">
        <div class="w-10 h-10 bg-primary/10 rounded-lg flex items-center justify-center">
          <Icon name="package" size={20} className="text-primary" />
        </div>
        <h2 class="text-xl font-bold text-foreground">Nieuwe Organization</h2>
      </div>
      
      <form on:submit|preventDefault={createOrganization} class="space-y-4">
        <div>
          <Label for="org-name">Naam</Label>
          <Input
            id="org-name"
            type="text"
            bind:value={newOrgName}
            required
            placeholder="Bijv. Mijn Bedrijf"
            class="mt-1"
          />
        </div>

        <div>
          <Label for="org-description">Beschrijving</Label>
          <Textarea
            id="org-description"
            bind:value={newOrgDescription}
            placeholder="Optionele beschrijving van de organization"
            class="mt-1"
            rows={3}
          />
        </div>

        <div>
          <Label>Kleur</Label>
          <div class="flex flex-wrap gap-2 mt-2">
            {#each colorOptions as color}
              <button
                type="button"
                class="w-8 h-8 rounded-full border-2 transition-all {newOrgColor === color ? 'border-foreground scale-110' : 'border-transparent hover:scale-105'}"
                style="background-color: {color}"
                on:click={() => newOrgColor = color}
              ></button>
            {/each}
          </div>
        </div>

        <div class="flex justify-end space-x-3 pt-4">
          <Button
            type="button"
            variant="outline"
            on:click={() => showCreateModal = false}
          >
            Annuleren
          </Button>
          <Button type="submit">
            Organization Aanmaken
          </Button>
        </div>
      </form>
    </Card>
  </div>
{/if}