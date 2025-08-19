---
title: Navigation Components - CloudBox Design System
description: Navigation patterns including sidebar, breadcrumbs, and menu components
feature: design-system
last-updated: 2025-08-19
version: 1.0.0
related-files:
  - ../tokens/colors.md
  - ../style-guide.md
  - buttons.md
status: approved
---

# Navigation Components

## Overview
Navigation components provide clear, consistent wayfinding throughout CloudBox. Inspired by Supabase's clean navigation patterns with enhanced visual hierarchy and accessibility features.

## Sidebar Navigation

### Primary Sidebar
**Main Navigation**: Primary navigation for dashboard and core features

**Visual Specifications**:
- **Width**: `256px` (expanded), `64px` (collapsed)
- **Background**: `#ffffff` (light), `#1e293b` (dark)
- **Border**: `1px solid #e5e7eb` (right border)
- **Height**: `100vh` (full viewport height)
- **Z-index**: `40` (above content, below modals)

```css
.sidebar {
  width: 256px;
  height: 100vh;
  background-color: #ffffff;
  border-right: 1px solid #e5e7eb;
  position: fixed;
  left: 0;
  top: 0;
  z-index: 40;
  transition: all 250ms ease-out;
  overflow-y: auto;
}

.sidebar.collapsed {
  width: 64px;
}

/* Dark mode */
.dark .sidebar {
  background-color: #1e293b;
  border-color: #475569;
}

/* Mobile */
@media (max-width: 768px) {
  .sidebar {
    width: 280px;
    transform: translateX(-100%);
  }
  
  .sidebar.open {
    transform: translateX(0);
  }
}
```

### Sidebar Header
**Brand and Controls**: Logo area and collapse toggle

```html
<div class="sidebar-header">
  <div class="flex items-center justify-between px-4 py-3 border-b border-neutral-100">
    <div class="flex items-center space-x-3">
      <div class="sidebar-logo">
        <svg class="w-8 h-8 text-primary-600" viewBox="0 0 32 32">
          <!-- CloudBox logo -->
        </svg>
      </div>
      <div class="sidebar-brand">
        <h1 class="text-lg font-semibold text-neutral-900">CloudBox</h1>
      </div>
    </div>
    <button class="sidebar-toggle btn-icon" aria-label="Toggle sidebar">
      <svg class="w-5 h-5" fill="currentColor" viewBox="0 0 20 20">
        <!-- Collapse icon -->
      </svg>
    </button>
  </div>
</div>
```

```css
.sidebar-header {
  flex-shrink: 0;
  border-bottom: 1px solid #f3f4f6;
}

.sidebar-logo {
  flex-shrink: 0;
}

.sidebar-brand {
  opacity: 1;
  transition: opacity 200ms ease-out;
}

.sidebar.collapsed .sidebar-brand {
  opacity: 0;
  width: 0;
  overflow: hidden;
}

.sidebar-toggle {
  opacity: 1;
  transition: opacity 200ms ease-out;
}

.sidebar.collapsed .sidebar-toggle {
  opacity: 0;
  pointer-events: none;
}
```

### Navigation Menu
**Menu Structure**: Hierarchical navigation with sections and items

```html
<nav class="sidebar-nav">
  <div class="sidebar-section">
    <div class="sidebar-section-header">
      <h2 class="sidebar-section-title">Dashboard</h2>
    </div>
    <ul class="sidebar-menu">
      <li>
        <a href="/dashboard" class="sidebar-item active">
          <div class="sidebar-item-icon">
            <svg class="w-5 h-5" fill="currentColor">
              <!-- Dashboard icon -->
            </svg>
          </div>
          <span class="sidebar-item-text">Overview</span>
        </a>
      </li>
      <li>
        <a href="/projects" class="sidebar-item">
          <div class="sidebar-item-icon">
            <svg class="w-5 h-5" fill="currentColor">
              <!-- Projects icon -->
            </svg>
          </div>
          <span class="sidebar-item-text">Projects</span>
          <span class="sidebar-item-badge">3</span>
        </a>
      </li>
    </ul>
  </div>
  
  <div class="sidebar-section">
    <div class="sidebar-section-header">
      <h2 class="sidebar-section-title">Management</h2>
    </div>
    <ul class="sidebar-menu">
      <li>
        <a href="/settings" class="sidebar-item">
          <div class="sidebar-item-icon">
            <svg class="w-5 h-5" fill="currentColor">
              <!-- Settings icon -->
            </svg>
          </div>
          <span class="sidebar-item-text">Settings</span>
        </a>
      </li>
    </ul>
  </div>
</nav>
```

