# geheimtipp

The football pool from `ihleven.de`, rebuilt as a layer of this app.

> **"pool"** throughout this document and the code comments means the
> *Tipprunde* — a group who each predict the weekend's results and are scored
> against each other. English has no better single word for it, and the German
> one reads badly inside English sentences. The term is developer-facing only:
> every string a user sees is German, and the navigation says *Tipprunde*.
>
> Other terms kept in German because the payloads and the screens use them:
> *Spieltag* (matchday), *Wertung* (standings), *Tipp* (a prediction),
> *Ausgabe* / edition (a season of the pool), *Trikot* (the jersey a ranking is
> named after), *Klassement* (one such ranking).

## Why

`ihleven.de` is this app's target domain, and it is occupied: a Nuxt 3 frontend
serving the pool at the root, with its Go backend under `/api/v1` and `/media`.
The family app cannot move there while that frontend is running, so the pool's
pages have to be served from here instead.

This is a **frontend** migration only. The pool's backend, its database and its
accounts are untouched and stay where they are — a tip placed here lands in the
same database it always did. What is replaced is the Nuxt app in front of it.

## Requirements

These were stated, and each one shaped the design:

1. **The pool's code is read-only.** `~/src/geheimtipp` is a reference. Nothing
   is written there, including its git state.
2. **`/geheimtipp` boundaries are not broken.** Everything the pool keeps lives
   under that prefix. Their `/avatare` became `/geheimtipp/avatare`; their
   `/login` became `/geheimtipp/login`.
3. **The pool's sign-in is not this app's sign-in.** The pool has its own users
   in its own database. An ihlvn account grants nothing there and the reverse.
   The two share an origin and nothing else.
4. **Migrate the frontend as it is** — reproduce the screens rather than redesign
   them, except where something was stated as a deliberate change.
5. **ihlvn stays on `tschabrun.de`** until the integrated frontend is ready. The
   cutover is a separate step.

## How it works

### Reaching the backend

The pool's API is on another origin, so a browser here cannot call it. A reverse
proxy in `app/geheimtipp` (Go) mounts it under this origin:

| route | forwards to | default |
|---|---|---|
| `/ght/{path...}` | the pool's API | `https://ihleven.de/api/v1` |
| `/ght/media/{path...}` | its media | `https://ihleven.de/media` |

Both upstreams are configuration (`GHT_API`, `GHT_MEDIA`), and the UI reads its
base from `public.geheimtippBase`. So serving both from one origin later is a
config change, not a rewrite — and the proxy may disappear entirely.

### Sign-in

**There is one sign-in, and it is this app's.** The form on `/` takes a pool
login and password just as the pool's own did, so someone who has always played
types what they have always typed. The first time, an account is created behind
them from `geheimtipp_migration` — see `Staging`-style notes in
`app/db/migrations/0003_geheimtipp.sql`. After that they are an ordinary account
here, confined to the pool.

**The pool's credential never reaches the browser.** The backend reads its JWT
from a **`token` cookie only** — an `Authorization` header is ignored — so the
token has to be a cookie *on the request the pool sees*, which is not the same
as a cookie in the browser. `app/geheimtipp` mints one per forwarded request
from whoever is signed in here, and sets it on the outbound request only.

**A `token` the browser sends is never forwarded**, whether or not one is minted
to replace it. The pool's own sign-in used to leave one lasting weeks, and
honouring it would be a second way in — one that needs no account here, survives
signing out, and cannot be revoked. Anyone still carrying one signs in again,
which was the deliberate choice: there is one credential and this app issues it.

Consequences worth knowing:

- signing out of this app is signing out of the pool, because the credential is
  derived from the session rather than stored;
- there is nothing to expire: the minted token lives for one request;
- a deployment with no `GHT_SECRET` still proxies, anonymously, so the pool's
  public pages work without the pool being configured at all.

The proxy still rewrites `Path` on every cookie it forwards, so the pool's own
`Path=/` session cookie cannot land at the root of the family app.

