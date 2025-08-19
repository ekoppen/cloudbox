---
title: Button Components - CloudBox Design System
description: Comprehensive button specifications with variants and states
feature: design-system
last-updated: 2025-08-19
version: 1.0.0
related-files:
  - ../tokens/colors.md
  - ../style-guide.md
  - forms.md
status: approved
---

# Button Components

## Overview
Buttons are the primary interactive elements in CloudBox, designed for clarity, accessibility, and consistent user experience. Inspired by Supabase's clean aesthetic with enhanced visual hierarchy.

## Button Variants

### Primary Button
**Main Actions**: For primary calls-to-action and important user actions

**Visual Specifications**:
- **Background**: `#059669` (primary-600)
- **Text Color**: `#ffffff` (white)
- **Border**: None
- **Font Weight**: 500 (medium)
- **Border Radius**: `6px`
- **Shadow**: `0 1px 2px rgba(0, 0, 0, 0.05)`

**Sizes**:
```css
/* Small */
.btn-primary-sm {
  height: 32px;
  padding: 0 12px;
  font-size: 14px;
  line-height: 20px;
}

/* Medium (Default) */
.btn-primary {
  height: 40px;
  padding: 0 16px;
  font-size: 16px;
  line-height: 24px;
}

/* Large */
.btn-primary-lg {
  height: 48px;
  padding: 0 20px;
  font-size: 18px;
  line-height: 28px;
}
```

**States**:
```css
/* Default */
.btn-primary {
  background-color: #059669;
  color: #ffffff;
  border: none;
  box-shadow: 0 1px 2px rgba(0, 0, 0, 0.05);
  transition: all 150ms ease-out;
}

/* Hover */
.btn-primary:hover {
  background-color: #047857;
  box-shadow: 0 1px 3px rgba(0, 0, 0, 0.1);
  transform: translateY(-1px);
}

/* Active/Pressed */
.btn-primary:active {
  background-color: #065f46;
  box-shadow: 0 1px 2px rgba(0, 0, 0, 0.05);
  transform: translateY(0);
}

/* Focus */
.btn-primary:focus {
  outline: none;
  box-shadow: 0 0 0 3px rgba(5, 150, 105, 0.2);
}

/* Disabled */
.btn-primary:disabled {
  background-color: #d1d5db;
  color: #9ca3af;
  cursor: not-allowed;
  box-shadow: none;
  transform: none;
}

/* Loading */
.btn-primary.loading {
  color: transparent;
  position: relative;
}

.btn-primary.loading::after {
  content: '';
  position: absolute;
  top: 50%;
  left: 50%;
  width: 16px;
  height: 16px;
  margin: -8px 0 0 -8px;
  border: 2px solid transparent;
  border-top: 2px solid #ffffff;
  border-radius: 50%;
  animation: spin 1s linear infinite;
}

@keyframes spin {
  0% { transform: rotate(0deg); }
  100% { transform: rotate(360deg); }
}
```

### Secondary Button
**Supporting Actions**: For secondary actions that complement primary buttons

**Visual Specifications**:
- **Background**: `transparent`
- **Text Color**: `#059669` (primary-600)
- **Border**: `1px solid #059669`
- **Font Weight**: 500 (medium)
- **Border Radius**: `6px`

**States**:
```css
/* Default */
.btn-secondary {
  background-color: transparent;
  color: #059669;
  border: 1px solid #059669;
  transition: all 150ms ease-out;
}

/* Hover */
.btn-secondary:hover {
  background-color: #ecfdf5;
  color: #047857;
  border-color: #047857;
}

/* Active */
.btn-secondary:active {
  background-color: #d1fae5;
  color: #065f46;
  border-color: #065f46;
}

/* Focus */
.btn-secondary:focus {
  outline: none;
  box-shadow: 0 0 0 3px rgba(5, 150, 105, 0.2);
}

/* Disabled */
.btn-secondary:disabled {
  background-color: transparent;
  color: #d1d5db;
  border-color: #d1d5db;
  cursor: not-allowed;
}
```

