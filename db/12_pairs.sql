create table if not exists public.pairs
(
    id            serial
        constraint pairs_pk
            primary key
        constraint pairs_id_key
            unique,
    base_unit     varchar,
    quote_unit    varchar,
    price         numeric(20, 8) default 0.00000000                not null,
    base_decimal  integer        default 2                         not null,
    quote_decimal integer        default 8                         not null,
    type          varchar        default 'spot'::character varying not null,
    status        boolean        default false                     not null
);

alter table public.pairs
    owner to envoys;

alter table public.pairs
    add unique (id);

create unique index if not exists pairs_id_uindex
    on public.pairs (id);

insert into public.pairs (id, base_unit, quote_unit, price, base_decimal, quote_decimal, status, type)
values  (4, 'trx', 'usd', 0.06819350, 6, 2, false, 'spot'),
        (5, 'btc', 'usd', 22689.08400000, 8, 2, false, 'spot'),
        (6, 'link', 'usd', 7.29913333, 6, 2, false, 'spot'),
        (7, 'omg', 'usd', 2.17870000, 6, 2, false, 'spot'),
        (8, 'bnb', 'usd', 304.17666667, 6, 2, false, 'spot'),
        (14, 'trx', 'eth', 0.00003381, 6, 6, true, 'spot'),
        (15, 'btc', 'eth', 15.68550907, 8, 4, true, 'spot'),
        (16, 'eth', 'uah', 72575.42159590, 6, 2, true, 'spot'),
        (17, 'eth', 'eur', 1745.48910890, 6, 2, true, 'spot'),
        (18, 'eth', 'gbp', 1535.04867985, 6, 2, true, 'spot'),
        (19, 'bnb', 'uah', 11801.81412884, 6, 2, true, 'spot'),
        (20, 'usdt', 'uah', 37.39584044, 2, 2, true, 'spot'),
        (22, 'trx', 'usdt', 0.06478058, 2, 8, true, 'spot'),
        (26, 'eth', 'usd', 1915.64464069, 6, 2, true, 'spot'),
        (27, 'goog', 'usd', 105.23999787, 8, 2, true, 'stock'),
        (28, 'ibm', 'usd', 128.58999635, 8, 2, true, 'stock'),
        (1, 'eth', 'link', 264.68774752, 6, 4, true, 'spot'),
        (2, 'eth', 'omg', 1264.22250317, 6, 4, true, 'spot'),
        (3, 'eth', 'bnb', 6.01392309, 2, 4, true, 'spot'),
        (9, 'aave', 'usd', 77.77934891, 6, 2, true, 'spot'),
        (10, 'btc', 'bnb', 94.48256872, 8, 4, true, 'spot'),
        (11, 'eth', 'usdt', 1915.81337882, 8, 2, true, 'spot'),
        (12, 'bnb', 'gbp', 256.19415596, 6, 2, true, 'spot'),
        (13, 'bnb', 'trx', 4958.71188188, 6, 8, true, 'spot');

select pg_catalog.setval('public.pairs_id_seq', 28, true);