```css
.sidebar-nav {
  flex: 1;
  padding: 16px 0;
  overflow-y: auto;
}

.sidebar-section {
  margin-bottom: 24px;
}

.sidebar-section:last-child {
  margin-bottom: 0;
}

.sidebar-section-header {
  padding: 0 16px 8px;
}

.sidebar-section-title {
  font-size: 12px;
  font-weight: 600;
  color: #9ca3af;
  text-transform: uppercase;
  letter-spacing: 0.05em;
  line-height: 16px;
  opacity: 1;
  transition: opacity 200ms ease-out;
}

.sidebar.collapsed .sidebar-section-title {
  opacity: 0;
}

.sidebar-menu {
  list-style: none;
  margin: 0;
  padding: 0;
}

.sidebar-item {
  display: flex;
  align-items: center;
  width: 100%;
  padding: 8px 16px;
  color: #6b7280;
  text-decoration: none;
  border-radius: 6px;
  margin: 0 8px;
  transition: all 150ms ease-out;
  position: relative;
}

.sidebar-item:hover {
  background-color: #f3f4f6;
  color: #374151;
}

.sidebar-item.active {
  background-color: #ecfdf5;
  color: #059669;
  font-weight: 500;
}

.sidebar-item.active::before {
  content: '';
  position: absolute;
  left: 0;
  top: 0;
  bottom: 0;
  width: 3px;
  background-color: #059669;
  border-radius: 0 2px 2px 0;
}

.sidebar-item-icon {
  flex-shrink: 0;
  width: 20px;
  height: 20px;
  margin-right: 12px;
}

.sidebar-item-text {
  flex: 1;
  font-size: 14px;
  font-weight: 400;
  line-height: 20px;
  opacity: 1;
  transition: opacity 200ms ease-out;
}

.sidebar.collapsed .sidebar-item-text {
  opacity: 0;
  width: 0;
  overflow: hidden;
}

.sidebar-item-badge {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  height: 18px;
  min-width: 18px;
  padding: 0 6px;
  background-color: #e5e7eb;
  color: #6b7280;
  border-radius: 9px;
  font-size: 11px;
  font-weight: 500;
  line-height: 1;
  opacity: 1;
  transition: opacity 200ms ease-out;
}

.sidebar.collapsed .sidebar-item-badge {
  opacity: 0;
}

.sidebar-item.active .sidebar-item-badge {
  background-color: #d1fae5;
  color: #065f46;
}

/* Dark mode */
.dark .sidebar-section-title { color: #94a3b8; }
.dark .sidebar-item { color: #cbd5e1; }
.dark .sidebar-item:hover {
  background-color: #334155;
  color: #f1f5f9;
}
.dark .sidebar-item.active {
  background-color: #0f2419;
  color: #22c55e;
}
```

### Sidebar Footer
**User Account**: User profile and account actions

```html
<div class="sidebar-footer">
  <div class="sidebar-user">
    <button class="sidebar-user-button">
      <div class="sidebar-user-avatar">
        <img src="/avatar.jpg" alt="User avatar" class="w-8 h-8 rounded-full" />
      </div>
      <div class="sidebar-user-info">
        <div class="sidebar-user-name">John Doe</div>
        <div class="sidebar-user-email">john@example.com</div>
      </div>
      <div class="sidebar-user-action">
        <svg class="w-4 h-4" fill="currentColor" viewBox="0 0 20 20">
          <!-- Chevron down icon -->
        </svg>
      </div>
    </button>
  </div>
</div>
```

