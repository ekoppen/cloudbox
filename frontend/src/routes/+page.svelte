<script lang="ts">
  import Card from "$lib/components/ui/card.svelte"
  import Button from "$lib/components/ui/button.svelte"
  import Icon from "$lib/components/ui/icon.svelte"
</script>

<div class="min-h-screen bg-background">
  <div class="relative max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
    <!-- Header -->
    <header class="pt-16 pb-24">
      <div class="text-center">
        <h1 class="text-6xl font-bold text-foreground mb-6 flex items-center justify-center gap-4">
          <Icon name="cloud" size={64} />
          CloudBox
        </h1>
        <p class="text-xl text-muted-foreground max-w-2xl mx-auto">
          Self-hosted Backend-as-a-Service solution for modern applications
        </p>
      </div>
    </header>

    <!-- Features Grid -->
    <div class="grid md:grid-cols-2 lg:grid-cols-3 gap-8 mb-24">
      <Card class="p-6 hover:shadow-lg transition-shadow">
        <div class="text-primary mb-4">
          <Icon name="rocket" size={32} />
        </div>
        <h3 class="text-lg font-semibold mb-2 text-card-foreground">Fast Setup</h3>
        <p class="text-muted-foreground">Get your backend running in minutes with Docker</p>
      </Card>
      
      <Card class="p-6 hover:shadow-lg transition-shadow">
        <div class="text-primary mb-4">
          <Icon name="shield" size={32} />
        </div>
        <h3 class="text-lg font-semibold mb-2 text-card-foreground">Security First</h3>
        <p class="text-muted-foreground">JWT authentication, rate limiting, and CORS protection</p>
      </Card>
      
      <Card class="p-6 hover:shadow-lg transition-shadow">
        <div class="text-primary mb-4">
          <Icon name="package" size={32} />
        </div>
        <h3 class="text-lg font-semibold mb-2 text-card-foreground">Project Management</h3>
        <p class="text-muted-foreground">Organize your apps with project-based API namespaces</p>
      </Card>
      
      <Card class="p-6 hover:shadow-lg transition-shadow">
        <div class="text-primary mb-4">
          <Icon name="globe" size={32} />
        </div>
        <h3 class="text-lg font-semibold mb-2 text-card-foreground">Web Hosting</h3>
        <p class="text-muted-foreground">Deploy static web applications with ease</p>
      </Card>
      
      <Card class="p-6 hover:shadow-lg transition-shadow">
        <div class="text-primary mb-4">
          <Icon name="tool" size={32} />
        </div>
        <h3 class="text-lg font-semibold mb-2 text-card-foreground">SDK Support</h3>
        <p class="text-muted-foreground">JavaScript, Python, Go, and TypeScript SDKs</p>
      </Card>
      
      <Card class="p-6 hover:shadow-lg transition-shadow">
        <div class="text-primary mb-4">
          <Icon name="backup" size={32} />
        </div>
        <h3 class="text-lg font-semibold mb-2 text-card-foreground">Backup & Restore</h3>
        <p class="text-muted-foreground">Automated backups with version control</p>
      </Card>
    </div>

    <!-- CTA Section -->
    <div class="text-center">
      <Card class="p-12 max-w-2xl mx-auto">
        <h2 class="text-3xl font-bold text-card-foreground mb-4">
          Ready to get started?
        </h2>
        <p class="text-muted-foreground mb-8 text-lg">
          CloudBox is currently in development. Sign up for early access!
        </p>
        <div class="flex gap-4 justify-center">
          <Button href="/login" size="lg">
            Sign In
          </Button>
          <Button href="/register" variant="outline" size="lg">
            Create Account
          </Button>
        </div>
      </Card>
    </div>

    <!-- Status -->
    <footer class="text-center py-12 text-muted-foreground">
      <p>CloudBox v1.0.0-alpha • Status: MVP Development</p>
    </footer>
  </div>
</div>
