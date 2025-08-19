---
title: Card Components - CloudBox Design System
description: Card layout specifications with visual hierarchy and content organization
feature: design-system
last-updated: 2025-08-19
version: 1.0.0
related-files:
  - ../tokens/colors.md
  - ../tokens/spacing.md
  - buttons.md
status: approved
---

# Card Components

## Overview
Cards are fundamental content containers in CloudBox, providing structured layouts for information display. Designed with Supabase-inspired aesthetics emphasizing clean backgrounds, subtle shadows, and clear content hierarchy.

## Base Card Specifications

### Visual Foundation
**Core Design Elements**:
- **Background**: `#ffffff` (pure white)
- **Border**: `1px solid #e5e7eb` (neutral-200)
- **Border Radius**: `8px` (slightly rounded for modern feel)
- **Shadow**: `0 1px 3px rgba(0, 0, 0, 0.1), 0 1px 2px rgba(0, 0, 0, 0.06)`
- **Padding**: `24px` (desktop), `16px` (mobile)

### Base Card CSS
```css
.card {
  background-color: #ffffff;
  border: 1px solid #e5e7eb;
  border-radius: 8px;
  box-shadow: 0 1px 3px rgba(0, 0, 0, 0.1), 0 1px 2px rgba(0, 0, 0, 0.06);
  padding: 24px;
  transition: all 200ms ease-out;
}

/* Mobile responsive */
@media (max-width: 768px) {
  .card {
    padding: 16px;
    border-radius: 6px;
  }
}

/* Dark mode */
.dark .card {
  background-color: #1e293b;
  border-color: #475569;
  box-shadow: 0 1px 3px rgba(0, 0, 0, 0.2), 0 1px 2px rgba(0, 0, 0, 0.1);
}
```

## Card Variants

### Elevated Card
**Enhanced Prominence**: For important content requiring visual emphasis

**Visual Specifications**:
- **Shadow**: `0 4px 6px rgba(0, 0, 0, 0.07), 0 1px 3px rgba(0, 0, 0, 0.06)`
- **Hover**: `0 10px 15px rgba(0, 0, 0, 0.1), 0 4px 6px rgba(0, 0, 0, 0.05)`
- **Transform**: `translateY(-2px)` on hover

```css
.card-elevated {
  box-shadow: 0 4px 6px rgba(0, 0, 0, 0.07), 0 1px 3px rgba(0, 0, 0, 0.06);
  transition: all 250ms ease-out;
}

.card-elevated:hover {
  box-shadow: 0 10px 15px rgba(0, 0, 0, 0.1), 0 4px 6px rgba(0, 0, 0, 0.05);
  transform: translateY(-2px);
}
```

### Interactive Card
**Clickable Content**: Cards that function as large interactive elements

```css
.card-interactive {
  cursor: pointer;
  transition: all 200ms ease-out;
}

.card-interactive:hover {
  border-color: #d1d5db;
  box-shadow: 0 4px 6px rgba(0, 0, 0, 0.07), 0 1px 3px rgba(0, 0, 0, 0.06);
  transform: translateY(-1px);
}

.card-interactive:active {
  transform: translateY(0);
  box-shadow: 0 1px 3px rgba(0, 0, 0, 0.1), 0 1px 2px rgba(0, 0, 0, 0.06);
}

.card-interactive:focus {
  outline: none;
  border-color: #059669;
  box-shadow: 0 0 0 3px rgba(5, 150, 105, 0.1);
}
```

### Status Cards
**Semantic Feedback**: Cards with contextual coloring for different states

```css
/* Success Card */
.card-success {
  border-left: 4px solid #059669;
  background-color: #f0fdf4;
  border-color: #d1fae5;
}

/* Warning Card */
.card-warning {
  border-left: 4px solid #f59e0b;
  background-color: #fffbeb;
  border-color: #fef3c7;
}

/* Error Card */
.card-error {
  border-left: 4px solid #dc2626;
  background-color: #fef2f2;
  border-color: #fee2e2;
}

/* Info Card */
.card-info {
  border-left: 4px solid #3b82f6;
  background-color: #eff6ff;
  border-color: #dbeafe;
}
```

## Card Layout Patterns

### Project Card
**Project Dashboard**: Standard layout for project items

