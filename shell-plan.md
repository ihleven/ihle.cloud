# The app shell

## What this is

A plan to rebuild the shell simpler than it is. Begun 2026-10-09 by collecting
requirements rather than by refactoring the files that exist.

The mobile half is settled, and settled by measurement on the device rather than
by argument. The desktop half is open. This document is the record of both, so
the next session starts from the decisions and not from the screenshots.

## Where this has got to

Done, and seen in a browser at both widths:

1. `AppBar` + the shell's ceiling → **`Header.vue`**, one component with two
   renderings. The wide one is sticky. `--top-height` replaces the measured
   `shell-ceiling`, and `bar` leaves the page config.
2. `AppMenu` → **`Menu.vue`**, rendered by `Header`, its four layers teleported
   to the end of the body. One corner at both widths, so `pinned`, `above`,
   `standing`, `useMediaQuery` and the two-case `clearance` all go. `menu`
   leaves the page config.
3. **`AppShell` retired.** The player and the document's colour move to
   `app.vue`; `shell-floor` is the only measurement left in the shell.
4. **The area dropdowns are gone** from the wide row, and the menu panel lists
   the areas under their concerns — written out from `grouped`, which `useAreas`
   now returns, rather than from `sections`, whose shape is for a navigation
   component this panel deliberately does not use.
5. **The header takes a title and actions** at `#header-title` and
   `#header-actions`. This forced the two renderings into **one element** —
   `fixed lg:sticky`, `bg-(--shell-top) lg:bg-default` — because two of them
   would have meant two targets sharing one id. The mediathek detail page is the
   first to use it: its own bar is gone, and its player sticks at
   `top-(--top-height)` instead of hiding under the header. `/musik` gives up
   its title and its way home — the header offers none from an area's own root,
   which is the rule — and keeps the filter, the year and the view toggle, which
   are the page's own business. Its row had the same `sticky top-0` bug. A
   content entry hands its name over from `[...slugs]`, once, rather than from
   each of the four views.
6. **The header is no longer gated on a session; its contents are.** The strip
   is drawn for every page, the menu button and its hole only with a session.
   `public: true` is now **`unguarded: true`**, in `app.vue`, the type and the
   nine pages that carry it.
7. **One `--content-shell-offset`**, built from `--top-height` plus the band,
   where the band counts only at the width it is drawn. The phantom scroll is
   gone: at 390×844 the document is 844 and the overflow is 0, where it used to
   be 948 in a 900 window. At 1440 the offset measures 86px — a 54px header and
   a 32px band.

The colour contract is built with them: `top: { colour, ink }` in page meta,
published as `--shell-top` and `--shell-ink`, read by the header and by the rule
that paints `html`/`body`. Retro is blue.

All eight steps are done. The haze, which was what came next, is measured and
settled — see below. The restructure it was thought to need was built, failed
and was reverted.

## What the device settled

The status bar on an installed iPhone is not something the app decorates — it is
something the app paints. Four things are needed together, and each one alone
does nothing:

1. `apple-mobile-web-app-status-bar-style: black`, which is what lets the page
   reach under the bar at all
2. `viewport-fit=cover`, without which every `env(safe-area-inset-*)` reports
   `0` and nothing can be sized to the bar — **and the inset alone is 30px
   short of the haze on every device** — see *The haze, and what ends it*
   below.

   `--status-bar` is `env(safe-area-inset-top)` where the app is installed and
   `0px` in a browser tab, where the page is inside the window's chrome rather
   than under it. Set from `app.vue`, not from a `display-mode` media query:
   iPadOS does not reliably answer one, and `useStandalone` checks the query
   *and* `navigator.standalone`, which iOS has answered since long before it
   supported the query
3. a fixed opaque strip of `env(safe-area-inset-top) + 3rem`
4. `html, body` painted the same colour below `lg` — at scroll offset zero iOS
   blurs what lies behind the scroll *content*, not the fixed strip over it

All four are deployed and the bar is black on every page, scrolled or not.

The scroll edge effect itself is still a system material and cannot be
suppressed. What it can be is invisible, by giving it one colour to blur on both
sides of the seam.

## The haze, and what ends it

**Measured on 2026-10-10 and settled.** What ends it is not a shape. It is one
number, and a button that gives the blur nothing to work on.

### What the haze is, and how it was finally read

On a scrolling page an installed app gets a blurred band at the top: iOS's
scroll edge effect, part of Liquid Glass. It cannot be switched off — eight
experiments established that, and they are listed in
[[ios-scroll-edge-effect]].

