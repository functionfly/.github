# FunctionFly UI Design Specification

## Overview

This document defines the complete UI architecture, design system, and component specifications for the FunctionFly dashboard - a stunning, modern multi-cloud failover management interface.

---

## Design Philosophy

**"Dark-First, Developer-Centric, Motion-Enhanced"**

- **Dark-first aesthetic**: Deep slate backgrounds with vibrant accent gradients
- **Developer-centric**: Information-dense layouts with clear data hierarchy
- **Motion-enhanced**: Subtle animations for state changes and interactions
- **Glassmorphism**: Frosted glass effects for cards and overlays
- **Gradient accents**: Dynamic gradients for CTAs and status indicators

---

## Color System

### Primary Palette

```css
/* Background Colors */
--bg-primary: #0a0a0f;        /* Deep void black */
--bg-secondary: #12121a;      /* Elevated surfaces */
--bg-tertiary: #1a1a25;       /* Cards, panels */
--bg-hover: #252535;          /* Hover states */

/* Text Colors */
--text-primary: #ffffff;      /* Primary text */
--text-secondary: #a0a0b0;    /* Secondary text */
--text-muted: #6b6b7b;        /* Tertiary/muted */
--text-accent: #6366f1;       /* Links, highlights */

/* Brand Gradients */
--gradient-primary: linear-gradient(135deg, #6366f1 0%, #8b5cf6 50%, #d946ef 100%);
--gradient-success: linear-gradient(135deg, #10b981 0%, #34d399 100%);
--gradient-warning: linear-gradient(135deg, #f59e0b 0%, #fbbf24 100%);
--gradient-danger: linear-gradient(135deg, #ef4444 0%, #f87171 100%);
--gradient-info: linear-gradient(135deg, #3b82f6 0%, #60a5fa 100%);

/* Provider Colors */
--cloudflare: #f48120;
--vercel: #000000;
--fly: #7b68ee;
--deno: #000000;
```

### Semantic Colors

```css
/* Status Colors */
--status-online: #10b981;
--status-offline: #ef4444;
--status-degraded: #f59e0b;
--status-pending: #6b7280;

/* Border Colors */
--border-subtle: rgba(255, 255, 255, 0.08);
--border-default: rgba(255, 255, 255, 0.12);
--border-focus: rgba(99, 102, 241, 0.5);
```

---

## Typography

### Font Stack

```css
--font-sans: 'Inter', system-ui, -apple-system, sans-serif;
--font-mono: 'JetBrains Mono', 'Fira Code', monospace;
--font-display: 'Cal Sans', 'Inter', sans-serif; /* For headings */
```

### Type Scale

| Token | Size | Line Height | Weight | Usage |
|-------|------|-------------|--------|-------|
| text-xs | 12px | 16px | 400 | Labels, badges |
| text-sm | 14px | 20px | 400 | Secondary text |
| text-base | 16px | 24px | 400 | Body text |
| text-lg | 18px | 28px | 500 | Lead paragraphs |
| text-xl | 20px | 28px | 600 | Small headings |
| text-2xl | 24px | 32px | 600 | Section headings |
| text-3xl | 30px | 36px | 700 | Page titles |
| text-4xl | 36px | 40px | 700 | Hero text |
| text-5xl | 48px | 48px | 800 | Display text |
| text-6xl | 60px | 60px | 800 | Hero display |

---

## Spacing System

```css
--space-1: 4px;
--space-2: 8px;
--space-3: 12px;
--space-4: 16px;
--space-5: 20px;
--space-6: 24px;
--space-8: 32px;
--space-10: 40px;
--space-12: 48px;
--space-16: 64px;
--space-20: 80px;
--space-24: 96px;
```

---

## Component Specifications

### Buttons

**Primary Button**
- Background: `var(--gradient-primary)`
- Text: white, font-weight 600
- Padding: 12px 24px
- Border-radius: 8px
- Hover: Scale 1.02, brightness 1.1
- Active: Scale 0.98

**Secondary Button**
- Background: transparent
- Border: 1px solid `var(--border-default)`
- Text: `var(--text-primary)`
- Hover: Background `var(--bg-hover)`

**Ghost Button**
- Background: transparent
- Text: `var(--text-secondary)`
- Hover: Text `var(--text-primary)`, bg `var(--bg-hover)`

### Cards

