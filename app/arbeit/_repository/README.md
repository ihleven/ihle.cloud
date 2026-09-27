# package repository — the cloud11 storage layer, kept for reading

This is the Postgres implementation of `arbeit.Repository` and the schema it was
written against, copied verbatim out of `github.com/ihleven/pkg` at
`origin/develop` (`f8c5da4`, 2022-10-16), where it was `arbeit/repository/`.
`schema.sql` sat at that repository's root and is filed here because this is
what it describes. Nothing is built, imported or wired to anything.

It is here because the `arbeit` package next door has a storage seam and no
implementation, and because this is the fullest surviving description of the
shape the data had. A future implementation against this app's pgx pool and
migrations has to decide the same questions this code answered; better to
read the answers than to guess at them again.

## Why the directory is named `_repository`

The Go tool ignores any directory whose name begins with `_`, exactly as it
ignores `testdata`. So these stay readable, greppable and syntax-highlighted
while `go build ./...`, `go vet ./...` and `go test ./...` never see them —
which matters, because they import `github.com/ihleven/errors`,
`github.com/ihleven/pkg/arbeit`, `jmoiron/sqlx` and `lib/pq`, none of which
this app depends on.

The underscore is the only change to the name: the package is `repository` and
the directory says so, so a file found later can be matched against the thing
it was. The HTTP layer is next door under the same rule, in `_handler`.

Do not fix the import paths. The moment these compile they are code, and code
has to be maintained.

## What is in here

| file | what it holds |
|---|---|
| `schema.sql` | the `c11_*` tables. Not runnable as-is: it opens with seven unqualified `DROP TABLE`, and it creates an `account` table that collides with this app's |
| `db.go` | the sqlx/lib-pq connection. Credentials are in the source — database `cloud11`, user `matt`, password `dummy` |
| `arbeitsjahr.go`, `arbeitsmonat.go`, `arbeitswoche.go`, `arbeitstag.go`, `zeitspanne.go`, `urlaub.go`, `job.go`, `kalendertag.go` | the queries, one file per table |
| `setup.go` | `SetupJobs`, which creates the four employers and every year and month inside them. Hardcoded for account 1 |
| `migration.go` | a one-off import from the *predecessor* database — see below |

## What was verified about it

The read path matches `schema.sql` exactly, including the `d.datum as tagdatum`
alias that keeps the joined calendar day from colliding with the working day's
own date. `RetrieveArbeitsjahre` scans its 22 columns in declaration order.

Three write paths do not, and did not before the move:

1. **`InsertArbeitstag` cannot execute.** Sixteen columns, fourteen
   placeholders — Postgres rejects it at parse time. `jahr` and `monat` are
   `NOT NULL` with no default and no value is passed for them, because
   `Arbeitstag.Jahr` and `.Monat` were commented out of the domain struct and
   the column list was never followed through. This is what `SetupArbeitsjahr`
   calls 365 times.
2. **`UpsertZeitspanne` violates a not-null.** It omits `job`, which is
   `NOT NULL` and half of the foreign key back to `c11_arbeitstag`. Same cause:
   `Zeitspanne.Job` is commented out. The error is `23502`; the code only
   tolerates `23505`.
3. **The use case skips a foreign key.** `c11_arbeitstag` references
   `c11_job (account, code)`, but nothing reachable through the `Repository`
   interface creates a job row. Only `SetupJobs` here does, hardcoded.

So there is no working implementation to copy, which is the likely reason the
calendar pages never had a backend.

## The predecessor schema, for when the Python app turns up

`migration.go` is not a schema migration. It reads a *different, older*
database and writes into the `c11_*` one, and in doing so it records what that
older schema looked like. It is not the only such record — see the second one
below, which is better. Django's fingerprints are on it —
`<table>_id` foreign keys, a `user_id` on everything:

- **`kalendertag`** — `id, jahr_id, monat, tag, datum, feiertag, kw_jahr, kw_nr,
  kw_tag, jahrtag, ordinal`, and (named in a comment, not selected)
  `monatsname, tagesname, kalendermonat_id, kalenderwoche_id`
- **`arbeitsjahr`** — `id, urlaub_vorjahr, urlaub_anspruch, urlaub_tage,
  urlaub_geplant, urlaub_rest, soll, ist, differenz, tage_arbeit, tage_krank,
  tage_freizeitausgleich, tage_buero, tage_dienstreise, tage_homeoffice,
  tage_frei, jahr_id, user_id`
- **`arbeitstag`** — `id, status, typ, urlaubstage, freizeitausgleich, krank,
  krankheit, soll, start, ende, brutto, pausen, netto, differenz, kommentar,
  tagebuch, arbeitsjahr_id, arbeitsmonat_id, arbeitswoche_id, dienstreise_id,
  tag_id, urlaub_id, user_id`
- **`arbeits_zeitspanne`** — `id, nr, typ, von, bis, dauer, beschreibung,
  arbeitstag_id`

Two things the older schema had that the `c11_*` one dropped: a `tagebuch` on
the day, and a `beschreibung` on the span. Two it had that were folded away:
`krank`/`krankheit` and `freizeitausgleich` became values of `status` and
`kategorie`.

The identifiers are dates: an `arbeitstag.id` is `YYYYMMDD` times 1000 plus the
account, and `tag_id` is the plain `YYYYMMDD` (`DateFromID`). `JobID` maps a
date to an employer for account 1, which is how the old rows — which had no job
at all — were assigned one. That mapping is the migration's one real judgement
and is not recoverable from data:

    DBIS   2001-08-20 … 2007-10-31   ALU
    UKLFR  2008-04-01 … 2013-01-31   Uniklinik Freiburg
    AVE    2013-02-01 … 2018-07-31   Averbis GmbH
    IC     2018-08-01 …              Inter Chalet

## The better source: the ihleven/drive repository

`git@github.com:ihleven/drive`, last touched June 2019 — the original cloud11
monolith, before anything was split out of it. Its `arbeit/` module is a Go
program that ran **directly against the predecessor database**, so it describes
that schema first-hand rather than through an import script, and it describes
more of it:

- `arbeit/entities.go` — `Arbeitsjahr` carries the columns one for one
  (`UrlaubVorjahr`, `UrlaubAnspruch`, `UrlaubTage`, `UrlaubGeplant`,
  `UrlaubRest`, `Soll`, `Ist`, `Differenz`, `tageArbeit`, `tageKrank`,
  `tageFreizeitausgleich`, `tageBuero`, `tageDienstreise`, `tageHomeoffice`,
  `tageFrei`, `jahrID`, `userID`), and its `Kalendertag` has the four fields
  `migration.go` names in a comment but never selects: `ordinal`, `monatsname`,
  `tagesname`, `yearday`.
- `arbeit/pg_arbeit/` — the queries themselves, against the unprefixed
  `arbeitstag` and `arbeitsjahr` tables.

So there are two independent records of the same schema and they can be checked
against each other. Start here; use `migration.go` for what this does not cover,
which is chiefly how the old rows were mapped onto jobs.

## What is not here

`arbeit/app.go` stayed behind: one live function (`IsContained`, now in
`zeitspanne.go` next door) and two hundred lines of commented-out sketch. It is
still in the `pkg` repo:

    git -C ~/src/pkg show origin/develop:arbeit/app.go
