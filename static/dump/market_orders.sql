create table market_orders
(
    symbol    varchar,
    uid       integer,
    price     numeric(32, 16)          default 0.0000000000000000 not null,
    size      numeric(32, 16)          default 0.0000000000000000,
    side      varchar,
    type      varchar,
    cid       integer,
    create_at timestamp with time zone default CURRENT_TIMESTAMP,
    id        varchar
        unique,
    volume    numeric(32, 16)          default 0.0000000000000000
);

alter table market_orders
    owner to envoys;