import { EventEmitter } from 'eventemitter3';
import type { CloudBox } from '../CloudBox';
import type { Channel, Message, ChannelMember, RealtimeSubscription, RealtimeCallback } from '../types';

export class Messaging extends EventEmitter {
  private cloudbox: CloudBox;
  private websocket: WebSocket | null = null;
  private subscriptions: Map<string, Set<RealtimeCallback>> = new Map();
  private reconnectAttempts = 0;
  private maxReconnectAttempts = 5;
  private reconnectDelay = 1000;

  constructor(cloudbox: CloudBox) {
    super();
    this.cloudbox = cloudbox;
  }

  /**
   * Get all channels
   */
  async getChannels(): Promise<Channel[]> {
    return this.cloudbox.apiClient.get<Channel[]>(
      `${this.cloudbox.getProjectApiPath()}/messaging/channels`
    );
  }

  /**
   * Create a new channel
   */
  async createChannel(
    name: string,
    options: {
      description?: string;
      type?: 'public' | 'private' | 'direct';
    } = {}
  ): Promise<Channel> {
    return this.cloudbox.apiClient.post<Channel>(
      `${this.cloudbox.getProjectApiPath()}/messaging/channels`,
      {
        name,
        description: options.description,
        type: options.type || 'public'
      }
    );
  }

  /**
   * Get a specific channel
   */
  async getChannel(channelId: string): Promise<Channel> {
    return this.cloudbox.apiClient.get<Channel>(
      `${this.cloudbox.getProjectApiPath()}/messaging/channels/${channelId}`
    );
  }

  /**
   * Update a channel
   */
  async updateChannel(
    channelId: string,
    updates: {
      name?: string;
      description?: string;
      is_active?: boolean;
    }
  ): Promise<Channel> {
    return this.cloudbox.apiClient.put<Channel>(
      `${this.cloudbox.getProjectApiPath()}/messaging/channels/${channelId}`,
      updates
    );
  }

  /**
   * Delete a channel
   */
  async deleteChannel(channelId: string): Promise<void> {
    await this.cloudbox.apiClient.delete(
      `${this.cloudbox.getProjectApiPath()}/messaging/channels/${channelId}`
    );
  }

  /**
   * Join a channel
   */
  async joinChannel(channelId: string): Promise<ChannelMember> {
    return this.cloudbox.apiClient.post<ChannelMember>(
      `${this.cloudbox.getProjectApiPath()}/messaging/channels/${channelId}/members`
    );
  }

  /**
   * Leave a channel
   */
  async leaveChannel(channelId: string, userId?: string): Promise<void> {
    const userIdParam = userId || 'me';
    await this.cloudbox.apiClient.delete(
      `${this.cloudbox.getProjectApiPath()}/messaging/channels/${channelId}/members/${userIdParam}`
    );
  }

  /**
   * Get channel members
   */
  async getChannelMembers(channelId: string): Promise<ChannelMember[]> {
    return this.cloudbox.apiClient.get<ChannelMember[]>(
      `${this.cloudbox.getProjectApiPath()}/messaging/channels/${channelId}/members`
    );
  }

  /**
   * Send a message to a channel
   */
  async sendMessage(
    channelId: string,
    content: string,
    options: {
      messageType?: 'text' | 'image' | 'file' | 'system';
      metadata?: Record<string, any>;
    } = {}
  ): Promise<Message> {
    return this.cloudbox.apiClient.post<Message>(
      `${this.cloudbox.getProjectApiPath()}/messaging/channels/${channelId}/messages`,
      {
        content,
        message_type: options.messageType || 'text',
        metadata: options.metadata
      }
    );
  }

  /**
   * Get messages from a channel
   */
  async getMessages(
    channelId: string,
    options: {
      limit?: number;
      offset?: number;
      before?: string;
      after?: string;
    } = {}
  ): Promise<Message[]> {
    const params = new URLSearchParams();
    if (options.limit) params.append('limit', options.limit.toString());
    if (options.offset) params.append('offset', options.offset.toString());
    if (options.before) params.append('before', options.before);
    if (options.after) params.append('after', options.after);

    const queryString = params.toString();
    const url = `${this.cloudbox.getProjectApiPath()}/messaging/channels/${channelId}/messages${queryString ? `?${queryString}` : ''}`;

    return this.cloudbox.apiClient.get<Message[]>(url);
  }

