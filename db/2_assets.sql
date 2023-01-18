create table if not exists public.spot_assets
(
    id      serial
        constraint spot_assets_pk
            primary key
        constraint spot_assets_id_key
            unique,
    user_id integer,
    symbol  varchar,
    balance numeric(32, 19) default 0.0000000000000000 not null
);

alter table public.spot_assets
    owner to envoys;

create unique index if not exists spot_assets_id_uindex
    on public.spot_assets (id);