There is a test for each.

### The gate

Pages under `/geheimtipp/**` keep `definePageMeta({ public: true })`, and that
flag must stay. It is about this app's sign-in overlay, not about access:
without it every visitor to the pool would meet an ihlvn sign-in dialog laid
over pages they are welcome to read. Removing it "for consistency" now that the
two sign-ins are one is the likeliest way to break this.

A layer-local global middleware (`ght.global.ts`) checks this app's session
first — being signed in here is being signed in to the pool — and only asks the
pool for someone with no account at all. An unknown caller goes to `/`, the one
sign-in form. It is a UX guard, not a security boundary: the pool's backend
checks its token on every write.

`/geheimtipp/login` still exists but only forwards to `/`. The route is kept
because links to it are: the pool's own pages pointed there for years.

### State

Four composables, each owning one thing:

- `useGhtSession` — `GET /aktuell`. Both the session probe (`authkey`) and the
  landing page's content, so it is one request and one piece of state.
- `useGhtEdition` — `GET /ausgaben/:code`. The tippers of an edition, which is
  where every screen gets names and avatars from; the other payloads key by login
  and carry neither.
- `useGhtTipps` / `useGhtResults` — the two writes.
- `useGhtDates`, `useGht` — formatting and fetching.

### Layout

Two bars, and which menu belongs to which is the point of having two. The header
over the grass is the pool: the same on every page whatever edition is open, so
the **account** menu sits there. The bar below is the **edition** — its pages,
its name, and the registration this person plays it under.

Below everything, and only sometimes, the **family app's footer**. A pool
account is confined — the server hands it one area and refuses it everywhere
else — so an account entitled to anything besides the pool is one that arrived
here from a family app it can go back to, and the footer is that way back. For
everybody else, including a visitor with no account, the page is unchanged: the
footer is a column of areas, and for a pool player it would list none of them
and offer doors that bounce whoever opens them.

It reads entitlements rather than a "confined" flag, because that is the
question it is actually asking — is there anywhere else for this person to go?

## Routes

### Ported

| route | from | notes |
|---|---|---|
| `/geheimtipp` | `geheimtipp/index.vue` | last/current/next matchday, season from `/aktuell` rather than hardcoded |
| `/geheimtipp/spieltag-<saison>-<spieltag>` | same name | the tip matrix, both standings tables, the league table |
| `/geheimtipp/matches/[id]` | `geheimtipp/matches/[id].vue` | |
| `/geheimtipp/ranking` | `geheimtipp/ranking.vue` | Wertung |
| `/geheimtipp/spieltage` | `geheimtipp/spieltage/index.vue` | |
| `/geheimtipp/register` | `geheimtipp/register.vue` | |
| `/geheimtipp/login` | `login.vue` (their root) | no layout |
| `/geheimtipp/avatare` | `avatare.vue` (their root) | now addresses images by id |

### Left out

| route | why |
|---|---|
| `geheimtipp/tipper/index.vue`, `tipper/[ght].vue` | nothing links to them; still to do |
| `geheimtipp/[ausgabe]*` — 16 pages | a parallel, older navigation: edition → competition → matchday, plus `teams`, `groups`, `klassements`, `spiele`. Their own nav does not point at any of it and the flat routes cover the same data. Recommend dropping. |
| `geheimtipp/[ausgabe]/register.vue` | duplicate of the flat one |
| `/` (their root) | `/geheimtipp` is the pool's landing page; the root belongs to the family app |
| `/BL[[tier]]/**` | a league browser around the pool rather than part of it; its entry point 500s |
| `/wiki/**`, `/WM[season]/**`, `/debug`, `/EM2024*` | read before deciding: the wiki page has its `<ContentDoc>` commented out, the WM page is a layout sketch, debug dumps the runtime config |

## Problems on the way

