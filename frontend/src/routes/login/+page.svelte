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

<!-- Animated gradient background -->
<div class="login-container">
  <div class="animated-gradient"></div>
  
  <!-- Login card -->
  <div class="login-wrapper">
    <div class="login-card">
      <!-- Top content -->
      <div class="top-content">
        <!-- CloudBox logo -->
        <div class="logo-section">
          <Icon name="cloud" size={48} />
          <h1 class="logo-text-cutout">CloudBox</h1>
        </div>
        
        <!-- Subtitle -->
        <p class="subtitle">
          Sign in to your account
        </p>
      </div>
      
      <!-- Bottom content -->
      <div class="bottom-content">
        {#if error}
          <div class="error-message">
            <div class="flex items-center space-x-2">
              <Icon name="alert-triangle" size={16} className="text-red-500" />
              <p class="text-sm font-medium">{error}</p>
            </div>
          </div>
        {/if}

        <form on:submit|preventDefault={handleLogin} class="login-form">
            <div class="form-group">
              <Input
                id="email"
                type="email"
                bind:value={email}
                required
                placeholder="Email address"
                className="modern-input"
              />
            </div>

            <div class="form-group">
              <Input
                id="password"
                type="password"
                bind:value={password}
                required
                placeholder="Password"
                className="modern-input"
              />
            </div>

            <div class="form-options">
              <label class="checkbox-label">
                <input
                  type="checkbox"
                  bind:checked={rememberMe}
                  class="modern-checkbox"
                />
                <span>Remember me</span>
              </label>
            </div>

            <Button type="submit" disabled={loading} variant="primary" size="lg" class="w-full" loading={loading}>
              {#if loading}
                <span>Signing in...</span>
              {:else}
                <span>Sign In</span>
              {/if}
            </Button>
          </form>
        
        <!-- Back to home link -->
        <a href="/" class="text-link back-link">
          ← Back to Home
        </a>
      </div>
    </div>
  </div>
</div>

<style>
  .login-container {
    position: relative;
    width: 100vw;
    height: 100vh;
    overflow: hidden;
    display: flex;
    align-items: center;
    justify-content: center;
  }

  .animated-gradient {
    position: absolute;
    top: 0;
    left: 0;
    right: 0;
    bottom: 0;
    background: linear-gradient(
      45deg,
      #667eea 0%,
      #764ba2 25%,
      #f093fb 50%,
      #f5576c 75%,
      #4facfe 100%
    );
    background-size: 400% 400%;
    animation: gradientShift 15s ease infinite;
    z-index: 1;
  }

  .animated-gradient::before {
    content: '';
    position: absolute;
    top: 0;
    left: 0;
    right: 0;
    bottom: 0;
    background: linear-gradient(
      135deg,
      rgba(255, 255, 255, 0.1) 0%,
      rgba(255, 255, 255, 0) 50%,
      rgba(0, 0, 0, 0.1) 100%
    );
    animation: overlayShift 20s ease-in-out infinite;
  }

  @keyframes gradientShift {
    0% { background-position: 0% 50%; }
    50% { background-position: 100% 50%; }
    100% { background-position: 0% 50%; }
  }

  @keyframes overlayShift {
    0%, 100% { opacity: 0.3; }
    50% { opacity: 0.7; }
  }

  .login-wrapper {
    position: relative;
    z-index: 2;
    width: 100%;
    max-width: 420px;
    padding: 0 24px;
  }

  .login-card {
    background: rgba(255, 255, 255, 0.95);
    backdrop-filter: blur(20px);
    border: 1px solid rgba(255, 255, 255, 0.2);
    border-radius: 24px;
    padding: 48px 32px;
    box-shadow: 
      0 30px 60px -15px rgba(0, 0, 0, 0.3),
      0 25px 35px -15px rgba(0, 0, 0, 0.2),
      0 10px 20px -10px rgba(0, 0, 0, 0.15),
      0 0 0 1px rgba(255, 255, 255, 0.1) inset;
    animation: cardFloat 6s ease-in-out infinite;
    display: flex;
    flex-direction: column;
    justify-content: space-between;
    min-height: 500px;
  }

  @keyframes cardFloat {
    0%, 100% { transform: translateY(0px); }
    50% { transform: translateY(-10px); }
  }

  .logo-section {
    display: flex;
    flex-direction: column;
    align-items: center;
    margin-bottom: 24px;
    gap: 12px;
  }

  .logo-text-cutout {
    font-size: 64px;
    font-weight: 800;
    line-height: 1.2;
    margin: 0;
    letter-spacing: -0.025em;
    background: linear-gradient(45deg, #667eea 0%, #764ba2 25%, #f093fb 50%, #f5576c 75%, #4facfe 100%);
    background-size: 400% 400%;
    -webkit-background-clip: text;
    background-clip: text;
    -webkit-text-fill-color: transparent;
    animation: gradientShift 15s ease infinite;
    filter: brightness(1.3) saturate(1.2);
  }

  .subtitle {
    text-align: center;
    color: #6b7280;
    font-size: 16px;
    margin: 0 0 32px 0;
    opacity: 0.8;
  }

  .error-message {
    margin-bottom: 16px;
    padding: 12px 16px;
    background: rgba(239, 68, 68, 0.1);
    border: 1px solid rgba(239, 68, 68, 0.2);
    border-radius: 8px;
    color: #dc2626;
  }

  .login-form {
    display: flex;
    flex-direction: column;
    gap: 16px;
    margin-bottom: 24px;
  }

  .form-group {
    display: flex;
    flex-direction: column;
  }

  :global(.modern-input) {
    padding: 12px 16px;
    border: 2px solid rgba(0, 0, 0, 0.1);
    border-radius: 8px;
    background: rgba(255, 255, 255, 0.9);
    backdrop-filter: blur(10px);
    font-size: 16px;
    transition: all 0.2s ease;
    outline: none;
  }

  :global(.modern-input:focus) {
    border-color: #667eea;
    background: rgba(255, 255, 255, 1);
    box-shadow: 0 0 0 3px rgba(102, 126, 234, 0.1);
  }

  .form-options {
    display: flex;
    align-items: center;
    justify-content: flex-start;
  }

  .checkbox-label {
    display: flex;
    align-items: center;
    gap: 8px;
    color: #6b7280;
    font-size: 14px;
    cursor: pointer;
  }

  .modern-checkbox {
    width: 16px;
    height: 16px;
    border: 2px solid rgba(0, 0, 0, 0.2);
    border-radius: 4px;
    background: rgba(255, 255, 255, 0.9);
    cursor: pointer;
  }

  .modern-checkbox:checked {
    background: #667eea;
    border-color: #667eea;
  }

  .modern-button {
    position: relative;
    display: flex;
    align-items: center;
    justify-content: center;
    padding: 14px 24px;
    border: none;
    border-radius: 8px;
    font-weight: 600;
    font-size: 16px;
    text-align: center;
    transition: all 0.3s cubic-bezier(0.4, 0.0, 0.2, 1);
    overflow: hidden;
    cursor: pointer;
    gap: 8px;
  }

  .modern-button.primary {
    background: linear-gradient(135deg, #667eea 0%, #764ba2 50%, #667eea 100%);
    background-size: 200% 200%;
    color: white;
    border: 2px solid rgba(255, 255, 255, 0.2);
  }

  .modern-button.primary:hover:not(:disabled) {
    background-position: 100% 0%;
    transform: translateY(-2px);
    box-shadow: 0 20px 40px rgba(102, 126, 234, 0.4);
  }

  .modern-button:disabled {
    opacity: 0.7;
    cursor: not-allowed;
    transform: none;
  }

  .text-link {
    display: block;
    text-align: center;
    color: #6b7280;
    font-size: 14px;
    font-weight: 500;
    text-decoration: none;
    opacity: 0.8;
    transition: all 0.2s ease;
    margin-top: 8px;
  }

  .text-link:hover {
    color: #4b5563;
    opacity: 1;
    text-decoration: underline;
    text-decoration-color: rgba(75, 85, 99, 0.3);
    text-underline-offset: 4px;
  }

  .back-link {
    margin-top: 16px;
  }

  /* Dark mode support - CloudBox theme system */
  :global(.cloudbox-dark) .login-card {
    background: rgba(26, 26, 26, 0.8);
    border: 1px solid rgba(255, 255, 255, 0.1);
    backdrop-filter: blur(30px);
    box-shadow: 
      0 30px 60px -15px rgba(0, 0, 0, 0.6),
      0 25px 35px -15px rgba(0, 0, 0, 0.4),
      0 10px 20px -10px rgba(0, 0, 0, 0.3),
      0 0 0 1px rgba(255, 255, 255, 0.05) inset;
  }
  
  :global(.cloudbox-dark) .logo-text-cutout {
    filter: brightness(1.5) saturate(1.3);
  }
  
  :global(.cloudbox-dark) .subtitle {
    color: #cbd5e1;
  }
  
  :global(.cloudbox-dark .modern-input) {
    background: rgba(30, 41, 59, 0.8);
    border-color: rgba(255, 255, 255, 0.1);
    color: #f1f5f9;
    backdrop-filter: blur(10px);
  }
  
  :global(.cloudbox-dark .modern-input:focus) {
    background: rgba(30, 41, 59, 0.9);
    border-color: #667eea;
    box-shadow: 0 0 0 3px rgba(102, 126, 234, 0.15);
  }
  
  :global(.cloudbox-dark) .checkbox-label {
    color: #cbd5e1;
  }
  
  :global(.cloudbox-dark) .modern-checkbox {
    background: rgba(30, 41, 59, 0.8);
    border-color: rgba(255, 255, 255, 0.2);
  }
  
  :global(.cloudbox-dark) .modern-checkbox:checked {
    background: #667eea;
    border-color: #667eea;
  }
  
  :global(.cloudbox-dark) .error-message {
    background: rgba(239, 68, 68, 0.2);
    border-color: rgba(239, 68, 68, 0.4);
    color: #fca5a5;
    backdrop-filter: blur(10px);
  }
  
  :global(.cloudbox-dark) .text-link {
    color: #cbd5e1;
  }
  
  :global(.cloudbox-dark) .text-link:hover {
    color: #e2e8f0;
  }

  /* Mobile responsiveness */
  @media (max-width: 480px) {
    .login-wrapper {
      max-width: 100%;
      padding: 0 16px;
    }
    
    .login-card {
      padding: 32px 24px;
      border-radius: 16px;
      min-height: 450px;
    }
    
    .logo-text-cutout {
      font-size: 40px;
    }
    
    :global(.modern-input) {
      font-size: 16px; /* Prevent zoom on iOS */
    }
  }

  /* Reduce motion for accessibility */
  @media (prefers-reduced-motion: reduce) {
    .animated-gradient,
    .animated-gradient::before,
    .login-card {
      animation: none;
    }
  }
</style>