**Default Card**
- Background: `var(--bg-tertiary)`
- Border: 1px solid `var(--border-subtle)`
- Border-radius: 12px
- Padding: 24px
- Shadow: none (flat design)

**Glass Card**
- Background: rgba(26, 26, 37, 0.6)
- Backdrop-filter: blur(12px)
- Border: 1px solid `var(--border-subtle)`
- Border-radius: 16px

**Hover Card**
- Transition: all 200ms ease
- Hover: Border color `var(--border-default)`, translateY(-2px)

### Inputs

**Text Input**
- Background: `var(--bg-secondary)`
- Border: 1px solid `var(--border-subtle)`
- Border-radius: 8px
- Padding: 12px 16px
- Focus: Border `var(--border-focus)`, ring 2px `var(--border-focus)`

### Badges

**Status Badge**
- Online: Green dot + "Online" text
- Offline: Red dot + "Offline" text
- Degraded: Amber dot + "Degraded" text
- Dot size: 8px with pulse animation for online

---

## Layout Architecture

### Dashboard Layout

```
┌─────────────────────────────────────────────────────────┐
│  Sidebar    │  TopBar (fixed)                          │
│  (fixed)    ├──────────────────────────────────────────┤
│             │                                          │
│  Logo       │  Main Content Area (scrollable)          │
│             │                                          │
│  Nav        │  ┌────────────────────────────────────┐  │
│  - Dashboard│  │  Page Header                       │  │
│  - Functions│  ├────────────────────────────────────┤  │
│  - Providers│  │                                    │  │
│  - Analytics│  │  Content                           │  │
│  - Settings │  │                                    │  │
│             │  └────────────────────────────────────┘  │
│  User Menu  │                                          │
└─────────────┴──────────────────────────────────────────┘
```

### Sidebar Specifications

- Width: 260px (desktop), 280px (wide)
- Background: `var(--bg-secondary)`
- Border-right: 1px solid `var(--border-subtle)`
- Padding: 24px 16px
- Collapsible on mobile (drawer)

**Navigation Items**
- Height: 44px
- Border-radius: 8px
- Icon + Label layout
- Active: Background `var(--bg-hover)`, left border 3px gradient
- Hover: Background `var(--bg-hover)`

### TopBar Specifications

- Height: 64px
- Background: `var(--bg-primary)` with backdrop blur
- Border-bottom: 1px solid `var(--border-subtle)`
- Sticky positioning

**Elements**
- Breadcrumb navigation
- Global search (cmd+k shortcut)
- Notification bell
- User avatar dropdown

---

## Page Specifications

### 1. Landing Page

**Hero Section**
- Full viewport height (100vh)
- Centered content
- Animated gradient background (subtle mesh gradient)
- Main headline: "Multi-Cloud Failover for Indie SaaS"
- Subheadline: "Deploy once. Run everywhere. Never go down."
- CTA: "Get Started Free" (primary gradient button)
- Secondary CTA: "View Documentation" (ghost button)

**Features Grid**
- 3-column grid on desktop
- Feature cards with icon, title, description
- Icons: Zap, Globe, Shield

**Pricing Section**
- 2-tier pricing cards (Starter/Pro)
- Featured card with gradient border
- Feature checklists
- CTA buttons

### 2. Auth Pages

**Layout**
- Split screen: Form left, illustration right
- Centered card on mobile
- Max-width: 420px for form

**Login Form**
- Email input
- Password input with toggle visibility
- "Remember me" checkbox
- Submit button
- "Forgot password?" link
- Divider with "or continue with"
- OAuth buttons (GitHub, Google)

### 3. Dashboard Home

**Stats Row**
- 4 stat cards in a row
- Metric name, value, change indicator
- Sparkline mini-charts

**Metrics**
- Active Functions (count)
- Avg Latency (ms)
- Uptime %
- Requests This Month

**Recent Activity**
- Timeline of recent events
- Deployment events
- Failover events
- Health status changes

**Provider Status**
- Mini cards showing each provider
- Status indicator
- Last checked timestamp

### 4. Functions List

**Toolbar**
- Search input
- Filter dropdown (provider, status)
- "Deploy New" button (primary)

**Table**
- Columns: Name, Providers, Status, Last Deployed, Actions
- Row hover effect
- Status badges
- Action menu (edit, delete, view logs)

**Empty State**
- Illustration
- "No functions yet" message
- "Deploy your first function" CTA

### 5. Function Editor

