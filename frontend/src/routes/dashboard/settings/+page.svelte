<script lang="ts">
  import { theme, type AccentColor } from '$lib/stores/theme';
  import { auth } from '$lib/stores/auth';
  import { toast } from '$lib/stores/toast';
  import { API_ENDPOINTS } from '$lib/config';
  import Card from '$lib/components/ui/card.svelte';
  import Button from '$lib/components/ui/button.svelte';
  import Input from '$lib/components/ui/input.svelte';
  import Label from '$lib/components/ui/label.svelte';
  import Icon from '$lib/components/ui/icon.svelte';
  
  let showEditProfileModal = false;
  let editProfileData = {
    name: $auth.user?.name || '',
    email: $auth.user?.email || ''
  };
  let updatingProfile = false;
  
  let showChangePasswordModal = false;
  let changePasswordData = {
    currentPassword: '',
    newPassword: '',
    confirmPassword: ''
  };
  let changingPassword = false;
  
  const accentColors: { name: AccentColor; label: string; preview: string; darkPreview: string }[] = [
    { name: 'blue', label: 'Blauw', preview: 'bg-blue-500', darkPreview: 'bg-blue-400' },
    { name: 'green', label: 'Groen', preview: 'bg-green-600', darkPreview: 'bg-green-400' },
    { name: 'purple', label: 'Paars', preview: 'bg-purple-500', darkPreview: 'bg-purple-300' },
    { name: 'orange', label: 'Oranje', preview: 'bg-orange-500', darkPreview: 'bg-orange-400' },
    { name: 'red', label: 'Rood', preview: 'bg-red-500', darkPreview: 'bg-red-300' },
    { name: 'pink', label: 'Roze', preview: 'bg-pink-500', darkPreview: 'bg-pink-300' },
  ];

  function handleAccentColorChange(accentColor: AccentColor) {
    console.log('Settings: changing accent color to', accentColor);
    theme.setAccentColor(accentColor);
  }

  function openEditProfileModal() {
    editProfileData = {
      name: $auth.user?.name || '',
      email: $auth.user?.email || ''
    };
    showEditProfileModal = true;
  }

  async function updateProfile() {
    if (!editProfileData.name.trim()) {
      toast.error('Naam is verplicht');
      return;
    }

    updatingProfile = true;
    try {
      const response = await fetch(API_ENDPOINTS.auth.me, {
        method: 'PUT',
        headers: {
          'Authorization': `Bearer ${$auth.token}`,
          'Content-Type': 'application/json',
        },
        body: JSON.stringify({
          name: editProfileData.name.trim()
        }),
      });

      if (response.ok) {
        // Update the auth store with the new name
        auth.updateUser({ ...$auth.user, name: editProfileData.name.trim() });
        toast.success('Profiel succesvol bijgewerkt');
        showEditProfileModal = false;
      } else {
        const errorData = await response.json();
        toast.error(errorData.error || 'Fout bij het bijwerken van profiel');
      }
    } catch (error) {
      console.error('Profile update error:', error);
      toast.error('Netwerk fout bij het bijwerken van profiel');
    } finally {
      updatingProfile = false;
    }
  }

  function openChangePasswordModal() {
    changePasswordData = {
      currentPassword: '',
      newPassword: '',
      confirmPassword: ''
    };
    showChangePasswordModal = true;
  }

  async function changePassword() {
    if (!changePasswordData.currentPassword) {
      toast.error('Huidig wachtwoord is verplicht');
      return;
    }
    
    if (!changePasswordData.newPassword) {
      toast.error('Nieuw wachtwoord is verplicht');
      return;
    }
    
    if (changePasswordData.newPassword.length < 6) {
      toast.error('Nieuw wachtwoord moet minimaal 6 karakters lang zijn');
      return;
    }
    
    if (changePasswordData.newPassword !== changePasswordData.confirmPassword) {
      toast.error('Wachtwoorden komen niet overeen');
      return;
    }

    changingPassword = true;
    try {
      const response = await fetch(API_ENDPOINTS.auth.changePassword, {
        method: 'PUT',
        headers: {
          'Authorization': `Bearer ${$auth.token}`,
          'Content-Type': 'application/json',
        },
        body: JSON.stringify({
          current_password: changePasswordData.currentPassword,
          new_password: changePasswordData.newPassword
        }),
      });

      if (response.ok) {
        toast.success('Wachtwoord succesvol gewijzigd');
        showChangePasswordModal = false;
      } else {
        const errorData = await response.json();
        toast.error(errorData.error || 'Fout bij het wijzigen van wachtwoord');
      }
    } catch (error) {
      console.error('Password change error:', error);
      toast.error('Netwerk fout bij het wijzigen van wachtwoord');
    } finally {
      changingPassword = false;
    }
  }

  // Debug function to check current classes
  function debugClasses() {
    if (typeof document !== 'undefined') {
      const classes = document.documentElement.classList.toString();
      console.log('Current HTML classes:', classes);
      
      // Check computed styles
      const computedStyle = getComputedStyle(document.documentElement);
      console.log('CSS Variables:', {
        background: computedStyle.getPropertyValue('--background'),
        foreground: computedStyle.getPropertyValue('--foreground'),
        primary: computedStyle.getPropertyValue('--primary'),
        card: computedStyle.getPropertyValue('--card')
      });
    }
  }
