<script lang="ts">
  import { page } from '$app/stores';
  import { onMount } from 'svelte';
  import { auth } from '$lib/stores/auth';
  import { toast } from '$lib/stores/toast';
  import { API_BASE_URL } from '$lib/config';
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

  let messages: Message[] = [];
  let templates: MessageTemplate[] = [];
  let messagingStats = {
    total_sent: 0,
    total_delivered: 0,
    total_opened: 0,
    total_clicked: 0,
    bounce_rate: 0,
    open_rate: 0,
    click_rate: 0
  };
  let loading = true;
  let backendAvailable = true;

  let activeTab = 'messages';
  let showCreateMessage = false;
  let showCreateTemplate = false;
  let systemNotifications = [];
  let showNotifications = true;
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

  onMount(() => {
    loadMessages();
    loadTemplates();
    loadMessagingStats();
    loadSystemNotifications();
  });

  async function loadMessages() {
    try {
      const response = await fetch(`${API_BASE_URL}/p/${projectId}/api/messaging/messages`, {
        headers: {
          'Authorization': `Bearer ${$auth.token}`,
          'Content-Type': 'application/json',
        },
      });

      if (response.ok) {
        messages = await response.json();
      } else {
        console.error('Failed to load messages:', response.status);
        if (response.status === 404) {
          backendAvailable = false;
        } else {
          toast.error('Fout bij het laden van berichten');
        }
      }
    } catch (error) {
      console.error('Error loading messages:', error);
      if (error.message.includes('Failed to fetch')) {
        backendAvailable = false;
      } else {
        toast.error('Netwerkfout bij het laden van berichten');
      }
    }
  }

  async function loadTemplates() {
    try {
      const response = await fetch(`${API_BASE_URL}/p/${projectId}/api/messaging/templates`, {
        headers: {
          'Authorization': `Bearer ${$auth.token}`,
          'Content-Type': 'application/json',
        },
      });

      if (response.ok) {
        templates = await response.json();
      } else {
        console.error('Failed to load templates:', response.status);
        if (response.status === 404) {
          backendAvailable = false;
        }
      }
    } catch (error) {
      console.error('Error loading templates:', error);
      if (error.message.includes('Failed to fetch')) {
        backendAvailable = false;
      }
    }
  }

  async function loadMessagingStats() {
    try {
      const response = await fetch(`${API_BASE_URL}/p/${projectId}/api/messaging/stats`, {
        headers: {
          'Authorization': `Bearer ${$auth.token}`,
          'Content-Type': 'application/json',
        },
      });

      if (response.ok) {
        const data = await response.json();
        messagingStats = {
          total_sent: data.total_sent || 0,
          total_delivered: data.total_delivered || 0,
          total_opened: data.total_opened || 0,
          total_clicked: data.total_clicked || 0,
          bounce_rate: data.bounce_rate || 0,
          open_rate: data.open_rate || 0,
          click_rate: data.click_rate || 0,
        };
      } else {
        console.error('Failed to load messaging stats:', response.status);
        if (response.status === 404) {
          backendAvailable = false;
        }
      }
    } catch (error) {
      console.error('Error loading messaging stats:', error);
      if (error.message.includes('Failed to fetch')) {
        backendAvailable = false;
      }
    } finally {
      loading = false;
    }
  }

  async function loadSystemNotifications() {
    try {
      // First load channels to find the system notifications channel
      const channelsResponse = await fetch(`${API_BASE_URL}/p/${projectId}/api/messaging/channels`, {
        headers: {
          'Authorization': `Bearer ${$auth.token}`,
          'Content-Type': 'application/json',
        },
      });

      if (channelsResponse.ok) {
        const channels = await channelsResponse.json();
        const systemChannel = channels.find(channel => 
          channel.type === 'system' && channel.name === 'System Notifications'
        );

        if (systemChannel) {
          // Load messages from the system channel
          const messagesResponse = await fetch(`${API_BASE_URL}/p/${projectId}/api/messaging/channels/${systemChannel.id}/messages`, {
            headers: {
              'Authorization': `Bearer ${$auth.token}`,
              'Content-Type': 'application/json',
            },
          });

          if (messagesResponse.ok) {
            const messages = await messagesResponse.json();
            // Filter for GitHub-related notifications and sort by newest first
            systemNotifications = messages
              .filter(msg => msg.metadata?.type === 'github_update' || msg.metadata?.type === 'deployment_started')
              .sort((a, b) => new Date(b.created_at) - new Date(a.created_at))
              .slice(0, 10); // Show last 10 notifications
          }
        }
      } else {
        console.error('Failed to load system notifications:', channelsResponse.status);
      }
    } catch (error) {
      console.error('Error loading system notifications:', error);
    }
  }

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

  async function duplicateMessage(messageId: string) {
    const message = messages.find(m => m.id === messageId);
    if (!message) return;

    try {
      const response = await fetch(`${API_BASE_URL}/p/${projectId}/api/messaging/messages`, {
        method: 'POST',
        headers: {
          'Authorization': `Bearer ${$auth.token}`,
          'Content-Type': 'application/json',
        },
        body: JSON.stringify({
          subject: `${message.subject} (kopie)`,
          type: message.type,
          content: message.content,
          status: 'draft'
        }),
      });

      if (response.ok) {
        const duplicate = await response.json();
        messages = [duplicate, ...messages];
        toast.success('Bericht gekopieerd');
      } else {
        toast.error('Fout bij het kopiëren van bericht');
      }
    } catch (error) {
      console.error('Error duplicating message:', error);
      toast.error('Netwerkfout bij het kopiëren van bericht');
    }
  }

  async function deleteMessage(messageId: string) {
    if (!confirm('Weet je zeker dat je dit bericht wilt verwijderen?')) return;

    try {
      const response = await fetch(`${API_BASE_URL}/p/${projectId}/api/messaging/messages/${messageId}`, {
        method: 'DELETE',
        headers: {
          'Authorization': `Bearer ${$auth.token}`,
          'Content-Type': 'application/json',
        },
      });

      if (response.ok) {
        messages = messages.filter(m => m.id !== messageId);
        toast.success('Bericht verwijderd');
      } else {
        toast.error('Fout bij het verwijderen van bericht');
      }
    } catch (error) {
      console.error('Error deleting message:', error);
      toast.error('Netwerkfout bij het verwijderen van bericht');
    }
  }

  async function deleteTemplate(templateId: string) {
    if (!confirm('Weet je zeker dat je deze template wilt verwijderen?')) return;

    try {
      const response = await fetch(`${API_BASE_URL}/p/${projectId}/api/messaging/templates/${templateId}`, {
        method: 'DELETE',
        headers: {
          'Authorization': `Bearer ${$auth.token}`,
          'Content-Type': 'application/json',
        },
      });

      if (response.ok) {
        templates = templates.filter(t => t.id !== templateId);
        toast.success('Template verwijderd');
      } else {
        toast.error('Fout bij het verwijderen van template');
      }
    } catch (error) {
      console.error('Error deleting template:', error);
      toast.error('Netwerkfout bij het verwijderen van template');
    }
  }

  async function createMessage() {
    if (!newMessage.subject || !newMessage.content) return;

    try {
      const response = await fetch(`${API_BASE_URL}/p/${projectId}/api/messaging/messages`, {
        method: 'POST',
        headers: {
          'Authorization': `Bearer ${$auth.token}`,
          'Content-Type': 'application/json',
        },
        body: JSON.stringify({
          subject: newMessage.subject,
          type: newMessage.type,
          content: newMessage.content,
          recipients: newMessage.recipients,
          schedule_at: newMessage.schedule_at || undefined,
          status: newMessage.schedule_at ? 'pending' : 'draft'
        }),
      });

      if (response.ok) {
        const message = await response.json();
        messages = [message, ...messages];
        showCreateMessage = false;
        newMessage = {
          subject: '',
          type: 'email',
          content: '',
          recipients: 'all',
          schedule_at: ''
        };
        toast.success('Bericht aangemaakt');
      } else {
        const error = await response.json();
        toast.error(error.message || 'Fout bij het aanmaken van bericht');
      }
    } catch (error) {
      console.error('Error creating message:', error);
      toast.error('Netwerkfout bij het aanmaken van bericht');
    }
  }

  async function createTemplate() {
    if (!newTemplate.name || !newTemplate.content) return;

    try {
      const response = await fetch(`${API_BASE_URL}/p/${projectId}/api/messaging/templates`, {
        method: 'POST',
        headers: {
          'Authorization': `Bearer ${$auth.token}`,
          'Content-Type': 'application/json',
        },
        body: JSON.stringify({
          name: newTemplate.name,
          type: newTemplate.type,
          subject: newTemplate.subject,
          content: newTemplate.content,
          variables: extractVariables(newTemplate.content + ' ' + newTemplate.subject)
        }),
      });

      if (response.ok) {
        const template = await response.json();
        templates = [template, ...templates];
        showCreateTemplate = false;
        newTemplate = {
          name: '',
          type: 'email',
          subject: '',
          content: ''
        };
        toast.success('Template aangemaakt');
      } else {
        const error = await response.json();
        toast.error(error.message || 'Fout bij het aanmaken van template');
      }
    } catch (error) {
      console.error('Error creating template:', error);
      toast.error('Netwerkfout bij het aanmaken van template');
    }
  }

  function extractVariables(text: string): string[] {
    const matches = text.match(/\{\{([^}]+)\}\}/g);
    if (!matches) return [];
    return [...new Set(matches.map(match => match.slice(2, -2).trim()))];
  }

  async function sendMessage(messageId: string) {
    if (!confirm('Weet je zeker dat je dit bericht wilt versturen?')) return;

    try {
      const response = await fetch(`${API_BASE_URL}/p/${projectId}/api/messaging/messages/${messageId}/send`, {
        method: 'POST',
        headers: {
          'Authorization': `Bearer ${$auth.token}`,
          'Content-Type': 'application/json',
        },
      });

      if (response.ok) {
        messages = messages.map(m => 
          m.id === messageId ? { 
            ...m, 
            status: 'sent',
            sent_at: new Date().toISOString()
          } : m
        );
        toast.success('Bericht verstuurd');
      } else {
        toast.error('Fout bij het versturen van bericht');
      }
    } catch (error) {
      console.error('Error sending message:', error);
      toast.error('Netwerkfout bij het versturen van bericht');
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

  async function deployFromNotification(notification) {
    if (!notification.metadata?.repository_id) return;
    
    const repositoryId = notification.metadata.repository_id;
    const repositoryName = notification.metadata.repository_name;
    
    if (!confirm(`Weet je zeker dat je de pending update voor ${repositoryName} wilt deployen?`)) return;

    try {
      const response = await fetch(`${API_BASE_URL}/api/v1/projects/${projectId}/github-repositories/${repositoryId}/deploy-pending`, {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
          ...auth.getAuthHeader()
        }
      });

      if (response.ok) {
        const result = await response.json();
        toast.success(`Deployment gestart voor ${result.deployments_started} environment(s)`);
        // Reload notifications to update the UI
        await loadSystemNotifications();
      } else {
        const error = await response.json();
        toast.error(error.error || 'Fout bij starten deployment');
      }
    } catch (error) {
      console.error('Error deploying from notification:', error);
      toast.error('Netwerkfout bij deployment');
    }
  }

  function formatNotificationTime(dateStr: string): string {
    const date = new Date(dateStr);
    const now = new Date();
    const diffInMinutes = Math.floor((now - date) / (1000 * 60));
    
    if (diffInMinutes < 1) return 'Zojuist';
    if (diffInMinutes < 60) return `${diffInMinutes}m geleden`;
    if (diffInMinutes < 1440) return `${Math.floor(diffInMinutes / 60)}u geleden`;
    return `${Math.floor(diffInMinutes / 1440)}d geleden`;
  }
</script>

<svelte:head>
  <title>Berichten - CloudBox</title>
</svelte:head>

<div class="space-y-6">
  <!-- Backend Status Notice -->
  {#if !backendAvailable}
    <Card class="bg-yellow-50 dark:bg-yellow-900/20 border-yellow-200 dark:border-yellow-800 p-4">
      <div class="flex items-center space-x-3">
        <Icon name="warning" size={20} className="text-yellow-600 dark:text-yellow-400" />
        <div>
          <h3 class="text-sm font-medium text-yellow-800 dark:text-yellow-200">Backend Server Niet Beschikbaar</h3>
          <p class="text-xs text-yellow-700 dark:text-yellow-300 mt-1">
            De backend server draait niet of de API endpoints zijn nog niet geïmplementeerd. Start de backend server om alle functionaliteit te gebruiken.
          </p>
        </div>
      </div>
    </Card>
  {/if}

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
        on:click={() => activeTab = 'notifications'}
        class="flex items-center space-x-2 py-2 px-1 border-b-2 text-sm font-medium transition-colors {
          activeTab === 'notifications' 
            ? 'border-primary text-primary' 
            : 'border-transparent text-muted-foreground hover:text-foreground hover:border-border'
        }"
      >
        <Icon name="github" size={16} />
        <span>GitHub Meldingen ({systemNotifications.length})</span>
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

  <!-- GitHub Notifications Tab -->
  {#if activeTab === 'notifications'}
    <div class="space-y-4">
      {#if systemNotifications.length > 0}
        {#each systemNotifications as notification}
          <Card class="p-6">
            <div class="flex items-start justify-between">
              <div class="flex-1">
                <div class="flex items-center space-x-3">
                  <div class="w-8 h-8 bg-orange-100 dark:bg-orange-900 rounded-lg flex items-center justify-center">
                    <Icon name="github" size={16} className="text-orange-600 dark:text-orange-400" />
                  </div>
                  <div class="flex-1">
                    <div class="flex items-center space-x-2">
                      <h3 class="text-lg font-medium text-foreground">
                        {notification.metadata?.repository_name || 'Repository'}
                      </h3>
                      {#if notification.metadata?.type === 'github_update'}
                        <Badge class="bg-orange-100 dark:bg-orange-900 text-orange-800 dark:text-orange-200">
                          Update Beschikbaar
                        </Badge>
                      {:else if notification.metadata?.type === 'deployment_started'}
                        <Badge class="bg-blue-100 dark:bg-blue-900 text-blue-800 dark:text-blue-200">
                          Deployment Gestart
                        </Badge>
                      {/if}
                    </div>
                    <div class="flex items-center space-x-4 text-sm text-muted-foreground mt-1">
                      <span>Commit: {notification.metadata?.commit_hash?.substring(0, 8) || 'Unknown'}</span>
                      <span>{formatNotificationTime(notification.created_at)}</span>
                    </div>
                  </div>
                </div>
                
                <div class="mt-3 text-sm text-foreground">
                  {@html notification.content.replace(/\*\*(.*?)\*\*/g, '<strong>$1</strong>').replace(/`(.*?)`/g, '<code class="bg-gray-100 dark:bg-gray-800 px-1 py-0.5 rounded text-xs">$1</code>').replace(/\n/g, '<br>')}
                </div>
              </div>

              <div class="flex items-center space-x-3 ml-4">
                {#if notification.metadata?.type === 'github_update' && notification.metadata?.can_deploy}
                  <Button
                    size="sm"
                    on:click={() => deployFromNotification(notification)}
                    class="bg-orange-600 text-white hover:bg-orange-700"
                  >
                    <Icon name="rocket" size={14} className="mr-1" />
                    Deploy
                  </Button>
                {/if}
                
                {#if notification.metadata?.github_url}
                  <Button
                    variant="outline"
                    size="sm"
                    on:click={() => window.open(notification.metadata.github_url, '_blank')}
                  >
                    <Icon name="github" size={14} className="mr-1" />
                    Bekijk op GitHub
                  </Button>
                {/if}
              </div>
            </div>
          </Card>
        {/each}
      {:else}
        <Card class="p-12 text-center">
          <div class="w-16 h-16 bg-muted rounded-lg flex items-center justify-center mx-auto mb-4">
            <Icon name="github" size={32} className="text-muted-foreground" />
          </div>
          <h3 class="text-lg font-medium text-foreground mb-2">Geen GitHub meldingen</h3>
          <p class="text-muted-foreground mb-4">Hier verschijnen meldingen over repository updates en deployments</p>
          <div class="text-sm text-muted-foreground">
            <p>Meldingen worden automatisch aangemaakt wanneer:</p>
            <ul class="mt-2 space-y-1">
              <li>• Er nieuwe commits zijn gepusht naar je repository</li>
              <li>• Een deployment wordt gestart</li>
              <li>• Er problemen zijn met deployments</li>
            </ul>
          </div>
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