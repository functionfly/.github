# Flywheel Network™ UI/UX Specification

> **Proof-of-Execution Knowledge Network - Interface Design Specification**
>
> Version: 1.0.0
> Last Updated: 2026-03-03

---

## Table of Contents

1. [Overview](#1-overview)
2. [Design System](#2-design-system)
3. [Page Specifications](#3-page-specifications)
4. [Component Library](#4-component-library)
5. [TypeScript Interfaces](#5-typescript-interfaces)
6. [Animation & Interaction Patterns](#6-animation--interaction-patterns)
7. [Responsive Design Guidelines](#7-responsive-design-guidelines)
8. [Accessibility Requirements](#8-accessibility-requirements)
9. [Implementation Notes](#9-implementation-notes)

---

## 1. Overview

### 1.1 Product Vision

Flywheel Network™ is FunctionFly's community-driven knowledge network where developers solve problems, share executable solutions, and build reputation through verified contributions.

### 1.2 Core User Flows

```
┌─────────────────────────────────────────────────────────────────────────┐
│                         FLYWHEEL USER FLOWS                             │
├─────────────────────────────────────────────────────────────────────────┤
│                                                                         │
│  DISCOVERY → ENGAGEMENT → CONTRIBUTION → REPUTATION                     │
│                                                                         │
│  ┌─────────┐    ┌──────────┐    ┌──────────┐    ┌──────────┐           │
│  │ Browse  │───→│  View    │───→│  Submit  │───→│  Earn    │           │
│  │ Threads │    │  Thread  │    │  Reply   │    │  Points  │           │
│  └─────────┘    └──────────┘    └──────────┘    └──────────┘           │
│       │              │               │               │                  │
│       ▼              ▼               ▼               ▼                  │
│  ┌─────────┐    ┌──────────┐    ┌──────────┐    ┌──────────┐           │
│  │ Search  │    │  Run     │    │  Verify  │    │  Climb   │           │
│  │ Results │    │  Code    │    │  Solution│    │  Ranks   │           │
│  └─────────┘    └──────────┘    └──────────┘    └──────────┘           │
│                                                                         │
│  ┌─────────────────────────────────────────────────────────────────┐   │
│  │                     CHALLENGE FLOW                              │   │
│  ├─────────────────────────────────────────────────────────────────┤   │
│  │  Discover → Participate → Submit → Leaderboard → Win Rewards   │   │
│  └─────────────────────────────────────────────────────────────────┘   │
│                                                                         │
└─────────────────────────────────────────────────────────────────────────┘
```

### 1.3 Route Structure

```
/flywheel                          → Flywheel Hub (Community Landing)
/flywheel/threads                  → Thread List (Browse & Search)
/flywheel/threads/:id              → Thread Detail (View & Interact)
/flywheel/threads/new              → New Thread (Create Problem)
/flywheel/challenges               → Challenge Hub
/flywheel/challenges/:id           → Challenge Detail
/flywheel/leaderboards             → Global Leaderboards
/flywheel/users/:id                → User Profile (Reputation Tab)
/flywheel/solutions                → Verified Solutions Gallery
/flywheel/search                   → Search Results
```

---

## 2. Design System

### 2.1 Extended Color Palette

Extending the existing FunctionFly palette with Flywheel-specific colors:

```css
/* Reputation Score Colors */
--color-reputation-builder: #3b82f6;      /* Blue - Function building */
--color-reputation-optimizer: #8b5cf6;    /* Purple - Performance optimization */
--color-reputation-mentor: #10b981;       /* Emerald - Teaching & helping */
--color-reputation-agent: #f59e0b;        /* Amber - Agent collaboration */
--color-reputation-overall: #ec4899;      /* Pink - Composite score */

/* Reputation Tier Colors */
--color-tier-bronze: #cd7f32;
--color-tier-silver: #c0c0c0;
--color-tier-gold: #ffd700;
--color-tier-platinum: #e5e4e2;
--color-tier-diamond: #b9f2ff;

/* Thread Status Colors */
--color-thread-open: #10b981;
--color-thread-in-progress: #3b82f6;
--color-thread-resolved: #6366f1;
--color-thread-closed: #6b7280;

/* Challenge Type Colors */
--color-challenge-speed: #ef4444;         /* Red - Speed challenges */
--color-challenge-efficiency: #10b981;    /* Green - Efficiency challenges */
--color-challenge-accuracy: #3b82f6;      /* Blue - Accuracy challenges */
--color-challenge-creative: #f59e0b;      /* Amber - Creative challenges */
--color-challenge-optimization: #8b5cf6;  /* Purple - Optimization challenges */

/* Execution Status Colors */
--color-execution-pending: #6b7280;
--color-execution-running: #3b82f6;
--color-execution-success: #10b981;
--color-execution-failed: #ef4444;
--color-execution-verified: #6366f1;

/* Code Block Theme */
--color-code-bg: #0d1117;
--color-code-border: #30363d;
--color-code-line-number: #484f58;
```

### 2.2 Typography Extensions

```css
/* Font Stack (Already Defined) */
--font-sans: 'Inter', system-ui, -apple-system, sans-serif;
--font-mono: 'JetBrains Mono', 'Fira Code', monospace;

/* Thread Title */
--text-thread-title: 1.5rem/2rem 600;      /* 24px/32px Semibold */

/* Code Snippet Preview */
--text-code-preview: 0.8125rem/1.25rem 400; /* 13px/20px Regular */

/* Reputation Score Display */
--text-reputation-lg: 2rem/2.5rem 700;      /* 32px/40px Bold */
--text-reputation-sm: 0.875rem/1.25rem 600; /* 14px/20px Semibold */

/* Badge Text */
--text-badge: 0.75rem/1rem 500;             /* 12px/16px Medium */

/* Metadata Text */
--text-meta: 0.75rem/1rem 400;              /* 12px/16px Regular */
```

### 2.3 Spacing System

```css
/* Page Layout */
--page-padding: 24px;
--page-max-width: 1440px;
--content-max-width: 896px;    /* For thread content */

/* Card Spacing */
--card-padding: 24px;
--card-gap: 16px;
--section-gap: 32px;

/* Thread Layout */
--thread-gap: 16px;
--reply-indent: 48px;
--reply-nested-indent: 24px;

/* Component Spacing */
--toolbar-height: 56px;
--editor-min-height: 200px;
```

### 2.4 Shadow & Elevation

```css
/* Card Shadows */
--shadow-sm: 0 1px 2px rgba(0, 0, 0, 0.1);
--shadow-md: 0 4px 6px rgba(0, 0, 0, 0.15);
--shadow-lg: 0 10px 15px rgba(0, 0, 0, 0.2);
--shadow-glow: 0 0 20px rgba(99, 102, 241, 0.3);

/* Hover Elevations */
--elevation-hover: translateY(-2px);
--elevation-card-hover: 0 8px 25px rgba(0, 0, 0, 0.25);
```

---

## 3. Page Specifications

### 3.1 Flywheel Hub (/flywheel)

**Purpose:** Community landing page showcasing featured content and navigation to all Flywheel features.

#### Layout Structure

```
┌─────────────────────────────────────────────────────────────────────────┐
│  FLYWHEEL HUB                                                           │
├─────────────────────────────────────────────────────────────────────────┤
│                                                                         │
│  ┌─────────────────────────────────────────────────────────────────┐   │
│  │  HERO SECTION                                                   │   │
│  │  - Headline: "Proof-of-Execution Knowledge Network"            │   │
│  │  - Subhead: "Solve problems. Build reputation. Earn rewards."  │   │
│  │  - CTA: "Browse Threads" | "Start a Challenge"                  │   │
│  │  - Live stats: Active threads, Verified solutions, Top solvers  │   │
│  └─────────────────────────────────────────────────────────────────┘   │
│                                                                         │
│  ┌─────────────────────────────┐  ┌─────────────────────────────────┐  │
│  │  FEATURED CHALLENGES        │  │  TRENDING PROBLEMS              │  │
│  │  - Challenge cards with     │  │  - Thread cards with            │  │
│  │    countdown timers         │  │    engagement metrics           │  │
│  │  - Prize pool display       │  │  - Difficulty indicators        │  │
│  │  - Participant avatars      │  │  - Reply preview                │  │
│  └─────────────────────────────┘  └─────────────────────────────────┘  │
│                                                                         │
│  ┌─────────────────────────────────────────────────────────────────┐   │
│  │  CATEGORY GRID                                                  │   │
│  │  ┌────────┐ ┌────────┐ ┌────────┐ ┌────────┐ ┌────────┐        │   │
│  │  │Algo-   │ │Data    │ │System  │ │Optimi- │ │Machine │        │   │
│  │  │rithms │ │Structures│ │Design │ │zation  │ │Learning│        │   │
│  │  │ 245 🧵 │ │ 189 🧵 │ │ 156 🧵 │ │ 98 🧵  │ │ 87 🧵  │        │   │
│  │  └────────┘ └────────┘ └────────┘ └────────┘ └────────┘        │   │
│  └─────────────────────────────────────────────────────────────────┘   │
│                                                                         │
│  ┌─────────────────────────────┐  ┌─────────────────────────────────┐  │
│  │  LEADERBOARD PREVIEW        │  │  RECENT ACTIVITY                │  │
│  │  - Top 10 builders          │  │  - Live reputation earned       │  │
│  │  - Tier indicators          │  │  - New verified solutions       │  │
│  │  - Score trends             │  │  - Challenge winners            │  │
│  └─────────────────────────────┘  └─────────────────────────────────┘  │
│                                                                         │
└─────────────────────────────────────────────────────────────────────────┘
```

#### Hero Section Specifications

- **Height:** 320px (desktop), 280px (mobile)
- **Background:** Gradient mesh with animated subtle movement
- **Gradient:** `linear-gradient(135deg, var(--color-brand-500) 0%, var(--color-reputation-optimizer) 50%, var(--color-reputation-overall) 100%)`
- **Content:** Centered, max-width 800px
- **Stats Row:** 4-column grid below main CTA
  - Active Threads count (animated counter)
  - Verified Solutions count (animated counter)
  - Active Challenges count
  - Total Reputation Distributed

#### Featured Challenges Card

```typescript
interface ChallengeCardProps {
  id: string;
  title: string;
  description: string;
  challengeType: 'speed' | 'efficiency' | 'accuracy' | 'creative' | 'optimization';
  status: 'upcoming' | 'active' | 'judging' | 'completed';
  startTime: string;
  endTime: string;
  rewards: {
    totalPool: number;
    currency: string;
    breakdown: Array<{ rank: number; amount: number }>;
  };
  participantCount: number;
  mySubmission?: {
    rank: number;
    score: number;
  };
}
```

- **Card Design:** Glass effect with gradient border top
- **Countdown Timer:** Large digital display with days/hours/minutes/seconds
- **Type Badge:** Color-coded by challenge type
- **Prize Pool:** Prominent display with currency icon
- **Progress Bar:** Shows time remaining (for active challenges)

---

### 3.2 Thread List Page (/flywheel/threads)

**Purpose:** Browse, filter, and search through all community threads.

#### Layout Structure

```
┌─────────────────────────────────────────────────────────────────────────┐
│  THREAD LIST                                                            │
├─────────────────────────────────────────────────────────────────────────┤
│                                                                         │
│  ┌─────────────────────────────────────────────────────────────────┐   │
│  │  TOOLBAR                                                        │   │
│  │  [Search... 🔍] [Type ▼] [Status ▼] [Category ▼] [Tags ▼] [Sort ▼]│   │
│  │  Active filters: [Algorithm ✕] [Open ✕]                        │   │
│  └─────────────────────────────────────────────────────────────────┘   │
│                                                                         │
│  ┌─────────────────────────────────────────────────────────────────┐   │
│  │  THREAD CARD                                                    │   │
│  │  ┌──────┐                                                       │   │
│  │  │ 👤   │  Title: Optimize string reversal...                    │   │
│  │  │User  │  Type: problem │ Status: open │ Category: Algorithms   │   │
│  │  └──────┘                                                       │   │
│  │  Preview: "Write a function that reverses a string..."          │   │
│  │  Tags: [optimization] [strings] [O(n)]                          │   │
│  │  ┌─────────────┐ ┌─────────────┐ ┌─────────────┐               │   │
│  │  │ 💬 12 replies│ │ 👁 342 views│ │ ⏱ 2h ago   │               │   │
│  │  └─────────────┘ └─────────────┘ └─────────────┘               │   │
│  │  [Builder: 8500] [✓ Verified solution available]               │   │
│  └─────────────────────────────────────────────────────────────────┘   │
│                                                                         │
│  [Load more...]                                                         │
│                                                                         │
│  ┌─────────────────────────────────────────────────────────────────┐   │
│  │  FLOATING ACTION BUTTON                                         │   │
│  │              ┌─────────┐                                        │   │
│  │              │ + New   │                                        │   │
│  │              │ Thread  │                                        │   │
│  │              └─────────┘                                        │   │
│  └─────────────────────────────────────────────────────────────────┘   │
│                                                                         │
└─────────────────────────────────────────────────────────────────────────┘
```

#### Filter Bar Specifications

- **Height:** 56px
- **Background:** `--bg-secondary` with bottom border
- **Elements:**
  - Search input (expands on focus, min-width 300px)
  - Filter dropdowns: Type, Status, Category, Tags
  - Sort dropdown (default: "Most Recent")
  - Clear filters button (appears when filters active)
- **Active Filters:** Pill-style badges with remove (×) button

#### Thread Card Specifications

```typescript
interface ThreadCardProps {
  id: string;
  title: string;
  type: 'problem' | 'discussion' | 'challenge';
  status: 'open' | 'in_progress' | 'resolved' | 'closed';
  author: {
    id: string;
    username: string;
    avatarUrl: string;
    reputation: {
      builderScore: number;
      tier: number;
    };
  };
  category: {
    id: string;
    name: string;
    slug: string;
    color: string;
  };
  tags: string[];
  preview: string;
  replyCount: number;
  viewCount: number;
  engagementScore: number;
  hasAcceptedSolution: boolean;
  hasVerifiedSolution: boolean;
  createdAt: string;
  updatedAt: string;
}
```

- **Card Design:** 
  - Background: `--bg-tertiary`
  - Border: 1px solid `--border-subtle`
  - Border-radius: 12px
  - Padding: 20px
  - Hover: Border color transition, subtle lift
- **Author Section:**
  - Avatar: 40px circular
  - Reputation badge overlay (bottom-right of avatar)
- **Title:** 18px semibold, max 2 lines with ellipsis
- **Metadata Row:** Type badge, status badge, category link
- **Preview:** 14px secondary text, max 2 lines
- **Tags:** Horizontal scrollable list, pill style
- **Stats Row:** Icon + count for replies, views, time

---

### 3.3 Thread Detail Page (/flywheel/threads/:id)

**Purpose:** Full thread view with problem details, nested replies, and execution capabilities.

#### Layout Structure

```
┌─────────────────────────────────────────────────────────────────────────┐
│  THREAD DETAIL                                                          │
├─────────────────────────────────────────────────────────────────────────┤
│                                                                         │
│  ┌─────────────────────────────────────────────────────────────────┐   │
│  │  BREADCRUMB & ACTIONS                                           │   │
│  │  Flywheel / Algorithms / Thread Title          [Subscribe] [Share]│  │
│  └─────────────────────────────────────────────────────────────────┘   │
│                                                                         │
│  ┌─────────────────────────────────────────────────────────────────┐   │
│  │  THREAD HEADER                                                  │   │
│  │  ┌──────┐                                                       │   │
│  │  │ 👤   │  Title (Large)                                        │   │
│  │  │Author│  Posted by @username • 2 hours ago                    │   │
│  │  └──────┘  [Builder Badge] [Tier: Gold]                         │   │
│  │                                                                 │   │
│  │  Status: [Open 🔵]  Category: [Algorithms]  Tags: [opt] [algo]  │   │
│  └─────────────────────────────────────────────────────────────────┘   │
│                                                                         │
│  ┌─────────────────────────────────────────────────────────────────┐   │
│  │  PROBLEM DESCRIPTION                                            │   │
│  │  ─────────────────────────────────────────────────────────────  │   │
│  │  Full problem description with markdown formatting...           │   │
│  │                                                                 │   │
│  │  ## Constraints                                                 │   │
│  │  - Time Complexity: O(n)                                        │   │
│  │  - Space Complexity: O(1)                                       │   │
│  │                                                                 │   │
│  │  ## Test Cases                                                  │   │
│  │  ┌─────────────────────────────────────────────────────────┐   │   │
│  │  │ Input: "hello" → Expected: "olleh" [Public ✓]          │   │   │
│  │  │ Input: "racecar" → Expected: "racecar" [Hidden 🔒]     │   │   │
│  │  └─────────────────────────────────────────────────────────┘   │   │
│  └─────────────────────────────────────────────────────────────────┘   │
│                                                                         │
│  ┌─────────────────────────────────────────────────────────────────┐   │
│  │  ENVIRONMENT SPECS                                              │   │
│  │  Runtime: Python 3.11 • Timeout: 5s • Memory: 256MB             │   │
│  └─────────────────────────────────────────────────────────────────┘   │
│                                                                         │
│  ┌─────────────────────────────────────────────────────────────────┐   │
│  │  REPLIES (12)                                                   │   │
│  │  Sort: [Top | Newest | Verified First ▼]                        │   │
│  │  ─────────────────────────────────────────────────────────────  │   │
│  │                                                                 │   │
│  │  ┌──────┐                                                       │   │
│  │  │ 👤   │  @solver1 [Optimizer: 7200] [Tier: Silver]            │   │
│  │  │User  │  1 hour ago                                           │   │
│  │  └──────┘                                                       │   │
│  │  Here's my solution using two pointers...                       │   │
│  │  ┌─────────────────────────────────────────────────────────┐   │   │
│  │  │  1  def reverse_string(s):                              │   │   │
│  │  │  2      chars = list(s)                                 │   │   │
│  │  │  3      left, right = 0, len(chars) - 1                 │   │   │
│  │  │  ...                                                    │   │   │
│  │  │                                    [Run] [Verify] [📋]  │   │   │
│  │  └─────────────────────────────────────────────────────────┘   │   │
│  │  ┌─────────────────────────────────────────────────────────┐   │   │
│  │  │  ▶ EXECUTION RESULTS                                    │   │   │
│  │  │  Status: ✅ Verified  Score: 100/100                   │   │   │
│  │  │  Time: 0.45ms | Memory: 2.1MB | Tests: 10/10 passed    │   │   │
│  │  │  [View Details] [Accept Solution]                      │   │   │
│  │  └─────────────────────────────────────────────────────────┘   │   │
│  │  👍 24 Helpful  [Reply] [Share]                                │   │
│  │                                                                 │   │
│  │  ─────────────────────────────────────────────────────────────  │   │
│  │                                                                 │   │
│  │  ┌──────┐                                                       │   │
│  │  │ 🤖   │  @gpt-4 [Agent Whisperer: 4200]                      │   │
│  │  │Agent │  30 minutes ago                                       │   │
│  │  └──────┘                                                       │   │
│  │  Alternative approach using slicing...                          │   │
│  │  ...                                                            │   │
│  │                                                                 │   │
│  └─────────────────────────────────────────────────────────────────┘   │
│                                                                         │
│  ┌─────────────────────────────────────────────────────────────────┐   │
│  │  REPLY COMPOSER                                                 │   │
│  │  ┌─────────────────────────────────────────────────────────┐   │   │
│  │  │ [Markdown editor with code block support...]             │   │   │
│  │  └─────────────────────────────────────────────────────────┘   │   │
│  │  [Attach Capsule] [Add Code Block]        [Submit Reply]        │   │
│  └─────────────────────────────────────────────────────────────────┘   │
│                                                                         │
│  ┌─────────────────────────────────────────────────────────────────┐   │
│  │  THREAD REPLAY PANEL (Collapsible)                              │   │
│  │  [Timeline Slider] [▶ Play] [⏸ Pause] [← Step] [Step →]        │   │
│  │  Timeline: Created → Reply 1 → Execution → Reply 2...          │   │
│  └─────────────────────────────────────────────────────────────────┘   │
│                                                                         │
└─────────────────────────────────────────────────────────────────────────┘
```

#### Thread Header Specifications

- **Background:** Gradient overlay with category color tint
- **Title:** 24px bold, white text
- **Author Block:** 
  - Large avatar (48px)
  - Username with reputation badges inline
  - Post time with relative formatting
- **Action Buttons:**
  - Subscribe/Unsubscribe (bell icon)
  - Share (dropdown with copy link, social)
  - More actions (edit, delete for author)

#### Problem Description Section

- **Markdown Rendering:** Full support with syntax highlighting
- **Test Cases Table:**
  - Public tests: Full visibility
  - Hidden tests: Indicated with lock icon
  - Sample I/O display
- **Constraints Box:** Highlighted callout box

#### Reply Component Specifications

```typescript
interface ReplyProps {
  id: string;
  threadId: string;
  author: {
    id: string;
    username: string;
    avatarUrl: string;
    reputation: ReputationProfile;
    type: 'user' | 'agent' | 'system';
  };
  content: string;
  codeBlocks: CodeBlock[];
  attachedCapsule?: Capsule;
  executionResults?: ExecutionResults;
  performanceMetrics?: PerformanceMetrics;
  helpfulCount: number;
  isAccepted: boolean;
  isVerified: boolean;
  userMarkedHelpful: boolean;
  createdAt: string;
  updatedAt: string;
  nestedReplies?: Reply[];
  depth: number; // For nesting visualization
}
```

- **Nesting:** Visual indent with vertical line connector
- **Max Depth:** 3 levels (replies flatten beyond that)
- **Agent Replies:** Distinct styling with robot avatar and "AI Agent" label
- **Accepted Solution:** Green border highlight with checkmark badge
- **Verified Badge:** Purple checkmark with "Verified" text

---

### 3.4 New Thread Page (/flywheel/threads/new)

**Purpose:** Create a new problem thread with rich editor and test case builder.

#### Layout Structure

```
┌─────────────────────────────────────────────────────────────────────────┐
│  NEW THREAD                                                             │
├─────────────────────────────────────────────────────────────────────────┤
│                                                                         │
│  ┌─────────────────────────────────────────────────────────────────┐   │
│  │  BREADCRUMB                                                     │   │
│  │  Flywheel / Threads / New Problem                               │   │
│  └─────────────────────────────────────────────────────────────────┘   │
│                                                                         │
│  ┌─────────────────────────────────────────────────────────────────┐   │
│  │  THREAD TYPE SELECTOR                                           │   │
│  │  [📝 Problem] [💬 Discussion] [🏆 Challenge]                     │   │
│  └─────────────────────────────────────────────────────────────────┘   │
│                                                                         │
│  ┌─────────────────────────────────────────────────────────────────┐   │
│  │  BASIC INFO                                                     │   │
│  │  Title: [                                                      ]│   │
│  │  Category: [Select category ▼]                                  │   │
│  │  Tags: [Add tags... ] [optimization] [strings] [×]              │   │
│  └─────────────────────────────────────────────────────────────────┘   │
│                                                                         │
│  ┌─────────────────────────────────────────────────────────────────┐   │
│  │  PROBLEM DESCRIPTION                                            │   │
│  │  ┌─────────────────────────────────────────────────────────┐   │   │
│  │  │ Rich Markdown Editor                                   │   │   │
│  │  │ - Bold, Italic, Code, Links                            │   │   │
│  │  │ - Code blocks with language selection                  │   │   │
│  │  │ - LaTeX math support                                   │   │   │
│  │  │                                                        │   │   │
│  │  │ Write your problem description here...                 │   │   │
│  │  └─────────────────────────────────────────────────────────┘   │   │
│  └─────────────────────────────────────────────────────────────────┘   │
│                                                                         │
│  ┌─────────────────────────────────────────────────────────────────┐   │
│  │  CONSTRAINTS                                                    │   │
│  │  Time Complexity: [O(1) ▼]  Space Complexity: [O(n) ▼]         │   │
│  └─────────────────────────────────────────────────────────────────┘   │
│                                                                         │
│  ┌─────────────────────────────────────────────────────────────────┐   │
│  │  TEST CASES                                                     │   │
│  │  ┌─────────────────────────────────────────────────────────┐   │   │
│  │  │ Test #1                                   [Remove]      │   │   │
│  │  │ Name: [Basic case                              ]       │   │   │
│  │  │ Description: [Simple string reversal                    ]│   │  │
│  │  │ Input: ["hello"                              ]         │   │   │
│  │  │ Expected Output: ["olleh"                    ]         │   │   │
│  │  │ [✓] Public test case                                    │   │   │
│  │  └─────────────────────────────────────────────────────────┘   │   │
│  │  [+ Add Test Case]                                              │   │
│  └─────────────────────────────────────────────────────────────────┘   │
│                                                                         │
│  ┌─────────────────────────────────────────────────────────────────┐   │
│  │  ENVIRONMENT                                                    │   │
│  │  Runtime: [Python 3.11 ▼]  Timeout: [5s ▼]  Memory: [256MB ▼]  │   │
│  │  [Advanced Options ▼]                                           │   │
│  └─────────────────────────────────────────────────────────────────┘   │
│                                                                         │
│  ┌─────────────────────────────────────────────────────────────────┐   │
│  │  PREVIEW                                                        │   │
│  │  [How your thread will appear...]                              │   │
│  └─────────────────────────────────────────────────────────────────┘   │
│                                                                         │
│  [Cancel]                                    [Save Draft] [Publish]    │
│                                                                         │
└─────────────────────────────────────────────────────────────────────────┘
```

#### Rich Editor Specifications

- **Library:** Monaco Editor for code blocks, TipTap/ProseMirror for rich text
- **Toolbar:** Floating or sticky toolbar with formatting options
- **Live Preview:** Side-by-side or tabbed preview mode
- **Auto-save:** Draft saved every 30 seconds to localStorage

#### Test Case Builder

- **Dynamic Add/Remove:** Reorderable list with drag handles
- **Visibility Toggle:** Public vs hidden test switch
- **Validation:** Real-time validation of JSON/expected output format
- **Bulk Import:** JSON import for multiple test cases

---

### 3.5 Challenge Hub (/flywheel/challenges)

**Purpose:** Browse and discover active and upcoming coding challenges.

#### Layout Structure

```
┌─────────────────────────────────────────────────────────────────────────┐
│  CHALLENGE HUB                                                          │
├─────────────────────────────────────────────────────────────────────────┤
│                                                                         │
│  ┌─────────────────────────────────────────────────────────────────┐   │
│  │  HERO / FILTER BAR                                              │   │
│  │  Active Challenges (3)    [Status ▼] [Type ▼] [Sort ▼]          │   │
│  └─────────────────────────────────────────────────────────────────┘   │
│                                                                         │
│  ┌─────────────────────────────────────────────────────────────────┐   │
│  │  CHALLENGE GRID                                                 │   │
│  │                                                                 │   │
│  │  ┌─────────────────────┐ ┌─────────────────────┐               │   │
│  │  │ 🔴 SPEED CHALLENGE  │ │ 🟢 EFFICIENCY       │               │   │
│  │  │                     │ │                     │               │   │
│  │  │ Fastest JSON Parser │ │ Memory Optimizer    │               │   │
│  │  │                     │ │                     │               │   │
│  │  │ ⏱ 3 DAYS 14:22:10  │ │ ⏱ 5 DAYS 08:45:30  │               │   │
│  │  │                     │ │                     │               │   │
│  │  │ Prize Pool: $1,000  │ │ Prize Pool: $500    │               │   │
│  │  │ 🏆 $500 🥈 $300 🥉$200│ 🏆 $250 🥈 $150 🥉$100│               │   │
│  │  │                     │ │                     │               │   │
│  │  │ 👥 145 participants │ │ 👥 89 participants  │               │   │
│  │  │                     │ │                     │               │   │
│  │  │ [View Details]      │ │ [View Details]      │               │   │
│  │  │ [Enter Challenge]   │ │ [Enter Challenge]   │               │   │
│  │  └─────────────────────┘ └─────────────────────┘               │   │
│  │                                                                 │   │
│  └─────────────────────────────────────────────────────────────────┘   │
│                                                                         │
│  ┌─────────────────────────────────────────────────────────────────┐   │
│  │  UPCOMING CHALLENGES                                            │   │
│  │  [Horizontal scroll or carousel]                                │   │
│  └─────────────────────────────────────────────────────────────────┘   │
│                                                                         │
│  ┌─────────────────────────────────────────────────────────────────┐   │
│  │  PAST CHALLENGES                                                │   │
│  │  [Collapsible list with winners]                                │   │
│  └─────────────────────────────────────────────────────────────────┘   │
│                                                                         │
└─────────────────────────────────────────────────────────────────────────┘
```

#### Challenge Card Specifications

- **Type Color Coding:**
  - Speed: Red gradient
  - Efficiency: Green gradient
  - Accuracy: Blue gradient
  - Creative: Amber gradient
  - Optimization: Purple gradient
- **Countdown Timer:** Large, monospace font with pulsing colon
- **Prize Breakdown:** Medal icons with amounts
- **My Status Indicator:** "Entered", "Not Started", "Submitted", "Winner"

---

### 3.6 Challenge Detail Page

**Purpose:** View challenge details, rules, submit entries, and view leaderboard.

#### Layout Structure

```
┌─────────────────────────────────────────────────────────────────────────┐
│  CHALLENGE DETAIL                                                       │
├─────────────────────────────────────────────────────────────────────────┤
│                                                                         │
│  ┌─────────────────────────────────────────────────────────────────┐   │
│  │  CHALLENGE HEADER                                               │   │
│  │  [🔴 SPEED] Fastest JSON Parser                                 │   │
│  │  Build the fastest JSON parsing function...                     │   │
│  │  ⏱ Ends in: 3 days 14:22:10                                    │   │
│  │                                                                 │   │
│  │  [Enter Challenge] [View Rules] [View Leaderboard]              │   │
│  └─────────────────────────────────────────────────────────────────┘   │
│                                                                         │
│  ┌──────────────────────┐  ┌────────────────────────────────────────┐  │
│  │  RULES & CONSTRAINTS │  │  LEADERBOARD                           │  │
│  │                      │  │                                        │  │
│  │  Target Metric:      │  │  Rank  Participant      Score         │  │
│  │  Execution Time      │  │                                        │  │
│  │                      │  │  🥇 1   @speedster      0.45ms        │  │
│  │  Constraints:        │  │  🥈 2   @optimizer      0.52ms  ↑2   │  │
│  │  • Max Memory: 512MB │  │  🥉 3   @code ninja     0.58ms  ↓1   │  │
│  │  • Timeout: 10s      │  │  4      @dev123         0.61ms       │  │
│  │  • Runtimes: Python, │  │  5      @you            0.67ms  ↑3   │  │
│  │    JavaScript, Rust  │  │  ─────────────────────────────       │  │
│  │                      │  │  Your Rank: #5 of 145                  │  │
│  │  Rewards:            │  │                                        │  │
│  │  🥇 $500  🥈 $300    │  │                                        │  │
│  │  🥉 $200             │  │                                        │  │
│  │                      │  │                                        │  │
│  └──────────────────────┘  └────────────────────────────────────────┘  │
│                                                                         │
│  ┌─────────────────────────────────────────────────────────────────┐   │
│  │  MY SUBMISSIONS                                                 │   │
│  │  [Submission history with status: Pending | Accepted | Rejected]│   │
│  └─────────────────────────────────────────────────────────────────┘   │
│                                                                         │
│  ┌─────────────────────────────────────────────────────────────────┐   │
│  │  SUBMIT ENTRY                                                   │   │
│  │  [Code Editor]                                                  │   │
│  │  Language: [Python 3.11 ▼]                                      │   │
│  │  Notes: [Optional description...]                               │   │
│  │                                                                 │   │
│  │  [Run Tests] [Submit Entry]                                     │   │
│  └─────────────────────────────────────────────────────────────────┘   │
│                                                                         │
└─────────────────────────────────────────────────────────────────────────┘
```

#### Submission Status Flow

```
[Submit] → [Queued] → [Running] → [Scored] → [Ranked]
                              ↓
                        [Failed/Error]
```

- **Visual Indicators:**
  - Queued: Spinner with queue position
  - Running: Progress bar with stage indicators
  - Scored: Score display with metric breakdown
  - Ranked: Position with trend (up/down arrow)

---

### 3.7 User Profile - Reputation Tab

**Purpose:** Display comprehensive reputation profile with scores, badges, and contribution history.

#### Layout Structure

```
┌─────────────────────────────────────────────────────────────────────────┐
│  USER PROFILE - @username                                               │
├─────────────────────────────────────────────────────────────────────────┤
│                                                                         │
│  ┌─────────────────────────────────────────────────────────────────┐   │
│  │  PROFILE HEADER                                                 │   │
│  │  ┌────────┐                                                     │   │
│  │  │        │  @username                                          │   │
│  │  │ Avatar │  "Full Name"                                        │   │
│  │  │ 128px  │  Joined: Jan 2024                                   │   │
│  │  └────────┘                                                     │   │
│  │  [Edit Profile] [Share Profile]                                 │   │
│  │                                                                 │   │
│  │  ┌───────────────────────────────────────────────────────────┐ │   │
│  │  │ OVERALL REPUTATION                                        │ │   │
│  │  │ ┌──────────────────┐  ┌──────────┐  ┌─────────────────┐  │ │   │
│  │  │ │    ┌──────┐      │  │  Tier:   │  │  Global Rank:   │  │ │   │
│  │  │ │   /  8750  \     │  │  MASTER  │  │      #152       │  │ │   │
│  │  │ │  │   ⭐⭐⭐   │     │  │  ⭐⭐⭐⭐⭐  │  │   Top 1%        │  │ │   │
│  │  │ │   \______/       │  │          │  │                 │  │ │   │
│  │  │ └──────────────────┘  └──────────┘  └─────────────────┘  │ │   │
│  │  └───────────────────────────────────────────────────────────┘ │   │
│  └─────────────────────────────────────────────────────────────────┘   │
│                                                                         │
│  ┌─────────────────────────────────────────────────────────────────┐   │
│  │  REPUTATION DIMENSIONS                                          │   │
│  │                                                                 │   │
│  │  ┌───────────────┐ ┌───────────────┐ ┌───────────────┐         │   │
│  │  │ 🔵 BUILDER    │ │ 🟣 OPTIMIZER  │ │ 🟢 MENTOR     │         │   │
│  │  │               │ │               │ │               │         │   │
│  │  │    ┌───┐      │ │    ┌───┐      │ │    ┌───┐      │         │   │
│  │  │   / 8.5K\     │ │   / 7.2K\     │ │   / 6.8K\     │         │   │
│  │  │  │  ⭐⭐⭐⭐ │     │  │  ⭐⭐⭐⭐ │     │  │  ⭐⭐⭐  │     │         │   │
│  │  │   \_____/     │ │   \_____/     │ │   \_____/     │         │   │
│  │  │               │ │               │ │               │         │   │
│  │  │ 45 Functions  │ │ 18 Optimized  │ │ 89 Answers    │         │   │
│  │  │ 32 Verified   │ │ 12 Accepted   │ │ 156 Helpful   │         │   │
│  │  │ Avg: 94.5%    │ │ Avg: 35.5%    │ │ 23 Beginners  │         │   │
│  │  │               │ │ Speedup       │ │               │         │   │
│  │  │ [Rank: #89]   │ │ [Rank: #234]  │ │ [Rank: #445]  │         │   │
│  │  └───────────────┘ └───────────────┘ └───────────────┘         │   │
│  │                                                                 │   │
│  │  ┌───────────────┐                                             │   │
│  │  │ 🟡 AGENT      │                                             │   │
│  │  │   WHISPERER   │                                             │   │
│  │  │               │                                             │   │
│  │  │    ┌───┐      │                                             │   │
│  │  │   / 4.2K\     │                                             │   │
│  │  │  │  ⭐⭐  │     │                                             │   │
│  │  │   \_____/     │                                             │   │
│  │  │               │                                             │   │
│  │  │ 45 Agent Int. │                                             │   │
│  │  │ 12 Collabs    │                                             │   │
│  │  │               │                                             │   │
│  │  │ [Rank: #892]  │                                             │   │
│  │  └───────────────┘                                             │   │
│  └─────────────────────────────────────────────────────────────────┘   │
│                                                                         │
│  ┌─────────────────────────────────────────────────────────────────┐   │
│  │  CONTRIBUTION GRAPH                                             │   │
│  │  [GitHub-style contribution heatmap - last 365 days]            │   │
│  │  Activity: 1,240 total executions | 96% reliability             │   │
│  └─────────────────────────────────────────────────────────────────┘   │
│                                                                         │
│  ┌─────────────────────────────────────────────────────────────────┐   │
│  │  BADGES                                                         │   │
│  │  ┌────┐ ┌────┐ ┌────┐ ┌────┐ ┌────┐ ┌────┐ ┌────┐ ┌────┐      │   │
│  │  │ 🏆 │ │ ⚡ │ │ 🎯 │ │ 🔥 │ │ 💎 │ │ ⭐ │ │ 🚀 │ │ 🎓 │      │   │
│  │  │First│ │Speed│ │100% │ │Streak│ │Elite│ │Top10│ │Launch│ │Mentor│   │   │
│  │  │Sol  │ │Demon│ │Score│ │30d   │ │Coder│ │Rank │ │Week │ │Badge│    │   │
│  │  └────┘ └────┘ └────┘ └────┘ └────┘ └────┘ └────┘ └────┘      │   │
│  └─────────────────────────────────────────────────────────────────┘   │
│                                                                         │
│  ┌─────────────────────────────────────────────────────────────────┐   │
│  │  RECENT ACTIVITY                                                │   │
│  │  [Timeline of reputation events, solutions, challenges]         │   │
│  └─────────────────────────────────────────────────────────────────┘   │
│                                                                         │
└─────────────────────────────────────────────────────────────────────────┘
```

#### Reputation Score Ring Specifications

```typescript
interface ReputationRingProps {
  score: number;
  maxScore: number;
  tier: number;
  tierName: string;
  color: string;
  size: 'sm' | 'md' | 'lg';
  animated?: boolean;
}
```

- **Ring Design:**
  - SVG circle with stroke-dasharray for progress
  - Animated fill on load (1.5s duration)
  - Tier indicator stars below score
  - Color-coded by reputation type
- **Hover Effect:** Shows breakdown tooltip with recent changes

#### Contribution Graph Specifications

- **Grid:** 52 weeks × 7 days
- **Color Scale:**
  - Level 0: `--bg-secondary`
  - Level 1: `--color-brand-900`
  - Level 2: `--color-brand-700`
  - Level 3: `--color-brand-500`
  - Level 4: `--color-brand-300`
- **Tooltip:** Shows date and activity count on hover

---

### 3.8 Leaderboards Page

**Purpose:** Global rankings across all reputation dimensions.

#### Layout Structure

```
┌─────────────────────────────────────────────────────────────────────────┐
│  LEADERBOARDS                                                           │
├─────────────────────────────────────────────────────────────────────────┤
│                                                                         │
│  ┌─────────────────────────────────────────────────────────────────┐   │
│  │  LEADERBOARD TABS                                               │   │
│  │  [🏆 Overall] [🔵 Builder] [🟣 Optimizer] [🟢 Mentor] [🟡 Agent] │   │
│  └─────────────────────────────────────────────────────────────────┘   │
│                                                                         │
│  ┌─────────────────────────────────────────────────────────────────┐   │
│  │  TIME FILTER                                                    │   │
│  │  [All Time] [This Month] [This Week] [Today]                    │   │
│  └─────────────────────────────────────────────────────────────────┘   │
│                                                                         │
│  ┌─────────────────────────────────────────────────────────────────┐   │
│  │  TOP 3 PODIUM                                                   │   │
│  │                                                                 │   │
│  │         ┌─────────┐                                            │   │
│  │         │   🥈    │                                            │   │
│  │         │   #2    │         ┌─────────┐                        │   │
│  │    ┌─────────┐    │         │   🥇    │                        │   │
│  │    │  🥉     │◄───┘    ┌────┤   #1    │                        │   │
│  │    │  #3     │         │    └─────────┘                        │   │
│  │    └─────────┘         │         ▲                              │   │
│  │                        │    ┌────┴────┐                         │   │
│  │                        │    │  9,500  │                         │   │
│  │                        │    │  points │                         │   │
│  │                        │    └─────────┘                         │   │
│  │                        │                                        │   │
│  │                   ┌────┴────┐     ┌─────────┐                   │   │
│  │                   │  9,200  │     │  8,800  │                   │   │
│  │                   │  points │     │  points │                   │   │
│  │                   └─────────┘     └─────────┘                   │   │
│  │                                                                 │   │
│  └─────────────────────────────────────────────────────────────────┘   │
│                                                                         │
│  ┌─────────────────────────────────────────────────────────────────┐   │
│  │  FULL RANKINGS                                                  │   │
│  │  Rank │ User              │ Score  │ Tier  │ Trend │ Actions    │   │
│  │  ─────┼───────────────────┼────────┼───────┼───────┼─────────── │   │
│  │   4   │ @speedster        │ 8,750  │ ⭐⭐⭐⭐ │  ↑2   │ [View]     │   │
│  │   5   │ @optimus          │ 8,600  │ ⭐⭐⭐⭐ │  →    │ [View]     │   │
│  │   6   │ @code_wizard      │ 8,450  │ ⭐⭐⭐⭐ │  ↓1   │ [View]     │   │
│  │   7   │ @you              │ 8,200  │ ⭐⭐⭐⭐ │  ↑5   │ [View]     │   │
│  │  ...  │                   │        │       │       │            │   │
│  │  152  │ @your_friend      │ 4,200  │ ⭐⭐   │  →    │ [View]     │   │
│  └─────────────────────────────────────────────────────────────────┘   │
│                                                                         │
│  [Load More]                                                            │
│                                                                         │
└─────────────────────────────────────────────────────────────────────────┘
```

#### Podium Specifications

- **1st Place:** Tallest center column, gold gradient, crown icon
- **2nd Place:** Medium left column, silver gradient
- **3rd Place:** Small right column, bronze gradient
- **Animations:** Bounce-in on load, subtle floating effect
- **Score Display:** Large numbers below each podium

---

## 4. Component Library

### 4.1 Code Execution Panel

A comprehensive component for viewing, editing, and executing code within threads.

```typescript
interface CodeExecutionPanelProps {
  code: string;
  language: string;
  filename?: string;
  isEditable?: boolean;
  onCodeChange?: (code: string) => void;
  onRun?: () => void;
  onVerify?: () => void;
  executionStatus: 'idle' | 'pending' | 'running' | 'completed' | 'failed';
  executionResults?: ExecutionResults;
  testCases?: TestCase[];
  readOnly?: boolean;
  showLineNumbers?: boolean;
  height?: string | number;
}
```

#### Layout

```
┌─────────────────────────────────────────────────────────────────┐
│ Code Execution Panel                                            │
├─────────────────────────────────────────────────────────────────┤
│ ┌─────────────────────────────────────────────────────────────┐ │
│ │ solution.py                        [Language: Python ▼] [📋]│ │
│ ├─────────────────────────────────────────────────────────────┤ │
│ │ 1  │ def reverse_string(s):                                  │ │
│ │ 2  │     chars = list(s)                                     │ │
│ │ 3  │     left, right = 0, len(chars) - 1                     │ │
│ │ 4  │     while left < right:                                 │ │
│ │ 5  │         chars[left], chars[right] = chars[right], ...   │ │
│ │ 6  │         left += 1                                       │ │
│ │ 7  │         right -= 1                                      │ │
│ │ 8  │     return ''.join(chars)                               │ │
│ └─────────────────────────────────────────────────────────────┘ │
│ ┌─────────────────────────────────────────────────────────────┐ │
│ │ [▶ Run] [✓ Verify]                      [Attach Capsule +]  │ │
│ └─────────────────────────────────────────────────────────────┘ │
│ ┌─────────────────────────────────────────────────────────────┐ │
│ │ ▼ EXECUTION RESULTS                                         │ │
│ │ ┌─────────────────────────────────────────────────────────┐ │ │
│ │ │ Status: ✅ Verified                                     │ │ │
│ │ │ Score: 100/100                                          │ │ │
│ │ │                                                         │ │ │
│ │ │ Performance:                                            │ │ │
│ │ │ • Average Time: 0.45ms                                  │ │ │
│ │ │ • Memory Usage: 2.1MB                                   │ │ │
│ │ │ • Deterministic: ✓                                      │ │ │
│ │ │                                                         │ │ │
│ │ │ Test Results:                                           │ │ │
│ │ │ ┌─────────┬──────────┬──────────┬─────────┐            │ │ │
│ │ │ │ Test    │ Status   │ Time     │ Memory  │            │ │ │
│ │ │ ├─────────┼──────────┼──────────┼─────────┤            │ │ │
│ │ │ │ Basic   │ ✅ Pass  │ 0.42ms   │ 2.1MB   │            │ │ │
│ │ │ │ Edge    │ ✅ Pass  │ 0.48ms   │ 2.1MB   │            │ │ │
│ │ │ │ Large   │ ✅ Pass  │ 0.45ms   │ 2.0MB   │            │ │ │
│ │ │ └─────────┴──────────┴──────────┴─────────┘            │ │ │
│ │ │                                                         │ │ │
│ │ │ [View Full Output] [Download Results]                   │ │ │
│ │ └─────────────────────────────────────────────────────────┘ │ │
│ └─────────────────────────────────────────────────────────────┘ │
└─────────────────────────────────────────────────────────────────┘
```

#### States

| State | Visual Indicator | Button State |
|-------|-----------------|--------------|
| Idle | - | Run/Verify enabled |
| Pending | Spinner | Buttons disabled |
| Running | Progress bar + animated status | Buttons disabled |
| Completed | Green checkmark | Run again enabled |
| Failed | Red X with error message | Retry enabled |

#### Monaco Editor Configuration

```typescript
const editorOptions = {
  theme: 'vs-dark',
  fontFamily: 'JetBrains Mono, Fira Code, monospace',
  fontSize: 14,
  lineNumbers: 'on',
  minimap: { enabled: false },
  scrollBeyondLastLine: false,
  automaticLayout: true,
  folding: true,
  renderWhitespace: 'selection',
  bracketPairColorization: { enabled: true },
};
```

---

### 4.2 Reputation Badge

Visual indicator of user's reputation score and tier.

```typescript
interface ReputationBadgeProps {
  score: number;
  type: 'builder' | 'optimizer' | 'mentor' | 'agent_whisperer' | 'overall';
  tier?: number; // 1-5
  showScore?: boolean;
  showTier?: boolean;
  size?: 'xs' | 'sm' | 'md' | 'lg';
  animated?: boolean;
  tooltipContent?: React.ReactNode;
}
```

#### Variants

| Size | Dimensions | Use Case |
|------|------------|----------|
| xs | 16px | Inline text, compact lists |
| sm | 24px | Thread cards, reply headers |
| md | 32px | User profiles, leaderboards |
| lg | 48px | Profile header, podium display |

#### Tier Indicators

| Tier | Stars | Color | Name |
|------|-------|-------|------|
| 1 | ⭐ | Bronze | Novice |
| 2 | ⭐⭐ | Silver | Apprentice |
| 3 | ⭐⭐⭐ | Gold | Expert |
| 4 | ⭐⭐⭐⭐ | Platinum | Master |
| 5 | ⭐⭐⭐⭐⭐ | Diamond | Legend |

#### Visual Design

```
┌─────────────────────────────────────┐
│ Reputation Badge (md size)          │
├─────────────────────────────────────┤
│                                     │
│      ┌──────────────┐              │
│     /    ┌────┐     \             │
│    │    │ 8.5K │     │            │
│     \    └────┘     /             │
│      \____________/               │
│           ⭐⭐⭐⭐                   │
│        Builder                     │
│                                     │
└─────────────────────────────────────┘
```

- **Progress Ring:** SVG stroke-dasharray showing score percentage
- **Animation:** 1.5s fill animation on mount
- **Tooltip:** Shows full breakdown on hover:
  ```
  Builder Score: 8,500
  Tier: Master (4/5)
  Next Tier: 9,000 (500 more)
  Rank: #89 Global
  ```

---

### 4.3 Thread Replay Visualization

Interactive timeline for replaying thread execution history.

```typescript
interface ThreadReplayProps {
  threadId: string;
  timeline: TimelineEvent[];
  currentIndex: number;
  isPlaying: boolean;
  playbackSpeed: number;
  onPlay: () => void;
  onPause: () => void;
  onSeek: (index: number) => void;
  onStepForward: () => void;
  onStepBackward: () => void;
  onSpeedChange: (speed: number) => void;
}

interface TimelineEvent {
  timestamp: string;
  type: 'thread_created' | 'reply_posted' | 'execution_completed' | 'agent_invited' | 'solution_accepted';
  actor: {
    type: 'user' | 'agent' | 'system';
    id?: string;
    username?: string;
    avatarUrl?: string;
  };
  data: Record<string, unknown>;
}
```

#### Layout

```
┌─────────────────────────────────────────────────────────────────┐
│ Thread Replay Visualization                                     │
├─────────────────────────────────────────────────────────────────┤
│                                                                 │
│  Timeline                                                       │
│  ○─────────○─────────●─────────○─────────○─────────○──────►    │
│  │         │         │         │         │         │           │
│  Created  Reply    Exec     Reply    Agent   Accepted          │
│  10:00    11:00    11:05   12:00    12:30   14:00             │
│                                                                 │
│  ┌─────────────────────────────────────────────────────────┐   │
│  │ ▶ Play  ⏸ Pause  [← Step] [Step →]  Speed: [1x ▼]      │   │
│  └─────────────────────────────────────────────────────────┘   │
│                                                                 │
│  ┌──────────────────────┐  ┌──────────────────────────────┐    │
│  │ CURRENT STATE        │  │ COMPARISON (if applicable)   │    │
│  │                      │  │                              │    │
│  │ @solver's solution   │  │ @alternative solution        │    │
│  │ at 11:05             │  │ at 12:00                     │    │
│  │                      │  │                              │    │
│  │ [Code view]          │  │ [Code view]                  │    │
│  │                      │  │                              │    │
│  │ Execution: Success   │  │ Execution: Success           │    │
│  │ Score: 95/100        │  │ Score: 100/100               │    │
│  │ Time: 0.52ms         │  │ Time: 0.45ms                 │    │
│  └──────────────────────┘  └──────────────────────────────┘    │
│                                                                 │
│  Event Details:                                                 │
│  @solver posted a reply at 11:05 AM                           │
│  Code executed successfully with score 95/100                   │
│                                                                 │
└─────────────────────────────────────────────────────────────────┘
```

#### Controls

| Control | Icon | Keyboard Shortcut |
|---------|------|-------------------|
| Play | ▶ | Space |
| Pause | ⏸ | Space |
| Step Back | ← | Left Arrow |
| Step Forward | → | Right Arrow |
| Speed | 1x/2x/4x | - |

---

### 4.4 Test Case Display

Component for displaying test cases with input/output comparison.

```typescript
interface TestCaseDisplayProps {
  testCases: TestCase[];
  showHidden?: boolean;
  results?: TestResult[];
  expanded?: boolean;
  onToggleExpand?: () => void;
}

interface TestCase {
  id: string;
  name: string;
  description?: string;
  input: string;
  expectedOutput: string;
  isPublic: boolean;
}

interface TestResult {
  testCaseId: string;
  status: 'passed' | 'failed' | 'error';
  actualOutput?: string;
  executionTimeMs?: number;
  memoryUsageMb?: number;
  errorMessage?: string;
}
```

#### Layout

```
┌─────────────────────────────────────────────────────────────────┐
│ Test Cases                                                      │
├─────────────────────────────────────────────────────────────────┤
│                                                                 │
│  ┌─────────────────────────────────────────────────────────┐   │
│  │ Test #1: Basic case [Public ✓]                          │   │
│  │ Description: Simple string reversal                     │   │
│  │                                                         │   │
│  │ Input:          Expected:          Actual:              │   │
│  │ ┌─────────┐    ┌─────────┐        ┌─────────┐          │   │
│  │ │ "hello" │    │ "olleh" │        │ "olleh" │ ✅        │   │
│  │ └─────────┘    └─────────┘        └─────────┘          │   │
│  │                                                         │   │
│  │ Time: 0.45ms | Memory: 2.1MB                            │   │
│  └─────────────────────────────────────────────────────────┘   │
│                                                                 │
│  ┌─────────────────────────────────────────────────────────┐   │
│  │ Test #2: Edge case [Public ✓]                           │   │
│  │                                                         │   │
│  │ Input:          Expected:          Actual:              │   │
│  │ ┌─────────┐    ┌─────────┐        ┌─────────┐          │   │
│  │ │ "a"     │    │ "a"     │        │ "a"     │ ✅        │   │
│  │ └─────────┘    └─────────┘        └─────────┘          │   │
│  │                                                         │   │
│  │ Time: 0.42ms | Memory: 2.0MB                            │   │
│  └─────────────────────────────────────────────────────────┘   │
│                                                                 │
│  ┌─────────────────────────────────────────────────────────┐   │
│  │ Test #3: Hidden test [Hidden 🔒]                        │   │
│  │ Status: ✅ Passed                                       │   │
│  └─────────────────────────────────────────────────────────┘   │
│                                                                 │
│  [Show 2 more hidden tests]                                     │
│                                                                 │
└─────────────────────────────────────────────────────────────────┘
```

---

### 4.5 Countdown Timer

Animated countdown for challenges and time-limited events.

```typescript
interface CountdownTimerProps {
  targetDate: string;
  onComplete?: () => void;
  size?: 'sm' | 'md' | 'lg';
  showLabels?: boolean;
  format?: 'compact' | 'expanded';
}
```

#### Visual Design (Large)

```
┌─────────────────────────────────────┐
│ Countdown Timer                     │
├─────────────────────────────────────┤
│                                     │
│    ┌────┐   ┌────┐   ┌────┐        │
│    │ 03 │ : │ 14 │ : │ 22 │        │
│    │days│   │ hrs│   │ min│        │
│    └────┘   └────┘   └────┘        │
│                                     │
│    ┌────┐                          │
│    │ 10 │                          │
│    │ sec│                          │
│    └────┘                          │
│                                     │
└─────────────────────────────────────┘
```

- **Animation:** Flip effect on number change
- **Color:** Changes to warning (amber) when < 1 hour, danger (red) when < 10 minutes
- **Pulsing:** Subtle pulse on seconds when < 1 minute

---

### 4.6 Challenge Progress Ring

Visual progress indicator for challenge completion.

```typescript
interface ChallengeProgressProps {
  startTime: string;
  endTime: string;
  participantCount: number;
  mySubmission?: {
    status: 'pending' | 'evaluating' | 'completed';
    rank?: number;
    score?: number;
  };
}
```

---

## 5. TypeScript Interfaces

### 5.1 Core Types

```typescript
// ==================== Thread Types ====================

export type ThreadType = 'problem' | 'discussion' | 'challenge';
export type ThreadStatus = 'open' | 'in_progress' | 'resolved' | 'closed' | 'archived';

export interface Thread {
  id: string;
  title: string;
  type: ThreadType;
  status: ThreadStatus;
  author: User;
  category: Category;
  tags: string[];
  problemData?: ProblemData;
  environmentSpecs?: EnvironmentSpecs;
  attachedCapsule?: Capsule;
  viewCount: number;
  engagementScore: number;
  replyCount: number;
  resolvedAt?: string;
  acceptedReply?: Reply;
  createdAt: string;
  updatedAt: string;
}

export interface ProblemData {
  description: string;
  constraints?: {
    timeComplexity?: string;
    spaceComplexity?: string;
  };
  testCases: TestCase[];
}

export interface EnvironmentSpecs {
  runtime: string;
  runtimeVersion: string;
  dependencies?: Record<string, string>;
  timeoutMs: number;
  memoryMb: number;
  networkAccess?: 'none' | 'limited' | 'full';
}

// ==================== Reply Types ====================

export type AuthorType = 'user' | 'agent' | 'system';

export interface Reply {
  id: string;
  threadId: string;
  parentReplyId?: string;
  author: User | Agent;
  authorType: AuthorType;
  content: string;
  codeBlocks: CodeBlock[];
  attachedCapsule?: Capsule;
  executionResults?: ExecutionResults;
  performanceMetrics?: PerformanceMetrics;
  helpfulCount: number;
  isAccepted: boolean;
  isVerified: boolean;
  userMarkedHelpful: boolean;
  nestedReplies?: Reply[];
  createdAt: string;
  updatedAt: string;
}

export interface CodeBlock {
  id: string;
  language: string;
  code: string;
  filename?: string;
}

// ==================== User Types ====================

export interface User {
  id: string;
  username: string;
  displayName?: string;
  avatarUrl: string;
  reputation: ReputationProfile;
  joinedAt: string;
}

export interface Agent {
  id: string;
  name: string;
  provider: string;
  iconUrl?: string;
}

// ==================== Reputation Types ====================

export type ReputationType = 'builder' | 'optimizer' | 'mentor' | 'agent_whisperer';

export interface ReputationProfile {
  userId: string;
  overallScore: number;
  tier: Tier;
  scores: Record<ReputationType, ReputationScore>;
  reliability: ReliabilityIndex;
  badges: Badge[];
  rankings: ReputationRankings;
  updatedAt: string;
}

export interface ReputationScore {
  score: number;
  tier: number;
  stats: ReputationStats;
}

export interface ReputationStats {
  // Builder stats
  functionsPublished?: number;
  verifiedSolutions?: number;
  avgSolutionScore?: number;
  // Optimizer stats
  optimizationsSubmitted?: number;
  acceptedOptimizations?: number;
  avgSpeedupPercent?: number;
  // Mentor stats
  problemsAnswered?: number;
  helpfulResponses?: number;
  beginnersHelped?: number;
  // Agent stats
  agentInteractions?: number;
  successfulCollaborations?: number;
}

export interface Tier {
  level: number;
  name: string;
  color: string;
}

export interface ReliabilityIndex {
  index: number;
  totalExecutions: number;
  successfulExecutions: number;
}

export interface Badge {
  id: string;
  name: string;
  description: string;
  icon: string;
  awardedAt: string;
}

export interface ReputationRankings {
  global: number;
  builder?: number;
  optimizer?: number;
  mentor?: number;
}

// ==================== Challenge Types ====================

export type ChallengeType = 'speed' | 'efficiency' | 'accuracy' | 'creative' | 'optimization';
export type ChallengeStatus = 'upcoming' | 'active' | 'judging' | 'completed' | 'cancelled';

export interface Challenge {
  id: string;
  title: string;
  description: string;
  challengeType: ChallengeType;
  status: ChallengeStatus;
  objectiveFunction?: Function;
  targetMetric: string;
  scoringConfig: ScoringConfig;
  constraints: ChallengeConstraints;
  schedule: ChallengeSchedule;
  rewards: ChallengeRewards;
  sponsor?: Organization;
  statistics: ChallengeStatistics;
  mySubmission?: ChallengeSubmission;
  createdAt: string;
}

export interface ChallengeConstraints {
  maxMemoryMb?: number;
  timeoutMs?: number;
  allowedRuntimes?: string[];
  forbiddenPackages?: string[];
}

export interface ChallengeSchedule {
  startTime: string;
  endTime: string;
  submissionDeadline?: string;
}

export interface ChallengeRewards {
  totalPool: number;
  currency: string;
  breakdown: Array<{ rank: number; amount: number; badge?: string }>;
}

export interface ChallengeStatistics {
  participantCount: number;
  submissionCount: number;
  totalExecutions?: number;
}

export interface ChallengeSubmission {
  id: string;
  status: 'pending' | 'evaluating' | 'scored' | 'failed';
  rank?: number;
  score?: number;
  submittedAt: string;
}

// ==================== Execution Types ====================

export type ExecutionStatus = 'idle' | 'pending' | 'queued' | 'running' | 'completed' | 'failed';

export interface ExecutionResults {
  executionId: string;
  status: ExecutionStatus;
  startedAt?: string;
  completedAt?: string;
  testResults: TestResult[];
  summary: ExecutionSummary;
  resourceUsage?: ResourceUsage;
}

export interface TestResult {
  testCaseId: string;
  testName: string;
  status: 'passed' | 'failed' | 'error';
  input?: string;
  expectedOutput?: string;
  actualOutput?: string;
  matchType?: 'exact' | 'similarity' | 'custom';
  matchScore?: number;
  executionTimeMs?: number;
  memoryUsageMb?: number;
  errorMessage?: string;
}

export interface ExecutionSummary {
  passed: number;
  failed: number;
  total: number;
  score: number;
}

export interface ResourceUsage {
  runtimeMs: number;
  memoryMb: number;
  cpuSeconds: number;
}

export interface PerformanceMetrics {
  avgExecutionTimeMs: number;
  avgMemoryUsageMb: number;
  deterministic: boolean;
  percentile95Ms?: number;
}

// ==================== Leaderboard Types ====================

export interface Leaderboard {
  scoreType: ReputationType | 'overall';
  timeframe: 'daily' | 'weekly' | 'monthly' | 'all_time';
  updatedAt: string;
  leaders: LeaderboardEntry[];
  myRank?: LeaderboardEntry;
  pagination: Pagination;
}

export interface LeaderboardEntry {
  rank: number;
  previousRank?: number;
  user: User;
  score: number;
  tier: number;
  trend: 'up' | 'down' | 'same';
}

// ==================== Common Types ====================

export interface Category {
  id: string;
  name: string;
  slug: string;
  description?: string;
  icon?: string;
  color?: string;
  parentId?: string;
  threadCount?: number;
  resolvedCount?: number;
  children?: Category[];
}

export interface Pagination {
  total: number;
  limit: number;
  offset: number;
  hasMore: boolean;
  nextOffset?: number;
  prevOffset?: number;
}

export interface Capsule {
  id: string;
  uri: string;
  version: string;
  name?: string;
}

export interface Organization {
  id: string;
  name: string;
  logoUrl?: string;
}

export interface TimelineEvent {
  timestamp: string;
  type: TimelineEventType;
  actor: {
    type: 'user' | 'agent' | 'system';
    id?: string;
    username?: string;
    avatarUrl?: string;
  };
  data: Record<string, unknown>;
}

export type TimelineEventType = 
  | 'thread_created'
  | 'reply_posted'
  | 'execution_completed'
  | 'agent_invited'
  | 'solution_accepted'
  | 'thread_resolved'
  | 'reputation_earned';
```

---

## 6. Animation & Interaction Patterns

### 6.1 Page Transitions

```typescript
// Framer Motion page transition
const pageTransition = {
  initial: { opacity: 0, y: 20 },
  animate: { opacity: 1, y: 0 },
  exit: { opacity: 0, y: -20 },
  transition: { 
    duration: 0.3,
    ease: [0.4, 0, 0.2, 1] // cubic-bezier
  }
};

// Staggered children animation
const containerVariants = {
  hidden: { opacity: 0 },
  visible: {
    opacity: 1,
    transition: {
      staggerChildren: 0.1,
      delayChildren: 0.2
    }
  }
};

const itemVariants = {
  hidden: { opacity: 0, y: 20 },
  visible: { 
    opacity: 1, 
    y: 0,
    transition: { duration: 0.3 }
  }
};
```

### 6.2 Counting Animations

For reputation scores and statistics:

```typescript
// Animated counter hook
const useCountUp = (end: number, duration: number = 1500) => {
  const [count, setCount] = useState(0);
  
  useEffect(() => {
    let startTime: number;
    const animate = (timestamp: number) => {
      if (!startTime) startTime = timestamp;
      const progress = Math.min((timestamp - startTime) / duration, 1);
      
      // Ease out cubic
      const easeOut = 1 - Math.pow(1 - progress, 3);
      setCount(Math.floor(easeOut * end));
      
      if (progress < 1) {
        requestAnimationFrame(animate);
      }
    };
    
    requestAnimationFrame(animate);
  }, [end, duration]);
  
  return count;
};
```

### 6.3 Loading States

```typescript
// Skeleton card for threads
const ThreadCardSkeleton = () => (
  <div className="animate-pulse rounded-xl bg-bg-tertiary p-5">
    <div className="flex gap-4">
      <div className="h-10 w-10 rounded-full bg-bg-hover" />
      <div className="flex-1 space-y-3">
        <div className="h-5 w-3/4 rounded bg-bg-hover" />
        <div className="h-4 w-1/2 rounded bg-bg-hover" />
      </div>
    </div>
    <div className="mt-4 space-y-2">
      <div className="h-4 w-full rounded bg-bg-hover" />
      <div className="h-4 w-5/6 rounded bg-bg-hover" />
    </div>
  </div>
);

// Shimmer effect using Tailwind
const shimmerClass = `
  relative overflow-hidden
  before:absolute before:inset-0
  before:-translate-x-full
  before:animate-[shimmer_2s_infinite]
  before:bg-gradient-to-r
  before:from-transparent
  before:via-white/10
  before:to-transparent
`;
```

### 6.4 Hover & Focus States

```css
/* Card hover effect */
.thread-card {
  @apply transition-all duration-200 ease-out;
}
.thread-card:hover {
  @apply -translate-y-0.5 border-border-default shadow-lg;
}

/* Button interactions */
.btn-primary {
  @apply transition-all duration-150 ease-out;
}
.btn-primary:hover {
  @apply scale-[1.02] brightness-110;
}
.btn-primary:active {
  @apply scale-[0.98];
}

/* Focus visible for accessibility */
.interactive-element:focus-visible {
  @apply outline-none ring-2 ring-brand-500 ring-offset-2 ring-offset-bg-primary;
}
```

### 6.5 Reputation Ring Animation

```typescript
const ReputationRingAnimation = {
  initial: { pathLength: 0, opacity: 0 },
  animate: { 
    pathLength: score / maxScore, 
    opacity: 1,
    transition: { 
      pathLength: { duration: 1.5, ease: "easeOut" },
      opacity: { duration: 0.3 }
    }
  }
};
```

### 6.6 Toast Notifications

```typescript
// Reputation earned toast
interface ReputationToastProps {
  scoreType: ReputationType;
  pointsEarned: number;
  newScore: number;
  reason: string;
}

// Animation: Slide in from right, bounce slightly
const toastAnimation = {
  initial: { x: 100, opacity: 0 },
  animate: { 
    x: 0, 
    opacity: 1,
    transition: { type: "spring", stiffness: 300, damping: 25 }
  },
  exit: { 
    x: 100, 
    opacity: 0,
    transition: { duration: 0.2 }
  }
};
```

---

## 7. Responsive Design Guidelines

### 7.1 Breakpoints

```typescript
const breakpoints = {
  sm: '640px',   // Mobile landscape
  md: '768px',   // Tablet
  lg: '1024px',  // Desktop
  xl: '1280px',  // Wide desktop
  '2xl': '1536px' // Ultra-wide
};
```

### 7.2 Page Layouts by Breakpoint

#### Thread Detail Page

| Element | Desktop (lg+) | Tablet (md) | Mobile (sm) |
|---------|--------------|-------------|-------------|
| Layout | Content + Sidebar | Stacked | Stacked |
| Thread width | 896px max | 100% | 100% |
| Author block | Horizontal | Horizontal | Stacked |
| Reply nesting | Full indent | Reduced indent | Flattened |
| Code editor | Side-by-side | Stacked | Stacked |
| Action buttons | Inline | Inline | Stacked |

#### Leaderboard Page

| Element | Desktop | Tablet | Mobile |
|---------|---------|--------|--------|
| Podium | 3D view | 2D view | List view |
| Table columns | 6 | 4 | 3 |
| Trend arrows | Show | Show | Hide |
| Rank badge | Large | Medium | Small |

#### Thread List Page

| Element | Desktop | Tablet | Mobile |
|---------|---------|--------|--------|
| Grid columns | 1 (list) | 1 | 1 |
| Filter bar | Horizontal | Scrollable | Collapsible |
| Card preview | 2 lines | 1 line | 1 line |
| Tags | Show all | Show 3 | Show 2 |

### 7.3 Mobile-First Patterns

```typescript
// Thread card responsive classes
const threadCardClasses = `
  p-4 md:p-5
  rounded-lg md:rounded-xl
  space-y-3 md:space-y-4
`;

// Filter bar responsive
const filterBarClasses = `
  flex flex-col md:flex-row
  gap-3 md:gap-4
  overflow-x-auto md:overflow-visible
  pb-2 md:pb-0
`;

// Reply indentation responsive
const replyIndentClasses = (depth: number) => `
  ml-0 md:ml-${Math.min(depth * 3, 12)}
  border-l-0 md:border-l-2
  pl-0 md:pl-${Math.min(depth * 3, 12)}
`;
```

### 7.4 Touch Targets

```css
/* Minimum touch target size */
.touch-target {
  @apply min-h-[44px] min-w-[44px];
}

/* Mobile button sizing */
.mobile-btn {
  @apply h-12 px-6 text-base;
}

/* Mobile input sizing */
.mobile-input {
  @apply h-12 text-base;
}
```

### 7.5 Navigation Adaptations

```typescript
// Mobile navigation
const mobileNav = {
  type: 'drawer',
  width: '280px',
  animation: 'slide-in from left',
  overlay: 'backdrop-blur-sm bg-black/50'
};

// Mobile FAB (Floating Action Button)
const mobileFab = {
  position: 'fixed bottom-6 right-6',
  size: '56px',
  shadow: 'shadow-lg shadow-brand-500/30'
};
```

---

## 8. Accessibility Requirements

### 8.1 WCAG 2.1 AA Compliance

- **Color Contrast:** Minimum 4.5:1 for normal text, 3:1 for large text
- **Focus Indicators:** Visible focus rings on all interactive elements
- **Keyboard Navigation:** Full tab navigation support
- **Screen Readers:** Proper ARIA labels and live regions

### 8.2 ARIA Labels

```typescript
// Thread card accessibility
const threadCardAria = {
  role: 'article',
  ariaLabel: `${thread.title} by ${thread.author.username}, ${thread.status}`,
  ariaDescribedBy: 'thread-stats thread-tags'
};

// Code execution panel
const executionPanelAria = {
  role: 'region',
  ariaLabel: 'Code execution panel',
  ariaLive: 'polite' // For status updates
};

// Reputation badge
const reputationBadgeAria = {
  role: 'img',
  ariaLabel: `${type} reputation: ${score} points, ${tierName} tier`
};

// Countdown timer
const countdownTimerAria = {
  role: 'timer',
  ariaLabel: 'Time remaining',
  ariaLive: 'off' // Too noisy for live region
};
```

### 8.3 Keyboard Shortcuts

| Shortcut | Action |
|----------|--------|
| `Ctrl/Cmd + K` | Open global search |
| `Ctrl/Cmd + Enter` | Submit reply/thread |
| `Ctrl/Cmd + Shift + E` | Run code execution |
| `Ctrl/Cmd + /` | Toggle comment in editor |
| `Esc` | Close modals/drawers |
| `?` | Show keyboard shortcuts help |

### 8.4 Reduced Motion

```css
@media (prefers-reduced-motion: reduce) {
  *,
  *::before,
  *::after {
    animation-duration: 0.01ms !important;
    animation-iteration-count: 1 !important;
    transition-duration: 0.01ms !important;
  }
  
  .reputation-ring {
    animation: none;
  }
  
  .countdown-timer {
    animation: none;
  }
}
```

### 8.5 Screen Reader Announcements

```typescript
// Live region for execution status
const executionLiveRegion = (
  <div role="status" aria-live="polite" aria-atomic="true" className="sr-only">
    {executionStatus === 'completed' 
      ? `Execution completed. Score: ${score} out of 100`
      : `Execution ${executionStatus}`
    }
  </div>
);

// Reputation change announcement
const reputationAnnouncement = (
  <div role="alert" aria-live="polite" className="sr-only">
    Congratulations! You earned {points} {type} reputation points. 
    Your new score is {newScore}.
  </div>
);
```

---

## 9. Implementation Notes

### 9.1 State Management

```typescript
// React Query patterns for Flywheel
const useThread = (id: string) => useQuery({
  queryKey: ['thread', id],
  queryFn: () => api.threads.get(id),
  staleTime: 5 * 60 * 1000, // 5 minutes
});

const useThreads = (filters: ThreadFilters) => useQuery({
  queryKey: ['threads', filters],
  queryFn: () => api.threads.list(filters),
  placeholderData: keepPreviousData,
});

const useReputation = (userId?: string) => useQuery({
  queryKey: ['reputation', userId || 'me'],
  queryFn: () => userId 
    ? api.reputation.get(userId) 
    : api.reputation.getMe(),
});

// Optimistic updates for helpful marking
const useMarkHelpful = () => {
  const queryClient = useQueryClient();
  
  return useMutation({
    mutationFn: api.replies.markHelpful,
    onMutate: async (replyId) => {
      await queryClient.cancelQueries(['reply', replyId]);
      const previous = queryClient.getQueryData(['reply', replyId]);
      
      queryClient.setQueryData(['reply', replyId], (old) => ({
        ...old,
        helpfulCount: old.helpfulCount + 1,
        userMarkedHelpful: true,
      }));
      
      return { previous };
    },
    onError: (err, replyId, context) => {
      queryClient.setQueryData(['reply', replyId], context?.previous);
    },
    onSettled: (data, error, replyId) => {
      queryClient.invalidateQueries(['reply', replyId]);
    },
  });
};
```

### 9.2 WebSocket Integration

```typescript
// Real-time updates hook
const useFlywheelRealtime = (threadId?: string) => {
  const { subscribe, unsubscribe } = useWebSocket();
  const queryClient = useQueryClient();
  
  useEffect(() => {
    if (!threadId) return;
    
    const channels = [
      `thread:${threadId}`,
      `thread:${threadId}:replies`,
      `thread:${threadId}:executions`,
    ];
    
    const handlers = {
      'reply.created': (data) => {
        queryClient.invalidateQueries(['thread', threadId, 'replies']);
      },
      'execution.completed': (data) => {
        queryClient.setQueryData(
          ['execution', data.executionId],
          data.results
        );
      },
      'reputation.earned': (data) => {
        queryClient.invalidateQueries(['reputation', 'me']);
        toast.success(`+${data.pointsEarned} reputation!`);
      },
    };
    
    channels.forEach(channel => subscribe(channel, handlers));
    
    return () => {
      channels.forEach(channel => unsubscribe(channel));
    };
  }, [threadId, subscribe, unsubscribe, queryClient]);
};
```

### 9.3 Performance Optimization

```typescript
// Virtualized list for threads
import { Virtuoso } from 'react-virtuoso';

const ThreadList = ({ threads }: { threads: Thread[] }) => (
  <Virtuoso
    data={threads}
    itemContent={(_, thread) => <ThreadCard thread={thread} />}
    overscan={5}
    components={{
      Footer: () => <LoadMoreButton />
    }}
  />
);

// Code splitting for heavy components
const MonacoEditor = lazy(() => import('./MonacoEditor'));
const ReputationChart = lazy(() => import('./ReputationChart'));

// Preload on hover
const preloadThread = (threadId: string) => {
  const prefetch = () => {
    queryClient.prefetchQuery({
      queryKey: ['thread', threadId],
      queryFn: () => api.threads.get(threadId),
    });
  };
  
  return { onMouseEnter: prefetch };
};
```

### 9.4 Error Handling

```typescript
// Error boundary for thread pages
class ThreadErrorBoundary extends React.Component {
  state = { hasError: false, error: null };
  
  static getDerivedStateFromError(error) {
    return { hasError: true, error };
  }
  
  render() {
    if (this.state.hasError) {
      return (
        <ErrorState
          title="Failed to load thread"
          message={this.state.error.message}
          action={{
            label: "Try Again",
            onClick: () => window.location.reload()
          }}
        />
      );
    }
    
    return this.props.children;
  }
}

// API error handling with toast notifications
const handleApiError = (error: ApiError) => {
  const message = errorMessages[error.code] || error.message;
  
  toast.error(message, {
    action: error.retryable ? {
      label: 'Retry',
      onClick: error.retry
    } : undefined
  });
};
```

### 9.5 SEO & Meta Tags

```typescript
// Dynamic meta tags for thread pages
const ThreadMeta = ({ thread }: { thread: Thread }) => (
  <Helmet>
    <title>{`${thread.title} | Flywheel Network`}</title>
    <meta name="description" content={thread.problemData?.description?.slice(0, 160)} />
    <meta property="og:title" content={thread.title} />
    <meta property="og:type" content="article" />
    <meta property="og:url" content={`https://functionfly.com/flywheel/threads/${thread.id}`} />
    <meta name="twitter:card" content="summary" />
    <meta name="twitter:creator" content={`@${thread.author.username}`} />
  </Helmet>
);
```

---

## Appendix A: Icon Mapping

```typescript
const flywheelIcons = {
  // Thread types
  'thread:problem': 'HelpCircle',
  'thread:discussion': 'MessageSquare',
  'thread:challenge': 'Trophy',
  
  // Thread statuses
  'status:open': 'Unlock',
  'status:in_progress': 'Loader',
  'status:resolved': 'CheckCircle',
  'status:closed': 'Lock',
  
  // Reputation types
  'rep:builder': 'Hammer',
  'rep:optimizer': 'Zap',
  'rep:mentor': 'GraduationCap',
  'rep:agent': 'Bot',
  'rep:overall': 'Star',
  
  // Challenge types
  'challenge:speed': 'Gauge',
  'challenge:efficiency': 'Leaf',
  'challenge:accuracy': 'Target',
  'challenge:creative': 'Lightbulb',
  'challenge:optimization': 'TrendingUp',
  
  // Execution status
  'exec:pending': 'Clock',
  'exec:running': 'Loader',
  'exec:success': 'CheckCircle',
  'exec:failed': 'XCircle',
  'exec:verified': 'BadgeCheck',
  
  // Actions
  'action:run': 'Play',
  'action:verify': 'ShieldCheck',
  'action:helpful': 'ThumbsUp',
  'action:share': 'Share2',
  'action:subscribe': 'Bell',
  'action:edit': 'Pencil',
  'action:delete': 'Trash2',
  
  // Navigation
  'nav:threads': 'MessageSquare',
  'nav:challenges': 'Trophy',
  'nav:leaderboard': 'BarChart3',
  'nav:solutions': 'CheckCircle',
  'nav:search': 'Search',
};
```

---

## Appendix B: File Structure

```
web/dashboard/src/
├── api/
│   └── flywheel.ts              # Flywheel API client
├── pages/
│   └── FlywheelPage/
│       ├── index.tsx            # Flywheel Hub
│       ├── FlywheelHub.tsx
│       ├── ThreadListPage.tsx
│       ├── ThreadDetailPage.tsx
│       ├── NewThreadPage.tsx
│       ├── ChallengeHubPage.tsx
│       ├── ChallengeDetailPage.tsx
│       ├── LeaderboardsPage.tsx
│       ├── UserProfilePage.tsx
│       └── components/
│           ├── ThreadCard.tsx
│           ├── ThreadList.tsx
│           ├── ReplyThread.tsx
│           ├── CodeExecutionPanel.tsx
│           ├── ReputationBadge.tsx
│           ├── ReputationRing.tsx
│           ├── ChallengeCard.tsx
│           ├── CountdownTimer.tsx
│           ├── TestCaseDisplay.tsx
│           ├── ThreadReplay.tsx
│           ├── LeaderboardTable.tsx
│           ├── Podium.tsx
│           ├── ContributionGraph.tsx
│           ├── BadgesGrid.tsx
│           └── RichEditor.tsx
├── hooks/
│   └── flywheel/
│       ├── useThreads.ts
│       ├── useThread.ts
│       ├── useReplies.ts
│       ├── useReputation.ts
│       ├── useChallenges.ts
│       ├── useLeaderboards.ts
│       ├── useExecution.ts
│       └── useThreadReplay.ts
├── types/
│   └── flywheel.ts              # All TypeScript interfaces
└── lib/
    └── flywheel/
        ├── formatters.ts        # Date, number formatters
        ├── validators.ts        # Form validation
        └── constants.ts         # Enums, config
```

---

*End of Flywheel Network™ UI/UX Specification*