```css
.sidebar-footer {
  flex-shrink: 0;
  border-top: 1px solid #f3f4f6;
  padding: 16px;
}

.sidebar-user-button {
  display: flex;
  align-items: center;
  width: 100%;
  padding: 8px;
  background: transparent;
  border: none;
  border-radius: 6px;
  text-align: left;
  cursor: pointer;
  transition: all 150ms ease-out;
}

.sidebar-user-button:hover {
  background-color: #f3f4f6;
}

.sidebar-user-avatar {
  flex-shrink: 0;
  margin-right: 12px;
}

.sidebar-user-info {
  flex: 1;
  opacity: 1;
  transition: opacity 200ms ease-out;
}

.sidebar.collapsed .sidebar-user-info {
  opacity: 0;
  width: 0;
  overflow: hidden;
}

.sidebar-user-name {
  font-size: 14px;
  font-weight: 500;
  color: #1f2937;
  line-height: 20px;
}

.sidebar-user-email {
  font-size: 12px;
  color: #6b7280;
  line-height: 16px;
}

.sidebar-user-action {
  flex-shrink: 0;
  color: #9ca3af;
  opacity: 1;
  transition: opacity 200ms ease-out;
}

.sidebar.collapsed .sidebar-user-action {
  opacity: 0;
}
```

## Breadcrumb Navigation

### Standard Breadcrumbs
**Page Hierarchy**: Show current location in site hierarchy

```html
<nav class="breadcrumb" aria-label="Breadcrumb">
  <ol class="breadcrumb-list">
    <li class="breadcrumb-item">
      <a href="/dashboard" class="breadcrumb-link">
        <svg class="w-4 h-4 mr-1" fill="currentColor">
          <!-- Home icon -->
        </svg>
        Dashboard
      </a>
    </li>
    <li class="breadcrumb-separator" aria-hidden="true">
      <svg class="w-4 h-4" fill="currentColor" viewBox="0 0 20 20">
        <path fill-rule="evenodd" d="M7.293 14.707a1 1 0 010-1.414L10.586 10 7.293 6.707a1 1 0 111.414-1.414l4 4a1 1 0 010 1.414l-4 4a1 1 0 01-1.414 0z" clip-rule="evenodd" />
      </svg>
    </li>
    <li class="breadcrumb-item">
      <a href="/projects" class="breadcrumb-link">Projects</a>
    </li>
    <li class="breadcrumb-separator" aria-hidden="true">
      <svg class="w-4 h-4" fill="currentColor" viewBox="0 0 20 20">
        <path fill-rule="evenodd" d="M7.293 14.707a1 1 0 010-1.414L10.586 10 7.293 6.707a1 1 0 111.414-1.414l4 4a1 1 0 010 1.414l-4 4a1 1 0 01-1.414 0z" clip-rule="evenodd" />
      </svg>
    </li>
    <li class="breadcrumb-item breadcrumb-current" aria-current="page">
      <span>My Project</span>
    </li>
  </ol>
</nav>
```

```css
.breadcrumb {
  padding: 12px 0;
}

.breadcrumb-list {
  display: flex;
  align-items: center;
  list-style: none;
  margin: 0;
  padding: 0;
  flex-wrap: wrap;
}

.breadcrumb-item {
  display: flex;
  align-items: center;
}

.breadcrumb-link {
  display: flex;
  align-items: center;
  padding: 4px 8px;
  color: #6b7280;
  text-decoration: none;
  font-size: 14px;
  font-weight: 400;
  border-radius: 4px;
  transition: all 150ms ease-out;
}

.breadcrumb-link:hover {
  color: #059669;
  background-color: #f0fdf4;
}

.breadcrumb-separator {
  display: flex;
  align-items: center;
  color: #d1d5db;
  margin: 0 4px;
}

.breadcrumb-current {
  font-size: 14px;
  font-weight: 500;
  color: #1f2937;
  padding: 4px 8px;
}

/* Mobile responsive */
@media (max-width: 640px) {
  .breadcrumb-item:not(:last-child):not(:nth-last-child(2)) {
    display: none;
  }
  
  .breadcrumb-separator:not(:nth-last-child(2)) {
    display: none;
  }
  
  .breadcrumb-item:nth-last-child(3) {
    display: flex;
  }
  
  .breadcrumb-item:nth-last-child(3) .breadcrumb-link::before {
    content: '...';
    margin-right: 4px;
  }
}
```

## Top Navigation Bar

### Main Header
**Global Actions**: Search, notifications, user menu