**Layout**
- Two-column: Form left, Preview/Logs right

**Form Fields**
- Function name (text)
- Provider selection (multi-select chips)
- Region selection
- Code editor (Monaco/CodeMirror)
- Environment variables (key-value pairs)
- Secrets (masked inputs)

**Actions**
- "Deploy" button
- "Test" button (secondary)
- "Save Draft" button (ghost)

### 6. Providers Page

**Provider Cards**
- Provider logo/icon
- Connection status
- Region list
- "Connect" / "Disconnect" button
- "Configure" link

**Connection Flow**
- Modal with provider-specific instructions
- API key input
- Validation feedback
- Success confirmation

### 7. Analytics Page

**Chart Types**
- Line chart: Request volume over time
- Bar chart: Latency by provider
- Pie/donut: Traffic distribution
- Area chart: Error rate over time

**Filters**
- Date range picker
- Provider filter
- Function filter

**Metrics Grid**
- Total requests
- Average latency
- Error rate
- Cache hit rate

---

## Animation Specifications

### Page Transitions

```css
/* Fade + slide up */
.page-enter {
  opacity: 0;
  transform: translateY(20px);
}
.page-enter-active {
  opacity: 1;
  transform: translateY(0);
  transition: opacity 300ms, transform 300ms cubic-bezier(0.4, 0, 0.2, 1);
}
```

### Card Hover

```css
.card {
  transition: transform 200ms ease, border-color 200ms ease;
}
.card:hover {
  transform: translateY(-2px);
  border-color: var(--border-default);
}
```

### Button Interactions

```css
.button {
  transition: transform 150ms ease, filter 150ms ease;
}
.button:hover {
  transform: scale(1.02);
  filter: brightness(1.1);
}
.button:active {
  transform: scale(0.98);
}
```

### Loading States

**Skeleton Loading**
- Shimmer animation
- Background gradient sweep
- Duration: 1.5s infinite

**Spinner**
- Circular progress indicator
- Brand gradient stroke
- Size variants: sm (16px), md (24px), lg (32px)

---

## Responsive Breakpoints

```css
--breakpoint-sm: 640px;   /* Mobile landscape */
--breakpoint-md: 768px;   /* Tablet */
--breakpoint-lg: 1024px;  /* Desktop */
--breakpoint-xl: 1280px;  /* Wide desktop */
--breakpoint-2xl: 1536px; /* Ultra-wide */
```

### Mobile Adaptations

- Sidebar becomes drawer (hamburger menu)
- Stats grid becomes 2-column, then 1-column
- Tables become cards
- Charts stack vertically
- Reduced padding (16px vs 24px)

---

## Icon System

**Library**: Lucide React
**Size Scale**: 16px (sm), 20px (md), 24px (lg), 32px (xl)

**Key Icons**
- Dashboard: `LayoutDashboard`
- Functions: `FunctionSquare`
- Providers: `Cloud`
- Analytics: `BarChart3`
- Settings: `Settings`
- Deploy: `Rocket`
- Health: `Activity`
- Failover: `Shuffle`
- Success: `CheckCircle2`
- Error: `XCircle`
- Warning: `AlertTriangle`
- Info: `Info`

---

## Tech Stack

### Core
- **Framework**: React 18+ with TypeScript
- **Build Tool**: Vite
- **Styling**: TailwindCSS 3.4+
- **UI Components**: shadcn/ui
- **Animation**: Framer Motion
- **Icons**: Lucide React

### State Management
- **Server State**: TanStack Query (React Query)
- **Client State**: Zustand
- **Form State**: React Hook Form + Zod

### Data Visualization
- **Charts**: Recharts
- **Tables**: TanStack Table

### Routing
- **Router**: React Router v6

### Utilities
- **HTTP Client**: Axios
- **Date Formatting**: date-fns
- **Class Names**: clsx + tailwind-merge

---

## File Structure

