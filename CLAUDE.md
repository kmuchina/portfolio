# CLAUDE.md — kmuchina.com design & engineering brief

Read this whole file before doing anything. It is the source of truth for every
decision on this project. If a decision changes, **update this file in the same
change**.

---

## 1. Project summary

- **Who:** Kennedy Muchina Kinuthia, a Graduate Software Engineer in Nairobi, Kenya.
- **What:** a single-page portfolio, with room for project detail pages later
  (`/projects/<slug>/`). A small Go program generates it as a **static site**.
- **Hosting:** GitHub Pages on the custom domain **kmuchina.com** (set up in Phase 8).
- **Design reference:** `design/mockup.png` (local only, gitignored: it shows
  the phone number and the retired email) is a **layout reference only, not a
  spec**. This file overrides it. The site must be **more minimalist than the
  mockup**: more whitespace, fewer boxed cards, hairline dividers instead of
  heavy borders. `design/` is never copied into `dist/`.
- **Experience reference:** `design/experience-prototype.html` (Phase 4) is the
  approved, working prototype of the spine layout and the career graph. The
  site reproduces it (same markup classes, geometry, motion); it is the
  visual reference for that section. Its Google Fonts link, demo bar, stub
  sections and hand-drawn icons are prototype-only and never copied.
- **Sections reference:** `design/sections-prototype.html` (Phase 5) is the
  approved prototype of About, Skills, Projects, Education, Contact and the
  footer (§17). The site reproduces it. Its Google Fonts link, demo bar,
  `<symbol>` sprite of stand-in icons, hard-coded content and blanket
  `!important` reduced-motion rule are prototype-only and never copied.

### Roadmap (8 phases)

| # | Phase | Status |
|---|---|---|
| 0 | Foundations: fonts, colour, motion, theming, tooling, content, this brief | ✅ done (2026-10-30) |
| 1 | Go skeleton & data structs (typed `site.yaml`, templates, layout) | ✅ done (2026-10-30) |
| 2 | Design system components (buttons, card, tags, section shell, icons) | ✅ done (2026-10-30) |
| 3 | Header (sticky, scrolled state), nav, **mobile menu**, hero | ✅ done (2026-10-30) |
| 4 | Section system ("spine" layout, bands) + experience career graph | ✅ done (2026-10-30) |
| 5 | About, skills, projects, education, contact, footer: the sections prototype (§17) | ✅ done (2026-10-30) |
| 6 | Scrollspy (live `aria-current`) and interaction polish | ⬜ |
| 7 | Polish, SEO and images | ⬜ |
| 8 | Deploy to GitHub Pages (Actions, CNAME, custom domain) | ✅ done (2026-10-30): workflow, `CNAME`, `.nojekyll`, `make ci`, Pages + DNS (§12 Deployment) |

---

## 2. Hard stack constraints (non-negotiable)

- **Go** standard library + `html/template`. The **only** third-party Go
  dependency allowed is `gopkg.in/yaml.v3`.
- **CSS:** hand-written vanilla CSS with custom properties. **No** Tailwind,
  Bootstrap, Sass, PostCSS or any other CSS framework or preprocessor.
- **JS:** vanilla only, small, and progressive enhancement. **No** frameworks,
  animation libraries or jQuery. The page must be complete and readable with
  JS off.
- **No Node toolchain** in the build. (`npm pack` was used once to fetch the
  fonts; nothing in the build needs Node.)
- **Everything self-hosted:** fonts, icons and images. **No** runtime requests to
  Google Fonts, CDNs, analytics or any other third party.
- Never add a dependency without asking Kennedy first.

---

## 3. Typography: the "Editorial" pairing

Three families, three roles. Typography carries the personality of the site.

| Role | Family | Token | Used for |
|---|---|---|---|
| Display | **Fraunces** (variable serif: opsz, wght, SOFT, WONK) | `--font-display` | the name, section headings, the pull-quote |
| Body / UI | **Inter** (variable: opsz, wght) | `--font-body` | paragraphs, bullets, buttons, nav, card titles |
| Mono | **JetBrains Mono** (variable: wght 100–800) | `--font-mono` | eyebrow labels, dates, tech tags, small meta |

Fallback stacks (already in the tokens):
- Display: `"Iowan Old Style", Georgia, "Times New Roman", serif`
- Body: `system-ui, -apple-system, "Segoe UI", Roboto, sans-serif`
- Mono: `ui-monospace, "SF Mono", Consolas, "Liberation Mono", monospace`

### Fraunces
- Always `font-optical-sizing: auto` and `font-variation-settings: "SOFT" 50, "WONK" 0`
  (slightly softened, never wonky). Reset `font-feature-settings: normal`, so
  Inter's `ss01`/`cv11` never leak into it.
- **Hero name** (`.display-name`): `--fs-3xl`, weight **600**, letter-spacing
  **-0.02em**, line-height **1.05** (`--lh-tight`).
- **Section headings** (`h2`): `--fs-2xl`, weight **560**, letter-spacing
  **-0.015em**, line-height **1.1**.
- **Italic is used for exactly two things:**
  1. the middle name "*Muchina*" in the hero and the header brand, in
     `--accent-text` (`<h1 class="display-name">Kennedy <em>Muchina</em> Kinuthia</h1>`);
  2. the pull-quote "*Build solutions that make a difference.*" (`.quote`,
     `--fs-xl`, weight 400, `--lh-snug`, curly quotes added by CSS in `--accent`).

### Inter
- Body weight **400**, UI/buttons **500**, card titles (`h3`) **600**.
- **Optical size: Inter renders at its default `opsz` (14)** via
  `font-optical-sizing: none` on `body` (Phase 4, to match the approved
  prototype, whose Inter has no opsz axis). Fraunces elements always set
  `font-optical-sizing: auto`.
- Body line-height **1.65** (`--lh-body`).
- `font-feature-settings: "cv11", "ss01"` on `body`.
- `font-variant-numeric: tabular-nums` (`.tnum`) wherever numbers line up
  (dates in lists, stats).
- Paragraph measure: `max-width: 65ch` (`--measure`, set on `p` in base).

### JetBrains Mono
- **Eyebrow labels** (`.eyebrow`): `--fs-xs` (0.75rem), weight **500**,
  uppercase, letter-spacing **0.14em**, colour `--muted`.
- **Dates/meta** (`.meta`): `--fs-meta` (0.8125rem), weight **400**, tabular nums, `--muted`.
- **Tech tags** (`.tag`): `--fs-xs`, weight **400**, no uppercase.

### Fluid type scale (`00-tokens.css`)

| Token | Value | px range |
|---|---|---|
| `--fs-xs` | `0.75rem` | 12 |
| `--fs-meta` | `0.8125rem` | 13 |
| `--fs-sm` | `0.875rem` | 14 |
| `--fs-base` | `clamp(1rem, 0.96rem + 0.2vw, 1.0625rem)` | 16 → 17 (320–820px viewport) |
| `--fs-lg` | `clamp(1.125rem, 1.05rem + 0.35vw, 1.25rem)` | 18 → 20 (343–914) |
| `--fs-xl` | `clamp(1.25rem, 1.1rem + 0.7vw, 1.625rem)` | 20 → 26 (343–1200) |
| `--fs-2xl` | `clamp(1.75rem, 1.35rem + 1.8vw, 2.75rem)` | 28 → 44 (356–1244) |
| `--fs-3xl` | `clamp(2.5rem, 1.6rem + 4vw, 5rem)` | 40 → 80 (360–1360) |

Line heights: `--lh-tight: 1.05`, `--lh-snug: 1.2`, `--lh-body: 1.65`.

### Font files (`static/fonts/`, latin subset, from Fontsource v5.3.0)

| File | Contents |
|---|---|
| `fraunces-var.woff2` | Fraunces normal, full axes (opsz, wght, SOFT, WONK) |
| `fraunces-var-italic.woff2` | Fraunces italic, full axes |
| `inter-var.woff2` | Inter normal (opsz, wght) |
| `jetbrains-mono-var.woff2` | JetBrains Mono normal (wght 100–800) |
| `*-OFL.txt` | SIL Open Font License 1.1 texts (keep them) |

`@font-face` rules live in `01-fonts.css` (`font-display: swap`, latin
`unicode-range`). **The layout (`partials/head.html`) preloads only Fraunces normal and Inter:**

```html
<link rel="preload" href="/fonts/fraunces-var.woff2" as="font" type="font/woff2" crossorigin>
<link rel="preload" href="/fonts/inter-var.woff2" as="font" type="font/woff2" crossorigin>
```

---

## 4. Colour: mockup blue, dark by default

One ink, one paper, one accent. **Never hard-code a colour value outside
`static/css/00-tokens.css`.** Always use the custom properties.
There are two exceptions, because neither can read CSS variables: the initial
`<meta name="theme-color" content="#0B1220">` (`main.js` re-syncs it from
`--bg`) and `static/favicon.svg`.

### Dark theme (default): `:root, [data-theme="dark"]`, `color-scheme: dark`

| Token | Value | Use |
|---|---|---|
| `--bg` | `#0B1220` | page background (deep navy) |
| `--surface` | `#111A2E` | cards, raised areas |
| `--surface-2` | `#111B30` | alternating section bands (Phase 4: stronger contrast with `--bg`) |
| `--ink` | `#E6EAF2` | primary text |
| `--muted` | `#9AA6BD` | secondary text, labels |
| `--line` | `#1E2940` | hairline borders/dividers |
| `--accent` | `#3B82F6` | brand blue, decorative / large text only |
| `--accent-strong` | `#2563EB` | primary button fill |
| `--accent-strong-hover` | `#1D4ED8` | primary button hover |
| `--accent-text` | `#60A5FA` | links & accent text |
| `--accent-soft` | `#172746` | icon tiles, "Current" tag background |
| `--on-accent` | `#FFFFFF` | text on `--accent-strong` |
| `--line-strong` | `#2C3A57` | hover borders (cards, ghost button), toggle ring |
| `--focus` | `#60A5FA` | focus ring |
| `--shadow` | `0 1px 2px rgb(0 0 0 / .3), 0 8px 24px rgb(0 0 0 / .25)` | card hover only |

### Light theme: `[data-theme="light"]`, `color-scheme: light`

| Token | Value |
|---|---|
| `--bg` | `#F7F8FA` |
| `--surface` | `#FFFFFF` |
| `--surface-2` | `#EBEFF5` |
| `--ink` | `#0F1A2E` |
| `--muted` | `#4A5568` |
| `--line` | `#E2E7EF` |
| `--accent` | `#3B82F6` |
| `--accent-strong` | `#2563EB` |
| `--accent-strong-hover` | `#1D4ED8` |
| `--accent-text` | `#1D4ED8` |
| `--accent-soft` | `#E8F0FE` |
| `--on-accent` | `#FFFFFF` |
| `--line-strong` | `#CBD3E0` |
| `--focus` | `#2563EB` |
| `--shadow` | `0 1px 2px rgb(15 26 46 / .06), 0 8px 24px rgb(15 26 46 / .08)` |

