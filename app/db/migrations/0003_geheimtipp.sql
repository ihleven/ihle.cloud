-- Unified sign-in: one login form for both this app's accounts and the
-- geheimtipp pool's players.
--
-- Two additions. An account kind, so a pool player's account can be confined to
-- the pool; and a table holding what is needed to create such an account the
-- first time its owner signs in.

-- ---------------------------------------------------------------- account ---

-- What the account may do — not where it came from.
--
-- 'full' is governed by entitlements, as every account is today. 'geheimtipp'
-- is confined to the pool and refused everywhere else, whatever permissions it
-- may have acquired. Someone who is both a family member and a player is
-- 'full': the confinement is for people who have no business here beyond the
-- tipping.
--
-- Defaulting to 'full' leaves every existing row correct.
alter table account add column type text not null default 'full'
    check (type in ('full', 'geheimtipp'));

-- ------------------------------------------------------ geheimtipp_migration ---

-- The pool's players, as far as signing them in requires.
--
-- It exists so that a pool player meets the new login form, types the username
-- and password they have always used, and is signed in — without being asked to
-- do anything. On the first successful sign-in an account is created from this
-- row and the row is deleted, so the table shrinks toward empty and "is the
-- migration finished" is a count.
--
-- The duplication is deliberate. The pool's own data is moving to a database of
-- its own, owned by the geheimtipp package; copying the handful of fields
-- sign-in needs into this one is what keeps authentication from reaching across
-- into it. Everything adoption needs is here for that reason.
--
-- The name is kept in two parts, as the pool holds it, rather than pre-joined:
-- joining is lossless and splitting is not, so the transformation belongs where
-- it is used.
--
-- passwd holds the pool's password IN THE CLEAR. That is deliberate and
-- temporary — it is what lets the sign-in be invisible — and it ends when this
-- table does. It must never be selected into a struct that is logged or
-- returned, and no endpoint may expose it. When the table is empty the
-- migration is complete and it should be dropped.
create table geheimtipp_migration (
    login      text primary key,
    passwd     text        not null,
    vorname    text        not null default '',
    nachname   text        not null default '',
    email      text        not null default '',
    created_at timestamptz not null default now()
);
