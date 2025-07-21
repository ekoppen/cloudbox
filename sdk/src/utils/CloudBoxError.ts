import type { CloudBoxErrorDetails } from '../types';

export class CloudBoxError extends Error {
  public readonly code: string;
  public readonly details?: Record<string, any>;
  public readonly statusCode?: number;

  constructor(
    message: string,
    code: string = 'UNKNOWN_ERROR',
    details?: Record<string, any>,
    statusCode?: number
  ) {
    super(message);
    this.name = 'CloudBoxError';
    this.code = code;
    this.details = details;
    this.statusCode = statusCode;
    
    // Maintain proper stack trace
    if (Error.captureStackTrace) {
      Error.captureStackTrace(this, CloudBoxError);
    }
  }

  static fromResponse(response: Response, errorData?: any): CloudBoxError {
    const message = errorData?.error || errorData?.message || `HTTP ${response.status}: ${response.statusText}`;
    const code = errorData?.code || `HTTP_${response.status}`;
    
    return new CloudBoxError(message, code, errorData, response.status);
  }

  static fromNetworkError(error: Error): CloudBoxError {
    return new CloudBoxError(
      `Network error: ${error.message}`,
      'NETWORK_ERROR',
      { originalError: error.message }
    );
  }

  toJSON(): CloudBoxErrorDetails {
    return {
      code: this.code,
      message: this.message,
      details: this.details
    };
  }
}