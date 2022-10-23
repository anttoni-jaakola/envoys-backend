create table currencies
(
    id            serial
        constraint currencies_pk
            primary key,
    name          varchar(40),
    symbol        varchar(6),
    min_withdraw  numeric(8, 4)            default 0.0001             not null,
    max_withdraw  numeric(20, 8)           default 100                not null,
    min_deposit   numeric(8, 4)            default 0.0001             not null,
    min_trade     numeric(8, 4)            default 0.0001             not null,
    max_trade     numeric(20, 8)           default 1000000            not null,
    fees_trade    numeric(4, 4)            default 0.1                not null,
    fees_discount numeric(4, 4)            default 0                  not null,
    fees_charges  numeric(32, 16)          default 0.0000000000000000 not null,
    fees_costs    numeric(32, 16)          default 0.0000000000000000 not null,
    marker        boolean                  default false              not null,
    chains        jsonb                    default '[]'::jsonb        not null,
    status        boolean                  default false              not null,
    fin_type      integer                  default 0                  not null,
    create_at     timestamp with time zone default CURRENT_TIMESTAMP
);

comment on column currencies.fin_type is '0 - crypto
1 - fiat';

alter table currencies
    owner to envoys;

create unique index currencies_id_uindex
    on currencies (id);

create unique index currencies_symbol_uindex
    on currencies (symbol);

