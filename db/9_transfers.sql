create table if not exists public.spot_transfers
(
    id         serial
        constraint spot_transfers_pk
            primary key
        constraint spot_transfers_id_key
            unique
        constraint spot_transfers_id_key1
            unique,
    user_id    integer,
    order_id   integer,
    base_unit  varchar,
    quote_unit varchar,
    price      numeric(20, 8),
    quantity   numeric(32, 19),
    assigning  integer,
    fees       double precision,
    create_at  timestamp with time zone default CURRENT_TIMESTAMP not null
);

alter table public.spot_transfers
    owner to envoys;

create unique index if not exists spot_transfers_id_uindex
    on public.spot_transfers (id);