create table if not exists public.wallets
(
    id       serial
        constraint wallets_pk
            primary key
        constraint wallets_id_key
            unique
        constraint wallets_id_key1
            unique,
    address  varchar,
    user_id  integer,
    platform integer default 0 not null,
    protocol integer default 0 not null,
    symbol   varchar
);

alter table public.wallets
    owner to envoys;

alter table public.wallets
    add unique (id);

create unique index if not exists wallets_id_uindex
    on public.wallets (id);