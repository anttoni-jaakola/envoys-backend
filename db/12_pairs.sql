create table if not exists public.pairs
(
    id            serial
        constraint pairs_pk
            primary key
        constraint pairs_id_key
            unique,
    base_unit     varchar,
    quote_unit    varchar,
    price         numeric(20, 8) default 0.00000000 not null,
    base_decimal  integer        default 2          not null,
    quote_decimal integer        default 8          not null,
    status        boolean        default false      not null
);

alter table public.pairs
    owner to envoys;

alter table public.pairs
    add unique (id);

create unique index if not exists pairs_id_uindex
    on public.pairs (id);

insert into public.pairs (id, base_unit, quote_unit, price, base_decimal, quote_decimal, status)
values  (4, 'trx', 'usd', 0.06819350, 6, 2, false),
        (5, 'btc', 'usd', 22689.08400000, 8, 2, false),
        (6, 'link', 'usd', 7.29913333, 6, 2, false),
        (7, 'omg', 'usd', 2.17870000, 6, 2, false),
        (8, 'bnb', 'usd', 304.17666667, 6, 2, false),
        (22, 'trx', 'usdt', 0.07156860, 2, 8, true),
        (26, 'eth', 'usd', 1709.47799584, 6, 2, true),
        (1, 'eth', 'link', 209.89255124, 6, 4, true),
        (2, 'eth', 'omg', 891.62499457, 6, 4, true),
        (3, 'eth', 'bnb', 5.38724998, 2, 4, true),
        (9, 'aave', 'usd', 91.46255714, 6, 2, true),
        (10, 'btc', 'bnb', 78.51044107, 8, 4, true),
        (11, 'eth', 'usdt', 1709.71932009, 8, 2, true),
        (12, 'bnb', 'gbp', 263.40453237, 6, 2, true),
        (13, 'bnb', 'trx', 4351.31478038, 6, 8, true),
        (14, 'trx', 'eth', 0.00004187, 6, 6, true),
        (15, 'btc', 'eth', 14.60122795, 8, 4, true),
        (16, 'eth', 'uah', 67179.27764229, 6, 2, true),
        (17, 'eth', 'eur', 1600.24794456, 6, 2, true),
        (18, 'eth', 'gbp', 1420.53327339, 6, 2, true),
        (19, 'bnb', 'uah', 12608.72073956, 6, 2, true),
        (20, 'usdt', 'uah', 39.69596824, 2, 2, true);

select pg_catalog.setval('public.pairs_id_seq', 22, true);