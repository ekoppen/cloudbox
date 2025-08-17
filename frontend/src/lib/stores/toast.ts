import { writable } from 'svelte/store';

export type ToastType = 'success' | 'error' | 'warning' | 'info';

export interface Toast {
  id: string;
  type: ToastType;
  title?: string;
  message: string;
  duration?: number;
  dismissible?: boolean;
}

interface ToastStore {
  subscribe: (fn: (toasts: Toast[]) => void) => () => void;
  toasts: Toast[];
  add: (toast: Omit<Toast, 'id'>) => void;
  remove: (id: string) => void;
  clear: () => void;
  success: (message: string, title?: string) => void;
  error: (message: string, title?: string) => void;
  warning: (message: string, title?: string) => void;
  info: (message: string, title?: string) => void;
}

function createToastStore(): ToastStore {
  const { subscribe, set, update } = writable<Toast[]>([]);

  return {
    subscribe,
    toasts: [],
    add: (toast: Omit<Toast, 'id'>) => {
      const id = Math.random().toString(36).substr(2, 9);
      const newToast: Toast = {
        id,
        duration: 5000,
        dismissible: true,
        ...toast,
      };

      update(toasts => [...toasts, newToast]);

      // Auto remove after duration
      if (newToast.duration && newToast.duration > 0) {
        setTimeout(() => {
          update(toasts => toasts.filter(t => t.id !== id));
        }, newToast.duration);
      }
    },
    remove: (id: string) => {
      update(toasts => toasts.filter(t => t.id !== id));
    },
    clear: () => {
      set([]);
    },
    success: (message: string, title?: string) => {
      toastStore.add({ type: 'success', message, title });
    },
    error: (message: string, title?: string) => {
      toastStore.add({ type: 'error', message, title, duration: 8000 });
    },
    warning: (message: string, title?: string) => {
      toastStore.add({ type: 'warning', message, title });
    },
    info: (message: string, title?: string) => {
      toastStore.add({ type: 'info', message, title });
    },
  };
}

export const toastStore = createToastStore();
export const toast = toastStore; // Alias for backward compatibility

// Helper function aliases
export const addToast = (message: string, type: ToastType = 'info', title?: string) => {
  toastStore.add({ type, message, title });
};

// Helper function to get current store value
function get<T>(store: { subscribe: (fn: (value: T) => void) => () => void }): T {
  let value: T;
  const unsubscribe = store.subscribe((v) => value = v);
  unsubscribe();
  return value!;
}