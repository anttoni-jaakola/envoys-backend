create table if not exists public.spot_pairs
(
    id            serial
        constraint spot_pairs_pk
            primary key
        constraint spot_pairs_id_key
            unique,
    base_unit     varchar,
    quote_unit    varchar,
    price         numeric(20, 8) default 0.00000000 not null,
    base_decimal  integer        default 2          not null,
    quote_decimal integer        default 8          not null,
    status        boolean        default false      not null
);

alter table public.spot_pairs
    owner to envoys;

create unique index if not exists spot_pairs_id_uindex
    on public.spot_pairs (id);

insert into public.spot_pairs (id, base_unit, quote_unit, price, base_decimal, quote_decimal, status)
values  (1, 'eth', 'link', 220.16359184, 6, 4, false),
        (2, 'eth', 'omg', 736.91967577, 6, 4, false),
        (3, 'eth', 'bnb', 5.27048079, 2, 4, false),
        (4, 'trx', 'usd', 0.06819350, 6, 2, false),
        (5, 'btc', 'usd', 22689.08400000, 8, 2, false),
        (6, 'link', 'usd', 7.29913333, 6, 2, false),
        (7, 'omg', 'usd', 2.17870000, 6, 2, false),
        (8, 'bnb', 'usd', 304.17666667, 6, 2, false),
        (9, 'aave', 'usd', 95.36833333, 6, 2, false),
        (10, 'btc', 'bnb', 138.33670553, 8, 4, false),
        (11, 'eth', 'usdt', 1609.71294407, 2, 8, false),
        (12, 'bnb', 'gbp', 189.00000000, 6, 2, false),
        (13, 'bnb', 'trx', 3412.38216665, 6, 8, false),
        (14, 'trx', 'eth', 0.00004233, 6, 6, false),
        (15, 'btc', 'eth', 14.21413631, 8, 4, false),
        (16, 'eth', 'uah', 63020.19688061, 6, 2, false),
        (17, 'eth', 'eur', 1577.38801106, 6, 2, false),
        (18, 'eth', 'gbp', 1329.20534028, 6, 2, false),
        (19, 'bnb', 'uah', 11679.32986432, 6, 2, false),
        (20, 'usdt', 'uah', 39.18388945, 2, 2, false),
        (26, 'eth', 'usd', 1182.07705187, 6, 2, true),
        (21, 'eth', 'tst', 2552.66167413, 8, 4, true),
        (22, 'trx', 'usdt', 0.05367425, 2, 8, true);