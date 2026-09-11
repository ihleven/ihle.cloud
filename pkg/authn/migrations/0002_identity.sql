-- Reshape identity: real constraints on the account table, a password column in
-- place of the credentials blob, and the tables a passkey login needs — one row
-- per credential, server-side ceremony state, revocable sessions, and
-- single-use enrollment links.
--
-- Existing accounts and their password hashes are carried across; nothing is
-- deleted except the two tables that were never wired up.

create extension if not exists pgcrypto;

-- ---------------------------------------------------------------- account ---

alter table account
    add column handle        bytea,
    add column password_hash text,
    add column disabled      boolean     not null default false,
    add column created_at    timestamptz not null default now();

-- A random opaque WebAuthn user handle. The authenticator stores it, so it must
-- not be the email and has to survive an email change.
update account set handle = gen_random_bytes(32) where handle is null;

-- The password moves out of the credentials blob and into its own column. The
-- stored bcrypt hashes are carried across unchanged and keep verifying; the
-- login path re-hashes them to argon2id on the next successful sign-in.
update account set password_hash = credentials ->> 'password' where password_hash is null;

update account set display_name = '' where display_name is null;

alter table account
    alter column hidrive type jsonb using coalesce(hidrive::jsonb, '{}'::jsonb);
alter table account
    alter column hidrive set default '{}'::jsonb,
    alter column hidrive set not null;

alter table account
    alter column name         set not null,
    alter column email        set not null,
    alter column display_name set not null,
    alter column display_name set default '',
    alter column handle       set not null;

-- name is the identity the CMS sees as an entry's owner and the key cmsauth
-- joins on, yet nothing enforced that it was unique.
alter table account
    add constraint account_name_key   unique (name),
    add constraint account_email_key  unique (email),
    add constraint account_handle_key unique (handle);

alter table account
    drop column credentials,
    drop column inv_code,
    drop column inv_expiry;

-- ---------------------------------------------------------------- cmsauth ---

alter table cmsauth add column account_id bigint;

update cmsauth c set account_id = a.id from account a where a.name = c.id;

-- A row that names no account would silently lose someone their permissions, so
-- refuse the migration rather than dropping it.
do $$
begin
    if exists (select 1 from cmsauth where account_id is null) then
        raise exception 'cmsauth rows reference no account: %',
            (select string_agg(id, ', ') from cmsauth where account_id is null);
    end if;
end $$;

update cmsauth set groups      = coalesce(groups, '{}');
update cmsauth set permissions = coalesce(permissions, '{}');

alter table cmsauth drop constraint cmsauth_pkey;
alter table cmsauth drop column id;

alter table cmsauth
    alter column account_id  set not null,
    alter column groups      set not null,
    alter column groups      set default '{}',
    alter column permissions set not null,
    alter column permissions set default '{}';

alter table cmsauth
    add primary key (account_id),
    add constraint cmsauth_account_id_fkey
        foreign key (account_id) references account (id) on delete cascade;

-- --------------------------------------------------------------- passkeys ---

-- One row per credential, so refreshing a sign counter touches one row and a
-- single device can be named and revoked. The credential is stored whole: the
-- library's type is the specification's "credential record" and is built to be
-- persisted and restored intact, so decomposing it would only invite drift.
create table passkey (
    id              bytea       primary key,
    account_id      bigint      not null references account (id) on delete cascade,
    credential      jsonb       not null,
    name            text        not null default '',
    -- The WebAuthn backup-eligible flag, lifted out of the credential so the
    -- device-bound policy is queryable: registration refuses a credential that
    -- can sync, and an audit can find any row that would fail today's rule.
    backup_eligible boolean     not null,
    created_at      timestamptz not null default now(),
    last_used_at    timestamptz
);
create index passkey_account_id_idx on passkey (account_id);

-- Ceremony state. It has to be server-side and single-use: a client-controlled
-- challenge would defeat the point of the signature.
create table webauthn_challenge (
    id         bytea       primary key,
    -- Null for usernameless login, where the account is only known once the
    -- authenticator answers.
    account_id bigint      references account (id) on delete cascade,
    purpose    text        not null check (purpose in ('register', 'login')),
    data       jsonb       not null,
    -- Set when the account re-entered its password to authorise this
    -- registration. Finishing refuses a challenge without it, so the step-up
    -- cannot be skipped by calling the finish endpoint directly.
    reauth_at  timestamptz,
    expires_at timestamptz not null,
    created_at timestamptz not null default now()
);
create index webauthn_challenge_expires_at_idx on webauthn_challenge (expires_at);

-- --------------------------------------------------------------- sessions ---

-- Sessions live server-side so that logging out and revoking access are real
-- rather than advisory. Only the hash of the cookie value is stored.
create table session (
    token_hash bytea       primary key,
    account_id bigint      not null references account (id) on delete cascade,
    expires_at timestamptz not null,
    created_at timestamptz not null default now(),
    last_seen  timestamptz not null default now()
);
create index session_account_id_idx on session (account_id);
create index session_expires_at_idx on session (expires_at);

-- Single-use links for enrolling a passkey, and the recovery path for an
-- account that has lost its keys. Only the hash is stored, so a database leak
-- yields no usable links.
create table enroll_token (
    token_hash bytea       primary key,
    account_id bigint      not null references account (id) on delete cascade,
    expires_at timestamptz not null,
    used_at    timestamptz,
    created_at timestamptz not null default now()
);
create index enroll_token_account_id_idx on enroll_token (account_id);

-- ---------------------------------------------------------------- removals ---

drop table webauthn;
drop table hidrive;