```html
<header class="top-nav">
  <div class="top-nav-container">
    <!-- Mobile menu button -->
    <button class="mobile-menu-button md:hidden" aria-label="Open menu">
      <svg class="w-6 h-6" fill="currentColor" viewBox="0 0 20 20">
        <!-- Hamburger icon -->
      </svg>
    </button>
    
    <!-- Search -->
    <div class="top-nav-search">
      <div class="search-container">
        <div class="search-icon">
          <svg class="w-5 h-5" fill="currentColor" viewBox="0 0 20 20">
            <!-- Search icon -->
          </svg>
        </div>
        <input 
          type="text" 
          placeholder="Search projects, deployments..."
          class="search-input"
          aria-label="Search"
        />
        <kbd class="search-shortcut">⌘K</kbd>
      </div>
    </div>
    
    <!-- Actions -->
    <div class="top-nav-actions">
      <!-- Notifications -->
      <button class="top-nav-action" aria-label="Notifications">
        <svg class="w-5 h-5" fill="currentColor" viewBox="0 0 20 20">
          <!-- Bell icon -->
        </svg>
        <span class="notification-badge">3</span>
      </button>
      
      <!-- Help -->
      <button class="top-nav-action" aria-label="Help">
        <svg class="w-5 h-5" fill="currentColor" viewBox="0 0 20 20">
          <!-- Help icon -->
        </svg>
      </button>
      
      <!-- User menu -->
      <div class="user-menu">
        <button class="user-menu-button" aria-label="User menu">
          <img src="/avatar.jpg" alt="User" class="w-8 h-8 rounded-full" />
        </button>
      </div>
    </div>
  </div>
</header>
```

```css
.top-nav {
  height: 64px;
  background-color: #ffffff;
  border-bottom: 1px solid #e5e7eb;
  position: fixed;
  top: 0;
  left: 256px;
  right: 0;
  z-index: 30;
  transition: left 250ms ease-out;
}

.sidebar.collapsed + .top-nav {
  left: 64px;
}

.top-nav-container {
  display: flex;
  align-items: center;
  height: 100%;
  padding: 0 24px;
  gap: 16px;
}

.mobile-menu-button {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 40px;
  height: 40px;
  color: #6b7280;
  background: transparent;
  border: none;
  border-radius: 6px;
  cursor: pointer;
  transition: all 150ms ease-out;
}

.mobile-menu-button:hover {
  background-color: #f3f4f6;
  color: #374151;
}

.top-nav-search {
  flex: 1;
  max-width: 400px;
}

.search-container {
  position: relative;
  display: flex;
  align-items: center;
}

.search-icon {
  position: absolute;
  left: 12px;
  color: #9ca3af;
  pointer-events: none;
  z-index: 1;
}

.search-input {
  width: 100%;
  height: 40px;
  padding: 0 80px 0 40px;
  background-color: #f9fafb;
  border: 1px solid #e5e7eb;
  border-radius: 6px;
  font-size: 14px;
  color: #1f2937;
  placeholder-color: #9ca3af;
  transition: all 150ms ease-out;
}

.search-input:focus {
  outline: none;
  background-color: #ffffff;
  border-color: #059669;
  box-shadow: 0 0 0 3px rgba(5, 150, 105, 0.1);
}

.search-shortcut {
  position: absolute;
  right: 12px;
  padding: 2px 6px;
  background-color: #e5e7eb;
  color: #6b7280;
  border-radius: 3px;
  font-size: 11px;
  font-weight: 500;
  line-height: 1;
  pointer-events: none;
}

.top-nav-actions {
  display: flex;
  align-items: center;
  gap: 8px;
}

.top-nav-action {
  position: relative;
  display: flex;
  align-items: center;
  justify-content: center;
  width: 40px;
  height: 40px;
  color: #6b7280;
  background: transparent;
  border: none;
  border-radius: 6px;
  cursor: pointer;
  transition: all 150ms ease-out;
}

.top-nav-action:hover {
  background-color: #f3f4f6;
  color: #374151;
}

.notification-badge {
  position: absolute;
  top: 8px;
  right: 8px;
  display: flex;
  align-items: center;
  justify-content: center;
  width: 16px;
  height: 16px;
  background-color: #dc2626;
  color: #ffffff;
  border-radius: 8px;
  font-size: 10px;
  font-weight: 600;
  line-height: 1;
}

.user-menu-button {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 40px;
  height: 40px;
  background: transparent;
  border: none;
  border-radius: 6px;
  cursor: pointer;
  transition: all 150ms ease-out;
}

.user-menu-button:hover {
  background-color: #f3f4f6;
}

/* Mobile responsive */
@media (max-width: 768px) {
  .top-nav {
    left: 0;
  }
  
  .top-nav-search {
    display: none;
  }
}
```

