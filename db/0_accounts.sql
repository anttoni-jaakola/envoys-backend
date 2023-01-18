create table if not exists public.spot_accounts
(
    id        serial
        constraint spot_accounts_pk
            primary key
        constraint spot_accounts_id_key
            unique,
    name      varchar(25)              default ''::character varying not null,
    email     varchar                  default ''::character varying not null,
    secure    varchar                  default ''::character varying not null,
    password  varchar,
    entropy   bytea,
    sample    jsonb                    default '[]'::jsonb           not null,
    rules     jsonb                    default '{}'::jsonb           not null,
    status    boolean                  default false                 not null,
    create_at timestamp with time zone default CURRENT_TIMESTAMP
);

alter table public.spot_accounts
    owner to envoys;

create unique index if not exists spot_accounts_email_uindex
    on public.spot_accounts (email);

create unique index if not exists spot_accounts_id_uindex
    on public.spot_accounts (id);

insert into public.spot_accounts (id, name, email, secure, password, entropy, sample, rules, status, create_at)
values  (1, 'Aleksandr Konotopskiy', 'paymex.center@gmail.com', '', 'METU-dMG21lHzumRUuGkF_6WOF6hqXmPz3XxWl7Q4x4=', '0x34C9C7411729A47624E50DEA7E0DF7D5', '["news"]', '{"spot": ["currencies", "chains", "pairs", "contracts", "listing"], "default": ["accounts", "news", "support", "advertising"]}', true, '2022-06-05 12:33:37.234748 +00:00');