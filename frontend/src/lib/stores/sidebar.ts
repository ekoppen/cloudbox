import { writable } from 'svelte/store';

interface SidebarState {
  isHovered: boolean;
  isCollapsed: boolean;
  context: 'dashboard' | 'project' | 'admin';
  projectId?: string;
  projectName?: string;
}

function createSidebarStore() {
  const { subscribe, set, update } = writable<SidebarState>({
    isHovered: false,
    isCollapsed: true,
    context: 'dashboard'
  });

  return {
    subscribe,
    setHovered: (hovered: boolean) => update(state => ({ ...state, isHovered: hovered })),
    setContext: (context: 'dashboard' | 'project' | 'admin', projectId?: string, projectName?: string) => 
      update(state => ({ ...state, context, projectId, projectName })),
    setCollapsed: (collapsed: boolean) => update(state => ({ ...state, isCollapsed: collapsed })),
    reset: () => set({
      isHovered: false,
      isCollapsed: true,
      context: 'dashboard'
    })
  };
}

export const sidebarStore = createSidebarStore();