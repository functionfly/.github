# FunctionFly Logo Assets

This directory contains all logo and brand assets for FunctionFly.

## Directory Structure

```
assets/
├── logo/                    # Logo variations
│   ├── functionfly-logo-horizontal-fullcolor.svg
│   ├── functionfly-logo-horizontal-darkmode.svg
│   ├── functionfly-logo-symbol-fullcolor.svg
│   ├── functionfly-logo-symbol-darkmode.svg
│   ├── functionfly-logo-symbol-indigo.svg
│   ├── functionfly-logo-symbol-white.svg
│   ├── functionfly-logo-symbol-dark.svg
│   ├── functionfly-logo-wordmark-indigo.svg
│   └── logo-styles.css
├── icons/                   # App icons
│   ├── app-icon-1024.svg
│   └── social-avatar.svg
└── favicons/               # Favicon suite
    ├── favicon-16x16.svg
    ├── favicon-32x32.svg
    └── apple-touch-icon.svg
```

## Logo Variations

### Full Color (Primary)
Use on light backgrounds (white, light gray).
- [`functionfly-logo-horizontal-fullcolor.svg`](logo/functionfly-logo-horizontal-fullcolor.svg) - Horizontal with wordmark
- [`functionfly-logo-symbol-fullcolor.svg`](logo/functionfly-logo-symbol-fullcolor.svg) - Symbol only

### Dark Mode
Use on dark backgrounds.
- [`functionfly-logo-horizontal-darkmode.svg`](logo/functionfly-logo-horizontal-darkmode.svg) - Horizontal with wordmark (white text)
- [`functionfly-logo-symbol-darkmode.svg`](logo/functionfly-logo-symbol-darkmode.svg) - Symbol only (brighter gradient)

### Monochrome
Use for limited color environments or print.
- [`functionfly-logo-symbol-white.svg`](logo/functionfly-logo-symbol-white.svg) - White version for dark backgrounds
- [`functionfly-logo-symbol-dark.svg`](logo/functionfly-logo-symbol-dark.svg) - Slate 900 for light backgrounds/print
- [`functionfly-logo-symbol-indigo.svg`](logo/functionfly-logo-symbol-indigo.svg) - Single color indigo
- [`functionfly-logo-wordmark-indigo.svg`](logo/functionfly-logo-wordmark-indigo.svg) - Wordmark only

## App Icons

### iOS/macOS (1024x1024)
- [`app-icon-1024.svg`](icons/app-icon-1024.svg) - App Store, iOS home screen, macOS

### Social Media (400x400)
- [`social-avatar.svg`](icons/social-avatar.svg) - Twitter/X, LinkedIn, GitHub, Discord

## Favicons

- [`favicon-16x16.svg`](favicons/favicon-16x16.svg) - Browser tabs, bookmarks
- [`favicon-32x32.svg`](favicons/favicon-32x32.svg) - Retina displays
- [`apple-touch-icon.svg`](favicons/apple-touch-icon.svg) - Apple touch icon (180x180)

## CSS Integration

Import the logo styles in your project:

```css
@import url('assets/logo/logo-styles.css');
```

Or use the CSS variables directly:

```css
:root {
  --logo-gradient: linear-gradient(135deg, #6366f1 0%, #8b5cf6 50%, #d946ef 100%);
  --logo-indigo: #6366f1;
  --logo-height-lg: 48px;
}
```

## Usage Guidelines

### Clearspace
Maintain clearspace equal to symbol height around the logo.

### Minimum Size
- Digital: 100px width minimum
- Print: 25mm / 1 inch width minimum
- Favicon: 16x16px

### Color Usage
| Background | Logo Version |
|------------|--------------|
| White/Light | Full Color |
| Dark | Dark Mode or White Monochrome |
| Brand Gradient | White Monochrome |
| Limited Color | Indigo Single Color |

### Don'ts
❌ Never:
- Change logo colors outside brand palette
- Stretch or distort proportions
- Add drop shadows or effects
- Place on busy/patterned backgrounds
- Use low contrast combinations
- Rotate or tilt the logo
- Change wordmark font
- Crop the symbol

## Brand Colors

| Name | Hex | Usage |
|------|-----|-------|
| Indigo 500 | `#6366f1` | Primary brand color |
| Violet 500 | `#8b5cf6` | Gradient mid-point |
| Fuchsia 500 | `#d946ef` | Gradient end |
| Indigo 400 | `#818cf8` | Symbol top |
| Violet 400 | `#a78bfa` | Symbol mid |
| Fuchsia 400 | `#e879f9` | Symbol bottom |
| Amber 400 | `#fbbf24` | Lightning bolt |
| Slate 900 | `#0f172a` | Dark monochrome |

## License

These brand assets are proprietary to FunctionFly and should only be used in accordance with the brand guidelines.
