# Flywheel Network™ Components

This directory contains all components for the Flywheel Network™ - FunctionFly's community-driven knowledge network.

## Directory Structure

```
flywheel/
├── layout/           # Layout components
│   ├── FlywheelLayout.tsx
│   ├── FlywheelSidebar.tsx
│   ├── FlywheelTopBar.tsx
│   └── FlywheelMobileNav.tsx
├── thread/           # Thread components
│   ├── ThreadCard.tsx
│   ├── ThreadHeader.tsx
│   ├── ThreadContent.tsx
│   ├── ThreadReplyComposer.tsx
│   └── ThreadList.tsx
├── execution/        # Code execution components
│   ├── RunCapsuleButton.tsx
│   ├── ExecutionStatusIndicator.tsx
│   ├── RuntimeMetricsDisplay.tsx
│   ├── OutputComparisonViewer.tsx
│   ├── ExecutionTraceTimeline.tsx
│   └── VerifiedFixBadge.tsx
├── reputation/       # Reputation components
│   ├── ReputationBadge.tsx
│   ├── ReputationSummaryCard.tsx
│   ├── SkillDomainBadgeCloud.tsx
│   ├── ReliabilityIndexGauge.tsx
│   └── LeaderboardTable.tsx
├── challenge/        # Challenge components
│   ├── ChallengeCard.tsx
│   ├── ChallengeLeaderboard.tsx
│   └── DailyChallengeCard.tsx
├── ai/               # AI assistant components
│   ├── ThreadSuggestionDropdown.tsx
│   └── AutoFixRecommendationCard.tsx
└── types.ts          # Shared TypeScript types
```

## Design System

- Background: slate-950
- Cards: slate-900 with border-slate-800
- Primary: indigo-500/violet-500
- Success: emerald-500