### Rules
- **Accent ≤ 5% of any screen.** It appears only on links (and `.link-arrow`),
  the primary button, the active nav item, focus rings, the "*Muchina*"
  highlight, the hero eyebrow (`.eyebrow--accent`), `.tile` icons, `.list`
  dots, the `.quote` marks, the "Current" tag (`.tag--accent`) and, since
  Phase 5, skill evidence (count badge, open pill), the project kind, the
  "First Class Honours" line, the award icon, the lit meter bars, the
  email's 2px underline and the phone stage's glow. Everything else is
  neutral.
- **`--accent-strong` vs `--accent-text`:** white on `#3B82F6` is only about 3.7:1,
  so **buttons use `--accent-strong`** (with `--on-accent` text) and
  **body-size accent text uses `--accent-text`**, which changes per theme.
  `--accent` itself is only for decorative use or large text (the `.quote`
  marks, `.list` dots, the hero glow). Since Phase 3 the "Muchina" highlight
  uses `--accent-text` too (≥ 4.5:1 in both themes).
- `::selection` = accent at 30% (`color-mix(in srgb, var(--accent) 30%, transparent)`).
- The hero **follows the theme**; it is not forced dark. Section rhythm comes
  from alternating `--bg` and `--surface-2`.
- The decorative gradients on the site are the hero glow (a radial
  `--accent` at ≤ 8%, `color-mix`, with eased stops so no disc edge or
  banding shows) and, since Phase 5, the project phone stage (a radial
  `--accent` at 14% from the bottom edge, as in the prototype). (The career graph's dashed lines and fading stub use
  `repeating-linear-gradient` / mask gradients as drawing techniques, not
  decoration.)
- **Employment-type colours (Phase 4).** `--type-<kind>` and
  `--type-<kind>-soft` for internship (= `--accent-text` / `--accent-soft`),
  volunteer (dark `#34D399` / `#10291F`, light `#047857` / `#E3F5EC`),
  contract (dark `#FBBF24` / `#2A220C`, light `#AD5009` / `#FDF3E1`; light
  nudged from `#B45309`, which was 4.35:1 on `--surface-2`), freelance (dark
  `#A78BFA` / `#1E1A38`, light `#6D28D9` / `#F0EBFE`), full-time, part-time
  and milestone (= `--ink` / `--surface`). Contract and freelance are unused
  today. **These are the only hues besides the accent, and they may appear
  only inside the experience section and its legend.** Rows set local `--c`
  / `--soft`; components read those, never `--type-*` directly.
- In dark mode, use borders (`--line`), not shadows. Shadows appear only on
  hover, with one exception: the project phone's resting drop shadow
  (`--shadow-phone`, Phase 5), which grounds the device on its stage.
- **Project phones (Phase 5):** `--screen` / `--screen-ink` (dark `#0E1526` /
  `#C9D3E6`, light `#F4F6FA` / `#2B3648`) are the placeholder app screen and
  its ink; `--leaf` (dark `#22C55E`, light `#15803D`) is the Apple Care
  scanner green; `--shadow-phone` is the phone's drop shadow. They appear
  only inside `.phone`.
- **No `color-mix()` for text colours.** Text colours must be real tokens so
  `tools/contrast` can verify them. `color-mix()` is fine for borders and
  backgrounds that carry no text (e.g. `::selection`, the faint prose underline).
- **`make contrast` must pass** (every text pair ≥ 4.5:1 in both themes) after
  any colour change. Body-text pairs checked: ink, muted and accent-text on
  bg / surface / surface-2; accent-text and ink on accent-soft; on-accent on
  both button fills; **every `--type-*` on `--surface-2` and on its own
  `-soft`** (the tool resolves `var(--alias)` tokens); and (Phase 5)
  screen-ink and leaf on screen. Large-text pairs: accent on bg / surface-2
  (≥ 3:1). The scanner's 9px "Analysing" pill sits on a `color-mix` tint of
  `--leaf`; it is part of the decorative, `aria-hidden` placeholder picture.

---

## 5. Theming: dark / system / light (default dark)

- The preference is stored in `localStorage["theme"]` as `"dark" | "system" | "light"`.
  **Nothing stored means `"dark"`.** No-JS visitors get dark, because `:root` = dark.
- **`static/js/theme-init.js`** is inlined in `<head>` **before any CSS**, to
  prevent a flash of the wrong theme. It reads the preference (storage wrapped
  in try/catch), resolves `"system"` with `matchMedia('(prefers-color-scheme: light)')`,
  sets `html[data-theme]` to the **resolved** value and `html[data-theme-pref]`
  to the **preference**, and adds `html.js`.
