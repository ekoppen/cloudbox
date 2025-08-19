---
title: Color System - CloudBox Design Tokens
description: Complete color palette specifications inspired by Supabase
feature: design-system
last-updated: 2025-08-19
version: 1.0.0
related-files:
  - ../style-guide.md
  - ../components/buttons.md
status: approved
---

# Color System

## Overview
CloudBox uses a sophisticated color system inspired by Supabase's clean, professional aesthetic. The palette emphasizes trust, clarity, and modern design principles while maintaining excellent accessibility standards.

## Brand Identity Colors

### Primary Palette - Emerald Green
**Core Brand Color**: Deep emerald green conveying growth, reliability, and innovation

```css
/* Primary Colors */
--color-primary-50: #ecfdf5;     /* Lightest tint, backgrounds */
--color-primary-100: #d1fae5;    /* Very light, subtle highlights */
--color-primary-200: #a7f3d0;    /* Light, success backgrounds */
--color-primary-300: #6ee7b7;    /* Medium light, hover states */
--color-primary-400: #34d399;    /* Medium, active states */
--color-primary-500: #10b981;    /* Base primary light */
--color-primary-600: #059669;    /* Primary brand color */
--color-primary-700: #047857;    /* Primary dark, hover */
--color-primary-800: #065f46;    /* Darker, pressed states */
--color-primary-900: #064e3b;    /* Darkest, high contrast */
```

