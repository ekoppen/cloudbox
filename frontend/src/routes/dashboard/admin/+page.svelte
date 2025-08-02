<script lang="ts">
  import { onMount } from 'svelte';
  import { API_BASE_URL } from '$lib/config';
  import { auth } from '$lib/stores/auth';
  import { toast } from '$lib/stores/toast';
  import Card from '$lib/components/ui/card.svelte';
  import Button from '$lib/components/ui/button.svelte';
  import Input from '$lib/components/ui/input.svelte';
  import Label from '$lib/components/ui/label.svelte';
  import Icon from '$lib/components/ui/icon.svelte';
  import Modal from '$lib/components/ui/modal.svelte';

  let settings = {};
  let instructions = null;
  let loading = true;
  let savingSettings = {};
  let settingTimeout;

  onMount(async () => {
    await loadSystemSettings();
  });

  async function loadSystemSettings() {
    loading = true;
    try {
      const response = await fetch(`${API_BASE_URL}/api/v1/admin/system/settings`, {
        headers: {
          'Content-Type': 'application/json',
          ...auth.getAuthHeader()
        }
      });

      if (response.ok) {
        const data = await response.json();
        settings = data.settings;
        instructions = data.instructions;
      } else {
        const error = await response.json();
        toast.error(error.error || 'Fout bij laden van systeeminstellingen');
      }
    } catch (error) {
      console.error('Error loading system settings:', error);
      toast.error('Netwerkfout bij laden systeeminstellingen');
    } finally {
      loading = false;
    }
  }

  async function updateSetting(key: string, value: string, settingName: string) {
    savingSettings[key] = true;
    
    try {
      const response = await fetch(`${API_BASE_URL}/api/v1/admin/system/settings/${key}`, {
        method: 'PUT',
        headers: {
          'Content-Type': 'application/json',
          ...auth.getAuthHeader()
        },
        body: JSON.stringify({ value })
      });

      if (response.ok) {
        toast.success(`${settingName} bijgewerkt`);
        // Reload instructions if domain/protocol changed
        if (key.includes('site_domain') || key.includes('site_protocol')) {
          await loadSystemSettings();
        }
      } else {
        const error = await response.json();
        toast.error(error.error || `Fout bij bijwerken ${settingName}`);
      }
    } catch (error) {
      console.error('Error updating setting:', error);
      toast.error(`Netwerkfout bij bijwerken ${settingName}`);
    } finally {
      savingSettings[key] = false;
    }
  }


  function getSettingValue(category: string, key: string) {
    const categorySettings = settings[category] || [];
    const setting = categorySettings.find(s => s.key === key);
    return setting?.value || setting?.default_value || '';
  }

  function handleSettingChange(category: string, key: string, value: string, settingName: string) {
    // Update local state immediately for better UX
    if (settings[category]) {
      const settingIndex = settings[category].findIndex(s => s.key === key);
      if (settingIndex !== -1) {
        settings[category][settingIndex].value = value;
      }
    }
    
    // Debounced save after user stops typing
    clearTimeout(settingTimeout);
    settingTimeout = setTimeout(() => {
      updateSetting(key, value, settingName);
    }, 1000);
  }

</script>

<svelte:head>
  <title>Admin Settings - CloudBox</title>
</svelte:head>

<div>

  {#if loading}
    <div class="text-center py-8">
      <div class="inline-block animate-spin rounded-full h-8 w-8 border-b-2 border-primary"></div>
      <p class="mt-2 text-muted-foreground">Laden van instellingen...</p>
    </div>
  {:else}
    <!-- General Settings -->
    {#if settings.general}
      <Card class="p-6">
        <div class="flex items-center space-x-3 mb-6">
          <Icon name="settings" size={24} />
          <h2 class="text-xl font-semibold text-card-foreground">Algemene Instellingen</h2>
        </div>

        <div class="grid grid-cols-1 md:grid-cols-2 gap-6">
          {#each settings.general as setting}
            <div class="space-y-2">
              <Label for={setting.key}>
                {setting.name}
                {#if setting.is_required}
                  <span class="text-red-500">*</span>
                {/if}
              </Label>
              {#if setting.value_type === 'boolean'}
                <select
                  id={setting.key}
                  class="w-full px-3 py-2 border border-border rounded-md bg-background text-foreground focus:outline-none focus:ring-2 focus:ring-primary focus:border-transparent"
                  value={setting.value}
                  on:change={(e) => handleSettingChange('general', setting.key, e.target.value, setting.name)}
                >
                  <option value="false">Uitgeschakeld</option>
                  <option value="true">Ingeschakeld</option>
                </select>
              {:else}
                <Input
                  id={setting.key}
                  type="text"
                  value={setting.value}
                  placeholder={setting.default_value}
                  on:input={(e) => handleSettingChange('general', setting.key, e.target.value, setting.name)}
                />
              {/if}
              {#if setting.description}
                <p class="text-xs text-muted-foreground">{setting.description}</p>
              {/if}
            </div>
          {/each}
        </div>
      </Card>
    {/if}

  {/if}
</div>

