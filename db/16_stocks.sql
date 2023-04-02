create table if not exists public.stocks
(
    id            serial,
    symbol        varchar,
    "group"       varchar        default 'action'::character varying not null,
    zone          varchar        default 'usd'::character varying    not null,
    price         numeric(20, 8) default 0.00000000                  not null,
    base_decimal  integer        default 8                           not null,
    quote_decimal integer        default 8                           not null,
    name          varchar        default ''::character varying       not null,
    status        boolean        default false                       not null
);

alter table public.stocks
    owner to envoys;

alter table public.stocks
    add constraint stocks_pkey
        primary key (id);

alter table public.stocks
    add constraint stocks_symbol_key
        unique (symbol);

insert into public.stocks (id, symbol, "group", zone, price, status, name, base_decimal, quote_decimal)
values (1, 'goog', 'action', 'usd', 106.07001790, true, 'Alphabet Inc Class C', 8, 8),
       (2, 'ibm', 'action', 'usd', 125.32997456, true, 'International Business Machines Corporation', 8, 8);

select pg_catalog.setval('public.stocks_id_seq', 2, true);