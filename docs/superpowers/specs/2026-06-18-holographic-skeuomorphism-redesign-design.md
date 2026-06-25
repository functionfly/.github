# Holographic Skeuomorphism Redesign — Design Spec

**Status:** Approved and implemented (foundation + global rollout)  
**Author:** Kilo  
**Date:** 2026-06-18  
**Scope:** `web/site` (FunctionFly marketing site)

## Implementation Log

- **2026-06-18** — Foundation completed:
  - Loaded `hs-tokens.css` and `hs.css` globally in `Layout.astro`.
  - Fixed `HSHero`, `ChromaticEdge`, and added `.hs-slider` styles.
  - Added user-facing "Enhanced visuals" toggle in header; wired to `ReducedMotionGate`.
  - Simplified `RefractiveObject` to reliable WebGL; removed experimental WebGPU path.
  - Fixed Vercel adapter build failure by moving `/blog` redirect from `astro.config.mjs` to `vercel.json`.
- **2026-06-18** — Homepage and global chrome:
  - Replaced homepage hero and CTA buttons with `TactileButton`.
  - Converted header to glass (`hs-glass`) and HS tokens; added enhanced-visuals toggle.
  - Converted footer to HS tokens.
- **2026-06-18** — Cross-page rollout:
  - Added `hs-legacy-compat.css` to map legacy page card classes (`ff-pillar-card`, `ff-flow-step`, `ff-split-card`, `ff-endpoint-card`) to HS glass/neumorphic styling across all marketing and legal pages.
- **Build status:** `bun run build` succeeds for all 25 pages.

---

## 1. Context

The Holographic Skeuomorphism (HS) design system is already partially implemented in `web/site/src/components/hs/`. It includes `LightSourceProvider`, `GlassPanel`, `NeumorphicSurface`, tactile controls, `RefractiveObject`, `HolographicBackdrop`, `ReducedMotionGate`, `ScrollParallaxCamera`, and static fallbacks. The homepage (`src/components/HomePage.tsx`) already imports and uses several of these components.

However, the implementation is not yet production-ready:

- `hs-tokens.css` and `hs.css` are not loaded globally, so `--hs-*` CSS custom properties are undefined on most pages and undefined during SSR on the homepage.
- `HSHero` does not actually render `HolographicBackdrop` or `ScrollParallaxCamera` as its docstring claims.
- `TactileSlider` references `.hs-slider`, but no slider styles exist in `hs.css`.
- `ChromaticEdge` sets a `--hs-chromatic-accent` custom property that `hs.css` never consumes.
- Only the homepage uses HS components; all other marketing pages still use the older Velocity-brand design-system classes.

## 2. Goal

Make the HS design system the consistent, production-ready visual language across the entire FunctionFly marketing site, while keeping implementation and maintenance costs low.

## 3. Non-Goals (Cost Guardrails)

To avoid an expensive rewrite or visual over-engineering:

- **No framework migration.** We will keep Astro + React islands. The brief mentions React + Vite/Next.js, but the site is already Astro and the HS library is written for Astro islands.
- **No new design-tool dependencies.** We will not add Tailwind, Spline subscriptions, or paid font licenses. We already have custom CSS (`design-system.css`, `hs-tokens.css`, `hs.css`) and Google Fonts/Commit Mono.
- **No full-page WebGPU/Spline wallpaper.** Real-time refraction is reserved for one signature moment on the homepage hero. All other pages use static or CSS-only atmospheric backgrounds.
- **No rewrite of content.** Existing page structure, copy, and i18n remain. We only replace surface styling (cards, panels, buttons, inputs) with HS equivalents.

## 4. Architecture Decisions

| Decision | Choice | Rationale |
|----------|--------|-----------|
| Framework | Astro 6 + React 19 islands | Existing stack; best SEO/i18n performance for marketing content. |
| Styling | Hand-written CSS custom properties (`hs-tokens.css` + `hs.css`) | Already in place; cheaper and more consistent than adding Tailwind for glass/neumorphic surfaces. |
| 3D/Refraction | `@react-three/fiber` + `@react-three/drei` `MeshTransmissionMaterial` | Already in dependencies; real refraction for the hero only. |
| Renderer | WebGL default with WebGPU feature-detect fallback | `RefractiveObject` currently only checks for WebGPU but still uses the default R3F renderer. We will keep WebGL as the default and only enable WebGPU if explicitly supported and beneficial. |
| Motion | Framer Motion for component press/hover states; CSS for scroll reveals | Framer Motion is already installed; use it only where it meaningfully improves tactile feedback. |
| Icons | Keep existing inline SVGs and emoji icons for now | Avoids adding `lucide-react` and refactoring every page. |

## 5. Foundation Fixes (Do First)

These must be completed before any page rollout because every HS component depends on them.

1. **Load HS tokens and styles globally.** Add `import '../styles/hs-tokens.css'` and `import '../components/hs/hs.css'` to `src/layouts/Layout.astro` so every page has the variables. Remove the local `import './hs/hs.css'` from `HomePage.tsx`.
2. **Fix `HSHero`.** Make it render `HolographicBackdrop` (static fallback by default, Spline optional behind a prop) and wrap refractive children with `ScrollParallaxCamera` only when explicitly enabled.
3. **Add missing `.hs-slider` styles** to `hs.css` so `TactileSlider` has a consistent neumorphic track and thumb.
4. **Fix `ChromaticEdge`.** Update `hs.css` to consume `--hs-chromatic-accent` for the colored edge stop, or remove the unused custom property from `HSHero`/`HSFeatureCard`.
5. **Strengthen `ReducedMotionGate`.** Add a user-facing "Enhanced visuals" toggle in the footer/header that writes to `localStorage` and is read by `ReducedMotionGate`, not just OS-level `prefers-reduced-motion`.
6. **GPU/WebGPU sanity check.** `RefractiveObject` should default to WebGL and only opt into WebGPU when `navigator.gpu` is available and the GPU tier is `high`. Low/medium tiers always see the static fallback.

