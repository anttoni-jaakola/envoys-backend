create extension if not exists timescaledb cascade;
create table if not exists public.trades
(
    id         serial,
    base_unit  text,
    quote_unit text,
    price      numeric(20, 8)           default 0.00000000         not null,
    quantity   numeric(32, 18)          default 0.0000000000000000 not null,
    assigning  integer                  default 0                  not null,
    market     boolean                  default false              not null,
    create_at  timestamp with time zone default CURRENT_TIMESTAMP  not null
        constraint trades_create_at_key
            unique
);

alter table public.trades
    owner to envoys;

create index if not exists trades_create_at_idx
    on public.trades (create_at desc);;

select create_hypertable('trades', 'create_at');