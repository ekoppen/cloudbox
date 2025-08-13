/**
 * CloudBox SDK - TypeScript Entry Point
 * Official CloudBox SDK for JavaScript/TypeScript applications
 * 
 * @version 1.0.0
 * @author VibCode
 * @package @ekoppen/cloudbox-sdk
 */

// Export main client
export { CloudBoxClient } from './client';

// Export all managers
export { CollectionManager } from './managers/collections';
export { StorageManager } from './managers/storage';
export { UserManager } from './managers/users';
export { FunctionManager } from './managers/functions';

// Export all types
export * from './types';

// Default export for convenience
export { CloudBoxClient as default } from './client';