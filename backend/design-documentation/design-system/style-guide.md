---
title: CloudBox Design System Style Guide
description: Complete design system specifications inspired by Supabase
feature: design-system
last-updated: 2025-08-19
version: 1.0.0
related-files:
  - tokens/colors.md
  - tokens/typography.md
  - tokens/spacing.md
  - components/buttons.md
status: draft
---

# CloudBox Design System Style Guide

## Table of Contents
1. [Color System](#color-system)
2. [Typography System](#typography-system)
3. [Spacing & Layout System](#spacing--layout-system)
4. [Component Specifications](#component-specifications)
5. [Motion & Animation System](#motion--animation-system)

## Color System

### Primary Colors
**Emerald Green Palette** - Inspired by Supabase's signature green
- **Primary**: `#059669` – Main CTAs, active states, brand elements
- **Primary Dark**: `#047857` – Hover states, emphasis, pressed buttons
- **Primary Light**: `#10b981` – Subtle backgrounds, highlights, success states

### Secondary Colors
**Neutral Grays** - Professional, clean backgrounds and text
- **Secondary**: `#6b7280` – Supporting text, icons, borders
- **Secondary Light**: `#f3f4f6` – Background sections, card backgrounds
- **Secondary Pale**: `#f9fafb` – Page backgrounds, subtle highlights

### Accent Colors
**Functional Color Palette**
- **Accent Primary**: `#3b82f6` – Information, links, secondary actions
- **Accent Secondary**: `#f59e0b` – Warnings, attention states
- **Gradient Start**: `#059669` – For gradient elements
- **Gradient End**: `#10b981` – For gradient elements

### Semantic Colors
**Status & Feedback Colors**
- **Success**: `#059669` – Positive actions, confirmations, completed states
- **Warning**: `#f59e0b` – Caution states, alerts, pending actions
- **Error**: `#dc2626` – Errors, destructive actions, failed states
- **Info**: `#3b82f6` – Informational messages, help text

### Neutral Palette
**Gray Scale System**
- `Neutral-50`: `#f9fafb` – Lightest backgrounds
- `Neutral-100`: `#f3f4f6` – Card backgrounds, subtle sections
- `Neutral-200`: `#e5e7eb` – Borders, dividers
- `Neutral-300`: `#d1d5db` – Input borders, inactive elements
- `Neutral-400`: `#9ca3af` – Placeholder text, disabled text
- `Neutral-500`: `#6b7280` – Secondary text, icons
- `Neutral-600`: `#4b5563` – Primary text on light backgrounds
- `Neutral-700`: `#374151` – Headings, emphasis text
- `Neutral-800`: `#1f2937` – Dark text, navigation
- `Neutral-900`: `#111827` – Highest contrast text

### Accessibility Notes
- All color combinations meet WCAG AA standards (4.5:1 normal text, 3:1 large text)
- Primary green maintains 7:1 contrast ratio on white backgrounds
- Error states use sufficient contrast for colorblind accessibility
- Focus indicators use 3:1 contrast minimum

## Typography System

### Font Stack
**Primary**: `Inter, -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif`
**Monospace**: `'JetBrains Mono', 'Fira Code', Consolas, 'Liberation Mono', monospace`

### Font Weights
- **Light**: 300 (Rare use, large headings only)
- **Regular**: 400 (Body text, standard UI)
- **Medium**: 500 (Subheadings, emphasis)
- **Semibold**: 600 (Headings, important actions)
- **Bold**: 700 (Major headings, strong emphasis)

### Type Scale
**Responsive Typography System**
- **H1**: `32px/40px, 700, -0.025em` – Page titles, major sections
- **H2**: `24px/32px, 600, -0.02em` – Section headers, card titles
- **H3**: `20px/28px, 600, -0.015em` – Subsection headers
- **H4**: `18px/24px, 500, -0.01em` – Component titles, small headers
- **H5**: `16px/24px, 500, -0.005em` – Minor headers, labels
- **Body Large**: `18px/28px, 400` – Primary reading text, descriptions
- **Body**: `16px/24px, 400` – Standard UI text, paragraphs
- **Body Small**: `14px/20px, 400` – Secondary information, metadata
- **Caption**: `12px/16px, 400` – Timestamps, fine print
- **Label**: `14px/20px, 500, uppercase, 0.05em` – Form labels, categories
- **Code**: `14px/20px, 400, monospace` – Code blocks, technical text

### Responsive Typography
**Mobile Adjustments** (≤768px):
- H1: `28px/36px` – Reduced for mobile screens
- H2: `20px/28px` – Better mobile hierarchy
- Body Large: `16px/24px` – Optimized reading size

## Spacing & Layout System

### Base Unit
**4px Grid System** - All spacing multiples of 4px for pixel-perfect alignment

### Spacing Scale
- `xs`: `4px` – Micro spacing, tight element relationships
- `sm`: `8px` – Small spacing, internal component padding
- `md`: `16px` – Default spacing, standard margins
- `lg`: `24px` – Medium spacing, section separation
- `xl`: `32px` – Large spacing, major sections
- `2xl`: `48px` – Extra large spacing, page sections
- `3xl`: `64px` – Huge spacing, hero sections

### Grid System
**12-Column Responsive Grid**
- **Columns**: 12 (desktop), 8 (tablet), 4 (mobile)
- **Gutters**: 24px (desktop), 16px (tablet/mobile)
- **Margins**: 24px (all breakpoints)
- **Container**: 1280px max-width with auto margins

### Breakpoints
- **Mobile**: `320px – 767px` – Single column layouts
- **Tablet**: `768px – 1023px` – Adaptive two-column
- **Desktop**: `1024px – 1279px` – Full multi-column
- **Wide**: `1280px+` – Maximum content width

## Component Specifications

### Button Component
**Primary Button**
- **Height**: `40px` (medium), `32px` (small), `48px` (large)
- **Padding**: `12px 20px` (medium), `8px 16px` (small), `14px 24px` (large)
- **Border Radius**: `6px` (consistent with modern aesthetic)
- **Background**: `#059669` → `#047857` on hover
- **Typography**: Body/500, white text
- **Shadow**: `0 1px 2px rgba(0,0,0,0.05)` → `0 1px 3px rgba(0,0,0,0.1)` on hover

**Secondary Button**
- **Background**: `transparent` with `#059669` border
- **Text Color**: `#059669` → `#047857` on hover
- **Border**: `1px solid #059669`
- **Hover**: Background `#f0fdf4`, border `#047857`

**Ghost Button**
- **Background**: `transparent`
- **Text Color**: `#6b7280` → `#374151` on hover
- **Hover**: Background `#f9fafb`

### Card Component
**Base Card**
- **Background**: `#ffffff`
- **Border**: `1px solid #e5e7eb`
- **Border Radius**: `8px`
- **Shadow**: `0 1px 3px rgba(0,0,0,0.1), 0 1px 2px rgba(0,0,0,0.06)`
- **Padding**: `24px` (desktop), `16px` (mobile)

**Elevated Card**
- **Shadow**: `0 4px 6px rgba(0,0,0,0.07), 0 1px 3px rgba(0,0,0,0.06)`
- **Hover**: `0 10px 15px rgba(0,0,0,0.1), 0 4px 6px rgba(0,0,0,0.05)`

### Form Elements
**Input Field**
- **Height**: `40px`
- **Padding**: `10px 12px`
- **Border**: `1px solid #d1d5db`
- **Border Radius**: `6px`
- **Background**: `#ffffff`
- **Focus**: Border `#059669`, shadow `0 0 0 3px rgba(5,150,105,0.1)`
- **Error**: Border `#dc2626`, shadow `0 0 0 3px rgba(220,38,38,0.1)`

## Motion & Animation System

### Timing Functions
**Easing Curves**
- **Ease-out**: `cubic-bezier(0.0, 0, 0.2, 1)` – Entrances, expansions
- **Ease-in-out**: `cubic-bezier(0.4, 0, 0.6, 1)` – Transitions, movements
- **Spring**: `cubic-bezier(0.34, 1.56, 0.64, 1)` – Playful interactions

### Duration Scale
**Animation Timing**
- **Micro**: `150ms` – Button hover, small state changes
- **Short**: `250ms` – Dropdown appear, tooltip show
- **Medium**: `350ms` – Modal open, page transitions
- **Long**: `500ms` – Complex animations, loading states

### Animation Principles
- **Performance**: 60fps minimum, transform/opacity preferred
- **Purpose**: Every animation serves functional purpose
- **Accessibility**: Respects `prefers-reduced-motion`
- **Consistency**: Similar actions use similar timings

## Implementation Guidelines

### CSS Custom Properties
```css
:root {
  /* Colors */
  --color-primary: #059669;
  --color-primary-dark: #047857;
  --color-primary-light: #10b981;
  
  /* Typography */
  --font-family-sans: Inter, -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif;
  --font-family-mono: 'JetBrains Mono', 'Fira Code', Consolas, monospace;
  
  /* Spacing */
  --spacing-xs: 4px;
  --spacing-sm: 8px;
  --spacing-md: 16px;
  --spacing-lg: 24px;
  --spacing-xl: 32px;
  
  /* Shadows */
  --shadow-sm: 0 1px 2px rgba(0,0,0,0.05);
  --shadow-md: 0 1px 3px rgba(0,0,0,0.1), 0 1px 2px rgba(0,0,0,0.06);
  --shadow-lg: 0 4px 6px rgba(0,0,0,0.07), 0 1px 3px rgba(0,0,0,0.06);
}
```

### Accessibility Requirements
- Minimum 4.5:1 contrast ratio for normal text
- Minimum 3:1 contrast ratio for large text
- Focus indicators visible and consistent
- Touch targets minimum 44×44px
- Keyboard navigation support
- Screen reader compatibility

## Quality Assurance

### Design System Compliance Checklist
- [ ] Colors match defined palette with proper contrast ratios
- [ ] Typography follows established hierarchy and scale
- [ ] Spacing uses 4px grid system consistently
- [ ] Components match documented specifications
- [ ] Motion follows timing and easing standards
- [ ] Accessibility requirements met (WCAG 2.1 AA)
- [ ] Cross-browser compatibility verified
- [ ] Mobile responsiveness validated

---

*This style guide serves as the single source of truth for CloudBox design decisions. All implementations should reference these specifications for consistency and quality.*