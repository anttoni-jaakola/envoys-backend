create table if not exists public.spot_trades
(
    id         serial,
    base_unit  text,
    quote_unit text,
    price      numeric(20, 8)           default 0.00000000         not null,
    quantity   numeric(32, 19)          default 0.0000000000000000 not null,
    assigning  integer                  default 0                  not null,
    market     boolean                  default false              not null,
    create_at  timestamp with time zone default CURRENT_TIMESTAMP  not null
        constraint spot_trades_create_at_key
            unique
);

alter table public.spot_trades
    owner to envoys;

create index if not exists spot_trades_create_at_idx
    on public.spot_trades (create_at desc);;

select create_hypertable('spot_trades', 'create_at');
