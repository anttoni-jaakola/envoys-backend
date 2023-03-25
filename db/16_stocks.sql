create table if not exists public.stocks
(
    id     serial,
    symbol varchar,
    tag    integer        default 0                        not null,
    zone   varchar        default 'usd'::character varying not null,
    price  numeric(20, 8) default 0.00000000               not null,
    status boolean        default false                    not null,
    name   varchar        default ''::character varying    not null
);

alter table public.stocks
    owner to envoys;

alter table public.stocks
    add constraint stocks_pkey
        primary key (id);

alter table public.stocks
    add constraint stocks_symbol_key
        unique (symbol);