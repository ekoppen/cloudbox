<script lang="ts">
  import { onMount } from 'svelte';
  import { API_BASE_URL, API_ENDPOINTS } from '$lib/config';
  import { auth } from '$lib/stores/auth';
  import { toast } from '$lib/stores/toast';
  import Card from '$lib/components/ui/card.svelte';
  import Button from '$lib/components/ui/button.svelte';
  import Input from '$lib/components/ui/input.svelte';
  import Label from '$lib/components/ui/label.svelte';
  import Icon from '$lib/components/ui/icon.svelte';
  import Badge from '$lib/components/ui/badge.svelte';

  interface User {
    id: number;
    email: string;
    name: string;
    role: string;
    is_active: boolean;
    created_at: string;
    last_login_at?: string;
    organization_admins?: OrganizationAdmin[];
  }

  interface Organization {
    id: number;
    name: string;
    color: string;
    is_active: boolean;
  }

  interface OrganizationAdmin {
    id: number;
    user_id: number;
    organization_id: number;
    organization: Organization;
    role: string;
    is_active: boolean;
    assigned_at: string;
  }

  let users: User[] = [];
  let organizations: Organization[] = [];
  let loading = true;
  let showCreateModal = false;
  let showEditModal = false;
  let showOrgAdminModal = false;
  let selectedUser: User | null = null;
  
  let newUser = {
    email: '',
    name: '',
    password: '',
    role: 'user',
    is_active: true
  };
  
  let editUser = {
    email: '',
    name: '',
    role: '',
    is_active: true
  };
  
  let creating = false;
  let updating = false;
  let deleting = false;
  let assigning = false;
  
  let orgAdminData = {
    organizationId: 0,
    role: 'admin'
  };

  onMount(async () => {
    await Promise.all([loadUsers(), loadOrganizations()]);
  });

  async function loadUsers() {
    loading = true;
    try {
      const response = await fetch(`${API_BASE_URL}/api/v1/admin/users`, {
        headers: {
          'Content-Type': 'application/json',
          ...auth.getAuthHeader()
        }
      });

      if (response.ok) {
        users = await response.json();
      } else {
        const error = await response.json();
        toast.error(error.error || 'Fout bij laden van gebruikers');
      }
    } catch (error) {
      console.error('Error loading users:', error);
      toast.error('Netwerkfout bij laden gebruikers');
    } finally {
      loading = false;
    }
  }

  async function loadOrganizations() {
    try {
      const response = await fetch(`${API_BASE_URL}/api/v1/organizations`, {
        headers: {
          'Content-Type': 'application/json',
          ...auth.getAuthHeader()
        }
      });

      if (response.ok) {
        organizations = await response.json();
      } else {
        console.error('Failed to load organizations');
      }
    } catch (error) {
      console.error('Error loading organizations:', error);
    }
  }

  async function createUser() {
    if (!newUser.email || !newUser.name || !newUser.password) {
      toast.error('Alle velden zijn verplicht');
      return;
    }

    creating = true;
    try {
      const response = await fetch(`${API_BASE_URL}/api/v1/admin/users`, {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
          ...auth.getAuthHeader()
        },
        body: JSON.stringify(newUser)
      });

      if (response.ok) {
        const user = await response.json();
        users = [...users, user];
        toast.success('Gebruiker succesvol aangemaakt');
        showCreateModal = false;
        newUser = { email: '', name: '', password: '', role: 'user', is_active: true };
      } else {
        const error = await response.json();
        toast.error(error.error || 'Fout bij aanmaken gebruiker');
      }
    } catch (error) {
      console.error('Error creating user:', error);
      toast.error('Netwerkfout bij aanmaken gebruiker');
    } finally {
      creating = false;
    }
  }

  async function updateUser() {
    if (!selectedUser || !editUser.email || !editUser.name) {
      toast.error('Alle velden zijn verplicht');
      return;
    }

    updating = true;
    try {
      const response = await fetch(`${API_BASE_URL}/api/v1/admin/users/${selectedUser.id}`, {
        method: 'PUT',
        headers: {
          'Content-Type': 'application/json',
          ...auth.getAuthHeader()
        },
        body: JSON.stringify(editUser)
      });

      if (response.ok) {
        const updatedUser = await response.json();
        users = users.map(u => u.id === selectedUser.id ? updatedUser : u);
        toast.success('Gebruiker succesvol bijgewerkt');
        showEditModal = false;
        selectedUser = null;
      } else {
        const error = await response.json();
        toast.error(error.error || 'Fout bij bijwerken gebruiker');
      }
    } catch (error) {
      console.error('Error updating user:', error);
      toast.error('Netwerkfout bij bijwerken gebruiker');
    } finally {
      updating = false;
    }
  }

  async function deleteUser(user: User) {
    if (!confirm(`Weet je zeker dat je gebruiker "${user.name}" wilt verwijderen?`)) {
      return;
    }

    deleting = true;
    try {
      const response = await fetch(`${API_BASE_URL}/api/v1/admin/users/${user.id}`, {
        method: 'DELETE',
        headers: {
          'Content-Type': 'application/json',
          ...auth.getAuthHeader()
        }
      });

      if (response.ok) {
        users = users.filter(u => u.id !== user.id);
        toast.success('Gebruiker succesvol verwijderd');
      } else {
        const error = await response.json();
        toast.error(error.error || 'Fout bij verwijderen gebruiker');
      }
    } catch (error) {
      console.error('Error deleting user:', error);
      toast.error('Netwerkfout bij verwijderen gebruiker');
    } finally {
      deleting = false;
    }
  }

  async function assignOrganizationAdmin() {
    if (!selectedUser || !orgAdminData.organizationId) {
      toast.error('Selecteer een organization');
      return;
    }

    assigning = true;
    try {
      const response = await fetch(API_ENDPOINTS.admin.organizationAdmins.assign, {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
          ...auth.getAuthHeader()
        },
        body: JSON.stringify({
          user_id: selectedUser.id,
          organization_id: orgAdminData.organizationId,
          role: orgAdminData.role
        })
      });

      if (response.ok) {
        toast.success('Organization admin toegewezen');
        showOrgAdminModal = false;
        await loadUsers(); // Reload to get updated organization_admins
      } else {
        const error = await response.json();
        toast.error(error.error || 'Fout bij toewijzen organization admin');
      }
    } catch (error) {
      console.error('Error assigning organization admin:', error);
      toast.error('Netwerkfout bij toewijzen organization admin');
    } finally {
      assigning = false;
    }
  }

  async function revokeOrganizationAdmin(userId: number, orgId: number) {
    if (!confirm('Weet je zeker dat je deze organization admin rechten wilt intrekken?')) {
      return;
    }

    try {
      const response = await fetch(API_ENDPOINTS.admin.organizationAdmins.revoke(userId.toString(), orgId.toString()), {
        method: 'DELETE',
        headers: {
          'Content-Type': 'application/json',
          ...auth.getAuthHeader()
        }
      });

      if (response.ok) {
        toast.success('Organization admin rechten ingetrokken');
        await loadUsers(); // Reload to get updated list
      } else {
        const error = await response.json();
        toast.error(error.error || 'Fout bij intrekken organization admin rechten');
      }
    } catch (error) {
      console.error('Error revoking organization admin:', error);
      toast.error('Netwerkfout bij intrekken organization admin rechten');
    }
  }

  function openEditModal(user: User) {
    selectedUser = user;
    editUser = {
      email: user.email,
      name: user.name,
      role: user.role,
      is_active: user.is_active
    };
    showEditModal = true;
  }

  function openOrgAdminModal(user: User) {
    selectedUser = user;
    orgAdminData = {
      organizationId: 0,
      role: 'admin'
    };
    showOrgAdminModal = true;
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

  function getRoleBadgeVariant(role: string) {
    switch (role) {
      case 'superadmin': return 'destructive';
      case 'admin': return 'default';
      case 'user': return 'secondary';
      default: return 'secondary';
    }
  }

  function getRoleLabel(role: string) {
    switch (role) {
      case 'superadmin': return 'Super Admin';
      case 'admin': return 'Project Admin';
      case 'user': return 'Gebruiker';
      default: return role;
    }
  }
</script>

<svelte:head>
  <title>Gebruikersbeheer - CloudBox Admin</title>
</svelte:head>

<div class="space-y-8">
  <!-- Page Header -->
  <div class="flex items-center justify-between">
    <div class="flex items-center space-x-4">
      <div class="w-12 h-12 bg-primary/10 rounded-2xl flex items-center justify-center">
        <Icon name="users" size={24} className="text-primary" />
      </div>
      <div>
        <h1 class="text-2xl font-bold text-foreground">Gebruikersbeheer</h1>
        <p class="text-sm text-muted-foreground">Beheer gebruikers en machtigingen</p>
      </div>
    </div>
    <Button 
      on:click={() => showCreateModal = true}
      variant="floating"
      size="icon-lg"
      iconOnly={true}
      tooltip="Nieuwe Gebruiker"
    >
      <Icon name="user-plus" size={20} />
    </Button>
  </div>

  {#if loading}
    <div class="p-16 text-center">
      <div class="w-12 h-12 bg-primary/10 rounded-xl flex items-center justify-center mx-auto mb-4">
        <div class="animate-spin rounded-full h-6 w-6 border-b-2 border-primary"></div>
      </div>
      <p class="text-foreground font-medium">Loading users...</p>
      <p class="text-sm text-muted-foreground mt-1">Please wait while we fetch user data</p>
    </div>
  {:else}
    <!-- Users Table -->
    <div class="bg-background border border-border rounded-2xl overflow-hidden">
      <div class="px-8 py-6 border-b border-border">
        <div class="flex items-center justify-between">
          <span class="text-lg font-semibold text-foreground">Alle gebruikers ({users.length})</span>
        </div>
      </div>

      <div class="overflow-x-auto">
        <table class="w-full">
          <thead>
            <tr class="border-b border-border bg-muted/30">
              <th class="text-left py-4 px-6 text-sm font-semibold text-foreground">User</th>
              <th class="text-left py-4 px-6 text-sm font-semibold text-foreground">Role</th>
              <th class="text-left py-4 px-6 text-sm font-semibold text-foreground">Organizations</th>
              <th class="text-left py-4 px-6 text-sm font-semibold text-foreground">Status</th>
              <th class="text-left py-4 px-6 text-sm font-semibold text-foreground">Last Login</th>
              <th class="text-right py-4 px-6 text-sm font-semibold text-foreground">Actions</th>
            </tr>
          </thead>
          <tbody>
            {#each users as user (user.id)}
              <tr class="border-b border-border hover:bg-muted/30 transition-colors">
                <td class="py-6 px-6">
                  <div class="flex items-center space-x-3">
                    <div class="w-10 h-10 bg-primary/10 rounded-xl flex items-center justify-center">
                      <Icon name="user" size={16} className="text-primary" />
                    </div>
                    <div>
                      <div class="font-semibold text-foreground">{user.name}</div>
                      <div class="text-sm text-muted-foreground">{user.email}</div>
                    </div>
                  </div>
                </td>
                <td class="py-6 px-6">
                  <Badge variant={getRoleBadgeVariant(user.role)} className="font-medium">
                    {getRoleLabel(user.role)}
                  </Badge>
                </td>
                <td class="py-6 px-6">
                  {#if user.organization_admins && user.organization_admins.length > 0}
                    <div class="flex flex-wrap gap-2">
                      {#each user.organization_admins as orgAdmin}
                        <div class="flex items-center space-x-2 bg-muted px-3 py-1 rounded-full text-sm">
                          <div 
                            class="w-2 h-2 rounded-full"
                            style="background-color: {orgAdmin.organization.color}"
                          ></div>
                          <span class="font-medium">{orgAdmin.organization.name}</span>
                          <button
                            on:click={() => revokeOrganizationAdmin(user.id, orgAdmin.organization_id)}
                            class="text-red-600 hover:text-red-700 ml-1 p-0.5 rounded-full hover:bg-red-50"
                            title="Remove assignment"
                          >
                            <Icon name="x" size={12} />
                          </button>
                        </div>
                      {/each}
                    </div>
                  {:else}
                    <span class="text-sm text-muted-foreground">No assignments</span>
                  {/if}
                </td>
                <td class="py-6 px-6">
                  <Badge variant={user.is_active ? "default" : "secondary"} className="font-medium">
                    {user.is_active ? 'Active' : 'Inactive'}
                  </Badge>
                </td>
                <td class="py-6 px-6 text-sm text-muted-foreground">
                  {user.last_login_at ? formatDate(user.last_login_at) : 'Never'}
                </td>
                <td class="py-6 px-6 text-right">
                  <div class="flex items-center justify-end space-x-2">
                    {#if user.role === 'admin'}
                      <Button
                        variant="ghost"
                        size="sm"
                        on:click={() => openOrgAdminModal(user)}
                        title="Assign to organization"
                        className="text-muted-foreground hover:text-foreground"
                      >
                        <Icon name="building" size={16} />
                      </Button>
                    {/if}
                    <Button
                      variant="ghost"
                      size="sm"
                      on:click={() => openEditModal(user)}
                      className="text-muted-foreground hover:text-foreground"
                    >
                      <Icon name="edit" size={16} />
                    </Button>
                    {#if user.id !== $auth.user?.id}
                      <Button
                        variant="ghost"
                        size="sm"
                        disabled={deleting}
                        on:click={() => deleteUser(user)}
                        className="text-red-600 hover:text-red-700 hover:bg-red-50"
                      >
                        <Icon name="trash" size={16} />
                      </Button>
                    {/if}
                  </div>
                </td>
              </tr>
            {/each}
          </tbody>
        </table>
      </div>
    </div>
  {/if}
</div>

<!-- Create User Modal -->
{#if showCreateModal}
  <div class="fixed inset-0 modal-backdrop-enhanced flex items-start justify-center p-4 pt-16 sm:pt-20 overflow-y-auto z-50"
       on:click={() => showCreateModal = false}>
    <div class="max-w-md w-full my-auto modal-content-wrapper" on:click|stopPropagation>
      <Card class="p-6 border-2 shadow-2xl">
      <div class="flex items-center space-x-3 mb-6">
        <div class="w-10 h-10 bg-primary/10 rounded-lg flex items-center justify-center">
          <Icon name="user" size={20} className="text-primary" />
        </div>
        <div>
          <h2 class="text-xl font-bold text-foreground">Nieuwe Gebruiker</h2>
          <p class="text-sm text-muted-foreground">Maak een nieuwe CloudBox gebruiker aan</p>
        </div>
      </div>
      
      <form on:submit|preventDefault={createUser} class="space-y-4">
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
            placeholder="email@example.com"
          />
        </div>

        <div class="space-y-2">
          <Label for="create-password">Wachtwoord</Label>
          <Input
            id="create-password"
            type="password"
            bind:value={newUser.password}
            required
            placeholder="Minimaal 8 karakters"
          />
        </div>

        <div class="space-y-2">
          <Label for="create-role">Rol</Label>
          <select
            id="create-role"
            bind:value={newUser.role}
            class="w-full px-3 py-2 border border-border rounded-md bg-background text-foreground focus:outline-none focus:ring-2 focus:ring-primary focus:border-transparent"
          >
            <option value="user">Gebruiker</option>
            <option value="admin">Project Admin</option>
            <option value="superadmin">Super Admin</option>
          </select>
        </div>

        <div class="flex items-center space-x-2">
          <input
            id="create-active"
            type="checkbox"
            bind:checked={newUser.is_active}
            class="rounded border-border text-primary focus:ring-primary"
          />
          <Label for="create-active">Account actief</Label>
        </div>

        <div class="flex space-x-3 pt-4">
          <Button
            type="submit"
            disabled={creating}
            class="flex-1"
          >
            {#if creating}
              <Icon name="loader" size={16} class="mr-2 animate-spin" />
              Aanmaken...
            {:else}
              <Icon name="user" size={16} class="mr-2" />
              Gebruiker Aanmaken
            {/if}
          </Button>
          <Button
            type="button"
            variant="outline"
            on:click={() => showCreateModal = false}
            disabled={creating}
          >
            Annuleren
          </Button>
        </div>
      </form>
      </Card>
    </div>
  </div>
{/if}

<!-- Edit User Modal -->
{#if showEditModal && selectedUser}
  <div class="fixed inset-0 modal-backdrop-enhanced flex items-start justify-center p-4 pt-16 sm:pt-20 overflow-y-auto z-50"
       on:click={() => showEditModal = false}>
    <div class="max-w-md w-full my-auto modal-content-wrapper" on:click|stopPropagation>
      <Card class="p-6 border-2 shadow-2xl">
      <div class="flex items-center space-x-3 mb-6">
        <div class="w-10 h-10 bg-primary/10 rounded-lg flex items-center justify-center">
          <Icon name="edit" size={20} className="text-primary" />
        </div>
        <div>
          <h2 class="text-xl font-bold text-foreground">Gebruiker Bewerken</h2>
          <p class="text-sm text-muted-foreground">Wijzig gebruikersgegevens</p>
        </div>
      </div>
      
      <form on:submit|preventDefault={updateUser} class="space-y-4">
        <div class="space-y-2">
          <Label for="edit-name">Naam</Label>
          <Input
            id="edit-name"
            type="text"
            bind:value={editUser.name}
            required
            placeholder="Volledige naam"
          />
        </div>

        <div class="space-y-2">
          <Label for="edit-email">Email</Label>
          <Input
            id="edit-email"
            type="email"
            bind:value={editUser.email}
            required
            placeholder="email@example.com"
          />
        </div>

        <div class="space-y-2">
          <Label for="edit-role">Rol</Label>
          <select
            id="edit-role"
            bind:value={editUser.role}
            class="w-full px-3 py-2 border border-border rounded-md bg-background text-foreground focus:outline-none focus:ring-2 focus:ring-primary focus:border-transparent"
          >
            <option value="user">Gebruiker</option>
            <option value="admin">Project Admin</option>
            <option value="superadmin">Super Admin</option>
          </select>
        </div>

        <div class="flex items-center space-x-2">
          <input
            id="edit-active"
            type="checkbox"
            bind:checked={editUser.is_active}
            class="rounded border-border text-primary focus:ring-primary"
          />
          <Label for="edit-active">Account actief</Label>
        </div>

        <div class="flex space-x-3 pt-4">
          <Button
            type="submit"
            disabled={updating}
            class="flex-1"
          >
            {#if updating}
              <Icon name="loader" size={16} class="mr-2 animate-spin" />
              Bijwerken...
            {:else}
              <Icon name="check" size={16} class="mr-2" />
              Opslaan
            {/if}
          </Button>
          <Button
            type="button"
            variant="outline"
            on:click={() => showEditModal = false}
            disabled={updating}
          >
            Annuleren
          </Button>
        </div>
      </form>
      </Card>
    </div>
  </div>
{/if}
<!-- Assign Organization Admin Modal -->
{#if showOrgAdminModal && selectedUser}
  <div class="fixed inset-0 modal-backdrop-enhanced flex items-start justify-center p-4 pt-16 sm:pt-20 overflow-y-auto z-50"
       on:click={() => showOrgAdminModal = false}>
    <div class="max-w-md w-full my-auto modal-content-wrapper" on:click|stopPropagation>
      <Card class="p-6 border-2 shadow-2xl">
      <div class="flex items-center space-x-3 mb-6">
        <div class="w-10 h-10 bg-primary/10 rounded-lg flex items-center justify-center">
          <Icon name="package" size={20} className="text-primary" />
        </div>
        <div>
          <h2 class="text-xl font-bold text-foreground">Organization Admin Toewijzen</h2>
          <p class="text-sm text-muted-foreground">Wijs {selectedUser.name} toe aan een organization</p>
        </div>
      </div>
      
      <form on:submit|preventDefault={assignOrganizationAdmin} class="space-y-4">
        <div class="space-y-2">
          <Label for="org-select">Organization</Label>
          <select
            id="org-select"
            bind:value={orgAdminData.organizationId}
            required
            class="w-full px-3 py-2 border border-border rounded-md bg-background text-foreground focus:outline-none focus:ring-2 focus:ring-primary focus:border-transparent"
          >
            <option value={0}>Selecteer een organization</option>
            {#each organizations as org}
              <option value={org.id}>{org.name}</option>
            {/each}
          </select>
        </div>

        <div class="space-y-2">
          <Label for="org-role">Rol</Label>
          <select
            id="org-role"
            bind:value={orgAdminData.role}
            class="w-full px-3 py-2 border border-border rounded-md bg-background text-foreground focus:outline-none focus:ring-2 focus:ring-primary focus:border-transparent"
          >
            <option value="admin">Admin</option>
          </select>
          <p class="text-xs text-muted-foreground">
            Admin heeft beheerrechten over alle projecten binnen de organization
          </p>
        </div>

        <div class="flex space-x-3 pt-4">
          <Button
            type="submit"
            disabled={assigning || !orgAdminData.organizationId}
            class="flex-1"
          >
            {#if assigning}
              <Icon name="loader" size={16} class="mr-2 animate-spin" />
              Toewijzen...
            {:else}
              <Icon name="package" size={16} class="mr-2" />
              Toewijzen
            {/if}
          </Button>
          <Button
            type="button"
            variant="outline"
            on:click={() => showOrgAdminModal = false}
            disabled={assigning}
          >
            Annuleren
          </Button>
        </div>
      </form>
      </Card>
    </div>
  </div>
{/if}
