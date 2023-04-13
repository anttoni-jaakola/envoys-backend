create table if not exists public.assets
(
    id            serial
        constraint assets_pk
            primary key
        constraint assets_id_key
            unique,
    name          varchar(60),
    symbol        varchar(6),
    min_withdraw  numeric(8, 4)            default 0.0001                      not null,
    max_withdraw  numeric(20, 8)           default 100                         not null,
    min_trade     numeric(8, 4)            default 0.0001                      not null,
    max_trade     numeric(20, 8)           default 1000000                     not null,
    fees_trade    numeric(4, 4)            default 0.1                         not null,
    fees_discount numeric(4, 4)            default 0                           not null,
    fees_charges  numeric(32, 18)          default 0.000000000000000000        not null,
    fees_costs    numeric(32, 18)          default 0.000000000000000000        not null,
    marker        boolean                  default false                       not null,
    chains        jsonb                    default '[]'::jsonb                 not null,
    status        boolean                  default false                       not null,
    "group"       varchar                  default 'crypto'::character varying not null,
    type          varchar                  default 'spot'::character varying   not null,
    create_at     timestamp with time zone default CURRENT_TIMESTAMP
);

alter table public.assets
    owner to envoys;

alter table public.assets
    add unique (symbol);

alter table public.assets
    add unique (id);

create unique index if not exists assets_id_uindex
    on public.assets (id);

create unique index if not exists assets_symbol_uindex
    on public.assets (symbol);

insert into public.assets (id, name, symbol, min_withdraw, max_withdraw, min_trade, max_trade, fees_trade, fees_discount, fees_charges, fees_costs, marker, chains, status, "group", type, create_at)
values  (1, 'Omisego', 'omg', 0.0100, 100.00000000, 0.0100, 1000000.00000000, 0.1500, 0.0500, 0.000000000000000000, 0.000000000000000000, false, '[2]', true, 'crypto', 'spot', '2021-12-26 10:27:02.914683 +00:00'),
        (2, 'Binance', 'bnb', 0.0100, 100.00000000, 0.0010, 1000000.00000000, 0.1500, 0.0500, 0.000000000000000000, 0.000000000000000000, true, '[3, 2]', true, 'crypto', 'spot', '2021-12-26 10:27:02.914683 +00:00'),
        (3, 'Chain Link', 'link', 0.0100, 100.00000000, 0.0100, 1000000.00000000, 0.1500, 0.0500, 0.000000000000000000, 0.000000000000000000, false, '[2]', true, 'crypto', 'spot', '2021-12-26 10:27:02.914683 +00:00'),
        (4, 'Aave', 'aave', 0.0001, 100.00000000, 0.0001, 1000000.00000000, 0.1500, 0.0500, 0.000000000000000000, 0.000000000000000000, false, '[2]', true, 'crypto', 'spot', '2021-12-26 10:27:02.914683 +00:00'),
        (6, 'US Dollar', 'usd', 10.0000, 1000.00000000, 1.0000, 100000000000.00000000, 0.1500, 0.0500, 0.000000000000000000, 0.000000000000000000, true, '[7, 6]', true, 'fiat', 'spot', '2022-08-02 14:18:27.610763 +00:00'),
        (7, 'Euro', 'eur', 0.0001, 100.00000000, 0.0100, 100000000000.00000000, 0.1500, 0.0500, 0.000000000000000000, 0.000000000000000000, true, '[7, 6]', true, 'fiat', 'spot', '2022-06-11 12:23:00.914358 +00:00'),
        (10, 'Ukrainian Hryvnia', 'uah', 0.0001, 100.0000000, 0.0100, 1000000.00000000, 0.1500, 0.0500, 0.000000000000000000, 0.000000000000000000, true, '[6, 7]', true, 'fiat', 'spot', '2022-06-17 14:23:27.806669 +00:00'),
        (11, 'Pound Sterling', 'gbp', 0.0001, 100.00000000, 0.0100, 100000000000.00000000, 0.1500, 0.0500, 0.000000000000000000, 0.000000000000000000, true, '[7, 6]', true, 'fiat', 'spot', '2022-06-11 12:40:55.332645 +00:00'),
        (15, 'Tron', 'trx', 100.0000, 1000000.00000000, 0.0001, 1000000.00000000, 0.1000, 0.0500, 0.000000000000000000, 0.000000000000000000, true, '[1]', true, 'crypto', 'spot', '2021-12-26 10:27:02.914683 +00:00'),
        (5, 'Bitcoin', 'btc', 0.0001, 100.00000000, 0.0001, 1000000.00000000, 0.1500, 0.0500, 0.000000000000000000, 0.000000000000000000, true, '[4, 2]', true, 'crypto', 'spot', '2021-12-26 10:27:02.914683 +00:00'),
        (12, 'Ethereum', 'eth', 0.0010, 100.00000000, 0.0100, 1000000.00000000, 0.1500, 0.0500, 0.000000000000000000, 0.000000000000000000, true, '[3, 2]', true, 'crypto', 'spot', '2021-12-26 10:27:02.914683 +00:00'),
        (14, 'Tether USD', 'usdt', 10.0000, 1000.00000000, 1.0000, 1000000.00000000, 0.1500, 0.0500, 0.000000000000000000, 0.000000000000000000, false, '[1, 2, 3]', true, 'crypto', 'spot', '2021-12-26 10:27:02.914683 +00:00'),
        (16, 'Alphabet Inc Class C', 'goog', 100.0000, 1000000.00000000, 1.0000, 1000000.00000000, 0.0000, 0.0000, 0.000000000000000000, 0.000000000000000000, false, '[]', true, 'action', 'stock', '2023-04-05 20:11:24.727000 +00:00'),
        (17, 'International Business Machines Corporation', 'ibm', 100.0000, 1000000.00000000, 1.0000, 1000000.00000000, 0.0000, 0.0000, 0.000000000000000000, 0.000000000000000000, false, '[]', true, 'action', 'stock', '2023-04-05 20:12:30.562000 +00:00');

select pg_catalog.setval('public.assets_id_seq', 17, true);