  /**
   * Get a specific message
   */
  async getMessage(channelId: string, messageId: string): Promise<Message> {
    return this.cloudbox.apiClient.get<Message>(
      `${this.cloudbox.getProjectApiPath()}/messaging/channels/${channelId}/messages/${messageId}`
    );
  }

  /**
   * Update a message
   */
  async updateMessage(
    channelId: string,
    messageId: string,
    content: string,
    metadata?: Record<string, any>
  ): Promise<Message> {
    return this.cloudbox.apiClient.put<Message>(
      `${this.cloudbox.getProjectApiPath()}/messaging/channels/${channelId}/messages/${messageId}`,
      { content, metadata }
    );
  }

  /**
   * Delete a message
   */
  async deleteMessage(channelId: string, messageId: string): Promise<void> {
    await this.cloudbox.apiClient.delete(
      `${this.cloudbox.getProjectApiPath()}/messaging/channels/${channelId}/messages/${messageId}`
    );
  }

  /**
   * Search messages in a channel
   */
  async searchMessages(
    channelId: string,
    query: string,
    options: {
      limit?: number;
      offset?: number;
    } = {}
  ): Promise<Message[]> {
    const params = new URLSearchParams();
    params.append('q', query);
    if (options.limit) params.append('limit', options.limit.toString());
    if (options.offset) params.append('offset', options.offset.toString());

    return this.cloudbox.apiClient.get<Message[]>(
      `${this.cloudbox.getProjectApiPath()}/messaging/channels/${channelId}/messages/search?${params.toString()}`
    );
  }

  /**
   * Mark messages as read
   */
  async markAsRead(channelId: string, messageId: string): Promise<void> {
    await this.cloudbox.apiClient.post(
      `${this.cloudbox.getProjectApiPath()}/messaging/channels/${channelId}/messages/${messageId}/read`
    );
  }

  /**
   * Connect to realtime messaging
   */
  async connectRealtime(): Promise<void> {
    if (this.websocket) {
      return; // Already connected
    }

    const wsUrl = this.cloudbox.apiClient.getBaseUrl()
      .replace('http://', 'ws://')
      .replace('https://', 'wss://') + 
      `${this.cloudbox.getProjectApiPath()}/messaging/ws?api_key=${this.cloudbox.config.apiKey}`;

    this.websocket = new WebSocket(wsUrl);

    return new Promise((resolve, reject) => {
      if (!this.websocket) return reject(new Error('WebSocket not initialized'));

      this.websocket.onopen = () => {
        console.log('CloudBox Messaging WebSocket connected');
        this.reconnectAttempts = 0;
        this.emit('connected');
        resolve();
      };

      this.websocket.onmessage = (event) => {
        try {
          const data = JSON.parse(event.data);
          this.handleRealtimeMessage(data);
        } catch (error) {
          console.error('Failed to parse WebSocket message:', error);
        }
      };

      this.websocket.onclose = (event) => {
        console.log('CloudBox Messaging WebSocket disconnected:', event.code, event.reason);
        this.emit('disconnected', { code: event.code, reason: event.reason });
        this.websocket = null;

        // Attempt to reconnect if not manually closed
        if (event.code !== 1000 && this.reconnectAttempts < this.maxReconnectAttempts) {
          this.attemptReconnect();
        }
      };

      this.websocket.onerror = (error) => {
        console.error('CloudBox Messaging WebSocket error:', error);
        this.emit('error', error);
        reject(error);
      };
    });
  }

  /**
   * Disconnect from realtime messaging
   */
  disconnectRealtime(): void {
    if (this.websocket) {
      this.websocket.close(1000, 'Client disconnect');
      this.websocket = null;
    }
    this.subscriptions.clear();
    this.emit('disconnected');
  }