### Ghost Button
**Subtle Actions**: For tertiary actions and navigation elements

**Visual Specifications**:
- **Background**: `transparent`
- **Text Color**: `#6b7280` (neutral-500)
- **Border**: None
- **Font Weight**: 400 (regular)
- **Border Radius**: `6px`

**States**:
```css
/* Default */
.btn-ghost {
  background-color: transparent;
  color: #6b7280;
  border: none;
  transition: all 150ms ease-out;
}

/* Hover */
.btn-ghost:hover {
  background-color: #f3f4f6;
  color: #374151;
}

/* Active */
.btn-ghost:active {
  background-color: #e5e7eb;
  color: #1f2937;
}

/* Focus */
.btn-ghost:focus {
  outline: none;
  background-color: #f3f4f6;
  box-shadow: 0 0 0 2px rgba(107, 114, 128, 0.2);
}
```

### Destructive Button
**Dangerous Actions**: For delete, remove, or destructive operations

**Visual Specifications**:
- **Background**: `#dc2626` (error-600)
- **Text Color**: `#ffffff` (white)
- **Border**: None
- **Font Weight**: 500 (medium)
- **Border Radius**: `6px`

**States**:
```css
/* Default */
.btn-destructive {
  background-color: #dc2626;
  color: #ffffff;
  border: none;
  box-shadow: 0 1px 2px rgba(0, 0, 0, 0.05);
}

/* Hover */
.btn-destructive:hover {
  background-color: #b91c1c;
  box-shadow: 0 1px 3px rgba(0, 0, 0, 0.1);
}

/* Active */
.btn-destructive:active {
  background-color: #991b1b;
}

/* Focus */
.btn-destructive:focus {
  outline: none;
  box-shadow: 0 0 0 3px rgba(220, 38, 38, 0.2);
}
```

## Icon Buttons

### Icon-Only Buttons
**Compact Actions**: For toolbar buttons and space-constrained interfaces

**Specifications**:
- **Size**: `32px × 32px` (small), `40px × 40px` (medium), `48px × 48px` (large)
- **Icon Size**: `16px` (small), `20px` (medium), `24px` (large)
- **Border Radius**: `6px`

```css
.btn-icon {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  width: 40px;
  height: 40px;
  padding: 0;
  background-color: transparent;
  border: none;
  border-radius: 6px;
  color: #6b7280;
  transition: all 150ms ease-out;
}

.btn-icon:hover {
  background-color: #f3f4f6;
  color: #374151;
}

.btn-icon:focus {
  outline: none;
  box-shadow: 0 0 0 2px rgba(107, 114, 128, 0.2);
}
```

### Icon with Text
**Enhanced Clarity**: Icons combined with text labels

```css
.btn-icon-text {
  display: inline-flex;
  align-items: center;
  gap: 8px;
}

.btn-icon-text .icon {
  width: 16px;
  height: 16px;
  flex-shrink: 0;
}
```

## Button Groups

### Horizontal Button Group
**Related Actions**: Grouping related buttons together

```css
.btn-group {
  display: inline-flex;
  border-radius: 6px;
  box-shadow: 0 1px 2px rgba(0, 0, 0, 0.05);
}

.btn-group .btn {
  border-radius: 0;
  border-right: 1px solid #e5e7eb;
}

.btn-group .btn:first-child {
  border-top-left-radius: 6px;
  border-bottom-left-radius: 6px;
}

.btn-group .btn:last-child {
  border-top-right-radius: 6px;
  border-bottom-right-radius: 6px;
  border-right: none;
}

.btn-group .btn:only-child {
  border-radius: 6px;
  border-right: none;
}
```

## Special Variants

### Loading State Button
**Async Actions**: Buttons that trigger loading states

