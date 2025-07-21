import type { CloudBox } from '../CloudBox';
import type { CloudFunction, FunctionExecution, ExecuteFunctionOptions } from '../types';

export class Functions {
  private cloudbox: CloudBox;

  constructor(cloudbox: CloudBox) {
    this.cloudbox = cloudbox;
  }

  /**
   * Execute a function by name (public API)
   */
  async execute(
    functionName: string, 
    data?: Record<string, any>, 
    options: ExecuteFunctionOptions = {}
  ): Promise<any> {
    const requestData = {
      data: data || {},
      headers: options.headers || {}
    };

    // For public function execution via project API
    return this.cloudbox.apiClient.post(
      `${this.cloudbox.getProjectApiPath()}/functions/${functionName}`,
      requestData,
      {
        signal: options.timeout ? AbortSignal.timeout(options.timeout) : undefined
      }
    );
  }

  /**
   * Execute a function with GET method (for simple invocations)
   */
  async get(functionName: string, params?: Record<string, any>): Promise<any> {
    const queryParams = params ? new URLSearchParams(params).toString() : '';
    const url = `${this.cloudbox.getProjectApiPath()}/functions/${functionName}${queryParams ? `?${queryParams}` : ''}`;
    
    return this.cloudbox.apiClient.get(url);
  }

  /**
   * Execute a function with PUT method
   */
  async put(functionName: string, data?: Record<string, any>): Promise<any> {
    return this.cloudbox.apiClient.put(
      `${this.cloudbox.getProjectApiPath()}/functions/${functionName}`,
      { data: data || {} }
    );
  }

  /**
   * Execute a function with DELETE method
   */
  async delete(functionName: string, data?: Record<string, any>): Promise<any> {
    return this.cloudbox.apiClient.delete(
      `${this.cloudbox.getProjectApiPath()}/functions/${functionName}`,
      {
        body: JSON.stringify({ data: data || {} }),
        headers: { 'Content-Type': 'application/json' }
      }
    );
  }

  // Admin Functions (require authentication)

  /**
   * Get all functions (admin only)
   */
  async getFunctions(): Promise<CloudFunction[]> {
    return this.cloudbox.apiClient.get<CloudFunction[]>(
      `${this.cloudbox.getAdminApiPath()}/functions`
    );
  }

  /**
   * Create a new function (admin only)
   */
  async createFunction(functionData: {
    name: string;
    description?: string;
    runtime: string;
    language: string;
    code: string;
    entry_point?: string;
    timeout?: number;
    memory?: number;
    environment?: Record<string, any>;
    is_public?: boolean;
  }): Promise<CloudFunction> {
    return this.cloudbox.apiClient.post<CloudFunction>(
      `${this.cloudbox.getAdminApiPath()}/functions`,
      functionData
    );
  }

  /**
   * Get a specific function (admin only)
   */
  async getFunction(functionId: string): Promise<CloudFunction> {
    return this.cloudbox.apiClient.get<CloudFunction>(
      `${this.cloudbox.getAdminApiPath()}/functions/${functionId}`
    );
  }

  /**
   * Update a function (admin only)
   */
  async updateFunction(functionId: string, updates: Partial<{
    name: string;
    description: string;
    code: string;
    runtime: string;
    language: string;
    entry_point: string;
    timeout: number;
    memory: number;
    environment: Record<string, any>;
    is_public: boolean;
    is_active: boolean;
  }>): Promise<CloudFunction> {
    return this.cloudbox.apiClient.put<CloudFunction>(
      `${this.cloudbox.getAdminApiPath()}/functions/${functionId}`,
      updates
    );
  }

  /**
   * Delete a function (admin only)
   */
  async deleteFunction(functionId: string): Promise<void> {
    await this.cloudbox.apiClient.delete(
      `${this.cloudbox.getAdminApiPath()}/functions/${functionId}`
    );
  }