## 6. Component Adjustments

### Keep as-is
- `LightSourceProvider`, `useLightSource`, `tokens.ts` — foundation is solid.
- `GlassPanel`, `NeumorphicSurface`, `TactileButton`, `TactileToggle`, `TactileInput` — minor CSS polish only.
- `StaticHolographicBackdrop`, `StaticRefractiveObject`, `StaticGlassPanel` — fallbacks are correct.
- `ScrollParallaxCamera` — keep the subtle DOM/CSS parallax; no changes.

### Modify
- `HSHero` — fix composition as described above.
- `HSFeatureCard` — ensure `ChromaticEdge` accent color is actually applied.
- `RefractiveObject` — default to WebGL; gate WebGPU behind capability + high GPU tier.
- `HolographicBackdrop` — default to static gradient; Spline only when a `scene` prop is provided and load succeeds.

### Add
- `HSPageSection`: a reusable section wrapper with consistent vertical rhythm and optional glass/neumorphic background.
- `HSContainer`: max-width + padding wrapper matching the existing marketing site widths.

## 7. Rollout Order (Pages)

We will roll out page by page, reusing existing content and replacing surface components.

1. **Homepage (`src/components/HomePage.tsx`)** — already partially HS. Fix foundation issues first, then polish.
2. **Header + Footer (`src/components/Header.astro`, `src/layouts/Layout.astro`)** — convert header to `GlassPanel` / tactile nav; footer links remain but adopt HS tokens.
3. **High-traffic marketing pages:**
   - `/for-agents` (`ForAgentsPage.tsx`)
   - `/trust` (`TrustPage.tsx`)
   - `/plans` (`PlansPage.tsx`)
   - `/pricing` (`PricingPage.tsx`)
   - `/state-fabric` (`StateFabricPage.tsx`)
   - `/mcp` (`MCPPage` if exists, otherwise Astro page)
4. **Secondary pages:**
   - `/about`, `/careers`, `/contact`, `/partnerships`, `/registry`, `/agent-execution`, `/bundles`
   - Legal/compliance: `/privacy`, `/terms`, `/sla`, `/security`, `/compliance`, `/vulnerability`
5. **Changelog/blog shells** — adopt HS tokens for cards and code blocks; keep content structure.

## 8. Performance & Accessibility Requirements

- **Glass contrast:** Text on `GlassPanel` must meet WCAG AA against the underlying background in both dark and light themes. Test with the worst-case background (e.g., bright gradient behind the panel).
- **Focus rings:** All tactile controls must have a visible `:focus-visible` outline independent of neumorphic shadows.
- **Reduced motion:** `ReducedMotionGate` must disable the refractive hero, parallax, and scroll animations for users who prefer reduced motion or toggle off enhanced visuals.
- **Low-end GPU:** Low/medium GPU tiers get static fallbacks for all 3D/shader content.
- **Backdrop-filter limits:** Cap nested `GlassPanel` depth at 3 in code/docs; warn in code comments.
- **Bundle size:** Keep `@react-three/fiber` and `three` out of the initial page bundle except on the homepage. Use `client:visible` or `client:media` Astro directives where possible.

## 9. Implementation Phases

### Phase 1: Foundation (this session)
- Load HS CSS globally in `Layout.astro`.
- Fix `HSHero`, `ChromaticEdge`, `TactileSlider` styles.
- Add user-facing enhanced-visuals toggle.
- Repair `RefractiveObject` WebGPU/WebGL logic.
- Verify homepage builds and renders correctly.

### Phase 2: Homepage Polish
- Replace remaining legacy hero/button classes with HS components.
- Ensure typewriter + static headline fallback works.
- Verify refractive hero performance on desktop and mobile.

### Phase 3: Header/Footer + Top Pages
- Convert header/footer to HS.
- Roll out HS to `/for-agents`, `/trust`, `/plans`, `/pricing`, `/state-fabric`, `/mcp`.

### Phase 4: Secondary Pages
- Roll out HS to remaining marketing and legal pages.

### Phase 5: QA & Hardening
- Cross-browser test (Chrome, Safari, Firefox).
- Lighthouse/axe accessibility pass.
- GPU fallback verification.
- Build size audit.

## 10. Risks & Mitigations

| Risk | Mitigation |
|------|------------|
| `backdrop-filter` performance on mobile | Cap nested depth; use `StaticGlassPanel` fallback on low GPU. |
| Three.js bundle bloat | Keep R3F components on homepage only; lazy-load elsewhere. |
| Light/dark theme contrast failures | Test every HS surface in both themes; keep `--hs-glass-bg` opaque enough. |
| SSR mismatch / hydration errors | `RefractiveObject` already renders static fallback on server; keep this pattern. |
| Broken existing page styles | Roll out incrementally; revert a page if it regresses. |

## 11. Success Criteria

- `bun run build` succeeds for `web/site` with no new warnings.
- Homepage renders correctly in dark/light mode with working refractive hero.
- All marketing pages use HS tokens for cards, buttons, and inputs.
- Lighthouse Performance score stays ≥ 70 on mobile (marketing baseline).
- axe-core reports no contrast or focus-ring violations on updated pages.

---

**Next step:** Review this spec. Once approved, write the implementation plan and begin Phase 1.
