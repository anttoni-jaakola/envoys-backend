create table if not exists public.spot_wallets
(
    id       serial
        constraint spot_wallets_pk
            primary key
        constraint spot_wallets_id_key
            unique
        constraint spot_wallets_id_key1
            unique,
    address  varchar,
    user_id  integer,
    platform integer default 0 not null,
    protocol integer default 0 not null,
    symbol   varchar
);

alter table public.spot_wallets
    owner to envoys;

create unique index if not exists spot_wallets_id_uindex
    on public.spot_wallets (id);