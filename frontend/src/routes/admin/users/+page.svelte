<script lang="ts">
  import { onMount } from 'svelte';
  import { API_BASE_URL, API_ENDPOINTS, createApiRequest } from '$lib/config';
  import { auth } from '$lib/stores/auth';
  import { toast } from '$lib/stores/toast';
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
    role: string;
    is_active: boolean;
    created_at: string;
    updated_at: string;
    last_login_at?: string;
  }

  let users: User[] = [];
  let loading = true;
  let error = '';
  let showEditModal = false;
  let showDeleteModal = false;
  let showCreateModal = false;
  let editingUser: User | null = null;
  let passwordValidation = {
    minLength: false,
    hasUppercase: false,
    hasLowercase: false,
    hasNumber: false,
    hasSpecialChar: false
  };
  let deletingUser: User | null = null;
  let newUser = {
    name: '',
    email: '',
    password: '',
    role: 'user',
    is_active: true
  };
  let searchTerm = '';
  let selectedRole = 'all';
  let selectedStatus = 'all';

  const roles = [
    { value: 'all', label: 'Alle rollen' },
    { value: 'admin', label: 'Admin' },
    { value: 'user', label: 'User' }
  ];

  const statusOptions = [
    { value: 'all', label: 'Alle statussen' },
    { value: 'active', label: 'Actief' },
    { value: 'inactive', label: 'Inactief' }
  ];

  onMount(() => {
    loadUsers();
  });

  async function loadUsers() {
    loading = true;
    error = '';

    try {
      const response = await createApiRequest(API_ENDPOINTS.admin.users.list, {
        headers: {
          'Authorization': `Bearer ${$auth.token}`,
          'Content-Type': 'application/json',
        },
      });

      if (response.ok) {
        users = await response.json();
      } else {
        const data = await response.json();
        error = data.error || 'Fout bij laden van gebruikers';
      }
    } catch (err) {
      error = 'Netwerkfout bij laden van gebruikers';
      console.error('Load users error:', err);
    } finally {
      loading = false;
    }
  }

  async function updateUser(user: User) {
    try {
      const response = await fetch(API_ENDPOINTS.admin.users.update(user.id.toString()), {
        method: 'PUT',
        headers: {
          'Authorization': `Bearer ${$auth.token}`,
          'Content-Type': 'application/json',
        },
        body: JSON.stringify({
          email: user.email,
          name: user.name,
          role: user.role,
          is_active: user.is_active
        }),
      });

      if (response.ok) {
        const updatedUser = await response.json();
        users = users.map(u => u.id === user.id ? updatedUser : u);
        showEditModal = false;
        editingUser = null;
        toast.success('Gebruiker bijgewerkt');
      } else {
        const data = await response.json();
        error = data.error || 'Fout bij bijwerken van gebruiker';
      }
    } catch (err) {
      error = 'Netwerkfout bij bijwerken van gebruiker';
      console.error('Update user error:', err);
    }
  }

  async function deleteUser(user: User) {
    try {
      const response = await fetch(API_ENDPOINTS.admin.users.delete(user.id.toString()), {
        method: 'DELETE',
        headers: {
          'Authorization': `Bearer ${$auth.token}`,
          'Content-Type': 'application/json',
        },
      });

      if (response.ok) {
        users = users.filter(u => u.id !== user.id);
        showDeleteModal = false;
        deletingUser = null;
        toast.success('Gebruiker verwijderd');
      } else {
        const data = await response.json();
        error = data.error || 'Fout bij verwijderen van gebruiker';
      }
    } catch (err) {
      error = 'Netwerkfout bij verwijderen van gebruiker';
      console.error('Delete user error:', err);
    }
  }

  async function toggleUserStatus(user: User) {
    const updatedUser = { ...user, is_active: !user.is_active };
    await updateUser(updatedUser);
  }

  function validatePassword(password: string) {
    passwordValidation = {
      minLength: password.length >= 8,
      hasUppercase: /[A-Z]/.test(password),
      hasLowercase: /[a-z]/.test(password),
      hasNumber: /\d/.test(password),
      hasSpecialChar: /[!@#$%^&*()_+\-=\[\]{};':"\\|,.<>\/?]/.test(password)
    };
    
    return Object.values(passwordValidation).every(Boolean);
  }

  function isPasswordValid(): boolean {
    return validatePassword(newUser.password);
  }

  async function createUser() {
    // Validate password before submitting
    if (!isPasswordValid()) {
      toast.error('Wachtwoord voldoet niet aan alle vereisten');
      return;
    }
    try {
      const response = await fetch(API_ENDPOINTS.admin.users.create, {
        method: 'POST',
        headers: {
          'Authorization': `Bearer ${$auth.token}`,
          'Content-Type': 'application/json',
        },
        body: JSON.stringify(newUser),
      });

      if (response.ok) {
        const createdUser = await response.json();
        users = [...users, createdUser];
        showCreateModal = false;
        newUser = {
          name: '',
          email: '',
          password: '',
          role: 'user',
          is_active: true
        };
        toast.success('Gebruiker aangemaakt');
      } else {
        const data = await response.json();
        error = data.error || 'Fout bij aanmaken van gebruiker';
      }
    } catch (err) {
      error = 'Netwerkfout bij aanmaken van gebruiker';
      console.error('Create user error:', err);
    }
  }

  function formatDate(dateStr: string) {
    return new Date(dateStr).toLocaleDateString('nl-NL', {
      year: 'numeric',
      month: 'short',
      day: 'numeric',
      hour: '2-digit',
      minute: '2-digit'
    });
  }

  function openEditModal(user: User) {
    editingUser = { ...user };
    showEditModal = true;
  }

  function openDeleteModal(user: User) {
    deletingUser = user;
    showDeleteModal = true;
  }

  // Filter users based on search term, role, and status
  $: filteredUsers = users.filter(user => {
    const matchesSearch = user.email.toLowerCase().includes(searchTerm.toLowerCase()) ||
                         user.name.toLowerCase().includes(searchTerm.toLowerCase());
    const matchesRole = selectedRole === 'all' || user.role === selectedRole;
    const matchesStatus = selectedStatus === 'all' || 
                         (selectedStatus === 'active' && user.is_active) ||
                         (selectedStatus === 'inactive' && !user.is_active);
    
    return matchesSearch && matchesRole && matchesStatus;
  });

  $: userStats = {
    total: users.length,
    active: users.filter(u => u.is_active).length,
    inactive: users.filter(u => !u.is_active).length,
    admins: users.filter(u => u.role === 'admin').length
  };
</script>

<svelte:head>
  <title>Gebruikers Beheer - CloudBox Admin</title>
</svelte:head>

<div class="space-y-8">
  <!-- Header -->
  <div class="flex items-center justify-between">
    <div class="flex items-center space-x-4">
      <div class="w-12 h-12 bg-primary rounded-xl flex items-center justify-center">
        <Icon name="users" size={24} color="white" />
      </div>
      <div>
        <h1 class="text-3xl font-bold text-foreground">Gebruikers Beheer</h1>
        <p class="text-muted-foreground">
          Beheer alle CloudBox gebruikers en hun rechten
        </p>
      </div>
    </div>
    <div class="flex items-center space-x-3">
      <Button
        variant="secondary"
        href="/admin"
        size="lg"
        class="flex items-center space-x-2"
      >
        <Icon name="backup" size={16} />
        <span>Terug naar Dashboard</span>
      </Button>
      <Button
        variant="primary"
        on:click={() => showCreateModal = true}
        size="lg"
        class="flex items-center space-x-2"
      >
        <Icon name="users" size={16} />
        <span>Nieuwe Gebruiker</span>
      </Button>
      <Button
        on:click={loadUsers}
        variant="ghost"
        size="icon"
        class="hover:rotate-180 transition-transform duration-300"
        title="Vernieuw gebruikerslijst"
      >
        <Icon name="refresh-cw" size={16} />
      </Button>
    </div>
  </div>

  <!-- User Stats -->
  <div class="grid grid-cols-1 md:grid-cols-4 gap-6">
    <Card class="p-6">
      <div class="flex items-center justify-between">
        <div>
          <p class="text-sm font-medium text-muted-foreground">Totaal Gebruikers</p>
          <p class="text-2xl font-bold text-foreground">{userStats.total}</p>
        </div>
        <div class="w-10 h-10 bg-blue-100 dark:bg-gray-800 rounded-lg flex items-center justify-center">
          <Icon name="users" size={20} className="text-blue-600 dark:text-blue-400" />
        </div>
      </div>
    </Card>
    
    <Card class="p-6">
      <div class="flex items-center justify-between">
        <div>
          <p class="text-sm font-medium text-muted-foreground">Actieve Gebruikers</p>
          <p class="text-2xl font-bold text-green-600">{userStats.active}</p>
        </div>
        <div class="w-10 h-10 bg-green-100 dark:bg-green-900 rounded-lg flex items-center justify-center">
          <Icon name="check" size={20} className="text-green-600 dark:text-green-400" />
        </div>
      </div>
    </Card>
    
    <Card class="p-6">
      <div class="flex items-center justify-between">
        <div>
          <p class="text-sm font-medium text-muted-foreground">Inactieve Gebruikers</p>
          <p class="text-2xl font-bold text-gray-600">{userStats.inactive}</p>
        </div>
        <div class="w-10 h-10 bg-gray-100 dark:bg-gray-900 rounded-lg flex items-center justify-center">
          <Icon name="x" size={20} className="text-gray-600 dark:text-gray-400" />
        </div>
      </div>
    </Card>
    
    <Card class="p-6">
      <div class="flex items-center justify-between">
        <div>
          <p class="text-sm font-medium text-muted-foreground">Administrators</p>
          <p class="text-2xl font-bold text-purple-600">{userStats.admins}</p>
        </div>
        <div class="w-10 h-10 bg-purple-100 dark:bg-purple-900 rounded-lg flex items-center justify-center">
          <Icon name="shield" size={20} className="text-purple-600 dark:text-purple-400" />
        </div>
      </div>
    </Card>
  </div>

  <!-- Filters -->
  <Card class="p-6">
    <div class="grid grid-cols-1 md:grid-cols-4 gap-4">
      <div>
        <Label for="search">Zoeken</Label>
        <div class="relative">
          <Icon name="search" size={16} className="absolute left-3 top-1/2 transform -translate-y-1/2 text-muted-foreground" />
          <Input
            id="search"
            type="text"
            placeholder="Email of naam..."
            bind:value={searchTerm}
            class="pl-10"
          />
        </div>
      </div>
      <div>
        <Label for="role">Rol</Label>
        <select 
          id="role"
          bind:value={selectedRole}
          class="w-full px-3 py-2 border border-input rounded-md bg-background text-foreground focus:outline-none focus:ring-2 focus:ring-ring focus:border-transparent"
        >
          {#each roles as role}
            <option value={role.value}>{role.label}</option>
          {/each}
        </select>
      </div>
      <div>
        <Label for="status">Status</Label>
        <select 
          id="status"
          bind:value={selectedStatus}
          class="w-full px-3 py-2 border border-input rounded-md bg-background text-foreground focus:outline-none focus:ring-2 focus:ring-ring focus:border-transparent"
        >
          {#each statusOptions as status}
            <option value={status.value}>{status.label}</option>
          {/each}
        </select>
      </div>
      <div class="flex items-end">
        <Button
          variant="outline"
          on:click={() => {
            searchTerm = '';
            selectedRole = 'all';
            selectedStatus = 'all';
          }}
          class="flex items-center space-x-2"
        >
          <Icon name="x" size={16} />
          <span>Wissen</span>
        </Button>
      </div>
    </div>
  </Card>

  <!-- Error message -->
  {#if error}
    <Card class="bg-destructive/10 border-destructive/20 p-4">
      <div class="flex justify-between items-center">
        <p class="text-destructive text-sm">{error}</p>
        <Button
          variant="ghost"
          size="sm"
          on:click={() => error = ''}
          class="text-destructive hover:text-destructive/80"
        >
          ×
        </Button>
      </div>
    </Card>
  {/if}

  <!-- Users Table -->
  {#if loading}
    <Card class="p-12">
      <div class="text-center">
        <div class="animate-spin rounded-full h-8 w-8 border-b-2 border-primary mx-auto"></div>
        <p class="mt-4 text-muted-foreground">Gebruikers laden...</p>
      </div>
    </Card>
  {:else if filteredUsers.length === 0}
    <Card class="p-12">
      <div class="text-center">
        <div class="w-16 h-16 bg-muted rounded-full flex items-center justify-center mx-auto mb-4">
          <Icon name="users" size={32} className="text-muted-foreground" />
        </div>
        <h3 class="text-lg font-medium text-foreground mb-2">Geen gebruikers gevonden</h3>
        <p class="text-muted-foreground mb-6 max-w-sm mx-auto">
          {searchTerm || selectedRole !== 'all' || selectedStatus !== 'all' 
            ? 'Geen gebruikers voldoen aan de filteropties.' 
            : 'Er zijn nog geen gebruikers geregistreerd.'}
        </p>
      </div>
    </Card>
  {:else}
    <Card class="overflow-hidden">
      <div class="overflow-x-auto">
        <table class="w-full">
          <thead class="bg-muted">
            <tr>
              <th class="text-left p-4 font-medium text-muted-foreground">Gebruiker</th>
              <th class="text-left p-4 font-medium text-muted-foreground">Email</th>
              <th class="text-left p-4 font-medium text-muted-foreground">Rol</th>
              <th class="text-left p-4 font-medium text-muted-foreground">Status</th>
              <th class="text-left p-4 font-medium text-muted-foreground">Geregistreerd</th>
              <th class="text-left p-4 font-medium text-muted-foreground">Laatste Login</th>
              <th class="text-left p-4 font-medium text-muted-foreground">Acties</th>
            </tr>
          </thead>
          <tbody class="divide-y divide-border">
            {#each filteredUsers as user}
              <tr class="hover:bg-muted/50">
                <td class="p-4">
                  <div class="flex items-center space-x-3">
                    <div class="w-10 h-10 bg-primary/10 rounded-full flex items-center justify-center">
                      <Icon name="user" size={16} className="text-primary" />
                    </div>
                    <div>
                      <div class="font-medium text-foreground">{user.name}</div>
                      <div class="text-sm text-muted-foreground">ID: {user.id}</div>
                    </div>
                  </div>
                </td>
                <td class="p-4 text-foreground">{user.email}</td>
                <td class="p-4">
                  <Badge variant={user.role === 'superadmin' ? 'destructive' : user.role === 'admin' ? 'default' : 'secondary'} class="flex items-center space-x-1">
                    <Icon name={user.role === 'superadmin' ? 'crown' : user.role === 'admin' ? 'shield' : 'user'} size={12} />
                    <span>{user.role === 'superadmin' ? 'SuperAdmin' : user.role === 'admin' ? 'Admin' : 'User'}</span>
                  </Badge>
                </td>
                <td class="p-4">
                  <Badge variant={user.is_active ? 'default' : 'secondary'} class="flex items-center space-x-1">
                    <div class="w-2 h-2 rounded-full {user.is_active ? 'bg-green-500' : 'bg-gray-400'}"></div>
                    <span>{user.is_active ? 'Actief' : 'Inactief'}</span>
                  </Badge>
                </td>
                <td class="p-4 text-muted-foreground text-sm">{formatDate(user.created_at)}</td>
                <td class="p-4 text-muted-foreground text-sm">
                  {user.last_login_at ? formatDate(user.last_login_at) : 'Nooit'}
                </td>
                <td class="p-4">
                  <div class="flex items-center space-x-2">
                    <Button
                      variant="outline"
                      size="sm"
                      on:click={() => toggleUserStatus(user)}
                      class="flex items-center space-x-1"
                    >
                      <Icon name={user.is_active ? 'x' : 'check'} size={14} />
                      <span>{user.is_active ? 'Deactiveren' : 'Activeren'}</span>
                    </Button>
                    <Button
                      variant="outline"
                      size="sm"
                      on:click={() => openEditModal(user)}
                      class="flex items-center space-x-1"
                    >
                      <Icon name="edit" size={14} />
                    </Button>
                    <Button
                      variant="destructive"
                      size="sm"
                      on:click={() => openDeleteModal(user)}
                      class="flex items-center space-x-1"
                    >
                      <Icon name="trash" size={14} />
                    </Button>
                  </div>
                </td>
              </tr>
            {/each}
          </tbody>
        </table>
      </div>
    </Card>
  {/if}
</div>

<!-- Edit User Modal -->
{#if showEditModal && editingUser}
  <div class="fixed inset-0 modal-backdrop-enhanced flex items-start justify-center p-4 pt-16 sm:pt-20 overflow-y-auto z-50"
       role="dialog"
       aria-modal="true"
       aria-labelledby="edit-modal-title"
       on:click={() => { showEditModal = false; editingUser = null; }}
       on:keydown={(e) => e.key === 'Escape' && (showEditModal = false, editingUser = null)}>
    <div class="max-w-lg w-full my-auto modal-content-wrapper" 
         role="document"
         on:click|stopPropagation
         on:keydown|stopPropagation>
      <Card class="p-6 border-2 shadow-2xl">
      <div class="flex items-center space-x-3 mb-6">
        <div class="w-10 h-10 bg-primary/10 rounded-lg flex items-center justify-center">
          <Icon name="edit" size={20} className="text-primary" />
        </div>
        <div>
          <h2 id="edit-modal-title" class="text-xl font-bold text-foreground">Gebruiker Bewerken</h2>
          <p class="text-sm text-muted-foreground">Bewerk gebruikersgegevens en rechten</p>
        </div>
      </div>
      
      <form on:submit|preventDefault={() => updateUser(editingUser)} class="space-y-6">
        <div class="space-y-2">
          <Label for="edit-name">Naam</Label>
          <Input
            id="edit-name"
            type="text"
            bind:value={editingUser.name}
            required
            placeholder="Volledige naam"
          />
        </div>
        
        <div class="space-y-2">
          <Label for="edit-email">Email</Label>
          <Input
            id="edit-email"
            type="email"
            bind:value={editingUser.email}
            required
            placeholder="email@voorbeeld.nl"
          />
        </div>

        <div class="space-y-2">
          <Label for="edit-role">Rol</Label>
          <select 
            id="edit-role"
            bind:value={editingUser.role}
            class="w-full px-3 py-2 border border-input rounded-md bg-background text-foreground focus:outline-none focus:ring-2 focus:ring-ring focus:border-transparent"
          >
            <option value="user">User</option>
            <option value="admin">Admin</option>
            <option value="superadmin">SuperAdmin</option>
          </select>
        </div>

        <div class="flex items-center space-x-2">
          <input 
            id="edit-active"
            type="checkbox" 
            bind:checked={editingUser.is_active}
            class="rounded border-gray-300 text-primary focus:ring-primary focus:ring-offset-0"
          />
          <Label for="edit-active">Account actief</Label>
        </div>
        
        <div class="flex space-x-3 pt-4">
          <Button
            type="button"
            variant="outline"
            on:click={() => {
              showEditModal = false;
              editingUser = null;
            }}
            class="flex-1"
          >
            Annuleren
          </Button>
          <Button
            type="submit"
            class="flex-1"
          >
            Opslaan
          </Button>
        </div>
      </form>
      </Card>
    </div>
  </div>
{/if}

<!-- Delete User Modal -->
{#if showDeleteModal && deletingUser}
  <div class="fixed inset-0 modal-backdrop-enhanced flex items-start justify-center p-4 pt-16 sm:pt-20 overflow-y-auto z-50"
       role="dialog"
       aria-modal="true"
       aria-labelledby="delete-modal-title"
       on:click={() => { showDeleteModal = false; deletingUser = null; }}
       on:keydown={(e) => e.key === 'Escape' && (showDeleteModal = false, deletingUser = null)}>
    <div class="max-w-lg w-full my-auto modal-content-wrapper" 
         role="document"
         on:click|stopPropagation
         on:keydown|stopPropagation>
      <Card class="p-6 border-2 shadow-2xl">
      <div class="flex items-center space-x-3 mb-6">
        <div class="w-10 h-10 bg-destructive/10 rounded-lg flex items-center justify-center">
          <Icon name="trash" size={20} className="text-destructive" />
        </div>
        <div>
          <h2 id="delete-modal-title" class="text-xl font-bold text-foreground">Gebruiker Verwijderen</h2>
          <p class="text-sm text-muted-foreground">Dit kan niet ongedaan gemaakt worden</p>
        </div>
      </div>
      
      <div class="space-y-4">
        <p class="text-foreground">
          Weet je zeker dat je <strong>{deletingUser.name}</strong> ({deletingUser.email}) wilt verwijderen?
        </p>
        
        <div class="bg-destructive/10 border border-destructive/20 rounded-lg p-4">
          <div class="flex items-center space-x-2 text-destructive">
            <Icon name="warning" size={16} />
            <span class="font-medium text-sm">Waarschuwing</span>
          </div>
          <p class="text-destructive text-sm mt-2">
            Alle projecten en data van deze gebruiker blijven bestaan, maar zijn niet meer toegankelijk.
          </p>
        </div>
        
        <div class="flex space-x-3 pt-4">
          <Button
            type="button"
            variant="outline"
            on:click={() => {
              showDeleteModal = false;
              deletingUser = null;
            }}
            class="flex-1"
          >
            Annuleren
          </Button>
          <Button
            variant="destructive"
            on:click={() => deleteUser(deletingUser)}
            class="flex-1"
          >
            Verwijderen
          </Button>
        </div>
      </div>
      </Card>
    </div>
  </div>
{/if}

<!-- Create User Modal -->
{#if showCreateModal}
  <div class="fixed inset-0 modal-backdrop-enhanced flex items-start justify-center p-4 pt-16 sm:pt-20 overflow-y-auto z-50"
       role="dialog"
       aria-modal="true"
       aria-labelledby="create-modal-title"
       on:click={() => showCreateModal = false}
       on:keydown={(e) => e.key === 'Escape' && (showCreateModal = false)}>
    <div class="max-w-lg w-full my-auto modal-content-wrapper" 
         role="document"
         on:click|stopPropagation
         on:keydown|stopPropagation>
      <Card class="p-6 border-2 shadow-2xl">
      <div class="flex items-center space-x-3 mb-6">
        <div class="w-10 h-10 bg-primary/10 rounded-lg flex items-center justify-center">
          <Icon name="users" size={20} className="text-primary" />
        </div>
        <div>
          <h2 id="create-modal-title" class="text-xl font-bold text-foreground">Nieuwe Gebruiker</h2>
          <p class="text-sm text-muted-foreground">Maak een nieuwe gebruiker aan</p>
        </div>
      </div>
      
      <form on:submit|preventDefault={createUser} class="space-y-6">
        <div class="space-y-2">
          <Label for="create-name">Naam</Label>
          <Input
            id="create-name"
            type="text"
            bind:value={newUser.name}
            required
            placeholder="Volledige naam"
          />
        </div>
        
        <div class="space-y-2">
          <Label for="create-email">Email</Label>
          <Input
            id="create-email"
            type="email"
            bind:value={newUser.email}
            required
            placeholder="email@voorbeeld.nl"
          />
        </div>

        <div class="space-y-2">
          <Label for="create-password">Wachtwoord</Label>
          <Input
            id="create-password"
            type="password"
            bind:value={newUser.password}
            on:input={() => validatePassword(newUser.password)}
            required
            placeholder="Wachtwoord"
          />
          
          <!-- Password Requirements -->
          <div class="mt-3 p-3 bg-gray-50 dark:bg-gray-800 rounded-lg">
            <p class="text-sm font-medium text-gray-700 dark:text-gray-300 mb-2">Wachtwoord vereisten:</p>
            <div class="space-y-1">
              <div class="flex items-center space-x-2">
                <Icon name={passwordValidation.minLength ? "check" : "x"} 
                      size={14} 
                      className={passwordValidation.minLength ? "text-green-500" : "text-red-500"} />
                <span class="text-xs {passwordValidation.minLength ? 'text-green-600' : 'text-red-600'}">
                  Minimaal 8 karakters
                </span>
              </div>
              <div class="flex items-center space-x-2">
                <Icon name={passwordValidation.hasUppercase ? "check" : "x"} 
                      size={14} 
                      className={passwordValidation.hasUppercase ? "text-green-500" : "text-red-500"} />
                <span class="text-xs {passwordValidation.hasUppercase ? 'text-green-600' : 'text-red-600'}">
                  Hoofdletter (A-Z)
                </span>
              </div>
              <div class="flex items-center space-x-2">
                <Icon name={passwordValidation.hasLowercase ? "check" : "x"} 
                      size={14} 
                      className={passwordValidation.hasLowercase ? "text-green-500" : "text-red-500"} />
                <span class="text-xs {passwordValidation.hasLowercase ? 'text-green-600' : 'text-red-600'}">
                  Kleine letter (a-z)
                </span>
              </div>
              <div class="flex items-center space-x-2">
                <Icon name={passwordValidation.hasNumber ? "check" : "x"} 
                      size={14} 
                      className={passwordValidation.hasNumber ? "text-green-500" : "text-red-500"} />
                <span class="text-xs {passwordValidation.hasNumber ? 'text-green-600' : 'text-red-600'}">
                  Cijfer (0-9)
                </span>
              </div>
              <div class="flex items-center space-x-2">
                <Icon name={passwordValidation.hasSpecialChar ? "check" : "x"} 
                      size={14} 
                      className={passwordValidation.hasSpecialChar ? "text-green-500" : "text-red-500"} />
                <span class="text-xs {passwordValidation.hasSpecialChar ? 'text-green-600' : 'text-red-600'}">
                  Speciaal karakter (!@#$%^&*)
                </span>
              </div>
            </div>
          </div>
        </div>

        <div class="space-y-2">
          <Label for="create-role">Rol</Label>
          <select 
            id="create-role"
            bind:value={newUser.role}
            class="w-full px-3 py-2 border border-input rounded-md bg-background text-foreground focus:outline-none focus:ring-2 focus:ring-ring focus:border-transparent"
          >
            <option value="user">User</option>
            <option value="admin">Admin</option>
            <option value="superadmin">SuperAdmin</option>
          </select>
        </div>

        <div class="flex items-center space-x-2">
          <input 
            id="create-active"
            type="checkbox" 
            bind:checked={newUser.is_active}
            class="rounded border-gray-300 text-primary focus:ring-primary focus:ring-offset-0"
          />
          <Label for="create-active">Account actief</Label>
        </div>
        
        <div class="flex space-x-3 pt-4">
          <Button
            type="button"
            variant="outline"
            on:click={() => {
              showCreateModal = false;
              newUser = {
                name: '',
                email: '',
                password: '',
                role: 'user',
                is_active: true
              };
            }}
            class="flex-1"
          >
            Annuleren
          </Button>
          <Button
            type="submit"
            class="flex-1"
          >
            Aanmaken
          </Button>
        </div>
      </form>
      </Card>
    </div>
  </div>
{/if}