**The reference is `origin/develop`, not the working tree.** The checkout sits on
`feature/ausgabe-2024` (2023), three years stale — it reads `tipp.ht`/`tipp.at`
where the API sends `Result: [2,1]`, and hardcodes season 2024. Read it with
`git show origin/develop:<path>`.

**Reading source is not looking at the site.** This cost the most time, twice.
The pool was recorded as "public" from a 302 that was never followed — every page
redirects to `/login`. And the edition dropdown was modelled on `TipperMenu.vue`,
which is used on exactly one page and is not one of these — what renders is the
`Menu` inside `Navigation.vue`, white and without the sign-out that was copied
from the wrong component. A component existing in the tree is not evidence that
the page uses it. Render it and measure.

**Crests.** Their `assets/logos/clubs/` has 74 SVGs, and three current clubs
(`SGE`, `SVW`, `HOF`) are not among them — they exist only inlined in
`ClubLogo.vue`, which is ~3,500 lines of `v-if`. Those were extracted, the
aliases copied, and two files repaired that were missing XML namespace
declarations, which silently makes an SVG unrenderable as an `<img>`.

**Every crest rendered zero pixels wide.** The preflight rule
`img { max-width: 100% }` resolves against a table cell that has no width of its
own, and collapses it. Their component inlines an `<svg>`, which that rule does
not match — shipping the crests as files is what introduced this. `max-w-none`
fixes it.

**Assets cannot live under `/ght`** (the proxy owns every path there) **or under
`/geheimtipp`** (a directory of files shadows the SPA route of the same name and
the server 301s). They are at `/assets/geheimtipp/`.

**Custom classes.** `.text-outline` — white text with a four-way black outline —
is in the pool's own `tailwind.css`, not Tailwind's. Nothing in the ported markup
referenced it, so the numbers on the jerseys rendered dark. The layer's own
classes now live in `app/assets/css/geheimtipp.css`, prefixed `ght-` so the lint
rule that checks classes against Tailwind can be told about them by shape.

**Two copies of the same person.** A tipper is described both in `/aktuell` as
`registration` and in the edition as `tipper[]`. Saving the registration page
refreshed only the first, so the ranking and the matchday headings kept the old
name until the page was reloaded — the same bug their site has. Both are
refreshed now.

**`101:0` in the tip matrix is real data.** Someone typed it. Their site shows
the same.

## Verifying without touching the pool

The pool is productive and shared with eleven other people. A stray tip is a real
prediction; a stray result rescores everyone's season.

So writes were driven against a **replay stub**: recorded GETs served back, PUTs
recorded and answered, with the binary pointed at it by `GHT_API`. That confirmed
the URLs, the form bodies, batching, the per-match refusal path and the round
trip through to "gespeichert am". Recordings are deleted afterwards — `/aktuell`
carries an email address.

The live tip write was confirmed separately by the user. **The result write has
not been exercised against the real backend**; its contract was read off their
`submitResults`.

## Open points

**To decide**

