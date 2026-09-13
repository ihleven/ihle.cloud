---
name: changelog
description: Use when a user-visible change has just been completed, when the user asks to update, write or start a CHANGELOG, or when cutting a release. Keeps CHANGELOG.md written as the work happens rather than reconstructed from commits at release time.
version: 1.0.0
---

# Keeping the changelog

The changelog is written **while the change is being made**, by whoever made it,
because that is the only moment the reason is still known. Reconstructing it from
commit messages before a release produces a list of what was touched, not a
record of what changed for anyone — the "why" is gone and the reader gets a diff
in prose.

So: finishing a user-visible change includes writing its entry. Not later.

## Where

`CHANGELOG.md` at the repository root. Format is
[Keep a Changelog](https://keepachangelog.com/en/1.0.0/):

```markdown
## Unreleased

### Added
### Changed
### Fixed
### Removed
### Security
```

New work always goes under `## Unreleased`, at the top of the file. Released
sections sit below it as `## [1.2.0] - [2026-04-10]`, newest first.

**Read the existing `###` headings before writing.** Merge the entry into the
heading that is already there; do not append a second `### Changed`. Changelogs
maintained by appending drift into three `### Changed` sections and nobody
notices until release.

## What an entry looks like

A **bold lead sentence** naming the outcome, then prose explaining what changed
and why it matters — written for someone who uses the software and does not read
its source.

```markdown
- **A film is addressed by which film it is, not by where its file sits**:
  previously the page was handed the path of the video inside the storage
  provider and asked for it by that path, which put the storage layout into the
  address bar. Now the page asks for the film and the server looks up where the
  bytes are, so moving a file is a change to the film's entry rather than to
  every link that pointed at it.
```

Rules for the prose:

- **What and why, never how.** No file names, function names, type names, flags,
  routes or commands. If an entry cannot be written without them, it is probably
  not a user-visible change (see below).
- **Outcome first.** The lead sentence should be readable alone, in a list of
  twenty others.
- **Say what a reader must do.** If a credential was exposed, say it should be
  replaced. If behaviour changed in a way that needs attention, say so.
- **Explain the fix's consequence, not its mechanism.** "Captions never appeared,
  because the timings were stored in one place and read from another" — not the
  parsing change that fixed it.

## Which section

- **Added** — something that was not possible before.
- **Changed** — something that behaves differently than it did.
- **Fixed** — something that was supposed to work and did not. Fixed entries
  describe the *symptom the user saw*, then why.
- **Removed** — something that is gone. Say what replaces it, or why it is not
  needed.
- **Security** — or mark the entry `(security fix)` inside Fixed. Always say
  whether action is required.

## What does not get an entry

Refactors, renames, test additions, formatting, dependency bumps and internal
restructuring — **unless** they change what someone outside the code can do. A
package split is not an entry; "another application can now reuse these endpoints
without adopting this project's web framework" is.

When in doubt: could a user or an integrator notice, or would they want to know?
If no, skip it. A changelog padded with internal churn stops being read.

## Revising, not accumulating

An unreleased entry is not yet published, so it can still be **edited**. When
later work changes or supersedes something already written under `## Unreleased`,
rewrite that entry to describe the final state. Do not add a second entry that
contradicts the first, and do not narrate the intermediate steps — nobody needs
to know that a rule was 4px before it was 1px.

This is the main advantage of writing continuously: the Unreleased section is a
draft of the release notes, kept true, rather than an append-only log.

## Cutting a release

1. Reread `## Unreleased` as a whole. It should read as release notes: merge
   duplicated headings, drop anything that turned out not to matter, and make
   sure superseded entries describe the end state.
2. Rename it to `## [x.y.z] - [YYYY-MM-DD]`.
3. Add a fresh empty `## Unreleased` above it.

Nothing is generated from commits at this point. If something is missing, that is
a signal the entry was skipped when the work was done, not a reason to start
scanning the log.

## Starting a changelog in a project that has none

Create the file with the header, the Keep a Changelog link, and a single
`## Unreleased` section covering the work in progress. **Do not reconstruct past
releases from git history** — that is guesswork presented as record. Say so
instead:

```markdown
This file starts here; anything before the entries below is only in the git
history.
```

## Checks worth running before finishing

- Does the file contain a repeated `###` heading inside one section?
- Does any entry contain a file name, function name, flag or route?
- Does each lead sentence stand on its own?
- Is anything in Unreleased now false, because later work changed it?
