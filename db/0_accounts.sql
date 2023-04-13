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
values  (2, 'Test Account', 'paymex.center2@gmail.com', '', 'vUPtjVOPvsL2-TIoWDioSnIg1WFWMbYEL9rQVgO8oLE=', E'\\xD186B2BA717227426E3D17BBE345FFBE', '[]', '{"spot": ["reserves", "contracts", "pairs", "chains", "assets", "repayments", "listing"], "market": ["pairs", "assets"], "default": ["accounts", "advertising"]}', false, '', true, '2023-04-03 08:56:56.430222 +00:00'),
        (4, 'Sergey', 'const.subject@gmail.com', '', 'Hd5Oz4qXK6q3yFG1JfFgILSa6rgPmi4qKrIiIs3Y-44=', E'\\xB1680E6005110E807148F1C212B0FE60', '[]', '{"spot": ["reserves", "contracts", "pairs", "chains", "assets", "repayments", "listing"], "market": ["pairs", "assets"], "default": ["accounts", "advertising"]}', false, '', true, '2023-04-11 17:35:17.921736 +00:00'),
        (5, 'Dmytro', 'dmytro.taran.dev@gmail.com', '', '7Zn7p82Fkj_bn26rIsGkNMkF1Vg-ZmSNgyqSlTj2i3g=', E'\\xFD7EE762E1CA03DF1603E824B41656F6', '[]', '{"spot": ["reserves", "contracts", "pairs", "chains", "assets", "repayments", "listing"], "market": ["pairs", "assets"], "default": ["accounts", "advertising"]}', false, '', true, '2023-04-12 13:50:51.949103 +00:00'),
        (3, 'Aleksandr', 'alexpro401@gmail.com', '', 'tdcZRpOlgs9sFQTY_Y3Vjz132e_GxOwEm15SUJM81Jc=', E'\\x0AE33D4DB135F00C25C8DA5EEFCCA47B', '[]', '{"spot": ["reserves", "contracts", "pairs", "chains", "assets", "repayments", "listing"], "market": ["pairs", "assets"], "default": ["accounts", "advertising"]}', false, '', true, '2023-04-10 08:02:11.934437 +00:00'),
        (1, 'Konotopskiy Aleksandr', 'paymex.center@gmail.com', '', 'vUPtjVOPvsL2-TIoWDioSnIg1WFWMbYEL9rQVgO8oLE=', E'\\xA903C868AE1FECE210190A07C5C1D98B', '[]', '{"spot": ["reserves", "contracts", "pairs", "chains", "assets", "repayments", "listing"], "market": ["pairs", "assets"], "default": ["accounts", "advertising"]}', false, '', true, '2023-02-17 12:36:36.560573 +00:00');

select pg_catalog.setval('public.accounts_id_seq', 5, true);