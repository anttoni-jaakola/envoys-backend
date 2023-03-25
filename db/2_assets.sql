create table if not exists public.assets
(
    id      serial
        constraint assets_pk
            primary key
        constraint assets_id_key
            unique,
    user_id integer,
    symbol  varchar,
    balance numeric(32, 18) default 0.000000000000000000 not null,
    type    integer         default 0                    not null
);

alter table public.assets
    owner to envoys;

alter table public.assets
    add unique (id);

create unique index if not exists assets_id_uindex
    on public.assets (id);