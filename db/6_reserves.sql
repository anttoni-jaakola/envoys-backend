create table if not exists public.reserves
(
    id       serial
        constraint reserves_pk
            primary key
        unique,
    user_id  integer,
    address  varchar,
    symbol   varchar,
    platform integer,
    protocol integer,
    value    numeric(32, 18) default 0.000000000000000000 not null,
    reverse  numeric(32, 18) default 0.000000000000000000 not null,
    lock     boolean         default false                not null
);

alter table public.reserves
    owner to envoys;

alter table public.reserves
    add unique (id);

create unique index if not exists reserves_id_uindex
    on public.reserves (id);;