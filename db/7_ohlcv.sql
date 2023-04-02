create extension if not exists timescaledb cascade;
create table if not exists public.ohlcv
(
    id         serial,
    base_unit  varchar,
    quote_unit varchar,
    price      numeric(20, 8)           default 0.00000000                  not null,
    quantity   numeric(32, 18)          default 0.0000000000000000          not null,
    assigning  varchar                  default 'supply'::character varying not null,
    create_at  timestamp with time zone default CURRENT_TIMESTAMP           not null
        constraint ohlcv_create_at_key
            unique
);

alter table public.ohlcv
    owner to envoys;

create index if not exists ohlcv_create_at_idx
    on public.ohlcv (create_at desc);;

select create_hypertable('ohlcv', 'create_at');