```html
<div class="card card-interactive">
  <!-- Card Header -->
  <div class="card-header">
    <div class="flex items-start justify-between">
      <div class="flex items-center space-x-3">
        <div class="card-icon">
          <svg class="w-6 h-6 text-primary-600" fill="currentColor" viewBox="0 0 20 20">
            <!-- Project icon -->
          </svg>
        </div>
        <div>
          <h3 class="card-title">Project Name</h3>
          <p class="card-subtitle">Last updated 2 hours ago</p>
        </div>
      </div>
      <div class="card-actions">
        <button class="btn-icon">
          <svg class="w-5 h-5" fill="currentColor" viewBox="0 0 20 20">
            <!-- Menu icon -->
          </svg>
        </button>
      </div>
    </div>
  </div>
  
  <!-- Card Content -->
  <div class="card-content">
    <p class="card-description">Project description with key details and current status information.</p>
  </div>
  
  <!-- Card Footer -->
  <div class="card-footer">
    <div class="flex items-center justify-between">
      <div class="flex items-center space-x-4">
        <span class="card-badge card-badge-success">Active</span>
        <span class="card-meta">5 deployments</span>
      </div>
      <div class="flex items-center space-x-2">
        <button class="btn-secondary btn-sm">Settings</button>
        <button class="btn-primary btn-sm">Deploy</button>
      </div>
    </div>
  </div>
</div>
```

### Metric Card
**Dashboard Metrics**: For displaying key performance indicators

```html
<div class="card">
  <div class="card-header">
    <div class="flex items-center justify-between">
      <h3 class="card-title">Total Deployments</h3>
      <div class="card-icon-small">
        <svg class="w-5 h-5 text-neutral-500" fill="currentColor" viewBox="0 0 20 20">
          <!-- Metric icon -->
        </svg>
      </div>
    </div>
  </div>
  
  <div class="card-content">
    <div class="metric-value">1,247</div>
    <div class="metric-change metric-change-positive">
      <svg class="w-4 h-4" fill="currentColor" viewBox="0 0 20 20">
        <!-- Trend up icon -->
      </svg>
      <span>+12.5% from last month</span>
    </div>
  </div>
</div>
```

### Feature Card
**Feature Showcase**: For highlighting features or capabilities

```html
<div class="card card-elevated">
  <div class="card-media">
    <div class="card-media-icon">
      <svg class="w-8 h-8 text-primary-600" fill="currentColor" viewBox="0 0 20 20">
        <!-- Feature icon -->
      </svg>
    </div>
  </div>
  
  <div class="card-header">
    <h3 class="card-title">Feature Title</h3>
  </div>
  
  <div class="card-content">
    <p class="card-description">Detailed description of the feature, its benefits, and how it helps users achieve their goals.</p>
  </div>
  
  <div class="card-footer">
    <button class="btn-ghost">Learn More →</button>
  </div>
</div>
```

## Card Typography & Content Elements

### Typography Hierarchy
```css
/* Card Title */
.card-title {
  font-size: 18px;
  font-weight: 600;
  color: #1f2937;
  line-height: 24px;
  margin: 0;
}

/* Card Subtitle */
.card-subtitle {
  font-size: 14px;
  font-weight: 400;
  color: #6b7280;
  line-height: 20px;
  margin: 2px 0 0 0;
}

/* Card Description */
.card-description {
  font-size: 16px;
  font-weight: 400;
  color: #4b5563;
  line-height: 24px;
  margin: 12px 0 0 0;
}

/* Card Meta Text */
.card-meta {
  font-size: 14px;
  font-weight: 400;
  color: #9ca3af;
  line-height: 20px;
}

/* Dark mode adjustments */
.dark .card-title { color: #f1f5f9; }
.dark .card-subtitle { color: #cbd5e1; }
.dark .card-description { color: #e2e8f0; }
.dark .card-meta { color: #94a3b8; }
```

### Card Structure Elements
```css
/* Card Header */
.card-header {
  margin-bottom: 16px;
}

.card-header:last-child {
  margin-bottom: 0;
}

/* Card Content */
.card-content {
  margin-bottom: 16px;
}

.card-content:last-child {
  margin-bottom: 0;
}

/* Card Footer */
.card-footer {
  padding-top: 16px;
  border-top: 1px solid #f3f4f6;
  margin-top: 16px;
}

.dark .card-footer {
  border-color: #374151;
}

/* Card Actions */
.card-actions {
  display: flex;
  align-items: center;
  gap: 8px;
}
```

## Card Media Elements

