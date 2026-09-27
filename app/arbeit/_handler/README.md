# package handler — the cloud11 HTTP layer, kept for reading

`handler.go` and its test, copied verbatim out of `github.com/ihleven/pkg` at
`origin/develop` (`f8c5da4`, 2022-10-16), where they were `arbeit/handler/`.
Nothing here is built, imported or wired to anything: the leading underscore
makes the Go tool skip the directory, exactly as it skips `testdata`. The same
applies to `_repository` next door, where the longer note lives.

Do not fix the import paths. The moment these compile they are code, and code
has to be maintained.

## What it is

One `http.HandlerFunc` over the whole of `arbeit.Usecase`. It predates
`net/http`'s pattern routing and everything this app routes with, so it is here
as a record of what the API *was* rather than as something to mount.

It dispatches by picking the URL apart with a regular expression —

    /(\d{4})/?(kw|KW|Kw|w|W)?(\d{1,2})?/?(\d{1,2})?

— and then branching on which of year, month, week and day came out. Four
shapes: a day (GET, and PUT falling through to GET), a month, a year, and the
list of years. The week branch is empty.

`arbeit_test.go` is the one test that came with it, and it is worth keeping: it
pins the nine URL shapes the regular expression was meant to accept, including
`2020/KW5` and a day inside a week. If the routes are ever rebuilt on
`ServeMux` patterns, this is the list to rebuild them against.

## What is wrong with it

Worth knowing before anyone reaches for it as a starting point:

- **The account is the literal `1`** in all six use-case calls. There is no
  identity in it at all.
- **CORS is wide open** — `Allow-Origin: *` with all methods and all headers,
  set on every response by two near-identical functions.
- **Every response body is empty.** It encodes into `w` through a
  `json.Encoder`, then writes a `var bytes []byte` that was never filled. The
  content it sends is the encoder's, and the explicit `Write` adds nothing;
  `Content-Type` is set after the body has already gone out, so it never
  reaches the client.
- **PUT falls through to GET**, which is how the updated day gets returned —
  but `UpdateArbeitstag` returns `(nil, nil)` regardless, so the re-read is
  what carries the answer, and only because the fallthrough is there.
