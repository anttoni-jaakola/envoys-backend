create table if not exists public.transfers
(
    id         serial
        constraint transfers_pk
            primary key
        constraint transfers_id_key
            unique
        constraint transfers_id_key1
            unique,
    user_id    integer,
    order_id   integer,
    base_unit  varchar,
    quote_unit varchar,
    price      numeric(20, 8),
    quantity   numeric(32, 18),
    assigning  integer,
    fees       double precision,
    create_at  timestamp with time zone default CURRENT_TIMESTAMP not null
);

alter table public.transfers
    owner to envoys;

alter table public.transfers
    add unique (id);

create unique index if not exists transfers_id_uindex
    on public.transfers (id);