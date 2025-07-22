<script lang="ts">
  import { onMount } from 'svelte';
  import { auth } from '$lib/stores/auth';
  import { toast } from '$lib/stores/toast';
  import Card from '$lib/components/ui/card.svelte';
  import Button from '$lib/components/ui/button.svelte';
  import Input from '$lib/components/ui/input.svelte';
  import Label from '$lib/components/ui/label.svelte';
  import Badge from '$lib/components/ui/badge.svelte';
  import Icon from '$lib/components/ui/icon.svelte';

  let systemInfo = {
    version: '1.0.0',
    uptime: '0 dagen',
    environment: 'development',
    nodeVersion: '',
    goVersion: '',
    databaseVersion: '',
    redisVersion: '',
  };

  let systemSettings = {
    maintenanceMode: false,
    registrationEnabled: true,
    emailNotifications: true,
    backupEnabled: true,
    logLevel: 'info'
  };

  let loading = true;
  let saving = false;

  const logLevels = [
    { value: 'debug', label: 'Debug' },
    { value: 'info', label: 'Info' },
    { value: 'warn', label: 'Warning' },
    { value: 'error', label: 'Error' }
  ];

  onMount(() => {
    loadSystemInfo();
    loadSystemSettings();
  });

  async function loadSystemInfo() {
    try {
      const response = await fetch('http://localhost:8080/api/v1/admin/system/info', {
        headers: {
          'Authorization': `Bearer ${$auth.token}`,
          'Content-Type': 'application/json',
        },
      });

      if (response.ok) {
        systemInfo = await response.json();
      } else {
        console.error('Failed to load system info:', response.status);
      }
    } catch (error) {
      console.error('Error loading system info:', error);
    }
    loading = false;
  }

  async function loadSystemSettings() {
    try {
      const response = await fetch('http://localhost:8080/api/v1/admin/system/settings', {
        headers: {
          'Authorization': `Bearer ${$auth.token}`,
          'Content-Type': 'application/json',
        },
      });

      if (response.ok) {
        systemSettings = await response.json();
      } else {
        console.error('Failed to load system settings:', response.status);
      }
    } catch (error) {
      console.error('Error loading system settings:', error);
    }
  }

  async function saveSettings() {
    saving = true;
    try {
      const response = await fetch('http://localhost:8080/api/v1/admin/system/settings', {
        method: 'PUT',
        headers: {
          'Authorization': `Bearer ${$auth.token}`,
          'Content-Type': 'application/json',
        },
        body: JSON.stringify(systemSettings),
      });

      if (response.ok) {
        toast.success('Systeem instellingen opgeslagen');
      } else {
        const data = await response.json();
        toast.error(data.error || 'Fout bij opslaan instellingen');
      }
    } catch (error) {
      console.error('Save settings error:', error);
      toast.error('Netwerkfout bij opslaan instellingen');
    } finally {
      saving = false;
    }
  }

  async function restartService() {
    const confirmed = confirm('Weet je zeker dat je de service wilt herstarten?');
    if (!confirmed) return;

    try {
      toast.info('Service wordt herstart...');
      const response = await fetch('http://localhost:8080/api/v1/admin/system/restart', {
        method: 'POST',
        headers: {
          'Authorization': `Bearer ${$auth.token}`,
          'Content-Type': 'application/json',
        },
      });

      if (response.ok) {
        toast.success('Service herstart voltooid');
      } else {
        const data = await response.json();
        toast.error(data.error || 'Fout bij herstarten service');
      }
    } catch (error) {
      console.error('Restart service error:', error);
      toast.error('Netwerkfout bij herstarten service');
    }
  }

  async function clearCache() {
    try {
      toast.info('Cache wordt geleegd...');
      const response = await fetch('http://localhost:8080/api/v1/admin/system/clear-cache', {
        method: 'POST',
        headers: {
          'Authorization': `Bearer ${$auth.token}`,
          'Content-Type': 'application/json',
        },
      });

      if (response.ok) {
        toast.success('Cache geleegd');
      } else {
        const data = await response.json();
        toast.error(data.error || 'Fout bij legen cache');
      }
    } catch (error) {
      console.error('Clear cache error:', error);
      toast.error('Netwerkfout bij legen cache');
    }
  }

  async function runBackup() {
    try {
      toast.info('Backup wordt gestart...');
      const response = await fetch('http://localhost:8080/api/v1/admin/system/backup', {
        method: 'POST',
        headers: {
          'Authorization': `Bearer ${$auth.token}`,
          'Content-Type': 'application/json',
        },
      });

      if (response.ok) {
        const result = await response.json();
        toast.success(`Backup voltooid: ${result.filename || 'backup.sql'}`);
      } else {
        const data = await response.json();
        toast.error(data.error || 'Fout bij maken backup');
      }
    } catch (error) {
      console.error('Backup error:', error);
      toast.error('Netwerkfout bij maken backup');
    }
  }
