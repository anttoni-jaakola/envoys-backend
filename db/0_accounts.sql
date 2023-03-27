create table if not exists public.accounts
(
    id            serial
        constraint accounts_pk
            primary key
        unique,
    name          varchar(25)              default ''::character varying not null,
    email         varchar                  default ''::character varying not null
        unique,
    email_code    varchar                  default ''::character varying not null,
    password      varchar,
    entropy       bytea,
    sample        jsonb                    default '[]'::jsonb           not null,
    rules         jsonb                    default '{}'::jsonb           not null,
    factor_secure boolean                  default false                 not null,
    factor_secret varchar                  default ''::character varying not null,
    status        boolean                  default false                 not null,
    create_at     timestamp with time zone default CURRENT_TIMESTAMP
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

insert into public.accounts (id, name, email, email_code, password, entropy, sample, rules, factor_secure, factor_secret, status, create_at)
values  (1, 'Konotopskiy Aleksandr', 'paymex.center@gmail.com', '', 'vUPtjVOPvsL2-TIoWDioSnIg1WFWMbYEL9rQVgO8oLE=', E'\\xA903C868AE1FECE210190A07C5C1D98B', '[]', '{"spot": ["reserves", "contracts", "pairs", "chains", "assets", "repayments"], "default": ["accounts", "advertising"]}', false, '', true, '2023-02-17 12:36:36.560573 +00:00');

select pg_catalog.setval('public.accounts_id_seq', 1, true);