- **Toggle control** (logic in `main.js`; markup in
  `templates/partials/theme-toggle.html`; styled as the `.theme-toggle`
  component). **There are two instances per page** (Phase 3): one in the
  header (≥ 1024px) and one in the mobile menu. The partial takes
  `(dict "Theme" .Site.UI.Theme "Label" …)`; each instance gets its own
  radiogroup label (`ui.theme.label`, `ui.theme.menu_label`; validation
  requires them to differ). `main.js` initialises every `[data-theme-toggle]`
  and `applyTheme` updates `aria-checked` and the roving tabindex in all of
  them, so they always agree. Without JS the toggle is hidden (it couldn't work).
  - A segmented control with options in this order: **Dark (Lucide `moon`),
    System (`monitor`), Light (`sun`)**.
  - `<div role="radiogroup" aria-label="Colour theme" data-theme-toggle>` holds
    three `<button type="button" role="radio" aria-checked data-value="dark|system|light" title="…">`,
    each with an `aria-hidden` SVG (`{{icon}}`) and an `.sr-only` text label (labels come from `ui.theme`).
  - **Roving tabindex:** only the checked option has `tabindex="0"`.
    Arrow keys move focus *and* selection (with wrap-around); Home and End jump
    to the ends.
  - On change, `main.js` saves the preference, applies `data-theme`, updates
    `aria-checked` and `tabindex`, sets `<meta name="theme-color">` to the
    computed `--bg`, and dispatches a `themechange` event on `document`.
  - In **system** mode it listens to `prefers-color-scheme` and re-applies live.
  - **Cross-fade:** `main.js` adds `html.theme-fade` on the first user toggle.
    That class turns `--theme-fade` into colour transitions at `--dur-fast`, so
    the initial page load never animates.

---

## 6. Motion: subtle, purposeful, accessible

Motion only clarifies; it never decorates. **Only `opacity` and `transform`
move.** Colour properties (`color`, `background-color`, `border-color`, `fill`,
`stroke`) may cross-fade at `--dur-fast`. Never animate layout properties
(width, height, top, margin…) or `box-shadow`; fade a shadow pseudo-element's
`opacity` instead.

### Tokens
```css
--ease-out: cubic-bezier(0.22, 1, 0.36, 1);
--dur-fast: 180ms;   /* hovers, colour changes, underlines */
--dur-slow: 600ms;   /* reveals, hero entrance */
--stagger: 80ms;     /* delay between items in a group */
--reveal-distance: 16px;
--transition-ui: var(--theme-fade), color/background-color/border-color/
                 text-decoration-color/opacity/transform at --dur-fast --ease-out;
```
**One easing curve and two durations for the entire site.** Nothing else.
**Every component that transitions uses `transition: var(--transition-ui)`**,
which already starts with `--theme-fade`. Because it is one shared list, no
layer can silently drop another layer's transitioned properties.

### The complete motion vocabulary (nothing beyond this, ever)

All of it is implemented in `static/css/09-motion.css` (the top cascade layer):

1. **Hero entrance on load (CSS only, Phase 3):** each hero text item has
   `data-hero` and an inline `style="--i: N"` (eyebrow 0, name 1, role and
   summary 2, CTAs, social row and quote 3). Under `html.js` they run
   `@keyframes hero-in` (opacity 0→1, `translateY(--reveal-distance)`→0,
   `--dur-slow`, `--ease-out`, `animation-fill-mode: both`, delay
   `calc(var(--i) * var(--stagger))`): ≈ 840ms in total. The photo
   (`[data-hero-photo]`) runs `hero-photo-in` (opacity + scale 1.02→1) after
   one `--stagger`. No JS, and the hero has **no** `data-reveal`.
2. **Scroll reveal:** `[data-reveal]` elements, or the direct children of
   `[data-reveal-group]`, are `opacity: 0` until revealed, **only under
   `html.js`** (content is never hidden without JS). An `IntersectionObserver`
   (threshold `0.15`, rootMargin `0px 0px -10% 0px`) adds `.is-visible` once,
   then unobserves. That starts a one-shot `@keyframes reveal` (from
   `opacity: 0; translateY(var(--reveal-distance))`, `--dur-slow`,
   `animation-fill-mode: backwards`). `main.js` sets `--i` on each group child;
   the delay is `calc(var(--i) * var(--stagger))`. **Why an animation, not a
   transition (Phase 2):** a reveal transition on a card would also govern its
   hover transitions (600ms plus the stagger delay). A finished animation
   leaves no trace.
3. **Card hover** (`.lift`, and `.card--interactive`): `translateY(-2px)`,
   plus `--shadow` on `::after` fading in via opacity, plus the border shifting
   to `--line-strong`. The movement only applies under `@media (hover: hover)`;
   `:focus-within` also gets the `--line-strong` border.
4. **Link / nav underline** (`.underline-grow`, and the `.link` component): a
   1px pseudo-element that grows with `transform: scaleX(0 → 1)`,
   `transform-origin: left`, on hover, focus-visible or `[aria-current]`. Use
   it for single-line links (nav, footer, CTAs). Links in running text
   (`.prose a`) keep the native underline, whose colour fades from faint to
   full on hover (a pseudo-element can't follow a wrapped line).
5. **Button press** (`.press`): `transform: scale(0.98)` on `:active`.
6. **Theme cross-fade:** see §5. The default `*, *::before, *::after {
   transition: var(--theme-fade) }` lives in the **base** layer (so components
   can override it). Any element that declares its own `transition` uses
   `var(--transition-ui)`, or lists `var(--theme-fade)` **first**.
7. **Arrow nudge** (Phase 2): the trailing icon of `.link-arrow` moves
   `translateX(3px)` over `--dur-fast` on hover or focus-visible.
8. **Mobile menu** (Phase 3): the open `<dialog class="menu">` fades in and
   drops `--space-2` (8px) at `--dur-fast`; its `::backdrop` fades in.
9. **Smooth anchor scrolling:** `html { scroll-behavior: smooth }` in the
   motion layer, only under `prefers-reduced-motion: no-preference`.

`.underline-grow` reads two optional variables: `--underline-inset` (move the
line in from the sides, e.g. past a link's padding) and `--underline-offset`
(distance from the box bottom). The nav sets both.

10. **Career graph: the line draws itself as you scroll** (Phase 4). Each
    row's segment ink (`.seg::after`) is revealed by `clip-path: inset()`
    and the merge curve by `stroke-dashoffset`, on a **scroll-driven
    animation** (`view-timeline: --row` on `.career__row`; `animation-range`
    with **px offsets**: above-segments and curves `cover 150px` → `cover
    300px`, full segments `cover 150px` → `exit 0%`, below-segments `cover
    300px` → `exit 0%`). The tip leads the reader at about two-thirds of the
    viewport. `clip-path` and `stroke-dashoffset` are the amendment to
    "only opacity and transform": they are the only way to draw a line, and
    they follow the scroll, not a clock. **Fallback** (`@supports not
    (animation-timeline: view())`): `main.js` adds `.is-lit` to each row
    once (IntersectionObserver, rootMargin `0px 0px -35% 0px`) and the same
    visuals transition once (500ms).
11. **Node ignite:** the node fill scales 0 → 1.35 → 1 as the ink reaches it
    (scroll-driven, `cover 290px` → `cover 370px`); in the fallback it
    transitions on **`--ease-spring`** (`cubic-bezier(0.34, 1.56, 0.64, 1)`),
    the only overshoot easing on the site, used **only here**.
12. **The current role breathes:** a `::before` halo on `.node--current`
    scales 1 → 2.8 and fades .55 → 0, 2.4s, `infinite`. **This is the one
    looping animation on the whole site** (the single exception to "nothing
    loops"). `main.js` pauses it (`.is-paused` → `animation-play-state:
    paused`) whenever the node is off-screen.
13. **Focus follows the reader:** hovering or focusing an entry
    (`:has(:hover, :focus-within)`) grows its node (`scale: 1.3`) and dims
    the other rows' graph cells to 35%. CSS only; text never dims.
14. **Legend filtering:** pressing a legend button highlights one kind:
    non-matching rows get `.is-quiet` (tile greyed, text to `--muted`,
    still ≥ 4.5:1) and the other lane drops to 20% opacity per lane (so the
    highlighted lane stays continuous, e.g. the volunteer lane through the
    KRB merge). Pressing it again or "All" resets. Dims, never hides.

15. **Project phone rise** (Phase 5): hovering a project card (`(hover:
    hover)` only) lifts the card 2px (item 3, `.lift`) and raises its phone
    `translateY(-10px)` over `--dur-slow`.
16. **Email underline** (Phase 5, an extension of item 4): the contact
    address's 2px `--accent` underline rests at `scaleX(0.18)` and grows to
    `scaleX(1)` on hover/focus over `--dur-slow`.
17. **Diagonal nudge** (Phase 5, a sibling of item 7): `arrow-up-right`
    icons (`.link-arrow--external`, `.project__link`, `.channels__go`) move
    `translate(2px, -2px)` over `--dur-fast`; the footer's Back to top
    arrow moves `translateY(-2px)`.
18. **Copy confirmation** (Phase 5): the copy button shows "Copied" plus a
    check for 2.2s (`main.js`), then resets. The skill "used in" panel
    appears with **no** animation (it toggles `hidden`).

**Banned:** parallax, scroll-jacking, typing/typewriter text, cursor followers,
particles, auto-playing carousels, bounce/elastic easings (`--ease-spring`
above is the one deliberate, graph-only exception), and anything that loops
(except item 12).

### Reduced motion
Under `@media (prefers-reduced-motion: reduce)`, revealed elements are visible
immediately (`opacity: 1`, `animation: none`), the hero entrance and the menu
animation are off, scrolling is instant, hover lifts, press scaling and
the arrow nudges, the project phone rise and the email underline's growth
(it switches at once) are disabled, the career graph is fully drawn and lit and
static (no scroll drawing, no ignite, no pulse; hover and filter changes are
instant), the underline fades in instead of growing, and
animations are cut to 0.01ms. Only short opacity and colour changes remain.
`main.js` skips the observer entirely and marks every target `.is-visible`.
Because `motion` is the top layer, this block needs **no `!important`**; its
selectors repeat the ones they cancel, so they win on source order.

---

## 7. Layout & spacing tokens

Spacing uses a 4px base (`--space-N` = N × 4px):

| Token | rem | px |
|---|---|---|
| `--space-1` | 0.25 | 4 |
| `--space-2` | 0.5 | 8 |
| `--space-3` | 0.75 | 12 |
| `--space-4` | 1 | 16 |
| `--space-5` | 1.25 | 20 |
| `--space-6` | 1.5 | 24 |
| `--space-8` | 2 | 32 |
| `--space-10` | 2.5 | 40 |
| `--space-12` | 3 | 48 |
| `--space-16` | 4 | 64 |
| `--space-20` | 5 | 80 |
| `--space-24` | 6 | 96 |

- Icons: `--icon-sm: 16px` (buttons, link-arrows, toggle), `--icon-md: 20px`
  (default, icon buttons, media tiles), `--icon-lg: 24px`.
- Touch: `--tap: 44px`, the minimum size of buttons and icon buttons. Every
  interactive element is ≥ 40px; inline links in running text are exempt
  (WCAG 2.5.8's inline exception).
- Section padding: `--section-pad: clamp(4rem, 3rem + 5vw, 8rem)`.
- Container: `.container` is at most `--container: 1200px` wide **including**
  its gutters (`padding-inline: var(--gutter)`, `--gutter: clamp(1.25rem, 4vw,
  2rem)`), so content is up to 1136px. (Phase 4: was 1120px of content; now
  matches the prototype.)
- Radii: `--radius-sm: 6px`, `--radius: 12px`, `--radius-pill: 999px`.
- Breakpoints are mobile-first: **640px, 900px, 1200px**, plus **1024px, the
  nav breakpoint** (desktop links + toggle from 1024px; menu button below).
  Use them literally in `@media (min-width: …)`, because custom properties
  don't work in media queries. (The hero also stacks its CTAs below 400px.)
- `--header-h`: **64px**, **72px from 1024px** (tokens). The header's 1px
  border is inside it. Every `[id]` has `scroll-margin-top: calc(var(--header-h)
  + var(--space-4))` (base layer), so anchors never land under the header.

### The spine: section system (Phase 4, `04-layout.css`)
Every non-hero section:
```html
<section id="…" class="section [section--alt] …" aria-labelledby="…-title">
  <div class="container section__grid">
    <div class="section__aside" data-reveal>eyebrow · h2.section__title · extras · .section__action</div>
    <div class="section__body">the section's content</div>
  </div>
</section>
```
- **Separation:** each section is a band: `--section-pad` above and below, a
  full-bleed `1px --line` top border (not on the hero), and `--band` as its
  background (`--bg`, or `--surface-2` with `.section--alt`). Nothing inside
  a section comes close to the gap between sections.
- **Band order:** Hero `--bg` (no border) · About `--bg` · Experience alt ·
  Skills `--bg` · Projects alt · Education `--bg` · Contact alt.
- **From 1024px:** `.section__grid` is `minmax(15rem, 1fr) 2.6fr` with gap
  `clamp(2rem, 5vw, 5rem)`; `.section__aside` is sticky at `calc(var(--header-h)
  + var(--space-8))`, so the title stays in view while the body scrolls. Below
  1024px they stack (gap `--space-10`) and nothing is sticky.
- `.section__body` takes the remaining width; paragraphs still cap at 65ch.
- The eyebrow is optional (About has none: `.section__title:first-child`
  drops its top margin). `.section__lead` is an optional `--muted` line
  under the title, 30ch (Skills, Contact).
- The action link (`.link-arrow.section__action.tap-target`) lives in the
  aside: a 16px box `--space-8` below what precedes it, with the 44px hit
  area from `.tap-target`.

### Layout primitives (`04-layout.css`)
Sections compose from these instead of writing one-off CSS. **Vary them with
custom properties, not modifier classes.** The variables are registered with
`@property … inherits: false`, so a nested primitive never picks up its
parent's override.

| Class | Does | Variables (default) |
|---|---|---|
| `.container` | max content width + gutter | (`--container`, `--gutter`) |
| `.stack` | vertical flow (flex column + `gap`) | `--stack-space` (`--space-4`) |
| `.cluster` | wrapping row, centred | `--cluster-space` (`--space-3`) |
| `.grid-auto` | as many columns as fit | `--grid-min` (16rem), `--grid-gap` (`--space-6`) |
| `.split` | 1 column; 2 from 900px | `--split-ratio` (`1fr 1fr`); gap `clamp(--space-8, 5vw, --space-12)` |

`.stack` uses `gap`, not `> * + * { margin-block-start }`: with margins the
variable would be read on the child, so a nested stack's own override would
leak into its outer spacing. Children's own margins don't collapse into the gap.

---

## 8. Minimalism rules

- Use whitespace instead of boxes. Separate content with hairline `--line`
  dividers (1px).
- **At most one card style** in the whole site: `.card`. **Cards are for
  projects and the degree only** (Phase 5; the about facts became a
  hairline fact sheet). Experience, skills, the school column and the about
  section are not cards; they use whitespace and hairline dividers.
- No gradients, except the subtle hero backdrop and the project phone stage.
- **No company or organisation logos** (employers, schools). Use Lucide
  icons or text instead. Technology logos on skill pills (Kotlin, Git,
  Cisco…) are fine: they name a tool, monochrome in `currentColor`.
- Content width ≤ 65ch.
- When in doubt, remove it.

---

## 9. Icons

**Decision (Phase 1): one SVG file per icon in `templates/icons/<name>.svg`,
inlined with the `icon` template func.** Because the SVG is inline,
`currentColor` works, there are no extra requests, and the build can check
every icon name.

- **Inline SVG only.** No icon fonts, no `<img>` icons and no sprite.
- **Usage:** `{{icon "github"}}` gives a decorative icon
  (`class="icon" aria-hidden="true" focusable="false"`).
  `{{icon "github" "GitHub"}}` gives a meaningful icon (`class="icon" role="img"`
  plus `<title>GitHub</title>`). Use the label form only when the icon is the
  only content of a link or button. An unknown name fails the build.
- **Normalised files:** one line each, with a `viewBox` and no
  `width`/`height`/`class`/`stroke-width`. Lucide icons carry
  `fill="none" stroke="currentColor"` with round caps and joins; Simple Icons
  carry `fill="currentColor"`.
- **Size and stroke come from CSS:** `.icon` (in `05-components.css`) is
  `var(--icon-size, var(--icon-md))` square with `stroke-width: 1.5`.
  Components set `--icon-size` (it inherits), e.g. `.btn { --icon-size: var(--icon-sm) }`.
- **UI icons, from Lucide 1.65.0 (ISC):** `code smartphone brain graduation-cap
  map-pin users wrench layers phone mail download arrow-up-right arrow-right
  moon sun monitor menu x external-link calendar`, plus (Phase 4, the
  experience org tiles) `flower-2 route solar-panel trophy`, plus (Phase 5)
  `book-open award globe clock briefcase copy check arrow-up plus leaf`
  and the skill icons below. (`phone` is still in the folder but unused:
  there is no phone number on the site.)
- **`linkedin`:** from Lucide **0.577.0**, the last Lucide release before it
  removed brand icons. Simple Icons has also removed LinkedIn.
- **Brands, from Simple Icons 16.36.0 (CC0):** `github`, `googleplay`, plus
  (Phase 5, skill pills) `kotlin jetpackcompose javascript nodedotjs express
  html5 css cplusplus python tensorflow git androidstudio cisco googlemeet
  android ios`. **Never** use an older Simple Icons release to bring back a
  removed brand, and never draw a logo by hand.
- **Skill icons (Phase 5).** Every skill pill has an icon (validation
  requires it), monochrome `currentColor` at 16px, no brand colours.

  | Skill | Icon file | Source |
  |---|---|---|
  | Kotlin | `kotlin` | Simple Icons |
  | Jetpack Compose | `jetpackcompose` | Simple Icons |
  | JavaScript | `javascript` | Simple Icons |
  | Node.js | `nodedotjs` | Simple Icons |
  | Express | `express` | Simple Icons |
  | HTML5 | `html5` | Simple Icons |
  | CSS3 | `css` | Simple Icons |
  | C++ | `cplusplus` | Simple Icons |
  | Python | `python` | Simple Icons |
  | TensorFlow Lite | `tensorflow` | Simple Icons |
  | CNN | `network` | Lucide |
  | Git & GitHub | `git` | Simple Icons |
  | Android Studio | `androidstudio` | Simple Icons |
  | VS Code | `code-xml` | Lucide fallback (Simple Icons removed `visualstudiocode`) |
  | Business Central | `building-2` | Lucide fallback (Simple Icons removed `dynamics365`) |
  | Cisco networking | `cisco` | Simple Icons |
  | Slack | `message-square` | Lucide fallback (Simple Icons removed `slack`) |
  | Google Meet | `googlemeet` | Simple Icons |
  | REST APIs | `webhook` | Lucide |
  | Android | `android` | Simple Icons |
  | iOS | `ios` | Simple Icons |
  | Flow architecture | `workflow` | Lucide |
  | Backend | `server` | Lucide |
  | Full-stack | `layers` | Lucide |
  | Project management | `kanban` | Lucide |
  | Team leadership | `users` | Lucide |
  | Self-directed | `compass` | Lucide |
  | Accountable | `shield-check` | Lucide |
  | Analytical | `chart-line` | Lucide |
- **Licences:** in `templates/icons/LICENSES/`, which also has a README of
  sources and versions.
- **Adding an icon:** fetch it once (`npm pack lucide-static` or `npm pack
  simple-icons` into a scratch folder), normalise it to the format above, and
  copy only that file.

---

## 10. Content rules

- **All copy lives in `content/site.yaml`**, the single source of truth.
  Templates never hard-code text. Interface copy (skip link, nav labels,
  button labels, theme-toggle labels, "Current", "Languages", 404 text) lives
  in the **`ui:`** block. The only exception is the dev-only styleguide page.
- The contact email is always **muchina537@gmail.com**. The build **fails** if
  `kmuchina22` appears anywhere in the content.
- **No phone number on the site** (Kennedy's decision, 2026-10-30). There
  is no phone field in the content model (`contact.phone`, `site.show_phone`
  and `TelHref` were removed in Phase 5) and no template branch renders one.
  A build test fails if any file in `dist/` contains `tel:`, `703 537`,
  `703537` or `+254` (also as `&#43;254`, html/template's escape).
- **No references section.** References are deliberately left off the public
  site *and* the résumé; Kennedy provides three referees on request.
- Alt text comes from the YAML: `person.photo_alt`; a project screenshot's alt
  is `ui.projects.screenshot_alt` + the title ("Screenshot of Apple Care").
- **Section actions (Phase 2):** `experience.action` and `projects.action` are
  optional `{ label, url }`, rendered as a `.link-arrow` in the section header.
  `url` is either an absolute `https://` URL or a file path under `static/`
  (rendered through `{{asset}}`; `Action.IsExternal()` decides).
- **Navigation (Phase 3):** `ui.nav` (`label` + `href`) is the single list
  behind both the header nav and the mobile menu. "Home" was dropped (the
  brand links to `#top`). Validation checks every `href` against the section
  ids; a build test checks every rendered link has a matching `id`. Also in
  `ui:` since Phase 3: `nav_label` ("Primary"), `home_label` (the brand's
  accessible name is "<full name> — <home_label>"), `menu.{label, open,
  close}`, and `theme.menu_label`.
- **Experience schema (Phase 4).** Each `experience.items[]` entry:
  `title`, `org`, `type`, `icon`, `location`, `start`, `end`, `bullets`,
  optional `skills`.
  - `type` is one of **Internship | Volunteer | Contract | Freelance |
    Full-time | Part-time** (as on LinkedIn); anything else fails the build.
  - **No "intern" in titles:** a title matching `(?i)\bintern(s|ship)?\b`
    fails the build ("International" is fine). "Intern" is the employment
    type, never part of the title.
  - `start` / `end` are `YYYY-MM`; `end` may be `present` (the current role;
    there is no `current` flag). `end` before `start` fails.
  - `icon` is a Lucide icon standing for the organisation (no logos).
  - Bullets may use **`**bold**`** (for metrics, as in the PDF). That is the
    only inline syntax: `internal/textfmt` HTML-escapes everything, then turns
    each closed `**…**` into `<strong>`; an unclosed `**` stays literal.
    Templates call it as `{{bold .}}`. No Markdown dependency.
  - `skills` terms come from the bullets; nothing invented.
- **Job anchors (Phase 5):** each job `<article>` has `id="job-<slug>"`, the
  slug of `org` up to its first comma (`Job.ID()`: "Kenya Roads Board,
  Ministry of…" → `job-kenya-roads-board`). Validation keeps them unique.
- **Milestones:** an education item with `timeline: { date: "YYYY-MM", title,
  detail }` also appears in the career graph as a milestone row.
- `ui.career`: `highlight`, `legend_label`, `all`, `milestone`, `present`,
  `skills_label` (the graph's interface copy). Durations ("1 yr 1 mo") and the
  summary sentence ("roles since 2023, up to three at once") are English
  generated in Go (`internal/career`).
- **Phase 5 schema** (details in §17): `about.eyebrow` optional, `about.body`
  takes `**bold**`, `about.focus[]` `{label, icon, detail}`, `about.facts[]`
  `{label, value, detail}`; `skills.lead`, `skills.groups[].note`,
  `skills.groups[].items[]` `{name, icon}`; projects `metric: {value,
  label}`, `note`, `screenshot`, `placeholder` (`journal | scanner |
  none`), no more `image` / `image_alt`; education `date` (`YYYY-MM`) or
  `start` / `end` (`YYYY`), `icon`, `honours`, `honours_detail`, `award:
  {title, detail}`, `grades[]` `{label, value}` (no more `highlights`);
  `languages[].level` from LinkedIn's scale; `contact.timezone: {iana,
  label}`, `contact.links[].handle`; and the `ui.about`,
  `ui.skills`, `ui.projects`, `ui.contact`, `ui.footer` blocks.
- Fields added in Phase 1: `person.greeting` (the hero eyebrow, "Hello, I'm"),
  `contact.eyebrow`, `contact.heading`, `contact.body`,
  `projects.items[].image_alt`, and the `ui:` block.
- Decoding is **strict** (`KnownFields`), so a misspelt key fails the build.
  Every new YAML key needs a struct field in `internal/site/types.go`.

---

## 11. Accessibility baseline

- Semantic landmarks (`header`, `nav`, `main`, `footer`), **exactly one `<h1>`**
  (the hero name; on the 404 and styleguide pages, the page title), and a
  logical heading order: an h2 per section and an h3 per item, with no skipped levels.
- Visible focus everywhere: `outline: 2px solid var(--focus); outline-offset: 3px`
  (`:focus-visible` in base). Never remove it.
- A skip link (`.skip-link` → `#main`) as the first focusable element.
- All text ≥ 4.5:1 in both themes, enforced by `make contrast`.
- Everything works by keyboard, including the theme toggle's roving tabindex.
- Respect `prefers-reduced-motion` (§6) and `prefers-color-scheme` in system mode.
- Utility: `.sr-only` for screen-reader-only labels (utilities layer, no
  `!important`). `.tap-target` extends a small single-line link's hit area to
  `--tap` height with an invisible `::before` (only for links in a row).
- Icon-only links and buttons (`.icon-btn`) must have an accessible name: the
  `icon` label form or `aria-label`.
- `:focus:not(:focus-visible)` has no outline. `.card--interactive` shows the
  ring on the card (via `:has(.card__link:focus-visible)`), not on the title link.
- Images always have `width`/`height` attributes (via `imgSize`). The hero
  photo is eager with `fetchpriority="high"`; every other image uses `loading="lazy"`
  and `decoding="async"`.

---

## 12. Build pipeline

```
content/site.yaml ─▶ internal/site    Load (strict) ─▶ Validate (all problems, numbered)
static/ ───────────▶ internal/assets  wipe dist/ ─▶ copy ─▶ concat+fingerprint CSS ─▶ fingerprint main.js ─▶ manifest
templates/ ────────▶ internal/render  layout + partials + pages/*.html ─▶ dist/<page>.html
                     internal/build   Run(opts) orchestrates the three, then writes dist/CNAME + dist/.nojekyll;
                                      cmd/build and cmd/serve call it
```

### Packages
| Package | Responsibility |
|---|---|
| `internal/site` | `types.go`: structs mirroring `site.yaml`, plus helpers (`Person.FullName`, `Contact.MailtoHref`, `Action.IsExternal` / `AssetPath`, `Site.Facts` (the derived "Currently"), `Job.ID`, `SkillGroup.PanelID`, `Slug`, `Education.Degree` / `Schools`, `School.DateText`, `Language.Meter` / `MeterLabel` / `MeterBars`, `Project.HasLinks`, `EmploymentTypes`, `Placeholders`, `LanguageLevels`). `load.go`: `Load`/`Parse` with `KnownFields(true)`. `validate.go`: `Validate(site, staticDir, iconDir) (warnings, error)`, which collects every problem into a numbered `Problems` error. `SectionIDs` lists the page anchors. |
| `internal/assets` | `Build(static, out, protect…)`: wipes `out` (refusing if it contains a source dir), copies `static/`, fingerprints, and returns a `Manifest`. `ImageSize` reads JPEG/PNG dimensions with `image.DecodeConfig`. |
| `internal/render` | `render.go`: parses `layout.html` + `partials/*.html`, clones them per page, executes `"layout"`. `funcs.go`: the FuncMap. |
| `internal/build` | `Run(Options)`, `Defaults()`, and a `Result` summary that includes warnings. Also writes the GitHub Pages files: `CNAME` (the host of `site.url`, never hard-coded) and an empty `.nojekyll`. |
| `internal/career` | Pure: `Build(entries, now, labels) Graph`, the career-graph view model (§16). |
| `internal/evidence` | Pure: `Build(site, labels) []Group`, each skill's evidence (§17). |
| `internal/textfmt` | `Bold(s)`: the `**bold**` formatter (escape first, then `<strong>`). |

### What validation checks
- Required fields: the name parts, role, summary, photo alt, email, site
  url/title/description/lang, and every section heading.
- The email format, and that `kmuchina22` appears nowhere.
- `site.url` and every link URL are absolute `https://` URLs.
- `person.photo` exists under `static/`; `person.resume_pdf` too, when set.
  **`resume_pdf` is optional:** empty hides the hero "Download résumé"
  button (and the styleguide's ghost-button demo).
- Project slugs match `^[a-z0-9-]+$` and are unique.
- Every YAML `icon:` has a `templates/icons/<name>.svg`.
- `ui.nav[].href` is one of the section anchors.
- `experience.action` / `projects.action`, when present, have a label and a
  `url` that is an `https://` URL or an existing file under `static/`.
- `contact.timezone.iana` loads with
  `time.LoadLocation`.
- Skills: every group has a name and icon, every item a name and an
  **existing icon file**; panel ids (`used-<group>-<skill>`) are unique.
- Projects: `placeholder` is in the enum; a `metric` has value and label;
  links have labels.
- Education: degree and school set; `date` is `YYYY-MM`; `start`/`end` are
  both `YYYY` and in order; every item has `honours` or `grades` (else it
  wouldn't render); at most one has honours. `languages[].level` is one of
  the LinkedIn levels. Job ids are unique. Every `ui.*` label is set.
- Experience: required fields, `type` in the enum, no "intern" in a title,
  `YYYY-MM` dates (or `end: present`), `end` ≥ `start`, at least one bullet,
  the icon file exists. Education `timeline.date` is `YYYY-MM`. The
  `ui.career` labels are set.
- A missing project screenshot is only a **warning**; the placeholder UI
  renders instead.

### Fingerprinting & manifest
- `static/css/*.css` are concatenated in filename order and written as
  `dist/css/site.<sha256[:8]>.css`. `static/js/main.js` becomes
  `dist/js/main.<sha256[:8]>.js`. `static/js/theme-init.js` is **not**
  copied; the layout inlines it.
- Every other static file is copied unchanged and also recorded in the manifest
  (`"img/profile.jpg" → "/img/profile.jpg"`). Templates reference **every**
  asset through `{{asset "…"}}`, so a missing file fails the build.
- Dotfiles are skipped. `design/` is never read or copied.

### Template FuncMap (`internal/render/funcs.go`)
| Func | Purpose |
|---|---|
| `asset "css/site.css"` | The fingerprinted URL path from the manifest; an unknown key is a build error. |
| `icon name [label]` | Inline SVG (§9). |
| `imgSize "img/profile.jpg"` | `{W, H}` for `width`/`height` attributes. |
| `exists "img/projects/x.png"` | Whether the file exists under `static/`. |
| `year` | The current year (footer). |
| `join list sep` | `strings.Join`. |
| `dict "k" v …` | a `map[string]any`, to pass several values to a partial (the theme toggle). |
| `bold s` | `template.HTML` of `textfmt.Bold(s)`: escaped text with `**x**` → `<strong>x</strong>`. |

There's no `safeURL` and no URL that needs one.

### Templates
```
templates/
├── layout.html            {{define "layout"}}: the <html> shell
├── pages/                 one output file per page: dist/<name>.html
│   ├── index.html         defines "main": the sections in page order
│   ├── 404.html           noindex; GitHub Pages serves it for unknown paths
│   └── styleguide.html    noindex; dev-only design-system verification page
├── partials/              head, skip-link, nav, theme-toggle, social-links,
│                          section-action, hero, about, experience, skills,
│                          projects, education, contact, footer
└── icons/                 <name>.svg + LICENSES/
```
- A page can define the blocks `"main"` (required), `"title"`, `"head-extra"`
  (page-only `<style>`), and `"noindex"`. Defining `{{define "noindex"}}{{end}}`
  makes the page `noindex` and omits its canonical link.
- Template data is `render.Page`: `.Site` (`*site.Site`), `.Name`, `.Path`,
  `.Home` (`""` on the index, `"/"` elsewhere; prefix in-page anchors with it,
  as in `{{$.Home}}#about`), `.Canonical`, `.NoIndex`, `.ThemeInit` and
  `.Career` (the `*career.Graph`, built once per build for the current month),
  `.Skills` (`[]evidence.Group`) and `.LocalTime` (`{Text, ISO}`: the build's
  time in `contact.timezone`, the contact clock's no-JS value).
- **`<head>` order (`partials/head.html`):** charset → viewport → inlined
  `theme-init.js` → `<title>` → description → canonical or robots →
  `theme-color` → font preloads (Fraunces normal + Inter only) →
  fingerprinted CSS → fingerprinted `main.js` (defer) → favicon →
  `{{/* Phase 7: SEO/OG/JSON-LD */}}` → `head-extra`.
- `<html lang="{{.Site.Meta.Lang}}" data-theme="dark">`, so no-JS visitors get dark.
- `static/favicon.svg` is a placeholder "KM" monogram. Its colours mirror
  `--bg` and `--accent`, because SVG favicons can't read CSS variables. It and
  the `theme-color` meta are the only literal colours outside `00-tokens.css`.

### Page structure & anchors
`skip-link → div.scroll-sentinel → header.site-header (brand, nav, theme toggle, menu button) → dialog.menu → main#main → footer.site-footer`

| Section | id | Heading id | Background |
|---|---|---|---|
| Hero | `top` | `top-title` (the `<h1>`) | `--bg` |
| About | `about` | `about-title` | `--bg` |
| Experience | `experience` | `experience-title` | `.section--alt` |
| Skills | `skills` | `skills-title` | `--bg` |
| Projects | `projects` | `projects-title` | `.section--alt` |
| Education (includes Languages) | `education` | `education-title` | `--bg` |
| Contact | `contact` | `contact-title` | `.section--alt` |

Every section is a `<section id aria-labelledby="<id>-title">`. Each project
is `<article id="project-<slug>">`; each job `<article class="job"
id="job-<slug>">`; each evidence panel `<div class="used"
id="used-<group>-<skill>">`. The footer's Back to top links to `#top` on the
index and `#main` elsewhere.

**Nav order decision:** the mockup's *nav* lists Projects before Skills, but
its *page* shows Skills before Projects. The page order wins, and the nav
(`ui.nav`) matches it: About, Experience, Skills, Projects, Education, Contact
(no "Home" since Phase 3: the brand does that job).

**Reveal hooks:** every `.section__aside` has `data-reveal`, and so does the
about intro. Lists and grids use `data-reveal-group` (`.focus`, `.facts`,
`.skill-table`, `.projects__grid`, `.edu`, `.contact-body`). The career graph has **no** reveal hooks: its
motion is the graph itself (§6 items 10–14). The hero has **no** reveal hooks: its
entrance is CSS-only (`data-hero`, §6 item 1).

### CSS architecture: cascade layers (Phase 2)
`00-tokens.css` opens with the layer order, so **the layer decides which rule
wins before specificity or file order does**:

```css
@layer reset, tokens, base, layout, components, sections, utilities, motion;
```

Numbered files in `static/css/` are concatenated in filename order:

| File | Layer | Contents |
|---|---|---|
| `00-tokens.css` | `tokens` (+ the `@layer` declaration) | all tokens and theme colours |
| `01-fonts.css` | none | `@font-face` (not part of the cascade) |
| `02-base.css` | `reset`, then `base` | reset; bare elements, the theme-fade default, focus, selection |
| `04-layout.css` | `layout` | the primitives (§7), `.section`, `.section--alt` |
| `05-components.css` | `components` | the component catalogue (§14) and type roles |
| `06-sections.css` | `sections` | per-section composition: header, menu, hero (§15), the career graph (§16), about … footer (§17) |
| `08-utilities.css` | `utilities` | `.sr-only`, `.muted`, `.tnum`, `.tap-target` |
| `09-motion.css` | `motion` | the motion vocabulary (§6), reveal, reduced motion |

**Which layer does a rule belong in?**
- Styles a bare element with no class → `base` (`reset` if it only normalises the browser).
- Arranges children and is content-agnostic → `layout`.
- A reusable, named piece used in more than one place → `components`.
- Only makes sense inside one section (`.hero__…`, `.career`, `.projects__grid`) → `sections`.
- One property, meant to override anything → `utilities`.
- Moves something, or cancels movement for reduced motion → `motion`.

Rules: **every rule lives in a layer** except `@font-face`. Page-only CSS (the
styleguide's `head-extra`) is wrapped in `@layer sections`, because unlayered
CSS would beat every layer, including reduced motion. **No `!important`**
anywhere (none is needed, even in the reduced-motion block).

### Commands

| | make | plain Go |
|---|---|---|
| Build | `make build` | `go run ./cmd/build` |
| Serve (:8080) | `make serve` | `go run ./cmd/serve` |
| Test | `make test` | `go test ./...` |
| Contrast | `make contrast` | `go run ./tools/contrast` |
| Vet | `make vet` | `go vet ./...` |
| Clean | `make clean` | `rm -rf dist` |
| CI, locally | `make ci` | vet → test → build → output check (`check-dist`) |

Flags for both `cmd/build` and `cmd/serve`: `-content`, `-templates`,
`-static`, `-out`. `cmd/serve` also takes `-addr` (default `:8080`). It
rebuilds on any request for `/`, a path ending in `/`, or `*.html` (guarded by
an RWMutex), serves with `Cache-Control: no-store`, and turns a failed build
into a plain-text 500.

**Tests:** `internal/site` loads and validates the real content, checks the
language meter mapping (all five levels, an unknown level fails), the
derived "Currently" fact (and its absence without a current role), the job
ids and their uniqueness, `Slug`, the degree/school split, the screenshot
warning, a table of Phase 5 rules (missing/empty skill icon, duplicate
skill, bad placeholder, bad time zone, non-https source URL, education
dates and years, an item with neither honours nor grades, a half metric),
checks that an unknown field or the retired email fails,
that problems are collected, that a bad section `action` fails (missing file,
missing label, non-https URL), and table-tests the experience rules (intern in
a title, the type enum, `YYYY-MM` dates, end before start, a missing icon) and
the milestone date. `internal/career` table-tests the real graph (§16) plus
synthetic cases (no volunteer roles, only volunteer roles, a single entry, a
fork, overlapping internships, a lane-1 role ending after lane 0's newest
started, an ongoing volunteer role), durations (same month, 1 month apart,
12 and 14 months, Jul–Jul, two years, present), tenure widths, the legend and
the summary. `internal/textfmt` tests `**bold**`, `<script>` escaping and an
unclosed `**`. `internal/evidence` table-tests the real evidence (the 8
skills of §17 and that every other skill has none), the hrefs and kinds,
normalisation ("Node.js" = "nodejs", "Git & GitHub" = "git github", "C++"
≠ "C"), a duplicate match counted once, the three panel wordings and the
empty case. `internal/build`
runs a full build into a temp dir and asserts: one `<h1`, every section id, the
email, no retired email, every header and menu nav `href` matching a
rendered `id` (12 links, no "Home"), two theme toggles, two section
`.link-arrow` actions, the 5 career rows in order with their kinds, the
summary text, one merge curve, `**bold**` rendered as `<strong>`, no "intern"
in a job title, the `@layer` order declared before the first layer block,
fingerprinted assets that exist, theme-init inlined and not copied, no
`design/`, the 404/styleguide/favicon files present, `CNAME` equal to
`kmuchina.com`, and an empty `.nojekyll`. Phase 5 adds: no phone number in
any output file, the "About me" heading, the intro first in the about body
with its `<strong>`, "Currently" as the first fact, 29 skill pills each
with an icon, 8 evidence buttons with panels (open without JS) whose links
all resolve, 2 `aria-hidden` placeholder stages, the interactive journal
card, the degree/grades/meter markup, the channels in order, the
build-time Nairobi clock, the copy button and live region, the 4 job ids,
the source link and Back to top. A second test builds with a temporary
screenshot and checks it replaces the scanner placeholder (size, alt, no
`aria-hidden`).

Go is installed for the user at `~/.local/go`. If `go` isn't on `PATH`, run
`export PATH="$HOME/.local/go/bin:$PATH"`.

### Deployment (GitHub Pages, set up 2026-10-30)

- **`main` is production.** `.github/workflows/deploy.yml` runs on every push
  to `main` (and by hand via `workflow_dispatch`): checkout → `setup-go`
  (version from `go.mod`) → `go vet` + `go test` → `go run ./cmd/build -out
  dist` → output check (`index.html`, `404.html`, `.nojekyll`, `CNAME` is
  exactly `kmuchina.com`) → `configure-pages` → `upload-pages-artifact` →
  `deploy-pages`. Any failing step means nothing is deployed and the live
  site is unchanged. `concurrency: pages` cancels a superseded run.
- `make ci` runs the same steps locally, in the same order.
- `CNAME` and `.nojekyll` come from the Go build (`internal/build`), not
  the workflow, so `make build` output is always Pages-ready. URLs are
  root-relative (`/css/…`, `/#about` on the 404) or absolute, which is
  correct because the site is served from the domain root.
- **Never commit `dist/`.** It is in `.gitignore`; `git ls-files dist` must
  be empty.
- **Repo:** `kmuchina/portfolio` (`origin` = `git@github.com:kmuchina/portfolio.git`),
  public (free-plan Pages needs it). Pages source: **GitHub Actions**;
  custom domain `kmuchina.com`, verified at the account level (Settings →
  Pages → TXT record), **Enforce HTTPS** on once DNS is green.
- **DNS (Spaceship):** `@` A → `185.199.108.153`, `.109.153`, `.110.153`,
  `.111.153`; `@` AAAA → `2606:50c0:8000::153`, `8001`, `8002`, `8003`;
  `www` CNAME → `kmuchina.github.io` (Pages redirects www to the bare
  domain), plus the GitHub verification TXT. Email MX/TXT records stay.

---

## 13. Class names & BEM convention

The convention is **Block__element--modifier**, all lowercase and hyphenated.
A block is a standalone component; elements only make sense inside their
block; modifiers are variants. Global type roles and utilities (`.eyebrow`,
`.meta`, `.lead`, `.display-name`, `.quote`, `.prose`, `.list`, `.divider`,
`.sr-only`, `.muted`, `.tnum`, `.tap-target`, `.icon`), the layout primitives
(`.container`, `.stack`, `.cluster`, `.grid-auto`, `.split`) and the motion
hooks (`.lift`, `.underline-grow`, `.press`) can be mixed into any element. Phases 2–5 style these classes; change the markup only
when a class is genuinely missing, and update this list when you do.

| Block | Elements / modifiers | Where |
|---|---|---|
| `.skip-link` | | layout |
| `.scroll-sentinel` | | layout (1px marker for the scrolled header) |
| `.site-header` | `--scrolled`, `__inner`, `__menu-btn` | nav.html |
| `.brand` | (`<em>` = "Muchina") | nav.html |
| `.nav` | `__list`, `__item`, `__link` (`[aria-current="true"]` = active) | nav.html |
| `.menu` (`<dialog>`) | `__inner`, `__bar`, `__close`, `__list`, `__item`, `__link`, `__footer` | nav.html |
| `.theme-toggle` | `__option` (`[aria-checked]` is the state) | theme-toggle.html |
| `.section` | `--alt`, `__grid`, `__aside`, `__body`, `__title`, `__action` | every section (the spine) |
| `.hero` | `__inner`, `__content`, `__eyebrow`, `__name`, `__role`, `__summary`, `__actions`, `__social`, `__aside` (photo column), `__media`, `__photo`, `__quote` (`<figure>`) | hero.html |
| `.btn` | `--primary`, `--ghost`, `__label` | hero, 404 |
| `.icon-btn` | | social rows |
| `.link` | | menu email |
| `.link-arrow` | `--external` (`arrow-up-right`, diagonal nudge) | section actions |
| `.eyebrow` | `--accent` (hero only) | every section header |
| `.card` | `--interactive`, `__link` | projects, the degree |
| `.tile` | (`--tile-size`, `--tile-radius`; reads `--soft` / `--c`) | about focus, skill groups, degree, job org tiles |
| `.social` | `__item`, `__item--location`, `__link`, `__label` | hero, footer, menu |
| `.intro` | (`<strong>` = the ink phrase) | about.html |
| `.focus` | `__item`, `__title`, `__text` | about.html |
| `.facts` (`<dl>`) | `__item`, `__label` (`<dt>`), `__value` (`<dd>`), `__detail` | about.html |
| `.experience` | | experience.html |
| `.summary` | `__num`, `__text` | experience aside |
| `.legend` | `__label`, `__item` (`[aria-pressed]`), `__swatch` (`--all`, `--<kind>`), `__count` | experience aside |
| `.career` (`<ol>`, `[data-filter]`, `[data-lane]`) | `__row` (`--internship`, `--volunteer`, `--contract`, `--freelance`, `--full-time`, `--part-time`, `--milestone`; `.is-quiet`, `.is-lit`), `__graph` | experience.html |
| `.seg` | `--l0`, `--l1`, `--above`, `--below`, `--full`, `--dashed`, `--fade` | career graph cell |
| `.node` | `--l0`, `--l1`, `--diamond`, `--ring`, `--current` (`.is-paused`) | career graph cell, legend swatch |
| `.curve` (`<svg>`) | `--fork`; paths `.rest`, `.ink` | career graph cell |
| `.job` (`<article>`) | `__tile` (the org tile), `__head`, `__title`, `__org`, `__meta`, `__type`, `__dates`, `__tenure`, `__place`, `__body`, `__bullets`, `__tags` | experience.html |
| `.tenure-bar` | (inner `span`, width `--w`) | job meta |
| `.tag` | `--accent` ("Current" only) | experience, skills, projects |
| `.tag-list` | (children are `li.tag`) | experience, skills, projects |
| `.skill-table` (was the prototype's `.stack`, which clashes with the primitive) | `__row` (`[data-skill-row]`), `__head`, `__name`, `__note` | skills.html |
| `.skill-list` / `.skill` | `--plain`, `__count`; `button.skill[aria-expanded]` | skills.html |
| `.used` | `__label`, `__list`, `__link`, `__kind` | skills.html (evidence panel) |
| `.projects` | `__grid` | projects.html |
| `.project` | `--staged`, `__text`, `__kind`, `__title`, `__summary`, `__foot`, `__metric`, `__link`, `__note`, `__stage` | projects.html |
| `.phone` | `__screen`, `__notch`, `__image` | project stage |
| `.ui` (placeholder screen) | `__bar`, `__title`, `__dot`, `__chips`, `__chip` (`--on`), `__card`, `__line` (`--accent`; width `--w`), `__fab`, `__finder`, `__scan`, `__sheet`, `__pill` | project stage |
| `.edu` | | education.html (the grid) |
| `.degree` | `__head`, `__title`, `__school` | education.html |
| `.honours` | (`<small>` = detail) | degree card |
| `.award` | | degree card |
| `.school` | `__title`, `__meta` | education.html |
| `.grades` (`<dl>`) | | school column |
| `.langs` | `__label` | education.html |
| `.lang` | `__name`, `__level` | education.html |
| `.meter` | (`i`, `i.on`) | languages |
| `.contact-body` | | contact.html |
| `.mail` | `__address` | contact.html |
| `.copy` | `__icon` (`--copy`, `--done`), `__label`; `.is-done` | contact.html |
| `.channels` | `__link`, `__icon`, `__name`, `__handle`, `__go` | contact.html |
| `.here` | `__item` | contact.html |
| `.site-footer` | `__inner`, `__name`, `__built`, `__end`, `__social` | footer.html |
| `.to-top` | | footer.html |
| `.not-found` | `__title`, `__body`, `__action` | 404.html |
| `.icon` | | every `{{icon}}` |

The styleguide's own classes are prefixed `sg-` (plus `.swatch`, `.demo-*`
and `.font-checks`), and its styles live in that page's `head-extra`.

---

## 14. Components (Phase 2, `05-components.css`)

Tokens only, mobile-first. Components declare `transition: var(--transition-ui)`;
their movement lives in `09-motion.css`. All of them are shown, with real
content from `site.yaml`, on `/styleguide.html` (Components section).

| Component | Purpose | Modifiers / parts | Use it for |
|---|---|---|---|
| `.btn` | 44px button or button-link, Inter 500 `--fs-sm`, `--radius-sm`, `--icon-sm` icons | `--primary` (`--accent-strong` → `-hover`), `--ghost` (`--line` border → `--line-strong` + `--surface`) | Hero CTAs, 404. Always add `.press`. One primary per view. |
| `.icon-btn` | 44px round icon-only target, `--muted` → `--ink` + `--surface` | — | Social rows. **Needs an accessible name.** |
| `.link` | standalone single-line `--accent-text` link; underline grows | — | The menu's email. In a row, add `.tap-target`. |
| `.link-arrow` | `--accent-text` Inter 500 `--fs-sm` link (line-height 1) + trailing `arrow-right`, 44px tall | `.section__action` (16px box + `.tap-target`); `--external` (`arrow-up-right`) | Section actions ("View full résumé →", "More on GitHub ↗") in the spine aside; the `section-action` partial picks the variant. |
| `.prose` | 65ch running text, `--space-4` between children; links keep a native underline that fades in | — | Any multi-paragraph copy with inline links. |
| `.tag-list` / `.tag` | reset `<ul>` of mono `--fs-xs` pills with a `--line-strong` border (Phase 4, as in the prototype) | `.tag--accent` (`--accent-soft` bg, `--accent-text`) | Skills, project stacks. `--accent` **only** for "Current". |
| `.card` | `--surface`, 1px `--line`, `--radius`, fluid padding, no shadow | `--interactive` (+ `.card__link`: whole card clickable, lift on hover) | **Projects and the degree only.** |
| `.tile` | 44px icon tile, radius 10px, `--soft`/`--c` or the accent pair, 22px icon | `--tile-size`, `--tile-radius`, `--icon-size` | About focus areas, skill groups (36px), the degree, the job org tiles. |
| `.skill` | pill, Inter 500 `--fs-sm`, `--line-strong` border, 16px `--muted` icon | `--plain` (400, `--muted`); `button.skill` + `__count`: hover/open = `--accent-text` | Skill pills; evidence toggles. |
| `.used` | `--surface` panel, 1px `--line`, `--radius`, links in `--accent-text` | — | A skill's "used in" panel. |
| `.copy` / `.to-top` | 44px pills, `--line-strong` border, `--icon-sm` | `.copy.is-done` (check icon, `--accent-text`) | Copy email (hidden without JS), Back to top. |
| `.phone` / `.ui` | 11.5rem device frame (`--ink` bezel, `--screen`), peeking out of its stage; `.ui*` draws an abstract placeholder app | — | Project cards. |
| `.legend` | a group of toggle buttons (`aria-pressed`): swatch, label, count; 40px tall | swatch shape/colour set per kind in `06-sections.css` | The career-graph legend. |
| `.eyebrow` | mono `--fs-xs` 500 uppercase 0.14em `--muted` | `--accent` | Above section titles; `--accent` for the hero "Hello, I'm" only. |
| `.lead` | `--fs-lg`, `--muted` | — | A section's opening paragraph. |
| `.meta` | mono `--fs-meta`, tabular nums, `--muted` | — | Dates, kinds, small metadata. |
| `.quote` | Fraunces italic `--fs-xl`, `--lh-snug`, `--ink`, curly quotes in `--accent` | — | The hero pull-quote only. |
| `.list` | bullets as a 5px `--accent` dot on the first line's x-height; `--space-2` apart; 65ch | — | Bulleted lists (styleguide; none on the page today). |
| `.divider` | `<hr>`, 1px `--line`, no margin | — | Separating items (parent controls spacing). |
| `.theme-toggle` | 40px segmented pill (3px padding) of three 32px round options; checked = `--surface` + `--ink` + 1px `--line-strong` ring; each option's `::before` makes a 36 × 40px hit area | — | Header (≥ 1024px) and mobile menu; kept in sync. Hidden without JS. |
| `.display-name` | the hero `<h1>` (§3) | `<em>` = `--accent-text` italic | Hero only. |
| `.skip-link` | 44px `--accent-strong` pill, shown on focus | — | Layout, first focusable element. |

Focus: one ring everywhere (`2px solid var(--focus)`, offset 3px). Outlines
follow `border-radius`, so buttons, icon buttons and toggle options get a ring
in their own shape.

## 15. Header, menu & hero (Phase 3, `06-sections.css`)

### Header — `.site-header`
- **Sticky** (`top: 0`, `z-index: 10`), `--header-h` tall. Contents: `.brand`
  (Fraunces 600 `--fs-lg`, links to `#top`), `.nav` (Inter 500 `--fs-sm`,
  `--muted` → `--ink`, 40px tall, `--space-3` side padding, underline-grow),
  the theme toggle, and the menu button (`.icon-btn`, `menu` icon).
- **Scrolled state:** with JS, the header is transparent with no border at the
  very top. `main.js` watches `div.scroll-sentinel` (1px, absolute at the top
  of the page) with an `IntersectionObserver` (**no scroll listener**) and sets
  `.site-header--scrolled` once it leaves the viewport: background
  `color-mix(in srgb, var(--bg) 82%, transparent)`, `backdrop-filter:
  blur(12px) saturate(140%)`, `1px --line` border, colours cross-fading at
  `--dur-fast`. Where `backdrop-filter` is unsupported (`@supports`), it stays
  solid `--bg`. **Without JS it is always solid.**
- **< 1024px with JS:** nav and toggle hidden; the menu button opens the menu.
  **< 1024px without JS:** no menu button; the nav is a sideways-scrolling row
  under the brand, and the header is not sticky (two rows are taller than
  `--header-h`).

### Mobile menu — `<dialog class="menu" id="site-menu">`
- Opened with `showModal()`: native focus trap, `Esc`, inert page. Menu button
  has `aria-controls="site-menu"`, `aria-expanded`, label `ui.menu.open`; the
  dialog has its own close button (`x`, `ui.menu.close`) placed directly under
  the menu button. Focus moves to it on open and returns to the menu button
  on close.
- Panel: full width, `inset-block-start: var(--header-h)`, `--bg`; fills the
  viewport height below 640px. Links: Fraunces 560 `--fs-xl`, hairlines
  between them (borders on the `<li>`s, since an `<hr>` can't sit in a `<ul>`);
  the section under the header when it opens gets `aria-current="true"`
  (`--accent-text`). Below: the second theme toggle, the email (`.link`), the
  social `.icon-btn` row.
- Closes on: a link click (then the link navigates), `Esc`, a backdrop click
  (a click whose target is the dialog and lies outside its box), or a resize
  to ≥ 1024px. Scroll lock: `html:has(dialog[open]) { overflow: hidden }`.
  Backdrop: `color-mix(in srgb, var(--bg) 60%, transparent)`.

### Hero — `#top.hero`
- Order: eyebrow (`.eyebrow--accent`) → `h1.hero__name` (`.display-name`) →
  `.hero__role` (Inter 400 `--fs-xl` `--muted`, `--space-3` above) →
  `.hero__summary` (`.lead`, full YAML text, `52ch`, `--space-5` above) → CTAs
  (`.cluster`, `--space-8` above; full-width below 400px) → `.hero__social`
  (`--space-6` above: icon buttons, then the
  location, after a little extra gap). Photo and quote in `.hero__aside`.
- Layout: one column below 900px (text, then the photo at `max-width: 22rem`,
  centred; top padding is half `--section-pad`). From 900px a `.split` with
  `--split-ratio: 1.15fr 0.85fr`, both columns centred, hero `min-height:
  min(calc(100svh - var(--header-h)), 54rem)`; photo column up to 26rem,
  aligned to the end. From 1200px the column gap is `--space-20`. The hero
  clips overflow, so nothing scrolls sideways down to 320px.
- **Photo treatment: (B) framed portrait.** `profile.jpg` is a pale-grey
  studio shot (≈ #B6 at the top corners to #E0 lower down), half body, face
  centred in the upper third (eyes at ~27% of the height). That grey can't
  dissolve into the navy dark `--bg`, so a cut-out fade (A) was rejected. The
  photo is already exactly 4:5, so `aspect-ratio: 4/5; object-fit: cover;
  object-position: 50% 30%` crops nothing. `--radius`, `1px --line` border, an
  `--accent-soft` rounded rectangle offset `--space-4` down-right
  (`.hero__media::before`), and the ≤ 8% `--accent` glow
  (`.hero__aside::before`). Eager, `fetchpriority="high"`, never lazy.
- **Quote placement rule:** `figure.hero__quote` sits **under the photo,
  aligned to the end, at every width** (max 24ch, balanced, two lines).
  It never overlaps the photo: light `--ink` on the pale-grey backdrop fails
  contrast in dark mode, and the photo's top corners are beside the face.
  Readable by screen readers (no `aria-hidden`).

## 16. Experience: the career graph (Phase 4)

Reference: `design/experience-prototype.html`. Markup in
`templates/partials/experience.html`, CSS in `06-sections.css` (graph,
entry) and `09-motion.css` (items 10–12), JS in `main.js` (legend, live
duration, fallback observer, pulse pause; ~2.6 KB).

### View model (`internal/career`)
`Build(entries, now, labels) Graph`. "now" is passed in (tests are
deterministic; the site uses the build's month).
- **Rows:** one per experience item plus one per education `timeline`
  milestone (start = end = its date), **sorted by start, newest first** (ties:
  later end first). Each `Row` carries the entry, its `Kind` (the CSS
  modifier), `Lane`, `Above []Seg` (rendered before the curve and node:
  `above`, then `full` segments), `Curve *Curve`, `Node`, `Below []Seg`, and
  display values: `TypeLabel`, `StartISO`/`StartText`, `EndISO`/`EndText`,
  `Duration`, `TenurePct`.
- **Lanes by category:** lane 0 = work (Internship, Contract, Freelance,
  Full-time, Part-time) and milestones; lane 1 = Volunteer. Overlapping
  entries of the same lane **share** it (no extra lanes).
- **Segments between consecutive nodes of a lane:** the newer row gets
  `seg--below`, the older `seg--above`, rows between `seg--full`. **Solid** if
  the older one ran until at least the month before the newer one started
  (`older.end + 1 ≥ newer.start`: continuous or overlapping); otherwise
  **dashed** (a time gap).
- **Lane 0's newest node, if current:** a `seg--above seg--fade` stub (the
  career is still going).
- **Lane 1 above its newest node:** with `E` = the lane's latest end, the lane
  runs up through every row that started before `E` (`seg--above` on its node
  row, `seg--full` between). At the topmost such row: if that row's node is in
  lane 0, a **merge** curve (lane 0 at the top of the row → lane 1 at the node)
  plus `seg--below`; otherwise it just ends. An ongoing volunteer role runs to
  the top row and fades.
- **Lane 1 below its oldest node:** if a lane-0 node below was active when the
  lane began, a **fork** curve on that row (`seg--below` / `seg--full` down to
  it); otherwise nothing.
- **Nodes:** `node node--l<lane>` plus `node--current` (end: present),
  `node--ring` (milestone) or `node--diamond` (volunteer); plain = work.
- **Duration** is inclusive of both months ("3 mos", "9 mos", "1 yr 1 mo",
  "2 yrs"); the current role's is wrapped in `data-start` and `main.js`
  recomputes it on load (the Go value is the no-JS fallback). **Tenure bar**
  width = `months / longest role's months`, rounded (`style="--w:NN%"`).
- **Legend:** "All" (all rows) then each kind present, in the enum order,
  then milestones, each with its count and the lane it lights.
- **Summary:** roles (excluding milestones), "roles since {earliest year}", plus
  ", up to {N} at once" when the most roles held in one month is > 1 (words
  for small numbers).

### Your data, as built
```
         lane 0   lane 1
           ┆               ← fades out upward (still going)
 KFC       ◉                ICT & Digital Innovation      Internship  Aug 2026 – Present
           ┆   dashed       (Dec 2025 → Aug 2026 gap)
 Grad      ○                Graduated, BSc CS             Milestone   Dec 2025
           ┆   dashed       (Jan → Dec 2025 gap)
           │╲               Tech4Good ended Dec 2024 while KRB ran: merge
 KRB       ●  │             ICT Officer                   Internship  May 2024 – Jan 2025
              │  solid
 T4G          ◆             Project Lead & Tech Lead      Volunteer   Feb – Dec 2024
              │  solid      (overlaps Chair: shared lane)
 Chair        ◆             Chairperson                   Volunteer   Jul 2023 – Jul 2024
```
Summary: **4** "roles since 2023, up to three at once" (May–Jul 2024: KRB,
Tech4Good and Chair).

### Geometry and look (from the prototype)
Graph column `--gw` 52px (68px from 640px), lanes at `--x0` 16/20px and
`--x1` 37/46px, node centre `--node-y` 22px. The merge/fork curve is an SVG
path in a 68 × 22 viewBox stretched to `--gw`. Resting line `--line-strong`;
lit ink in the lane colour (lane 0 internship blue, lane 1 volunteer green).
Entry: a 44px org tile (`.job__tile`, the type's `--soft` tint, 22px icon),
Fraunces 600 `--fs-xl` title, Inter 600 org, a meta block (type in Fraunces
italic in the type colour, mono dates, tenure bar + duration, location with
a pin) that becomes a right-aligned 11rem column from 1200px, bullets with an
8px type-coloured dash and `<strong>` metrics in `--ink`, and skill tags. The
milestone entry is quieter (dashed 36px tile, italic title).

## 17. About, skills, projects, education, contact & footer (Phase 5)

Reference: `design/sections-prototype.html`. Markup in the partials, CSS in
`06-sections.css` (sections) and `05-components.css` (`.tile`, `.skill`,
`.used`, `.copy`, `.to-top`, `.phone`, `.ui*`), motion in `09-motion.css`
(§6 items 15–18), JS in `main.js` (skill toggles, copy, clock, year; ≈ 1.9 KB
without comments). Every section uses the spine (§7).

- **About.** Aside: only the `<h2>` "About me" (no eyebrow; the old tagline
  is gone). Body: `p.intro` first (Inter 400 `--fs-lg`, line-height 1.7,
  56ch, `--muted`; its one `**bold**` phrase is `<strong>` in `--ink` 500),
  then `ul.focus` (three items: `.tile`, title, detail; side by side with
  vertical hairlines from 760px), then `dl.facts` (2 × 2 from 760px, label
  column 7.5rem, hairline rows).
  **"Currently" is derived** (`Site.Facts`): the experience item whose `end`
  is `present` gives the value (title) and detail (org), rendered first, then
  `about.facts`. No current role → no "Currently".
- **Skills.** `.skill-table` rows: a 36px `.tile` + the group name (`<h3>`,
  Fraunces italic 500 `--fs-lg`) and optional mono `note`; the pills on the
  right from 760px (13rem head column). Every pill has its icon (§9).
  **Skill evidence** (`internal/evidence`): a skill matches a project if it
  equals an entry of `project.stack`, and a job if it equals an entry of
  `job.skills`, after normalising both (lower-case; drop everything but
  letters, digits, `+` and `#`). Nothing fuzzier. Each place counts once.
  Project evidence links to `#project-<slug>` (kind `ui.skills.project`);
  job evidence to `#job-<slug>`, labelled "{title}, {org before the comma}",
  kind = the employment type. **Panel label:** "{skill}, `used_in`" when
  every place is a project; "`used_at`" when every place is a paid role;
  "`shown_at`" otherwise (a mix, or any volunteer role: leadership shown,
  not a tool used). Today: Kotlin, Jetpack Compose → the journal; Python,
  TensorFlow Lite, CNN → Apple Care; Business Central, Cisco networking →
  ICT Officer, KRB (used at); Team leadership → Tech4Good and the IEEE
  branch (2, shown at). A skill with evidence is a `<button class="skill"
  aria-expanded aria-controls>` with a count badge, and its `.used` panel
  sits in the same row. **Without JS** every panel is visible inline
  (`aria-expanded="true"`, no `hidden`); `main.js` hides them and makes the
  buttons toggles, one open per row. The open-pill styling is `html.js`
  only.
- **Projects.** Two columns from 900px. A card is `.project.card.lift`
  (+ `.card--interactive` with a link; its first link is the `.card__link`
  stretched over the whole card, phone included). Text: kind (Fraunces
  italic, `--accent-text`), title, summary, tags, a foot (metric: big
  Fraunces value + label; links; or the mono `note`). Then the **stage**
  with the phone. **Screenshot rule:** if `projects.items[].screenshot`
  exists under `static/`, it renders as `img.phone__image` (`imgSize`,
  `alt="{ui.projects.screenshot_alt} {title}"`, lazy, async) and the stage
  is exposed; otherwise the `placeholder` UI (`journal` | `scanner`) draws
  in CSS and the stage is `aria-hidden="true"`; `none` (with no screenshot)
  renders no stage. A missing screenshot is a build warning. Screenshots:
  portrait PNG/JPEG, ideally 1080 × 2220.
- **Education.** `.edu`: 1.4fr / 1fr from 900px. The first item with
  `honours` is the **degree card** (`.card.degree`: tile, title, school +
  `Dec 2025`, "First Class Honours" in Fraunces italic `--fs-2xl`
  `--accent-text` with its detail, the award row). Items with `grades` form
  the quieter `.school` column (title, school + mono years, a 3-cell
  `dl.grades`; each column is at least its label's one-line width, so
  "Mean grade" never wraps and the values stay on one line). **Languages** (`.langs`, full width under a hairline): name,
  level and a 5-bar `.meter` (`role="img"`, `aria-label` "Level N of 5"
  from Go). **`languages[].level` is LinkedIn's scale:** Elementary (1),
  Limited working (2), Professional working (3), Full professional (4),
  Native (5); anything else fails the build.
- **Contact.** Aside: eyebrow, title, `section__lead` (`contact.body`).
  Body: the **email as the hero** (Fraunces 500, up to 3.5rem,
  `overflow-wrap: anywhere` so it fits at 320px) with the growing underline,
  and the **Copy email** button (clipboard, else it selects the address and
  says "Press Ctrl+C"; "Copied" + check for 2.2s; announced through a
  `role="status"` live region; labels from `ui.contact` via `data-*`;
  hidden without JS). Then `ul.channels` (`contact.links` in order: icon,
  name, `handle`, ↗; three columns from 760px, hairline rows, 64px tall).
  Then `p.here`: location, **local time** ("It's 15:26 here (EAT,
  UTC+3)"; Go renders the build's time in `contact.timezone.iana` as the
  no-JS value; `main.js` formats `Intl.DateTimeFormat` with the `data-tz`
  zone every 30s) and availability.
- **Footer.** Name (Fraunces), "© {year}. {ui.footer.built}" ("Built with
  Golang"; Kennedy, 2026-10-30: no "View the source" link; `main.js`
  refreshes the year), then the social icon buttons and the **Back to top**
  pill.
- **Deliberate differences from the prototype:** the real Lucide / Simple
  Icons files instead of its stand-in sprite (e.g. Lucide `brain`, not a
  chip; the Google Play logo, not a triangle); self-hosted fonts; the real
  header and 3-state toggle; all content from YAML; tokens instead of its
  hex values; motion in the `motion` layer without `!important`; the
  site's `.link-arrow` keeps its 0.01em letter-spacing (shared with the
  Phase 4 port); short paragraphs keep the site's 65ch `max-width` box
  (no visual change); the footer credit is "Built with Golang" with no
  source link; the grades columns grow to keep "Mean grade" on one line.

## 18. Working agreement for future sessions

1. Read this file first.
2. Propose a short plan before any large change.
3. Keep changes scoped to the current phase.
4. Before saying a phase is done, run `go vet ./...`, `go test ./...`, the
   build, and `make contrast`, and check the page in a browser in both themes,
   including with reduced motion.
5. Update this file whenever a decision changes, and tick off the roadmap.
6. Never add dependencies without asking Kennedy.
7. Commit locally only when asked; don't push unless asked.
8. **`main` is production. Every push deploys. Run `make ci` before pushing.**

---

## 19. Open TODOs (from Kennedy)

> **First push (2026-10-30), Kennedy's decision:** the history was squashed
> into one fresh root commit so the phone number from the Phase 0–1 commits
> never reached GitHub (the old local history is on branch
> `backup/pre-squash`; never push it). The old résumé PDF was moved to the
> gitignored `tmp/`, and the site shipped without a résumé until the new
> one replaced it (same day).

- [ ] Project screenshots: `static/img/projects/<slug>.png`
  (`prayer-gratitude-journal.png`, `apple-care.png`), portrait, ideally
  1080 × 2220. The build warns and shows the placeholder UI until they exist.
- [ ] The journal app's direct Play Store URL (its link points at the
  developer page for now: `projects.items[0].links`).
- [x] `static/img/picture_2.jpg` is gone from `static/` (checked 2026-10-30).
- [x] **New résumé added** (2026-10-30): no phone number; references
  "available upon request". Any future résumé must pass
  `pdftotext static/Kennedy_Muchina_Resume.pdf - | grep -i -e 703 -e referee`
  with no output before it is committed.
- [x] The phone number in the Phase 0–1 commits: resolved by squashing the
  history before the first push (see the note above).
- [x] Mockup, profile photo (`img/profile.jpg`, 3327×4160, 1.1 MB, optimise
  in Phase 7) and résumé PDF added.
- [x] The résumé PDF now shows muchina537@gmail.com (checked with `pdftotext` in Phase 1).
