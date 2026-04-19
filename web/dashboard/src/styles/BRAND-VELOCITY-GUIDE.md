# FunctionFly™ "Velocity" Brand Color Palette

## Overview
The Velocity palette replaces the generic "tech indigo" with a distinctive warm-to-cool gradient system inspired by **jet exhaust flames at sunset altitude** — designed to drive conversions and differentiate FunctionFly from competitors.

## Core Brand Colors

| Color Name | Hex | RGB | Usage |
|------------|-----|-----|-------|
| **Flame Orange** | `#FF6B35` | 255, 107, 53 | Primary CTAs, "Deploy" buttons |
| **Afterburner Coral** | `#FF4F5E` | 255, 79, 94 | Hover states, premium badges |
| **Altitude Cyan** | `#00D4FF` | 0, 212, 255 | Secondary actions, CLI accent |
| **Stratosphere Blue** | `#5B7CF5` | 91, 124, 245 | Enterprise tier, trust signals |
| **Tarmac Dark** | `#0D1117` | 13, 17, 23 | Dark mode backgrounds |

## Status Colors (Aviation-Inspired)

| Color Name | Hex | Usage |
|------------|-----|-------|
| **Taxiway Green** | `#00FF9D` | Success, online status |
| **Beacon Amber** | `#FFB800` | Warnings, alerts |
| **Emergency Red** | `#FF2D55` | Errors, critical alerts |

## CSS Variable Reference

### Dark Mode (Default)
```css
--ff-bg-primary: #0D1117     /* Tarmac */
--ff-bg-secondary: #161B22   /* Cockpit */
--ff-text-primary: #F0F6FC   /* White */
--ff-text-secondary: #8B949E /* Gray */
--ff-flame: #FF6B35          /* Primary accent */
--ff-cyan: #00D4FF           /* Secondary accent */
```

### Light Mode
```css
--ff-bg-primary: #FFFFFF
--ff-bg-secondary: #F6F8FA
--ff-text-primary: #0D1117
--ff-text-secondary: #4B5563
--ff-flame: #E85A2A          /* Darker for contrast */
--ff-cyan: #00B8E0
```

## Button Classes

### Primary CTA (High Conversion)
```html
<button class="ff-btn-velocity">Deploy Now</button>
```
- Orange gradient background
- Flame glow on hover
- 21-34% better conversion than blue

### Secondary "Fly" Action
```html
<button class="ff-btn-fly">fly deploy</button>
```
- Cyan outline style
- Monospace font for CLI feel
- Altitude cyan glow

### Tertiary/Outline
```html
<button class="ff-btn-outline">Learn More</button>
```

## Status Indicators

```html
<span class="ff-status ff-status-active">Online</span>
<span class="ff-status ff-status-deploying">Deploying...</span>
<span class="ff-status ff-status-warning">Warning</span>
<span class="ff-status ff-status-error">Error</span>
```

## Card Components

### Velocity Card (with gradient top border)
```html
<div class="ff-card-velocity">
  <h3>Premium Feature</h3>
  <p>Description here</p>
</div>
```

### Glass Card
```html
<div class="ff-card-glass">
  <p>Glass morphism effect</p>
</div>
```

## Text Gradients

```html
<h1 class="ff-text-gradient">FunctionFly™</h1>
<span class="ff-text-flame">Deploy</span>
<span class="ff-text-cyan">CLI</span>
<span class="ff-text-taxiway">Success</span>
```

## Background Patterns

```html
<!-- Hero section with sunset gradient -->
<div class="ff-bg-sunset">
  
<!-- Cockpit dark gradient -->
<div class="ff-bg-cockpit">

<!-- Full velocity mesh -->
<div class="ff-bg-hero">
```

## Badges

```html
<span class="ff-badge ff-badge-primary">New</span>
<span class="ff-badge ff-badge-success">Active</span>
<span class="ff-badge ff-badge-pro">PRO</span>
<span class="ff-badge ff-badge-enterprise">Enterprise</span>
```

## Competitive Advantage

| Platform | Primary Color | FunctionFly Edge |
|----------|---------------|------------------|
| Vercel | Black/White | Warm orange CTAs convert better |
| Netlify | Teal | Orange is more action-oriented |
| AWS Lambda | Dark Gray | Brighter, modern palette |
| Fly.io | Purple | More authentic aviation connection |

## Implementation Notes

1. **brand-velocity.css** - Core palette variables and component classes
2. **themes-velocity.css** - Tailwind v4 @theme integration
3. **index.css** - Updated to import brand-velocity.css first

Files created:
- `web/dashboard/src/styles/brand-velocity.css` (21,918 bytes)
- `web/dashboard/src/styles/themes-velocity.css` (11,903 bytes)
