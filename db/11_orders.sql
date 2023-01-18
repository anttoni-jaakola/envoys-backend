create table if not exists public.spot_orders
(
    id         serial
        constraint spot_orders_pk
            primary key
        constraint spot_orders_id_key
            unique,
    assigning  integer                  default 0                  not null,
    price      numeric(20, 8)           default 0.00000000         not null,
    value      numeric(32, 19)          default 0.0000000000000000 not null,
    quantity   numeric(32, 19)          default 0.0000000000000000 not null,
    base_unit  varchar,
    quote_unit varchar,
    user_id    integer,
    type       integer                  default 0                  not null,
    status     integer                  default 2                  not null,
    create_at  timestamp with time zone default CURRENT_TIMESTAMP
);

comment on column public.spot_orders.assigning is 'Buy = 0 or Sell = 1 type.';

comment on column public.spot_orders.status is 'CANCEL = 0, FILLED = 1, PENDING = 2';

alter table public.spot_orders
    owner to envoys;

create unique index if not exists spot_orders_id_uindex
    on public.spot_orders (id);