### Card Icons
```css
/* Large icon container */
.card-icon {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 40px;
  height: 40px;
  background-color: #f0fdf4;
  border-radius: 8px;
  flex-shrink: 0;
}

/* Small icon */
.card-icon-small {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 32px;
  height: 32px;
  flex-shrink: 0;
}

/* Media icon for feature cards */
.card-media {
  text-align: center;
  margin-bottom: 16px;
}

.card-media-icon {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  width: 64px;
  height: 64px;
  background-color: #f0fdf4;
  border-radius: 12px;
  margin: 0 auto;
}
```

### Card Badges
```css
.card-badge {
  display: inline-flex;
  align-items: center;
  padding: 4px 8px;
  border-radius: 4px;
  font-size: 12px;
  font-weight: 500;
  line-height: 16px;
  text-transform: uppercase;
  letter-spacing: 0.05em;
}

.card-badge-success {
  background-color: #d1fae5;
  color: #065f46;
}

.card-badge-warning {
  background-color: #fef3c7;
  color: #92400e;
}

.card-badge-error {
  background-color: #fee2e2;
  color: #991b1b;
}

.card-badge-info {
  background-color: #dbeafe;
  color: #1e40af;
}

.card-badge-neutral {
  background-color: #f3f4f6;
  color: #374151;
}
```

## Metric Card Specifications

### Metric Display
```css
.metric-value {
  font-size: 32px;
  font-weight: 700;
  color: #1f2937;
  line-height: 40px;
  margin: 8px 0;
}

.metric-change {
  display: flex;
  align-items: center;
  gap: 4px;
  font-size: 14px;
  font-weight: 500;
  line-height: 20px;
}

.metric-change-positive {
  color: #059669;
}

.metric-change-negative {
  color: #dc2626;
}

.metric-change-neutral {
  color: #6b7280;
}

/* Dark mode */
.dark .metric-value { color: #f1f5f9; }
```

## Card Grid Layouts

### Responsive Grid
```css
.card-grid {
  display: grid;
  gap: 24px;
  grid-template-columns: repeat(auto-fit, minmax(320px, 1fr));
}

/* Metric cards grid */
.metric-grid {
  display: grid;
  gap: 24px;
  grid-template-columns: repeat(auto-fit, minmax(240px, 1fr));
}

/* Feature cards grid */
.feature-grid {
  display: grid;
  gap: 32px;
  grid-template-columns: repeat(auto-fit, minmax(280px, 1fr));
}

/* Mobile adjustments */
@media (max-width: 768px) {
  .card-grid,
  .metric-grid,
  .feature-grid {
    grid-template-columns: 1fr;
    gap: 16px;
  }
}
```

## Accessibility Specifications

### Focus Management
```css
/* Interactive cards */
.card-interactive:focus {
  outline: none;
  border-color: #059669;
  box-shadow: 0 0 0 3px rgba(5, 150, 105, 0.1);
}

/* Focus within cards */
.card:focus-within {
  border-color: #d1d5db;
}

/* High contrast mode */
@media (prefers-contrast: high) {
  .card {
    border: 2px solid #374151;
  }
  
  .card-interactive:focus {
    border: 2px solid #059669;
  }
}
```

### Screen Reader Support
```html
<!-- Interactive card with proper semantics -->
<article class="card card-interactive" role="button" tabindex="0" 
         aria-labelledby="project-title" aria-describedby="project-desc">
  <header class="card-header">
    <h3 id="project-title" class="card-title">Project Name</h3>
  </header>
  <div class="card-content">
    <p id="project-desc" class="card-description">Project description</p>
  </div>
</article>

<!-- Metric card -->
<div class="card" role="region" aria-labelledby="metric-title">
  <h3 id="metric-title" class="card-title">Total Deployments</h3>
  <div class="metric-value" aria-label="1,247 total deployments">1,247</div>
  <div class="metric-change metric-change-positive" aria-label="Increased by 12.5% from last month">
    +12.5% from last month
  </div>
</div>
```

## Usage Guidelines

### When to Use Cards

**Appropriate Uses**:
- Grouping related information and actions
- Displaying projects, items, or entities
- Creating scannable content layouts
- Showcasing features or capabilities
- Presenting metrics and key data

**Avoid Using Cards For**:
- Single pieces of information that don't need grouping
- Long-form content (use article layouts instead)
- Navigation elements (use proper navigation components)
- Simple lists (use list components instead)

### Content Guidelines

**Card Titles**:
- Keep concise (1-3 words when possible)
- Use sentence case, not title case
- Make descriptive and scannable

**Card Descriptions**:
- Limit to 2-3 lines for scanability
- Focus on key benefits or status
- Use plain language, avoid jargon

