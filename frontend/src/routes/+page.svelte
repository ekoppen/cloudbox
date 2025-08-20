<script lang="ts">
  import Button from "$lib/components/ui/button.svelte"
  import Icon from "$lib/components/ui/icon.svelte"
</script>

<!-- Animated gradient background -->
<div class="landing-container">
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
          Self-hosted Backend-as-a-Service
        </p>
      </div>
      
      <!-- Bottom content -->
      <div class="bottom-content">
        <!-- Login buttons -->
        <div class="button-section">
          <a href="/login" class="modern-button primary">
            <span>Sign In</span>
            <div class="button-glow"></div>
          </a>
          <a href="/register" class="text-link">
            Create Account
          </a>
        </div>
        
        <!-- Status indicator -->
        <div class="status">
          <span class="status-dot"></span>
          Development Preview
        </div>
      </div>
    </div>
  </div>
</div>

<style>
  .landing-container {
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
    min-height: 400px;
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

  .button-section {
    margin-bottom: 24px;
    display: flex;
    flex-direction: column;
    align-items: center;
    gap: 12px;
  }

  .modern-button {
    position: relative;
    display: block;
    padding: 12px 24px;
    border-radius: 8px;
    text-decoration: none;
    font-weight: 600;
    font-size: 14px;
    text-align: center;
    transition: all 0.3s cubic-bezier(0.4, 0.0, 0.2, 1);
    overflow: hidden;
    border: 2px solid transparent;
    max-width: 200px;
    margin: 0 auto;
  }

  .modern-button span {
    position: relative;
    z-index: 2;
  }

  .modern-button.primary {
    background: linear-gradient(135deg, #667eea 0%, #764ba2 50%, #667eea 100%);
    background-size: 200% 200%;
    color: white;
    border: 2px solid rgba(255, 255, 255, 0.2);
  }

  .modern-button.primary:hover {
    background-position: 100% 0%;
    transform: translateY(-2px);
    box-shadow: 0 20px 40px rgba(102, 126, 234, 0.4);
  }

  .button-glow {
    position: absolute;
    top: 0;
    left: -100%;
    width: 100%;
    height: 100%;
    background: linear-gradient(90deg, transparent, rgba(255, 255, 255, 0.4), transparent);
    transition: left 0.6s;
  }

  .modern-button.primary:hover .button-glow {
    left: 100%;
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

  .status {
    display: flex;
    align-items: center;
    justify-content: center;
    gap: 8px;
    color: #6b7280;
    font-size: 14px;
    opacity: 0.7;
  }

  .status-dot {
    width: 8px;
    height: 8px;
    border-radius: 50%;
    background: linear-gradient(135deg, #10b981, #34d399);
    animation: pulse 2s ease-in-out infinite;
  }

  @keyframes pulse {
    0%, 100% { opacity: 1; transform: scale(1); }
    50% { opacity: 0.7; transform: scale(1.2); }
  }

  /* Dark mode support - CloudBox theme system */
  :global(.cloudbox-dark) .login-card {
    background: rgba(15, 23, 42, 0.8);
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
  
  :global(.cloudbox-dark) .status {
    color: #cbd5e1;
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
    }
    
    .logo-text-cutout {
      font-size: 40px;
    }
  }

  /* Reduce motion for accessibility */
  @media (prefers-reduced-motion: reduce) {
    .animated-gradient,
    .animated-gradient::before,
    .login-card,
    .status-dot {
      animation: none;
    }
  }
</style>