  /**
   * Deploy a function (admin only)
   */
  async deployFunction(functionId: string): Promise<{ message: string; status: string }> {
    return this.cloudbox.apiClient.post(
      `${this.cloudbox.getAdminApiPath()}/functions/${functionId}/deploy`
    );
  }

  /**
   * Execute a function by ID (admin only)
   */
  async executeById(functionId: string, data?: Record<string, any>): Promise<{
    execution_id: string;
    status: string;
    execution_time: number;
    response: any;
  }> {
    return this.cloudbox.apiClient.post(
      `${this.cloudbox.getAdminApiPath()}/functions/${functionId}/execute`,
      { data: data || {} }
    );
  }

  /**
   * Get function execution logs (admin only)
   */
  async getFunctionLogs(functionId: string, limit?: number): Promise<FunctionExecution[]> {
    const params = limit ? `?limit=${limit}` : '';
    return this.cloudbox.apiClient.get<FunctionExecution[]>(
      `${this.cloudbox.getAdminApiPath()}/functions/${functionId}/logs${params}`
    );
  }

  /**
   * Get function metrics (admin only)
   */
  async getFunctionMetrics(functionId: string, period?: '24h' | '7d' | '30d'): Promise<{
    total_executions: number;
    avg_execution_time: number;
    success_rate: number;
    error_rate: number;
    total_memory_used: number;
  }> {
    const params = period ? `?period=${period}` : '';
    return this.cloudbox.apiClient.get(
      `${this.cloudbox.getAdminApiPath()}/functions/${functionId}/metrics${params}`
    );
  }

  /**
   * Create a JavaScript function with template
   */
  async createJavaScriptFunction(
    name: string,
    description?: string,
    customCode?: string
  ): Promise<CloudFunction> {
    const defaultCode = `exports.handler = async (event, context) => {
  console.log('Function invoked with event:', event);
  
  return {
    statusCode: 200,
    body: {
      message: 'Hello from CloudBox Function!',
      timestamp: new Date().toISOString(),
      input: event
    }
  };
};`;

    return this.createFunction({
      name,
      description: description || `JavaScript function: ${name}`,
      runtime: 'nodejs18',
      language: 'javascript',
      code: customCode || defaultCode,
      entry_point: 'index.handler',
      timeout: 30,
      memory: 128,
      is_public: true
    });
  }

  /**
   * Create a Python function with template
   */
  async createPythonFunction(
    name: string,
    description?: string,
    customCode?: string
  ): Promise<CloudFunction> {
    const defaultCode = `def handler(event, context):
    print(f'Function invoked with event: {event}')
    
    return {
        'statusCode': 200,
        'body': {
            'message': 'Hello from CloudBox Function!',
            'timestamp': datetime.now().isoformat(),
            'input': event
        }
    }`;

    return this.createFunction({
      name,
      description: description || `Python function: ${name}`,
      runtime: 'python3.9',
      language: 'python',
      code: customCode || defaultCode,
      entry_point: 'main.handler',
      timeout: 30,
      memory: 128,
      is_public: true
    });
  }

  /**
   * Create a Go function with template
   */
  async createGoFunction(
    name: string,
    description?: string,
    customCode?: string
  ): Promise<CloudFunction> {
    const defaultCode = `package main

import (
    "context"
    "encoding/json"
    "fmt"
    "log"
    "time"
)

type Event map[string]interface{}
type Response struct {
    StatusCode int         \`json:"statusCode"\`
    Body       interface{} \`json:"body"\`
}

func Handler(ctx context.Context, event Event) (Response, error) {
    log.Printf("Function invoked with event: %+v", event)
    
    return Response{
        StatusCode: 200,
        Body: map[string]interface{}{
            "message":   "Hello from CloudBox Function!",
            "timestamp": time.Now().Format(time.RFC3339),
            "input":     event,
        },
    }, nil
}`;

    return this.createFunction({
      name,
      description: description || `Go function: ${name}`,
      runtime: 'go1.19',
      language: 'go',
      code: customCode || defaultCode,
      entry_point: 'main.Handler',
      timeout: 30,
      memory: 128,
      is_public: true
    });
  }
}