```html
<!-- Loading state markup -->
<button class="btn-primary loading" disabled>
  <span class="btn-text">Processing...</span>
</button>
```

### Split Button
**Primary Action + Options**: Combined action and dropdown

```css
.btn-split {
  display: inline-flex;
  border-radius: 6px;
  overflow: hidden;
  box-shadow: 0 1px 2px rgba(0, 0, 0, 0.05);
}

.btn-split .btn-main {
  border-top-right-radius: 0;
  border-bottom-right-radius: 0;
  border-right: 1px solid rgba(255, 255, 255, 0.2);
}

.btn-split .btn-dropdown {
  border-top-left-radius: 0;
  border-bottom-left-radius: 0;
  padding: 0 8px;
  min-width: 32px;
}
```

## Accessibility Specifications

### Keyboard Navigation
```css
/* Focus management */
.btn:focus-visible {
  outline: 2px solid #059669;
  outline-offset: 2px;
}

/* High contrast mode */
@media (prefers-contrast: high) {
  .btn {
    border: 1px solid currentColor;
  }
}

/* Reduced motion */
@media (prefers-reduced-motion: reduce) {
  .btn {
    transition: none;
  }
  
  .btn:hover {
    transform: none;
  }
}
```

### ARIA Attributes
```html
<!-- Loading button -->
<button class="btn-primary" aria-busy="true" disabled>
  <span aria-hidden="true">Processing...</span>
  <span class="sr-only">Please wait, processing your request</span>
</button>

<!-- Toggle button -->
<button class="btn-secondary" aria-pressed="false" aria-label="Toggle sidebar">
  <span aria-hidden="true">☰</span>
</button>

<!-- Destructive action -->
<button class="btn-destructive" aria-describedby="delete-warning">
  Delete Project
</button>
<div id="delete-warning" class="sr-only">
  This action cannot be undone
</div>
```

### Screen Reader Support
```css
/* Screen reader only text */
.sr-only {
  position: absolute;
  width: 1px;
  height: 1px;
  padding: 0;
  margin: -1px;
  overflow: hidden;
  clip: rect(0, 0, 0, 0);
  white-space: nowrap;
  border: 0;
}
```

## Usage Guidelines

### When to Use Each Variant

**Primary Button**:
- Main call-to-action on a page or section
- Form submissions (Save, Submit, Create)
- Critical user actions (Sign In, Purchase)
- Limit to 1-2 per screen/section

**Secondary Button**:
- Supporting actions alongside primary buttons
- Less critical actions (Cancel, Edit, View Details)
- Alternative paths or options

**Ghost Button**:
- Navigation elements
- Tertiary actions with minimal visual weight
- Close buttons, minimize buttons
- Actions within content areas

**Destructive Button**:
- Delete, remove, or destructive operations
- Actions that cannot be easily undone
- Use sparingly and with confirmation dialogs

### Best Practices

**Do's**:
- Use clear, action-oriented labels ("Create Project", not "Submit")
- Maintain consistent button hierarchy on each page
- Provide loading states for async actions
- Include proper focus management and keyboard support
- Use icons to enhance understanding, not replace text

**Don'ts**:
- Don't use multiple primary buttons in the same context
- Don't use destructive styling for non-destructive actions
- Don't make buttons too small (minimum 32px height)
- Don't rely solely on color to convey button purpose
- Don't use button styling for non-interactive elements

### Button Sizing Guidelines

**Small (32px height)**:
- Dense interfaces, tables, forms
- Secondary actions in compact spaces
- Mobile interfaces where space is limited

**Medium (40px height)**:
- Default size for most interfaces
- Primary and secondary actions
- Form elements and standard interactions

**Large (48px height)**:
- Hero sections, landing pages
- Critical actions requiring emphasis
- Touch-friendly mobile interfaces

## Implementation Examples

