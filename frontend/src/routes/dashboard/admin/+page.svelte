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
  let showInstructionsModal = false;
  let testingOAuth = false;
  let oauthTestResult = null;
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

  async function testGitHubOAuth() {
    testingOAuth = true;
    oauthTestResult = null;
    
    try {
      const response = await fetch(`${API_BASE_URL}/api/v1/admin/system/test-github-oauth`, {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
          ...auth.getAuthHeader()
        }
      });

      if (response.ok) {
        oauthTestResult = await response.json();
        if (oauthTestResult.status === 'configured') {
          toast.success('✅ GitHub OAuth configuratie is correct!');
        }
      } else {
        const error = await response.json();
        oauthTestResult = error;
        if (error.status === 'disabled') {
          toast.error('❌ GitHub OAuth is uitgeschakeld');
        } else if (error.status === 'incomplete') {
          toast.error('❌ GitHub OAuth configuratie is incompleet');
        } else {
          toast.error(error.error || 'Fout bij testen GitHub OAuth');
        }
      }
    } catch (error) {
      console.error('Error testing GitHub OAuth:', error);
      toast.error('Netwerkfout bij testen GitHub OAuth');
    } finally {
      testingOAuth = false;
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

  function showInstructions() {
    showInstructionsModal = true;
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

    <!-- GitHub OAuth Settings -->
    {#if settings.github}
      <Card class="p-6">
        <div class="flex items-center justify-between mb-6">
          <div class="flex items-center space-x-3">
            <Icon name="github" size={24} />
            <h2 class="text-xl font-semibold text-card-foreground">GitHub OAuth Instellingen</h2>
          </div>
          <div class="flex space-x-2">
            <Button on:click={testGitHubOAuth} size="sm" variant="outline" disabled={testingOAuth}>
              {#if testingOAuth}
                <div class="animate-spin rounded-full h-4 w-4 border-b-2 border-primary mr-2"></div>
                Testen...
              {:else}
                <Icon name="shield" size={16} className="mr-2" />
                Test Configuratie
              {/if}
            </Button>
            <Button on:click={showInstructions} size="sm">
              <Icon name="info" size={16} className="mr-2" />
              Setup Guide
            </Button>
          </div>
        </div>

        <div class="grid grid-cols-1 gap-6">
          {#each settings.github as setting}
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
                  on:change={(e) => handleSettingChange('github', setting.key, e.target.value, setting.name)}
                >
                  <option value="false">Uitgeschakeld</option>
                  <option value="true">Ingeschakeld</option>
                </select>
              {:else}
                <Input
                  id={setting.key}
                  type={setting.is_secret ? 'password' : 'text'}
                  value={setting.value}
                  placeholder={setting.is_secret && setting.value ? '••••••••' : setting.default_value}
                  on:input={(e) => handleSettingChange('github', setting.key, e.target.value, setting.name)}
                />
              {/if}
              {#if setting.description}
                <p class="text-xs text-muted-foreground">{setting.description}</p>
              {/if}
            </div>
          {/each}
        </div>

        <!-- OAuth Test Result -->
        {#if oauthTestResult}
          <div class="mt-6 p-4 rounded-lg border {oauthTestResult.status === 'configured' ? 'bg-green-50 border-green-200' : 'bg-red-50 border-red-200'}">
            <div class="flex items-start space-x-3">
              <Icon 
                name={oauthTestResult.status === 'configured' ? 'check' : 'x'} 
                size={20} 
                className={oauthTestResult.status === 'configured' ? 'text-green-600 mt-0.5' : 'text-red-600 mt-0.5'} 
              />
              <div class="flex-1">
                <h4 class="font-medium {oauthTestResult.status === 'configured' ? 'text-green-800' : 'text-red-800'}">
                  {#if oauthTestResult.status === 'configured'}
                    ✅ GitHub OAuth Configuratie OK
                  {:else if oauthTestResult.status === 'disabled'}
                    ❌ GitHub OAuth Uitgeschakeld
                  {:else if oauthTestResult.status === 'incomplete'}
                    ❌ Configuratie Incompleet
                  {:else}
                    ❌ Configuratie Fout
                  {/if}
                </h4>
                {#if oauthTestResult.callback_url}
                  <p class="text-sm {oauthTestResult.status === 'configured' ? 'text-green-700' : 'text-red-700'} mt-1">
                    Callback URL: <code class="bg-white px-1 rounded text-xs">{oauthTestResult.callback_url}</code>
                  </p>
                {/if}
                {#if oauthTestResult.error}
                  <p class="text-sm text-red-700 mt-1">{oauthTestResult.error}</p>
                {/if}
              </div>
            </div>
          </div>
        {/if}
      </Card>
    {/if}
  {/if}
</div>

<!-- Instructions Modal -->
{#if showInstructionsModal && instructions}
  <Modal open={showInstructionsModal} on:close={() => showInstructionsModal = false} size="3xl">
    <div class="p-8 max-h-[80vh] overflow-y-auto">
      <div class="flex items-center space-x-3 mb-6">
        <Icon name="github" size={32} />
        <div>
          <h2 class="text-2xl font-semibold">GitHub OAuth Setup</h2>
          <p class="text-muted-foreground">Configureer per-repository OAuth autorisatie</p>
        </div>
      </div>

      <div class="space-y-8">
        {#each instructions.steps as step}
          <div class="flex space-x-4">
            <div class="flex-shrink-0 w-8 h-8 bg-primary text-primary-foreground rounded-full flex items-center justify-center font-medium text-sm">
              {step.step}
            </div>
            <div class="flex-1">
              <h3 class="font-semibold text-lg mb-2">{step.title}</h3>
              <p class="text-muted-foreground mb-4">{step.description}</p>
              
              {#if step.url}
                <a 
                  href={step.url} 
                  target="_blank" 
                  class="inline-flex items-center text-primary hover:text-primary/80 mb-4"
                >
                  <Icon name="external-link" size={16} className="mr-2" />
                  {step.url}
                </a>
              {/if}

              {#if step.fields}
                <div class="bg-muted p-4 rounded-lg space-y-3">
                  <h4 class="font-medium">Vul de volgende velden in:</h4>
                  {#each Object.entries(step.fields) as [field, value]}
                    <div class="flex justify-between items-center">
                      <span class="font-medium">{field}:</span>
                      <code class="bg-background px-2 py-1 rounded text-sm">{value}</code>
                    </div>
                  {/each}
                </div>
              {/if}

              {#if step.actions}
                <ul class="list-disc list-inside space-y-1 text-muted-foreground">
                  {#each step.actions as action}
                    <li>{action}</li>
                  {/each}
                </ul>
              {/if}

              {#if step.action}
                <div class="bg-blue-50 border border-blue-200 p-3 rounded-lg">
                  <p class="text-blue-800 font-medium">{step.action}</p>
                </div>
              {/if}
            </div>
          </div>
        {/each}

        <!-- Current Configuration Summary -->
        <div class="bg-green-50 border border-green-200 p-4 rounded-lg">
          <h3 class="font-semibold text-green-800 mb-2">🎯 Huidige Configuratie</h3>
          <div class="space-y-2 text-sm">
            <div class="flex justify-between">
              <span class="text-green-700">Domein:</span>
              <code class="bg-white px-2 py-1 rounded">{instructions.current_domain}</code>
            </div>
            <div class="flex justify-between">
              <span class="text-green-700">Protocol:</span>
              <code class="bg-white px-2 py-1 rounded">{instructions.current_protocol}</code>
            </div>
            <div class="flex justify-between">
              <span class="text-green-700">Callback URL:</span>
              <code class="bg-white px-2 py-1 rounded text-xs">{instructions.callback_url}</code>
            </div>
            <div class="flex justify-between">
              <span class="text-green-700">Homepage URL:</span>
              <code class="bg-white px-2 py-1 rounded text-xs">{instructions.homepage_url}</code>
            </div>
          </div>
        </div>

        {#if !instructions.is_production}
          <div class="bg-yellow-50 border border-yellow-200 p-4 rounded-lg">
            <h3 class="font-semibold text-yellow-800 mb-2">🚧 Development Mode</h3>
            <p class="text-yellow-700 text-sm">
              Je bent in development mode. Voor productie moet je de URLs aanpassen naar je productie domein.
            </p>
          </div>
        {/if}
      </div>

      <div class="flex justify-end space-x-3 mt-8 pt-6 border-t border-border">
        <Button variant="outline" on:click={() => showInstructionsModal = false}>
          Sluiten
        </Button>
        <Button on:click={testGitHubOAuth} disabled={testingOAuth}>
          {#if testingOAuth}
            <div class="animate-spin rounded-full h-4 w-4 border-b-2 border-primary-foreground mr-2"></div>
            Testen...
          {:else}
            <Icon name="shield" size={16} className="mr-2" />
            Test Configuratie
          {/if}
        </Button>
      </div>
    </div>
  </Modal>
{/if}