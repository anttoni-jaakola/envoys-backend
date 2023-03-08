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
values  (1, 'Konotopskiy Aleksandr', 'paymex.center@gmail.com', '', 'vUPtjVOPvsL2-TIoWDioSnIg1WFWMbYEL9rQVgO8oLE=', E'\\xA903C868AE1FECE210190A07C5C1D98B', '[]', '{"spot": ["reserves", "contracts", "pairs", "chains", "currencies", "repayments"], "default": ["accounts", "advertising"]}', false, '', true, '2023-02-17 12:36:36.560573 +00:00'),
        (4, 'Evgeniy', 'Oldgoodbatman@gmail.com', '', 'NJigtctDhgOS950FcjZtPMG01KBTqRv3GtOzQNHrWhU=', E'\\xF947BB76783111A257EDD1526235DFC7', '[]', '{"spot": ["reserves", "contracts", "pairs", "chains", "currencies"], "default": ["accounts", "deny-record", "advertising"]}', false, '', false, '2023-02-27 09:28:15.881397 +00:00'),
        (3, 'Александр', 'alexpro401@gmail.com', '', 'tdcZRpOlgs9sFQTY_Y3Vjz132e_GxOwEm15SUJM81Jc=', E'\\xEF4192674F204608BF45275C95781417', '[]', '{"spot": ["reserves", "contracts", "pairs", "chains", "currencies"], "default": ["deny-record", "advertising", "accounts"]}', false, '', true, '2023-02-27 09:27:07.504725 +00:00'),
        (2, 'Test Account', 'paymex.center2@gmail.com', '', 'vUPtjVOPvsL2-TIoWDioSnIg1WFWMbYEL9rQVgO8oLE=', E'\\x25DE4AC5A98BF203C4DDDF56230C00DD', '[]', '{"spot": ["reserves", "contracts", "pairs", "chains", "currencies"], "default": ["advertising", "accounts"]}', false, '', true, '2023-02-23 11:56:51.035995 +00:00');