```
web/dashboard/
├── public/
│   ├── logo.svg
│   └── favicon.ico
├── src/
│   ├── api/
│   │   ├── client.ts
│   │   ├── auth.ts
│   │   ├── apps.ts
│   │   ├── functions.ts
│   │   ├── providers.ts
│   │   └── analytics.ts
│   ├── components/
│   │   ├── ui/              # shadcn components
│   │   ├── layout/
│   │   │   ├── Sidebar.tsx
│   │   │   ├── TopBar.tsx
│   │   │   ├── PageLayout.tsx
│   │   │   └── MobileNav.tsx
│   │   ├── common/
│   │   │   ├── Logo.tsx
│   │   │   ├── StatusBadge.tsx
│   │   │   ├── ProviderIcon.tsx
│   │   │   ├── StatCard.tsx
│   │   │   └── Sparkline.tsx
│   │   └── charts/
│   │       ├── LineChart.tsx
│   │       ├── BarChart.tsx
│   │       └── PieChart.tsx
│   ├── hooks/
│   │   ├── useAuth.ts
│   │   ├── useApps.ts
│   │   ├── useDeployments.ts
│   │   └── useTheme.ts
│   ├── lib/
│   │   ├── utils.ts
│   │   ├── constants.ts
│   │   └── formatters.ts
│   ├── pages/
│   │   ├── LandingPage/
│   │   │   ├── Hero.tsx
│   │   │   ├── Features.tsx
│   │   │   ├── Pricing.tsx
│   │   │   └── Footer.tsx
│   │   ├── AuthPage/
│   │   │   ├── LoginForm.tsx
│   │   │   └── SignupForm.tsx
│   │   ├── DashboardPage/
│   │   │   ├── index.tsx
│   │   │   ├── StatsRow.tsx
│   │   │   ├── ActivityFeed.tsx
│   │   │   └── ProviderStatus.tsx
│   │   ├── FunctionsPage/
│   │   │   ├── index.tsx
│   │   │   ├── FunctionList.tsx
│   │   │   ├── FunctionEditor.tsx
│   │   │   └── DeployModal.tsx
│   │   ├── ProvidersPage/
│   │   │   ├── index.tsx
│   │   │   ├── ProviderCard.tsx
│   │   │   └── ConnectModal.tsx
│   │   ├── AnalyticsPage/
│   │   │   ├── index.tsx
│   │   │   ├── MetricsGrid.tsx
│   │   │   └── ChartsSection.tsx
│   │   └── SettingsPage/
│   │       ├── index.tsx
│   │       ├── AccountSettings.tsx
│   │       └── BillingSettings.tsx
│   ├── stores/
│   │   ├── authStore.ts
│   │   └── themeStore.ts
│   ├── types/
│   │   ├── api.ts
│   │   ├── models.ts
│   │   └── index.ts
│   ├── App.tsx
│   ├── main.tsx
│   └── index.css
├── index.html
├── package.json
├── tailwind.config.js
├── tsconfig.json
└── vite.config.ts
```

---

## API Integration Patterns

### Authentication Flow

```typescript
// JWT stored in memory (React Context)
// Refresh token in httpOnly cookie

const useAuth = () => {
  const login = async (email: string, password: string) => {
    const { token, user } = await api.auth.login(email, password);
    setToken(token);
    setUser(user);
  };
  
  const logout = () => {
    clearToken();
    queryClient.clear();
  };
};
```

### Data Fetching Pattern

```typescript
// Using TanStack Query
const useApps = () => {
  return useQuery({
    queryKey: ['apps'],
    queryFn: api.apps.list,
    staleTime: 5 * 60 * 1000, // 5 minutes
  });
};

const useCreateApp = () => {
  const queryClient = useQueryClient();
  
  return useMutation({
    mutationFn: api.apps.create,
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['apps'] });
      toast.success('App created successfully');
    },
  });
};
```

---

## Accessibility Requirements

- WCAG 2.1 AA compliance
- Keyboard navigation support
- Focus visible states
- ARIA labels for interactive elements
- Color contrast ratio 4.5:1 minimum
- Reduced motion support (`prefers-reduced-motion`)
- Screen reader friendly tables

---

## Performance Targets

- First Contentful Paint: < 1.5s
- Largest Contentful Paint: < 2.5s
- Time to Interactive: < 3.5s
- Bundle size: < 200KB (gzipped, initial)
- Lighthouse score: > 90

---

## MVP Scope Notes

**Phase 1 (MVP)**
- Landing page
- Auth (login/signup)
- Dashboard home with stats
- Functions list and basic deploy
- Provider connection
- Simple analytics (tables over charts)

**Phase 2**
- Advanced charts
- Function editor with Monaco
- Real-time updates (WebSockets)
- Team collaboration
- Advanced analytics

---

*This specification serves as the blueprint for implementing the FunctionFly dashboard UI.*
