<script lang="ts">
  import { page } from '$app/stores';
  import { onMount } from 'svelte';
  import { auth } from '$lib/stores/auth';
  import Card from '$lib/components/ui/card.svelte';
  import Button from '$lib/components/ui/button.svelte';
  import Input from '$lib/components/ui/input.svelte';
  import Label from '$lib/components/ui/label.svelte';
  import Badge from '$lib/components/ui/badge.svelte';
  import Textarea from '$lib/components/ui/textarea.svelte';
  import Icon from '$lib/components/ui/icon.svelte';

  interface Message {
    id: string;
    subject: string;
    type: 'email' | 'sms' | 'push';
    status: 'sent' | 'pending' | 'failed' | 'draft';
    recipients: number;
    sent_at?: string;
    created_at: string;
    content: string;
  }

  interface MessageTemplate {
    id: string;
    name: string;
    type: 'email' | 'sms' | 'push';
    subject: string;
    content: string;
    variables: string[];
  }

  let messages: Message[] = [
    {
      id: '1',
      subject: 'Welkom bij CloudBox!',
      type: 'email',
      status: 'sent',
      recipients: 127,
      sent_at: '2025-01-19T10:30:00Z',
      created_at: '2025-01-19T10:25:00Z',
      content: 'Welkom bij onze service! We zijn blij dat je er bent.'
    },
    {
      id: '2',
      subject: 'Wachtwoord reset',
      type: 'email',
      status: 'sent',
      recipients: 23,
      sent_at: '2025-01-19T14:20:00Z',
      created_at: '2025-01-19T14:18:00Z',
      content: 'Klik op de link om je wachtwoord te resetten.'
    },
    {
      id: '3',
      subject: 'Nieuwe functie beschikbaar',
      type: 'push',
      status: 'pending',
      recipients: 89,
      created_at: '2025-01-19T15:00:00Z',
      content: 'Ontdek onze nieuwe functie in de app!'
    }
  ];

  let templates: MessageTemplate[] = [
    {
      id: '1',
      name: 'Welkom Email',
      type: 'email',
      subject: 'Welkom bij {{app_name}}!',
      content: 'Hallo {{user_name}},\n\nWelkom bij {{app_name}}! We zijn blij dat je er bent.',
      variables: ['app_name', 'user_name']
    },
    {
      id: '2',
      name: 'Wachtwoord Reset',
      type: 'email',
      subject: 'Reset je wachtwoord',
      content: 'Hallo {{user_name}},\n\nKlik op deze link om je wachtwoord te resetten: {{reset_link}}',
      variables: ['user_name', 'reset_link']
    },
    {
      id: '3',
      name: 'Push Notificatie',
      type: 'push',
      subject: 'Nieuwe update!',
      content: 'Er is een nieuwe update beschikbaar in {{app_name}}',
      variables: ['app_name']
    }
  ];

  let messagingStats = {
    total_sent: 1247,
    total_delivered: 1189,
    total_opened: 567,
    total_clicked: 234,
    bounce_rate: 4.7,
    open_rate: 47.7,
    click_rate: 19.7
  };

  let activeTab = 'messages';
  let showCreateMessage = false;
  let showCreateTemplate = false;
  let newMessage = {
    subject: '',
    type: 'email' as 'email' | 'sms' | 'push',
    content: '',
    recipients: 'all',
    schedule_at: ''
  };
  let newTemplate = {
    name: '',
    type: 'email' as 'email' | 'sms' | 'push',
    subject: '',
    content: ''
  };

  $: projectId = $page.params.id;

  function getStatusColor(status: string): string {
    switch (status) {
      case 'sent': return 'bg-green-100 dark:bg-green-900 text-green-800 dark:text-green-200';
      case 'pending': return 'bg-yellow-100 dark:bg-yellow-900 text-yellow-800 dark:text-yellow-200';
      case 'failed': return 'bg-red-100 dark:bg-red-900 text-red-800 dark:text-red-200';
      case 'draft': return 'bg-muted text-muted-foreground';
      default: return 'bg-muted text-muted-foreground';
    }
  }

  function getTypeIcon(type: string): string {
    switch (type) {
      case 'email': return '✉️';
      case 'sms': return '💬';
      case 'push': return '🔔';
      default: return '📨';
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

  function duplicateMessage(messageId: string) {
    const message = messages.find(m => m.id === messageId);
    if (message) {
      const duplicate: Message = {
        ...message,
        id: Date.now().toString(),
        subject: `${message.subject} (kopie)`,
        status: 'draft',
        sent_at: undefined,
        created_at: new Date().toISOString()
      };
      messages = [duplicate, ...messages];
    }
  }

  function deleteMessage(messageId: string) {
    if (confirm('Weet je zeker dat je dit bericht wilt verwijderen?')) {
      messages = messages.filter(m => m.id !== messageId);
    }
  }

  function deleteTemplate(templateId: string) {
    if (confirm('Weet je zeker dat je deze template wilt verwijderen?')) {
      templates = templates.filter(t => t.id !== templateId);
    }
  }

  async function createMessage() {
    if (!newMessage.subject || !newMessage.content) return;

    const message: Message = {
      id: Date.now().toString(),
      subject: newMessage.subject,
      type: newMessage.type,
      status: newMessage.schedule_at ? 'pending' : 'draft',
      recipients: newMessage.recipients === 'all' ? 127 : 50,
      content: newMessage.content,
      created_at: new Date().toISOString()
    };

    messages = [message, ...messages];
    showCreateMessage = false;
    newMessage = {
      subject: '',
      type: 'email',
      content: '',
      recipients: 'all',
      schedule_at: ''
    };
  }

  async function createTemplate() {
    if (!newTemplate.name || !newTemplate.content) return;

    const template: MessageTemplate = {
      id: Date.now().toString(),
      name: newTemplate.name,
      type: newTemplate.type,
      subject: newTemplate.subject,
      content: newTemplate.content,
      variables: extractVariables(newTemplate.content + ' ' + newTemplate.subject)
    };

    templates = [template, ...templates];
    showCreateTemplate = false;
    newTemplate = {
      name: '',
      type: 'email',
      subject: '',
      content: ''
    };
  }

  function extractVariables(text: string): string[] {
    const matches = text.match(/\{\{([^}]+)\}\}/g);
    if (!matches) return [];
    return [...new Set(matches.map(match => match.slice(2, -2).trim()))];
  }

  function sendMessage(messageId: string) {
    if (confirm('Weet je zeker dat je dit bericht wilt versturen?')) {
      messages = messages.map(m => 
        m.id === messageId ? { 
          ...m, 
          status: 'sent',
          sent_at: new Date().toISOString()
        } : m
      );
    }
  }

  function useTemplate(template: MessageTemplate) {
    newMessage = {
      subject: template.subject,
      type: template.type,
      content: template.content,
      recipients: 'all',
      schedule_at: ''
    };
    showCreateMessage = true;
  }
</script>

<svelte:head>
  <title>Berichten - CloudBox</title>
</svelte:head>

<div class="space-y-6">
  <!-- Header -->
  <div class="flex items-center justify-between">
    <div class="flex items-center space-x-4">
      <div class="w-10 h-10 bg-primary/10 rounded-lg flex items-center justify-center">
        <Icon name="messaging" size={20} className="text-primary" />
      </div>
      <div>
        <h1 class="text-2xl font-bold text-foreground">Berichten</h1>
        <p class="text-sm text-muted-foreground">
          Verstuur emails, SMS en push notificaties naar gebruikers
        </p>
      </div>
    </div>
    <div class="flex space-x-3">
      <Button
        variant="outline"
        on:click={() => showCreateTemplate = true}
        class="flex items-center space-x-2"
      >
        <Icon name="backup" size={16} />
        <span>Nieuwe Template</span>
      </Button>
      <Button
        on:click={() => showCreateMessage = true}
        class="flex items-center space-x-2"
      >
        <Icon name="messaging" size={16} />
        <span>Nieuw Bericht</span>
      </Button>
    </div>
  </div>

  <!-- Messaging Stats -->
  <div class="grid grid-cols-1 md:grid-cols-4 gap-4">
    <Card class="p-6">
      <div class="flex items-center justify-between">
        <div>
          <p class="text-sm font-medium text-muted-foreground">Verzonden</p>
          <p class="text-2xl font-bold text-foreground">{messagingStats.total_sent.toLocaleString()}</p>
        </div>
        <div class="w-10 h-10 bg-blue-100 dark:bg-blue-900 rounded-lg flex items-center justify-center">
          <Icon name="messaging" size={20} className="text-blue-600 dark:text-blue-400" />
        </div>
      </div>
    </Card>

    <Card class="p-6">
      <div class="flex items-center justify-between">
        <div>
          <p class="text-sm font-medium text-muted-foreground">Bezorgd</p>
          <p class="text-2xl font-bold text-foreground">{messagingStats.total_delivered.toLocaleString()}</p>
          <p class="text-xs text-muted-foreground">{((messagingStats.total_delivered / messagingStats.total_sent) * 100).toFixed(1)}%</p>
        </div>
        <div class="w-10 h-10 bg-green-100 dark:bg-green-900 rounded-lg flex items-center justify-center">
          <Icon name="auth" size={20} className="text-green-600 dark:text-green-400" />
        </div>
      </div>
    </Card>

    <Card class="p-6">
      <div class="flex items-center justify-between">
        <div>
          <p class="text-sm font-medium text-muted-foreground">Geopend</p>
          <p class="text-2xl font-bold text-foreground">{messagingStats.total_opened}</p>
          <p class="text-xs text-muted-foreground">{messagingStats.open_rate}%</p>
        </div>
        <div class="w-10 h-10 bg-yellow-100 dark:bg-yellow-900 rounded-lg flex items-center justify-center">
          <Icon name="backup" size={20} className="text-yellow-600 dark:text-yellow-400" />
        </div>
      </div>
    </Card>

    <Card class="p-6">
      <div class="flex items-center justify-between">
        <div>
          <p class="text-sm font-medium text-muted-foreground">Geklikt</p>
          <p class="text-2xl font-bold text-foreground">{messagingStats.total_clicked}</p>
          <p class="text-xs text-muted-foreground">{messagingStats.click_rate}%</p>
        </div>
        <div class="w-10 h-10 bg-purple-100 dark:bg-purple-900 rounded-lg flex items-center justify-center">
          <Icon name="functions" size={20} className="text-purple-600 dark:text-purple-400" />
        </div>
      </div>
    </Card>
  </div>

  <!-- Tabs -->
  <div class="border-b border-border">
    <nav class="flex space-x-8">
      <button
        on:click={() => activeTab = 'messages'}
        class="flex items-center space-x-2 py-2 px-1 border-b-2 text-sm font-medium transition-colors {
          activeTab === 'messages' 
            ? 'border-primary text-primary' 
            : 'border-transparent text-muted-foreground hover:text-foreground hover:border-border'
        }"
      >
        <Icon name="messaging" size={16} />
        <span>Berichten ({messages.length})</span>
      </button>
      <button
        on:click={() => activeTab = 'templates'}
        class="flex items-center space-x-2 py-2 px-1 border-b-2 text-sm font-medium transition-colors {
          activeTab === 'templates' 
            ? 'border-primary text-primary' 
            : 'border-transparent text-muted-foreground hover:text-foreground hover:border-border'
        }"
      >
        <Icon name="backup" size={16} />
        <span>Templates ({templates.length})</span>
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

  <!-- Messages Tab -->
  {#if activeTab === 'messages'}
    <div class="space-y-4">
      {#each messages as message}
        <Card class="p-6">
          <div class="flex items-center justify-between">
            <div class="flex-1">
              <div class="flex items-center space-x-3">
                <div class="w-8 h-8 bg-primary/10 rounded-lg flex items-center justify-center">
                  <Icon name="messaging" size={16} className="text-primary" />
                </div>
                <div>
                  <h3 class="text-lg font-medium text-foreground">{message.subject}</h3>
                  <div class="flex items-center space-x-4 text-sm text-muted-foreground">
                    <Badge class={getStatusColor(message.status)}>
                      {message.status}
                    </Badge>
                    <span>{message.recipients} ontvangers</span>
                    <span>Aangemaakt: {formatDate(message.created_at)}</span>
                    {#if message.sent_at}
                      <span>Verzonden: {formatDate(message.sent_at)}</span>
                    {/if}
                  </div>
                </div>
              </div>
              
              <p class="mt-2 text-sm text-muted-foreground line-clamp-2">{message.content}</p>
            </div>

            <div class="flex items-center space-x-3">
              {#if message.status === 'draft' || message.status === 'pending'}
                <Button
                  size="sm"
                  on:click={() => sendMessage(message.id)}
                  class="flex items-center space-x-1"
                >
                  <Icon name="messaging" size={14} />
                  <span>Versturen</span>
                </Button>
              {/if}
              <Button
                variant="ghost"
                size="sm"
                on:click={() => duplicateMessage(message.id)}
              >
                Kopiëren
              </Button>
              <Button
                variant="ghost"
                size="sm"
              >
                Statistieken
              </Button>
              <Button
                variant="ghost"
                size="sm"
                class="text-destructive hover:text-destructive"
                on:click={() => deleteMessage(message.id)}
              >
                Verwijderen
              </Button>
            </div>
          </div>
        </Card>
      {/each}

      {#if messages.length === 0}
        <Card class="p-12 text-center">
          <div class="w-16 h-16 bg-muted rounded-lg flex items-center justify-center mx-auto mb-4">
            <Icon name="messaging" size={32} className="text-muted-foreground" />
          </div>
          <h3 class="text-lg font-medium text-foreground mb-2">Geen berichten</h3>
          <p class="text-muted-foreground mb-4">Verstuur je eerste bericht naar gebruikers</p>
          <Button
            on:click={() => showCreateMessage = true}
            class="flex items-center space-x-2"
          >
            <Icon name="messaging" size={16} />
            <span>Nieuw Bericht</span>
          </Button>
        </Card>
      {/if}
    </div>
  {/if}

  <!-- Templates Tab -->
  {#if activeTab === 'templates'}
    <div class="space-y-4">
      <div class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
        {#each templates as template}
          <Card class="p-6">
            <div class="flex items-start justify-between">
              <div class="flex-1">
                <div class="flex items-center space-x-2 mb-2">
                  <div class="w-6 h-6 bg-primary/10 rounded flex items-center justify-center">
                    <Icon name="messaging" size={12} className="text-primary" />
                  </div>
                  <h3 class="text-lg font-medium text-foreground">{template.name}</h3>
                </div>
                
                <p class="text-sm text-muted-foreground mb-3">{template.subject}</p>
                
                <div class="text-xs text-muted-foreground mb-3">
                  <p class="line-clamp-3">{template.content}</p>
                </div>

                {#if template.variables.length > 0}
                  <div class="mb-3">
                    <p class="text-xs text-muted-foreground mb-1">Variabelen:</p>
                    <div class="flex flex-wrap gap-1">
                      {#each template.variables as variable}
                        <Badge variant="outline" class="text-xs">
                          {variable}
                        </Badge>
                      {/each}
                    </div>
                  </div>
                {/if}
              </div>
            </div>

            <div class="flex space-x-2">
              <Button
                size="sm"
                on:click={() => useTemplate(template)}
                class="flex-1"
              >
                Gebruiken
              </Button>
              <Button
                variant="ghost"
                size="sm"
                class="w-8 h-8 p-0"
              >
                <Icon name="settings" size={14} />
              </Button>
              <Button
                variant="ghost"
                size="sm"
                class="w-8 h-8 p-0 text-destructive hover:text-destructive"
                on:click={() => deleteTemplate(template.id)}
              >
                <Icon name="backup" size={14} />
              </Button>
            </div>
          </Card>
        {/each}
      </div>

      {#if templates.length === 0}
        <Card class="p-12 text-center">
          <div class="w-16 h-16 bg-muted rounded-lg flex items-center justify-center mx-auto mb-4">
            <Icon name="backup" size={32} className="text-muted-foreground" />
          </div>
          <h3 class="text-lg font-medium text-foreground mb-2">Geen templates</h3>
          <p class="text-muted-foreground mb-4">Maak herbruikbare templates voor berichten</p>
          <Button
            on:click={() => showCreateTemplate = true}
            class="flex items-center space-x-2"
          >
            <Icon name="backup" size={16} />
            <span>Nieuwe Template</span>
          </Button>
        </Card>
      {/if}
    </div>
  {/if}

  <!-- Settings Tab -->
  {#if activeTab === 'settings'}
    <div class="space-y-6">
      <div>
        <h2 class="text-lg font-medium text-foreground">Berichten Instellingen</h2>
        <p class="text-sm text-muted-foreground">Configureer email, SMS en push instellingen</p>
      </div>

      <!-- Email Settings -->
      <Card class="p-6">
        <div class="flex items-center space-x-2 mb-4">
          <Icon name="messaging" size={20} className="text-primary" />
          <h3 class="text-lg font-medium text-foreground">Email Instellingen</h3>
        </div>
        <div class="space-y-4">
          <div class="grid grid-cols-1 md:grid-cols-2 gap-4">
            <div>
              <Label>Afzender naam</Label>
              <Input type="text" value="CloudBox" class="mt-1" />
            </div>
            <div>
              <Label>Afzender email</Label>
              <Input type="email" value="noreply@cloudbox.nl" class="mt-1" />
            </div>
          </div>
          <div class="grid grid-cols-1 md:grid-cols-2 gap-4">
            <div>
              <Label>SMTP Server</Label>
              <Input type="text" value="smtp.mailgun.org" class="mt-1" />
            </div>
            <div>
              <Label>SMTP Poort</Label>
              <Input type="number" value="587" class="mt-1" />
            </div>
          </div>
        </div>
      </Card>

      <!-- SMS Settings -->
      <Card class="p-6">
        <div class="flex items-center space-x-2 mb-4">
          <Icon name="cloud" size={20} className="text-primary" />
          <h3 class="text-lg font-medium text-foreground">SMS Instellingen</h3>
        </div>
        <div class="space-y-4">
          <div class="grid grid-cols-1 md:grid-cols-2 gap-4">
            <div>
              <Label>SMS Provider</Label>
              <select class="w-full p-2 border border-border rounded-md bg-background text-foreground mt-1 focus:ring-2 focus:ring-primary">
                <option>Twilio</option>
                <option>MessageBird</option>
                <option>CM.com</option>
              </select>
            </div>
            <div>
              <Label>Afzender ID</Label>
              <Input type="text" value="CloudBox" class="mt-1" />
            </div>
          </div>
        </div>
      </Card>

      <!-- Push Settings -->
      <Card class="p-6">
        <div class="flex items-center space-x-2 mb-4">
          <Icon name="functions" size={20} className="text-primary" />
          <h3 class="text-lg font-medium text-foreground">Push Notificatie Instellingen</h3>
        </div>
        <div class="space-y-4">
          <div>
            <label class="flex items-center">
              <input type="checkbox" checked class="rounded border-border text-primary focus:ring-primary" />
              <span class="ml-2 text-sm text-foreground">Push notificaties inschakelen</span>
            </label>
          </div>
          <div class="grid grid-cols-1 md:grid-cols-2 gap-4">
            <div>
              <Label>FCM Server Key</Label>
              <Input type="password" value="••••••••••••••••" class="mt-1" />
            </div>
            <div>
              <Label>APNs Certificate</Label>
              <input type="file" class="w-full p-2 border border-border rounded-md bg-background text-foreground mt-1 focus:ring-2 focus:ring-primary file:mr-4 file:py-2 file:px-4 file:rounded-full file:border-0 file:text-sm file:font-semibold file:bg-primary/10 file:text-primary hover:file:bg-primary/20" />
            </div>
          </div>
        </div>
      </Card>

      <div class="flex justify-end">
        <Button>
          Instellingen Opslaan
        </Button>
      </div>
    </div>
  {/if}
</div>

<!-- Create Message Modal -->
{#if showCreateMessage}
  <div class="fixed inset-0 bg-black/80 backdrop-blur-sm flex items-center justify-center p-4 z-50">
    <Card class="max-w-2xl w-full p-6 max-h-screen overflow-y-auto border-2 shadow-2xl">
      <div class="flex items-center space-x-3 mb-4">
        <div class="w-10 h-10 bg-primary/10 rounded-lg flex items-center justify-center">
          <Icon name="messaging" size={20} className="text-primary" />
        </div>
        <h2 class="text-xl font-bold text-foreground">Nieuw Bericht</h2>
      </div>
      
      <form on:submit|preventDefault={createMessage} class="space-y-4">
        <div class="grid grid-cols-1 md:grid-cols-2 gap-4">
          <div>
            <Label for="message-type">Type</Label>
            <select
              id="message-type"
              bind:value={newMessage.type}
              class="w-full p-2 border border-border rounded-md bg-background text-foreground mt-1 focus:ring-2 focus:ring-primary"
            >
              <option value="email">✉️ Email</option>
              <option value="sms">💬 SMS</option>
              <option value="push">🔔 Push Notificatie</option>
            </select>
          </div>

          <div>
            <Label for="message-recipients">Ontvangers</Label>
            <select
              id="message-recipients"
              bind:value={newMessage.recipients}
              class="w-full p-2 border border-border rounded-md bg-background text-foreground mt-1 focus:ring-2 focus:ring-primary"
            >
              <option value="all">Alle gebruikers</option>
              <option value="active">Alleen actieve gebruikers</option>
              <option value="custom">Custom selectie</option>
            </select>
          </div>
        </div>

        <div>
          <Label for="message-subject">Onderwerp</Label>
          <Input
            id="message-subject"
            type="text"
            bind:value={newMessage.subject}
            required
            class="mt-1"
            placeholder="Onderwerp van je bericht"
          />
        </div>

        <div>
          <Label for="message-content">Bericht</Label>
          <Textarea
            id="message-content"
            bind:value={newMessage.content}
            required
            class="mt-1"
            rows={6}
            placeholder="Schrijf je bericht hier..."
          />
        </div>

        <div>
          <Label for="message-schedule">Inplannen (optioneel)</Label>
          <Input
            id="message-schedule"
            type="datetime-local"
            bind:value={newMessage.schedule_at}
            class="mt-1"
          />
        </div>
        
        <div class="flex space-x-3 pt-4">
          <Button
            type="button"
            variant="outline"
            on:click={() => { showCreateMessage = false; }}
            class="flex-1"
          >
            Annuleren
          </Button>
          <Button
            type="submit"
            disabled={!newMessage.subject || !newMessage.content}
            class="flex-1"
          >
            {newMessage.schedule_at ? 'Inplannen' : 'Als Concept Opslaan'}
          </Button>
        </div>
      </form>
    </Card>
  </div>
{/if}

<!-- Create Template Modal -->
{#if showCreateTemplate}
  <div class="fixed inset-0 bg-black/80 backdrop-blur-sm flex items-center justify-center p-4 z-50">
    <Card class="max-w-2xl w-full p-6 max-h-screen overflow-y-auto border-2 shadow-2xl">
      <div class="flex items-center space-x-3 mb-4">
        <div class="w-10 h-10 bg-primary/10 rounded-lg flex items-center justify-center">
          <Icon name="backup" size={20} className="text-primary" />
        </div>
        <h2 class="text-xl font-bold text-foreground">Nieuwe Template</h2>
      </div>
      
      <form on:submit|preventDefault={createTemplate} class="space-y-4">
        <div class="grid grid-cols-1 md:grid-cols-2 gap-4">
          <div>
            <Label for="template-name">Template naam</Label>
            <Input
              id="template-name"
              type="text"
              bind:value={newTemplate.name}
              required
              class="mt-1"
              placeholder="bijv. Welkom Email"
            />
          </div>

          <div>
            <Label for="template-type">Type</Label>
            <select
              id="template-type"
              bind:value={newTemplate.type}
              class="w-full p-2 border border-border rounded-md bg-background text-foreground mt-1 focus:ring-2 focus:ring-primary"
            >
              <option value="email">✉️ Email</option>
              <option value="sms">💬 SMS</option>
              <option value="push">🔔 Push Notificatie</option>
            </select>
          </div>
        </div>

        <div>
          <Label for="template-subject">Onderwerp</Label>
          <Input
            id="template-subject"
            type="text"
            bind:value={newTemplate.subject}
            class="mt-1"
            placeholder="Gebruik {{variabele}} voor dynamische content"
          />
        </div>

        <div>
          <Label for="template-content">Template inhoud</Label>
          <Textarea
            id="template-content"
            bind:value={newTemplate.content}
            required
            class="mt-1"
            rows={8}
            placeholder="Gebruik {{variabele}} voor dynamische content, bijv: Hallo {{user_name}}, welkom bij {{app_name}}!"
          />
          <p class="mt-1 text-xs text-muted-foreground">
            Gebruik &#123;&#123;variabele_naam&#125;&#125; voor dynamische waarden
          </p>
        </div>
        
        <div class="flex space-x-3 pt-4">
          <Button
            type="button"
            variant="outline"
            on:click={() => { showCreateTemplate = false; }}
            class="flex-1"
          >
            Annuleren
          </Button>
          <Button
            type="submit"
            disabled={!newTemplate.name || !newTemplate.content}
            class="flex-1"
          >
            Template Aanmaken
          </Button>
        </div>
      </form>
    </Card>
  </div>
{/if}