**Its depth was assumed to be unreadable. It is not.** That assumption is what
cost every round before this one: that the blur's reach was its own
undocumented thing, different per device and per OS version, with no number a
page could find. What was missing was never a theory. It was a ruler.

`/debug/haze` is the ruler — a page with none of the shell on it
(`layout: false`), in a colour nothing else is, with sixteen numbered 10px
stripes fixed at the top of the window and `env(safe-area-inset-top)` printed
on the page as a number, because an installed app has no inspector and no
address bar to read it from. Both devices, installed:

| | OS | `inset-top` | last blurred stripe | difference |
|---|---|---|---|---|
| iPhone | 26 | 50 | 80 | **30** |
| iPad, no notch | 27.0.1 | 0 | 30 | **30** |

**haze depth = `env(safe-area-inset-top)` + 30px.** One rule, both devices, ±
the ruler's 10px step.

**And the haze is not anchored to the inset at all.** The iPad reports zero —
its page does not reach under the clock — and is hazed anyway. Every earlier
attempt sized the opaque band from the inset alone and was therefore short by
exactly 30px on both devices; on the iPad, where the inset is zero, those 30px
were the whole of the haze. That is why one build looked right on a phone and
wrong on a tablet, and why four rounds of raising a floor never converged.

**What the status-bar tag does, measured rather than read.** It moves the iPad
between modes and does not move the iPhone: under `default` the phone still
reported an inset of 50 — full-bleed — while the iPad reported 0 and had no
bleed at all. That is not the documented split, in which only
`black-translucent` is full-bleed. The app ships `black`. Installed, the value
is read once, when the icon is created, so testing a change means a build,
deleting the icon and adding it again.

### The frame, and why it was thought to end it

