create table wallets
(
    id       serial
        constraint wallets_pk
            primary key,
    address  varchar,
    user_id  integer,
    platform integer default 0 not null,
    protocol integer default 0 not null,
    symbol   varchar
);

alter table wallets
    owner to envoys;

create unique index wallets_id_uindex
    on wallets (id);
