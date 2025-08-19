<script lang="ts">
  import Badge from '$lib/components/ui/badge.svelte';
  import Icon from '$lib/components/ui/icon.svelte';

  export let status: string;
  export let isEnabled: boolean = false;
  export let size: 'sm' | 'md' = 'sm';

  function getStatusConfig(status: string, isEnabled: boolean) {
    if (!isEnabled) {
      return {
        color: 'bg-gray-100 dark:bg-gray-800 text-gray-600 dark:text-gray-400',
        icon: 'pause',
        text: 'Inactief'
      };
    }

    switch (status) {
      case 'enabled':
        return {
          color: 'bg-green-100 dark:bg-green-900 text-green-800 dark:text-green-200',
          icon: 'auth',
          text: 'Actief'
        };
      case 'error':
        return {
          color: 'bg-red-100 dark:bg-red-900 text-red-800 dark:text-red-200',
          icon: 'backup',
          text: 'Fout'
        };
      case 'disabled':
        return {
          color: 'bg-gray-100 dark:bg-gray-800 text-gray-600 dark:text-gray-400',
          icon: 'pause',
          text: 'Uitgeschakeld'
        };
      case 'loading':
        return {
          color: 'bg-blue-100 dark:bg-blue-900 text-blue-800 dark:text-blue-200',
          icon: 'package',
          text: 'Laden...'
        };
      default:
        return {
          color: 'bg-muted text-muted-foreground',
          icon: 'package',
          text: status
        };
    }
  }

  $: statusConfig = getStatusConfig(status, isEnabled);
  $: iconSize = size === 'sm' ? 12 : 16;
</script>

<span class="inline-flex items-center px-2 py-1 text-xs font-medium rounded-full {statusConfig.color}">
  <Icon name={statusConfig.icon} size={iconSize} className="mr-1" />
  {statusConfig.text}
</span>