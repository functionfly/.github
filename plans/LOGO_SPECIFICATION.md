# FunctionFly Logo & Icon Specification

## Brand Identity

### Core Concept
FunctionFly represents **speed**, **resilience**, and **multi-cloud freedom**. The brand embodies the idea of functions that effortlessly fly across cloud providers, maintaining high availability and performance.

### Brand Keywords
- **Velocity**: Speed of execution, rapid failover
- **Elevation**: Rising above single-cloud limitations
- **Resilience**: Unbreakable, always-on reliability
- **Edge**: Cutting-edge technology, edge computing
- **Freedom**: Multi-cloud portability, no vendor lock-in

---

## Logo Anatomy

### Primary Logo (Horizontal)
```
┌─────────────────────────────────────────────────────────────┐
│  [SYMBOL]  FunctionFly                                      │
│   48px      Wordmark                                        │
│                                                             │
│  ┌──────┐                                                   │
│  │  ⚡  │  FunctionFly                                      │
│  │  ✈   │                                                   │
│  └──────┘                                                   │
└─────────────────────────────────────────────────────────────┘
```

### Symbol Mark (Icon Only)
The FunctionFly symbol combines three visual metaphors:
1. **Lightning Bolt (⚡)** - Represents speed, execution, functions
2. **Flight/Wings (✈)** - Represents flying across clouds, freedom, elevation
3. **Hexagon Base** - Represents structure, reliability, cloud infrastructure

```
      Symbol Construction Grid
      
      ┌─────────────────────────────┐
      │     ◇  Wing Tip             │
      │    /  \                     │
      │   / ⚡ \   Lightning Core   │
      │  │  ✈  │   Flight Path      │
      │   \   /                     │
      │    \ /                      │
      │     ◇  Wing Tip             │
      └─────────────────────────────┘
      
      Grid: 24x24px base unit
      Safe Zone: 4px padding on all sides
```

### Wordmark Typography
- **Font**: Inter Bold (weight 700) or Cal Sans SemiBold
- **Letter-spacing**: -0.02em (tight tracking)
- **Case**: PascalCase (FunctionFly)
- **Styling**: Clean, no italics, no shadows

---

## Color System

### Primary Brand Colors

```css
/* Core Brand Gradient */
--brand-gradient-start: #6366f1;    /* Indigo 500 */
--brand-gradient-mid: #8b5cf6;      /* Violet 500 */
--brand-gradient-end: #d946ef;      /* Fuchsia 500 */

/* Primary Logo Colors */
--logo-primary: linear-gradient(135deg, #6366f1 0%, #8b5cf6 50%, #d946ef 100%);
--logo-solid: #6366f1;              /* Fallback solid color */

/* Symbol Gradient (3-tone) */
--symbol-top: #818cf8;              /* Indigo 400 */
--symbol-mid: #a78bfa;              /* Violet 400 */
--symbol-bottom: #e879f9;           /* Fuchsia 400 */
```

### Logo Color Variations