</script>

<svelte:head>
  <title>Systeem Beheer - CloudBox Admin</title>
</svelte:head>

<div class="space-y-8">
  <!-- Header -->
  <div class="flex items-center justify-between">
    <div class="flex items-center space-x-4">
      <div class="w-12 h-12 bg-orange-500 rounded-xl flex items-center justify-center">
        <Icon name="settings" size={24} color="white" />
      </div>
      <div>
        <h1 class="text-3xl font-bold text-foreground">Systeem Beheer</h1>
        <p class="text-muted-foreground">
          Beheer systeem instellingen en onderhoud
        </p>
      </div>
    </div>
    <div class="flex items-center space-x-3">
      <Button
        variant="outline"
        href="/admin"
        size="lg"
        class="flex items-center space-x-2"
      >
        <Icon name="backup" size={16} />
        <span>Terug naar Dashboard</span>
      </Button>
    </div>
  </div>

  {#if loading}
    <Card class="p-12">
      <div class="text-center">
        <div class="animate-spin rounded-full h-8 w-8 border-b-2 border-primary mx-auto"></div>
        <p class="mt-4 text-muted-foreground">Systeem informatie laden...</p>
      </div>
    </Card>
  {:else}
    <!-- System Information -->
    <div class="grid grid-cols-1 lg:grid-cols-2 gap-8">
      <!-- System Info -->
      <Card class="p-6">
        <div class="flex items-center space-x-3 mb-6">
          <Icon name="info" size={20} className="text-blue-600" />
          <h2 class="text-xl font-semibold text-foreground">Systeem Informatie</h2>
        </div>
        
        <div class="space-y-4">
          <div class="flex justify-between">
            <span class="text-muted-foreground">CloudBox Versie</span>
            <Badge variant="outline">{systemInfo.version}</Badge>
          </div>
          <div class="flex justify-between">
            <span class="text-muted-foreground">Uptime</span>
            <span class="font-medium text-foreground">{systemInfo.uptime}</span>
          </div>
          <div class="flex justify-between">
            <span class="text-muted-foreground">Omgeving</span>
            <Badge variant={systemInfo.environment === 'production' ? 'default' : 'secondary'}>
              {systemInfo.environment}
            </Badge>
          </div>
          <div class="flex justify-between">
            <span class="text-muted-foreground">Node.js</span>
            <span class="font-medium text-foreground">{systemInfo.nodeVersion}</span>
          </div>
          <div class="flex justify-between">
            <span class="text-muted-foreground">Go</span>
            <span class="font-medium text-foreground">{systemInfo.goVersion}</span>
          </div>
          <div class="flex justify-between">
            <span class="text-muted-foreground">Database</span>
            <span class="font-medium text-foreground">{systemInfo.databaseVersion}</span>
          </div>
          <div class="flex justify-between">
            <span class="text-muted-foreground">Redis</span>
            <span class="font-medium text-foreground">{systemInfo.redisVersion}</span>
          </div>
        </div>
      </Card>

      <!-- System Actions -->
      <Card class="p-6">
        <div class="flex items-center space-x-3 mb-6">
          <Icon name="zap" size={20} className="text-yellow-600" />
          <h2 class="text-xl font-semibold text-foreground">Systeem Acties</h2>
        </div>
        
        <div class="space-y-4">
          <Button
            on:click={restartService}
            variant="outline"
            class="w-full flex items-center justify-center space-x-2"
          >
            <Icon name="refresh" size={16} />
            <span>Service Herstarten</span>
          </Button>
          
          <Button
            on:click={clearCache}
            variant="outline"
            class="w-full flex items-center justify-center space-x-2"
          >
            <Icon name="trash" size={16} />
            <span>Cache Legen</span>
          </Button>
          
          <Button
            on:click={runBackup}
            variant="outline"
            class="w-full flex items-center justify-center space-x-2"
          >
            <Icon name="database" size={16} />
            <span>Backup Maken</span>
          </Button>
        </div>
      </Card>
    </div>

    <!-- System Settings -->
    <Card class="p-6">
      <div class="flex items-center justify-between mb-6">
        <div class="flex items-center space-x-3">
          <Icon name="cog" size={20} className="text-purple-600" />
          <h2 class="text-xl font-semibold text-foreground">Systeem Instellingen</h2>
        </div>
        <Button
          on:click={saveSettings}
          disabled={saving}
          class="flex items-center space-x-2"
        >
          {#if saving}
            <div class="animate-spin rounded-full h-4 w-4 border-b-2 border-white"></div>
          {:else}
            <Icon name="save" size={16} />
          {/if}
          <span>{saving ? 'Opslaan...' : 'Opslaan'}</span>
        </Button>
      </div>
      
      <div class="grid grid-cols-1 md:grid-cols-2 gap-6">
        <!-- Column 1 -->
        <div class="space-y-6">
          <div class="flex items-center justify-between">
            <div>
              <Label class="text-sm font-medium">Onderhouds Modus</Label>
              <p class="text-xs text-muted-foreground">Schakel de applicatie tijdelijk uit</p>
            </div>
            <label class="relative inline-flex items-center cursor-pointer">
              <input
                type="checkbox"
                bind:checked={systemSettings.maintenanceMode}
                class="sr-only peer"
              />
              <div class="w-11 h-6 bg-gray-200 peer-focus:outline-none peer-focus:ring-4 peer-focus:ring-blue-300 dark:peer-focus:ring-blue-800 rounded-full peer dark:bg-gray-700 peer-checked:after:translate-x-full peer-checked:after:border-white after:content-[''] after:absolute after:top-[2px] after:left-[2px] after:bg-white after:border-gray-300 after:border after:rounded-full after:h-5 after:w-5 after:transition-all dark:border-gray-600 peer-checked:bg-blue-600"></div>
            </label>
          </div>

          <div class="flex items-center justify-between">
            <div>
              <Label class="text-sm font-medium">Registratie Ingeschakeld</Label>
              <p class="text-xs text-muted-foreground">Sta nieuwe gebruikers toe om te registreren</p>
            </div>
            <label class="relative inline-flex items-center cursor-pointer">
              <input
                type="checkbox"
                bind:checked={systemSettings.registrationEnabled}
                class="sr-only peer"
              />
              <div class="w-11 h-6 bg-gray-200 peer-focus:outline-none peer-focus:ring-4 peer-focus:ring-blue-300 dark:peer-focus:ring-blue-800 rounded-full peer dark:bg-gray-700 peer-checked:after:translate-x-full peer-checked:after:border-white after:content-[''] after:absolute after:top-[2px] after:left-[2px] after:bg-white after:border-gray-300 after:border after:rounded-full after:h-5 after:w-5 after:transition-all dark:border-gray-600 peer-checked:bg-blue-600"></div>
            </label>
          </div>
        </div>

        <!-- Column 2 -->
        <div class="space-y-6">
          <div class="flex items-center justify-between">
            <div>
              <Label class="text-sm font-medium">Email Notificaties</Label>
              <p class="text-xs text-muted-foreground">Verstuur systeem notificaties via email</p>
            </div>
            <label class="relative inline-flex items-center cursor-pointer">
              <input
                type="checkbox"
                bind:checked={systemSettings.emailNotifications}
                class="sr-only peer"
              />
              <div class="w-11 h-6 bg-gray-200 peer-focus:outline-none peer-focus:ring-4 peer-focus:ring-blue-300 dark:peer-focus:ring-blue-800 rounded-full peer dark:bg-gray-700 peer-checked:after:translate-x-full peer-checked:after:border-white after:content-[''] after:absolute after:top-[2px] after:left-[2px] after:bg-white after:border-gray-300 after:border after:rounded-full after:h-5 after:w-5 after:transition-all dark:border-gray-600 peer-checked:bg-blue-600"></div>
            </label>
          </div>

          <div class="flex items-center justify-between">
            <div>
              <Label class="text-sm font-medium">Automatische Backup</Label>
              <p class="text-xs text-muted-foreground">Maak dagelijks automatisch backups</p>
            </div>
            <label class="relative inline-flex items-center cursor-pointer">
              <input
                type="checkbox"
                bind:checked={systemSettings.backupEnabled}
                class="sr-only peer"
              />
              <div class="w-11 h-6 bg-gray-200 peer-focus:outline-none peer-focus:ring-4 peer-focus:ring-blue-300 dark:peer-focus:ring-blue-800 rounded-full peer dark:bg-gray-700 peer-checked:after:translate-x-full peer-checked:after:border-white after:content-[''] after:absolute after:top-[2px] after:left-[2px] after:bg-white after:border-gray-300 after:border after:rounded-full after:h-5 after:w-5 after:transition-all dark:border-gray-600 peer-checked:bg-blue-600"></div>
            </label>
          </div>
        </div>

        <!-- Log Level Setting -->
        <div class="md:col-span-2">
          <Label for="logLevel" class="text-sm font-medium">Log Level</Label>
          <p class="text-xs text-muted-foreground mb-2">Stel het detail niveau van logging in</p>
          <select
            id="logLevel"
            bind:value={systemSettings.logLevel}
            class="w-full px-3 py-2 border border-input rounded-md bg-background text-foreground focus:outline-none focus:ring-2 focus:ring-ring focus:border-transparent"
          >
            {#each logLevels as level}
              <option value={level.value}>{level.label}</option>
            {/each}
          </select>
        </div>
      </div>
    </Card>
  {/if}
</div>