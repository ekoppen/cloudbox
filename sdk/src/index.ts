/**
 * CloudBox JavaScript SDK
 * Official SDK for CloudBox Backend-as-a-Service
 */

export { CloudBox } from './CloudBox';
export { Auth } from './services/Auth';
export { Database } from './services/Database';
export { Storage } from './services/Storage';
export { Functions } from './services/Functions';
export { Messaging } from './services/Messaging';
export { Users } from './services/Users';

// Types
export * from './types';

// Utils
export { CloudBoxError } from './utils/CloudBoxError';

// Default export
import { CloudBox } from './CloudBox';
export default CloudBox;