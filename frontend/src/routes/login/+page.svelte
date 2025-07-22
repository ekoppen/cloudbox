<script lang="ts">
  import { goto } from '$app/navigation';
  import { auth } from '$lib/stores/auth';
  import { API_ENDPOINTS, createApiRequest } from '$lib/config';
  import Card from '$lib/components/ui/card.svelte';
  import Button from '$lib/components/ui/button.svelte';
  import Input from '$lib/components/ui/input.svelte';
  import Label from '$lib/components/ui/label.svelte';
  
  let email = '';
  let password = '';
  let rememberMe = false;
  let loading = false;
  let error = '';

  async function handleLogin() {
    if (!email || !password) {
      error = 'Vul alle velden in';
      return;
    }

    loading = true;
    error = '';

    try {
      const response = await createApiRequest(API_ENDPOINTS.auth.login, {
        method: 'POST',
        body: JSON.stringify({ email, password, remember_me: rememberMe }),
      });

      const data = await response.json();

      if (response.ok) {
        // Use auth store to manage state
        auth.login(data.user, data.token, data.refresh_token);
        
        // Redirect to dashboard
        goto('/dashboard');
      } else {
        error = data.error || 'Inloggen mislukt';
      }
    } catch (err) {
      error = 'Netwerkfout. Controleer of de backend draait.';
      console.error('Login error:', err);
    } finally {
      loading = false;
    }
  }
</script>

<svelte:head>
  <title>Inloggen - CloudBox</title>
</svelte:head>

<div class="min-h-screen bg-background flex items-center justify-center p-4">
  <Card class="w-full max-w-md p-8">
    <div class="text-center mb-8">
      <h1 class="text-3xl font-bold text-foreground mb-2">☁️ CloudBox</h1>
      <h2 class="text-xl font-semibold text-foreground mb-2">Inloggen</h2>
      <p class="text-muted-foreground">Voer je gegevens in om door te gaan</p>
    </div>

    {#if error}
      <Card class="bg-destructive/10 border-destructive/20 p-4 mb-6">
        <p class="text-destructive text-sm">{error}</p>
      </Card>
    {/if}

    <form on:submit|preventDefault={handleLogin} class="space-y-6">
      <div class="space-y-2">
        <Label for="email">Email adres</Label>
        <Input
          id="email"
          type="email"
          bind:value={email}
          required
          placeholder="jouw@email.com"
        />
      </div>

      <div class="space-y-2">
        <Label for="password">Wachtwoord</Label>
        <Input
          id="password"
          type="password"
          bind:value={password}
          required
          placeholder="••••••••"
        />
      </div>

      <div class="flex items-center space-x-2">
        <input
          id="remember"
          type="checkbox"
          bind:checked={rememberMe}
          class="rounded border-border text-primary focus:ring-primary"
        />
        <Label for="remember" class="text-sm text-muted-foreground cursor-pointer">
          Ingelogd blijven (30 dagen)
        </Label>
      </div>

      <Button
        type="submit"
        disabled={loading}
        class="w-full"
        size="lg"
      >
        {loading ? 'Bezig met inloggen...' : 'Inloggen'}
      </Button>
    </form>

  </Card>
</div>