#### 1. Full Color (Primary)
- Symbol: Multi-gradient (indigo → violet → fuchsia)
- Wordmark: Solid indigo (#6366f1) or gradient
- Usage: Default for all light backgrounds

#### 2. Dark Mode
- Symbol: Brighter gradient (#818cf8 → #c084fc → #f0abfc)
- Wordmark: White (#ffffff) or light gradient
- Usage: Dark backgrounds, night mode

#### 3. Monochrome Light
- Symbol: White (#ffffff)
- Wordmark: White (#ffffff)
- Usage: Dark backgrounds, colored surfaces

#### 4. Monochrome Dark
- Symbol: Slate 900 (#0f172a)
- Wordmark: Slate 900 (#0f172a)
- Usage: Light backgrounds, print

#### 5. Single Color
- Symbol & Wordmark: Brand Indigo (#6366f1)
- Usage: Limited color environments, favicons

### Background Color Pairings

| Background Type | Logo Version | Contrast Ratio |
|----------------|--------------|----------------|
| White (#fff) | Full Color | 4.5:1 |
| Light Gray (#f8fafc) | Full Color | 4.5:1 |
| Dark (#0a0a0f) | Monochrome Light | 15:1 |
| Brand Gradient | Monochrome Light | 8:1 |
| Colored Surface | Monochrome Light/Dark | 4.5:1+ |

---

## Usage Guidelines

### Clearspace

Minimum clearspace around logo: **equal to symbol height**

```
┌──────────────────────────────────────────────────────────┐
│                                                          │
│    ┌────┐                                                │
│    │    │  FunctionFly                                   │
│    │ ⚡ │                                                │
│    └────┘                                                │
│                                                          │
│    │←── Symbol Height ──→│                              │
│    │    = Clearspace     │                              │
│                                                          │
└──────────────────────────────────────────────────────────┘
```

### Minimum Size

| Format | Minimum Width | Minimum Height |
|--------|---------------|----------------|
| Print | 25mm / 1 inch | 8mm |
| Digital | 100px | 32px |
| Favicon | 16px | 16px |
| App Icon | 60px @2x | 60px @2x |

### Proportions

- Symbol to Wordmark spacing: 12px (desktop), 8px (mobile)
- Logo lockup ratio: Symbol height = 1.2x cap height of wordmark
- Never stretch, compress, or rotate the logo

### Incorrect Usage (Don'ts)

❌ **Never:**
- Change logo colors outside brand palette
- Stretch or distort proportions
- Add drop shadows or effects
- Place on busy/patterned backgrounds
- Use low contrast combinations
- Rotate or tilt the logo
- Change wordmark font
- Crop the symbol

---

## Icon System

### Favicon Suite

| Size | Format | Usage |
|------|--------|-------|
| 16x16 | .ico, .png | Browser tabs, bookmarks |
| 32x32 | .png | Retina displays |
| 48x48 | .png | Windows taskbar |
| 180x180 | .png | Apple touch icon |

### App Icons

| Platform | Size | Corner Radius | Format |
|----------|------|---------------|--------|
| iOS | 1024x1024 | 180px | .png |
| Android | 512x512 | Adaptive | .png, .svg |
| macOS | 1024x1024 | 180px | .png |
| Windows | 256x256 | Square | .png, .ico |

### Social Media Avatars

| Platform | Size | Format |
|----------|------|--------|
| Twitter/X | 400x400 | .png, .jpg |
| LinkedIn | 300x300 | .png, .jpg |
| GitHub | 420x420 | .png |
| Discord | 128x128 | .png, .jpg |

### SVG Icons (UI Components)

```
Icon Sizes for UI:
┌─────────────┬──────────┬──────────────┐
│   Size      │  Usage   │   Stroke     │
├─────────────┼──────────┼──────────────┤
│   16px      │  Inline  │   1.5px      │
│   20px      │  Buttons │   1.5px      │
│   24px      │  Default │   2px        │
│   32px      │  Feature │   2px        │
│   48px      │  Hero    │   2.5px      │
└─────────────┴──────────┴──────────────┘
```

---

## File Formats & Delivery

### Export Formats

| Format | Extension | Usage |
|--------|-----------|-------|
| SVG | .svg | Web, scalable, preferred |
| PNG | .png | Raster fallback, transparency |
| WebP | .webp | Modern web, compression |
| PDF | .pdf | Print, vector |
| EPS | .eps | Legacy print |

### File Naming Convention

```
functionfly-logo-[variant]-[color]-[size].[ext]

Examples:
- functionfly-logo-horizontal-fullcolor.svg
- functionfly-logo-symbol-white.svg
- functionfly-logo-wordmark-indigo.svg
- functionfly-icon-16x16.png
- functionfly-appicon-ios-1024.png
```

---

## Implementation Notes

### CSS Variables for Web

```css
:root {
  /* Logo Colors */
  --logo-gradient: linear-gradient(135deg, #6366f1 0%, #8b5cf6 50%, #d946ef 100%);
  --logo-indigo: #6366f1;
  --logo-violet: #8b5cf6;
  --logo-fuchsia: #d946ef;
  
  /* Logo Sizes */
  --logo-height-sm: 24px;
  --logo-height-md: 32px;
  --logo-height-lg: 48px;
  
  /* Spacing */
  --logo-spacing: 12px;
  --logo-clearspace: 32px;
}
```

### Animation Guidelines

When animating the logo:
- Fade in: 300ms ease-out
- Scale: 200ms cubic-bezier(0.4, 0, 0.2, 1)
- Gradient shift: 8s infinite loop (subtle)
- Never rotate or distort

---

## Summary

The FunctionFly logo combines the concepts of **speed** (lightning) and **freedom** (flight) into a cohesive symbol that represents our multi-cloud serverless platform. The gradient from indigo to fuchsia represents innovation and creativity, while the geometric base conveys reliability and structure.

**Key Takeaways:**
- Use gradient version on light backgrounds
- Use monochrome white on dark backgrounds
- Maintain clearspace equal to symbol height
- Never distort or modify the logo
- Always use SVG when possible for crisp rendering
