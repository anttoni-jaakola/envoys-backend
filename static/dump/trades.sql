create table trades
(
    id         serial,
    base_unit  varchar,
    quote_unit varchar,
    price      numeric(20, 8)           default 0.00000000         not null,
    quantity   numeric(32, 16)          default 0.0000000000000000 not null,
    assigning  integer                  default 0                  not null,
    market     boolean                  default false              not null,
    ask_price  numeric(20, 8)           default 0.00000000         not null,
    bid_price  numeric(20, 8)           default 0.00000000         not null,
    create_at  timestamp with time zone default CURRENT_TIMESTAMP  not null
);

alter table trades
    owner to envoys;

create index trades_create_at_idx
    on trades (create_at desc);

