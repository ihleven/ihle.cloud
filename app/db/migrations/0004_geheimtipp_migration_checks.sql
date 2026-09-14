-- Make a useless migration row impossible to insert.
--
-- The columns other than login and passwd default to '', so the obvious way to
-- add somebody by hand —
--
--     insert into geheimtipp_migration (login, passwd) values ('paul', 'spacedog');
--
-- — leaves no address. Adoption then refuses, because account.email is unique
-- and is what a commit is signed with, so there is no placeholder worth
-- inventing. The refusal is logged, but what the person sees is "invalid
-- credentials", indistinguishable from a wrong password. That is correct for
-- sign-in and useless for whoever added the row.
--
-- Better to fail where the mistake is made. An address is required, and so is a
-- password: a row with neither cannot ever produce an account, so it is not a
-- row worth having.
--
-- NOT VALID so that rows already added are left as they are. They still cannot
-- be adopted — nothing about the refusal changes — but this is a table of
-- people, and quietly deleting some of them to satisfy a constraint would be a
-- worse trade than leaving them to be noticed.

alter table geheimtipp_migration
    add constraint geheimtipp_migration_email_present check (email <> '') not valid;

alter table geheimtipp_migration
    add constraint geheimtipp_migration_passwd_present check (passwd <> '') not valid;
