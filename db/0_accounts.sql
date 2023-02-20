-- The purpose of this code is to create a table in the public schema named 'accounts' and add constraints to the table.
-- It also adds a unique index to the 'email' and 'id' columns, and sets the owner of the table to 'envoys'.
create table if not exists public.accounts
(
    id            serial
        constraint accounts_pk
            primary key
        unique
        constraint accounts_id_key1
            unique,
    name          varchar(25)              default ''::character varying not null,
    email         varchar                  default ''::character varying not null,
    email_code    varchar                  default ''::character varying not null,
    password      varchar,
    entropy       bytea,
    sample        jsonb                    default '[]'::jsonb           not null,
    rules         jsonb                    default '{}'::jsonb           not null,
    status        boolean                  default false                 not null,
    create_at     timestamp with time zone default CURRENT_TIMESTAMP,
    factor_secure boolean                  default false                 not null,
    factor_secret varchar                  default ''::character varying not null
);

alter table public.accounts
    owner to envoys;

alter table public.accounts
    add unique (id);

alter table public.accounts
    add unique (email);

create unique index accounts_email_uindex
    on accounts (email);

create unique index accounts_id_uindex
    on accounts (id);

insert into public.accounts (id, name, email, email_code, password, entropy, sample, rules, status, create_at, factor_secure, factor_secret)
values  (1, 'Konotopskiy Aleksandr', 'paymex.center@gmail.com', '', 'vUPtjVOPvsL2-TIoWDioSnIg1WFWMbYEL9rQVgO8oLE=', E''\\xA903C868AE1FECE210190A07C5C1D98B'', '[]', '{"spot": ["reserves", "contracts", "pairs", "chains", "currencies"], "default": ["accounts", "advertising"]}', true, '2023-02-17 12:36:36.560573 +00:00', false, '');