## Tab Navigation

### Standard Tabs
**Content Switching**: For switching between related content sections

```html
<div class="tabs">
  <div class="tab-list" role="tablist">
    <button class="tab-trigger active" role="tab" aria-selected="true" aria-controls="overview-panel">
      Overview
    </button>
    <button class="tab-trigger" role="tab" aria-selected="false" aria-controls="deployments-panel">
      Deployments
      <span class="tab-badge">12</span>
    </button>
    <button class="tab-trigger" role="tab" aria-selected="false" aria-controls="settings-panel">
      Settings
    </button>
  </div>
  
  <div class="tab-content">
    <div class="tab-panel active" id="overview-panel" role="tabpanel">
      <!-- Overview content -->
    </div>
    <div class="tab-panel" id="deployments-panel" role="tabpanel">
      <!-- Deployments content -->
    </div>
    <div class="tab-panel" id="settings-panel" role="tabpanel">
      <!-- Settings content -->
    </div>
  </div>
</div>
```

```css
.tabs {
  width: 100%;
}

.tab-list {
  display: flex;
  align-items: center;
  border-bottom: 1px solid #e5e7eb;
  margin-bottom: 24px;
  overflow-x: auto;
}

.tab-trigger {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 12px 16px;
  background: transparent;
  border: none;
  border-bottom: 2px solid transparent;
  color: #6b7280;
  font-size: 14px;
  font-weight: 500;
  line-height: 20px;
  white-space: nowrap;
  cursor: pointer;
  transition: all 150ms ease-out;
}

.tab-trigger:hover {
  color: #374151;
  border-bottom-color: #d1d5db;
}

.tab-trigger.active {
  color: #059669;
  border-bottom-color: #059669;
}

.tab-trigger:focus {
  outline: none;
  background-color: #f9fafb;
}

.tab-badge {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  height: 18px;
  min-width: 18px;
  padding: 0 6px;
  background-color: #e5e7eb;
  color: #6b7280;
  border-radius: 9px;
  font-size: 11px;
  font-weight: 500;
  line-height: 1;
}

.tab-trigger.active .tab-badge {
  background-color: #d1fae5;
  color: #065f46;
}

.tab-content {
  position: relative;
}

.tab-panel {
  display: none;
}

.tab-panel.active {
  display: block;
}

/* Mobile responsive */
@media (max-width: 640px) {
  .tab-trigger {
    padding: 12px;
    font-size: 13px;
  }
  
  .tab-list {
    margin-bottom: 16px;
  }
}
```

## Accessibility Specifications

### Keyboard Navigation
```css
/* Focus management */
.sidebar-item:focus,
.breadcrumb-link:focus,
.tab-trigger:focus {
  outline: 2px solid #059669;
  outline-offset: 2px;
}

/* Skip link */
.skip-link {
  position: absolute;
  top: -40px;
  left: 6px;
  background: #059669;
  color: white;
  padding: 8px;
  text-decoration: none;
  border-radius: 4px;
  z-index: 100;
}

.skip-link:focus {
  top: 6px;
}
```

### ARIA Support
```html
<!-- Sidebar with proper semantics -->
<nav class="sidebar" aria-label="Main navigation">
  <div class="sidebar-header">
    <h1>CloudBox</h1>
    <button aria-label="Collapse sidebar" aria-expanded="true">
      <!-- Toggle icon -->
    </button>
  </div>
  
  <ul class="sidebar-menu" role="list">
    <li role="listitem">
      <a href="/dashboard" class="sidebar-item" aria-current="page">
        Overview
      </a>
    </li>
  </ul>
</nav>

<!-- Breadcrumbs with proper semantics -->
<nav aria-label="Breadcrumb">
  <ol class="breadcrumb-list">
    <li><a href="/">Home</a></li>
    <li aria-current="page">Current Page</li>
  </ol>
</nav>

<!-- Tabs with proper semantics -->
<div class="tabs">
  <div class="tab-list" role="tablist" aria-label="Content sections">
    <button role="tab" aria-selected="true" aria-controls="panel-1">
      Tab 1
    </button>
  </div>
  <div id="panel-1" role="tabpanel" aria-labelledby="tab-1">
    Panel content
  </div>
</div>
```

