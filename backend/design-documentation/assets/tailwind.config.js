/** @type {import('tailwindcss').Config} */
module.exports = {
  content: [
    './src/**/*.{js,jsx,ts,tsx,html,svelte}',
    './components/**/*.{js,jsx,ts,tsx,html,svelte}',
    './pages/**/*.{js,jsx,ts,tsx,html,svelte}',
  ],
  darkMode: 'class',
  theme: {
    extend: {
      // CloudBox Color System
      colors: {
        // Primary Brand Colors - Emerald Green
        primary: {
          50: '#ecfdf5',
          100: '#d1fae5',
          200: '#a7f3d0',
          300: '#6ee7b7',
          400: '#34d399',
          500: '#10b981',
          600: '#059669', // Main brand color
          700: '#047857',
          800: '#065f46',
          900: '#064e3b',
        },
        
        // Neutral Colors - Professional Grays
        neutral: {
          50: '#f9fafb',
          100: '#f3f4f6',
          200: '#e5e7eb',
          300: '#d1d5db',
          400: '#9ca3af',
          500: '#6b7280',
          600: '#4b5563',
          700: '#374151',
          800: '#1f2937',
          900: '#111827',
        },
        
        // Semantic Colors
        success: {
          50: '#ecfdf5',
          100: '#d1fae5',
          200: '#a7f3d0',
          300: '#6ee7b7',
          400: '#34d399',
          500: '#10b981',
          600: '#059669',
          700: '#047857',
          800: '#065f46',
          900: '#064e3b',
        },
        
        warning: {
          50: '#fffbeb',
          100: '#fef3c7',
          200: '#fde68a',
          300: '#fcd34d',
          400: '#fbbf24',
          500: '#f59e0b',
          600: '#d97706',
          700: '#b45309',
          800: '#92400e',
          900: '#78350f',
        },
        
        error: {
          50: '#fef2f2',
          100: '#fee2e2',
          200: '#fecaca',
          300: '#fca5a5',
          400: '#f87171',
          500: '#ef4444',
          600: '#dc2626',
          700: '#b91c1c',
          800: '#991b1b',
          900: '#7f1d1d',
        },
        
        info: {
          50: '#eff6ff',
          100: '#dbeafe',
          200: '#bfdbfe',
          300: '#93c5fd',
          400: '#60a5fa',
          500: '#3b82f6',
          600: '#2563eb',
          700: '#1d4ed8',
          800: '#1e40af',
          900: '#1e3a8a',
        },
        
        // Accent Colors
        blue: {
          50: '#eff6ff',
          100: '#dbeafe',
          200: '#bfdbfe',
          300: '#93c5fd',
          400: '#60a5fa',
          500: '#3b82f6',
          600: '#2563eb',
          700: '#1d4ed8',
          800: '#1e40af',
          900: '#1e3a8a',
        },
        
        purple: {
          50: '#faf5ff',
          100: '#f3e8ff',
          200: '#e9d5ff',
          300: '#d8b4fe',
          400: '#c084fc',
          500: '#a855f7',
          600: '#9333ea',
          700: '#7c3aed',
          800: '#6b21a8',
          900: '#581c87',
        },
        
        // Dark Mode Colors
        dark: {
          background: '#0f172a',
          surface: '#1e293b',
          'surface-elevated': '#334155',
          border: '#475569',
          'text-primary': '#f1f5f9',
          'text-secondary': '#cbd5e1',
          'text-tertiary': '#94a3b8',
        },
      },
      
      // Typography System
      fontFamily: {
        sans: ['Inter', 'system-ui', '-apple-system', 'BlinkMacSystemFont', '"Segoe UI"', 'Roboto', 'sans-serif'],
        mono: ['"JetBrains Mono"', '"Fira Code"', 'Consolas', '"Liberation Mono"', 'monospace'],
      },
      
      fontSize: {
        // Custom type scale
        'xs': ['12px', '16px'],
        'sm': ['14px', '20px'],
        'base': ['16px', '24px'],
        'lg': ['18px', '28px'],
        'xl': ['20px', '28px'],
        '2xl': ['24px', '32px'],
        '3xl': ['32px', '40px'],
      },
      
      fontWeight: {
        light: '300',
        normal: '400',
        medium: '500',
        semibold: '600',
        bold: '700',
      },
      
      letterSpacing: {
        tighter: '-0.025em',
        tight: '-0.02em',
        normal: '0',
        wide: '0.025em',
        wider: '0.05em',
      },
      
      // Spacing System - 4px grid
      spacing: {
        '0.5': '2px',
        '1': '4px',
        '1.5': '6px',
        '2': '8px',
        '2.5': '10px',
        '3': '12px',
        '3.5': '14px',
        '4': '16px',
        '5': '20px',
        '6': '24px',
        '7': '28px',
        '8': '32px',
        '9': '36px',
        '10': '40px',
        '11': '44px',
        '12': '48px',
        '14': '56px',
        '16': '64px',
        '18': '72px',
        '20': '80px',
        '24': '96px',
        '28': '112px',
        '32': '128px',
      },
      
      // Border Radius
      borderRadius: {
        'none': '0',
        'sm': '2px',
        DEFAULT: '4px',
        'md': '6px',
        'lg': '8px',
        'xl': '12px',
        '2xl': '16px',
        'full': '9999px',
      },
      
      // Box Shadows
      boxShadow: {
        'xs': '0 1px 2px rgba(0, 0, 0, 0.05)',
        'sm': '0 1px 3px rgba(0, 0, 0, 0.1), 0 1px 2px rgba(0, 0, 0, 0.06)',
        'md': '0 4px 6px rgba(0, 0, 0, 0.07), 0 1px 3px rgba(0, 0, 0, 0.06)',
        'lg': '0 10px 15px rgba(0, 0, 0, 0.1), 0 4px 6px rgba(0, 0, 0, 0.05)',
        'xl': '0 20px 25px rgba(0, 0, 0, 0.1), 0 10px 10px rgba(0, 0, 0, 0.04)',
        'inner': 'inset 0 2px 4px rgba(0, 0, 0, 0.06)',
        'none': '0 0 #0000',
        
        // Focus shadows
        'focus-primary': '0 0 0 3px rgba(5, 150, 105, 0.2)',
        'focus-error': '0 0 0 3px rgba(220, 38, 38, 0.2)',
        'focus-neutral': '0 0 0 2px rgba(107, 114, 128, 0.2)',
      },
      
      // Animation & Transitions
      transitionDuration: {
        '150': '150ms',
        '200': '200ms',
        '250': '250ms',
        '300': '300ms',
        '350': '350ms',
        '400': '400ms',
        '500': '500ms',
      },
      
      transitionTimingFunction: {
        'ease-out': 'cubic-bezier(0.0, 0, 0.2, 1)',
        'ease-in-out': 'cubic-bezier(0.4, 0, 0.6, 1)',
        'spring': 'cubic-bezier(0.34, 1.56, 0.64, 1)',
      },
      
      // Animation keyframes
      keyframes: {
        'fade-in': {
          '0%': { opacity: '0' },
          '100%': { opacity: '1' },
        },
        'fade-out': {
          '0%': { opacity: '1' },
          '100%': { opacity: '0' },
        },
        'slide-in-right': {
          '0%': { transform: 'translateX(100%)' },
          '100%': { transform: 'translateX(0)' },
        },
        'slide-in-left': {
          '0%': { transform: 'translateX(-100%)' },
          '100%': { transform: 'translateX(0)' },
        },
        'slide-up': {
          '0%': { transform: 'translateY(100%)' },
          '100%': { transform: 'translateY(0)' },
        },
        'scale-in': {
          '0%': { transform: 'scale(0.9)', opacity: '0' },
          '100%': { transform: 'scale(1)', opacity: '1' },
        },
        'spin': {
          '0%': { transform: 'rotate(0deg)' },
          '100%': { transform: 'rotate(360deg)' },
        },
        'pulse': {
          '0%, 100%': { opacity: '1' },
          '50%': { opacity: '0.5' },
        },
      },
      
      animation: {
        'fade-in': 'fade-in 200ms ease-out',
        'fade-out': 'fade-out 200ms ease-out',
        'slide-in-right': 'slide-in-right 350ms ease-out',
        'slide-in-left': 'slide-in-left 350ms ease-out',
        'slide-up': 'slide-up 350ms ease-out',
        'scale-in': 'scale-in 200ms ease-out',
        'spin': 'spin 1s linear infinite',
        'pulse': 'pulse 2s cubic-bezier(0.4, 0, 0.6, 1) infinite',
      },
      
      // Breakpoints (for reference - Tailwind uses these by default)
      screens: {
        'sm': '640px',
        'md': '768px',
        'lg': '1024px',
        'xl': '1280px',
        '2xl': '1536px',
      },
      
      // Container
      container: {
        center: true,
        padding: {
          DEFAULT: '1rem',
          sm: '1.5rem',
          lg: '2rem',
        },
        screens: {
          sm: '640px',
          md: '768px',
          lg: '1024px',
          xl: '1280px',
        },
      },
    },
  },
  plugins: [
    // Focus-visible plugin for better focus management
    require('@tailwindcss/forms')({
      strategy: 'class',
    }),
    
    // Typography plugin for rich text content
    require('@tailwindcss/typography'),
    
    // Custom component utilities
    function({ addUtilities, addComponents, theme }) {
      // Custom button utilities
      addComponents({
        '.btn': {
          '@apply inline-flex items-center justify-center font-medium rounded-md transition-all duration-150 focus:outline-none focus:ring-2 focus:ring-offset-2': {},
        },
        '.btn-sm': {
          '@apply h-8 px-3 text-sm': {},
        },
        '.btn-md': {
          '@apply h-10 px-4 text-base': {},
        },
        '.btn-lg': {
          '@apply h-12 px-5 text-lg': {},
        },
        '.btn-primary': {
          '@apply bg-primary-600 hover:bg-primary-700 text-white shadow-sm focus:ring-primary-500': {},
        },
        '.btn-secondary': {
          '@apply border border-primary-600 text-primary-600 hover:bg-primary-50 focus:ring-primary-500': {},
        },
        '.btn-ghost': {
          '@apply text-neutral-500 hover:bg-neutral-100 hover:text-neutral-700 focus:ring-neutral-500': {},
        },
        '.btn-destructive': {
          '@apply bg-error-600 hover:bg-error-700 text-white shadow-sm focus:ring-error-500': {},
        },
      });
      
      // Card components
      addComponents({
        '.card': {
          '@apply bg-white border border-neutral-200 rounded-lg shadow-sm p-6 transition-all duration-200': {},
        },
        '.card-elevated': {
          '@apply shadow-md hover:shadow-lg hover:-translate-y-1 transition-all duration-250': {},
        },
        '.card-interactive': {
          '@apply cursor-pointer hover:border-neutral-300 hover:shadow-md hover:-translate-y-0.5 focus:outline-none focus:border-primary-600 focus:ring-2 focus:ring-primary-500 focus:ring-offset-1': {},
        },
      });
      
      // Form components
      addComponents({
        '.form-input': {
          '@apply block w-full px-3 py-2 border border-neutral-300 rounded-md shadow-sm placeholder-neutral-400 focus:outline-none focus:ring-2 focus:ring-primary-500 focus:border-primary-500 transition-colors duration-150': {},
        },
        '.form-label': {
          '@apply block text-sm font-medium text-neutral-700 mb-1': {},
        },
        '.form-error': {
          '@apply text-sm text-error-600 mt-1': {},
        },
      });
      
      // Utility classes
      addUtilities({
        '.text-gradient': {
          'background': `linear-gradient(135deg, ${theme('colors.primary.600')}, ${theme('colors.primary.500')})`,
          'background-clip': 'text',
          '-webkit-background-clip': 'text',
          'color': 'transparent',
        },
        '.bg-gradient-primary': {
          'background': `linear-gradient(135deg, ${theme('colors.primary.600')}, ${theme('colors.primary.500')})`,
        },
        '.bg-gradient-dark': {
          'background': `linear-gradient(135deg, ${theme('colors.neutral.800')}, ${theme('colors.neutral.900')})`,
        },
      });
    },
  ],
};
