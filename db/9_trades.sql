create table if not exists public.trades
(
    id         serial
        constraint trades_pk
            primary key
        constraint trades_id_key
            unique
        constraint trades_id_key1
            unique,
    user_id    integer,
    order_id   integer,
    base_unit  varchar,
    quote_unit varchar,
    price      numeric(20, 8),
    quantity   numeric(32, 18),
    assigning  varchar                  default 'buy'::character varying not null,
    fees       double precision,
    maker      boolean                  default false             not null,
    create_at  timestamp with time zone default CURRENT_TIMESTAMP not null
);

alter table public.trades
    owner to envoys;

alter table public.trades
    add unique (id);

create unique index if not exists trades_id_uindex
    on public.trades (id);