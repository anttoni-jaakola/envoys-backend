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
    status        boolean                  default false                 not null,
    factor_secure boolean                  default false                 not null,
    factor_secret varchar                  default ''::character varying not null,
    kyc_secure    boolean                  default false                 not null,
    kyc_secret    varchar                  default ''::character varying not null,
    kyc_process   boolean                  default false                 not null,
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

insert into public.accounts (id, name, email, email_code, password, entropy, sample, rules, status, factor_secure, factor_secret, kyc_secure, kyc_secret, kyc_process, create_at)
values  (1, 'Konotopskiy Aleksandr', 'paymex.center@gmail.com', '', 'vUPtjVOPvsL2-TIoWDioSnIg1WFWMbYEL9rQVgO8oLE=', E''\\xA903C868AE1FECE210190A07C5C1D98B'', '[]', '{"spot": ["reserves", "contracts", "pairs", "chains", "currencies"], "default": ["accounts", "advertising"]}', true, false, '', false, '', false, '2023-02-17 12:36:36.560573 +00:00');