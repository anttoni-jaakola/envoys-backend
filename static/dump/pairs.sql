create table pairs
(
    id            serial
        constraint pairs_pk
            primary key,
    base_unit     varchar,
    quote_unit    varchar,
    price         numeric(20, 8) default 0.00000000 not null,
    base_decimal  integer        default 2          not null,
    quote_decimal integer        default 8          not null,
    status        boolean        default false      not null
);

alter table pairs
    owner to envoys;

create unique index pairs_id_uindex
    on pairs (id);

INSERT INTO public.pairs (id, base_unit, quote_unit, price, base_decimal, quote_decimal, status) VALUES (1, 'eth', 'link', 220.16359184, 6, 4, false);
INSERT INTO public.pairs (id, base_unit, quote_unit, price, base_decimal, quote_decimal, status) VALUES (2, 'eth', 'omg', 736.91967577, 6, 4, false);
INSERT INTO public.pairs (id, base_unit, quote_unit, price, base_decimal, quote_decimal, status) VALUES (3, 'eth', 'bnb', 5.27048079, 2, 4, false);
INSERT INTO public.pairs (id, base_unit, quote_unit, price, base_decimal, quote_decimal, status) VALUES (4, 'trx', 'usd', 0.06819350, 6, 2, false);
INSERT INTO public.pairs (id, base_unit, quote_unit, price, base_decimal, quote_decimal, status) VALUES (5, 'btc', 'usd', 22689.08400000, 8, 2, false);
INSERT INTO public.pairs (id, base_unit, quote_unit, price, base_decimal, quote_decimal, status) VALUES (6, 'link', 'usd', 7.29913333, 6, 2, false);
INSERT INTO public.pairs (id, base_unit, quote_unit, price, base_decimal, quote_decimal, status) VALUES (7, 'omg', 'usd', 2.17870000, 6, 2, false);
INSERT INTO public.pairs (id, base_unit, quote_unit, price, base_decimal, quote_decimal, status) VALUES (8, 'bnb', 'usd', 304.17666667, 6, 2, false);
INSERT INTO public.pairs (id, base_unit, quote_unit, price, base_decimal, quote_decimal, status) VALUES (9, 'aave', 'usd', 95.36833333, 6, 2, false);
INSERT INTO public.pairs (id, base_unit, quote_unit, price, base_decimal, quote_decimal, status) VALUES (10, 'btc', 'bnb', 138.33670553, 8, 4, false);
INSERT INTO public.pairs (id, base_unit, quote_unit, price, base_decimal, quote_decimal, status) VALUES (11, 'eth', 'usdt', 1609.71294407, 2, 8, false);
INSERT INTO public.pairs (id, base_unit, quote_unit, price, base_decimal, quote_decimal, status) VALUES (12, 'bnb', 'gbp', 189.00000000, 6, 2, false);
INSERT INTO public.pairs (id, base_unit, quote_unit, price, base_decimal, quote_decimal, status) VALUES (13, 'bnb', 'trx', 3412.38216665, 6, 8, false);
INSERT INTO public.pairs (id, base_unit, quote_unit, price, base_decimal, quote_decimal, status) VALUES (14, 'trx', 'eth', 0.00004233, 6, 6, false);
INSERT INTO public.pairs (id, base_unit, quote_unit, price, base_decimal, quote_decimal, status) VALUES (15, 'btc', 'eth', 14.21413631, 8, 4, false);
INSERT INTO public.pairs (id, base_unit, quote_unit, price, base_decimal, quote_decimal, status) VALUES (16, 'eth', 'uah', 63020.19688061, 6, 2, false);
INSERT INTO public.pairs (id, base_unit, quote_unit, price, base_decimal, quote_decimal, status) VALUES (17, 'eth', 'eur', 1577.38801106, 6, 2, false);
INSERT INTO public.pairs (id, base_unit, quote_unit, price, base_decimal, quote_decimal, status) VALUES (18, 'eth', 'gbp', 1329.20534028, 6, 2, false);
INSERT INTO public.pairs (id, base_unit, quote_unit, price, base_decimal, quote_decimal, status) VALUES (19, 'bnb', 'uah', 11679.32986432, 6, 2, false);
INSERT INTO public.pairs (id, base_unit, quote_unit, price, base_decimal, quote_decimal, status) VALUES (20, 'usdt', 'uah', 39.18388945, 2, 2, false);
INSERT INTO public.pairs (id, base_unit, quote_unit, price, base_decimal, quote_decimal, status) VALUES (21, 'eth', 'tst', 2548.32508030, 8, 4, true);
INSERT INTO public.pairs (id, base_unit, quote_unit, price, base_decimal, quote_decimal, status) VALUES (22, 'trx', 'usdt', 0.06144743, 2, 8, true);
INSERT INTO public.pairs (id, base_unit, quote_unit, price, base_decimal, quote_decimal, status) VALUES (23, 'rub', 'kgs', 1.31000000, 2, 8, true);
INSERT INTO public.pairs (id, base_unit, quote_unit, price, base_decimal, quote_decimal, status) VALUES (24, 'eur', 'rub', 62.42266735, 2, 4, true);
INSERT INTO public.pairs (id, base_unit, quote_unit, price, base_decimal, quote_decimal, status) VALUES (25, 'usd', 'rub', 62.49988306, 2, 4, true);
INSERT INTO public.pairs (id, base_unit, quote_unit, price, base_decimal, quote_decimal, status) VALUES (26, 'eth', 'usd', 1309.12558276, 6, 2, true);
INSERT INTO public.pairs (id, base_unit, quote_unit, price, base_decimal, quote_decimal, status) VALUES (27, 'kgs', 'usd', 0.01200000, 2, 6, true);

SELECT pg_catalog.setval('public.pairs_id_seq', 27, true);