</script>

<svelte:head>
  <title>Profiel - CloudBox</title>
</svelte:head>

<div class="space-y-8">
  <!-- Page Header -->
  <div class="flex items-center justify-between">
    <div>
      <h1 class="text-3xl font-bold text-foreground">Profiel</h1>
      <p class="text-muted-foreground mt-2">Pas uw profiel, voorkeuren en thema aan</p>
    </div>
  </div>

  <!-- Theme Settings -->
  <Card class="p-6">
    <div class="flex items-center space-x-3 mb-6">
      <Icon name="palette" size={24} />
      <h2 class="text-xl font-semibold text-card-foreground">Thema & Uiterlijk</h2>
    </div>

    <div class="space-y-6">
      <!-- Dark/Light Mode -->
      <div>
        <h3 class="text-lg font-medium text-card-foreground mb-3">Thema Modus</h3>
        <div class="flex space-x-4">
          <button
            on:click={() => theme.setTheme('light')}
            class="flex items-center space-x-3 p-4 border-2 rounded-lg transition-colors"
            class:border-primary={$theme.theme === 'light'}
            class:border-border={$theme.theme !== 'light'}
            class:bg-primary-50={$theme.theme === 'light'}
          >
            <Icon name="sun" size={20} />
            <div class="text-left">
              <div class="font-medium text-card-foreground">Licht</div>
              <div class="text-sm text-muted-foreground">Lichte achtergrond</div>
            </div>
          </button>
          
          <button
            on:click={() => theme.setTheme('dark')}
            class="flex items-center space-x-3 p-4 border-2 rounded-lg transition-colors"
            class:border-primary={$theme.theme === 'dark'}
            class:border-border={$theme.theme !== 'dark'}
            class:bg-primary-50={$theme.theme === 'dark'}
          >
            <Icon name="moon" size={20} />
            <div class="text-left">
              <div class="font-medium text-card-foreground">Donker</div>
              <div class="text-sm text-muted-foreground">Donkere achtergrond</div>
            </div>
          </button>
        </div>
      </div>

      <!-- Accent Color -->
      <div>
        <h3 class="text-lg font-medium text-card-foreground mb-3">Accent Kleur</h3>
        <div class="grid grid-cols-2 gap-3">
          {#each accentColors as color}
            <button
              on:click={() => handleAccentColorChange(color.name)}
              class="flex items-center space-x-3 p-3 border-2 rounded-lg transition-all duration-200 hover:scale-105"
              class:border-primary={$theme.accentColor === color.name}
              class:border-border={$theme.accentColor !== color.name}
              class:bg-muted={$theme.accentColor === color.name}
              class:shadow-md={$theme.accentColor === color.name}
            >
              <div class="w-5 h-5 rounded-full {$theme.theme === 'dark' ? color.darkPreview : color.preview} shadow-sm"></div>
              <span class="font-medium text-card-foreground">{color.label}</span>
              {#if $theme.accentColor === color.name}
                <Icon name="user" size={14} className="ml-auto text-primary" />
              {/if}
            </button>
          {/each}
        </div>
      </div>

      <!-- Theme Preview -->
      <div>
        <h3 class="text-lg font-medium text-card-foreground mb-3">Live Preview</h3>
        <div class="border border-border rounded-lg p-4 space-y-4">
          <!-- Header Preview -->
          <div class="flex items-center justify-between p-3 bg-card rounded border border-border">
            <div class="flex items-center space-x-3">
              <div class="w-8 h-8 bg-primary rounded-full flex items-center justify-center">
                <Icon name="cloud" size={16} color="white" />
              </div>
              <div>
                <div class="font-medium text-card-foreground">CloudBox</div>
                <div class="text-sm text-muted-foreground">Preview van je theme</div>
              </div>
            </div>
            <div class="bg-primary text-primary-foreground px-3 py-1 rounded text-sm font-medium">
              Actief
            </div>
          </div>
          
          <!-- Button Preview -->
          <div class="space-y-3">
            <div class="flex space-x-3">
              <button class="bg-primary text-primary-foreground px-4 py-2 rounded font-medium hover:opacity-90 transition-opacity">
                Primary Button
              </button>
              <button class="bg-secondary text-secondary-foreground px-4 py-2 rounded font-medium border border-border hover:bg-muted transition-colors">
                Secondary Button
              </button>
            </div>
            
            <!-- Test different approaches -->
            <div class="space-y-2">
              <div class="flex space-x-3">
                <button style="background-color: hsl(var(--primary)); color: hsl(var(--primary-foreground));" class="px-4 py-2 rounded font-medium">
                  Inline Style
                </button>
                <button class="theme-primary px-4 py-2 rounded font-medium">
                  Utility Class
                </button>
              </div>
              <div class="text-xs text-muted-foreground">
                Primary CSS var: <span class="font-mono">{$theme.theme === 'dark' ? 'dark' : 'light'} mode</span>
              </div>
            </div>
          </div>
          
          <!-- Color Bar Preview -->
          <div class="flex space-x-2">
            <div class="bg-primary h-3 rounded flex-1" title="Primary Color"></div>
            <div class="bg-secondary h-3 rounded flex-1" title="Secondary Color"></div>
            <div class="bg-muted h-3 rounded flex-1" title="Muted Color"></div>
            <div class="bg-accent h-3 rounded flex-1" title="Accent Color"></div>
          </div>
          
          <!-- Info -->
          <div class="text-xs text-muted-foreground text-center p-2 bg-muted rounded">
            Huidige theme: <span class="font-medium text-foreground">{$theme.theme === 'dark' ? 'Donker' : 'Licht'}</span> • 
            Accent: <span class="font-medium text-primary">{accentColors.find(c => c.name === $theme.accentColor)?.label}</span>
          </div>
          
          <!-- Debug Button -->
          <button 
            on:click={debugClasses}
            class="w-full text-xs bg-secondary text-secondary-foreground px-2 py-1 rounded hover:bg-muted transition-colors"
          >
            🔍 Debug Classes (Check Console)
          </button>
        </div>
      </div>
    </div>
  </Card>

  <!-- Account Settings -->
  <Card class="p-6">
    <div class="flex items-center space-x-3 mb-6">
      <Icon name="user" size={24} />
      <h2 class="text-xl font-semibold text-card-foreground">Account</h2>
    </div>

    <div class="space-y-4">
      <div>
        <div class="block text-sm font-medium text-card-foreground mb-2">Email</div>
        <div class="p-3 bg-muted rounded-lg text-muted-foreground">
          {$auth.user?.email || 'Niet beschikbaar'}
        </div>
      </div>
      
      <div>
        <div class="block text-sm font-medium text-card-foreground mb-2">Naam</div>
        <div class="p-3 bg-muted rounded-lg text-muted-foreground">
          {$auth.user?.name || 'Niet beschikbaar'}
        </div>
      </div>

      <div class="pt-4 flex space-x-3">
        <Button variant="outline" size="sm" on:click={openEditProfileModal}>
          <Icon name="edit" size={16} class="mr-2" />
          Profiel Bewerken
        </Button>
        <Button variant="outline" size="sm" on:click={openChangePasswordModal}>
          <Icon name="key" size={16} class="mr-2" />
          Wachtwoord Wijzigen
        </Button>
      </div>
    </div>
  </Card>

</div>

<!-- Edit Profile Modal -->
{#if showEditProfileModal}
  <div class="fixed inset-0 bg-black/80 backdrop-blur-sm flex items-center justify-center p-4 z-50">
    <Card class="max-w-md w-full p-6 border-2 shadow-2xl">
      <div class="flex items-center space-x-3 mb-6">
        <div class="w-10 h-10 bg-primary/10 rounded-lg flex items-center justify-center">
          <Icon name="user" size={20} className="text-primary" />
        </div>
        <div>
          <h2 class="text-xl font-bold text-foreground">Profiel Bewerken</h2>
          <p class="text-sm text-muted-foreground">Werk je profielgegevens bij</p>
        </div>
      </div>
      
      <form on:submit|preventDefault={updateProfile} class="space-y-4">
        <div class="space-y-2">
          <Label for="edit-name">Naam</Label>
          <Input
            id="edit-name"
            type="text"
            bind:value={editProfileData.name}
            required
            placeholder="Je naam"
          />
        </div>

        <div class="space-y-2">
          <Label for="edit-email">Email</Label>
          <Input
            id="edit-email"
            type="email"
            bind:value={editProfileData.email}
            disabled
            class="bg-muted text-muted-foreground cursor-not-allowed"
            placeholder="Email kan niet worden gewijzigd"
          />
          <p class="text-xs text-muted-foreground">
            Email adres kan niet worden gewijzigd. Neem contact op met een administrator.
          </p>
        </div>

        <div class="flex space-x-3 pt-4">
          <Button
            type="submit"
            disabled={updatingProfile}
            class="flex-1"
          >
            {#if updatingProfile}
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
            on:click={() => showEditProfileModal = false}
            disabled={updatingProfile}
          >
            Annuleren
          </Button>
        </div>
      </form>
    </Card>
  </div>
{/if}

<!-- Change Password Modal -->
{#if showChangePasswordModal}
  <div class="fixed inset-0 bg-black/80 backdrop-blur-sm flex items-center justify-center p-4 z-50">
    <Card class="max-w-md w-full p-6 border-2 shadow-2xl">
      <div class="flex items-center space-x-3 mb-6">
        <div class="w-10 h-10 bg-primary/10 rounded-lg flex items-center justify-center">
          <Icon name="key" size={20} className="text-primary" />
        </div>
        <div>
          <h2 class="text-xl font-bold text-foreground">Wachtwoord Wijzigen</h2>
          <p class="text-sm text-muted-foreground">Wijzig je account wachtwoord</p>
        </div>
      </div>
      
      <form on:submit|preventDefault={changePassword} class="space-y-4">
        <div class="space-y-2">
          <Label for="current-password">Huidig Wachtwoord</Label>
          <Input
            id="current-password"
            type="password"
            bind:value={changePasswordData.currentPassword}
            required
            placeholder="Voer je huidige wachtwoord in"
          />
        </div>

        <div class="space-y-2">
          <Label for="new-password">Nieuw Wachtwoord</Label>
          <Input
            id="new-password"
            type="password"
            bind:value={changePasswordData.newPassword}
            required
            placeholder="Voer een nieuw wachtwoord in"
          />
          <p class="text-xs text-muted-foreground">
            Minimaal 6 karakters lang
          </p>
        </div>

        <div class="space-y-2">
          <Label for="confirm-password">Bevestig Nieuw Wachtwoord</Label>
          <Input
            id="confirm-password"
            type="password"
            bind:value={changePasswordData.confirmPassword}
            required
            placeholder="Bevestig je nieuwe wachtwoord"
          />
        </div>

        <div class="flex space-x-3 pt-4">
          <Button
            type="submit"
            disabled={changingPassword}
            class="flex-1"
          >
            {#if changingPassword}
              <Icon name="loader" size={16} class="mr-2 animate-spin" />
              Wijzigen...
            {:else}
              <Icon name="key" size={16} class="mr-2" />
              Wachtwoord Wijzigen
            {/if}
          </Button>
          <Button
            type="button"
            variant="outline"
            on:click={() => showChangePasswordModal = false}
            disabled={changingPassword}
          >
            Annuleren
          </Button>
        </div>
      </form>
    </Card>
  </div>
{/if}