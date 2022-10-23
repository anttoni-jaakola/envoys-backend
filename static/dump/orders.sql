create table orders
(
    id         serial
        constraint orders_pk
            primary key,
    assigning  integer                  default 0                  not null,
    price      numeric(20, 8)           default 0.00000000         not null,
    value      numeric(32, 16)          default 0.0000000000000000 not null,
    quantity   numeric(32, 16)          default 0.0000000000000000 not null,
    base_unit  varchar,
    quote_unit varchar,
    user_id    integer,
    type       integer                  default 0                  not null,
    status     integer                  default 2                  not null,
    create_at  timestamp with time zone default CURRENT_TIMESTAMP
);

comment on column orders.assigning is 'Buy = 0 or Sell = 1 type.';

comment on column orders.status is 'CANCEL = 0, FILLED = 1, PENDING = 2';

alter table orders
    owner to envoys;

create unique index orders_id_uindex
    on orders (id);