**Card Actions**:
- Limit to 2-3 primary actions per card
- Use clear, action-oriented labels
- Consider card-level vs. individual element actions

### Layout Best Practices

**Spacing**:
- Maintain consistent gaps in card grids (24px desktop, 16px mobile)
- Use proper internal padding (24px desktop, 16px mobile)
- Ensure adequate breathing room around content

**Hierarchy**:
- Establish clear visual hierarchy with typography
- Use icons consistently to enhance recognition
- Align related elements for easy scanning

**Responsive Design**:
- Cards should stack vertically on mobile
- Reduce padding and font sizes appropriately
- Maintain touch-friendly interactive areas (44px minimum)

## Implementation Examples

### React Card Component
```jsx
import React from 'react';
import { cn } from '@/lib/utils';

const Card = React.forwardRef(({ className, children, ...props }, ref) => (
  <div
    ref={ref}
    className={cn(
      "bg-white border border-neutral-200 rounded-lg shadow-sm p-6 transition-all duration-200",
      className
    )}
    {...props}
  >
    {children}
  </div>
));

const CardHeader = ({ className, children, ...props }) => (
  <div
    className={cn("mb-4 last:mb-0", className)}
    {...props}
  >
    {children}
  </div>
);

const CardTitle = ({ className, children, ...props }) => (
  <h3
    className={cn("text-lg font-semibold text-neutral-900 leading-6", className)}
    {...props}
  >
    {children}
  </h3>
);

const CardContent = ({ className, children, ...props }) => (
  <div
    className={cn("mb-4 last:mb-0", className)}
    {...props}
  >
    {children}
  </div>
);

const CardFooter = ({ className, children, ...props }) => (
  <div
    className={cn("pt-4 border-t border-neutral-100 mt-4", className)}
    {...props}
  >
    {children}
  </div>
);

export { Card, CardHeader, CardTitle, CardContent, CardFooter };
```

### Tailwind CSS Implementation
```html
<!-- Project Card -->
<div class="bg-white border border-neutral-200 rounded-lg shadow-sm p-6 hover:shadow-md hover:border-neutral-300 transition-all duration-200 cursor-pointer">
  <div class="flex items-start justify-between mb-4">
    <div class="flex items-center space-x-3">
      <div class="flex items-center justify-center w-10 h-10 bg-primary-50 rounded-lg">
        <svg class="w-6 h-6 text-primary-600" fill="currentColor">
          <!-- Icon -->
        </svg>
      </div>
      <div>
        <h3 class="text-lg font-semibold text-neutral-900">Project Name</h3>
        <p class="text-sm text-neutral-500">Updated 2 hours ago</p>
      </div>
    </div>
  </div>
  
  <div class="mb-4">
    <p class="text-base text-neutral-600">Project description with current status and key information.</p>
  </div>
  
  <div class="pt-4 border-t border-neutral-100">
    <div class="flex items-center justify-between">
      <div class="flex items-center space-x-4">
        <span class="inline-flex items-center px-2 py-1 rounded text-xs font-medium uppercase tracking-wide bg-success-100 text-success-800">Active</span>
        <span class="text-sm text-neutral-400">5 deployments</span>
      </div>
      <div class="flex space-x-2">
        <button class="btn-secondary btn-sm">Settings</button>
        <button class="btn-primary btn-sm">Deploy</button>
      </div>
    </div>
  </div>
</div>
```

## Testing Checklist

### Visual Testing
- [ ] Cards display correctly in grid layouts
- [ ] Shadows and borders render consistently
- [ ] Typography hierarchy is clear and readable
- [ ] Interactive states (hover, focus, active) work properly
- [ ] Status badges and icons align correctly

### Responsive Testing
- [ ] Cards stack properly on mobile devices
- [ ] Padding and spacing adapt correctly
- [ ] Touch targets meet minimum size requirements
- [ ] Text remains readable at all screen sizes

### Accessibility Testing
- [ ] Cards are keyboard navigable when interactive
- [ ] Focus indicators are visible and consistent
- [ ] Screen readers announce card content properly
- [ ] Color contrast meets WCAG AA standards
- [ ] Semantic HTML is used appropriately

### Performance Testing
- [ ] Large card grids scroll smoothly
- [ ] Animations don't cause layout thrashing
- [ ] Images and media load efficiently
- [ ] Cards render quickly in large datasets

---

*These card specifications provide the foundation for consistent, accessible content containers throughout CloudBox. Proper implementation ensures optimal user experience and maintainable code.*