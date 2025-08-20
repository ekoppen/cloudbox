<script lang="ts">
  import { page } from '$app/stores';
  import { onMount } from 'svelte';
  import { auth } from '$lib/stores/auth';
  import { toast } from '$lib/stores/toast';
  import { API_BASE_URL, API_ENDPOINTS } from '$lib/config';
  import Card from '$lib/components/ui/card.svelte';
  import Button from '$lib/components/ui/button.svelte';
  import Input from '$lib/components/ui/input.svelte';
  import Label from '$lib/components/ui/label.svelte';
  import Badge from '$lib/components/ui/badge.svelte';
  import Icon from '$lib/components/ui/icon.svelte';

  interface User {
    id: number;
    email: string;
    name: string;
    created_at: string;
    last_login?: string;
    status: 'active' | 'suspended' | 'pending';
    email_verified: boolean;
  }

  interface AuthProvider {
    id: string;
    name: string;
    enabled: boolean;
    icon: string;
  }

  let users: User[] = [];

  let authProviders: AuthProvider[] = [
    { id: 'email', name: 'Email/Password', enabled: true, icon: '✉️' },
    { id: 'google', name: 'Google OAuth', enabled: false, icon: '🌐' },
    { id: 'github', name: 'GitHub OAuth', enabled: false, icon: '⚫' },
    { id: 'apple', name: 'Apple ID', enabled: false, icon: '🍎' }
  ];

  let authSettings = {
    email_verification: true,
    password_min_length: 8,
    session_duration: 24,
    max_login_attempts: 5,
    lockout_duration: 30
  };

  let activeTab = 'users';
  let showCreateUser = false;
  let newUser = { email: '', name: '', password: '', send_invitation: true };
  let loading = false;
  let backendAvailable = true;

  $: projectId = $page.params.id;

  onMount(() => {
    loadUsers();
    loadAuthSettings();
  });

  async function loadUsers() {
    try {
      const response = await fetch(`${API_BASE_URL}/api/v1/projects/${projectId}/auth/users`, {
        headers: {
          'Authorization': `Bearer ${$auth.token}`,
          'Content-Type': 'application/json',
        },
      });

      if (response.ok) {
        const data = await response.json();
        users = data.users || [];
      } else {
        console.error('Failed to load users:', response.status);
        users = []; // Ensure users is always an array
        // Don't show error toast if backend is not available - this is expected during development
        if (response.status === 404) {
          backendAvailable = false;
        } else if (response.status !== 0) {
          toast.error('Fout bij het laden van gebruikers');
        }
      }
    } catch (error) {
      console.error('Error loading users:', error);
      users = []; // Ensure users is always an array
      // Check if backend is unavailable
      if (error.message.includes('Failed to fetch')) {
        backendAvailable = false;
      } else {
        toast.error('Netwerkfout bij het laden van gebruikers');
      }
    }
  }

  async function loadAuthSettings() {
    try {
      const response = await fetch(`${API_BASE_URL}/api/v1/projects/${projectId}/auth/settings`, {
        headers: {
          'Authorization': `Bearer ${$auth.token}`,
          'Content-Type': 'application/json',
        },
      });

      if (response.ok) {
        const settings = await response.json();
        authSettings = { ...authSettings, ...settings };
        authProviders = settings.providers || authProviders;
      } else {
        console.error('Failed to load auth settings:', response.status);
        // Don't show error toast if backend is not available
        if (response.status === 404) {
          backendAvailable = false;
        }
      }
    } catch (error) {
      console.error('Error loading auth settings:', error);
      // Only show error toast if it's not a connection error
      if (!error.message.includes('Failed to fetch')) {
        toast.error('Netwerkfout bij het laden van instellingen');
      }
    }
  }

  function getStatusColor(status: string): string {
    switch (status) {
      case 'active': return 'bg-green-100 dark:bg-green-900 text-green-800 dark:text-green-200';
      case 'suspended': return 'bg-red-100 dark:bg-red-900 text-red-800 dark:text-red-200';
      case 'pending': return 'bg-yellow-100 dark:bg-yellow-900 text-yellow-800 dark:text-yellow-200';
      default: return 'bg-muted text-muted-foreground';
    }
  }

  function formatDate(dateStr: string): string {
    return new Date(dateStr).toLocaleDateString('nl-NL', {
      year: 'numeric',
      month: 'short',
      day: 'numeric',
      hour: '2-digit',
      minute: '2-digit'
    });
  }

  async function toggleUserStatus(userId: number) {
    try {
      const user = users.find(u => u.id === userId);
      if (!user) return;

      const newStatus = user.status === 'active' ? 'suspended' : 'active';
      
      const response = await fetch(`${API_BASE_URL}/api/v1/projects/${projectId}/auth/users/${userId}`, {
        method: 'PATCH',
        headers: {
          'Authorization': `Bearer ${$auth.token}`,
          'Content-Type': 'application/json',
        },
        body: JSON.stringify({ status: newStatus }),
      });

      if (response.ok) {
        users = users.map(user => 
          user.id === userId ? { ...user, status: newStatus } : user
        );
        toast.success(`Gebruiker ${newStatus === 'active' ? 'geactiveerd' : 'gesuspendeerd'}`);
      } else {
        toast.error('Fout bij het wijzigen van gebruikersstatus');
      }
    } catch (error) {
      console.error('Error toggling user status:', error);
      toast.error('Netwerkfout bij het wijzigen van gebruikersstatus');
    }
  }

  async function deleteUser(userId: number) {
    if (!confirm('Weet je zeker dat je deze gebruiker wilt verwijderen?')) return;

    try {
      const response = await fetch(`${API_BASE_URL}/api/v1/projects/${projectId}/auth/users/${userId}`, {
        method: 'DELETE',
        headers: {
          'Authorization': `Bearer ${$auth.token}`,
          'Content-Type': 'application/json',
        },
      });

      if (response.ok) {
        users = users.filter(user => user.id !== userId);
        toast.success('Gebruiker verwijderd');
      } else {
        toast.error('Fout bij het verwijderen van gebruiker');
      }
    } catch (error) {
      console.error('Error deleting user:', error);
      toast.error('Netwerkfout bij het verwijderen van gebruiker');
    }
  }

  async function createUser() {
    if (!newUser.email || !newUser.name) return;

    loading = true;
    try {
      const response = await fetch(`${API_BASE_URL}/api/v1/projects/${projectId}/auth/users`, {
        method: 'POST',
        headers: {
          'Authorization': `Bearer ${$auth.token}`,
          'Content-Type': 'application/json',
        },
        body: JSON.stringify({
          email: newUser.email,
          name: newUser.name,
          password: newUser.password || undefined,
          send_invitation: newUser.send_invitation,
        }),
      });

      if (response.ok) {
        const user = await response.json();
        users = [user, ...users];
        showCreateUser = false;
        newUser = { email: '', name: '', password: '', send_invitation: true };
        toast.success('Gebruiker succesvol aangemaakt');
      } else {
        const error = await response.json();
        toast.error(error.message || 'Fout bij het aanmaken van gebruiker');
      }
    } catch (err) {
      console.error('Create user error:', err);
      toast.error('Netwerkfout bij het aanmaken van gebruiker');
    } finally {
      loading = false;
    }
  }

  async function toggleProvider(providerId: string) {
    try {
      const provider = authProviders.find(p => p.id === providerId);
      if (!provider) return;

      const newEnabled = !provider.enabled;
      
      const response = await fetch(`${API_BASE_URL}/api/v1/projects/${projectId}/auth/providers/${providerId}`, {
        method: 'PATCH',
        headers: {
          'Authorization': `Bearer ${$auth.token}`,
          'Content-Type': 'application/json',
        },
        body: JSON.stringify({ enabled: newEnabled }),
      });

      if (response.ok) {
        authProviders = authProviders.map(provider => 
          provider.id === providerId ? { ...provider, enabled: newEnabled } : provider
        );
        toast.success(`${provider.name} ${newEnabled ? 'ingeschakeld' : 'uitgeschakeld'}`);
      } else {
        toast.error('Fout bij het wijzigen van provider instellingen');
      }
    } catch (error) {
      console.error('Error toggling provider:', error);
      toast.error('Netwerkfout bij het wijzigen van provider instellingen');
    }
  }

  async function saveAuthSettings() {
    loading = true;
    try {
      const response = await fetch(`${API_BASE_URL}/api/v1/projects/${projectId}/auth/settings`, {
        method: 'PUT',
        headers: {
          'Authorization': `Bearer ${$auth.token}`,
          'Content-Type': 'application/json',
        },
        body: JSON.stringify(authSettings),
      });

      if (response.ok) {
        toast.success('Authenticatie instellingen opgeslagen');
      } else {
        toast.error('Fout bij het opslaan van instellingen');
      }
    } catch (error) {
      console.error('Error saving auth settings:', error);
      toast.error('Netwerkfout bij het opslaan van instellingen');
    } finally {
      loading = false;
    }
  }
