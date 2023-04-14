create table if not exists public.orders
(
    id         serial
        constraint orders_pk
            primary key
        constraint orders_id_key
            unique,
    assigning  varchar                  default 'buy'::character varying     not null,
    price      numeric(20, 8)           default 0.00000000                   not null,
    value      numeric(32, 18)          default 0.000000000000000000         not null,
    quantity   numeric(32, 18)          default 0.000000000000000000         not null,
    base_unit  varchar,
    quote_unit varchar,
    user_id    integer,
    extra      jsonb                    default '{}'::jsonb                  not null,
    type       varchar                  default 'spot'::character varying    not null,
    trading    varchar                  default 'limit'::character varying   not null,
    status     varchar                  default 'pending'::character varying not null,
    create_at  timestamp with time zone default CURRENT_TIMESTAMP
);

alter table public.orders
    owner to envoys;

alter table public.orders
    add unique (id);

create unique index if not exists orders_id_uindex
    on public.orders (id);