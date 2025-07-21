<script lang="ts">
  import { page } from '$app/stores';
  import { onMount } from 'svelte';
  import { auth } from '$lib/stores/auth';
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

  let users: User[] = [
    {
      id: 1,
      email: 'jan@voorbeeld.nl',
      name: 'Jan de Vries',
      created_at: '2025-01-15T10:30:00Z',
      last_login: '2025-01-19T14:20:00Z',
      status: 'active',
      email_verified: true
    },
    {
      id: 2,
      email: 'sarah@test.nl',
      name: 'Sarah Johnson',
      created_at: '2025-01-16T11:45:00Z',
      last_login: '2025-01-19T09:10:00Z',
      status: 'active',
      email_verified: true
    },
    {
      id: 3,
      email: 'mike@demo.com',
      name: 'Mike Peters',
      created_at: '2025-01-17T09:15:00Z',
      status: 'pending',
      email_verified: false
    }
  ];

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

  $: projectId = $page.params.id;

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

  function toggleUserStatus(userId: number) {
    users = users.map(user => 
      user.id === userId ? { 
        ...user, 
        status: user.status === 'active' ? 'suspended' : 'active' 
      } : user
    );
  }

  function deleteUser(userId: number) {
    if (confirm('Weet je zeker dat je deze gebruiker wilt verwijderen?')) {
      users = users.filter(user => user.id !== userId);
    }
  }

  async function createUser() {
    if (!newUser.email || !newUser.name) return;

    loading = true;
    try {
      const user: User = {
        id: Date.now(),
        email: newUser.email,
        name: newUser.name,
        created_at: new Date().toISOString(),
        status: 'pending',
        email_verified: false
      };

      users = [user, ...users];
      showCreateUser = false;
      newUser = { email: '', name: '', password: '', send_invitation: true };
    } catch (err) {
      console.error('Create user error:', err);
    } finally {
      loading = false;
    }
  }

  function toggleProvider(providerId: string) {
    authProviders = authProviders.map(provider => 
      provider.id === providerId ? { ...provider, enabled: !provider.enabled } : provider
    );
  }

  function saveAuthSettings() {
    alert('Authenticatie instellingen opgeslagen!');
  }
</script>

<svelte:head>
  <title>Authenticatie - CloudBox</title>
</svelte:head>

<div class="space-y-6">
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
  <div class="border-b border-border">
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
        <Card class="p-6">
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
        <Card class="p-6">
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
        <Card class="p-6">
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
        <Card class="p-6">
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
      <Card>
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
          <Card class="p-6">
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

      <Card class="p-6">
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
    <Card class="max-w-md w-full p-6 border-2 shadow-2xl">
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