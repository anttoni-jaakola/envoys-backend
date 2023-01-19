create table if not exists public.spot_transactions
(
    id           serial
        constraint spot_transactions_pk
            primary key
        constraint spot_transactions_id_key
            unique
        constraint spot_transactions_id_key1
            unique,
    symbol       varchar,
    hash         varchar                  default ''::character varying not null,
    value        numeric(32, 18),
    fees         double precision         default 0                     not null,
    confirmation integer                  default 0                     not null,
    "to"         varchar                  default ''::character varying not null,
    block        integer                  default 0                     not null,
    chain_id     integer                  default 0                     not null,
    user_id      integer,
    tx_type      integer                  default 0                     not null,
    fin_type     integer                  default 0                     not null,
    platform     integer,
    protocol     integer,
    claim        boolean                  default false                 not null,
    price        double precision         default 0                     not null,
    status       integer                  default 2                     not null,
    create_at    timestamp with time zone default CURRENT_TIMESTAMP
);

alter table public.spot_transactions
    owner to envoys;

create unique index if not exists spot_transactions_id_uindex
    on public.spot_transactions (id);;