  /**
   * Subscribe to channel messages
   */
  subscribeToChannel(channelId: string, callback: RealtimeCallback): RealtimeSubscription {
    const key = `channel:${channelId}`;
    
    if (!this.subscriptions.has(key)) {
      this.subscriptions.set(key, new Set());
    }
    
    this.subscriptions.get(key)!.add(callback);

    // Send subscription message to server
    if (this.websocket && this.websocket.readyState === WebSocket.OPEN) {
      this.websocket.send(JSON.stringify({
        type: 'subscribe',
        channel: channelId
      }));
    }

    return {
      unsubscribe: () => {
        const callbacks = this.subscriptions.get(key);
        if (callbacks) {
          callbacks.delete(callback);
          if (callbacks.size === 0) {
            this.subscriptions.delete(key);
            
            // Send unsubscribe message to server
            if (this.websocket && this.websocket.readyState === WebSocket.OPEN) {
              this.websocket.send(JSON.stringify({
                type: 'unsubscribe',
                channel: channelId
              }));
            }
          }
        }
      }
    };
  }

  /**
   * Subscribe to user presence updates
   */
  subscribeToPresence(callback: RealtimeCallback): RealtimeSubscription {
    const key = 'presence';
    
    if (!this.subscriptions.has(key)) {
      this.subscriptions.set(key, new Set());
    }
    
    this.subscriptions.get(key)!.add(callback);

    return {
      unsubscribe: () => {
        const callbacks = this.subscriptions.get(key);
        if (callbacks) {
          callbacks.delete(callback);
          if (callbacks.size === 0) {
            this.subscriptions.delete(key);
          }
        }
      }
    };
  }

  /**
   * Set user presence status
   */
  async setPresence(status: 'online' | 'away' | 'offline', customData?: Record<string, any>): Promise<void> {
    if (this.websocket && this.websocket.readyState === WebSocket.OPEN) {
      this.websocket.send(JSON.stringify({
        type: 'presence',
        status,
        data: customData
      }));
    }
  }

  private handleRealtimeMessage(data: any): void {
    const { type, channel, payload } = data;

    switch (type) {
      case 'message':
        this.notifySubscribers(`channel:${channel}`, {
          event: 'message',
          payload,
          timestamp: new Date().toISOString()
        });
        break;

      case 'message_updated':
        this.notifySubscribers(`channel:${channel}`, {
          event: 'message_updated',
          payload,
          timestamp: new Date().toISOString()
        });
        break;

      case 'message_deleted':
        this.notifySubscribers(`channel:${channel}`, {
          event: 'message_deleted',
          payload,
          timestamp: new Date().toISOString()
        });
        break;

      case 'user_joined':
        this.notifySubscribers(`channel:${channel}`, {
          event: 'user_joined',
          payload,
          timestamp: new Date().toISOString()
        });
        break;

      case 'user_left':
        this.notifySubscribers(`channel:${channel}`, {
          event: 'user_left',
          payload,
          timestamp: new Date().toISOString()
        });
        break;

      case 'presence_update':
        this.notifySubscribers('presence', {
          event: 'presence_update',
          payload,
          timestamp: new Date().toISOString()
        });
        break;

      default:
        console.warn('Unknown realtime message type:', type);
    }
  }

  private notifySubscribers(key: string, message: any): void {
    const callbacks = this.subscriptions.get(key);
    if (callbacks) {
      callbacks.forEach(callback => {
        try {
          callback(message);
        } catch (error) {
          console.error('Error in realtime callback:', error);
        }
      });
    }
  }

  private attemptReconnect(): void {
    this.reconnectAttempts++;
    const delay = this.reconnectDelay * Math.pow(2, this.reconnectAttempts - 1);

    console.log(`Attempting to reconnect in ${delay}ms (attempt ${this.reconnectAttempts}/${this.maxReconnectAttempts})`);

    setTimeout(async () => {
      try {
        await this.connectRealtime();
        console.log('Reconnected successfully');
      } catch (error) {
        console.error('Reconnection failed:', error);
        if (this.reconnectAttempts < this.maxReconnectAttempts) {
          this.attemptReconnect();
        } else {
          console.error('Max reconnection attempts reached');
          this.emit('reconnect_failed');
        }
      }
    }, delay);
  }

  /**
   * Get connection status
   */
  get isConnected(): boolean {
    return this.websocket !== null && this.websocket.readyState === WebSocket.OPEN;
  }
}