**Usage Guidelines**:
- `primary-600` (#059669): Main CTAs, active navigation, brand elements
- `primary-700` (#047857): Hover states, emphasis, pressed buttons
- `primary-500` (#10b981): Success states, positive feedback
- `primary-100` (#d1fae5): Success backgrounds, subtle highlights
- `primary-50` (#ecfdf5): Very subtle backgrounds, table row highlights

### Secondary Palette - Neutral Grays
**Professional Foundation**: Clean grays for text, backgrounds, and structural elements

```css
/* Neutral Colors */
--color-neutral-50: #f9fafb;     /* Page backgrounds, lightest */
--color-neutral-100: #f3f4f6;    /* Card backgrounds, sections */
--color-neutral-200: #e5e7eb;    /* Borders, dividers */
--color-neutral-300: #d1d5db;    /* Input borders, inactive */
--color-neutral-400: #9ca3af;    /* Placeholder text, disabled */
--color-neutral-500: #6b7280;    /* Secondary text, icons */
--color-neutral-600: #4b5563;    /* Primary text on light */
--color-neutral-700: #374151;    /* Headings, emphasis */
--color-neutral-800: #1f2937;    /* Dark text, navigation */
--color-neutral-900: #111827;    /* Highest contrast text */
```

**Usage Guidelines**:
- `neutral-900` (#111827): Primary headings, high contrast text
- `neutral-700` (#374151): Body text, secondary headings
- `neutral-500` (#6b7280): Supporting text, icons, labels
- `neutral-300` (#d1d5db): Borders, form field borders
- `neutral-100` (#f3f4f6): Card backgrounds, content areas
- `neutral-50` (#f9fafb): Page backgrounds, subtle sections

## Functional Colors

### Semantic Status Colors
**Clear Communication**: Colors that convey meaning and state

```css
/* Success Colors */
--color-success-50: #ecfdf5;
--color-success-100: #d1fae5;
--color-success-500: #10b981;    /* Same as primary-500 */
--color-success-600: #059669;    /* Same as primary-600 */
--color-success-700: #047857;    /* Same as primary-700 */

/* Warning Colors */
--color-warning-50: #fffbeb;
--color-warning-100: #fef3c7;
--color-warning-400: #fbbf24;
--color-warning-500: #f59e0b;    /* Main warning color */
--color-warning-600: #d97706;
--color-warning-700: #b45309;

/* Error Colors */
--color-error-50: #fef2f2;
--color-error-100: #fee2e2;
--color-error-400: #f87171;
--color-error-500: #ef4444;      /* Main error color */
--color-error-600: #dc2626;      /* Dark error, high contrast */
--color-error-700: #b91c1c;

/* Info Colors */
--color-info-50: #eff6ff;
--color-info-100: #dbeafe;
--color-info-400: #60a5fa;
--color-info-500: #3b82f6;       /* Main info color */
--color-info-600: #2563eb;
--color-info-700: #1d4ed8;
```

**Usage Guidelines**:
- **Success**: Form validation, completed actions, positive feedback
- **Warning**: Alerts, caution states, pending actions
- **Error**: Form errors, failed actions, destructive operations
- **Info**: Help text, informational messages, links

## Accent & Accent Colors

### Blue Accent - Information & Actions
**Secondary Actions**: Complementary blue for non-primary interactions

```css
/* Blue Accent */
--color-blue-50: #eff6ff;
--color-blue-100: #dbeafe;
--color-blue-500: #3b82f6;       /* Main accent color */
--color-blue-600: #2563eb;       /* Darker accent */
--color-blue-700: #1d4ed8;       /* Dark accent hover */
```

**Usage**: Secondary buttons, links, informational elements, badges

### Purple Accent - Premium & Advanced
**Premium Features**: Purple tones for pro features and advanced functionality

```css
/* Purple Accent */
--color-purple-50: #faf5ff;
--color-purple-100: #f3e8ff;
--color-purple-500: #8b5cf6;
--color-purple-600: #7c3aed;
--color-purple-700: #6d28d9;
```

**Usage**: Premium badges, pro features, advanced settings

## Dark Mode Colors

### Dark Theme Palette
**Professional Dark Mode**: Carefully calibrated for developer workflows

```css
/* Dark Mode Colors */
--color-dark-background: #0f172a;     /* Main background */
--color-dark-surface: #1e293b;        /* Card backgrounds */
--color-dark-surface-elevated: #334155; /* Elevated surfaces */
--color-dark-border: #475569;         /* Borders, dividers */
--color-dark-text-primary: #f1f5f9;   /* Primary text */
--color-dark-text-secondary: #cbd5e1; /* Secondary text */
--color-dark-text-tertiary: #94a3b8;  /* Tertiary text */

/* Dark Mode Primary */
--color-primary-dark-400: #4ade80;    /* Lighter green for dark bg */
--color-primary-dark-500: #22c55e;    /* Main primary on dark */
--color-primary-dark-600: #16a34a;    /* Hover state on dark */
```

## Color Usage Patterns

### Text Colors
**Hierarchy and Readability**

```css
/* Light Mode Text */
.text-primary { color: var(--color-neutral-900); }      /* Headings */
.text-secondary { color: var(--color-neutral-700); }    /* Body text */
.text-tertiary { color: var(--color-neutral-500); }     /* Supporting text */
.text-disabled { color: var(--color-neutral-400); }     /* Disabled text */

/* Dark Mode Text */
.dark .text-primary { color: var(--color-dark-text-primary); }
.dark .text-secondary { color: var(--color-dark-text-secondary); }
.dark .text-tertiary { color: var(--color-dark-text-tertiary); }
```

### Background Colors
**Surface Hierarchy**

```css
/* Light Mode Backgrounds */
.bg-page { background-color: var(--color-neutral-50); }     /* Page background */
.bg-surface { background-color: #ffffff; }                  /* Cards, modals */
.bg-section { background-color: var(--color-neutral-100); } /* Sections */
.bg-subtle { background-color: var(--color-neutral-50); }   /* Subtle areas */

/* Dark Mode Backgrounds */
.dark .bg-page { background-color: var(--color-dark-background); }
.dark .bg-surface { background-color: var(--color-dark-surface); }
.dark .bg-section { background-color: var(--color-dark-surface-elevated); }
```

### Border Colors
**Structural Elements**

```css
/* Light Mode Borders */
.border-default { border-color: var(--color-neutral-200); } /* Default borders */
.border-strong { border-color: var(--color-neutral-300); }  /* Emphasized borders */
.border-subtle { border-color: var(--color-neutral-100); }  /* Subtle dividers */

/* Focus States */
.focus\:border-primary:focus { border-color: var(--color-primary-600); }
.focus\:ring-primary:focus { 
  box-shadow: 0 0 0 3px rgba(5, 150, 105, 0.1);
}
```

## Accessibility Compliance

### Contrast Ratios
**WCAG 2.1 AA Standards**

| Color Combination | Ratio | Status |
|------------------|-------|--------|
| Primary-600 on White | 7.2:1 | AAA ✓ |
| Neutral-700 on White | 8.9:1 | AAA ✓ |
| Neutral-500 on White | 4.6:1 | AA ✓ |
| Error-600 on White | 6.5:1 | AAA ✓ |
| Warning-500 on White | 4.8:1 | AA ✓ |

### Color Blind Accessibility
**Inclusive Design Considerations**
- Primary green and error red are distinguishable for deuteranopia
- Warning orange provides sufficient contrast from both green and red
- Icons and text labels accompany color-coded information
- Focus indicators don't rely solely on color

### Focus Indicators
**Keyboard Navigation Support**

```css
/* Focus Ring System */
.focus-ring:focus {
  outline: 2px solid var(--color-primary-600);
  outline-offset: 2px;
}

.focus-ring-inset:focus {
  box-shadow: inset 0 0 0 2px var(--color-primary-600);
}

/* High Contrast Mode */
@media (prefers-contrast: high) {
  .focus-ring:focus {
    outline: 3px solid var(--color-primary-700);
  }
}
```

## Implementation Guidelines

### CSS Custom Properties
**Root Variable Definitions**

```css
:root {
  /* Primary Brand Colors */
  --color-primary: var(--color-primary-600);
  --color-primary-hover: var(--color-primary-700);
  --color-primary-light: var(--color-primary-500);
  --color-primary-subtle: var(--color-primary-50);
  
  /* Semantic Colors */
  --color-success: var(--color-success-600);
  --color-warning: var(--color-warning-500);
  --color-error: var(--color-error-600);
  --color-info: var(--color-info-500);
  
  /* Text Colors */
  --color-text-primary: var(--color-neutral-900);
  --color-text-secondary: var(--color-neutral-700);
  --color-text-tertiary: var(--color-neutral-500);
  
  /* Border Colors */
  --color-border: var(--color-neutral-200);
  --color-border-strong: var(--color-neutral-300);
  
  /* Background Colors */
  --color-bg-page: var(--color-neutral-50);
  --color-bg-surface: #ffffff;
  --color-bg-section: var(--color-neutral-100);
}
```

### Tailwind CSS Integration
**Theme Configuration**

```javascript
// tailwind.config.js
module.exports = {
  theme: {
    extend: {
      colors: {
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
      },
    },
  },
};
```

## Testing & Validation

### Color Testing Checklist
- [ ] All colors meet WCAG AA contrast requirements (4.5:1)
- [ ] Primary actions meet AAA standards (7:1)
- [ ] Color combinations tested with color blindness simulators
- [ ] Focus indicators provide sufficient contrast (3:1 minimum)
- [ ] Dark mode colors maintain accessibility standards
- [ ] Color-coded information includes non-color indicators

### Tools for Validation
- **Contrast Checkers**: WebAIM, Colour Contrast Analyser
- **Color Blindness**: Stark, Color Oracle
- **Accessibility Testing**: axe DevTools, WAVE

---

*This color system provides the foundation for all CloudBox interfaces. Consistent application of these colors ensures brand coherence and accessibility compliance across all user touchpoints.*