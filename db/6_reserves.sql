create table if not exists public.spot_reserves
(
    id       serial
        constraint spot_reserves_pk
            primary key
        constraint spot_reserves_id_key
            unique,
    user_id  integer,
    address  varchar,
    symbol   varchar,
    platform integer,
    protocol integer,
    value    numeric(32, 18) default 0.0000000000000000 not null,
    lock     boolean         default false              not null
);

alter table public.spot_reserves
    owner to envoys;

create unique index if not exists spot_reserves_id_uindex
    on public.spot_reserves (id);;