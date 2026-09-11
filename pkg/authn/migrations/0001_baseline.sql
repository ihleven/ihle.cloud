-- The authdb as it stood before this package existed.
--
-- The schema had never been checked in: it lived only in the running database.
-- This file records it verbatim so the starting point is reviewable and so a
-- fresh environment can be built from scratch. Every statement is guarded, so
-- running it against the existing database changes nothing; 0002 does the
-- actual reshaping.

-- Identity. Note account.name is the login (and what the Go code calls
-- Account.ID), while display_name is the human name.
create table if not exists account (
    id           bigint generated always as identity primary key,
    email        text,
    name         text,
    display_name text,
    -- An invitation code sketch: plaintext, one per account, with no record of
    -- having been used. 0002 replaces it with the enroll_token table.
    inv_code     text,
    inv_expiry   timestamp with time zone,
    -- {"username": ..., "password": ...} — the username half is never read.
    credentials  json,
    -- {"alias": ..., "home": ..., "root": ...}; alias points into hitoken.
    hidrive      json
);

-- The CMS-facing half of an account, keyed by account.name.
create table if not exists cmsauth (
    id          text primary key,
    groups      text[],
    permissions text[]
);

-- HiDrive OAuth refresh tokens, keyed by the alias an account points at.
-- Aliases are shared: the anonymous account uses 'ihleven'.
create table if not exists hitoken (
    alias         text primary key,
    scope         text,
    refresh_token text,
    expires_at    timestamp with time zone
);

-- A second, unused HiDrive token table that predates hitoken. Dropped by 0002.
create table if not exists hidrive (
    id            text primary key,
    refresh_token text,
    expiry        timestamp with time zone
);

-- An abandoned first attempt at passkey storage: no primary key, no foreign
-- key, every credential for a person in one JSON array, and keyed by an email
-- that does not match account.name. Dropped by 0002.
create table if not exists webauthn (
    name         text,
    display_name text,
    credentials  json,
    id           bigint
);
