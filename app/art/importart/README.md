# importart

Reads the old *Kunst* PostgreSQL database and writes CMS entries for the
artworks and exhibitions.

```
./ihlvn import --bilder --ausstellungen            # builds the entries and prints them
./ihlvn import --bilder --deleted --write          # ...and writes them
```

Nothing is written without `--write`; `--write` needs `REPO_CONTENT` set. Run it
with the working directory in this repository — `.env` is read from the process's
working directory, so running the binary from elsewhere silently picks up a
different repository's configuration.

## The dumps

Three, all kept here rather than at a repository root, because they are the
input this package reads and are meaningless apart from it. They came from the
`art` repo, which is being retired; the live `wi20241020` Postgres they were
taken from is a local database that does not survive a machine.

### `pg_backup_wi-20241020.sql` — the one the importer reads

A plain-text `pg_dump` (client 14.13) taken 20 October 2024, complete, with 22
constraint/index statements and no sequences — the ids are plain integers carried
over from the legacy system.

```
createdb wi20241020
psql -d wi20241020 -f pg_backup_wi-20241020.sql
```

That is the database `--db` defaults to (`postgresql://localhost/wi20241020`).

**The data has not changed since.** A dump taken on 20 September 2026 from the
same database — by then on server 14.24 — was identical to this one apart from
the `\restrict` / `\unrestrict` wrapper that newer `pg_dump` clients emit. It was
discarded as a duplicate. So this file is both the current state of the archive
and exactly what the 2024-12-10 import read, which is why the entries in the
content repository match it row for row.

| table | rows | |
|---|---|---|
| `foto` | 551 | photographs: `bild_id, name, path, format, width, height, taken, caption, kommentar, status` |
| `bild` | 429 | the artworks — 308 live, 121 marked `DEL` |
| `ausstellung` | 123 | the exhibitions |
| `foto2` / `bild2` / `serie2` | 113 / 66 / 2 | an older parallel set; `bild2` has no `dir`, `gattung` or `status` |
| `serie` | 3 | series: `slug, jahr, titel, technik, hoehe, breite, tiefe, phase` |
| `katalog` | 2 | catalogues |
| `dokument`, `enthalten`, `enthalten2` | 0 | empty |

### `wi-20210528.sql` and `wi-20211022.sql` — earlier generations

Both `pg_dump` 11.9, and kept only as history: the archive roughly tripled
between them and 2024.

| | May 2021 | Oct 2021 | Oct 2024 |
|---|---|---|---|
| `bild` | 152 | 174 | 429 |
| `ausstellung` | 15 | 15 | 123 |
| `foto` | 257 | 298 | 551 |

They also hold a table that no longer exists: `ausstellung2`, 17 rows shaped like
`ausstellung` without `slug` and `fotos`. It was folded into the main table
before 2024 — 14 of its 15 distinct titles appear there. The one that does not is
*"Schwarzwald Bild 4"*, so these two files are the only record of it.

## What this importer takes, and what it leaves

Only `bild` and `ausstellung`. Everything else in the dumps has never been
imported, and there is no code here that would.

That matters most for `foto`. An artwork carries a single `Foto int` — a
`foto_id` pointing into a table that was never brought across — so the entries in
the content repository have no photo paths, dimensions, captions or upload dates.
Anything that wants to show a picture of a work, rather than its measurements,
needs those 551 rows first. The same goes for `serie` and `serie_nr`, which
reference series that were never imported, so an artwork knows it is number 3 of
a series it cannot name.

## The state of the import

It was run on 2024-12-10 and wrote 308 artworks and 123 exhibitions — the live
rows only, without the 121 marked `DEL`. `--deleted` includes those; they arrive
as inactive entries keeping their `DEL` status, which is indexed, so `?status=DEL`
finds them. Nothing has been imported since.
