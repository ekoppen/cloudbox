/**
 * Function Manager - CloudBox SDK
 * 
 * Manages serverless functions and their execution with full type safety
 */

import {
  CloudFunction,
  ExecuteFunctionRequest,
  ExecuteFunctionResponse
} from '../types';

import type { CloudBoxClient } from '../client';

export class FunctionManager {
  constructor(private client: CloudBoxClient) {}

  // FUNCTION CRUD OPERATIONS

  /**
   * Create a new serverless function
   */
  async create(functionData: {
    name: string;
    description?: string;
    code: string;
    runtime: string;
    environment_variables?: Record<string, string>;
    is_active?: boolean;
  }): Promise<CloudFunction> {
    return this.client.request<CloudFunction>('/functions', {
      method: 'POST',
      body: {
        ...functionData,
        environment_variables: functionData.environment_variables || {},
        is_active: functionData.is_active !== undefined ? functionData.is_active : true
      }
    });
  }

  /**
   * List all functions in the project
   */
  async list(options: {
    limit?: number;
    offset?: number;
    filter?: Record<string, any>;
  } = {}): Promise<CloudFunction[]> {
    return this.client.request<CloudFunction[]>('/functions', {
      params: options
    });
  }

  /**
   * Get a specific function by ID
   */
  async get(functionId: string): Promise<CloudFunction> {
    return this.client.request<CloudFunction>(`/functions/${functionId}`);
  }

  /**
   * Update a function's code or configuration
   */
  async update(functionId: string, updates: {
    name?: string;
    description?: string;
    code?: string;
    runtime?: string;
    environment_variables?: Record<string, string>;
    is_active?: boolean;
  }): Promise<CloudFunction> {
    return this.client.request<CloudFunction>(`/functions/${functionId}`, {
      method: 'PUT',
      body: updates
    });
  }

  /**
   * Delete a function
   */
  async delete(functionId: string): Promise<void> {
    await this.client.request(`/functions/${functionId}`, {
      method: 'DELETE'
    });
  }

  // FUNCTION EXECUTION

  /**
   * Execute a function with input data
   */
  async execute(functionId: string, input: ExecuteFunctionRequest): Promise<ExecuteFunctionResponse> {
    return this.client.request<ExecuteFunctionResponse>(`/functions/${functionId}/execute`, {
      method: 'POST',
      body: input
    });
  }

  /**
   * Execute a function by name (alternative to ID-based execution)
   */
  async executeByName(functionName: string, input: ExecuteFunctionRequest): Promise<ExecuteFunctionResponse> {
    return this.client.request<ExecuteFunctionResponse>(`/functions/by-name/${functionName}/execute`, {
      method: 'POST',
      body: input
    });
  }

  /**
   * Execute a function asynchronously (fire-and-forget)
   */
  async executeAsync(functionId: string, input: ExecuteFunctionRequest): Promise<string> {
    const response = await this.client.request<{ execution_id: string }>(`/functions/${functionId}/execute-async`, {
      method: 'POST',
      body: input
    });
    return response.execution_id;
  }

  // FUNCTION MANAGEMENT

  /**
   * Get function execution logs
   */
  async getLogs(functionId: string, options: {
    limit?: number;
    offset?: number;
    start_time?: string;
    end_time?: string;
  } = {}): Promise<Array<{
    execution_id: string;
    timestamp: string;
    level: string;
    message: string;
    execution_time?: number;
    memory_used?: number;
  }>> {
    return this.client.request(`/functions/${functionId}/logs`, {
      params: options
    });
  }

  /**
   * Get function execution statistics
   */
  async getStats(functionId: string, period: '1h' | '24h' | '7d' | '30d' = '24h'): Promise<{
    total_executions: number;
    successful_executions: number;
    failed_executions: number;
    avg_execution_time: number;
    avg_memory_used: number;
    success_rate: number;
    period: string;
  }> {
    return this.client.request(`/functions/${functionId}/stats`, {
      params: { period }
    });
  }

  /**
   * Test a function with sample data
   */
  async test(functionId: string, testInput: ExecuteFunctionRequest): Promise<{
    success: boolean;
    result?: any;
    error?: string;
    execution_time: number;
    memory_used: number;
  }> {
    return this.client.request(`/functions/${functionId}/test`, {
      method: 'POST',
      body: testInput
    });
  }

  // FUNCTION DEPLOYMENT

  /**
   * Deploy function code changes
   */
  async deploy(functionId: string): Promise<void> {
    await this.client.request(`/functions/${functionId}/deploy`, {
      method: 'POST'
    });
  }

  /**
   * Get function deployment status
   */
  async getDeploymentStatus(functionId: string): Promise<{
    status: 'deployed' | 'deploying' | 'failed' | 'pending';
    last_deployed: string;
    deployment_id?: string;
    error?: string;
  }> {
    return this.client.request(`/functions/${functionId}/deployment-status`);
  }

  // BATCH OPERATIONS

  /**
   * Delete multiple functions at once
   */
  async deleteMany(functionIds: string[]): Promise<void> {
    await this.client.request('/functions/batch', {
      method: 'DELETE',
      body: { function_ids: functionIds }
    });
  }

  /**
   * Enable or disable multiple functions
   */
  async setActiveStatus(functionIds: string[], isActive: boolean): Promise<void> {
    await this.client.request('/functions/batch', {
      method: 'PUT',
      body: { 
        function_ids: functionIds,
        is_active: isActive 
      }
    });
  }
}