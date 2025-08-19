<script lang="ts">
  import { goto } from '$app/navigation';
  import { auth } from '$lib/stores/auth';
  import { API_ENDPOINTS, createApiRequest } from '$lib/config';
  import Card from '$lib/components/ui/card.svelte';
  import Button from '$lib/components/ui/button.svelte';
  import Input from '$lib/components/ui/input.svelte';
  import Label from '$lib/components/ui/label.svelte';
  import Icon from '$lib/components/ui/icon.svelte';
  
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

<div class="min-h-screen bg-gradient-to-br from-background to-muted/30 flex items-center justify-center p-4">
  <div class="w-full max-w-md">
    <!-- Logo and Header -->
    <div class="text-center mb-8">
      <div class="w-16 h-16 bg-primary/10 rounded-2xl flex items-center justify-center mx-auto mb-4">
        <span class="text-2xl">☁️</span>
      </div>
      <h1 class="text-3xl font-bold text-foreground mb-2 font-['Inter']">CloudBox</h1>
      <p class="text-muted-foreground">Sign in to your account</p>
    </div>

    <!-- Login Card -->
    <div class="bg-background border border-border rounded-2xl shadow-xl p-8">
      <div class="mb-6">
        <h2 class="text-2xl font-semibold text-foreground mb-2">Welcome back</h2>
        <p class="text-sm text-muted-foreground">Enter your credentials to continue</p>
      </div>

      {#if error}
        <div class="bg-red-50 border border-red-200 rounded-xl p-4 mb-6">
          <div class="flex items-center space-x-2">
            <div class="w-5 h-5 bg-red-100 rounded-full flex items-center justify-center">
              <Icon name="alert-triangle" size={12} className="text-red-600" />
            </div>
            <p class="text-red-800 text-sm font-medium">{error}</p>
          </div>
        </div>
      {/if}

      <form on:submit|preventDefault={handleLogin} class="space-y-6">
        <div class="space-y-4">
          <div>
            <Label for="email" className="text-sm font-medium text-foreground mb-2 block">Email address</Label>
            <Input
              id="email"
              type="email"
              bind:value={email}
              required
              placeholder="you@example.com"
              className="h-12 text-base border-border focus:border-primary focus:ring-1 focus:ring-primary"
            />
          </div>

          <div>
            <Label for="password" className="text-sm font-medium text-foreground mb-2 block">Password</Label>
            <Input
              id="password"
              type="password"
              bind:value={password}
              required
              placeholder="Enter your password"
              className="h-12 text-base border-border focus:border-primary focus:ring-1 focus:ring-primary"
            />
          </div>
        </div>

        <div class="flex items-center justify-between">
          <div class="flex items-center space-x-2">
            <input
              id="remember"
              type="checkbox"
              bind:checked={rememberMe}
              class="w-4 h-4 rounded border-border text-primary focus:ring-primary focus:ring-2 focus:ring-offset-2"
            />
            <Label for="remember" class="text-sm text-muted-foreground cursor-pointer">
              Remember me for 30 days
            </Label>
          </div>
        </div>

        <Button
          type="submit"
          disabled={loading}
          class="w-full h-12 text-base font-medium"
          size="lg"
        >
          {#if loading}
            <div class="animate-spin rounded-full h-5 w-5 border-b-2 border-white mr-2"></div>
            Signing in...
          {:else}
            Sign in to CloudBox
          {/if}
        </Button>
      </form>

      <!-- Footer -->
      <div class="mt-8 pt-6 border-t border-border text-center">
        <p class="text-sm text-muted-foreground">
          Need help? Contact your administrator
        </p>
      </div>
    </div>
  </div>
</div>