</script>

<svelte:head>
  <title>Authenticatie - CloudBox</title>
</svelte:head>

<div class="space-y-6">
  <!-- Backend Status Notice -->
  {#if !backendAvailable}
    <Card class="glassmorphism-card bg-yellow-50 dark:bg-yellow-900/20 border-yellow-200 dark:border-yellow-800 p-4">
      <div class="flex items-center space-x-3">
        <Icon name="warning" size={20} className="text-yellow-600 dark:text-yellow-400" />
        <div>
          <h3 class="text-sm font-medium text-yellow-800 dark:text-yellow-200">Backend Server Niet Beschikbaar</h3>
          <p class="text-xs text-yellow-700 dark:text-yellow-300 mt-1">
            De backend server draait niet op localhost:8080. Start de backend server om alle functionaliteit te gebruiken.
          </p>
        </div>
      </div>
    </Card>
  {/if}

  <!-- Header -->
  <div class="flex items-center space-x-4">
    <div class="w-10 h-10 bg-primary/10 rounded-lg flex items-center justify-center">
      <Icon name="auth" size={20} className="text-primary" />
    </div>
    <div>
      <h1 class="text-2xl font-bold text-foreground">Authenticatie</h1>
      <p class="text-sm text-muted-foreground">
        Beheer gebruikers, authenticatie providers en beveiliging
      </p>
    </div>
  </div>

  <!-- Tabs -->
  <div class="glassmorphism-nav border-b border-border">
    <nav class="flex space-x-8">
      <button
        on:click={() => activeTab = 'users'}
        class="flex items-center space-x-2 py-2 px-1 border-b-2 text-sm font-medium transition-colors {
          activeTab === 'users' 
            ? 'border-primary text-primary' 
            : 'border-transparent text-muted-foreground hover:text-foreground hover:border-border'
        }"
      >
        <Icon name="user" size={16} />
        <span>Gebruikers ({users.length})</span>
      </button>
      <button
        on:click={() => activeTab = 'providers'}
        class="flex items-center space-x-2 py-2 px-1 border-b-2 text-sm font-medium transition-colors {
          activeTab === 'providers' 
            ? 'border-primary text-primary' 
            : 'border-transparent text-muted-foreground hover:text-foreground hover:border-border'
        }"
      >
        <Icon name="auth" size={16} />
        <span>Providers</span>
      </button>
      <button
        on:click={() => activeTab = 'settings'}
        class="flex items-center space-x-2 py-2 px-1 border-b-2 text-sm font-medium transition-colors {
          activeTab === 'settings' 
            ? 'border-primary text-primary' 
            : 'border-transparent text-muted-foreground hover:text-foreground hover:border-border'
        }"
      >
        <Icon name="settings" size={16} />
        <span>Instellingen</span>
      </button>
    </nav>
  </div>

  <!-- Users Tab -->
  {#if activeTab === 'users'}
    <div class="space-y-6">
      <div class="flex justify-between items-center">
        <div>
          <h2 class="text-lg font-medium text-foreground">Gebruikers</h2>
          <p class="text-sm text-muted-foreground">Beheer geregistreerde gebruikers</p>
        </div>
        <Button
          on:click={() => showCreateUser = true}
          class="flex items-center space-x-2"
        >
          <Icon name="user" size={16} />
          <span>Nieuwe Gebruiker</span>
        </Button>
      </div>

      <!-- Quick Stats -->
      <div class="grid grid-cols-1 md:grid-cols-4 gap-4">
        <Card class="glassmorphism-card p-6">
          <div class="flex items-center justify-between">
            <div>
              <p class="text-sm font-medium text-muted-foreground">Actieve gebruikers</p>
              <p class="text-2xl font-bold text-foreground">{users.filter(u => u.status === 'active').length}</p>
            </div>
            <div class="w-10 h-10 bg-green-100 dark:bg-green-900 rounded-lg flex items-center justify-center">
              <Icon name="user" size={20} className="text-green-600 dark:text-green-400" />
            </div>
          </div>
        </Card>
        <Card class="glassmorphism-card p-6">
          <div class="flex items-center justify-between">
            <div>
              <p class="text-sm font-medium text-muted-foreground">In afwachting</p>
              <p class="text-2xl font-bold text-foreground">{users.filter(u => u.status === 'pending').length}</p>
            </div>
            <div class="w-10 h-10 bg-yellow-100 dark:bg-yellow-900 rounded-lg flex items-center justify-center">
              <Icon name="backup" size={20} className="text-yellow-600 dark:text-yellow-400" />
            </div>
          </div>
        </Card>
        <Card class="glassmorphism-card p-6">
          <div class="flex items-center justify-between">
            <div>
              <p class="text-sm font-medium text-muted-foreground">Geverifieerd</p>
              <p class="text-2xl font-bold text-foreground">{users.filter(u => u.email_verified).length}</p>
            </div>
            <div class="w-10 h-10 bg-blue-100 dark:bg-blue-900 rounded-lg flex items-center justify-center">
              <Icon name="auth" size={20} className="text-blue-600 dark:text-blue-400" />
            </div>
          </div>
        </Card>
        <Card class="glassmorphism-card p-6">
          <div class="flex items-center justify-between">
            <div>
              <p class="text-sm font-medium text-muted-foreground">Recent actief</p>
              <p class="text-2xl font-bold text-foreground">{users.filter(u => u.last_login).length}</p>
            </div>
            <div class="w-10 h-10 bg-purple-100 dark:bg-purple-900 rounded-lg flex items-center justify-center">
              <Icon name="functions" size={20} className="text-purple-600 dark:text-purple-400" />
            </div>
          </div>
        </Card>
      </div>

      <!-- Users Table -->
      <Card class="glassmorphism-table">
        <div class="overflow-x-auto">
          <table class="min-w-full divide-y divide-border">
            <thead class="bg-muted/30">
              <tr>
                <th class="px-6 py-3 text-left text-xs font-medium text-muted-foreground uppercase tracking-wider">Gebruiker</th>
                <th class="px-6 py-3 text-left text-xs font-medium text-muted-foreground uppercase tracking-wider">Status</th>
                <th class="px-6 py-3 text-left text-xs font-medium text-muted-foreground uppercase tracking-wider">Aangemaakt</th>
                <th class="px-6 py-3 text-left text-xs font-medium text-muted-foreground uppercase tracking-wider">Laatste login</th>
                <th class="px-6 py-3 text-right text-xs font-medium text-muted-foreground uppercase tracking-wider">Acties</th>
              </tr>
            </thead>
            <tbody class="bg-card divide-y divide-border">
              {#each users as user}
                <tr class="hover:bg-muted/30">
                  <td class="px-6 py-4">
                    <div class="flex items-center">
                      <div class="flex-shrink-0 h-10 w-10">
                        <div class="h-10 w-10 rounded-full bg-primary/10 flex items-center justify-center">
                          <span class="text-primary font-medium">{user.name.charAt(0)}</span>
                        </div>
                      </div>
                      <div class="ml-4">
                        <div class="text-sm font-medium text-foreground">{user.name}</div>
                        <div class="text-sm text-muted-foreground flex items-center">
                          {user.email}
                          {#if user.email_verified}
                            <Icon name="auth" size={12} className="ml-1 text-green-500" />
                          {/if}
                        </div>
                      </div>
                    </div>
                  </td>
                  <td class="px-6 py-4">
                    <Badge class={getStatusColor(user.status)}>
                      {user.status}
                    </Badge>
                  </td>
                  <td class="px-6 py-4 text-sm text-muted-foreground">
                    {formatDate(user.created_at)}
                  </td>
                  <td class="px-6 py-4 text-sm text-muted-foreground">
                    {user.last_login ? formatDate(user.last_login) : 'Nooit'}
                  </td>
                  <td class="px-6 py-4 text-right">
                    <div class="flex justify-end space-x-2">
                      <Button
                        variant="ghost"
                        size="sm"
                        on:click={() => toggleUserStatus(user.id)}
                      >
                        {user.status === 'active' ? 'Suspendeer' : 'Activeer'}
                      </Button>
                      <Button
                        variant="ghost"
                        size="sm"
                        class="text-destructive hover:text-destructive"
                        on:click={() => deleteUser(user.id)}
                      >
                        Verwijder
                      </Button>
                    </div>
                  </td>
                </tr>
              {/each}
            </tbody>
          </table>
        </div>
      </Card>
    </div>
  {/if}

  <!-- Providers Tab -->
  {#if activeTab === 'providers'}
    <div class="space-y-6">
      <div>
        <h2 class="text-lg font-medium text-foreground">Authenticatie Providers</h2>
        <p class="text-sm text-muted-foreground">Configureer authenticatie methoden</p>
      </div>

      <div class="grid gap-4">
        {#each authProviders as provider}
          <Card class="glassmorphism-card p-6">
            <div class="flex items-center justify-between">
              <div class="flex items-center space-x-4">
                <div class="w-12 h-12 bg-primary/10 rounded-lg flex items-center justify-center">
                  <Icon name="auth" size={24} className="text-primary" />
                </div>
                <div>
                  <h3 class="text-lg font-medium text-foreground">{provider.name}</h3>
                  <p class="text-sm text-muted-foreground">
                    {provider.enabled ? 'Ingeschakeld' : 'Uitgeschakeld'}
                  </p>
                </div>
              </div>
              <div class="flex items-center space-x-3">
                <label class="flex items-center">
                  <input
                    type="checkbox"
                    checked={provider.enabled}
                    on:change={() => toggleProvider(provider.id)}
                    class="rounded border-border text-primary focus:ring-primary"
                  />
                  <span class="ml-2 text-sm text-foreground">Ingeschakeld</span>
                </label>
                {#if provider.id !== 'email'}
                  <Button variant="outline">
                    Configureren
                  </Button>
                {/if}
              </div>
            </div>
          </Card>
        {/each}
      </div>
    </div>
  {/if}

  <!-- Settings Tab -->
  {#if activeTab === 'settings'}
    <div class="space-y-6">
      <div>
        <h2 class="text-lg font-medium text-foreground">Beveiliging Instellingen</h2>
        <p class="text-sm text-muted-foreground">Configureer wachtwoord en sessie beleid</p>
      </div>

      <Card class="glassmorphism-form p-6">
        <form on:submit|preventDefault={saveAuthSettings} class="space-y-6">
          <div class="grid grid-cols-1 md:grid-cols-2 gap-6">
            <div>
              <label class="flex items-center">
                <input
                  type="checkbox"
                  bind:checked={authSettings.email_verification}
                  class="rounded border-border text-primary focus:ring-primary"
                />
                <span class="ml-2 text-sm font-medium text-foreground">Email verificatie vereist</span>
              </label>
            </div>

            <div>
              <Label>Minimale wachtwoord lengte</Label>
              <Input
                type="number"
                bind:value={authSettings.password_min_length}
                min="6"
                max="32"
                class="mt-1"
              />
            </div>

            <div>
              <Label>Sessie duur (uren)</Label>
              <Input
                type="number"
                bind:value={authSettings.session_duration}
                min="1"
                max="168"
                class="mt-1"
              />
            </div>

            <div>
              <Label>Max inlog pogingen</Label>
              <Input
                type="number"
                bind:value={authSettings.max_login_attempts}
                min="3"
                max="10"
                class="mt-1"
              />
            </div>

            <div>
              <Label>Blokkeer duur (minuten)</Label>
              <Input
                type="number"
                bind:value={authSettings.lockout_duration}
                min="5"
                max="1440"
                class="mt-1"
              />
            </div>
          </div>

          <div class="flex justify-end">
            <Button type="submit">
              Instellingen Opslaan
            </Button>
          </div>
        </form>
      </Card>
    </div>
  {/if}
</div>

<!-- Create User Modal -->
{#if showCreateUser}
  <div class="fixed inset-0 bg-black/80 backdrop-blur-sm flex items-center justify-center p-4 z-50">
    <Card class="glassmorphism-modal max-w-md w-full p-6 border-2 shadow-2xl">
      <div class="flex items-center space-x-3 mb-4">
        <div class="w-10 h-10 bg-primary/10 rounded-lg flex items-center justify-center">
          <Icon name="user" size={20} className="text-primary" />
        </div>
        <h2 class="text-xl font-bold text-foreground">Nieuwe Gebruiker</h2>
      </div>
      
      <form on:submit|preventDefault={createUser} class="space-y-4">
        <div>
          <Label for="user-email">Email</Label>
          <Input
            id="user-email"
            type="email"
            bind:value={newUser.email}
            required
            class="mt-1"
            placeholder="gebruiker@voorbeeld.nl"
          />
        </div>

        <div>
          <Label for="user-name">Naam</Label>
          <Input
            id="user-name"
            type="text"
            bind:value={newUser.name}
            required
            class="mt-1"
            placeholder="Volledige naam"
          />
        </div>

        <div>
          <Label for="user-password">Wachtwoord</Label>
          <Input
            id="user-password"
            type="password"
            bind:value={newUser.password}
            class="mt-1"
            placeholder="Laat leeg voor automatisch wachtwoord"
          />
        </div>

        <div>
          <label class="flex items-center">
            <input
              type="checkbox"
              bind:checked={newUser.send_invitation}
              class="rounded border-border text-primary focus:ring-primary"
            />
            <span class="ml-2 text-sm text-foreground">Uitnodiging email versturen</span>
          </label>
        </div>
        
        <div class="flex space-x-3 pt-4">
          <Button
            type="button"
            variant="outline"
            on:click={() => { showCreateUser = false; newUser = { email: '', name: '', password: '', send_invitation: true }; }}
            class="flex-1"
          >
            Annuleren
          </Button>
          <Button
            type="submit"
            disabled={loading || !newUser.email || !newUser.name}
            class="flex-1"
          >
            {loading ? 'Aanmaken...' : 'Gebruiker Aanmaken'}
          </Button>
        </div>
      </form>
    </Card>
  </div>
{/if}