## Usage Guidelines

### Navigation Hierarchy

**Primary Navigation (Sidebar)**:
- Main sections and core functionality
- Always visible for orientation
- Limited to 7±2 main items for cognitive load
- Group related items into sections

**Secondary Navigation (Tabs)**:
- Related content within a section
- Context-specific views
- Maximum 5-6 tabs for usability
- Use badges for counts when relevant

**Tertiary Navigation (Breadcrumbs)**:
- Show current location in hierarchy
- Enable quick backtracking
- Always include home/dashboard link
- Truncate appropriately on mobile

### Content Organization

**Sidebar Sections**:
- Group by functional area (Dashboard, Management, etc.)
- Use clear, action-oriented labels
- Include visual indicators for active states
- Show item counts when relevant

**Information Architecture**:
- Maintain consistent navigation across all pages
- Use progressive disclosure for complex hierarchies
- Provide multiple ways to reach important content
- Include search for findability

### Responsive Behavior

**Mobile Strategy**:
- Sidebar converts to overlay on mobile
- Top navigation shows hamburger menu
- Breadcrumbs show only last 2 levels
- Tabs become horizontally scrollable

**Breakpoint Behaviors**:
- **Desktop** (1024px+): Full sidebar always visible
- **Tablet** (768px-1023px): Collapsible sidebar
- **Mobile** (<768px): Overlay sidebar, simplified navigation

## Implementation Examples

### React Navigation Components
```jsx
import React, { useState } from 'react';
import { cn } from '@/lib/utils';

const Sidebar = ({ collapsed, onToggle, children }) => {
  return (
    <nav className={cn(
      "sidebar",
      collapsed && "collapsed"
    )} aria-label="Main navigation">
      <div className="sidebar-header">
        <div className="flex items-center justify-between px-4 py-3">
          <div className="flex items-center space-x-3">
            <div className="sidebar-logo">
              {/* Logo component */}
            </div>
            <h1 className="sidebar-brand">CloudBox</h1>
          </div>
          <button 
            onClick={onToggle}
            className="sidebar-toggle"
            aria-label={collapsed ? 'Expand sidebar' : 'Collapse sidebar'}
            aria-expanded={!collapsed}
          >
            {/* Toggle icon */}
          </button>
        </div>
      </div>
      <div className="sidebar-nav">
        {children}
      </div>
    </nav>
  );
};

const SidebarItem = ({ href, icon, children, active, badge }) => {
  return (
    <a 
      href={href}
      className={cn(
        "sidebar-item",
        active && "active"
      )}
      aria-current={active ? 'page' : undefined}
    >
      <div className="sidebar-item-icon">{icon}</div>
      <span className="sidebar-item-text">{children}</span>
      {badge && <span className="sidebar-item-badge">{badge}</span>}
    </a>
  );
};

export { Sidebar, SidebarItem };
```

## Testing Checklist

### Visual Testing
- [ ] Navigation renders correctly in all states
- [ ] Active states are clearly visible
- [ ] Hover and focus states work properly
- [ ] Icons and text align correctly
- [ ] Responsive behavior works on all devices

### Accessibility Testing
- [ ] All navigation is keyboard accessible
- [ ] Focus indicators are visible and consistent
- [ ] Screen readers announce navigation properly
- [ ] ARIA attributes are correct and complete
- [ ] Tab order is logical and intuitive

### Functional Testing
- [ ] All navigation links work correctly
- [ ] Active states update properly
- [ ] Collapse/expand functionality works
- [ ] Search functionality is responsive
- [ ] Mobile menu toggles properly

### Performance Testing
- [ ] Navigation renders quickly
- [ ] Smooth animations on state changes
- [ ] No layout shift during loading
- [ ] Icons load efficiently

---

*These navigation specifications ensure consistent, accessible wayfinding throughout CloudBox. Proper implementation provides users with clear orientation and efficient task completion.*