WebKit skips the blur — and draws a **hard edge** instead — when the element at
the top of the viewport sits in a `position: fixed` or `sticky` container that
**spans the viewport**: roughly ≥90% wide and up to ~105% tall, with an opaque
background. [kp-rueck#143][pr] is that fix in another app.

The operative words are *spans the viewport*. A 48px band never qualifies,
however deep it is made. This is also why the note in [[ios-scroll-edge-effect]]
recording that "an app with its own scroll container" had been tried and failed
is not a contradiction: a scroll container alone is not the condition. The
container has to be fixed, full-viewport and opaque.

### The shape

Installed, the app stops being a document that scrolls and becomes a fixed frame
with a scroller inside it. In a browser tab it keeps the document scroll — see
*Installed only* below.

```
html                     opaque, the app's own colour
└── app shell            position: fixed; inset: 0 — opaque
    ├── Header           an ordinary row at the top: not fixed, not sticky
    └── scroller         overflow-y: auto — everything else scrolls in here
```

### What that deletes

Most of what this document currently describes about the top of the screen
exists to work around the page scrolling under a fixed bar. With the frame,
those reasons go — in the installed shape. The tab shape keeps them, which is
why none of it can simply be deleted; it is a branch, not a replacement:

- **`--top-height` as a reservation.** The header is the first row of a flex
  column and the scroller takes the rest. Nothing is reserved, nothing is
  padded, nothing has to know a height.
- **Every `sticky top-(--top-height)` on a page becomes `sticky top-0`.** Inside
  the scroller, *the top* already is below the header. Four pages currently have
  to know the header's height; none of them would.
- **The stacking fight.** The header is a sibling above the scroller rather than
  an overlay competing with positioned page content for document order — so it
  need not be fixed, need not be last in the layout, and the Musik tiles cannot
  reach it.
- **The band under the status bar**, and with it the floor, the guard and the
  question of how deep either should be.
- **`scrolls: true` as a special case.** A header inside the scroller leaves
  with the page by default; a page that wants it to stay says `sticky` instead.
  The special case inverts and gets simpler.

### Installed only

**The frame is applied where the app is installed and nowhere else.**
`position: fixed; inset: 0` stops the document scrolling, and iOS Safari only
hides its toolbars while the document scrolls — so in a browser tab the frame
would cost about 100px of screen to Safari's own chrome, permanently, to solve
a problem that does not exist there. The haze is a standalone phenomenon;
`useStandalone` already decides the question, and it decides this one too.

Which means the app has two shapes, and both have to work: a document that
scrolls in a tab, a frame with a scroller inside it when installed. That is the
price of the frame and it should be stated plainly rather than discovered.

### What was checked, and what it costs

Each of these was looked at before any of it was written.

**Breaks — `useScrollGravity`.** It reads `window.scrollY` and listens for
`scroll` on `window`; in a frame neither ever moves, so the tile labels on `/`
stop being carried by the scroll. No error, just a dead effect. It already
takes a container ref, so it listens on the scroller and reads `scrollTop`
instead. One consumer, one page.

**Breaks — the menu's scroll lock.** `Menu.vue` sets
`document.body.style.overflow = 'hidden'` while the panel is open. In a frame
the body does not scroll, so the page would keep moving behind the open menu.
The lock moves to the scroller.

**Fine — viewport-height sizing.** Nine places use `min-h-screen`, `h-screen`
or `h-dvh`: the error page, the pool door, enrolment, three HiDrive media
views, the retro reader's sheet, and geheimtipp's layout with two of its pages.
Inside a frame that is itself the viewport, `100vh` still equals the frame, so
they keep meaning what they meant. The only one that changes is the layout's own
`min-h-screen` column, which the frame replaces.

**Fine — the footer.** It sits after the column so it begins at the fold. In the
frame it sits after the column inside the scroller, which is the same
arrangement; `min-h-screen` on the column becomes `min-h-full` of the scroller.

**Fine, and the one number that survives — `--content-shell-offset`.** Three
shared pages size themselves with `calc(100dvh - var(--content-shell-offset))`:
the content browser, search, and the asset library. Inside the frame `100dvh` is
still the window and the offset is still the header's height, so the arithmetic
is unchanged. Those files live in the CMS repo and are shared with the CMS's own
app, which has no frame — **they must not learn about one**. The offset keeps
meaning *what the shell spent above you*, and that is what crosses the boundary.

Nothing on this list blocks the frame. The frame itself is a handful of lines;
the two breaks above are the work.

### Tried, and taken back out

**The frame was built and it did not work.** Seventh experiment; it joins the
six in [[ios-scroll-edge-effect]]. The haze is a band of constant depth at the
top of the screen covering whatever is in it, and a fixed viewport-spanning
opaque container did not stop it — it only looked as though it had, because
everything in that band was black on black and the one light thing, the round
button, was outside the frame until it was not.

What the frame cost while it lasted: a second shape of the whole app, two paths
through the layout, a scroller, a three-way teleport target, and changes to
`useScrollGravity` and the menu's scroll lock. All reverted.

What survives it: `--sticky-top`, which is a genuine simplification independent
of any of this — six pages had each been choosing between the header's height
and the band's, and two of them chose wrong.

### What was done about it

`--status-bar` stays `env(safe-area-inset-top)` and nothing added to it. The
band is not made deeper and nothing is moved down.

Clearing the hazed zone geometrically means starting everything below
`inset + 30`. On the iPhone that puts the round button 81px from the top where
it sits at 51 — a larger downward shift than the 60px floor that was tried and
rejected, and paid for by the device that was already right. Two earlier
attempts at the same trade were rejected for the same reason.

**So the button gave up its white disc instead.** It is three bars and nothing
behind them: no fill, no ring, no shadow, taking `--shell-ink` from the band it
sits on and the menu's own ink once the menu is open. This is the reasoning the
way-back chevron beside it already followed — *the round, bordered, shadowed
button existed to be legible over a photograph, and over flat colour it is a
button wearing a coat indoors* — and it was simply never applied here.

The haze is not removed. It cannot be. What is removed is the thing it had to
work on: a blur over a white circle on a dark band is the strongest edge on the
screen, and over three thin bars there is almost nothing to smear.

A floor remains available if Apple's extent ever moves far enough to need one.
The number to put in it is now measurable rather than guessable: re-read
`/debug/haze` and take `last blurred stripe − inset-top`.

[forum]: https://developer.apple.com/forums/thread/803378
[pr]: https://github.com/feuerwehr-oberwil/kp-rueck/pull/143

## The rules that follow

These are not preferences. Breaking one of them puts the bug back.

- **One colour, two surfaces.** The header and the document background are a
  single value. Set apart, the status bar changes colour as the page scrolls.
- **A layout that draws a page draws the header.** Below `lg` that is
  `env(safe-area-inset-top) + 3rem` of flat colour. A layout may leave it out,
  and then that area's own pages are the status bar.
- **Nothing high-contrast in the top `inset + 30px`.** That zone is hazed on
  every device, including one that reports an inset of zero, and on a phone it
  is out of a thumb's reach as well. Flat colour there costs nothing; an edge
  there is exactly what the blur shows. Where a control has to sit in it, it
  goes without a fill, a border or a shadow.
- **A flat colour.** Not a gradient, not an image: the system samples it.
- **The indicators are not ours.** iOS picks light or dark from the same
  surface. Black and white are safe; a mid-tone is a coin toss. This is the real
  limit on "any colour", and it applies to our own ink as much as to Apple's.

## Mobile, decided

- The header carries **the way back on the left and the round menu button on the
  right**, in its 3rem control row.
- **No bar along the foot.** `AreaBar.vue` and `MainMenu.vue` were already
  orphaned at `HEAD` — nothing rendered them — so this confirms a decision that
  had been taken and not written down. The comment in today's `AppBar.vue` that
  still describes a foot bar is stale.
- The only thing left at the foot is **the player**.

The cost is stated rather than argued away: the top right corner is the farthest
point from a thumb on a phone.

## Desktop, decided

- **The area dropdowns go.** One concern per dropdown was how twelve areas were
  made to fit a row, and the panel already holds the same list with room to
  group it properly. A dropdown costs two interactions, open and choose; so does
  the menu. Navigation is the menu, at every width.
- **The menu groups by concern.** The panel renders a flat list today, which is
  why the dropdowns were the only place the grouping showed. `useAreas` already
  returns the grouped shape — `sections`, with a label per concern — and nothing
  reads it.
- **The menu button stays, at both widths.** With the dropdowns gone it is the
  navigation, not an extra; and the panel remains the only place the account and
  signing out live.
- **The disc stays.** Its cost was never the animation, it was the button
  sitting in two different corners: two origins, a `useMediaQuery`, a
  `clearance` with two cases, and arithmetic to keep it clear of the player. One
  corner removes all of that. What is left is one `ResizeObserver` on the panel
  and `discRadius`, which is a tested pure function.
- **The breakpoint stays `lg`.** The wide rendering needs room for a wordmark,
  three or four dropdowns, a title, page actions and the button; at `md` that is
  already tight, and an iPad in portrait is a touch device where the narrow
  rendering's reasoning holds.
- **The header is not gated on a session; its contents are.** Today the shell is
  drawn only for a session, so a public page has no strip — and on an installed
  iPhone that means the status bar takes whatever that page happens to have at
  the top. Navigation genuinely needs a session and the way back needs somewhere
  to go, but a colour needs only a page. So: the strip always, the areas and the
  menu with a session, the way back when there is one.
- **Three renderings of the area list**, each with a job of its own: the tiles
  on the front page, the grouped column in the panel, and the footer's.
- **The retro reader keeps the header.** It turns the bar off today and keeps a
  floating button as its only way out; once the button is in the header, the two
  cannot be separated.

Nine pages carry the flag that exempts them: `enroll.vue`, and eight of the
pool's. It is called `public: true`, and it does not mean what it says — in
`app.vue` it means only *do not raise ihlvn's sign-in modal here*. The pool is
not open to the world; every page of it sends you to its own login, and the flag
is there because it is gated by its own authentication rather than by this app's
session. `enroll.vue` carries it for the opposite reason: a one-time link, and
no account yet.

**Rename it `unguarded`.** The word `public` has already cost a plan rewrite —
it was read here as "open to the world" and written into a document as fact.
`unguarded` says the thing that is true and acted upon: this app does not guard
this page. The pool-versus-enrolment distinction is real but nothing branches on
it — not the modal, not the header's contents, not the entitlement middleware,
and not an account menu, since the pool draws its own. A third value nobody
reads would join `useMenuSize`'s unused sizes and `public.api.origin` on the
list below.

The pool draws its own chrome from its own layout, so the only page the session
rule changes is `enroll.vue`. Worth noting separately: the pool's top is a
photograph, so on an installed phone its status bar takes whatever the grass
gives — not wrong, but it should be a decision rather than an accident.

Still open: whether the wide header scrolls away with the page or stays. Sticky
costs nothing — it stays in the flow, so it still reserves its own height — and
the reason recorded for scrolling away is an iOS one, which is a phone matter
and does not reach the wide rendering.

## One header, two renderings

Today there are two tops: `AppBar.vue`, drawn by the layout from `lg` up, and
the "ceiling" inside `AppShell.vue`, drawn below it. They are the same job and
they become one component, **`Header.vue`**, with the width switch inside it.

The name pairs with `Footer.vue`, which is already there unprefixed —
`components` is registered with `pathPrefix: false`, so the tag is the bare file
name and `<Header />` sits beside `<Footer />`.

With the area dropdowns gone the two renderings hold the same things: the way
back or the wordmark, the page's title, the page's actions, and the menu button.
What is left of the difference is not contents at all:

- the zone under the status bar, which exists only below `lg`
- in the flow from `lg`, fixed below it
- the band, which hangs under the wide rendering only
- the row stops at `--shell-width` on a wide screen; on a narrow one it is the
  window

The one difference that reaches outside the component is positioning. From `lg`
the header is in the normal flow and pushes the page down by itself; below `lg`
it is fixed and the page has to leave room for it. Both heights are known, so
this is a variable and not a measurement:

```css
--top-height: calc(env(safe-area-inset-top) + 3rem);  /* fixed, the page pays */
@media (width >= 64rem) { --top-height: 0px; }        /* in flow, pays itself */
```

```html
<main :style="{ paddingTop: 'var(--top-height)' }">
```

Two files for two breakpoints is the version that rots, because they drift. If
`Header.vue` grows unwieldy — `AppBar.vue` is 163 lines before it absorbs
anything — the split to reach for is by *contents*, an areas row and a controls
row, with the outer element staying one component.

## What the page contributes

The header takes a **title** and **actions** from the page. Without that, pages
grow bars of their own — the mediathek detail page has one, and it duplicates
the way back that the header already draws.

By teleport, which is what this house already uses to put a page's form into the
shell's sidebar:

```vue
<!-- Header.vue -->
<div class="grow truncate"><slot name="title" /></div>
<div class="flex shrink-0 items-center gap-1" id="header-actions" />
```

```vue
<!-- a page -->
<Teleport to="#header-actions" defer>
  <UButton size="xs" variant="ghost" icon="i-lucide-maximize" … />
</Teleport>
```

`defer` matters: the page mounts before the header on a first load, and without
it the target does not exist yet.

The row reads **back · title · actions · menu** at both widths, and there is
room for it at both: the areas are in the menu, not in the row.

The title is **centred in the window**, which means it is out of the row's flow.
As a third flex item it was centred between two boxes of different widths, and a
long title ran under the buttons on the right instead of shortening. Absolutely
positioned with gutters wide enough for either side, it shortens first and
collides with neither.

Where no page has claimed the title, the header says which area this is, so the
middle of the row is never empty. `PageTitle` is what a page uses instead of a
bare teleport: mounting claims the title and unmounting gives it back, which is
how the header knows whether to print its own.

At the left, where there is no way back — the front page and the area roots —
the way home is **the cloud alone**, not the whole wordmark. The name is the
size of a heading and that corner holds a control.

**What this does to the mediathek detail page**, which is the case that prompted
it: its `<nav>` loses the back link and the title and keeps the two buttons that
belong to the video — a width toggle and fullscreen. Those stay with the player,
because they are about the picture rather than about the page.

**And it has a bug to fix on the way.** That player is `sticky top-0`, which
sticks it to the top of the window — underneath the fixed header. On a phone the
top ~95px of the video is hidden, invisible today only because both are black.
With the height published as `--top-height` it becomes `sticky
top-(--top-height)`. A page cannot read Nuxt state from a class, which is the
second reason that height is a CSS variable and not a measurement in state.

## What a page still configures

Fourteen pages set `shell` meta today, and the use is lopsided: eleven of them
set nothing but `background`.

| key | | |
|---|---|---|
| `background` | stays | the page's own colour. Eleven users, untouched by any of this |
| `top: { colour, ink }` | new | the header and the document behind it — the status bar |
| `footer` | stays | desktop only, one user |
| `band` `bandHold` `bandTo` `bandVia` | stay | the desktop gradient — kalender, and `/content` through the layer hook |
| `bar` | **goes** | no page refuses the header any more; presence is the layout's |
| `menu` | **goes** | the values were about where a floating button sits, and it no longer floats |

`bar` was a tri-state: drawn, not drawn, or pinned.

**Pinned is the default.** The wide header stays put while the page scrolls.
With the dropdowns gone the menu button is the whole of the navigation, and it
should not be able to leave the screen. A page can still say otherwise — that is
what the key was for — but no value for it is added until a page asks, or it
joins the dead configurability below.

**Nothing turns the header off any more.** The two pages that did both keep it:
the retro reader, because with the button in the header, turning the header off
takes the way out with it; and `[...slugs]`, the content viewer, for the reason
in the next paragraph.

`menu` had three users and each loses its reason. `filme`'s `'fixed'` and the
reader's `'static'` were about pinning a button that floated over the page; in
the header it floats over nothing.

`[...slugs]`'s `false` says "a content entry is shown full-bleed and carries its
own way around" — and the second half of that is not true of any of them.
`PageExhibition`, `PagePerson` and `PageWork` have no navigation at all, and
`PageFilm`'s chevrons are the player's transport, a frame back and a frame on.
Installed, with no address bar and no back gesture, all four content types are
dead ends. The header costs them 3rem over a photograph and buys a way out; and
on a phone that zone is the status bar anyway, so without a header it is a
picture that decides what colour the bar is.

None of them has anything to give up, then — but all four could hand a title to
the header rather than leave it empty.

`menuPinned` in `useShell` goes with them: it existed only so a page could pin
the button separately from the bar.

## Where things live

Almost everything lives in the layout. One thing cannot.

**The player must be outside the layout.** Moving between a page that uses
`default` and one that uses geheimtipp's layout destroys the first layout and
everything in it. For an `<audio>` element that means the music stops.

**Everything else may be inside it,** including the header, the way back and the
menu. They are rebuilt when the layout changes, and that costs nothing. A layout
that wants a header draws one; a layout that does not, does not.

This replaces an earlier belief that chrome had to sit outside the layout in
order to appear on pages with `layout: false`. It doesn't:

- `/` sets `layout: false` but renders `<NuxtLayout name="default">` itself, for
  everyone except a pool-only account — and that account is meant to have none
  of this app's chrome. The `layout: false` exists to stop `setPageLayout`
  writing a layout name onto `/`'s route meta, which would then follow the
  visitor into the pool's pages.
- `/geheimtipp/login` renders no layout at all. It is ours, in our layer, and it
  draws itself on purpose: the pool is an area of this app with its own
  authentication and its own identity, and this app's chrome would offer a
  visitor standing at that door places they may not go.

So the split is one component wide: the player outside, the rest in.

## The page, schematically

Mobile, below `lg` — `Header`'s narrow rendering:

```
┌───────────────────────────────────────────────┐ ─┐
│                                               │  │  env(safe-area-inset-top)
│        21:16        ((o))  ▮▮▮  [18]          │  │  ~47px · system draws the
│                                               │  │  indicators · HAZED
├ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ┤  │  nothing of ours in here
│                                               │  │
│  ‹   Walking With Beasts · Eine…  ↔ ⛶   ☰     │ 3rem  control row
│  back        title (page)   actions(page) menu│  │  ink = var(--shell-ink)
└───────────────────────────────────────────────┘ ─┘  fixed · var(--shell-top)
┌───────────────────────────────────────────────┐
│                                               │  <main>
│                   <NuxtPage>                  │  padding-top: var(--top-height)
│                                               │  background = var(--page-bg)
│                  scrolls under                │
│                    the header                 │
│                                               │
├───────────────────────────────────────────────┤
│  ▶  ──────○────────────────────  1:04         │  the player, outside
└───────────────────────────────────────────────┘  measured → shell-floor
                                                   → main padding-bottom

behind everything:  html, body = var(--shell-top)
                    ← this is what the status bar samples at scroll offset 0
```

Desktop, from `lg` — the same component, wide rendering:

```
┌───────────────────────────────────────────────┐
│  ihleven     Walking With Beasts · …   ↔ ⛶ ☰  │  Header · --bar-height
│  ─────────────────────────────────────────    │  in flow (see Open)
├───────────────────────────────────────────────┤  rule stops short of ☰
│▒▒▒▒▒▒▒▒▒▒▒▒▒▒▒▒▒▒▒▒▒▒▒▒▒▒▒▒▒▒▒▒▒▒▒▒▒▒▒▒▒▒▒▒▒▒▒│  band · 2rem gradient
├───────────────────────────────────────────────┤  --ui-bg → --band-to
│                                               │
│                   <NuxtPage>                  │  <main>, nothing reserved
│                                               │
└───────────────────────────────────────────────┘  ← the fold
┌───────────────────────────────────────────────┐
│  Footer                                       │  reached by scrolling
└───────────────────────────────────────────────┘

no document override up here — that media query is below-lg only
--content-shell-offset = header + band, for the embedded CMS page
```

Who renders what:

```
app.vue
└── <UApp>                        providers: toasts, overlays, tooltips
    │
    ├── <NuxtLayout>              default │ geheimtipp │ none, for the pool door
    │   │
    │   └── layouts/default.vue
    │       ├── Header            one component, two renderings
    │       │     ├── < lg        inset · ‹ back · title · actions · ☰
    │       │     ├── ≥ lg        wordmark/back · title · actions · ☰
    │       │     └── Menu        the ☰ and what it opens
    │       ├── band              lg only
    │       ├── <main> <slot />
    │       └── Footer            lg only
    │
    │   …and, teleported to the end of <body> by Menu:
    │       scrim · disc · panel
    │
    ├── Player                    outside, or the music stops
    └── LoginForm (modal)         no session
```

**The menu lives in the header, button and all.** `Menu` already owns the button
it opens from, and the button belongs to the row it sits in, so `Header` is
where it is written.

**The hole stays, and so does the metric table.** The button cannot join the
row's flow: what it opens is teleported to the end of the document and therefore
paints over the header, so a button in the flow would be covered the moment the
menu opened — you could not press it to close. It is teleported with the rest,
fixed in the corner, and the row leaves it a hole as it always did. `metric.hole`
and `metric.rule` keep their jobs. What the move does buy is that the corner is
the same one at both widths.

What it opens — the scrim, the disc and the panel — is wrapped in
`<Teleport to="body">`. The app sets no `z-index` and stacks by document order,
and teleporting is how that rule is satisfied rather than worked around: it is
why Nuxt UI's own overlays ship without a `z-index` too. So the menu need not be
the last thing written in any file; its DOM ends up last in the body either way,
and a modal opened afterwards still comes out above it.

The disc still measures from the button's centre in viewport coordinates, and
`position: fixed` inside a teleport behaves normally, so none of the geometry
changes.

**The stacking question the move raised, and what it cost.** `Header` written
at the top of the layout is written before the page, and this app stacks by
document order with no z-index anywhere. A positioned element beats unpositioned
content whatever the order — but page content that is *itself* positioned beats
the header, and plenty of it is: Musik's album tiles are `relative`, and they
scrolled straight over the bar.

So the header is **written last in the layout and fixed at both widths**. Sticky
cannot be: it has to sit in the flow at the top to work, which is exactly where
it loses. The room it needs is `pt-(--top-height)` on the column, inside
`min-h-screen` rather than added to it, so a screenful is still a screenful.

A small consequence worth knowing: `document.querySelector('header')` now finds
the CMS content module's bar rather than this one.

## What that simplifies

The round button is top right at **every** width. That is the largest single
simplification on the table, and it comes from a layout decision rather than
from a refactor:

- `Menu` loses its `above` and `standing` props and the arithmetic that kept the
  button clear of whatever was playing
- the `useMediaQuery` inside it goes
- `clearance` collapses from two cases to one
- the panel's `max-lg:flex-col-reverse` — areas last so the thumb reaches them
  first — loses its reason and becomes one layout
- the disc grows from one place instead of two
- the panel hangs from the top at both widths, so `max-lg:flex-col-reverse` and
  the wordmark's two positions collapse to one each

And `AppShell` itself goes: its ceiling merges into `Header`, its menu goes into
`Header` with the button, and what is left outside is `<Player />`, which is
where it was before.

## Measuring, and not measuring

Nothing about the header is measured. Its height is known at both widths, so it
is a variable — `--top-height` — and the page reserves it in CSS.
`useState('shell-ceiling')` and the `ResizeObserver` that fills it both go.

**The device is measured, and that is the one place it belongs.** `/debug/haze`,
reached from the menu, is a page with none of the shell on it that exists to be
read off a phone rather than reasoned about: a loud background, a numbered 10px
ruler fixed at the top of the window, and `env(safe-area-inset-top)` printed as
a number. It is kept rather than deleted. The haze moved in four OS releases by
Apple's own bug report, and when it moves again this is the thing that says by
how much. Two rules it has to keep, both learned by breaking them: the
background must be a colour nothing else on the screen is — a dark test page
under a dark system bar is indistinguishable from the thing it is testing — and
the ruler has to sit above anything opaque the test draws, or it measures
nothing.

`shell-floor` is gone too, and it never worked. The player was wrapped in a box
whose height was measured and published — but the player positions itself with
`fixed`, and a fixed child adds no height to its parent, so the box measured
zero and the page was told nothing. What the box did do was carry
`pointer-events-none`, which the player inherited: its buttons could not be
pressed and a click went through to whatever was behind it. On the music page
that was an album, so pressing skip opened a track list — and while that dialog
was open the overlay library marks the rest of the document unclickable, which
is why the menu stopped navigating as well.

The player is rendered unwrapped now. Nothing reserves room for it, which is
what was already true; whether it should is an open question.

What the chrome costs the embedded CMS page is published **once**, as
`--content-shell-offset`. Today two channels disagree — the layout forces the
offset to `0px` below `lg` while `<main>` is padded by the measured ceiling —
and that page, which frames itself against the window, is 48px too tall as a
result. Measured at 900px wide: `--content-shell-offset: 0px`,
`padding-top: 48px`, document 948px in a 900px window.

## A page that sends the header away

A film page is one picture and a list under it, and the thing worth seeing is
the picture. Such a page says so:

```ts
top: { colour: 'black', ink: 'light', scrolls: true }
```

and the header's **row** becomes `absolute` instead of `fixed`. That is the
whole of the difference: out of the flow either way, so the room reserved for it
does not change, but one of them leaves with the page and the other does not.

**The band under the status bar is not part of that.** It is a separate fixed
strip of `--status-bar` height and it always stays. Installed, the system draws
its clock and battery over the top of the page and takes their colour from what
is painted there — and when the whole header left, nothing was, so the page
itself scrolled through it, blurred. An iPad showed this plainly; a notched
iPhone hid it, because there the film was black and so was everything else. The round
button leaves with it — a white circle sitting on a film is the noise the page
is trying not to have — while what the button opens stays fixed, because an
overlay that scrolled with the page behind it would be no overlay.

The picture then sticks at `top-(--status-bar)` rather than at
`top-(--top-height)`: the row is gone, but the band is not, and a picture rising
past it would put the page back under the clock. Everything the page keeps under the picture
stays under it.

The cost is stated: scrolled in, there is no way back and no menu until you
scroll up again, and installed there is no browser back either. On a page whose
whole purpose is one picture, that is the trade worth making.

Two pages ask for it: the mediathek's episodes, and `[...slugs]` — which is all
four content types rather than films alone, because `definePageMeta` is resolved
at build time and cannot ask what kind of entry it is about to show.

## The colour contract

**Built.** It is the one part of this plan that exists: `useShell` carries `top`,
`AppShell` publishes `--shell-top` and `--shell-ink` on the document, the strip
and the chevron read them, and `main.css` paints `html`/`body` from the same
value. It will move with the strip when that becomes `Header`.

Retro is the first page to use it: `var(--color-blue-800)` with light ink, on
both the shelf and the reader.

Whether there is a header is the layout's business. What colour it is, is the
page's.

```ts
type Top = { colour: string, ink: 'light' | 'dark' }

const shell = computed(() => ({
  // The header and the document behind the page: together, on an installed
  // iPhone, the status bar. The ink is stated rather than derived, so choosing
  // a colour means looking at what has to be read against it.
  top: { colour: 'black', ink: 'light' } as Top,
  background: 'var(--color-ral-7004)',
  bar: true, footer: true, menu: true, band: '2rem', bandHold: '40%',
  ...(route.meta.shell ?? {}),
}))
```

Both halves are written, always. There is no colour-only form: a page that picks
a colour and leaves the ink to a default is a page where nobody looked at the
contrast, and the way back and the menu button are the two things on the screen
that must not disappear.

Set in three places:

- **globally** — the default above
- **per area** — the layer's `pages:extend` hook, as `/content` already does for
  its band colour
- **per page** — `definePageMeta({ shell: { top: { colour: '…', ink: 'dark' } } })`

Published once and read everywhere, so the surfaces cannot drift:

```ts
const inks = { light: 'white', dark: 'black' }

useHead({ htmlAttrs: { style: computed(() =>
  `--shell-top:${top.value.colour};--shell-ink:${inks[top.value.ink]}`) } })
```

```css
@media (width < 64rem) {
  html, body { background-color: var(--shell-top, black); }
}
```

The header is `background: var(--shell-top)`, and the chevron and the menu
button take `text-(--shell-ink)` and `border-(--shell-ink)` — which is also what
lets the round button drop the border, shadow and white fill it wears to stay
legible over a photograph. In a flat header it is a button wearing a coat
indoors.

## Open

- **What the pool paints at its top**, now that it is what its status bar
  samples.
- **Whether the page should leave room for the player.** It never has: the
  measurement that was supposed to arrange it was taken from a box a fixed
  element contributes nothing to. The full-width rendering therefore covers the
  last line of a page, and the compact one is a floating pill that need not.

## To rename

- `AppBar.vue` → `Header.vue`, absorbing the narrow rendering
- `AppMenu.vue` → `Menu.vue`, moving inside `Header`
- page meta `public: true` → `unguarded: true`, in `app.vue` and the nine pages
  that carry it

## Deleted

- `AppShell.vue` — its three parts went to `Header`, to `Menu` and to `app.vue`
- `splitForBar` and `concernIcons` in `utils/modules.ts` — the foot bar and the
  dropdowns were their only readers. `appModules` stays: it is the list the
  others are built from, and only its *external* use was zero
- `useAreas().menus` and `sections` — the same two readers
- `public.api.origin` in `nuxt.config.ts` — read by nobody

And one correction rather than a deletion: `apiBaseURL`'s comment claimed the
entry layer reads it. It does not; it reads `contentApiBase`, which is the
difference between a host configuring a layer and a host giving up a name.

## Kept, and made configurable

`useMenuSize` stays, and the size is now a page's to choose:
`definePageMeta({ shell: { menuSize: 'lg' } })`. It used to be shared state that
nothing ever set, so only the middle size existed. Both the button and the
header read it from one place — the header leaves the button a hole and stops
its rule short of it, and a disagreement between them is the button hanging off
the end of its own line.