### React Component
```jsx
import React from 'react';
import { cn } from '@/lib/utils';

const Button = React.forwardRef(({ 
  className, 
  variant = 'primary', 
  size = 'medium', 
  loading = false,
  disabled = false,
  children, 
  ...props 
}, ref) => {
  const baseClasses = 'inline-flex items-center justify-center font-medium rounded-md transition-all duration-150 focus:outline-none focus:ring-2 focus:ring-offset-2';
  
  const variants = {
    primary: 'bg-primary-600 hover:bg-primary-700 text-white shadow-sm focus:ring-primary-500',
    secondary: 'border border-primary-600 text-primary-600 hover:bg-primary-50 focus:ring-primary-500',
    ghost: 'text-neutral-500 hover:bg-neutral-100 hover:text-neutral-700 focus:ring-neutral-500',
    destructive: 'bg-error-600 hover:bg-error-700 text-white shadow-sm focus:ring-error-500'
  };
  
  const sizes = {
    small: 'h-8 px-3 text-sm',
    medium: 'h-10 px-4 text-base',
    large: 'h-12 px-5 text-lg'
  };
  
  return (
    <button
      ref={ref}
      className={cn(
        baseClasses,
        variants[variant],
        sizes[size],
        (disabled || loading) && 'opacity-50 cursor-not-allowed',
        loading && 'relative text-transparent',
        className
      )}
      disabled={disabled || loading}
      aria-busy={loading}
      {...props}
    >
      {loading && (
        <div className="absolute inset-0 flex items-center justify-center">
          <div className="w-4 h-4 border-2 border-current border-t-transparent rounded-full animate-spin" />
        </div>
      )}
      {children}
    </button>
  );
});

Button.displayName = 'Button';

export { Button };
```

### Tailwind CSS Classes
```html
<!-- Primary Button -->
<button class="inline-flex items-center justify-center h-10 px-4 text-base font-medium text-white bg-primary-600 border border-transparent rounded-md shadow-sm hover:bg-primary-700 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-primary-500 transition-all duration-150">
  Create Project
</button>

<!-- Secondary Button -->
<button class="inline-flex items-center justify-center h-10 px-4 text-base font-medium text-primary-600 bg-transparent border border-primary-600 rounded-md hover:bg-primary-50 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-primary-500 transition-all duration-150">
  Cancel
</button>

<!-- Icon Button -->
<button class="inline-flex items-center justify-center w-10 h-10 text-neutral-500 bg-transparent border-none rounded-md hover:bg-neutral-100 hover:text-neutral-700 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-neutral-500 transition-all duration-150">
  <svg class="w-5 h-5" fill="currentColor" viewBox="0 0 20 20">
    <path fill-rule="evenodd" d="M4.293 4.293a1 1 0 011.414 0L10 8.586l4.293-4.293a1 1 0 111.414 1.414L11.414 10l4.293 4.293a1 1 0 01-1.414 1.414L10 11.414l-4.293 4.293a1 1 0 01-1.414-1.414L8.586 10 4.293 5.707a1 1 0 010-1.414z" clip-rule="evenodd" />
  </svg>
</button>
```

## Testing Checklist

### Visual Testing
- [ ] All button variants display correctly
- [ ] Hover and focus states work as expected
- [ ] Loading states animate properly
- [ ] Button sizes are consistent
- [ ] Icons align properly with text

### Accessibility Testing
- [ ] All buttons are keyboard accessible
- [ ] Focus indicators are visible and consistent
- [ ] Screen readers announce button purpose correctly
- [ ] Color contrast meets WCAG AA standards
- [ ] Disabled states are properly communicated

### Functional Testing
- [ ] Click handlers work correctly
- [ ] Loading states prevent multiple submissions
- [ ] Form validation integrates properly
- [ ] Button groups function as expected
- [ ] Responsive behavior works on mobile

---

*These button specifications ensure consistent, accessible, and visually appealing interactive elements throughout CloudBox. All implementations should follow these guidelines for optimal user experience.*