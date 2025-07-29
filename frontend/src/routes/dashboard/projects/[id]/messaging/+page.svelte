<script lang="ts">
  import { page } from '$app/stores';
  import { onMount, onDestroy } from 'svelte';
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


  let messages: Message[] = [];
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
  let systemNotifications: any[] = [];
  let showNotifications = true;
  let expandedNotifications = new Set(); // Track which notifications are expanded
  let refreshInterval: number;
  let lastViewedTime = Date.now(); // Track when user last viewed notifications
  let unreadCount = 0;
  let newMessage = {
    subject: '',
    type: 'email' as 'email' | 'sms' | 'push',
    content: '',
    recipients: 'all',
    schedule_at: ''
  };

  $: projectId = $page.params.id;

  onMount(() => {
    loadMessages();
    loadMessagingStats();
    loadSystemNotifications();
    
    // Set up auto-refresh for real-time updates every 30 seconds
    refreshInterval = setInterval(() => {
      loadSystemNotifications();
      loadMessages();
    }, 30000);
  });

  onDestroy(() => {
    if (refreshInterval) {
      clearInterval(refreshInterval);
    }
  });

  async function loadMessages() {
    try {
      const response = await fetch(`${API_BASE_URL}/api/v1/projects/${projectId}/messaging/messages`, {
        headers: {
          'Authorization': `Bearer ${$auth.token}`,
          'Content-Type': 'application/json',
        },
      });

      if (response.ok) {
        const data = await response.json();
        messages = Array.isArray(data) ? data : [];
      } else {
        console.error('Failed to load messages:', response.status);
        messages = [];
        if (response.status === 404) {
          backendAvailable = false;
        } else {
          toast.error('Fout bij het laden van berichten');
        }
      }
    } catch (error) {
      console.error('Error loading messages:', error);
      messages = [];
      if (error.message.includes('Failed to fetch')) {
        backendAvailable = false;
      } else {
        toast.error('Netwerkfout bij het laden van berichten');
      }
    }
  }


  async function loadMessagingStats() {
    try {
      const response = await fetch(`${API_BASE_URL}/api/v1/projects/${projectId}/messaging/stats`, {
        headers: {
          'Authorization': `Bearer ${$auth.token}`,
          'Content-Type': 'application/json',
        },
      });

      if (response.ok) {
        const data = await response.json();
        messagingStats = {
          total_sent: data.total_sent || data.total_messages || 0,
          total_delivered: data.total_delivered || data.total_messages || 0,
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
      const channelsResponse = await fetch(`${API_BASE_URL}/api/v1/projects/${projectId}/messaging/channels`, {
        headers: {
          'Authorization': `Bearer ${$auth.token}`,
          'Content-Type': 'application/json',
        },
      });

      if (channelsResponse.ok) {
        const channelsData = await channelsResponse.json();
        const channels = Array.isArray(channelsData) ? channelsData : [];
        const systemChannel = channels.find(channel => 
          channel.type === 'system' && channel.name === 'System Notifications'
        );

        if (systemChannel) {
          // Load messages from the system channel - but this endpoint doesn't exist in admin API
          // So we'll use the all messages endpoint and filter
          const messagesResponse = await fetch(`${API_BASE_URL}/api/v1/projects/${projectId}/messaging/messages`, {
            headers: {
              'Authorization': `Bearer ${$auth.token}`,
              'Content-Type': 'application/json',
            },
          });

          if (messagesResponse.ok) {
            const messagesData = await messagesResponse.json();
            const messages = Array.isArray(messagesData) ? messagesData : [];
            // Debug: Show all messages first, then filter for GitHub-related notifications
            console.log('All messages:', messages);
            systemNotifications = messages
              .filter(msg => {
                console.log('Message metadata:', msg.metadata);
                // Include all GitHub webhook-related message types
                const isGitHubMessage = msg.metadata?.type === 'github_update' || 
                                      msg.metadata?.type === 'deployment_started' || 
                                      msg.metadata?.type === 'webhook_test' ||
                                      msg.type === 'system';
                return isGitHubMessage;
              })
              .sort((a, b) => new Date(b.created_at) - new Date(a.created_at))
              .slice(0, 10); // Show last 10 notifications
              
            // Calculate unread count (messages newer than last viewed time)
            if (activeTab !== 'notifications') {
              unreadCount = systemNotifications.filter(msg => 
                new Date(msg.created_at).getTime() > lastViewedTime
              ).length;
            }
            
            console.log('Filtered system notifications:', systemNotifications);
          } else {
            systemNotifications = [];
          }
        } else {
          systemNotifications = [];
        }
      } else {
        console.error('Failed to load system notifications:', channelsResponse.status);
        systemNotifications = [];
      }
    } catch (error) {
      console.error('Error loading system notifications:', error);
      systemNotifications = [];
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
    toast.error('Bericht kopiëren is nog niet geïmplementeerd');
  }

  async function deleteMessage(messageId: string) {
    toast.error('Bericht verwijderen is nog niet geïmplementeerd');
  }


  async function createMessage() {
    toast.error('Bericht aanmaken is nog niet geïmplementeerd');
    showCreateMessage = false;
  }



  async function sendMessage(messageId: string) {
    toast.error('Bericht versturen is nog niet geïmplementeerd');
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

  function toggleNotificationExpansion(notificationId: string) {
    if (expandedNotifications.has(notificationId)) {
      expandedNotifications.delete(notificationId);
    } else {
      expandedNotifications.add(notificationId);
    }
    expandedNotifications = expandedNotifications; // Trigger reactivity
  }

  function markNotificationsAsRead() {
    lastViewedTime = Date.now();
    unreadCount = 0;
  }

  async function deleteNotification(messageId: string) {
    try {
      const response = await fetch(`${API_BASE_URL}/api/v1/projects/${projectId}/messaging/messages/${messageId}`, {
        method: 'DELETE',
        headers: {
          'Authorization': `Bearer ${$auth.token}`,
          'Content-Type': 'application/json',
        },
      });

      if (response.ok) {
        toast.success('Bericht verwijderd');
        // Remove from local array immediately for better UX
        systemNotifications = systemNotifications.filter(n => n.id !== messageId);
        // Reload to ensure consistency
        await loadSystemNotifications();
      } else {
        const error = await response.json();
        toast.error(error.error || 'Kon bericht niet verwijderen');
      }
    } catch (error) {
      console.error('Error deleting notification:', error);
      toast.error('Netwerkfout bij verwijderen bericht');
    }
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
        on:click={() => {
          activeTab = 'notifications';
          markNotificationsAsRead();
        }}
        class="flex items-center space-x-2 py-2 px-1 border-b-2 text-sm font-medium transition-colors {
          activeTab === 'notifications' 
            ? 'border-primary text-primary' 
            : 'border-transparent text-muted-foreground hover:text-foreground hover:border-border'
        }"
      >
        <Icon name="github" size={16} />
        <span class="flex items-center space-x-2">
          <span>GitHub Meldingen ({systemNotifications.length})</span>
          {#if unreadCount > 0}
            <Badge class="bg-red-500 text-white text-xs px-2 py-0.5 min-w-[20px] h-5 flex items-center justify-center">
              {unreadCount}
            </Badge>
          {/if}
        </span>
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


  <!-- GitHub Notifications Tab -->
  {#if activeTab === 'notifications'}
    <div class="space-y-4">
      {#if systemNotifications.length > 0}
        {#each systemNotifications as notification}
          <Card 
            class="p-6 cursor-pointer hover:bg-muted/30 hover:border-muted-foreground/20 transition-colors"
            on:click={() => toggleNotificationExpansion(notification.id)}
          >
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
                      {:else if notification.metadata?.type === 'webhook_test'}
                        <Badge class="bg-green-100 dark:bg-green-900 text-green-800 dark:text-green-200">
                          Webhook Test
                        </Badge>
                      {/if}
                    </div>
                    <div class="flex items-center space-x-4 text-sm text-muted-foreground mt-1">
                      {#if notification.metadata?.type === 'webhook_test'}
                        <span>Hook ID: #{notification.metadata?.hook_id || 'Unknown'}</span>
                      {:else}
                        <span>Commit: {notification.metadata?.commit_hash?.substring(0, 8) || 'Unknown'}</span>
                      {/if}
                      <span>{formatNotificationTime(notification.created_at)}</span>
                    </div>
                  </div>
                </div>
                
                <div class="mt-3 text-sm text-foreground">
                  {#if expandedNotifications.has(notification.id)}
                    {@html notification.content.replace(/\*\*(.*?)\*\*/g, '<strong>$1</strong>').replace(/`(.*?)`/g, '<code class="bg-gray-100 dark:bg-gray-800 px-1 py-0.5 rounded text-xs">$1</code>').replace(/\n/g, '<br>')}
                  {:else}
                    <!-- Show preview of content -->
                    <div class="line-clamp-2">
                      {@html notification.content.replace(/\*\*(.*?)\*\*/g, '<strong>$1</strong>').replace(/`(.*?)`/g, '<code class="bg-gray-100 dark:bg-gray-800 px-1 py-0.5 rounded text-xs">$1</code>').replace(/\n/g, '<br>')}
                    </div>
                    <div class="mt-2 text-xs text-muted-foreground">
                      Klik om volledig bericht te bekijken
                    </div>
                  {/if}
                </div>
              </div>

              <div class="flex items-center space-x-3 ml-4">
                <!-- Expand/Collapse icon -->
                <Button
                  variant="ghost"
                  size="sm"
                  class="w-8 h-8 p-0"
                  on:click={(e) => {
                    e.stopPropagation();
                    toggleNotificationExpansion(notification.id);
                  }}
                >
                  <Icon 
                    name={expandedNotifications.has(notification.id) ? 'arrow-down' : 'arrow-right'} 
                    size={14} 
                    className="text-muted-foreground"
                  />
                </Button>
                
                {#if notification.metadata?.type === 'github_update' && notification.metadata?.can_deploy}
                  <Button
                    size="sm"
                    on:click={(e) => {
                      e.stopPropagation();
                      deployFromNotification(notification);
                    }}
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
                    on:click={(e) => {
                      e.stopPropagation();
                      window.open(notification.metadata.github_url, '_blank');
                    }}
                  >
                    <Icon name="github" size={14} className="mr-1" />
                    Bekijk op GitHub
                  </Button>
                {/if}
                
                <Button
                  variant="ghost"
                  size="sm"
                  on:click={(e) => {
                    e.stopPropagation();
                    deleteNotification(notification.id);
                  }}
                  class="text-red-600 hover:text-red-700 hover:bg-red-50"
                >
                  <Icon name="backup" size={14} className="mr-1" />
                  Verwijder
                </Button>
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
        <p class="text-sm text-muted-foreground">Configureer Gmail en Discord notificaties</p>
      </div>

      <!-- Gmail Settings -->
      <Card class="p-6">
        <div class="flex items-center space-x-2 mb-4">
          <Icon name="messaging" size={20} className="text-primary" />
          <h3 class="text-lg font-medium text-foreground">Gmail Instellingen</h3>
        </div>
        <div class="space-y-4">
          <div class="grid grid-cols-1 md:grid-cols-2 gap-4">
            <div>
              <Label>Gmail adres</Label>
              <Input type="email" placeholder="je-gmail@gmail.com" class="mt-1" />
            </div>
            <div>
              <Label>App wachtwoord</Label>
              <Input type="password" placeholder="xxxx-xxxx-xxxx-xxxx" class="mt-1" />
            </div>
          </div>
          <div class="text-sm text-muted-foreground bg-muted p-3 rounded-md">
            <p><strong>💡 Hoe Gmail App Wachtwoord instellen:</strong></p>
            <ol class="mt-2 space-y-1 list-decimal list-inside">
              <li>Ga naar Google Account instellingen</li>
              <li>Activeer 2-staps verificatie</li>
              <li>Ga naar "App wachtwoorden"</li>
              <li>Genereer een nieuw wachtwoord voor "Mail"</li>
              <li>Gebruik dit wachtwoord hier (niet je gewone Gmail wachtwoord)</li>
            </ol>
          </div>
          <div class="flex items-center space-x-2">
            <input type="checkbox" class="rounded border-border" />
            <Label class="text-sm">Gmail notificaties inschakelen voor GitHub webhooks</Label>
          </div>
        </div>
      </Card>

      <!-- Discord Settings -->
      <Card class="p-6">
        <div class="flex items-center space-x-2 mb-4">
          <Icon name="cloud" size={20} className="text-primary" />
          <h3 class="text-lg font-medium text-foreground">Discord Instellingen</h3>
        </div>
        <div class="space-y-4">
          <div>
            <Label>Discord Webhook URL</Label>
            <Input type="url" placeholder="https://discord.com/api/webhooks/..." class="mt-1" />
          </div>
          <div class="text-sm text-muted-foreground bg-muted p-3 rounded-md">
            <p><strong>💬 Hoe Discord Webhook instellen:</strong></p>
            <ol class="mt-2 space-y-1 list-decimal list-inside">
              <li>Ga naar je Discord server</li>
              <li>Klik op "Server Settings" → "Integrations" → "Webhooks"</li>
              <li>Klik "Create Webhook" of bewerk een bestaande</li>
              <li>Kies het kanaal waar berichten moeten komen</li>
              <li>Kopieer de Webhook URL en plak deze hier</li>
            </ol>
          </div>
          <div class="flex items-center space-x-2">
            <input type="checkbox" class="rounded border-border" />
            <Label class="text-sm">Discord notificaties inschakelen voor GitHub webhooks</Label>
          </div>
          <div class="grid grid-cols-1 md:grid-cols-2 gap-4">
            <div>
              <Label>Bot naam</Label>
              <Input type="text" value="CloudBox Bot" class="mt-1" />
            </div>
            <div>
              <Label>Kanaal voor notificaties</Label>
              <Input type="text" placeholder="#deployments" class="mt-1" />
            </div>
          </div>
        </div>
      </Card>

      <!-- Push Notifications Info -->
      <Card class="p-6">
        <div class="flex items-center space-x-2 mb-4">
          <Icon name="functions" size={20} className="text-primary" />
          <h3 class="text-lg font-medium text-foreground">Browser Push Notificaties</h3>
        </div>
        <div class="space-y-4">
          <div class="text-sm text-muted-foreground bg-blue-50 dark:bg-blue-900/20 p-4 rounded-md border border-blue-200 dark:border-blue-800">
            <h4 class="font-medium text-blue-900 dark:text-blue-100 mb-2">📱 Hoe werken Push Notificaties?</h4>
            <div class="space-y-2 text-blue-800 dark:text-blue-200">
              <p><strong>Browser Notificaties:</strong> CloudBox gebruikt de Web Push API om berichten rechtstreeks naar je browser te sturen, zelfs als de pagina niet open is.</p>
              
              <p><strong>Service Worker:</strong> Een achtergrond script verwerkt en toont notificaties, zelfs wanneer CloudBox niet actief is in je browser.</p>
              
              <p><strong>Permission vereist:</strong> Je browser vraagt eerst toestemming om notificaties te mogen tonen. Klik "Toestaan" voor de beste ervaring.</p>
              
              <p><strong>Cross-platform:</strong> Werkt op desktop (Chrome, Firefox, Safari) en mobiel (Android Chrome, iOS Safari 16.4+).</p>
            </div>
          </div>
          
          <div class="flex items-center space-x-2">
            <input type="checkbox" class="rounded border-border" />
            <Label class="text-sm">Browser push notificaties inschakelen voor GitHub webhooks</Label>
          </div>
          
          <Button variant="outline" class="w-full">
            <Icon name="functions" size={16} className="mr-2" />
            Test Push Notificatie
          </Button>
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

