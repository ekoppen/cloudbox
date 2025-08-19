<script lang="ts">
  // Import Lucide Svelte icons
  import {
    Activity, AlertCircle, AlertTriangle, Archive, ArrowDown, ArrowLeft, ArrowRight, ArrowUp,
    Building, Calendar, Check, ChevronDown, ChevronLeft, ChevronRight, Clock, Cloud, Code,
    Cog, Crown, Database, Download, Edit, File, Folder, Globe, Grid, Home, Info, Key,
    Layout, Link, List, Lock, LogOut, MessageCircle, Moon, Package, Palette, Plus,
    Puzzle, RefreshCw, Rocket, Save, Search, Server, Settings, Shield, ShieldCheck,
    ShieldX, Sun, Trash2, TrendingUp, Upload, User, Users, Wrench, Zap, X, HardDrive, FileText,
    Terminal, Layers, Cpu, BarChart3, Monitor, GitBranch, Store, PlusCircle, CheckCircle,
    PauseCircle, XCircle, Eye, LayoutDashboard, Star,
  } from 'lucide-svelte';
  
  export let name: string;
  export let size: number = 24;
  export let color: string = 'currentColor';
  export let className: string = '';
  export let strokeWidth: number = 2;
  
  // Default to icon-contrast class for better theme handling if no explicit class is provided
  $: finalClassName = className || 'icon-contrast';
  
  // Icon mapping for backwards compatibility
  const iconMap: Record<string, any> = {
    // Dashboard and navigation
    'dashboard': Layout,
    'grid': Grid,
    'layout': Layout,
    'home': Home,
    'building': Building,
    'package': Package,
    'projects': Package,
    
    // System and technical
    'cloud': Cloud,
    'server': Server,
    'database': Database,
    'storage': HardDrive,
    'functions': Code,
    'messaging': MessageCircle,
    'deployments': Zap,
    'rocket': Rocket,
    
    // Authentication and security
    'auth': Lock,
    'lock': Lock,
    'key': Key,
    'shield': Shield,
    'shield-check': ShieldCheck,
    'shield-x': ShieldX,
    
    // User interface
    'settings': Settings,
    'cog': Cog,
    'user': User,
    'users': Users,
    'crown': Crown,
    'palette': Palette,
    
    // Actions
    'edit': Edit,
    'save': Save,
    'download': Download,
    'upload': Upload,
    'backup': Archive,
    'trash': Trash2,
    'refresh': RefreshCw,
    'refresh-cw': RefreshCw,
    'plus': Plus,
    
    // Navigation arrows
    'arrow-left': ArrowLeft,
    'arrow-right': ArrowRight,
    'arrow-up': ArrowUp,
    'arrow-down': ArrowDown,
    'chevron-down': ChevronDown,
    'chevron-left': ChevronLeft,
    'chevron-right': ChevronRight,
    'back': ArrowLeft,
    'forward': ArrowRight,
    
    // Status and feedback
    'check': Check,
    'x': X,
    'info': Info,
    'warning': AlertTriangle,
    'alert': AlertCircle,
    'clock': Clock,
    'calendar': Calendar,
    
    // Files and content
    'file': File,
    'file-text': FileText,
    'folder': Folder,
    'link': Link,
    'search': Search,
    'code': Code,
    'terminal': Terminal,
    
    // Theme
    'sun': Sun,
    'moon': Moon,
    
    // External services
    'github': GitBranch,
    'globe': Globe,
    
    // System monitoring
    'activity': Activity,
    'trending-up': TrendingUp,
    'bar-chart': BarChart3,
    'monitor': Monitor,
    'cpu': Cpu,
    'layers': Layers,
    
    // Plugin and marketplace specific
    'store': Store,
    'plus-circle': PlusCircle,
    'check-circle': CheckCircle,
    'pause-circle': PauseCircle,
    'x-circle': XCircle,
    'eye': Eye,
    'layout-dashboard': LayoutDashboard,
    'template': LayoutDashboard, // alias for template
    
    // Utility
    'tool': Wrench,
    'puzzle': Puzzle,
    'plugin': Puzzle,
    'zap': Zap,
    'list': List,
    'log-out': LogOut,
    'logout': LogOut,
    'star': Star,
  };
  
  // Get the icon component or fallback to AlertCircle for missing icons
  $: IconComponent = iconMap[name] || AlertCircle;
</script>

<svelte:component
  this={IconComponent}
  {size}
  {color}
  class={finalClassName}
  stroke-width={strokeWidth}
/>