INSERT INTO public.currencies (id, name, symbol, min_withdraw, max_withdraw, min_deposit, min_trade, max_trade, fees_trade, fees_discount, fees_charges, fees_costs, marker, chains, status, fin_type, create_at) VALUES (1, 'Omisego', 'omg', 0.0100, 100.00000000, 0.0100, 0.0100, 1000000.00000000, 0.1500, 0.0500, 0.0000000000000000, 0.0000000000000000, false, '[2]', true, 0, '2021-12-26 10:27:02.914683 +00:00');
INSERT INTO public.currencies (id, name, symbol, min_withdraw, max_withdraw, min_deposit, min_trade, max_trade, fees_trade, fees_discount, fees_charges, fees_costs, marker, chains, status, fin_type, create_at) VALUES (2, 'Binance', 'bnb', 0.0100, 100.00000000, 0.0100, 0.0010, 1000000.00000000, 0.1500, 0.0500, 0.0000000000000000, 0.0000000000000000, true, '[3, 2]', true, 0, '2021-12-26 10:27:02.914683 +00:00');
INSERT INTO public.currencies (id, name, symbol, min_withdraw, max_withdraw, min_deposit, min_trade, max_trade, fees_trade, fees_discount, fees_charges, fees_costs, marker, chains, status, fin_type, create_at) VALUES (3, 'Chain Link', 'link', 0.0100, 100.00000000, 0.0100, 0.0100, 1000000.00000000, 0.1500, 0.0500, 0.0000000000000000, 0.0000000000000000, false, '[2]', true, 0, '2021-12-26 10:27:02.914683 +00:00');
INSERT INTO public.currencies (id, name, symbol, min_withdraw, max_withdraw, min_deposit, min_trade, max_trade, fees_trade, fees_discount, fees_charges, fees_costs, marker, chains, status, fin_type, create_at) VALUES (4, 'Aave', 'aave', 0.0001, 100.00000000, 0.0001, 0.0001, 1000000.00000000, 0.1500, 0.0500, 0.0000000000000000, 0.0000000000000000, false, '[2]', true, 0, '2021-12-26 10:27:02.914683 +00:00');
INSERT INTO public.currencies (id, name, symbol, min_withdraw, max_withdraw, min_deposit, min_trade, max_trade, fees_trade, fees_discount, fees_charges, fees_costs, marker, chains, status, fin_type, create_at) VALUES (5, 'Bitcoin', 'btc', 0.0001, 100.00000000, 0.0001, 0.0001, 1000000.00000000, 0.1500, 0.0500, 0.0000000000000000, 0.0000000000000000, true, '[4, 2]', true, 0, '2021-12-26 10:27:02.914683 +00:00');
INSERT INTO public.currencies (id, name, symbol, min_withdraw, max_withdraw, min_deposit, min_trade, max_trade, fees_trade, fees_discount, fees_charges, fees_costs, marker, chains, status, fin_type, create_at) VALUES (6, 'US Dollar', 'usd', 10.0000, 1000.00000000, 5.0000, 1.0000, 100000000000.00000000, 0.1500, 0.0500, 0.0000000000000000, 0.0000000000000000, true, '[7, 6]', true, 1, '2022-08-02 14:18:27.610763 +00:00');
INSERT INTO public.currencies (id, name, symbol, min_withdraw, max_withdraw, min_deposit, min_trade, max_trade, fees_trade, fees_discount, fees_charges, fees_costs, marker, chains, status, fin_type, create_at) VALUES (7, 'Euro', 'eur', 0.0001, 100.00000000, 0.0100, 0.0100, 100000000000.00000000, 0.1500, 0.0500, 0.0000000000000000, 0.0000000000000000, true, '[7, 6]', true, 1, '2022-06-11 12:23:00.914358 +00:00');
INSERT INTO public.currencies (id, name, symbol, min_withdraw, max_withdraw, min_deposit, min_trade, max_trade, fees_trade, fees_discount, fees_charges, fees_costs, marker, chains, status, fin_type, create_at) VALUES (8, 'Russian ruble', 'rub', 0.0001, 1000000.00000000, 1000.0000, 400.0000, 1000000.00000000, 0.1500, 0.0500, 0.0000000000000000, 0.0000000000000000, false, '[6, 7]', true, 1, '2022-10-10 09:24:01.773841 +00:00');
INSERT INTO public.currencies (id, name, symbol, min_withdraw, max_withdraw, min_deposit, min_trade, max_trade, fees_trade, fees_discount, fees_charges, fees_costs, marker, chains, status, fin_type, create_at) VALUES (9, 'Kyrgyzstan Som', 'kgs', 0.0001, 100.00000000, 10.0000, 1.0000, 100000000000.00000000, 0.1500, 0.0500, 0.0000000000000000, 0.0000000000000000, true, '[6, 7]', true, 1, '2022-08-02 13:47:53.279849 +00:00');
INSERT INTO public.currencies (id, name, symbol, min_withdraw, max_withdraw, min_deposit, min_trade, max_trade, fees_trade, fees_discount, fees_charges, fees_costs, marker, chains, status, fin_type, create_at) VALUES (10, 'Ukrainian Hryvnia', 'uah', 0.0001, 100.00000000, 0.0100, 0.0100, 1000000.00000000, 0.1500, 0.0500, 0.0000000000000000, 0.0000000000000000, true, '[6, 7]', true, 1, '2022-06-17 14:23:27.806669 +00:00');
INSERT INTO public.currencies (id, name, symbol, min_withdraw, max_withdraw, min_deposit, min_trade, max_trade, fees_trade, fees_discount, fees_charges, fees_costs, marker, chains, status, fin_type, create_at) VALUES (11, 'Pound Sterling', 'gbp', 0.0001, 100.00000000, 0.0100, 0.0100, 100000000000.00000000, 0.1500, 0.0500, 0.0000000000000000, 0.0000000000000000, true, '[7, 6]', true, 1, '2022-06-11 12:40:55.332645 +00:00');
INSERT INTO public.currencies (id, name, symbol, min_withdraw, max_withdraw, min_deposit, min_trade, max_trade, fees_trade, fees_discount, fees_charges, fees_costs, marker, chains, status, fin_type, create_at) VALUES (12, 'Ethereum', 'eth', 0.0010, 100.00000000, 0.0100, 0.0100, 1000000.00000000, 0.1500, 0.0500, 0.0000000000000000, 0.0000000000000000, true, '[3, 2]', true, 0, '2021-12-26 10:27:02.914683 +00:00');
INSERT INTO public.currencies (id, name, symbol, min_withdraw, max_withdraw, min_deposit, min_trade, max_trade, fees_trade, fees_discount, fees_charges, fees_costs, marker, chains, status, fin_type, create_at) VALUES (13, 'Test Token', 'tst', 100.0000, 100000.00000000, 100.0000, 100.0000, 10000000000.00000000, 0.1500, 0.0500, 0.0000000000000000, 0.0000000000000000, false, '[2]', true, 0, '2022-08-07 10:50:25.489160 +00:00');
INSERT INTO public.currencies (id, name, symbol, min_withdraw, max_withdraw, min_deposit, min_trade, max_trade, fees_trade, fees_discount, fees_charges, fees_costs, marker, chains, status, fin_type, create_at) VALUES (14, 'Tether USD', 'usdt', 10.0000, 1000.00000000, 5.0000, 1.0000, 1000000.00000000, 0.1500, 0.0500, 0.0000000000000000, 0.0000000000000000, false, '[1, 2, 3]', true, 0, '2021-12-26 10:27:02.914683 +00:00');
INSERT INTO public.currencies (id, name, symbol, min_withdraw, max_withdraw, min_deposit, min_trade, max_trade, fees_trade, fees_discount, fees_charges, fees_costs, marker, chains, status, fin_type, create_at) VALUES (15, 'Tron', 'trx', 100.0000, 1000000.00000000, 0.0100, 0.0001, 1000000.00000000, 0.1500, 0.0500, 0.0000000000000000, 0.0000000000000000, true, '[1]', true, 0, '2021-12-26 10:27:02.914683 +00:00');

SELECT pg_catalog.setval('public.currencies_id_seq', 15, true);