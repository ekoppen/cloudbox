/** @type {import('tailwindcss').Config} */
export default {
  darkMode: ["class"],
  content: [
    './src/**/*.{html,js,svelte,ts}',
  ],
  theme: {
    container: {
      center: true,
      padding: "2rem",
      screens: {
        "2xl": "1400px",
      },
    },
    extend: {
      colors: {
        // Supabase-inspired color system
        border: "hsl(var(--border))",
        input: "hsl(var(--input))",
        ring: "hsl(var(--ring))",
        background: "hsl(var(--background))",
        foreground: "hsl(var(--foreground))",
        
        // Primary brand colors - Supabase green
        primary: {
          DEFAULT: "hsl(var(--primary))",
          foreground: "hsl(var(--primary-foreground))",
          50: "hsl(var(--primary-50))",
          100: "hsl(var(--primary-100))",
          200: "hsl(var(--primary-200))",
          300: "hsl(var(--primary-300))",
          400: "hsl(var(--primary-400))",
          500: "hsl(var(--primary-500))",
          600: "hsl(var(--primary-600))",
          700: "hsl(var(--primary-700))",
          800: "hsl(var(--primary-800))",
          900: "hsl(var(--primary-900))",
        },
        
        // Neutral grays
        gray: {
          50: "hsl(var(--gray-50))",
          100: "hsl(var(--gray-100))",
          200: "hsl(var(--gray-200))",
          300: "hsl(var(--gray-300))",
          400: "hsl(var(--gray-400))",
          500: "hsl(var(--gray-500))",
          600: "hsl(var(--gray-600))",
          700: "hsl(var(--gray-700))",
          800: "hsl(var(--gray-800))",
          900: "hsl(var(--gray-900))",
        },
        
        // Semantic colors
        secondary: {
          DEFAULT: "hsl(var(--secondary))",
          foreground: "hsl(var(--secondary-foreground))",
        },
        destructive: {
          DEFAULT: "hsl(var(--destructive))",
          foreground: "hsl(var(--destructive-foreground))",
        },
        warning: {
          DEFAULT: "hsl(var(--warning))",
          foreground: "hsl(var(--warning-foreground))",
        },
        success: {
          DEFAULT: "hsl(var(--success))",
          foreground: "hsl(var(--success-foreground))",
        },
        
        // Surface colors
        muted: {
          DEFAULT: "hsl(var(--muted))",
          foreground: "hsl(var(--muted-foreground))",
        },
        accent: {
          DEFAULT: "hsl(var(--accent))",
          foreground: "hsl(var(--accent-foreground))",
        },
        popover: {
          DEFAULT: "hsl(var(--popover))",
          foreground: "hsl(var(--popover-foreground))",
        },
        card: {
          DEFAULT: "hsl(var(--card))",
          foreground: "hsl(var(--card-foreground))",
        },
        
        // Sidebar colors
        sidebar: {
          DEFAULT: "hsl(var(--sidebar))",
          foreground: "hsl(var(--sidebar-foreground))",
          border: "hsl(var(--sidebar-border))",
          hover: "hsl(var(--sidebar-hover))",
        },
        
        // CloudBox accent colors
        'cloudbox-blue': '#375D8D',
        'cloudbox-green': '#9DD8BA', 
        'cloudbox-purple': '#948BA4',
      },
      
      borderRadius: {
        lg: "var(--radius)",
        md: "calc(var(--radius) - 2px)",
        sm: "calc(var(--radius) - 4px)",
      },
      
      fontFamily: {
        sans: [
          "Roboto",
          "-apple-system",
          "BlinkMacSystemFont",
          "Segoe UI",
          "Oxygen",
          "Ubuntu",
          "Cantarell",
          "sans-serif",
        ],
        heading: [
          "Roboto Condensed",
          "Roboto",
          "-apple-system",
          "BlinkMacSystemFont",
          "Segoe UI",
          "sans-serif",
        ],
        mono: [
          "JetBrains Mono",
          "SF Mono",
          "Monaco",
          "Inconsolata",
          "Roboto Mono",
          "monospace",
        ],
      },
      
      fontSize: {
        // Enhanced typography scale for modern design
        'xs': ['0.75rem', { lineHeight: '1rem' }],          // 12px
        'sm': ['0.875rem', { lineHeight: '1.25rem' }],      // 14px
        'base': ['1rem', { lineHeight: '1.5rem' }],         // 16px
        'lg': ['1.125rem', { lineHeight: '1.75rem' }],      // 18px
        'xl': ['1.25rem', { lineHeight: '1.75rem' }],       // 20px
        '2xl': ['1.5rem', { lineHeight: '2rem' }],          // 24px
        '3xl': ['1.875rem', { lineHeight: '2.25rem' }],     // 30px
        '4xl': ['2.25rem', { lineHeight: '2.5rem' }],       // 36px
        '5xl': ['3rem', { lineHeight: '1' }],               // 48px
        '6xl': ['3.75rem', { lineHeight: '1' }],            // 60px
        
        // UI-specific sizes
        'ui-xs': ['0.6875rem', { lineHeight: '1rem' }],     // 11px - Fine print
        'ui-sm': ['0.8125rem', { lineHeight: '1.125rem' }], // 13px - Labels
        'ui-base': ['0.9375rem', { lineHeight: '1.375rem' }], // 15px - UI text
        
        // Display sizes
        'display-sm': ['2.25rem', { lineHeight: '2.5rem', letterSpacing: '-0.025em' }], // 36px
        'display-md': ['2.875rem', { lineHeight: '3.25rem', letterSpacing: '-0.025em' }], // 46px  
        'display-lg': ['3.5rem', { lineHeight: '4rem', letterSpacing: '-0.025em' }],     // 56px
        'display-xl': ['4.5rem', { lineHeight: '5rem', letterSpacing: '-0.025em' }],     // 72px
      },
      
      fontWeight: {
        light: '300',
        normal: '400',
        medium: '500',
        semibold: '600',
        bold: '700',
      },
      
      lineHeight: {
        none: '1',
        tight: '1.25',
        snug: '1.375',
        normal: '1.5',
        relaxed: '1.625',
        loose: '2',
      },
      
      letterSpacing: {
        tighter: '-0.05em',
        tight: '-0.025em',
        normal: '0em',
        wide: '0.025em',
        wider: '0.05em',
        widest: '0.1em',
      },
      
      spacing: {
        '4.5': '1.125rem',
        '18': '4.5rem',
        '88': '22rem',
      },
      
      width: {
        'sidebar': '240px',
        'sidebar-collapsed': '64px',
      },
      
      margin: {
        'sidebar': '240px',
        'sidebar-collapsed': '64px',
      },
      
      keyframes: {
        "accordion-down": {
          from: { height: 0 },
          to: { height: "var(--radix-accordion-content-height)" },
        },
        "accordion-up": {
          from: { height: "var(--radix-accordion-content-height)" },
          to: { height: 0 },
        },
        "fade-in": {
          "0%": { opacity: "0", transform: "translateY(-10px)" },
          "100%": { opacity: "1", transform: "translateY(0)" },
        },
        "fade-out": {
          "0%": { opacity: "1", transform: "translateY(0)" },
          "100%": { opacity: "0", transform: "translateY(-10px)" },
        },
      },
      
      animation: {
        "accordion-down": "accordion-down 0.2s ease-out",
        "accordion-up": "accordion-up 0.2s ease-out",
        "fade-in": "fade-in 0.2s ease-out",
        "fade-out": "fade-out 0.2s ease-out",
      },
      
      boxShadow: {
        'brand': '0 0 0 1px hsl(var(--primary) / 0.2), 0 1px 2px 0 rgb(0 0 0 / 0.05)',
        'brand-lg': '0 0 0 1px hsl(var(--primary) / 0.2), 0 10px 15px -3px rgb(0 0 0 / 0.1)',
      },
    },
  },
  plugins: [
    require('@tailwindcss/forms'),
    require('@tailwindcss/typography'),
    require('daisyui'),
  ],
  
  // DaisyUI Configuration
  daisyui: {
    themes: [
      {
        cloudbox: {
          // CloudBox light theme - improved contrast and accessibility
          'primary': '#375D8D',        // Blue from your palette
          'primary-content': '#ffffff', // White text on primary
          'secondary': '#8D83A0',      // Enhanced purple with better contrast
          'secondary-content': '#ffffff', // White text on secondary
          'accent': '#8DD4B8',         // Enhanced green with better contrast
          'accent-content': '#141414', // Dark text on accent
          'neutral': '#848485',        // Enhanced gray with better contrast
          'neutral-content': '#ffffff', // White text on neutral
          'base-100': '#FCFBFB',       // User requested rgb(252,251,251) background
          'base-200': '#F0F0F0',       // Darker for better contrast on light background
          'base-300': '#CCCCCC',       // Much darker borders for better visibility
          'base-content': '#141414',   // Darker text for WCAG AA compliance
          'info': '#375D8D',           // Blue
          'info-content': '#ffffff',   // White text on info
          'success': '#8DD4B8',        // Enhanced green
          'success-content': '#141414', // Dark text on success
          'warning': '#f59e0b',        // Amber
          'warning-content': '#141414', // Dark text on warning
          'error': '#ef4444',          // Red
          'error-content': '#ffffff',  // White text on error
        },
        'cloudbox-dark': {
          // CloudBox dark theme - enhanced contrast and saturation
          'primary': '#5A82C4',        // Brighter, more saturated blue for dark mode
          'primary-content': '#141414', // Dark text on primary
          'secondary': '#A394B8',      // Enhanced purple for dark mode
          'secondary-content': '#141414', // Dark text on secondary
          'accent': '#9FE4CC',         // Enhanced green for dark mode
          'accent-content': '#141414', // Dark text on accent
          'neutral': '#7A7A7B',        // Enhanced gray for dark mode
          'neutral-content': '#ffffff', // White text on neutral
          'base-100': '#171717',       // Enhanced dark background
          'base-200': '#1C1C1C',       // Better card contrast
          'base-300': '#333333',       // Better border visibility
          'base-content': '#FAFAFA',   // Light text for dark background
          'info': '#5A82C4',           // Enhanced blue
          'info-content': '#141414',   // Dark text on info
          'success': '#9FE4CC',        // Enhanced green
          'success-content': '#141414', // Dark text on success
          'warning': '#F5B341',        // Enhanced amber for dark
          'warning-content': '#141414', // Dark text on warning
          'error': '#F56565',          // Enhanced red for dark
          'error-content': '#ffffff',  // White text on error
        }
      }
    ],
    darkTheme: 'cloudbox-dark',
    base: true,
    styled: true,
    utils: true,
    prefix: '',
    logs: false,
    themeRoot: ':root',
  },
}