- **`DeSandro.vue`** — the wordmark. The effect is a copied CSS demo (David
  DeSandro's, via David Walsh). The version on `develop` is already stripped of
  the shadow effect, so porting is small; the real question is its font stack
  (`proxima-nova`, Yanone Kaffeesatz).
- **`Footer.vue`** — half of it is dead Bootstrap markup with hardcoded links
  back to 2004. A faithful copy is probably not what is wanted.
- **`HeaderGhtMenu`'s remaining content** — their panel also carries a disabled
  "Daten ändern".
- **The `[ausgabe]` subtree** — drop or port.
- **`icon.png`** ships at 865×865 and 404 KB for a 48px render.

**To do**

- Tipper list and profile.
- The refresh button on the matchday header.
- Exercise the result write against the real backend once.
- The cutover to `ihleven.de`, at which point the pool's frontend is retired and
  `/ght` either repoints or goes away.

## The club crests: provenance and rights

**Not legal advice.** This records what was checked and what remains open.

### What the originals are

Their `ClubLogo.vue` is one component of ~3,500 lines holding every crest as
inline `<svg>`, selected by a chain of `v-if="club === 'BVB'"`. Alongside it
`assets/logos/clubs/` holds 74 standalone SVGs — an older, wider set, including
historic variants (`SGE_1980-1999.svg`) and clubs not in the league.

Neither set is a source of truth on its own. Three clubs currently in the league
— `SGE`, `SVW`, `HOF` — have **no file** and exist only inside the component,
and some of the files are older designs than the ones the component draws.

### Why we ship files instead

The component was not ported. Reasons, in order of weight:

1. **It is a chain, not a lookup.** Adding a promoted club means editing a
   3,500-line file; a missing one fails silently by rendering nothing.
2. **Every crest is in every bundle.** Inline SVG in a component is shipped to
   anyone who loads the page, whether their matchday involves that club or not.
   As files, the browser requests the eighteen it needs and caches them.
3. **The data does not carry them.** `teams[].logo` is always `""`, so the crest
   has to be resolved from the team code either way — a filename is the smaller
   indirection.

The 24 crests that exist only inline were extracted to files, three aliases
copied (`SVW`←`BRE`, `SGE`←`FRA`, `HOF`←`TSG`, which is how the component maps
them), and two files repaired that were missing XML namespace declarations.

One thing the inline form got for free and files do not: preflight's
`img { max-width: 100% }` does not match an `<svg>`, so their crests never
collapsed in an auto-width table cell. Ours did, until `max-w-none`.

### What was checked about rights

All 77 files were inspected for provenance and licence metadata:

- **47** carry an Adobe Illustrator generator comment, **2** Inkscape, **1**
  CorelDRAW.
- **27** carry no marker at all — and those are exactly the ones extracted from
  the component here: the originals do have the comment, and extraction started
  at `<svg`, dropping everything before it. Not evidence of anything.
- **No file carries a licence.** No `<cc:license>`, no `<dc:rights>`, no
  `<dc:creator>`, no licence named in any comment. One empty `<dc:title>`.
- The three files that match "commons" match only on the `xmlns:cc` namespace
  declaration that Inkscape writes into every file it saves. Not a licence
  statement.

**So the Wikipedia question cannot be answered from the files.** There is no
provenance in them. Bundesliga crests do exist on Wikimedia Commons, and could
plausibly be where some came from — but Illustrator exports are equally
consistent with having been traced, bought, or taken from a press kit.

### What matters more than the licence

Even granting a free licence, it would settle **copyright** only. On Commons
these crests are typically tagged as below the German threshold of originality
(*Schöpfungshöhe*) — hence no copyright — while being explicitly marked as still
carrying a **trademark**. A club crest is a registered mark whatever the SVG's
copyright status, and displaying one identifies a football club, which is the
one thing trademark law is about.

Practically: the pool has shown these crests for years on the club's own terms —
a private tipping game among a dozen people, not a commercial use, not implying
endorsement. Serving the same pictures from a different host does not change
that. It is worth knowing, not worth stopping for.

**Open:** if the pool ever becomes public-facing or commercial, the crests are
the first thing to revisit — either licensed properly, replaced with the team
codes the components already fall back to, or dropped.

## Notes

- **Email addresses.** `GET /aktuell` returns the *caller's own* record and that
  is where the account menu takes name and address from. `GET /geheimtipper`
  returns one for everybody who has ever played. Nothing here reads that
  endpoint's copy, and nothing should.
- **Result entry** is offered to `public.geheimtippAdmin` (default `matt`). Their
  frontend writes that login into a component; which account runs a deployment is
  a property of the deployment. It is presentation either way — the backend
  refuses the write from anyone else.
- `origin/feature/modernize` is `develop` on nuxt 4 + @nuxt/ui 3 + bun. The
  commit is build-level only — files moved, components untouched — but it proves
  their pages run on nuxt 4 unmodified.
